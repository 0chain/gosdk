// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package uniswapnetwork

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

// UniswapMetaData contains all meta data concerning the Uniswap contract.
var UniswapMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"msg\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"v\",\"type\":\"uint256\"}],\"name\":\"DebugMsg\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"zcnAmount\",\"type\":\"uint256\"}],\"name\":\"getEstimatedETHforZCN\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOutMin\",\"type\":\"uint256\"}],\"name\":\"swapETHForZCNExactAmountIn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOut\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOutDesired\",\"type\":\"uint256\"}],\"name\":\"swapETHForZCNExactAmountOut\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOut\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountIn\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountOutMin\",\"type\":\"uint256\"}],\"name\":\"swapUSDCForZCNExactAmountIn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOut\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOutDesired\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountInMax\",\"type\":\"uint256\"}],\"name\":\"swapUSDCForZCNExactAmountOut\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOut\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountIn\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountOutMin\",\"type\":\"uint256\"}],\"name\":\"swapZCNForUSDCExactAmountIn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOut\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOutDesired\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountInMax\",\"type\":\"uint256\"}],\"name\":\"swapZCNForUSDCExactAmountOut\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOut\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountIn\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountOutMin\",\"type\":\"uint256\"}],\"name\":\"swapZCNForWETHExactAmountIn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOut\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOutDesired\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountInMax\",\"type\":\"uint256\"}],\"name\":\"swapZCNForWETHExactAmountOut\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOut\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// UniswapABI is the input ABI used to generate the binding from.
// Deprecated: Use UniswapMetaData.ABI instead.
var UniswapABI = UniswapMetaData.ABI

// Uniswap is an auto generated Go binding around an Ethereum contract.
type Uniswap struct {
	UniswapCaller     // Read-only binding to the contract
	UniswapTransactor // Write-only binding to the contract
	UniswapFilterer   // Log filterer for contract events
}

// UniswapCaller is an auto generated read-only Go binding around an Ethereum contract.
type UniswapCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapTransactor is an auto generated write-only Go binding around an Ethereum contract.
type UniswapTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type UniswapFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type UniswapSession struct {
	Contract     *Uniswap          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// UniswapCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type UniswapCallerSession struct {
	Contract *UniswapCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// UniswapTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type UniswapTransactorSession struct {
	Contract     *UniswapTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// UniswapRaw is an auto generated low-level Go binding around an Ethereum contract.
type UniswapRaw struct {
	Contract *Uniswap // Generic contract binding to access the raw methods on
}

// UniswapCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type UniswapCallerRaw struct {
	Contract *UniswapCaller // Generic read-only contract binding to access the raw methods on
}

// UniswapTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type UniswapTransactorRaw struct {
	Contract *UniswapTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUniswap creates a new instance of Uniswap, bound to a specific deployed contract.
