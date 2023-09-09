// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bancor

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

// IBancorNetworkMetaData contains all meta data concerning the IBancorNetwork contract.
var IBancorNetworkMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"cancelWithdrawal\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"pool\",\"type\":\"address\"}],\"name\":\"collectionByPool\",\"outputs\":[{\"internalType\":\"contractIPoolCollection\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken[]\",\"name\":\"tokens\",\"type\":\"address[]\"},{\"internalType\":\"contractIPoolCollection\",\"name\":\"poolCollection\",\"type\":\"address\"}],\"name\":\"createPools\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"pool\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenAmount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"contractToken\",\"name\":\"pool\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenAmount\",\"type\":\"uint256\"}],\"name\":\"depositFor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"contractIFlashLoanRecipient\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"flashLoan\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getRoleMember\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleMemberCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIPoolToken\",\"name\":\"poolToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"poolTokenAmount\",\"type\":\"uint256\"}],\"name\":\"initWithdrawal\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"liquidityPools\",\"outputs\":[{\"internalType\":\"contractToken[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"availableAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"originalAmount\",\"type\":\"uint256\"}],\"name\":\"migrateLiquidity\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken[]\",\"name\":\"pools\",\"type\":\"address[]\"},{\"internalType\":\"contractIPoolCollection\",\"name\":\"newPoolCollection\",\"type\":\"address\"}],\"name\":\"migratePools\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"poolCollections\",\"outputs\":[{\"internalType\":\"contractIPoolCollection[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"contractToken\",\"name\":\"targetToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"sourceAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minReturnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"tradeBySourceAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"contractToken\",\"name\":\"targetToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"sourceAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minReturnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"tradeBySourceAmountArb\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"contractToken\",\"name\":\"targetToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"targetAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxSourceAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"tradeByTargetAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"contractToken\",\"name\":\"targetToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"targetAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxSourceAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"tradeByTargetAmountArb\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"withdrawNetworkFees\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"pool\",\"type\":\"address\"}],\"name\":\"withdrawPOL\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// IBancorNetworkABI is the input ABI used to generate the binding from.
// Deprecated: Use IBancorNetworkMetaData.ABI instead.
var IBancorNetworkABI = IBancorNetworkMetaData.ABI

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
	parsed, err := IBancorNetworkMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
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

