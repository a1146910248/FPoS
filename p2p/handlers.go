package p2p

import (
	. "FPoS/types"
	"fmt"
)

func (n *Layer2Node) SetTransactionHandler(handler TransactionHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.handlers.TxHandler = handler
}

func (n *Layer2Node) SetBlockHandler(handler BlockHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.handlers.BlockHandler = handler
}

func (n *Layer2Node) validateTransaction(tx Transaction) bool {
	n.mu.RLock()
	handler := n.handlers.TxHandler
	n.mu.RUnlock()

	if handler == nil {
		return n.defaultTxValidation(&tx)
	}
	return handler(tx)
}

func (n *Layer2Node) validateBlock(block Block) bool {
	n.mu.RLock()
	handler := n.handlers.BlockHandler
	n.mu.RUnlock()

	if handler == nil {
		return n.defaultBlockValidation(block)
	}
	return handler(block)
}

func (n *Layer2Node) defaultTxValidation(tx *Transaction) bool {
	if isRight, err := CalculateTxHash(tx); !isRight || err != nil {
		fmt.Println("交易哈希错误")
		return false
	}

	if _, exists := n.txPool.Load(tx.Hash); exists {
		return false
	}

	if tx.Timestamp.IsZero() {
		return false
	}

	if tx.From == "" || tx.To == "" {
		return false
	}

	if len(tx.Signature) == 0 {
		return false
	}

	// 检查nonce值
	currentNonce := n.stateDB.GetNonce(tx.From)
	if tx.Nonce != currentNonce+1 {
		fmt.Printf("交易nonce无效: 期望 %d, 实际 %d\n", currentNonce+1, tx.Nonce)
		return false
	}

	// Gas和余额检查
	if err := n.stateDB.ValidateTransaction(tx, n.minGasPrice); err != nil {
		fmt.Println("交易验证不通过：", err)
		return false
	}
	if err := VerifyTransactionSignature(tx, n); err != nil {
		fmt.Printf("Transaction signature verification failed: %v\n", err)
		return false
	}
	return true
}

func (n *Layer2Node) defaultBlockValidation(block Block) bool {
	n.mu.RLock()
	currentHeight := n.latestBlock
	n.mu.RUnlock()

	// 检查block hash
	if hash, err := CalculateBlockHash(&block); hash != block.Hash || err != nil {
		return false
	}
	if block.Height <= currentHeight {
		return false
	}

	previousBlock, exists := n.blockCache.Load(block.Height - 1)
	if exists {
		if prev, ok := previousBlock.(Block); ok {
			if prev.Hash != block.PreviousHash {
				return false
			}
		}
	}

	if block.Timestamp.IsZero() {
		return false
	}

	for _, tx := range block.Transactions {
		if !n.validateTxForBlock(&tx) {
			return false
		}
	}

	if block.Proposer == "" || len(block.Signature) == 0 {
		return false
	}
	err := VerifyBlockSignature(&block, n)
	if err != nil {
		return false
	}
	return true
}

// 验证区块中的每条交易，需要在交易池中存在，与同步交易刚好相反
func (n *Layer2Node) validateTxForBlock(tx *Transaction) bool {
	if isRight, err := CalculateTxHash(tx); !isRight || err != nil {
		fmt.Println("交易哈希错误")
		return false
	}
	if n.isSequencer {
		if _, exists := n.txPool.Load(tx.Hash); exists {
			return false
		}
	} else {
		if _, exists := n.txPool.Load(tx.Hash); !exists {
			return false
		}
	}

	if tx.Timestamp.IsZero() {
		return false
	}

	if tx.From == "" || tx.To == "" {
		return false
	}

	if len(tx.Signature) == 0 {
		return false
	}

	// 检查nonce值
	currentNonce := n.stateDB.GetNonce(tx.From)
	if tx.Nonce != currentNonce+1 {
		fmt.Printf("交易nonce无效: 期望 %d, 实际 %d\n", currentNonce+1, tx.Nonce)
		return false
	}

	// Gas和余额检查
	if err := n.stateDB.ValidateTransaction(tx, n.minGasPrice); err != nil {
		fmt.Println("交易验证不通过：", err)
		return false
	}
	if err := VerifyTransactionSignature(tx, n); err != nil {
		fmt.Printf("Transaction signature verification failed: %v\n", err)
		return false
	}
	return true
}