func NewUniswap(address common.Address, backend bind.ContractBackend) (*Uniswap, error) {
	contract, err := bindUniswap(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Uniswap{UniswapCaller: UniswapCaller{contract: contract}, UniswapTransactor: UniswapTransactor{contract: contract}, UniswapFilterer: UniswapFilterer{contract: contract}}, nil
}

// NewUniswapCaller creates a new read-only instance of Uniswap, bound to a specific deployed contract.
func NewUniswapCaller(address common.Address, caller bind.ContractCaller) (*UniswapCaller, error) {
	contract, err := bindUniswap(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &UniswapCaller{contract: contract}, nil
}

// NewUniswapTransactor creates a new write-only instance of Uniswap, bound to a specific deployed contract.
func NewUniswapTransactor(address common.Address, transactor bind.ContractTransactor) (*UniswapTransactor, error) {
	contract, err := bindUniswap(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &UniswapTransactor{contract: contract}, nil
}

// NewUniswapFilterer creates a new log filterer instance of Uniswap, bound to a specific deployed contract.
func NewUniswapFilterer(address common.Address, filterer bind.ContractFilterer) (*UniswapFilterer, error) {
	contract, err := bindUniswap(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &UniswapFilterer{contract: contract}, nil
}

// bindUniswap binds a generic wrapper to an already deployed contract.
func bindUniswap(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := UniswapMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Uniswap *UniswapRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Uniswap.Contract.UniswapCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Uniswap *UniswapRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Uniswap.Contract.UniswapTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Uniswap *UniswapRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Uniswap.Contract.UniswapTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Uniswap *UniswapCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Uniswap.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Uniswap *UniswapTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Uniswap.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Uniswap *UniswapTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Uniswap.Contract.contract.Transact(opts, method, params...)
}

// GetEstimatedETHforZCN is a free data retrieval call binding the contract method 0x1a34ff1c.
//
// Solidity: function getEstimatedETHforZCN(uint256 zcnAmount) view returns(uint256[])
func (_Uniswap *UniswapCaller) GetEstimatedETHforZCN(opts *bind.CallOpts, zcnAmount *big.Int) ([]*big.Int, error) {
	var out []interface{}
	err := _Uniswap.contract.Call(opts, &out, "getEstimatedETHforZCN", zcnAmount)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetEstimatedETHforZCN is a free data retrieval call binding the contract method 0x1a34ff1c.
//
// Solidity: function getEstimatedETHforZCN(uint256 zcnAmount) view returns(uint256[])
func (_Uniswap *UniswapSession) GetEstimatedETHforZCN(zcnAmount *big.Int) ([]*big.Int, error) {
	return _Uniswap.Contract.GetEstimatedETHforZCN(&_Uniswap.CallOpts, zcnAmount)
}

// GetEstimatedETHforZCN is a free data retrieval call binding the contract method 0x1a34ff1c.
//
// Solidity: function getEstimatedETHforZCN(uint256 zcnAmount) view returns(uint256[])
func (_Uniswap *UniswapCallerSession) GetEstimatedETHforZCN(zcnAmount *big.Int) ([]*big.Int, error) {
	return _Uniswap.Contract.GetEstimatedETHforZCN(&_Uniswap.CallOpts, zcnAmount)
}

// SwapETHForZCNExactAmountIn is a paid mutator transaction binding the contract method 0xb33d99b1.
//
// Solidity: function swapETHForZCNExactAmountIn(uint256 amountOutMin) payable returns(uint256 amountOut)
func (_Uniswap *UniswapTransactor) SwapETHForZCNExactAmountIn(opts *bind.TransactOpts, amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.contract.Transact(opts, "swapETHForZCNExactAmountIn", amountOutMin)
}

// SwapETHForZCNExactAmountIn is a paid mutator transaction binding the contract method 0xb33d99b1.
//
// Solidity: function swapETHForZCNExactAmountIn(uint256 amountOutMin) payable returns(uint256 amountOut)
func (_Uniswap *UniswapSession) SwapETHForZCNExactAmountIn(amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapETHForZCNExactAmountIn(&_Uniswap.TransactOpts, amountOutMin)
}

// SwapETHForZCNExactAmountIn is a paid mutator transaction binding the contract method 0xb33d99b1.
//
// Solidity: function swapETHForZCNExactAmountIn(uint256 amountOutMin) payable returns(uint256 amountOut)
func (_Uniswap *UniswapTransactorSession) SwapETHForZCNExactAmountIn(amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapETHForZCNExactAmountIn(&_Uniswap.TransactOpts, amountOutMin)
}

// SwapETHForZCNExactAmountOut is a paid mutator transaction binding the contract method 0x18ae74a4.
//
// Solidity: function swapETHForZCNExactAmountOut(uint256 amountOutDesired) payable returns(uint256 amountOut)
func (_Uniswap *UniswapTransactor) SwapETHForZCNExactAmountOut(opts *bind.TransactOpts, amountOutDesired *big.Int) (*types.Transaction, error) {
	return _Uniswap.contract.Transact(opts, "swapETHForZCNExactAmountOut", amountOutDesired)
}

// SwapETHForZCNExactAmountOut is a paid mutator transaction binding the contract method 0x18ae74a4.
//
// Solidity: function swapETHForZCNExactAmountOut(uint256 amountOutDesired) payable returns(uint256 amountOut)
func (_Uniswap *UniswapSession) SwapETHForZCNExactAmountOut(amountOutDesired *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapETHForZCNExactAmountOut(&_Uniswap.TransactOpts, amountOutDesired)
}

// SwapETHForZCNExactAmountOut is a paid mutator transaction binding the contract method 0x18ae74a4.
//
// Solidity: function swapETHForZCNExactAmountOut(uint256 amountOutDesired) payable returns(uint256 amountOut)
func (_Uniswap *UniswapTransactorSession) SwapETHForZCNExactAmountOut(amountOutDesired *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapETHForZCNExactAmountOut(&_Uniswap.TransactOpts, amountOutDesired)
}

// SwapUSDCForZCNExactAmountIn is a paid mutator transaction binding the contract method 0x0976c3c2.
//
// Solidity: function swapUSDCForZCNExactAmountIn(uint256 amountIn, uint256 amountOutMin) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactor) SwapUSDCForZCNExactAmountIn(opts *bind.TransactOpts, amountIn *big.Int, amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.contract.Transact(opts, "swapUSDCForZCNExactAmountIn", amountIn, amountOutMin)
}

// SwapUSDCForZCNExactAmountIn is a paid mutator transaction binding the contract method 0x0976c3c2.
//
// Solidity: function swapUSDCForZCNExactAmountIn(uint256 amountIn, uint256 amountOutMin) returns(uint256 amountOut)
func (_Uniswap *UniswapSession) SwapUSDCForZCNExactAmountIn(amountIn *big.Int, amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapUSDCForZCNExactAmountIn(&_Uniswap.TransactOpts, amountIn, amountOutMin)
}

// SwapUSDCForZCNExactAmountIn is a paid mutator transaction binding the contract method 0x0976c3c2.
//
// Solidity: function swapUSDCForZCNExactAmountIn(uint256 amountIn, uint256 amountOutMin) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactorSession) SwapUSDCForZCNExactAmountIn(amountIn *big.Int, amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapUSDCForZCNExactAmountIn(&_Uniswap.TransactOpts, amountIn, amountOutMin)
}

// SwapUSDCForZCNExactAmountOut is a paid mutator transaction binding the contract method 0x97a40b34.
//
// Solidity: function swapUSDCForZCNExactAmountOut(uint256 amountOutDesired, uint256 amountInMax) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactor) SwapUSDCForZCNExactAmountOut(opts *bind.TransactOpts, amountOutDesired *big.Int, amountInMax *big.Int) (*types.Transaction, error) {
	return _Uniswap.contract.Transact(opts, "swapUSDCForZCNExactAmountOut", amountOutDesired, amountInMax)
}

// SwapUSDCForZCNExactAmountOut is a paid mutator transaction binding the contract method 0x97a40b34.
//
// Solidity: function swapUSDCForZCNExactAmountOut(uint256 amountOutDesired, uint256 amountInMax) returns(uint256 amountOut)
func (_Uniswap *UniswapSession) SwapUSDCForZCNExactAmountOut(amountOutDesired *big.Int, amountInMax *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapUSDCForZCNExactAmountOut(&_Uniswap.TransactOpts, amountOutDesired, amountInMax)
}

// SwapUSDCForZCNExactAmountOut is a paid mutator transaction binding the contract method 0x97a40b34.
//
// Solidity: function swapUSDCForZCNExactAmountOut(uint256 amountOutDesired, uint256 amountInMax) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactorSession) SwapUSDCForZCNExactAmountOut(amountOutDesired *big.Int, amountInMax *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapUSDCForZCNExactAmountOut(&_Uniswap.TransactOpts, amountOutDesired, amountInMax)
}

// SwapZCNForUSDCExactAmountIn is a paid mutator transaction binding the contract method 0xe60b51b6.
//
// Solidity: function swapZCNForUSDCExactAmountIn(uint256 amountIn, uint256 amountOutMin) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactor) SwapZCNForUSDCExactAmountIn(opts *bind.TransactOpts, amountIn *big.Int, amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.contract.Transact(opts, "swapZCNForUSDCExactAmountIn", amountIn, amountOutMin)
}

