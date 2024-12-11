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

// EthereumMetaData contains all meta data concerning the Ethereum contract.
var EthereumMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"height\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"BlockSubmitted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"name\":\"blocks\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"height\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"latestHeight\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"height\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"}],\"name\":\"submitBlock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052348015600e575f5ffd5b506105588061001c5f395ff3fe608060405234801561000f575f5ffd5b506004361061003f575f3560e01c8063690d7a021461004357806387bbb63314610076578063e405bbc314610092575b5f5ffd5b61005d600480360381019061005891906102cd565b6100b0565b60405161006d9493929190610337565b60405180910390f35b610090600480360381019061008b91906103a4565b6100ee565b005b61009a610273565b6040516100a791906103f4565b60405180910390f35b5f602052805f5260405f205f91509050805f015f9054906101000a900467ffffffffffffffff16908060010154908060020154908060030154905084565b6001805f9054906101000a900467ffffffffffffffff1661010f919061043a565b67ffffffffffffffff168367ffffffffffffffff1614610164576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161015b906104cf565b60405180910390fd5b60405180608001604052808467ffffffffffffffff168152602001838152602001828152602001428152505f5f8567ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f820151815f015f6101000a81548167ffffffffffffffff021916908367ffffffffffffffff1602179055506020820151816001015560408201518160020155606082015181600301559050508260015f6101000a81548167ffffffffffffffff021916908367ffffffffffffffff1602179055508267ffffffffffffffff167f5e2237be8935308b23e4831f5316c6e809ea2b183d86bbcf405427329916c9d1838342604051610266939291906104ed565b60405180910390a2505050565b60015f9054906101000a900467ffffffffffffffff1681565b5f5ffd5b5f67ffffffffffffffff82169050919050565b6102ac81610290565b81146102b6575f5ffd5b50565b5f813590506102c7816102a3565b92915050565b5f602082840312156102e2576102e161028c565b5b5f6102ef848285016102b9565b91505092915050565b61030181610290565b82525050565b5f819050919050565b61031981610307565b82525050565b5f819050919050565b6103318161031f565b82525050565b5f60808201905061034a5f8301876102f8565b6103576020830186610310565b6103646040830185610310565b6103716060830184610328565b95945050505050565b61038381610307565b811461038d575f5ffd5b50565b5f8135905061039e8161037a565b92915050565b5f5f5f606084860312156103bb576103ba61028c565b5b5f6103c8868287016102b9565b93505060206103d986828701610390565b92505060406103ea86828701610390565b9150509250925092565b5f6020820190506104075f8301846102f8565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61044482610290565b915061044f83610290565b9250828201905067ffffffffffffffff81111561046f5761046e61040d565b5b92915050565b5f82825260208201905092915050565b7f496e76616c696420626c6f636b206865696768740000000000000000000000005f82015250565b5f6104b9601483610475565b91506104c482610485565b602082019050919050565b5f6020820190508181035f8301526104e6816104ad565b9050919050565b5f6060820190506105005f830186610310565b61050d6020830185610310565b61051a6040830184610328565b94935050505056fea264697066735822122056ee58c8a865038a125f51444eaac527a54995d6e76204636d8a7b1384697cc164736f6c634300081c0033",
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
