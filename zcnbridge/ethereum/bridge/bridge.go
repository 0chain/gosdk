// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bridge

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

// BridgeMetaData contains all meta data concerning the Bridge contract.
var BridgeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"contractIAuthorizers\",\"name\":\"_authorizers\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousAuthorizers\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newAuthorizers\",\"type\":\"address\"}],\"name\":\"AuthorizersTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"clientId\",\"type\":\"bytes\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"Burned\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"txid\",\"type\":\"bytes\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"Minted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"authorizers\",\"outputs\":[{\"internalType\":\"contractIAuthorizers\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"balance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenToRescue\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"rescueFunds\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"userNonceMinted\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"getUserNonceMinted\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_clientId\",\"type\":\"bytes\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"_signatures\",\"type\":\"bytes[]\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_for\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"_signatures\",\"type\":\"bytes[]\"}],\"name\":\"mintFor\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"isAuthorizationValid\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052600060045534801561001557600080fd5b506040516200141e3803806200141e8339810160408190526100369161010f565b8082610041336100a7565b600180546001600160a01b039283166001600160a01b0319918216179091556002805492841692909116821790556040516000907fc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d908290a35050600160035550610149565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6001600160a01b038116811461010c57600080fd5b50565b6000806040838503121561012257600080fd5b825161012d816100f7565b602084015190925061013e816100f7565b809150509250929050565b6112c580620001596000396000f3fe608060405234801561001057600080fd5b50600436106100cf5760003560e01c8063715018a61161008c578063e563e52611610066578063e563e526146101ab578063f2fde38b146101d4578063fc0c546a146101e7578063fe9d9303146101fa57600080fd5b8063715018a61461018a5780638da5cb5b14610192578063b69ef8a8146101a357600080fd5b8063062f950e146100d457806323b318801461010757806325250e0a1461011c57806356741b2c1461013f57806356f6dd62146101645780636ccae05414610177575b600080fd5b6100f46100e2366004610d74565b60056020526000908152604090205481565b6040519081526020015b60405180910390f35b61011a610115366004610e26565b61020d565b005b61012f61012a366004610ebc565b610377565b60405190151581526020016100fe565b6002546001600160a01b03165b6040516001600160a01b0390911681526020016100fe565b61011a610172366004610e26565b61041b565b61012f610185366004610f3f565b61056a565b61011a610667565b6000546001600160a01b031661014c565b6100f461067b565b6100f46101b9366004610d74565b6001600160a01b031660009081526005602052604090205490565b61011a6101e2366004610d74565b6106ed565b60015461014c906001600160a01b031681565b61011a610208366004610f80565b610766565b6001600160a01b038716600090815260056020526040902054879084908190610237906001610fcc565b1461025d5760405162461bcd60e51b815260040161025490610ff2565b60405180910390fd5b604080516080810182526001600160a01b038b16815260208082018b90528251601f8a0182900482028101820184528981528c938c938c938c938c938c938c936000939192830191908990899081908401838280828437600092018290525093855250505060209091018690529091506102df6002546001600160a01b031690565b6001600160a01b0316630b249ae48a8a8a8a8a6040518663ffffffff1660e01b8152600401610312959493929190611089565b6020604051808303816000875af1158015610331573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061035591906110c3565b9050610363828286866107ac565b505050505050505050505050505050505050565b60008061038c6002546001600160a01b031690565b6001600160a01b0316630b249ae4338a8a8a8a6040518663ffffffff1660e01b81526004016103bf959493929190611089565b6020604051808303816000875af11580156103de573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061040291906110c3565b905061040f818585610a63565b98975050505050505050565b6001600160a01b038716600090815260056020526040902054879084908190610445906001610fcc565b146104625760405162461bcd60e51b815260040161025490610ff2565b600060405180608001604052808b6001600160a01b031681526020018a815260200189898080601f016020809104026020016040519081016040528093929190818152602001838380828437600092018290525093855250505060209091018890529091506104d96002546001600160a01b031690565b6001600160a01b0316630b249ae48c8c8c8c8c6040518663ffffffff1660e01b815260040161050c959493929190611089565b6020604051808303816000875af115801561052b573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061054f91906110c3565b905061055d828288886107ac565b5050505050505050505050565b6000610574610b0c565b6001546001600160a01b038086169116036105ec5760405162461bcd60e51b815260206004820152603260248201527f546f6b656e506f6f6c3a2043616e6e6f7420636c61696d20746f6b656e2068656044820152711b1908189e481d1a194818dbdb9d1c9858dd60721b6064820152608401610254565b60405163a9059cbb60e01b81526001600160a01b0384811660048301526024820184905285169063a9059cbb906044016020604051808303816000875af115801561063b573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061065f91906110dc565b949350505050565b61066f610b0c565b6106796000610b66565b565b6001546040516370a0823160e01b81523060048201526000916001600160a01b0316906370a0823190602401602060405180830381865afa1580156106c4573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106e891906110c3565b905090565b6106f5610b0c565b6001600160a01b03811661075a5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b6064820152608401610254565b61076381610b66565b50565b6107a7338484848080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250610bb692505050565b505050565b6002600354036107fe5760405162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c006044820152606401610254565b600260038190555460405163e304688f60e01b81528491849184916001600160a01b03169063e304688f9061083b908690869086906004016110fe565b6020604051808303816000875af115801561085a573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061087e91906110dc565b61089a5760405162461bcd60e51b8152600401610254906111a6565b306001600160a01b031663fc0c546a6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156108d8573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906108fc91906111ec565b8751602089015160405163a9059cbb60e01b81526001600160a01b039283166004820152602481019190915291169063a9059cbb906044016020604051808303816000875af1158015610953573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061097791906110dc565b6109cf5760405162461bcd60e51b815260206004820152602360248201527f4272696467653a207472616e73666572206f7574206f6620706f6f6c206661696044820152621b195960ea1b6064820152608401610254565b86606001516005600089600001516001600160a01b03166001600160a01b0316815260200190815260200160002081905550866060015187600001516001600160a01b03167fe04478a4154dc31a079fa36b9ee1af057f492a47c1524ac67f2ea4c214c3de9289602001518a60400151604051610a4d929190611239565b60405180910390a3505060016003555050505050565b60025460405163e304688f60e01b81526000918591859185916001600160a01b039091169063e304688f90610aa0908690869086906004016110fe565b6020604051808303816000875af1158015610abf573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610ae391906110dc565b610aff5760405162461bcd60e51b8152600401610254906111a6565b5060019695505050505050565b6000546001600160a01b031633146106795760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e65726044820152606401610254565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b306001600160a01b031663fc0c546a6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610bf4573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610c1891906111ec565b6040516323b872dd60e01b81526001600160a01b0385811660048301523060248301526044820185905291909116906323b872dd906064016020604051808303816000875af1158015610c6f573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610c9391906110dc565b610cee5760405162461bcd60e51b815260206004820152602660248201527f4272696467653a207472616e7366657220696e746f206275726e20706f6f6c2060448201526519985a5b195960d21b6064820152608401610254565b600454610cfc906001610fcc565b6004819055604051610d0f908390611273565b6040518091039020846001600160a01b03167f2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df285604051610d5291815260200190565b60405180910390a4505050565b6001600160a01b038116811461076357600080fd5b600060208284031215610d8657600080fd5b8135610d9181610d5f565b9392505050565b60008083601f840112610daa57600080fd5b50813567ffffffffffffffff811115610dc257600080fd5b602083019150836020828501011115610dda57600080fd5b9250929050565b60008083601f840112610df357600080fd5b50813567ffffffffffffffff811115610e0b57600080fd5b6020830191508360208260051b8501011115610dda57600080fd5b600080600080600080600060a0888a031215610e4157600080fd5b8735610e4c81610d5f565b965060208801359550604088013567ffffffffffffffff80821115610e7057600080fd5b610e7c8b838c01610d98565b909750955060608a0135945060808a0135915080821115610e9c57600080fd5b50610ea98a828b01610de1565b989b979a50959850939692959293505050565b60008060008060008060808789031215610ed557600080fd5b86359550602087013567ffffffffffffffff80821115610ef457600080fd5b610f008a838b01610d98565b9097509550604089013594506060890135915080821115610f2057600080fd5b50610f2d89828a01610de1565b979a9699509497509295939492505050565b600080600060608486031215610f5457600080fd5b8335610f5f81610d5f565b92506020840135610f6f81610d5f565b929592945050506040919091013590565b600080600060408486031215610f9557600080fd5b83359250602084013567ffffffffffffffff811115610fb357600080fd5b610fbf86828701610d98565b9497909650939450505050565b60008219821115610fed57634e487b7160e01b600052601160045260246000fd5b500190565b60208082526048908201527f69664e6f744d696e7465643a206e6f6e63652070726f7669646564206d75737460408201527f20312067726561746572207468616e207468652070726576696f75732062757260608201526737103737b731b29760c11b608082015260a00190565b81835281816020850137506000828201602090810191909152601f909101601f19169091010190565b60018060a01b03861681528460208201526080604082015260006110b1608083018587611060565b90508260608301529695505050505050565b6000602082840312156110d557600080fd5b5051919050565b6000602082840312156110ee57600080fd5b81518015158114610d9157600080fd5b60006040820185835260206040818501528185835260608501905060608660051b86010192508660005b8781101561119857868503605f190183528135368a9003601e1901811261114e57600080fd5b8901848101903567ffffffffffffffff81111561116a57600080fd5b80360382131561117957600080fd5b611184878284611060565b965050509183019190830190600101611128565b509298975050505050505050565b60208082526026908201527f417574686f72697a6572733a207369676e617475726573206e6f7420617574686040820152651bdc9a5e995960d21b606082015260800190565b6000602082840312156111fe57600080fd5b8151610d9181610d5f565b60005b8381101561122457818101518382015260200161120c565b83811115611233576000848401525b50505050565b828152604060208201526000825180604084015261125e816060850160208701611209565b601f01601f1916919091016060019392505050565b60008251611285818460208701611209565b919091019291505056fea2646970667358221220ff339aeb671ed7aad629a2c68ec0ddf7695b44909257d409cc095374fdb0862e64736f6c634300080f0033",
}

