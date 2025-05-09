package ethereum

import (
	"FPoS/types"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	etype "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
)

type EthereumConfig struct {
	RPCURL          string `yaml:"rpc_url"`
	ContractAddress string `yaml:"contract_address"`
	PrivateKey      string `yaml:"private_key"`
	GasLimit        uint64 `yaml:"gas_limit"`
	GasPrice        int64  `yaml:"gas_price"`
	ConfirmBlocks   uint64 `yaml:"confirm_blocks"`
}

type EthereumClient struct {
	client     *ethclient.Client
	chainID    *big.Int
	contract   *Ethereum
	config     *EthereumConfig
	privKey    *ecdsa.PrivateKey
	statusChan chan TxStatusEvent // 状态通知通道
}

// 添加交易状态事件类型
type TxStatusEvent struct {
	Block       types.Block
	Status      int
	L1TxHash    string
	L1Timestamp time.Time
}

// BlockMetadata 包含提交到Layer 1的区块元数据
type BlockMetadata struct {
	Block          *types.Block
	DACAccountRoot string   // DAC账户状态根（十六进制字符串）
	DACTxRoot      string   // DAC交易状态根（十六进制字符串）
	DACProofs      [][]byte // DAC证明数据
}

func NewEthereumClient(config *EthereumConfig) (*EthereumClient, error) {
	client, err := ethclient.Dial(config.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum node: %v", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	// 解析私钥
	privKey, err := crypto.HexToECDSA(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	// 使用 NewEthereum 而不是 NewL2Contract
	contractAddress := common.HexToAddress(config.ContractAddress)
	contract, err := NewEthereum(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to load ethereum contract: %v", err)
	}

	return &EthereumClient{
		client:     client,
		chainID:    chainID,
		contract:   contract,
		config:     config,
		privKey:    privKey,
		statusChan: make(chan TxStatusEvent, 100),
	}, nil
}

// GetStatusChannel 获取状态通知通道
func (ec *EthereumClient) GetStatusChannel() <-chan TxStatusEvent {
	return ec.statusChan
}

// getTransactOpts 获取交易选项
func (ec *EthereumClient) getTransactOpts() (*bind.TransactOpts, error) {
	nonce, err := ec.client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(ec.privKey.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	gasPrice := big.NewInt(ec.config.GasPrice)
	auth, err := bind.NewKeyedTransactorWithChainID(ec.privKey, ec.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %v", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = ec.config.GasLimit
	auth.GasPrice = gasPrice

	return auth, nil
}

// SubmitBlock 提交区块到L1
func (ec *EthereumClient) SubmitBlock(block *types.Block) error {
	// 创建简单的元数据，不包含DAC状态
	metadata := &BlockMetadata{
		Block: block,
	}

	auth, err := ec.getTransactOpts()
	if err != nil {
		return fmt.Errorf("failed to get transaction options: %v", err)
	}
	ec.statusChan <- TxStatusEvent{
		Block:  *metadata.Block,
		Status: types.TxStatusL1Submit,
	}

	// 将区块哈希和状态根转换为[32]byte
	blockHash := common.HexToHash(metadata.Block.Hash)
	stateRoot := common.HexToHash(metadata.Block.StateRoot)

	// 使用生成的合约方法（只提交block，不包含DAC数据）
	tx, err := ec.contract.SubmitBlock(auth, metadata.Block.Height, blockHash, stateRoot, [32]byte{}, [32]byte{})
	if err != nil {
		ec.statusChan <- TxStatusEvent{
			Block:  *metadata.Block,
			Status: types.TxStatusL1Failed,
		}
		return fmt.Errorf("failed to submit block: %v", err)
	}

	// 等待交易确认
	receipt, err := ec.waitForTransaction(tx.Hash())
	if err != nil {
		ec.statusChan <- TxStatusEvent{
			Block:  *metadata.Block,
			Status: types.TxStatusL1Failed,
		}
		return fmt.Errorf("failed to wait for transaction confirmation: %v", err)
	}

	if receipt.Status == 0 {
		ec.statusChan <- TxStatusEvent{
			Block:  *metadata.Block,
			Status: types.TxStatusL1Failed,
		}
		return fmt.Errorf("transaction failed")
	}
	ec.statusChan <- TxStatusEvent{
		Block:       *metadata.Block,
		Status:      types.TxStatusL1Confirmed,
		L1TxHash:    tx.Hash().String(),
		L1Timestamp: time.Now(),
	}
	return nil
}

// 添加用于提交带DAC状态的区块的方法
// SubmitBlockWithDAC 提交带DAC状态的区块到L1
func (ec *EthereumClient) SubmitBlockWithDAC(block *types.Block, accountRootHex, txRootHex string, proofs [][]byte) error {
	// 创建包含DAC状态的元数据
	metadata := &BlockMetadata{
		Block:          block,
		DACAccountRoot: accountRootHex,
		DACTxRoot:      txRootHex,
		DACProofs:      proofs,
	}

	auth, err := ec.getTransactOpts()
	if err != nil {
		return fmt.Errorf("failed to get transaction options: %v", err)
	}

	ec.statusChan <- TxStatusEvent{
		Block:  *metadata.Block,
		Status: types.TxStatusL1Submit,
	}

	// 将区块哈希和状态根转换为[32]byte
	blockHash := common.HexToHash(metadata.Block.Hash)
	stateRoot := common.HexToHash(metadata.Block.StateRoot)

	// 将十六进制字符串转换为字节数组
	accountRootBytes, err := hex.DecodeString(accountRootHex)
	if err != nil {
		return fmt.Errorf("解析账户根失败: %w", err)
	}

	txRootBytes, err := hex.DecodeString(txRootHex)
	if err != nil {
		return fmt.Errorf("解析交易根失败: %w", err)
	}

	// 确保字节数组转为32字节格式
	var accountRoot [32]byte
	var txRoot [32]byte

	// 正确处理长度问题 - 如果小于32字节，放在数组末尾
	if len(accountRootBytes) <= 32 {
		copy(accountRoot[32-len(accountRootBytes):], accountRootBytes)
	} else {
		// 如果超过32字节，取最后32字节
		copy(accountRoot[:], accountRootBytes[len(accountRootBytes)-32:])
	}

	if len(txRootBytes) <= 32 {
		copy(txRoot[32-len(txRootBytes):], txRootBytes)
	} else {
		copy(txRoot[:], txRootBytes[len(txRootBytes)-32:])
	}

	// 使用生成的合约方法
	tx, err := ec.contract.SubmitBlock(auth, metadata.Block.Height, blockHash, stateRoot, accountRoot, txRoot)
	if err != nil {
		ec.statusChan <- TxStatusEvent{
			Block:  *metadata.Block,
			Status: types.TxStatusL1Failed,
		}
		return fmt.Errorf("failed to submit block with DAC: %v", err)
	}

	// 等待交易确认
	receipt, err := ec.waitForTransaction(tx.Hash())
	if err != nil {
		ec.statusChan <- TxStatusEvent{
			Block:  *metadata.Block,
			Status: types.TxStatusL1Failed,
		}
		return fmt.Errorf("failed to wait for transaction confirmation: %v", err)
	}

	if receipt.Status == 0 {
		ec.statusChan <- TxStatusEvent{
			Block:  *metadata.Block,
			Status: types.TxStatusL1Failed,
		}
		return fmt.Errorf("transaction failed")
	}
	// 如果有证明，需要提交证明
	//if len(proofs) > 0 {
	//	// 等待区块提交确认后再提交证明
	//	receipt, err := ec.waitForTransaction(tx.Hash())
	//	if err != nil {
	//		ec.statusChan <- TxStatusEvent{
	//			Block:  *metadata.Block,
	//			Status: types.TxStatusL1Failed,
	//		}
	//		return fmt.Errorf("failed to wait for transaction confirmation: %v", err)
	//	}
	//
	//	if receipt.Status == 0 {
	//		ec.statusChan <- TxStatusEvent{
	//			Block:  *metadata.Block,
	//			Status: types.TxStatusL1Failed,
	//		}
	//		return fmt.Errorf("block transaction failed")
	//	}
	//
	//	// 使用新的交易选项提交证明
	//	authProof, err := ec.getTransactOpts()
	//	if err != nil {
	//		return fmt.Errorf("failed to get transaction options for proof submission: %v", err)
	//	}
	//
	//	// 调用合约方法提交证明
	//	proofTx, err := ec.contract.SubmitDACProof(authProof, metadata.Block.Height, accountRoot, txRoot, proofs)
	//	if err != nil {
	//		return fmt.Errorf("failed to submit DAC proof: %v", err)
	//	}
	//
	//	// 等待证明交易确认
	//	proofReceipt, err := ec.waitForTransaction(proofTx.Hash())
	//	if err != nil {
	//		return fmt.Errorf("failed to wait for proof transaction confirmation: %v", err)
	//	}
	//
	//	if proofReceipt.Status == 0 {
	//		return fmt.Errorf("proof transaction failed")
	//	}
	//
	//	fmt.Printf("DAC证明已提交到Layer 1: 区块高度=%d, 交易哈希=%s\n",
	//		metadata.Block.Height, proofTx.Hash().String())
	//}

	ec.statusChan <- TxStatusEvent{
		Block:       *metadata.Block,
		Status:      types.TxStatusL1Confirmed,
		L1TxHash:    tx.Hash().String(),
		L1Timestamp: time.Now(),
	}

	return nil
}

// GetRandomNumber 从L1合约获取随机数，并将其映射到uint64范围内
func (ec *EthereumClient) GetRandomNumber() (uint64, error) {
	// 调用合约的 getRandomNumber 方法
	randomBig, err := ec.contract.GetRandomNumber(&bind.CallOpts{})
	if err != nil {
		return 0, fmt.Errorf("failed to get random number from contract: %v", err)
	}

	// 将大数映射到uint64范围内
	// 使用取模运算将大数映射到uint64的范围
	maxUint64 := new(big.Int).SetUint64(^uint64(0))
	mappedNumber := new(big.Int).Mod(randomBig, maxUint64)

	return mappedNumber.Uint64(), nil
}

// GetFullRandomNumber 获取完整的随机数（big.Int）
func (ec *EthereumClient) GetFullRandomNumber() (*big.Int, error) {
	return ec.contract.GetRandomNumber(&bind.CallOpts{})
}

// WatchRandomNumberUpdated 监听随机数更新事件
func (ec *EthereumClient) WatchRandomNumberUpdated(sink chan<- *EthereumRandomNumberUpdated) (event.Subscription, error) {
	return ec.contract.WatchRandomNumberUpdated(&bind.WatchOpts{}, sink)
}

// FilterRandomNumberUpdated 过滤随机数更新事件
func (ec *EthereumClient) FilterRandomNumberUpdated(opts *bind.FilterOpts) ([]*EthereumRandomNumberUpdated, error) {
	iterator, err := ec.contract.FilterRandomNumberUpdated(opts)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var events []*EthereumRandomNumberUpdated
	for iterator.Next() {
		events = append(events, iterator.Event)
	}

	return events, iterator.Error()
}

// waitForTransaction 等待交易确认
func (ec *EthereumClient) waitForTransaction(txHash common.Hash) (*etype.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	for {
		receipt, err := ec.client.TransactionReceipt(ctx, txHash)
		if err != nil {
			if err == ethereum.NotFound {
				time.Sleep(time.Second)
				continue
			}
			return nil, err
		}

		// 等待足够的确认数
		if receipt.BlockNumber != nil {
			currentBlock, err := ec.client.BlockNumber(ctx)
			if err != nil {
				return nil, err
			}

			confirmations := currentBlock - receipt.BlockNumber.Uint64()
			if confirmations >= ec.config.ConfirmBlocks {
				return receipt, nil
			}
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for transaction confirmation")
		case <-time.After(time.Second):
			continue
		}
	}
}

// GetBalance 获取账户余额
func (ec *EthereumClient) GetBalance(address common.Address) (*big.Int, error) {
	return ec.client.BalanceAt(context.Background(), address, nil)
}

// GetBlockNumber 获取当前区块高度
func (ec *EthereumClient) GetBlockNumber() (uint64, error) {
	return ec.client.BlockNumber(context.Background())
}

// 获取DAC最新状态根并转换为字符串
func (ec *EthereumClient) GetLatestDACRootsAsString() (acRootString string, txRootString string, err error) {
	acRoot, txRoot, err := ec.contract.GetLatestDACRoots(&bind.CallOpts{})
	if err != nil {
		return "", "", fmt.Errorf("获取DAC状态根失败: %v", err)
	}

	// 将[32]byte转换为字符串（不带0x前缀）
	acRootString = common.Bytes2Hex(acRoot[:])
	txRootString = common.Bytes2Hex(txRoot[:])

	return acRootString, txRootString, nil
}

//// 如果你需要带0x前缀的字符串
//func (ec *EthereumClient) GetLatestDACRootsAsHexString() (string, string, error) {
//	acRoot, txRoot, err := ec.contract.GetLatestDACRoots(&bind.CallOpts{})
//	if err != nil {
//		return "", "", fmt.Errorf("获取DAC状态根失败: %v", err)
//	}
//
//	// 将[32]byte转换为带0x前缀的十六进制字符串
//	acRootString := common.BytesToHash(acRoot[:]).Hex()
//	txRootString := common.BytesToHash(txRoot[:]).Hex()
//
//	return acRootString, txRootString, nil
//}

//// 获取DAC最新状态根并转换为带0x前缀的字符串
//func (ec *EthereumClient) GetLatestDACRootsAsHexutil() (string, string, error) {
//	acRoot, txRoot, err := ec.contract.GetLatestDACRoots(&bind.CallOpts{})
//	if err != nil {
//		return "", "", fmt.Errorf("获取DAC状态根失败: %v", err)
//	}
//
//	// 使用hexutil.Encode自动添加0x前缀
//	acRootString := hexutil.Encode(acRoot[:])
//	txRootString := hexutil.Encode(txRoot[:])
//
//	return acRootString, txRootString, nil
//}
//
//// 获取DAC最新状态根并转换为普通字符串（无0x前缀）
//func (ec *EthereumClient) GetLatestDACRootsWithEncodingHex() (string, string, error) {
//	acRoot, txRoot, err := ec.contract.GetLatestDACRoots(&bind.CallOpts{})
//	if err != nil {
//		return "", "", fmt.Errorf("获取DAC状态根失败: %v", err)
//	}
//
//	// 使用标准库encoding/hex
//	acRootString := hex.EncodeToString(acRoot[:])
//	txRootString := hex.EncodeToString(txRoot[:])
//
//	return acRootString, txRootString, nil
//}

// Close 关闭客户端连接
func (ec *EthereumClient) Close() {
	ec.client.Close()
}
