package p2p

import (
	"FPoS/config"
	"FPoS/core/consensus"
	"FPoS/core/ethereum"
	"FPoS/types"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

//const _MaxBlockGasLimit_ = 810000

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
		maxBlockGasLimit: viper.GetUint64("L2.maxBlockGasLimit"), // 区块 gas 上限为 30,000,000
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
		Height:        s.node.latestBlock + 1,
		Timestamp:     time.Now(),
		Transactions:  transactions,
		StateRoot:     s.node.stateDB.GetStateRoot(),
		TxRoot:        types.CalculateMerkleRoot(transactions),
		Proposer:      proPub,
		GasUsed:       totalGas, // 记录区块使用的总gas
		GasLimit:      s.maxBlockGasLimit,
		FinalProof:    mockPlonkProof().FinalProof,
		KZGCommitment: mockKZGCommit().Commitment,
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

	// 1. 请求DAC状态根
	accountRoot, txRoot, err := s.node.RequestDACStateRoot()
	if err != nil {
		fmt.Printf("请求DAC状态根失败: %v\n", err)
		return
	}

	// 2. 从Layer 1获取状态根
	l1StateRoot, l1TxRoot, err := s.getLayer1StateRoot()
	if err != nil {
		fmt.Printf("获取Layer 1状态根失败: %v\n", err)
		return
	}

	// 3. 验证DAC状态根与Layer 1状态根的一致性
	stateIsValid := s.verifyStateRootConsistency(accountRoot, txRoot, l1StateRoot, l1TxRoot, block.Height)
	if !stateIsValid {
		logger.Warn("DAC状态根与Layer 1状态根不一致\n")
		return
	}

	// 4. 提交区块到DAC网络并等待响应
	dacProofs, newAccountRoot, newTxRoot, err := s.submitBlockAndWaitForProofs(block, l1StateRoot)
	if err != nil {
		fmt.Printf("获取DAC证明失败: %v，将继续提交区块\n", err)
		// 即使DAC证明获取失败，也继续提交区块
	}

	// 将交易保存到历史记录
	for _, tx := range block.Transactions {
		// 更新状态
		if tx.StatLog.Status != types.TxStatusL1Confirmed && tx.StatLog.Status != types.TxStatusL1Failed {
			tx.StatLog.Status = types.TxStatusConfirmed
		}
		tx.StatLog.BlockHash = blockHash
		tx.StatLog.BlockHeight = block.Height
		s.node.txHistory.Store(tx.Hash, tx)
	}

	// 一旦当选应该立即置否以防止连续出块
	s.node.mu.Lock()
	s.node.isSequencer = false
	s.node.mu.Unlock()

	// 收集投票
	err = s.PubVoteReq(block)
	if err != nil {
		return
	}

	for i := 0; i < 2; i++ {
		select {
		case vote := <-s.blockVoteChan:
			block.Votes = append(block.Votes, vote)
			if !vote.Approve {
				block.IsSus = true
			}
		}
	}

	// 5. 将区块和DAC证明一起提交到Layer 1
	if err := s.submitBlockWithProofsToLayer1(block, newAccountRoot, newTxRoot, dacProofs); err != nil {
		fmt.Printf("提交区块和证明到Layer 1失败: %v\n", err)
		// 不中断区块链流程，只记录错误
	}

	// 广播区块
	if err := s.node.BroadcastBlock(block); err != nil {
		fmt.Printf("Failed to broadcast block: %v\n", err)
		return
	}

	fmt.Printf("New block produced: height=%d, txs=%d, gasUsed=%d\n",
		block.Height, len(block.Transactions), totalGas)
}

// 从Layer 1获取状态根
func (s *Sequencer) getLayer1StateRoot() (ac string, tx string, err error) {
	ac, tx, err = s.ethClient.GetLatestDACRootsAsString()
	if err != nil {
		return "", "", err
	}

	return ac, tx, nil
}

