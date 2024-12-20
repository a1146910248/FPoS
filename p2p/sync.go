package p2p

import (
	"FPoS/core/consensus"
	. "FPoS/types"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const MaxMessageSize = 3 * 1024 * 1024 // 0.5MB限制

// 添加消息类型标识
type MessageType int

const (
	SyncReq MessageType = iota
	SyncResponse
	SyncChunk
	TxSyncRequest
	TxSyncResponse
)

type StateChunk struct {
	Type        MessageType `json:"type"` // 添加消息类型
	RequestID   string      `json:"request_id"`
	ChunkIndex  int         `json:"chunk_index"`
	TotalChunks int         `json:"total_chunks"`
	Data        []byte      `json:"data"`
	IsFinal     bool        `json:"is_final"`
}

func (n *Layer2Node) setupTopics() error {
	txTopic, err := n.pubsub.Join("l2_transactions")
	if err != nil {
		return err
	}
	n.topic.txTopic = txTopic

	blockTopic, err := n.pubsub.Join("l2_blocks")
	if err != nil {
		return err
	}
	n.topic.blockTopic = blockTopic

	stateTopic, err := n.pubsub.Join("l2_state")
	if err != nil {
		return err
	}
	n.topic.stateTopic = stateTopic

	// 添加验证者同步专用 topic
	validatorTopic, err := n.pubsub.Join("l2_validator_sync") // 修改 topic 名称
	if err != nil {
		return fmt.Errorf("failed to join validator topic: %v", err)
	}
	n.topic.validatorTopic = validatorTopic

	// 添加交易同步专用 topic
	txSyncTopic, err := n.pubsub.Join("l2_tx_sync")
	if err != nil {
		return err
	}
	n.topic.txSyncTopic = txSyncTopic

	// 添加交易状态 topic
	txStatTopic, err := n.pubsub.Join("l2_tx_stat")
	if err != nil {
		return err
	}
	n.topic.txStatTopic = txStatTopic

	go n.handleTxMessages()
	go n.handleBlockMessages()
	go n.handleStateMessages()
	go n.handleTxSyncMessages()
	go n.handleValidatorSyncMessage()
	go n.handleTxStatMessage()

	return nil
}

type pendingMessage struct {
	msg       *pubsub.Message
	timestamp time.Time
}

func (n *Layer2Node) handleTxMessages() {
	sub, err := n.topic.txTopic.Subscribe()
	if err != nil {
		return
	}

	// 创建一个带缓冲的通道用于消息队列
	pendingMsgs := make(chan *pendingMessage, 10000)

	// 启动消费者协程
	go func() {
		for pending := range pendingMsgs {
			// 等待初始化和同步完成
			for {
				n.mu.RLock()
				isSyncing := n.isSyncing
				n.mu.RUnlock()

				if !isSyncing {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			if time.Since(pending.timestamp) > 5*time.Minute {
				fmt.Printf("Pending transaction expired, skipping\n")
				continue
			}

			var tx Transaction
			if err := json.Unmarshal(pending.msg.Data, &tx); err == nil {
				if n.validateTransaction(tx) {
					n.txPool.Store(tx.Hash, tx)
					n.IncrementTxCount()
					// 更新活跃用户统计
					stats := GetStats()
					stats.UpdateActiveUser(tx.From)
					stats.UpdateActiveUser(tx.To)
					fmt.Printf("Processed pending transaction: from=%s, nonce=%d\n",
						tx.From, tx.Nonce)
				}
			}
		}
	}()

	// 生产者循环
	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			continue
		}

		n.mu.RLock()
		//initialized := n.initialized
		isSyncing := n.isSyncing
		n.mu.RUnlock()

		// 如果已经初始化，直接丢弃消息
		//if initialized {
		//	continue
		//}

		// 如果正在同步，或者消息队列不为空，将消息放入队列
		if isSyncing || len(pendingMsgs) > 0 {
			select {
			case pendingMsgs <- &pendingMessage{
				msg:       msg,
				timestamp: time.Now(),
			}:
				fmt.Printf("Transaction queued for processing after sync\n")
			default:
				fmt.Printf("Pending message queue full, dropping transaction\n")
			}
			continue
		}

		// 如果不在同步且队列为空，直接处理消息
		var tx Transaction
		if err := json.Unmarshal(msg.Data, &tx); err == nil {
			if n.validateTransaction(tx) {
				n.txPool.Store(tx.Hash, tx)
				n.IncrementTxCount()
				// 更新活跃用户统计
				stats := GetStats()
				stats.UpdateActiveUser(tx.From)
				stats.UpdateActiveUser(tx.To)
				fmt.Printf("Processed transaction directly: from=%s, nonce=%d\n",
					tx.From, tx.Nonce)
				//n.BroadcastTransaction(tx)
			}
		}
	}
}

func (n *Layer2Node) handleBlockMessages() {
	sub, err := n.topic.blockTopic.Subscribe()
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
			if n.validateBlock(block, false) {
				isHistoricalBlock := block.Height <= n.latestBlock
				err = n.processNewBlock(block, isHistoricalBlock)
				if err != nil {
					fmt.Printf("process block fail")
					return
				}
				fmt.Printf("接收到新区块！\nblockHash：%s\nPrevBlockHash:%s\nBlockHeight:%d\nSig:%x\nProposor:%s\n\n", block.Hash, block.PreviousHash, block.Height, block.Signature, block.Proposer)
			}
		} else {
			fmt.Printf("process block Unmashal fail!")
		}
	}
}

