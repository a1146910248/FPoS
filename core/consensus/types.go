package consensus

import (
	"crypto"
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

// Validator 验证者信息
type Validator struct {
	Address        string           // 验证者地址
	PublicKey      crypto.PublicKey // 公钥
	Status         ValidatorStatus  // 状态
	StakeAmount    uint64           // 质押金额
	JoinTime       time.Time        // 加入时间
	BlocksProduced uint64
	LastBlockTime  time.Time
	MissedBlocks   int // 错过的区块数
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
