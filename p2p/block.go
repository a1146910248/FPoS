package p2p

import (
	. "FPoS/types"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"time"
)

// 处理新区块
func (n *Layer2Node) processNewBlock(block Block) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 检查区块高度
	if block.Height <= n.latestBlock {
		return fmt.Errorf("block height %d is not higher than current height %d",
			block.Height, n.latestBlock)
	}

	// 应用交易前，清理这些交易相关的待处理状态
	for _, tx := range block.Transactions {
		n.stateDB.CleanPendingState(tx.From)
	}

	// 应用交易
	if err := n.applyTransactions(block.Transactions); err != nil {
		return err
	}

	// 更新状态
	n.latestBlock = block.Height
	n.stateRoot = block.StateRoot
	n.blockCache.Store(block.Height, block)

	// 从交易池中移除已处理的交易
	for _, tx := range block.Transactions {
		n.txPool.Delete(tx.Hash)
	}

	return nil
}

// 广播交易
func (n *Layer2Node) BroadcastTransaction(tx Transaction) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	return n.txTopic.Publish(n.ctx, data)
}

// 广播区块
func (n *Layer2Node) BroadcastBlock(block Block) error {
	data, err := json.Marshal(block)
	if err != nil {
		return err
	}

	return n.blockTopic.Publish(n.ctx, data)
}

// 请求状态同步
func (n *Layer2Node) RequestSync(fromHeight uint64) error {
	req := SyncRequest{
		FromHeight: fromHeight,
		ToHeight:   n.latestBlock,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	msg := Message{
		Type:    MsgTypeSyncRequest,
		Payload: data,
	}

	// 向所有连接的节点发送同步请求
	peers := n.host.Network().Peers()
	for _, peer := range peers {
		if err := n.sendMessage(peer, msg); err != nil {
			fmt.Printf("Failed to send sync request to peer %s: %s\n", peer, err)
		}
	}

	return nil
}

func (n *Layer2Node) applyTransactions(txs []Transaction) error {
	for _, tx := range txs {
		// 执行交易，更新账户状态
		if err := n.stateDB.ExecuteTransaction(&tx); err != nil {
			return fmt.Errorf("failed to execute transaction: %w", err)
		}
	}
	return nil
}

func (n *Layer2Node) updateState(stateRoot string, height uint64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.stateRoot = stateRoot
	n.latestBlock = height
}

func SignBlock(block *Block, node *Layer2Node) error {
	// 序列化区块数据
	blockData := struct {
		Height       uint64
		Timestamp    time.Time
		Transactions []Transaction
		PreviousHash string
		StateRoot    string
		Proposer     string
		GasUsed      uint64
		GasLimit     uint64
	}{
		Height:       block.Height,
		Timestamp:    block.Timestamp,
		Transactions: block.Transactions,
		PreviousHash: block.PreviousHash,
		StateRoot:    block.StateRoot,
		Proposer:     block.Proposer,
		GasUsed:      block.GasUsed,
		GasLimit:     block.GasLimit,
	}

	data, err := json.Marshal(blockData)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// 使用节点私钥签名
	signature, err := node.privateKey.Sign(data)
	if err != nil {
		return fmt.Errorf("failed to sign block: %w", err)
	}

	block.Signature = signature
	return nil
}

// 验证区块签名的方法
func VerifyBlockSignature(block *Block, n *Layer2Node) error {
	// 重建区块数据
	blockData, err := json.Marshal(struct {
		Height       uint64
		Timestamp    time.Time
		Transactions []Transaction
		PreviousHash string
		StateRoot    string
		Proposer     string
		GasUsed      uint64
		GasLimit     uint64
	}{
		Height:       block.Height,
		Timestamp:    block.Timestamp,
		Transactions: block.Transactions,
		PreviousHash: block.PreviousHash,
		StateRoot:    block.StateRoot,
		Proposer:     block.Proposer,
		GasUsed:      block.GasUsed,
		GasLimit:     block.GasLimit,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal block for verification: %w", err)
	}

	// 从From地址反推公钥
	proposer := block.Proposer
	if len(proposer) < 2 || proposer[:2] != "0x" {
		return fmt.Errorf("invalid address format")
	}

	// 遍历所有连接的节点，查找匹配的地址
	found := false
	var pubKey crypto.PubKey
	// 发送交易的地址可能是自己的，也可能是对等节点的其他人的
	if addr, err := PublicKeyToAddress(n.publicKey); addr == proposer {
		if err != nil {
			return fmt.Errorf("invalid address format")
		}
		pubKey = n.publicKey
		found = true
	} else {
		for _, peerID := range n.host.Network().Peers() {
			if pk := n.host.Peerstore().PubKey(peerID); pk != nil {
				addr, err = PublicKeyToAddress(pk)
				if err != nil {
					continue
				}
				if addr == proposer {
					pubKey = pk
					found = true
					break
				}
			}
		}
	}

	if !found {
		return fmt.Errorf("could not find public key for address: %s", proposer)
	}

	// 验证签名
	valid, err := pubKey.Verify(blockData, block.Signature)
	if err != nil {
		return fmt.Errorf("signature verification error: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid block signature")
	}

	return nil
}

// 计算区块哈希
func CalculateBlockHash(block *Block) (string, error) {
	// 创建用于哈希计算的区块数据结构
	blockData := struct {
		Height       uint64
		PreviousHash string
		Timestamp    time.Time
		Transactions []Transaction
		StateRoot    string
		Proposer     string
		GasUsed      uint64
		GasLimit     uint64
		Signature    []byte
	}{
		Height:       block.Height,
		PreviousHash: block.PreviousHash,
		Timestamp:    block.Timestamp,
		Transactions: block.Transactions,
		StateRoot:    block.StateRoot,
		Proposer:     block.Proposer,
		GasUsed:      block.GasUsed,
		GasLimit:     block.GasLimit,
		Signature:    block.Signature,
	}

	// 序列化区块数据
	data, err := json.Marshal(blockData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal block: %w", err)
	}

	// 计算哈希
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
