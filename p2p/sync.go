package p2p

import (
	. "FPoS/types"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
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

		var tx Transaction
		if err := json.Unmarshal(msg.Data, &tx); err == nil {
			if n.validateTransaction(tx) {
				n.txPool.Store(tx.Hash, tx)
				fmt.Printf("已收到一条消息！\ntxHash：%s\ntxFrom:%s\ntxTo:%s\ntxNonce:%d\ntxSig:%x\n\n", tx.Hash, tx.From, tx.To, tx.Nonce, tx.Signature)
				n.stateDB.ExecuteTransaction(&tx)
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
		if err := json.Unmarshal(msg.Data, &block); err == nil {
			if n.validateBlock(block) {
				n.processNewBlock(block)
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
		n.updateLocalState(resp.Accounts)
		fmt.Printf("----------------------------------------Successfully synced state from peers----------------------------------------\n\n")
		return nil
	case <-timeout:
		return fmt.Errorf("state sync timed out")
	}
}

// 更新本地状态
func (n *Layer2Node) updateLocalState(accounts map[string]*AccountState) {
	n.stateDB.mu.Lock()
	defer n.stateDB.mu.Unlock()

	// 合并状态，保留较大的nonce值
	for addr, newState := range accounts {
		if existingState, exists := n.stateDB.accounts[addr]; exists {
			if newState.Nonce > existingState.Nonce {
				existingState.Nonce = newState.Nonce
			}
			if newState.Balance > existingState.Balance {
				existingState.Balance = newState.Balance
			}
		} else {
			n.stateDB.accounts[addr] = &AccountState{
				Balance: newState.Balance,
				Nonce:   newState.Nonce,
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
		n.stateDB.mu.RUnlock()

		response := &StateSync{
			RequestID: syncReq.RequestID,
			Accounts:  accounts,
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
