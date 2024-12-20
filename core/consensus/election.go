package consensus

import (
	"FPoS/core/ethereum"
	"FPoS/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"sync"
	"time"
)

type ElectionManager struct {
	mu sync.RWMutex

	state      *ElectionState
	config     *ConsensusConfig
	Validators map[string]*Validator

	// 一层相关
	ethClient    *ethereum.EthereumClient // 用于获取L1随机数
	randomNumber uint64                   // 当前轮次的随机数

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
			NextRotationTime: time.Now().Add(config.RotationInterval),
		},
		config:     config,
		blockCh:    make(chan uint64, 1),
		Validators: make(map[string]*Validator),
		rotationCh: make(chan string, 1),
		ctx:        ctx,
		cancel:     cancel,
	}
}
func (em *ElectionManager) SetEth(eth *ethereum.EthereumClient) {
	em.ethClient = eth
}
func (em *ElectionManager) GetEth() *ethereum.EthereumClient {
	return em.ethClient
}

func (em *ElectionManager) GetState() ElectionState {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return *em.state
}

func (em *ElectionManager) SetState(state *ElectionState) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.state = state
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

		case <-time.After(time.Second): // 定期检查轮换时间
			em.mu.RLock()
			nextRotation := em.state.NextRotationTime
			em.mu.RUnlock()

			if time.Now().After(nextRotation) {
				fmt.Printf("Rotation timer expired, rotating sequencer\n")
				em.RotateSequencer()
			}
		}
	}
}

