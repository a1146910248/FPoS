package consensus

import (
	"FPoS/types"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"math/big"
	"sort"
	"time"
)

const (
	MaxBucketStake = 1000000 // 每个桶的最大质押金额
)

// calculateMappedValue 计算固定的映射值 x_i = (s_i/MaxBucketStake) * (2^160-1)
func calculateMappedValue(stakeAmount uint64) *big.Int {
	stake := new(big.Int).SetUint64(stakeAmount)
	maxStake := new(big.Int).SetUint64(MaxBucketStake)

	maxValue := new(big.Int).Sub(
		new(big.Int).Lsh(big.NewInt(1), 160), // 2^160
		big.NewInt(1),                        // -1
	)

	mappedValue := new(big.Int).Mul(stake, maxValue)
	mappedValue.Div(mappedValue, maxStake)

	return mappedValue
}

// RegisterValidator 修改注册逻辑
func (em *ElectionManager) RegisterValidator(pubKey crypto.PubKey, stake uint64) (*Validator, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	addr, _ := types.PublicKeyToAddress(pubKey)

	bucketCount := (stake + MaxBucketStake - 1) / MaxBucketStake
	buckets := make(map[uint64]*StakeBucket)

	for i := uint64(0); i < bucketCount; i++ {
		bucketStake := stake
		if bucketStake > MaxBucketStake {
			bucketStake = MaxBucketStake
		}
		stake -= bucketStake

		mappedValue := calculateMappedValue(bucketStake)
		buckets[i] = &StakeBucket{
			ID:            i,
			StakeAmount:   bucketStake,
			MappedValue:   mappedValue,
			CurrentWeight: mappedValue, // 初始权重等于映射值
		}
	}

	validator := &Validator{
		Address:   addr,
		PublicKey: pubKey,
		Status:    Active,
		Buckets:   buckets,
		JoinTime:  time.Now(),
	}

	em.Validators[addr] = validator
	return validator, nil
}