func (n *Layer2Node) handleStateMessages() {
	sub, err := n.topic.stateTopic.Subscribe()
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

// 处理交易同步消息
func (n *Layer2Node) handleTxSyncMessages() {
	sub, err := n.topic.txSyncTopic.Subscribe()
	if err != nil {
		return
	}

	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			fmt.Printf("Error receiving message: %v\n", err)
			continue
		}

		// 跳过自己发送的消息
		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		var msgType struct {
			Type MessageType `json:"type"`
		}
		if err := json.Unmarshal(msg.Data, &msgType); err != nil {
			continue
		}

		switch msgType.Type {
		case TxSyncRequest:
			var req TxSyncReq
			if err := json.Unmarshal(msg.Data, &req); err != nil {
				continue
			}
			// 检查是否持有请求的交易
			n.handleTxSyncRequest(msg)

		case TxSyncResponse:
			var resp TxSyncRsp
			if err := json.Unmarshal(msg.Data, &resp); err != nil {
				continue
			}
			// 检查是否是发给自己的响应
			if resp.RequestID == n.currentSyncRequestID {
				n.handleTxSyncResponse(msg)
			}
		}
	}
}

func (n *Layer2Node) handleValidatorSyncMessage() {
	sub, err := n.topic.validatorTopic.Subscribe()
	if err != nil {
		return
	}

	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			fmt.Printf("Error receiving message: %v\n", err)
			continue
		}

		// 忽略自己发送的消息
		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		// 添加日志以调试消息接收
		fmt.Printf("Received validator message from: %s\n", msg.ReceivedFrom)

		if err = n.electionMgr.HandleValidatorMessage(msg.Data); err != nil {
			fmt.Printf("Failed to handle validator message: %v\n", err)
		}
	}
}

func (n *Layer2Node) handleTxStatMessage() {
	sub, err := n.topic.txStatTopic.Subscribe()
	if err != nil {
		return
	}

	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			fmt.Printf("Error receiving message: %v\n", err)
			continue
		}

		// 忽略自己发送的消息
		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		// 添加日志以调试消息接收
		fmt.Printf("Received validator message from: %s\n", msg.ReceivedFrom)

		var txs []Transaction
		if err := json.Unmarshal(msg.Data, &txs); err != nil {
			fmt.Printf("Failed to handle validator message: %v\n", err)
			continue
		}
		for _, tx := range txs {
			// 更新交易历史
			n.txHistory.Store(tx.Hash, tx)
		}
	}
}

// 从其他节点同步状态
func (n *Layer2Node) syncStateFromPeers() error {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("Retry attempt %d/%d for state sync...\n", attempt+1, maxRetries)
			time.Sleep(retryDelay)
		}

		if err := n.attemptStateSync(); err != nil {
			lastErr = err
			fmt.Printf("State sync attempt %d failed: %v\n", attempt+1, err)
			continue
		}

		// 如果成功，直接返回
		return nil
	}
	os.Exit(1)
	return fmt.Errorf("state sync failed after %d attempts, last error: %v", maxRetries, lastErr)
}

