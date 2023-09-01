// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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

// IBancorNetworkMetaData contains all meta data concerning the IBancorNetwork contract.
var IBancorNetworkMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIReserveToken\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"contractIReserveToken\",\"name\":\"targetToken\",\"type\":\"address\"}],\"name\":\"conversionPath\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"path\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"sourceAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minReturn\",\"type\":\"uint256\"},{\"internalType\":\"addresspayable\",\"name\":\"beneficiary\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"affiliateAccount\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"affiliateFee\",\"type\":\"uint256\"}],\"name\":\"convertByPath\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"path\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"sourceAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minReturn\",\"type\":\"uint256\"},{\"internalType\":\"addresspayable\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"convertByPath2\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"path\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"sourceAmount\",\"type\":\"uint256\"}],\"name\":\"rateByPath\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"d734fa19": "conversionPath(address,address)",
		"b77d239b": "convertByPath(address[],uint256,uint256,address,address,uint256)",
		"f8394fb7": "convertByPath2(address[],uint256,uint256,address)",
		"7f9c0ecd": "rateByPath(address[],uint256)",
	},
}

// IBancorNetworkABI is the input ABI used to generate the binding from.
// Deprecated: Use IBancorNetworkMetaData.ABI instead.
var IBancorNetworkABI = IBancorNetworkMetaData.ABI

// Deprecated: Use IBancorNetworkMetaData.Sigs instead.
// IBancorNetworkFuncSigs maps the 4-byte function signature to its string representation.
var IBancorNetworkFuncSigs = IBancorNetworkMetaData.Sigs

// IBancorNetwork is an auto generated Go binding around an Ethereum contract.
type IBancorNetwork struct {
	IBancorNetworkCaller     // Read-only binding to the contract
	IBancorNetworkTransactor // Write-only binding to the contract
	IBancorNetworkFilterer   // Log filterer for contract events
}

