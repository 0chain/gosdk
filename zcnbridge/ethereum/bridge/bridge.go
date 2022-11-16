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
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"contractIAuthorizers\",\"name\":\"_authorizers\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousAuthorizers\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newAuthorizers\",\"type\":\"address\"}],\"name\":\"AuthorizersTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"clientId\",\"type\":\"bytes\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"Burned\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"txid\",\"type\":\"bytes\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"clientId\",\"type\":\"bytes\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"Minted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"authorizers\",\"outputs\":[{\"internalType\":\"contractIAuthorizers\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"balance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenToRescue\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"rescueFunds\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_clientId\",\"type\":\"bytes\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"_clientId\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"_signatures\",\"type\":\"bytes[]\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_for\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"_clientId\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"_signatures\",\"type\":\"bytes[]\"}],\"name\":\"mintFor\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_txid\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"_clientId\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"isAuthorizationValid\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052600060045534801561001557600080fd5b50604051620014c8380380620014c88339810160408190526100369161010f565b8082610041336100a7565b600180546001600160a01b039283166001600160a01b0319918216179091556002805492841692909116821790556040516000907fc44d874e85f1c5b65d10c0c33020d49211b91e9f2704457f2ef269e5fb7a6b5d908290a35050600160035550610149565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6001600160a01b038116811461010c57600080fd5b50565b6000806040838503121561012257600080fd5b825161012d816100f7565b602084015190925061013e816100f7565b809150509250929050565b61136f80620001596000396000f3fe608060405234801561001057600080fd5b50600436106100a95760003560e01c8063b0b70c5f11610071578063b0b70c5f14610129578063b69ef8a81461013c578063e2d6408914610152578063f2fde38b14610165578063fc0c546a14610178578063fe9d93031461018b57600080fd5b80634b099d95146100ae57806356741b2c146100d65780636ccae054146100fb578063715018a61461010e5780638da5cb5b14610118575b600080fd5b6100c16100bc366004610e1b565b61019e565b60405190151581526020015b60405180910390f35b6002546001600160a01b03165b6040516001600160a01b0390911681526020016100cd565b6100c1610109366004610edd565b610248565b61011661034a565b005b6000546001600160a01b03166100e3565b610116610137366004610f1e565b61035e565b61014461050e565b6040519081526020016100cd565b610116610160366004610e1b565b610580565b610116610173366004610fe0565b610708565b6001546100e3906001600160a01b031681565b610116610199366004611004565b610781565b6000806101b36002546001600160a01b031690565b6001600160a01b031663207b601e338c8c8c8c8c8c6040518863ffffffff1660e01b81526004016101ea9796959493929190611079565b6020604051808303816000875af1158015610209573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061022d91906110c9565b905061023a8185856107c7565b9a9950505050505050505050565b6000610252610870565b6001546001600160a01b038086169116036102cf5760405162461bcd60e51b815260206004820152603260248201527f546f6b656e506f6f6c3a2043616e6e6f7420636c61696d20746f6b656e2068656044820152711b1908189e481d1a194818dbdb9d1c9858dd60721b60648201526084015b60405180910390fd5b60405163a9059cbb60e01b81526001600160a01b0384811660048301526024820184905285169063a9059cbb906044016020604051808303816000875af115801561031e573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061034291906110e2565b949350505050565b610352610870565b61035c60006108ca565b565b8484848060058484604051610374929190611104565b908152602001604051809103902054600161038f9190611114565b146103ac5760405162461bcd60e51b81526004016102c69061113a565b6040805160a0810182526001600160a01b038e16815260208082018e90528251601f8d0182900482028101820184528c81528f938f938f938f938f938f938f938f938f936000939290830191908b908b9081908401838280828437600092019190915250505090825250604080516020601f8a01819004810282018101909252888152918101919089908990819084018382808284376000920182905250938552505050602090910186905290915061046d6002546001600160a01b031690565b6001600160a01b031663207b601e8c8c8c8c8c8c8c6040518863ffffffff1660e01b81526004016104a49796959493929190611079565b6020604051808303816000875af11580156104c3573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906104e791906110c9565b90506104f58282868661091a565b5050505050505050505050505050505050505050505050565b6001546040516370a0823160e01b81523060048201526000916001600160a01b0316906370a0823190602401602060405180830381865afa158015610557573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061057b91906110c9565b905090565b8484848060058484604051610596929190611104565b90815260200160405180910390205460016105b19190611114565b146105ce5760405162461bcd60e51b81526004016102c69061113a565b6040805160a08101825233815260208082018e90528251601f8d0182900482028101820184528c81526000938301918e908e9081908401838280828437600092019190915250505090825250604080516020601f8d018190048102820181019092528b815291810191908c908c9081908401838280828437600092018290525093855250505060209091018990529091506106716002546001600160a01b031690565b6001600160a01b031663207b601e338f8f8f8f8f8f6040518863ffffffff1660e01b81526004016106a89796959493929190611079565b6020604051808303816000875af11580156106c7573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106eb91906110c9565b90506106f98282898961091a565b50505050505050505050505050565b610710610870565b6001600160a01b0381166107755760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016102c6565b61077e816108ca565b50565b6107c2338484848080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250610be492505050565b505050565b60025460405163e304688f60e01b81526000918591859185916001600160a01b039091169063e304688f90610804908690869086906004016111a8565b6020604051808303816000875af1158015610823573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061084791906110e2565b6108635760405162461bcd60e51b81526004016102c690611250565b5060019695505050505050565b6000546001600160a01b0316331461035c5760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e657260448201526064016102c6565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60026003540361096c5760405162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c0060448201526064016102c6565b600260038190555460405163e304688f60e01b81528491849184916001600160a01b03169063e304688f906109a9908690869086906004016111a8565b6020604051808303816000875af11580156109c8573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109ec91906110e2565b610a085760405162461bcd60e51b81526004016102c690611250565b306001600160a01b031663fc0c546a6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610a46573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610a6a9190611296565b8751602089015160405163a9059cbb60e01b81526001600160a01b039283166004820152602481019190915291169063a9059cbb906044016020604051808303816000875af1158015610ac1573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610ae591906110e2565b610b3d5760405162461bcd60e51b815260206004820152602360248201527f4272696467653a207472616e73666572206f7574206f6620706f6f6c206661696044820152621b195960ea1b60648201526084016102c6565b866080015160058860600151604051610b5691906112e3565b90815260200160405180910390208190555086608001518760600151604051610b7f91906112e3565b604051809103902088600001516001600160a01b03167f154c9d73b1247a0e3b766a2dfdf858043dc6d1fb8a20536197f6a6359e871ed38a602001518b60400151604051610bce9291906112ff565b60405180910390a4505060016003555050505050565b306001600160a01b031663fc0c546a6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610c22573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610c469190611296565b6040516323b872dd60e01b81526001600160a01b0385811660048301523060248301526044820185905291909116906323b872dd906064016020604051808303816000875af1158015610c9d573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610cc191906110e2565b610d1c5760405162461bcd60e51b815260206004820152602660248201527f4272696467653a207472616e7366657220696e746f206275726e20706f6f6c2060448201526519985a5b195960d21b60648201526084016102c6565b600454610d2a906001611114565b6004819055604051610d3d9083906112e3565b6040518091039020846001600160a01b03167f2b1155a5de2441854f3781130b980daa499b3412053ee40fcde076774bb12df285604051610d8091815260200190565b60405180910390a4505050565b60008083601f840112610d9f57600080fd5b50813567ffffffffffffffff811115610db757600080fd5b602083019150836020828501011115610dcf57600080fd5b9250929050565b60008083601f840112610de857600080fd5b50813567ffffffffffffffff811115610e0057600080fd5b6020830191508360208260051b8501011115610dcf57600080fd5b60008060008060008060008060a0898b031215610e3757600080fd5b88359750602089013567ffffffffffffffff80821115610e5657600080fd5b610e628c838d01610d8d565b909950975060408b0135915080821115610e7b57600080fd5b610e878c838d01610d8d565b909750955060608b0135945060808b0135915080821115610ea757600080fd5b50610eb48b828c01610dd6565b999c989b5096995094979396929594505050565b6001600160a01b038116811461077e57600080fd5b600080600060608486031215610ef257600080fd5b8335610efd81610ec8565b92506020840135610f0d81610ec8565b929592945050506040919091013590565b600080600080600080600080600060c08a8c031215610f3c57600080fd5b8935610f4781610ec8565b985060208a0135975060408a013567ffffffffffffffff80821115610f6b57600080fd5b610f778d838e01610d8d565b909950975060608c0135915080821115610f9057600080fd5b610f9c8d838e01610d8d565b909750955060808c0135945060a08c0135915080821115610fbc57600080fd5b50610fc98c828d01610dd6565b915080935050809150509295985092959850929598565b600060208284031215610ff257600080fd5b8135610ffd81610ec8565b9392505050565b60008060006040848603121561101957600080fd5b83359250602084013567ffffffffffffffff81111561103757600080fd5b61104386828701610d8d565b9497909650939450505050565b81835281816020850137506000828201602090810191909152601f909101601f19169091010190565b60018060a01b038816815286602082015260a0604082015260006110a160a083018789611050565b82810360608401526110b4818688611050565b91505082608083015298975050505050505050565b6000602082840312156110db57600080fd5b5051919050565b6000602082840312156110f457600080fd5b81518015158114610ffd57600080fd5b8183823760009101908152919050565b6000821982111561113557634e487b7160e01b600052601160045260246000fd5b500190565b60208082526048908201527f69664e6f744d696e7465643a206e6f6e63652070726f7669646564206d75737460408201527f20312067726561746572207468616e207468652070726576696f75732062757260608201526737103737b731b29760c11b608082015260a00190565b60006040820185835260206040818501528185835260608501905060608660051b86010192508660005b8781101561124257868503605f190183528135368a9003601e190181126111f857600080fd5b8901848101903567ffffffffffffffff81111561121457600080fd5b80360382131561122357600080fd5b61122e878284611050565b9650505091830191908301906001016111d2565b509298975050505050505050565b60208082526026908201527f417574686f72697a6572733a207369676e617475726573206e6f7420617574686040820152651bdc9a5e995960d21b606082015260800190565b6000602082840312156112a857600080fd5b8151610ffd81610ec8565b60005b838110156112ce5781810151838201526020016112b6565b838111156112dd576000848401525b50505050565b600082516112f58184602087016112b3565b9190910192915050565b82815260406020820152600082518060408401526113248160608501602087016112b3565b601f01601f191691909101606001939250505056fea2646970667358221220ffcd2bdd5431db4b98390e93cbc166948946c85fe3db6cc13703e129330b7ace64736f6c634300080f0033",
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

