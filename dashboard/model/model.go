package model

import "math/big"

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
