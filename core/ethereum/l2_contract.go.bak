// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ethereum

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// L2ContractBlockInfo is an auto generated low-level Go binding around an user-defined struct.
type L2ContractBlockInfo struct {
	Height    uint64
	BlockHash [32]byte
	StateRoot [32]byte
	Timestamp *big.Int
}

// EthereumMetaData contains all meta data concerning the Ethereum contract.
var EthereumMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"height\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"BlockSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"oldValue\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"blockHeight\",\"type\":\"uint64\"}],\"name\":\"RandomNumberUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"lastHeight\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"StateReset\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"name\":\"blocks\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"height\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentRandomNumber\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"fromHeight\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"toHeight\",\"type\":\"uint64\"}],\"name\":\"getBlockRange\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"height\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"internalType\":\"structL2Contract.BlockInfo[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRandomNumber\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"latestHeight\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"resetState\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"height\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"}],\"name\":\"submitBlock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f5ffd5b503360035f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550424360014361005f91906100cb565b4060405160200161007293929190610147565b604051602081830303815290604052805190602001205f1c600281905550610183565b5f819050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f6100d582610095565b91506100e083610095565b92508282039050818111156100f8576100f761009e565b5b92915050565b5f819050919050565b61011861011382610095565b6100fe565b82525050565b5f819050919050565b5f819050919050565b61014161013c8261011e565b610127565b82525050565b5f6101528286610107565b6020820191506101628285610107565b6020820191506101728284610130565b602082019150819050949350505050565b6112cd806101905f395ff3fe608060405234801561000f575f5ffd5b5060043610610091575f3560e01c80638da5cb5b116100645780638da5cb5b14610132578063dbdff2c114610150578063e405bbc31461016e578063f2fde38b1461018c578063f9e19fa5146101a857610091565b8063011a8f2d146100955780632334d212146100b3578063690d7a02146100e357806387bbb63314610116575b5f5ffd5b61009d6101b2565b6040516100aa91906109da565b60405180910390f35b6100cd60048036038101906100c89190610a34565b6101b8565b6040516100da9190610ba3565b60405180910390f35b6100fd60048036038101906100f89190610bc3565b6103e3565b60405161010d9493929190610c0c565b60405180910390f35b610130600480360381019061012b9190610c79565b610421565b005b61013a610623565b6040516101479190610d08565b60405180910390f35b610158610648565b60405161016591906109da565b60405180910390f35b610176610651565b6040516101839190610d21565b60405180910390f35b6101a660048036038101906101a19190610d64565b61066a565b005b6101b06107aa565b005b60025481565b60608167ffffffffffffffff168367ffffffffffffffff161115610211576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161020890610de9565b60405180910390fd5b60015f9054906101000a900467ffffffffffffffff1667ffffffffffffffff168267ffffffffffffffff16111561027d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161027490610e51565b60405180910390fd5b5f6001848461028c9190610e9c565b6102969190610ed7565b90505f8167ffffffffffffffff1667ffffffffffffffff8111156102bd576102bc610f12565b5b6040519080825280602002602001820160405280156102f657816020015b6102e3610994565b8152602001906001900390816102db5790505b5090505f5f90505b8267ffffffffffffffff168167ffffffffffffffff1610156103d7575f5f82886103289190610ed7565b67ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f206040518060800160405290815f82015f9054906101000a900467ffffffffffffffff1667ffffffffffffffff1667ffffffffffffffff1681526020016001820154815260200160028201548152602001600382015481525050828267ffffffffffffffff16815181106103bf576103be610f3f565b5b602002602001018190525080806001019150506102fe565b50809250505092915050565b5f602052805f5260405f205f91509050805f015f9054906101000a900467ffffffffffffffff16908060010154908060020154908060030154905084565b6001805f9054906101000a900467ffffffffffffffff166104429190610ed7565b67ffffffffffffffff168367ffffffffffffffff1614610497576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161048e90610fb6565b60405180910390fd5b60405180608001604052808467ffffffffffffffff168152602001838152602001828152602001428152505f5f8567ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f820151815f015f6101000a81548167ffffffffffffffff021916908367ffffffffffffffff1602179055506020820151816001015560408201518160020155606082015181600301559050508260015f6101000a81548167ffffffffffffffff021916908367ffffffffffffffff1602179055505f60025490508282424360025460405160200161057c959493929190611014565b604051602081830303815290604052805190602001205f1c6002819055508367ffffffffffffffff167f5e2237be8935308b23e4831f5316c6e809ea2b183d86bbcf405427329916c9d18484426040516105d893929190611072565b60405180910390a27f67e80fae26f6e2110364a6c27237dd1440736f63389c36175e936ec2194e123c8160025486604051610615939291906110a7565b60405180910390a150505050565b60035f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b5f600254905090565b60015f9054906101000a900467ffffffffffffffff1681565b60035f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146106f9576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106f09061114c565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603610767576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161075e906111b4565b60405180910390fd5b8060035f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050565b60035f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610839576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108309061114c565b60405180910390fd5b5f60015f9054906101000a900467ffffffffffffffff1690505f5f90505b60015f9054906101000a900467ffffffffffffffff1667ffffffffffffffff168167ffffffffffffffff16116108ef575f5f8267ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f5f82015f6101000a81549067ffffffffffffffff0219169055600182015f9055600282015f9055600382015f9055505080806108e7906111d2565b915050610857565b505f60015f6101000a81548167ffffffffffffffff021916908367ffffffffffffffff16021790555042436001436109279190611201565b4060405160200161093a93929190611234565b604051602081830303815290604052805190602001205f1c6002819055507f0ab0fce6606ff44d9ab55e12700809b531f2d93f2557609ef19e74349231243f8142604051610989929190611270565b60405180910390a150565b60405180608001604052805f67ffffffffffffffff1681526020015f81526020015f81526020015f81525090565b5f819050919050565b6109d4816109c2565b82525050565b5f6020820190506109ed5f8301846109cb565b92915050565b5f5ffd5b5f67ffffffffffffffff82169050919050565b610a13816109f7565b8114610a1d575f5ffd5b50565b5f81359050610a2e81610a0a565b92915050565b5f5f60408385031215610a4a57610a496109f3565b5b5f610a5785828601610a20565b9250506020610a6885828601610a20565b9150509250929050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b610aa4816109f7565b82525050565b5f819050919050565b610abc81610aaa565b82525050565b610acb816109c2565b82525050565b608082015f820151610ae55f850182610a9b565b506020820151610af86020850182610ab3565b506040820151610b0b6040850182610ab3565b506060820151610b1e6060850182610ac2565b50505050565b5f610b2f8383610ad1565b60808301905092915050565b5f602082019050919050565b5f610b5182610a72565b610b5b8185610a7c565b9350610b6683610a8c565b805f5b83811015610b96578151610b7d8882610b24565b9750610b8883610b3b565b925050600181019050610b69565b5085935050505092915050565b5f6020820190508181035f830152610bbb8184610b47565b905092915050565b5f60208284031215610bd857610bd76109f3565b5b5f610be584828501610a20565b91505092915050565b610bf7816109f7565b82525050565b610c0681610aaa565b82525050565b5f608082019050610c1f5f830187610bee565b610c2c6020830186610bfd565b610c396040830185610bfd565b610c4660608301846109cb565b95945050505050565b610c5881610aaa565b8114610c62575f5ffd5b50565b5f81359050610c7381610c4f565b92915050565b5f5f5f60608486031215610c9057610c8f6109f3565b5b5f610c9d86828701610a20565b9350506020610cae86828701610c65565b9250506040610cbf86828701610c65565b9150509250925092565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610cf282610cc9565b9050919050565b610d0281610ce8565b82525050565b5f602082019050610d1b5f830184610cf9565b92915050565b5f602082019050610d345f830184610bee565b92915050565b610d4381610ce8565b8114610d4d575f5ffd5b50565b5f81359050610d5e81610d3a565b92915050565b5f60208284031215610d7957610d786109f3565b5b5f610d8684828501610d50565b91505092915050565b5f82825260208201905092915050565b7f496e76616c6964206865696768742072616e67650000000000000000000000005f82015250565b5f610dd3601483610d8f565b9150610dde82610d9f565b602082019050919050565b5f6020820190508181035f830152610e0081610dc7565b9050919050565b7f486569676874206f7574206f662072616e6765000000000000000000000000005f82015250565b5f610e3b601383610d8f565b9150610e4682610e07565b602082019050919050565b5f6020820190508181035f830152610e6881610e2f565b9050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f610ea6826109f7565b9150610eb1836109f7565b9250828203905067ffffffffffffffff811115610ed157610ed0610e6f565b5b92915050565b5f610ee1826109f7565b9150610eec836109f7565b9250828201905067ffffffffffffffff811115610f0c57610f0b610e6f565b5b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b7f496e76616c696420626c6f636b206865696768740000000000000000000000005f82015250565b5f610fa0601483610d8f565b9150610fab82610f6c565b602082019050919050565b5f6020820190508181035f830152610fcd81610f94565b9050919050565b5f819050919050565b610fee610fe982610aaa565b610fd4565b82525050565b5f819050919050565b61100e611009826109c2565b610ff4565b82525050565b5f61101f8288610fdd565b60208201915061102f8287610fdd565b60208201915061103f8286610ffd565b60208201915061104f8285610ffd565b60208201915061105f8284610ffd565b6020820191508190509695505050505050565b5f6060820190506110855f830186610bfd565b6110926020830185610bfd565b61109f60408301846109cb565b949350505050565b5f6060820190506110ba5f8301866109cb565b6110c760208301856109cb565b6110d46040830184610bee565b949350505050565b7f4f6e6c79206f776e65722063616e2063616c6c20746869732066756e6374696f5f8201527f6e00000000000000000000000000000000000000000000000000000000000000602082015250565b5f611136602183610d8f565b9150611141826110dc565b604082019050919050565b5f6020820190508181035f8301526111638161112a565b9050919050565b7f4e6577206f776e65722063616e6e6f74206265207a65726f20616464726573735f82015250565b5f61119e602083610d8f565b91506111a98261116a565b602082019050919050565b5f6020820190508181035f8301526111cb81611192565b9050919050565b5f6111dc826109f7565b915067ffffffffffffffff82036111f6576111f5610e6f565b5b600182019050919050565b5f61120b826109c2565b9150611216836109c2565b925082820390508181111561122e5761122d610e6f565b5b92915050565b5f61123f8286610ffd565b60208201915061124f8285610ffd565b60208201915061125f8284610fdd565b602082019150819050949350505050565b5f6040820190506112835f830185610bee565b61129060208301846109cb565b939250505056fea26469706673582212202bc7f40aeb1c74844d9833a86098d029fb137b84df59a2a5fc4ea17653ccae5e64736f6c634300081c0033",
}