// BridgeABI is the input ABI used to generate the binding from.
// Deprecated: Use BridgeMetaData.ABI instead.
var BridgeABI = BridgeMetaData.ABI

// BridgeBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use BridgeMetaData.Bin instead.
var BridgeBin = BridgeMetaData.Bin

// DeployBridge deploys a new Ethereum contract, binding an instance of Bridge to it.
func DeployBridge(auth *bind.TransactOpts, backend bind.ContractBackend, _token common.Address, _authorizers common.Address) (common.Address, *types.Transaction, *Bridge, error) {
	parsed, err := BridgeMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(BridgeBin), backend, _token, _authorizers)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

// Bridge is an auto generated Go binding around an Ethereum contract.
type Bridge struct {
	BridgeCaller     // Read-only binding to the contract
	BridgeTransactor // Write-only binding to the contract
	BridgeFilterer   // Log filterer for contract events
}

// BridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type BridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BridgeSession struct {
	Contract     *Bridge           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BridgeCallerSession struct {
	Contract *BridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BridgeTransactorSession struct {
	Contract     *BridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type BridgeRaw struct {
	Contract *Bridge // Generic contract binding to access the raw methods on
}

// BridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BridgeCallerRaw struct {
	Contract *BridgeCaller // Generic read-only contract binding to access the raw methods on
}

// BridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BridgeTransactorRaw struct {
	Contract *BridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBridge creates a new instance of Bridge, bound to a specific deployed contract.
func NewBridge(address common.Address, backend bind.ContractBackend) (*Bridge, error) {
	contract, err := bindBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

// NewBridgeCaller creates a new read-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeCaller(address common.Address, caller bind.ContractCaller) (*BridgeCaller, error) {
	contract, err := bindBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeCaller{contract: contract}, nil
}

// NewBridgeTransactor creates a new write-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*BridgeTransactor, error) {
	contract, err := bindBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeTransactor{contract: contract}, nil
}