// IsAuthorizationValid is a paid mutator transaction binding the contract method 0x4b099d95.
//
// Solidity: function isAuthorizationValid(uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce, bytes[] signatures) returns(bool)
func (_Bridge *BridgeTransactor) IsAuthorizationValid(opts *bind.TransactOpts, _amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "isAuthorizationValid", _amount, _txid, _clientId, _nonce, signatures)
}

// IsAuthorizationValid is a paid mutator transaction binding the contract method 0x4b099d95.
//
// Solidity: function isAuthorizationValid(uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce, bytes[] signatures) returns(bool)
func (_Bridge *BridgeSession) IsAuthorizationValid(_amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.IsAuthorizationValid(&_Bridge.TransactOpts, _amount, _txid, _clientId, _nonce, signatures)
}

// IsAuthorizationValid is a paid mutator transaction binding the contract method 0x4b099d95.
//
// Solidity: function isAuthorizationValid(uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce, bytes[] signatures) returns(bool)
func (_Bridge *BridgeTransactorSession) IsAuthorizationValid(_amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.IsAuthorizationValid(&_Bridge.TransactOpts, _amount, _txid, _clientId, _nonce, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0xe2d64089.
//
// Solidity: function mint(uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeTransactor) Mint(opts *bind.TransactOpts, _amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "mint", _amount, _txid, _clientId, _nonce, _signatures)
}

// Mint is a paid mutator transaction binding the contract method 0xe2d64089.
//
// Solidity: function mint(uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeSession) Mint(_amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.Mint(&_Bridge.TransactOpts, _amount, _txid, _clientId, _nonce, _signatures)
}

