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

// BindingsMetaData contains all meta data concerning the Bindings contract.
var BindingsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"uri_\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"vrfCoordinator_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"vrfLink_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"vrfKeyHash_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"vrfFee_\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"previous\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"updated\",\"type\":\"uint256\"}],\"name\":\"BatchUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"previous\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"updated\",\"type\":\"string\"}],\"name\":\"HiddenUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"previous\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"updated\",\"type\":\"uint256\"}],\"name\":\"MaxUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"previous\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"updated\",\"type\":\"bool\"}],\"name\":\"MintableUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previous\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"updated\",\"type\":\"address\"}],\"name\":\"PackUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"previous\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"updated\",\"type\":\"uint256\"}],\"name\":\"PriceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previous\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"updated\",\"type\":\"address\"}],\"name\":\"ReceiverUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"previous\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"updated\",\"type\":\"bool\"}],\"name\":\"RevealableUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"previous\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"updated\",\"type\":\"uint256\"}],\"name\":\"RoyaltyUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"requestId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"tokens\",\"type\":\"uint256[]\"}],\"name\":\"TokenReveal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"previous\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"updated\",\"type\":\"string\"}],\"name\":\"UriFallbackUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"previous\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"updated\",\"type\":\"string\"}],\"name\":\"UriUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"allocation\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hidden\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"max\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mintOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"mintable\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"order\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pack\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pending\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"price\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"randomness\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"requestId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"randomness\",\"type\":\"uint256\"}],\"name\":\"rawFulfillRandomness\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"receiver\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"tokens\",\"type\":\"uint256[]\"}],\"name\":\"reveal\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"revealable\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"revealed\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"royalty\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"salePrice\",\"type\":\"uint256\"}],\"name\":\"royaltyInfo\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"allocation_\",\"type\":\"string\"}],\"name\":\"setAllocation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"batch_\",\"type\":\"uint256\"}],\"name\":\"setBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"hidden_\",\"type\":\"string\"}],\"name\":\"setHidden\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"max_\",\"type\":\"uint256\"}],\"name\":\"setMax\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"status_\",\"type\":\"bool\"}],\"name\":\"setMintable\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pack_\",\"type\":\"address\"}],\"name\":\"setPack\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"price_\",\"type\":\"uint256\"}],\"name\":\"setPrice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"receiver_\",\"type\":\"address\"}],\"name\":\"setReceiver\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"status_\",\"type\":\"bool\"}],\"name\":\"setRevealable\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"royalty_\",\"type\":\"uint256\"}],\"name\":\"setRoyalty\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"uri_\",\"type\":\"string\"}],\"name\":\"setURI\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"uri_\",\"type\":\"string\"}],\"name\":\"setURIFallback\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"shuffle\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURIFallback\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"total\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"update\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"uri\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"uriFallback\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// BindingsABI is the input ABI used to generate the binding from.
// Deprecated: Use BindingsMetaData.ABI instead.
var BindingsABI = BindingsMetaData.ABI

// Bindings is an auto generated Go binding around an Ethereum contract.
type Bindings struct {
	BindingsCaller     // Read-only binding to the contract
	BindingsTransactor // Write-only binding to the contract
	BindingsFilterer   // Log filterer for contract events
}

