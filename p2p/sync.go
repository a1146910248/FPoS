package p2p

import (
	. "FPoS/types"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"sort"
	"sync"
	"time"
)

func (n *Layer2Node) setupTopics() error {
	txTopic, err := n.pubsub.Join("l2_transactions")
	if err != nil {
		return err
	}
	n.txTopic = txTopic

	blockTopic, err := n.pubsub.Join("l2_blocks")
	if err != nil {
		return err
	}
	n.blockTopic = blockTopic

	stateTopic, err := n.pubsub.Join("l2_state")
	if err != nil {
		return err
	}
	n.stateTopic = stateTopic

	go n.handleTxMessages()
	go n.handleBlockMessages()
	go n.handleStateMessages()

	return nil
}

func (n *Layer2Node) handleTxMessages() {
	sub, err := n.txTopic.Subscribe()
	if err != nil {
		return
	}

	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			continue
		}
		myAddr, _ := PublicKeyToAddress(n.publicKey)
		var tx Transaction
		if err := json.Unmarshal(msg.Data, &tx); err == nil {
			if n.validateTransaction(tx) {
				n.txPool.Store(tx.Hash, tx)
				fmt.Printf("已收到一条消息！\ntxHash：%s\ntxFrom:%s\ntxTo:%s\ntxNonce:%d\ntxSig:%x\n余额:%d\n", tx.Hash, tx.From, tx.To, tx.Nonce, tx.Signature, n.stateDB.GetBalance(myAddr))
				// 打包成区块再执行
				//n.stateDB.ExecuteTransaction(&tx)
				n.BroadcastTransaction(tx)
			}
		}
	}
}

func (n *Layer2Node) handleBlockMessages() {
	sub, err := n.blockTopic.Subscribe()
	if err != nil {
		return
	}

	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			continue
		}

		var block Block
		if err = json.Unmarshal(msg.Data, &block); err == nil {
			if n.validateBlock(block) {
				err = n.processNewBlock(block)
				if err != nil {
					fmt.Printf("process block fail")
					return
				}
				fmt.Printf("接收到新区块！\nblockHash：%s\nPrevBlockHash:%s\nBlockHeight:%d\nSig:%x\nProposor:%s\n\n", block.Hash, block.PreviousHash, block.Height, block.Signature, block.Proposer)
			}
		}
	}
}

func (n *Layer2Node) handleStateMessages() {
	sub, err := n.stateTopic.Subscribe()
	if err != nil {
		return
	}

	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			fmt.Printf("Error receiving state message: %s\n", err)
			continue
		}

		// 忽略自己发送的消息
		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		n.handleStateSync(msg)
	}
	//for {
	//	msg, err := sub.Next(n.ctx)
	//	if err != nil {
	//		continue
	//	}
	//
	//	// 尝试解析为节点信息
	//	var peerInfo peer.AddrInfo
	//	if err := json.Unmarshal(msg.Data, &peerInfo); err == nil {
	//		// 收到新节点信息，尝试连接
	//		if peerInfo.ID != n.host.ID() && n.host.Network().Connectedness(peerInfo.ID) != network.Connected {
	//			if err := n.host.Connect(n.ctx, peerInfo); err == nil {
	//				fmt.Printf("Connected to broadcasted peer: %s\n", peerInfo.ID)
	//			}
	//		}
	//		continue
	//	}
	//
	//	// 如果不是节点信息，则尝试解析为状态更新
	//	var state struct {
	//		StateRoot string `json:"stateRoot"`
	//		Height    uint64 `json:"height"`
	//	}
	//	if err := json.Unmarshal(msg.Data, &state); err == nil {
	//		n.updateState(state.StateRoot, state.Height)
	//	}
	//}
}

