// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bancornetwork

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

// BancorMetaData contains all meta data concerning the Bancor contract.
var BancorMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractITokenGovernance\",\"name\":\"initBNTGovernance\",\"type\":\"address\"},{\"internalType\":\"contractITokenGovernance\",\"name\":\"initVBNTGovernance\",\"type\":\"address\"},{\"internalType\":\"contractINetworkSettings\",\"name\":\"initNetworkSettings\",\"type\":\"address\"},{\"internalType\":\"contractIMasterVault\",\"name\":\"initMasterVault\",\"type\":\"address\"},{\"internalType\":\"contractIExternalProtectionVault\",\"name\":\"initExternalProtectionVault\",\"type\":\"address\"},{\"internalType\":\"contractIPoolToken\",\"name\":\"initBNTPoolToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"bancorArbitrage\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"carbonPOL\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"AccessDenied\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"AlreadyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"AlreadyInitialized\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DeadlineExpired\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DepositingDisabled\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DoesNotExist\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InsufficientFlashLoanReturn\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidFee\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidPool\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidToken\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NativeTokenAmountMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotEmpty\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotWhitelisted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotWhitelistedForPOL\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"Overflow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PoolNotInSurplus\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroValue\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractToken\",\"name\":\"zcntoken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"borrower\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"feeAmount\",\"type\":\"uint256\"}],\"name\":\"FlashLoanCompleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"contextId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"contractToken\",\"name\":\"zcntoken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"availableAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"originalAmount\",\"type\":\"uint256\"}],\"name\":\"FundsMigrated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"NetworkFeesWithdrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"oldRewardsPPM\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"newRewardsPPM\",\"type\":\"uint32\"}],\"name\":\"POLRewardsPPMUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"zcntoken\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"polTokenAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"userReward\",\"type\":\"uint256\"}],\"name\":\"POLWithdrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Paused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractToken\",\"name\":\"pool\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"contractIPoolCollection\",\"name\":\"poolCollection\",\"type\":\"address\"}],\"name\":\"PoolAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint16\",\"name\":\"poolType\",\"type\":\"uint16\"},{\"indexed\":true,\"internalType\":\"contractIPoolCollection\",\"name\":\"poolCollection\",\"type\":\"address\"}],\"name\":\"PoolCollectionAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint16\",\"name\":\"poolType\",\"type\":\"uint16\"},{\"indexed\":true,\"internalType\":\"contractIPoolCollection\",\"name\":\"poolCollection\",\"type\":\"address\"}],\"name\":\"PoolCollectionRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractToken\",\"name\":\"pool\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"contractIPoolCollection\",\"name\":\"poolCollection\",\"type\":\"address\"}],\"name\":\"PoolCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractToken\",\"name\":\"pool\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"contractIPoolCollection\",\"name\":\"poolCollection\",\"type\":\"address\"}],\"name\":\"PoolRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"contextId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"contractToken\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"contractToken\",\"name\":\"targetToken\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"sourceAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"targetAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"bntAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"targetFeeAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"bntFeeAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"}],\"name\":\"TokensTraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Unpaused\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"cancelWithdrawal\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"pool\",\"type\":\"address\"}],\"name\":\"collectionByPool\",\"outputs\":[{\"internalType\":\"contractIPoolCollection\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken[]\",\"name\":\"tokens\",\"type\":\"address[]\"},{\"internalType\":\"contractIPoolCollection\",\"name\":\"poolCollection\",\"type\":\"address\"}],\"name\":\"createPools\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"pool\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenAmount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"contractToken\",\"name\":\"pool\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenAmount\",\"type\":\"uint256\"}],\"name\":\"depositFor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"depositingEnabled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"enableDepositing\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"zcntoken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"contractIFlashLoanRecipient\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"flashLoan\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getRoleMember\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleMemberCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIPoolToken\",\"name\":\"poolToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"poolTokenAmount\",\"type\":\"uint256\"}],\"name\":\"initWithdrawal\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIBNTPool\",\"name\":\"initBNTPool\",\"type\":\"address\"},{\"internalType\":\"contractIPendingWithdrawals\",\"name\":\"initPendingWithdrawals\",\"type\":\"address\"},{\"internalType\":\"contractIPoolMigrator\",\"name\":\"initPoolMigrator\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"liquidityPools\",\"outputs\":[{\"internalType\":\"contractToken[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"zcntoken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"availableAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"originalAmount\",\"type\":\"uint256\"}],\"name\":\"migrateLiquidity\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken[]\",\"name\":\"pools\",\"type\":\"address[]\"},{\"internalType\":\"contractIPoolCollection\",\"name\":\"newPoolCollection\",\"type\":\"address\"}],\"name\":\"migratePools\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingNetworkFeeAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"polRewardsPPM\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"poolCollections\",\"outputs\":[{\"internalType\":\"contractIPoolCollection[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"postUpgrade\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIPoolCollection\",\"name\":\"newPoolCollection\",\"type\":\"address\"}],\"name\":\"registerPoolCollection\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"resume\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"roleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"roleEmergencyStopper\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"roleMigrationManager\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"roleNetworkFeeManager\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"newRewardsPPM\",\"type\":\"uint32\"}],\"name\":\"setPOLRewardsPPM\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"contractToken\",\"name\":\"targetToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"sourceAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minReturnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"tradeBySourceAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"contractToken\",\"name\":\"targetToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"sourceAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minReturnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"tradeBySourceAmountArb\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"contractToken\",\"name\":\"targetToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"targetAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxSourceAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"tradeByTargetAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"contractToken\",\"name\":\"targetToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"targetAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxSourceAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"tradeByTargetAmountArb\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIPoolCollection\",\"name\":\"poolCollection\",\"type\":\"address\"}],\"name\":\"unregisterPoolCollection\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"withdrawNetworkFees\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractToken\",\"name\":\"pool\",\"type\":\"address\"}],\"name\":\"withdrawPOL\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// BancorABI is the input ABI used to generate the binding from.
// Deprecated: Use BancorMetaData.ABI instead.
var BancorABI = BancorMetaData.ABI

// Bancor is an auto generated Go binding around an Ethereum contract.
type Bancor struct {
	BancorCaller     // Read-only binding to the contract
	BancorTransactor // Write-only binding to the contract
	BancorFilterer   // Log filterer for contract events
}