// IBancorNetworkCaller is an auto generated read-only Go binding around an Ethereum contract.
type IBancorNetworkCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBancorNetworkTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IBancorNetworkTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBancorNetworkFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IBancorNetworkFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBancorNetworkSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IBancorNetworkSession struct {
	Contract     *IBancorNetwork   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IBancorNetworkCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IBancorNetworkCallerSession struct {
	Contract *IBancorNetworkCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// IBancorNetworkTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IBancorNetworkTransactorSession struct {
	Contract     *IBancorNetworkTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// IBancorNetworkRaw is an auto generated low-level Go binding around an Ethereum contract.
type IBancorNetworkRaw struct {
	Contract *IBancorNetwork // Generic contract binding to access the raw methods on
}

// IBancorNetworkCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IBancorNetworkCallerRaw struct {
	Contract *IBancorNetworkCaller // Generic read-only contract binding to access the raw methods on
}

// IBancorNetworkTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IBancorNetworkTransactorRaw struct {
	Contract *IBancorNetworkTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIBancorNetwork creates a new instance of IBancorNetwork, bound to a specific deployed contract.
func NewIBancorNetwork(address common.Address, backend bind.ContractBackend) (*IBancorNetwork, error) {
	contract, err := bindIBancorNetwork(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IBancorNetwork{IBancorNetworkCaller: IBancorNetworkCaller{contract: contract}, IBancorNetworkTransactor: IBancorNetworkTransactor{contract: contract}, IBancorNetworkFilterer: IBancorNetworkFilterer{contract: contract}}, nil
}

// NewIBancorNetworkCaller creates a new read-only instance of IBancorNetwork, bound to a specific deployed contract.
func NewIBancorNetworkCaller(address common.Address, caller bind.ContractCaller) (*IBancorNetworkCaller, error) {
	contract, err := bindIBancorNetwork(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IBancorNetworkCaller{contract: contract}, nil
}

// NewIBancorNetworkTransactor creates a new write-only instance of IBancorNetwork, bound to a specific deployed contract.
func NewIBancorNetworkTransactor(address common.Address, transactor bind.ContractTransactor) (*IBancorNetworkTransactor, error) {
	contract, err := bindIBancorNetwork(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IBancorNetworkTransactor{contract: contract}, nil
}

// NewIBancorNetworkFilterer creates a new log filterer instance of IBancorNetwork, bound to a specific deployed contract.
func NewIBancorNetworkFilterer(address common.Address, filterer bind.ContractFilterer) (*IBancorNetworkFilterer, error) {
	contract, err := bindIBancorNetwork(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IBancorNetworkFilterer{contract: contract}, nil
}

// bindIBancorNetwork binds a generic wrapper to an already deployed contract.
func bindIBancorNetwork(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IBancorNetworkABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IBancorNetwork *IBancorNetworkRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IBancorNetwork.Contract.IBancorNetworkCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IBancorNetwork *IBancorNetworkRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.IBancorNetworkTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IBancorNetwork *IBancorNetworkRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.IBancorNetworkTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IBancorNetwork *IBancorNetworkCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IBancorNetwork.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IBancorNetwork *IBancorNetworkTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IBancorNetwork *IBancorNetworkTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.contract.Transact(opts, method, params...)
}

// ConversionPath is a free data retrieval call binding the contract method 0xd734fa19.
//
// Solidity: function conversionPath(address sourceToken, address targetToken) view returns(address[])
func (_IBancorNetwork *IBancorNetworkCaller) ConversionPath(opts *bind.CallOpts, sourceToken common.Address, targetToken common.Address) ([]common.Address, error) {
	var out []interface{}
	err := _IBancorNetwork.contract.Call(opts, &out, "conversionPath", sourceToken, targetToken)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// ConversionPath is a free data retrieval call binding the contract method 0xd734fa19.
//
// Solidity: function conversionPath(address sourceToken, address targetToken) view returns(address[])
func (_IBancorNetwork *IBancorNetworkSession) ConversionPath(sourceToken common.Address, targetToken common.Address) ([]common.Address, error) {
	return _IBancorNetwork.Contract.ConversionPath(&_IBancorNetwork.CallOpts, sourceToken, targetToken)
}

// ConversionPath is a free data retrieval call binding the contract method 0xd734fa19.
//
// Solidity: function conversionPath(address sourceToken, address targetToken) view returns(address[])
func (_IBancorNetwork *IBancorNetworkCallerSession) ConversionPath(sourceToken common.Address, targetToken common.Address) ([]common.Address, error) {
	return _IBancorNetwork.Contract.ConversionPath(&_IBancorNetwork.CallOpts, sourceToken, targetToken)
}

// RateByPath is a free data retrieval call binding the contract method 0x7f9c0ecd.
//
// Solidity: function rateByPath(address[] path, uint256 sourceAmount) view returns(uint256)
func (_IBancorNetwork *IBancorNetworkCaller) RateByPath(opts *bind.CallOpts, path []common.Address, sourceAmount *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _IBancorNetwork.contract.Call(opts, &out, "rateByPath", path, sourceAmount)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RateByPath is a free data retrieval call binding the contract method 0x7f9c0ecd.
//
// Solidity: function rateByPath(address[] path, uint256 sourceAmount) view returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) RateByPath(path []common.Address, sourceAmount *big.Int) (*big.Int, error) {
	return _IBancorNetwork.Contract.RateByPath(&_IBancorNetwork.CallOpts, path, sourceAmount)
}

// RateByPath is a free data retrieval call binding the contract method 0x7f9c0ecd.
//
// Solidity: function rateByPath(address[] path, uint256 sourceAmount) view returns(uint256)
func (_IBancorNetwork *IBancorNetworkCallerSession) RateByPath(path []common.Address, sourceAmount *big.Int) (*big.Int, error) {
	return _IBancorNetwork.Contract.RateByPath(&_IBancorNetwork.CallOpts, path, sourceAmount)
}

// ConvertByPath is a paid mutator transaction binding the contract method 0xb77d239b.
//
// Solidity: function convertByPath(address[] path, uint256 sourceAmount, uint256 minReturn, address beneficiary, address affiliateAccount, uint256 affiliateFee) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) ConvertByPath(opts *bind.TransactOpts, path []common.Address, sourceAmount *big.Int, minReturn *big.Int, beneficiary common.Address, affiliateAccount common.Address, affiliateFee *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "convertByPath", path, sourceAmount, minReturn, beneficiary, affiliateAccount, affiliateFee)
}

// ConvertByPath is a paid mutator transaction binding the contract method 0xb77d239b.
//
// Solidity: function convertByPath(address[] path, uint256 sourceAmount, uint256 minReturn, address beneficiary, address affiliateAccount, uint256 affiliateFee) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) ConvertByPath(path []common.Address, sourceAmount *big.Int, minReturn *big.Int, beneficiary common.Address, affiliateAccount common.Address, affiliateFee *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.ConvertByPath(&_IBancorNetwork.TransactOpts, path, sourceAmount, minReturn, beneficiary, affiliateAccount, affiliateFee)
}

// ConvertByPath is a paid mutator transaction binding the contract method 0xb77d239b.
//
// Solidity: function convertByPath(address[] path, uint256 sourceAmount, uint256 minReturn, address beneficiary, address affiliateAccount, uint256 affiliateFee) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) ConvertByPath(path []common.Address, sourceAmount *big.Int, minReturn *big.Int, beneficiary common.Address, affiliateAccount common.Address, affiliateFee *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.ConvertByPath(&_IBancorNetwork.TransactOpts, path, sourceAmount, minReturn, beneficiary, affiliateAccount, affiliateFee)
}

// ConvertByPath2 is a paid mutator transaction binding the contract method 0xf8394fb7.
//
// Solidity: function convertByPath2(address[] path, uint256 sourceAmount, uint256 minReturn, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) ConvertByPath2(opts *bind.TransactOpts, path []common.Address, sourceAmount *big.Int, minReturn *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "convertByPath2", path, sourceAmount, minReturn, beneficiary)
}

// ConvertByPath2 is a paid mutator transaction binding the contract method 0xf8394fb7.
//
// Solidity: function convertByPath2(address[] path, uint256 sourceAmount, uint256 minReturn, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) ConvertByPath2(path []common.Address, sourceAmount *big.Int, minReturn *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.ConvertByPath2(&_IBancorNetwork.TransactOpts, path, sourceAmount, minReturn, beneficiary)
}

// ConvertByPath2 is a paid mutator transaction binding the contract method 0xf8394fb7.
//
// Solidity: function convertByPath2(address[] path, uint256 sourceAmount, uint256 minReturn, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) ConvertByPath2(path []common.Address, sourceAmount *big.Int, minReturn *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.ConvertByPath2(&_IBancorNetwork.TransactOpts, path, sourceAmount, minReturn, beneficiary)
}