// 从其他节点同步状态
func (n *Layer2Node) syncStateFromPeers() error {
	// 首先等待一段时间，确保有足够的节点连接
	time.Sleep(2 * time.Second)
	// 创建状态同步请求
	syncReq := &StateSync{
		RequestID: uuid.New().String(),
	}

	// 序列化请求
	reqData, err := json.Marshal(syncReq)
	if err != nil {
		return fmt.Errorf("failed to marshal sync request: %w", err)
	}

	// 通过状态主题发布同步请求
	if err = n.stateTopic.Publish(n.ctx, reqData); err != nil {
		return fmt.Errorf("failed to publish sync request: %w", err)
	}

	// 等待响应的通道
	responseChan := make(chan *StateSync)
	timeout := time.After(10 * time.Second)

	// 设置一次性的消息处理器来接收响应
	sub, err := n.stateTopic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to state topic: %w", err)
	}
	defer sub.Cancel()

	// 在goroutine中处理响应
	go func() {
		for {
			msg, err := sub.Next(n.ctx)
			if err != nil {
				fmt.Printf("Error receiving sync response: %s\n", err)
				return
			}

			// 忽略自己发送的消息
			if msg.ReceivedFrom == n.host.ID() {
				continue
			}

			var syncResp StateSync
			if err := json.Unmarshal(msg.Data, &syncResp); err != nil {
				fmt.Printf("Error unmarshaling sync response: %s\n", err)
				continue
			}

			// 如果是对我们请求的响应
			if syncResp.Accounts != nil {
				responseChan <- &syncResp
				return
			}
		}
	}()

	// 等待响应或超时
	select {
	case resp := <-responseChan:
		// 更新本地状态
		n.updateLocalState(resp.Accounts, resp.PendingTxs)
		fmt.Printf("----------------------------------------Successfully synced state from peers----------------------------------------\n\n")
		return nil
	case <-timeout:
		return fmt.Errorf("state sync timed out")
	}
}

func (n *Layer2Node) updateLocalState(accounts map[string]*AccountState, pendingTxs []Transaction) {

	n.stateDB.mu.Lock()

	// 更新账户状态
	for addr, newState := range accounts {
		if existingState, exists := n.stateDB.accounts[addr]; exists {
			if newState.Nonce > existingState.Nonce {
				existingState.Nonce = newState.Nonce
			}
			existingState.Balance = newState.Balance
		} else {
			n.stateDB.accounts[addr] = &AccountState{
				Balance: newState.Balance,
				Nonce:   newState.Nonce,
			}
		}
	}
	n.stateDB.mu.Unlock()

	// 清空现有的交易池和待处理状态
	n.txPool = &sync.Map{}
	n.stateDB.pendingTxs = make(map[string]*PendingState)

	// 按发送方地址和nonce排序交易
	sort.Slice(pendingTxs, func(i, j int) bool {
		if pendingTxs[i].From == pendingTxs[j].From {
			return pendingTxs[i].Nonce < pendingTxs[j].Nonce
		}
		return pendingTxs[i].From < pendingTxs[j].From
	})

	// 重新验证并添加交易，同时更新待处理状态
	for _, tx := range pendingTxs {
		if err := n.stateDB.ValidateTransaction(&tx, n.minGasPrice); err == nil {
			n.stateDB.mu.Lock()
			// 交易有效，添加到交易池
			n.txPool.Store(tx.Hash, tx)

			// 更新或创建待处理状态
			pending, exists := n.stateDB.pendingTxs[tx.From]
			if !exists {
				account := n.stateDB.GetAccount(tx.From)
				pending = &PendingState{
					pendingBalance: account.Balance,
					pendingNonce:   account.Nonce,
					mu:             sync.RWMutex{},
				}
				n.stateDB.pendingTxs[tx.From] = pending
			}

			pending.mu.Lock()
			gasFeeCost := tx.GasUsed * tx.GasPrice
			totalCost := gasFeeCost + tx.Value
			pending.pendingBalance -= totalCost
			pending.pendingNonce++
			pending.mu.Unlock()
			n.stateDB.mu.Unlock()
		} else {
			fmt.Printf("Invalid transaction during sync: %s, error: %v\n", tx.Hash, err)
		}
	}

	// 验证一致性
	var txCount int
	n.txPool.Range(func(_, _ interface{}) bool {
		txCount++
		return true
	})

	pendingCount := 0
	for addr, pending := range n.stateDB.pendingTxs {
		pending.mu.RLock()
		pendingDiff := pending.pendingNonce - n.stateDB.GetAccount(addr).Nonce
		pending.mu.RUnlock()
		pendingCount += int(pendingDiff)
	}

	if txCount != pendingCount {
		fmt.Printf("Warning: Inconsistency detected - TxPool count: %d, Pending count: %d\n",
			txCount, pendingCount)
		// 打印详细信息以调试
		fmt.Printf("Pending states by address:\n")
		for addr, pending := range n.stateDB.pendingTxs {
			pending.mu.RLock()
			fmt.Printf("Address: %s, Current Nonce: %d, Pending Nonce: %d, Diff: %d\n",
				addr,
				n.stateDB.GetAccount(addr).Nonce,
				pending.pendingNonce,
				pending.pendingNonce-n.stateDB.GetAccount(addr).Nonce)
			pending.mu.RUnlock()
		}

		fmt.Printf("State updated: accounts=%d, pendingStates=%d, pendingTxs=%d\n",
			len(accounts), len(n.stateDB.pendingTxs), txCount)
	}
}

