//go:build !mobile
// +build !mobile

package zcncore

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/block"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/node"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
)

// Provider represents the type of provider.
type Provider int

const (
	ProviderMiner Provider = iota + 1
	ProviderSharder
	ProviderBlobber
	ProviderValidator
	ProviderAuthorizer
)

type TransactionVelocity = float64

// Transaction velocity vs cost factor
// TODO: Pass it to miner to calculate real time factor
const (
	RegularTransaction TransactionVelocity = 1.0
	FastTransaction    TransactionVelocity = 1.3
	FasterTransaction  TransactionVelocity = 1.6
)

type ConfirmationStatus int

const (
	Undefined ConfirmationStatus = iota
	Success

	// ChargeableError is an error that still charges the user for the transaction.
	ChargeableError
)

type Miner struct {
	ID         string      `json:"id"`
	N2NHost    string      `json:"n2n_host"`
	Host       string      `json:"host"`
	Port       int         `json:"port"`
	PublicKey  string      `json:"public_key"`
	ShortName  string      `json:"short_name"`
	BuildTag   string      `json:"build_tag"`
	TotalStake int64       `json:"total_stake"`
	Stat       interface{} `json:"stat"`
}

// Node represents a node (miner or sharder) in the network.
type Node struct {
	Miner     Miner `json:"simple_miner"`
	StakePool `json:"stake_pool"`
}

// MinerSCNodes list of nodes registered to the miner smart contract
type MinerSCNodes struct {
	Nodes []Node `json:"Nodes"`
}

type VestingSCConfig struct {
	MinLock              common.Balance `json:"min_lock"`
	MinDuration          time.Duration  `json:"min_duration"`
	MaxDuration          time.Duration  `json:"max_duration"`
	MaxDestinations      int            `json:"max_destinations"`
	MaxDescriptionLength int            `json:"max_description_length"`
}

type DelegatePool struct {
	Balance      int64  `json:"balance"`
	Reward       int64  `json:"reward"`
	Status       int    `json:"status"`
	RoundCreated int64  `json:"round_created"` // used for cool down
	DelegateID   string `json:"delegate_id"`
}

type StakePool struct {
	Pools    map[string]*DelegatePool `json:"pools"`
	Reward   int64                    `json:"rewards"`
	Settings StakePoolSettings        `json:"settings"`
	Minter   int                      `json:"minter"`
}

type stakePoolRequest struct {
	ProviderType Provider `json:"provider_type,omitempty"`
	ProviderID   string   `json:"provider_id,omitempty"`
}

type MinerSCDelegatePoolInfo struct {
	ID         common.Key     `json:"id"`
	Balance    common.Balance `json:"balance"`
	Reward     common.Balance `json:"reward"`      // uncollected reread
	RewardPaid common.Balance `json:"reward_paid"` // total reward all time
	Status     string         `json:"status"`
}

// MinerSCUserPoolsInfo represents the user stake pools information
type MinerSCUserPoolsInfo struct {
	Pools map[string][]*MinerSCDelegatePoolInfo `json:"pools"`
}

type TransactionCommon interface {
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input interface{}, val uint64, feeOpts ...FeeOption) (*transaction.Transaction, error)
	// Send implements sending token to a given clientid
	Send(toClientID string, val uint64, desc string) error

	//RegisterMultiSig registers a group wallet and subwallets with MultisigSC
	RegisterMultiSig(walletstr, mswallet string) error

	VestingAdd(ar *VestingAddRequest, value uint64) error

	MinerSCLock(providerId string, providerType Provider, lock uint64) error
	MinerSCUnlock(providerId string, providerType Provider) error
	MinerSCCollectReward(providerID string, providerType Provider) error
	MinerSCKill(providerID string, providerType Provider) error

	StorageSCCollectReward(providerID string, providerType Provider) error

	FinalizeAllocation(allocID string) error
	CancelAllocation(allocID string) error
	CreateAllocation(car *CreateAllocationRequest, lock uint64) error //
	CreateReadPool() error
	ReadPoolLock(allocID string, blobberID string, duration int64, lock uint64) error
	ReadPoolUnlock() error
	StakePoolLock(providerId string, providerType Provider, lock uint64) error
	StakePoolUnlock(providerId string, providerType Provider) error
	UpdateBlobberSettings(blobber *Blobber) error
	UpdateValidatorSettings(validator *Validator) error
	UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock uint64) error
	WritePoolLock(allocID string, blobberID string, duration int64, lock uint64) error
	WritePoolUnlock(allocID string) error

	VestingUpdateConfig(*InputMap) error
	MinerScUpdateConfig(*InputMap) error
	MinerScUpdateGlobals(*InputMap) error
	StorageScUpdateConfig(*InputMap) error
	AddHardfork(ip *InputMap) (err error)
	FaucetUpdateConfig(*InputMap) error
	ZCNSCUpdateGlobalConfig(*InputMap) error

	MinerSCMinerSettings(*MinerSCMinerInfo) error
	MinerSCSharderSettings(*MinerSCMinerInfo) error
	MinerSCDeleteMiner(*MinerSCMinerInfo) error
	MinerSCDeleteSharder(*MinerSCMinerInfo) error

	// ZCNSCUpdateAuthorizerConfig updates authorizer config by ID
	ZCNSCUpdateAuthorizerConfig(*AuthorizerNode) error
	// ZCNSCAddAuthorizer adds authorizer
	ZCNSCAddAuthorizer(*AddAuthorizerPayload) error

	// ZCNSCAuthorizerHealthCheck provides health check for authorizer
	ZCNSCAuthorizerHealthCheck(*AuthorizerHealthCheckPayload) error

	// GetVerifyConfirmationStatus implements the verification status from sharders
	GetVerifyConfirmationStatus() ConfirmationStatus

	// ZCNSCDeleteAuthorizer deletes authorizer
	ZCNSCDeleteAuthorizer(*DeleteAuthorizerPayload) error

	ZCNSCCollectReward(providerID string, providerType Provider) error
}