// IReserveTokenMetaData contains all meta data concerning the IReserveToken contract.
var IReserveTokenMetaData = &bind.MetaData{
	ABI: "[]",
}

// IReserveTokenABI is the input ABI used to generate the binding from.
// Deprecated: Use IReserveTokenMetaData.ABI instead.
var IReserveTokenABI = IReserveTokenMetaData.ABI

// IReserveToken is an auto generated Go binding around an Ethereum contract.
type IReserveToken struct {
	IReserveTokenCaller     // Read-only binding to the contract
	IReserveTokenTransactor // Write-only binding to the contract
	IReserveTokenFilterer   // Log filterer for contract events
}

// IReserveTokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type IReserveTokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IReserveTokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IReserveTokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IReserveTokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IReserveTokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IReserveTokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IReserveTokenSession struct {
	Contract     *IReserveToken    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IReserveTokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IReserveTokenCallerSession struct {
	Contract *IReserveTokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// IReserveTokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IReserveTokenTransactorSession struct {
	Contract     *IReserveTokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// IReserveTokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type IReserveTokenRaw struct {
	Contract *IReserveToken // Generic contract binding to access the raw methods on
}

// IReserveTokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IReserveTokenCallerRaw struct {
	Contract *IReserveTokenCaller // Generic read-only contract binding to access the raw methods on
}

// IReserveTokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IReserveTokenTransactorRaw struct {
	Contract *IReserveTokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIReserveToken creates a new instance of IReserveToken, bound to a specific deployed contract.
func NewIReserveToken(address common.Address, backend bind.ContractBackend) (*IReserveToken, error) {
	contract, err := bindIReserveToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IReserveToken{IReserveTokenCaller: IReserveTokenCaller{contract: contract}, IReserveTokenTransactor: IReserveTokenTransactor{contract: contract}, IReserveTokenFilterer: IReserveTokenFilterer{contract: contract}}, nil
}

// NewIReserveTokenCaller creates a new read-only instance of IReserveToken, bound to a specific deployed contract.
func NewIReserveTokenCaller(address common.Address, caller bind.ContractCaller) (*IReserveTokenCaller, error) {
	contract, err := bindIReserveToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IReserveTokenCaller{contract: contract}, nil
}

// NewIReserveTokenTransactor creates a new write-only instance of IReserveToken, bound to a specific deployed contract.
func NewIReserveTokenTransactor(address common.Address, transactor bind.ContractTransactor) (*IReserveTokenTransactor, error) {
	contract, err := bindIReserveToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IReserveTokenTransactor{contract: contract}, nil
}

// NewIReserveTokenFilterer creates a new log filterer instance of IReserveToken, bound to a specific deployed contract.
func NewIReserveTokenFilterer(address common.Address, filterer bind.ContractFilterer) (*IReserveTokenFilterer, error) {
	contract, err := bindIReserveToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IReserveTokenFilterer{contract: contract}, nil
}

// bindIReserveToken binds a generic wrapper to an already deployed contract.
func bindIReserveToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IReserveTokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IReserveToken *IReserveTokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IReserveToken.Contract.IReserveTokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IReserveToken *IReserveTokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IReserveToken.Contract.IReserveTokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IReserveToken *IReserveTokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IReserveToken.Contract.IReserveTokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IReserveToken *IReserveTokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IReserveToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IReserveToken *IReserveTokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IReserveToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IReserveToken *IReserveTokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IReserveToken.Contract.contract.Transact(opts, method, params...)
}
