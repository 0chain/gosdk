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
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"authorizerCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"authorizers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isAuthorizer\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"message\",\"type\":\"bytes32\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"authorize\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"_clientId\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"messageHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAuthorizer\",\"type\":\"address\"}],\"name\":\"addAuthorizers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_authorizer\",\"type\":\"address\"}],\"name\":\"removeAuthorizers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052600060025534801561001557600080fd5b5061001f33610024565b610074565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b611004806100836000396000f3fe608060405234801561001057600080fd5b506004361061009e5760003560e01c80638da5cb5b116100665780638da5cb5b1461012e578063c85501bb14610149578063e304688f14610151578063f2fde38b14610174578063f36bf4011461018757600080fd5b806309c7a20f146100a3578063207b601e146100e757806343ab2c9e14610108578063715018a61461011d5780637ac3d68d14610125575b600080fd5b6100cd6100b1366004610cb1565b6001602081905260009182526040909120805491015460ff1682565b604080519283529015156020830152015b60405180910390f35b6100fa6100f5366004610d15565b61019a565b6040519081526020016100de565b61011b610116366004610cb1565b6101e3565b005b61011b6103cd565b6100fa60025481565b6000546040516001600160a01b0390911681526020016100de565b6100fa6103e1565b61016461015f366004610da8565b610449565b60405190151581526020016100de565b61011b610182366004610cb1565b6106ed565b61011b610195366004610cb1565b610763565b60006101d7888888888888886040516020016101bc9796959493929190610e27565b6040516020818303038152906040528051906020012061088c565b98975050505050505050565b6101eb6108df565b6001600160a01b03811661023d5760405162461bcd60e51b8152602060048201526014602482015273696e76616c6964207a65726f206164647265737360601b60448201526064015b60405180910390fd5b6001600160a01b0381166000908152600160208190526040909120015460ff16156102aa5760405162461bcd60e51b815260206004820152601d60248201527f4164647265737320697320616c726561647920617574686f72697a65720000006044820152606401610234565b6003541561036b576040518060400160405280600360016003805490506102d19190610e8a565b815481106102e1576102e1610ea1565b600091825260208083209190910154835260019281018390526001600160a01b03851682528281526040909120835181559201519101805460ff1916911515919091179055600380548061033757610337610eb7565b600190038181906000526020600020016000905590556001600260008282546103609190610ecd565b909155506103ca9050565b604080518082018252600280548252600160208084018281526001600160a01b038716600090815291839052948120935184559351928101805460ff1916931515939093179092558054919290916103c4908490610ecd565b90915550505b50565b6103d56108df565b6103df6000610939565b565b6000600360025410156103f5575060025490565b600060036002546104069190610efb565b9050600060036002546104199190610f0f565b9050801561043e578061042d836002610f23565b6104379190610ecd565b9250505090565b610437826002610f23565b600080600254116104915760405162461bcd60e51b8152602060048201526012602482015271139bc8185d5d1a1bdc9a5e995c9cc81e595d60721b6044820152606401610234565b6104996103e1565b8210156104dc5760405162461bcd60e51b815260206004820152601160248201527053696720636f756e7420746f6f206c6f7760781b6044820152606401610234565b600060025467ffffffffffffffff8111156104f9576104f9610f42565b604051908082528060200260200182016040528015610522578160200160208202803683370190505b50905060005b838110156106e157600061059f86868481811061054757610547610ea1565b90506020028101906105599190610f58565b8080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061059992508b915061088c9050565b90610989565b6001600160a01b0381166000908152600160208190526040909120015490915060ff1661060e5760405162461bcd60e51b815260206004820152601960248201527f4d657373616765206973206e6f7420617574686f72697a6564000000000000006044820152606401610234565b6001600160a01b0381166000908152600160205260409020548351849190811061063a5761063a610ea1565b60200260200101511561068f5760405162461bcd60e51b815260206004820152601d60248201527f4475706c69636174656420617574686f72697a657220697320757365640000006044820152606401610234565b6001600160a01b038116600090815260016020819052604090912054845185919081106106be576106be610ea1565b9115156020928302919091019091015250806106d981610f9f565b915050610528565b50600195945050505050565b6106f56108df565b6001600160a01b03811661075a5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b6064820152608401610234565b6103ca81610939565b61076b6108df565b6001600160a01b0381166107b85760405162461bcd60e51b8152602060048201526014602482015273696e76616c6964207a65726f206164647265737360601b6044820152606401610234565b6001600160a01b0381166000908152600160208190526040909120015460ff166108245760405162461bcd60e51b815260206004820152601a60248201527f41646472657373206e6f7420616e20417574686f72697a6572730000000000006044820152606401610234565b6001600160a01b03811660009081526001602081905260408220808201805460ff19169055546003805480840182559084527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b015560028054919290916103c4908490610e8a565b6040517f19457468657265756d205369676e6564204d6573736167653a0a3332000000006020820152603c8101829052600090605c01604051602081830303815290604052805190602001209050919050565b6000546001600160a01b031633146103df5760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e65726044820152606401610234565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b600080600061099885856109ad565b915091506109a5816109f2565b509392505050565b60008082516041036109e35760208301516040840151606085015160001a6109d787828585610ba8565b945094505050506109eb565b506000905060025b9250929050565b6000816004811115610a0657610a06610fb8565b03610a0e5750565b6001816004811115610a2257610a22610fb8565b03610a6f5760405162461bcd60e51b815260206004820152601860248201527f45434453413a20696e76616c6964207369676e617475726500000000000000006044820152606401610234565b6002816004811115610a8357610a83610fb8565b03610ad05760405162461bcd60e51b815260206004820152601f60248201527f45434453413a20696e76616c6964207369676e6174757265206c656e677468006044820152606401610234565b6003816004811115610ae457610ae4610fb8565b03610b3c5760405162461bcd60e51b815260206004820152602260248201527f45434453413a20696e76616c6964207369676e6174757265202773272076616c604482015261756560f01b6064820152608401610234565b6004816004811115610b5057610b50610fb8565b036103ca5760405162461bcd60e51b815260206004820152602260248201527f45434453413a20696e76616c6964207369676e6174757265202776272076616c604482015261756560f01b6064820152608401610234565b6000807f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0831115610bdf5750600090506003610c8c565b8460ff16601b14158015610bf757508460ff16601c14155b15610c085750600090506004610c8c565b6040805160008082526020820180845289905260ff881692820192909252606081018690526080810185905260019060a0016020604051602081039080840390855afa158015610c5c573d6000803e3d6000fd5b5050604051601f1901519150506001600160a01b038116610c8557600060019250925050610c8c565b9150600090505b94509492505050565b80356001600160a01b0381168114610cac57600080fd5b919050565b600060208284031215610cc357600080fd5b610ccc82610c95565b9392505050565b60008083601f840112610ce557600080fd5b50813567ffffffffffffffff811115610cfd57600080fd5b6020830191508360208285010111156109eb57600080fd5b600080600080600080600060a0888a031215610d3057600080fd5b610d3988610c95565b965060208801359550604088013567ffffffffffffffff80821115610d5d57600080fd5b610d698b838c01610cd3565b909750955060608a0135915080821115610d8257600080fd5b50610d8f8a828b01610cd3565b989b979a50959894979596608090950135949350505050565b600080600060408486031215610dbd57600080fd5b83359250602084013567ffffffffffffffff80821115610ddc57600080fd5b818601915086601f830112610df057600080fd5b813581811115610dff57600080fd5b8760208260051b8501011115610e1457600080fd5b6020830194508093505050509250925092565b6bffffffffffffffffffffffff198860601b168152866014820152848660348301376000858201603481016000815285878237506034940193840192909252505060540195945050505050565b634e487b7160e01b600052601160045260246000fd5b600082821015610e9c57610e9c610e74565b500390565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052603160045260246000fd5b60008219821115610ee057610ee0610e74565b500190565b634e487b7160e01b600052601260045260246000fd5b600082610f0a57610f0a610ee5565b500490565b600082610f1e57610f1e610ee5565b500690565b6000816000190483118215151615610f3d57610f3d610e74565b500290565b634e487b7160e01b600052604160045260246000fd5b6000808335601e19843603018112610f6f57600080fd5b83018035915067ffffffffffffffff821115610f8a57600080fd5b6020019150368190038213156109eb57600080fd5b600060018201610fb157610fb1610e74565b5060010190565b634e487b7160e01b600052602160045260246000fdfea264697066735822122045e37bb81309c5ba4df43cb05552f34f5bc44d013d13971d32584a85b75c83b064736f6c634300080f0033",
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
// Solidity: function authorize(bytes32 message, bytes[] signatures) view returns(bool)
func (_Authorizers *AuthorizersCaller) Authorize(opts *bind.CallOpts, message [32]byte, signatures [][]byte) (bool, error) {
	var out []interface{}
	err := _Authorizers.contract.Call(opts, &out, "authorize", message, signatures)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Authorize is a free data retrieval call binding the contract method 0xe304688f.
//
// Solidity: function authorize(bytes32 message, bytes[] signatures) view returns(bool)
func (_Authorizers *AuthorizersSession) Authorize(message [32]byte, signatures [][]byte) (bool, error) {
	return _Authorizers.Contract.Authorize(&_Authorizers.CallOpts, message, signatures)
}

// Authorize is a free data retrieval call binding the contract method 0xe304688f.
//
// Solidity: function authorize(bytes32 message, bytes[] signatures) view returns(bool)
func (_Authorizers *AuthorizersCallerSession) Authorize(message [32]byte, signatures [][]byte) (bool, error) {
	return _Authorizers.Contract.Authorize(&_Authorizers.CallOpts, message, signatures)
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

// MessageHash is a free data retrieval call binding the contract method 0x207b601e.
//
// Solidity: function messageHash(address _to, uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce) pure returns(bytes32)
func (_Authorizers *AuthorizersCaller) MessageHash(opts *bind.CallOpts, _to common.Address, _amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Authorizers.contract.Call(opts, &out, "messageHash", _to, _amount, _txid, _clientId, _nonce)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MessageHash is a free data retrieval call binding the contract method 0x207b601e.
//
// Solidity: function messageHash(address _to, uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce) pure returns(bytes32)
func (_Authorizers *AuthorizersSession) MessageHash(_to common.Address, _amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int) ([32]byte, error) {
	return _Authorizers.Contract.MessageHash(&_Authorizers.CallOpts, _to, _amount, _txid, _clientId, _nonce)
}

// MessageHash is a free data retrieval call binding the contract method 0x207b601e.
//
// Solidity: function messageHash(address _to, uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce) pure returns(bytes32)
func (_Authorizers *AuthorizersCallerSession) MessageHash(_to common.Address, _amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int) ([32]byte, error) {
	return _Authorizers.Contract.MessageHash(&_Authorizers.CallOpts, _to, _amount, _txid, _clientId, _nonce)
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
// Solidity: function addAuthorizers(address newAuthorizer) returns()
func (_Authorizers *AuthorizersTransactor) AddAuthorizers(opts *bind.TransactOpts, newAuthorizer common.Address) (*types.Transaction, error) {
	return _Authorizers.contract.Transact(opts, "addAuthorizers", newAuthorizer)
}

// AddAuthorizers is a paid mutator transaction binding the contract method 0x43ab2c9e.
//
// Solidity: function addAuthorizers(address newAuthorizer) returns()
func (_Authorizers *AuthorizersSession) AddAuthorizers(newAuthorizer common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.AddAuthorizers(&_Authorizers.TransactOpts, newAuthorizer)
}

// AddAuthorizers is a paid mutator transaction binding the contract method 0x43ab2c9e.
//
// Solidity: function addAuthorizers(address newAuthorizer) returns()
func (_Authorizers *AuthorizersTransactorSession) AddAuthorizers(newAuthorizer common.Address) (*types.Transaction, error) {
	return _Authorizers.Contract.AddAuthorizers(&_Authorizers.TransactOpts, newAuthorizer)
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