// BancorCaller is an auto generated read-only Go binding around an Ethereum contract.
type BancorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BancorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BancorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BancorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BancorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BancorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BancorSession struct {
	Contract     *Bancor           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BancorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BancorCallerSession struct {
	Contract *BancorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BancorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BancorTransactorSession struct {
	Contract     *BancorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BancorRaw is an auto generated low-level Go binding around an Ethereum contract.
type BancorRaw struct {
	Contract *Bancor // Generic contract binding to access the raw methods on
}

// BancorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BancorCallerRaw struct {
	Contract *BancorCaller // Generic read-only contract binding to access the raw methods on
}

// BancorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BancorTransactorRaw struct {
	Contract *BancorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBancor creates a new instance of Bancor, bound to a specific deployed contract.
func NewBancor(address common.Address, backend bind.ContractBackend) (*Bancor, error) {
	contract, err := bindBancor(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bancor{BancorCaller: BancorCaller{contract: contract}, BancorTransactor: BancorTransactor{contract: contract}, BancorFilterer: BancorFilterer{contract: contract}}, nil
}

// NewBancorCaller creates a new read-only instance of Bancor, bound to a specific deployed contract.
func NewBancorCaller(address common.Address, caller bind.ContractCaller) (*BancorCaller, error) {
	contract, err := bindBancor(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BancorCaller{contract: contract}, nil
}

// NewBancorTransactor creates a new write-only instance of Bancor, bound to a specific deployed contract.
func NewBancorTransactor(address common.Address, transactor bind.ContractTransactor) (*BancorTransactor, error) {
	contract, err := bindBancor(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BancorTransactor{contract: contract}, nil
}

// NewBancorFilterer creates a new log filterer instance of Bancor, bound to a specific deployed contract.
func NewBancorFilterer(address common.Address, filterer bind.ContractFilterer) (*BancorFilterer, error) {
	contract, err := bindBancor(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BancorFilterer{contract: contract}, nil
}

// bindBancor binds a generic wrapper to an already deployed contract.
func bindBancor(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BancorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bancor *BancorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bancor.Contract.BancorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bancor *BancorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bancor.Contract.BancorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bancor *BancorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bancor.Contract.BancorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bancor *BancorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bancor.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bancor *BancorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bancor.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bancor *BancorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bancor.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Bancor *BancorCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Bancor *BancorSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Bancor.Contract.DEFAULTADMINROLE(&_Bancor.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Bancor *BancorCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Bancor.Contract.DEFAULTADMINROLE(&_Bancor.CallOpts)
}

// CollectionByPool is a free data retrieval call binding the contract method 0x9bca0e70.
//
// Solidity: function collectionByPool(address pool) view returns(address)
func (_Bancor *BancorCaller) CollectionByPool(opts *bind.CallOpts, pool common.Address) (common.Address, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "collectionByPool", pool)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CollectionByPool is a free data retrieval call binding the contract method 0x9bca0e70.
//
// Solidity: function collectionByPool(address pool) view returns(address)
func (_Bancor *BancorSession) CollectionByPool(pool common.Address) (common.Address, error) {
	return _Bancor.Contract.CollectionByPool(&_Bancor.CallOpts, pool)
}

// CollectionByPool is a free data retrieval call binding the contract method 0x9bca0e70.
//
// Solidity: function collectionByPool(address pool) view returns(address)
func (_Bancor *BancorCallerSession) CollectionByPool(pool common.Address) (common.Address, error) {
	return _Bancor.Contract.CollectionByPool(&_Bancor.CallOpts, pool)
}

// DepositingEnabled is a free data retrieval call binding the contract method 0x71f43f9a.
//
// Solidity: function depositingEnabled() view returns(bool)
func (_Bancor *BancorCaller) DepositingEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "depositingEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// DepositingEnabled is a free data retrieval call binding the contract method 0x71f43f9a.
//
// Solidity: function depositingEnabled() view returns(bool)
func (_Bancor *BancorSession) DepositingEnabled() (bool, error) {
	return _Bancor.Contract.DepositingEnabled(&_Bancor.CallOpts)
}

// DepositingEnabled is a free data retrieval call binding the contract method 0x71f43f9a.
//
// Solidity: function depositingEnabled() view returns(bool)
func (_Bancor *BancorCallerSession) DepositingEnabled() (bool, error) {
	return _Bancor.Contract.DepositingEnabled(&_Bancor.CallOpts)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Bancor *BancorCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Bancor *BancorSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Bancor.Contract.GetRoleAdmin(&_Bancor.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Bancor *BancorCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Bancor.Contract.GetRoleAdmin(&_Bancor.CallOpts, role)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Bancor *BancorCaller) GetRoleMember(opts *bind.CallOpts, role [32]byte, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "getRoleMember", role, index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Bancor *BancorSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _Bancor.Contract.GetRoleMember(&_Bancor.CallOpts, role, index)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Bancor *BancorCallerSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _Bancor.Contract.GetRoleMember(&_Bancor.CallOpts, role, index)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Bancor *BancorCaller) GetRoleMemberCount(opts *bind.CallOpts, role [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "getRoleMemberCount", role)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Bancor *BancorSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _Bancor.Contract.GetRoleMemberCount(&_Bancor.CallOpts, role)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Bancor *BancorCallerSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _Bancor.Contract.GetRoleMemberCount(&_Bancor.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Bancor *BancorCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Bancor *BancorSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Bancor.Contract.HasRole(&_Bancor.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Bancor *BancorCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Bancor.Contract.HasRole(&_Bancor.CallOpts, role, account)
}

// LiquidityPools is a free data retrieval call binding the contract method 0xd6efd7c3.
//
// Solidity: function liquidityPools() view returns(address[])
func (_Bancor *BancorCaller) LiquidityPools(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "liquidityPools")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// LiquidityPools is a free data retrieval call binding the contract method 0xd6efd7c3.
//
// Solidity: function liquidityPools() view returns(address[])
func (_Bancor *BancorSession) LiquidityPools() ([]common.Address, error) {
	return _Bancor.Contract.LiquidityPools(&_Bancor.CallOpts)
}

// LiquidityPools is a free data retrieval call binding the contract method 0xd6efd7c3.
//
// Solidity: function liquidityPools() view returns(address[])
func (_Bancor *BancorCallerSession) LiquidityPools() ([]common.Address, error) {
	return _Bancor.Contract.LiquidityPools(&_Bancor.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Bancor *BancorCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Bancor *BancorSession) Paused() (bool, error) {
	return _Bancor.Contract.Paused(&_Bancor.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Bancor *BancorCallerSession) Paused() (bool, error) {
	return _Bancor.Contract.Paused(&_Bancor.CallOpts)
}

// PendingNetworkFeeAmount is a free data retrieval call binding the contract method 0x7bf6a425.
//
// Solidity: function pendingNetworkFeeAmount() view returns(uint256)
func (_Bancor *BancorCaller) PendingNetworkFeeAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "pendingNetworkFeeAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PendingNetworkFeeAmount is a free data retrieval call binding the contract method 0x7bf6a425.
//
// Solidity: function pendingNetworkFeeAmount() view returns(uint256)
func (_Bancor *BancorSession) PendingNetworkFeeAmount() (*big.Int, error) {
	return _Bancor.Contract.PendingNetworkFeeAmount(&_Bancor.CallOpts)
}

// PendingNetworkFeeAmount is a free data retrieval call binding the contract method 0x7bf6a425.
//
// Solidity: function pendingNetworkFeeAmount() view returns(uint256)
func (_Bancor *BancorCallerSession) PendingNetworkFeeAmount() (*big.Int, error) {
	return _Bancor.Contract.PendingNetworkFeeAmount(&_Bancor.CallOpts)
}

// PolRewardsPPM is a free data retrieval call binding the contract method 0x1329db29.
//
// Solidity: function polRewardsPPM() view returns(uint32)
func (_Bancor *BancorCaller) PolRewardsPPM(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "polRewardsPPM")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// PolRewardsPPM is a free data retrieval call binding the contract method 0x1329db29.
//
// Solidity: function polRewardsPPM() view returns(uint32)
func (_Bancor *BancorSession) PolRewardsPPM() (uint32, error) {
	return _Bancor.Contract.PolRewardsPPM(&_Bancor.CallOpts)
}

// PolRewardsPPM is a free data retrieval call binding the contract method 0x1329db29.
//
// Solidity: function polRewardsPPM() view returns(uint32)
func (_Bancor *BancorCallerSession) PolRewardsPPM() (uint32, error) {
	return _Bancor.Contract.PolRewardsPPM(&_Bancor.CallOpts)
}

// PoolCollections is a free data retrieval call binding the contract method 0x39fadf98.
//
// Solidity: function poolCollections() view returns(address[])
func (_Bancor *BancorCaller) PoolCollections(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "poolCollections")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// PoolCollections is a free data retrieval call binding the contract method 0x39fadf98.
//
// Solidity: function poolCollections() view returns(address[])
func (_Bancor *BancorSession) PoolCollections() ([]common.Address, error) {
	return _Bancor.Contract.PoolCollections(&_Bancor.CallOpts)
}

// PoolCollections is a free data retrieval call binding the contract method 0x39fadf98.
//
// Solidity: function poolCollections() view returns(address[])
func (_Bancor *BancorCallerSession) PoolCollections() ([]common.Address, error) {
	return _Bancor.Contract.PoolCollections(&_Bancor.CallOpts)
}

// RoleAdmin is a free data retrieval call binding the contract method 0x93867fb5.
//
// Solidity: function roleAdmin() pure returns(bytes32)
func (_Bancor *BancorCaller) RoleAdmin(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "roleAdmin")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RoleAdmin is a free data retrieval call binding the contract method 0x93867fb5.
//
// Solidity: function roleAdmin() pure returns(bytes32)
func (_Bancor *BancorSession) RoleAdmin() ([32]byte, error) {
	return _Bancor.Contract.RoleAdmin(&_Bancor.CallOpts)
}

// RoleAdmin is a free data retrieval call binding the contract method 0x93867fb5.
//
// Solidity: function roleAdmin() pure returns(bytes32)
func (_Bancor *BancorCallerSession) RoleAdmin() ([32]byte, error) {
	return _Bancor.Contract.RoleAdmin(&_Bancor.CallOpts)
}

// RoleEmergencyStopper is a free data retrieval call binding the contract method 0x41f435b3.
//
// Solidity: function roleEmergencyStopper() pure returns(bytes32)
func (_Bancor *BancorCaller) RoleEmergencyStopper(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "roleEmergencyStopper")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RoleEmergencyStopper is a free data retrieval call binding the contract method 0x41f435b3.
//
// Solidity: function roleEmergencyStopper() pure returns(bytes32)
func (_Bancor *BancorSession) RoleEmergencyStopper() ([32]byte, error) {
	return _Bancor.Contract.RoleEmergencyStopper(&_Bancor.CallOpts)
}

// RoleEmergencyStopper is a free data retrieval call binding the contract method 0x41f435b3.
//
// Solidity: function roleEmergencyStopper() pure returns(bytes32)
func (_Bancor *BancorCallerSession) RoleEmergencyStopper() ([32]byte, error) {
	return _Bancor.Contract.RoleEmergencyStopper(&_Bancor.CallOpts)
}

// RoleMigrationManager is a free data retrieval call binding the contract method 0xe6aac07e.
//
// Solidity: function roleMigrationManager() pure returns(bytes32)
func (_Bancor *BancorCaller) RoleMigrationManager(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "roleMigrationManager")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RoleMigrationManager is a free data retrieval call binding the contract method 0xe6aac07e.
//
// Solidity: function roleMigrationManager() pure returns(bytes32)
func (_Bancor *BancorSession) RoleMigrationManager() ([32]byte, error) {
	return _Bancor.Contract.RoleMigrationManager(&_Bancor.CallOpts)
}

// RoleMigrationManager is a free data retrieval call binding the contract method 0xe6aac07e.
//
// Solidity: function roleMigrationManager() pure returns(bytes32)
func (_Bancor *BancorCallerSession) RoleMigrationManager() ([32]byte, error) {
	return _Bancor.Contract.RoleMigrationManager(&_Bancor.CallOpts)
}

// RoleNetworkFeeManager is a free data retrieval call binding the contract method 0xc8447487.
//
// Solidity: function roleNetworkFeeManager() pure returns(bytes32)
func (_Bancor *BancorCaller) RoleNetworkFeeManager(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "roleNetworkFeeManager")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RoleNetworkFeeManager is a free data retrieval call binding the contract method 0xc8447487.
//
// Solidity: function roleNetworkFeeManager() pure returns(bytes32)
func (_Bancor *BancorSession) RoleNetworkFeeManager() ([32]byte, error) {
	return _Bancor.Contract.RoleNetworkFeeManager(&_Bancor.CallOpts)
}

// RoleNetworkFeeManager is a free data retrieval call binding the contract method 0xc8447487.
//
// Solidity: function roleNetworkFeeManager() pure returns(bytes32)
func (_Bancor *BancorCallerSession) RoleNetworkFeeManager() ([32]byte, error) {
	return _Bancor.Contract.RoleNetworkFeeManager(&_Bancor.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Bancor *BancorCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Bancor *BancorSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Bancor.Contract.SupportsInterface(&_Bancor.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Bancor *BancorCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Bancor.Contract.SupportsInterface(&_Bancor.CallOpts, interfaceId)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(uint16)
func (_Bancor *BancorCaller) Version(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Bancor.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(uint16)
func (_Bancor *BancorSession) Version() (uint16, error) {
	return _Bancor.Contract.Version(&_Bancor.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(uint16)
func (_Bancor *BancorCallerSession) Version() (uint16, error) {
	return _Bancor.Contract.Version(&_Bancor.CallOpts)
}

// CancelWithdrawal is a paid mutator transaction binding the contract method 0x3efcfda4.
//
// Solidity: function cancelWithdrawal(uint256 id) returns(uint256)
func (_Bancor *BancorTransactor) CancelWithdrawal(opts *bind.TransactOpts, id *big.Int) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "cancelWithdrawal", id)
}

// CancelWithdrawal is a paid mutator transaction binding the contract method 0x3efcfda4.
//
// Solidity: function cancelWithdrawal(uint256 id) returns(uint256)
func (_Bancor *BancorSession) CancelWithdrawal(id *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.CancelWithdrawal(&_Bancor.TransactOpts, id)
}

// CancelWithdrawal is a paid mutator transaction binding the contract method 0x3efcfda4.
//
// Solidity: function cancelWithdrawal(uint256 id) returns(uint256)
func (_Bancor *BancorTransactorSession) CancelWithdrawal(id *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.CancelWithdrawal(&_Bancor.TransactOpts, id)
}

// CreatePools is a paid mutator transaction binding the contract method 0x42659964.
//
// Solidity: function createPools(address[] tokens, address poolCollection) returns()
func (_Bancor *BancorTransactor) CreatePools(opts *bind.TransactOpts, tokens []common.Address, poolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "createPools", tokens, poolCollection)
}

// CreatePools is a paid mutator transaction binding the contract method 0x42659964.
//
// Solidity: function createPools(address[] tokens, address poolCollection) returns()
func (_Bancor *BancorSession) CreatePools(tokens []common.Address, poolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.CreatePools(&_Bancor.TransactOpts, tokens, poolCollection)
}

// CreatePools is a paid mutator transaction binding the contract method 0x42659964.
//
// Solidity: function createPools(address[] tokens, address poolCollection) returns()
func (_Bancor *BancorTransactorSession) CreatePools(tokens []common.Address, poolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.CreatePools(&_Bancor.TransactOpts, tokens, poolCollection)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address pool, uint256 tokenAmount) payable returns(uint256)
func (_Bancor *BancorTransactor) Deposit(opts *bind.TransactOpts, pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "deposit", pool, tokenAmount)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address pool, uint256 tokenAmount) payable returns(uint256)
func (_Bancor *BancorSession) Deposit(pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.Deposit(&_Bancor.TransactOpts, pool, tokenAmount)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address pool, uint256 tokenAmount) payable returns(uint256)
func (_Bancor *BancorTransactorSession) Deposit(pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.Deposit(&_Bancor.TransactOpts, pool, tokenAmount)
}

// DepositFor is a paid mutator transaction binding the contract method 0xb3db428b.
//
// Solidity: function depositFor(address provider, address pool, uint256 tokenAmount) payable returns(uint256)
func (_Bancor *BancorTransactor) DepositFor(opts *bind.TransactOpts, provider common.Address, pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "depositFor", provider, pool, tokenAmount)
}

// DepositFor is a paid mutator transaction binding the contract method 0xb3db428b.
//
// Solidity: function depositFor(address provider, address pool, uint256 tokenAmount) payable returns(uint256)
func (_Bancor *BancorSession) DepositFor(provider common.Address, pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.DepositFor(&_Bancor.TransactOpts, provider, pool, tokenAmount)
}

// DepositFor is a paid mutator transaction binding the contract method 0xb3db428b.
//
// Solidity: function depositFor(address provider, address pool, uint256 tokenAmount) payable returns(uint256)
func (_Bancor *BancorTransactorSession) DepositFor(provider common.Address, pool common.Address, tokenAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.DepositFor(&_Bancor.TransactOpts, provider, pool, tokenAmount)
}

// EnableDepositing is a paid mutator transaction binding the contract method 0x26e6b697.
//
// Solidity: function enableDepositing(bool status) returns()
func (_Bancor *BancorTransactor) EnableDepositing(opts *bind.TransactOpts, status bool) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "enableDepositing", status)
}

// EnableDepositing is a paid mutator transaction binding the contract method 0x26e6b697.
//
// Solidity: function enableDepositing(bool status) returns()
func (_Bancor *BancorSession) EnableDepositing(status bool) (*types.Transaction, error) {
	return _Bancor.Contract.EnableDepositing(&_Bancor.TransactOpts, status)
}

// EnableDepositing is a paid mutator transaction binding the contract method 0x26e6b697.
//
// Solidity: function enableDepositing(bool status) returns()
func (_Bancor *BancorTransactorSession) EnableDepositing(status bool) (*types.Transaction, error) {
	return _Bancor.Contract.EnableDepositing(&_Bancor.TransactOpts, status)
}

// FlashLoan is a paid mutator transaction binding the contract method 0xadf51de1.
//
// Solidity: function flashLoan(address zcntoken, uint256 amount, address recipient, bytes data) returns()
func (_Bancor *BancorTransactor) FlashLoan(opts *bind.TransactOpts, token common.Address, amount *big.Int, recipient common.Address, data []byte) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "flashLoan", token, amount, recipient, data)
}

// FlashLoan is a paid mutator transaction binding the contract method 0xadf51de1.
//
// Solidity: function flashLoan(address zcntoken, uint256 amount, address recipient, bytes data) returns()
func (_Bancor *BancorSession) FlashLoan(token common.Address, amount *big.Int, recipient common.Address, data []byte) (*types.Transaction, error) {
	return _Bancor.Contract.FlashLoan(&_Bancor.TransactOpts, token, amount, recipient, data)
}

// FlashLoan is a paid mutator transaction binding the contract method 0xadf51de1.
//
// Solidity: function flashLoan(address zcntoken, uint256 amount, address recipient, bytes data) returns()
func (_Bancor *BancorTransactorSession) FlashLoan(token common.Address, amount *big.Int, recipient common.Address, data []byte) (*types.Transaction, error) {
	return _Bancor.Contract.FlashLoan(&_Bancor.TransactOpts, token, amount, recipient, data)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Bancor *BancorTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Bancor *BancorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.GrantRole(&_Bancor.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Bancor *BancorTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.GrantRole(&_Bancor.TransactOpts, role, account)
}

// InitWithdrawal is a paid mutator transaction binding the contract method 0x357a0333.
//
// Solidity: function initWithdrawal(address poolToken, uint256 poolTokenAmount) returns(uint256)
func (_Bancor *BancorTransactor) InitWithdrawal(opts *bind.TransactOpts, poolToken common.Address, poolTokenAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "initWithdrawal", poolToken, poolTokenAmount)
}

// InitWithdrawal is a paid mutator transaction binding the contract method 0x357a0333.
//
// Solidity: function initWithdrawal(address poolToken, uint256 poolTokenAmount) returns(uint256)
func (_Bancor *BancorSession) InitWithdrawal(poolToken common.Address, poolTokenAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.InitWithdrawal(&_Bancor.TransactOpts, poolToken, poolTokenAmount)
}

// InitWithdrawal is a paid mutator transaction binding the contract method 0x357a0333.
//
// Solidity: function initWithdrawal(address poolToken, uint256 poolTokenAmount) returns(uint256)
func (_Bancor *BancorTransactorSession) InitWithdrawal(poolToken common.Address, poolTokenAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.InitWithdrawal(&_Bancor.TransactOpts, poolToken, poolTokenAmount)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address initBNTPool, address initPendingWithdrawals, address initPoolMigrator) returns()
func (_Bancor *BancorTransactor) Initialize(opts *bind.TransactOpts, initBNTPool common.Address, initPendingWithdrawals common.Address, initPoolMigrator common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "initialize", initBNTPool, initPendingWithdrawals, initPoolMigrator)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address initBNTPool, address initPendingWithdrawals, address initPoolMigrator) returns()
func (_Bancor *BancorSession) Initialize(initBNTPool common.Address, initPendingWithdrawals common.Address, initPoolMigrator common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.Initialize(&_Bancor.TransactOpts, initBNTPool, initPendingWithdrawals, initPoolMigrator)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address initBNTPool, address initPendingWithdrawals, address initPoolMigrator) returns()
func (_Bancor *BancorTransactorSession) Initialize(initBNTPool common.Address, initPendingWithdrawals common.Address, initPoolMigrator common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.Initialize(&_Bancor.TransactOpts, initBNTPool, initPendingWithdrawals, initPoolMigrator)
}

// MigrateLiquidity is a paid mutator transaction binding the contract method 0x3d1c24e7.
//
// Solidity: function migrateLiquidity(address zcntoken, address provider, uint256 amount, uint256 availableAmount, uint256 originalAmount) payable returns()
func (_Bancor *BancorTransactor) MigrateLiquidity(opts *bind.TransactOpts, token common.Address, provider common.Address, amount *big.Int, availableAmount *big.Int, originalAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "migrateLiquidity", token, provider, amount, availableAmount, originalAmount)
}

// MigrateLiquidity is a paid mutator transaction binding the contract method 0x3d1c24e7.
//
// Solidity: function migrateLiquidity(address zcntoken, address provider, uint256 amount, uint256 availableAmount, uint256 originalAmount) payable returns()
func (_Bancor *BancorSession) MigrateLiquidity(token common.Address, provider common.Address, amount *big.Int, availableAmount *big.Int, originalAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.MigrateLiquidity(&_Bancor.TransactOpts, token, provider, amount, availableAmount, originalAmount)
}

// MigrateLiquidity is a paid mutator transaction binding the contract method 0x3d1c24e7.
//
// Solidity: function migrateLiquidity(address zcntoken, address provider, uint256 amount, uint256 availableAmount, uint256 originalAmount) payable returns()
func (_Bancor *BancorTransactorSession) MigrateLiquidity(token common.Address, provider common.Address, amount *big.Int, availableAmount *big.Int, originalAmount *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.MigrateLiquidity(&_Bancor.TransactOpts, token, provider, amount, availableAmount, originalAmount)
}

// MigratePools is a paid mutator transaction binding the contract method 0xc109ba13.
//
// Solidity: function migratePools(address[] pools, address newPoolCollection) returns()
func (_Bancor *BancorTransactor) MigratePools(opts *bind.TransactOpts, pools []common.Address, newPoolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "migratePools", pools, newPoolCollection)
}

// MigratePools is a paid mutator transaction binding the contract method 0xc109ba13.
//
// Solidity: function migratePools(address[] pools, address newPoolCollection) returns()
func (_Bancor *BancorSession) MigratePools(pools []common.Address, newPoolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.MigratePools(&_Bancor.TransactOpts, pools, newPoolCollection)
}

// MigratePools is a paid mutator transaction binding the contract method 0xc109ba13.
//
// Solidity: function migratePools(address[] pools, address newPoolCollection) returns()
func (_Bancor *BancorTransactorSession) MigratePools(pools []common.Address, newPoolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.MigratePools(&_Bancor.TransactOpts, pools, newPoolCollection)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Bancor *BancorTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Bancor *BancorSession) Pause() (*types.Transaction, error) {
	return _Bancor.Contract.Pause(&_Bancor.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Bancor *BancorTransactorSession) Pause() (*types.Transaction, error) {
	return _Bancor.Contract.Pause(&_Bancor.TransactOpts)
}

// PostUpgrade is a paid mutator transaction binding the contract method 0x8cd2403d.
//
// Solidity: function postUpgrade(bytes data) returns()
func (_Bancor *BancorTransactor) PostUpgrade(opts *bind.TransactOpts, data []byte) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "postUpgrade", data)
}

// PostUpgrade is a paid mutator transaction binding the contract method 0x8cd2403d.
//
// Solidity: function postUpgrade(bytes data) returns()
func (_Bancor *BancorSession) PostUpgrade(data []byte) (*types.Transaction, error) {
	return _Bancor.Contract.PostUpgrade(&_Bancor.TransactOpts, data)
}

// PostUpgrade is a paid mutator transaction binding the contract method 0x8cd2403d.
//
// Solidity: function postUpgrade(bytes data) returns()
func (_Bancor *BancorTransactorSession) PostUpgrade(data []byte) (*types.Transaction, error) {
	return _Bancor.Contract.PostUpgrade(&_Bancor.TransactOpts, data)
}

// RegisterPoolCollection is a paid mutator transaction binding the contract method 0xa8bf9046.
//
// Solidity: function registerPoolCollection(address newPoolCollection) returns()
func (_Bancor *BancorTransactor) RegisterPoolCollection(opts *bind.TransactOpts, newPoolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "registerPoolCollection", newPoolCollection)
}

// RegisterPoolCollection is a paid mutator transaction binding the contract method 0xa8bf9046.
//
// Solidity: function registerPoolCollection(address newPoolCollection) returns()
func (_Bancor *BancorSession) RegisterPoolCollection(newPoolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.RegisterPoolCollection(&_Bancor.TransactOpts, newPoolCollection)
}

// RegisterPoolCollection is a paid mutator transaction binding the contract method 0xa8bf9046.
//
// Solidity: function registerPoolCollection(address newPoolCollection) returns()
func (_Bancor *BancorTransactorSession) RegisterPoolCollection(newPoolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.RegisterPoolCollection(&_Bancor.TransactOpts, newPoolCollection)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Bancor *BancorTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Bancor *BancorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.RenounceRole(&_Bancor.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Bancor *BancorTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.RenounceRole(&_Bancor.TransactOpts, role, account)
}

// Resume is a paid mutator transaction binding the contract method 0x046f7da2.
//
// Solidity: function resume() returns()
func (_Bancor *BancorTransactor) Resume(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "resume")
}

// Resume is a paid mutator transaction binding the contract method 0x046f7da2.
//
// Solidity: function resume() returns()
func (_Bancor *BancorSession) Resume() (*types.Transaction, error) {
	return _Bancor.Contract.Resume(&_Bancor.TransactOpts)
}

// Resume is a paid mutator transaction binding the contract method 0x046f7da2.
//
// Solidity: function resume() returns()
func (_Bancor *BancorTransactorSession) Resume() (*types.Transaction, error) {
	return _Bancor.Contract.Resume(&_Bancor.TransactOpts)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Bancor *BancorTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Bancor *BancorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.RevokeRole(&_Bancor.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Bancor *BancorTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.RevokeRole(&_Bancor.TransactOpts, role, account)
}

// SetPOLRewardsPPM is a paid mutator transaction binding the contract method 0x53300772.
//
// Solidity: function setPOLRewardsPPM(uint32 newRewardsPPM) returns()
func (_Bancor *BancorTransactor) SetPOLRewardsPPM(opts *bind.TransactOpts, newRewardsPPM uint32) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "setPOLRewardsPPM", newRewardsPPM)
}

// SetPOLRewardsPPM is a paid mutator transaction binding the contract method 0x53300772.
//
// Solidity: function setPOLRewardsPPM(uint32 newRewardsPPM) returns()
func (_Bancor *BancorSession) SetPOLRewardsPPM(newRewardsPPM uint32) (*types.Transaction, error) {
	return _Bancor.Contract.SetPOLRewardsPPM(&_Bancor.TransactOpts, newRewardsPPM)
}

// SetPOLRewardsPPM is a paid mutator transaction binding the contract method 0x53300772.
//
// Solidity: function setPOLRewardsPPM(uint32 newRewardsPPM) returns()
func (_Bancor *BancorTransactorSession) SetPOLRewardsPPM(newRewardsPPM uint32) (*types.Transaction, error) {
	return _Bancor.Contract.SetPOLRewardsPPM(&_Bancor.TransactOpts, newRewardsPPM)
}

// TradeBySourceAmount is a paid mutator transaction binding the contract method 0xd3a4acd3.
//
// Solidity: function tradeBySourceAmount(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorTransactor) TradeBySourceAmount(opts *bind.TransactOpts, sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "tradeBySourceAmount", sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeBySourceAmount is a paid mutator transaction binding the contract method 0xd3a4acd3.
//
// Solidity: function tradeBySourceAmount(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorSession) TradeBySourceAmount(sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.TradeBySourceAmount(&_Bancor.TransactOpts, sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeBySourceAmount is a paid mutator transaction binding the contract method 0xd3a4acd3.
//
// Solidity: function tradeBySourceAmount(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorTransactorSession) TradeBySourceAmount(sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.TradeBySourceAmount(&_Bancor.TransactOpts, sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeBySourceAmountArb is a paid mutator transaction binding the contract method 0xd895feee.
//
// Solidity: function tradeBySourceAmountArb(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorTransactor) TradeBySourceAmountArb(opts *bind.TransactOpts, sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "tradeBySourceAmountArb", sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeBySourceAmountArb is a paid mutator transaction binding the contract method 0xd895feee.
//
// Solidity: function tradeBySourceAmountArb(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorSession) TradeBySourceAmountArb(sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.TradeBySourceAmountArb(&_Bancor.TransactOpts, sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeBySourceAmountArb is a paid mutator transaction binding the contract method 0xd895feee.
//
// Solidity: function tradeBySourceAmountArb(address sourceToken, address targetToken, uint256 sourceAmount, uint256 minReturnAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorTransactorSession) TradeBySourceAmountArb(sourceToken common.Address, targetToken common.Address, sourceAmount *big.Int, minReturnAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.TradeBySourceAmountArb(&_Bancor.TransactOpts, sourceToken, targetToken, sourceAmount, minReturnAmount, deadline, beneficiary)
}

// TradeByTargetAmount is a paid mutator transaction binding the contract method 0x45d6602c.
//
// Solidity: function tradeByTargetAmount(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorTransactor) TradeByTargetAmount(opts *bind.TransactOpts, sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "tradeByTargetAmount", sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// TradeByTargetAmount is a paid mutator transaction binding the contract method 0x45d6602c.
//
// Solidity: function tradeByTargetAmount(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorSession) TradeByTargetAmount(sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.TradeByTargetAmount(&_Bancor.TransactOpts, sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// TradeByTargetAmount is a paid mutator transaction binding the contract method 0x45d6602c.
//
// Solidity: function tradeByTargetAmount(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorTransactorSession) TradeByTargetAmount(sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.TradeByTargetAmount(&_Bancor.TransactOpts, sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// TradeByTargetAmountArb is a paid mutator transaction binding the contract method 0xd0d14581.
//
// Solidity: function tradeByTargetAmountArb(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorTransactor) TradeByTargetAmountArb(opts *bind.TransactOpts, sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "tradeByTargetAmountArb", sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// TradeByTargetAmountArb is a paid mutator transaction binding the contract method 0xd0d14581.
//
// Solidity: function tradeByTargetAmountArb(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorSession) TradeByTargetAmountArb(sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.TradeByTargetAmountArb(&_Bancor.TransactOpts, sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// TradeByTargetAmountArb is a paid mutator transaction binding the contract method 0xd0d14581.
//
// Solidity: function tradeByTargetAmountArb(address sourceToken, address targetToken, uint256 targetAmount, uint256 maxSourceAmount, uint256 deadline, address beneficiary) payable returns(uint256)
func (_Bancor *BancorTransactorSession) TradeByTargetAmountArb(sourceToken common.Address, targetToken common.Address, targetAmount *big.Int, maxSourceAmount *big.Int, deadline *big.Int, beneficiary common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.TradeByTargetAmountArb(&_Bancor.TransactOpts, sourceToken, targetToken, targetAmount, maxSourceAmount, deadline, beneficiary)
}

// UnregisterPoolCollection is a paid mutator transaction binding the contract method 0x230df83a.
//
// Solidity: function unregisterPoolCollection(address poolCollection) returns()
func (_Bancor *BancorTransactor) UnregisterPoolCollection(opts *bind.TransactOpts, poolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "unregisterPoolCollection", poolCollection)
}

// UnregisterPoolCollection is a paid mutator transaction binding the contract method 0x230df83a.
//
// Solidity: function unregisterPoolCollection(address poolCollection) returns()
func (_Bancor *BancorSession) UnregisterPoolCollection(poolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.UnregisterPoolCollection(&_Bancor.TransactOpts, poolCollection)
}

// UnregisterPoolCollection is a paid mutator transaction binding the contract method 0x230df83a.
//
// Solidity: function unregisterPoolCollection(address poolCollection) returns()
func (_Bancor *BancorTransactorSession) UnregisterPoolCollection(poolCollection common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.UnregisterPoolCollection(&_Bancor.TransactOpts, poolCollection)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 id) returns(uint256)
func (_Bancor *BancorTransactor) Withdraw(opts *bind.TransactOpts, id *big.Int) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "withdraw", id)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 id) returns(uint256)
func (_Bancor *BancorSession) Withdraw(id *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.Withdraw(&_Bancor.TransactOpts, id)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 id) returns(uint256)
func (_Bancor *BancorTransactorSession) Withdraw(id *big.Int) (*types.Transaction, error) {
	return _Bancor.Contract.Withdraw(&_Bancor.TransactOpts, id)
}

