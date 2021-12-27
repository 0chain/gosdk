// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package authorizers

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

// AuthorizersMetaData contains all meta data concerning the Authorizers contract.
var AuthorizersMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_new_authorizer\",\"type\":\"address\"}],\"name\":\"addAuthorizers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"message\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"name\":\"authorize\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"authorizerCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"authorizers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isAuthorizer\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"message\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"message\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"recoverSigner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_authorizer\",\"type\":\"address\"}],\"name\":\"removeAuthorizers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"43ab2c9e": "addAuthorizers(address)",
		"296bb816": "authorize(bytes32,bytes)",
		"7ac3d68d": "authorizerCount()",
		"09c7a20f": "authorizers(address)",
		"f4fd62e3": "message(address,uint256,bytes,uint256)",
		"c85501bb": "minThreshold()",
		"8da5cb5b": "owner()",
		"97aba7f9": "recoverSigner(bytes32,bytes)",
		"f36bf401": "removeAuthorizers(address)",
		"715018a6": "renounceOwnership()",
		"f2fde38b": "transferOwnership(address)",
	},
	Bin: "0x6080604052600060025534801561001557600080fd5b50600061002061006f565b600080546001600160a01b0319166001600160a01b0383169081178255604051929350917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908290a350610073565b3390565b610cf7806100826000396000f3fe608060405234801561001057600080fd5b50600436106100a95760003560e01c80638da5cb5b116100715780638da5cb5b146101c257806397aba7f9146101e6578063c85501bb1461025d578063f2fde38b14610265578063f36bf4011461028b578063f4fd62e3146102b1576100a9565b806309c7a20f146100ae578063296bb816146100ed57806343ab2c9e14610178578063715018a6146101a05780637ac3d68d146101a8575b600080fd5b6100d4600480360360208110156100c457600080fd5b50356001600160a01b0316610336565b6040805192835290151560208301528051918290030190f35b6101646004803603604081101561010357600080fd5b8135919081019060408101602082013564010000000081111561012557600080fd5b82018360208201111561013757600080fd5b8035906020019184600183028401116401000000008311171561015957600080fd5b509092509050610352565b604080519115158252519081900360200190f35b61019e6004803603602081101561018e57600080fd5b50356001600160a01b03166105a7565b005b61019e61076c565b6101b0610818565b60408051918252519081900360200190f35b6101ca61081e565b604080516001600160a01b039092168252519081900360200190f35b6101ca600480360360408110156101fc57600080fd5b8135919081019060408101602082013564010000000081111561021e57600080fd5b82018360208201111561023057600080fd5b8035906020019184600183028401116401000000008311171561025257600080fd5b50909250905061082e565b6101b06108f7565b61019e6004803603602081101561027b57600080fd5b50356001600160a01b031661094d565b61019e600480360360208110156102a157600080fd5b50356001600160a01b0316610a4f565b6101b0600480360360808110156102c757600080fd5b6001600160a01b03823516916020810135918101906060810160408201356401000000008111156102f757600080fd5b82018360208201111561030957600080fd5b8035906020019184600183028401116401000000008311171561032b57600080fd5b919350915035610b82565b6001602081905260009182526040909120805491015460ff1682565b600060418206156103a3576040805162461bcd60e51b815260206004820152601660248201527544617461206e6f742065787065637465642073697a6560501b604482015290519081900360640190fd5b604182046103af6108f7565b8110156103f7576040805162461bcd60e51b815260206004820152601160248201527053696720636f756e7420746f6f206c6f7760781b604482015290519081900360640190fd5b606060025467ffffffffffffffff8111801561041257600080fd5b5060405190808252806020026020018201604052801561043c578160200160208202803683370190505b50905060005b8281101561059a57604180820290810160006104698a61046484868c8e610c53565b61082e565b6001600160a01b0381166000908152600160208190526040909120015490915060ff166104d6576040805162461bcd60e51b815260206004820152601660248201527513595cdcd859d948139bdd08105d5d1a1bdc9a5e995960521b604482015290519081900360640190fd5b6001600160a01b038116600090815260016020526040902054855186919081106104fc57fe5b602002602001015115610556576040805162461bcd60e51b815260206004820152601960248201527f4475706c696361746520417574686f72697a6572205573656400000000000000604482015290519081900360640190fd5b6001600160a01b0381166000908152600160208190526040909120548651879190811061057f57fe5b91151560209283029190910190910152505050600101610442565b5060019695505050505050565b6105af610be5565b6001600160a01b03166105c061081e565b6001600160a01b031614610609576040805162461bcd60e51b81526020600482018190526024820152600080516020610ca2833981519152604482015290519081900360640190fd5b6001600160a01b0381166000908152600160208190526040909120015460ff161561067b576040805162461bcd60e51b815260206004820152601d60248201527f4164647265737320697320416c726561647920417574686f72697a6572000000604482015290519081900360640190fd5b600354156107195760405180604001604052806003600160038054905003815481106106a357fe5b600091825260208083209190910154835260019281018390526001600160a01b03851682528281526040909120835181559201519101805460ff191691151591909117905560038054806106f357fe5b600082815260208120820160001990810191909155019055600280546001019055610769565b604080518082018252600280548252600160208084018281526001600160a01b038716600090815291839052949020925183559251918301805460ff191692151592909217909155805490910190555b50565b610774610be5565b6001600160a01b031661078561081e565b6001600160a01b0316146107ce576040805162461bcd60e51b81526020600482018190526024820152600080516020610ca2833981519152604482015290519081900360640190fd5b600080546040516001600160a01b03909116907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908390a3600080546001600160a01b0319169055565b60025481565b6000546001600160a01b03165b90565b60006041821461083d57600080fd5b600080600061088186868080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250610be992505050565b92509250925060018784848460405160008152602001604052604051808581526020018460ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa1580156108e1573d6000803e3d6000fd5b5050604051601f19015198975050505050505050565b60006003600254101561090d575060025461082b565b600060036002548161091b57fe5b049050600060036002548161092c57fe5b069050801561094257600290910201905061082b565b50600202905061082b565b610955610be5565b6001600160a01b031661096661081e565b6001600160a01b0316146109af576040805162461bcd60e51b81526020600482018190526024820152600080516020610ca2833981519152604482015290519081900360640190fd5b6001600160a01b0381166109f45760405162461bcd60e51b8152600401808060200182810382526026815260200180610c7c6026913960400191505060405180910390fd5b600080546040516001600160a01b03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a3600080546001600160a01b0319166001600160a01b0392909216919091179055565b610a57610be5565b6001600160a01b0316610a6861081e565b6001600160a01b031614610ab1576040805162461bcd60e51b81526020600482018190526024820152600080516020610ca2833981519152604482015290519081900360640190fd5b6001600160a01b0381166000908152600160208190526040909120015460ff16610b22576040805162461bcd60e51b815260206004820152601a60248201527f41646472657373206e6f7420616e20417574686f72697a657273000000000000604482015290519081900360640190fd5b6001600160a01b031660009081526001602081905260408220808201805460ff19169055546003805492830181559092527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b015560028054600019019055565b6000610bdb868686868660405160200180866001600160a01b031660601b8152601401858152602001848480828437919091019283525050604080518083038152602092830190915280519101209350610c0292505050565b9695505050505050565b3390565b6020810151604082015160609092015160001a92909190565b604080517f19457468657265756d205369676e6564204d6573736167653a0a333200000000602080830191909152603c8083019490945282518083039094018452605c909101909152815191012090565b60008085851115610c62578182fd5b83861115610c6e578182fd5b505082019391909203915056fe4f776e61626c653a206e6577206f776e657220697320746865207a65726f20616464726573734f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572a2646970667358221220a1626ea9fe51beed4aca526f7671a9046517990553891e345da203d026d3ecd164736f6c63430007050033",
}

