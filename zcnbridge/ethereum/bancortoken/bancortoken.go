// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bancortoken

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

// BancortokenMetaData contains all meta data concerning the Bancortoken contract.
var BancortokenMetaData = &bind.MetaData{
	ABI: "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_disable\",\"type\":\"bool\"}],\"name\":\"disableTransfers\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"standard\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"withdrawTokens\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"issue\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"destroy\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"transfersEnabled\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"newOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"inputs\":[{\"name\":\"_name\",\"type\":\"string\"},{\"name\":\"_symbol\",\"type\":\"string\"},{\"name\":\"_decimals\",\"type\":\"uint8\"}],\"payable\":false,\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"NewSmartToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"Issuance\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"Destruction\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_prevOwner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"OwnerUpdate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]",
}

// BancortokenABI is the input ABI used to generate the binding from.
// Deprecated: Use BancortokenMetaData.ABI instead.
var BancortokenABI = BancortokenMetaData.ABI

// Bancortoken is an auto generated Go binding around an Ethereum contract.
type Bancortoken struct {
	BancortokenCaller     // Read-only binding to the contract
	BancortokenTransactor // Write-only binding to the contract
	BancortokenFilterer   // Log filterer for contract events
}

// BancortokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type BancortokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BancortokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BancortokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BancortokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BancortokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BancortokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BancortokenSession struct {
	Contract     *Bancortoken      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BancortokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BancortokenCallerSession struct {
	Contract *BancortokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// BancortokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BancortokenTransactorSession struct {
	Contract     *BancortokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// BancortokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type BancortokenRaw struct {
	Contract *Bancortoken // Generic contract binding to access the raw methods on
}

// BancortokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BancortokenCallerRaw struct {
	Contract *BancortokenCaller // Generic read-only contract binding to access the raw methods on
}

// BancortokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BancortokenTransactorRaw struct {
	Contract *BancortokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBancortoken creates a new instance of Bancortoken, bound to a specific deployed contract.
func NewBancortoken(address common.Address, backend bind.ContractBackend) (*Bancortoken, error) {
	contract, err := bindBancortoken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bancortoken{BancortokenCaller: BancortokenCaller{contract: contract}, BancortokenTransactor: BancortokenTransactor{contract: contract}, BancortokenFilterer: BancortokenFilterer{contract: contract}}, nil
}

// NewBancortokenCaller creates a new read-only instance of Bancortoken, bound to a specific deployed contract.
func NewBancortokenCaller(address common.Address, caller bind.ContractCaller) (*BancortokenCaller, error) {
	contract, err := bindBancortoken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BancortokenCaller{contract: contract}, nil
}

// NewBancortokenTransactor creates a new write-only instance of Bancortoken, bound to a specific deployed contract.
func NewBancortokenTransactor(address common.Address, transactor bind.ContractTransactor) (*BancortokenTransactor, error) {
	contract, err := bindBancortoken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BancortokenTransactor{contract: contract}, nil
}

// NewBancortokenFilterer creates a new log filterer instance of Bancortoken, bound to a specific deployed contract.
func NewBancortokenFilterer(address common.Address, filterer bind.ContractFilterer) (*BancortokenFilterer, error) {
	contract, err := bindBancortoken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BancortokenFilterer{contract: contract}, nil
}

