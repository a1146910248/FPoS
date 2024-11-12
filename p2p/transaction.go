package p2p

import (
	"FPoS/types"
	crand "crypto/rand"
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
				toAddress, err := getRandomToPubKey()
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
				if err := n.signTransaction(&tx); err != nil {
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
func (n *Layer2Node) signTransaction(tx *types.Transaction) error {
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

	signature, err := n.privateKey.Sign(message)
	if err != nil {
		return err
	}

	tx.Signature = signature
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

// TODO 不能生成随意的pubkey， 而是应该从世界状态池中选择一个链接的节点进行发送
func getRandomToPubKey() (string, error) {
	// 生成一个随机的目标节点公钥
	targetPrivKey, _, err := crypto.GenerateKeyPairWithReader(
		crypto.Ed25519,
		2048,
		crand.Reader,
	)
	if err != nil {
		fmt.Printf("生成目标公钥失败: %v\n", err)
		return "", err
	}

	// 从目标公钥生成地址
	toAddress, err := PublicKeyToAddress(targetPrivKey.GetPublic())
	if err != nil {
		fmt.Printf("生成接收方地址失败: %v\n", err)
		return "", err
	}
	return toAddress, nil
}
