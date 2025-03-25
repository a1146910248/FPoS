package p2p

import (
	"FPoS/types"
	. "FPoS/types"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// 消息类型常量
const (
	DACStateRootRequestType MessageType = iota + 200
	DACStateRootResponseType
	DACBlockCommitRequestType
	DACBlockCommitResponseType
	DACProofSubmitRequestType
	DACProofSubmitResponseType
)

// DACStateRootRequest 请求DAC状态根
type DACStateRootRequest struct {
	Type      MessageType `json:"type"`
	RequestID string      `json:"request_id"`
	Signature []byte      `json:"signature"` // 排序器签名
}

// DACStateRootResponse DAC状态根响应
type DACStateRootResponse struct {
	Type        MessageType `json:"type"`
	RequestID   string      `json:"request_id"`
	AccountRoot string      `json:"account_root"`  // 账户状态根
	TxRoot      string      `json:"tx_root"`       // 交易状态根
	BlockHeight uint64      `json:"block_height"`  // 当前区块高度
	DACMemberID string      `json:"dac_member_id"` // DAC成员ID
	Signature   []byte      `json:"signature"`     // DAC成员签名
}

// DACBlockCommitRequest 区块提交请求
type DACBlockCommitRequest struct {
	Type        MessageType `json:"type"`
	RequestID   string      `json:"request_id"`
	Block       Block       `json:"block"`         // 完整区块
	L1StateRoot string      `json:"l1_state_root"` // 一层状态根
	Signature   []byte      `json:"signature"`     // 排序器签名
}

// DACBlockCommitResponse 区块提交响应
type DACBlockCommitResponse struct {
	Type           MessageType `json:"type"`
	RequestID      string      `json:"request_id"`
	BlockHeight    uint64      `json:"block_height"`
	NewAccountRoot string      `json:"new_account_root"` // 新账户状态根
	NewTxRoot      string      `json:"new_tx_root"`      // 新交易状态根
	Proof          [][]byte    `json:"proof"`            // 默克尔证明
	DACMemberID    string      `json:"dac_member_id"`    // DAC成员ID
	Signature      []byte      `json:"signature"`        // DAC成员签名
}

// DACProofSubmitRequest 证明提交请求
type DACProofSubmitRequest struct {
	Type        MessageType `json:"type"`
	RequestID   string      `json:"request_id"`
	BlockHeight uint64      `json:"block_height"`
	BlockHash   string      `json:"block_hash"`
	StateRoot   string      `json:"state_root"`   // 二层状态根
	AccountRoot string      `json:"account_root"` // DAC账户根
	TxRoot      string      `json:"tx_root"`      // DAC交易根
	Proofs      [][]byte    `json:"proofs"`       // 证明集合
	Signature   []byte      `json:"signature"`    // 排序器签名
}

// DACProofSubmitResponse 证明提交响应
type DACProofSubmitResponse struct {
	Type        MessageType `json:"type"`
	RequestID   string      `json:"request_id"`
	Success     bool        `json:"success"`
	TxHash      string      `json:"tx_hash"`       // L1交易哈希
	DACMemberID string      `json:"dac_member_id"` // DAC成员ID
	Signature   []byte      `json:"signature"`     // DAC成员签名
}

// 存储DAC状态根响应
type DACStateRootResp struct {
	accountRoot string
	txRoot      string
	blockHeight uint64
	responses   map[string]bool   // DAC成员ID -> 是否响应
	signatures  map[string][]byte // DAC成员ID -> 签名
}

// 存储区块提交响应
type DACBlockCommitResp struct {
	mu             sync.RWMutex // 添加读写锁
	blockHeight    uint64
	newAccountRoot string
	newTxRoot      string
	proofs         map[string][][]byte // DAC成员ID -> 证明
	responses      map[string]bool     // DAC成员ID -> 是否响应
	signatures     map[string][]byte   // DAC成员ID -> 签名
}

// 响应收集映射
var (
	stateRootResponses   sync.Map // requestID -> DACStateRootResp
	blockCommitResponses sync.Map // requestID -> DACBlockCommitResp
)

// 添加全局变量用于请求跟踪和超时处理
var dacRequestHandlers sync.Map // requestID -> 处理函数

// 排序器请求DAC状态根
func (n *Layer2Node) RequestDACStateRoot() (string, string, error) {
	// 只有排序器可以请求状态根
	if !n.isSequencer {
		return "", "", fmt.Errorf("只有排序器节点可以请求DAC状态根")
	}

	// 创建请求
	requestID := uuid.New().String()
	req := DACStateRootRequest{
		Type:      DACStateRootRequestType,
		RequestID: requestID,
	}

	// 签名请求
	reqData, err := json.Marshal(struct {
		Type      MessageType `json:"type"`
		RequestID string      `json:"request_id"`
	}{
		Type:      req.Type,
		RequestID: req.RequestID,
	})

	if err != nil {
		return "", "", fmt.Errorf("序列化签名数据失败: %w", err)
	}

	signature, err := n.privateKey.Sign(reqData)
	if err != nil {
		return "", "", fmt.Errorf("签名请求失败: %w", err)
	}
	req.Signature = signature

	// 广播请求
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", "", fmt.Errorf("序列化请求失败: %w", err)
	}

	err = n.topic.dacTopic.Publish(n.ctx, reqBytes)
	if err != nil {
		return "", "", fmt.Errorf("广播请求失败: %w", err)
	}

	logger.Infof("已发送DAC状态根请求，请求ID: %s", requestID)

	// 等待足够的响应
	accountRoot, txRoot, err := n.waitForStateRootResponses(requestID)
	if err != nil {
		return "", "", err
	}

	return accountRoot, txRoot, nil
}

