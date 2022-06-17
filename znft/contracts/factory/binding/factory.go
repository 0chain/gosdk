// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package binding

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
)

// BindingMetaData contains all meta data concerning the Binding contract.
var BindingMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"ModuleRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"TokenCreated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"uri\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"create\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"moduleRegistry\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"module\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"tokenList\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokenMapping\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// BindingABI is the input ABI used to generate the binding from.
// Deprecated: Use BindingMetaData.ABI instead.
var BindingABI = BindingMetaData.ABI

// Binding is an auto generated Go binding around an Ethereum contract.
type Binding struct {
	BindingCaller     // Read-only binding to the contract
	BindingTransactor // Write-only binding to the contract
	BindingFilterer   // Log filterer for contract events
}

// BindingCaller is an auto generated read-only Go binding around an Ethereum contract.
type BindingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BindingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BindingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BindingSession struct {
	Contract     *Binding          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BindingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BindingCallerSession struct {
	Contract *BindingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// BindingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BindingTransactorSession struct {
	Contract     *BindingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// BindingRaw is an auto generated low-level Go binding around an Ethereum contract.
type BindingRaw struct {
	Contract *Binding // Generic contract binding to access the raw methods on
}

// BindingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BindingCallerRaw struct {
	Contract *BindingCaller // Generic read-only contract binding to access the raw methods on
}

// BindingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BindingTransactorRaw struct {
	Contract *BindingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBinding creates a new instance of Binding, bound to a specific deployed contract.
func NewBinding(address common.Address, backend bind.ContractBackend) (*Binding, error) {
	contract, err := bindBinding(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Binding{BindingCaller: BindingCaller{contract: contract}, BindingTransactor: BindingTransactor{contract: contract}, BindingFilterer: BindingFilterer{contract: contract}}, nil
}

// NewBindingCaller creates a new read-only instance of Binding, bound to a specific deployed contract.
func NewBindingCaller(address common.Address, caller bind.ContractCaller) (*BindingCaller, error) {
	contract, err := bindBinding(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BindingCaller{contract: contract}, nil
}

// NewBindingTransactor creates a new write-only instance of Binding, bound to a specific deployed contract.
func NewBindingTransactor(address common.Address, transactor bind.ContractTransactor) (*BindingTransactor, error) {
	contract, err := bindBinding(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BindingTransactor{contract: contract}, nil
}

// NewBindingFilterer creates a new log filterer instance of Binding, bound to a specific deployed contract.
func NewBindingFilterer(address common.Address, filterer bind.ContractFilterer) (*BindingFilterer, error) {
	contract, err := bindBinding(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BindingFilterer{contract: contract}, nil
}

// bindBinding binds a generic wrapper to an already deployed contract.
func bindBinding(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BindingABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Binding *BindingRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Binding.Contract.BindingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Binding *BindingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Binding.Contract.BindingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Binding *BindingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Binding.Contract.BindingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Binding *BindingCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Binding.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Binding *BindingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Binding.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Binding *BindingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Binding.Contract.contract.Transact(opts, method, params...)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint256)
func (_Binding *BindingCaller) Count(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "count")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint256)
func (_Binding *BindingSession) Count() (*big.Int, error) {
	return _Binding.Contract.Count(&_Binding.CallOpts)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint256)
func (_Binding *BindingCallerSession) Count() (*big.Int, error) {
	return _Binding.Contract.Count(&_Binding.CallOpts)
}

// ModuleRegistry is a free data retrieval call binding the contract method 0x082a1051.
//
// Solidity: function moduleRegistry(address ) view returns(bool)
func (_Binding *BindingCaller) ModuleRegistry(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "moduleRegistry", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ModuleRegistry is a free data retrieval call binding the contract method 0x082a1051.
//
// Solidity: function moduleRegistry(address ) view returns(bool)
func (_Binding *BindingSession) ModuleRegistry(arg0 common.Address) (bool, error) {
	return _Binding.Contract.ModuleRegistry(&_Binding.CallOpts, arg0)
}

// ModuleRegistry is a free data retrieval call binding the contract method 0x082a1051.
//
// Solidity: function moduleRegistry(address ) view returns(bool)
func (_Binding *BindingCallerSession) ModuleRegistry(arg0 common.Address) (bool, error) {
	return _Binding.Contract.ModuleRegistry(&_Binding.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Binding *BindingCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Binding *BindingSession) Owner() (common.Address, error) {
	return _Binding.Contract.Owner(&_Binding.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Binding *BindingCallerSession) Owner() (common.Address, error) {
	return _Binding.Contract.Owner(&_Binding.CallOpts)
}

// TokenList is a free data retrieval call binding the contract method 0x9ead7222.
//
// Solidity: function tokenList(uint256 ) view returns(address)
func (_Binding *BindingCaller) TokenList(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "tokenList", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenList is a free data retrieval call binding the contract method 0x9ead7222.
//
// Solidity: function tokenList(uint256 ) view returns(address)
func (_Binding *BindingSession) TokenList(arg0 *big.Int) (common.Address, error) {
	return _Binding.Contract.TokenList(&_Binding.CallOpts, arg0)
}

// TokenList is a free data retrieval call binding the contract method 0x9ead7222.
//
// Solidity: function tokenList(uint256 ) view returns(address)
func (_Binding *BindingCallerSession) TokenList(arg0 *big.Int) (common.Address, error) {
	return _Binding.Contract.TokenList(&_Binding.CallOpts, arg0)
}

// TokenMapping is a free data retrieval call binding the contract method 0xba27f50b.
//
// Solidity: function tokenMapping(address ) view returns(bool)
func (_Binding *BindingCaller) TokenMapping(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "tokenMapping", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// TokenMapping is a free data retrieval call binding the contract method 0xba27f50b.
//
// Solidity: function tokenMapping(address ) view returns(bool)
func (_Binding *BindingSession) TokenMapping(arg0 common.Address) (bool, error) {
	return _Binding.Contract.TokenMapping(&_Binding.CallOpts, arg0)
}

// TokenMapping is a free data retrieval call binding the contract method 0xba27f50b.
//
// Solidity: function tokenMapping(address ) view returns(bool)
func (_Binding *BindingCallerSession) TokenMapping(arg0 common.Address) (bool, error) {
	return _Binding.Contract.TokenMapping(&_Binding.CallOpts, arg0)
}

// Create is a paid mutator transaction binding the contract method 0x9eeeaa75.
//
// Solidity: function create(address module, string name, string symbol, string uri, bytes data) returns(address)
func (_Binding *BindingTransactor) Create(opts *bind.TransactOpts, module common.Address, name string, symbol string, uri string, data []byte) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "create", module, name, symbol, uri, data)
}

// Create is a paid mutator transaction binding the contract method 0x9eeeaa75.
//
// Solidity: function create(address module, string name, string symbol, string uri, bytes data) returns(address)
func (_Binding *BindingSession) Create(module common.Address, name string, symbol string, uri string, data []byte) (*types.Transaction, error) {
	return _Binding.Contract.Create(&_Binding.TransactOpts, module, name, symbol, uri, data)
}

// Create is a paid mutator transaction binding the contract method 0x9eeeaa75.
//
// Solidity: function create(address module, string name, string symbol, string uri, bytes data) returns(address)
func (_Binding *BindingTransactorSession) Create(module common.Address, name string, symbol string, uri string, data []byte) (*types.Transaction, error) {
	return _Binding.Contract.Create(&_Binding.TransactOpts, module, name, symbol, uri, data)
}

// Register is a paid mutator transaction binding the contract method 0xab01b469.
//
// Solidity: function register(address module, bool status) returns()
func (_Binding *BindingTransactor) Register(opts *bind.TransactOpts, module common.Address, status bool) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "register", module, status)
}

// Register is a paid mutator transaction binding the contract method 0xab01b469.
//
// Solidity: function register(address module, bool status) returns()
func (_Binding *BindingSession) Register(module common.Address, status bool) (*types.Transaction, error) {
	return _Binding.Contract.Register(&_Binding.TransactOpts, module, status)
}

// Register is a paid mutator transaction binding the contract method 0xab01b469.
//
// Solidity: function register(address module, bool status) returns()
func (_Binding *BindingTransactorSession) Register(module common.Address, status bool) (*types.Transaction, error) {
	return _Binding.Contract.Register(&_Binding.TransactOpts, module, status)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Binding *BindingTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Binding *BindingSession) RenounceOwnership() (*types.Transaction, error) {
	return _Binding.Contract.RenounceOwnership(&_Binding.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Binding *BindingTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Binding.Contract.RenounceOwnership(&_Binding.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Binding *BindingTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Binding *BindingSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Binding.Contract.TransferOwnership(&_Binding.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Binding *BindingTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Binding.Contract.TransferOwnership(&_Binding.TransactOpts, newOwner)
}

// BindingModuleRegisteredIterator is returned from FilterModuleRegistered and is used to iterate over the raw logs and unpacked data for ModuleRegistered events raised by the Binding contract.
type BindingModuleRegisteredIterator struct {
	Event *BindingModuleRegistered // Event containing the contract specifics and raw log

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
func (it *BindingModuleRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingModuleRegistered)
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
		it.Event = new(BindingModuleRegistered)
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
func (it *BindingModuleRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingModuleRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingModuleRegistered represents a ModuleRegistered event raised by the Binding contract.
type BindingModuleRegistered struct {
	Module common.Address
	Status bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterModuleRegistered is a free log retrieval operation binding the contract event 0x30091f391c7fccd6f943ec02ae164c4197ea3e43a67edc2edf908f3efc371876.
//
// Solidity: event ModuleRegistered(address indexed module, bool status)
func (_Binding *BindingFilterer) FilterModuleRegistered(opts *bind.FilterOpts, module []common.Address) (*BindingModuleRegisteredIterator, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _Binding.contract.FilterLogs(opts, "ModuleRegistered", moduleRule)
	if err != nil {
		return nil, err
	}
	return &BindingModuleRegisteredIterator{contract: _Binding.contract, event: "ModuleRegistered", logs: logs, sub: sub}, nil
}

// WatchModuleRegistered is a free log subscription operation binding the contract event 0x30091f391c7fccd6f943ec02ae164c4197ea3e43a67edc2edf908f3efc371876.
//
// Solidity: event ModuleRegistered(address indexed module, bool status)
func (_Binding *BindingFilterer) WatchModuleRegistered(opts *bind.WatchOpts, sink chan<- *BindingModuleRegistered, module []common.Address) (event.Subscription, error) {

	var moduleRule []interface{}
	for _, moduleItem := range module {
		moduleRule = append(moduleRule, moduleItem)
	}

	logs, sub, err := _Binding.contract.WatchLogs(opts, "ModuleRegistered", moduleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingModuleRegistered)
				if err := _Binding.contract.UnpackLog(event, "ModuleRegistered", log); err != nil {
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

// ParseModuleRegistered is a log parse operation binding the contract event 0x30091f391c7fccd6f943ec02ae164c4197ea3e43a67edc2edf908f3efc371876.
//
// Solidity: event ModuleRegistered(address indexed module, bool status)
func (_Binding *BindingFilterer) ParseModuleRegistered(log types.Log) (*BindingModuleRegistered, error) {
	event := new(BindingModuleRegistered)
	if err := _Binding.contract.UnpackLog(event, "ModuleRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Binding contract.
type BindingOwnershipTransferredIterator struct {
	Event *BindingOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BindingOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingOwnershipTransferred)
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
		it.Event = new(BindingOwnershipTransferred)
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
func (it *BindingOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingOwnershipTransferred represents a OwnershipTransferred event raised by the Binding contract.
type BindingOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Binding *BindingFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BindingOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Binding.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BindingOwnershipTransferredIterator{contract: _Binding.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Binding *BindingFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BindingOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Binding.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingOwnershipTransferred)
				if err := _Binding.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Binding *BindingFilterer) ParseOwnershipTransferred(log types.Log) (*BindingOwnershipTransferred, error) {
	event := new(BindingOwnershipTransferred)
	if err := _Binding.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingTokenCreatedIterator is returned from FilterTokenCreated and is used to iterate over the raw logs and unpacked data for TokenCreated events raised by the Binding contract.
type BindingTokenCreatedIterator struct {
	Event *BindingTokenCreated // Event containing the contract specifics and raw log

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
func (it *BindingTokenCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingTokenCreated)
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
		it.Event = new(BindingTokenCreated)
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
func (it *BindingTokenCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingTokenCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingTokenCreated represents a TokenCreated event raised by the Binding contract.
type BindingTokenCreated struct {
	Owner common.Address
	Token common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTokenCreated is a free log retrieval operation binding the contract event 0xd5f9bdf12adf29dab0248c349842c3822d53ae2bb4f36352f301630d018c8139.
//
// Solidity: event TokenCreated(address indexed owner, address token)
func (_Binding *BindingFilterer) FilterTokenCreated(opts *bind.FilterOpts, owner []common.Address) (*BindingTokenCreatedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Binding.contract.FilterLogs(opts, "TokenCreated", ownerRule)
	if err != nil {
		return nil, err
	}
	return &BindingTokenCreatedIterator{contract: _Binding.contract, event: "TokenCreated", logs: logs, sub: sub}, nil
}

// WatchTokenCreated is a free log subscription operation binding the contract event 0xd5f9bdf12adf29dab0248c349842c3822d53ae2bb4f36352f301630d018c8139.
//
// Solidity: event TokenCreated(address indexed owner, address token)
func (_Binding *BindingFilterer) WatchTokenCreated(opts *bind.WatchOpts, sink chan<- *BindingTokenCreated, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Binding.contract.WatchLogs(opts, "TokenCreated", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingTokenCreated)
				if err := _Binding.contract.UnpackLog(event, "TokenCreated", log); err != nil {
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

// ParseTokenCreated is a log parse operation binding the contract event 0xd5f9bdf12adf29dab0248c349842c3822d53ae2bb4f36352f301630d018c8139.
//
// Solidity: event TokenCreated(address indexed owner, address token)
func (_Binding *BindingFilterer) ParseTokenCreated(log types.Log) (*BindingTokenCreated, error) {
	event := new(BindingTokenCreated)
	if err := _Binding.contract.UnpackLog(event, "TokenCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
