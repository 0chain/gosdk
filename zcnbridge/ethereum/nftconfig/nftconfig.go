// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package nftconfig

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

// NFTConfigMetaData contains all meta data concerning the NFTConfig contract.
var NFTConfigMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"previous\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"updated\",\"type\":\"uint256\"}],\"name\":\"ConfigUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"setUint256\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"value\",\"type\":\"address\"}],\"name\":\"setAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"}],\"name\":\"getUint256\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"}],\"name\":\"getAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true}]",
}

// NFTConfigABI is the input ABI used to generate the binding from.
// Deprecated: Use NFTConfigMetaData.ABI instead.
var NFTConfigABI = NFTConfigMetaData.ABI

// NFTConfig is an auto generated Go binding around an Ethereum contract.
type NFTConfig struct {
	NFTConfigCaller     // Read-only binding to the contract
	NFTConfigTransactor // Write-only binding to the contract
	NFTConfigFilterer   // Log filterer for contract events
}

// NFTConfigCaller is an auto generated read-only Go binding around an Ethereum contract.
type NFTConfigCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NFTConfigTransactor is an auto generated write-only Go binding around an Ethereum contract.
type NFTConfigTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NFTConfigFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type NFTConfigFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NFTConfigSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type NFTConfigSession struct {
	Contract     *NFTConfig        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// NFTConfigCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type NFTConfigCallerSession struct {
	Contract *NFTConfigCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// NFTConfigTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type NFTConfigTransactorSession struct {
	Contract     *NFTConfigTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// NFTConfigRaw is an auto generated low-level Go binding around an Ethereum contract.
type NFTConfigRaw struct {
	Contract *NFTConfig // Generic contract binding to access the raw methods on
}

// NFTConfigCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type NFTConfigCallerRaw struct {
	Contract *NFTConfigCaller // Generic read-only contract binding to access the raw methods on
}

// NFTConfigTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type NFTConfigTransactorRaw struct {
	Contract *NFTConfigTransactor // Generic write-only contract binding to access the raw methods on
}

