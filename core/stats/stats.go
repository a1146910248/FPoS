package stats

import (
	"FPoS/core/consensus"
	"FPoS/core/ethereum"
	"FPoS/p2p"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/viper"
)

type Stats struct {
	mu sync.RWMutex

	// TPS相关
	currentTPS    float64
	peakTPS       float64
	tpsUpdateTime time.Time
	txCount       uint64

	// 区块相关
	blockCount  uint64
	blockHeight uint64

	// 用户相关
	activeUsers map[string]time.Time

	// 组件引用
	node        *p2p.Layer2Node
	ethClient   *ethereum.EthereumClient
	electionMgr *consensus.ElectionManager
}

var globalStats *Stats
var once sync.Once

func GetStats() *Stats {
	once.Do(func() {
		globalStats = &Stats{
			activeUsers:   make(map[string]time.Time),
			tpsUpdateTime: time.Now(),
		}
	})
	return globalStats
}

// InitStats 初始化统计模块
func InitStats(node *p2p.Layer2Node, ethClient *ethereum.EthereumClient, electionMgr *consensus.ElectionManager) {
	stats := GetStats()
	stats.node = node
	stats.ethClient = ethClient
	stats.electionMgr = electionMgr
}

// UpdateActiveUser 更新活跃用户
func (s *Stats) UpdateActiveUser(address string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeUsers[address] = time.Now()
}

// UpdateTxCount 更新交易计数
func (s *Stats) UpdateTxCount(count uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	duration := now.Sub(s.tpsUpdateTime).Seconds()
	if duration > 0 {
		currentTPS := float64(count-s.txCount) / duration
		s.currentTPS = currentTPS
		if currentTPS > s.peakTPS {
			s.peakTPS = currentTPS
		}
	}

	s.txCount = count
	s.tpsUpdateTime = now
}

// UpdateBlockHeight 更新区块高度
func (s *Stats) UpdateBlockHeight(height uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blockHeight = height
	s.blockCount++
}

// GetCurrentTPS 获取当前TPS
func (s *Stats) GetCurrentTPS() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentTPS
}

// GetPeakTPS 获取峰值TPS
func (s *Stats) GetPeakTPS() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.peakTPS
}

// GetTotalTransactions 获取总交易数
func (s *Stats) GetTotalTransactions() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.txCount
}

// GetCurrentHeight 获取当前区块高度
func (s *Stats) GetCurrentHeight() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.blockHeight
}

// GetActiveUsers 获取活跃用户数
func (s *Stats) GetActiveUsers() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	activeCount := uint64(0)
	for addr, lastActive := range s.activeUsers {
		if now.Sub(lastActive) > time.Hour {
			delete(s.activeUsers, addr)
		} else {
			activeCount++
		}
	}
	return activeCount
}

// GetL1Stats 获取L1链统计信息
func (s *Stats) GetL1Stats() (blockCount uint64, balance *big.Int) {
	if s.ethClient == nil {
		return 0, big.NewInt(0)
	}

	blockCount, err := s.ethClient.GetBlockNumber()
	if err != nil {
		return 0, nil
	}
	address, err := getCommonAddress()
	if err != nil {
		println("获取公钥地址失败")
		return 0, nil
	}
	balance, err = s.ethClient.GetBalance(address)
	if err != nil {
		return 0, nil
	}
	return
}

// GetL2Stats 获取L2链统计信息
func (s *Stats) GetL2Stats() (blocks uint64, tps float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.blockCount, s.currentTPS
}

func getCommonAddress() (common.Address, error) {
	// 示例公钥（以太坊公钥的格式：0x前缀的64个字符）
	pubKeyHex := viper.GetString("eth.PUBKEY")

	// 将公钥从十六进制字符串转换为字节
	pubKey, err := crypto.HexToECDSA(pubKeyHex)
	if err != nil {
		return common.Address{}, err
	}

	// 获取公钥对应的地址
	address := crypto.PubkeyToAddress(pubKey.PublicKey)
	return address, nil
}
