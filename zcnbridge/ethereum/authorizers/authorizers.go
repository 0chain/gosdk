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
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"authorizerCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"authorizers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isAuthorizer\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"message_\",\"type\":\"bytes32\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures_\",\"type\":\"bytes[]\"}],\"name\":\"authorize\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"txid_\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"nonce_\",\"type\":\"uint256\"}],\"name\":\"messageHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"authorizer_\",\"type\":\"address\"}],\"name\":\"addAuthorizers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"authorizer_\",\"type\":\"address\"}],\"name\":\"removeAuthorizers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052600060025534801561001557600080fd5b5061001f33610024565b610074565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b610f33806100836000396000f3fe608060405234801561001057600080fd5b506004361061009e5760003560e01c80638da5cb5b116100665780638da5cb5b1461012e578063c85501bb14610149578063e304688f14610151578063f2fde38b14610174578063f36bf4011461018757600080fd5b806309c7a20f146100a35780630b249ae4146100e757806343ab2c9e14610108578063715018a61461011d5780637ac3d68d14610125575b600080fd5b6100cd6100b1366004610c17565b6001602081905260009182526040909120805491015460ff1682565b604080519283529015156020830152015b60405180910390f35b6100fa6100f5366004610c39565b61019a565b6040519081526020016100de565b61011b610116366004610c17565b6101dd565b005b61011b6103bd565b6100fa60025481565b6000546040516001600160a01b0390911681526020016100de565b6100fa6103d1565b61016461015f366004610cc8565b61040a565b60405190151581526020016100de565b61011b610182366004610c17565b6106f6565b61011b610195366004610c17565b61076c565b60006101d386868686866040516020016101b8959493929190610d47565b60405160208183030381529060405280519060200120610887565b9695505050505050565b6101e56108da565b6001600160a01b0381166102145760405162461bcd60e51b815260040161020b90610d7b565b60405180910390fd5b6001600160a01b0381166000908152600160208190526040909120015460ff161561029a5760405162461bcd60e51b815260206004820152603060248201527f417574686f72697a657220776974682074686520676976656e2061646472657360448201526f7320616c72656164792065786973747360801b606482015260840161020b565b6003541561035b576040518060400160405280600360016003805490506102c19190610dd5565b815481106102d1576102d1610dec565b600091825260208083209190910154835260019281018390526001600160a01b03851682528281526040909120835181559201519101805460ff1916911515919091179055600380548061032757610327610e02565b600190038181906000526020600020016000905590556001600260008282546103509190610e18565b909155506103ba9050565b604080518082018252600280548252600160208084018281526001600160a01b038716600090815291839052948120935184559351928101805460ff1916931515939093179092558054919290916103b4908490610e18565b90915550505b50565b6103c56108da565b6103cf6000610934565b565b60006003600160025460026103e69190610e30565b6103f09190610dd5565b6103fa9190610e4f565b610405906001610e18565b905090565b6000806002541161046b5760405162461bcd60e51b815260206004820152602560248201527f4e6f20617574686f72697a657220686173206265656e2072656769737465726560448201526419081e595d60da1b606482015260840161020b565b6104736103d1565b8210156104cd5760405162461bcd60e51b815260206004820152602260248201527f416d6f756e74206f66207369676e617475726573206973206e6f7420656e6f756044820152610ced60f31b606482015260840161020b565b600060025467ffffffffffffffff8111156104ea576104ea610e71565b604051908082528060200260200182016040528015610513578160200160208202803683370190505b50905060005b838110156106ea57600061059086868481811061053857610538610dec565b905060200281019061054a9190610e87565b8080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061058a92508b91506108879050565b90610984565b6001600160a01b0381166000908152600160208190526040909120015490915060ff166105ff5760405162461bcd60e51b815260206004820152601f60248201527f476976656e206d657373616765206973206e6f7420617574686f72697a656400604482015260640161020b565b6001600160a01b0381166000908152600160205260409020548351849190811061062b5761062b610dec565b6020026020010151156106985760405162461bcd60e51b815260206004820152602f60248201527f476976656e207369676e61747572657320636f6e7461696e206475706c69636160448201526e74656420617574686f72697a65727360881b606482015260840161020b565b6001600160a01b038116600090815260016020819052604090912054845185919081106106c7576106c7610dec565b9115156020928302919091019091015250806106e281610ece565b915050610519565b50600195945050505050565b6106fe6108da565b6001600160a01b0381166107635760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b606482015260840161020b565b6103ba81610934565b6107746108da565b6001600160a01b03811661079a5760405162461bcd60e51b815260040161020b90610d7b565b6001600160a01b0381166000908152600160208190526040909120015460ff1661081f5760405162461bcd60e51b815260206004820152603060248201527f417574686f72697a657220776974682074686520676976656e2061646472657360448201526f1cc8191bd95cc81b9bdd08195e1a5cdd60821b606482015260840161020b565b6001600160a01b03811660009081526001602081905260408220808201805460ff19169055546003805480840182559084527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b015560028054919290916103b4908490610dd5565b6040517f19457468657265756d205369676e6564204d6573736167653a0a3332000000006020820152603c8101829052600090605c01604051602081830303815290604052805190602001209050919050565b6000546001600160a01b031633146103cf5760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015260640161020b565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b600080600061099385856109a8565b915091506109a0816109ed565b509392505050565b60008082516041036109de5760208301516040840151606085015160001a6109d287828585610b37565b945094505050506109e6565b506000905060025b9250929050565b6000816004811115610a0157610a01610ee7565b03610a095750565b6001816004811115610a1d57610a1d610ee7565b03610a6a5760405162461bcd60e51b815260206004820152601860248201527f45434453413a20696e76616c6964207369676e61747572650000000000000000604482015260640161020b565b6002816004811115610a7e57610a7e610ee7565b03610acb5760405162461bcd60e51b815260206004820152601f60248201527f45434453413a20696e76616c6964207369676e6174757265206c656e67746800604482015260640161020b565b6003816004811115610adf57610adf610ee7565b036103ba5760405162461bcd60e51b815260206004820152602260248201527f45434453413a20696e76616c6964207369676e6174757265202773272076616c604482015261756560f01b606482015260840161020b565b6000807f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0831115610b6e5750600090506003610bf2565b6040805160008082526020820180845289905260ff881692820192909252606081018690526080810185905260019060a0016020604051602081039080840390855afa158015610bc2573d6000803e3d6000fd5b5050604051601f1901519150506001600160a01b038116610beb57600060019250925050610bf2565b9150600090505b94509492505050565b80356001600160a01b0381168114610c1257600080fd5b919050565b600060208284031215610c2957600080fd5b610c3282610bfb565b9392505050565b600080600080600060808688031215610c5157600080fd5b610c5a86610bfb565b945060208601359350604086013567ffffffffffffffff80821115610c7e57600080fd5b818801915088601f830112610c9257600080fd5b813581811115610ca157600080fd5b896020828501011115610cb357600080fd5b96999598505060200195606001359392505050565b600080600060408486031215610cdd57600080fd5b83359250602084013567ffffffffffffffff80821115610cfc57600080fd5b818601915086601f830112610d1057600080fd5b813581811115610d1f57600080fd5b8760208260051b8501011115610d3457600080fd5b6020830194508093505050509250925092565b6bffffffffffffffffffffffff198660601b1681528460148201528284603483013760349201918201526054019392505050565b60208082526024908201527f496e76616c696420617574686f72697a6572206164647265737320776173206760408201526334bb32b760e11b606082015260800190565b634e487b7160e01b600052601160045260246000fd5b600082821015610de757610de7610dbf565b500390565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052603160045260246000fd5b60008219821115610e2b57610e2b610dbf565b500190565b6000816000190483118215151615610e4a57610e4a610dbf565b500290565b600082610e6c57634e487b7160e01b600052601260045260246000fd5b500490565b634e487b7160e01b600052604160045260246000fd5b6000808335601e19843603018112610e9e57600080fd5b83018035915067ffffffffffffffff821115610eb957600080fd5b6020019150368190038213156109e657600080fd5b600060018201610ee057610ee0610dbf565b5060010190565b634e487b7160e01b600052602160045260246000fdfea2646970667358221220d36512e87ac563a1b5880c9eb0c665cc8b3875c8c66d26ea6aa50ec191d4719364736f6c634300080f0033",
}