// NewBridgeFilterer creates a new log filterer instance of Bridge, bound to a specific deployed contract.
func NewBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*BridgeFilterer, error) {
	contract, err := bindBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BridgeFilterer{contract: contract}, nil
}

// bindBridge binds a generic wrapper to an already deployed contract.
func bindBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BridgeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.BridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transact(opts, method, params...)
}

// Authorizers is a free data retrieval call binding the contract method 0x56741b2c.
//
// Solidity: function authorizers() view returns(address)
func (_Bridge *BridgeCaller) Authorizers(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "authorizers")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Authorizers is a free data retrieval call binding the contract method 0x56741b2c.
//
// Solidity: function authorizers() view returns(address)
func (_Bridge *BridgeSession) Authorizers() (common.Address, error) {
	return _Bridge.Contract.Authorizers(&_Bridge.CallOpts)
}

// Authorizers is a free data retrieval call binding the contract method 0x56741b2c.
//
// Solidity: function authorizers() view returns(address)
func (_Bridge *BridgeCallerSession) Authorizers() (common.Address, error) {
	return _Bridge.Contract.Authorizers(&_Bridge.CallOpts)
}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_Bridge *BridgeCaller) Balance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "balance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_Bridge *BridgeSession) Balance() (*big.Int, error) {
	return _Bridge.Contract.Balance(&_Bridge.CallOpts)
}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_Bridge *BridgeCallerSession) Balance() (*big.Int, error) {
	return _Bridge.Contract.Balance(&_Bridge.CallOpts)
}

