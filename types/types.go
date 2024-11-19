package types

import (
	"encoding/json"
	"time"
)

// 常量Gas，纯转账交易固定为21000Gas
const (
	TransferGas uint64 = 210      // 转账交易固定消耗的gas数量,以太坊为21000
	GasLimit    uint64 = 10000000 // 默认上限值
	GasPrice    uint64 = 20       // 默认gas价格
)

// 交易结构
type Transaction struct {
	Hash      string    `json:"hash"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Value     uint64    `json:"value"`
	Nonce     uint64    `json:"nonce"`
	GasPrice  uint64    `json:"gasPrice"` // 用户愿意支付的每单位gas的价格
	GasLimit  uint64    `json:"gasLimit"` // 用户愿意支付的最大gas数量
	GasUsed   uint64    `json:"gasUsed"`  // 实际使用的gas数量
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
	TxRoot       string        `json:"txRoot"` // 交易默克尔根
	Proposer     string        `json:"proposer"`
	GasUsed      uint64        `json:"gasUsed"`  // 区块中所有交易消耗的总gas
	GasLimit     uint64        `json:"gasLimit"` // 区块gas上限
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
