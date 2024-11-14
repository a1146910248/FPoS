package p2p

import (
	"FPoS/types"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const _MaxBlockGasLimit_ = 500

type Sequencer struct {
	node             *Layer2Node
	blockHeight      uint64
	mu               sync.Mutex
	maxBlockGasLimit uint64
}

func NewSequencer(node *Layer2Node) *Sequencer {
	node.isSequencer = true
	return &Sequencer{
		node:        node,
		blockHeight: 0,
		//maxBlockGasLimit: 30_000_000, // 区块 gas 上限为 30,000,000
		maxBlockGasLimit: _MaxBlockGasLimit_, // 区块 gas 上限为 30,000,000
	}
}

func (s *Sequencer) Start() {
	go s.blockProducingLoop()
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
	if err := s.signBlock(&block); err != nil {
		fmt.Printf("Failed to sign block: %v\n", err)
		return
	}

	// 计算区块哈希
	blockHash, err := calculateBlockHash(&block)
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

func (s *Sequencer) signBlock(block *types.Block) error {
	// 序列化区块数据
	blockData := struct {
		Height       uint64
		Timestamp    time.Time
		Transactions []types.Transaction
		PreviousHash string
		StateRoot    string
		Proposer     string
		GasUsed      uint64
		GasLimit     uint64
	}{
		Height:       block.Height,
		Timestamp:    block.Timestamp,
		Transactions: block.Transactions,
		PreviousHash: block.PreviousHash,
		StateRoot:    block.StateRoot,
		Proposer:     block.Proposer,
		GasUsed:      block.GasUsed,
		GasLimit:     block.GasLimit,
	}

	data, err := json.Marshal(blockData)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// 使用节点私钥签名
	signature, err := s.node.privateKey.Sign(data)
	if err != nil {
		return fmt.Errorf("failed to sign block: %w", err)
	}

	block.Signature = signature
	return nil
}

// 计算区块哈希
func calculateBlockHash(block *types.Block) (string, error) {
	// 创建用于哈希计算的区块数据结构
	blockData := struct {
		Height       uint64
		PreviousHash string
		Timestamp    time.Time
		Transactions []types.Transaction
		StateRoot    string
		Proposer     string
		GasUsed      uint64
		GasLimit     uint64
		Signature    []byte
	}{
		Height:       block.Height,
		PreviousHash: block.PreviousHash,
		Timestamp:    block.Timestamp,
		Transactions: block.Transactions,
		StateRoot:    block.StateRoot,
		Proposer:     block.Proposer,
		GasUsed:      block.GasUsed,
		GasLimit:     block.GasLimit,
		Signature:    block.Signature,
	}

	// 序列化区块数据
	data, err := json.Marshal(blockData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal block: %w", err)
	}

	// 计算哈希
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