// 等待状态根响应
func (n *Layer2Node) waitForStateRootResponses(requestID string) (string, string, error) {
	// 创建超时通道
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查响应
			respValue, exists := stateRootResponses.Load(requestID)
			if !exists {
				continue
			}

			resp := respValue.(*DACStateRootResp)

			// 获取DAC成员数量
			dacState := n.dacMgr.GetState()
			requiredResponses := len(dacState.CurrentMembers) * 2 / 3 // 2/3多数

			if len(resp.responses) >= requiredResponses {
				// 检查是否所有响应一致
				if isConsistent := n.verifyStateRootConsistency(resp); isConsistent {
					// 清理响应
					stateRootResponses.Delete(requestID)
					return resp.accountRoot, resp.txRoot, nil
				}
			}

		case <-timeout:
			// 超时
			stateRootResponses.Delete(requestID)
			return "", "", fmt.Errorf("等待DAC状态根响应超时")
		}
	}
}

// 验证状态根一致性
func (n *Layer2Node) verifyStateRootConsistency(resp *DACStateRootResp) bool {
	// 简单实现，确保所有响应的状态根一致
	// 实际应该对多个响应进行投票
	return true
}

// 处理DAC状态根请求
func (n *Layer2Node) handleDACStateRootRequest(msg *pubsub.Message) {
	// 只有DAC成员处理
	if !n.isDACMember {
		return
	}

	var req DACStateRootRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		logger.Errorf("解析DAC状态根请求失败: %v", err)
		return
	}

	// 验证排序器签名
	// 获取当前排序器地址
	state := GetStats()
	sequencerAddr := state.CurrentSequencer
	sequencer, exists := n.electionMgr.Validators[sequencerAddr]
	if !exists || sequencer.PublicKey == nil {
		logger.Errorf("找不到排序器公钥信息")
		return
	}

	// 验证签名
	reqData, err := json.Marshal(struct {
		Type      MessageType `json:"type"`
		RequestID string      `json:"request_id"`
	}{
		Type:      req.Type,
		RequestID: req.RequestID,
	})

	if err != nil {
		logger.Errorf("序列化验证数据失败: %v", err)
		return
	}

	valid, err := sequencer.PublicKey.Verify(reqData, req.Signature)
	if err != nil || !valid {
		logger.Errorf("排序器签名验证失败: %v", err)
		return
	}

	// 获取当前状态根
	accountRoot := n.dacMgr.GetAccountRoot()
	txRoot := n.dacMgr.GetTxRoot()
	blockHeight := n.latestBlock

	// 准备响应
	addr, _ := PublicKeyToAddress(n.publicKey)
	resp := DACStateRootResponse{
		Type:        DACStateRootResponseType,
		RequestID:   req.RequestID,
		AccountRoot: accountRoot,
		TxRoot:      txRoot,
		BlockHeight: blockHeight,
		DACMemberID: addr,
	}

	// 签名响应
	respData, err := json.Marshal(struct {
		Type        MessageType `json:"type"`
		RequestID   string      `json:"request_id"`
		AccountRoot string      `json:"account_root"`
		TxRoot      string      `json:"tx_root"`
		BlockHeight uint64      `json:"block_height"`
		DACMemberID string      `json:"dac_member_id"`
	}{
		Type:        resp.Type,
		RequestID:   resp.RequestID,
		AccountRoot: resp.AccountRoot,
		TxRoot:      resp.TxRoot,
		BlockHeight: resp.BlockHeight,
		DACMemberID: resp.DACMemberID,
	})

	if err != nil {
		logger.Errorf("序列化签名数据失败: %v", err)
		return
	}

	signature, err := n.privateKey.Sign(respData)
	if err != nil {
		logger.Errorf("签名响应失败: %v", err)
		return
	}
	resp.Signature = signature

	// 发送响应
	respBytes, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("序列化响应失败: %v", err)
		return
	}

	err = n.topic.dacTopic.Publish(n.ctx, respBytes)
	if err != nil {
		logger.Errorf("发送状态根响应失败: %v", err)
		return
	}

	logger.Infof("已发送DAC状态根响应，账户根: %s, 交易根: %s", accountRoot, txRoot)
}