// 单次同步尝试
func (n *Layer2Node) attemptStateSync() error {
	// 首先等待一段时间，确保有足够的节点连接
	time.Sleep(1 * time.Second)

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
	if err = n.topic.stateTopic.Publish(n.ctx, reqData); err != nil {
		return fmt.Errorf("failed to publish sync request: %w", err)
	}

	// 等待响应的通道
	responseChan := make(chan *StateSync)
	timeout := time.After(30 * time.Second)

	// 用于存储接收到的分片
	chunkMap := make(map[int][]byte)
	var totalChunks int = -1
	requestID := syncReq.RequestID // 保存请求ID

	// 设置一次性的消息处理器来接收响应
	sub, err := n.topic.stateTopic.Subscribe()
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

			var chunk StateChunk
			if err := json.Unmarshal(msg.Data, &chunk); err != nil {
				fmt.Printf("Error unmarshaling chunk: %s\n", err)
				continue
			}

			// 验证请求ID
			if chunk.RequestID != requestID {
				fmt.Printf("Received chunk for different request: %s (expected %s)\n",
					chunk.RequestID, requestID)
				continue
			}

			// 打印接收到的分片信息
			fmt.Printf("Received chunk %d/%d, size: %d bytes\n",
				chunk.ChunkIndex+1, chunk.TotalChunks, len(chunk.Data))

			// 存储分片
			chunkMap[chunk.ChunkIndex] = chunk.Data
			if totalChunks == -1 {
				totalChunks = chunk.TotalChunks
				fmt.Printf("Total chunks to receive: %d\n", totalChunks)
			}

			// 检查是否收到所有分片
			if len(chunkMap) == totalChunks {
				fmt.Printf("Received all %d chunks, reconstructing data...\n", totalChunks)

				// 重建完整响应
				fullData := make([]byte, 0)
				for i := 0; i < totalChunks; i++ {
					if data, exists := chunkMap[i]; exists {
						fullData = append(fullData, data...)
					} else {
						fmt.Printf("Missing chunk %d\n", i)
						return
					}
				}

				var syncResp StateSync
				if err := json.Unmarshal(fullData, &syncResp); err != nil {
					fmt.Printf("Error unmarshaling complete response: %s\n", err)
					return
				}

				responseChan <- &syncResp
				return
			}
		}
	}()

	// 等待响应或超时
	select {
	case resp := <-responseChan:
		n.updateLocalState(resp.Accounts, resp.PendingTxs, resp.Blocks, resp.ToHeight, resp.Validators, resp.SelectState)
		fmt.Printf("---------------------------------------------Successfully synced state from peers---------------------------------------------\n")
		return nil
	case <-timeout:
		return fmt.Errorf("state sync timed out")
	}
}

