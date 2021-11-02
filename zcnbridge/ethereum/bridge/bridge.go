// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bridge

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

// AuthorizableMetaData contains all meta data concerning the Authorizable contract.
var AuthorizableMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousAuthorizers\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newAuthorizers\",\"type\":\"address\"}],\"name\":\"AuthorizersTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"authorizers\",\"outputs\":[{\"internalType\":\"contractIAuthorizers\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"56741b2c": "authorizers()",
		"8da5cb5b": "owner()",
		"715018a6": "renounceOwnership()",
		"f2fde38b": "transferOwnership(address)",
	},
}

// AuthorizableABI is the input ABI used to generate the binding from.
// Deprecated: Use AuthorizableMetaData.ABI instead.
var AuthorizableABI = AuthorizableMetaData.ABI

// Deprecated: Use AuthorizableMetaData.Sigs instead.
// AuthorizableFuncSigs maps the 4-byte function signature to its string representation.
var AuthorizableFuncSigs = AuthorizableMetaData.Sigs

// Authorizable is an auto generated Go binding around an Ethereum contract.
type Authorizable struct {
	AuthorizableCaller     // Read-only binding to the contract
	AuthorizableTransactor // Write-only binding to the contract
	AuthorizableFilterer   // Log filterer for contract events
}