// 处理DAC状态根响应
func (n *Layer2Node) handleDACStateRootResponse(msg *pubsub.Message) {
	// 只有排序器处理
	if !n.isSequencer {
		return
	}

	var resp DACStateRootResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		logger.Errorf("解析DAC状态根响应失败: %v", err)
		return
	}

	// 验证DAC成员签名
	members := n.dacMgr.GetState().Members
	member, exists := members[resp.DACMemberID]
	if !exists || member.PublicKey == nil {
		logger.Errorf("找不到DAC成员公钥信息")
		return
	}

	// 验证签名
	respData, err := json.Marshal(struct {
		Type        MessageType `json:"type"`
		RequestID   string      `json:"request_id"`
		AccountRoot string      `json:"account_root"`
		TxRoot      string      `json:"tx_root"`
		BlockHeight uint64      `json:"block_height"`
		DACMemberID string      `json:"dac_member_id"`
	}{
		Type:        resp.Type,
		RequestID:   resp.RequestID,
		AccountRoot: resp.AccountRoot,
		TxRoot:      resp.TxRoot,
		BlockHeight: resp.BlockHeight,
		DACMemberID: resp.DACMemberID,
	})

	if err != nil {
		logger.Errorf("序列化验证数据失败: %v", err)
		return
	}

	valid, err := member.PublicKey.Verify(respData, resp.Signature)
	if err != nil || !valid {
		logger.Errorf("DAC成员签名验证失败: %v", err)
		return
	}

	// 记录响应
	respValue, _ := stateRootResponses.LoadOrStore(resp.RequestID, &DACStateRootResp{
		accountRoot: resp.AccountRoot,
		txRoot:      resp.TxRoot,
		blockHeight: resp.BlockHeight,
		responses:   make(map[string]bool),
		signatures:  make(map[string][]byte),
	})

	responseData := respValue.(*DACStateRootResp)
	responseData.responses[resp.DACMemberID] = true
	responseData.signatures[resp.DACMemberID] = resp.Signature

	// 如果是首次记录，设置状态根
	responseData.accountRoot = resp.AccountRoot
	responseData.txRoot = resp.TxRoot
	responseData.blockHeight = resp.BlockHeight

	logger.Infof("收到DAC成员 %s 的状态根响应: 账户根=%s, 交易根=%s",
		resp.DACMemberID, resp.AccountRoot, resp.TxRoot)
}