// CollectionByPool is a free data retrieval call binding the contract method 0x9bca0e70.
//
// Solidity: function collectionByPool(address pool) view returns(address)
func (_IBancorNetwork *IBancorNetworkCaller) CollectionByPool(opts *bind.CallOpts, pool common.Address) (common.Address, error) {
	var out []interface{}
	err := _IBancorNetwork.contract.Call(opts, &out, "collectionByPool", pool)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CollectionByPool is a free data retrieval call binding the contract method 0x9bca0e70.
//
// Solidity: function collectionByPool(address pool) view returns(address)
func (_IBancorNetwork *IBancorNetworkSession) CollectionByPool(pool common.Address) (common.Address, error) {
	return _IBancorNetwork.Contract.CollectionByPool(&_IBancorNetwork.CallOpts, pool)
}

// CollectionByPool is a free data retrieval call binding the contract method 0x9bca0e70.
//
// Solidity: function collectionByPool(address pool) view returns(address)
func (_IBancorNetwork *IBancorNetworkCallerSession) CollectionByPool(pool common.Address) (common.Address, error) {
	return _IBancorNetwork.Contract.CollectionByPool(&_IBancorNetwork.CallOpts, pool)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_IBancorNetwork *IBancorNetworkCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _IBancorNetwork.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_IBancorNetwork *IBancorNetworkSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _IBancorNetwork.Contract.GetRoleAdmin(&_IBancorNetwork.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_IBancorNetwork *IBancorNetworkCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _IBancorNetwork.Contract.GetRoleAdmin(&_IBancorNetwork.CallOpts, role)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_IBancorNetwork *IBancorNetworkCaller) GetRoleMember(opts *bind.CallOpts, role [32]byte, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _IBancorNetwork.contract.Call(opts, &out, "getRoleMember", role, index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_IBancorNetwork *IBancorNetworkSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _IBancorNetwork.Contract.GetRoleMember(&_IBancorNetwork.CallOpts, role, index)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_IBancorNetwork *IBancorNetworkCallerSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _IBancorNetwork.Contract.GetRoleMember(&_IBancorNetwork.CallOpts, role, index)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_IBancorNetwork *IBancorNetworkCaller) GetRoleMemberCount(opts *bind.CallOpts, role [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _IBancorNetwork.contract.Call(opts, &out, "getRoleMemberCount", role)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _IBancorNetwork.Contract.GetRoleMemberCount(&_IBancorNetwork.CallOpts, role)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_IBancorNetwork *IBancorNetworkCallerSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _IBancorNetwork.Contract.GetRoleMemberCount(&_IBancorNetwork.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_IBancorNetwork *IBancorNetworkCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _IBancorNetwork.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_IBancorNetwork *IBancorNetworkSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _IBancorNetwork.Contract.HasRole(&_IBancorNetwork.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_IBancorNetwork *IBancorNetworkCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _IBancorNetwork.Contract.HasRole(&_IBancorNetwork.CallOpts, role, account)
}

// LiquidityPools is a free data retrieval call binding the contract method 0xd6efd7c3.
//
// Solidity: function liquidityPools() view returns(address[])
func (_IBancorNetwork *IBancorNetworkCaller) LiquidityPools(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _IBancorNetwork.contract.Call(opts, &out, "liquidityPools")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// LiquidityPools is a free data retrieval call binding the contract method 0xd6efd7c3.
//
// Solidity: function liquidityPools() view returns(address[])
func (_IBancorNetwork *IBancorNetworkSession) LiquidityPools() ([]common.Address, error) {
	return _IBancorNetwork.Contract.LiquidityPools(&_IBancorNetwork.CallOpts)
}

// LiquidityPools is a free data retrieval call binding the contract method 0xd6efd7c3.
//
// Solidity: function liquidityPools() view returns(address[])
func (_IBancorNetwork *IBancorNetworkCallerSession) LiquidityPools() ([]common.Address, error) {
	return _IBancorNetwork.Contract.LiquidityPools(&_IBancorNetwork.CallOpts)
}

// PoolCollections is a free data retrieval call binding the contract method 0x39fadf98.
//
// Solidity: function poolCollections() view returns(address[])
func (_IBancorNetwork *IBancorNetworkCaller) PoolCollections(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _IBancorNetwork.contract.Call(opts, &out, "poolCollections")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// PoolCollections is a free data retrieval call binding the contract method 0x39fadf98.
//
// Solidity: function poolCollections() view returns(address[])
func (_IBancorNetwork *IBancorNetworkSession) PoolCollections() ([]common.Address, error) {
	return _IBancorNetwork.Contract.PoolCollections(&_IBancorNetwork.CallOpts)
}

// PoolCollections is a free data retrieval call binding the contract method 0x39fadf98.
//
// Solidity: function poolCollections() view returns(address[])
func (_IBancorNetwork *IBancorNetworkCallerSession) PoolCollections() ([]common.Address, error) {
	return _IBancorNetwork.Contract.PoolCollections(&_IBancorNetwork.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(uint16)
func (_IBancorNetwork *IBancorNetworkCaller) Version(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _IBancorNetwork.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(uint16)
func (_IBancorNetwork *IBancorNetworkSession) Version() (uint16, error) {
	return _IBancorNetwork.Contract.Version(&_IBancorNetwork.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(uint16)
func (_IBancorNetwork *IBancorNetworkCallerSession) Version() (uint16, error) {
	return _IBancorNetwork.Contract.Version(&_IBancorNetwork.CallOpts)
}

// CancelWithdrawal is a paid mutator transaction binding the contract method 0x3efcfda4.
//
// Solidity: function cancelWithdrawal(uint256 id) returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) CancelWithdrawal(opts *bind.TransactOpts, id *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "cancelWithdrawal", id)
}

// CancelWithdrawal is a paid mutator transaction binding the contract method 0x3efcfda4.
//
// Solidity: function cancelWithdrawal(uint256 id) returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) CancelWithdrawal(id *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.CancelWithdrawal(&_IBancorNetwork.TransactOpts, id)
}

// CancelWithdrawal is a paid mutator transaction binding the contract method 0x3efcfda4.
//
// Solidity: function cancelWithdrawal(uint256 id) returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) CancelWithdrawal(id *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.CancelWithdrawal(&_IBancorNetwork.TransactOpts, id)
}

// CreatePools is a paid mutator transaction binding the contract method 0x42659964.
//
// Solidity: function createPools(address[] tokens, address poolCollection) returns()
func (_IBancorNetwork *IBancorNetworkTransactor) CreatePools(opts *bind.TransactOpts, tokens []common.Address, poolCollection common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "createPools", tokens, poolCollection)
}

// CreatePools is a paid mutator transaction binding the contract method 0x42659964.
//
// Solidity: function createPools(address[] tokens, address poolCollection) returns()
func (_IBancorNetwork *IBancorNetworkSession) CreatePools(tokens []common.Address, poolCollection common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.CreatePools(&_IBancorNetwork.TransactOpts, tokens, poolCollection)
}

// CreatePools is a paid mutator transaction binding the contract method 0x42659964.
//
// Solidity: function createPools(address[] tokens, address poolCollection) returns()
func (_IBancorNetwork *IBancorNetworkTransactorSession) CreatePools(tokens []common.Address, poolCollection common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.CreatePools(&_IBancorNetwork.TransactOpts, tokens, poolCollection)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address pool, uint256 tokenAmount) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) Deposit(opts *bind.TransactOpts, pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "deposit", pool, tokenAmount)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address pool, uint256 tokenAmount) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) Deposit(pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.Deposit(&_IBancorNetwork.TransactOpts, pool, tokenAmount)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address pool, uint256 tokenAmount) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) Deposit(pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.Deposit(&_IBancorNetwork.TransactOpts, pool, tokenAmount)
}

// DepositFor is a paid mutator transaction binding the contract method 0xb3db428b.
//
// Solidity: function depositFor(address provider, address pool, uint256 tokenAmount) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) DepositFor(opts *bind.TransactOpts, provider common.Address, pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "depositFor", provider, pool, tokenAmount)
}

// DepositFor is a paid mutator transaction binding the contract method 0xb3db428b.
//
// Solidity: function depositFor(address provider, address pool, uint256 tokenAmount) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) DepositFor(provider common.Address, pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.DepositFor(&_IBancorNetwork.TransactOpts, provider, pool, tokenAmount)
}

// DepositFor is a paid mutator transaction binding the contract method 0xb3db428b.
//
// Solidity: function depositFor(address provider, address pool, uint256 tokenAmount) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) DepositFor(provider common.Address, pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.DepositFor(&_IBancorNetwork.TransactOpts, provider, pool, tokenAmount)
}

// FlashLoan is a paid mutator transaction binding the contract method 0xadf51de1.
//
// Solidity: function flashLoan(address token, uint256 amount, address recipient, bytes data) returns()
func (_IBancorNetwork *IBancorNetworkTransactor) FlashLoan(opts *bind.TransactOpts, token common.Address, amount *big.Int, recipient common.Address, data []byte) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "flashLoan", token, amount, recipient, data)
}

