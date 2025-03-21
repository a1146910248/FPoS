package dac

// StateProvider 定义了获取状态数据的接口
// 这样dac包就不需要直接依赖p2p包
type StateProvider interface {
	// GetBalance 获取账户余额
	GetBalance(address string) uint64

	// GetNonce 获取账户nonce
	GetNonce(address string) uint64
}
