//go:build !mobile
// +build !mobile

package zcncore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/block"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/core/zcncrypto"
)

type Provider int

const (
	ProviderMiner Provider = iota + 1
	ProviderSharder
	ProviderBlobber
	ProviderValidator
	ProviderAuthorizer
)

type ConfirmationStatus int

const (
	Undefined ConfirmationStatus = iota
	Success
	ChargeableError
)

type ChainConfig struct {
	ChainID                 string   `json:"chain_id,omitempty"`
	BlockWorker             string   `json:"block_worker"`
	Miners                  []string `json:"miners"`
	Sharders                []string `json:"sharders"`
	SignatureScheme         string   `json:"signature_scheme"`
	MinSubmit               int      `json:"min_submit"`
	MinConfirmation         int      `json:"min_confirmation"`
	ConfirmationChainLength int      `json:"confirmation_chain_length"`
	EthNode                 string   `json:"eth_node"`
}

type localConfig struct {
	chain         ChainConfig
	wallet        zcncrypto.Wallet
	authUrl       string
	isConfigured  bool
	isValidWallet bool
	isSplitWallet bool
}

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

type Node struct {
	Miner     Miner `json:"simple_miner"`
	StakePool `json:"stake_pool"`
}

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

type MinerSCDelegatePoolInfo struct {
	ID         common.Key     `json:"id"`
	Balance    common.Balance `json:"balance"`
	Reward     common.Balance `json:"reward"`      // uncollected reread
	RewardPaid common.Balance `json:"reward_paid"` // total reward all time
	Status     string         `json:"status"`
}

type MinerSCUserPoolsInfo struct {
	Pools map[string][]*MinerSCDelegatePoolInfo `json:"pools"`
}

type TransactionCommon interface {
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input interface{}, val uint64) (*transaction.Transaction, error)
	// Send implements sending token to a given clientid
	Send(toClientID string, val uint64, desc string) error
	// SetTransactionFee implements method to set the transaction fee
	SetTransactionFee(txnFee uint64) error

	//RegisterMultiSig registers a group wallet and subwallets with MultisigSC
	RegisterMultiSig(walletstr, mswallet string) error

	VestingAdd(ar *VestingAddRequest, value uint64) error

	MinerSCLock(minerID string, lock uint64) error
	MinerSCCollectReward(string, string, Provider) error

	StorageSCCollectReward(string, string, Provider) error

	FinalizeAllocation(allocID string, fee uint64) error
	CancelAllocation(allocID string, fee uint64) error
	CreateAllocation(car *CreateAllocationRequest, lock uint64, fee uint64) error //
	CreateReadPool(fee uint64) error
	ReadPoolLock(allocID string, blobberID string, duration int64, lock uint64, fee uint64) error
	ReadPoolUnlock(poolID string, fee uint64) error
	StakePoolLock(blobberID string, lock uint64, fee uint64) error
	StakePoolUnlock(blobberID string, poolID string, fee uint64) error
	UpdateBlobberSettings(blobber *Blobber, fee uint64) error
	UpdateValidatorSettings(validator *Validator, fee uint64) error
	UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock uint64, fee uint64) error
	WritePoolLock(allocID string, blobberID string, duration int64, lock uint64, fee uint64) error
	WritePoolUnlock(poolID string, fee uint64) error

	VestingUpdateConfig(*InputMap) error
	MinerScUpdateConfig(*InputMap) error
	MinerScUpdateGlobals(*InputMap) error
	StorageScUpdateConfig(*InputMap) error
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

	// GetVerifyConfirmationStatus implements the verification status from sharders
	GetVerifyConfirmationStatus() ConfirmationStatus
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
	DelegateWallet string         `json:"delegate_wallet"`
	MinStake       common.Balance `json:"min_stake"`
	MaxStake       common.Balance `json:"max_stake"`
	NumDelegates   int            `json:"num_delegates"`
	ServiceCharge  float64        `json:"service_charge"`
}

type Terms struct {
	ReadPrice        common.Balance `json:"read_price"`  // tokens / GB
	WritePrice       common.Balance `json:"write_price"` // tokens / GB
	MinLockDemand    float64        `json:"min_lock_demand"`
	MaxOfferDuration time.Duration  `json:"max_offer_duration"`
}