// GetUserNonceMinted is a free data retrieval call binding the contract method 0xe563e526.
//
// Solidity: function getUserNonceMinted(address to) view returns(uint256)
func (_Bridge *BridgeCaller) GetUserNonceMinted(opts *bind.CallOpts, to common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getUserNonceMinted", to)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUserNonceMinted is a free data retrieval call binding the contract method 0xe563e526.
//
// Solidity: function getUserNonceMinted(address to) view returns(uint256)
func (_Bridge *BridgeSession) GetUserNonceMinted(to common.Address) (*big.Int, error) {
	return _Bridge.Contract.GetUserNonceMinted(&_Bridge.CallOpts, to)
}

// GetUserNonceMinted is a free data retrieval call binding the contract method 0xe563e526.
//
// Solidity: function getUserNonceMinted(address to) view returns(uint256)
func (_Bridge *BridgeCallerSession) GetUserNonceMinted(to common.Address) (*big.Int, error) {
	return _Bridge.Contract.GetUserNonceMinted(&_Bridge.CallOpts, to)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeSession) Owner() (common.Address, error) {
	return _Bridge.Contract.Owner(&_Bridge.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeCallerSession) Owner() (common.Address, error) {
	return _Bridge.Contract.Owner(&_Bridge.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Bridge *BridgeCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Bridge *BridgeSession) Token() (common.Address, error) {
	return _Bridge.Contract.Token(&_Bridge.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Bridge *BridgeCallerSession) Token() (common.Address, error) {
	return _Bridge.Contract.Token(&_Bridge.CallOpts)
}

// UserNonceMinted is a free data retrieval call binding the contract method 0x062f950e.
//
// Solidity: function userNonceMinted(address ) view returns(uint256)
func (_Bridge *BridgeCaller) UserNonceMinted(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "userNonceMinted", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UserNonceMinted is a free data retrieval call binding the contract method 0x062f950e.
//
// Solidity: function userNonceMinted(address ) view returns(uint256)
func (_Bridge *BridgeSession) UserNonceMinted(arg0 common.Address) (*big.Int, error) {
	return _Bridge.Contract.UserNonceMinted(&_Bridge.CallOpts, arg0)
}

// UserNonceMinted is a free data retrieval call binding the contract method 0x062f950e.
//
// Solidity: function userNonceMinted(address ) view returns(uint256)
func (_Bridge *BridgeCallerSession) UserNonceMinted(arg0 common.Address) (*big.Int, error) {
	return _Bridge.Contract.UserNonceMinted(&_Bridge.CallOpts, arg0)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 _amount, bytes _clientId) returns()
func (_Bridge *BridgeTransactor) Burn(opts *bind.TransactOpts, _amount *big.Int, _clientId []byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "burn", _amount, _clientId)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 _amount, bytes _clientId) returns()
func (_Bridge *BridgeSession) Burn(_amount *big.Int, _clientId []byte) (*types.Transaction, error) {
	return _Bridge.Contract.Burn(&_Bridge.TransactOpts, _amount, _clientId)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 _amount, bytes _clientId) returns()
func (_Bridge *BridgeTransactorSession) Burn(_amount *big.Int, _clientId []byte) (*types.Transaction, error) {
	return _Bridge.Contract.Burn(&_Bridge.TransactOpts, _amount, _clientId)
}

// IsAuthorizationValid is a paid mutator transaction binding the contract method 0x25250e0a.
//
// Solidity: function isAuthorizationValid(uint256 _amount, bytes _txid, uint256 _nonce, bytes[] signatures) returns(bool)
func (_Bridge *BridgeTransactor) IsAuthorizationValid(opts *bind.TransactOpts, _amount *big.Int, _txid []byte, _nonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "isAuthorizationValid", _amount, _txid, _nonce, signatures)
}

// IsAuthorizationValid is a paid mutator transaction binding the contract method 0x25250e0a.
//
// Solidity: function isAuthorizationValid(uint256 _amount, bytes _txid, uint256 _nonce, bytes[] signatures) returns(bool)
func (_Bridge *BridgeSession) IsAuthorizationValid(_amount *big.Int, _txid []byte, _nonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.IsAuthorizationValid(&_Bridge.TransactOpts, _amount, _txid, _nonce, signatures)
}

// IsAuthorizationValid is a paid mutator transaction binding the contract method 0x25250e0a.
//
// Solidity: function isAuthorizationValid(uint256 _amount, bytes _txid, uint256 _nonce, bytes[] signatures) returns(bool)
func (_Bridge *BridgeTransactorSession) IsAuthorizationValid(_amount *big.Int, _txid []byte, _nonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.IsAuthorizationValid(&_Bridge.TransactOpts, _amount, _txid, _nonce, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x56f6dd62.
//
// Solidity: function mint(address _to, uint256 _amount, bytes _txid, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeTransactor) Mint(opts *bind.TransactOpts, _to common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "mint", _to, _amount, _txid, _nonce, _signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x56f6dd62.
//
// Solidity: function mint(address _to, uint256 _amount, bytes _txid, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeSession) Mint(_to common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.Mint(&_Bridge.TransactOpts, _to, _amount, _txid, _nonce, _signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x56f6dd62.
//
// Solidity: function mint(address _to, uint256 _amount, bytes _txid, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeTransactorSession) Mint(_to common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.Mint(&_Bridge.TransactOpts, _to, _amount, _txid, _nonce, _signatures)
}

// MintFor is a paid mutator transaction binding the contract method 0x23b31880.
//
// Solidity: function mintFor(address _for, uint256 _amount, bytes _txid, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeTransactor) MintFor(opts *bind.TransactOpts, _for common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "mintFor", _for, _amount, _txid, _nonce, _signatures)
}

// MintFor is a paid mutator transaction binding the contract method 0x23b31880.
//
// Solidity: function mintFor(address _for, uint256 _amount, bytes _txid, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeSession) MintFor(_for common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.MintFor(&_Bridge.TransactOpts, _for, _amount, _txid, _nonce, _signatures)
}

// MintFor is a paid mutator transaction binding the contract method 0x23b31880.
//
// Solidity: function mintFor(address _for, uint256 _amount, bytes _txid, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeTransactorSession) MintFor(_for common.Address, _amount *big.Int, _txid []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.MintFor(&_Bridge.TransactOpts, _for, _amount, _txid, _nonce, _signatures)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bridge.Contract.RenounceOwnership(&_Bridge.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bridge.Contract.RenounceOwnership(&_Bridge.TransactOpts)
}

// RescueFunds is a paid mutator transaction binding the contract method 0x6ccae054.
//
// Solidity: function rescueFunds(address tokenToRescue, address to, uint256 amount) returns(bool)
func (_Bridge *BridgeTransactor) RescueFunds(opts *bind.TransactOpts, tokenToRescue common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "rescueFunds", tokenToRescue, to, amount)
}

