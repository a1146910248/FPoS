package p2p

import (
	"FPoS/config"
	"FPoS/core/consensus"
	"FPoS/core/ethereum"
	"FPoS/types"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

const _MaxBlockGasLimit_ = 810000

type Sequencer struct {
	node             *Layer2Node
	blockHeight      uint64
	mu               sync.Mutex
	maxBlockGasLimit uint64
	blockVoteChan    chan types.BlockVote

	electionMgr *consensus.ElectionManager
	ethClient   *ethereum.EthereumClient
}

func NewSequencer(node *Layer2Node, config *config.Config) (*Sequencer, error) {
	ethClient, err := ethereum.NewEthereumClient(config.Ethereum)
	if err != nil {
		fmt.Printf("connect ethereum failed:" + err.Error())
		return nil, err
	}
	node.electionMgr.SetEth(ethClient)
	//node.isSequencer = true
	seq := &Sequencer{
		node:        node,
		blockHeight: 0,
		//maxBlockGasLimit: 30_000_000, // 区块 gas 上限为 30,000,000
		maxBlockGasLimit: _MaxBlockGasLimit_, // 区块 gas 上限为 30,000,000
		ethClient:        ethClient,
		blockVoteChan:    make(chan types.BlockVote, 2),
	}
	node.sequencer = seq
	return seq, nil
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
			s.node.mu.Lock()
			s.node.isSequencer = true
			s.node.sequencer = s
			s.node.mu.Unlock()
			addr, _ := types.PublicKeyToAddress(s.node.publicKey)
			fmt.Printf("Node %s became the new sequencer\n", addr)
		} else {
			// 没被选上，将排序器清空
			s.node.mu.Lock()
			s.node.isSequencer = false
			s.node.mu.Unlock()
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
				// 一旦当选应该立即置否以防止连续出块
				s.node.mu.Lock()
				s.node.isSequencer = false
				s.node.mu.Unlock()
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
		Height:       s.node.latestBlock + 1,
		Timestamp:    time.Now(),
		Transactions: transactions,
		StateRoot:    s.node.stateDB.GetStateRoot(),
		TxRoot:       types.CalculateMerkleRoot(transactions),
		Proposer:     proPub,
		GasUsed:      totalGas, // 记录区块使用的总gas
		GasLimit:     s.maxBlockGasLimit,
	}
	s.blockHeight = s.node.latestBlock
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
	// 收集投票
	err = s.PubVoteReq(block)
	if err != nil {
		return
	}

	for i := 0; i < 2; i++ {
		select {
		case vote := <-s.blockVoteChan:
			block.Votes = append(block.Votes, vote)
		}
	}

	// 提交区块到L1
	go func() {
		if err := s.ethClient.SubmitBlock(&block); err != nil {
			fmt.Printf("Failed to submit block to L1: %v\n", err)
			// 不要因为L1提交失败而影响L2的共识
		}
	}()

	// 广播区块
	if err := s.node.BroadcastBlock(block); err != nil {
		fmt.Printf("Failed to broadcast block: %v\n", err)
		return
	}

	fmt.Printf("New block produced: height=%d, txs=%d, gasUsed=%d\n",
		block.Height, len(block.Transactions), totalGas)
}

func (s *Sequencer) PubVoteReq(block types.Block) error {
	requestID := uuid.New().String()
	s.node.mu.Lock()
	s.node.currentBlockVoteRequestID = requestID
	s.node.mu.Unlock()
	address, _ := types.PublicKeyToAddress(s.node.publicKey)
	req := BlockVoteReq{
		Type:      BlockVoteRequest,
		RequestID: requestID,
		Address:   address,
		Block:     block,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal tx sync request: %w", err)
	}

	// 发布同步请求
	if err := s.node.topic.blockVoteTopic.Publish(s.node.ctx, data); err != nil {
		return fmt.Errorf("failed to publish tx sync request: %w", err)
	}
	logger.Info("等待收集投票")
	return nil
}
