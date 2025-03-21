package dac

import (
	"FPoS/core/consensus"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"sort"
	"time"
)

const (
	MaxBucketStake = 1000000 // 每个桶的最大质押金额
)

// CalculateMappedValue 计算固定的映射值 x_i = (s_i/MaxBucketStake) * (2^160-1)
func CalculateMappedValue(stakeAmount uint64) *big.Int {
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

// RotateMembers 轮换 DAC 成员，使用与ElectionManager.RotateSequencer相同的算法
func (dm *DACManager) RotateMembers() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 获取活跃DAC成员
	var activeMembers []*DACMember
	for _, member := range dm.state.Members {
		if member.Status == Active && member.StakeAmount >= dm.config.MinStakeAmount {
			activeMembers = append(activeMembers, member)
		}
	}

	// 如果没有活跃成员，直接返回
	if len(activeMembers) == 0 {
		return
	}

	// 从以太坊获取随机数，如果无法获取则使用当前时间戳
	var fullRandom *big.Int
	if dm.electionMgr != nil && dm.electionMgr.GetEth() != nil {
		var err error
		fullRandom, err = dm.electionMgr.GetEth().GetFullRandomNumber()
		if err != nil {
			fmt.Printf("Failed to get random number from L1: %v\n", err)
			// 使用当前时间戳作为备选随机数
			fullRandom = big.NewInt(time.Now().UnixNano())
		}
	} else {
		// 使用当前时间戳作为随机数
		fullRandom = big.NewInt(time.Now().UnixNano())
	}

	type bucketWeight struct {
		member *DACMember
		bucket *consensus.StakeBucket
		weight *big.Int
	}

	var weights []bucketWeight

	// 人数不够最低要求，等待下一轮
	minRequiredMembers := 1 // 设置最低成员要求，与排序器逻辑一致
	if len(activeMembers) < minRequiredMembers {
		fmt.Printf("DAC member number too low, waiting for next rotation\n")
		return
	}

	// 计算所有桶的权重
	for _, m := range activeMembers {
		for _, bucket := range m.Buckets {
			newWeight := dm.calculateBucketWeight(m, bucket, fullRandom)
			weights = append(weights, bucketWeight{
				member: m,
				bucket: bucket,
				weight: newWeight,
			})
			bucket.CurrentWeight = newWeight
		}
	}

	// 按权重排序
	sort.Slice(weights, func(i, j int) bool {
		return weights[i].weight.Cmp(weights[j].weight) > 0
	})

	// 选择前X个最高权重的DAC成员
	selectedMembers := make(map[string]struct{})
	var topMembers []struct {
		member *DACMember
		bucket *consensus.StakeBucket
		weight *big.Int
	}

	// 确定要选择的成员数量
	maxDACMembers := 10 // 可以根据需求调整
	if maxDACMembers > len(activeMembers) {
		maxDACMembers = len(activeMembers)
	}

	// 确保不会选择同一个成员多次
	for _, bw := range weights {
		if len(topMembers) >= maxDACMembers {
			break
		}

		// 检查这个成员是否已经被选中
		if _, exists := selectedMembers[bw.member.Address]; !exists {
			topMembers = append(topMembers, struct {
				member *DACMember
				bucket *consensus.StakeBucket
				weight *big.Int
			}{
				member: bw.member,
				bucket: bw.bucket,
				weight: bw.weight,
			})
			selectedMembers[bw.member.Address] = struct{}{}
		}
	}

	// 重置选中的桶权重
	for _, selected := range topMembers {
		selected.bucket.CurrentWeight = big.NewInt(0)
	}

	// 更新状态
	now := time.Now()
	dm.state.CurrentTerm++
	dm.state.LastRotation = now
	dm.state.NextRotationTime = now.Add(dm.state.RotationInterval)

	// 更新DAC成员列表
	dm.state.CurrentMembers = make([]string, len(topMembers))
	for i, m := range topMembers {
		dm.state.CurrentMembers[i] = m.member.Address
	}

	// 打印选举结果
	fmt.Printf("New DAC members selected for term %d:\n", dm.state.CurrentTerm)
	for i, m := range topMembers {
		fmt.Printf("DAC Member %d:\n"+
			"  Address: %s\n"+
			"  Bucket ID: %d\n"+
			"  Bucket Stake: %d\n"+
			"  Weight: %s\n",
			i+1,
			m.member.Address,
			m.bucket.ID,
			m.bucket.StakeAmount,
			m.weight.String(),
		)
	}
	fmt.Printf("Next DAC rotation at: %s\n",
		dm.state.NextRotationTime.Format(time.RFC3339))

	// 通知轮转通道
	select {
	case dm.rotationCh <- dm.state.CurrentMembers:
	default:
	}

	// 计算活跃DAC成员数量
	activeCount := uint64(0)
	for _, m := range dm.state.Members {
		if m.Status == Active {
			activeCount++
		}
	}

	// 通知状态变更
	if dm.onStateChange != nil {
		dm.onStateChange(
			dm.state.CurrentMembers,
			uint64(len(dm.state.Members)),
			activeCount,
		)
	}
}

// calculateBucketWeight 计算单个桶的新权重
func (dm *DACManager) calculateBucketWeight(m *DACMember, bucket *consensus.StakeBucket, randomNumber *big.Int) *big.Int {
	pubKeyBytes, _ := m.PublicKey.Raw()
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
