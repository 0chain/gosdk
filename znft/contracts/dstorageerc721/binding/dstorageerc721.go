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
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"uri_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"previous\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"updated\",\"type\":\"uint256\"}],\"name\":\"MaxUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"previous\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"updated\",\"type\":\"bool\"}],\"name\":\"MintableUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previous\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"updated\",\"type\":\"address\"}],\"name\":\"ReceiverUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"previous\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"updated\",\"type\":\"uint256\"}],\"name\":\"RoyaltyUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"previous\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"updated\",\"type\":\"string\"}],\"name\":\"UriFallbackUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"previous\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"updated\",\"type\":\"string\"}],\"name\":\"UriUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"allocation\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"max\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mintOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"mintable\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"price\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"receiver\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"royalty\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"salePrice\",\"type\":\"uint256\"}],\"name\":\"royaltyInfo\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"allocation_\",\"type\":\"string\"}],\"name\":\"setAllocation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"max_\",\"type\":\"uint256\"}],\"name\":\"setMax\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"status_\",\"type\":\"bool\"}],\"name\":\"setMintable\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"receiver_\",\"type\":\"address\"}],\"name\":\"setReceiver\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"royalty_\",\"type\":\"uint256\"}],\"name\":\"setRoyalty\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"uri_\",\"type\":\"string\"}],\"name\":\"setURI\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"uri_\",\"type\":\"string\"}],\"name\":\"setURIFallback\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURIFallback\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"total\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"uri\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"uriFallback\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// Allocation is a free data retrieval call binding the contract method 0x88a17bde.
//
// Solidity: function allocation() view returns(string)
func (_Binding *BindingCaller) Allocation(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "allocation")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Allocation is a free data retrieval call binding the contract method 0x88a17bde.
//
// Solidity: function allocation() view returns(string)
func (_Binding *BindingSession) Allocation() (string, error) {
	return _Binding.Contract.Allocation(&_Binding.CallOpts)
}