// 重新验证交易池中的所有交易
func (n *Layer2Node) revalidateTransactionPool() {
	invalidTxs := make([]string, 0)

	// 收集所有需要移除的交易
	n.txPool.Range(func(key, value interface{}) bool {
		if tx, ok := value.(Transaction); ok {
			// 使用新的状态验证交易
			if err := n.stateDB.ValidateTransaction(&tx, n.minGasPrice); err != nil {
				invalidTxs = append(invalidTxs, tx.Hash)
			}
		}
		return true
	})

	// 移除无效交易
	for _, hash := range invalidTxs {
		if txInterface, exists := n.txPool.Load(hash); exists {
			if tx, ok := txInterface.(Transaction); ok {
				n.removeFromTxPool(&tx)
			}
		}
	}
}

// 处理状态同步请求
func (n *Layer2Node) handleStateSync(msg *pubsub.Message) {
	var syncReq StateSync
	if err := json.Unmarshal(msg.Data, &syncReq); err != nil {
		fmt.Printf("Error unmarshaling sync request: %s\n", err)
		return
	}

	// 如果是同步请求（没有accounts字段）
	if syncReq.Accounts == nil {
		// 准备响应
		n.stateDB.mu.RLock()
		accounts := make(map[string]*AccountState)
		for addr, state := range n.stateDB.accounts {
			accounts[addr] = &AccountState{
				Balance: state.Balance,
				Nonce:   state.Nonce,
			}
		}
		// 收集交易池中的交易
		var pendingTxs []Transaction
		n.txPool.Range(func(_, value interface{}) bool {
			if tx, ok := value.(Transaction); ok {
				pendingTxs = append(pendingTxs, tx)
			}
			return true
		})
		n.stateDB.mu.RUnlock()

		// 添加待处理状态信息
		pendingStates := make(map[string]*PendingState)
		for addr, pending := range n.stateDB.pendingTxs {
			pending.mu.RLock()
			pendingStates[addr] = &PendingState{
				pendingBalance: pending.pendingBalance,
				pendingNonce:   pending.pendingNonce,
			}
			pending.mu.RUnlock()
		}

		response := &StateSync{
			RequestID:    syncReq.RequestID,
			Accounts:     accounts,
			PendingState: pendingStates, // 添加待处理状态
			PendingTxs:   pendingTxs,
		}

		// 序列化并发送响应
		respData, err := json.Marshal(response)
		if err != nil {
			fmt.Printf("Error marshaling sync response: %s\n", err)
			return
		}

		if err := n.stateTopic.Publish(n.ctx, respData); err != nil {
			fmt.Printf("Error publishing sync response: %s\n", err)
			return
		}
	}
}
