package p2p

import (
	"FPoS/types"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"golang.org/x/crypto/sha3"
)

func (n *Layer2Node) StartPeriodicTransaction() {
	// 添加标志防止重复启动
	n.mu.Lock()
	if n.periodicTxStarted { // 需要在 Layer2Node 结构体中添加此字段
		n.mu.Unlock()
		return
	}
	n.periodicTxStarted = true
	n.mu.Unlock()

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		// 获取本节点的地址
		fromAddress, err := types.PublicKeyToAddress(n.privateKey.GetPublic())
		if err != nil {
			fmt.Printf("生成发送方地址失败: %v\n", err)
			return
		}
		for {
			select {
			case <-n.ctx.Done():
				logger.Error("定时交易退出")
				return
			case <-ticker.C:
				//default:
				// 检查是否为 sequencer
				n.mu.RLock()
				isSeq := n.isSequencer
				n.mu.RUnlock()
				if isSeq {
					fmt.Printf("Skipping transaction generation for sequencer node\n")
					continue
				}
				// 获取一个随机的目标地址
				toAddress, err := n.getRandomToPubKey()
				if err != nil {
					fmt.Printf("生成目标地址失败: %v\n", err)
					continue
				}
				// 创建一个新交易
				tx := types.Transaction{
					From:      fromAddress,
					To:        toAddress, // 生成随机的目标地址
					Value:     uint64(rand.Intn(10000000)),
					Nonce:     n.stateDB.GetNonce(fromAddress) + 1,
					GasLimit:  types.GasLimit,
					GasUsed:   types.TransferGas,
					GasPrice:  types.GasPrice,
					Timestamp: time.Now(),
				}
				// 交易初始状态
				tx.StatLog.Status = types.TxStatusPending
				// 计算交易哈希
				hash, err := calculateTxHash(&tx)
				if err != nil {
					fmt.Printf("计算交易哈希失败: %v\n", err)
					continue
				}
				tx.Hash = hash

				// 签名交易
				if err := SignTransaction(&tx, n); err != nil {
					fmt.Printf("签名交易失败: %v\n", err)
					continue
				}

				// 广播交易
				if err := n.BroadcastTransaction(tx); err != nil {
					fmt.Printf("广播交易失败: %v\n", err)
					continue
				}

				//fmt.Printf("发送交易成功: %s\n", tx.Hash)
			}
		}
	}()
}

// 添加签名方法
func SignTransaction(tx *types.Transaction, node *Layer2Node) error {
	// 使用节点的私钥对交易进行签名
	message, err := json.Marshal(struct {
		From      string
		To        string
		Value     uint64
		Nonce     uint64
		GasLimit  uint64
		GasUsed   uint64
		GasPrice  uint64
		Timestamp time.Time
	}{
		From:      tx.From,
		To:        tx.To,
		Value:     tx.Value,
		GasLimit:  tx.GasLimit,
		GasUsed:   tx.GasUsed,
		GasPrice:  tx.GasPrice,
		Nonce:     tx.Nonce,
		Timestamp: tx.Timestamp,
	})
	if err != nil {
		return err
	}

	signature, err := node.privateKey.Sign(message)
	if err != nil {
		return err
	}

	tx.Signature = signature
	return nil
}

// 验证交易签名的方法
func VerifyTransactionSignature(tx *types.Transaction, n *Layer2Node) error {
	// 重建签名消息
	message, err := json.Marshal(struct {
		From      string
		To        string
		Value     uint64
		Nonce     uint64
		GasLimit  uint64
		GasUsed   uint64
		GasPrice  uint64
		Timestamp time.Time
	}{
		From:      tx.From,
		To:        tx.To,
		Value:     tx.Value,
		GasLimit:  tx.GasLimit,
		GasUsed:   tx.GasUsed,
		GasPrice:  tx.GasPrice,
		Nonce:     tx.Nonce,
		Timestamp: tx.Timestamp,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal transaction for verification: %w", err)
	}

	// 从From地址反推公钥
	fromAddr := tx.From
	if len(fromAddr) < 2 || fromAddr[:2] != "0x" {
		return fmt.Errorf("invalid address format")
	}

	// 首先检查是否是本节点的地址
	if addr, err := types.PublicKeyToAddress(n.publicKey); err == nil && addr == fromAddr {
		valid, err := n.publicKey.Verify(message, tx.Signature)
		if err != nil {
			return fmt.Errorf("signature verification error: %w", err)
		}
		if !valid {
			return fmt.Errorf("invalid block signature")
		}
		return nil
	}

	// 从账户状态中获取公钥
	pubKey, err := n.stateDB.GetAccountPublicKey(fromAddr)
	if err != nil {
		return fmt.Errorf("failed to get proposer public key: %w", err)
	}
	// 验证签名
	valid, err := pubKey.Verify(message, tx.Signature)
	if err != nil {
		return fmt.Errorf("signature verification error: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid transaction signature")
	}

	return nil
}

// 计算交易哈希的函数
func calculateTxHash(tx *types.Transaction) (string, error) {
	// 创建一个不包含哈希的交易结构用于序列化
	txData := struct {
		From      string    `json:"from"`
		To        string    `json:"to"`
		Value     uint64    `json:"value"`
		Nonce     uint64    `json:"nonce"`
		GasPrice  uint64    `json:"gasPrice"` // 用户愿意支付的每单位gas的价格
		GasLimit  uint64    `json:"gasLimit"` // 用户愿意支付的最大gas数量
		GasUsed   uint64    `json:"gasUsed"`  // 实际使用的gas数量
		Timestamp time.Time `json:"timestamp"`
	}{
		From:      tx.From,
		To:        tx.To,
		Value:     tx.Value,
		GasLimit:  tx.GasLimit,
		GasUsed:   tx.GasUsed,
		GasPrice:  tx.GasPrice,
		Nonce:     tx.Nonce,
		Timestamp: tx.Timestamp,
	}

	// 序列化交易数据
	data, err := json.Marshal(txData)
	if err != nil {
		return "", err
	}

	// 使用 Keccak-256 哈希算法
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)

	// 转换为十六进制字符串，添加"0x"前缀
	return "0x" + hex.EncodeToString(hash.Sum(nil)), nil
}