// 修改提交区块到DAC网络的方法，使每次提交都有唯一的请求ID
func (n *Layer2Node) SubmitBlockToDACNetwork(requestID string, block types.Block, l1StateRoot string) error {
	// 只有排序器可以提交区块
	if !n.isSequencer {
		return fmt.Errorf("只有排序器节点可以提交区块到DAC网络")
	}

	req := DACBlockCommitRequest{
		Type:        DACBlockCommitRequestType,
		RequestID:   requestID,
		Block:       block,
		L1StateRoot: l1StateRoot,
	}

	// 签名请求
	reqData, err := json.Marshal(struct {
		BlockHeight uint64 `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		L1StateRoot string `json:"l1_state_root"`
	}{
		BlockHeight: block.Height,
		BlockHash:   block.Hash,
		L1StateRoot: l1StateRoot,
	})

	if err != nil {
		return fmt.Errorf("序列化签名数据失败: %w", err)
	}

	signature, err := n.privateKey.Sign(reqData)
	if err != nil {
		return fmt.Errorf("签名请求失败: %w", err)
	}
	req.Signature = signature

	// 初始化响应收集
	blockCommitResponses.Store(requestID, &DACBlockCommitResp{
		blockHeight: block.Height,
		proofs:      make(map[string][][]byte),
		responses:   make(map[string]bool),
		signatures:  make(map[string][]byte),
	})

	// 广播请求
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	err = n.topic.dacTopic.Publish(n.ctx, reqBytes)
	if err != nil {
		return fmt.Errorf("广播请求失败: %w", err)
	}

	fmt.Printf("已发送区块提交请求，区块高度: %d, 哈希: %s, 请求ID: %s\n",
		block.Height, block.Hash, requestID)

	return nil
}

// 处理区块提交请求
func (n *Layer2Node) handleDACBlockCommitRequest(msg *pubsub.Message) {
	// 只有DAC成员处理
	if !n.isDACMember {
		return
	}

	var req DACBlockCommitRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		logger.Errorf("解析区块提交请求失败: %v", err)
		return
	}

	// 验证排序器签名
	// 获取当前排序器地址
	state := GetStats()
	sequencerAddr := state.CurrentSequencer
	sequencer, exists := n.electionMgr.Validators[sequencerAddr]
	if !exists || sequencer.PublicKey == nil {
		logger.Errorf("找不到排序器公钥信息")
		return
	}

	// 验证签名
	reqData, err := json.Marshal(struct {
		BlockHeight uint64 `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		L1StateRoot string `json:"l1_state_root"`
	}{
		BlockHeight: req.Block.Height,
		BlockHash:   req.Block.Hash,
		L1StateRoot: req.L1StateRoot,
	})

	if err != nil {
		logger.Errorf("序列化验证数据失败: %v", err)
		return
	}

	valid, err := sequencer.PublicKey.Verify(reqData, req.Signature)
	if err != nil || !valid {
		logger.Errorf("排序器签名验证失败: %v", err)
		return
	}

	// 验证Layer 1状态根
	// 这里需要与实际的Layer 1验证逻辑对接
	// ...

	// 处理区块中的交易，更新DAC状态
	if err := n.processDACBlockTransactions(req.Block); err != nil {
		logger.Errorf("处理区块交易失败: %v", err)
		return
	}

	// 获取新的状态根
	newAccountRoot := n.dacMgr.GetAccountRoot()
	newTxRoot := n.dacMgr.GetTxRoot()

	// 生成默克尔证明 (以第一个账户为例)
	var proof [][]byte
	if len(req.Block.Transactions) > 0 {
		firstTx := req.Block.Transactions[0]
		proof, err = n.dacMgr.GetAccountProof(firstTx.From)
		if err != nil {
			logger.Errorf("生成默克尔证明失败: %v", err)
			return
		}
	}

	// 准备响应
	addr, _ := PublicKeyToAddress(n.publicKey)
	resp := DACBlockCommitResponse{
		Type:           DACBlockCommitResponseType,
		RequestID:      req.RequestID,
		BlockHeight:    req.Block.Height,
		NewAccountRoot: newAccountRoot,
		NewTxRoot:      newTxRoot,
		Proof:          proof,
		DACMemberID:    addr,
	}

	// 签名响应
	respData, err := json.Marshal(struct {
		RequestID      string `json:"request_id"`
		BlockHeight    uint64 `json:"block_height"`
		NewAccountRoot string `json:"new_account_root"`
		NewTxRoot      string `json:"new_tx_root"`
		DACMemberID    string `json:"dac_member_id"`
	}{
		RequestID:      resp.RequestID,
		BlockHeight:    resp.BlockHeight,
		NewAccountRoot: resp.NewAccountRoot,
		NewTxRoot:      resp.NewTxRoot,
		DACMemberID:    resp.DACMemberID,
	})

	if err != nil {
		logger.Errorf("序列化签名数据失败: %v", err)
		return
	}

	signature, err := n.privateKey.Sign(respData)
	if err != nil {
		logger.Errorf("签名响应失败: %v", err)
		return
	}
	resp.Signature = signature

	// 发送响应
	respBytes, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("序列化响应失败: %v", err)
		return
	}

	err = n.topic.dacTopic.Publish(n.ctx, respBytes)
	if err != nil {
		logger.Errorf("发送区块提交响应失败: %v", err)
		return
	}

	logger.Infof("已发送区块提交响应，区块高度: %d, 新账户根: %s, 新交易根: %s",
		req.Block.Height, newAccountRoot, newTxRoot)
}

