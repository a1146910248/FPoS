package model

import (
	"math/big"
	"time"
)

// 链状态
type ChainStats struct {
	CurrentTPS  float64  `json:"current_tps"`
	PeakTPS     float64  `json:"peak_tps"`
	TotalTx     uint64   `json:"total_tx"`
	BlockHeight uint64   `json:"block_height"`
	ActiveUsers uint64   `json:"active_users"`
	L1Blocks    uint64   `json:"l1_blocks"`
	L2Blocks    uint64   `json:"l2_blocks"`
	L1Balance   *big.Int `json:"l1_balance"`
	L2TPS       float64  `json:"l2_tps"`
}

// TransactionList 交易列表响应结构
type TransactionList struct {
	Total int64         `json:"total"`
	List  []Transaction `json:"list"`
}

// Transaction 交易信息结构
type Transaction struct {
	Hash      string    `json:"hash"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Value     uint64    `json:"value"`
	Nonce     uint64    `json:"nonce"`
	GasPrice  uint64    `json:"gas_price"`
	GasLimit  uint64    `json:"gas_limit"`
	GasUsed   uint64    `json:"gas_used"`
	Timestamp time.Time `json:"timestamp"`
	Status    int       `json:"status"`     // 使用枚举值 0-4
	BlockHash string    `json:"block_hash"` // 所属区块hash
}