type Blobber struct {
	ID                common.Key        `json:"id"`
	BaseURL           string            `json:"url"`
	Terms             Terms             `json:"terms"`
	Capacity          common.Size       `json:"capacity"`
	Allocated         common.Size       `json:"allocated"`
	LastHealthCheck   common.Timestamp  `json:"last_health_check"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

type Validator struct {
	ID                common.Key        `json:"id"`
	BaseURL           string            `json:"url"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

type AddAuthorizerPayload struct {
	PublicKey         string                      `json:"public_key"`
	URL               string                      `json:"url"`
	StakePoolSettings AuthorizerStakePoolSettings `json:"stake_pool_settings"` // Used to initially create stake pool
}

type AuthorizerStakePoolSettings struct {
	DelegateWallet string         `json:"delegate_wallet"`
	MinStake       common.Balance `json:"min_stake"`
	MaxStake       common.Balance `json:"max_stake"`
	NumDelegates   int            `json:"num_delegates"`
	ServiceCharge  float64        `json:"service_charge"`
}

type AuthorizerConfig struct {
	Fee common.Balance `json:"fee"`
}

type InputMap struct {
	Fields map[string]string `json:"Fields"`
}

// NewTransaction allocation new generic transaction object for any operation
func NewTransaction(cb TransactionCallback, txnFee uint64, nonce int64) (TransactionScheme, error) {
	err := CheckConfig()
	if err != nil {
		return nil, err
	}
	if _config.isSplitWallet {
		if _config.authUrl == "" {
			return nil, errors.New("", "auth url not set")
		}
		logging.Info("New transaction interface with auth")
		return newTransactionWithAuth(cb, txnFee, nonce)
	}
	logging.Info("New transaction interface")
	t, err := newTransaction(cb, txnFee, nonce)
	return t, err
}

func (t *Transaction) ExecuteSmartContract(address, methodName string, input interface{}, val uint64) (*transaction.Transaction, error) {
	err := t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return nil, err
	}
	go func() {
		t.setNonceAndSubmit()
	}()
	return t.txn, nil
}

func (t *Transaction) SetTransactionFee(txnFee uint64) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction fee.")
	}
	t.txn.TransactionFee = txnFee
	return nil
}

func (t *Transaction) Send(toClientID string, val uint64, desc string) error {
	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSend
		t.txn.ToClientID = toClientID
		t.txn.Value = val
		t.txn.TransactionData = string(txnData)
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) SendWithSignatureHash(toClientID string, val uint64, desc string, sig string, CreationDate int64, hash string) error {
	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSend
		t.txn.ToClientID = toClientID
		t.txn.Value = val
		t.txn.Hash = hash
		t.txn.TransactionData = string(txnData)
		t.txn.Signature = sig
		t.txn.CreationDate = CreationDate
		t.setNonceAndSubmit()
	}()
	return nil
}