// WithdrawNetworkFees is a paid mutator transaction binding the contract method 0x3cd11924.
//
// Solidity: function withdrawNetworkFees(address recipient) returns(uint256)
func (_Bancor *BancorTransactor) WithdrawNetworkFees(opts *bind.TransactOpts, recipient common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "withdrawNetworkFees", recipient)
}

// WithdrawNetworkFees is a paid mutator transaction binding the contract method 0x3cd11924.
//
// Solidity: function withdrawNetworkFees(address recipient) returns(uint256)
func (_Bancor *BancorSession) WithdrawNetworkFees(recipient common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.WithdrawNetworkFees(&_Bancor.TransactOpts, recipient)
}

// WithdrawNetworkFees is a paid mutator transaction binding the contract method 0x3cd11924.
//
// Solidity: function withdrawNetworkFees(address recipient) returns(uint256)
func (_Bancor *BancorTransactorSession) WithdrawNetworkFees(recipient common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.WithdrawNetworkFees(&_Bancor.TransactOpts, recipient)
}

// WithdrawPOL is a paid mutator transaction binding the contract method 0x8ffcca07.
//
// Solidity: function withdrawPOL(address pool) returns(uint256)
func (_Bancor *BancorTransactor) WithdrawPOL(opts *bind.TransactOpts, pool common.Address) (*types.Transaction, error) {
	return _Bancor.contract.Transact(opts, "withdrawPOL", pool)
}