// RescueFunds is a paid mutator transaction binding the contract method 0x6ccae054.
//
// Solidity: function rescueFunds(address tokenToRescue, address to, uint256 amount) returns(bool)
func (_Bridge *BridgeSession) RescueFunds(tokenToRescue common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.RescueFunds(&_Bridge.TransactOpts, tokenToRescue, to, amount)
}

// RescueFunds is a paid mutator transaction binding the contract method 0x6ccae054.
//
// Solidity: function rescueFunds(address tokenToRescue, address to, uint256 amount) returns(bool)
func (_Bridge *BridgeTransactorSession) RescueFunds(tokenToRescue common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.RescueFunds(&_Bridge.TransactOpts, tokenToRescue, to, amount)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.TransferOwnership(&_Bridge.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.TransferOwnership(&_Bridge.TransactOpts, newOwner)
}

// BridgeAuthorizersTransferredIterator is returned from FilterAuthorizersTransferred and is used to iterate over the raw logs and unpacked data for AuthorizersTransferred events raised by the Bridge contract.
type BridgeAuthorizersTransferredIterator struct {
	Event *BridgeAuthorizersTransferred // Event containing the contract specifics and raw log

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
func (it *BridgeAuthorizersTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeAuthorizersTransferred)
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
		it.Event = new(BridgeAuthorizersTransferred)
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
func (it *BridgeAuthorizersTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeAuthorizersTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeAuthorizersTransferred represents a AuthorizersTransferred event raised by the Bridge contract.
type BridgeAuthorizersTransferred struct {
	PreviousAuthorizers common.Address
	NewAuthorizers      common.Address
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterAuthorizersTransferred is a free log retrieval operation binding the contract event 0xc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d.
//
// Solidity: event AuthorizersTransferred(address indexed previousAuthorizers, address indexed newAuthorizers)
func (_Bridge *BridgeFilterer) FilterAuthorizersTransferred(opts *bind.FilterOpts, previousAuthorizers []common.Address, newAuthorizers []common.Address) (*BridgeAuthorizersTransferredIterator, error) {

	var previousAuthorizersRule []interface{}
	for _, previousAuthorizersItem := range previousAuthorizers {
		previousAuthorizersRule = append(previousAuthorizersRule, previousAuthorizersItem)
	}
	var newAuthorizersRule []interface{}
	for _, newAuthorizersItem := range newAuthorizers {
		newAuthorizersRule = append(newAuthorizersRule, newAuthorizersItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "AuthorizersTransferred", previousAuthorizersRule, newAuthorizersRule)
	if err != nil {
		return nil, err
	}
	return &BridgeAuthorizersTransferredIterator{contract: _Bridge.contract, event: "AuthorizersTransferred", logs: logs, sub: sub}, nil
}

// WatchAuthorizersTransferred is a free log subscription operation binding the contract event 0xc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d.
//
// Solidity: event AuthorizersTransferred(address indexed previousAuthorizers, address indexed newAuthorizers)
func (_Bridge *BridgeFilterer) WatchAuthorizersTransferred(opts *bind.WatchOpts, sink chan<- *BridgeAuthorizersTransferred, previousAuthorizers []common.Address, newAuthorizers []common.Address) (event.Subscription, error) {

	var previousAuthorizersRule []interface{}
	for _, previousAuthorizersItem := range previousAuthorizers {
		previousAuthorizersRule = append(previousAuthorizersRule, previousAuthorizersItem)
	}
	var newAuthorizersRule []interface{}
	for _, newAuthorizersItem := range newAuthorizers {
		newAuthorizersRule = append(newAuthorizersRule, newAuthorizersItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "AuthorizersTransferred", previousAuthorizersRule, newAuthorizersRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeAuthorizersTransferred)
				if err := _Bridge.contract.UnpackLog(event, "AuthorizersTransferred", log); err != nil {
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

// ParseAuthorizersTransferred is a log parse operation binding the contract event 0xc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d.
//
// Solidity: event AuthorizersTransferred(address indexed previousAuthorizers, address indexed newAuthorizers)
func (_Bridge *BridgeFilterer) ParseAuthorizersTransferred(log types.Log) (*BridgeAuthorizersTransferred, error) {
	event := new(BridgeAuthorizersTransferred)
	if err := _Bridge.contract.UnpackLog(event, "AuthorizersTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeBurnedIterator is returned from FilterBurned and is used to iterate over the raw logs and unpacked data for Burned events raised by the Bridge contract.
type BridgeBurnedIterator struct {
	Event *BridgeBurned // Event containing the contract specifics and raw log

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
func (it *BridgeBurnedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeBurned)
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
		it.Event = new(BridgeBurned)
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
func (it *BridgeBurnedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeBurnedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeBurned represents a Burned event raised by the Bridge contract.
type BridgeBurned struct {
	From     common.Address
	Amount   *big.Int
	ClientId common.Hash
	Nonce    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBurned is a free log retrieval operation binding the contract event 0x2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df2.
//
// Solidity: event Burned(address indexed from, uint256 amount, bytes indexed clientId, uint256 indexed nonce)
func (_Bridge *BridgeFilterer) FilterBurned(opts *bind.FilterOpts, from []common.Address, clientId [][]byte, nonce []*big.Int) (*BridgeBurnedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	var clientIdRule []interface{}
	for _, clientIdItem := range clientId {
		clientIdRule = append(clientIdRule, clientIdItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Burned", fromRule, clientIdRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return &BridgeBurnedIterator{contract: _Bridge.contract, event: "Burned", logs: logs, sub: sub}, nil
}

// WatchBurned is a free log subscription operation binding the contract event 0x2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df2.
//
// Solidity: event Burned(address indexed from, uint256 amount, bytes indexed clientId, uint256 indexed nonce)
func (_Bridge *BridgeFilterer) WatchBurned(opts *bind.WatchOpts, sink chan<- *BridgeBurned, from []common.Address, clientId [][]byte, nonce []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	var clientIdRule []interface{}
	for _, clientIdItem := range clientId {
		clientIdRule = append(clientIdRule, clientIdItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Burned", fromRule, clientIdRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeBurned)
				if err := _Bridge.contract.UnpackLog(event, "Burned", log); err != nil {
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

// ParseBurned is a log parse operation binding the contract event 0x2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df2.
//
// Solidity: event Burned(address indexed from, uint256 amount, bytes indexed clientId, uint256 indexed nonce)
func (_Bridge *BridgeFilterer) ParseBurned(log types.Log) (*BridgeBurned, error) {
	event := new(BridgeBurned)
	if err := _Bridge.contract.UnpackLog(event, "Burned", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeMintedIterator is returned from FilterMinted and is used to iterate over the raw logs and unpacked data for Minted events raised by the Bridge contract.
type BridgeMintedIterator struct {
	Event *BridgeMinted // Event containing the contract specifics and raw log

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
func (it *BridgeMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeMinted)
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
		it.Event = new(BridgeMinted)
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
func (it *BridgeMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeMinted represents a Minted event raised by the Bridge contract.
type BridgeMinted struct {
	To     common.Address
	Amount *big.Int
	Txid   []byte
	Nonce  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterMinted is a free log retrieval operation binding the contract event 0xe04478a4154dc31a079fa36b9ee1af057f492a47c1524ac67f2ea4c214c3de92.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, uint256 indexed nonce)
func (_Bridge *BridgeFilterer) FilterMinted(opts *bind.FilterOpts, to []common.Address, nonce []*big.Int) (*BridgeMintedIterator, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Minted", toRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return &BridgeMintedIterator{contract: _Bridge.contract, event: "Minted", logs: logs, sub: sub}, nil
}

// WatchMinted is a free log subscription operation binding the contract event 0xe04478a4154dc31a079fa36b9ee1af057f492a47c1524ac67f2ea4c214c3de92.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, uint256 indexed nonce)
func (_Bridge *BridgeFilterer) WatchMinted(opts *bind.WatchOpts, sink chan<- *BridgeMinted, to []common.Address, nonce []*big.Int) (event.Subscription, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Minted", toRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeMinted)
				if err := _Bridge.contract.UnpackLog(event, "Minted", log); err != nil {
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

// ParseMinted is a log parse operation binding the contract event 0xe04478a4154dc31a079fa36b9ee1af057f492a47c1524ac67f2ea4c214c3de92.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, uint256 indexed nonce)
func (_Bridge *BridgeFilterer) ParseMinted(log types.Log) (*BridgeMinted, error) {
	event := new(BridgeMinted)
	if err := _Bridge.contract.UnpackLog(event, "Minted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Bridge contract.
type BridgeOwnershipTransferredIterator struct {
	Event *BridgeOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BridgeOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeOwnershipTransferred)
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
		it.Event = new(BridgeOwnershipTransferred)
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
func (it *BridgeOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeOwnershipTransferred represents a OwnershipTransferred event raised by the Bridge contract.
type BridgeOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bridge *BridgeFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BridgeOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BridgeOwnershipTransferredIterator{contract: _Bridge.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bridge *BridgeFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BridgeOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeOwnershipTransferred)
				if err := _Bridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Bridge *BridgeFilterer) ParseOwnershipTransferred(log types.Log) (*BridgeOwnershipTransferred, error) {
	event := new(BridgeOwnershipTransferred)
	if err := _Bridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