type VestingDest struct {
	ID     string     `json:"id"`     // destination ID
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

func (t *Transaction) MinerSCLock(nodeID string, lock uint64) (err error) {

	var mscl MinerSCLock
	mscl.ID = nodeID

	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, &mscl, lock)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCCollectReward(providerId, poolId string, providerType Provider) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		PoolId:       poolId,
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

func (t *Transaction) StorageSCCollectReward(providerId, poolId string, providerType Provider) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		PoolId:       poolId,
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
func (t *Transaction) FinalizeAllocation(allocID string, fee uint64) (
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
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// CancelAllocation transaction.
func (t *Transaction) CancelAllocation(allocID string, fee uint64) (
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
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// CreateAllocation transaction.
func (t *Transaction) CreateAllocation(car *CreateAllocationRequest,
	lock uint64, fee uint64) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_ALLOCATION, car, lock)
	if err != nil {
		logging.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// CreateReadPool for current user.
func (t *Transaction) CreateReadPool(fee uint64) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_READ_POOL, nil, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// ReadPoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) ReadPoolLock(allocID, blobberID string,
	duration int64, lock, fee uint64) (err error) {

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
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// ReadPoolUnlock for current user and given pool.
func (t *Transaction) ReadPoolUnlock(poolID string, fee uint64) (err error) {
	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_UNLOCK, &unlockRequest{
			PoolID: poolID,
		}, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// StakePoolLock used to lock tokens in a stake pool of a blobber.
func (t *Transaction) StakePoolLock(blobberID string, lock, fee uint64) (
	err error) {

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_LOCK, &spr, lock)
	if err != nil {
		logging.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// StakePoolUnlock by blobberID and poolID.
func (t *Transaction) StakePoolUnlock(blobberID, poolID string,
	fee uint64) (err error) {

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
		PoolID    string `json:"pool_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID
	spr.PoolID = poolID

	err = t.createSmartContractTxn(StorageSmartContractAddress, transaction.STORAGESC_STAKE_POOL_UNLOCK, &spr, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// UpdateBlobberSettings update settings of a blobber.
func (t *Transaction) UpdateBlobberSettings(b *Blobber, fee uint64) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, b, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// UpdateAllocation transaction.
func (t *Transaction) UpdateAllocation(allocID string, sizeDiff int64,
	expirationDiff int64, lock, fee uint64) (err error) {

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
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// WritePoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) WritePoolLock(allocID, blobberID string, duration int64,
	lock, fee uint64) (err error) {

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
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// WritePoolUnlock for current user and given pool.
func (t *Transaction) WritePoolUnlock(poolID string, fee uint64) (
	err error) {

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_UNLOCK, &unlockRequest{
			PoolID: poolID,
		}, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	t.SetTransactionFee(fee)
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

//RegisterMultiSig register a multisig wallet with the SC.
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
			nonce = transaction.Cache.GetNextNonce(t.txn.ClientID)
		} else {
			transaction.Cache.Set(t.txn.ClientID, nonce)
		}
		t.txn.TransactionNonce = nonce

		t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
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

//RegisterVote register a multisig wallet with the SC.
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
			nonce = transaction.Cache.GetNextNonce(t.txn.ClientID)
		} else {
			transaction.Cache.Set(t.txn.ClientID, nonce)
		}
		t.txn.TransactionNonce = nonce
		t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
		t.submitTxn()
	}()
	return nil
}

type MinerSCDelegatePool struct {
	Settings StakePoolSettings `json:"settings"`
}

type SimpleMiner struct {
	ID string `json:"id"`
}

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

type AuthorizerNode struct {
	ID     string            `json:"id"`
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
			transaction.Cache.Evict(t.txn.ClientID)
			return errors.New("", "invalid transaction. cannot be verified.")
		}
	}
	// If transaction is verify only start from current time
	if t.txn.CreationDate == 0 {
		t.txn.CreationDate = int64(common.Now())
	}

	tq, err := NewTransactionQuery(_config.chain.Sharders)
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

				var tr map[string]json.RawMessage
				if err := json.Unmarshal(txnJson, &tr); err != nil {
					return
				}

				txStatus := tr["transaction_status"]
				switch string(txStatus) {
				case "1":
					t.completeVerifyWithConStatus(StatusSuccess, int(Success), string(output), nil)
				case "2":
					txOutput := tr["transaction_output"]
					t.completeVerifyWithConStatus(StatusSuccess, int(ChargeableError), string(txOutput), nil)
				default:
					t.completeVerify(StatusError, string(output), nil)
				}
				return
			}
		}
	}()
	return nil
}

// ConvertToValue converts ZCN tokens to value
func ConvertToValue(token float64) uint64 {
	return uint64(token * float64(TOKEN_UNIT))
}

func GetLatestFinalized(ctx context.Context, numSharders int) (b *block.Header, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders, GET_LATEST_FINALIZED, result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	for i := 0; i < numSharders; i++ {
		var rsp = <-result

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

func GetLatestFinalizedMagicBlock(ctx context.Context, numSharders int) (m *block.MagicBlock, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders, GET_LATEST_FINALIZED_MAGIC_BLOCK, result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	type respObj struct {
		MagicBlock *block.MagicBlock `json:"magic_block"`
	}

	for i := 0; i < numSharders; i++ {
		var rsp = <-result

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

	var numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders, GET_CHAIN_STATS, result)
	var rsp *util.GetResponse
	for i := 0; i < numSharders; i++ {
		var x = <-result
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

func GetBlockByRound(ctx context.Context, numSharders int, round int64) (b *block.Block, err error) {

	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders,
		fmt.Sprintf("%sround=%d&content=full,header", GET_BLOCK_INFO, round),
		result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	type respObj struct {
		Block  *block.Block  `json:"block"`
		Header *block.Header `json:"header"`
	}

	for i := 0; i < numSharders; i++ {
		var rsp = <-result

		logging.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			logging.Error(rsp.Body)
			continue
		}

		var respo respObj
		if err = json.Unmarshal([]byte(rsp.Body), &respo); err != nil {
			logging.Error("block parse error: ", err)
			err = nil
			continue
		}

		if respo.Block == nil {
			logging.Debug(rsp.Url, "no block in response:", rsp.Body)
			continue
		}

		if respo.Header == nil {
			logging.Debug(rsp.Url, "no block header in response:", rsp.Body)
			continue
		}

		if respo.Header.Hash != string(respo.Block.Hash) {
			logging.Debug(rsp.Url, "header and block hash mismatch:", rsp.Body)
			continue
		}

		b = respo.Block
		b.Header = respo.Header

		var h = encryption.FastHash([]byte(b.Hash))
		if roundConsensus[h]++; roundConsensus[h] > maxConsensus {
			maxConsensus = roundConsensus[h]
		}
	}

	if maxConsensus == 0 {
		return nil, errors.New("", "round info not found")
	}

	return
}

func GetMagicBlockByNumber(ctx context.Context, numSharders int, number int64) (m *block.MagicBlock, err error) {

	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders,
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

		timeout, _ := context.WithTimeout(context.Background(), time.Second)
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

func WithEthereumNode(uri string) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.EthNode = uri
		return nil
	}
}

func WithChainID(id string) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.ChainID = id
		return nil
	}
}

func WithMinSubmit(m int) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.MinSubmit = m
		return nil
	}
}

func WithMinConfirmation(m int) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.MinConfirmation = m
		return nil
	}
}

func WithConfirmationChainLength(m int) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.ConfirmationChainLength = m
		return nil
	}
}

// UpdateValidatorSettings update settings of a validator.
func (t *Transaction) UpdateValidatorSettings(v *Validator, fee uint64) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_VALIDATOR_SETTINGS, v, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// InitZCNSDK initializes the SDK with miner, sharder and signature scheme provided.
func InitZCNSDK(blockWorker string, signscheme string, configs ...func(*ChainConfig) error) error {
	if signscheme != "ed25519" && signscheme != "bls0chain" {
		return errors.New("", "invalid/unsupported signature scheme")
	}
	_config.chain.BlockWorker = blockWorker
	_config.chain.SignatureScheme = signscheme

	err := UpdateNetworkDetails()
	if err != nil {
		fmt.Println("UpdateNetworkDetails:", err)
		return err
	}

	go updateNetworkDetailsWorker(context.Background())

	for _, conf := range configs {
		err := conf(&_config.chain)
		if err != nil {
			return errors.Wrap(err, "invalid/unsupported options.")
		}
	}
	assertConfig()
	_config.isConfigured = true
	logging.Info("******* Wallet SDK Version:", version.VERSIONSTR, " ******* (InitZCNSDK)")

	cfg := &conf.Config{
		BlockWorker:             _config.chain.BlockWorker,
		MinSubmit:               _config.chain.MinSubmit,
		MinConfirmation:         _config.chain.MinConfirmation,
		ConfirmationChainLength: _config.chain.ConfirmationChainLength,
		SignatureScheme:         _config.chain.SignatureScheme,
		ChainID:                 _config.chain.ChainID,
		EthereumNode:            _config.chain.EthNode,
	}

	conf.InitClientConfig(cfg)

	return nil
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
	go getInfoFromSharders(WithParams(GET_VESTING_CLIENT_POOLS, Params{
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
	getInfoFromSharders(WithParams(GET_VESTING_POOL_INFO, Params{
		"pool_id": poolID,
	}), 0, cb)
	return
}

func GetVestingSCConfig(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go getInfoFromSharders(GET_VESTING_CONFIG, 0, cb)
	return
}

// faucet

func GetFaucetSCConfig(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go getInfoFromSharders(GET_FAUCETSC_CONFIG, 0, cb)
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
