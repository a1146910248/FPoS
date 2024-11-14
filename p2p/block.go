package p2p

import (
	. "FPoS/types"
	"encoding/json"
	"fmt"
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