// compiler time check
var (
	_ TransactionScheme = (*Transaction)(nil)
	_ TransactionScheme = (*TransactionWithAuth)(nil)
)

// TransactionScheme implements few methods for block chain.
//
// Note: to be buildable on MacOSX all arguments should have names.
type TransactionScheme interface {
	TransactionCommon
	// SetTransactionCallback implements storing the callback
	// used to call after the transaction or verification is completed
	SetTransactionCallback(cb TransactionCallback) error
	// StoreData implements store the data to blockchain
	StoreData(data string) error
	// ExecuteFaucetSCWallet implements the `Faucet Smart contract` for a given wallet
	ExecuteFaucetSCWallet(walletStr string, methodName string, input []byte) error
	// GetTransactionHash implements retrieval of hash of the submitted transaction
	GetTransactionHash() string
	// SetTransactionHash implements verify a previous transaction status
	SetTransactionHash(hash string) error
	// SetTransactionNonce implements method to set the transaction nonce
	SetTransactionNonce(txnNonce int64) error
	// Verify implements verify the transaction
	Verify() error
	// GetVerifyOutput implements the verification output from sharders
	GetVerifyOutput() string
	// GetTransactionError implements error string in case of transaction failure
	GetTransactionError() string
	// GetVerifyError implements error string in case of verify failure error
	GetVerifyError() string
	// GetTransactionNonce returns nonce
	GetTransactionNonce() int64

	// Output of transaction.
	Output() []byte

	// Hash Transaction status regardless of status
	Hash() string

	// Vesting SC

	VestingTrigger(poolID string) error
	VestingStop(sr *VestingStopRequest) error
	VestingUnlock(poolID string) error
	VestingDelete(poolID string) error

	// Miner SC
}

// PriceRange represents a price range allowed by user to filter blobbers.
type PriceRange struct {
	Min common.Balance `json:"min"`
	Max common.Balance `json:"max"`
}

// CreateAllocationRequest is information to create allocation.
type CreateAllocationRequest struct {
	DataShards      int              `json:"data_shards"`
	ParityShards    int              `json:"parity_shards"`
	Size            common.Size      `json:"size"`
	Expiration      common.Timestamp `json:"expiration_date"`
	Owner           string           `json:"owner_id"`
	OwnerPublicKey  string           `json:"owner_public_key"`
	Blobbers        []string         `json:"blobbers"`
	ReadPriceRange  PriceRange       `json:"read_price_range"`
	WritePriceRange PriceRange       `json:"write_price_range"`
}

type StakePoolSettings struct {
	DelegateWallet *string  `json:"delegate_wallet,omitempty"`
	NumDelegates   *int     `json:"num_delegates,omitempty"`
	ServiceCharge  *float64 `json:"service_charge,omitempty"`
}

type Terms struct {
	ReadPrice        common.Balance `json:"read_price"`  // tokens / GB
	WritePrice       common.Balance `json:"write_price"` // tokens / GB `
	MaxOfferDuration time.Duration  `json:"max_offer_duration"`
}

// Blobber represents a blobber node.
type Blobber struct {
	// ID is the blobber ID.
	ID common.Key `json:"id"`
	// BaseURL is the blobber's base URL used to access the blobber
	BaseURL string `json:"url"`
	// Terms of storage service of the blobber (read/write price, max offer duration)
	Terms Terms `json:"terms"`
	// Capacity is the total capacity of the blobber
	Capacity common.Size `json:"capacity"`
	// Used is the capacity of the blobber used to create allocations
	Allocated common.Size `json:"allocated"`
	// LastHealthCheck is the last time the blobber was checked for health
	LastHealthCheck common.Timestamp `json:"last_health_check"`
	// StakePoolSettings is the settings of the blobber's stake pool
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
	// NotAvailable is true if the blobber is not available
	NotAvailable bool `json:"not_available"`
	// IsRestricted is true if the blobber is restricted.
	// Restricted blobbers needs to be authenticated using AuthTickets in order to be used for allocation creation.
	// Check Restricted Blobbers documentation for more details.
	IsRestricted bool `json:"is_restricted"`
}