func (n *Layer2Node) updateLocalState(accounts map[string]*AccountState, pendingTxs []Transaction, blocks []Block,
	newHeight uint64, validators map[string]consensus.Validator, selectState consensus.ElectionState) {
	n.mu.Lock()
	n.isSyncing = true
	wasInitialized := n.initialized
	n.mu.Unlock()

	defer func() {
		n.mu.Lock()
		n.isSyncing = false
		if !wasInitialized {
			n.initialized = true
			fmt.Println("Node initialization completed")
		}
		n.mu.Unlock()
	}()

	n.stateDB.Lock()
	// 更新账户状态
	for addr, newState := range accounts {
		if existingState, exists := n.stateDB.accounts[addr]; exists {
			existingState.Balance = newState.Balance
		} else {
			n.stateDB.accounts[addr] = &AccountState{
				Balance:       newState.Balance,
				Nonce:         0,
				PublicKey:     newState.PublicKey,
				PublicKeyType: newState.PublicKeyType,
			}
		}
	}
	n.stateDB.Unlock()

	// 清空现有的交易池和待处理状态
	n.txPool = &sync.Map{}
	n.stateDB.Lock()
	n.stateDB.pendingTxs = make(map[string]*PendingState)

	// 按发送方地址和nonce排序交易
	sort.Slice(pendingTxs, func(i, j int) bool {
		if pendingTxs[i].From == pendingTxs[j].From {
			return pendingTxs[i].Nonce < pendingTxs[j].Nonce
		}
		return pendingTxs[i].From < pendingTxs[j].From
	})
	n.stateDB.Unlock()
	// 重新验证并添加交易，同时更新待处理状态
	for _, tx := range pendingTxs {
		if err := n.stateDB.ValidateTransaction(&tx, n.minGasPrice); err == nil {
			txAccount := n.stateDB.GetAccount(tx.From)
			n.stateDB.Lock()
			// 交易有效，添加到交易池
			n.txPool.Store(tx.Hash, tx)

			// 更新或创建待处理状态
			pending, exists := n.stateDB.pendingTxs[tx.From]
			if !exists {
				account := txAccount
				pending = &PendingState{
					pendingBalance: account.Balance,
					pendingNonce:   account.Nonce,
					mu:             sync.RWMutex{},
				}
				n.stateDB.pendingTxs[tx.From] = pending
			}

			n.stateDB.Unlock()
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

	n.mu.Lock()
	n.latestBlock = newHeight
	// 如果是 sequencer 节点，同时更新 sequencer 的区块高度
	if n.isSequencer && n.sequencer != nil {
		n.sequencer.mu.Lock()
		n.sequencer.blockHeight = newHeight
		n.sequencer.mu.Unlock()
	}

	// 签名的交易和账户重构完成后再同步区块，否则会出现nonce等不一致区块
	for _, block := range blocks {
		// 验证区块
		if !n.validateBlockInternal(block, true) {
			fmt.Printf("Invalid block during sync: height=%d, hash=%s\n",
				block.Height, block.Hash)
			continue
		}

		// 处理区块
		if err := n.processNewBlockInternal(block, true); err != nil {
			fmt.Printf("Failed to process block during sync: %v\n", err)
			continue
		}
	}

	// 同步选举器的state
	n.electionMgr.SetState(&selectState)
	// 同步现有的validators
	n.electionMgr.InitValidators()
	for addr, validator := range validators {
		if _, ok := n.electionMgr.Validators[addr]; ok {
			continue
		}
		n.electionMgr.Validators[addr] = &validator
	}
	//// 如果没有别的候选者应该在初始化后令唯一的候选者为排序器
	//if len(n.electionMgr.Validators) == 1 {
	//	n.electionMgr.RotateSequencer()
	//}
	n.mu.Unlock()
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
	if pb, _ := PublicKeyToAddress(n.publicKey); pb != "0x8f00527c4f08eb89f9158f9fe14545b868e0498d" {
		return
	}
	// 首先尝试解析为StateChunk
	var chunk StateChunk
	if err := json.Unmarshal(msg.Data, &chunk); err == nil && chunk.Type == SyncChunk {
		// 这是一个分片消息，直接返回
		return
	}

	// 如果不是分片消息，则尝试解析为StateSync
	var syncReq StateSync
	if err := json.Unmarshal(msg.Data, &syncReq); err != nil {
		fmt.Printf("Error unmarshaling message: %s\n", err)
		return
	}

	// 只处理同步请求
	if syncReq.Type != SyncReq {
		return
	}

	// 准备响应数据
	n.stateDB.mu.RLock()
	accounts := make(map[string]*AccountState)
	for addr, state := range n.stateDB.accounts {
		accounts[addr] = &AccountState{
			Balance:       state.Balance,
			Nonce:         state.Nonce,
			PublicKey:     state.PublicKey,
			PublicKeyType: state.PublicKeyType,
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

	// 收集区块信息
	blocks := make([]Block, 0)
	for height := syncReq.FromHeight; height <= n.latestBlock; height++ {
		if blockInterface, exists := n.blockCache.Load(height); exists {
			if block, ok := blockInterface.(Block); ok {
				blocks = append(blocks, block)
			}
		}
	}

	response := &StateSync{
		Type:         SyncResponse,
		RequestID:    syncReq.RequestID,
		Accounts:     accounts,
		PendingState: pendingStates,
		PendingTxs:   pendingTxs,
		Blocks:       blocks,
		FromHeight:   syncReq.FromHeight,
		ToHeight:     n.latestBlock,
		Validators:   n.electionMgr.GetValidators(),
		SelectState:  n.electionMgr.GetState(),
	}

	// 序列化完整响应
	fullData, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("Error marshaling sync response: %s\n", err)
		return
	}

	// 计算分片
	totalChunks := (len(fullData) + MaxMessageSize - 1) / MaxMessageSize
	fmt.Printf("Splitting response into %d chunks\n", totalChunks)

	// 分片发送
	for i := 0; i < totalChunks; i++ {
		start := i * MaxMessageSize
		end := start + MaxMessageSize
		if end > len(fullData) {
			end = len(fullData)
		}

		chunk := StateChunk{
			Type:        SyncChunk,
			RequestID:   syncReq.RequestID,
			ChunkIndex:  i,
			TotalChunks: totalChunks,
			Data:        fullData[start:end],
			IsFinal:     i == totalChunks-1,
		}

		chunkData, err := json.Marshal(chunk)
		if err != nil {
			fmt.Printf("Error marshaling chunk %d: %s\n", i, err)
			continue
		}

		if err := n.topic.stateTopic.Publish(n.ctx, chunkData); err != nil {
			fmt.Printf("Error publishing chunk %d/%d: %s\n", i+1, totalChunks, err)
			continue
		}

		fmt.Printf("Sent chunk %d/%d, size: %d bytes\n", i+1, totalChunks, len(chunk.Data))
		//time.Sleep(200 * time.Millisecond)
	}
}

// 广播验证者加入
func (n *Layer2Node) BroadcastValidatorJoin(validator consensus.Validator, msgType consensus.ValidatorMessageType) error {
	// 确保在广播前设置 PublicKeyBytes
	if validator.PublicKey != nil {
		pubKeyBytes, err := crypto.MarshalPublicKey(validator.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to get public key bytes: %v", err)
		}
		validator.PublicKeyBytes = pubKeyBytes
	}
	msg := consensus.ValidatorMessage{
		Type:      msgType,
		Validator: validator,
		Signature: nil,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	err = n.topic.validatorTopic.Publish(n.ctx, data)
	return err
}

// 广播交易状态
func (n *Layer2Node) broadcastTxStat(txs []Transaction) error {
	data, err := json.Marshal(txs)
	if err != nil {
		return err
	}
	err = n.topic.txStatTopic.Publish(n.ctx, data)
	return err
}