// AuthorizableCaller is an auto generated read-only Go binding around an Ethereum contract.
type AuthorizableCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuthorizableTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AuthorizableTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuthorizableFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AuthorizableFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuthorizableSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AuthorizableSession struct {
	Contract     *Authorizable     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AuthorizableCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AuthorizableCallerSession struct {
	Contract *AuthorizableCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// AuthorizableTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AuthorizableTransactorSession struct {
	Contract     *AuthorizableTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// AuthorizableRaw is an auto generated low-level Go binding around an Ethereum contract.
type AuthorizableRaw struct {
	Contract *Authorizable // Generic contract binding to access the raw methods on
}

// AuthorizableCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AuthorizableCallerRaw struct {
	Contract *AuthorizableCaller // Generic read-only contract binding to access the raw methods on
}

// AuthorizableTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AuthorizableTransactorRaw struct {
	Contract *AuthorizableTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAuthorizable creates a new instance of Authorizable, bound to a specific deployed contract.
func NewAuthorizable(address common.Address, backend bind.ContractBackend) (*Authorizable, error) {
	contract, err := bindAuthorizable(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Authorizable{AuthorizableCaller: AuthorizableCaller{contract: contract}, AuthorizableTransactor: AuthorizableTransactor{contract: contract}, AuthorizableFilterer: AuthorizableFilterer{contract: contract}}, nil
}

// NewAuthorizableCaller creates a new read-only instance of Authorizable, bound to a specific deployed contract.
func NewAuthorizableCaller(address common.Address, caller bind.ContractCaller) (*AuthorizableCaller, error) {
	contract, err := bindAuthorizable(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AuthorizableCaller{contract: contract}, nil
}

// NewAuthorizableTransactor creates a new write-only instance of Authorizable, bound to a specific deployed contract.
func NewAuthorizableTransactor(address common.Address, transactor bind.ContractTransactor) (*AuthorizableTransactor, error) {
	contract, err := bindAuthorizable(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AuthorizableTransactor{contract: contract}, nil
}

// NewAuthorizableFilterer creates a new log filterer instance of Authorizable, bound to a specific deployed contract.
func NewAuthorizableFilterer(address common.Address, filterer bind.ContractFilterer) (*AuthorizableFilterer, error) {
	contract, err := bindAuthorizable(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AuthorizableFilterer{contract: contract}, nil
}

// bindAuthorizable binds a generic wrapper to an already deployed contract.
func bindAuthorizable(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AuthorizableABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Authorizable *AuthorizableRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Authorizable.Contract.AuthorizableCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Authorizable *AuthorizableRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Authorizable.Contract.AuthorizableTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Authorizable *AuthorizableRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Authorizable.Contract.AuthorizableTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Authorizable *AuthorizableCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Authorizable.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Authorizable *AuthorizableTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Authorizable.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Authorizable *AuthorizableTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Authorizable.Contract.contract.Transact(opts, method, params...)
}

// Authorizers is a free data retrieval call binding the contract method 0x56741b2c.
//
// Solidity: function authorizers() view returns(address)
func (_Authorizable *AuthorizableCaller) Authorizers(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Authorizable.contract.Call(opts, &out, "authorizers")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Authorizers is a free data retrieval call binding the contract method 0x56741b2c.
//
// Solidity: function authorizers() view returns(address)
func (_Authorizable *AuthorizableSession) Authorizers() (common.Address, error) {
	return _Authorizable.Contract.Authorizers(&_Authorizable.CallOpts)
}

// Authorizers is a free data retrieval call binding the contract method 0x56741b2c.
//
// Solidity: function authorizers() view returns(address)
func (_Authorizable *AuthorizableCallerSession) Authorizers() (common.Address, error) {
	return _Authorizable.Contract.Authorizers(&_Authorizable.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Authorizable *AuthorizableCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Authorizable.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Authorizable *AuthorizableSession) Owner() (common.Address, error) {
	return _Authorizable.Contract.Owner(&_Authorizable.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Authorizable *AuthorizableCallerSession) Owner() (common.Address, error) {
	return _Authorizable.Contract.Owner(&_Authorizable.CallOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Authorizable *AuthorizableTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Authorizable.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Authorizable *AuthorizableSession) RenounceOwnership() (*types.Transaction, error) {
	return _Authorizable.Contract.RenounceOwnership(&_Authorizable.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Authorizable *AuthorizableTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Authorizable.Contract.RenounceOwnership(&_Authorizable.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Authorizable *AuthorizableTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Authorizable.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Authorizable *AuthorizableSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Authorizable.Contract.TransferOwnership(&_Authorizable.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Authorizable *AuthorizableTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Authorizable.Contract.TransferOwnership(&_Authorizable.TransactOpts, newOwner)
}

// AuthorizableAuthorizersTransferredIterator is returned from FilterAuthorizersTransferred and is used to iterate over the raw logs and unpacked data for AuthorizersTransferred events raised by the Authorizable contract.
type AuthorizableAuthorizersTransferredIterator struct {
	Event *AuthorizableAuthorizersTransferred // Event containing the contract specifics and raw log

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
func (it *AuthorizableAuthorizersTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuthorizableAuthorizersTransferred)
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
		it.Event = new(AuthorizableAuthorizersTransferred)
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
func (it *AuthorizableAuthorizersTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuthorizableAuthorizersTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuthorizableAuthorizersTransferred represents a AuthorizersTransferred event raised by the Authorizable contract.
type AuthorizableAuthorizersTransferred struct {
	PreviousAuthorizers common.Address
	NewAuthorizers      common.Address
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterAuthorizersTransferred is a free log retrieval operation binding the contract event 0xc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d.
//
// Solidity: event AuthorizersTransferred(address indexed previousAuthorizers, address indexed newAuthorizers)
func (_Authorizable *AuthorizableFilterer) FilterAuthorizersTransferred(opts *bind.FilterOpts, previousAuthorizers []common.Address, newAuthorizers []common.Address) (*AuthorizableAuthorizersTransferredIterator, error) {

	var previousAuthorizersRule []interface{}
	for _, previousAuthorizersItem := range previousAuthorizers {
		previousAuthorizersRule = append(previousAuthorizersRule, previousAuthorizersItem)
	}
	var newAuthorizersRule []interface{}
	for _, newAuthorizersItem := range newAuthorizers {
		newAuthorizersRule = append(newAuthorizersRule, newAuthorizersItem)
	}

	logs, sub, err := _Authorizable.contract.FilterLogs(opts, "AuthorizersTransferred", previousAuthorizersRule, newAuthorizersRule)
	if err != nil {
		return nil, err
	}
	return &AuthorizableAuthorizersTransferredIterator{contract: _Authorizable.contract, event: "AuthorizersTransferred", logs: logs, sub: sub}, nil
}

// WatchAuthorizersTransferred is a free log subscription operation binding the contract event 0xc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d.
//
// Solidity: event AuthorizersTransferred(address indexed previousAuthorizers, address indexed newAuthorizers)
func (_Authorizable *AuthorizableFilterer) WatchAuthorizersTransferred(opts *bind.WatchOpts, sink chan<- *AuthorizableAuthorizersTransferred, previousAuthorizers []common.Address, newAuthorizers []common.Address) (event.Subscription, error) {

	var previousAuthorizersRule []interface{}
	for _, previousAuthorizersItem := range previousAuthorizers {
		previousAuthorizersRule = append(previousAuthorizersRule, previousAuthorizersItem)
	}
	var newAuthorizersRule []interface{}
	for _, newAuthorizersItem := range newAuthorizers {
		newAuthorizersRule = append(newAuthorizersRule, newAuthorizersItem)
	}

	logs, sub, err := _Authorizable.contract.WatchLogs(opts, "AuthorizersTransferred", previousAuthorizersRule, newAuthorizersRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuthorizableAuthorizersTransferred)
				if err := _Authorizable.contract.UnpackLog(event, "AuthorizersTransferred", log); err != nil {
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

// ParseAuthorizersTransferred is a log parse operation binding the contract event 0xc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d.
//
// Solidity: event AuthorizersTransferred(address indexed previousAuthorizers, address indexed newAuthorizers)
func (_Authorizable *AuthorizableFilterer) ParseAuthorizersTransferred(log types.Log) (*AuthorizableAuthorizersTransferred, error) {
	event := new(AuthorizableAuthorizersTransferred)
	if err := _Authorizable.contract.UnpackLog(event, "AuthorizersTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuthorizableOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Authorizable contract.
type AuthorizableOwnershipTransferredIterator struct {
	Event *AuthorizableOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *AuthorizableOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuthorizableOwnershipTransferred)
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
		it.Event = new(AuthorizableOwnershipTransferred)
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
func (it *AuthorizableOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuthorizableOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuthorizableOwnershipTransferred represents a OwnershipTransferred event raised by the Authorizable contract.
type AuthorizableOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Authorizable *AuthorizableFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AuthorizableOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Authorizable.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AuthorizableOwnershipTransferredIterator{contract: _Authorizable.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Authorizable *AuthorizableFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AuthorizableOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Authorizable.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuthorizableOwnershipTransferred)
				if err := _Authorizable.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Authorizable *AuthorizableFilterer) ParseOwnershipTransferred(log types.Log) (*AuthorizableOwnershipTransferred, error) {
	event := new(AuthorizableOwnershipTransferred)
	if err := _Authorizable.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeMetaData contains all meta data concerning the Bridge contract.
var BridgeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"contractIAuthorizers\",\"name\":\"_authorizers\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousAuthorizers\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newAuthorizers\",\"type\":\"address\"}],\"name\":\"AuthorizersTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"clientId\",\"type\":\"bytes\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"Burned\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"txid\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"Minted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"authorizers\",\"outputs\":[{\"internalType\":\"contractIAuthorizers\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"balance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_clientId\",\"type\":\"bytes\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"isAuthorizationValid\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_for\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"name\":\"mintFor\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenToRescue\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"rescueFunds\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"56741b2c": "authorizers()",
		"b69ef8a8": "balance()",
		"fe9d9303": "burn(uint256,bytes)",
		"408a12e6": "isAuthorizationValid(uint256,bytes,uint256,bytes)",
		"4d02be9f": "mint(uint256,bytes,uint256,bytes)",
		"d44a8430": "mintFor(address,uint256,bytes,uint256,bytes)",
		"8da5cb5b": "owner()",
		"715018a6": "renounceOwnership()",
		"6ccae054": "rescueFunds(address,address,uint256)",
		"fc0c546a": "token()",
		"f2fde38b": "transferOwnership(address)",
	},
	Bin: "0x6080604052600060035534801561001557600080fd5b506040516112af3803806112af833981016040819052610034916100f2565b600080546001600160a01b03191633908117825560405183928592918291907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908290a350600180546001600160a01b039283166001600160a01b0319918216179091556002805492841692909116821790556040516000907fc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d908290a350505061012c565b6001600160a01b03811681146100ef57600080fd5b50565b6000806040838503121561010557600080fd5b8251610110816100da565b6020840151909250610121816100da565b809150509250929050565b6111748061013b6000396000f3fe608060405234801561001057600080fd5b50600436106100a95760003560e01c80638da5cb5b116100715780638da5cb5b1461012b578063b69ef8a81461013c578063d44a843014610152578063f2fde38b14610165578063fc0c546a14610178578063fe9d93031461018b57600080fd5b8063408a12e6146100ae5780634d02be9f146100d657806356741b2c146100eb5780636ccae05414610110578063715018a614610123575b600080fd5b6100c16100bc366004610d4f565b61019e565b60405190151581526020015b60405180910390f35b6100e96100e4366004610d4f565b610251565b005b6002546001600160a01b03165b6040516001600160a01b0390911681526020016100cd565b6100c161011e366004610dea565b610395565b6100e96104c3565b6000546001600160a01b03166100f8565b610144610537565b6040519081526020016100cd565b6100e9610160366004610e2b565b6105b8565b6100e9610173366004610ec1565b6106f8565b6001546100f8906001600160a01b031681565b6100e9610199366004610ee5565b6107e2565b6000806101b36002546001600160a01b031690565b6001600160a01b031663f4fd62e3338a8a8a8a6040518663ffffffff1660e01b81526004016101e6959493929190610f5a565b602060405180830381600087803b15801561020057600080fd5b505af1158015610214573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906102389190610f94565b9050610245818585610828565b98975050505050505050565b60008381526004602052604090205460ff16156102aa5760405162461bcd60e51b8152602060048201526012602482015271139bdb98d948185b1c9958591e481d5cd95960721b60448201526064015b60405180910390fd5b60006102be6002546001600160a01b031690565b6001600160a01b031663f4fd62e333898989896040518663ffffffff1660e01b81526004016102f1959493929190610f5a565b602060405180830381600087803b15801561030b57600080fd5b505af115801561031f573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103439190610f94565b905061038c338888888080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508a9250879150899050886108e0565b50505050505050565b600080546001600160a01b031633146103c05760405162461bcd60e51b81526004016102a190610fad565b6001546001600160a01b03858116911614156104395760405162461bcd60e51b815260206004820152603260248201527f546f6b656e506f6f6c3a2043616e6e6f7420636c61696d20746f6b656e2068656044820152711b1908189e481d1a194818dbdb9d1c9858dd60721b60648201526084016102a1565b60405163a9059cbb60e01b81526001600160a01b0384811660048301526024820184905285169063a9059cbb90604401602060405180830381600087803b15801561048357600080fd5b505af1158015610497573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906104bb9190610fe2565b949350505050565b6000546001600160a01b031633146104ed5760405162461bcd60e51b81526004016102a190610fad565b600080546040516001600160a01b03909116907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908390a3600080546001600160a01b0319169055565b6001546040516370a0823160e01b81523060048201526000916001600160a01b0316906370a082319060240160206040518083038186803b15801561057b57600080fd5b505afa15801561058f573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105b39190610f94565b905090565b60008381526004602052604090205460ff161561060c5760405162461bcd60e51b8152602060048201526012602482015271139bdb98d948185b1c9958591e481d5cd95960721b60448201526064016102a1565b60006106206002546001600160a01b031690565b6001600160a01b031663f4fd62e389898989896040518663ffffffff1660e01b8152600401610653959493929190610f5a565b602060405180830381600087803b15801561066d57600080fd5b505af1158015610681573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106a59190610f94565b90506106ee888888888080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508a9250879150899050886108e0565b5050505050505050565b6000546001600160a01b031633146107225760405162461bcd60e51b81526004016102a190610fad565b6001600160a01b0381166107875760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016102a1565b600080546040516001600160a01b03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a3600080546001600160a01b0319166001600160a01b0392909216919091179055565b610823338484848080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250610b3f92505050565b505050565b6002546040516314b5dc0b60e11b81526000918591859185916001600160a01b039091169063296bb8169061086590869086908690600401611004565b602060405180830381600087803b15801561087f57600080fd5b505af1158015610893573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906108b79190610fe2565b6108d35760405162461bcd60e51b81526004016102a190611027565b5060019695505050505050565b6002546040516314b5dc0b60e11b81528491849184916001600160a01b03169063296bb8169061091890869086908690600401611004565b602060405180830381600087803b15801561093257600080fd5b505af1158015610946573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061096a9190610fe2565b6109865760405162461bcd60e51b81526004016102a190611027565b306001600160a01b031663fc0c546a6040518163ffffffff1660e01b815260040160206040518083038186803b1580156109bf57600080fd5b505afa1580156109d3573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109f7919061106d565b60405163a9059cbb60e01b81526001600160a01b038c81166004830152602482018c9052919091169063a9059cbb90604401602060405180830381600087803b158015610a4357600080fd5b505af1158015610a57573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610a7b9190610fe2565b610ad35760405162461bcd60e51b815260206004820152602360248201527f4272696467653a207472616e73666572206f7574206f6620706f6f6c206661696044820152621b195960ea1b60648201526084016102a1565b60008781526004602052604090819020805460ff19166001179055516001600160a01b038b16907fe04478a4154dc31a079fa36b9ee1af057f492a47c1524ac67f2ea4c214c3de9290610b2b908c908c908c906110ba565b60405180910390a250505050505050505050565b306001600160a01b031663fc0c546a6040518163ffffffff1660e01b815260040160206040518083038186803b158015610b7857600080fd5b505afa158015610b8c573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610bb0919061106d565b6040516323b872dd60e01b81526001600160a01b0385811660048301523060248301526044820185905291909116906323b872dd90606401602060405180830381600087803b158015610c0257600080fd5b505af1158015610c16573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610c3a9190610fe2565b610c955760405162461bcd60e51b815260206004820152602660248201527f4272696467653a207472616e7366657220696e746f206275726e20706f6f6c2060448201526519985a5b195960d21b60648201526084016102a1565b600354610ca39060016110fc565b6003819055604051610cb6908390611122565b6040518091039020846001600160a01b03167f2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df285604051610cf991815260200190565b60405180910390a4505050565b60008083601f840112610d1857600080fd5b50813567ffffffffffffffff811115610d3057600080fd5b602083019150836020828501011115610d4857600080fd5b9250929050565b60008060008060008060808789031215610d6857600080fd5b86359550602087013567ffffffffffffffff80821115610d8757600080fd5b610d938a838b01610d06565b9097509550604089013594506060890135915080821115610db357600080fd5b50610dc089828a01610d06565b979a9699509497509295939492505050565b6001600160a01b0381168114610de757600080fd5b50565b600080600060608486031215610dff57600080fd5b8335610e0a81610dd2565b92506020840135610e1a81610dd2565b929592945050506040919091013590565b600080600080600080600060a0888a031215610e4657600080fd5b8735610e5181610dd2565b965060208801359550604088013567ffffffffffffffff80821115610e7557600080fd5b610e818b838c01610d06565b909750955060608a0135945060808a0135915080821115610ea157600080fd5b50610eae8a828b01610d06565b989b979a50959850939692959293505050565b600060208284031215610ed357600080fd5b8135610ede81610dd2565b9392505050565b600080600060408486031215610efa57600080fd5b83359250602084013567ffffffffffffffff811115610f1857600080fd5b610f2486828701610d06565b9497909650939450505050565b81835281816020850137506000828201602090810191909152601f909101601f19169091010190565b60018060a01b0386168152846020820152608060408201526000610f82608083018587610f31565b90508260608301529695505050505050565b600060208284031215610fa657600080fd5b5051919050565b6020808252818101527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604082015260600190565b600060208284031215610ff457600080fd5b81518015158114610ede57600080fd5b83815260406020820152600061101e604083018486610f31565b95945050505050565b60208082526026908201527f417574686f72697a6572733a207369676e617475726573206e6f7420617574686040820152651bdc9a5e995960d21b606082015260800190565b60006020828403121561107f57600080fd5b8151610ede81610dd2565b60005b838110156110a557818101518382015260200161108d565b838111156110b4576000848401525b50505050565b83815260606020820152600083518060608401526110df81608085016020880161108a565b604083019390935250601f91909101601f19160160800192915050565b6000821982111561111d57634e487b7160e01b600052601160045260246000fd5b500190565b6000825161113481846020870161108a565b919091019291505056fea2646970667358221220b552b7380e61d622d53ca633a6557f71c075ff38b34d9a7f8dc4e58fb3248e4f64736f6c63430008090033",
}

// BridgeABI is the input ABI used to generate the binding from.
// Deprecated: Use BridgeMetaData.ABI instead.
var BridgeABI = BridgeMetaData.ABI

// Deprecated: Use BridgeMetaData.Sigs instead.
// BridgeFuncSigs maps the 4-byte function signature to its string representation.
var BridgeFuncSigs = BridgeMetaData.Sigs

// BridgeBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use BridgeMetaData.Bin instead.
var BridgeBin = BridgeMetaData.Bin

// DeployBridge deploys a new Ethereum contract, binding an instance of Bridge to it.
func DeployBridge(auth *bind.TransactOpts, backend bind.ContractBackend, _token common.Address, _authorizers common.Address) (common.Address, *types.Transaction, *Bridge, error) {
	parsed, err := BridgeMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(BridgeBin), backend, _token, _authorizers)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

// Bridge is an auto generated Go binding around an Ethereum contract.
type Bridge struct {
	BridgeCaller     // Read-only binding to the contract
	BridgeTransactor // Write-only binding to the contract
	BridgeFilterer   // Log filterer for contract events
}

// BridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type BridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BridgeSession struct {
	Contract     *Bridge           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BridgeCallerSession struct {
	Contract *BridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BridgeTransactorSession struct {
	Contract     *BridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type BridgeRaw struct {
	Contract *Bridge // Generic contract binding to access the raw methods on
}

// BridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BridgeCallerRaw struct {
	Contract *BridgeCaller // Generic read-only contract binding to access the raw methods on
}

// BridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BridgeTransactorRaw struct {
	Contract *BridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBridge creates a new instance of Bridge, bound to a specific deployed contract.
func NewBridge(address common.Address, backend bind.ContractBackend) (*Bridge, error) {
	contract, err := bindBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

// NewBridgeCaller creates a new read-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeCaller(address common.Address, caller bind.ContractCaller) (*BridgeCaller, error) {
	contract, err := bindBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeCaller{contract: contract}, nil
}

// NewBridgeTransactor creates a new write-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*BridgeTransactor, error) {
	contract, err := bindBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeTransactor{contract: contract}, nil
}

// NewBridgeFilterer creates a new log filterer instance of Bridge, bound to a specific deployed contract.
func NewBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*BridgeFilterer, error) {
	contract, err := bindBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BridgeFilterer{contract: contract}, nil
}

// bindBridge binds a generic wrapper to an already deployed contract.
func bindBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BridgeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.BridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transact(opts, method, params...)
}

// Authorizers is a free data retrieval call binding the contract method 0x56741b2c.
//
// Solidity: function authorizers() view returns(address)
func (_Bridge *BridgeCaller) Authorizers(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "authorizers")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Authorizers is a free data retrieval call binding the contract method 0x56741b2c.
//
// Solidity: function authorizers() view returns(address)
func (_Bridge *BridgeSession) Authorizers() (common.Address, error) {
	return _Bridge.Contract.Authorizers(&_Bridge.CallOpts)
}

// Authorizers is a free data retrieval call binding the contract method 0x56741b2c.
//
// Solidity: function authorizers() view returns(address)
func (_Bridge *BridgeCallerSession) Authorizers() (common.Address, error) {
	return _Bridge.Contract.Authorizers(&_Bridge.CallOpts)
}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_Bridge *BridgeCaller) Balance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "balance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_Bridge *BridgeSession) Balance() (*big.Int, error) {
	return _Bridge.Contract.Balance(&_Bridge.CallOpts)
}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_Bridge *BridgeCallerSession) Balance() (*big.Int, error) {
	return _Bridge.Contract.Balance(&_Bridge.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeSession) Owner() (common.Address, error) {
	return _Bridge.Contract.Owner(&_Bridge.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeCallerSession) Owner() (common.Address, error) {
	return _Bridge.Contract.Owner(&_Bridge.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Bridge *BridgeCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Bridge *BridgeSession) Token() (common.Address, error) {
	return _Bridge.Contract.Token(&_Bridge.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Bridge *BridgeCallerSession) Token() (common.Address, error) {
	return _Bridge.Contract.Token(&_Bridge.CallOpts)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 _amount, bytes _clientId) returns()
func (_Bridge *BridgeTransactor) Burn(opts *bind.TransactOpts, _amount *big.Int, _clientId []byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "burn", _amount, _clientId)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 _amount, bytes _clientId) returns()
func (_Bridge *BridgeSession) Burn(_amount *big.Int, _clientId []byte) (*types.Transaction, error) {
	return _Bridge.Contract.Burn(&_Bridge.TransactOpts, _amount, _clientId)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 _amount, bytes _clientId) returns()
func (_Bridge *BridgeTransactorSession) Burn(_amount *big.Int, _clientId []byte) (*types.Transaction, error) {
	return _Bridge.Contract.Burn(&_Bridge.TransactOpts, _amount, _clientId)
}

// IsAuthorizationValid is a paid mutator transaction binding the contract method 0x408a12e6.
//
// Solidity: function isAuthorizationValid(uint256 _amount, bytes _txid, uint256 _nonce, bytes signature) returns(bool)
func (_Bridge *BridgeTransactor) IsAuthorizationValid(opts *bind.TransactOpts, _amount *big.Int, _txid []byte, _nonce *big.Int, signature []byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "isAuthorizationValid", _amount, _txid, _nonce, signature)
}

// IsAuthorizationValid is a paid mutator transaction binding the contract method 0x408a12e6.
//
// Solidity: function isAuthorizationValid(uint256 _amount, bytes _txid, uint256 _nonce, bytes signature) returns(bool)
func (_Bridge *BridgeSession) IsAuthorizationValid(_amount *big.Int, _txid []byte, _nonce *big.Int, signature []byte) (*types.Transaction, error) {
	return _Bridge.Contract.IsAuthorizationValid(&_Bridge.TransactOpts, _amount, _txid, _nonce, signature)
}

// IsAuthorizationValid is a paid mutator transaction binding the contract method 0x408a12e6.
//
// Solidity: function isAuthorizationValid(uint256 _amount, bytes _txid, uint256 _nonce, bytes signature) returns(bool)
func (_Bridge *BridgeTransactorSession) IsAuthorizationValid(_amount *big.Int, _txid []byte, _nonce *big.Int, signature []byte) (*types.Transaction, error) {
	return _Bridge.Contract.IsAuthorizationValid(&_Bridge.TransactOpts, _amount, _txid, _nonce, signature)
}

// Mint is a paid mutator transaction binding the contract method 0x4d02be9f.
//
// Solidity: function mint(uint256 _amount, bytes _txid, uint256 _nonce, bytes signatures) returns()
func (_Bridge *BridgeTransactor) Mint(opts *bind.TransactOpts, _amount *big.Int, _txid []byte, _nonce *big.Int, signatures []byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "mint", _amount, _txid, _nonce, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x4d02be9f.
//
// Solidity: function mint(uint256 _amount, bytes _txid, uint256 _nonce, bytes signatures) returns()
func (_Bridge *BridgeSession) Mint(_amount *big.Int, _txid []byte, _nonce *big.Int, signatures []byte) (*types.Transaction, error) {
	return _Bridge.Contract.Mint(&_Bridge.TransactOpts, _amount, _txid, _nonce, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x4d02be9f.
//
// Solidity: function mint(uint256 _amount, bytes _txid, uint256 _nonce, bytes signatures) returns()
func (_Bridge *BridgeTransactorSession) Mint(_amount *big.Int, _txid []byte, _nonce *big.Int, signatures []byte) (*types.Transaction, error) {
	return _Bridge.Contract.Mint(&_Bridge.TransactOpts, _amount, _txid, _nonce, signatures)
}

// MintFor is a paid mutator transaction binding the contract method 0xd44a8430.
//
// Solidity: function mintFor(address _for, uint256 _amount, bytes _txid, uint256 _nonce, bytes signatures) returns()
func (_Bridge *BridgeTransactor) MintFor(opts *bind.TransactOpts, _for common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int, signatures []byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "mintFor", _for, _amount, _txid, _nonce, signatures)
}

// MintFor is a paid mutator transaction binding the contract method 0xd44a8430.
//
// Solidity: function mintFor(address _for, uint256 _amount, bytes _txid, uint256 _nonce, bytes signatures) returns()
func (_Bridge *BridgeSession) MintFor(_for common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int, signatures []byte) (*types.Transaction, error) {
	return _Bridge.Contract.MintFor(&_Bridge.TransactOpts, _for, _amount, _txid, _nonce, signatures)
}

// MintFor is a paid mutator transaction binding the contract method 0xd44a8430.
//
// Solidity: function mintFor(address _for, uint256 _amount, bytes _txid, uint256 _nonce, bytes signatures) returns()
func (_Bridge *BridgeTransactorSession) MintFor(_for common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int, signatures []byte) (*types.Transaction, error) {
	return _Bridge.Contract.MintFor(&_Bridge.TransactOpts, _for, _amount, _txid, _nonce, signatures)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bridge.Contract.RenounceOwnership(&_Bridge.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bridge.Contract.RenounceOwnership(&_Bridge.TransactOpts)
}

// RescueFunds is a paid mutator transaction binding the contract method 0x6ccae054.
//
// Solidity: function rescueFunds(address tokenToRescue, address to, uint256 amount) returns(bool)
func (_Bridge *BridgeTransactor) RescueFunds(opts *bind.TransactOpts, tokenToRescue common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "rescueFunds", tokenToRescue, to, amount)
}

// RescueFunds is a paid mutator transaction binding the contract method 0x6ccae054.
//
// Solidity: function rescueFunds(address tokenToRescue, address to, uint256 amount) returns(bool)
func (_Bridge *BridgeSession) RescueFunds(tokenToRescue common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.RescueFunds(&_Bridge.TransactOpts, tokenToRescue, to, amount)
}

// RescueFunds is a paid mutator transaction binding the contract method 0x6ccae054.
//
// Solidity: function rescueFunds(address tokenToRescue, address to, uint256 amount) returns(bool)
func (_Bridge *BridgeTransactorSession) RescueFunds(tokenToRescue common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.RescueFunds(&_Bridge.TransactOpts, tokenToRescue, to, amount)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.TransferOwnership(&_Bridge.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.TransferOwnership(&_Bridge.TransactOpts, newOwner)
}

// BridgeAuthorizersTransferredIterator is returned from FilterAuthorizersTransferred and is used to iterate over the raw logs and unpacked data for AuthorizersTransferred events raised by the Bridge contract.
type BridgeAuthorizersTransferredIterator struct {
	Event *BridgeAuthorizersTransferred // Event containing the contract specifics and raw log

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
func (it *BridgeAuthorizersTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeAuthorizersTransferred)
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
		it.Event = new(BridgeAuthorizersTransferred)
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
func (it *BridgeAuthorizersTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeAuthorizersTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeAuthorizersTransferred represents a AuthorizersTransferred event raised by the Bridge contract.
type BridgeAuthorizersTransferred struct {
	PreviousAuthorizers common.Address
	NewAuthorizers      common.Address
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterAuthorizersTransferred is a free log retrieval operation binding the contract event 0xc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d.
//
// Solidity: event AuthorizersTransferred(address indexed previousAuthorizers, address indexed newAuthorizers)
func (_Bridge *BridgeFilterer) FilterAuthorizersTransferred(opts *bind.FilterOpts, previousAuthorizers []common.Address, newAuthorizers []common.Address) (*BridgeAuthorizersTransferredIterator, error) {

	var previousAuthorizersRule []interface{}
	for _, previousAuthorizersItem := range previousAuthorizers {
		previousAuthorizersRule = append(previousAuthorizersRule, previousAuthorizersItem)
	}
	var newAuthorizersRule []interface{}
	for _, newAuthorizersItem := range newAuthorizers {
		newAuthorizersRule = append(newAuthorizersRule, newAuthorizersItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "AuthorizersTransferred", previousAuthorizersRule, newAuthorizersRule)
	if err != nil {
		return nil, err
	}
	return &BridgeAuthorizersTransferredIterator{contract: _Bridge.contract, event: "AuthorizersTransferred", logs: logs, sub: sub}, nil
}

// WatchAuthorizersTransferred is a free log subscription operation binding the contract event 0xc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d.
//
// Solidity: event AuthorizersTransferred(address indexed previousAuthorizers, address indexed newAuthorizers)
func (_Bridge *BridgeFilterer) WatchAuthorizersTransferred(opts *bind.WatchOpts, sink chan<- *BridgeAuthorizersTransferred, previousAuthorizers []common.Address, newAuthorizers []common.Address) (event.Subscription, error) {

	var previousAuthorizersRule []interface{}
	for _, previousAuthorizersItem := range previousAuthorizers {
		previousAuthorizersRule = append(previousAuthorizersRule, previousAuthorizersItem)
	}
	var newAuthorizersRule []interface{}
	for _, newAuthorizersItem := range newAuthorizers {
		newAuthorizersRule = append(newAuthorizersRule, newAuthorizersItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "AuthorizersTransferred", previousAuthorizersRule, newAuthorizersRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeAuthorizersTransferred)
				if err := _Bridge.contract.UnpackLog(event, "AuthorizersTransferred", log); err != nil {
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

// ParseAuthorizersTransferred is a log parse operation binding the contract event 0xc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d.
//
// Solidity: event AuthorizersTransferred(address indexed previousAuthorizers, address indexed newAuthorizers)
func (_Bridge *BridgeFilterer) ParseAuthorizersTransferred(log types.Log) (*BridgeAuthorizersTransferred, error) {
	event := new(BridgeAuthorizersTransferred)
	if err := _Bridge.contract.UnpackLog(event, "AuthorizersTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeBurnedIterator is returned from FilterBurned and is used to iterate over the raw logs and unpacked data for Burned events raised by the Bridge contract.
type BridgeBurnedIterator struct {
	Event *BridgeBurned // Event containing the contract specifics and raw log

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
func (it *BridgeBurnedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeBurned)
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
		it.Event = new(BridgeBurned)
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
func (it *BridgeBurnedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeBurnedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeBurned represents a Burned event raised by the Bridge contract.
type BridgeBurned struct {
	From     common.Address
	Amount   *big.Int
	ClientId common.Hash
	Nonce    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBurned is a free log retrieval operation binding the contract event 0x2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df2.
//
// Solidity: event Burned(address indexed from, uint256 amount, bytes indexed clientId, uint256 indexed nonce)
func (_Bridge *BridgeFilterer) FilterBurned(opts *bind.FilterOpts, from []common.Address, clientId [][]byte, nonce []*big.Int) (*BridgeBurnedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	var clientIdRule []interface{}
	for _, clientIdItem := range clientId {
		clientIdRule = append(clientIdRule, clientIdItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Burned", fromRule, clientIdRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return &BridgeBurnedIterator{contract: _Bridge.contract, event: "Burned", logs: logs, sub: sub}, nil
}

// WatchBurned is a free log subscription operation binding the contract event 0x2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df2.
//
// Solidity: event Burned(address indexed from, uint256 amount, bytes indexed clientId, uint256 indexed nonce)
func (_Bridge *BridgeFilterer) WatchBurned(opts *bind.WatchOpts, sink chan<- *BridgeBurned, from []common.Address, clientId [][]byte, nonce []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	var clientIdRule []interface{}
	for _, clientIdItem := range clientId {
		clientIdRule = append(clientIdRule, clientIdItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Burned", fromRule, clientIdRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeBurned)
				if err := _Bridge.contract.UnpackLog(event, "Burned", log); err != nil {
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

// ParseBurned is a log parse operation binding the contract event 0x2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df2.
//
// Solidity: event Burned(address indexed from, uint256 amount, bytes indexed clientId, uint256 indexed nonce)
func (_Bridge *BridgeFilterer) ParseBurned(log types.Log) (*BridgeBurned, error) {
	event := new(BridgeBurned)
	if err := _Bridge.contract.UnpackLog(event, "Burned", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeMintedIterator is returned from FilterMinted and is used to iterate over the raw logs and unpacked data for Minted events raised by the Bridge contract.
type BridgeMintedIterator struct {
	Event *BridgeMinted // Event containing the contract specifics and raw log

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
func (it *BridgeMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeMinted)
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
		it.Event = new(BridgeMinted)
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
func (it *BridgeMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeMinted represents a Minted event raised by the Bridge contract.
type BridgeMinted struct {
	To     common.Address
	Amount *big.Int
	Txid   []byte
	Nonce  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterMinted is a free log retrieval operation binding the contract event 0xe04478a4154dc31a079fa36b9ee1af057f492a47c1524ac67f2ea4c214c3de92.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, uint256 nonce)
func (_Bridge *BridgeFilterer) FilterMinted(opts *bind.FilterOpts, to []common.Address) (*BridgeMintedIterator, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Minted", toRule)
	if err != nil {
		return nil, err
	}
	return &BridgeMintedIterator{contract: _Bridge.contract, event: "Minted", logs: logs, sub: sub}, nil
}

// WatchMinted is a free log subscription operation binding the contract event 0xe04478a4154dc31a079fa36b9ee1af057f492a47c1524ac67f2ea4c214c3de92.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, uint256 nonce)
func (_Bridge *BridgeFilterer) WatchMinted(opts *bind.WatchOpts, sink chan<- *BridgeMinted, to []common.Address) (event.Subscription, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Minted", toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeMinted)
				if err := _Bridge.contract.UnpackLog(event, "Minted", log); err != nil {
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

// ParseMinted is a log parse operation binding the contract event 0xe04478a4154dc31a079fa36b9ee1af057f492a47c1524ac67f2ea4c214c3de92.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, uint256 nonce)
func (_Bridge *BridgeFilterer) ParseMinted(log types.Log) (*BridgeMinted, error) {
	event := new(BridgeMinted)
	if err := _Bridge.contract.UnpackLog(event, "Minted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Bridge contract.
type BridgeOwnershipTransferredIterator struct {
	Event *BridgeOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BridgeOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeOwnershipTransferred)
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
		it.Event = new(BridgeOwnershipTransferred)
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
func (it *BridgeOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeOwnershipTransferred represents a OwnershipTransferred event raised by the Bridge contract.
type BridgeOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bridge *BridgeFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BridgeOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BridgeOwnershipTransferredIterator{contract: _Bridge.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bridge *BridgeFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BridgeOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeOwnershipTransferred)
				if err := _Bridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Bridge *BridgeFilterer) ParseOwnershipTransferred(log types.Log) (*BridgeOwnershipTransferred, error) {
	event := new(BridgeOwnershipTransferred)
	if err := _Bridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContextMetaData contains all meta data concerning the Context contract.
var ContextMetaData = &bind.MetaData{
	ABI: "[]",
}

// ContextABI is the input ABI used to generate the binding from.
// Deprecated: Use ContextMetaData.ABI instead.
var ContextABI = ContextMetaData.ABI

// Context is an auto generated Go binding around an Ethereum contract.
type Context struct {
	ContextCaller     // Read-only binding to the contract
	ContextTransactor // Write-only binding to the contract
	ContextFilterer   // Log filterer for contract events
}

// ContextCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContextCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContextTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContextTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContextFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContextFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContextSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContextSession struct {
	Contract     *Context          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContextCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContextCallerSession struct {
	Contract *ContextCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// ContextTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContextTransactorSession struct {
	Contract     *ContextTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// ContextRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContextRaw struct {
	Contract *Context // Generic contract binding to access the raw methods on
}

// ContextCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContextCallerRaw struct {
	Contract *ContextCaller // Generic read-only contract binding to access the raw methods on
}

// ContextTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContextTransactorRaw struct {
	Contract *ContextTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContext creates a new instance of Context, bound to a specific deployed contract.
func NewContext(address common.Address, backend bind.ContractBackend) (*Context, error) {
	contract, err := bindContext(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Context{ContextCaller: ContextCaller{contract: contract}, ContextTransactor: ContextTransactor{contract: contract}, ContextFilterer: ContextFilterer{contract: contract}}, nil
}

// NewContextCaller creates a new read-only instance of Context, bound to a specific deployed contract.
func NewContextCaller(address common.Address, caller bind.ContractCaller) (*ContextCaller, error) {
	contract, err := bindContext(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContextCaller{contract: contract}, nil
}

// NewContextTransactor creates a new write-only instance of Context, bound to a specific deployed contract.
func NewContextTransactor(address common.Address, transactor bind.ContractTransactor) (*ContextTransactor, error) {
	contract, err := bindContext(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContextTransactor{contract: contract}, nil
}

// NewContextFilterer creates a new log filterer instance of Context, bound to a specific deployed contract.
func NewContextFilterer(address common.Address, filterer bind.ContractFilterer) (*ContextFilterer, error) {
	contract, err := bindContext(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContextFilterer{contract: contract}, nil
}

// bindContext binds a generic wrapper to an already deployed contract.
func bindContext(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContextABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Context *ContextRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Context.Contract.ContextCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Context *ContextRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Context.Contract.ContextTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Context *ContextRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Context.Contract.ContextTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Context *ContextCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Context.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Context *ContextTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Context.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Context *ContextTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Context.Contract.contract.Transact(opts, method, params...)
}

// IAuthorizersMetaData contains all meta data concerning the IAuthorizers contract.
var IAuthorizersMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"message\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"name\":\"authorize\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"message\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"296bb816": "authorize(bytes32,bytes)",
		"f4fd62e3": "message(address,uint256,bytes,uint256)",
	},
}

// IAuthorizersABI is the input ABI used to generate the binding from.
// Deprecated: Use IAuthorizersMetaData.ABI instead.
var IAuthorizersABI = IAuthorizersMetaData.ABI

// Deprecated: Use IAuthorizersMetaData.Sigs instead.
// IAuthorizersFuncSigs maps the 4-byte function signature to its string representation.
var IAuthorizersFuncSigs = IAuthorizersMetaData.Sigs

// IAuthorizers is an auto generated Go binding around an Ethereum contract.
type IAuthorizers struct {
	IAuthorizersCaller     // Read-only binding to the contract
	IAuthorizersTransactor // Write-only binding to the contract
	IAuthorizersFilterer   // Log filterer for contract events
}

// IAuthorizersCaller is an auto generated read-only Go binding around an Ethereum contract.
type IAuthorizersCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IAuthorizersTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IAuthorizersTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IAuthorizersFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IAuthorizersFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IAuthorizersSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IAuthorizersSession struct {
	Contract     *IAuthorizers     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IAuthorizersCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IAuthorizersCallerSession struct {
	Contract *IAuthorizersCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// IAuthorizersTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IAuthorizersTransactorSession struct {
	Contract     *IAuthorizersTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// IAuthorizersRaw is an auto generated low-level Go binding around an Ethereum contract.
type IAuthorizersRaw struct {
	Contract *IAuthorizers // Generic contract binding to access the raw methods on
}

// IAuthorizersCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IAuthorizersCallerRaw struct {
	Contract *IAuthorizersCaller // Generic read-only contract binding to access the raw methods on
}

// IAuthorizersTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IAuthorizersTransactorRaw struct {
	Contract *IAuthorizersTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIAuthorizers creates a new instance of IAuthorizers, bound to a specific deployed contract.
func NewIAuthorizers(address common.Address, backend bind.ContractBackend) (*IAuthorizers, error) {
	contract, err := bindIAuthorizers(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IAuthorizers{IAuthorizersCaller: IAuthorizersCaller{contract: contract}, IAuthorizersTransactor: IAuthorizersTransactor{contract: contract}, IAuthorizersFilterer: IAuthorizersFilterer{contract: contract}}, nil
}

// NewIAuthorizersCaller creates a new read-only instance of IAuthorizers, bound to a specific deployed contract.
func NewIAuthorizersCaller(address common.Address, caller bind.ContractCaller) (*IAuthorizersCaller, error) {
	contract, err := bindIAuthorizers(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IAuthorizersCaller{contract: contract}, nil
}

// NewIAuthorizersTransactor creates a new write-only instance of IAuthorizers, bound to a specific deployed contract.
func NewIAuthorizersTransactor(address common.Address, transactor bind.ContractTransactor) (*IAuthorizersTransactor, error) {
	contract, err := bindIAuthorizers(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IAuthorizersTransactor{contract: contract}, nil
}

// NewIAuthorizersFilterer creates a new log filterer instance of IAuthorizers, bound to a specific deployed contract.
func NewIAuthorizersFilterer(address common.Address, filterer bind.ContractFilterer) (*IAuthorizersFilterer, error) {
	contract, err := bindIAuthorizers(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IAuthorizersFilterer{contract: contract}, nil
}

// bindIAuthorizers binds a generic wrapper to an already deployed contract.
func bindIAuthorizers(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IAuthorizersABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IAuthorizers *IAuthorizersRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IAuthorizers.Contract.IAuthorizersCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IAuthorizers *IAuthorizersRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IAuthorizers.Contract.IAuthorizersTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IAuthorizers *IAuthorizersRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IAuthorizers.Contract.IAuthorizersTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IAuthorizers *IAuthorizersCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IAuthorizers.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IAuthorizers *IAuthorizersTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IAuthorizers.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IAuthorizers *IAuthorizersTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IAuthorizers.Contract.contract.Transact(opts, method, params...)
}

// Authorize is a paid mutator transaction binding the contract method 0x296bb816.
//
// Solidity: function authorize(bytes32 message, bytes signatures) returns(bool)
func (_IAuthorizers *IAuthorizersTransactor) Authorize(opts *bind.TransactOpts, message [32]byte, signatures []byte) (*types.Transaction, error) {
	return _IAuthorizers.contract.Transact(opts, "authorize", message, signatures)
}

// Authorize is a paid mutator transaction binding the contract method 0x296bb816.
//
// Solidity: function authorize(bytes32 message, bytes signatures) returns(bool)
func (_IAuthorizers *IAuthorizersSession) Authorize(message [32]byte, signatures []byte) (*types.Transaction, error) {
	return _IAuthorizers.Contract.Authorize(&_IAuthorizers.TransactOpts, message, signatures)
}

// Authorize is a paid mutator transaction binding the contract method 0x296bb816.
//
// Solidity: function authorize(bytes32 message, bytes signatures) returns(bool)
func (_IAuthorizers *IAuthorizersTransactorSession) Authorize(message [32]byte, signatures []byte) (*types.Transaction, error) {
	return _IAuthorizers.Contract.Authorize(&_IAuthorizers.TransactOpts, message, signatures)
}

// Message is a paid mutator transaction binding the contract method 0xf4fd62e3.
//
// Solidity: function message(address _to, uint256 _amount, bytes _txid, uint256 _nonce) returns(bytes32)
func (_IAuthorizers *IAuthorizersTransactor) Message(opts *bind.TransactOpts, _to common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int) (*types.Transaction, error) {
	return _IAuthorizers.contract.Transact(opts, "message", _to, _amount, _txid, _nonce)
}

// Message is a paid mutator transaction binding the contract method 0xf4fd62e3.
//
// Solidity: function message(address _to, uint256 _amount, bytes _txid, uint256 _nonce) returns(bytes32)
func (_IAuthorizers *IAuthorizersSession) Message(_to common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int) (*types.Transaction, error) {
	return _IAuthorizers.Contract.Message(&_IAuthorizers.TransactOpts, _to, _amount, _txid, _nonce)
}

// Message is a paid mutator transaction binding the contract method 0xf4fd62e3.
//
// Solidity: function message(address _to, uint256 _amount, bytes _txid, uint256 _nonce) returns(bytes32)
func (_IAuthorizers *IAuthorizersTransactorSession) Message(_to common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int) (*types.Transaction, error) {
	return _IAuthorizers.Contract.Message(&_IAuthorizers.TransactOpts, _to, _amount, _txid, _nonce)
}

// IBridgeMetaData contains all meta data concerning the IBridge contract.
var IBridgeMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"clientId\",\"type\":\"bytes\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"Burned\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"txid\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"Minted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_clientId\",\"type\":\"bytes\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"fe9d9303": "burn(uint256,bytes)",
		"4d02be9f": "mint(uint256,bytes,uint256,bytes)",
	},
}

// IBridgeABI is the input ABI used to generate the binding from.
// Deprecated: Use IBridgeMetaData.ABI instead.
var IBridgeABI = IBridgeMetaData.ABI

// Deprecated: Use IBridgeMetaData.Sigs instead.
// IBridgeFuncSigs maps the 4-byte function signature to its string representation.
var IBridgeFuncSigs = IBridgeMetaData.Sigs

// IBridge is an auto generated Go binding around an Ethereum contract.
type IBridge struct {
	IBridgeCaller     // Read-only binding to the contract
	IBridgeTransactor // Write-only binding to the contract
	IBridgeFilterer   // Log filterer for contract events
}

// IBridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type IBridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IBridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IBridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IBridgeSession struct {
	Contract     *IBridge          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IBridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IBridgeCallerSession struct {
	Contract *IBridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// IBridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IBridgeTransactorSession struct {
	Contract     *IBridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// IBridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type IBridgeRaw struct {
	Contract *IBridge // Generic contract binding to access the raw methods on
}

// IBridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IBridgeCallerRaw struct {
	Contract *IBridgeCaller // Generic read-only contract binding to access the raw methods on
}

// IBridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IBridgeTransactorRaw struct {
	Contract *IBridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIBridge creates a new instance of IBridge, bound to a specific deployed contract.
func NewIBridge(address common.Address, backend bind.ContractBackend) (*IBridge, error) {
	contract, err := bindIBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IBridge{IBridgeCaller: IBridgeCaller{contract: contract}, IBridgeTransactor: IBridgeTransactor{contract: contract}, IBridgeFilterer: IBridgeFilterer{contract: contract}}, nil
}

// NewIBridgeCaller creates a new read-only instance of IBridge, bound to a specific deployed contract.
func NewIBridgeCaller(address common.Address, caller bind.ContractCaller) (*IBridgeCaller, error) {
	contract, err := bindIBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IBridgeCaller{contract: contract}, nil
}

// NewIBridgeTransactor creates a new write-only instance of IBridge, bound to a specific deployed contract.
func NewIBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*IBridgeTransactor, error) {
	contract, err := bindIBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IBridgeTransactor{contract: contract}, nil
}

// NewIBridgeFilterer creates a new log filterer instance of IBridge, bound to a specific deployed contract.
func NewIBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*IBridgeFilterer, error) {
	contract, err := bindIBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IBridgeFilterer{contract: contract}, nil
}

// bindIBridge binds a generic wrapper to an already deployed contract.
func bindIBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IBridgeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IBridge *IBridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IBridge.Contract.IBridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IBridge *IBridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IBridge.Contract.IBridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IBridge *IBridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IBridge.Contract.IBridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IBridge *IBridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IBridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IBridge *IBridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IBridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IBridge *IBridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IBridge.Contract.contract.Transact(opts, method, params...)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 _amount, bytes _clientId) returns()
func (_IBridge *IBridgeTransactor) Burn(opts *bind.TransactOpts, _amount *big.Int, _clientId []byte) (*types.Transaction, error) {
	return _IBridge.contract.Transact(opts, "burn", _amount, _clientId)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 _amount, bytes _clientId) returns()
func (_IBridge *IBridgeSession) Burn(_amount *big.Int, _clientId []byte) (*types.Transaction, error) {
	return _IBridge.Contract.Burn(&_IBridge.TransactOpts, _amount, _clientId)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 _amount, bytes _clientId) returns()
func (_IBridge *IBridgeTransactorSession) Burn(_amount *big.Int, _clientId []byte) (*types.Transaction, error) {
	return _IBridge.Contract.Burn(&_IBridge.TransactOpts, _amount, _clientId)
}

// Mint is a paid mutator transaction binding the contract method 0x4d02be9f.
//
// Solidity: function mint(uint256 _amount, bytes _txid, uint256 _nonce, bytes signatures) returns()
func (_IBridge *IBridgeTransactor) Mint(opts *bind.TransactOpts, _amount *big.Int, _txid []byte, _nonce *big.Int, signatures []byte) (*types.Transaction, error) {
	return _IBridge.contract.Transact(opts, "mint", _amount, _txid, _nonce, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x4d02be9f.
//
// Solidity: function mint(uint256 _amount, bytes _txid, uint256 _nonce, bytes signatures) returns()
func (_IBridge *IBridgeSession) Mint(_amount *big.Int, _txid []byte, _nonce *big.Int, signatures []byte) (*types.Transaction, error) {
	return _IBridge.Contract.Mint(&_IBridge.TransactOpts, _amount, _txid, _nonce, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x4d02be9f.
//
// Solidity: function mint(uint256 _amount, bytes _txid, uint256 _nonce, bytes signatures) returns()
func (_IBridge *IBridgeTransactorSession) Mint(_amount *big.Int, _txid []byte, _nonce *big.Int, signatures []byte) (*types.Transaction, error) {
	return _IBridge.Contract.Mint(&_IBridge.TransactOpts, _amount, _txid, _nonce, signatures)
}

// IBridgeBurnedIterator is returned from FilterBurned and is used to iterate over the raw logs and unpacked data for Burned events raised by the IBridge contract.
type IBridgeBurnedIterator struct {
	Event *IBridgeBurned // Event containing the contract specifics and raw log

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
func (it *IBridgeBurnedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IBridgeBurned)
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
		it.Event = new(IBridgeBurned)
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
func (it *IBridgeBurnedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IBridgeBurnedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IBridgeBurned represents a Burned event raised by the IBridge contract.
type IBridgeBurned struct {
	From     common.Address
	Amount   *big.Int
	ClientId common.Hash
	Nonce    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBurned is a free log retrieval operation binding the contract event 0x2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df2.
//
// Solidity: event Burned(address indexed from, uint256 amount, bytes indexed clientId, uint256 indexed nonce)
func (_IBridge *IBridgeFilterer) FilterBurned(opts *bind.FilterOpts, from []common.Address, clientId [][]byte, nonce []*big.Int) (*IBridgeBurnedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	var clientIdRule []interface{}
	for _, clientIdItem := range clientId {
		clientIdRule = append(clientIdRule, clientIdItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _IBridge.contract.FilterLogs(opts, "Burned", fromRule, clientIdRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return &IBridgeBurnedIterator{contract: _IBridge.contract, event: "Burned", logs: logs, sub: sub}, nil
}

// WatchBurned is a free log subscription operation binding the contract event 0x2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df2.
//
// Solidity: event Burned(address indexed from, uint256 amount, bytes indexed clientId, uint256 indexed nonce)
func (_IBridge *IBridgeFilterer) WatchBurned(opts *bind.WatchOpts, sink chan<- *IBridgeBurned, from []common.Address, clientId [][]byte, nonce []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	var clientIdRule []interface{}
	for _, clientIdItem := range clientId {
		clientIdRule = append(clientIdRule, clientIdItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _IBridge.contract.WatchLogs(opts, "Burned", fromRule, clientIdRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IBridgeBurned)
				if err := _IBridge.contract.UnpackLog(event, "Burned", log); err != nil {
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

// ParseBurned is a log parse operation binding the contract event 0x2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df2.
//
// Solidity: event Burned(address indexed from, uint256 amount, bytes indexed clientId, uint256 indexed nonce)
func (_IBridge *IBridgeFilterer) ParseBurned(log types.Log) (*IBridgeBurned, error) {
	event := new(IBridgeBurned)
	if err := _IBridge.contract.UnpackLog(event, "Burned", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IBridgeMintedIterator is returned from FilterMinted and is used to iterate over the raw logs and unpacked data for Minted events raised by the IBridge contract.
type IBridgeMintedIterator struct {
	Event *IBridgeMinted // Event containing the contract specifics and raw log

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
func (it *IBridgeMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IBridgeMinted)
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
		it.Event = new(IBridgeMinted)
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
func (it *IBridgeMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IBridgeMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IBridgeMinted represents a Minted event raised by the IBridge contract.
type IBridgeMinted struct {
	To     common.Address
	Amount *big.Int
	Txid   []byte
	Nonce  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterMinted is a free log retrieval operation binding the contract event 0xe04478a4154dc31a079fa36b9ee1af057f492a47c1524ac67f2ea4c214c3de92.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, uint256 nonce)
func (_IBridge *IBridgeFilterer) FilterMinted(opts *bind.FilterOpts, to []common.Address) (*IBridgeMintedIterator, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IBridge.contract.FilterLogs(opts, "Minted", toRule)
	if err != nil {
		return nil, err
	}
	return &IBridgeMintedIterator{contract: _IBridge.contract, event: "Minted", logs: logs, sub: sub}, nil
}

// WatchMinted is a free log subscription operation binding the contract event 0xe04478a4154dc31a079fa36b9ee1af057f492a47c1524ac67f2ea4c214c3de92.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, uint256 nonce)
func (_IBridge *IBridgeFilterer) WatchMinted(opts *bind.WatchOpts, sink chan<- *IBridgeMinted, to []common.Address) (event.Subscription, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IBridge.contract.WatchLogs(opts, "Minted", toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IBridgeMinted)
				if err := _IBridge.contract.UnpackLog(event, "Minted", log); err != nil {
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

// ParseMinted is a log parse operation binding the contract event 0xe04478a4154dc31a079fa36b9ee1af057f492a47c1524ac67f2ea4c214c3de92.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, uint256 nonce)
func (_IBridge *IBridgeFilterer) ParseMinted(log types.Log) (*IBridgeMinted, error) {
	event := new(IBridgeMinted)
	if err := _IBridge.contract.UnpackLog(event, "Minted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IERC20MetaData contains all meta data concerning the IERC20 contract.
var IERC20MetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"dd62ed3e": "allowance(address,address)",
		"095ea7b3": "approve(address,uint256)",
		"70a08231": "balanceOf(address)",
		"18160ddd": "totalSupply()",
		"a9059cbb": "transfer(address,uint256)",
		"23b872dd": "transferFrom(address,address,uint256)",
	},
}

// IERC20ABI is the input ABI used to generate the binding from.
// Deprecated: Use IERC20MetaData.ABI instead.
var IERC20ABI = IERC20MetaData.ABI

// Deprecated: Use IERC20MetaData.Sigs instead.
// IERC20FuncSigs maps the 4-byte function signature to its string representation.
var IERC20FuncSigs = IERC20MetaData.Sigs

// IERC20 is an auto generated Go binding around an Ethereum contract.
type IERC20 struct {
	IERC20Caller     // Read-only binding to the contract
	IERC20Transactor // Write-only binding to the contract
	IERC20Filterer   // Log filterer for contract events
}

// IERC20Caller is an auto generated read-only Go binding around an Ethereum contract.
type IERC20Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20Transactor is an auto generated write-only Go binding around an Ethereum contract.
type IERC20Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IERC20Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IERC20Session struct {
	Contract     *IERC20           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IERC20CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IERC20CallerSession struct {
	Contract *IERC20Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// IERC20TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IERC20TransactorSession struct {
	Contract     *IERC20Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IERC20Raw is an auto generated low-level Go binding around an Ethereum contract.
type IERC20Raw struct {
	Contract *IERC20 // Generic contract binding to access the raw methods on
}

// IERC20CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IERC20CallerRaw struct {
	Contract *IERC20Caller // Generic read-only contract binding to access the raw methods on
}

// IERC20TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IERC20TransactorRaw struct {
	Contract *IERC20Transactor // Generic write-only contract binding to access the raw methods on
}

// NewIERC20 creates a new instance of IERC20, bound to a specific deployed contract.
func NewIERC20(address common.Address, backend bind.ContractBackend) (*IERC20, error) {
	contract, err := bindIERC20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IERC20{IERC20Caller: IERC20Caller{contract: contract}, IERC20Transactor: IERC20Transactor{contract: contract}, IERC20Filterer: IERC20Filterer{contract: contract}}, nil
}

// NewIERC20Caller creates a new read-only instance of IERC20, bound to a specific deployed contract.
func NewIERC20Caller(address common.Address, caller bind.ContractCaller) (*IERC20Caller, error) {
	contract, err := bindIERC20(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IERC20Caller{contract: contract}, nil
}

// NewIERC20Transactor creates a new write-only instance of IERC20, bound to a specific deployed contract.
func NewIERC20Transactor(address common.Address, transactor bind.ContractTransactor) (*IERC20Transactor, error) {
	contract, err := bindIERC20(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IERC20Transactor{contract: contract}, nil
}

// NewIERC20Filterer creates a new log filterer instance of IERC20, bound to a specific deployed contract.
func NewIERC20Filterer(address common.Address, filterer bind.ContractFilterer) (*IERC20Filterer, error) {
	contract, err := bindIERC20(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IERC20Filterer{contract: contract}, nil
}

// bindIERC20 binds a generic wrapper to an already deployed contract.
func bindIERC20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IERC20ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IERC20 *IERC20Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IERC20.Contract.IERC20Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IERC20 *IERC20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IERC20.Contract.IERC20Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IERC20 *IERC20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IERC20.Contract.IERC20Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IERC20 *IERC20CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IERC20.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IERC20 *IERC20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IERC20.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IERC20 *IERC20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IERC20.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_IERC20 *IERC20Caller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IERC20.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_IERC20 *IERC20Session) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _IERC20.Contract.Allowance(&_IERC20.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_IERC20 *IERC20CallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _IERC20.Contract.Allowance(&_IERC20.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_IERC20 *IERC20Caller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IERC20.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_IERC20 *IERC20Session) BalanceOf(account common.Address) (*big.Int, error) {
	return _IERC20.Contract.BalanceOf(&_IERC20.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_IERC20 *IERC20CallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _IERC20.Contract.BalanceOf(&_IERC20.CallOpts, account)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_IERC20 *IERC20Caller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IERC20.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_IERC20 *IERC20Session) TotalSupply() (*big.Int, error) {
	return _IERC20.Contract.TotalSupply(&_IERC20.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_IERC20 *IERC20CallerSession) TotalSupply() (*big.Int, error) {
	return _IERC20.Contract.TotalSupply(&_IERC20.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_IERC20 *IERC20Transactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_IERC20 *IERC20Session) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Approve(&_IERC20.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_IERC20 *IERC20TransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Approve(&_IERC20.TransactOpts, spender, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20Transactor) Transfer(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.contract.Transact(opts, "transfer", recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20Session) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Transfer(&_IERC20.TransactOpts, recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20TransactorSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Transfer(&_IERC20.TransactOpts, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20Transactor) TransferFrom(opts *bind.TransactOpts, sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.contract.Transact(opts, "transferFrom", sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20Session) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.TransferFrom(&_IERC20.TransactOpts, sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20TransactorSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.TransferFrom(&_IERC20.TransactOpts, sender, recipient, amount)
}

// IERC20ApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the IERC20 contract.
type IERC20ApprovalIterator struct {
	Event *IERC20Approval // Event containing the contract specifics and raw log

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
func (it *IERC20ApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IERC20Approval)
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
		it.Event = new(IERC20Approval)
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
func (it *IERC20ApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IERC20ApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IERC20Approval represents a Approval event raised by the IERC20 contract.
type IERC20Approval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_IERC20 *IERC20Filterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*IERC20ApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _IERC20.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &IERC20ApprovalIterator{contract: _IERC20.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_IERC20 *IERC20Filterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *IERC20Approval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _IERC20.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IERC20Approval)
				if err := _IERC20.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_IERC20 *IERC20Filterer) ParseApproval(log types.Log) (*IERC20Approval, error) {
	event := new(IERC20Approval)
	if err := _IERC20.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IERC20TransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the IERC20 contract.
type IERC20TransferIterator struct {
	Event *IERC20Transfer // Event containing the contract specifics and raw log

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
func (it *IERC20TransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IERC20Transfer)
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
		it.Event = new(IERC20Transfer)
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
func (it *IERC20TransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IERC20TransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IERC20Transfer represents a Transfer event raised by the IERC20 contract.
type IERC20Transfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_IERC20 *IERC20Filterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*IERC20TransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IERC20.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &IERC20TransferIterator{contract: _IERC20.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_IERC20 *IERC20Filterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *IERC20Transfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IERC20.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IERC20Transfer)
				if err := _IERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_IERC20 *IERC20Filterer) ParseTransfer(log types.Log) (*IERC20Transfer, error) {
	event := new(IERC20Transfer)
	if err := _IERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OwnableMetaData contains all meta data concerning the Ownable contract.
var OwnableMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"8da5cb5b": "owner()",
		"715018a6": "renounceOwnership()",
		"f2fde38b": "transferOwnership(address)",
	},
}

// OwnableABI is the input ABI used to generate the binding from.
// Deprecated: Use OwnableMetaData.ABI instead.
var OwnableABI = OwnableMetaData.ABI

// Deprecated: Use OwnableMetaData.Sigs instead.
// OwnableFuncSigs maps the 4-byte function signature to its string representation.
var OwnableFuncSigs = OwnableMetaData.Sigs

// Ownable is an auto generated Go binding around an Ethereum contract.
type Ownable struct {
	OwnableCaller     // Read-only binding to the contract
	OwnableTransactor // Write-only binding to the contract
	OwnableFilterer   // Log filterer for contract events
}

// OwnableCaller is an auto generated read-only Go binding around an Ethereum contract.
type OwnableCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnableTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OwnableTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnableFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OwnableFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnableSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OwnableSession struct {
	Contract     *Ownable          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OwnableCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OwnableCallerSession struct {
	Contract *OwnableCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// OwnableTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OwnableTransactorSession struct {
	Contract     *OwnableTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// OwnableRaw is an auto generated low-level Go binding around an Ethereum contract.
type OwnableRaw struct {
	Contract *Ownable // Generic contract binding to access the raw methods on
}

// OwnableCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OwnableCallerRaw struct {
	Contract *OwnableCaller // Generic read-only contract binding to access the raw methods on
}

// OwnableTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OwnableTransactorRaw struct {
	Contract *OwnableTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOwnable creates a new instance of Ownable, bound to a specific deployed contract.
func NewOwnable(address common.Address, backend bind.ContractBackend) (*Ownable, error) {
	contract, err := bindOwnable(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Ownable{OwnableCaller: OwnableCaller{contract: contract}, OwnableTransactor: OwnableTransactor{contract: contract}, OwnableFilterer: OwnableFilterer{contract: contract}}, nil
}

// NewOwnableCaller creates a new read-only instance of Ownable, bound to a specific deployed contract.
func NewOwnableCaller(address common.Address, caller bind.ContractCaller) (*OwnableCaller, error) {
	contract, err := bindOwnable(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OwnableCaller{contract: contract}, nil
}

// NewOwnableTransactor creates a new write-only instance of Ownable, bound to a specific deployed contract.
func NewOwnableTransactor(address common.Address, transactor bind.ContractTransactor) (*OwnableTransactor, error) {
	contract, err := bindOwnable(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OwnableTransactor{contract: contract}, nil
}

// NewOwnableFilterer creates a new log filterer instance of Ownable, bound to a specific deployed contract.
func NewOwnableFilterer(address common.Address, filterer bind.ContractFilterer) (*OwnableFilterer, error) {
	contract, err := bindOwnable(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OwnableFilterer{contract: contract}, nil
}

// bindOwnable binds a generic wrapper to an already deployed contract.
func bindOwnable(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OwnableABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ownable *OwnableRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ownable.Contract.OwnableCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ownable *OwnableRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ownable.Contract.OwnableTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ownable *OwnableRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ownable.Contract.OwnableTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ownable *OwnableCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ownable.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ownable *OwnableTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ownable.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ownable *OwnableTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ownable.Contract.contract.Transact(opts, method, params...)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ownable *OwnableCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Ownable.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ownable *OwnableSession) Owner() (common.Address, error) {
	return _Ownable.Contract.Owner(&_Ownable.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ownable *OwnableCallerSession) Owner() (common.Address, error) {
	return _Ownable.Contract.Owner(&_Ownable.CallOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Ownable *OwnableTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ownable.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Ownable *OwnableSession) RenounceOwnership() (*types.Transaction, error) {
	return _Ownable.Contract.RenounceOwnership(&_Ownable.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Ownable *OwnableTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Ownable.Contract.RenounceOwnership(&_Ownable.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ownable *OwnableTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Ownable.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ownable *OwnableSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Ownable.Contract.TransferOwnership(&_Ownable.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ownable *OwnableTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Ownable.Contract.TransferOwnership(&_Ownable.TransactOpts, newOwner)
}

// OwnableOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Ownable contract.
type OwnableOwnershipTransferredIterator struct {
	Event *OwnableOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *OwnableOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OwnableOwnershipTransferred)
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
		it.Event = new(OwnableOwnershipTransferred)
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
func (it *OwnableOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OwnableOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OwnableOwnershipTransferred represents a OwnershipTransferred event raised by the Ownable contract.
type OwnableOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Ownable *OwnableFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*OwnableOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Ownable.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &OwnableOwnershipTransferredIterator{contract: _Ownable.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Ownable *OwnableFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *OwnableOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Ownable.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OwnableOwnershipTransferred)
				if err := _Ownable.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Ownable *OwnableFilterer) ParseOwnershipTransferred(log types.Log) (*OwnableOwnershipTransferred, error) {
	event := new(OwnableOwnershipTransferred)
	if err := _Ownable.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TokenPoolMetaData contains all meta data concerning the TokenPool contract.
var TokenPoolMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"_token\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"balance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenToRescue\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"rescueFunds\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"b69ef8a8": "balance()",
		"8da5cb5b": "owner()",
		"715018a6": "renounceOwnership()",
		"6ccae054": "rescueFunds(address,address,uint256)",
		"fc0c546a": "token()",
		"f2fde38b": "transferOwnership(address)",
	},
	Bin: "0x608060405234801561001057600080fd5b5060405161060438038061060483398101604081905261002f91610095565b600080546001600160a01b031916339081178255604051909182917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908290a350600180546001600160a01b0319166001600160a01b03929092169190911790556100c5565b6000602082840312156100a757600080fd5b81516001600160a01b03811681146100be57600080fd5b9392505050565b610530806100d46000396000f3fe608060405234801561001057600080fd5b50600436106100625760003560e01c80636ccae05414610067578063715018a61461008f5780638da5cb5b14610099578063b69ef8a8146100be578063f2fde38b146100d4578063fc0c546a146100e7575b600080fd5b61007a61007536600461042c565b6100fa565b60405190151581526020015b60405180910390f35b610097610231565b005b6000546001600160a01b03165b6040516001600160a01b039091168152602001610086565b6100c66102a5565b604051908152602001610086565b6100976100e2366004610468565b610326565b6001546100a6906001600160a01b031681565b600080546001600160a01b0316331461012e5760405162461bcd60e51b81526004016101259061048a565b60405180910390fd5b6001546001600160a01b03858116911614156101a75760405162461bcd60e51b815260206004820152603260248201527f546f6b656e506f6f6c3a2043616e6e6f7420636c61696d20746f6b656e2068656044820152711b1908189e481d1a194818dbdb9d1c9858dd60721b6064820152608401610125565b60405163a9059cbb60e01b81526001600160a01b0384811660048301526024820184905285169063a9059cbb90604401602060405180830381600087803b1580156101f157600080fd5b505af1158015610205573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061022991906104bf565b949350505050565b6000546001600160a01b0316331461025b5760405162461bcd60e51b81526004016101259061048a565b600080546040516001600160a01b03909116907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908390a3600080546001600160a01b0319169055565b6001546040516370a0823160e01b81523060048201526000916001600160a01b0316906370a082319060240160206040518083038186803b1580156102e957600080fd5b505afa1580156102fd573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061032191906104e1565b905090565b6000546001600160a01b031633146103505760405162461bcd60e51b81526004016101259061048a565b6001600160a01b0381166103b55760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b6064820152608401610125565b600080546040516001600160a01b03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a3600080546001600160a01b0319166001600160a01b0392909216919091179055565b80356001600160a01b038116811461042757600080fd5b919050565b60008060006060848603121561044157600080fd5b61044a84610410565b925061045860208501610410565b9150604084013590509250925092565b60006020828403121561047a57600080fd5b61048382610410565b9392505050565b6020808252818101527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604082015260600190565b6000602082840312156104d157600080fd5b8151801515811461048357600080fd5b6000602082840312156104f357600080fd5b505191905056fea2646970667358221220c6b2cdbdf1df2f88753f576d0fbc0a1d04edc2ac0942542f7d998300f8a4568d64736f6c63430008090033",
}

// TokenPoolABI is the input ABI used to generate the binding from.
// Deprecated: Use TokenPoolMetaData.ABI instead.
var TokenPoolABI = TokenPoolMetaData.ABI

// Deprecated: Use TokenPoolMetaData.Sigs instead.
// TokenPoolFuncSigs maps the 4-byte function signature to its string representation.
var TokenPoolFuncSigs = TokenPoolMetaData.Sigs

// TokenPoolBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use TokenPoolMetaData.Bin instead.
var TokenPoolBin = TokenPoolMetaData.Bin

// DeployTokenPool deploys a new Ethereum contract, binding an instance of TokenPool to it.
func DeployTokenPool(auth *bind.TransactOpts, backend bind.ContractBackend, _token common.Address) (common.Address, *types.Transaction, *TokenPool, error) {
	parsed, err := TokenPoolMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(TokenPoolBin), backend, _token)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &TokenPool{TokenPoolCaller: TokenPoolCaller{contract: contract}, TokenPoolTransactor: TokenPoolTransactor{contract: contract}, TokenPoolFilterer: TokenPoolFilterer{contract: contract}}, nil
}

// TokenPool is an auto generated Go binding around an Ethereum contract.
type TokenPool struct {
	TokenPoolCaller     // Read-only binding to the contract
	TokenPoolTransactor // Write-only binding to the contract
	TokenPoolFilterer   // Log filterer for contract events
}

// TokenPoolCaller is an auto generated read-only Go binding around an Ethereum contract.
type TokenPoolCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenPoolTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TokenPoolTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenPoolFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TokenPoolFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenPoolSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TokenPoolSession struct {
	Contract     *TokenPool        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TokenPoolCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TokenPoolCallerSession struct {
	Contract *TokenPoolCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// TokenPoolTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TokenPoolTransactorSession struct {
	Contract     *TokenPoolTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// TokenPoolRaw is an auto generated low-level Go binding around an Ethereum contract.
type TokenPoolRaw struct {
	Contract *TokenPool // Generic contract binding to access the raw methods on
}

// TokenPoolCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TokenPoolCallerRaw struct {
	Contract *TokenPoolCaller // Generic read-only contract binding to access the raw methods on
}

// TokenPoolTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TokenPoolTransactorRaw struct {
	Contract *TokenPoolTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTokenPool creates a new instance of TokenPool, bound to a specific deployed contract.
func NewTokenPool(address common.Address, backend bind.ContractBackend) (*TokenPool, error) {
	contract, err := bindTokenPool(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TokenPool{TokenPoolCaller: TokenPoolCaller{contract: contract}, TokenPoolTransactor: TokenPoolTransactor{contract: contract}, TokenPoolFilterer: TokenPoolFilterer{contract: contract}}, nil
}

// NewTokenPoolCaller creates a new read-only instance of TokenPool, bound to a specific deployed contract.
func NewTokenPoolCaller(address common.Address, caller bind.ContractCaller) (*TokenPoolCaller, error) {
	contract, err := bindTokenPool(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TokenPoolCaller{contract: contract}, nil
}

// NewTokenPoolTransactor creates a new write-only instance of TokenPool, bound to a specific deployed contract.
func NewTokenPoolTransactor(address common.Address, transactor bind.ContractTransactor) (*TokenPoolTransactor, error) {
	contract, err := bindTokenPool(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TokenPoolTransactor{contract: contract}, nil
}

// NewTokenPoolFilterer creates a new log filterer instance of TokenPool, bound to a specific deployed contract.
func NewTokenPoolFilterer(address common.Address, filterer bind.ContractFilterer) (*TokenPoolFilterer, error) {
	contract, err := bindTokenPool(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TokenPoolFilterer{contract: contract}, nil
}

// bindTokenPool binds a generic wrapper to an already deployed contract.
func bindTokenPool(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TokenPoolABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TokenPool *TokenPoolRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TokenPool.Contract.TokenPoolCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TokenPool *TokenPoolRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TokenPool.Contract.TokenPoolTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TokenPool *TokenPoolRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TokenPool.Contract.TokenPoolTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TokenPool *TokenPoolCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TokenPool.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TokenPool *TokenPoolTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TokenPool.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TokenPool *TokenPoolTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TokenPool.Contract.contract.Transact(opts, method, params...)
}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_TokenPool *TokenPoolCaller) Balance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _TokenPool.contract.Call(opts, &out, "balance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_TokenPool *TokenPoolSession) Balance() (*big.Int, error) {
	return _TokenPool.Contract.Balance(&_TokenPool.CallOpts)
}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_TokenPool *TokenPoolCallerSession) Balance() (*big.Int, error) {
	return _TokenPool.Contract.Balance(&_TokenPool.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_TokenPool *TokenPoolCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TokenPool.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_TokenPool *TokenPoolSession) Owner() (common.Address, error) {
	return _TokenPool.Contract.Owner(&_TokenPool.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_TokenPool *TokenPoolCallerSession) Owner() (common.Address, error) {
	return _TokenPool.Contract.Owner(&_TokenPool.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_TokenPool *TokenPoolCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TokenPool.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_TokenPool *TokenPoolSession) Token() (common.Address, error) {
	return _TokenPool.Contract.Token(&_TokenPool.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_TokenPool *TokenPoolCallerSession) Token() (common.Address, error) {
	return _TokenPool.Contract.Token(&_TokenPool.CallOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_TokenPool *TokenPoolTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TokenPool.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_TokenPool *TokenPoolSession) RenounceOwnership() (*types.Transaction, error) {
	return _TokenPool.Contract.RenounceOwnership(&_TokenPool.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_TokenPool *TokenPoolTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _TokenPool.Contract.RenounceOwnership(&_TokenPool.TransactOpts)
}

// RescueFunds is a paid mutator transaction binding the contract method 0x6ccae054.
//
// Solidity: function rescueFunds(address tokenToRescue, address to, uint256 amount) returns(bool)
func (_TokenPool *TokenPoolTransactor) RescueFunds(opts *bind.TransactOpts, tokenToRescue common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TokenPool.contract.Transact(opts, "rescueFunds", tokenToRescue, to, amount)
}

// RescueFunds is a paid mutator transaction binding the contract method 0x6ccae054.
//
// Solidity: function rescueFunds(address tokenToRescue, address to, uint256 amount) returns(bool)
func (_TokenPool *TokenPoolSession) RescueFunds(tokenToRescue common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TokenPool.Contract.RescueFunds(&_TokenPool.TransactOpts, tokenToRescue, to, amount)
}

// RescueFunds is a paid mutator transaction binding the contract method 0x6ccae054.
//
// Solidity: function rescueFunds(address tokenToRescue, address to, uint256 amount) returns(bool)
func (_TokenPool *TokenPoolTransactorSession) RescueFunds(tokenToRescue common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _TokenPool.Contract.RescueFunds(&_TokenPool.TransactOpts, tokenToRescue, to, amount)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_TokenPool *TokenPoolTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _TokenPool.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_TokenPool *TokenPoolSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _TokenPool.Contract.TransferOwnership(&_TokenPool.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_TokenPool *TokenPoolTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _TokenPool.Contract.TransferOwnership(&_TokenPool.TransactOpts, newOwner)
}

// TokenPoolOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the TokenPool contract.
type TokenPoolOwnershipTransferredIterator struct {
	Event *TokenPoolOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *TokenPoolOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TokenPoolOwnershipTransferred)
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
		it.Event = new(TokenPoolOwnershipTransferred)
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
func (it *TokenPoolOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TokenPoolOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TokenPoolOwnershipTransferred represents a OwnershipTransferred event raised by the TokenPool contract.
type TokenPoolOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_TokenPool *TokenPoolFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*TokenPoolOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _TokenPool.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &TokenPoolOwnershipTransferredIterator{contract: _TokenPool.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_TokenPool *TokenPoolFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *TokenPoolOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _TokenPool.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TokenPoolOwnershipTransferred)
				if err := _TokenPool.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_TokenPool *TokenPoolFilterer) ParseOwnershipTransferred(log types.Log) (*TokenPoolOwnershipTransferred, error) {
	event := new(TokenPoolOwnershipTransferred)
	if err := _TokenPool.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
