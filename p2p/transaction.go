package p2p

import (
	"FPoS/types"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.org/x/crypto/sha3"
	"math/rand"
	"time"
)

func (n *Layer2Node) StartPeriodicTransaction() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		// 获取本节点的地址
		fromAddress, err := PublicKeyToAddress(n.privateKey.GetPublic())
		if err != nil {
			fmt.Printf("生成发送方地址失败: %v\n", err)
			return
		}

		for {
			select {
			case <-n.ctx.Done():
				return
			case <-ticker.C:
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

				fmt.Printf("发送交易成功: %s\n", tx.Hash)
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

	// 遍历所有连接的节点，查找匹配的地址
	found := false
	var pubKey crypto.PubKey
	// 发送交易的地址可能是自己的，也可能是对等节点的其他人的
	if addr, err := PublicKeyToAddress(n.publicKey); addr == fromAddr {
		if err != nil {
			return fmt.Errorf("invalid address format")
		}
		pubKey = n.publicKey
		found = true
	} else {
		for _, peerID := range n.host.Network().Peers() {
			if pk := n.host.Peerstore().PubKey(peerID); pk != nil {
				addr, err = PublicKeyToAddress(pk)
				if err != nil {
					continue
				}
				if addr == fromAddr {
					pubKey = pk
					found = true
					break
				}
			}
		}
	}

	if !found {
		return fmt.Errorf("could not find public key for address: %s", fromAddr)
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

// 从公钥生成地址
func PublicKeyToAddress(pub crypto.PubKey) (string, error) {
	// 获取公钥的原始字节
	pubBytes, err := pub.Raw()
	if err != nil {
		return "", err
	}

	// 使用 Keccak-256 哈希公钥
	hash := sha3.NewLegacyKeccak256()
	hash.Write(pubBytes)

	// 取最后20字节作为地址（类似以太坊）
	address := hash.Sum(nil)[12:]

	// 返回带0x前缀的地址
	return "0x" + hex.EncodeToString(address), nil
}

func (n *Layer2Node) getRandomToPubKey() (string, error) {
	n.stateDB.mu.RLock()
	defer n.stateDB.mu.RUnlock()

	// 获取所有账户地址
	addresses := make([]string, 0)
	for addr := range n.stateDB.accounts {
		// 排除自己的地址
		if pb, _ := PublicKeyToAddress(n.publicKey); addr != pb {
			addresses = append(addresses, addr)
		}
	}
	if len(addresses) == 0 {
		return "", fmt.Errorf("no other addresses available in  state")
	}

	return addresses[rand.Intn(len(addresses))], nil
}

// 当交易从交易池移除时（超时或其他原因）
func (n *Layer2Node) removeFromTxPool(tx *types.Transaction) {
	n.txPool.Delete(tx.Hash)
	n.stateDB.RestorePendingState(tx)
}
