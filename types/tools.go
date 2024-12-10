package types

import (
	"encoding/hex"
	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.org/x/crypto/sha3"
)

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