// BindingsCaller is an auto generated read-only Go binding around an Ethereum contract.
type BindingsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BindingsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BindingsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BindingsSession struct {
	Contract     *Bindings         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BindingsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BindingsCallerSession struct {
	Contract *BindingsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// BindingsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BindingsTransactorSession struct {
	Contract     *BindingsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// BindingsRaw is an auto generated low-level Go binding around an Ethereum contract.
type BindingsRaw struct {
	Contract *Bindings // Generic contract binding to access the raw methods on
}

// BindingsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BindingsCallerRaw struct {
	Contract *BindingsCaller // Generic read-only contract binding to access the raw methods on
}

// BindingsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BindingsTransactorRaw struct {
	Contract *BindingsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBindings creates a new instance of Bindings, bound to a specific deployed contract.
func NewBindings(address common.Address, backend bind.ContractBackend) (*Bindings, error) {
	contract, err := bindBindings(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bindings{BindingsCaller: BindingsCaller{contract: contract}, BindingsTransactor: BindingsTransactor{contract: contract}, BindingsFilterer: BindingsFilterer{contract: contract}}, nil
}

// NewBindingsCaller creates a new read-only instance of Bindings, bound to a specific deployed contract.
func NewBindingsCaller(address common.Address, caller bind.ContractCaller) (*BindingsCaller, error) {
	contract, err := bindBindings(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BindingsCaller{contract: contract}, nil
}

// NewBindingsTransactor creates a new write-only instance of Bindings, bound to a specific deployed contract.
func NewBindingsTransactor(address common.Address, transactor bind.ContractTransactor) (*BindingsTransactor, error) {
	contract, err := bindBindings(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BindingsTransactor{contract: contract}, nil
}

// NewBindingsFilterer creates a new log filterer instance of Bindings, bound to a specific deployed contract.
func NewBindingsFilterer(address common.Address, filterer bind.ContractFilterer) (*BindingsFilterer, error) {
	contract, err := bindBindings(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BindingsFilterer{contract: contract}, nil
}

// bindBindings binds a generic wrapper to an already deployed contract.
func bindBindings(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BindingsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bindings *BindingsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bindings.Contract.BindingsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bindings *BindingsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bindings.Contract.BindingsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bindings *BindingsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bindings.Contract.BindingsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bindings *BindingsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bindings.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bindings *BindingsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bindings.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bindings *BindingsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bindings.Contract.contract.Transact(opts, method, params...)
}

// Allocation is a free data retrieval call binding the contract method 0x88a17bde.
//
// Solidity: function allocation() view returns(string)
func (_Bindings *BindingsCaller) Allocation(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "allocation")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Allocation is a free data retrieval call binding the contract method 0x88a17bde.
//
// Solidity: function allocation() view returns(string)
func (_Bindings *BindingsSession) Allocation() (string, error) {
	return _Bindings.Contract.Allocation(&_Bindings.CallOpts)
}

// Allocation is a free data retrieval call binding the contract method 0x88a17bde.
//
// Solidity: function allocation() view returns(string)
func (_Bindings *BindingsCallerSession) Allocation() (string, error) {
	return _Bindings.Contract.Allocation(&_Bindings.CallOpts)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_Bindings *BindingsCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_Bindings *BindingsSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _Bindings.Contract.BalanceOf(&_Bindings.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_Bindings *BindingsCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _Bindings.Contract.BalanceOf(&_Bindings.CallOpts, owner)
}

// Batch is a free data retrieval call binding the contract method 0xaf713566.
//
// Solidity: function batch() view returns(uint256)
func (_Bindings *BindingsCaller) Batch(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "batch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Batch is a free data retrieval call binding the contract method 0xaf713566.
//
// Solidity: function batch() view returns(uint256)
func (_Bindings *BindingsSession) Batch() (*big.Int, error) {
	return _Bindings.Contract.Batch(&_Bindings.CallOpts)
}

// Batch is a free data retrieval call binding the contract method 0xaf713566.
//
// Solidity: function batch() view returns(uint256)
func (_Bindings *BindingsCallerSession) Batch() (*big.Int, error) {
	return _Bindings.Contract.Batch(&_Bindings.CallOpts)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_Bindings *BindingsCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_Bindings *BindingsSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _Bindings.Contract.GetApproved(&_Bindings.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_Bindings *BindingsCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _Bindings.Contract.GetApproved(&_Bindings.CallOpts, tokenId)
}

// Hidden is a free data retrieval call binding the contract method 0xaef6d4b1.
//
// Solidity: function hidden() view returns(string)
func (_Bindings *BindingsCaller) Hidden(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "hidden")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Hidden is a free data retrieval call binding the contract method 0xaef6d4b1.
//
// Solidity: function hidden() view returns(string)
func (_Bindings *BindingsSession) Hidden() (string, error) {
	return _Bindings.Contract.Hidden(&_Bindings.CallOpts)
}

// Hidden is a free data retrieval call binding the contract method 0xaef6d4b1.
//
// Solidity: function hidden() view returns(string)
func (_Bindings *BindingsCallerSession) Hidden() (string, error) {
	return _Bindings.Contract.Hidden(&_Bindings.CallOpts)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_Bindings *BindingsCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_Bindings *BindingsSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _Bindings.Contract.IsApprovedForAll(&_Bindings.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_Bindings *BindingsCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _Bindings.Contract.IsApprovedForAll(&_Bindings.CallOpts, owner, operator)
}

// Max is a free data retrieval call binding the contract method 0x6ac5db19.
//
// Solidity: function max() view returns(uint256)
func (_Bindings *BindingsCaller) Max(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "max")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Max is a free data retrieval call binding the contract method 0x6ac5db19.
//
// Solidity: function max() view returns(uint256)
func (_Bindings *BindingsSession) Max() (*big.Int, error) {
	return _Bindings.Contract.Max(&_Bindings.CallOpts)
}

// Max is a free data retrieval call binding the contract method 0x6ac5db19.
//
// Solidity: function max() view returns(uint256)
func (_Bindings *BindingsCallerSession) Max() (*big.Int, error) {
	return _Bindings.Contract.Max(&_Bindings.CallOpts)
}

// Mintable is a free data retrieval call binding the contract method 0x4bf365df.
//
// Solidity: function mintable() view returns(bool)
func (_Bindings *BindingsCaller) Mintable(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "mintable")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Mintable is a free data retrieval call binding the contract method 0x4bf365df.
//
// Solidity: function mintable() view returns(bool)
func (_Bindings *BindingsSession) Mintable() (bool, error) {
	return _Bindings.Contract.Mintable(&_Bindings.CallOpts)
}

// Mintable is a free data retrieval call binding the contract method 0x4bf365df.
//
// Solidity: function mintable() view returns(bool)
func (_Bindings *BindingsCallerSession) Mintable() (bool, error) {
	return _Bindings.Contract.Mintable(&_Bindings.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Bindings *BindingsCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Bindings *BindingsSession) Name() (string, error) {
	return _Bindings.Contract.Name(&_Bindings.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Bindings *BindingsCallerSession) Name() (string, error) {
	return _Bindings.Contract.Name(&_Bindings.CallOpts)
}

// Order is a free data retrieval call binding the contract method 0x21603f43.
//
// Solidity: function order(uint256 ) view returns(uint256)
func (_Bindings *BindingsCaller) Order(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "order", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Order is a free data retrieval call binding the contract method 0x21603f43.
//
// Solidity: function order(uint256 ) view returns(uint256)
func (_Bindings *BindingsSession) Order(arg0 *big.Int) (*big.Int, error) {
	return _Bindings.Contract.Order(&_Bindings.CallOpts, arg0)
}

// Order is a free data retrieval call binding the contract method 0x21603f43.
//
// Solidity: function order(uint256 ) view returns(uint256)
func (_Bindings *BindingsCallerSession) Order(arg0 *big.Int) (*big.Int, error) {
	return _Bindings.Contract.Order(&_Bindings.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bindings *BindingsCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bindings *BindingsSession) Owner() (common.Address, error) {
	return _Bindings.Contract.Owner(&_Bindings.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bindings *BindingsCallerSession) Owner() (common.Address, error) {
	return _Bindings.Contract.Owner(&_Bindings.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Bindings *BindingsCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Bindings *BindingsSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Bindings.Contract.OwnerOf(&_Bindings.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Bindings *BindingsCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Bindings.Contract.OwnerOf(&_Bindings.CallOpts, tokenId)
}

// Pack is a free data retrieval call binding the contract method 0xef082838.
//
// Solidity: function pack() view returns(address)
func (_Bindings *BindingsCaller) Pack(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "pack")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Pack is a free data retrieval call binding the contract method 0xef082838.
//
// Solidity: function pack() view returns(address)
func (_Bindings *BindingsSession) Pack() (common.Address, error) {
	return _Bindings.Contract.Pack(&_Bindings.CallOpts)
}

// Pack is a free data retrieval call binding the contract method 0xef082838.
//
// Solidity: function pack() view returns(address)
func (_Bindings *BindingsCallerSession) Pack() (common.Address, error) {
	return _Bindings.Contract.Pack(&_Bindings.CallOpts)
}

// Pending is a free data retrieval call binding the contract method 0xe20ccec3.
//
// Solidity: function pending() view returns(uint256)
func (_Bindings *BindingsCaller) Pending(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "pending")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Pending is a free data retrieval call binding the contract method 0xe20ccec3.
//
// Solidity: function pending() view returns(uint256)
func (_Bindings *BindingsSession) Pending() (*big.Int, error) {
	return _Bindings.Contract.Pending(&_Bindings.CallOpts)
}

// Pending is a free data retrieval call binding the contract method 0xe20ccec3.
//
// Solidity: function pending() view returns(uint256)
func (_Bindings *BindingsCallerSession) Pending() (*big.Int, error) {
	return _Bindings.Contract.Pending(&_Bindings.CallOpts)
}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() view returns(uint256)
func (_Bindings *BindingsCaller) Price(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "price")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() view returns(uint256)
func (_Bindings *BindingsSession) Price() (*big.Int, error) {
	return _Bindings.Contract.Price(&_Bindings.CallOpts)
}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() view returns(uint256)
func (_Bindings *BindingsCallerSession) Price() (*big.Int, error) {
	return _Bindings.Contract.Price(&_Bindings.CallOpts)
}

// Randomness is a free data retrieval call binding the contract method 0x36013189.
//
// Solidity: function randomness() view returns(uint256)
func (_Bindings *BindingsCaller) Randomness(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "randomness")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Randomness is a free data retrieval call binding the contract method 0x36013189.
//
// Solidity: function randomness() view returns(uint256)
func (_Bindings *BindingsSession) Randomness() (*big.Int, error) {
	return _Bindings.Contract.Randomness(&_Bindings.CallOpts)
}

// Randomness is a free data retrieval call binding the contract method 0x36013189.
//
// Solidity: function randomness() view returns(uint256)
func (_Bindings *BindingsCallerSession) Randomness() (*big.Int, error) {
	return _Bindings.Contract.Randomness(&_Bindings.CallOpts)
}

// Receiver is a free data retrieval call binding the contract method 0xf7260d3e.
//
// Solidity: function receiver() view returns(address)
func (_Bindings *BindingsCaller) Receiver(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "receiver")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Receiver is a free data retrieval call binding the contract method 0xf7260d3e.
//
// Solidity: function receiver() view returns(address)
func (_Bindings *BindingsSession) Receiver() (common.Address, error) {
	return _Bindings.Contract.Receiver(&_Bindings.CallOpts)
}

// Receiver is a free data retrieval call binding the contract method 0xf7260d3e.
//
// Solidity: function receiver() view returns(address)
func (_Bindings *BindingsCallerSession) Receiver() (common.Address, error) {
	return _Bindings.Contract.Receiver(&_Bindings.CallOpts)
}

// Revealable is a free data retrieval call binding the contract method 0x03d16985.
//
// Solidity: function revealable() view returns(bool)
func (_Bindings *BindingsCaller) Revealable(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "revealable")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Revealable is a free data retrieval call binding the contract method 0x03d16985.
//
// Solidity: function revealable() view returns(bool)
func (_Bindings *BindingsSession) Revealable() (bool, error) {
	return _Bindings.Contract.Revealable(&_Bindings.CallOpts)
}

// Revealable is a free data retrieval call binding the contract method 0x03d16985.
//
// Solidity: function revealable() view returns(bool)
func (_Bindings *BindingsCallerSession) Revealable() (bool, error) {
	return _Bindings.Contract.Revealable(&_Bindings.CallOpts)
}

// Revealed is a free data retrieval call binding the contract method 0x51830227.
//
// Solidity: function revealed() view returns(uint256)
func (_Bindings *BindingsCaller) Revealed(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "revealed")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Revealed is a free data retrieval call binding the contract method 0x51830227.
//
// Solidity: function revealed() view returns(uint256)
func (_Bindings *BindingsSession) Revealed() (*big.Int, error) {
	return _Bindings.Contract.Revealed(&_Bindings.CallOpts)
}

// Revealed is a free data retrieval call binding the contract method 0x51830227.
//
// Solidity: function revealed() view returns(uint256)
func (_Bindings *BindingsCallerSession) Revealed() (*big.Int, error) {
	return _Bindings.Contract.Revealed(&_Bindings.CallOpts)
}

// Royalty is a free data retrieval call binding the contract method 0x29ee566c.
//
// Solidity: function royalty() view returns(uint256)
func (_Bindings *BindingsCaller) Royalty(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "royalty")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Royalty is a free data retrieval call binding the contract method 0x29ee566c.
//
// Solidity: function royalty() view returns(uint256)
func (_Bindings *BindingsSession) Royalty() (*big.Int, error) {
	return _Bindings.Contract.Royalty(&_Bindings.CallOpts)
}

// Royalty is a free data retrieval call binding the contract method 0x29ee566c.
//
// Solidity: function royalty() view returns(uint256)
func (_Bindings *BindingsCallerSession) Royalty() (*big.Int, error) {
	return _Bindings.Contract.Royalty(&_Bindings.CallOpts)
}

// RoyaltyInfo is a free data retrieval call binding the contract method 0x2a55205a.
//
// Solidity: function royaltyInfo(uint256 tokenId, uint256 salePrice) view returns(address, uint256)
func (_Bindings *BindingsCaller) RoyaltyInfo(opts *bind.CallOpts, tokenId *big.Int, salePrice *big.Int) (common.Address, *big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "royaltyInfo", tokenId, salePrice)

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
func (_Bindings *BindingsSession) RoyaltyInfo(tokenId *big.Int, salePrice *big.Int) (common.Address, *big.Int, error) {
	return _Bindings.Contract.RoyaltyInfo(&_Bindings.CallOpts, tokenId, salePrice)
}

// RoyaltyInfo is a free data retrieval call binding the contract method 0x2a55205a.
//
// Solidity: function royaltyInfo(uint256 tokenId, uint256 salePrice) view returns(address, uint256)
func (_Bindings *BindingsCallerSession) RoyaltyInfo(tokenId *big.Int, salePrice *big.Int) (common.Address, *big.Int, error) {
	return _Bindings.Contract.RoyaltyInfo(&_Bindings.CallOpts, tokenId, salePrice)
}

// Shuffle is a free data retrieval call binding the contract method 0xef6537b5.
//
// Solidity: function shuffle(uint256 ) view returns(uint256)
func (_Bindings *BindingsCaller) Shuffle(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "shuffle", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Shuffle is a free data retrieval call binding the contract method 0xef6537b5.
//
// Solidity: function shuffle(uint256 ) view returns(uint256)
func (_Bindings *BindingsSession) Shuffle(arg0 *big.Int) (*big.Int, error) {
	return _Bindings.Contract.Shuffle(&_Bindings.CallOpts, arg0)
}

// Shuffle is a free data retrieval call binding the contract method 0xef6537b5.
//
// Solidity: function shuffle(uint256 ) view returns(uint256)
func (_Bindings *BindingsCallerSession) Shuffle(arg0 *big.Int) (*big.Int, error) {
	return _Bindings.Contract.Shuffle(&_Bindings.CallOpts, arg0)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Bindings *BindingsCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Bindings *BindingsSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Bindings.Contract.SupportsInterface(&_Bindings.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Bindings *BindingsCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Bindings.Contract.SupportsInterface(&_Bindings.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Bindings *BindingsCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Bindings *BindingsSession) Symbol() (string, error) {
	return _Bindings.Contract.Symbol(&_Bindings.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Bindings *BindingsCallerSession) Symbol() (string, error) {
	return _Bindings.Contract.Symbol(&_Bindings.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Bindings *BindingsCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Bindings *BindingsSession) TokenURI(tokenId *big.Int) (string, error) {
	return _Bindings.Contract.TokenURI(&_Bindings.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Bindings *BindingsCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _Bindings.Contract.TokenURI(&_Bindings.CallOpts, tokenId)
}

// TokenURIFallback is a free data retrieval call binding the contract method 0xc7c8f564.
//
// Solidity: function tokenURIFallback(uint256 tokenId) view returns(string)
func (_Bindings *BindingsCaller) TokenURIFallback(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "tokenURIFallback", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURIFallback is a free data retrieval call binding the contract method 0xc7c8f564.
//
// Solidity: function tokenURIFallback(uint256 tokenId) view returns(string)
func (_Bindings *BindingsSession) TokenURIFallback(tokenId *big.Int) (string, error) {
	return _Bindings.Contract.TokenURIFallback(&_Bindings.CallOpts, tokenId)
}

// TokenURIFallback is a free data retrieval call binding the contract method 0xc7c8f564.
//
// Solidity: function tokenURIFallback(uint256 tokenId) view returns(string)
func (_Bindings *BindingsCallerSession) TokenURIFallback(tokenId *big.Int) (string, error) {
	return _Bindings.Contract.TokenURIFallback(&_Bindings.CallOpts, tokenId)
}

// Total is a free data retrieval call binding the contract method 0x2ddbd13a.
//
// Solidity: function total() view returns(uint256)
func (_Bindings *BindingsCaller) Total(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "total")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Total is a free data retrieval call binding the contract method 0x2ddbd13a.
//
// Solidity: function total() view returns(uint256)
func (_Bindings *BindingsSession) Total() (*big.Int, error) {
	return _Bindings.Contract.Total(&_Bindings.CallOpts)
}

// Total is a free data retrieval call binding the contract method 0x2ddbd13a.
//
// Solidity: function total() view returns(uint256)
func (_Bindings *BindingsCallerSession) Total() (*big.Int, error) {
	return _Bindings.Contract.Total(&_Bindings.CallOpts)
}

// Uri is a free data retrieval call binding the contract method 0xeac989f8.
//
// Solidity: function uri() view returns(string)
func (_Bindings *BindingsCaller) Uri(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "uri")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Uri is a free data retrieval call binding the contract method 0xeac989f8.
//
// Solidity: function uri() view returns(string)
func (_Bindings *BindingsSession) Uri() (string, error) {
	return _Bindings.Contract.Uri(&_Bindings.CallOpts)
}

// Uri is a free data retrieval call binding the contract method 0xeac989f8.
//
// Solidity: function uri() view returns(string)
func (_Bindings *BindingsCallerSession) Uri() (string, error) {
	return _Bindings.Contract.Uri(&_Bindings.CallOpts)
}

// UriFallback is a free data retrieval call binding the contract method 0x6dd8e21a.
//
// Solidity: function uriFallback() view returns(string)
func (_Bindings *BindingsCaller) UriFallback(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Bindings.contract.Call(opts, &out, "uriFallback")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UriFallback is a free data retrieval call binding the contract method 0x6dd8e21a.
//
// Solidity: function uriFallback() view returns(string)
func (_Bindings *BindingsSession) UriFallback() (string, error) {
	return _Bindings.Contract.UriFallback(&_Bindings.CallOpts)
}

// UriFallback is a free data retrieval call binding the contract method 0x6dd8e21a.
//
// Solidity: function uriFallback() view returns(string)
func (_Bindings *BindingsCallerSession) UriFallback() (string, error) {
	return _Bindings.Contract.UriFallback(&_Bindings.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_Bindings *BindingsTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_Bindings *BindingsSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.Approve(&_Bindings.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_Bindings *BindingsTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.Approve(&_Bindings.TransactOpts, to, tokenId)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount) payable returns()
func (_Bindings *BindingsTransactor) Mint(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "mint", amount)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount) payable returns()
func (_Bindings *BindingsSession) Mint(amount *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.Mint(&_Bindings.TransactOpts, amount)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount) payable returns()
func (_Bindings *BindingsTransactorSession) Mint(amount *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.Mint(&_Bindings.TransactOpts, amount)
}

// MintOwner is a paid mutator transaction binding the contract method 0x33f88d22.
//
// Solidity: function mintOwner(uint256 amount) returns()
func (_Bindings *BindingsTransactor) MintOwner(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "mintOwner", amount)
}

// MintOwner is a paid mutator transaction binding the contract method 0x33f88d22.
//
// Solidity: function mintOwner(uint256 amount) returns()
func (_Bindings *BindingsSession) MintOwner(amount *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.MintOwner(&_Bindings.TransactOpts, amount)
}

// MintOwner is a paid mutator transaction binding the contract method 0x33f88d22.
//
// Solidity: function mintOwner(uint256 amount) returns()
func (_Bindings *BindingsTransactorSession) MintOwner(amount *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.MintOwner(&_Bindings.TransactOpts, amount)
}

// RawFulfillRandomness is a paid mutator transaction binding the contract method 0x94985ddd.
//
// Solidity: function rawFulfillRandomness(bytes32 requestId, uint256 randomness) returns()
func (_Bindings *BindingsTransactor) RawFulfillRandomness(opts *bind.TransactOpts, requestId [32]byte, randomness *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "rawFulfillRandomness", requestId, randomness)
}

// RawFulfillRandomness is a paid mutator transaction binding the contract method 0x94985ddd.
//
// Solidity: function rawFulfillRandomness(bytes32 requestId, uint256 randomness) returns()
func (_Bindings *BindingsSession) RawFulfillRandomness(requestId [32]byte, randomness *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.RawFulfillRandomness(&_Bindings.TransactOpts, requestId, randomness)
}

// RawFulfillRandomness is a paid mutator transaction binding the contract method 0x94985ddd.
//
// Solidity: function rawFulfillRandomness(bytes32 requestId, uint256 randomness) returns()
func (_Bindings *BindingsTransactorSession) RawFulfillRandomness(requestId [32]byte, randomness *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.RawFulfillRandomness(&_Bindings.TransactOpts, requestId, randomness)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bindings *BindingsTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bindings *BindingsSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bindings.Contract.RenounceOwnership(&_Bindings.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bindings *BindingsTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bindings.Contract.RenounceOwnership(&_Bindings.TransactOpts)
}

// Reveal is a paid mutator transaction binding the contract method 0xb93f208a.
//
// Solidity: function reveal(uint256[] tokens) returns(bytes32)
func (_Bindings *BindingsTransactor) Reveal(opts *bind.TransactOpts, tokens []*big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "reveal", tokens)
}

// Reveal is a paid mutator transaction binding the contract method 0xb93f208a.
//
// Solidity: function reveal(uint256[] tokens) returns(bytes32)
func (_Bindings *BindingsSession) Reveal(tokens []*big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.Reveal(&_Bindings.TransactOpts, tokens)
}

// Reveal is a paid mutator transaction binding the contract method 0xb93f208a.
//
// Solidity: function reveal(uint256[] tokens) returns(bytes32)
func (_Bindings *BindingsTransactorSession) Reveal(tokens []*big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.Reveal(&_Bindings.TransactOpts, tokens)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_Bindings *BindingsTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_Bindings *BindingsSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.SafeTransferFrom(&_Bindings.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_Bindings *BindingsTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.SafeTransferFrom(&_Bindings.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes _data) returns()
func (_Bindings *BindingsTransactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, _data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes _data) returns()
func (_Bindings *BindingsSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _Bindings.Contract.SafeTransferFrom0(&_Bindings.TransactOpts, from, to, tokenId, _data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes _data) returns()
func (_Bindings *BindingsTransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _Bindings.Contract.SafeTransferFrom0(&_Bindings.TransactOpts, from, to, tokenId, _data)
}

// SetAllocation is a paid mutator transaction binding the contract method 0x970a1fa8.
//
// Solidity: function setAllocation(string allocation_) returns()
func (_Bindings *BindingsTransactor) SetAllocation(opts *bind.TransactOpts, allocation_ string) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setAllocation", allocation_)
}

// SetAllocation is a paid mutator transaction binding the contract method 0x970a1fa8.
//
// Solidity: function setAllocation(string allocation_) returns()
func (_Bindings *BindingsSession) SetAllocation(allocation_ string) (*types.Transaction, error) {
	return _Bindings.Contract.SetAllocation(&_Bindings.TransactOpts, allocation_)
}

// SetAllocation is a paid mutator transaction binding the contract method 0x970a1fa8.
//
// Solidity: function setAllocation(string allocation_) returns()
func (_Bindings *BindingsTransactorSession) SetAllocation(allocation_ string) (*types.Transaction, error) {
	return _Bindings.Contract.SetAllocation(&_Bindings.TransactOpts, allocation_)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_Bindings *BindingsTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_Bindings *BindingsSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _Bindings.Contract.SetApprovalForAll(&_Bindings.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_Bindings *BindingsTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _Bindings.Contract.SetApprovalForAll(&_Bindings.TransactOpts, operator, approved)
}

// SetBatch is a paid mutator transaction binding the contract method 0xb76060f7.
//
// Solidity: function setBatch(uint256 batch_) returns()
func (_Bindings *BindingsTransactor) SetBatch(opts *bind.TransactOpts, batch_ *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setBatch", batch_)
}

// SetBatch is a paid mutator transaction binding the contract method 0xb76060f7.
//
// Solidity: function setBatch(uint256 batch_) returns()
func (_Bindings *BindingsSession) SetBatch(batch_ *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.SetBatch(&_Bindings.TransactOpts, batch_)
}

// SetBatch is a paid mutator transaction binding the contract method 0xb76060f7.
//
// Solidity: function setBatch(uint256 batch_) returns()
func (_Bindings *BindingsTransactorSession) SetBatch(batch_ *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.SetBatch(&_Bindings.TransactOpts, batch_)
}

// SetHidden is a paid mutator transaction binding the contract method 0xd309aa2c.
//
// Solidity: function setHidden(string hidden_) returns()
func (_Bindings *BindingsTransactor) SetHidden(opts *bind.TransactOpts, hidden_ string) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setHidden", hidden_)
}

// SetHidden is a paid mutator transaction binding the contract method 0xd309aa2c.
//
// Solidity: function setHidden(string hidden_) returns()
func (_Bindings *BindingsSession) SetHidden(hidden_ string) (*types.Transaction, error) {
	return _Bindings.Contract.SetHidden(&_Bindings.TransactOpts, hidden_)
}

// SetHidden is a paid mutator transaction binding the contract method 0xd309aa2c.
//
// Solidity: function setHidden(string hidden_) returns()
func (_Bindings *BindingsTransactorSession) SetHidden(hidden_ string) (*types.Transaction, error) {
	return _Bindings.Contract.SetHidden(&_Bindings.TransactOpts, hidden_)
}

// SetMax is a paid mutator transaction binding the contract method 0x1fe9eabc.
//
// Solidity: function setMax(uint256 max_) returns()
func (_Bindings *BindingsTransactor) SetMax(opts *bind.TransactOpts, max_ *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setMax", max_)
}

// SetMax is a paid mutator transaction binding the contract method 0x1fe9eabc.
//
// Solidity: function setMax(uint256 max_) returns()
func (_Bindings *BindingsSession) SetMax(max_ *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.SetMax(&_Bindings.TransactOpts, max_)
}

// SetMax is a paid mutator transaction binding the contract method 0x1fe9eabc.
//
// Solidity: function setMax(uint256 max_) returns()
func (_Bindings *BindingsTransactorSession) SetMax(max_ *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.SetMax(&_Bindings.TransactOpts, max_)
}

// SetMintable is a paid mutator transaction binding the contract method 0x285d70d4.
//
// Solidity: function setMintable(bool status_) returns()
func (_Bindings *BindingsTransactor) SetMintable(opts *bind.TransactOpts, status_ bool) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setMintable", status_)
}

// SetMintable is a paid mutator transaction binding the contract method 0x285d70d4.
//
// Solidity: function setMintable(bool status_) returns()
func (_Bindings *BindingsSession) SetMintable(status_ bool) (*types.Transaction, error) {
	return _Bindings.Contract.SetMintable(&_Bindings.TransactOpts, status_)
}

// SetMintable is a paid mutator transaction binding the contract method 0x285d70d4.
//
// Solidity: function setMintable(bool status_) returns()
func (_Bindings *BindingsTransactorSession) SetMintable(status_ bool) (*types.Transaction, error) {
	return _Bindings.Contract.SetMintable(&_Bindings.TransactOpts, status_)
}

// SetPack is a paid mutator transaction binding the contract method 0x353c65c5.
//
// Solidity: function setPack(address pack_) returns()
func (_Bindings *BindingsTransactor) SetPack(opts *bind.TransactOpts, pack_ common.Address) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setPack", pack_)
}

// SetPack is a paid mutator transaction binding the contract method 0x353c65c5.
//
// Solidity: function setPack(address pack_) returns()
func (_Bindings *BindingsSession) SetPack(pack_ common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.SetPack(&_Bindings.TransactOpts, pack_)
}

// SetPack is a paid mutator transaction binding the contract method 0x353c65c5.
//
// Solidity: function setPack(address pack_) returns()
func (_Bindings *BindingsTransactorSession) SetPack(pack_ common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.SetPack(&_Bindings.TransactOpts, pack_)
}

// SetPrice is a paid mutator transaction binding the contract method 0x91b7f5ed.
//
// Solidity: function setPrice(uint256 price_) returns()
func (_Bindings *BindingsTransactor) SetPrice(opts *bind.TransactOpts, price_ *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setPrice", price_)
}

// SetPrice is a paid mutator transaction binding the contract method 0x91b7f5ed.
//
// Solidity: function setPrice(uint256 price_) returns()
func (_Bindings *BindingsSession) SetPrice(price_ *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.SetPrice(&_Bindings.TransactOpts, price_)
}

// SetPrice is a paid mutator transaction binding the contract method 0x91b7f5ed.
//
// Solidity: function setPrice(uint256 price_) returns()
func (_Bindings *BindingsTransactorSession) SetPrice(price_ *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.SetPrice(&_Bindings.TransactOpts, price_)
}

// SetReceiver is a paid mutator transaction binding the contract method 0x718da7ee.
//
// Solidity: function setReceiver(address receiver_) returns()
func (_Bindings *BindingsTransactor) SetReceiver(opts *bind.TransactOpts, receiver_ common.Address) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setReceiver", receiver_)
}

// SetReceiver is a paid mutator transaction binding the contract method 0x718da7ee.
//
// Solidity: function setReceiver(address receiver_) returns()
func (_Bindings *BindingsSession) SetReceiver(receiver_ common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.SetReceiver(&_Bindings.TransactOpts, receiver_)
}

// SetReceiver is a paid mutator transaction binding the contract method 0x718da7ee.
//
// Solidity: function setReceiver(address receiver_) returns()
func (_Bindings *BindingsTransactorSession) SetReceiver(receiver_ common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.SetReceiver(&_Bindings.TransactOpts, receiver_)
}

// SetRevealable is a paid mutator transaction binding the contract method 0xb06194d3.
//
// Solidity: function setRevealable(bool status_) returns()
func (_Bindings *BindingsTransactor) SetRevealable(opts *bind.TransactOpts, status_ bool) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setRevealable", status_)
}

// SetRevealable is a paid mutator transaction binding the contract method 0xb06194d3.
//
// Solidity: function setRevealable(bool status_) returns()
func (_Bindings *BindingsSession) SetRevealable(status_ bool) (*types.Transaction, error) {
	return _Bindings.Contract.SetRevealable(&_Bindings.TransactOpts, status_)
}

// SetRevealable is a paid mutator transaction binding the contract method 0xb06194d3.
//
// Solidity: function setRevealable(bool status_) returns()
func (_Bindings *BindingsTransactorSession) SetRevealable(status_ bool) (*types.Transaction, error) {
	return _Bindings.Contract.SetRevealable(&_Bindings.TransactOpts, status_)
}

// SetRoyalty is a paid mutator transaction binding the contract method 0x4209a2e1.
//
// Solidity: function setRoyalty(uint256 royalty_) returns()
func (_Bindings *BindingsTransactor) SetRoyalty(opts *bind.TransactOpts, royalty_ *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setRoyalty", royalty_)
}

// SetRoyalty is a paid mutator transaction binding the contract method 0x4209a2e1.
//
// Solidity: function setRoyalty(uint256 royalty_) returns()
func (_Bindings *BindingsSession) SetRoyalty(royalty_ *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.SetRoyalty(&_Bindings.TransactOpts, royalty_)
}

// SetRoyalty is a paid mutator transaction binding the contract method 0x4209a2e1.
//
// Solidity: function setRoyalty(uint256 royalty_) returns()
func (_Bindings *BindingsTransactorSession) SetRoyalty(royalty_ *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.SetRoyalty(&_Bindings.TransactOpts, royalty_)
}

// SetURI is a paid mutator transaction binding the contract method 0x02fe5305.
//
// Solidity: function setURI(string uri_) returns()
func (_Bindings *BindingsTransactor) SetURI(opts *bind.TransactOpts, uri_ string) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setURI", uri_)
}

// SetURI is a paid mutator transaction binding the contract method 0x02fe5305.
//
// Solidity: function setURI(string uri_) returns()
func (_Bindings *BindingsSession) SetURI(uri_ string) (*types.Transaction, error) {
	return _Bindings.Contract.SetURI(&_Bindings.TransactOpts, uri_)
}

// SetURI is a paid mutator transaction binding the contract method 0x02fe5305.
//
// Solidity: function setURI(string uri_) returns()
func (_Bindings *BindingsTransactorSession) SetURI(uri_ string) (*types.Transaction, error) {
	return _Bindings.Contract.SetURI(&_Bindings.TransactOpts, uri_)
}

// SetURIFallback is a paid mutator transaction binding the contract method 0x0c8ab6e1.
//
// Solidity: function setURIFallback(string uri_) returns()
func (_Bindings *BindingsTransactor) SetURIFallback(opts *bind.TransactOpts, uri_ string) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "setURIFallback", uri_)
}

// SetURIFallback is a paid mutator transaction binding the contract method 0x0c8ab6e1.
//
// Solidity: function setURIFallback(string uri_) returns()
func (_Bindings *BindingsSession) SetURIFallback(uri_ string) (*types.Transaction, error) {
	return _Bindings.Contract.SetURIFallback(&_Bindings.TransactOpts, uri_)
}

// SetURIFallback is a paid mutator transaction binding the contract method 0x0c8ab6e1.
//
// Solidity: function setURIFallback(string uri_) returns()
func (_Bindings *BindingsTransactorSession) SetURIFallback(uri_ string) (*types.Transaction, error) {
	return _Bindings.Contract.SetURIFallback(&_Bindings.TransactOpts, uri_)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_Bindings *BindingsTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_Bindings *BindingsSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.TransferFrom(&_Bindings.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_Bindings *BindingsTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Bindings.Contract.TransferFrom(&_Bindings.TransactOpts, from, to, tokenId)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bindings *BindingsTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bindings *BindingsSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.TransferOwnership(&_Bindings.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bindings *BindingsTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bindings.Contract.TransferOwnership(&_Bindings.TransactOpts, newOwner)
}

// Update is a paid mutator transaction binding the contract method 0xa2e62045.
//
// Solidity: function update() returns()
func (_Bindings *BindingsTransactor) Update(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "update")
}

// Update is a paid mutator transaction binding the contract method 0xa2e62045.
//
// Solidity: function update() returns()
func (_Bindings *BindingsSession) Update() (*types.Transaction, error) {
	return _Bindings.Contract.Update(&_Bindings.TransactOpts)
}

// Update is a paid mutator transaction binding the contract method 0xa2e62045.
//
// Solidity: function update() returns()
func (_Bindings *BindingsTransactorSession) Update() (*types.Transaction, error) {
	return _Bindings.Contract.Update(&_Bindings.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_Bindings *BindingsTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bindings.contract.Transact(opts, "withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_Bindings *BindingsSession) Withdraw() (*types.Transaction, error) {
	return _Bindings.Contract.Withdraw(&_Bindings.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_Bindings *BindingsTransactorSession) Withdraw() (*types.Transaction, error) {
	return _Bindings.Contract.Withdraw(&_Bindings.TransactOpts)
}

// BindingsApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the Bindings contract.
type BindingsApprovalIterator struct {
	Event *BindingsApproval // Event containing the contract specifics and raw log

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
func (it *BindingsApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsApproval)
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
		it.Event = new(BindingsApproval)
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
func (it *BindingsApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsApproval represents a Approval event raised by the Bindings contract.
type BindingsApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_Bindings *BindingsFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*BindingsApprovalIterator, error) {

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

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &BindingsApprovalIterator{contract: _Bindings.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_Bindings *BindingsFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *BindingsApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsApproval)
				if err := _Bindings.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_Bindings *BindingsFilterer) ParseApproval(log types.Log) (*BindingsApproval, error) {
	event := new(BindingsApproval)
	if err := _Bindings.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the Bindings contract.
type BindingsApprovalForAllIterator struct {
	Event *BindingsApprovalForAll // Event containing the contract specifics and raw log

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
func (it *BindingsApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsApprovalForAll)
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
		it.Event = new(BindingsApprovalForAll)
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
func (it *BindingsApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsApprovalForAll represents a ApprovalForAll event raised by the Bindings contract.
type BindingsApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Bindings *BindingsFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*BindingsApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &BindingsApprovalForAllIterator{contract: _Bindings.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Bindings *BindingsFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *BindingsApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsApprovalForAll)
				if err := _Bindings.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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
func (_Bindings *BindingsFilterer) ParseApprovalForAll(log types.Log) (*BindingsApprovalForAll, error) {
	event := new(BindingsApprovalForAll)
	if err := _Bindings.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsBatchUpdatedIterator is returned from FilterBatchUpdated and is used to iterate over the raw logs and unpacked data for BatchUpdated events raised by the Bindings contract.
type BindingsBatchUpdatedIterator struct {
	Event *BindingsBatchUpdated // Event containing the contract specifics and raw log

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
func (it *BindingsBatchUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsBatchUpdated)
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
		it.Event = new(BindingsBatchUpdated)
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
func (it *BindingsBatchUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsBatchUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsBatchUpdated represents a BatchUpdated event raised by the Bindings contract.
type BindingsBatchUpdated struct {
	Previous *big.Int
	Updated  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBatchUpdated is a free log retrieval operation binding the contract event 0x656359bd8624a98c9559c454e7835a5e93f0867eacab61bfda9d2d0fce4e3097.
//
// Solidity: event BatchUpdated(uint256 previous, uint256 updated)
func (_Bindings *BindingsFilterer) FilterBatchUpdated(opts *bind.FilterOpts) (*BindingsBatchUpdatedIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "BatchUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingsBatchUpdatedIterator{contract: _Bindings.contract, event: "BatchUpdated", logs: logs, sub: sub}, nil
}

// WatchBatchUpdated is a free log subscription operation binding the contract event 0x656359bd8624a98c9559c454e7835a5e93f0867eacab61bfda9d2d0fce4e3097.
//
// Solidity: event BatchUpdated(uint256 previous, uint256 updated)
func (_Bindings *BindingsFilterer) WatchBatchUpdated(opts *bind.WatchOpts, sink chan<- *BindingsBatchUpdated) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "BatchUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsBatchUpdated)
				if err := _Bindings.contract.UnpackLog(event, "BatchUpdated", log); err != nil {
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

// ParseBatchUpdated is a log parse operation binding the contract event 0x656359bd8624a98c9559c454e7835a5e93f0867eacab61bfda9d2d0fce4e3097.
//
// Solidity: event BatchUpdated(uint256 previous, uint256 updated)
func (_Bindings *BindingsFilterer) ParseBatchUpdated(log types.Log) (*BindingsBatchUpdated, error) {
	event := new(BindingsBatchUpdated)
	if err := _Bindings.contract.UnpackLog(event, "BatchUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsHiddenUpdatedIterator is returned from FilterHiddenUpdated and is used to iterate over the raw logs and unpacked data for HiddenUpdated events raised by the Bindings contract.
type BindingsHiddenUpdatedIterator struct {
	Event *BindingsHiddenUpdated // Event containing the contract specifics and raw log

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
func (it *BindingsHiddenUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsHiddenUpdated)
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
		it.Event = new(BindingsHiddenUpdated)
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
func (it *BindingsHiddenUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsHiddenUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsHiddenUpdated represents a HiddenUpdated event raised by the Bindings contract.
type BindingsHiddenUpdated struct {
	Previous string
	Updated  string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterHiddenUpdated is a free log retrieval operation binding the contract event 0x57189fe32dda35d8b092ccd5a7d65eeecb46be25ce11a28ad02b5b0d3f662505.
//
// Solidity: event HiddenUpdated(string previous, string updated)
func (_Bindings *BindingsFilterer) FilterHiddenUpdated(opts *bind.FilterOpts) (*BindingsHiddenUpdatedIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "HiddenUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingsHiddenUpdatedIterator{contract: _Bindings.contract, event: "HiddenUpdated", logs: logs, sub: sub}, nil
}

// WatchHiddenUpdated is a free log subscription operation binding the contract event 0x57189fe32dda35d8b092ccd5a7d65eeecb46be25ce11a28ad02b5b0d3f662505.
//
// Solidity: event HiddenUpdated(string previous, string updated)
func (_Bindings *BindingsFilterer) WatchHiddenUpdated(opts *bind.WatchOpts, sink chan<- *BindingsHiddenUpdated) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "HiddenUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsHiddenUpdated)
				if err := _Bindings.contract.UnpackLog(event, "HiddenUpdated", log); err != nil {
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

// ParseHiddenUpdated is a log parse operation binding the contract event 0x57189fe32dda35d8b092ccd5a7d65eeecb46be25ce11a28ad02b5b0d3f662505.
//
// Solidity: event HiddenUpdated(string previous, string updated)
func (_Bindings *BindingsFilterer) ParseHiddenUpdated(log types.Log) (*BindingsHiddenUpdated, error) {
	event := new(BindingsHiddenUpdated)
	if err := _Bindings.contract.UnpackLog(event, "HiddenUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsMaxUpdatedIterator is returned from FilterMaxUpdated and is used to iterate over the raw logs and unpacked data for MaxUpdated events raised by the Bindings contract.
type BindingsMaxUpdatedIterator struct {
	Event *BindingsMaxUpdated // Event containing the contract specifics and raw log

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
func (it *BindingsMaxUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsMaxUpdated)
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
		it.Event = new(BindingsMaxUpdated)
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
func (it *BindingsMaxUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsMaxUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsMaxUpdated represents a MaxUpdated event raised by the Bindings contract.
type BindingsMaxUpdated struct {
	Previous *big.Int
	Updated  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMaxUpdated is a free log retrieval operation binding the contract event 0xaa6f6b0a509f2b07cf30d89dbd3bb410883aaa429ad4da41fdf36c02398cf1a0.
//
// Solidity: event MaxUpdated(uint256 previous, uint256 updated)
func (_Bindings *BindingsFilterer) FilterMaxUpdated(opts *bind.FilterOpts) (*BindingsMaxUpdatedIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "MaxUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingsMaxUpdatedIterator{contract: _Bindings.contract, event: "MaxUpdated", logs: logs, sub: sub}, nil
}

// WatchMaxUpdated is a free log subscription operation binding the contract event 0xaa6f6b0a509f2b07cf30d89dbd3bb410883aaa429ad4da41fdf36c02398cf1a0.
//
// Solidity: event MaxUpdated(uint256 previous, uint256 updated)
func (_Bindings *BindingsFilterer) WatchMaxUpdated(opts *bind.WatchOpts, sink chan<- *BindingsMaxUpdated) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "MaxUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsMaxUpdated)
				if err := _Bindings.contract.UnpackLog(event, "MaxUpdated", log); err != nil {
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
func (_Bindings *BindingsFilterer) ParseMaxUpdated(log types.Log) (*BindingsMaxUpdated, error) {
	event := new(BindingsMaxUpdated)
	if err := _Bindings.contract.UnpackLog(event, "MaxUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsMintableUpdatedIterator is returned from FilterMintableUpdated and is used to iterate over the raw logs and unpacked data for MintableUpdated events raised by the Bindings contract.
type BindingsMintableUpdatedIterator struct {
	Event *BindingsMintableUpdated // Event containing the contract specifics and raw log

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
func (it *BindingsMintableUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsMintableUpdated)
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
		it.Event = new(BindingsMintableUpdated)
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
func (it *BindingsMintableUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsMintableUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsMintableUpdated represents a MintableUpdated event raised by the Bindings contract.
type BindingsMintableUpdated struct {
	Previous bool
	Updated  bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMintableUpdated is a free log retrieval operation binding the contract event 0x8d9383d773c0600295154578f39da3106938ba8d1fe1767bcfabe8bf05f555f4.
//
// Solidity: event MintableUpdated(bool previous, bool updated)
func (_Bindings *BindingsFilterer) FilterMintableUpdated(opts *bind.FilterOpts) (*BindingsMintableUpdatedIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "MintableUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingsMintableUpdatedIterator{contract: _Bindings.contract, event: "MintableUpdated", logs: logs, sub: sub}, nil
}

// WatchMintableUpdated is a free log subscription operation binding the contract event 0x8d9383d773c0600295154578f39da3106938ba8d1fe1767bcfabe8bf05f555f4.
//
// Solidity: event MintableUpdated(bool previous, bool updated)
func (_Bindings *BindingsFilterer) WatchMintableUpdated(opts *bind.WatchOpts, sink chan<- *BindingsMintableUpdated) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "MintableUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsMintableUpdated)
				if err := _Bindings.contract.UnpackLog(event, "MintableUpdated", log); err != nil {
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
func (_Bindings *BindingsFilterer) ParseMintableUpdated(log types.Log) (*BindingsMintableUpdated, error) {
	event := new(BindingsMintableUpdated)
	if err := _Bindings.contract.UnpackLog(event, "MintableUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Bindings contract.
type BindingsOwnershipTransferredIterator struct {
	Event *BindingsOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BindingsOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsOwnershipTransferred)
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
		it.Event = new(BindingsOwnershipTransferred)
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
func (it *BindingsOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsOwnershipTransferred represents a OwnershipTransferred event raised by the Bindings contract.
type BindingsOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bindings *BindingsFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BindingsOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BindingsOwnershipTransferredIterator{contract: _Bindings.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bindings *BindingsFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BindingsOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsOwnershipTransferred)
				if err := _Bindings.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Bindings *BindingsFilterer) ParseOwnershipTransferred(log types.Log) (*BindingsOwnershipTransferred, error) {
	event := new(BindingsOwnershipTransferred)
	if err := _Bindings.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsPackUpdatedIterator is returned from FilterPackUpdated and is used to iterate over the raw logs and unpacked data for PackUpdated events raised by the Bindings contract.
type BindingsPackUpdatedIterator struct {
	Event *BindingsPackUpdated // Event containing the contract specifics and raw log

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
func (it *BindingsPackUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsPackUpdated)
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
		it.Event = new(BindingsPackUpdated)
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
func (it *BindingsPackUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsPackUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsPackUpdated represents a PackUpdated event raised by the Bindings contract.
type BindingsPackUpdated struct {
	Previous common.Address
	Updated  common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPackUpdated is a free log retrieval operation binding the contract event 0x69c956910dd41d384c8eaa85a91003a538ae5d5ceb57ae4e530072e3908b10f4.
//
// Solidity: event PackUpdated(address previous, address updated)
func (_Bindings *BindingsFilterer) FilterPackUpdated(opts *bind.FilterOpts) (*BindingsPackUpdatedIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "PackUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingsPackUpdatedIterator{contract: _Bindings.contract, event: "PackUpdated", logs: logs, sub: sub}, nil
}

// WatchPackUpdated is a free log subscription operation binding the contract event 0x69c956910dd41d384c8eaa85a91003a538ae5d5ceb57ae4e530072e3908b10f4.
//
// Solidity: event PackUpdated(address previous, address updated)
func (_Bindings *BindingsFilterer) WatchPackUpdated(opts *bind.WatchOpts, sink chan<- *BindingsPackUpdated) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "PackUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsPackUpdated)
				if err := _Bindings.contract.UnpackLog(event, "PackUpdated", log); err != nil {
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

// ParsePackUpdated is a log parse operation binding the contract event 0x69c956910dd41d384c8eaa85a91003a538ae5d5ceb57ae4e530072e3908b10f4.
//
// Solidity: event PackUpdated(address previous, address updated)
func (_Bindings *BindingsFilterer) ParsePackUpdated(log types.Log) (*BindingsPackUpdated, error) {
	event := new(BindingsPackUpdated)
	if err := _Bindings.contract.UnpackLog(event, "PackUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsPriceUpdatedIterator is returned from FilterPriceUpdated and is used to iterate over the raw logs and unpacked data for PriceUpdated events raised by the Bindings contract.
type BindingsPriceUpdatedIterator struct {
	Event *BindingsPriceUpdated // Event containing the contract specifics and raw log

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
func (it *BindingsPriceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsPriceUpdated)
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
		it.Event = new(BindingsPriceUpdated)
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
func (it *BindingsPriceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsPriceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsPriceUpdated represents a PriceUpdated event raised by the Bindings contract.
type BindingsPriceUpdated struct {
	Previous *big.Int
	Updated  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPriceUpdated is a free log retrieval operation binding the contract event 0x945c1c4e99aa89f648fbfe3df471b916f719e16d960fcec0737d4d56bd696838.
//
// Solidity: event PriceUpdated(uint256 previous, uint256 updated)
func (_Bindings *BindingsFilterer) FilterPriceUpdated(opts *bind.FilterOpts) (*BindingsPriceUpdatedIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "PriceUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingsPriceUpdatedIterator{contract: _Bindings.contract, event: "PriceUpdated", logs: logs, sub: sub}, nil
}

// WatchPriceUpdated is a free log subscription operation binding the contract event 0x945c1c4e99aa89f648fbfe3df471b916f719e16d960fcec0737d4d56bd696838.
//
// Solidity: event PriceUpdated(uint256 previous, uint256 updated)
func (_Bindings *BindingsFilterer) WatchPriceUpdated(opts *bind.WatchOpts, sink chan<- *BindingsPriceUpdated) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "PriceUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsPriceUpdated)
				if err := _Bindings.contract.UnpackLog(event, "PriceUpdated", log); err != nil {
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

// ParsePriceUpdated is a log parse operation binding the contract event 0x945c1c4e99aa89f648fbfe3df471b916f719e16d960fcec0737d4d56bd696838.
//
// Solidity: event PriceUpdated(uint256 previous, uint256 updated)
func (_Bindings *BindingsFilterer) ParsePriceUpdated(log types.Log) (*BindingsPriceUpdated, error) {
	event := new(BindingsPriceUpdated)
	if err := _Bindings.contract.UnpackLog(event, "PriceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsReceiverUpdatedIterator is returned from FilterReceiverUpdated and is used to iterate over the raw logs and unpacked data for ReceiverUpdated events raised by the Bindings contract.
type BindingsReceiverUpdatedIterator struct {
	Event *BindingsReceiverUpdated // Event containing the contract specifics and raw log

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
func (it *BindingsReceiverUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsReceiverUpdated)
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
		it.Event = new(BindingsReceiverUpdated)
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
func (it *BindingsReceiverUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsReceiverUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsReceiverUpdated represents a ReceiverUpdated event raised by the Bindings contract.
type BindingsReceiverUpdated struct {
	Previous common.Address
	Updated  common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterReceiverUpdated is a free log retrieval operation binding the contract event 0xbda2bcccbfa5ae883ab7d9f03480ab68fe68e9200c9b52c0c47abc21d2c90ec9.
//
// Solidity: event ReceiverUpdated(address previous, address updated)
func (_Bindings *BindingsFilterer) FilterReceiverUpdated(opts *bind.FilterOpts) (*BindingsReceiverUpdatedIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "ReceiverUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingsReceiverUpdatedIterator{contract: _Bindings.contract, event: "ReceiverUpdated", logs: logs, sub: sub}, nil
}

// WatchReceiverUpdated is a free log subscription operation binding the contract event 0xbda2bcccbfa5ae883ab7d9f03480ab68fe68e9200c9b52c0c47abc21d2c90ec9.
//
// Solidity: event ReceiverUpdated(address previous, address updated)
func (_Bindings *BindingsFilterer) WatchReceiverUpdated(opts *bind.WatchOpts, sink chan<- *BindingsReceiverUpdated) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "ReceiverUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsReceiverUpdated)
				if err := _Bindings.contract.UnpackLog(event, "ReceiverUpdated", log); err != nil {
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
func (_Bindings *BindingsFilterer) ParseReceiverUpdated(log types.Log) (*BindingsReceiverUpdated, error) {
	event := new(BindingsReceiverUpdated)
	if err := _Bindings.contract.UnpackLog(event, "ReceiverUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsRevealableUpdatedIterator is returned from FilterRevealableUpdated and is used to iterate over the raw logs and unpacked data for RevealableUpdated events raised by the Bindings contract.
type BindingsRevealableUpdatedIterator struct {
	Event *BindingsRevealableUpdated // Event containing the contract specifics and raw log

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
func (it *BindingsRevealableUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsRevealableUpdated)
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
		it.Event = new(BindingsRevealableUpdated)
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
func (it *BindingsRevealableUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsRevealableUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsRevealableUpdated represents a RevealableUpdated event raised by the Bindings contract.
type BindingsRevealableUpdated struct {
	Previous bool
	Updated  bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRevealableUpdated is a free log retrieval operation binding the contract event 0x84cc45418127802859711bdf440a92a3a6d3819145166e3f884db8d202aa5ebf.
//
// Solidity: event RevealableUpdated(bool previous, bool updated)
func (_Bindings *BindingsFilterer) FilterRevealableUpdated(opts *bind.FilterOpts) (*BindingsRevealableUpdatedIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "RevealableUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingsRevealableUpdatedIterator{contract: _Bindings.contract, event: "RevealableUpdated", logs: logs, sub: sub}, nil
}

// WatchRevealableUpdated is a free log subscription operation binding the contract event 0x84cc45418127802859711bdf440a92a3a6d3819145166e3f884db8d202aa5ebf.
//
// Solidity: event RevealableUpdated(bool previous, bool updated)
func (_Bindings *BindingsFilterer) WatchRevealableUpdated(opts *bind.WatchOpts, sink chan<- *BindingsRevealableUpdated) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "RevealableUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsRevealableUpdated)
				if err := _Bindings.contract.UnpackLog(event, "RevealableUpdated", log); err != nil {
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

// ParseRevealableUpdated is a log parse operation binding the contract event 0x84cc45418127802859711bdf440a92a3a6d3819145166e3f884db8d202aa5ebf.
//
// Solidity: event RevealableUpdated(bool previous, bool updated)
func (_Bindings *BindingsFilterer) ParseRevealableUpdated(log types.Log) (*BindingsRevealableUpdated, error) {
	event := new(BindingsRevealableUpdated)
	if err := _Bindings.contract.UnpackLog(event, "RevealableUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsRoyaltyUpdatedIterator is returned from FilterRoyaltyUpdated and is used to iterate over the raw logs and unpacked data for RoyaltyUpdated events raised by the Bindings contract.
type BindingsRoyaltyUpdatedIterator struct {
	Event *BindingsRoyaltyUpdated // Event containing the contract specifics and raw log

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
func (it *BindingsRoyaltyUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsRoyaltyUpdated)
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
		it.Event = new(BindingsRoyaltyUpdated)
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
func (it *BindingsRoyaltyUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsRoyaltyUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsRoyaltyUpdated represents a RoyaltyUpdated event raised by the Bindings contract.
type BindingsRoyaltyUpdated struct {
	Previous *big.Int
	Updated  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRoyaltyUpdated is a free log retrieval operation binding the contract event 0x54e506cda8889617ec187c699f1c3b373053eb5796248194796f7e1501dfab24.
//
// Solidity: event RoyaltyUpdated(uint256 previous, uint256 updated)
func (_Bindings *BindingsFilterer) FilterRoyaltyUpdated(opts *bind.FilterOpts) (*BindingsRoyaltyUpdatedIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "RoyaltyUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingsRoyaltyUpdatedIterator{contract: _Bindings.contract, event: "RoyaltyUpdated", logs: logs, sub: sub}, nil
}

// WatchRoyaltyUpdated is a free log subscription operation binding the contract event 0x54e506cda8889617ec187c699f1c3b373053eb5796248194796f7e1501dfab24.
//
// Solidity: event RoyaltyUpdated(uint256 previous, uint256 updated)
func (_Bindings *BindingsFilterer) WatchRoyaltyUpdated(opts *bind.WatchOpts, sink chan<- *BindingsRoyaltyUpdated) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "RoyaltyUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsRoyaltyUpdated)
				if err := _Bindings.contract.UnpackLog(event, "RoyaltyUpdated", log); err != nil {
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
func (_Bindings *BindingsFilterer) ParseRoyaltyUpdated(log types.Log) (*BindingsRoyaltyUpdated, error) {
	event := new(BindingsRoyaltyUpdated)
	if err := _Bindings.contract.UnpackLog(event, "RoyaltyUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsTokenRevealIterator is returned from FilterTokenReveal and is used to iterate over the raw logs and unpacked data for TokenReveal events raised by the Bindings contract.
type BindingsTokenRevealIterator struct {
	Event *BindingsTokenReveal // Event containing the contract specifics and raw log

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
func (it *BindingsTokenRevealIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsTokenReveal)
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
		it.Event = new(BindingsTokenReveal)
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
func (it *BindingsTokenRevealIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsTokenRevealIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsTokenReveal represents a TokenReveal event raised by the Bindings contract.
type BindingsTokenReveal struct {
	User      common.Address
	RequestId [32]byte
	Tokens    []*big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTokenReveal is a free log retrieval operation binding the contract event 0xf78fdbb14693a4df69f9724dd4d02cb841094dbff2f6dac48ad20cda511f2d12.
//
// Solidity: event TokenReveal(address indexed user, bytes32 requestId, uint256[] tokens)
func (_Bindings *BindingsFilterer) FilterTokenReveal(opts *bind.FilterOpts, user []common.Address) (*BindingsTokenRevealIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "TokenReveal", userRule)
	if err != nil {
		return nil, err
	}
	return &BindingsTokenRevealIterator{contract: _Bindings.contract, event: "TokenReveal", logs: logs, sub: sub}, nil
}

// WatchTokenReveal is a free log subscription operation binding the contract event 0xf78fdbb14693a4df69f9724dd4d02cb841094dbff2f6dac48ad20cda511f2d12.
//
// Solidity: event TokenReveal(address indexed user, bytes32 requestId, uint256[] tokens)
func (_Bindings *BindingsFilterer) WatchTokenReveal(opts *bind.WatchOpts, sink chan<- *BindingsTokenReveal, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "TokenReveal", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsTokenReveal)
				if err := _Bindings.contract.UnpackLog(event, "TokenReveal", log); err != nil {
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

// ParseTokenReveal is a log parse operation binding the contract event 0xf78fdbb14693a4df69f9724dd4d02cb841094dbff2f6dac48ad20cda511f2d12.
//
// Solidity: event TokenReveal(address indexed user, bytes32 requestId, uint256[] tokens)
func (_Bindings *BindingsFilterer) ParseTokenReveal(log types.Log) (*BindingsTokenReveal, error) {
	event := new(BindingsTokenReveal)
	if err := _Bindings.contract.UnpackLog(event, "TokenReveal", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the Bindings contract.
type BindingsTransferIterator struct {
	Event *BindingsTransfer // Event containing the contract specifics and raw log

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
func (it *BindingsTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsTransfer)
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
		it.Event = new(BindingsTransfer)
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
func (it *BindingsTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsTransfer represents a Transfer event raised by the Bindings contract.
type BindingsTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Bindings *BindingsFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*BindingsTransferIterator, error) {

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

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &BindingsTransferIterator{contract: _Bindings.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Bindings *BindingsFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *BindingsTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsTransfer)
				if err := _Bindings.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_Bindings *BindingsFilterer) ParseTransfer(log types.Log) (*BindingsTransfer, error) {
	event := new(BindingsTransfer)
	if err := _Bindings.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsUriFallbackUpdatedIterator is returned from FilterUriFallbackUpdated and is used to iterate over the raw logs and unpacked data for UriFallbackUpdated events raised by the Bindings contract.
type BindingsUriFallbackUpdatedIterator struct {
	Event *BindingsUriFallbackUpdated // Event containing the contract specifics and raw log

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
func (it *BindingsUriFallbackUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsUriFallbackUpdated)
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
		it.Event = new(BindingsUriFallbackUpdated)
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
func (it *BindingsUriFallbackUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsUriFallbackUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsUriFallbackUpdated represents a UriFallbackUpdated event raised by the Bindings contract.
type BindingsUriFallbackUpdated struct {
	Previous string
	Updated  string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterUriFallbackUpdated is a free log retrieval operation binding the contract event 0xe1b7ff5efe58018e39b7877b5cfa772bb90f32504be7b2330b078d2a9b114bbe.
//
// Solidity: event UriFallbackUpdated(string previous, string updated)
func (_Bindings *BindingsFilterer) FilterUriFallbackUpdated(opts *bind.FilterOpts) (*BindingsUriFallbackUpdatedIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "UriFallbackUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingsUriFallbackUpdatedIterator{contract: _Bindings.contract, event: "UriFallbackUpdated", logs: logs, sub: sub}, nil
}

// WatchUriFallbackUpdated is a free log subscription operation binding the contract event 0xe1b7ff5efe58018e39b7877b5cfa772bb90f32504be7b2330b078d2a9b114bbe.
//
// Solidity: event UriFallbackUpdated(string previous, string updated)
func (_Bindings *BindingsFilterer) WatchUriFallbackUpdated(opts *bind.WatchOpts, sink chan<- *BindingsUriFallbackUpdated) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "UriFallbackUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsUriFallbackUpdated)
				if err := _Bindings.contract.UnpackLog(event, "UriFallbackUpdated", log); err != nil {
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
func (_Bindings *BindingsFilterer) ParseUriFallbackUpdated(log types.Log) (*BindingsUriFallbackUpdated, error) {
	event := new(BindingsUriFallbackUpdated)
	if err := _Bindings.contract.UnpackLog(event, "UriFallbackUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingsUriUpdatedIterator is returned from FilterUriUpdated and is used to iterate over the raw logs and unpacked data for UriUpdated events raised by the Bindings contract.
type BindingsUriUpdatedIterator struct {
	Event *BindingsUriUpdated // Event containing the contract specifics and raw log

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
func (it *BindingsUriUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingsUriUpdated)
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
		it.Event = new(BindingsUriUpdated)
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
func (it *BindingsUriUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingsUriUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingsUriUpdated represents a UriUpdated event raised by the Bindings contract.
type BindingsUriUpdated struct {
	Previous string
	Updated  string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterUriUpdated is a free log retrieval operation binding the contract event 0x7d8ebb5abe647a67ba3a2649e11557ae5aa256cf3449245e0c840c98132e5a37.
//
// Solidity: event UriUpdated(string previous, string updated)
func (_Bindings *BindingsFilterer) FilterUriUpdated(opts *bind.FilterOpts) (*BindingsUriUpdatedIterator, error) {

	logs, sub, err := _Bindings.contract.FilterLogs(opts, "UriUpdated")
	if err != nil {
		return nil, err
	}
	return &BindingsUriUpdatedIterator{contract: _Bindings.contract, event: "UriUpdated", logs: logs, sub: sub}, nil
}

// WatchUriUpdated is a free log subscription operation binding the contract event 0x7d8ebb5abe647a67ba3a2649e11557ae5aa256cf3449245e0c840c98132e5a37.
//
// Solidity: event UriUpdated(string previous, string updated)
func (_Bindings *BindingsFilterer) WatchUriUpdated(opts *bind.WatchOpts, sink chan<- *BindingsUriUpdated) (event.Subscription, error) {

	logs, sub, err := _Bindings.contract.WatchLogs(opts, "UriUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingsUriUpdated)
				if err := _Bindings.contract.UnpackLog(event, "UriUpdated", log); err != nil {
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
func (_Bindings *BindingsFilterer) ParseUriUpdated(log types.Log) (*BindingsUriUpdated, error) {
	event := new(BindingsUriUpdated)
	if err := _Bindings.contract.UnpackLog(event, "UriUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
