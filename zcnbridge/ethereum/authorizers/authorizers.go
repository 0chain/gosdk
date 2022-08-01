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
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"authorizerCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"authorizers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isAuthorizer\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"message\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"recoverSigner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"message\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"name\":\"authorize\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"_clientId\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"messageHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAuthorizer\",\"type\":\"address\"}],\"name\":\"addAuthorizers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_authorizer\",\"type\":\"address\"}],\"name\":\"removeAuthorizers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052600060025534801561001557600080fd5b5061001f33610024565b610074565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b610cd0806100836000396000f3fe608060405234801561001057600080fd5b50600436106100a95760003560e01c80637ac3d68d116100715780637ac3d68d146101535780638da5cb5b1461015c57806397aba7f914610181578063c85501bb14610194578063f2fde38b1461019c578063f36bf401146101af57600080fd5b806309c7a20f146100ae578063207b601e146100f2578063296bb8161461011357806343ab2c9e14610136578063715018a61461014b575b600080fd5b6100d86100bc3660046109dc565b6001602081905260009182526040909120805491015460ff1682565b604080519283529015156020830152015b60405180910390f35b610105610100366004610a47565b6101c2565b6040519081526020016100e9565b610126610121366004610ada565b610250565b60405190151581526020016100e9565b6101496101443660046109dc565b6104d3565b005b61014961066b565b61010560025481565b6000546001600160a01b03165b6040516001600160a01b0390911681526020016100e9565b61016961018f366004610ada565b61067f565b610105610743565b6101496101aa3660046109dc565b6107ab565b6101496101bd3660046109dc565b610821565b6000610244888888888888886040516020016101e49796959493929190610b26565b60408051601f1981840301815282825280516020918201207f19457468657265756d205369676e6564204d6573736167653a0a33320000000084830152603c8085019190915282518085039091018152605c909301909152815191012090565b98975050505050505050565b600061025d604183610b89565b156102a85760405162461bcd60e51b815260206004820152601660248201527544617461206e6f742065787065637465642073697a6560501b60448201526064015b60405180910390fd5b60006102b5604184610bb3565b90506102bf610743565b8110156103025760405162461bcd60e51b815260206004820152601160248201527053696720636f756e7420746f6f206c6f7760781b604482015260640161029f565b600060025467ffffffffffffffff81111561031f5761031f610bc7565b604051908082528060200260200182016040528015610348578160200160208202803683370190505b50905060005b828110156104c6576000610363826041610bdd565b90506000610372826041610bfc565b905060006103868a61018f84868c8e610c14565b6001600160a01b0381166000908152600160208190526040909120015490915060ff166103ee5760405162461bcd60e51b815260206004820152601660248201527513595cdcd859d948139bdd08105d5d1a1bdc9a5e995960521b604482015260640161029f565b6001600160a01b0381166000908152600160205260409020548551869190811061041a5761041a610c3e565b60200260200101511561046f5760405162461bcd60e51b815260206004820152601960248201527f4475706c696361746520417574686f72697a6572205573656400000000000000604482015260640161029f565b6001600160a01b0381166000908152600160208190526040909120548651879190811061049e5761049e610c3e565b60200260200101901515908115158152505050505080806104be90610c54565b91505061034e565b5060019695505050505050565b6104db6108fd565b6001600160a01b0381166000908152600160208190526040909120015460ff16156105485760405162461bcd60e51b815260206004820152601d60248201527f4164647265737320697320416c726561647920417574686f72697a6572000000604482015260640161029f565b600354156106095760405180604001604052806003600160038054905061056f9190610c6d565b8154811061057f5761057f610c3e565b600091825260208083209190910154835260019281018390526001600160a01b03851682528281526040909120835181559201519101805460ff191691151591909117905560038054806105d5576105d5610c84565b600190038181906000526020600020016000905590556001600260008282546105fe9190610bfc565b909155506106689050565b604080518082018252600280548252600160208084018281526001600160a01b038716600090815291839052948120935184559351928101805460ff191693151593909317909255805491929091610662908490610bfc565b90915550505b50565b6106736108fd565b61067d6000610957565b565b60006041821461068e57600080fd5b60008060006106d286868080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152506109a792505050565b6040805160008152602081018083528c905260ff8516918101919091526060810183905260808101829052929550909350915060019060a0016020604051602081039080840390855afa15801561072d573d6000803e3d6000fd5b5050604051601f19015198975050505050505050565b600060036002541015610757575060025490565b600060036002546107689190610bb3565b90506000600360025461077b9190610b89565b905080156107a0578061078f836002610bdd565b6107999190610bfc565b9250505090565b610799826002610bdd565b6107b36108fd565b6001600160a01b0381166108185760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b606482015260840161029f565b61066881610957565b6108296108fd565b6001600160a01b0381166000908152600160208190526040909120015460ff166108955760405162461bcd60e51b815260206004820152601a60248201527f41646472657373206e6f7420616e20417574686f72697a657273000000000000604482015260640161029f565b6001600160a01b03811660009081526001602081905260408220808201805460ff19169055546003805480840182559084527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b01556002805491929091610662908490610c6d565b6000546001600160a01b0316331461067d5760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015260640161029f565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6020810151604082015160609092015160001a92909190565b80356001600160a01b03811681146109d757600080fd5b919050565b6000602082840312156109ee57600080fd5b6109f7826109c0565b9392505050565b60008083601f840112610a1057600080fd5b50813567ffffffffffffffff811115610a2857600080fd5b602083019150836020828501011115610a4057600080fd5b9250929050565b600080600080600080600060a0888a031215610a6257600080fd5b610a6b886109c0565b965060208801359550604088013567ffffffffffffffff80821115610a8f57600080fd5b610a9b8b838c016109fe565b909750955060608a0135915080821115610ab457600080fd5b50610ac18a828b016109fe565b989b979a50959894979596608090950135949350505050565b600080600060408486031215610aef57600080fd5b83359250602084013567ffffffffffffffff811115610b0d57600080fd5b610b19868287016109fe565b9497909650939450505050565b6bffffffffffffffffffffffff198860601b168152866014820152848660348301376000858201603481016000815285878237506034940193840192909252505060540195945050505050565b634e487b7160e01b600052601260045260246000fd5b600082610b9857610b98610b73565b500690565b634e487b7160e01b600052601160045260246000fd5b600082610bc257610bc2610b73565b500490565b634e487b7160e01b600052604160045260246000fd5b6000816000190483118215151615610bf757610bf7610b9d565b500290565b60008219821115610c0f57610c0f610b9d565b500190565b60008085851115610c2457600080fd5b83861115610c3157600080fd5b5050820193919092039150565b634e487b7160e01b600052603260045260246000fd5b600060018201610c6657610c66610b9d565b5060010190565b600082821015610c7f57610c7f610b9d565b500390565b634e487b7160e01b600052603160045260246000fdfea2646970667358221220e72e52db134d0238877d5eea24e07dac8fa6b62dee0080fb0c9b1a0f879386be64736f6c634300080f0033",
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