// NewNFTConfig creates a new instance of NFTConfig, bound to a specific deployed contract.
func NewNFTConfig(address common.Address, backend bind.ContractBackend) (*NFTConfig, error) {
	contract, err := bindNFTConfig(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &NFTConfig{NFTConfigCaller: NFTConfigCaller{contract: contract}, NFTConfigTransactor: NFTConfigTransactor{contract: contract}, NFTConfigFilterer: NFTConfigFilterer{contract: contract}}, nil
}

// NewNFTConfigCaller creates a new read-only instance of NFTConfig, bound to a specific deployed contract.
func NewNFTConfigCaller(address common.Address, caller bind.ContractCaller) (*NFTConfigCaller, error) {
	contract, err := bindNFTConfig(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &NFTConfigCaller{contract: contract}, nil
}

// NewNFTConfigTransactor creates a new write-only instance of NFTConfig, bound to a specific deployed contract.
func NewNFTConfigTransactor(address common.Address, transactor bind.ContractTransactor) (*NFTConfigTransactor, error) {
	contract, err := bindNFTConfig(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &NFTConfigTransactor{contract: contract}, nil
}

// NewNFTConfigFilterer creates a new log filterer instance of NFTConfig, bound to a specific deployed contract.
func NewNFTConfigFilterer(address common.Address, filterer bind.ContractFilterer) (*NFTConfigFilterer, error) {
	contract, err := bindNFTConfig(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &NFTConfigFilterer{contract: contract}, nil
}

// bindNFTConfig binds a generic wrapper to an already deployed contract.
func bindNFTConfig(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(NFTConfigABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_NFTConfig *NFTConfigRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _NFTConfig.Contract.NFTConfigCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_NFTConfig *NFTConfigRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NFTConfig.Contract.NFTConfigTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_NFTConfig *NFTConfigRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _NFTConfig.Contract.NFTConfigTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_NFTConfig *NFTConfigCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _NFTConfig.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_NFTConfig *NFTConfigTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NFTConfig.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_NFTConfig *NFTConfigTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _NFTConfig.Contract.contract.Transact(opts, method, params...)
}

// GetAddress is a free data retrieval call binding the contract method 0x21f8a721.
//
// Solidity: function getAddress(bytes32 key) view returns(address)
func (_NFTConfig *NFTConfigCaller) GetAddress(opts *bind.CallOpts, key [32]byte) (common.Address, error) {
	var out []interface{}
	err := _NFTConfig.contract.Call(opts, &out, "getAddress", key)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAddress is a free data retrieval call binding the contract method 0x21f8a721.
//
// Solidity: function getAddress(bytes32 key) view returns(address)
func (_NFTConfig *NFTConfigSession) GetAddress(key [32]byte) (common.Address, error) {
	return _NFTConfig.Contract.GetAddress(&_NFTConfig.CallOpts, key)
}

// GetAddress is a free data retrieval call binding the contract method 0x21f8a721.
//
// Solidity: function getAddress(bytes32 key) view returns(address)
func (_NFTConfig *NFTConfigCallerSession) GetAddress(key [32]byte) (common.Address, error) {
	return _NFTConfig.Contract.GetAddress(&_NFTConfig.CallOpts, key)
}

// GetUint256 is a free data retrieval call binding the contract method 0x33598b00.
//
// Solidity: function getUint256(bytes32 key) view returns(uint256)
func (_NFTConfig *NFTConfigCaller) GetUint256(opts *bind.CallOpts, key [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _NFTConfig.contract.Call(opts, &out, "getUint256", key)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUint256 is a free data retrieval call binding the contract method 0x33598b00.
//
// Solidity: function getUint256(bytes32 key) view returns(uint256)
func (_NFTConfig *NFTConfigSession) GetUint256(key [32]byte) (*big.Int, error) {
	return _NFTConfig.Contract.GetUint256(&_NFTConfig.CallOpts, key)
}

// GetUint256 is a free data retrieval call binding the contract method 0x33598b00.
//
// Solidity: function getUint256(bytes32 key) view returns(uint256)
func (_NFTConfig *NFTConfigCallerSession) GetUint256(key [32]byte) (*big.Int, error) {
	return _NFTConfig.Contract.GetUint256(&_NFTConfig.CallOpts, key)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_NFTConfig *NFTConfigCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _NFTConfig.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_NFTConfig *NFTConfigSession) Owner() (common.Address, error) {
	return _NFTConfig.Contract.Owner(&_NFTConfig.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_NFTConfig *NFTConfigCallerSession) Owner() (common.Address, error) {
	return _NFTConfig.Contract.Owner(&_NFTConfig.CallOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_NFTConfig *NFTConfigTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NFTConfig.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_NFTConfig *NFTConfigSession) RenounceOwnership() (*types.Transaction, error) {
	return _NFTConfig.Contract.RenounceOwnership(&_NFTConfig.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_NFTConfig *NFTConfigTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _NFTConfig.Contract.RenounceOwnership(&_NFTConfig.TransactOpts)
}

// SetAddress is a paid mutator transaction binding the contract method 0xca446dd9.
//
// Solidity: function setAddress(bytes32 key, address value) returns()
func (_NFTConfig *NFTConfigTransactor) SetAddress(opts *bind.TransactOpts, key [32]byte, value common.Address) (*types.Transaction, error) {
	return _NFTConfig.contract.Transact(opts, "setAddress", key, value)
}

// SetAddress is a paid mutator transaction binding the contract method 0xca446dd9.
//
// Solidity: function setAddress(bytes32 key, address value) returns()
func (_NFTConfig *NFTConfigSession) SetAddress(key [32]byte, value common.Address) (*types.Transaction, error) {
	return _NFTConfig.Contract.SetAddress(&_NFTConfig.TransactOpts, key, value)
}

// SetAddress is a paid mutator transaction binding the contract method 0xca446dd9.
//
// Solidity: function setAddress(bytes32 key, address value) returns()
func (_NFTConfig *NFTConfigTransactorSession) SetAddress(key [32]byte, value common.Address) (*types.Transaction, error) {
	return _NFTConfig.Contract.SetAddress(&_NFTConfig.TransactOpts, key, value)
}

// SetUint256 is a paid mutator transaction binding the contract method 0x4f3029c2.
//
// Solidity: function setUint256(bytes32 key, uint256 value) returns()
func (_NFTConfig *NFTConfigTransactor) SetUint256(opts *bind.TransactOpts, key [32]byte, value *big.Int) (*types.Transaction, error) {
	return _NFTConfig.contract.Transact(opts, "setUint256", key, value)
}

// SetUint256 is a paid mutator transaction binding the contract method 0x4f3029c2.
//
// Solidity: function setUint256(bytes32 key, uint256 value) returns()
func (_NFTConfig *NFTConfigSession) SetUint256(key [32]byte, value *big.Int) (*types.Transaction, error) {
	return _NFTConfig.Contract.SetUint256(&_NFTConfig.TransactOpts, key, value)
}

// SetUint256 is a paid mutator transaction binding the contract method 0x4f3029c2.
//
// Solidity: function setUint256(bytes32 key, uint256 value) returns()
func (_NFTConfig *NFTConfigTransactorSession) SetUint256(key [32]byte, value *big.Int) (*types.Transaction, error) {
	return _NFTConfig.Contract.SetUint256(&_NFTConfig.TransactOpts, key, value)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_NFTConfig *NFTConfigTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _NFTConfig.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_NFTConfig *NFTConfigSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _NFTConfig.Contract.TransferOwnership(&_NFTConfig.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_NFTConfig *NFTConfigTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _NFTConfig.Contract.TransferOwnership(&_NFTConfig.TransactOpts, newOwner)
}

// NFTConfigConfigUpdatedIterator is returned from FilterConfigUpdated and is used to iterate over the raw logs and unpacked data for ConfigUpdated events raised by the NFTConfig contract.
type NFTConfigConfigUpdatedIterator struct {
	Event *NFTConfigConfigUpdated // Event containing the contract specifics and raw log

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
func (it *NFTConfigConfigUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NFTConfigConfigUpdated)
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
		it.Event = new(NFTConfigConfigUpdated)
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
func (it *NFTConfigConfigUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NFTConfigConfigUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NFTConfigConfigUpdated represents a ConfigUpdated event raised by the NFTConfig contract.
type NFTConfigConfigUpdated struct {
	Key      [32]byte
	Previous *big.Int
	Updated  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterConfigUpdated is a free log retrieval operation binding the contract event 0xac2ccce3de9c0816ae772598f7f65fe69f9893b637f7c490497378cbb3ea043e.
//
// Solidity: event ConfigUpdated(bytes32 indexed key, uint256 previous, uint256 updated)
func (_NFTConfig *NFTConfigFilterer) FilterConfigUpdated(opts *bind.FilterOpts, key [][32]byte) (*NFTConfigConfigUpdatedIterator, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}

	logs, sub, err := _NFTConfig.contract.FilterLogs(opts, "ConfigUpdated", keyRule)
	if err != nil {
		return nil, err
	}
	return &NFTConfigConfigUpdatedIterator{contract: _NFTConfig.contract, event: "ConfigUpdated", logs: logs, sub: sub}, nil
}

// WatchConfigUpdated is a free log subscription operation binding the contract event 0xac2ccce3de9c0816ae772598f7f65fe69f9893b637f7c490497378cbb3ea043e.
//
// Solidity: event ConfigUpdated(bytes32 indexed key, uint256 previous, uint256 updated)
func (_NFTConfig *NFTConfigFilterer) WatchConfigUpdated(opts *bind.WatchOpts, sink chan<- *NFTConfigConfigUpdated, key [][32]byte) (event.Subscription, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}

	logs, sub, err := _NFTConfig.contract.WatchLogs(opts, "ConfigUpdated", keyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NFTConfigConfigUpdated)
				if err := _NFTConfig.contract.UnpackLog(event, "ConfigUpdated", log); err != nil {
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

// ParseConfigUpdated is a log parse operation binding the contract event 0xac2ccce3de9c0816ae772598f7f65fe69f9893b637f7c490497378cbb3ea043e.
//
// Solidity: event ConfigUpdated(bytes32 indexed key, uint256 previous, uint256 updated)
func (_NFTConfig *NFTConfigFilterer) ParseConfigUpdated(log types.Log) (*NFTConfigConfigUpdated, error) {
	event := new(NFTConfigConfigUpdated)
	if err := _NFTConfig.contract.UnpackLog(event, "ConfigUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NFTConfigOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the NFTConfig contract.
type NFTConfigOwnershipTransferredIterator struct {
	Event *NFTConfigOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *NFTConfigOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NFTConfigOwnershipTransferred)
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
		it.Event = new(NFTConfigOwnershipTransferred)
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
func (it *NFTConfigOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NFTConfigOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NFTConfigOwnershipTransferred represents a OwnershipTransferred event raised by the NFTConfig contract.
type NFTConfigOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_NFTConfig *NFTConfigFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*NFTConfigOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _NFTConfig.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &NFTConfigOwnershipTransferredIterator{contract: _NFTConfig.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_NFTConfig *NFTConfigFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *NFTConfigOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _NFTConfig.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NFTConfigOwnershipTransferred)
				if err := _NFTConfig.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_NFTConfig *NFTConfigFilterer) ParseOwnershipTransferred(log types.Log) (*NFTConfigOwnershipTransferred, error) {
	event := new(NFTConfigOwnershipTransferred)
	if err := _NFTConfig.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