// FlashLoan is a paid mutator transaction binding the contract method 0xadf51de1.
//
// Solidity: function flashLoan(address token, uint256 amount, address recipient, bytes data) returns()
func (_IBancorNetwork *IBancorNetworkSession) FlashLoan(token common.Address, amount *big.Int, recipient common.Address, data []byte) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.FlashLoan(&_IBancorNetwork.TransactOpts, token, amount, recipient, data)
}

// FlashLoan is a paid mutator transaction binding the contract method 0xadf51de1.
//
// Solidity: function flashLoan(address token, uint256 amount, address recipient, bytes data) returns()
func (_IBancorNetwork *IBancorNetworkTransactorSession) FlashLoan(token common.Address, amount *big.Int, recipient common.Address, data []byte) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.FlashLoan(&_IBancorNetwork.TransactOpts, token, amount, recipient, data)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_IBancorNetwork *IBancorNetworkTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_IBancorNetwork *IBancorNetworkSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.GrantRole(&_IBancorNetwork.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_IBancorNetwork *IBancorNetworkTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.GrantRole(&_IBancorNetwork.TransactOpts, role, account)
}

// InitWithdrawal is a paid mutator transaction binding the contract method 0x357a0333.
//
// Solidity: function initWithdrawal(address poolToken, uint256 poolTokenAmount) returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) InitWithdrawal(opts *bind.TransactOpts, poolToken common.Address, poolTokenAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "initWithdrawal", poolToken, poolTokenAmount)
}