type Validator struct {
	ID                common.Key        `json:"id"`
	BaseURL           string            `json:"url"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

// AddAuthorizerPayload represents the payload for adding an authorizer.
type AddAuthorizerPayload struct {
	PublicKey         string                      `json:"public_key"`
	URL               string                      `json:"url"`
	StakePoolSettings AuthorizerStakePoolSettings `json:"stake_pool_settings"` // Used to initially create stake pool
}

// DeleteAuthorizerPayload represents the payload for deleting an authorizer.
type DeleteAuthorizerPayload struct {
	ID string `json:"id"` // authorizer ID
}

// AuthorizerHealthCheckPayload represents the payload for authorizer health check.
type AuthorizerHealthCheckPayload struct {
	ID string `json:"id"` // authorizer ID
}

// AuthorizerStakePoolSettings represents the settings for an authorizer's stake pool.
type AuthorizerStakePoolSettings struct {
	DelegateWallet string  `json:"delegate_wallet"`
	NumDelegates   int     `json:"num_delegates"`
	ServiceCharge  float64 `json:"service_charge"`
}

type AuthorizerConfig struct {
	Fee common.Balance `json:"fee"`
}

// InputMap represents a map of input fields.
type InputMap struct {
	Fields map[string]string `json:"Fields"`
}

func newTransaction(cb TransactionCallback, txnFee uint64, nonce int64) (*Transaction, error) {
	t := &Transaction{}
	t.txn = transaction.NewTransactionEntity(_config.wallet.ClientID, _config.chain.ChainID, _config.wallet.ClientKey, nonce)
	t.txnStatus, t.verifyStatus = StatusUnknown, StatusUnknown
	t.txnCb = cb
	t.txn.TransactionNonce = nonce
	t.txn.TransactionFee = txnFee
	return t, nil
}

// NewTransaction new generic transaction object for any operation
//   - cb: callback for transaction state
//   - txnFee: Transaction fees (in SAS tokens)
//   - nonce: latest nonce of current wallet. please set it with 0 if you don't know the latest value
func NewTransaction(cb TransactionCallback, txnFee uint64, nonce int64) (TransactionScheme, error) {
	err := CheckConfig()
	if err != nil {
		return nil, err
	}
	if _config.isSplitWallet {
		logging.Info("New transaction interface with auth")
		return newTransactionWithAuth(cb, txnFee, nonce)
	}

	logging.Info("New transaction interface")
	return newTransaction(cb, txnFee, nonce)
}

func (t *Transaction) createSmartContractTxn(address, methodName string, input interface{}, value uint64, opts ...FeeOption) error {
	sn := transaction.SmartContractTxnData{Name: methodName, InputArgs: input}
	snBytes, err := json.Marshal(sn)
	if err != nil {
		return errors.Wrap(err, "create smart contract failed due to invalid data")
	}

	t.txn.TransactionType = transaction.TxnTypeSmartContract
	t.txn.ToClientID = address
	t.txn.TransactionData = string(snBytes)
	t.txn.Value = value

	if t.txn.TransactionFee > 0 {
		return nil
	}

	tf := &TxnFeeOption{}
	for _, opt := range opts {
		opt(tf)
	}

	if tf.noEstimateFee {
		return nil
	}

	// TODO: check if transaction is exempt to avoid unnecessary fee estimation
	minFee, err := transaction.EstimateFee(t.txn, _config.chain.Miners, 0.2)
	if err != nil {
		return err
	}

	t.txn.TransactionFee = minFee

	return nil
}

func (t *Transaction) createFaucetSCWallet(walletStr string, methodName string, input []byte) (*zcncrypto.Wallet, error) {
	w, err := getWallet(walletStr)
	if err != nil {
		fmt.Printf("Error while parsing the wallet. %v\n", err)
		return nil, err
	}
	err = t.createSmartContractTxn(FaucetSmartContractAddress, methodName, input, 0)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (t *Transaction) ExecuteSmartContract(address, methodName string, input interface{}, val uint64, opts ...FeeOption) (*transaction.Transaction, error) {
	err := t.createSmartContractTxn(address, methodName, input, val, opts...)
	if err != nil {
		return nil, err
	}
	go func() {
		t.setNonceAndSubmit()
	}()
	return t.txn, nil
}

func (t *Transaction) Send(toClientID string, val uint64, desc string) error {
	txnData, err := json.Marshal(transaction.SmartContractTxnData{Name: "transfer", InputArgs: SendTxnData{Note: desc}})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}

	t.txn.TransactionType = transaction.TxnTypeSend
	t.txn.ToClientID = toClientID
	t.txn.Value = val
	t.txn.TransactionData = string(txnData)
	if t.txn.TransactionFee == 0 {
		fee, err := transaction.EstimateFee(t.txn, _config.chain.Miners, 0.2)
		if err != nil {
			return err
		}
		t.txn.TransactionFee = fee
	}

	go func() {
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) SendWithSignatureHash(toClientID string, val uint64, desc string, sig string, CreationDate int64, hash string) error {
	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	t.txn.TransactionType = transaction.TxnTypeSend
	t.txn.ToClientID = toClientID
	t.txn.Value = val
	t.txn.Hash = hash
	t.txn.TransactionData = string(txnData)
	t.txn.Signature = sig
	t.txn.CreationDate = CreationDate
	if t.txn.TransactionFee == 0 {
		fee, err := transaction.EstimateFee(t.txn, _config.chain.Miners, 0.2)
		if err != nil {
			return err
		}
		t.txn.TransactionFee = fee
	}

	go func() {
		t.setNonceAndSubmit()
	}()
	return nil
}

type VestingDest struct {
	ID     string         `json:"id"`     // destination ID
	Amount common.Balance `json:"amount"` // amount to vest for the destination
}

type VestingAddRequest struct {
	Description  string           `json:"description"`  // allow empty
	StartTime    common.Timestamp `json:"start_time"`   //
	Duration     time.Duration    `json:"duration"`     //
	Destinations []*VestingDest   `json:"destinations"` //
}

func (t *Transaction) VestingAdd(ar *VestingAddRequest, value uint64) (
	err error) {

	err = t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_ADD, ar, value)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) VestingStop(sr *VestingStopRequest) (err error) {
	err = t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_STOP, sr, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) vestingPoolTxn(function string, poolID string,
	value uint64) error {

	return t.createSmartContractTxn(VestingSmartContractAddress,
		function, vestingRequest{PoolID: common.Key(poolID)}, value)
}

func (t *Transaction) VestingTrigger(poolID string) (err error) {

	err = t.vestingPoolTxn(transaction.VESTING_TRIGGER, poolID, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) VestingUnlock(poolID string) (err error) {
	err = t.vestingPoolTxn(transaction.VESTING_UNLOCK, poolID, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) VestingDelete(poolID string) (err error) {
	err = t.vestingPoolTxn(transaction.VESTING_DELETE, poolID, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCLock(providerId string, providerType Provider, lock uint64) error {
	if lock > math.MaxInt64 {
		return errors.New("invalid_lock", "int64 overflow on lock value")
	}

	pr := &stakePoolRequest{
		ProviderID:   providerId,
		ProviderType: providerType,
	}
	err := t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, pr, lock)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return err
}
func (t *Transaction) MinerSCUnlock(providerId string, providerType Provider) error {
	pr := &stakePoolRequest{
		ProviderID:   providerId,
		ProviderType: providerType,
	}
	err := t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UNLOCK, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return err
}

func (t *Transaction) MinerSCCollectReward(providerId string, providerType Provider) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
	}
	err := t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_COLLECT_REWARD, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return err
}

func (t *Transaction) MinerSCKill(providerId string, providerType Provider) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
	}
	var name string
	switch providerType {
	case ProviderMiner:
		name = transaction.MINERSC_KILL_MINER
	case ProviderSharder:
		name = transaction.MINERSC_KILL_SHARDER
	default:
		return fmt.Errorf("kill provider type %v not implimented", providerType)
	}

	err := t.createSmartContractTxn(MinerSmartContractAddress, name, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return err
}

func (t *Transaction) StorageSCCollectReward(providerId string, providerType Provider) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
	}
	err := t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_COLLECT_REWARD, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go t.setNonceAndSubmit()
	return err
}

// FinalizeAllocation transaction.
func (t *Transaction) FinalizeAllocation(allocID string) (
	err error) {

	type finiRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_FINALIZE_ALLOCATION, &finiRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		logging.Error(err)
		return
	}

	go func() { t.setNonceAndSubmit() }()
	return
}

// CancelAllocation transaction.
func (t *Transaction) CancelAllocation(allocID string) (
	err error) {

	type cancelRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CANCEL_ALLOCATION, &cancelRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// CreateAllocation transaction.
func (t *Transaction) CreateAllocation(car *CreateAllocationRequest,
	lock uint64) (err error) {
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_ALLOCATION, car, lock)
	if err != nil {
		logging.Error(err)
		return
	}

	go func() { t.setNonceAndSubmit() }()
	return
}

// CreateReadPool for current user.
func (t *Transaction) CreateReadPool() (err error) {
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_READ_POOL, nil, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// ReadPoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) ReadPoolLock(allocID, blobberID string, duration int64, lock uint64) (err error) {
	if lock > math.MaxInt64 {
		return errors.New("invalid_lock", "int64 overflow on lock value")
	}

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_LOCK, &lr, lock)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// ReadPoolUnlock for current user and given pool.
func (t *Transaction) ReadPoolUnlock() (err error) {
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_UNLOCK, nil, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// StakePoolLock used to lock tokens in a stake pool of a blobber.
func (t *Transaction) StakePoolLock(providerId string, providerType Provider, lock uint64) error {

	if lock > math.MaxInt64 {
		return errors.New("invalid_lock", "int64 overflow on lock value")
	}

	type stakePoolRequest struct {
		ProviderType Provider `json:"provider_type,omitempty"`
		ProviderID   string   `json:"provider_id,omitempty"`
	}

	spr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerId,
	}

	err := t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_LOCK, &spr, lock)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// StakePoolUnlock by blobberID and poolID.
func (t *Transaction) StakePoolUnlock(providerId string, providerType Provider) error {

	type stakePoolRequest struct {
		ProviderType Provider `json:"provider_type,omitempty"`
		ProviderID   string   `json:"provider_id,omitempty"`
	}

	spr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerId,
	}

	err := t.createSmartContractTxn(StorageSmartContractAddress, transaction.STORAGESC_STAKE_POOL_UNLOCK, &spr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// UpdateBlobberSettings update settings of a blobber.
func (t *Transaction) UpdateBlobberSettings(b *Blobber) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, b, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// UpdateAllocation transaction.
func (t *Transaction) UpdateAllocation(allocID string, sizeDiff int64,
	expirationDiff int64, lock uint64) (err error) {

	if lock > math.MaxInt64 {
		return errors.New("invalid_lock", "int64 overflow on lock value")
	}

	type updateAllocationRequest struct {
		ID         string `json:"id"`              // allocation id
		Size       int64  `json:"size"`            // difference
		Expiration int64  `json:"expiration_date"` // difference
	}

	var uar updateAllocationRequest
	uar.ID = allocID
	uar.Size = sizeDiff
	uar.Expiration = expirationDiff

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_ALLOCATION, &uar, lock)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// WritePoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) WritePoolLock(allocID, blobberID string, duration int64,
	lock uint64) (err error) {

	if lock > math.MaxInt64 {
		return errors.New("invalid_lock", "int64 overflow on lock value")
	}

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_LOCK, &lr, lock)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// WritePoolUnlock for current user and given pool.
func (t *Transaction) WritePoolUnlock(allocID string) (
	err error) {

	var ur = struct {
		AllocationID string `json:"allocation_id"`
	}{
		AllocationID: allocID,
	}

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_UNLOCK, &ur, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) VestingUpdateConfig(vscc *InputMap) (err error) {

	err = t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_UPDATE_SETTINGS, vscc, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// faucet smart contract

func (t *Transaction) FaucetUpdateConfig(ip *InputMap) (err error) {

	err = t.createSmartContractTxn(FaucetSmartContractAddress,
		transaction.FAUCETSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

//
// miner SC
//

func (t *Transaction) MinerScUpdateConfig(ip *InputMap) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerScUpdateGlobals(ip *InputMap) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_GLOBALS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) StorageScUpdateConfig(ip *InputMap) (err error) {
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}
func (t *Transaction) AddHardfork(ip *InputMap) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.ADD_HARDFORK, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) ZCNSCUpdateGlobalConfig(ip *InputMap) (err error) {
	err = t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_GLOBAL_CONFIG, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go t.setNonceAndSubmit()
	return
}

func (t *Transaction) GetVerifyConfirmationStatus() ConfirmationStatus {
	return ConfirmationStatus(t.verifyConfirmationStatus)
}

// RegisterMultiSig register a multisig wallet with the SC.
func (t *Transaction) RegisterMultiSig(walletstr string, mswallet string) error {
	w, err := GetWallet(walletstr)
	if err != nil {
		fmt.Printf("Error while parsing the wallet. %v\n", err)
		return err
	}

	msw, err := GetMultisigPayload(mswallet)
	if err != nil {
		fmt.Printf("\nError in registering. %v\n", err)
		return err
	}
	sn := transaction.SmartContractTxnData{Name: MultiSigRegisterFuncName, InputArgs: msw}
	snBytes, err := json.Marshal(sn)
	if err != nil {
		return errors.Wrap(err, "execute multisig register failed due to invalid data.")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSmartContract
		t.txn.ToClientID = MultiSigSmartContractAddress
		t.txn.TransactionData = string(snBytes)
		t.txn.Value = 0
		nonce := t.txn.TransactionNonce
		if nonce < 1 {
			nonce = node.Cache.GetNextNonce(t.txn.ClientID)
		} else {
			node.Cache.Set(t.txn.ClientID, nonce)
		}
		t.txn.TransactionNonce = nonce

		if t.txn.TransactionFee == 0 {
			fee, err := transaction.EstimateFee(t.txn, _config.chain.Miners, 0.2)
			if err != nil {
				return
			}
			t.txn.TransactionFee = fee
		}

		err = t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
		if err != nil {
			return
		}
		t.submitTxn()
	}()
	return nil
}

// NewMSTransaction new transaction object for multisig operation
func NewMSTransaction(walletstr string, cb TransactionCallback) (*Transaction, error) {
	w, err := GetWallet(walletstr)
	if err != nil {
		fmt.Printf("Error while parsing the wallet. %v", err)
		return nil, err
	}
	t := &Transaction{}
	t.txn = transaction.NewTransactionEntity(w.ClientID, _config.chain.ChainID, w.ClientKey, w.Nonce)
	t.txnStatus, t.verifyStatus = StatusUnknown, StatusUnknown
	t.txnCb = cb
	return t, nil
}

// RegisterVote register a multisig wallet with the SC.
func (t *Transaction) RegisterVote(signerwalletstr string, msvstr string) error {

	w, err := GetWallet(signerwalletstr)
	if err != nil {
		fmt.Printf("Error while parsing the wallet. %v", err)
		return err
	}

	msv, err := GetMultisigVotePayload(msvstr)

	if err != nil {
		fmt.Printf("\nError in voting. %v\n", err)
		return err
	}
	sn := transaction.SmartContractTxnData{Name: MultiSigVoteFuncName, InputArgs: msv}
	snBytes, err := json.Marshal(sn)
	if err != nil {
		return errors.Wrap(err, "execute multisig vote failed due to invalid data.")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSmartContract
		t.txn.ToClientID = MultiSigSmartContractAddress
		t.txn.TransactionData = string(snBytes)
		t.txn.Value = 0
		nonce := t.txn.TransactionNonce
		if nonce < 1 {
			nonce = node.Cache.GetNextNonce(t.txn.ClientID)
		} else {
			node.Cache.Set(t.txn.ClientID, nonce)
		}
		t.txn.TransactionNonce = nonce

		if t.txn.TransactionFee == 0 {
			fee, err := transaction.EstimateFee(t.txn, _config.chain.Miners, 0.2)
			if err != nil {
				return
			}
			t.txn.TransactionFee = fee
		}

		err = t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
		if err != nil {
			return
		}
		t.submitTxn()
	}()
	return nil
}

type MinerSCDelegatePool struct {
	Settings StakePoolSettings `json:"settings"`
}

// SimpleMiner represents a node in the network, miner or sharder.
type SimpleMiner struct {
	ID string `json:"id"`
}

// MinerSCMinerInfo interface for miner/sharder info functions on miner smart contract.
type MinerSCMinerInfo struct {
	SimpleMiner         `json:"simple_miner"`
	MinerSCDelegatePool `json:"stake_pool"`
}

func (t *Transaction) MinerSCMinerSettings(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_SETTINGS, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCSharderSettings(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_SETTINGS, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCDeleteMiner(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_DELETE, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCDeleteSharder(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_DELETE, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// AuthorizerNode represents an authorizer node in the network
type AuthorizerNode struct {
	ID     string            `json:"id"`
	URL    string            `json:"url"`
	Config *AuthorizerConfig `json:"config"`
}

func (t *Transaction) ZCNSCUpdateAuthorizerConfig(ip *AuthorizerNode) (err error) {
	err = t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_AUTHORIZER_CONFIG, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go t.setNonceAndSubmit()
	return
}

func (t *Transaction) Verify() error {
	if t.txnHash == "" && t.txnStatus == StatusUnknown {
		return errors.New("", "invalid transaction. cannot be verified.")
	}
	if t.txnHash == "" && t.txnStatus == StatusSuccess {
		h := t.GetTransactionHash()
		if h == "" {
			node.Cache.Evict(t.txn.ClientID)
			return errors.New("", "invalid transaction. cannot be verified.")
		}
	}
	// If transaction is verify only start from current time
	if t.txn.CreationDate == 0 {
		t.txn.CreationDate = int64(common.Now())
	}

	tq, err := NewTransactionQuery(Sharders.Healthy(), _config.chain.Miners)
	if err != nil {
		logging.Error(err)
		return err
	}

	go func() {

		for {

			tq.Reset()
			// Get transaction confirmationBlock from a random sharder
			confirmBlockHeader, confirmationBlock, lfbBlockHeader, err := tq.getFastConfirmation(context.TODO(), t.txnHash)
			if err != nil {
				now := int64(common.Now())

				// maybe it is a network or server error
				if lfbBlockHeader == nil {
					logging.Info(err, " now: ", now)
				} else {
					logging.Info(err, " now: ", now, ", LFB creation time:", lfbBlockHeader.CreationDate)
				}

				// transaction is done or expired. it means random sharder might be outdated, try to query it from s/S sharders to confirm it
				if util.MaxInt64(lfbBlockHeader.getCreationDate(now), now) >= (t.txn.CreationDate + int64(defaultTxnExpirationSeconds)) {
					logging.Info("falling back to ", getMinShardersVerify(), " of ", len(_config.chain.Sharders), " Sharders")
					confirmBlockHeader, confirmationBlock, lfbBlockHeader, err = tq.getConsensusConfirmation(context.TODO(), getMinShardersVerify(), t.txnHash)
				}

				// txn not found in fast confirmation/consensus confirmation
				if err != nil {

					if lfbBlockHeader == nil {
						// no any valid lfb on all sharders. maybe they are network/server errors. try it again
						continue
					}

					// it is expired
					if t.isTransactionExpired(lfbBlockHeader.getCreationDate(now), now) {
						t.completeVerify(StatusError, "", errors.New("", `{"error": "verify transaction failed"}`))
						return
					}
					continue
				}
			}

			valid := validateChain(confirmBlockHeader)
			if valid {
				output, err := json.Marshal(confirmationBlock)
				if err != nil {
					t.completeVerify(StatusError, "", errors.New("", `{"error": "transaction confirmation json marshal error"`))
					return
				}
				confJson := confirmationBlock["confirmation"]

				var conf map[string]json.RawMessage
				if err := json.Unmarshal(confJson, &conf); err != nil {
					return
				}
				txnJson := conf["txn"]

				tt := transaction.Transaction{}
				if err := json.Unmarshal(txnJson, &tt); err != nil {
					return
				}

				*t.txn = tt
				txStatus := tt.Status

				switch txStatus {
				case 1:
					t.completeVerifyWithConStatus(StatusSuccess, int(Success), string(output), nil)
				case 2:
					t.completeVerifyWithConStatus(StatusSuccess, int(ChargeableError), tt.TransactionOutput, nil)
				default:
					t.completeVerify(StatusError, string(output), nil)
				}
				return
			}
		}
	}()
	return nil
}

// ConvertToValue converts ZCN tokens to SAS tokens
// # Inputs
//   - token: ZCN tokens
func ConvertToValue(token float64) uint64 {
	return uint64(token * common.TokenUnit)
}

func GetLatestFinalized(ctx context.Context, numSharders int) (b *block.Header, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(Sharders.Healthy()) // overwrite, use all
	Sharders.QueryFromShardersContext(ctx, numSharders, GET_LATEST_FINALIZED, result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	for i := 0; i < numSharders; i++ {
		var rsp = <-result
		if rsp == nil {
			logging.Error("nil response")
			continue
		}

		logging.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			logging.Error(rsp.Body)
			continue
		}

		if err = json.Unmarshal([]byte(rsp.Body), &b); err != nil {
			logging.Error("block parse error: ", err)
			err = nil
			continue
		}

		var h = encryption.FastHash([]byte(b.Hash))
		if roundConsensus[h]++; roundConsensus[h] > maxConsensus {
			maxConsensus = roundConsensus[h]
		}
	}

	if maxConsensus == 0 {
		return nil, errors.New("", "block info not found")
	}

	return
}

// GetLatestFinalizedMagicBlock gets latest finalized magic block
//   - numSharders: number of sharders
//   - timeout: request timeout
func GetLatestFinalizedMagicBlock(ctx context.Context, numSharders int) (m *block.MagicBlock, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(Sharders.Healthy()) // overwrite, use all
	Sharders.QueryFromShardersContext(ctx, numSharders, GET_LATEST_FINALIZED_MAGIC_BLOCK, result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	type respObj struct {
		MagicBlock *block.MagicBlock `json:"magic_block"`
	}

	for i := 0; i < numSharders; i++ {
		var rsp = <-result
		if rsp == nil {
			logging.Error("nil response")
			continue
		}

		logging.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			logging.Error(rsp.Body)
			continue
		}

		var respo respObj
		if err = json.Unmarshal([]byte(rsp.Body), &respo); err != nil {
			logging.Error(" magic block parse error: ", err)
			err = nil
			continue
		}

		m = respo.MagicBlock
		var h = encryption.FastHash([]byte(respo.MagicBlock.Hash))
		if roundConsensus[h]++; roundConsensus[h] > maxConsensus {
			maxConsensus = roundConsensus[h]
		}
	}

	if maxConsensus == 0 {
		return nil, errors.New("", "magic block info not found")
	}

	return
}

func GetChainStats(ctx context.Context) (b *block.ChainStats, err error) {
	var result = make(chan *util.GetResponse, 1)
	defer close(result)

	var numSharders = len(Sharders.Healthy()) // overwrite, use all
	Sharders.QueryFromShardersContext(ctx, numSharders, GET_CHAIN_STATS, result)
	var rsp *util.GetResponse
	for i := 0; i < numSharders; i++ {
		var x = <-result
		if x == nil {
			logging.Error("nil response")
			continue
		}
		if x.StatusCode != http.StatusOK {
			continue
		}
		rsp = x
	}

	if rsp == nil {
		return nil, errors.New("http_request_failed", "Request failed with status not 200")
	}

	if err = json.Unmarshal([]byte(rsp.Body), &b); err != nil {
		return nil, err
	}
	return
}

func GetFeeStats(ctx context.Context) (b *block.FeeStats, err error) {

	var numMiners = 4

	if numMiners > len(_config.chain.Miners) {
		numMiners = len(_config.chain.Miners)
	}

	var result = make(chan *util.GetResponse, numMiners)

	queryFromMinersContext(ctx, numMiners, GET_FEE_STATS, result)
	var rsp *util.GetResponse

loop:
	for i := 0; i < numMiners; i++ {
		select {
		case x := <-result:
			if x.StatusCode != http.StatusOK {
				continue
			}
			rsp = x
			if rsp != nil {
				break loop
			}
		case <-ctx.Done():
			err = ctx.Err()
			return nil, err
		}
	}
	if rsp == nil {
		return nil, errors.New("http_request_failed", "Request failed with status not 200")
	}
	if err = json.Unmarshal([]byte(rsp.Body), &b); err != nil {
		return nil, err
	}
	return
}

func GetBlockByRound(ctx context.Context, numSharders int, round int64) (b *block.Block, err error) {
	return Sharders.GetBlockByRound(ctx, numSharders, round)
}

// GetRoundFromSharders returns the current round number from the sharders
func GetRoundFromSharders() (int64, error) {
	return Sharders.GetRoundFromSharders()
}

// GetHardForkRound returns the round number of the hard fork
//   - hardFork: hard fork name
func GetHardForkRound(hardFork string) (int64, error) {
	return Sharders.GetHardForkRound(hardFork)
}

func GetMagicBlockByNumber(ctx context.Context, numSharders int, number int64) (m *block.MagicBlock, err error) {

	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(Sharders.Healthy()) // overwrite, use all
	Sharders.QueryFromShardersContext(ctx, numSharders,
		fmt.Sprintf("%smagic_block_number=%d", GET_MAGIC_BLOCK_INFO, number),
		result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	type respObj struct {
		MagicBlock *block.MagicBlock `json:"magic_block"`
	}

	for i := 0; i < numSharders; i++ {
		var rsp = <-result
		if rsp == nil {
			logging.Error("nil response")
			continue
		}
		logging.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			logging.Error(rsp.Body)
			continue
		}

		var respo respObj
		if err = json.Unmarshal([]byte(rsp.Body), &respo); err != nil {
			logging.Error(" magic block parse error: ", err)
			err = nil
			continue
		}

		m = respo.MagicBlock
		var h = encryption.FastHash([]byte(respo.MagicBlock.Hash))
		if roundConsensus[h]++; roundConsensus[h] > maxConsensus {
			maxConsensus = roundConsensus[h]
		}
	}

	if maxConsensus == 0 {
		return nil, errors.New("", "magic block info not found")
	}

	return
}

type NonceCache struct {
	cache map[string]int64
	guard sync.Mutex
}

func NewNonceCache() *NonceCache {
	return &NonceCache{cache: make(map[string]int64)}
}

func (nc *NonceCache) GetNextNonce(clientId string) int64 {
	nc.guard.Lock()
	defer nc.guard.Unlock()
	if _, ok := nc.cache[clientId]; !ok {
		back := &getNonceCallBack{
			nonceCh: make(chan int64),
			err:     nil,
		}
		if err := GetNonce(back); err != nil {
			return 0
		}

		timeout, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		select {
		case n := <-back.nonceCh:
			if back.err != nil {
				return 0
			}
			nc.cache[clientId] = n
		case <-timeout.Done():
			return 0
		}
	}

	nc.cache[clientId] += 1
	return nc.cache[clientId]
}

func (nc *NonceCache) Set(clientId string, nonce int64) {
	nc.guard.Lock()
	defer nc.guard.Unlock()
	nc.cache[clientId] = nonce
}

func (nc *NonceCache) Evict(clientId string) {
	nc.guard.Lock()
	defer nc.guard.Unlock()
	delete(nc.cache, clientId)
}

// WithEthereumNode set the ethereum node used for bridge operations.
//   - uri: ethereum node uri
func WithEthereumNode(uri string) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.EthNode = uri
		return nil
	}
}

// WithChainID set the chain id. Chain ID is a unique identifier for the blockchain which is set at the time of its creation.
//   - id: chain id
func WithChainID(id string) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.ChainID = id
		return nil
	}
}

// WithMinSubmit set the minimum submit, minimum number of miners that should receive the transaction submission.
//   - m: minimum submit
func WithMinSubmit(m int) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.MinSubmit = m
		return nil
	}
}

// WithMinConfirmation set the minimum confirmation, minimum number of nodes that should confirm the transaction.
//   - m: minimum confirmation
func WithMinConfirmation(m int) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.MinConfirmation = m
		return nil
	}
}

// WithConfirmationChainLength set the confirmation chain length, which is the number of blocks that need to be confirmed.
//   - m: confirmation chain length
func WithConfirmationChainLength(m int) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.ConfirmationChainLength = m
		return nil
	}
}

func WithSharderConsensous(m int) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.SharderConsensous = m
		return nil
	}
}

func WithIsSplitWallet(v bool) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.IsSplitWallet = v
		return nil
	}
}

// UpdateValidatorSettings update settings of a validator.
func (t *Transaction) UpdateValidatorSettings(v *Validator) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_VALIDATOR_SETTINGS, v, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

type VestingClientList struct {
	Pools []common.Key `json:"pools"`
}

func GetVestingClientList(clientID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	if clientID == "" {
		clientID = _config.wallet.ClientID // if not blank
	}
	go GetInfoFromSharders(WithParams(GET_VESTING_CLIENT_POOLS, Params{
		"client_id": clientID,
	}), 0, cb)
	return
}

type VestingDestInfo struct {
	ID     common.Key       `json:"id"`     // identifier
	Wanted common.Balance   `json:"wanted"` // wanted amount for entire period
	Earned common.Balance   `json:"earned"` // can unlock
	Vested common.Balance   `json:"vested"` // already vested
	Last   common.Timestamp `json:"last"`   // last time unlocked
}

type VestingPoolInfo struct {
	ID           common.Key         `json:"pool_id"`      // pool ID
	Balance      common.Balance     `json:"balance"`      // real pool balance
	Left         common.Balance     `json:"left"`         // owner can unlock
	Description  string             `json:"description"`  // description
	StartTime    common.Timestamp   `json:"start_time"`   // from
	ExpireAt     common.Timestamp   `json:"expire_at"`    // until
	Destinations []*VestingDestInfo `json:"destinations"` // receivers
	ClientID     common.Key         `json:"client_id"`    // owner
}

func GetVestingPoolInfo(poolID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	GetInfoFromSharders(WithParams(GET_VESTING_POOL_INFO, Params{
		"pool_id": poolID,
	}), 0, cb)
	return
}

func GetVestingSCConfig(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go GetInfoFromSharders(GET_VESTING_CONFIG, 0, cb)
	return
}

// faucet

func GetFaucetSCConfig(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go GetInfoFromSharders(GET_FAUCETSC_CONFIG, 0, cb)
	return
}

func (t *Transaction) ZCNSCAddAuthorizer(ip *AddAuthorizerPayload) (err error) {
	err = t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_ADD_AUTHORIZER, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go t.setNonceAndSubmit()
	return
}

func (t *Transaction) ZCNSCAuthorizerHealthCheck(ip *AuthorizerHealthCheckPayload) (err error) {
	err = t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_AUTHORIZER_HEALTH_CHECK, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go t.setNonceAndSubmit()
	return
}

func (t *Transaction) ZCNSCDeleteAuthorizer(ip *DeleteAuthorizerPayload) (err error) {
	err = t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_DELETE_AUTHORIZER, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go t.setNonceAndSubmit()
	return
}

func (t *Transaction) ZCNSCCollectReward(providerId string, providerType Provider) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
	}
	err := t.createSmartContractTxn(ZCNSCSmartContractAddress,
		transaction.ZCNSC_COLLECT_REWARD, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return err
}