// 处理区块中的交易，更新DAC状态
func (n *Layer2Node) processDACBlockTransactions(block Block) error {
	// 跟踪需要更新的账户地址
	affectedAccounts := make(map[string]bool)

	// 第一步：收集所有受影响的账户地址
	for _, tx := range block.Transactions {
		affectedAccounts[tx.From] = true
		affectedAccounts[tx.To] = true
	}
	seq := GetStats().CurrentSequencer
	affectedAccounts[seq] = true

	// 第二步：获取所有受影响账户的初始状态
	accountStates := make(map[string]*Account)
	for addr := range affectedAccounts {
		account, err := n.dacMgr.GetAccountState(addr)
		if err != nil {
			// 如果账户不存在，创建一个新账户
			account = &Account{
				Address: addr,
				Balance: 0,
				Nonce:   0,
			}
		}
		accountStates[addr] = account
	}

	// 第三步：累积所有交易对账户的影响
	for _, tx := range block.Transactions {
		// 获取当前状态
		fromAccount := accountStates[tx.From]
		toAccount := accountStates[tx.To]
		seqAccount := accountStates[seq]

		// 计算gas费用
		gas := tx.GasUsed * tx.GasPrice

		// 更新余额 - 累积变化
		fromAccount.Balance -= (tx.Value + gas)
		toAccount.Balance += tx.Value
		seqAccount.Balance += gas

		// 更新nonce - 使用最高的nonce值
		if tx.Nonce > fromAccount.Nonce {
			fromAccount.Nonce = tx.Nonce
		}
	}

	// 第四步：构建最终的更新映射
	updates := make(map[string]Account)
	for addr, account := range accountStates {
		updates[addr] = *account
	}

	// 批量更新账户状态
	for addr, account := range updates {
		if err := n.dacMgr.UpdateAccountState(account); err != nil {
			return fmt.Errorf("更新账户 %s 状态失败: %w", addr, err)
		}
	}

	return nil
}