// 验证交易哈希的函数
func CalculateTxHash(tx *types.Transaction) (bool, error) {
	// 创建一个不包含哈希的交易结构用于序列化
	txData := struct {
		From      string    `json:"from"`
		To        string    `json:"to"`
		Value     uint64    `json:"value"`
		Nonce     uint64    `json:"nonce"`
		GasPrice  uint64    `json:"gasPrice"` // 用户愿意支付的每单位gas的价格
		GasLimit  uint64    `json:"gasLimit"` // 用户愿意支付的最大gas数量
		GasUsed   uint64    `json:"gasUsed"`  // 实际使用的gas数量
		Timestamp time.Time `json:"timestamp"`
	}{
		From:      tx.From,
		To:        tx.To,
		Value:     tx.Value,
		GasLimit:  tx.GasLimit,
		GasUsed:   tx.GasUsed,
		GasPrice:  tx.GasPrice,
		Nonce:     tx.Nonce,
		Timestamp: tx.Timestamp,
	}

	// 序列化交易数据
	data, err := json.Marshal(txData)
	if err != nil {
		return false, err
	}

	// 使用 Keccak-256 哈希算法
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)

	// 转换为十六进制字符串，添加"0x"前缀
	return "0x"+hex.EncodeToString(hash.Sum(nil)) == tx.Hash, nil
}

func (n *Layer2Node) getRandomToPubKey() (string, error) {
	// 先复制地址列表
	n.stateDB.RLock()
	addresses := make([]string, 0, len(n.stateDB.accounts))
	for addr := range n.stateDB.accounts {
		addresses = append(addresses, addr)
	}
	n.stateDB.RUnlock()

	// 在无锁的情况下处理数据
	filteredAddrs := make([]string, 0)
	selfAddr, _ := types.PublicKeyToAddress(n.publicKey)
	for _, addr := range addresses {
		if addr != selfAddr {
			filteredAddrs = append(filteredAddrs, addr)
		}
	}

	if len(filteredAddrs) == 0 {
		return "", fmt.Errorf("no other addresses available")
	}

	return filteredAddrs[rand.Intn(len(filteredAddrs))], nil
}

// 当交易从交易池移除时（超时或其他原因）
func (n *Layer2Node) removeFromTxPool(tx *types.Transaction) {
	n.txPool.Delete(tx.Hash)
	n.stateDB.RestorePendingState(tx)
}

// GetTotalTxCount 获取总交易数
func (n *Layer2Node) GetTotalTxCount() uint64 {
	n.txCountMu.RLock()
	defer n.txCountMu.RUnlock()
	return n.txCount
}

// IncrementTxCount 增加交易计数
func (n *Layer2Node) IncrementTxCount() {
	n.txCountMu.Lock()
	n.txCount++
	n.txCountMu.Unlock()
}

// GetTransactions 获取指定范围的交易
func GetTransactions(limit int, offset int) []types.Transaction {
	node := GetNode()
	if node == nil {
		return []types.Transaction{}
	}
	return node.GetTransactions(limit, offset)
}

// GetTotalTransactions 获取总交易数
func GetTotalTransactions() int {
	node := GetNode()
	if node == nil {
		return 0
	}
	return node.GetTotalTransactions()
}

// 节点级别的方法保持不变，但改名以区分
func (n *Layer2Node) GetTransactions(limit int, offset int) []types.Transaction {
	transactions := make([]types.Transaction, 0)

	// 收集所有交易（包括待确认和已确认的）
	var allTxs []types.Transaction

	// 从交易池获取待确认交易
	n.txPool.Range(func(_, value interface{}) bool {
		if tx, ok := value.(types.Transaction); ok {
			allTxs = append(allTxs, tx)
		}
		return true
	})

	// 从历史记录获取已确认交易
	n.txHistory.Range(func(_, value interface{}) bool {
		if tx, ok := value.(types.Transaction); ok {
			allTxs = append(allTxs, tx)
		}
		return true
	})

	// 按时间戳排序，最新的交易在前
	sort.Slice(allTxs, func(i, j int) bool {
		return allTxs[i].Timestamp.After(allTxs[j].Timestamp)
	})

	// 应用分页
	start := offset
	end := offset + limit
	if start >= len(allTxs) {
		return transactions
	}
	if end > len(allTxs) {
		end = len(allTxs)
	}

	return allTxs[start:end]
}

func (n *Layer2Node) GetTotalTransactions() int {
	var count int
	// 计算交易池中的交易数量
	n.txPool.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	// 计算历史记录中的交易数量
	n.txHistory.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
