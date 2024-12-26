package consensus

import (
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"math/big"
	"time"
)

// ValidatorConfig 验证者配置
type ValidatorConfig struct {
	MinStake    uint64 // 最小质押金额
	StakeAmount uint64 // 当前质押金额
}

// ValidatorStatus 验证者状态
type ValidatorStatus int

const (
	Inactive ValidatorStatus = iota
	Active
	Pending
)

// StakeBucket 质押桶
type StakeBucket struct {
	ID            uint64   `json:"id"`
	StakeAmount   uint64   `json:"stake_amount"`
	MappedValue   *big.Int `json:"mapped_value"`   // 固定的映射值 x_i
	CurrentWeight *big.Int `json:"current_weight"` // 当前轮次的权重 w_i
}

// Validator 验证者信息
type Validator struct {
	Address        string                  `json:"address"`
	PublicKey      crypto.PubKey           `json:"-"`          // 不直接序列化
	PublicKeyBytes []byte                  `json:"public_key"` // 用于序列化的字段
	Status         ValidatorStatus         `json:"status"`
	StakeAmount    uint64                  `json:"stake_amount"`
	JoinTime       time.Time               `json:"join_time"`
	Buckets        map[uint64]*StakeBucket `json:"buckets"` // 桶ID到质押桶的映射
	BlocksProduced uint64                  `json:"blocks_produced"`
	LastBlockTime  time.Time               `json:"last_block_time"`
	MissedBlocks   int                     `json:"missed_blocks"`
	WeightScore    uint64                  `json:"weight_score"`
}

// ConsensusConfig 共识配置
type ConsensusConfig struct {
	MinStakeAmount   uint64        // 最小质押要求
	RotationInterval time.Duration // 轮换间隔
	ValidatorQuorum  int           // 最小验证者数量
}

// ElectionState 选举状态
type ElectionState struct {
	CurrentSequencer string                // 当前排序器地址
	CurrentTerm      uint64                // 当前任期
	LastRotation     time.Time             // 上次轮换时间
	Validators       map[string]*Validator // 验证者列表
	RotationInterval time.Duration         // 轮换间隔
	BlockTimeout     time.Duration         // 区块生成超时时间
	MaxMissedBlocks  int                   // 允许的最大错过区块数
	LastRandomNumber *big.Int              // 上一次使用的随机数
	NextRotationTime time.Time             `json:"next_rotation_time"` // 新增：下次轮换时间
	CurrentProposers []string              // 新增：当前的提案者列表
}

// ValidatorMessage 验证者消息类型
type ValidatorMessageType int

const (
	ValidatorJoin ValidatorMessageType = iota
	ValidatorLeave
	ValidatorUpdate
)

// ValidatorMessage 验证者状态同步消息
type ValidatorMessage struct {
	Type      ValidatorMessageType `json:"type"`
	Validator Validator            `json:"validator"`
	Signature []byte               `json:"signature"`
}

// MarshalJSON 自定义JSON序列化方法
func (v *Validator) MarshalJSON() ([]byte, error) {
	type ValidatorAlias Validator

	// 获取公钥字节
	var pubKeyBytes []byte
	if v.PublicKey != nil {
		// 使用 libp2p 的 MarshalPublicKey 方法
		var err error
		pubKeyBytes, err = crypto.MarshalPublicKey(v.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal public key: %v", err)
		}
	}

	return json.Marshal(&struct {
		*ValidatorAlias
		PublicKeyBytes []byte `json:"public_key"`
	}{
		ValidatorAlias: (*ValidatorAlias)(v),
		PublicKeyBytes: pubKeyBytes,
	})
}

// UnmarshalJSON 自定义JSON反序列化方法
func (v *Validator) UnmarshalJSON(data []byte) error {
	type ValidatorAlias Validator
	aux := &struct {
		*ValidatorAlias
		PublicKeyBytes []byte `json:"public_key"`
	}{
		ValidatorAlias: (*ValidatorAlias)(v),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 从字节恢复公钥
	if len(aux.PublicKeyBytes) > 0 {
		pubKey, err := crypto.UnmarshalPublicKey(aux.PublicKeyBytes)
		if err != nil {
			return fmt.Errorf("failed to unmarshal public key: %v", err)
		}
		v.PublicKey = pubKey
		v.PublicKeyBytes = aux.PublicKeyBytes // 保存原始字节
	}

	return nil
}