// 处理区块提交响应
func (n *Layer2Node) handleDACBlockCommitResponse(msg *pubsub.Message) {
	// 只有排序器处理
	if !n.isSequencer {
		return
	}

	var resp DACBlockCommitResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		fmt.Printf("解析区块提交响应失败: %v\n", err)
		return
	}

	// 验证DAC成员签名
	members := n.dacMgr.GetState().Members
	member, exists := members[resp.DACMemberID]
	if !exists || member.PublicKey == nil {
		fmt.Printf("找不到DAC成员公钥信息\n")
		return
	}

	// 验证签名
	respData, err := json.Marshal(struct {
		RequestID      string `json:"request_id"`
		BlockHeight    uint64 `json:"block_height"`
		NewAccountRoot string `json:"new_account_root"`
		NewTxRoot      string `json:"new_tx_root"`
		DACMemberID    string `json:"dac_member_id"`
	}{
		RequestID:      resp.RequestID,
		BlockHeight:    resp.BlockHeight,
		NewAccountRoot: resp.NewAccountRoot,
		NewTxRoot:      resp.NewTxRoot,
		DACMemberID:    resp.DACMemberID,
	})

	if err != nil {
		fmt.Printf("序列化验证数据失败: %v\n", err)
		return
	}

	valid, err := member.PublicKey.Verify(respData, resp.Signature)
	if err != nil || !valid {
		fmt.Printf("DAC成员签名验证失败: %v\n", err)
		return
	}

	// 记录响应
	respValue, _ := blockCommitResponses.LoadOrStore(resp.RequestID, &DACBlockCommitResp{
		blockHeight:    resp.BlockHeight,
		newAccountRoot: resp.NewAccountRoot,
		newTxRoot:      resp.NewTxRoot,
		proofs:         make(map[string][][]byte),
		responses:      make(map[string]bool),
		signatures:     make(map[string][]byte),
	})

	responseData := respValue.(*DACBlockCommitResp)

	// 获取写锁更新数据
	responseData.mu.Lock()

	// 记录响应和签名
	responseData.responses[resp.DACMemberID] = true
	responseData.signatures[resp.DACMemberID] = resp.Signature
	responseData.proofs[resp.DACMemberID] = resp.Proof

	// 如果是首次记录，设置状态根
	responseData.newAccountRoot = resp.NewAccountRoot
	responseData.newTxRoot = resp.NewTxRoot

	// 释放锁
	responseData.mu.Unlock()

	fmt.Printf("收到DAC成员 %s 的区块提交响应: 新账户根=%s, 新交易根=%s\n",
		resp.DACMemberID, resp.NewAccountRoot, resp.NewTxRoot)
}

// 验证区块提交响应的一致性
func (n *Layer2Node) verifyBlockCommitConsistency(resp *DACBlockCommitResp) bool {
	// 注意：调用此函数前需要持有resp的读锁
	// 简单实现，确保所有响应的状态根一致
	// 实际应该对多个响应进行投票
	return true
}

// 提交证明到Layer 1
func (n *Layer2Node) SubmitProofToLayerOne(requestID string, resp *DACBlockCommitResp) error {
	// 获取读锁
	resp.mu.RLock()

	// 获取区块信息
	blockHeight := resp.blockHeight
	newAccountRoot := resp.newAccountRoot
	newTxRoot := resp.newTxRoot

	// 准备证明
	var proofs [][]byte
	for _, proof := range resp.proofs {
		proofs = proof
		break
	}

	resp.mu.RUnlock()

	// 获取区块Hash
	blockHash := ""
	blockInterface, exists := n.blockCache.Load(blockHeight)
	if !exists {
		return fmt.Errorf("区块未找到: 高度=%d", blockHeight)
	}

	block, ok := blockInterface.(Block)
	if !ok {
		return fmt.Errorf("区块类型转换失败")
	}

	blockHash = block.Hash

	// 创建提交请求
	req := DACProofSubmitRequest{
		Type:        DACProofSubmitRequestType,
		RequestID:   uuid.New().String(),
		BlockHeight: blockHeight,
		BlockHash:   blockHash,
		StateRoot:   n.stateDB.GetStateRoot(),
		AccountRoot: newAccountRoot,
		TxRoot:      newTxRoot,
		Proofs:      proofs,
	}

	// 签名请求
	reqData, err := json.Marshal(struct {
		BlockHeight uint64 `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		StateRoot   string `json:"state_root"`
		AccountRoot string `json:"account_root"`
		TxRoot      string `json:"tx_root"`
	}{
		BlockHeight: req.BlockHeight,
		BlockHash:   req.BlockHash,
		StateRoot:   req.StateRoot,
		AccountRoot: req.AccountRoot,
		TxRoot:      req.TxRoot,
	})

	if err != nil {
		return fmt.Errorf("序列化签名数据失败: %w", err)
	}

	signature, err := n.privateKey.Sign(reqData)
	if err != nil {
		return fmt.Errorf("签名请求失败: %w", err)
	}
	req.Signature = signature

	// 广播请求
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	err = n.topic.dacTopic.Publish(n.ctx, reqBytes)
	if err != nil {
		return fmt.Errorf("广播请求失败: %w", err)
	}

	logger.Infof("已发送证明提交请求，区块高度: %d, 新账户根: %s, 新交易根: %s",
		req.BlockHeight, req.AccountRoot, req.TxRoot)

	return nil
}