// Mint is a paid mutator transaction binding the contract method 0xe2d64089.
//
// Solidity: function mint(uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeTransactorSession) Mint(_amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.Mint(&_Bridge.TransactOpts, _amount, _txid, _clientId, _nonce, _signatures)
}

// MintFor is a paid mutator transaction binding the contract method 0xb0b70c5f.
//
// Solidity: function mintFor(address _for, uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeTransactor) MintFor(opts *bind.TransactOpts, _for common.Address, _amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "mintFor", _for, _amount, _txid, _clientId, _nonce, _signatures)
}

// MintFor is a paid mutator transaction binding the contract method 0xb0b70c5f.
//
// Solidity: function mintFor(address _for, uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeSession) MintFor(_for common.Address, _amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.MintFor(&_Bridge.TransactOpts, _for, _amount, _txid, _clientId, _nonce, _signatures)
}

// MintFor is a paid mutator transaction binding the contract method 0xb0b70c5f.
//
// Solidity: function mintFor(address _for, uint256 _amount, bytes _txid, bytes _clientId, uint256 _nonce, bytes[] _signatures) returns()
func (_Bridge *BridgeTransactorSession) MintFor(_for common.Address, _amount *big.Int, _txid []byte, _clientId []byte, _nonce *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.MintFor(&_Bridge.TransactOpts, _for, _amount, _txid, _clientId, _nonce, _signatures)
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
	To       common.Address
	Amount   *big.Int
	Txid     []byte
	ClientId common.Hash
	Nonce    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMinted is a free log retrieval operation binding the contract event 0x154c9d73b1247a0e3b766a2dfdf858043dc6d1fb8a20536197f6a6359e871ed3.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, bytes indexed clientId, uint256 indexed nonce)
