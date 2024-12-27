package p2p

import (
	. "FPoS/types"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
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
	currentNonce := n.stateDB.GetNonce(tx.From)
	if tx.Nonce > currentNonce+1 {
		fmt.Printf("Transaction nonce gap detected: current=%d, received=%d\n",
			currentNonce, tx.Nonce)

		// 触发交易同步
		go func() {
			n.mu.Lock()
			if n.isSyncing {
				n.mu.Unlock()
				return
			}
			// 在接收到响应并且装载完后再解除锁定
			n.isSyncing = true
			n.mu.Unlock()

			// 请求缺失的交易
			if err := n.syncMissingTransactions(tx.From, currentNonce+1, tx.Nonce); err != nil {
				fmt.Printf("Missing transactions sync failed: %v\n", err)
			}

		}()
		return false
	} else if tx.Nonce < currentNonce+1 {
		fmt.Printf("Transaction nonce too low: expected %d, got %d\n",
			currentNonce+1, tx.Nonce)
		return false
	}

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
			if _, exists := n.txPool.Load(tx.Hash); exists {
				fmt.Println("本节点为排序器节点，但交易池中交易未正确清除")
				return false
			}
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
			if _, exists := n.txPool.Load(tx.Hash); !exists {
				fmt.Println("本节点为非排序器节点，但交易池中未含有该交易")
				return false
			}
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
	currentNonce := n.stateDB.GetNonce(tx.From)
	if !isHistoricalBlock && tx.Nonce > currentNonce+1 {
		fmt.Printf("block交易nonce无效: 期望 %d, 实际 %d\n", currentNonce+1, tx.Nonce)
		return false
	}

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