// SwapZCNForUSDCExactAmountIn is a paid mutator transaction binding the contract method 0xe60b51b6.
//
// Solidity: function swapZCNForUSDCExactAmountIn(uint256 amountIn, uint256 amountOutMin) returns(uint256 amountOut)
func (_Uniswap *UniswapSession) SwapZCNForUSDCExactAmountIn(amountIn *big.Int, amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapZCNForUSDCExactAmountIn(&_Uniswap.TransactOpts, amountIn, amountOutMin)
}

// SwapZCNForUSDCExactAmountIn is a paid mutator transaction binding the contract method 0xe60b51b6.
//
// Solidity: function swapZCNForUSDCExactAmountIn(uint256 amountIn, uint256 amountOutMin) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactorSession) SwapZCNForUSDCExactAmountIn(amountIn *big.Int, amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapZCNForUSDCExactAmountIn(&_Uniswap.TransactOpts, amountIn, amountOutMin)
}

// SwapZCNForUSDCExactAmountOut is a paid mutator transaction binding the contract method 0x4becb631.
//
// Solidity: function swapZCNForUSDCExactAmountOut(uint256 amountOutDesired, uint256 amountInMax) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactor) SwapZCNForUSDCExactAmountOut(opts *bind.TransactOpts, amountOutDesired *big.Int, amountInMax *big.Int) (*types.Transaction, error) {
	return _Uniswap.contract.Transact(opts, "swapZCNForUSDCExactAmountOut", amountOutDesired, amountInMax)
}

