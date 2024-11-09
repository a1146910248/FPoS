package types

import (
	"encoding/json"
	"time"
)

// 交易结构
type Transaction struct {
	Hash      string    `json:"hash"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Value     uint64    `json:"value"`
	Nonce     uint64    `json:"nonce"`
	Timestamp time.Time `json:"timestamp"`
	Signature []byte    `json:"signature"`
}

// 区块结构
type Block struct {
	Height       uint64        `json:"height"`
	Hash         string        `json:"hash"`
	PreviousHash string        `json:"previousHash"`
	Timestamp    time.Time     `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	StateRoot    string        `json:"stateRoot"`
	Proposer     string        `json:"proposer"`
	Signature    []byte        `json:"signature"`
}

// 状态同步请求
type SyncRequest struct {
	FromHeight uint64 `json:"fromHeight"`
	ToHeight   uint64 `json:"toHeight"`
}

// 协议消息类型
const (
	MsgTypeTx           = "tx"
	MsgTypeBlock        = "block"
	MsgTypeSyncRequest  = "sync_req"
	MsgTypeSyncResponse = "sync_resp"
	MsgTypeState        = "state"
)

// P2P消息结构
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// 处理器类型定义
type TransactionHandler func(tx Transaction) bool
type BlockHandler func(block Block) bool

type Handlers struct {
	TxHandler    TransactionHandler
	BlockHandler BlockHandler
}
