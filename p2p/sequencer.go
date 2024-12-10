package p2p

import (
	"FPoS/core/consensus"
	"FPoS/types"
	"fmt"
	"sync"
	"time"
)

const _MaxBlockGasLimit_ = 2100

type Sequencer struct {
	node             *Layer2Node
	blockHeight      uint64
	mu               sync.Mutex
	maxBlockGasLimit uint64

	electionMgr *consensus.ElectionManager
}

func NewSequencer(node *Layer2Node) *Sequencer {
	//node.isSequencer = true
	seq := &Sequencer{
		node:        node,
		blockHeight: 0,
		//maxBlockGasLimit: 30_000_000, // 区块 gas 上限为 30,000,000
		maxBlockGasLimit: _MaxBlockGasLimit_, // 区块 gas 上限为 30,000,000
	}
	//node.sequencer = seq
	return seq
}

func (s *Sequencer) Start() {
	// 监听排序器轮换
	go s.watchRotation()

	// 原有的区块生产逻辑
	go s.blockProducingLoop()
}

func (s *Sequencer) watchRotation() {
	rotationCh := s.node.electionMgr.GetRotationChannel()
	for range rotationCh {
		s.mu.Lock()
		// 当前节点被选上
		if s.node.IsCurrentSequencer() {
			s.node.isSequencer = true
			s.node.sequencer = s
			addr, _ := types.PublicKeyToAddress(s.node.publicKey)
			fmt.Printf("Node %s became the new sequencer\n", addr)
		} else {
			// 没被选上，将排序器清空
			s.node.isSequencer = false
			s.node.sequencer = nil
		}
		s.mu.Unlock()
	}
}

func (s *Sequencer) blockProducingLoop() {
	ticker := time.NewTicker(1 * time.Second) // 每秒检查一次是否需要打包
	defer ticker.Stop()

	for {
		select {
		case <-s.node.ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			isCurrentSeq := s.node.isSequencer
			s.mu.Unlock()

			if isCurrentSeq && s.shouldProduceBlock() {
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

	proPub, _ := types.PublicKeyToAddress(s.node.publicKey)
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