// EthereumABI is the input ABI used to generate the binding from.
// Deprecated: Use EthereumMetaData.ABI instead.
var EthereumABI = EthereumMetaData.ABI

// EthereumBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use EthereumMetaData.Bin instead.
var EthereumBin = EthereumMetaData.Bin

// DeployEthereum deploys a new Ethereum contract, binding an instance of Ethereum to it.
func DeployEthereum(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Ethereum, error) {
	parsed, err := EthereumMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(EthereumBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Ethereum{EthereumCaller: EthereumCaller{contract: contract}, EthereumTransactor: EthereumTransactor{contract: contract}, EthereumFilterer: EthereumFilterer{contract: contract}}, nil
}

// Ethereum is an auto generated Go binding around an Ethereum contract.
type Ethereum struct {
	EthereumCaller     // Read-only binding to the contract
	EthereumTransactor // Write-only binding to the contract
	EthereumFilterer   // Log filterer for contract events
}

// EthereumCaller is an auto generated read-only Go binding around an Ethereum contract.
type EthereumCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EthereumTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EthereumFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EthereumSession struct {
	Contract     *Ethereum         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EthereumCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EthereumCallerSession struct {
	Contract *EthereumCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// EthereumTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EthereumTransactorSession struct {
	Contract     *EthereumTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// EthereumRaw is an auto generated low-level Go binding around an Ethereum contract.
type EthereumRaw struct {
	Contract *Ethereum // Generic contract binding to access the raw methods on
}

// EthereumCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EthereumCallerRaw struct {
	Contract *EthereumCaller // Generic read-only contract binding to access the raw methods on
}

// EthereumTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EthereumTransactorRaw struct {
	Contract *EthereumTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEthereum creates a new instance of Ethereum, bound to a specific deployed contract.
func NewEthereum(address common.Address, backend bind.ContractBackend) (*Ethereum, error) {
	contract, err := bindEthereum(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Ethereum{EthereumCaller: EthereumCaller{contract: contract}, EthereumTransactor: EthereumTransactor{contract: contract}, EthereumFilterer: EthereumFilterer{contract: contract}}, nil
}

// NewEthereumCaller creates a new read-only instance of Ethereum, bound to a specific deployed contract.
func NewEthereumCaller(address common.Address, caller bind.ContractCaller) (*EthereumCaller, error) {
	contract, err := bindEthereum(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumCaller{contract: contract}, nil
}

// NewEthereumTransactor creates a new write-only instance of Ethereum, bound to a specific deployed contract.
func NewEthereumTransactor(address common.Address, transactor bind.ContractTransactor) (*EthereumTransactor, error) {
	contract, err := bindEthereum(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumTransactor{contract: contract}, nil
}

// NewEthereumFilterer creates a new log filterer instance of Ethereum, bound to a specific deployed contract.
func NewEthereumFilterer(address common.Address, filterer bind.ContractFilterer) (*EthereumFilterer, error) {
	contract, err := bindEthereum(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EthereumFilterer{contract: contract}, nil
}

// bindEthereum binds a generic wrapper to an already deployed contract.
func bindEthereum(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := EthereumMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ethereum *EthereumRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ethereum.Contract.EthereumCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ethereum *EthereumRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ethereum.Contract.EthereumTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ethereum *EthereumRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ethereum.Contract.EthereumTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ethereum *EthereumCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ethereum.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ethereum *EthereumTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ethereum.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ethereum *EthereumTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ethereum.Contract.contract.Transact(opts, method, params...)
}

// Blocks is a free data retrieval call binding the contract method 0x690d7a02.
//
// Solidity: function blocks(uint64 ) view returns(uint64 height, bytes32 blockHash, bytes32 stateRoot, uint256 timestamp)
func (_Ethereum *EthereumCaller) Blocks(opts *bind.CallOpts, arg0 uint64) (struct {
	Height    uint64
	BlockHash [32]byte
	StateRoot [32]byte
	Timestamp *big.Int
}, error) {
	var out []interface{}
	err := _Ethereum.contract.Call(opts, &out, "blocks", arg0)

	outstruct := new(struct {
		Height    uint64
		BlockHash [32]byte
		StateRoot [32]byte
		Timestamp *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Height = *abi.ConvertType(out[0], new(uint64)).(*uint64)
	outstruct.BlockHash = *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)
	outstruct.StateRoot = *abi.ConvertType(out[2], new([32]byte)).(*[32]byte)
	outstruct.Timestamp = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Blocks is a free data retrieval call binding the contract method 0x690d7a02.
//
// Solidity: function blocks(uint64 ) view returns(uint64 height, bytes32 blockHash, bytes32 stateRoot, uint256 timestamp)
func (_Ethereum *EthereumSession) Blocks(arg0 uint64) (struct {
	Height    uint64
	BlockHash [32]byte
	StateRoot [32]byte
	Timestamp *big.Int
}, error) {
	return _Ethereum.Contract.Blocks(&_Ethereum.CallOpts, arg0)
}

// Blocks is a free data retrieval call binding the contract method 0x690d7a02.
//
// Solidity: function blocks(uint64 ) view returns(uint64 height, bytes32 blockHash, bytes32 stateRoot, uint256 timestamp)
func (_Ethereum *EthereumCallerSession) Blocks(arg0 uint64) (struct {
	Height    uint64
	BlockHash [32]byte
	StateRoot [32]byte
	Timestamp *big.Int
}, error) {
	return _Ethereum.Contract.Blocks(&_Ethereum.CallOpts, arg0)
}

// CurrentRandomNumber is a free data retrieval call binding the contract method 0x011a8f2d.
//
// Solidity: function currentRandomNumber() view returns(uint256)
func (_Ethereum *EthereumCaller) CurrentRandomNumber(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Ethereum.contract.Call(opts, &out, "currentRandomNumber")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CurrentRandomNumber is a free data retrieval call binding the contract method 0x011a8f2d.
//
// Solidity: function currentRandomNumber() view returns(uint256)
func (_Ethereum *EthereumSession) CurrentRandomNumber() (*big.Int, error) {
	return _Ethereum.Contract.CurrentRandomNumber(&_Ethereum.CallOpts)
}

// CurrentRandomNumber is a free data retrieval call binding the contract method 0x011a8f2d.
//
// Solidity: function currentRandomNumber() view returns(uint256)
func (_Ethereum *EthereumCallerSession) CurrentRandomNumber() (*big.Int, error) {
	return _Ethereum.Contract.CurrentRandomNumber(&_Ethereum.CallOpts)
}

// GetBlockRange is a free data retrieval call binding the contract method 0x2334d212.
//
// Solidity: function getBlockRange(uint64 fromHeight, uint64 toHeight) view returns((uint64,bytes32,bytes32,uint256)[])
func (_Ethereum *EthereumCaller) GetBlockRange(opts *bind.CallOpts, fromHeight uint64, toHeight uint64) ([]L2ContractBlockInfo, error) {
	var out []interface{}
	err := _Ethereum.contract.Call(opts, &out, "getBlockRange", fromHeight, toHeight)

	if err != nil {
		return *new([]L2ContractBlockInfo), err
	}

	out0 := *abi.ConvertType(out[0], new([]L2ContractBlockInfo)).(*[]L2ContractBlockInfo)

	return out0, err

}

// GetBlockRange is a free data retrieval call binding the contract method 0x2334d212.
//
// Solidity: function getBlockRange(uint64 fromHeight, uint64 toHeight) view returns((uint64,bytes32,bytes32,uint256)[])
func (_Ethereum *EthereumSession) GetBlockRange(fromHeight uint64, toHeight uint64) ([]L2ContractBlockInfo, error) {
	return _Ethereum.Contract.GetBlockRange(&_Ethereum.CallOpts, fromHeight, toHeight)
}

// GetBlockRange is a free data retrieval call binding the contract method 0x2334d212.
//
// Solidity: function getBlockRange(uint64 fromHeight, uint64 toHeight) view returns((uint64,bytes32,bytes32,uint256)[])
func (_Ethereum *EthereumCallerSession) GetBlockRange(fromHeight uint64, toHeight uint64) ([]L2ContractBlockInfo, error) {
	return _Ethereum.Contract.GetBlockRange(&_Ethereum.CallOpts, fromHeight, toHeight)
}

// GetRandomNumber is a free data retrieval call binding the contract method 0xdbdff2c1.
//
// Solidity: function getRandomNumber() view returns(uint256)
func (_Ethereum *EthereumCaller) GetRandomNumber(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Ethereum.contract.Call(opts, &out, "getRandomNumber")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRandomNumber is a free data retrieval call binding the contract method 0xdbdff2c1.
//
// Solidity: function getRandomNumber() view returns(uint256)
func (_Ethereum *EthereumSession) GetRandomNumber() (*big.Int, error) {
	return _Ethereum.Contract.GetRandomNumber(&_Ethereum.CallOpts)
}

// GetRandomNumber is a free data retrieval call binding the contract method 0xdbdff2c1.
//
// Solidity: function getRandomNumber() view returns(uint256)
func (_Ethereum *EthereumCallerSession) GetRandomNumber() (*big.Int, error) {
	return _Ethereum.Contract.GetRandomNumber(&_Ethereum.CallOpts)
}

// LatestHeight is a free data retrieval call binding the contract method 0xe405bbc3.
//
// Solidity: function latestHeight() view returns(uint64)
func (_Ethereum *EthereumCaller) LatestHeight(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Ethereum.contract.Call(opts, &out, "latestHeight")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// LatestHeight is a free data retrieval call binding the contract method 0xe405bbc3.
//
// Solidity: function latestHeight() view returns(uint64)
func (_Ethereum *EthereumSession) LatestHeight() (uint64, error) {
	return _Ethereum.Contract.LatestHeight(&_Ethereum.CallOpts)
}

// LatestHeight is a free data retrieval call binding the contract method 0xe405bbc3.
//
// Solidity: function latestHeight() view returns(uint64)
func (_Ethereum *EthereumCallerSession) LatestHeight() (uint64, error) {
	return _Ethereum.Contract.LatestHeight(&_Ethereum.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ethereum *EthereumCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Ethereum.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ethereum *EthereumSession) Owner() (common.Address, error) {
	return _Ethereum.Contract.Owner(&_Ethereum.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ethereum *EthereumCallerSession) Owner() (common.Address, error) {
	return _Ethereum.Contract.Owner(&_Ethereum.CallOpts)
}

// ResetState is a paid mutator transaction binding the contract method 0xf9e19fa5.
//
// Solidity: function resetState() returns()
func (_Ethereum *EthereumTransactor) ResetState(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ethereum.contract.Transact(opts, "resetState")
}

// ResetState is a paid mutator transaction binding the contract method 0xf9e19fa5.
//
// Solidity: function resetState() returns()
func (_Ethereum *EthereumSession) ResetState() (*types.Transaction, error) {
	return _Ethereum.Contract.ResetState(&_Ethereum.TransactOpts)
}

// ResetState is a paid mutator transaction binding the contract method 0xf9e19fa5.
//
// Solidity: function resetState() returns()
func (_Ethereum *EthereumTransactorSession) ResetState() (*types.Transaction, error) {
	return _Ethereum.Contract.ResetState(&_Ethereum.TransactOpts)
}

// SubmitBlock is a paid mutator transaction binding the contract method 0x87bbb633.
//
// Solidity: function submitBlock(uint64 height, bytes32 blockHash, bytes32 stateRoot) returns()
func (_Ethereum *EthereumTransactor) SubmitBlock(opts *bind.TransactOpts, height uint64, blockHash [32]byte, stateRoot [32]byte) (*types.Transaction, error) {
	return _Ethereum.contract.Transact(opts, "submitBlock", height, blockHash, stateRoot)
}

// SubmitBlock is a paid mutator transaction binding the contract method 0x87bbb633.
//
// Solidity: function submitBlock(uint64 height, bytes32 blockHash, bytes32 stateRoot) returns()
func (_Ethereum *EthereumSession) SubmitBlock(height uint64, blockHash [32]byte, stateRoot [32]byte) (*types.Transaction, error) {
	return _Ethereum.Contract.SubmitBlock(&_Ethereum.TransactOpts, height, blockHash, stateRoot)
}

// SubmitBlock is a paid mutator transaction binding the contract method 0x87bbb633.
//
// Solidity: function submitBlock(uint64 height, bytes32 blockHash, bytes32 stateRoot) returns()
func (_Ethereum *EthereumTransactorSession) SubmitBlock(height uint64, blockHash [32]byte, stateRoot [32]byte) (*types.Transaction, error) {
	return _Ethereum.Contract.SubmitBlock(&_Ethereum.TransactOpts, height, blockHash, stateRoot)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ethereum *EthereumTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Ethereum.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ethereum *EthereumSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Ethereum.Contract.TransferOwnership(&_Ethereum.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ethereum *EthereumTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Ethereum.Contract.TransferOwnership(&_Ethereum.TransactOpts, newOwner)
}

// EthereumBlockSubmittedIterator is returned from FilterBlockSubmitted and is used to iterate over the raw logs and unpacked data for BlockSubmitted events raised by the Ethereum contract.
type EthereumBlockSubmittedIterator struct {
	Event *EthereumBlockSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EthereumBlockSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumBlockSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EthereumBlockSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EthereumBlockSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumBlockSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumBlockSubmitted represents a BlockSubmitted event raised by the Ethereum contract.
type EthereumBlockSubmitted struct {
	Height    uint64
	BlockHash [32]byte
	StateRoot [32]byte
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterBlockSubmitted is a free log retrieval operation binding the contract event 0x5e2237be8935308b23e4831f5316c6e809ea2b183d86bbcf405427329916c9d1.
//
// Solidity: event BlockSubmitted(uint64 indexed height, bytes32 blockHash, bytes32 stateRoot, uint256 timestamp)
func (_Ethereum *EthereumFilterer) FilterBlockSubmitted(opts *bind.FilterOpts, height []uint64) (*EthereumBlockSubmittedIterator, error) {

	var heightRule []interface{}
	for _, heightItem := range height {
		heightRule = append(heightRule, heightItem)
	}

	logs, sub, err := _Ethereum.contract.FilterLogs(opts, "BlockSubmitted", heightRule)
	if err != nil {
		return nil, err
	}
	return &EthereumBlockSubmittedIterator{contract: _Ethereum.contract, event: "BlockSubmitted", logs: logs, sub: sub}, nil
}

// WatchBlockSubmitted is a free log subscription operation binding the contract event 0x5e2237be8935308b23e4831f5316c6e809ea2b183d86bbcf405427329916c9d1.
//
// Solidity: event BlockSubmitted(uint64 indexed height, bytes32 blockHash, bytes32 stateRoot, uint256 timestamp)
func (_Ethereum *EthereumFilterer) WatchBlockSubmitted(opts *bind.WatchOpts, sink chan<- *EthereumBlockSubmitted, height []uint64) (event.Subscription, error) {

	var heightRule []interface{}
	for _, heightItem := range height {
		heightRule = append(heightRule, heightItem)
	}

	logs, sub, err := _Ethereum.contract.WatchLogs(opts, "BlockSubmitted", heightRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumBlockSubmitted)
				if err := _Ethereum.contract.UnpackLog(event, "BlockSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBlockSubmitted is a log parse operation binding the contract event 0x5e2237be8935308b23e4831f5316c6e809ea2b183d86bbcf405427329916c9d1.
//
// Solidity: event BlockSubmitted(uint64 indexed height, bytes32 blockHash, bytes32 stateRoot, uint256 timestamp)
func (_Ethereum *EthereumFilterer) ParseBlockSubmitted(log types.Log) (*EthereumBlockSubmitted, error) {
	event := new(EthereumBlockSubmitted)
	if err := _Ethereum.contract.UnpackLog(event, "BlockSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EthereumRandomNumberUpdatedIterator is returned from FilterRandomNumberUpdated and is used to iterate over the raw logs and unpacked data for RandomNumberUpdated events raised by the Ethereum contract.
type EthereumRandomNumberUpdatedIterator struct {
	Event *EthereumRandomNumberUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EthereumRandomNumberUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumRandomNumberUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EthereumRandomNumberUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EthereumRandomNumberUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumRandomNumberUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumRandomNumberUpdated represents a RandomNumberUpdated event raised by the Ethereum contract.
type EthereumRandomNumberUpdated struct {
	OldValue    *big.Int
	NewValue    *big.Int
	BlockHeight uint64
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterRandomNumberUpdated is a free log retrieval operation binding the contract event 0x67e80fae26f6e2110364a6c27237dd1440736f63389c36175e936ec2194e123c.
//
// Solidity: event RandomNumberUpdated(uint256 oldValue, uint256 newValue, uint64 blockHeight)
func (_Ethereum *EthereumFilterer) FilterRandomNumberUpdated(opts *bind.FilterOpts) (*EthereumRandomNumberUpdatedIterator, error) {

	logs, sub, err := _Ethereum.contract.FilterLogs(opts, "RandomNumberUpdated")
	if err != nil {
		return nil, err
	}
	return &EthereumRandomNumberUpdatedIterator{contract: _Ethereum.contract, event: "RandomNumberUpdated", logs: logs, sub: sub}, nil
}

// WatchRandomNumberUpdated is a free log subscription operation binding the contract event 0x67e80fae26f6e2110364a6c27237dd1440736f63389c36175e936ec2194e123c.
//
// Solidity: event RandomNumberUpdated(uint256 oldValue, uint256 newValue, uint64 blockHeight)
func (_Ethereum *EthereumFilterer) WatchRandomNumberUpdated(opts *bind.WatchOpts, sink chan<- *EthereumRandomNumberUpdated) (event.Subscription, error) {

	logs, sub, err := _Ethereum.contract.WatchLogs(opts, "RandomNumberUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumRandomNumberUpdated)
				if err := _Ethereum.contract.UnpackLog(event, "RandomNumberUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRandomNumberUpdated is a log parse operation binding the contract event 0x67e80fae26f6e2110364a6c27237dd1440736f63389c36175e936ec2194e123c.
//
// Solidity: event RandomNumberUpdated(uint256 oldValue, uint256 newValue, uint64 blockHeight)
func (_Ethereum *EthereumFilterer) ParseRandomNumberUpdated(log types.Log) (*EthereumRandomNumberUpdated, error) {
	event := new(EthereumRandomNumberUpdated)
	if err := _Ethereum.contract.UnpackLog(event, "RandomNumberUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EthereumStateResetIterator is returned from FilterStateReset and is used to iterate over the raw logs and unpacked data for StateReset events raised by the Ethereum contract.
type EthereumStateResetIterator struct {
	Event *EthereumStateReset // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EthereumStateResetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumStateReset)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EthereumStateReset)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EthereumStateResetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumStateResetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumStateReset represents a StateReset event raised by the Ethereum contract.
type EthereumStateReset struct {
	LastHeight uint64
	Timestamp  *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterStateReset is a free log retrieval operation binding the contract event 0x0ab0fce6606ff44d9ab55e12700809b531f2d93f2557609ef19e74349231243f.
//
// Solidity: event StateReset(uint64 lastHeight, uint256 timestamp)
func (_Ethereum *EthereumFilterer) FilterStateReset(opts *bind.FilterOpts) (*EthereumStateResetIterator, error) {

	logs, sub, err := _Ethereum.contract.FilterLogs(opts, "StateReset")
	if err != nil {
		return nil, err
	}
	return &EthereumStateResetIterator{contract: _Ethereum.contract, event: "StateReset", logs: logs, sub: sub}, nil
}

// WatchStateReset is a free log subscription operation binding the contract event 0x0ab0fce6606ff44d9ab55e12700809b531f2d93f2557609ef19e74349231243f.
//
// Solidity: event StateReset(uint64 lastHeight, uint256 timestamp)
func (_Ethereum *EthereumFilterer) WatchStateReset(opts *bind.WatchOpts, sink chan<- *EthereumStateReset) (event.Subscription, error) {

	logs, sub, err := _Ethereum.contract.WatchLogs(opts, "StateReset")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumStateReset)
				if err := _Ethereum.contract.UnpackLog(event, "StateReset", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStateReset is a log parse operation binding the contract event 0x0ab0fce6606ff44d9ab55e12700809b531f2d93f2557609ef19e74349231243f.
//
// Solidity: event StateReset(uint64 lastHeight, uint256 timestamp)
func (_Ethereum *EthereumFilterer) ParseStateReset(log types.Log) (*EthereumStateReset, error) {
	event := new(EthereumStateReset)
	if err := _Ethereum.contract.UnpackLog(event, "StateReset", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
