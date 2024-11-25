package p2p

import (
	"FPoS/types"
	"fmt"
	"sync"
	"time"
)

const _MaxBlockGasLimit_ = 21000

type Sequencer struct {
	node             *Layer2Node
	blockHeight      uint64
	mu               sync.Mutex
	maxBlockGasLimit uint64
}

func NewSequencer(node *Layer2Node) *Sequencer {
	node.isSequencer = true
	seq := &Sequencer{
		node:        node,
		blockHeight: 0,
		//maxBlockGasLimit: 30_000_000, // 区块 gas 上限为 30,000,000
		maxBlockGasLimit: _MaxBlockGasLimit_, // 区块 gas 上限为 30,000,000
	}
	node.sequencer = seq
	return seq
}

func (s *Sequencer) Start() {
	go func() {
		// 等待节点初始化完成
		for {
			s.node.mu.RLock()
			initialized := s.node.initialized
			s.node.mu.RUnlock()

			if initialized {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		fmt.Println("Sequencer starting block production after node initialization")
		s.blockProducingLoop()
	}()
}

func (s *Sequencer) blockProducingLoop() {
	ticker := time.NewTicker(1 * time.Second) // 每秒检查一次是否需要打包
	defer ticker.Stop()

	for {
		select {
		case <-s.node.ctx.Done():
			return
		case <-ticker.C:
			if s.shouldProduceBlock() {
				s.produceBlock()
			}
		}
	}
}

// 检查是否应该打包新区块
func (s *Sequencer) shouldProduceBlock() bool {
	var totalGas uint64 = 0

	s.node.txPool.Range(func(_, value interface{}) bool {
		if tx, ok := value.(types.Transaction); ok {
			totalGas += tx.GasUsed
		}
		return totalGas < s.maxBlockGasLimit
	})

	return totalGas >= s.maxBlockGasLimit
}

func (s *Sequencer) produceBlock() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var (
		transactions []types.Transaction
		totalGas     uint64 = 0
	)

	// 收集交易，直到达到gas上限
	s.node.txPool.Range(func(key, value interface{}) bool {
		if tx, ok := value.(types.Transaction); ok {
			if totalGas+tx.GasUsed <= s.maxBlockGasLimit {
				transactions = append(transactions, tx)
				totalGas += tx.GasUsed
				s.node.txPool.Delete(key)
				return true
			}
			return false
		}
		return true
	})

	if len(transactions) == 0 {
		return
	}

	proPub, _ := PublicKeyToAddress(s.node.publicKey)
	// 创建新区块
	block := types.Block{
		Height:       s.blockHeight + 1,
		Timestamp:    time.Now(),
		Transactions: transactions,
		StateRoot:    s.node.stateDB.GetStateRoot(),
		TxRoot:       types.CalculateMerkleRoot(transactions),
		Proposer:     proPub,
		GasUsed:      totalGas, // 记录区块使用的总gas
		GasLimit:     s.maxBlockGasLimit,
	}

	// 计算前一个区块的哈希
	if s.blockHeight > 0 {
		if prevBlock, ok := s.node.blockCache.Load(s.blockHeight); ok {
			if prev, ok := prevBlock.(types.Block); ok {
				block.PreviousHash = prev.Hash
			}
		}
	}

	// 签名区块
	if err := SignBlock(&block, s.node); err != nil {
		fmt.Printf("Failed to sign block: %v\n", err)
		return
	}

	// 计算区块哈希
	blockHash, err := CalculateBlockHash(&block)
	if err != nil {
		fmt.Printf("Failed to calculate block hash: %v\n", err)
		return
	}
	block.Hash = blockHash

	// 广播区块
	if err := s.node.BroadcastBlock(block); err != nil {
		fmt.Printf("Failed to broadcast block: %v\n", err)
		return
	}

	// 更新状态
	s.blockHeight++
	fmt.Printf("New block produced: height=%d, txs=%d, gasUsed=%d\n",
		block.Height, len(block.Transactions), totalGas)
}
