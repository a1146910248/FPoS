package dac

import (
	"FPoS/types"
	"crypto/sha256"
	"encoding/hex"
)

// UpdateWorldState 更新世界状态
func (dm *DACManager) UpdateWorldState(block types.Block) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 处理区块中的所有交易，只存储交易信息
	for _, tx := range block.Transactions {
		// 存储交易到交易树
		if err := dm.StoreTransaction(tx); err != nil {
			return err
		}
	}

	return nil
}

// ComputeStateRoot 计算世界状态根哈希
func (dm *DACManager) ComputeStateRoot() (string, error) {
	// 获取账户树根哈希
	accountRoot := dm.accountTree.GetRootHash()

	// 获取交易树根哈希
	txRoot := dm.txTree.GetRootHash()

	// 计算组合哈希
	h := sha256.New()
	h.Write(accountRoot)
	h.Write(txRoot)
	stateRoot := h.Sum(nil)

	return hex.EncodeToString(stateRoot), nil
}
