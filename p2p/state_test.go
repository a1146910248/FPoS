package p2p

import (
	"FPoS/types"
	"testing"
	"time"
)

var tx = &types.Transaction{
	Hash:            "123",
	From:            "321",
	To:              "123",
	Value:           0,
	Nonce:           1,
	GasPrice:        1,
	GasLimit:        2000000000,
	GasUsed:         1,
	Timestamp:       time.Time{},
	Signature:       nil,
	StatLog:         types.StatLog{},
	IsContract:      true,
	ContractAddress: "",
	Input:           -1,
	Output:          0,
	FuncString:      "retrieve",
}

func TestTransferFuncCode(t *testing.T) {
	state := NewStateDB()
	err := state.ExecuteContractTransaction(tx)
	if err != nil {
		t.Errorf("failed to execute contract transaction: %s", err)
	}
}