// InitWithdrawal is a paid mutator transaction binding the contract method 0x357a0333.
//
// Solidity: function initWithdrawal(address poolToken, uint256 poolTokenAmount) returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) InitWithdrawal(poolToken common.Address, poolTokenAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.InitWithdrawal(&_IBancorNetwork.TransactOpts, poolToken, poolTokenAmount)
}

// InitWithdrawal is a paid mutator transaction binding the contract method 0x357a0333.
//
// Solidity: function initWithdrawal(address poolToken, uint256 poolTokenAmount) returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) InitWithdrawal(poolToken common.Address, poolTokenAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.InitWithdrawal(&_IBancorNetwork.TransactOpts, poolToken, poolTokenAmount)
}

// MigrateLiquidity is a paid mutator transaction binding the contract method 0x3d1c24e7.
//
// Solidity: function migrateLiquidity(address token, address provider, uint256 amount, uint256 availableAmount, uint256 originalAmount) payable returns()
func (_IBancorNetwork *IBancorNetworkTransactor) MigrateLiquidity(opts *bind.TransactOpts, token common.Address, provider common.Address, amount *big.Int, availableAmount *big.Int, originalAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "migrateLiquidity", token, provider, amount, availableAmount, originalAmount)
}

// MigrateLiquidity is a paid mutator transaction binding the contract method 0x3d1c24e7.
//
// Solidity: function migrateLiquidity(address token, address provider, uint256 amount, uint256 availableAmount, uint256 originalAmount) payable returns()
func (_IBancorNetwork *IBancorNetworkSession) MigrateLiquidity(token common.Address, provider common.Address, amount *big.Int, availableAmount *big.Int, originalAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.MigrateLiquidity(&_IBancorNetwork.TransactOpts, token, provider, amount, availableAmount, originalAmount)
}

// MigrateLiquidity is a paid mutator transaction binding the contract method 0x3d1c24e7.
//
// Solidity: function migrateLiquidity(address token, address provider, uint256 amount, uint256 availableAmount, uint256 originalAmount) payable returns()
func (_IBancorNetwork *IBancorNetworkTransactorSession) MigrateLiquidity(token common.Address, provider common.Address, amount *big.Int, availableAmount *big.Int, originalAmount *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.MigrateLiquidity(&_IBancorNetwork.TransactOpts, token, provider, amount, availableAmount, originalAmount)
}

// MigratePools is a paid mutator transaction binding the contract method 0xc109ba13.
//
// Solidity: function migratePools(address[] pools, address newPoolCollection) returns()
func (_IBancorNetwork *IBancorNetworkTransactor) MigratePools(opts *bind.TransactOpts, pools []common.Address, newPoolCollection common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "migratePools", pools, newPoolCollection)
}

// MigratePools is a paid mutator transaction binding the contract method 0xc109ba13.
//
// Solidity: function migratePools(address[] pools, address newPoolCollection) returns()
func (_IBancorNetwork *IBancorNetworkSession) MigratePools(pools []common.Address, newPoolCollection common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.MigratePools(&_IBancorNetwork.TransactOpts, pools, newPoolCollection)
}

