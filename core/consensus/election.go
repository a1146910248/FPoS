package consensus

import (
	"FPoS/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"sort"
	"sync"
	"time"
)

type ElectionManager struct {
	mu sync.RWMutex

	state      *ElectionState
	config     *ConsensusConfig
	Validators map[string]*Validator

	// 通道用于通知排序器变更
	blockCh    chan uint64 // 用于接收新区块高度
	rotationCh chan string
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewElectionManager(ctx context.Context, config *ConsensusConfig) *ElectionManager {
	ctx, cancel := context.WithCancel(ctx)
	return &ElectionManager{
		state: &ElectionState{
			Validators:       make(map[string]*Validator),
			RotationInterval: config.RotationInterval,
		},
		config:     config,
		blockCh:    make(chan uint64, 1),
		Validators: make(map[string]*Validator),
		rotationCh: make(chan string, 1),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (em *ElectionManager) Start() {
	go em.rotationLoop()
}

// NotifyNewBlock 通知新区块生成
func (em *ElectionManager) NotifyNewBlock(height uint64) {
	select {
	case em.blockCh <- height:
	default:
		// 如果通道已满，跳过
	}
}

func (em *ElectionManager) rotationLoop() {
	var lastBlockHeight uint64
	rotationTimer := time.NewTimer(em.state.RotationInterval)
	defer rotationTimer.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return

		case height := <-em.blockCh:
			if height <= lastBlockHeight {
				continue
			}
			lastBlockHeight = height
			em.RotateSequencer()
			// 重置轮换计时器
			if !rotationTimer.Stop() {
				select {
				case <-rotationTimer.C:
				default:
				}
			}
			rotationTimer.Reset(em.state.RotationInterval)

		case <-rotationTimer.C:
			// 如果计时器到期，说明当前排序器在规定时间内没有生成区块
			fmt.Printf("Rotation timer expired at block height %d, rotating sequencer\n",
				lastBlockHeight)
			em.RotateSequencer()
			rotationTimer.Reset(em.state.RotationInterval)
		}
	}
}

func (em *ElectionManager) RotateSequencer() {
	em.mu.Lock()
	defer em.mu.Unlock()

	// 获取活跃验证者列表
	activeValidators := em.getActiveValidators()
	if len(activeValidators) == 0 {
		return
	}

	// 按质押金额和加入时间排序
	sort.Slice(activeValidators, func(i, j int) bool {
		if activeValidators[i].StakeAmount == activeValidators[j].StakeAmount {
			return activeValidators[i].JoinTime.Before(activeValidators[j].JoinTime)
		}
		return activeValidators[i].StakeAmount > activeValidators[j].StakeAmount
	})

	// 选择新的排序器
	var newSequencer string
	for _, v := range activeValidators {
		if v.Address != em.state.CurrentSequencer {
			newSequencer = v.Address
			break
		}
	}

	if newSequencer != "" {
		em.state.CurrentSequencer = newSequencer
		em.state.CurrentTerm++
		em.state.LastRotation = time.Now()

		// 通知排序器变更
		select {
		case em.rotationCh <- newSequencer:
		default:
		}
	}
}

// 在区块处理时调用
func (em *ElectionManager) OnBlockProduced(height uint64) {
	em.mu.RLock()
	currentSeq := em.state.CurrentSequencer
	em.mu.RUnlock()

	// 记录区块生成情况
	em.NotifyNewBlock(height)

	// 可以添加额外的统计，如生成区块的性能评分等
	if validator, exists := em.Validators[currentSeq]; exists {
		validator.BlocksProduced++
		validator.LastBlockTime = time.Now()
	}
}

// 获取validator
func (em *ElectionManager) GetValidators() map[string]Validator {
	if len(em.Validators) == 0 {
		return nil
	}
	validatorMap := make(map[string]Validator)
	for addr, validator := range em.Validators {
		validatorMap[addr] = *validator
	}
	return validatorMap
}

func (em *ElectionManager) RegisterValidator(pubKey crypto.PubKey, stake uint64) (*Validator, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// 检查质押金额
	if stake < em.config.MinStakeAmount {
		return nil, fmt.Errorf("insufficient stake amount: required %d, got %d",
			em.config.MinStakeAmount, stake)
	}
	addr, _ := types.PublicKeyToAddress(pubKey)

	// 创建验证者记录
	validator := &Validator{
		Address:     addr,
		PublicKey:   pubKey,
		Status:      Active,
		StakeAmount: stake,
		JoinTime:    time.Now(),
	}

	// 注册验证者
	em.Validators[addr] = validator
	return validator, nil
}

func (em *ElectionManager) getActiveValidators() []*Validator {
	var active []*Validator
	for _, v := range em.Validators {
		if v.Status == Active {
			active = append(active, v)
		}
	}
	return active
}

// GetRotationChannel 获取排序器轮换通知通道
func (em *ElectionManager) GetRotationChannel() <-chan string {
	return em.rotationCh
}

// IsCurrentSequencer 检查指定地址是否为当前排序器
func (em *ElectionManager) IsCurrentSequencer(pub crypto.PubKey) bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	address, _ := types.PublicKeyToAddress(pub)
	return em.state.CurrentSequencer == address
}

func (em *ElectionManager) InitValidators() {
	if em.Validators == nil || len(em.Validators) == 0 {
		em.Validators = make(map[string]*Validator)
	}
}

// HandleValidatorMessage 处理验证者消息
func (em *ElectionManager) HandleValidatorMessage(data []byte) error {
	var msg ValidatorMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	switch msg.Type {
	// 新验证者加入
	case ValidatorJoin:
		em.Validators[msg.Validator.Address] = &msg.Validator
		fmt.Printf("New validator joined: %s\n", msg.Validator.Address)
		// 新验证者加入
	case ValidatorLeave:
		delete(em.Validators, msg.Validator.Address)
		fmt.Printf("validator quited: %s\n", msg.Validator.Address)
	case ValidatorUpdate:
		em.Validators[msg.Validator.Address] = &msg.Validator
		fmt.Printf("validator updated: %s\n", msg.Validator.Address)
	}

	return nil
}
