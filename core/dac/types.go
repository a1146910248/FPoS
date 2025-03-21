package dac

import (
	"FPoS/core/consensus"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// DAC 成员状态
type DACMemberStatus int

const (
	Inactive DACMemberStatus = iota
	Active
	Slashed
)

// DAC 成员
type DACMember struct {
	Address     string                            // 成员地址
	PublicKey   crypto.PubKey                     // 公钥
	Status      DACMemberStatus                   // 状态
	StakeAmount uint64                            // 质押数量
	JoinTime    time.Time                         // 加入时间
	Buckets     map[uint64]*consensus.StakeBucket // 质押桶

	// 数据可用性验证指标
	DataProvided         uint64    // 提供的数据量
	SuccessfulChallenges uint64    // 成功应对的挑战次数
	FailedChallenges     uint64    // 失败的挑战次数
	LastActiveTime       time.Time // 最后活跃时间
}

// DAC 状态
type DACState struct {
	CurrentMembers   []string              // 当前活跃 DAC 成员地址
	CurrentTerm      uint64                // 当前任期
	LastRotation     time.Time             // 上次轮换时间
	Members          map[string]*DACMember // 所有成员列表
	RotationInterval time.Duration         // 轮换间隔
	NextRotationTime time.Time             // 下次轮换时间
}