// MigratePools is a paid mutator transaction binding the contract method 0xc109ba13.
//
// Solidity: function migratePools(address[] pools, address newPoolCollection) returns()
func (_IBancorNetwork *IBancorNetworkTransactorSession) MigratePools(pools []common.Address, newPoolCollection common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.MigratePools(&_IBancorNetwork.TransactOpts, pools, newPoolCollection)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_IBancorNetwork *IBancorNetworkTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_IBancorNetwork *IBancorNetworkSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.RenounceRole(&_IBancorNetwork.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_IBancorNetwork *IBancorNetworkTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.RenounceRole(&_IBancorNetwork.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_IBancorNetwork *IBancorNetworkTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_IBancorNetwork *IBancorNetworkSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.RevokeRole(&_IBancorNetwork.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_IBancorNetwork *IBancorNetworkTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.RevokeRole(&_IBancorNetwork.TransactOpts, role, account)
}

// TradeBySourceAmount is a paid mutator transaction binding the contract method 0xd3a4acd3.
//
// Solidity: function tradeBySourceAmount(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) TradeBySourceAmount(opts *bind.TransactOpts, sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "tradeBySourceAmount", sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeBySourceAmount is a paid mutator transaction binding the contract method 0xd3a4acd3.
//
// Solidity: function tradeBySourceAmount(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) TradeBySourceAmount(sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.TradeBySourceAmount(&_IBancorNetwork.TransactOpts, sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeBySourceAmount is a paid mutator transaction binding the contract method 0xd3a4acd3.
//
// Solidity: function tradeBySourceAmount(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) TradeBySourceAmount(sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.TradeBySourceAmount(&_IBancorNetwork.TransactOpts, sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeBySourceAmountArb is a paid mutator transaction binding the contract method 0xd895feee.
//
// Solidity: function tradeBySourceAmountArb(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) TradeBySourceAmountArb(opts *bind.TransactOpts, sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "tradeBySourceAmountArb", sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeBySourceAmountArb is a paid mutator transaction binding the contract method 0xd895feee.
//
// Solidity: function tradeBySourceAmountArb(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) TradeBySourceAmountArb(sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.TradeBySourceAmountArb(&_IBancorNetwork.TransactOpts, sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeBySourceAmountArb is a paid mutator transaction binding the contract method 0xd895feee.
//
// Solidity: function tradeBySourceAmountArb(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) TradeBySourceAmountArb(sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.TradeBySourceAmountArb(&_IBancorNetwork.TransactOpts, sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeByTargetAmount is a paid mutator transaction binding the contract method 0x45d6602c.
//
// Solidity: function tradeByTargetAmount(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) TradeByTargetAmount(opts *bind.TransactOpts, sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "tradeByTargetAmount", sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// TradeByTargetAmount is a paid mutator transaction binding the contract method 0x45d6602c.
//
// Solidity: function tradeByTargetAmount(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) TradeByTargetAmount(sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.TradeByTargetAmount(&_IBancorNetwork.TransactOpts, sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// TradeByTargetAmount is a paid mutator transaction binding the contract method 0x45d6602c.
//
// Solidity: function tradeByTargetAmount(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) TradeByTargetAmount(sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.TradeByTargetAmount(&_IBancorNetwork.TransactOpts, sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// TradeByTargetAmountArb is a paid mutator transaction binding the contract method 0xd0d14581.
//
// Solidity: function tradeByTargetAmountArb(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) TradeByTargetAmountArb(opts *bind.TransactOpts, sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "tradeByTargetAmountArb", sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// TradeByTargetAmountArb is a paid mutator transaction binding the contract method 0xd0d14581.
//
// Solidity: function tradeByTargetAmountArb(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) TradeByTargetAmountArb(sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.TradeByTargetAmountArb(&_IBancorNetwork.TransactOpts, sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// TradeByTargetAmountArb is a paid mutator transaction binding the contract method 0xd0d14581.
//
// Solidity: function tradeByTargetAmountArb(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) TradeByTargetAmountArb(sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.TradeByTargetAmountArb(&_IBancorNetwork.TransactOpts, sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 id) returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) Withdraw(opts *bind.TransactOpts, id *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "withdraw", id)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 id) returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) Withdraw(id *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.Withdraw(&_IBancorNetwork.TransactOpts, id)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 id) returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) Withdraw(id *big.Int) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.Withdraw(&_IBancorNetwork.TransactOpts, id)
}

// WithdrawNetworkFees is a paid mutator transaction binding the contract method 0x3cd11924.
//
// Solidity: function withdrawNetworkFees(address recipient) returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) WithdrawNetworkFees(opts *bind.TransactOpts, recipient common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "withdrawNetworkFees", recipient)
}

// WithdrawNetworkFees is a paid mutator transaction binding the contract method 0x3cd11924.
//
// Solidity: function withdrawNetworkFees(address recipient) returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) WithdrawNetworkFees(recipient common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.WithdrawNetworkFees(&_IBancorNetwork.TransactOpts, recipient)
}

// WithdrawNetworkFees is a paid mutator transaction binding the contract method 0x3cd11924.
//
// Solidity: function withdrawNetworkFees(address recipient) returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) WithdrawNetworkFees(recipient common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.WithdrawNetworkFees(&_IBancorNetwork.TransactOpts, recipient)
}

// WithdrawPOL is a paid mutator transaction binding the contract method 0x8ffcca07.
//
// Solidity: function withdrawPOL(address pool) returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactor) WithdrawPOL(opts *bind.TransactOpts, pool common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.contract.Transact(opts, "withdrawPOL", pool)
}