// Allocation is a free data retrieval call binding the contract method 0x88a17bde.
//
// Solidity: function allocation() view returns(string)
func (_Binding *BindingCallerSession) Allocation() (string, error) {
	return _Binding.Contract.Allocation(&_Binding.CallOpts)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_Binding *BindingCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_Binding *BindingSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _Binding.Contract.BalanceOf(&_Binding.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_Binding *BindingCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _Binding.Contract.BalanceOf(&_Binding.CallOpts, owner)
}

// Batch is a free data retrieval call binding the contract method 0xaf713566.
//
// Solidity: function batch() view returns(uint256)
func (_Binding *BindingCaller) Batch(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "batch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Batch is a free data retrieval call binding the contract method 0xaf713566.
//
// Solidity: function batch() view returns(uint256)
func (_Binding *BindingSession) Batch() (*big.Int, error) {
	return _Binding.Contract.Batch(&_Binding.CallOpts)
}

// Batch is a free data retrieval call binding the contract method 0xaf713566.
//
// Solidity: function batch() view returns(uint256)
func (_Binding *BindingCallerSession) Batch() (*big.Int, error) {
	return _Binding.Contract.Batch(&_Binding.CallOpts)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_Binding *BindingCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_Binding *BindingSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _Binding.Contract.GetApproved(&_Binding.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_Binding *BindingCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _Binding.Contract.GetApproved(&_Binding.CallOpts, tokenId)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_Binding *BindingCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_Binding *BindingSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _Binding.Contract.IsApprovedForAll(&_Binding.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_Binding *BindingCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _Binding.Contract.IsApprovedForAll(&_Binding.CallOpts, owner, operator)
}

// Max is a free data retrieval call binding the contract method 0x6ac5db19.
//
// Solidity: function max() view returns(uint256)
func (_Binding *BindingCaller) Max(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "max")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Max is a free data retrieval call binding the contract method 0x6ac5db19.
//
// Solidity: function max() view returns(uint256)
func (_Binding *BindingSession) Max() (*big.Int, error) {
	return _Binding.Contract.Max(&_Binding.CallOpts)
}

// Max is a free data retrieval call binding the contract method 0x6ac5db19.
//
// Solidity: function max() view returns(uint256)
func (_Binding *BindingCallerSession) Max() (*big.Int, error) {
	return _Binding.Contract.Max(&_Binding.CallOpts)
}

// Mintable is a free data retrieval call binding the contract method 0x4bf365df.
//
// Solidity: function mintable() view returns(bool)
func (_Binding *BindingCaller) Mintable(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "mintable")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Mintable is a free data retrieval call binding the contract method 0x4bf365df.
//
// Solidity: function mintable() view returns(bool)
func (_Binding *BindingSession) Mintable() (bool, error) {
	return _Binding.Contract.Mintable(&_Binding.CallOpts)
}

// Mintable is a free data retrieval call binding the contract method 0x4bf365df.
//
// Solidity: function mintable() view returns(bool)
func (_Binding *BindingCallerSession) Mintable() (bool, error) {
	return _Binding.Contract.Mintable(&_Binding.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Binding *BindingCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Binding *BindingSession) Name() (string, error) {
	return _Binding.Contract.Name(&_Binding.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Binding *BindingCallerSession) Name() (string, error) {
	return _Binding.Contract.Name(&_Binding.CallOpts)
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

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Binding *BindingCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Binding *BindingSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Binding.Contract.OwnerOf(&_Binding.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Binding *BindingCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Binding.Contract.OwnerOf(&_Binding.CallOpts, tokenId)
}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() view returns(uint256)
func (_Binding *BindingCaller) Price(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "price")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() view returns(uint256)
func (_Binding *BindingSession) Price() (*big.Int, error) {
	return _Binding.Contract.Price(&_Binding.CallOpts)
}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() view returns(uint256)
func (_Binding *BindingCallerSession) Price() (*big.Int, error) {
	return _Binding.Contract.Price(&_Binding.CallOpts)
}

// Receiver is a free data retrieval call binding the contract method 0xf7260d3e.
//
// Solidity: function receiver() view returns(address)
func (_Binding *BindingCaller) Receiver(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "receiver")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Receiver is a free data retrieval call binding the contract method 0xf7260d3e.
//
// Solidity: function receiver() view returns(address)
func (_Binding *BindingSession) Receiver() (common.Address, error) {
	return _Binding.Contract.Receiver(&_Binding.CallOpts)
}

// Receiver is a free data retrieval call binding the contract method 0xf7260d3e.
//
// Solidity: function receiver() view returns(address)
func (_Binding *BindingCallerSession) Receiver() (common.Address, error) {
	return _Binding.Contract.Receiver(&_Binding.CallOpts)
}

// Royalty is a free data retrieval call binding the contract method 0x29ee566c.
//
// Solidity: function royalty() view returns(uint256)
func (_Binding *BindingCaller) Royalty(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "royalty")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Royalty is a free data retrieval call binding the contract method 0x29ee566c.
//
// Solidity: function royalty() view returns(uint256)
func (_Binding *BindingSession) Royalty() (*big.Int, error) {
	return _Binding.Contract.Royalty(&_Binding.CallOpts)
}

// Royalty is a free data retrieval call binding the contract method 0x29ee566c.
//
// Solidity: function royalty() view returns(uint256)
func (_Binding *BindingCallerSession) Royalty() (*big.Int, error) {
	return _Binding.Contract.Royalty(&_Binding.CallOpts)
}

// RoyaltyInfo is a free data retrieval call binding the contract method 0x2a55205a.
//
// Solidity: function royaltyInfo(uint256 tokenId, uint256 salePrice) view returns(address, uint256)
func (_Binding *BindingCaller) RoyaltyInfo(opts *bind.CallOpts, tokenId *big.Int, salePrice *big.Int) (common.Address, *big.Int, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "royaltyInfo", tokenId, salePrice)

	if err != nil {
		return *new(common.Address), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// RoyaltyInfo is a free data retrieval call binding the contract method 0x2a55205a.
//
// Solidity: function royaltyInfo(uint256 tokenId, uint256 salePrice) view returns(address, uint256)
func (_Binding *BindingSession) RoyaltyInfo(tokenId *big.Int, salePrice *big.Int) (common.Address, *big.Int, error) {
	return _Binding.Contract.RoyaltyInfo(&_Binding.CallOpts, tokenId, salePrice)
}

// RoyaltyInfo is a free data retrieval call binding the contract method 0x2a55205a.
//
// Solidity: function royaltyInfo(uint256 tokenId, uint256 salePrice) view returns(address, uint256)
func (_Binding *BindingCallerSession) RoyaltyInfo(tokenId *big.Int, salePrice *big.Int) (common.Address, *big.Int, error) {
	return _Binding.Contract.RoyaltyInfo(&_Binding.CallOpts, tokenId, salePrice)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Binding *BindingCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Binding *BindingSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Binding.Contract.SupportsInterface(&_Binding.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Binding *BindingCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Binding.Contract.SupportsInterface(&_Binding.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Binding *BindingCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Binding *BindingSession) Symbol() (string, error) {
	return _Binding.Contract.Symbol(&_Binding.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Binding *BindingCallerSession) Symbol() (string, error) {
	return _Binding.Contract.Symbol(&_Binding.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Binding *BindingCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Binding *BindingSession) TokenURI(tokenId *big.Int) (string, error) {
	return _Binding.Contract.TokenURI(&_Binding.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Binding *BindingCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _Binding.Contract.TokenURI(&_Binding.CallOpts, tokenId)
}

// TokenURIFallback is a free data retrieval call binding the contract method 0xc7c8f564.
//
// Solidity: function tokenURIFallback(uint256 tokenId) view returns(string)
func (_Binding *BindingCaller) TokenURIFallback(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "tokenURIFallback", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURIFallback is a free data retrieval call binding the contract method 0xc7c8f564.
//
// Solidity: function tokenURIFallback(uint256 tokenId) view returns(string)
func (_Binding *BindingSession) TokenURIFallback(tokenId *big.Int) (string, error) {
	return _Binding.Contract.TokenURIFallback(&_Binding.CallOpts, tokenId)
}

// TokenURIFallback is a free data retrieval call binding the contract method 0xc7c8f564.
//
// Solidity: function tokenURIFallback(uint256 tokenId) view returns(string)
func (_Binding *BindingCallerSession) TokenURIFallback(tokenId *big.Int) (string, error) {
	return _Binding.Contract.TokenURIFallback(&_Binding.CallOpts, tokenId)
}

// Total is a free data retrieval call binding the contract method 0x2ddbd13a.
//
// Solidity: function total() view returns(uint256)
func (_Binding *BindingCaller) Total(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "total")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Total is a free data retrieval call binding the contract method 0x2ddbd13a.
//
// Solidity: function total() view returns(uint256)
func (_Binding *BindingSession) Total() (*big.Int, error) {
	return _Binding.Contract.Total(&_Binding.CallOpts)
}

// Total is a free data retrieval call binding the contract method 0x2ddbd13a.
//
// Solidity: function total() view returns(uint256)
func (_Binding *BindingCallerSession) Total() (*big.Int, error) {
	return _Binding.Contract.Total(&_Binding.CallOpts)
}

// Uri is a free data retrieval call binding the contract method 0xeac989f8.
//
// Solidity: function uri() view returns(string)
func (_Binding *BindingCaller) Uri(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "uri")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Uri is a free data retrieval call binding the contract method 0xeac989f8.
//
// Solidity: function uri() view returns(string)
func (_Binding *BindingSession) Uri() (string, error) {
	return _Binding.Contract.Uri(&_Binding.CallOpts)
}

// Uri is a free data retrieval call binding the contract method 0xeac989f8.
//
// Solidity: function uri() view returns(string)
func (_Binding *BindingCallerSession) Uri() (string, error) {
	return _Binding.Contract.Uri(&_Binding.CallOpts)
}

// UriFallback is a free data retrieval call binding the contract method 0x6dd8e21a.
//
// Solidity: function uriFallback() view returns(string)
func (_Binding *BindingCaller) UriFallback(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Binding.contract.Call(opts, &out, "uriFallback")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UriFallback is a free data retrieval call binding the contract method 0x6dd8e21a.
//
// Solidity: function uriFallback() view returns(string)
func (_Binding *BindingSession) UriFallback() (string, error) {
	return _Binding.Contract.UriFallback(&_Binding.CallOpts)
}

// UriFallback is a free data retrieval call binding the contract method 0x6dd8e21a.
//
// Solidity: function uriFallback() view returns(string)
func (_Binding *BindingCallerSession) UriFallback() (string, error) {
	return _Binding.Contract.UriFallback(&_Binding.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_Binding *BindingTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_Binding *BindingSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.Approve(&_Binding.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_Binding *BindingTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.Approve(&_Binding.TransactOpts, to, tokenId)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount) payable returns()
func (_Binding *BindingTransactor) Mint(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "mint", amount)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount) payable returns()
func (_Binding *BindingSession) Mint(amount *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.Mint(&_Binding.TransactOpts, amount)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount) payable returns()
func (_Binding *BindingTransactorSession) Mint(amount *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.Mint(&_Binding.TransactOpts, amount)
}

// MintOwner is a paid mutator transaction binding the contract method 0x33f88d22.
//
// Solidity: function mintOwner(uint256 amount) returns()
func (_Binding *BindingTransactor) MintOwner(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "mintOwner", amount)
}

// MintOwner is a paid mutator transaction binding the contract method 0x33f88d22.
//
// Solidity: function mintOwner(uint256 amount) returns()
func (_Binding *BindingSession) MintOwner(amount *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.MintOwner(&_Binding.TransactOpts, amount)
}

// MintOwner is a paid mutator transaction binding the contract method 0x33f88d22.
//
// Solidity: function mintOwner(uint256 amount) returns()
func (_Binding *BindingTransactorSession) MintOwner(amount *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.MintOwner(&_Binding.TransactOpts, amount)
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

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_Binding *BindingTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_Binding *BindingSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.SafeTransferFrom(&_Binding.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_Binding *BindingTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.SafeTransferFrom(&_Binding.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes _data) returns()
func (_Binding *BindingTransactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, _data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes _data) returns()
func (_Binding *BindingSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _Binding.Contract.SafeTransferFrom0(&_Binding.TransactOpts, from, to, tokenId, _data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes _data) returns()
func (_Binding *BindingTransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _Binding.Contract.SafeTransferFrom0(&_Binding.TransactOpts, from, to, tokenId, _data)
}

// SetAllocation is a paid mutator transaction binding the contract method 0x970a1fa8.
//
// Solidity: function setAllocation(string allocation_) returns()
func (_Binding *BindingTransactor) SetAllocation(opts *bind.TransactOpts, allocation_ string) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "setAllocation", allocation_)
}

// SetAllocation is a paid mutator transaction binding the contract method 0x970a1fa8.
//
// Solidity: function setAllocation(string allocation_) returns()
func (_Binding *BindingSession) SetAllocation(allocation_ string) (*types.Transaction, error) {
	return _Binding.Contract.SetAllocation(&_Binding.TransactOpts, allocation_)
}

// SetAllocation is a paid mutator transaction binding the contract method 0x970a1fa8.
//
// Solidity: function setAllocation(string allocation_) returns()
func (_Binding *BindingTransactorSession) SetAllocation(allocation_ string) (*types.Transaction, error) {
	return _Binding.Contract.SetAllocation(&_Binding.TransactOpts, allocation_)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_Binding *BindingTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_Binding *BindingSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _Binding.Contract.SetApprovalForAll(&_Binding.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_Binding *BindingTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _Binding.Contract.SetApprovalForAll(&_Binding.TransactOpts, operator, approved)
}

// SetMax is a paid mutator transaction binding the contract method 0x1fe9eabc.
//
// Solidity: function setMax(uint256 max_) returns()
func (_Binding *BindingTransactor) SetMax(opts *bind.TransactOpts, max_ *big.Int) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "setMax", max_)
}

// SetMax is a paid mutator transaction binding the contract method 0x1fe9eabc.
//
// Solidity: function setMax(uint256 max_) returns()
func (_Binding *BindingSession) SetMax(max_ *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.SetMax(&_Binding.TransactOpts, max_)
}

// SetMax is a paid mutator transaction binding the contract method 0x1fe9eabc.
//
// Solidity: function setMax(uint256 max_) returns()
func (_Binding *BindingTransactorSession) SetMax(max_ *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.SetMax(&_Binding.TransactOpts, max_)
}

// SetMintable is a paid mutator transaction binding the contract method 0x285d70d4.
//
// Solidity: function setMintable(bool status_) returns()
func (_Binding *BindingTransactor) SetMintable(opts *bind.TransactOpts, status_ bool) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "setMintable", status_)
}

// SetMintable is a paid mutator transaction binding the contract method 0x285d70d4.
//
// Solidity: function setMintable(bool status_) returns()
func (_Binding *BindingSession) SetMintable(status_ bool) (*types.Transaction, error) {
	return _Binding.Contract.SetMintable(&_Binding.TransactOpts, status_)
}

// SetMintable is a paid mutator transaction binding the contract method 0x285d70d4.
//
// Solidity: function setMintable(bool status_) returns()
func (_Binding *BindingTransactorSession) SetMintable(status_ bool) (*types.Transaction, error) {
	return _Binding.Contract.SetMintable(&_Binding.TransactOpts, status_)
}

// SetReceiver is a paid mutator transaction binding the contract method 0x718da7ee.
//
// Solidity: function setReceiver(address receiver_) returns()
func (_Binding *BindingTransactor) SetReceiver(opts *bind.TransactOpts, receiver_ common.Address) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "setReceiver", receiver_)
}

// SetReceiver is a paid mutator transaction binding the contract method 0x718da7ee.
//
// Solidity: function setReceiver(address receiver_) returns()
func (_Binding *BindingSession) SetReceiver(receiver_ common.Address) (*types.Transaction, error) {
	return _Binding.Contract.SetReceiver(&_Binding.TransactOpts, receiver_)
}

// SetReceiver is a paid mutator transaction binding the contract method 0x718da7ee.
//
// Solidity: function setReceiver(address receiver_) returns()
func (_Binding *BindingTransactorSession) SetReceiver(receiver_ common.Address) (*types.Transaction, error) {
	return _Binding.Contract.SetReceiver(&_Binding.TransactOpts, receiver_)
}

// SetRoyalty is a paid mutator transaction binding the contract method 0x4209a2e1.
//
// Solidity: function setRoyalty(uint256 royalty_) returns()
func (_Binding *BindingTransactor) SetRoyalty(opts *bind.TransactOpts, royalty_ *big.Int) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "setRoyalty", royalty_)
}

// SetRoyalty is a paid mutator transaction binding the contract method 0x4209a2e1.
//
// Solidity: function setRoyalty(uint256 royalty_) returns()
func (_Binding *BindingSession) SetRoyalty(royalty_ *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.SetRoyalty(&_Binding.TransactOpts, royalty_)
}

// SetRoyalty is a paid mutator transaction binding the contract method 0x4209a2e1.
//
// Solidity: function setRoyalty(uint256 royalty_) returns()
func (_Binding *BindingTransactorSession) SetRoyalty(royalty_ *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.SetRoyalty(&_Binding.TransactOpts, royalty_)
}

// SetURI is a paid mutator transaction binding the contract method 0x02fe5305.
//
// Solidity: function setURI(string uri_) returns()
func (_Binding *BindingTransactor) SetURI(opts *bind.TransactOpts, uri_ string) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "setURI", uri_)
}

// SetURI is a paid mutator transaction binding the contract method 0x02fe5305.
//
// Solidity: function setURI(string uri_) returns()
func (_Binding *BindingSession) SetURI(uri_ string) (*types.Transaction, error) {
	return _Binding.Contract.SetURI(&_Binding.TransactOpts, uri_)
}

// SetURI is a paid mutator transaction binding the contract method 0x02fe5305.
//
// Solidity: function setURI(string uri_) returns()
func (_Binding *BindingTransactorSession) SetURI(uri_ string) (*types.Transaction, error) {
	return _Binding.Contract.SetURI(&_Binding.TransactOpts, uri_)
}

// SetURIFallback is a paid mutator transaction binding the contract method 0x0c8ab6e1.
//
// Solidity: function setURIFallback(string uri_) returns()
func (_Binding *BindingTransactor) SetURIFallback(opts *bind.TransactOpts, uri_ string) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "setURIFallback", uri_)
}

// SetURIFallback is a paid mutator transaction binding the contract method 0x0c8ab6e1.
//
// Solidity: function setURIFallback(string uri_) returns()
func (_Binding *BindingSession) SetURIFallback(uri_ string) (*types.Transaction, error) {
	return _Binding.Contract.SetURIFallback(&_Binding.TransactOpts, uri_)
}

// SetURIFallback is a paid mutator transaction binding the contract method 0x0c8ab6e1.
//
// Solidity: function setURIFallback(string uri_) returns()
func (_Binding *BindingTransactorSession) SetURIFallback(uri_ string) (*types.Transaction, error) {
	return _Binding.Contract.SetURIFallback(&_Binding.TransactOpts, uri_)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_Binding *BindingTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_Binding *BindingSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.TransferFrom(&_Binding.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_Binding *BindingTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Binding.Contract.TransferFrom(&_Binding.TransactOpts, from, to, tokenId)
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

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_Binding *BindingTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Binding.contract.Transact(opts, "withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_Binding *BindingSession) Withdraw() (*types.Transaction, error) {
	return _Binding.Contract.Withdraw(&_Binding.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_Binding *BindingTransactorSession) Withdraw() (*types.Transaction, error) {
	return _Binding.Contract.Withdraw(&_Binding.TransactOpts)
}

// BindingApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the Binding contract.
type BindingApprovalIterator struct {
	Event *BindingApproval // Event containing the contract specifics and raw log

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
func (it *BindingApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingApproval)
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
		it.Event = new(BindingApproval)
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
func (it *BindingApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingApproval represents a Approval event raised by the Binding contract.
type BindingApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_Binding *BindingFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*BindingApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _Binding.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &BindingApprovalIterator{contract: _Binding.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_Binding *BindingFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *BindingApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _Binding.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingApproval)
				if err := _Binding.contract.UnpackLog(event, "Approval", log); err != nil {
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
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_Binding *BindingFilterer) ParseApproval(log types.Log) (*BindingApproval, error) {
	event := new(BindingApproval)
	if err := _Binding.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the Binding contract.
type BindingApprovalForAllIterator struct {
	Event *BindingApprovalForAll // Event containing the contract specifics and raw log

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
func (it *BindingApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingApprovalForAll)
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
		it.Event = new(BindingApprovalForAll)
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
func (it *BindingApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingApprovalForAll represents a ApprovalForAll event raised by the Binding contract.
type BindingApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Binding *BindingFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*BindingApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Binding.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &BindingApprovalForAllIterator{contract: _Binding.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Binding *BindingFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *BindingApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Binding.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingApprovalForAll)
				if err := _Binding.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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

// ParseApprovalForAll is a log parse operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Binding *BindingFilterer) ParseApprovalForAll(log types.Log) (*BindingApprovalForAll, error) {
	event := new(BindingApprovalForAll)
	if err := _Binding.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingMaxUpdatedIterator is returned from FilterMaxUpdated and is used to iterate over the raw logs and unpacked data for MaxUpdated events raised by the Binding contract.
type BindingMaxUpdatedIterator struct {
	Event *BindingMaxUpdated // Event containing the contract specifics and raw log

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
func (it *BindingMaxUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingMaxUpdated)
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
		it.Event = new(BindingMaxUpdated)
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
func (it *BindingMaxUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingMaxUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingMaxUpdated represents a MaxUpdated event raised by the Binding contract.
type BindingMaxUpdated struct {
	Previous *big.Int
	Updated  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMaxUpdated is a free log retrieval operation binding the contract event 0xaa6f6b0a509f2b07cf30d89dbd3bb410883aaa429ad4da41fdf36c02398cf1a0.
//
// Solidity: event MaxUpdated(uint256 previous, uint256 updated)
func (_Binding *BindingFilterer) FilterMaxUpdated(opts *bind.FilterOpts) (*BindingMaxUpdatedIterator, error) {

	logs, sub, err := _Binding.contract.FilterLogs(opts, "MaxUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingMaxUpdatedIterator{contract: _Binding.contract, event: "MaxUpdated", logs: logs, sub: sub}, nil
}

// WatchMaxUpdated is a free log subscription operation binding the contract event 0xaa6f6b0a509f2b07cf30d89dbd3bb410883aaa429ad4da41fdf36c02398cf1a0.
//
// Solidity: event MaxUpdated(uint256 previous, uint256 updated)
func (_Binding *BindingFilterer) WatchMaxUpdated(opts *bind.WatchOpts, sink chan<- *BindingMaxUpdated) (event.Subscription, error) {

	logs, sub, err := _Binding.contract.WatchLogs(opts, "MaxUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingMaxUpdated)
				if err := _Binding.contract.UnpackLog(event, "MaxUpdated", log); err != nil {
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

// ParseMaxUpdated is a log parse operation binding the contract event 0xaa6f6b0a509f2b07cf30d89dbd3bb410883aaa429ad4da41fdf36c02398cf1a0.
//
// Solidity: event MaxUpdated(uint256 previous, uint256 updated)
func (_Binding *BindingFilterer) ParseMaxUpdated(log types.Log) (*BindingMaxUpdated, error) {
	event := new(BindingMaxUpdated)
	if err := _Binding.contract.UnpackLog(event, "MaxUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingMintableUpdatedIterator is returned from FilterMintableUpdated and is used to iterate over the raw logs and unpacked data for MintableUpdated events raised by the Binding contract.
type BindingMintableUpdatedIterator struct {
	Event *BindingMintableUpdated // Event containing the contract specifics and raw log

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
func (it *BindingMintableUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingMintableUpdated)
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
		it.Event = new(BindingMintableUpdated)
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
func (it *BindingMintableUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingMintableUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingMintableUpdated represents a MintableUpdated event raised by the Binding contract.
type BindingMintableUpdated struct {
	Previous bool
	Updated  bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMintableUpdated is a free log retrieval operation binding the contract event 0x8d9383d773c0600295154578f39da3106938ba8d1fe1767bcfabe8bf05f555f4.
//
// Solidity: event MintableUpdated(bool previous, bool updated)
func (_Binding *BindingFilterer) FilterMintableUpdated(opts *bind.FilterOpts) (*BindingMintableUpdatedIterator, error) {

	logs, sub, err := _Binding.contract.FilterLogs(opts, "MintableUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingMintableUpdatedIterator{contract: _Binding.contract, event: "MintableUpdated", logs: logs, sub: sub}, nil
}

// WatchMintableUpdated is a free log subscription operation binding the contract event 0x8d9383d773c0600295154578f39da3106938ba8d1fe1767bcfabe8bf05f555f4.
//
// Solidity: event MintableUpdated(bool previous, bool updated)
func (_Binding *BindingFilterer) WatchMintableUpdated(opts *bind.WatchOpts, sink chan<- *BindingMintableUpdated) (event.Subscription, error) {

	logs, sub, err := _Binding.contract.WatchLogs(opts, "MintableUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingMintableUpdated)
				if err := _Binding.contract.UnpackLog(event, "MintableUpdated", log); err != nil {
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

// ParseMintableUpdated is a log parse operation binding the contract event 0x8d9383d773c0600295154578f39da3106938ba8d1fe1767bcfabe8bf05f555f4.
//
// Solidity: event MintableUpdated(bool previous, bool updated)
func (_Binding *BindingFilterer) ParseMintableUpdated(log types.Log) (*BindingMintableUpdated, error) {
	event := new(BindingMintableUpdated)
	if err := _Binding.contract.UnpackLog(event, "MintableUpdated", log); err != nil {
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

// BindingReceiverUpdatedIterator is returned from FilterReceiverUpdated and is used to iterate over the raw logs and unpacked data for ReceiverUpdated events raised by the Binding contract.
type BindingReceiverUpdatedIterator struct {
	Event *BindingReceiverUpdated // Event containing the contract specifics and raw log

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
func (it *BindingReceiverUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingReceiverUpdated)
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
		it.Event = new(BindingReceiverUpdated)
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
func (it *BindingReceiverUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingReceiverUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingReceiverUpdated represents a ReceiverUpdated event raised by the Binding contract.
type BindingReceiverUpdated struct {
	Previous common.Address
	Updated  common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterReceiverUpdated is a free log retrieval operation binding the contract event 0xbda2bcccbfa5ae883ab7d9f03480ab68fe68e9200c9b52c0c47abc21d2c90ec9.
//
// Solidity: event ReceiverUpdated(address previous, address updated)
func (_Binding *BindingFilterer) FilterReceiverUpdated(opts *bind.FilterOpts) (*BindingReceiverUpdatedIterator, error) {

	logs, sub, err := _Binding.contract.FilterLogs(opts, "ReceiverUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingReceiverUpdatedIterator{contract: _Binding.contract, event: "ReceiverUpdated", logs: logs, sub: sub}, nil
}

// WatchReceiverUpdated is a free log subscription operation binding the contract event 0xbda2bcccbfa5ae883ab7d9f03480ab68fe68e9200c9b52c0c47abc21d2c90ec9.
//
// Solidity: event ReceiverUpdated(address previous, address updated)
func (_Binding *BindingFilterer) WatchReceiverUpdated(opts *bind.WatchOpts, sink chan<- *BindingReceiverUpdated) (event.Subscription, error) {

	logs, sub, err := _Binding.contract.WatchLogs(opts, "ReceiverUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingReceiverUpdated)
				if err := _Binding.contract.UnpackLog(event, "ReceiverUpdated", log); err != nil {
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

// ParseReceiverUpdated is a log parse operation binding the contract event 0xbda2bcccbfa5ae883ab7d9f03480ab68fe68e9200c9b52c0c47abc21d2c90ec9.
//
// Solidity: event ReceiverUpdated(address previous, address updated)
func (_Binding *BindingFilterer) ParseReceiverUpdated(log types.Log) (*BindingReceiverUpdated, error) {
	event := new(BindingReceiverUpdated)
	if err := _Binding.contract.UnpackLog(event, "ReceiverUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingRoyaltyUpdatedIterator is returned from FilterRoyaltyUpdated and is used to iterate over the raw logs and unpacked data for RoyaltyUpdated events raised by the Binding contract.
type BindingRoyaltyUpdatedIterator struct {
	Event *BindingRoyaltyUpdated // Event containing the contract specifics and raw log

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
func (it *BindingRoyaltyUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingRoyaltyUpdated)
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
		it.Event = new(BindingRoyaltyUpdated)
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
func (it *BindingRoyaltyUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingRoyaltyUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingRoyaltyUpdated represents a RoyaltyUpdated event raised by the Binding contract.
type BindingRoyaltyUpdated struct {
	Previous *big.Int
	Updated  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRoyaltyUpdated is a free log retrieval operation binding the contract event 0x54e506cda8889617ec187c699f1c3b373053eb5796248194796f7e1501dfab24.
//
// Solidity: event RoyaltyUpdated(uint256 previous, uint256 updated)
func (_Binding *BindingFilterer) FilterRoyaltyUpdated(opts *bind.FilterOpts) (*BindingRoyaltyUpdatedIterator, error) {

	logs, sub, err := _Binding.contract.FilterLogs(opts, "RoyaltyUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingRoyaltyUpdatedIterator{contract: _Binding.contract, event: "RoyaltyUpdated", logs: logs, sub: sub}, nil
}

// WatchRoyaltyUpdated is a free log subscription operation binding the contract event 0x54e506cda8889617ec187c699f1c3b373053eb5796248194796f7e1501dfab24.
//
// Solidity: event RoyaltyUpdated(uint256 previous, uint256 updated)
func (_Binding *BindingFilterer) WatchRoyaltyUpdated(opts *bind.WatchOpts, sink chan<- *BindingRoyaltyUpdated) (event.Subscription, error) {

	logs, sub, err := _Binding.contract.WatchLogs(opts, "RoyaltyUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingRoyaltyUpdated)
				if err := _Binding.contract.UnpackLog(event, "RoyaltyUpdated", log); err != nil {
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

// ParseRoyaltyUpdated is a log parse operation binding the contract event 0x54e506cda8889617ec187c699f1c3b373053eb5796248194796f7e1501dfab24.
//
// Solidity: event RoyaltyUpdated(uint256 previous, uint256 updated)
func (_Binding *BindingFilterer) ParseRoyaltyUpdated(log types.Log) (*BindingRoyaltyUpdated, error) {
	event := new(BindingRoyaltyUpdated)
	if err := _Binding.contract.UnpackLog(event, "RoyaltyUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the Binding contract.
type BindingTransferIterator struct {
	Event *BindingTransfer // Event containing the contract specifics and raw log

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
func (it *BindingTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingTransfer)
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
		it.Event = new(BindingTransfer)
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
func (it *BindingTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingTransfer represents a Transfer event raised by the Binding contract.
type BindingTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Binding *BindingFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*BindingTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _Binding.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &BindingTransferIterator{contract: _Binding.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Binding *BindingFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *BindingTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _Binding.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingTransfer)
				if err := _Binding.contract.UnpackLog(event, "Transfer", log); err != nil {
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
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Binding *BindingFilterer) ParseTransfer(log types.Log) (*BindingTransfer, error) {
	event := new(BindingTransfer)
	if err := _Binding.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingUriFallbackUpdatedIterator is returned from FilterUriFallbackUpdated and is used to iterate over the raw logs and unpacked data for UriFallbackUpdated events raised by the Binding contract.
type BindingUriFallbackUpdatedIterator struct {
	Event *BindingUriFallbackUpdated // Event containing the contract specifics and raw log

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
func (it *BindingUriFallbackUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingUriFallbackUpdated)
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
		it.Event = new(BindingUriFallbackUpdated)
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
func (it *BindingUriFallbackUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingUriFallbackUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingUriFallbackUpdated represents a UriFallbackUpdated event raised by the Binding contract.
type BindingUriFallbackUpdated struct {
	Previous string
	Updated  string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterUriFallbackUpdated is a free log retrieval operation binding the contract event 0xe1b7ff5efe58018e39b7877b5cfa772bb90f32504be7b2330b078d2a9b114bbe.
//
// Solidity: event UriFallbackUpdated(string previous, string updated)
func (_Binding *BindingFilterer) FilterUriFallbackUpdated(opts *bind.FilterOpts) (*BindingUriFallbackUpdatedIterator, error) {

	logs, sub, err := _Binding.contract.FilterLogs(opts, "UriFallbackUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingUriFallbackUpdatedIterator{contract: _Binding.contract, event: "UriFallbackUpdated", logs: logs, sub: sub}, nil
}

// WatchUriFallbackUpdated is a free log subscription operation binding the contract event 0xe1b7ff5efe58018e39b7877b5cfa772bb90f32504be7b2330b078d2a9b114bbe.
//
// Solidity: event UriFallbackUpdated(string previous, string updated)
func (_Binding *BindingFilterer) WatchUriFallbackUpdated(opts *bind.WatchOpts, sink chan<- *BindingUriFallbackUpdated) (event.Subscription, error) {

	logs, sub, err := _Binding.contract.WatchLogs(opts, "UriFallbackUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingUriFallbackUpdated)
				if err := _Binding.contract.UnpackLog(event, "UriFallbackUpdated", log); err != nil {
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

// ParseUriFallbackUpdated is a log parse operation binding the contract event 0xe1b7ff5efe58018e39b7877b5cfa772bb90f32504be7b2330b078d2a9b114bbe.
//
// Solidity: event UriFallbackUpdated(string previous, string updated)
func (_Binding *BindingFilterer) ParseUriFallbackUpdated(log types.Log) (*BindingUriFallbackUpdated, error) {
	event := new(BindingUriFallbackUpdated)
	if err := _Binding.contract.UnpackLog(event, "UriFallbackUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingUriUpdatedIterator is returned from FilterUriUpdated and is used to iterate over the raw logs and unpacked data for UriUpdated events raised by the Binding contract.
type BindingUriUpdatedIterator struct {
	Event *BindingUriUpdated // Event containing the contract specifics and raw log

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
func (it *BindingUriUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingUriUpdated)
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
		it.Event = new(BindingUriUpdated)
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
func (it *BindingUriUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingUriUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingUriUpdated represents a UriUpdated event raised by the Binding contract.
type BindingUriUpdated struct {
	Previous string
	Updated  string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterUriUpdated is a free log retrieval operation binding the contract event 0x7d8ebb5abe647a67ba3a2649e11557ae5aa256cf3449245e0c840c98132e5a37.
//
// Solidity: event UriUpdated(string previous, string updated)
func (_Binding *BindingFilterer) FilterUriUpdated(opts *bind.FilterOpts) (*BindingUriUpdatedIterator, error) {

	logs, sub, err := _Binding.contract.FilterLogs(opts, "UriUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingUriUpdatedIterator{contract: _Binding.contract, event: "UriUpdated", logs: logs, sub: sub}, nil
}

// WatchUriUpdated is a free log subscription operation binding the contract event 0x7d8ebb5abe647a67ba3a2649e11557ae5aa256cf3449245e0c840c98132e5a37.
//
// Solidity: event UriUpdated(string previous, string updated)
func (_Binding *BindingFilterer) WatchUriUpdated(opts *bind.WatchOpts, sink chan<- *BindingUriUpdated) (event.Subscription, error) {

	logs, sub, err := _Binding.contract.WatchLogs(opts, "UriUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingUriUpdated)
				if err := _Binding.contract.UnpackLog(event, "UriUpdated", log); err != nil {
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

// ParseUriUpdated is a log parse operation binding the contract event 0x7d8ebb5abe647a67ba3a2649e11557ae5aa256cf3449245e0c840c98132e5a37.
//
// Solidity: event UriUpdated(string previous, string updated)
func (_Binding *BindingFilterer) ParseUriUpdated(log types.Log) (*BindingUriUpdated, error) {
	event := new(BindingUriUpdated)
	if err := _Binding.contract.UnpackLog(event, "UriUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