// AuthorizersABI is the input ABI used to generate the binding from.
// Deprecated: Use AuthorizersMetaData.ABI instead.
var AuthorizersABI = AuthorizersMetaData.ABI

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

// Authorize is a free data retrieval call binding the contract method 0xe304688f.
//
// Solidity: function authorize(bytes32 message_, bytes[] signatures_) view returns(bool)
func (_Authorizers *AuthorizersCaller) Authorize(opts *bind.CallOpts, message_ [32]byte, signatures_ [][]byte) (bool, error) {
	var out []interface{}
	err := _Authorizers.contract.Call(opts, &out, "authorize", message_, signatures_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Authorize is a free data retrieval call binding the contract method 0xe304688f.
//
// Solidity: function authorize(bytes32 message_, bytes[] signatures_) view returns(bool)
func (_Authorizers *AuthorizersSession) Authorize(message_ [32]byte, signatures_ [][]byte) (bool, error) {
	return _Authorizers.Contract.Authorize(&_Authorizers.CallOpts, message_, signatures_)
}

// Authorize is a free data retrieval call binding the contract method 0xe304688f.
//
// Solidity: function authorize(bytes32 message_, bytes[] signatures_) view returns(bool)
func (_Authorizers *AuthorizersCallerSession) Authorize(message_ [32]byte, signatures_ [][]byte) (bool, error) {
	return _Authorizers.Contract.Authorize(&_Authorizers.CallOpts, message_, signatures_)
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

// MessageHash is a free data retrieval call binding the contract method 0x0b249ae4.
//
// Solidity: function messageHash(address to_, uint256 amount_, bytes txid_, uint256 nonce_) pure returns(bytes32)
func (_Authorizers *AuthorizersCaller) MessageHash(opts *bind.CallOpts, to_ common.Address, amount_ *big.Int, txid_ []byte, nonce_ *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Authorizers.contract.Call(opts, &out, "messageHash", to_, amount_, txid_, nonce_)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MessageHash is a free data retrieval call binding the contract method 0x0b249ae4.
//
// Solidity: function messageHash(address to_, uint256 amount_, bytes txid_, uint256 nonce_) pure returns(bytes32)
func (_Authorizers *AuthorizersSession) MessageHash(to_ common.Address, amount_ *big.Int, txid_ []byte, nonce_ *big.Int) ([32]byte, error) {
	return _Authorizers.Contract.MessageHash(&_Authorizers.CallOpts, to_, amount_, txid_, nonce_)
}

// MessageHash is a free data retrieval call binding the contract method 0x0b249ae4.
//
// Solidity: function messageHash(address to_, uint256 amount_, bytes txid_, uint256 nonce_) pure returns(bytes32)
func (_Authorizers *AuthorizersCallerSession) MessageHash(to_ common.Address, amount_ *big.Int, txid_ []byte, nonce_ *big.Int) ([32]byte, error) {
	return _Authorizers.Contract.MessageHash(&_Authorizers.CallOpts, to_, amount_, txid_, nonce_)
}

// MinThreshold is a free data retrieval call binding the contract method 0xc85501bb.
//
// Solidity: function minThreshold() view returns(uint256)
func (_Authorizers *AuthorizersCaller) MinThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Authorizers.contract.Call(opts, &out, "minThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinThreshold is a free data retrieval call binding the contract method 0xc85501bb.
//
// Solidity: function minThreshold() view returns(uint256)
func (_Authorizers *AuthorizersSession) MinThreshold() (*big.Int, error) {
	return _Authorizers.Contract.MinThreshold(&_Authorizers.CallOpts)
}

// MinThreshold is a free data retrieval call binding the contract method 0xc85501bb.
//
// Solidity: function minThreshold() view returns(uint256)
func (_Authorizers *AuthorizersCallerSession) MinThreshold() (*big.Int, error) {
	return _Authorizers.Contract.MinThreshold(&_Authorizers.CallOpts)
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
// Solidity: function addAuthorizers(address authorizer_) returns()
func (_Authorizers *AuthorizersTransactor) AddAuthorizers(opts *bind.TransactOpts, authorizer_ common.Address) (*types.Transaction, error) {
	return _Authorizers.contract.Transact(opts, "addAuthorizers", authorizer_)
}

// AddAuthorizers is a paid mutator transaction binding the contract method 0x43ab2c9e.
//
// Solidity: function addAuthorizers(address authorizer_) returns()
func (_Authorizers *AuthorizersSession) AddAuthorizers(authorizer_ common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.AddAuthorizers(&_Authorizers.TransactOpts, authorizer_)
}

// AddAuthorizers is a paid mutator transaction binding the contract method 0x43ab2c9e.
//
// Solidity: function addAuthorizers(address authorizer_) returns()
func (_Authorizers *AuthorizersTransactorSession) AddAuthorizers(authorizer_ common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.AddAuthorizers(&_Authorizers.TransactOpts, authorizer_)
}

// RemoveAuthorizers is a paid mutator transaction binding the contract method 0xf36bf401.
//
// Solidity: function removeAuthorizers(address authorizer_) returns()
func (_Authorizers *AuthorizersTransactor) RemoveAuthorizers(opts *bind.TransactOpts, authorizer_ common.Address) (*types.Transaction, error) {
	return _Authorizers.contract.Transact(opts, "removeAuthorizers", authorizer_)
}

// RemoveAuthorizers is a paid mutator transaction binding the contract method 0xf36bf401.
//
// Solidity: function removeAuthorizers(address authorizer_) returns()
func (_Authorizers *AuthorizersSession) RemoveAuthorizers(authorizer_ common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.RemoveAuthorizers(&_Authorizers.TransactOpts, authorizer_)
}

// RemoveAuthorizers is a paid mutator transaction binding the contract method 0xf36bf401.
//
// Solidity: function removeAuthorizers(address authorizer_) returns()
func (_Authorizers *AuthorizersTransactorSession) RemoveAuthorizers(authorizer_ common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.RemoveAuthorizers(&_Authorizers.TransactOpts, authorizer_)
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