// SwapZCNForUSDCExactAmountOut is a paid mutator transaction binding the contract method 0x4becb631.
//
// Solidity: function swapZCNForUSDCExactAmountOut(uint256 amountOutDesired, uint256 amountInMax) returns(uint256 amountOut)
func (_Uniswap *UniswapSession) SwapZCNForUSDCExactAmountOut(amountOutDesired *big.Int, amountInMax *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapZCNForUSDCExactAmountOut(&_Uniswap.TransactOpts, amountOutDesired, amountInMax)
}

// SwapZCNForUSDCExactAmountOut is a paid mutator transaction binding the contract method 0x4becb631.
//
// Solidity: function swapZCNForUSDCExactAmountOut(uint256 amountOutDesired, uint256 amountInMax) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactorSession) SwapZCNForUSDCExactAmountOut(amountOutDesired *big.Int, amountInMax *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapZCNForUSDCExactAmountOut(&_Uniswap.TransactOpts, amountOutDesired, amountInMax)
}

// SwapZCNForWETHExactAmountIn is a paid mutator transaction binding the contract method 0x50a6cd6f.
//
// Solidity: function swapZCNForWETHExactAmountIn(uint256 amountIn, uint256 amountOutMin) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactor) SwapZCNForWETHExactAmountIn(opts *bind.TransactOpts, amountIn *big.Int, amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.contract.Transact(opts, "swapZCNForWETHExactAmountIn", amountIn, amountOutMin)
}

// SwapZCNForWETHExactAmountIn is a paid mutator transaction binding the contract method 0x50a6cd6f.
//
// Solidity: function swapZCNForWETHExactAmountIn(uint256 amountIn, uint256 amountOutMin) returns(uint256 amountOut)
func (_Uniswap *UniswapSession) SwapZCNForWETHExactAmountIn(amountIn *big.Int, amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapZCNForWETHExactAmountIn(&_Uniswap.TransactOpts, amountIn, amountOutMin)
}

// SwapZCNForWETHExactAmountIn is a paid mutator transaction binding the contract method 0x50a6cd6f.
//
// Solidity: function swapZCNForWETHExactAmountIn(uint256 amountIn, uint256 amountOutMin) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactorSession) SwapZCNForWETHExactAmountIn(amountIn *big.Int, amountOutMin *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapZCNForWETHExactAmountIn(&_Uniswap.TransactOpts, amountIn, amountOutMin)
}

// SwapZCNForWETHExactAmountOut is a paid mutator transaction binding the contract method 0xaae07c3e.
//
// Solidity: function swapZCNForWETHExactAmountOut(uint256 amountOutDesired, uint256 amountInMax) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactor) SwapZCNForWETHExactAmountOut(opts *bind.TransactOpts, amountOutDesired *big.Int, amountInMax *big.Int) (*types.Transaction, error) {
	return _Uniswap.contract.Transact(opts, "swapZCNForWETHExactAmountOut", amountOutDesired, amountInMax)
}

