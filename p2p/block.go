package p2p

import (
	. "FPoS/types"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// 处理新区块
func (n *Layer2Node) processNewBlock(block Block, isHistoricalBlock bool) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	fmt.Printf("Processing block %d, current account nonces:\n", block.Height)
	for addr := range n.stateDB.accounts {
		fmt.Printf("Account %s nonce: %d\n", addr, n.stateDB.GetNonce(addr))
	}

	// 检查区块高度
	if !isHistoricalBlock && block.Height <= n.latestBlock {
		return fmt.Errorf("block height %d is not higher than current height %d",
			block.Height, n.latestBlock)
	}

	// 使用新的方法原子性地处理交易池和待处理状态
	n.cleanTxPoolAndPendingStates(block.Transactions)

	// 更新状态
	n.latestBlock = block.Height
	n.stateRoot = block.StateRoot
	n.blockCache.Store(block.Height, block)

	if n.sequencer != nil {
		n.sequencer.blockHeight++
	}

	// 通知选举管理器新区块生成
	if n.electionMgr != nil {
		n.electionMgr.OnBlockProduced(block.Height)
	}

	// 更新统计信息
	stats := GetStats()
	stats.UpdateBlockHeight(block.Height)
	stats.UpdateTxCount(n.GetTotalTxCount()) // 需要实现此方法
	return nil
}

// 广播交易
func (n *Layer2Node) BroadcastTransaction(tx Transaction) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	return n.topic.txTopic.Publish(n.ctx, data)
}

// 广播区块
func (n *Layer2Node) BroadcastBlock(block Block) error {
	data, err := json.Marshal(block)
	if err != nil {
		return err
	}

	return n.topic.blockTopic.Publish(n.ctx, data)
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

	// 首先检查是否是本节点的地址
	if addr, err := PublicKeyToAddress(n.publicKey); err == nil && addr == proposer {
		valid, err := n.publicKey.Verify(blockData, block.Signature)
		if err != nil {
			return fmt.Errorf("signature verification error: %w", err)
		}
		if !valid {
			return fmt.Errorf("invalid block signature")
		}
		return nil
	}

	// 从账户状态中获取公钥
	pubKey, err := n.stateDB.GetAccountPublicKey(proposer)
	if err != nil {
		return fmt.Errorf("failed to get proposer public key: %w", err)
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
func (n *Layer2Node) cleanTxPoolAndPendingStates(txs []Transaction) {
	n.stateDB.Lock()

	// 创建交易 map 用于快速查找
	txHashes := make(map[string]struct{})
	txsByAddress := make(map[string][]Transaction)

	for _, tx := range txs {
		txHashes[tx.Hash] = struct{}{}
		txsByAddress[tx.From] = append(txsByAddress[tx.From], tx)
	}

	// 清理交易池
	var txsToRemove []interface{}
	n.txPool.Range(func(key, value interface{}) bool {
		if hash, ok := key.(string); ok {
			if _, exists := txHashes[hash]; exists {
				txsToRemove = append(txsToRemove, key)
			}
		}
		return true
	})

	// 删除已确认的交易
	for _, key := range txsToRemove {
		n.txPool.Delete(key)
	}
	n.stateDB.Unlock()

	// 应用交易
	n.applyTransactions(txs)
	// 更新每个地址的待处理状态
	for address := range txsByAddress {
		// 重置待处理状态到最新确认的 nonce
		n.stateDB.ResetPendingNonce(address)

		// 重新应用剩余的待处理交易
		n.txPool.Range(func(_, value interface{}) bool {
			if tx, ok := value.(Transaction); ok {
				if tx.From == address {
					if pending, exists := n.stateDB.pendingTxs[address]; exists {
						pending.mu.Lock()
						pending.pendingNonce++
						pending.mu.Unlock()
					}
				}
			}
			return true
		})
	}
}

// 处理新区块
func (n *Layer2Node) processNewBlockInternal(block Block, isHistoricalBlock bool) error {
	fmt.Printf("Processing block %d, current account nonces:\n", block.Height)
	for addr := range n.stateDB.accounts {
		fmt.Printf("Account %s nonce: %d\n", addr, n.stateDB.GetNonce(addr))
	}

	// 检查区块高度
	if !isHistoricalBlock && block.Height <= n.latestBlock {
		return fmt.Errorf("block height %d is not higher than current height %d",
			block.Height, n.latestBlock)
	}

	// 使用新的方法原子性地处理交易池和待处理状态
	n.cleanTxPoolAndPendingStates(block.Transactions)

	// 更新状态
	n.latestBlock = block.Height
	n.stateRoot = block.StateRoot
	n.blockCache.Store(block.Height, block)

	return nil
}
