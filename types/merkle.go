package types

import (
	"crypto/sha256"
	"encoding/hex"
)

// CalculateMerkleRoot 添加默克尔树相关函数
func CalculateMerkleRoot(txs []Transaction) string {
	if len(txs) == 0 {
		return ""
	}

	var hashes [][]byte
	for _, tx := range txs {
		txHash, _ := hex.DecodeString(tx.Hash[2:]) // 去掉0x前缀
		hashes = append(hashes, txHash)
	}

	// 构建默克尔树
	for len(hashes) > 1 {
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}
		var nextLevel [][]byte
		for i := 0; i < len(hashes); i += 2 {
			hash := sha256.Sum256(append(hashes[i], hashes[i+1]...))
			nextLevel = append(nextLevel, hash[:])
		}
		hashes = nextLevel
	}

	return "0x" + hex.EncodeToString(hashes[0])
}