func (_Bridge *BridgeFilterer) FilterMinted(opts *bind.FilterOpts, to []common.Address, clientId [][]byte, nonce []*big.Int) (*BridgeMintedIterator, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	var clientIdRule []interface{}
	for _, clientIdItem := range clientId {
		clientIdRule = append(clientIdRule, clientIdItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Minted", toRule, clientIdRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return &BridgeMintedIterator{contract: _Bridge.contract, event: "Minted", logs: logs, sub: sub}, nil
}

// WatchMinted is a free log subscription operation binding the contract event 0x154c9d73b1247a0e3b766a2dfdf858043dc6d1fb8a20536197f6a6359e871ed3.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, bytes indexed clientId, uint256 indexed nonce)
func (_Bridge *BridgeFilterer) WatchMinted(opts *bind.WatchOpts, sink chan<- *BridgeMinted, to []common.Address, clientId [][]byte, nonce []*big.Int) (event.Subscription, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	var clientIdRule []interface{}
	for _, clientIdItem := range clientId {
		clientIdRule = append(clientIdRule, clientIdItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Minted", toRule, clientIdRule, nonceRule)
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

// ParseMinted is a log parse operation binding the contract event 0x154c9d73b1247a0e3b766a2dfdf858043dc6d1fb8a20536197f6a6359e871ed3.
//
// Solidity: event Minted(address indexed to, uint256 amount, bytes txid, bytes indexed clientId, uint256 indexed nonce)
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
