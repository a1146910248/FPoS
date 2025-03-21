package dac

import (
	"FPoS/core/consensus"
	"FPoS/core/merkle"
	"FPoS/types"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// DACManager DAC 管理器
type DACManager struct {
	mu          sync.RWMutex
	state       *DACState
	config      *consensus.ConsensusConfig
	electionMgr *consensus.ElectionManager

	// 状态树
	accountTree *merkle.SparseMerkleTree // 账户状态树
	txTree      *merkle.SparseMerkleTree // 交易树

	// 通道
	blockCh    chan uint64 // 用于接收新区块高度
	rotationCh chan []string
	ctx        context.Context
	cancel     context.CancelFunc

	// 事件回调
	onStateChange func(members []string, totalMembers uint64, activeMembers uint64)
}

// NewDACManager 创建新的 DAC 管理器
func NewDACManager(ctx context.Context, config *consensus.ConsensusConfig) *DACManager {
	ctx, cancel := context.WithCancel(ctx)

	return &DACManager{
		state: &DACState{
			Members:          make(map[string]*DACMember),
			RotationInterval: config.RotationInterval,
			NextRotationTime: time.Now().Add(config.RotationInterval),
		},
		config:      config,
		blockCh:     make(chan uint64, 10), // 初始化区块高度通道，缓冲大小为10
		rotationCh:  make(chan []string, 1),
		ctx:         ctx,
		cancel:      cancel,
		accountTree: merkle.NewSparseMerkleTree(merkle.AccountTreeDepth),
		txTree:      merkle.NewSparseMerkleTree(256), // 交易哈希是 256 位
	}
}

// SetElectionManager 设置选举管理器
func (dm *DACManager) SetElectionManager(em *consensus.ElectionManager) {
	dm.electionMgr = em
}

// RegisterMember 注册 DAC 成员
func (dm *DACManager) RegisterMember(pubKey crypto.PubKey, stake uint64) (*DACMember, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 检查质押金额
	if stake < dm.config.MinStakeAmount {
		return nil, fmt.Errorf("insufficient stake amount: required %d, got %d",
			dm.config.MinStakeAmount, stake)
	}

	addr, _ := types.PublicKeyToAddress(pubKey)

	// 创建 DAC 成员
	member := &DACMember{
		Address:     addr,
		PublicKey:   pubKey,
		Status:      Active,
		StakeAmount: stake,
		JoinTime:    time.Now(),
		Buckets:     make(map[uint64]*consensus.StakeBucket),
	}

	// 创建质押桶，与选举管理器类似
	bucketCount := (stake + consensus.MaxBucketStake - 1) / consensus.MaxBucketStake

	for i := uint64(0); i < bucketCount; i++ {
		bucketStake := stake
		if bucketStake > consensus.MaxBucketStake {
			bucketStake = consensus.MaxBucketStake
		}
		stake -= bucketStake

		mappedValue := consensus.CalculateMappedValue(bucketStake)
		member.Buckets[i] = &consensus.StakeBucket{
			ID:            i,
			StakeAmount:   bucketStake,
			MappedValue:   mappedValue,
			CurrentWeight: mappedValue,
		}
	}

	// 注册 DAC 成员
	dm.state.Members[addr] = member
	return member, nil
}

// Start 启动 DAC 管理器
func (dm *DACManager) Start() {
	go dm.rotationLoop()
}

// 轮换循环
func (dm *DACManager) rotationLoop() {
	var lastBlockHeight uint64
	for {
		select {
		case <-dm.ctx.Done():
			return

		case height := <-dm.blockCh:
			if height <= lastBlockHeight {
				continue
			}
			lastBlockHeight = height
			dm.RotateMembers()

		case <-time.After(time.Second): // 定期检查轮换时间
			dm.mu.RLock()
			nextRotation := dm.state.NextRotationTime
			dm.mu.RUnlock()

			if time.Now().After(nextRotation) {
				fmt.Printf("Rotation timer expired, rotating dac\n")
				dm.RotateMembers()
			}
		}
	}
}

// NotifyNewBlock 通知新区块生成
func (dm *DACManager) NotifyNewBlock(height uint64) {
	// 检查区块高度是否有效
	if height == 0 {
		return // 忽略无效的区块高度
	}

	// 向区块通道发送消息，使用非阻塞方式
	select {
	case dm.blockCh <- height:
		// 成功发送
	case <-dm.ctx.Done():
		// 上下文已取消
		return
	default:
		// 通道已满，打印警告但不阻塞
		fmt.Printf("Warning: DAC block notification channel is full, height %d dropped\n", height)
	}
}

// UpdateAccountState 更新账户状态
func (dm *DACManager) UpdateAccountState(account types.Account) error {
	// 序列化账户数据
	accountData, err := json.Marshal(account)
	if err != nil {
		return err
	}

	// 更新账户树
	key := []byte(account.Address)
	return dm.accountTree.Update(key, accountData)
}

// GetAccountState 获取账户状态
func (dm *DACManager) GetAccountState(address string) (*types.Account, error) {
	// 从状态树获取账户数据
	data, err := dm.accountTree.Get([]byte(address))
	if err != nil {
		return nil, err
	}

	var account types.Account
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, err
	}

	return &account, nil
}

// StoreTransaction 存储交易
func (dm *DACManager) StoreTransaction(tx types.Transaction) error {
	// 序列化交易数据
	txData, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	// 更新交易树
	key := []byte(tx.Hash)
	return dm.txTree.Update(key, txData)
}