// 验证状态根一致性
func (s *Sequencer) verifyStateRootConsistency(accountRoot, txRoot, l1StateRoot, l1TxRoot string, height uint64) bool {
	if height == 1 {
		return true
	}
	// 根据应用需求实现验证逻辑
	if accountRoot == l1StateRoot && txRoot == l1TxRoot {
		return true
	}
	return false // 示例实现
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

// 提交区块到DAC并等待证明 - 修正版本
func (s *Sequencer) submitBlockAndWaitForProofs(block types.Block, l1StateRoot string) ([][]byte, string, string, error) {
	// 创建请求ID
	requestID := uuid.New().String()

	// 初始化响应收集
	responseChannel := make(chan *DACBlockCommitResp, 1)

	// 设置临时响应处理器
	s.registerTempResponseHandler(requestID, responseChannel)

	// 提交区块到DAC网络
	if err := s.node.SubmitBlockToDACNetwork(requestID, block, l1StateRoot); err != nil {
		return nil, "", "", fmt.Errorf("提交区块到DAC网络失败: %w", err)
	}

	// A. 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// B. 等待响应
	select {
	case resp := <-responseChannel:
		// 使用读锁安全地访问响应数据
		resp.mu.RLock()

		// 找到第一个有效证明
		var proofs [][]byte
		for _, proof := range resp.proofs {
			proofs = proof
			break
		}

		// 获取根值
		accountRoot := resp.newAccountRoot
		txRoot := resp.newTxRoot

		resp.mu.RUnlock()

		return proofs, accountRoot, txRoot, nil

	case <-ctx.Done():
		return nil, "", "", fmt.Errorf("等待DAC证明超时")
	}
}

// 注册临时响应处理器
func (s *Sequencer) registerTempResponseHandler(requestID string, responseChannel chan<- *DACBlockCommitResp) {
	go func() {
		// 设置超时
		timeout := time.After(5 * time.Second)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 检查是否有对应的响应
				respValue, exists := blockCommitResponses.Load(requestID)
				if !exists {
					continue
				}

				resp := respValue.(*DACBlockCommitResp)

				// 获取读锁
				resp.mu.RLock()

				// 获取DAC成员数量
				dacState := s.node.dacMgr.GetState()
				requiredResponses := len(dacState.CurrentMembers) * 2 / 3 // 2/3多数
				responseCount := len(resp.responses)

				// 检查是否有足够的响应
				hasEnoughResponses := responseCount >= requiredResponses

				// 释放读锁
				resp.mu.RUnlock()

				if hasEnoughResponses {
					// 再次获取读锁进行一致性检查
					resp.mu.RLock()
					consistent := s.node.verifyBlockCommitConsistency(resp)
					resp.mu.RUnlock()

					if consistent {
						// 发送到通道
						responseChannel <- resp
						return
					}
				}

			case <-timeout:
				// 超时退出
				return
			}
		}
	}()
}

// 将区块和DAC证明一起提交到Layer 1
func (s *Sequencer) submitBlockWithProofsToLayer1(block types.Block, accountRoot, txRoot string, proofs [][]byte) error {
	// 如果缺少证明，仍然提交区块
	if len(proofs) == 0 || accountRoot == "" || txRoot == "" {
		if err := s.ethClient.SubmitBlock(&block); err != nil {
			return fmt.Errorf("提交区块到Layer 1失败: %w", err)
		}
		fmt.Printf("已提交区块到Layer 1（无DAC证明）: 区块高度=%d\n", block.Height)
		return nil
	}

	// 调用Layer 1客户端提交带证明的区块
	if err := s.ethClient.SubmitBlockWithDAC(&block, accountRoot, txRoot, proofs); err != nil {
		return fmt.Errorf("提交区块和证明到Layer 1失败: %w", err)
	}

	fmt.Printf("已提交区块和DAC证明到Layer 1: 区块高度=%d, DAC状态根=%s\n",
		block.Height, accountRoot)
	return nil
}