// Authorize is a free data retrieval call binding the contract method 0x296bb816.
//
// Solidity: function authorize(bytes32 message, bytes signatures) view returns(bool)
func (_Authorizers *AuthorizersCaller) Authorize(opts *bind.CallOpts, message [32]byte, signatures []byte) (bool, error) {
	var out []interface{}
	err := _Authorizers.contract.Call(opts, &out, "authorize", message, signatures)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Authorize is a free data retrieval call binding the contract method 0x296bb816.
//
// Solidity: function authorize(bytes32 message, bytes signatures) view returns(bool)
func (_Authorizers *AuthorizersSession) Authorize(message [32]byte, signatures []byte) (bool, error) {
	return _Authorizers.Contract.Authorize(&_Authorizers.CallOpts, message, signatures)
}

// Authorize is a free data retrieval call binding the contract method 0x296bb816.
//
// Solidity: function authorize(bytes32 message, bytes signatures) view returns(bool)
func (_Authorizers *AuthorizersCallerSession) Authorize(message [32]byte, signatures []byte) (bool, error) {
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

// RecoverSigner is a free data retrieval call binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 message, bytes signature) pure returns(address)
func (_Authorizers *AuthorizersCaller) RecoverSigner(opts *bind.CallOpts, message [32]byte, signature []byte) (common.Address, error) {
	var out []interface{}
	err := _Authorizers.contract.Call(opts, &out, "recoverSigner", message, signature)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RecoverSigner is a free data retrieval call binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 message, bytes signature) pure returns(address)
func (_Authorizers *AuthorizersSession) RecoverSigner(message [32]byte, signature []byte) (common.Address, error) {
	return _Authorizers.Contract.RecoverSigner(&_Authorizers.CallOpts, message, signature)
}

// RecoverSigner is a free data retrieval call binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 message, bytes signature) pure returns(address)
func (_Authorizers *AuthorizersCallerSession) RecoverSigner(message [32]byte, signature []byte) (common.Address, error) {
	return _Authorizers.Contract.RecoverSigner(&_Authorizers.CallOpts, message, signature)
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