//func (em *ElectionManager) RotateSequencer() {
//	em.mu.Lock()
//	defer em.mu.Unlock()
//
//	activeValidators := em.getActiveValidators()
//	if len(activeValidators) == 0 {
//		return
//	}
//
//	// 获取L1链上的随机数
//	fullRandom, err := em.ethClient.GetFullRandomNumber()
//	if err != nil {
//		fmt.Printf("Failed to get random number from L1: %v\n", err)
//		return
//	}
//
//	// 计算每个验证者的综合权重
//	type validatorWeight struct {
//		validator *Validator
//		weight    *big.Int
//	}
//
//	weights := make([]validatorWeight, len(activeValidators))
//	totalWeight := new(big.Int)
//
//	for i, v := range activeValidators {
//		// 1. 基础权重 = 质押金额
//		baseWeight := new(big.Int).SetUint64(v.StakeAmount)
//
//		// 2. 计算公钥派生的权重因子
//		pubKeyBytes, err := v.PublicKey.Raw()
//		if err != nil {
//			continue
//		}
//		pubKeyHash := ethCrypto.Keccak256(pubKeyBytes)
//		pubKeyFactor := new(big.Int).SetBytes(pubKeyHash)
//
//		// 3. 累积权重分数
//		accumWeight := new(big.Int).SetUint64(v.WeightScore)
//		// 将累积权重转换为1.0到2.0之间的乘数
//		// weightMultiplier = 1.0 + (accumWeight / 1000)
//		weightMultiplier := new(big.Int).Add(
//			big.NewInt(10), // 基础值1000
//			accumWeight,    // 累积权重
//		)
//		// 4. 使用随机数和公钥计算动态系数
//		dynamicFactor := new(big.Int).Xor(fullRandom, pubKeyFactor)
//
//		// 5. 综合权重计算
//		// weight = baseWeight * weightMultiplier * dynamicFactor / 1000
//		weight := new(big.Int).Mul(baseWeight, weightMultiplier) // 应用累积权重
//		weight.Mul(weight, dynamicFactor)                        // 应用动态因子
//		weight.Div(weight, big.NewInt(1000))                     // 归一化累积权重
//
//		weights[i] = validatorWeight{
//			validator: v,
//			weight:    weight,
//		}
//		totalWeight.Add(totalWeight, weight)
//
//		fmt.Printf("Validator weight calculation: %s\n"+
//			"  Base Weight: %d\n"+
//			"  Accum Score: %d\n"+
//			"  Final Weight: %s\n",
//			v.Address,
//			v.StakeAmount,
//			v.WeightScore,
//			weight.String(),
//		)
//	}
//
//	// 使用随机数选择验证者
//	selection := new(big.Int).Mod(fullRandom, totalWeight)
//	var selectedValidator *Validator
//	accum := new(big.Int)
//
//	for _, vw := range weights {
//		accum.Add(accum, vw.weight)
//		if selection.Cmp(accum) < 0 {
//			selectedValidator = vw.validator
//			break
//		}
//	}
//
//	if selectedValidator == nil {
//		selectedValidator = activeValidators[0]
//	}
//
//	// 更新权重分数
//	for _, v := range activeValidators {
//		if v.Address == selectedValidator.Address {
//			// 被选中的验证者重置权重分数
//			v.WeightScore = 0
//		} else {
//			// 未被选中的验证者增加权重分数
//			v.WeightScore += 100 // 每轮增加100点权重
//		}
//	}
//
//	// 确保不会连续选择同一个排序器
//	if selectedValidator.Address == em.state.CurrentSequencer && len(activeValidators) > 1 {
//		// 使用不同的随机数部分重新选择
//		altRandom := new(big.Int).Rsh(fullRandom, 128)
//		nextBestWeight := new(big.Int).SetInt64(0)
//		var nextBestValidator *Validator
//
//		for _, vw := range weights {
//			if vw.validator.Address != em.state.CurrentSequencer {
//				weightWithAlt := new(big.Int).Xor(vw.weight, altRandom)
//				if nextBestValidator == nil || weightWithAlt.Cmp(nextBestWeight) > 0 {
//					nextBestValidator = vw.validator
//					nextBestWeight = weightWithAlt
//				}
//			}
//		}
//		selectedValidator = nextBestValidator
//	}
//
//	// 更新状态
//	now := time.Now()
//	em.state.CurrentSequencer = selectedValidator.Address
//	em.state.CurrentTerm++
//	em.state.LastRotation = now
//	em.state.LastRandomNumber = fullRandom
//	em.state.NextRotationTime = now.Add(em.state.RotationInterval) // 设置下次轮换时间
//
//	fmt.Printf("New sequencer selected: %s (term=%d)\n"+
//		"Random=%s\n"+
//		"Weight=%d\n"+
//		"AccumScore=%d\n"+
//		"Next rotation at: %s\n",
//		selectedValidator.Address,
//		em.state.CurrentTerm,
//		fullRandom.String(),
//		selectedValidator.StakeAmount,
//		selectedValidator.WeightScore,
//		em.state.NextRotationTime.Format(time.RFC3339),
//	)
//
//	select {
//	case em.rotationCh <- selectedValidator.Address:
//	default:
//	}
//}

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

//func (em *ElectionManager) RegisterValidator(pubKey crypto.PubKey, stake uint64) (*Validator, error) {
//	em.mu.Lock()
//	defer em.mu.Unlock()
//
//	// 检查质押金额
//	if stake < em.config.MinStakeAmount {
//		return nil, fmt.Errorf("insufficient stake amount: required %d, got %d",
//			em.config.MinStakeAmount, stake)
//	}
//	addr, _ := types.PublicKeyToAddress(pubKey)
//
//	// 创建验证者记录
//	validator := &Validator{
//		Address:     addr,
//		PublicKey:   pubKey,
//		Status:      Active,
//		StakeAmount: stake,
//		JoinTime:    time.Now(),
//	}
//
//	// 注册验证者
//	em.Validators[addr] = validator
//	return validator, nil
//}

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

	// 如果只有 PublicKeyBytes，需要恢复 PublicKey
	if msg.Validator.PublicKey == nil && len(msg.Validator.PublicKeyBytes) > 0 {
		pubKey, err := crypto.UnmarshalPublicKey(msg.Validator.PublicKeyBytes)
		if err != nil {
			return fmt.Errorf("failed to unmarshal public key: %v", err)
		}
		msg.Validator.PublicKey = pubKey
	}

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