// RotateSequencer 修改选举逻辑以支持多桶
func (em *ElectionManager) RotateSequencer() {
	em.mu.Lock()
	defer em.mu.Unlock()

	activeValidators := em.getActiveValidators()
	if len(activeValidators) == 0 {
		return
	}

	fullRandom, err := em.ethClient.GetFullRandomNumber()
	if err != nil {
		fmt.Printf("Failed to get random number from L1: %v\n", err)
		return
	}

	type bucketWeight struct {
		validator *Validator
		bucket    *StakeBucket
		weight    *big.Int
	}

	var weights []bucketWeight
	// 人数不够共识，则等待下一轮
	if len(activeValidators) < 3 {
		fmt.Printf("validators number to low")
		return
	}
	// 计算所有桶的权重
	for _, v := range activeValidators {
		for _, bucket := range v.Buckets {
			newWeight := em.calculateBucketWeight(v, bucket, fullRandom)
			weights = append(weights, bucketWeight{
				validator: v,
				bucket:    bucket,
				weight:    newWeight,
			})
			bucket.CurrentWeight = newWeight
		}
	}

	// 按权重排序
	sort.Slice(weights, func(i, j int) bool {
		return weights[i].weight.Cmp(weights[j].weight) > 0
	})

	// 选择前三个最高权重的提案者
	selectedProposers := make(map[string]struct{})
	var topProposers []struct {
		validator *Validator
		bucket    *StakeBucket
		weight    *big.Int
	}

	// 确保不会选择同一个验证者多次
	for _, bw := range weights {
		if len(topProposers) >= 3 {
			break
		}
		// 检查这个验证者是否已经被选中
		if _, exists := selectedProposers[bw.validator.Address]; !exists {
			topProposers = append(topProposers, struct {
				validator *Validator
				bucket    *StakeBucket
				weight    *big.Int
			}{
				validator: bw.validator,
				bucket:    bw.bucket,
				weight:    bw.weight,
			})
			selectedProposers[bw.validator.Address] = struct{}{}
		}
	}

	// 重置选中的桶权重
	for _, proposer := range topProposers {
		proposer.bucket.CurrentWeight = big.NewInt(0)
	}

	//// 找出权重最大的桶
	//var maxWeight *big.Int
	//var selectedBucket *StakeBucket
	//var selectedValidator *Validator
	//
	//for _, bw := range weights {
	//	// 更新为新的weight值
	//	bw.bucket.CurrentWeight = bw.weight
	//	if maxWeight == nil || bw.weight.Cmp(maxWeight) > 0 {
	//		maxWeight = bw.weight
	//		selectedBucket = bw.bucket
	//		selectedValidator = bw.validator
	//	}
	//}
	//
	//// 重置选中的桶权重，其他桶权重保持不变
	//selectedBucket.CurrentWeight = big.NewInt(0)

	// 更新状态
	now := time.Now()
	//em.state.CurrentSequencer = selectedValidator.Address
	em.state.CurrentTerm++
	em.state.LastRotation = now
	em.state.LastRandomNumber = fullRandom
	em.state.NextRotationTime = now.Add(em.state.RotationInterval)

	// 更新提案者列表
	em.state.CurrentProposers = make([]string, len(topProposers))
	for i, p := range topProposers {
		em.state.CurrentProposers[i] = p.validator.Address
	}

	// 主排序器为权重最高的提案者
	em.state.CurrentSequencer = topProposers[0].validator.Address

	//fmt.Printf("New sequencer selected: %s (term=%d)\n"+
	//	"Bucket ID: %d\n"+
	//	"Bucket Stake: %d\n"+
	//	"Final Weight: %s\n"+
	//	"Next rotation at: %s\n",
	//	selectedValidator.Address,
	//	em.state.CurrentTerm,
	//	selectedBucket.ID,
	//	selectedBucket.StakeAmount,
	//	maxWeight.String(),
	//	em.state.NextRotationTime.Format(time.RFC3339),
	//)

	// 打印选举结果
	fmt.Printf("New proposers selected for term %d:\n", em.state.CurrentTerm)
	for i, p := range topProposers {
		fmt.Printf("Proposer %d:\n"+
			"  Address: %s\n"+
			"  Bucket ID: %d\n"+
			"  Bucket Stake: %d\n"+
			"  Weight: %s\n",
			i+1,
			p.validator.Address,
			p.bucket.ID,
			p.bucket.StakeAmount,
			p.weight.String(),
		)
	}
	fmt.Printf("Next rotation at: %s\n",
		em.state.NextRotationTime.Format(time.RFC3339))

	select {
	case em.rotationCh <- em.state.CurrentSequencer:
	default:
	}

	// 计算活跃验证者数量
	activeCount := uint64(0)
	for _, v := range em.Validators {
		if v.Status == Active {
			activeCount++
		}
	}

	// 通知状态变更
	if em.onStateChange != nil {
		em.onStateChange(
			em.state.CurrentSequencer,
			em.state.CurrentProposers,
			uint64(len(em.Validators)),
			activeCount,
		)
	}
}

// calculateBucketWeight 计算单个桶的新权重
func (em *ElectionManager) calculateBucketWeight(v *Validator, bucket *StakeBucket, randomNumber *big.Int) *big.Int {
	pubKeyBytes, _ := v.PublicKey.Raw()
	bucketBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bucketBytes, bucket.ID)

	// 组合数据: randomNumber + pubKey + bucketID
	data := append(randomNumber.Bytes(), pubKeyBytes...)
	data = append(data, bucketBytes...)

	hash := sha256.Sum256(data)
	hashBig := new(big.Int).SetBytes(hash[:])

	// 使用固定的映射值计算 mod
	modResult := new(big.Int).Mod(hashBig, bucket.MappedValue)

	// r_i = x_i + [SHA256(r+Pub_i+i) mod x_i]
	r_i := new(big.Int).Add(bucket.MappedValue, modResult)

	// w_(k+1) = w_k + r_i
	newWeight := new(big.Int).Add(bucket.CurrentWeight, r_i)

	return newWeight
}
