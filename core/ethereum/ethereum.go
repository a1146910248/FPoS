package ethereum

import (
	"FPoS/types"
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	etype "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
	"time"
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
	client   *ethclient.Client
	chainID  *big.Int
	contract *Ethereum
	config   *EthereumConfig
	privKey  *ecdsa.PrivateKey
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
		client:   client,
		chainID:  chainID,
		contract: contract,
		config:   config,
		privKey:  privKey,
	}, nil
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
	auth, err := ec.getTransactOpts()
	if err != nil {
		return fmt.Errorf("failed to get transaction options: %v", err)
	}

	// 将区块哈希和状态根转换为[32]byte
	blockHash := common.HexToHash(block.Hash)
	stateRoot := common.HexToHash(block.StateRoot)

	// 使用生成的合约方法
	tx, err := ec.contract.SubmitBlock(auth, block.Height, blockHash, stateRoot)
	if err != nil {
		return fmt.Errorf("failed to submit block: %v", err)
	}

	// 等待交易确认
	receipt, err := ec.waitForTransaction(tx.Hash())
	if err != nil {
		return fmt.Errorf("failed to wait for transaction confirmation: %v", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction failed")
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

// Close 关闭客户端连接
func (ec *EthereumClient) Close() {
	ec.client.Close()
}
