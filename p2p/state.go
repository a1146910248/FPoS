package p2p

import (
	"FPoS/types"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
)

// AccountState 账户状态
type AccountState struct {
	Balance uint64
	Nonce   uint64
	mu      sync.RWMutex
}

// StateDB 状态数据库
type StateDB struct {
	accounts   map[string]*AccountState // 地址 -> 账户状态
	pendingTxs map[string]*PendingState
	mu         sync.RWMutex
}

// PendingState 待处理交易的结构
type PendingState struct {
	pendingBalance uint64
	pendingNonce   uint64
	mu             sync.RWMutex
}

// StateSync 状态同步请求和响应的消息结构
type StateSync struct {
	RequestID    string                   `json:"requestId"`
	FromHeight   uint64                   `json:"fromHeight"`
	ToHeight     uint64                   `json:"toHeight"`
	Accounts     map[string]*AccountState `json:"accounts"`
	PendingState map[string]*PendingState `json:"pendingState,omitempty"`
	PendingTxs   []types.Transaction      `json:"pendingTxs,omitempty"`
}

// NewStateDB 创建新的状态数据库
func NewStateDB() *StateDB {
	return &StateDB{
		accounts:   make(map[string]*AccountState),
		pendingTxs: make(map[string]*PendingState),
	}
}

// GetAccount 获取账户状态，如果不存在则创建
func (s *StateDB) GetAccount(address string) *AccountState {
	s.mu.RLock()
	account, exists := s.accounts[address]
	s.mu.RUnlock()

	if !exists {
		s.mu.Lock()
		// 双重检查
		if account, exists = s.accounts[address]; !exists {
			account = &AccountState{}
			s.accounts[address] = account
		}
		s.mu.Unlock()
	}

	return account
}

// GetBalance 获取余额
func (s *StateDB) GetBalance(address string) uint64 {
	account := s.GetAccount(address)
	account.mu.RLock()
	defer account.mu.RUnlock()
	return account.Balance
}

// GetNonce 获取nonce
func (s *StateDB) GetNonce(address string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if account, exists := s.accounts[address]; exists {
		return account.Nonce
	}
	return 0
}

// UpdateBalance 更新余额
func (s *StateDB) UpdateBalance(address string, newBalance uint64) {
	account := s.GetAccount(address)
	account.mu.Lock()
	defer account.mu.Unlock()
	account.Balance = newBalance
}

// IncrementNonce 增加nonce
func (s *StateDB) IncrementNonce(address string) {
	account := s.GetAccount(address)
	account.mu.Lock()
	defer account.mu.Unlock()
	account.Nonce++
}

// ValidateTransaction 验证交易
func (s *StateDB) ValidateTransaction(tx *types.Transaction, minGasPrice uint64) error {
	// 检查gas相关参数
	if tx.GasLimit < tx.GasUsed {
		return fmt.Errorf("gas usage overrun: %d > %d", tx.GasUsed, tx.GasLimit)
	}
	if tx.GasPrice < minGasPrice {
		return fmt.Errorf("gas price too low, minimum required: %d", minGasPrice)
	}

	sender := s.GetAccount(tx.From)
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取或创建待处理状态
	pending, exists := s.pendingTxs[tx.From]
	if !exists {
		pending = &PendingState{
			pendingBalance: sender.Balance,
			pendingNonce:   sender.Nonce,
		}
		s.pendingTxs[tx.From] = pending
	}

	pending.mu.Lock()
	defer pending.mu.Unlock()

	// 计算交易需要的总费用
	gasFeeCost := tx.GasUsed * tx.GasPrice
	totalCost := gasFeeCost + tx.Value

	// 检查发送方余额
	sender.mu.RLock()
	defer sender.mu.RUnlock()

	// 检查待处理余额是否足够
	if pending.pendingBalance < totalCost {
		return fmt.Errorf("insufficient balance (including pending): has %d, needs %d",
			pending.pendingBalance, totalCost)
	}

	// 更新待处理状态
	pending.pendingBalance -= totalCost
	pending.pendingNonce++

	return nil
}

// CleanPendingState 当交易被打包进区块时，清理待处理状态
func (s *StateDB) CleanPendingState(address string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pendingTxs, address)
}

// RestorePendingState 当交易从交易池移除时，恢复待处理状态
func (s *StateDB) RestorePendingState(tx *types.Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pending, exists := s.pendingTxs[tx.From]; exists {
		pending.mu.Lock()
		defer pending.mu.Unlock()

		gasFeeCost := tx.GasUsed * tx.GasPrice
		totalCost := gasFeeCost + tx.Value

		pending.pendingBalance += totalCost
		pending.pendingNonce--
	}
}

// ExecuteTransaction 执行交易
func (s *StateDB) ExecuteTransaction(tx *types.Transaction) error {
	sender := s.GetAccount(tx.From)
	receiver := s.GetAccount(tx.To)

	// 锁定发送方和接收方账户
	sender.mu.Lock()
	receiver.mu.Lock()
	defer sender.mu.Unlock()
	defer receiver.mu.Unlock()

	// 计算gas费用
	gasFee := tx.GasUsed * tx.GasPrice
	totalDeduction := tx.Value + gasFee

	// 再次检查余额（因为状态可能在验证后发生改变）
	if sender.Balance < totalDeduction {
		return fmt.Errorf("insufficient balance")
	}

	// 更新账户状态
	sender.Balance -= totalDeduction
	receiver.Balance += tx.Value
	sender.Nonce++

	return nil
}

// GetStateRoot 计算状态根哈希
func (s *StateDB) GetStateRoot() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 将所有账户状态排序并序列化
	data, err := json.Marshal(s.accounts)
	if err != nil {
		return ""
	}

	// 计算哈希
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