// SwapZCNForWETHExactAmountOut is a paid mutator transaction binding the contract method 0xaae07c3e.
//
// Solidity: function swapZCNForWETHExactAmountOut(uint256 amountOutDesired, uint256 amountInMax) returns(uint256 amountOut)
func (_Uniswap *UniswapSession) SwapZCNForWETHExactAmountOut(amountOutDesired *big.Int, amountInMax *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapZCNForWETHExactAmountOut(&_Uniswap.TransactOpts, amountOutDesired, amountInMax)
}

// SwapZCNForWETHExactAmountOut is a paid mutator transaction binding the contract method 0xaae07c3e.
//
// Solidity: function swapZCNForWETHExactAmountOut(uint256 amountOutDesired, uint256 amountInMax) returns(uint256 amountOut)
func (_Uniswap *UniswapTransactorSession) SwapZCNForWETHExactAmountOut(amountOutDesired *big.Int, amountInMax *big.Int) (*types.Transaction, error) {
	return _Uniswap.Contract.SwapZCNForWETHExactAmountOut(&_Uniswap.TransactOpts, amountOutDesired, amountInMax)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Uniswap *UniswapTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Uniswap.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Uniswap *UniswapSession) Receive() (*types.Transaction, error) {
	return _Uniswap.Contract.Receive(&_Uniswap.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Uniswap *UniswapTransactorSession) Receive() (*types.Transaction, error) {
	return _Uniswap.Contract.Receive(&_Uniswap.TransactOpts)
}

// UniswapDebugMsgIterator is returned from FilterDebugMsg and is used to iterate over the raw logs and unpacked data for DebugMsg events raised by the Uniswap contract.
type UniswapDebugMsgIterator struct {
	Event *UniswapDebugMsg // Event containing the contract specifics and raw log

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
func (it *UniswapDebugMsgIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UniswapDebugMsg)
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
		it.Event = new(UniswapDebugMsg)
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
func (it *UniswapDebugMsgIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UniswapDebugMsgIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UniswapDebugMsg represents a DebugMsg event raised by the Uniswap contract.
type UniswapDebugMsg struct {
	Msg string
	V   *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterDebugMsg is a free log retrieval operation binding the contract event 0xea30ed2bcbf4b7c5487f2c07ef639257ebf04932960dff4496fb914f769a6439.
//
// Solidity: event DebugMsg(string msg, uint256 v)
func (_Uniswap *UniswapFilterer) FilterDebugMsg(opts *bind.FilterOpts) (*UniswapDebugMsgIterator, error) {

	logs, sub, err := _Uniswap.contract.FilterLogs(opts, "DebugMsg")
	if err != nil {
		return nil, err
	}
	return &UniswapDebugMsgIterator{contract: _Uniswap.contract, event: "DebugMsg", logs: logs, sub: sub}, nil
}

// WatchDebugMsg is a free log subscription operation binding the contract event 0xea30ed2bcbf4b7c5487f2c07ef639257ebf04932960dff4496fb914f769a6439.
//
// Solidity: event DebugMsg(string msg, uint256 v)
func (_Uniswap *UniswapFilterer) WatchDebugMsg(opts *bind.WatchOpts, sink chan<- *UniswapDebugMsg) (event.Subscription, error) {

	logs, sub, err := _Uniswap.contract.WatchLogs(opts, "DebugMsg")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UniswapDebugMsg)
				if err := _Uniswap.contract.UnpackLog(event, "DebugMsg", log); err != nil {
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

// ParseDebugMsg is a log parse operation binding the contract event 0xea30ed2bcbf4b7c5487f2c07ef639257ebf04932960dff4496fb914f769a6439.
//
// Solidity: event DebugMsg(string msg, uint256 v)
func (_Uniswap *UniswapFilterer) ParseDebugMsg(log types.Log) (*UniswapDebugMsg, error) {
	event := new(UniswapDebugMsg)
	if err := _Uniswap.contract.UnpackLog(event, "DebugMsg", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