// WithdrawPOL is a paid mutator transaction binding the contract method 0x8ffcca07.
//
// Solidity: function withdrawPOL(address pool) returns(uint256)
func (_IBancorNetwork *IBancorNetworkSession) WithdrawPOL(pool common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.WithdrawPOL(&_IBancorNetwork.TransactOpts, pool)
}

// WithdrawPOL is a paid mutator transaction binding the contract method 0x8ffcca07.
//
// Solidity: function withdrawPOL(address pool) returns(uint256)
func (_IBancorNetwork *IBancorNetworkTransactorSession) WithdrawPOL(pool common.Address) (*types.Transaction, error) {
	return _IBancorNetwork.Contract.WithdrawPOL(&_IBancorNetwork.TransactOpts, pool)
}

// IBancorNetworkRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the IBancorNetwork contract.
type IBancorNetworkRoleAdminChangedIterator struct {
	Event *IBancorNetworkRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *IBancorNetworkRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IBancorNetworkRoleAdminChanged)
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
		it.Event = new(IBancorNetworkRoleAdminChanged)
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
func (it *IBancorNetworkRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IBancorNetworkRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IBancorNetworkRoleAdminChanged represents a RoleAdminChanged event raised by the IBancorNetwork contract.
type IBancorNetworkRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_IBancorNetwork *IBancorNetworkFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*IBancorNetworkRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _IBancorNetwork.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &IBancorNetworkRoleAdminChangedIterator{contract: _IBancorNetwork.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_IBancorNetwork *IBancorNetworkFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *IBancorNetworkRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _IBancorNetwork.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IBancorNetworkRoleAdminChanged)
				if err := _IBancorNetwork.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_IBancorNetwork *IBancorNetworkFilterer) ParseRoleAdminChanged(log types.Log) (*IBancorNetworkRoleAdminChanged, error) {
	event := new(IBancorNetworkRoleAdminChanged)
	if err := _IBancorNetwork.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IBancorNetworkRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the IBancorNetwork contract.
type IBancorNetworkRoleGrantedIterator struct {
	Event *IBancorNetworkRoleGranted // Event containing the contract specifics and raw log

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
func (it *IBancorNetworkRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IBancorNetworkRoleGranted)
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
		it.Event = new(IBancorNetworkRoleGranted)
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
func (it *IBancorNetworkRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IBancorNetworkRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IBancorNetworkRoleGranted represents a RoleGranted event raised by the IBancorNetwork contract.
type IBancorNetworkRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_IBancorNetwork *IBancorNetworkFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*IBancorNetworkRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _IBancorNetwork.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &IBancorNetworkRoleGrantedIterator{contract: _IBancorNetwork.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_IBancorNetwork *IBancorNetworkFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *IBancorNetworkRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _IBancorNetwork.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IBancorNetworkRoleGranted)
				if err := _IBancorNetwork.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_IBancorNetwork *IBancorNetworkFilterer) ParseRoleGranted(log types.Log) (*IBancorNetworkRoleGranted, error) {
	event := new(IBancorNetworkRoleGranted)
	if err := _IBancorNetwork.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IBancorNetworkRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the IBancorNetwork contract.
type IBancorNetworkRoleRevokedIterator struct {
	Event *IBancorNetworkRoleRevoked // Event containing the contract specifics and raw log

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
func (it *IBancorNetworkRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IBancorNetworkRoleRevoked)
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
		it.Event = new(IBancorNetworkRoleRevoked)
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
func (it *IBancorNetworkRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IBancorNetworkRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IBancorNetworkRoleRevoked represents a RoleRevoked event raised by the IBancorNetwork contract.
type IBancorNetworkRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_IBancorNetwork *IBancorNetworkFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*IBancorNetworkRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _IBancorNetwork.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &IBancorNetworkRoleRevokedIterator{contract: _IBancorNetwork.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_IBancorNetwork *IBancorNetworkFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *IBancorNetworkRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _IBancorNetwork.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IBancorNetworkRoleRevoked)
				if err := _IBancorNetwork.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_IBancorNetwork *IBancorNetworkFilterer) ParseRoleRevoked(log types.Log) (*IBancorNetworkRoleRevoked, error) {
	event := new(IBancorNetworkRoleRevoked)
	if err := _IBancorNetwork.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