// bindBancortoken binds a generic wrapper to an already deployed contract.
func bindBancortoken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BancortokenMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bancortoken *BancortokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bancortoken.Contract.BancortokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bancortoken *BancortokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bancortoken.Contract.BancortokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bancortoken *BancortokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bancortoken.Contract.BancortokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bancortoken *BancortokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bancortoken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bancortoken *BancortokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bancortoken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bancortoken *BancortokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bancortoken.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) returns(uint256)
func (_Bancortoken *BancortokenCaller) Allowance(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Bancortoken.contract.Call(opts, &out, "allowance", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) returns(uint256)
func (_Bancortoken *BancortokenSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Bancortoken.Contract.Allowance(&_Bancortoken.CallOpts, arg0, arg1)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) returns(uint256)
func (_Bancortoken *BancortokenCallerSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Bancortoken.Contract.Allowance(&_Bancortoken.CallOpts, arg0, arg1)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) returns(uint256)
func (_Bancortoken *BancortokenCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Bancortoken.contract.Call(opts, &out, "balanceOf", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) returns(uint256)
func (_Bancortoken *BancortokenSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _Bancortoken.Contract.BalanceOf(&_Bancortoken.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) returns(uint256)
func (_Bancortoken *BancortokenCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _Bancortoken.Contract.BalanceOf(&_Bancortoken.CallOpts, arg0)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() returns(uint8)
func (_Bancortoken *BancortokenCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Bancortoken.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() returns(uint8)
func (_Bancortoken *BancortokenSession) Decimals() (uint8, error) {
	return _Bancortoken.Contract.Decimals(&_Bancortoken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() returns(uint8)
func (_Bancortoken *BancortokenCallerSession) Decimals() (uint8, error) {
	return _Bancortoken.Contract.Decimals(&_Bancortoken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() returns(string)
func (_Bancortoken *BancortokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Bancortoken.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() returns(string)
func (_Bancortoken *BancortokenSession) Name() (string, error) {
	return _Bancortoken.Contract.Name(&_Bancortoken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() returns(string)
func (_Bancortoken *BancortokenCallerSession) Name() (string, error) {
	return _Bancortoken.Contract.Name(&_Bancortoken.CallOpts)
}

// NewOwner is a free data retrieval call binding the contract method 0xd4ee1d90.
//
// Solidity: function newOwner() returns(address)
func (_Bancortoken *BancortokenCaller) NewOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bancortoken.contract.Call(opts, &out, "newOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NewOwner is a free data retrieval call binding the contract method 0xd4ee1d90.
//
// Solidity: function newOwner() returns(address)
func (_Bancortoken *BancortokenSession) NewOwner() (common.Address, error) {
	return _Bancortoken.Contract.NewOwner(&_Bancortoken.CallOpts)
}

// NewOwner is a free data retrieval call binding the contract method 0xd4ee1d90.
//
// Solidity: function newOwner() returns(address)
func (_Bancortoken *BancortokenCallerSession) NewOwner() (common.Address, error) {
	return _Bancortoken.Contract.NewOwner(&_Bancortoken.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() returns(address)
func (_Bancortoken *BancortokenCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bancortoken.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() returns(address)
func (_Bancortoken *BancortokenSession) Owner() (common.Address, error) {
	return _Bancortoken.Contract.Owner(&_Bancortoken.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() returns(address)
func (_Bancortoken *BancortokenCallerSession) Owner() (common.Address, error) {
	return _Bancortoken.Contract.Owner(&_Bancortoken.CallOpts)
}

// Standard is a free data retrieval call binding the contract method 0x5a3b7e42.
//
// Solidity: function standard() returns(string)
func (_Bancortoken *BancortokenCaller) Standard(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Bancortoken.contract.Call(opts, &out, "standard")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Standard is a free data retrieval call binding the contract method 0x5a3b7e42.
//
// Solidity: function standard() returns(string)
func (_Bancortoken *BancortokenSession) Standard() (string, error) {
	return _Bancortoken.Contract.Standard(&_Bancortoken.CallOpts)
}

// Standard is a free data retrieval call binding the contract method 0x5a3b7e42.
//
// Solidity: function standard() returns(string)
func (_Bancortoken *BancortokenCallerSession) Standard() (string, error) {
	return _Bancortoken.Contract.Standard(&_Bancortoken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() returns(string)
func (_Bancortoken *BancortokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Bancortoken.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() returns(string)
func (_Bancortoken *BancortokenSession) Symbol() (string, error) {
	return _Bancortoken.Contract.Symbol(&_Bancortoken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() returns(string)
func (_Bancortoken *BancortokenCallerSession) Symbol() (string, error) {
	return _Bancortoken.Contract.Symbol(&_Bancortoken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() returns(uint256)
func (_Bancortoken *BancortokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bancortoken.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() returns(uint256)
func (_Bancortoken *BancortokenSession) TotalSupply() (*big.Int, error) {
	return _Bancortoken.Contract.TotalSupply(&_Bancortoken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() returns(uint256)
func (_Bancortoken *BancortokenCallerSession) TotalSupply() (*big.Int, error) {
	return _Bancortoken.Contract.TotalSupply(&_Bancortoken.CallOpts)
}

// TransfersEnabled is a free data retrieval call binding the contract method 0xbef97c87.
//
// Solidity: function transfersEnabled() returns(bool)
func (_Bancortoken *BancortokenCaller) TransfersEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Bancortoken.contract.Call(opts, &out, "transfersEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// TransfersEnabled is a free data retrieval call binding the contract method 0xbef97c87.
//
// Solidity: function transfersEnabled() returns(bool)
func (_Bancortoken *BancortokenSession) TransfersEnabled() (bool, error) {
	return _Bancortoken.Contract.TransfersEnabled(&_Bancortoken.CallOpts)
}

// TransfersEnabled is a free data retrieval call binding the contract method 0xbef97c87.
//
// Solidity: function transfersEnabled() returns(bool)
func (_Bancortoken *BancortokenCallerSession) TransfersEnabled() (bool, error) {
	return _Bancortoken.Contract.TransfersEnabled(&_Bancortoken.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() returns(string)
func (_Bancortoken *BancortokenCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Bancortoken.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() returns(string)
func (_Bancortoken *BancortokenSession) Version() (string, error) {
	return _Bancortoken.Contract.Version(&_Bancortoken.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() returns(string)
func (_Bancortoken *BancortokenCallerSession) Version() (string, error) {
	return _Bancortoken.Contract.Version(&_Bancortoken.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_Bancortoken *BancortokenTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bancortoken.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_Bancortoken *BancortokenSession) AcceptOwnership() (*types.Transaction, error) {
	return _Bancortoken.Contract.AcceptOwnership(&_Bancortoken.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_Bancortoken *BancortokenTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _Bancortoken.Contract.AcceptOwnership(&_Bancortoken.TransactOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool success)
func (_Bancortoken *BancortokenTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bancortoken.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool success)
func (_Bancortoken *BancortokenSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.Approve(&_Bancortoken.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool success)
func (_Bancortoken *BancortokenTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.Approve(&_Bancortoken.TransactOpts, _spender, _value)
}

// Destroy is a paid mutator transaction binding the contract method 0xa24835d1.
//
// Solidity: function destroy(address _from, uint256 _amount) returns()
func (_Bancortoken *BancortokenTransactor) Destroy(opts *bind.TransactOpts, _from common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bancortoken.contract.Transact(opts, "destroy", _from, _amount)
}

// Destroy is a paid mutator transaction binding the contract method 0xa24835d1.
//
// Solidity: function destroy(address _from, uint256 _amount) returns()
func (_Bancortoken *BancortokenSession) Destroy(_from common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.Destroy(&_Bancortoken.TransactOpts, _from, _amount)
}

// Destroy is a paid mutator transaction binding the contract method 0xa24835d1.
//
// Solidity: function destroy(address _from, uint256 _amount) returns()
func (_Bancortoken *BancortokenTransactorSession) Destroy(_from common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.Destroy(&_Bancortoken.TransactOpts, _from, _amount)
}

// DisableTransfers is a paid mutator transaction binding the contract method 0x1608f18f.
//
// Solidity: function disableTransfers(bool _disable) returns()
func (_Bancortoken *BancortokenTransactor) DisableTransfers(opts *bind.TransactOpts, _disable bool) (*types.Transaction, error) {
	return _Bancortoken.contract.Transact(opts, "disableTransfers", _disable)
}

// DisableTransfers is a paid mutator transaction binding the contract method 0x1608f18f.
//
// Solidity: function disableTransfers(bool _disable) returns()
func (_Bancortoken *BancortokenSession) DisableTransfers(_disable bool) (*types.Transaction, error) {
	return _Bancortoken.Contract.DisableTransfers(&_Bancortoken.TransactOpts, _disable)
}

// DisableTransfers is a paid mutator transaction binding the contract method 0x1608f18f.
//
// Solidity: function disableTransfers(bool _disable) returns()
func (_Bancortoken *BancortokenTransactorSession) DisableTransfers(_disable bool) (*types.Transaction, error) {
	return _Bancortoken.Contract.DisableTransfers(&_Bancortoken.TransactOpts, _disable)
}

// Issue is a paid mutator transaction binding the contract method 0x867904b4.
//
// Solidity: function issue(address _to, uint256 _amount) returns()
func (_Bancortoken *BancortokenTransactor) Issue(opts *bind.TransactOpts, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bancortoken.contract.Transact(opts, "issue", _to, _amount)
}

// Issue is a paid mutator transaction binding the contract method 0x867904b4.
//
// Solidity: function issue(address _to, uint256 _amount) returns()
func (_Bancortoken *BancortokenSession) Issue(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.Issue(&_Bancortoken.TransactOpts, _to, _amount)
}

// Issue is a paid mutator transaction binding the contract method 0x867904b4.
//
// Solidity: function issue(address _to, uint256 _amount) returns()
func (_Bancortoken *BancortokenTransactorSession) Issue(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.Issue(&_Bancortoken.TransactOpts, _to, _amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool success)
func (_Bancortoken *BancortokenTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bancortoken.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool success)
func (_Bancortoken *BancortokenSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.Transfer(&_Bancortoken.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool success)
func (_Bancortoken *BancortokenTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.Transfer(&_Bancortoken.TransactOpts, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool success)
func (_Bancortoken *BancortokenTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bancortoken.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool success)
func (_Bancortoken *BancortokenSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.TransferFrom(&_Bancortoken.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool success)
func (_Bancortoken *BancortokenTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.TransferFrom(&_Bancortoken.TransactOpts, _from, _to, _value)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _newOwner) returns()
func (_Bancortoken *BancortokenTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _Bancortoken.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _newOwner) returns()
func (_Bancortoken *BancortokenSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _Bancortoken.Contract.TransferOwnership(&_Bancortoken.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _newOwner) returns()
func (_Bancortoken *BancortokenTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _Bancortoken.Contract.TransferOwnership(&_Bancortoken.TransactOpts, _newOwner)
}

// WithdrawTokens is a paid mutator transaction binding the contract method 0x5e35359e.
//
// Solidity: function withdrawTokens(address _token, address _to, uint256 _amount) returns()
func (_Bancortoken *BancortokenTransactor) WithdrawTokens(opts *bind.TransactOpts, _token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bancortoken.contract.Transact(opts, "withdrawTokens", _token, _to, _amount)
}

// WithdrawTokens is a paid mutator transaction binding the contract method 0x5e35359e.
//
// Solidity: function withdrawTokens(address _token, address _to, uint256 _amount) returns()
func (_Bancortoken *BancortokenSession) WithdrawTokens(_token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.WithdrawTokens(&_Bancortoken.TransactOpts, _token, _to, _amount)
}

// WithdrawTokens is a paid mutator transaction binding the contract method 0x5e35359e.
//
// Solidity: function withdrawTokens(address _token, address _to, uint256 _amount) returns()
func (_Bancortoken *BancortokenTransactorSession) WithdrawTokens(_token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bancortoken.Contract.WithdrawTokens(&_Bancortoken.TransactOpts, _token, _to, _amount)
}

// BancortokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the Bancortoken contract.
type BancortokenApprovalIterator struct {
	Event *BancortokenApproval // Event containing the contract specifics and raw log

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
func (it *BancortokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancortokenApproval)
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
		it.Event = new(BancortokenApproval)
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
func (it *BancortokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancortokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancortokenApproval represents a Approval event raised by the Bancortoken contract.
type BancortokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _value)
func (_Bancortoken *BancortokenFilterer) FilterApproval(opts *bind.FilterOpts, _owner []common.Address, _spender []common.Address) (*BancortokenApprovalIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _Bancortoken.contract.FilterLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return &BancortokenApprovalIterator{contract: _Bancortoken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _value)
func (_Bancortoken *BancortokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *BancortokenApproval, _owner []common.Address, _spender []common.Address) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _Bancortoken.contract.WatchLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancortokenApproval)
				if err := _Bancortoken.contract.UnpackLog(event, "Approval", log); err != nil {
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
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _value)
func (_Bancortoken *BancortokenFilterer) ParseApproval(log types.Log) (*BancortokenApproval, error) {
	event := new(BancortokenApproval)
	if err := _Bancortoken.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancortokenDestructionIterator is returned from FilterDestruction and is used to iterate over the raw logs and unpacked data for Destruction events raised by the Bancortoken contract.
type BancortokenDestructionIterator struct {
	Event *BancortokenDestruction // Event containing the contract specifics and raw log

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
func (it *BancortokenDestructionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancortokenDestruction)
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
		it.Event = new(BancortokenDestruction)
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
func (it *BancortokenDestructionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancortokenDestructionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancortokenDestruction represents a Destruction event raised by the Bancortoken contract.
type BancortokenDestruction struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDestruction is a free log retrieval operation binding the contract event 0x9a1b418bc061a5d80270261562e6986a35d995f8051145f277be16103abd3453.
//
// Solidity: event Destruction(uint256 _amount)
func (_Bancortoken *BancortokenFilterer) FilterDestruction(opts *bind.FilterOpts) (*BancortokenDestructionIterator, error) {

	logs, sub, err := _Bancortoken.contract.FilterLogs(opts, "Destruction")
	if err != nil {
		return nil, err
	}
	return &BancortokenDestructionIterator{contract: _Bancortoken.contract, event: "Destruction", logs: logs, sub: sub}, nil
}

// WatchDestruction is a free log subscription operation binding the contract event 0x9a1b418bc061a5d80270261562e6986a35d995f8051145f277be16103abd3453.
//
// Solidity: event Destruction(uint256 _amount)
func (_Bancortoken *BancortokenFilterer) WatchDestruction(opts *bind.WatchOpts, sink chan<- *BancortokenDestruction) (event.Subscription, error) {

	logs, sub, err := _Bancortoken.contract.WatchLogs(opts, "Destruction")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancortokenDestruction)
				if err := _Bancortoken.contract.UnpackLog(event, "Destruction", log); err != nil {
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

// ParseDestruction is a log parse operation binding the contract event 0x9a1b418bc061a5d80270261562e6986a35d995f8051145f277be16103abd3453.
//
// Solidity: event Destruction(uint256 _amount)
func (_Bancortoken *BancortokenFilterer) ParseDestruction(log types.Log) (*BancortokenDestruction, error) {
	event := new(BancortokenDestruction)
	if err := _Bancortoken.contract.UnpackLog(event, "Destruction", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancortokenIssuanceIterator is returned from FilterIssuance and is used to iterate over the raw logs and unpacked data for Issuance events raised by the Bancortoken contract.
type BancortokenIssuanceIterator struct {
	Event *BancortokenIssuance // Event containing the contract specifics and raw log

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
func (it *BancortokenIssuanceIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancortokenIssuance)
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
		it.Event = new(BancortokenIssuance)
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
func (it *BancortokenIssuanceIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancortokenIssuanceIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancortokenIssuance represents a Issuance event raised by the Bancortoken contract.
type BancortokenIssuance struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterIssuance is a free log retrieval operation binding the contract event 0x9386c90217c323f58030f9dadcbc938f807a940f4ff41cd4cead9562f5da7dc3.
//
// Solidity: event Issuance(uint256 _amount)
func (_Bancortoken *BancortokenFilterer) FilterIssuance(opts *bind.FilterOpts) (*BancortokenIssuanceIterator, error) {

	logs, sub, err := _Bancortoken.contract.FilterLogs(opts, "Issuance")
	if err != nil {
		return nil, err
	}
	return &BancortokenIssuanceIterator{contract: _Bancortoken.contract, event: "Issuance", logs: logs, sub: sub}, nil
}

// WatchIssuance is a free log subscription operation binding the contract event 0x9386c90217c323f58030f9dadcbc938f807a940f4ff41cd4cead9562f5da7dc3.
//
// Solidity: event Issuance(uint256 _amount)
func (_Bancortoken *BancortokenFilterer) WatchIssuance(opts *bind.WatchOpts, sink chan<- *BancortokenIssuance) (event.Subscription, error) {

	logs, sub, err := _Bancortoken.contract.WatchLogs(opts, "Issuance")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancortokenIssuance)
				if err := _Bancortoken.contract.UnpackLog(event, "Issuance", log); err != nil {
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

// ParseIssuance is a log parse operation binding the contract event 0x9386c90217c323f58030f9dadcbc938f807a940f4ff41cd4cead9562f5da7dc3.
//
// Solidity: event Issuance(uint256 _amount)
func (_Bancortoken *BancortokenFilterer) ParseIssuance(log types.Log) (*BancortokenIssuance, error) {
	event := new(BancortokenIssuance)
	if err := _Bancortoken.contract.UnpackLog(event, "Issuance", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancortokenNewSmartTokenIterator is returned from FilterNewSmartToken and is used to iterate over the raw logs and unpacked data for NewSmartToken events raised by the Bancortoken contract.
type BancortokenNewSmartTokenIterator struct {
	Event *BancortokenNewSmartToken // Event containing the contract specifics and raw log

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
func (it *BancortokenNewSmartTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancortokenNewSmartToken)
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
		it.Event = new(BancortokenNewSmartToken)
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
func (it *BancortokenNewSmartTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancortokenNewSmartTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancortokenNewSmartToken represents a NewSmartToken event raised by the Bancortoken contract.
type BancortokenNewSmartToken struct {
	Token common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterNewSmartToken is a free log retrieval operation binding the contract event 0xf4cd1f8571e8d9c97ffcb81558807ab73f9803d54de5da6a0420593c82a4a9f0.
//
// Solidity: event NewSmartToken(address _token)
func (_Bancortoken *BancortokenFilterer) FilterNewSmartToken(opts *bind.FilterOpts) (*BancortokenNewSmartTokenIterator, error) {

	logs, sub, err := _Bancortoken.contract.FilterLogs(opts, "NewSmartToken")
	if err != nil {
		return nil, err
	}
	return &BancortokenNewSmartTokenIterator{contract: _Bancortoken.contract, event: "NewSmartToken", logs: logs, sub: sub}, nil
}

// WatchNewSmartToken is a free log subscription operation binding the contract event 0xf4cd1f8571e8d9c97ffcb81558807ab73f9803d54de5da6a0420593c82a4a9f0.
//
// Solidity: event NewSmartToken(address _token)
func (_Bancortoken *BancortokenFilterer) WatchNewSmartToken(opts *bind.WatchOpts, sink chan<- *BancortokenNewSmartToken) (event.Subscription, error) {

	logs, sub, err := _Bancortoken.contract.WatchLogs(opts, "NewSmartToken")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancortokenNewSmartToken)
				if err := _Bancortoken.contract.UnpackLog(event, "NewSmartToken", log); err != nil {
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

// ParseNewSmartToken is a log parse operation binding the contract event 0xf4cd1f8571e8d9c97ffcb81558807ab73f9803d54de5da6a0420593c82a4a9f0.
//
// Solidity: event NewSmartToken(address _token)
func (_Bancortoken *BancortokenFilterer) ParseNewSmartToken(log types.Log) (*BancortokenNewSmartToken, error) {
	event := new(BancortokenNewSmartToken)
	if err := _Bancortoken.contract.UnpackLog(event, "NewSmartToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancortokenOwnerUpdateIterator is returned from FilterOwnerUpdate and is used to iterate over the raw logs and unpacked data for OwnerUpdate events raised by the Bancortoken contract.
type BancortokenOwnerUpdateIterator struct {
	Event *BancortokenOwnerUpdate // Event containing the contract specifics and raw log

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
func (it *BancortokenOwnerUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancortokenOwnerUpdate)
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
		it.Event = new(BancortokenOwnerUpdate)
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
func (it *BancortokenOwnerUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancortokenOwnerUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancortokenOwnerUpdate represents a OwnerUpdate event raised by the Bancortoken contract.
type BancortokenOwnerUpdate struct {
	PrevOwner common.Address
	NewOwner  common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterOwnerUpdate is a free log retrieval operation binding the contract event 0x343765429aea5a34b3ff6a3785a98a5abb2597aca87bfbb58632c173d585373a.
//
// Solidity: event OwnerUpdate(address _prevOwner, address _newOwner)
func (_Bancortoken *BancortokenFilterer) FilterOwnerUpdate(opts *bind.FilterOpts) (*BancortokenOwnerUpdateIterator, error) {

	logs, sub, err := _Bancortoken.contract.FilterLogs(opts, "OwnerUpdate")
	if err != nil {
		return nil, err
	}
	return &BancortokenOwnerUpdateIterator{contract: _Bancortoken.contract, event: "OwnerUpdate", logs: logs, sub: sub}, nil
}

// WatchOwnerUpdate is a free log subscription operation binding the contract event 0x343765429aea5a34b3ff6a3785a98a5abb2597aca87bfbb58632c173d585373a.
//
// Solidity: event OwnerUpdate(address _prevOwner, address _newOwner)
func (_Bancortoken *BancortokenFilterer) WatchOwnerUpdate(opts *bind.WatchOpts, sink chan<- *BancortokenOwnerUpdate) (event.Subscription, error) {

	logs, sub, err := _Bancortoken.contract.WatchLogs(opts, "OwnerUpdate")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancortokenOwnerUpdate)
				if err := _Bancortoken.contract.UnpackLog(event, "OwnerUpdate", log); err != nil {
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

// ParseOwnerUpdate is a log parse operation binding the contract event 0x343765429aea5a34b3ff6a3785a98a5abb2597aca87bfbb58632c173d585373a.
//
// Solidity: event OwnerUpdate(address _prevOwner, address _newOwner)
func (_Bancortoken *BancortokenFilterer) ParseOwnerUpdate(log types.Log) (*BancortokenOwnerUpdate, error) {
	event := new(BancortokenOwnerUpdate)
	if err := _Bancortoken.contract.UnpackLog(event, "OwnerUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancortokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the Bancortoken contract.
type BancortokenTransferIterator struct {
	Event *BancortokenTransfer // Event containing the contract specifics and raw log

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
func (it *BancortokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancortokenTransfer)
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
		it.Event = new(BancortokenTransfer)
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
func (it *BancortokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancortokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancortokenTransfer represents a Transfer event raised by the Bancortoken contract.
type BancortokenTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _value)
func (_Bancortoken *BancortokenFilterer) FilterTransfer(opts *bind.FilterOpts, _from []common.Address, _to []common.Address) (*BancortokenTransferIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _Bancortoken.contract.FilterLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return &BancortokenTransferIterator{contract: _Bancortoken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _value)
func (_Bancortoken *BancortokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *BancortokenTransfer, _from []common.Address, _to []common.Address) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _Bancortoken.contract.WatchLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancortokenTransfer)
				if err := _Bancortoken.contract.UnpackLog(event, "Transfer", log); err != nil {
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
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _value)
func (_Bancortoken *BancortokenFilterer) ParseTransfer(log types.Log) (*BancortokenTransfer, error) {
	event := new(BancortokenTransfer)
	if err := _Bancortoken.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}