// GetTransaction 获取交易
func (dm *DACManager) GetTransaction(hash string) (*types.Transaction, error) {
	// 从交易树获取交易数据
	data, err := dm.txTree.Get([]byte(hash))
	if err != nil {
		return nil, err
	}

	var tx types.Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, err
	}

	return &tx, nil
}

// GetAccountProof 获取账户状态证明
func (dm *DACManager) GetAccountProof(address string) ([][]byte, error) {
	return dm.accountTree.GenerateProof([]byte(address))
}

// GetTransactionProof 获取交易证明
func (dm *DACManager) GetTransactionProof(hash string) ([][]byte, error) {
	return dm.txTree.GenerateProof([]byte(hash))
}

// SetStateChangeCallback 设置状态变更回调函数
func (dm *DACManager) SetStateChangeCallback(callback func(members []string, totalMembers uint64, activeMembers uint64)) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.onStateChange = callback
}

// BatchUpdateAccounts 批量更新账户状态
func (dm *DACManager) BatchUpdateAccounts(accounts []types.Account) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for _, account := range accounts {
		if err := dm.UpdateAccountState(account); err != nil {
			return err
		}
	}

	return nil
}

// 获取当前DAC状态
func (dm *DACManager) GetState() DACState {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// 创建并返回状态的深拷贝，避免外部修改
	stateCopy := &DACState{
		CurrentMembers:   make([]string, len(dm.state.CurrentMembers)),
		CurrentTerm:      dm.state.CurrentTerm,
		LastRotation:     dm.state.LastRotation,
		RotationInterval: dm.state.RotationInterval,
		NextRotationTime: dm.state.NextRotationTime,
		Members:          make(map[string]*DACMember),
	}

	// 复制CurrentMembers切片
	copy(stateCopy.CurrentMembers, dm.state.CurrentMembers)

	// 复制Members映射
	for k, v := range dm.state.Members {
		memberCopy := *v // 创建成员的副本
		stateCopy.Members[k] = &memberCopy
	}

	return *stateCopy
}

// 同步世界状态到DAC状态
func (dm *DACManager) SyncWorldState(stateProvider StateProvider, block types.Block) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 处理区块中的所有交易
	for _, tx := range block.Transactions {
		// 存储交易到交易树
		if err := dm.StoreTransaction(tx); err != nil {
			return err
		}

		// 从StateProvider获取发送方的最新状态
		senderAddr := tx.From
		senderAccount := types.Account{
			Address: senderAddr,
			Balance: stateProvider.GetBalance(senderAddr),
			Nonce:   stateProvider.GetNonce(senderAddr),
		}

		// 更新DAC中的发送方账户
		if err := dm.UpdateAccountState(senderAccount); err != nil {
			return fmt.Errorf("更新DAC发送方账户失败: %w", err)
		}

		// 从StateProvider获取接收方的最新状态
		receiverAddr := tx.To
		receiverAccount := types.Account{
			Address: receiverAddr,
			Balance: stateProvider.GetBalance(receiverAddr),
			Nonce:   stateProvider.GetNonce(receiverAddr),
		}

		// 更新DAC中的接收方账户
		if err := dm.UpdateAccountState(receiverAccount); err != nil {
			return fmt.Errorf("更新DAC接收方账户失败: %w", err)
		}
	}

	return nil
}

// OnBlockProduced 在区块处理时调用，用于更新DAC成员状态和触发轮转
func (dm *DACManager) OnBlockProduced(height uint64) {
	dm.mu.RLock()
	currentMembers := dm.state.CurrentMembers
	dm.mu.RUnlock()

	// 记录区块生成情况
	dm.NotifyNewBlock(height)

	// 更新所有当前DAC成员的统计信息
	dm.mu.Lock()
	for _, memberAddr := range currentMembers {
		if member, exists := dm.state.Members[memberAddr]; exists {
			// 增加参与的区块数
			member.DataProvided++
			member.LastActiveTime = time.Now()
		}
	}
	dm.mu.Unlock()

	// 验证是否需要触发DAC数据可用性挑战
	go dm.checkDataAvailability(height)
}

// checkDataAvailability 检查数据可用性，可能触发挑战
func (dm *DACManager) checkDataAvailability(height uint64) {
	// 这里可以实现数据可用性的检查逻辑
	// 例如，随机选择一些区块数据，要求DAC成员提供证明

	// TODO: 实现具体的数据可用性挑战逻辑
}

// 添加成员
func (dm *DACManager) AddMember(member *DACMember) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 检查成员是否已存在
	if _, exists := dm.state.Members[member.Address]; exists {
		return false
	}

	// 添加成员
	dm.state.Members[member.Address] = member
	return true
}

// 移除成员
func (dm *DACManager) RemoveMember(address string) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 检查成员是否存在
	if _, exists := dm.state.Members[address]; !exists {
		return false
	}

	// 移除成员
	delete(dm.state.Members, address)

	// 如果是当前活跃成员，也需要从活跃列表中移除
	for i, addr := range dm.state.CurrentMembers {
		if addr == address {
			dm.state.CurrentMembers = append(dm.state.CurrentMembers[:i], dm.state.CurrentMembers[i+1:]...)
			break
		}
	}

	return true
}

// 更新成员
func (dm *DACManager) UpdateMember(member *DACMember) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 检查成员是否存在
	if _, exists := dm.state.Members[member.Address]; !exists {
		return false
	}

	// 更新成员
	dm.state.Members[member.Address] = member
	return true
}

// 设置状态
func (dm *DACManager) SetState(state *DACState) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 更新状态
	dm.state = state
}
