package p2p

import (
	"FPoS/core/consensus"
	"FPoS/core/dac"
	. "FPoS/types"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func (n *Layer2Node) SetTransactionHandler(handler TransactionHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.handlers.TxHandler = handler
}

func (n *Layer2Node) SetBlockHandler(handler BlockHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.handlers.BlockHandler = handler
}

func (n *Layer2Node) validateTransaction(tx Transaction) bool {
	n.mu.RLock()
	handler := n.handlers.TxHandler
	n.mu.RUnlock()

	if handler == nil {
		return n.defaultTxValidation(&tx)
	}
	return handler(tx)
}

func (n *Layer2Node) validateBlock(block Block, isHistoricalBlock bool) bool {
	n.mu.RLock()
	handler := n.handlers.BlockHandler
	n.mu.RUnlock()

	if handler == nil {
		return n.defaultBlockValidation(block, isHistoricalBlock)
	}
	return handler(block)
}

func (n *Layer2Node) defaultTxValidation(tx *Transaction) bool {
	if isRight, err := CalculateTxHash(tx); !isRight || err != nil {
		fmt.Println("交易哈希错误")
		return false
	}

	if _, exists := n.txPool.Load(tx.Hash); exists {
		return false
	}

	if tx.Timestamp.IsZero() {
		return false
	}

	if tx.From == "" || tx.To == "" {
		return false
	}

	if len(tx.Signature) == 0 {
		return false
	}

	// 检查nonce值
	//currentNonce := n.stateDB.GetNonce(tx.From)
	//if tx.Nonce > currentNonce+1 {
	//	fmt.Printf("Transaction nonce gap detected: current=%d, received=%d\n",
	//		currentNonce, tx.Nonce)
	//
	//	// 触发交易同步
	//	go func() {
	//		n.mu.Lock()
	//		if n.isSyncing {
	//			n.mu.Unlock()
	//			return
	//		}
	//		// 在接收到响应并且装载完后再解除锁定
	//		n.isSyncing = true
	//		n.mu.Unlock()
	//
	//		// 请求缺失的交易
	//		if err := n.syncMissingTransactions(tx.From, currentNonce+1, tx.Nonce); err != nil {
	//			fmt.Printf("Missing transactions sync failed: %v\n", err)
	//		}
	//
	//	}()
	//	return false
	//} else if tx.Nonce < currentNonce+1 {
	//	fmt.Printf("Transaction nonce too low: expected %d, got %d\n",
	//		currentNonce+1, tx.Nonce)
	//	return false
	//}

	// Gas和余额检查
	if err := n.stateDB.ValidateTransaction(tx, n.minGasPrice); err != nil {
		fmt.Println("交易验证不通过：", err)
		return false
	}
	if err := VerifyTransactionSignature(tx, n); err != nil {
		fmt.Printf("Transaction signature verification failed: %v\n", err)
		return false
	}
	return true
}

func (n *Layer2Node) defaultBlockValidation(block Block, isHistoricalBlock bool) bool {
	n.mu.RLock()
	currentHeight := n.latestBlock
	n.mu.RUnlock()

	// 检查block hash
	if hash, err := CalculateBlockHash(&block); hash != block.Hash || err != nil {
		fmt.Printf("Block Hash invalid: current=%s, real=%s\n",
			block.Hash, hash)
		return false
	}
	// 只有非历史区块才检查高度必须大于当前高度
	if !isHistoricalBlock && block.Height <= currentHeight {
		fmt.Printf("Block height invalid: current=%d, new=%d\n",
			currentHeight, block.Height)
		return false
	}

	// 检查区块连续性
	previousBlock, exists := n.blockCache.Load(block.Height - 1)
	if exists {
		if prev, ok := previousBlock.(Block); ok {
			if prev.Hash != block.PreviousHash {
				return false
			}
		}
	}

	// 验证交易默克尔根
	if calculateTxRoot := CalculateMerkleRoot(block.Transactions); calculateTxRoot != block.TxRoot {
		fmt.Printf("Transaction merker verification failed\n")
	}

	if block.Timestamp.IsZero() {
		return false
	}

	// 验证每笔交易，是否是同步分开
	for _, tx := range block.Transactions {
		if !n.validateTxForBlock(&tx, isHistoricalBlock, block.Proposer) {
			return false
		}
	}

	if block.Proposer == "" || len(block.Signature) == 0 {
		return false
	}
	if len(block.Votes) != 2 {
		return false
	}
	err := VerifyBlockSignature(&block, n)
	if err != nil {
		return false
	}
	return true
}

func (n *Layer2Node) BlockVoteValidation(block Block, isHistoricalBlock bool) bool {
	n.mu.RLock()
	currentHeight := n.latestBlock
	n.mu.RUnlock()

	// 检查block hash
	if hash, err := CalculateBlockHash(&block); hash != block.Hash || err != nil {
		fmt.Printf("Block Hash invalid: current=%s, real=%s\n",
			block.Hash, hash)
		return false
	}
	// 只有非历史区块才检查高度必须大于当前高度
	if !isHistoricalBlock && block.Height <= currentHeight {
		fmt.Printf("Block height invalid: current=%d, new=%d\n",
			currentHeight, block.Height)
		return false
	}

	// 检查区块连续性
	previousBlock, exists := n.blockCache.Load(block.Height - 1)
	if exists {
		if prev, ok := previousBlock.(Block); ok {
			if prev.Hash != block.PreviousHash {
				return false
			}
		}
	}

	// 验证交易默克尔根
	if calculateTxRoot := CalculateMerkleRoot(block.Transactions); calculateTxRoot != block.TxRoot {
		fmt.Printf("Transaction merker verification failed\n")
	}

	if block.Timestamp.IsZero() {
		return false
	}

	// 验证每笔交易，是否是同步分开
	for _, tx := range block.Transactions {
		if !n.validateTxForBlock(&tx, isHistoricalBlock, block.Proposer) {
			return false
		}
	}

	if block.Proposer == "" || len(block.Signature) == 0 {
		return false
	}
	err := VerifyBlockSignature(&block, n)
	if err != nil {
		return false
	}
	return true
}

// 验证区块中的每条交易，需要在交易池中存在，与同步交易刚好相反
func (n *Layer2Node) validateTxForBlock(tx *Transaction, isHistoricalBlock bool, sequencerAddr string) bool {
	if isRight, err := CalculateTxHash(tx); !isRight || err != nil {
		fmt.Println("交易哈希错误")
		return false
	}
	// 只有当不是历史区块时才检查交易池
	myAddress, _ := PublicKeyToAddress(n.publicKey)
	if !isHistoricalBlock {
		if sequencerAddr == myAddress {
			//if _, exists := n.txPool.Load(tx.Hash); exists {
			//	fmt.Println("本节点为排序器节点，但交易池中交易未正确清除")
			//	return false
			//}
		} else {
			// 等待初始化和同步完成,未完全同步会导致找不到对应交易
			for {
				n.mu.RLock()
				isSyncing := n.isSyncing
				n.mu.RUnlock()

				if !isSyncing {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			//if _, exists := n.txPool.Load(tx.Hash); !exists {
			//	fmt.Println("本节点为非排序器节点，但交易池中未含有该交易")
			//	return false
			//}
		}
	}

	if tx.Timestamp.IsZero() {
		return false
	}

	if tx.From == "" || tx.To == "" {
		return false
	}

	if len(tx.Signature) == 0 {
		return false
	}

	// 检查 nonce 值,如果大于现在的 From 的 nonce 则不合法
	//currentNonce := n.stateDB.GetNonce(tx.From)
	//if !isHistoricalBlock && tx.Nonce > currentNonce+1 {
	//	fmt.Printf("block交易nonce无效: 期望 %d, 实际 %d\n", currentNonce+1, tx.Nonce)
	//	return false
	//}

	// Gas和余额检查
	if err := n.stateDB.ValidateTransactionForBlock(tx, n.minGasPrice); err != nil {
		fmt.Println("交易验证不通过：", err)
		return false
	}
	if err := VerifyTransactionSignature(tx, n); err != nil {
		fmt.Printf("Transaction signature verification failed: %v\n", err)
		return false
	}
	return true
}

func (n *Layer2Node) validateBlockInternal(block Block, isHistoricalBlock bool) bool {
	currentHeight := n.latestBlock

	// 检查block hash
	if hash, err := CalculateBlockHash(&block); hash != block.Hash || err != nil {
		return false
	}
	// 只有非历史区块才检查高度必须大于当前高度
	if !isHistoricalBlock && block.Height <= currentHeight {
		fmt.Printf("Block height invalid: current=%d, new=%d\n",
			currentHeight, block.Height)
		return false
	}

	// 检查区块连续性
	previousBlock, exists := n.blockCache.Load(block.Height - 1)
	if exists {
		if prev, ok := previousBlock.(Block); ok {
			if prev.Hash != block.PreviousHash {
				return false
			}
		}
	}

	// 验证交易默克尔根
	if calculateTxRoot := CalculateMerkleRoot(block.Transactions); calculateTxRoot != block.TxRoot {
		fmt.Printf("Transaction merker verification failed\n")
	}

	if block.Timestamp.IsZero() {
		return false
	}

	// 验证每笔交易，是否是同步分开
	for _, tx := range block.Transactions {
		if !n.validateTxForBlock(&tx, isHistoricalBlock, block.Proposer) {
			return false
		}
	}

	if block.Proposer == "" || len(block.Signature) == 0 {
		return false
	}
	if len(block.Votes) != 2 {
		return false
	}
	err := VerifyBlockSignature(&block, n)
	if err != nil {
		return false
	}
	return true
}

// 交易同步请求结构
type TxSyncReq struct {
	Type      MessageType `json:"type"`
	RequestID string      `json:"request_id"`
	Address   string      `json:"address"`
	FromNonce uint64      `json:"from_nonce"`
	ToNonce   uint64      `json:"to_nonce"`
}

// 交易同步响应结构
type TxSyncRsp struct {
	Type         MessageType  `json:"type"`
	RequestID    string       `json:"request_id"`
	Address      string       `json:"address"`
	Transactions []TxWithMeta `json:"transactions"`
}

// 添加交易元数据
type TxWithMeta struct {
	Transaction Transaction `json:"transaction"`
	Source      TxSource    `json:"source"` // 交易来源
}

// 交易来源
type TxSource string

const (
	TxSourcePool    TxSource = "pool"    // 来自交易池
	TxSourceHistory TxSource = "history" // 来自历史记录
)

// 同步缺失的交易
func (n *Layer2Node) syncMissingTransactions(address string, fromNonce, toNonce uint64) error {
	requestID := uuid.New().String()
	n.mu.Lock()
	n.currentSyncRequestID = requestID // 记录当前请求ID
	n.mu.Unlock()

	req := TxSyncReq{
		Type:      TxSyncRequest,
		RequestID: requestID,
		Address:   address,
		FromNonce: fromNonce,
		ToNonce:   toNonce,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal tx sync request: %w", err)
	}

	// 发布同步请求
	if err := n.topic.txSyncTopic.Publish(n.ctx, data); err != nil {
		return fmt.Errorf("failed to publish tx sync request: %w", err)
	}

	fmt.Printf("Requested missing transactions for %s from nonce %d to %d\n",
		address, fromNonce, toNonce)
	return nil
}

// 处理交易同步请求
func (n *Layer2Node) handleTxSyncRequest(msg *pubsub.Message) {
	var req TxSyncReq
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return
	}

	// 用 map 收集交易，确保不重复
	txMap := make(map[uint64]TxWithMeta)
	missingNonces := make(map[uint64]bool)

	// 初始化需要的 nonce 列表
	for nonce := req.FromNonce; nonce <= req.ToNonce; nonce++ {
		missingNonces[nonce] = true
	}

	// 收集请求范围内的交易
	n.txPool.Range(func(_, value interface{}) bool {
		if tx, ok := value.(Transaction); ok {
			if tx.From == req.Address &&
				tx.Nonce >= req.FromNonce &&
				tx.Nonce <= req.ToNonce {
				txMap[tx.Nonce] = TxWithMeta{
					Transaction: tx,
					Source:      TxSourcePool,
				}
				delete(missingNonces, tx.Nonce)
			}
		}
		return true
	})
	// 从历史记录中收集缺失的交易
	if len(missingNonces) > 0 {
		// 首先收集该地址的所有历史交易hash
		addrTxs := make(map[string]struct{})
		n.txHistory.Range(func(key, value interface{}) bool {
			if tx, ok := value.(Transaction); ok {
				if tx.From == req.Address {
					addrTxs[key.(string)] = struct{}{}
				}
			}
			return true
		})

		// 再次遍历找到缺失的nonce
		n.txHistory.Range(func(key, value interface{}) bool {
			if _, exists := addrTxs[key.(string)]; exists {
				if tx, ok := value.(Transaction); ok {
					if _, missing := missingNonces[tx.Nonce]; missing {
						txMap[tx.Nonce] = TxWithMeta{
							Transaction: tx,
							Source:      TxSourceHistory,
						}
						delete(missingNonces, tx.Nonce)
					}
				}
			}
			// 如果已经找到所有缺失的nonce，可以提前结束遍历
			if len(missingNonces) == 0 {
				return false
			}
			return true
		})

		// 检查是否找到所有缺失的交易
		if len(missingNonces) > 0 {
			logger.Infof("Incomplete transaction sequence for address %s",
				req.Address)
			return
		}
	}

	// 将 map 转换为有序数组
	transactions := make([]TxWithMeta, 0, req.ToNonce-req.FromNonce+1)
	for nonce := req.FromNonce; nonce <= req.ToNonce; nonce++ {
		txMeta, exists := txMap[nonce]
		if !exists {
			logger.Errorf("Unexpected missing transaction for nonce %d", nonce)
			return
		}
		transactions = append(transactions, txMeta)
	}

	// 发送响应
	resp := TxSyncRsp{
		Type:         TxSyncResponse,
		RequestID:    req.RequestID,
		Address:      req.Address,
		Transactions: transactions,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return
	}

	n.topic.txSyncTopic.Publish(n.ctx, data)
}

// 处理交易同步响应
func (n *Layer2Node) handleTxSyncResponse(msg *pubsub.Message) {
	var resp TxSyncRsp
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return
	}

	// 按顺序处理交易
	for _, tx := range resp.Transactions {
		if n.validateTransaction(tx.Transaction) {
			// 只有当前在交易池中的才加入，否则加入到历史池子中，不然会导致当选排序器时重复消费
			if tx.Source == TxSourcePool {
				n.txPool.Store(tx.Transaction.Hash, tx.Transaction)
			} else {
				n.txHistory.Store(tx.Transaction.Hash, tx.Transaction)
			}
			fmt.Printf("Synced missing transaction: from=%s, nonce=%d\n",
				tx.Transaction.From, tx.Transaction.Nonce)
		}
	}
	// 如果完美解决
	n.mu.Lock()
	n.isSyncing = false
	n.currentSyncRequestID = ""
	n.mu.Unlock()
}

// 区块投票请求结构
type BlockVoteReq struct {
	Type      MessageType `json:"type"`
	RequestID string      `json:"request_id"`
	Address   string      `json:"address"`
	Block     Block       `json:"block"`
}

// 区块投票响应结构
type BlockVoteRsp struct {
	Type      MessageType `json:"type"`
	RequestID string      `json:"request_id"`
	Address   string      `json:"address"`
	BlockVote BlockVote   `json:"block_vote"`
}

// 处理区块投票请求
func (n *Layer2Node) handleBlockVoteRequest(msg *pubsub.Message) {
	var req BlockVoteReq
	var vote BlockVote
	var err error
	if err = json.Unmarshal(msg.Data, &req); err != nil {
		return
	}
	// 如果自己不是是其他提案者，则不需要投票
	address, err := PublicKeyToAddress(n.publicKey)
	if err != nil {
		return
	}
	if !n.electionMgr.IsProposer(address) || address == req.Block.Proposer {
		return
	}

	if n.BlockVoteValidation(req.Block, false) {
		vote, err = n.createBlockVote(req.Block, true)
	} else {
		vote, err = n.createBlockVote(req.Block, false)
	}
	if err != nil {
		return
	}

	// 发送响应
	resp := BlockVoteRsp{
		Type:      BlockVoteResponse,
		RequestID: req.RequestID,
		Address:   req.Address,
		BlockVote: vote,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	n.topic.blockVoteTopic.Publish(n.ctx, data)
	logger.Info("已发送投票")
}

// 创建区块投票
func (n *Layer2Node) createBlockVote(block Block, isValid bool) (BlockVote, error) {
	addr, err := PublicKeyToAddress(n.publicKey)
	if err != nil {
		logger.Error("Failed to get address.")
		return BlockVote{}, err
	}
	vote := BlockVote{
		BlockHash:    block.Hash,
		BlockHeight:  block.Height,
		Approve:      isValid,
		VoterAddress: addr,
		Timestamp:    time.Now(),
	}

	// 签名投票
	err = SignBlockVote(&vote, n)
	if err != nil {
		return BlockVote{}, err
	}
	if err != nil {
		logger.Errorf("Failed to sign vote: %v", err)
		return BlockVote{}, err
	}

	return vote, nil
}

func SignBlockVote(vote *BlockVote, node *Layer2Node) error {
	// 序列化区块数据
	voteData := struct {
		BlockHash    string
		BlockHeight  uint64
		Approve      bool
		VoterAddress string
		Timestamp    time.Time
	}{
		BlockHash:    vote.BlockHash,
		BlockHeight:  vote.BlockHeight,
		Approve:      vote.Approve,
		VoterAddress: vote.VoterAddress,
		Timestamp:    vote.Timestamp,
	}

	data, err := json.Marshal(voteData)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// 使用节点私钥签名
	signature, err := node.privateKey.Sign(data)
	if err != nil {
		return fmt.Errorf("failed to sign block: %w", err)
	}

	vote.Signature = signature
	return nil
}

// DAC 消息类型
const (
	DACStateRequest MessageType = iota + 100
	DACStateResponse
	DACProofRequest
	DACProofResponse
	DACMemberJoin
	DACMemberLeave
	DACMemberUpdate
	DACStateCommitReq
	DACStateCommitResp
	DACProofSubmit
)

// DAC 状态请求
type DACStateReq struct {
	Type      MessageType `json:"type"`
	Address   string      `json:"address"`
	RequestID string      `json:"request_id"`
}

// DAC 状态响应
type DACStateRsp struct {
	Type      MessageType `json:"type"`
	RequestID string      `json:"request_id"`
	Account   Account     `json:"account"`
	Proof     [][]byte    `json:"proof"`
}

// DAC成员消息结构
type DACMemberMessage struct {
	Type      MessageType   `json:"type"`
	RequestID string        `json:"request_id"`
	Member    DACMemberData `json:"member"`
	Signature []byte        `json:"signature"`
}

//// DACMemberBucketData 表示质押桶的传输友好形式
//type DACMemberBucketData struct {
//	ID            uint64   `json:"id"`
//	StakeAmount   uint64   `json:"stake_amount"`
//	MappedValue   *big.Int `json:"mapped_value_string,string"`
//	CurrentWeight *big.Int `json:"current_weight_string,string"`
//}

// 扩展DACMemberData包含桶信息
type DACMemberData struct {
	Address        string                           `json:"address"`
	PublicKeyBytes []byte                           `json:"public_key_bytes"`
	Status         int                              `json:"status"`
	StakeAmount    uint64                           `json:"stake_amount"`
	JoinTime       time.Time                        `json:"join_time"`
	DataProvided   uint64                           `json:"data_provided"`
	LastActiveTime time.Time                        `json:"last_active_time"`
	Buckets        map[uint64]consensus.StakeBucket `json:"buckets"`
}

// DAC状态同步请求
type DACSyncReq struct {
	Type        MessageType `json:"type"`
	RequestID   string      `json:"request_id"`
	CurrentTerm uint64      `json:"current_term"`
}

// DAC状态同步响应
type DACSyncRsp struct {
	Type           MessageType              `json:"type"`
	RequestID      string                   `json:"request_id"`
	CurrentMembers []string                 `json:"current_members"`
	CurrentTerm    uint64                   `json:"current_term"`
	LastRotation   time.Time                `json:"last_rotation"`
	NextRotation   time.Time                `json:"next_rotation"`
	Members        map[string]DACMemberData `json:"members"`
}

// DAC分片结构，用于大数据传输
type DACSyncChunk struct {
	Type        MessageType `json:"type"`
	RequestID   string      `json:"request_id"`
	ChunkIndex  int         `json:"chunk_index"`
	TotalChunks int         `json:"total_chunks"`
	Data        []byte      `json:"data"`
	IsFinal     bool        `json:"is_final"`
}

// 处理 DAC 状态请求
func (n *Layer2Node) handleDACStateRequest(msg *pubsub.Message) {
	var req DACStateReq
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return
	}

	// 只有 DAC 成员才处理请求
	if !n.isDACMember {
		return
	}

	// 获取账户状态和证明
	account, err := n.dacMgr.GetAccountState(req.Address)
	if err != nil {
		return
	}

	proof, err := n.dacMgr.GetAccountProof(req.Address)
	if err != nil {
		return
	}

	// 发送响应
	resp := DACStateRsp{
		Type:      DACStateResponse,
		RequestID: req.RequestID,
		Account:   *account,
		Proof:     proof,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return
	}

	n.topic.dacTopic.Publish(n.ctx, data)
}

// 订阅DAC相关主题
func (n *Layer2Node) subscribeToDACTopics() error {
	var err error

	// 检查是否已经订阅
	if n.topic.dacTopic == nil {
		return nil
	}

	// 订阅DAC消息
	dacSub, err := n.topic.dacTopic.Subscribe()
	if err != nil {
		return fmt.Errorf("订阅DAC主题失败: %w", err)
	}

	// 处理DAC消息
	go func() {
		for {
			msg, err := dacSub.Next(n.ctx)
			if err != nil {
				if n.ctx.Err() != nil {
					// 上下文已取消，退出循环
					return
				}
				logger.Errorf("获取DAC消息失败: %v", err)
				continue
			}

			// 忽略自己发送的消息
			if msg.ReceivedFrom == n.host.ID() {
				continue
			}

			// 处理DAC消息
			go n.handleDACMessage(msg)
		}
	}()

	return nil
}

// 处理DAC消息
func (n *Layer2Node) handleDACMessage(msg *pubsub.Message) {
	// 解析消息类型
	var msgData struct {
		Type MessageType `json:"type"`
	}

	if err := json.Unmarshal(msg.Data, &msgData); err != nil {
		logger.Errorf("解析DAC消息类型失败: %v", err)
		return
	}

	// 根据消息类型分发处理
	switch msgData.Type {
	case DACStateRequest:
		n.handleDACStateRequest(msg)
	case DACMemberJoin, DACMemberLeave, DACMemberUpdate:
		n.handleDACMemberMessage(msg)
	case DACStateRootRequestType:
		n.handleDACStateRootRequest(msg)
	case DACStateRootResponseType:
		n.handleDACStateRootResponse(msg)
	case DACBlockCommitRequestType:
		n.handleDACBlockCommitRequest(msg)
	case DACBlockCommitResponseType:
		n.handleDACBlockCommitResponse(msg)
	case DACProofSubmitRequestType, DACProofSubmitResponseType:
		// 处理证明提交相关消息
		logger.Info("收到DAC证明提交相关消息")
	default:
		logger.Warn("未知DAC消息类型: %v", msgData.Type)
	}
}

// 处理DAC成员消息
func (n *Layer2Node) handleDACMemberMessage(msg *pubsub.Message) {
	var dacMsg DACMemberMessage
	if err := json.Unmarshal(msg.Data, &dacMsg); err != nil {
		logger.Errorf("解析DAC成员消息失败: %v", err)
		return
	}

	// 验证消息签名
	if !n.verifyDACMemberMessage(&dacMsg) {
		logger.Warn("DAC成员消息签名验证失败")
		return
	}

	// 将DACMemberData转换为DACMember
	member, err := n.convertToDACMember(dacMsg.Member)
	if err != nil {
		logger.Errorf("转换DAC成员数据失败: %v", err)
		return
	}

	// 根据消息类型处理
	switch dacMsg.Type {
	case DACMemberJoin:
		if n.dacMgr.AddMember(member) {
			logger.Infof("新DAC成员已加入: %s", member.Address)
		}
	case DACMemberLeave:
		if n.dacMgr.RemoveMember(member.Address) {
			logger.Infof("DAC成员已离开: %s", member.Address)
		}
	case DACMemberUpdate:
		if n.dacMgr.UpdateMember(member) {
			logger.Infof("DAC成员信息已更新: %s", member.Address)
		}
	}
}

// 验证DAC成员消息签名
func (n *Layer2Node) verifyDACMemberMessage(message *DACMemberMessage) bool {
	// 需要验证的数据
	signData := struct {
		Type      MessageType   `json:"type"`
		RequestID string        `json:"request_id"`
		Member    DACMemberData `json:"member"`
	}{
		Type:      message.Type,
		RequestID: message.RequestID,
		Member:    message.Member,
	}

	// 序列化数据
	data, err := json.Marshal(signData)
	if err != nil {
		logger.Errorf("序列化验证数据失败: %v", err)
		return false
	}

	// 从公钥字节恢复公钥
	pubKey, err := crypto.UnmarshalPublicKey(message.Member.PublicKeyBytes)
	if err != nil {
		logger.Errorf("解析公钥失败: %v", err)
		return false
	}

	// 验证签名
	valid, err := pubKey.Verify(data, message.Signature)
	if err != nil {
		logger.Errorf("验证签名失败: %v", err)
		return false
	}

	return valid
}

// 将DACMemberData转换为DACMember
func (n *Layer2Node) convertToDACMember(data DACMemberData) (*dac.DACMember, error) {
	// 从字节恢复公钥
	pubKey, err := crypto.UnmarshalPublicKey(data.PublicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	// 创建桶映射
	buckets := make(map[uint64]*consensus.StakeBucket)
	for id, bucketData := range data.Buckets {
		bucket := &consensus.StakeBucket{
			ID:            id,
			StakeAmount:   bucketData.StakeAmount,
			MappedValue:   bucketData.MappedValue,
			CurrentWeight: bucketData.CurrentWeight,
		}
		buckets[id] = bucket
	}

	// 创建DACMember实例
	member := &dac.DACMember{
		Address:        data.Address,
		PublicKey:      pubKey,
		Status:         dac.DACMemberStatus(data.Status),
		StakeAmount:    data.StakeAmount,
		JoinTime:       data.JoinTime,
		DataProvided:   data.DataProvided,
		LastActiveTime: data.LastActiveTime,
		Buckets:        buckets,
	}

	return member, nil
}

// 广播DAC成员消息
func (n *Layer2Node) BroadcastDACMemberMessage(dacMsg DACMemberMessage) error {
	// 序列化DAC消息
	data, err := json.Marshal(dacMsg)
	if err != nil {
		return fmt.Errorf("序列化DAC成员消息失败: %w", err)
	}

	// 通过DAC主题发布消息
	return n.topic.dacTopic.Publish(n.ctx, data)
}

// 创建DAC成员加入消息
func (n *Layer2Node) CreateDACMemberJoinMessage(member *dac.DACMember) (*DACMemberMessage, error) {
	if member == nil || member.PublicKey == nil {
		return nil, fmt.Errorf("无效的DAC成员数据")
	}

	// 获取公钥字节
	pubKeyBytes, err := crypto.MarshalPublicKey(member.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("序列化公钥失败: %w", err)
	}

	// 转换桶数据
	buckets := make(map[uint64]consensus.StakeBucket)
	for id, bucket := range member.Buckets {
		bucketData := consensus.StakeBucket{
			ID:            id,
			StakeAmount:   bucket.StakeAmount,
			MappedValue:   bucket.MappedValue,
			CurrentWeight: bucket.CurrentWeight,
		}
		buckets[id] = bucketData
	}

	// 创建成员数据
	memberData := DACMemberData{
		Address:        member.Address,
		PublicKeyBytes: pubKeyBytes,
		Status:         int(member.Status),
		StakeAmount:    member.StakeAmount,
		JoinTime:       member.JoinTime,
		DataProvided:   member.DataProvided,
		LastActiveTime: member.LastActiveTime,
		Buckets:        buckets, // 添加桶信息
	}

	// 创建消息
	message := DACMemberMessage{
		Type:      DACMemberJoin,
		RequestID: uuid.New().String(),
		Member:    memberData,
	}

	// 签名消息
	if err := n.signDACMemberMessage(&message); err != nil {
		return nil, err
	}

	return &message, nil
}

// 签名DAC成员消息
func (n *Layer2Node) signDACMemberMessage(message *DACMemberMessage) error {
	// 需要签名的数据
	signData := struct {
		Type      MessageType   `json:"type"`
		RequestID string        `json:"request_id"`
		Member    DACMemberData `json:"member"`
	}{
		Type:      message.Type,
		RequestID: message.RequestID,
		Member:    message.Member,
	}

	// 序列化数据
	data, err := json.Marshal(signData)
	if err != nil {
		return fmt.Errorf("序列化签名数据失败: %w", err)
	}

	// 使用节点私钥签名
	signature, err := n.privateKey.Sign(data)
	if err != nil {
		return fmt.Errorf("签名DAC成员消息失败: %w", err)
	}

	message.Signature = signature
	return nil
}

// 更新DAC状态
func (n *Layer2Node) updateDACState(resp DACSyncRsp) {
	if n.dacMgr == nil {
		logger.Error("DAC管理器未初始化")
		return
	}

	// 将DACMemberData转换为DACMember
	members := make(map[string]*dac.DACMember)
	for addr, memberData := range resp.Members {
		member, err := n.convertToDACMember(memberData)
		if err != nil {
			logger.Errorf("转换DAC成员数据失败: %v", err)
			continue
		}
		members[addr] = member
	}

	// 更新DAC状态
	state := &dac.DACState{
		CurrentMembers:   resp.CurrentMembers,
		CurrentTerm:      resp.CurrentTerm,
		LastRotation:     resp.LastRotation,
		NextRotationTime: resp.NextRotation,
		Members:          members,
		RotationInterval: n.dacMgr.GetState().RotationInterval, // 保持原有的轮换间隔
	}

	// 更新DAC管理器状态
	n.dacMgr.SetState(state)

	logger.Infof("已更新DAC状态: 当前任期=%d, 活跃成员数=%d", state.CurrentTerm, len(state.CurrentMembers))
}

// 创建DAC成员离开消息
func (n *Layer2Node) CreateDACMemberLeaveMessage(member *dac.DACMember) (*DACMemberMessage, error) {
	if member == nil || member.PublicKey == nil {
		return nil, fmt.Errorf("无效的DAC成员数据")
	}

	// 获取公钥字节
	pubKeyBytes, err := crypto.MarshalPublicKey(member.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("序列化公钥失败: %w", err)
	}

	// 转换桶数据
	buckets := make(map[uint64]consensus.StakeBucket)
	for id, bucket := range member.Buckets {
		bucketData := consensus.StakeBucket{
			ID:            id,
			StakeAmount:   bucket.StakeAmount,
			MappedValue:   bucket.MappedValue,
			CurrentWeight: bucket.CurrentWeight,
		}
		buckets[id] = bucketData
	}

	// 创建成员数据
	memberData := DACMemberData{
		Address:        member.Address,
		PublicKeyBytes: pubKeyBytes,
		Status:         int(member.Status),
		StakeAmount:    member.StakeAmount,
		JoinTime:       member.JoinTime,
		DataProvided:   member.DataProvided,
		LastActiveTime: member.LastActiveTime,
		Buckets:        buckets, // 添加桶信息
	}

	// 创建消息
	message := DACMemberMessage{
		Type:      DACMemberLeave,
		RequestID: uuid.New().String(),
		Member:    memberData,
	}

	// 签名消息
	if err := n.signDACMemberMessage(&message); err != nil {
		return nil, err
	}

	return &message, nil
}

// 创建DAC成员更新消息
func (n *Layer2Node) CreateDACMemberUpdateMessage(member *dac.DACMember) (*DACMemberMessage, error) {
	if member == nil || member.PublicKey == nil {
		return nil, fmt.Errorf("无效的DAC成员数据")
	}

	// 获取公钥字节
	pubKeyBytes, err := crypto.MarshalPublicKey(member.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("序列化公钥失败: %w", err)
	}

	// 转换桶数据
	buckets := make(map[uint64]consensus.StakeBucket)
	for id, bucket := range member.Buckets {
		bucketData := consensus.StakeBucket{
			ID:            id,
			StakeAmount:   bucket.StakeAmount,
			MappedValue:   bucket.MappedValue,
			CurrentWeight: bucket.CurrentWeight,
		}
		buckets[id] = bucketData
	}

	// 创建成员数据
	memberData := DACMemberData{
		Address:        member.Address,
		PublicKeyBytes: pubKeyBytes,
		Status:         int(member.Status),
		StakeAmount:    member.StakeAmount,
		JoinTime:       member.JoinTime,
		DataProvided:   member.DataProvided,
		LastActiveTime: member.LastActiveTime,
		Buckets:        buckets, // 添加桶信息
	}

	// 创建消息
	message := DACMemberMessage{
		Type:      DACMemberUpdate,
		RequestID: uuid.New().String(),
		Member:    memberData,
	}

	// 签名消息
	if err := n.signDACMemberMessage(&message); err != nil {
		return nil, err
	}

	return &message, nil
}