// WithdrawPOL is a paid mutator transaction binding the contract method 0x8ffcca07.
//
// Solidity: function withdrawPOL(address pool) returns(uint256)
func (_Bancor *BancorSession) WithdrawPOL(pool common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.WithdrawPOL(&_Bancor.TransactOpts, pool)
}

// WithdrawPOL is a paid mutator transaction binding the contract method 0x8ffcca07.
//
// Solidity: function withdrawPOL(address pool) returns(uint256)
func (_Bancor *BancorTransactorSession) WithdrawPOL(pool common.Address) (*types.Transaction, error) {
	return _Bancor.Contract.WithdrawPOL(&_Bancor.TransactOpts, pool)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Bancor *BancorTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bancor.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Bancor *BancorSession) Receive() (*types.Transaction, error) {
	return _Bancor.Contract.Receive(&_Bancor.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Bancor *BancorTransactorSession) Receive() (*types.Transaction, error) {
	return _Bancor.Contract.Receive(&_Bancor.TransactOpts)
}

// BancorFlashLoanCompletedIterator is returned from FilterFlashLoanCompleted and is used to iterate over the raw logs and unpacked data for FlashLoanCompleted events raised by the Bancor contract.
type BancorFlashLoanCompletedIterator struct {
	Event *BancorFlashLoanCompleted // Event containing the contract specifics and raw log

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
func (it *BancorFlashLoanCompletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorFlashLoanCompleted)
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
		it.Event = new(BancorFlashLoanCompleted)
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
func (it *BancorFlashLoanCompletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorFlashLoanCompletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorFlashLoanCompleted represents a FlashLoanCompleted event raised by the Bancor contract.
type BancorFlashLoanCompleted struct {
	Token     common.Address
	Borrower  common.Address
	Amount    *big.Int
	FeeAmount *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterFlashLoanCompleted is a free log retrieval operation binding the contract event 0x0da3485ef1bb570df7bb888887eae5aa01d81b83cd8ccc80c0ea0922a677ecef.
//
// Solidity: event FlashLoanCompleted(address indexed zcntoken, address indexed borrower, uint256 amount, uint256 feeAmount)
func (_Bancor *BancorFilterer) FilterFlashLoanCompleted(opts *bind.FilterOpts, token []common.Address, borrower []common.Address) (*BancorFlashLoanCompletedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var borrowerRule []interface{}
	for _, borrowerItem := range borrower {
		borrowerRule = append(borrowerRule, borrowerItem)
	}

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "FlashLoanCompleted", tokenRule, borrowerRule)
	if err != nil {
		return nil, err
	}
	return &BancorFlashLoanCompletedIterator{contract: _Bancor.contract, event: "FlashLoanCompleted", logs: logs, sub: sub}, nil
}

// WatchFlashLoanCompleted is a free log subscription operation binding the contract event 0x0da3485ef1bb570df7bb888887eae5aa01d81b83cd8ccc80c0ea0922a677ecef.
//
// Solidity: event FlashLoanCompleted(address indexed zcntoken, address indexed borrower, uint256 amount, uint256 feeAmount)
func (_Bancor *BancorFilterer) WatchFlashLoanCompleted(opts *bind.WatchOpts, sink chan<- *BancorFlashLoanCompleted, token []common.Address, borrower []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var borrowerRule []interface{}
	for _, borrowerItem := range borrower {
		borrowerRule = append(borrowerRule, borrowerItem)
	}

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "FlashLoanCompleted", tokenRule, borrowerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorFlashLoanCompleted)
				if err := _Bancor.contract.UnpackLog(event, "FlashLoanCompleted", log); err != nil {
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

// ParseFlashLoanCompleted is a log parse operation binding the contract event 0x0da3485ef1bb570df7bb888887eae5aa01d81b83cd8ccc80c0ea0922a677ecef.
//
// Solidity: event FlashLoanCompleted(address indexed zcntoken, address indexed borrower, uint256 amount, uint256 feeAmount)
func (_Bancor *BancorFilterer) ParseFlashLoanCompleted(log types.Log) (*BancorFlashLoanCompleted, error) {
	event := new(BancorFlashLoanCompleted)
	if err := _Bancor.contract.UnpackLog(event, "FlashLoanCompleted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorFundsMigratedIterator is returned from FilterFundsMigrated and is used to iterate over the raw logs and unpacked data for FundsMigrated events raised by the Bancor contract.
type BancorFundsMigratedIterator struct {
	Event *BancorFundsMigrated // Event containing the contract specifics and raw log

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
func (it *BancorFundsMigratedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorFundsMigrated)
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
		it.Event = new(BancorFundsMigrated)
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
func (it *BancorFundsMigratedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorFundsMigratedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorFundsMigrated represents a FundsMigrated event raised by the Bancor contract.
type BancorFundsMigrated struct {
	ContextId       [32]byte
	Token           common.Address
	Provider        common.Address
	Amount          *big.Int
	AvailableAmount *big.Int
	OriginalAmount  *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterFundsMigrated is a free log retrieval operation binding the contract event 0x102bce4e43a6a8cf0306fde6154221c1f5460f64ba63b92b156bce998ef0db56.
//
// Solidity: event FundsMigrated(bytes32 indexed contextId, address indexed zcntoken, address indexed provider, uint256 amount, uint256 availableAmount, uint256 originalAmount)
func (_Bancor *BancorFilterer) FilterFundsMigrated(opts *bind.FilterOpts, contextId [][32]byte, token []common.Address, provider []common.Address) (*BancorFundsMigratedIterator, error) {

	var contextIdRule []interface{}
	for _, contextIdItem := range contextId {
		contextIdRule = append(contextIdRule, contextIdItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "FundsMigrated", contextIdRule, tokenRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &BancorFundsMigratedIterator{contract: _Bancor.contract, event: "FundsMigrated", logs: logs, sub: sub}, nil
}

// WatchFundsMigrated is a free log subscription operation binding the contract event 0x102bce4e43a6a8cf0306fde6154221c1f5460f64ba63b92b156bce998ef0db56.
//
// Solidity: event FundsMigrated(bytes32 indexed contextId, address indexed zcntoken, address indexed provider, uint256 amount, uint256 availableAmount, uint256 originalAmount)
func (_Bancor *BancorFilterer) WatchFundsMigrated(opts *bind.WatchOpts, sink chan<- *BancorFundsMigrated, contextId [][32]byte, token []common.Address, provider []common.Address) (event.Subscription, error) {

	var contextIdRule []interface{}
	for _, contextIdItem := range contextId {
		contextIdRule = append(contextIdRule, contextIdItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "FundsMigrated", contextIdRule, tokenRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorFundsMigrated)
				if err := _Bancor.contract.UnpackLog(event, "FundsMigrated", log); err != nil {
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

// ParseFundsMigrated is a log parse operation binding the contract event 0x102bce4e43a6a8cf0306fde6154221c1f5460f64ba63b92b156bce998ef0db56.
//
// Solidity: event FundsMigrated(bytes32 indexed contextId, address indexed zcntoken, address indexed provider, uint256 amount, uint256 availableAmount, uint256 originalAmount)
func (_Bancor *BancorFilterer) ParseFundsMigrated(log types.Log) (*BancorFundsMigrated, error) {
	event := new(BancorFundsMigrated)
	if err := _Bancor.contract.UnpackLog(event, "FundsMigrated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorNetworkFeesWithdrawnIterator is returned from FilterNetworkFeesWithdrawn and is used to iterate over the raw logs and unpacked data for NetworkFeesWithdrawn events raised by the Bancor contract.
type BancorNetworkFeesWithdrawnIterator struct {
	Event *BancorNetworkFeesWithdrawn // Event containing the contract specifics and raw log

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
func (it *BancorNetworkFeesWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorNetworkFeesWithdrawn)
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
		it.Event = new(BancorNetworkFeesWithdrawn)
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
func (it *BancorNetworkFeesWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorNetworkFeesWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorNetworkFeesWithdrawn represents a NetworkFeesWithdrawn event raised by the Bancor contract.
type BancorNetworkFeesWithdrawn struct {
	Caller    common.Address
	Recipient common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNetworkFeesWithdrawn is a free log retrieval operation binding the contract event 0x328c9cc28e75030423307e732b07659ae452a620281f3e54e838000a7f467538.
//
// Solidity: event NetworkFeesWithdrawn(address indexed caller, address indexed recipient, uint256 amount)
func (_Bancor *BancorFilterer) FilterNetworkFeesWithdrawn(opts *bind.FilterOpts, caller []common.Address, recipient []common.Address) (*BancorNetworkFeesWithdrawnIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "NetworkFeesWithdrawn", callerRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &BancorNetworkFeesWithdrawnIterator{contract: _Bancor.contract, event: "NetworkFeesWithdrawn", logs: logs, sub: sub}, nil
}

// WatchNetworkFeesWithdrawn is a free log subscription operation binding the contract event 0x328c9cc28e75030423307e732b07659ae452a620281f3e54e838000a7f467538.
//
// Solidity: event NetworkFeesWithdrawn(address indexed caller, address indexed recipient, uint256 amount)
func (_Bancor *BancorFilterer) WatchNetworkFeesWithdrawn(opts *bind.WatchOpts, sink chan<- *BancorNetworkFeesWithdrawn, caller []common.Address, recipient []common.Address) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "NetworkFeesWithdrawn", callerRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorNetworkFeesWithdrawn)
				if err := _Bancor.contract.UnpackLog(event, "NetworkFeesWithdrawn", log); err != nil {
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

// ParseNetworkFeesWithdrawn is a log parse operation binding the contract event 0x328c9cc28e75030423307e732b07659ae452a620281f3e54e838000a7f467538.
//
// Solidity: event NetworkFeesWithdrawn(address indexed caller, address indexed recipient, uint256 amount)
func (_Bancor *BancorFilterer) ParseNetworkFeesWithdrawn(log types.Log) (*BancorNetworkFeesWithdrawn, error) {
	event := new(BancorNetworkFeesWithdrawn)
	if err := _Bancor.contract.UnpackLog(event, "NetworkFeesWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorPOLRewardsPPMUpdatedIterator is returned from FilterPOLRewardsPPMUpdated and is used to iterate over the raw logs and unpacked data for POLRewardsPPMUpdated events raised by the Bancor contract.
type BancorPOLRewardsPPMUpdatedIterator struct {
	Event *BancorPOLRewardsPPMUpdated // Event containing the contract specifics and raw log

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
func (it *BancorPOLRewardsPPMUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorPOLRewardsPPMUpdated)
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
		it.Event = new(BancorPOLRewardsPPMUpdated)
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
func (it *BancorPOLRewardsPPMUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorPOLRewardsPPMUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorPOLRewardsPPMUpdated represents a POLRewardsPPMUpdated event raised by the Bancor contract.
type BancorPOLRewardsPPMUpdated struct {
	OldRewardsPPM uint32
	NewRewardsPPM uint32
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterPOLRewardsPPMUpdated is a free log retrieval operation binding the contract event 0xa159b13d7eac36d9a65034b4fd6ace1d9cb070d063dc950c564a266f4d091802.
//
// Solidity: event POLRewardsPPMUpdated(uint32 oldRewardsPPM, uint32 newRewardsPPM)
func (_Bancor *BancorFilterer) FilterPOLRewardsPPMUpdated(opts *bind.FilterOpts) (*BancorPOLRewardsPPMUpdatedIterator, error) {

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "POLRewardsPPMUpdated")
	if err != nil {
		return nil, err
	}
	return &BancorPOLRewardsPPMUpdatedIterator{contract: _Bancor.contract, event: "POLRewardsPPMUpdated", logs: logs, sub: sub}, nil
}

// WatchPOLRewardsPPMUpdated is a free log subscription operation binding the contract event 0xa159b13d7eac36d9a65034b4fd6ace1d9cb070d063dc950c564a266f4d091802.
//
// Solidity: event POLRewardsPPMUpdated(uint32 oldRewardsPPM, uint32 newRewardsPPM)
func (_Bancor *BancorFilterer) WatchPOLRewardsPPMUpdated(opts *bind.WatchOpts, sink chan<- *BancorPOLRewardsPPMUpdated) (event.Subscription, error) {

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "POLRewardsPPMUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorPOLRewardsPPMUpdated)
				if err := _Bancor.contract.UnpackLog(event, "POLRewardsPPMUpdated", log); err != nil {
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

// ParsePOLRewardsPPMUpdated is a log parse operation binding the contract event 0xa159b13d7eac36d9a65034b4fd6ace1d9cb070d063dc950c564a266f4d091802.
//
// Solidity: event POLRewardsPPMUpdated(uint32 oldRewardsPPM, uint32 newRewardsPPM)
func (_Bancor *BancorFilterer) ParsePOLRewardsPPMUpdated(log types.Log) (*BancorPOLRewardsPPMUpdated, error) {
	event := new(BancorPOLRewardsPPMUpdated)
	if err := _Bancor.contract.UnpackLog(event, "POLRewardsPPMUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorPOLWithdrawnIterator is returned from FilterPOLWithdrawn and is used to iterate over the raw logs and unpacked data for POLWithdrawn events raised by the Bancor contract.
type BancorPOLWithdrawnIterator struct {
	Event *BancorPOLWithdrawn // Event containing the contract specifics and raw log

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
func (it *BancorPOLWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorPOLWithdrawn)
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
		it.Event = new(BancorPOLWithdrawn)
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
func (it *BancorPOLWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorPOLWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorPOLWithdrawn represents a POLWithdrawn event raised by the Bancor contract.
type BancorPOLWithdrawn struct {
	Caller         common.Address
	Token          common.Address
	PolTokenAmount *big.Int
	UserReward     *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterPOLWithdrawn is a free log retrieval operation binding the contract event 0x5ad7a2184454b6259cd118e4041a953dc9d6498302bbe528e4f967bed9197129.
//
// Solidity: event POLWithdrawn(address indexed caller, address indexed zcntoken, uint256 polTokenAmount, uint256 userReward)
func (_Bancor *BancorFilterer) FilterPOLWithdrawn(opts *bind.FilterOpts, caller []common.Address, token []common.Address) (*BancorPOLWithdrawnIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "POLWithdrawn", callerRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &BancorPOLWithdrawnIterator{contract: _Bancor.contract, event: "POLWithdrawn", logs: logs, sub: sub}, nil
}

// WatchPOLWithdrawn is a free log subscription operation binding the contract event 0x5ad7a2184454b6259cd118e4041a953dc9d6498302bbe528e4f967bed9197129.
//
// Solidity: event POLWithdrawn(address indexed caller, address indexed zcntoken, uint256 polTokenAmount, uint256 userReward)
func (_Bancor *BancorFilterer) WatchPOLWithdrawn(opts *bind.WatchOpts, sink chan<- *BancorPOLWithdrawn, caller []common.Address, token []common.Address) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "POLWithdrawn", callerRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorPOLWithdrawn)
				if err := _Bancor.contract.UnpackLog(event, "POLWithdrawn", log); err != nil {
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

// ParsePOLWithdrawn is a log parse operation binding the contract event 0x5ad7a2184454b6259cd118e4041a953dc9d6498302bbe528e4f967bed9197129.
//
// Solidity: event POLWithdrawn(address indexed caller, address indexed zcntoken, uint256 polTokenAmount, uint256 userReward)
func (_Bancor *BancorFilterer) ParsePOLWithdrawn(log types.Log) (*BancorPOLWithdrawn, error) {
	event := new(BancorPOLWithdrawn)
	if err := _Bancor.contract.UnpackLog(event, "POLWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the Bancor contract.
type BancorPausedIterator struct {
	Event *BancorPaused // Event containing the contract specifics and raw log

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
func (it *BancorPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorPaused)
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
		it.Event = new(BancorPaused)
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
func (it *BancorPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorPaused represents a Paused event raised by the Bancor contract.
type BancorPaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Bancor *BancorFilterer) FilterPaused(opts *bind.FilterOpts) (*BancorPausedIterator, error) {

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &BancorPausedIterator{contract: _Bancor.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Bancor *BancorFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *BancorPaused) (event.Subscription, error) {

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorPaused)
				if err := _Bancor.contract.UnpackLog(event, "Paused", log); err != nil {
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

// ParsePaused is a log parse operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Bancor *BancorFilterer) ParsePaused(log types.Log) (*BancorPaused, error) {
	event := new(BancorPaused)
	if err := _Bancor.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorPoolAddedIterator is returned from FilterPoolAdded and is used to iterate over the raw logs and unpacked data for PoolAdded events raised by the Bancor contract.
type BancorPoolAddedIterator struct {
	Event *BancorPoolAdded // Event containing the contract specifics and raw log

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
func (it *BancorPoolAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorPoolAdded)
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
		it.Event = new(BancorPoolAdded)
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
func (it *BancorPoolAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorPoolAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorPoolAdded represents a PoolAdded event raised by the Bancor contract.
type BancorPoolAdded struct {
	Pool           common.Address
	PoolCollection common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterPoolAdded is a free log retrieval operation binding the contract event 0x95f865c2808f8b2a85eea2611db7843150ee7835ef1403f9755918a97d76933c.
//
// Solidity: event PoolAdded(address indexed pool, address indexed poolCollection)
func (_Bancor *BancorFilterer) FilterPoolAdded(opts *bind.FilterOpts, pool []common.Address, poolCollection []common.Address) (*BancorPoolAddedIterator, error) {

	var poolRule []interface{}
	for _, poolItem := range pool {
		poolRule = append(poolRule, poolItem)
	}
	var poolCollectionRule []interface{}
	for _, poolCollectionItem := range poolCollection {
		poolCollectionRule = append(poolCollectionRule, poolCollectionItem)
	}

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "PoolAdded", poolRule, poolCollectionRule)
	if err != nil {
		return nil, err
	}
	return &BancorPoolAddedIterator{contract: _Bancor.contract, event: "PoolAdded", logs: logs, sub: sub}, nil
}

// WatchPoolAdded is a free log subscription operation binding the contract event 0x95f865c2808f8b2a85eea2611db7843150ee7835ef1403f9755918a97d76933c.
//
// Solidity: event PoolAdded(address indexed pool, address indexed poolCollection)
func (_Bancor *BancorFilterer) WatchPoolAdded(opts *bind.WatchOpts, sink chan<- *BancorPoolAdded, pool []common.Address, poolCollection []common.Address) (event.Subscription, error) {

	var poolRule []interface{}
	for _, poolItem := range pool {
		poolRule = append(poolRule, poolItem)
	}
	var poolCollectionRule []interface{}
	for _, poolCollectionItem := range poolCollection {
		poolCollectionRule = append(poolCollectionRule, poolCollectionItem)
	}

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "PoolAdded", poolRule, poolCollectionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorPoolAdded)
				if err := _Bancor.contract.UnpackLog(event, "PoolAdded", log); err != nil {
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

// ParsePoolAdded is a log parse operation binding the contract event 0x95f865c2808f8b2a85eea2611db7843150ee7835ef1403f9755918a97d76933c.
//
// Solidity: event PoolAdded(address indexed pool, address indexed poolCollection)
func (_Bancor *BancorFilterer) ParsePoolAdded(log types.Log) (*BancorPoolAdded, error) {
	event := new(BancorPoolAdded)
	if err := _Bancor.contract.UnpackLog(event, "PoolAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorPoolCollectionAddedIterator is returned from FilterPoolCollectionAdded and is used to iterate over the raw logs and unpacked data for PoolCollectionAdded events raised by the Bancor contract.
type BancorPoolCollectionAddedIterator struct {
	Event *BancorPoolCollectionAdded // Event containing the contract specifics and raw log

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
func (it *BancorPoolCollectionAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorPoolCollectionAdded)
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
		it.Event = new(BancorPoolCollectionAdded)
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
func (it *BancorPoolCollectionAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorPoolCollectionAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorPoolCollectionAdded represents a PoolCollectionAdded event raised by the Bancor contract.
type BancorPoolCollectionAdded struct {
	PoolType       uint16
	PoolCollection common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterPoolCollectionAdded is a free log retrieval operation binding the contract event 0x5ae87719d73cb0fabb219f0e4b6e0a614ed7506f8a08bdb20bebf313573151b7.
//
// Solidity: event PoolCollectionAdded(uint16 indexed poolType, address indexed poolCollection)
func (_Bancor *BancorFilterer) FilterPoolCollectionAdded(opts *bind.FilterOpts, poolType []uint16, poolCollection []common.Address) (*BancorPoolCollectionAddedIterator, error) {

	var poolTypeRule []interface{}
	for _, poolTypeItem := range poolType {
		poolTypeRule = append(poolTypeRule, poolTypeItem)
	}
	var poolCollectionRule []interface{}
	for _, poolCollectionItem := range poolCollection {
		poolCollectionRule = append(poolCollectionRule, poolCollectionItem)
	}

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "PoolCollectionAdded", poolTypeRule, poolCollectionRule)
	if err != nil {
		return nil, err
	}
	return &BancorPoolCollectionAddedIterator{contract: _Bancor.contract, event: "PoolCollectionAdded", logs: logs, sub: sub}, nil
}

// WatchPoolCollectionAdded is a free log subscription operation binding the contract event 0x5ae87719d73cb0fabb219f0e4b6e0a614ed7506f8a08bdb20bebf313573151b7.
//
// Solidity: event PoolCollectionAdded(uint16 indexed poolType, address indexed poolCollection)
func (_Bancor *BancorFilterer) WatchPoolCollectionAdded(opts *bind.WatchOpts, sink chan<- *BancorPoolCollectionAdded, poolType []uint16, poolCollection []common.Address) (event.Subscription, error) {

	var poolTypeRule []interface{}
	for _, poolTypeItem := range poolType {
		poolTypeRule = append(poolTypeRule, poolTypeItem)
	}
	var poolCollectionRule []interface{}
	for _, poolCollectionItem := range poolCollection {
		poolCollectionRule = append(poolCollectionRule, poolCollectionItem)
	}

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "PoolCollectionAdded", poolTypeRule, poolCollectionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorPoolCollectionAdded)
				if err := _Bancor.contract.UnpackLog(event, "PoolCollectionAdded", log); err != nil {
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

// ParsePoolCollectionAdded is a log parse operation binding the contract event 0x5ae87719d73cb0fabb219f0e4b6e0a614ed7506f8a08bdb20bebf313573151b7.
//
// Solidity: event PoolCollectionAdded(uint16 indexed poolType, address indexed poolCollection)
func (_Bancor *BancorFilterer) ParsePoolCollectionAdded(log types.Log) (*BancorPoolCollectionAdded, error) {
	event := new(BancorPoolCollectionAdded)
	if err := _Bancor.contract.UnpackLog(event, "PoolCollectionAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorPoolCollectionRemovedIterator is returned from FilterPoolCollectionRemoved and is used to iterate over the raw logs and unpacked data for PoolCollectionRemoved events raised by the Bancor contract.
type BancorPoolCollectionRemovedIterator struct {
	Event *BancorPoolCollectionRemoved // Event containing the contract specifics and raw log

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
func (it *BancorPoolCollectionRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorPoolCollectionRemoved)
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
		it.Event = new(BancorPoolCollectionRemoved)
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
func (it *BancorPoolCollectionRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorPoolCollectionRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorPoolCollectionRemoved represents a PoolCollectionRemoved event raised by the Bancor contract.
type BancorPoolCollectionRemoved struct {
	PoolType       uint16
	PoolCollection common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterPoolCollectionRemoved is a free log retrieval operation binding the contract event 0xa0c1e3924f995e5ba38f53b4effb6d4b3eeb84176a2951c589115140f638ac09.
//
// Solidity: event PoolCollectionRemoved(uint16 indexed poolType, address indexed poolCollection)
func (_Bancor *BancorFilterer) FilterPoolCollectionRemoved(opts *bind.FilterOpts, poolType []uint16, poolCollection []common.Address) (*BancorPoolCollectionRemovedIterator, error) {

	var poolTypeRule []interface{}
	for _, poolTypeItem := range poolType {
		poolTypeRule = append(poolTypeRule, poolTypeItem)
	}
	var poolCollectionRule []interface{}
	for _, poolCollectionItem := range poolCollection {
		poolCollectionRule = append(poolCollectionRule, poolCollectionItem)
	}

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "PoolCollectionRemoved", poolTypeRule, poolCollectionRule)
	if err != nil {
		return nil, err
	}
	return &BancorPoolCollectionRemovedIterator{contract: _Bancor.contract, event: "PoolCollectionRemoved", logs: logs, sub: sub}, nil
}

// WatchPoolCollectionRemoved is a free log subscription operation binding the contract event 0xa0c1e3924f995e5ba38f53b4effb6d4b3eeb84176a2951c589115140f638ac09.
//
// Solidity: event PoolCollectionRemoved(uint16 indexed poolType, address indexed poolCollection)
func (_Bancor *BancorFilterer) WatchPoolCollectionRemoved(opts *bind.WatchOpts, sink chan<- *BancorPoolCollectionRemoved, poolType []uint16, poolCollection []common.Address) (event.Subscription, error) {

	var poolTypeRule []interface{}
	for _, poolTypeItem := range poolType {
		poolTypeRule = append(poolTypeRule, poolTypeItem)
	}
	var poolCollectionRule []interface{}
	for _, poolCollectionItem := range poolCollection {
		poolCollectionRule = append(poolCollectionRule, poolCollectionItem)
	}

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "PoolCollectionRemoved", poolTypeRule, poolCollectionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorPoolCollectionRemoved)
				if err := _Bancor.contract.UnpackLog(event, "PoolCollectionRemoved", log); err != nil {
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

// ParsePoolCollectionRemoved is a log parse operation binding the contract event 0xa0c1e3924f995e5ba38f53b4effb6d4b3eeb84176a2951c589115140f638ac09.
//
// Solidity: event PoolCollectionRemoved(uint16 indexed poolType, address indexed poolCollection)
func (_Bancor *BancorFilterer) ParsePoolCollectionRemoved(log types.Log) (*BancorPoolCollectionRemoved, error) {
	event := new(BancorPoolCollectionRemoved)
	if err := _Bancor.contract.UnpackLog(event, "PoolCollectionRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorPoolCreatedIterator is returned from FilterPoolCreated and is used to iterate over the raw logs and unpacked data for PoolCreated events raised by the Bancor contract.
type BancorPoolCreatedIterator struct {
	Event *BancorPoolCreated // Event containing the contract specifics and raw log

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
func (it *BancorPoolCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorPoolCreated)
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
		it.Event = new(BancorPoolCreated)
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
func (it *BancorPoolCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorPoolCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorPoolCreated represents a PoolCreated event raised by the Bancor contract.
type BancorPoolCreated struct {
	Pool           common.Address
	PoolCollection common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterPoolCreated is a free log retrieval operation binding the contract event 0x4f2ce4e40f623ca765fc0167a25cb7842ceaafb8d82d3dec26ca0d0e0d2d4896.
//
// Solidity: event PoolCreated(address indexed pool, address indexed poolCollection)
func (_Bancor *BancorFilterer) FilterPoolCreated(opts *bind.FilterOpts, pool []common.Address, poolCollection []common.Address) (*BancorPoolCreatedIterator, error) {

	var poolRule []interface{}
	for _, poolItem := range pool {
		poolRule = append(poolRule, poolItem)
	}
	var poolCollectionRule []interface{}
	for _, poolCollectionItem := range poolCollection {
		poolCollectionRule = append(poolCollectionRule, poolCollectionItem)
	}

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "PoolCreated", poolRule, poolCollectionRule)
	if err != nil {
		return nil, err
	}
	return &BancorPoolCreatedIterator{contract: _Bancor.contract, event: "PoolCreated", logs: logs, sub: sub}, nil
}

// WatchPoolCreated is a free log subscription operation binding the contract event 0x4f2ce4e40f623ca765fc0167a25cb7842ceaafb8d82d3dec26ca0d0e0d2d4896.
//
// Solidity: event PoolCreated(address indexed pool, address indexed poolCollection)
func (_Bancor *BancorFilterer) WatchPoolCreated(opts *bind.WatchOpts, sink chan<- *BancorPoolCreated, pool []common.Address, poolCollection []common.Address) (event.Subscription, error) {

	var poolRule []interface{}
	for _, poolItem := range pool {
		poolRule = append(poolRule, poolItem)
	}
	var poolCollectionRule []interface{}
	for _, poolCollectionItem := range poolCollection {
		poolCollectionRule = append(poolCollectionRule, poolCollectionItem)
	}

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "PoolCreated", poolRule, poolCollectionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorPoolCreated)
				if err := _Bancor.contract.UnpackLog(event, "PoolCreated", log); err != nil {
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

// ParsePoolCreated is a log parse operation binding the contract event 0x4f2ce4e40f623ca765fc0167a25cb7842ceaafb8d82d3dec26ca0d0e0d2d4896.
//
// Solidity: event PoolCreated(address indexed pool, address indexed poolCollection)
func (_Bancor *BancorFilterer) ParsePoolCreated(log types.Log) (*BancorPoolCreated, error) {
	event := new(BancorPoolCreated)
	if err := _Bancor.contract.UnpackLog(event, "PoolCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorPoolRemovedIterator is returned from FilterPoolRemoved and is used to iterate over the raw logs and unpacked data for PoolRemoved events raised by the Bancor contract.
type BancorPoolRemovedIterator struct {
	Event *BancorPoolRemoved // Event containing the contract specifics and raw log

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
func (it *BancorPoolRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorPoolRemoved)
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
		it.Event = new(BancorPoolRemoved)
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
func (it *BancorPoolRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorPoolRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorPoolRemoved represents a PoolRemoved event raised by the Bancor contract.
type BancorPoolRemoved struct {
	Pool           common.Address
	PoolCollection common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterPoolRemoved is a free log retrieval operation binding the contract event 0x987eb3c2f78454541205f72f34839b434c306c9eaf4922efd7c0c3060fdb2e4c.
//
// Solidity: event PoolRemoved(address indexed pool, address indexed poolCollection)
func (_Bancor *BancorFilterer) FilterPoolRemoved(opts *bind.FilterOpts, pool []common.Address, poolCollection []common.Address) (*BancorPoolRemovedIterator, error) {

	var poolRule []interface{}
	for _, poolItem := range pool {
		poolRule = append(poolRule, poolItem)
	}
	var poolCollectionRule []interface{}
	for _, poolCollectionItem := range poolCollection {
		poolCollectionRule = append(poolCollectionRule, poolCollectionItem)
	}

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "PoolRemoved", poolRule, poolCollectionRule)
	if err != nil {
		return nil, err
	}
	return &BancorPoolRemovedIterator{contract: _Bancor.contract, event: "PoolRemoved", logs: logs, sub: sub}, nil
}

// WatchPoolRemoved is a free log subscription operation binding the contract event 0x987eb3c2f78454541205f72f34839b434c306c9eaf4922efd7c0c3060fdb2e4c.
//
// Solidity: event PoolRemoved(address indexed pool, address indexed poolCollection)
func (_Bancor *BancorFilterer) WatchPoolRemoved(opts *bind.WatchOpts, sink chan<- *BancorPoolRemoved, pool []common.Address, poolCollection []common.Address) (event.Subscription, error) {

	var poolRule []interface{}
	for _, poolItem := range pool {
		poolRule = append(poolRule, poolItem)
	}
	var poolCollectionRule []interface{}
	for _, poolCollectionItem := range poolCollection {
		poolCollectionRule = append(poolCollectionRule, poolCollectionItem)
	}

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "PoolRemoved", poolRule, poolCollectionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorPoolRemoved)
				if err := _Bancor.contract.UnpackLog(event, "PoolRemoved", log); err != nil {
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

// ParsePoolRemoved is a log parse operation binding the contract event 0x987eb3c2f78454541205f72f34839b434c306c9eaf4922efd7c0c3060fdb2e4c.
//
// Solidity: event PoolRemoved(address indexed pool, address indexed poolCollection)
func (_Bancor *BancorFilterer) ParsePoolRemoved(log types.Log) (*BancorPoolRemoved, error) {
	event := new(BancorPoolRemoved)
	if err := _Bancor.contract.UnpackLog(event, "PoolRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the Bancor contract.
type BancorRoleAdminChangedIterator struct {
	Event *BancorRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *BancorRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorRoleAdminChanged)
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
		it.Event = new(BancorRoleAdminChanged)
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
func (it *BancorRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorRoleAdminChanged represents a RoleAdminChanged event raised by the Bancor contract.
type BancorRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Bancor *BancorFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*BancorRoleAdminChangedIterator, error) {

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

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &BancorRoleAdminChangedIterator{contract: _Bancor.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Bancor *BancorFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *BancorRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorRoleAdminChanged)
				if err := _Bancor.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_Bancor *BancorFilterer) ParseRoleAdminChanged(log types.Log) (*BancorRoleAdminChanged, error) {
	event := new(BancorRoleAdminChanged)
	if err := _Bancor.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the Bancor contract.
type BancorRoleGrantedIterator struct {
	Event *BancorRoleGranted // Event containing the contract specifics and raw log

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
func (it *BancorRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorRoleGranted)
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
		it.Event = new(BancorRoleGranted)
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
func (it *BancorRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorRoleGranted represents a RoleGranted event raised by the Bancor contract.
type BancorRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Bancor *BancorFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BancorRoleGrantedIterator, error) {

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

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BancorRoleGrantedIterator{contract: _Bancor.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Bancor *BancorFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *BancorRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorRoleGranted)
				if err := _Bancor.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_Bancor *BancorFilterer) ParseRoleGranted(log types.Log) (*BancorRoleGranted, error) {
	event := new(BancorRoleGranted)
	if err := _Bancor.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the Bancor contract.
type BancorRoleRevokedIterator struct {
	Event *BancorRoleRevoked // Event containing the contract specifics and raw log

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
func (it *BancorRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorRoleRevoked)
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
		it.Event = new(BancorRoleRevoked)
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
func (it *BancorRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorRoleRevoked represents a RoleRevoked event raised by the Bancor contract.
type BancorRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Bancor *BancorFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BancorRoleRevokedIterator, error) {

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

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BancorRoleRevokedIterator{contract: _Bancor.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Bancor *BancorFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *BancorRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorRoleRevoked)
				if err := _Bancor.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_Bancor *BancorFilterer) ParseRoleRevoked(log types.Log) (*BancorRoleRevoked, error) {
	event := new(BancorRoleRevoked)
	if err := _Bancor.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorTokensTradedIterator is returned from FilterTokensTraded and is used to iterate over the raw logs and unpacked data for TokensTraded events raised by the Bancor contract.
type BancorTokensTradedIterator struct {
	Event *BancorTokensTraded // Event containing the contract specifics and raw log

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
func (it *BancorTokensTradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorTokensTraded)
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
		it.Event = new(BancorTokensTraded)
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
func (it *BancorTokensTradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorTokensTradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorTokensTraded represents a TokensTraded event raised by the Bancor contract.
type BancorTokensTraded struct {
	ContextId       [32]byte
	SourceToken     common.Address
	TargetToken     common.Address
	SourceAmount    *big.Int
	TargetAmount    *big.Int
	BntAmount       *big.Int
	TargetFeeAmount *big.Int
	BntFeeAmount    *big.Int
	Trader          common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterTokensTraded is a free log retrieval operation binding the contract event 0x5c02c2bb2d1d082317eb23916ca27b3e7c294398b60061a2ad54f1c3c018c318.
//
// Solidity: event TokensTraded(bytes32 indexed contextId, address indexed sourceToken, address indexed targetToken, uint256 sourceAmount, uint256 targetAmount, uint256 bntAmount, uint256 targetFeeAmount, uint256 bntFeeAmount, address trader)
func (_Bancor *BancorFilterer) FilterTokensTraded(opts *bind.FilterOpts, contextId [][32]byte, sourceToken []common.Address, targetToken []common.Address) (*BancorTokensTradedIterator, error) {

	var contextIdRule []interface{}
	for _, contextIdItem := range contextId {
		contextIdRule = append(contextIdRule, contextIdItem)
	}
	var sourceTokenRule []interface{}
	for _, sourceTokenItem := range sourceToken {
		sourceTokenRule = append(sourceTokenRule, sourceTokenItem)
	}
	var targetTokenRule []interface{}
	for _, targetTokenItem := range targetToken {
		targetTokenRule = append(targetTokenRule, targetTokenItem)
	}

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "TokensTraded", contextIdRule, sourceTokenRule, targetTokenRule)
	if err != nil {
		return nil, err
	}
	return &BancorTokensTradedIterator{contract: _Bancor.contract, event: "TokensTraded", logs: logs, sub: sub}, nil
}

// WatchTokensTraded is a free log subscription operation binding the contract event 0x5c02c2bb2d1d082317eb23916ca27b3e7c294398b60061a2ad54f1c3c018c318.
//
// Solidity: event TokensTraded(bytes32 indexed contextId, address indexed sourceToken, address indexed targetToken, uint256 sourceAmount, uint256 targetAmount, uint256 bntAmount, uint256 targetFeeAmount, uint256 bntFeeAmount, address trader)
func (_Bancor *BancorFilterer) WatchTokensTraded(opts *bind.WatchOpts, sink chan<- *BancorTokensTraded, contextId [][32]byte, sourceToken []common.Address, targetToken []common.Address) (event.Subscription, error) {

	var contextIdRule []interface{}
	for _, contextIdItem := range contextId {
		contextIdRule = append(contextIdRule, contextIdItem)
	}
	var sourceTokenRule []interface{}
	for _, sourceTokenItem := range sourceToken {
		sourceTokenRule = append(sourceTokenRule, sourceTokenItem)
	}
	var targetTokenRule []interface{}
	for _, targetTokenItem := range targetToken {
		targetTokenRule = append(targetTokenRule, targetTokenItem)
	}

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "TokensTraded", contextIdRule, sourceTokenRule, targetTokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorTokensTraded)
				if err := _Bancor.contract.UnpackLog(event, "TokensTraded", log); err != nil {
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

// ParseTokensTraded is a log parse operation binding the contract event 0x5c02c2bb2d1d082317eb23916ca27b3e7c294398b60061a2ad54f1c3c018c318.
//
// Solidity: event TokensTraded(bytes32 indexed contextId, address indexed sourceToken, address indexed targetToken, uint256 sourceAmount, uint256 targetAmount, uint256 bntAmount, uint256 targetFeeAmount, uint256 bntFeeAmount, address trader)
func (_Bancor *BancorFilterer) ParseTokensTraded(log types.Log) (*BancorTokensTraded, error) {
	event := new(BancorTokensTraded)
	if err := _Bancor.contract.UnpackLog(event, "TokensTraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BancorUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the Bancor contract.
type BancorUnpausedIterator struct {
	Event *BancorUnpaused // Event containing the contract specifics and raw log

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
func (it *BancorUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BancorUnpaused)
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
		it.Event = new(BancorUnpaused)
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
func (it *BancorUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BancorUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BancorUnpaused represents a Unpaused event raised by the Bancor contract.
type BancorUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Bancor *BancorFilterer) FilterUnpaused(opts *bind.FilterOpts) (*BancorUnpausedIterator, error) {

	logs, sub, err := _Bancor.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &BancorUnpausedIterator{contract: _Bancor.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Bancor *BancorFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *BancorUnpaused) (event.Subscription, error) {

	logs, sub, err := _Bancor.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BancorUnpaused)
				if err := _Bancor.contract.UnpackLog(event, "Unpaused", log); err != nil {
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

// ParseUnpaused is a log parse operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Bancor *BancorFilterer) ParseUnpaused(log types.Log) (*BancorUnpaused, error) {
	event := new(BancorUnpaused)
	if err := _Bancor.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