// AuthorizersABI is the input ABI used to generate the binding from.
// Deprecated: Use AuthorizersMetaData.ABI instead.
var AuthorizersABI = AuthorizersMetaData.ABI

// Deprecated: Use AuthorizersMetaData.Sigs instead.
// AuthorizersFuncSigs maps the 4-byte function signature to its string representation.
var AuthorizersFuncSigs = AuthorizersMetaData.Sigs

// AuthorizersBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use AuthorizersMetaData.Bin instead.
var AuthorizersBin = AuthorizersMetaData.Bin

// DeployAuthorizers deploys a new Ethereum contract, binding an instance of Authorizers to it.
func DeployAuthorizers(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Authorizers, error) {
	parsed, err := AuthorizersMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(AuthorizersBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Authorizers{AuthorizersCaller: AuthorizersCaller{contract: contract}, AuthorizersTransactor: AuthorizersTransactor{contract: contract}, AuthorizersFilterer: AuthorizersFilterer{contract: contract}}, nil
}

// Authorizers is an auto generated Go binding around an Ethereum contract.
type Authorizers struct {
	AuthorizersCaller     // Read-only binding to the contract
	AuthorizersTransactor // Write-only binding to the contract
	AuthorizersFilterer   // Log filterer for contract events
}

// AuthorizersCaller is an auto generated read-only Go binding around an Ethereum contract.
type AuthorizersCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuthorizersTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AuthorizersTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuthorizersFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AuthorizersFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuthorizersSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AuthorizersSession struct {
	Contract     *Authorizers      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AuthorizersCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AuthorizersCallerSession struct {
	Contract *AuthorizersCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// AuthorizersTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AuthorizersTransactorSession struct {
	Contract     *AuthorizersTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// AuthorizersRaw is an auto generated low-level Go binding around an Ethereum contract.
type AuthorizersRaw struct {
	Contract *Authorizers // Generic contract binding to access the raw methods on
}

// AuthorizersCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AuthorizersCallerRaw struct {
	Contract *AuthorizersCaller // Generic read-only contract binding to access the raw methods on
}

// AuthorizersTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AuthorizersTransactorRaw struct {
	Contract *AuthorizersTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAuthorizers creates a new instance of Authorizers, bound to a specific deployed contract.
func NewAuthorizers(address common.Address, backend bind.ContractBackend) (*Authorizers, error) {
	contract, err := bindAuthorizers(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Authorizers{AuthorizersCaller: AuthorizersCaller{contract: contract}, AuthorizersTransactor: AuthorizersTransactor{contract: contract}, AuthorizersFilterer: AuthorizersFilterer{contract: contract}}, nil
}

// NewAuthorizersCaller creates a new read-only instance of Authorizers, bound to a specific deployed contract.
func NewAuthorizersCaller(address common.Address, caller bind.ContractCaller) (*AuthorizersCaller, error) {
	contract, err := bindAuthorizers(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AuthorizersCaller{contract: contract}, nil
}

// NewAuthorizersTransactor creates a new write-only instance of Authorizers, bound to a specific deployed contract.
func NewAuthorizersTransactor(address common.Address, transactor bind.ContractTransactor) (*AuthorizersTransactor, error) {
	contract, err := bindAuthorizers(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AuthorizersTransactor{contract: contract}, nil
}

// NewAuthorizersFilterer creates a new log filterer instance of Authorizers, bound to a specific deployed contract.
func NewAuthorizersFilterer(address common.Address, filterer bind.ContractFilterer) (*AuthorizersFilterer, error) {
	contract, err := bindAuthorizers(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AuthorizersFilterer{contract: contract}, nil
}

// bindAuthorizers binds a generic wrapper to an already deployed contract.
func bindAuthorizers(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AuthorizersABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Authorizers *AuthorizersRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Authorizers.Contract.AuthorizersCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Authorizers *AuthorizersRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Authorizers.Contract.AuthorizersTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Authorizers *AuthorizersRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Authorizers.Contract.AuthorizersTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Authorizers *AuthorizersCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Authorizers.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Authorizers *AuthorizersTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Authorizers.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Authorizers *AuthorizersTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Authorizers.Contract.contract.Transact(opts, method, params...)
}

// AuthorizerCount is a free data retrieval call binding the contract method 0x7ac3d68d.
//
// Solidity: function authorizerCount() view returns(uint256)
func (_Authorizers *AuthorizersCaller) AuthorizerCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Authorizers.contract.Call(opts, &out, "authorizerCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AuthorizerCount is a free data retrieval call binding the contract method 0x7ac3d68d.
//
// Solidity: function authorizerCount() view returns(uint256)
func (_Authorizers *AuthorizersSession) AuthorizerCount() (*big.Int, error) {
	return _Authorizers.Contract.AuthorizerCount(&_Authorizers.CallOpts)
}

// AuthorizerCount is a free data retrieval call binding the contract method 0x7ac3d68d.
//
// Solidity: function authorizerCount() view returns(uint256)
func (_Authorizers *AuthorizersCallerSession) AuthorizerCount() (*big.Int, error) {
	return _Authorizers.Contract.AuthorizerCount(&_Authorizers.CallOpts)
}

// Authorizers is a free data retrieval call binding the contract method 0x09c7a20f.
//
// Solidity: function authorizers(address ) view returns(uint256 index, bool isAuthorizer)
func (_Authorizers *AuthorizersCaller) Authorizers(opts *bind.CallOpts, arg0 common.Address) (struct {
	Index        *big.Int
	IsAuthorizer bool
}, error) {
	var out []interface{}
	err := _Authorizers.contract.Call(opts, &out, "authorizers", arg0)

	outstruct := new(struct {
		Index        *big.Int
		IsAuthorizer bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Index = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.IsAuthorizer = *abi.ConvertType(out[1], new(bool)).(*bool)

	return *outstruct, err

}

// Authorizers is a free data retrieval call binding the contract method 0x09c7a20f.
//
// Solidity: function authorizers(address ) view returns(uint256 index, bool isAuthorizer)
func (_Authorizers *AuthorizersSession) Authorizers(arg0 common.Address) (struct {
	Index        *big.Int
	IsAuthorizer bool
}, error) {
	return _Authorizers.Contract.Authorizers(&_Authorizers.CallOpts, arg0)
}

// Authorizers is a free data retrieval call binding the contract method 0x09c7a20f.
//
// Solidity: function authorizers(address ) view returns(uint256 index, bool isAuthorizer)
func (_Authorizers *AuthorizersCallerSession) Authorizers(arg0 common.Address) (struct {
	Index        *big.Int
	IsAuthorizer bool
}, error) {
	return _Authorizers.Contract.Authorizers(&_Authorizers.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Authorizers *AuthorizersCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Authorizers.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Authorizers *AuthorizersSession) Owner() (common.Address, error) {
	return _Authorizers.Contract.Owner(&_Authorizers.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Authorizers *AuthorizersCallerSession) Owner() (common.Address, error) {
	return _Authorizers.Contract.Owner(&_Authorizers.CallOpts)
}

// AddAuthorizers is a paid mutator transaction binding the contract method 0x43ab2c9e.
//
// Solidity: function addAuthorizers(address _new_authorizer) returns()
func (_Authorizers *AuthorizersTransactor) AddAuthorizers(opts *bind.TransactOpts, _new_authorizer common.Address) (*types.Transaction, error) {
	return _Authorizers.contract.Transact(opts, "addAuthorizers", _new_authorizer)
}

// AddAuthorizers is a paid mutator transaction binding the contract method 0x43ab2c9e.
//
// Solidity: function addAuthorizers(address _new_authorizer) returns()
func (_Authorizers *AuthorizersSession) AddAuthorizers(_new_authorizer common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.AddAuthorizers(&_Authorizers.TransactOpts, _new_authorizer)
}

// AddAuthorizers is a paid mutator transaction binding the contract method 0x43ab2c9e.
//
// Solidity: function addAuthorizers(address _new_authorizer) returns()
func (_Authorizers *AuthorizersTransactorSession) AddAuthorizers(_new_authorizer common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.AddAuthorizers(&_Authorizers.TransactOpts, _new_authorizer)
}

// Authorize is a paid mutator transaction binding the contract method 0x296bb816.
//
// Solidity: function authorize(bytes32 message, bytes signatures) returns(bool)
func (_Authorizers *AuthorizersTransactor) Authorize(opts *bind.TransactOpts, message [32]byte, signatures []byte) (*types.Transaction, error) {
	return _Authorizers.contract.Transact(opts, "authorize", message, signatures)
}

// Authorize is a paid mutator transaction binding the contract method 0x296bb816.
//
// Solidity: function authorize(bytes32 message, bytes signatures) returns(bool)
func (_Authorizers *AuthorizersSession) Authorize(message [32]byte, signatures []byte) (*types.Transaction, error) {
	return _Authorizers.Contract.Authorize(&_Authorizers.TransactOpts, message, signatures)
}

// Authorize is a paid mutator transaction binding the contract method 0x296bb816.
//
// Solidity: function authorize(bytes32 message, bytes signatures) returns(bool)
func (_Authorizers *AuthorizersTransactorSession) Authorize(message [32]byte, signatures []byte) (*types.Transaction, error) {
	return _Authorizers.Contract.Authorize(&_Authorizers.TransactOpts, message, signatures)
}

// Message is a paid mutator transaction binding the contract method 0xf4fd62e3.
//
// Solidity: function message(address _to, uint256 _amount, bytes _txid, uint256 _nonce) returns(bytes32)
func (_Authorizers *AuthorizersTransactor) Message(opts *bind.TransactOpts, _to common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int) (*types.Transaction, error) {
	return _Authorizers.contract.Transact(opts, "message", _to, _amount, _txid, _nonce)
}

// Message is a paid mutator transaction binding the contract method 0xf4fd62e3.
//
// Solidity: function message(address _to, uint256 _amount, bytes _txid, uint256 _nonce) returns(bytes32)
func (_Authorizers *AuthorizersSession) Message(_to common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int) (*types.Transaction, error) {
	return _Authorizers.Contract.Message(&_Authorizers.TransactOpts, _to, _amount, _txid, _nonce)
}

// Message is a paid mutator transaction binding the contract method 0xf4fd62e3.
//
// Solidity: function message(address _to, uint256 _amount, bytes _txid, uint256 _nonce) returns(bytes32)
func (_Authorizers *AuthorizersTransactorSession) Message(_to common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int) (*types.Transaction, error) {
	return _Authorizers.Contract.Message(&_Authorizers.TransactOpts, _to, _amount, _txid, _nonce)
}

// MinThreshold is a paid mutator transaction binding the contract method 0xc85501bb.
//
// Solidity: function minThreshold() returns(uint256)
func (_Authorizers *AuthorizersTransactor) MinThreshold(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Authorizers.contract.Transact(opts, "minThreshold")
}

// MinThreshold is a paid mutator transaction binding the contract method 0xc85501bb.
//
// Solidity: function minThreshold() returns(uint256)
func (_Authorizers *AuthorizersSession) MinThreshold() (*types.Transaction, error) {
	return _Authorizers.Contract.MinThreshold(&_Authorizers.TransactOpts)
}

// MinThreshold is a paid mutator transaction binding the contract method 0xc85501bb.
//
// Solidity: function minThreshold() returns(uint256)
func (_Authorizers *AuthorizersTransactorSession) MinThreshold() (*types.Transaction, error) {
	return _Authorizers.Contract.MinThreshold(&_Authorizers.TransactOpts)
}

// RecoverSigner is a paid mutator transaction binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 message, bytes signature) returns(address)
func (_Authorizers *AuthorizersTransactor) RecoverSigner(opts *bind.TransactOpts, message [32]byte, signature []byte) (*types.Transaction, error) {
	return _Authorizers.contract.Transact(opts, "recoverSigner", message, signature)
}

// RecoverSigner is a paid mutator transaction binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 message, bytes signature) returns(address)
func (_Authorizers *AuthorizersSession) RecoverSigner(message [32]byte, signature []byte) (*types.Transaction, error) {
	return _Authorizers.Contract.RecoverSigner(&_Authorizers.TransactOpts, message, signature)
}

// RecoverSigner is a paid mutator transaction binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 message, bytes signature) returns(address)
func (_Authorizers *AuthorizersTransactorSession) RecoverSigner(message [32]byte, signature []byte) (*types.Transaction, error) {
	return _Authorizers.Contract.RecoverSigner(&_Authorizers.TransactOpts, message, signature)
}

// RemoveAuthorizers is a paid mutator transaction binding the contract method 0xf36bf401.
//
// Solidity: function removeAuthorizers(address _authorizer) returns()
func (_Authorizers *AuthorizersTransactor) RemoveAuthorizers(opts *bind.TransactOpts, _authorizer common.Address) (*types.Transaction, error) {
	return _Authorizers.contract.Transact(opts, "removeAuthorizers", _authorizer)
}

// RemoveAuthorizers is a paid mutator transaction binding the contract method 0xf36bf401.
//
// Solidity: function removeAuthorizers(address _authorizer) returns()
func (_Authorizers *AuthorizersSession) RemoveAuthorizers(_authorizer common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.RemoveAuthorizers(&_Authorizers.TransactOpts, _authorizer)
}

// RemoveAuthorizers is a paid mutator transaction binding the contract method 0xf36bf401.
//
// Solidity: function removeAuthorizers(address _authorizer) returns()
func (_Authorizers *AuthorizersTransactorSession) RemoveAuthorizers(_authorizer common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.RemoveAuthorizers(&_Authorizers.TransactOpts, _authorizer)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Authorizers *AuthorizersTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Authorizers.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Authorizers *AuthorizersSession) RenounceOwnership() (*types.Transaction, error) {
	return _Authorizers.Contract.RenounceOwnership(&_Authorizers.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Authorizers *AuthorizersTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Authorizers.Contract.RenounceOwnership(&_Authorizers.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Authorizers *AuthorizersTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Authorizers.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Authorizers *AuthorizersSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.TransferOwnership(&_Authorizers.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Authorizers *AuthorizersTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.TransferOwnership(&_Authorizers.TransactOpts, newOwner)
}

// AuthorizersOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Authorizers contract.
type AuthorizersOwnershipTransferredIterator struct {
	Event *AuthorizersOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *AuthorizersOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuthorizersOwnershipTransferred)
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
		it.Event = new(AuthorizersOwnershipTransferred)
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
func (it *AuthorizersOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuthorizersOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuthorizersOwnershipTransferred represents a OwnershipTransferred event raised by the Authorizers contract.
type AuthorizersOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Authorizers *AuthorizersFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AuthorizersOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Authorizers.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AuthorizersOwnershipTransferredIterator{contract: _Authorizers.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Authorizers *AuthorizersFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AuthorizersOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Authorizers.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuthorizersOwnershipTransferred)
				if err := _Authorizers.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Authorizers *AuthorizersFilterer) ParseOwnershipTransferred(log types.Log) (*AuthorizersOwnershipTransferred, error) {
	event := new(AuthorizersOwnershipTransferred)
	if err := _Authorizers.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
