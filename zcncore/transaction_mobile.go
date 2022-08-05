//go:build mobile
// +build mobile

package zcncore

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/block"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
)

const (
	Undefined int = iota
	Success
	ChargeableError
)

type TransactionCommon interface {
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input string, val string) error

	// Send implements sending token to a given clientid
	Send(toClientID string, val string, desc string) error
	// SetTransactionFee implements method to set the transaction fee
	SetTransactionFee(txnFee string) error

	VestingAdd(ar VestingAddRequest, value string) error

	MinerSCLock(minerID string, lock string) error
	MinerSCCollectReward(providerId string, poolId string, providerType int) error
	StorageSCCollectReward(providerId, poolId string, providerType int) error

	FinalizeAllocation(allocID string, fee string) error
	CancelAllocation(allocID string, fee string) error
	CreateAllocation(car *CreateAllocationRequest, lock, fee string) error //
	CreateReadPool(fee string) error
	ReadPoolLock(allocID string, blobberID string, duration int64, lock, fee string) error
	ReadPoolUnlock(poolID string, fee string) error
	StakePoolLock(blobberID string, lock, fee string) error
	StakePoolUnlock(blobberID string, poolID string, fee string) error
	UpdateBlobberSettings(blobber Blobber, fee string) error
	UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock, fee string) error
	WritePoolLock(allocID string, lock, fee string) error
	WritePoolUnlock(allocID string, fee string) error

	VestingUpdateConfig(InputMap) error
	MinerScUpdateConfig(InputMap) error
	MinerScUpdateGlobals(InputMap) error
	StorageScUpdateConfig(InputMap) error
	FaucetUpdateConfig(InputMap) error
	ZCNSCUpdateGlobalConfig(InputMap) error

	MinerSCMinerSettings(MinerSCMinerInfo) error
	MinerSCSharderSettings(MinerSCMinerInfo) error
	MinerSCDeleteMiner(MinerSCMinerInfo) error
	MinerSCDeleteSharder(MinerSCMinerInfo) error

	// ZCNSCUpdateAuthorizerConfig updates authorizer config by ID
	ZCNSCUpdateAuthorizerConfig(AuthorizerNode) error
	// ZCNSCAddAuthorizer adds authorizer
	ZCNSCAddAuthorizer(AddAuthorizerPayload) error

	GetVerifyConfirmationStatus() int
}


type chainConfig struct {
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
	chain         chainConfig
	wallet        zcncrypto.Wallet
	authUrl       string
	isConfigured  bool
	isValidWallet bool
	isSplitWallet bool
}

// priceRange represents a price range allowed by user to filter blobbers.
type priceRange struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

// createAllocationRequest is information to create allocation.
type createAllocationRequest struct {
	DataShards      int        `json:"data_shards"`
	ParityShards    int        `json:"parity_shards"`
	Size            int64      `json:"size"`
	Expiration      int64      `json:"expiration_date"`
	Owner           string     `json:"owner_id"`
	OwnerPublicKey  string     `json:"owner_public_key"`
	Blobbers        []string   `json:"blobbers"`
	ReadPriceRange  priceRange `json:"read_price_range"`
	WritePriceRange priceRange `json:"write_price_range"`
}

type CreateAllocationRequest struct {
	DataShards     int
	ParityShards   int
	Size           int64
	Expiration     int64
	Owner          string
	OwnerPublicKey string
	ReadPriceMin   int64
	ReadPriceMax   int64
	WritePriceMin  int64
	WritePriceMax  int64

	blobbers []string
}

func (car *CreateAllocationRequest) AddBlobber(blobber string) {
	car.blobbers = append(car.blobbers, blobber)
}

func (car *CreateAllocationRequest) toCreateAllocationSCInput() *createAllocationRequest {
	return &createAllocationRequest{
		DataShards:      car.DataShards,
		ParityShards:    car.ParityShards,
		Size:            car.Size,
		Expiration:      car.Expiration,
		Owner:           car.Owner,
		OwnerPublicKey:  car.OwnerPublicKey,
		Blobbers:        car.blobbers,
		ReadPriceRange:  priceRange{Min: car.ReadPriceMin, Max: car.ReadPriceMax},
		WritePriceRange: priceRange{Min: car.WritePriceMin, Max: car.WritePriceMax},
	}
}

type StakePoolSettings struct {
	DelegateWallet string  `json:"delegate_wallet"`
	MinStake       int64   `json:"min_stake"`
	MaxStake       int64   `json:"max_stake"`
	NumDelegates   int     `json:"num_delegates"`
	ServiceCharge  float64 `json:"service_charge"`
}

type Terms struct {
	ReadPrice        int64   `json:"read_price"`  // tokens / GB
	WritePrice       int64   `json:"write_price"` // tokens / GB
	MinLockDemand    float64 `json:"min_lock_demand"`
	MaxOfferDuration int64   `json:"max_offer_duration"`
}

type Blobber interface {
	SetTerms(readPrice int64, writePrice int64, minLockDemand float64, maxOfferDuration int64)
	SetStakePoolSettings(delegateWallet string, minStake int64, maxStake int64, numDelegates int, serviceCharge float64)
}

func NewBlobber(id, baseUrl string, capacity, allocated, lastHealthCheck int64) Blobber {
	return &blobber{
		ID:              id,
		BaseURL:         baseUrl,
		Capacity:        capacity,
		Allocated:       allocated,
		LastHealthCheck: lastHealthCheck,
	}
}

type blobber struct {
	ID                string            `json:"id"`
	BaseURL           string            `json:"url"`
	Capacity          int64             `json:"capacity"`
	Allocated         int64             `json:"allocated"`
	LastHealthCheck   int64             `json:"last_health_check"`
	Terms             Terms             `json:"terms"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

func (b *blobber) SetStakePoolSettings(delegateWallet string, minStake int64, maxStake int64, numDelegates int, serviceCharge float64) {
	b.StakePoolSettings = StakePoolSettings{
		DelegateWallet: delegateWallet,
		MinStake:       minStake,
		MaxStake:       maxStake,
		NumDelegates:   numDelegates,
		ServiceCharge:  serviceCharge,
	}
}

func (b *blobber) SetTerms(readPrice int64, writePrice int64, minLockDemand float64, maxOfferDuration int64) {
	b.Terms = Terms{
		ReadPrice:        readPrice,
		WritePrice:       writePrice,
		MinLockDemand:    minLockDemand,
		MaxOfferDuration: maxOfferDuration,
	}
}

type Validator interface {
	SetStakePoolSettings(delegateWallet string, minStake int64, maxStake int64, numDelegates int, serviceCharge float64)
}

func NewValidator(id string, baseUrl string) Validator {
	return &validator{
		ID:      common.Key(id),
		BaseURL: baseUrl,
	}
}

type validator struct {
	ID                common.Key        `json:"id"`
	BaseURL           string            `json:"url"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

func (v *validator) SetStakePoolSettings(delegateWallet string, minStake int64, maxStake int64, numDelegates int, serviceCharge float64) {
	v.StakePoolSettings = StakePoolSettings{
		DelegateWallet: delegateWallet,
		MinStake:       minStake,
		MaxStake:       maxStake,
		NumDelegates:   numDelegates,
		ServiceCharge:  serviceCharge,
	}
}

type AddAuthorizerPayload interface {
	SetStakePoolSettings(delegateWallet string, minStake int64, maxStake int64, numDelegates int, serviceCharge float64)
}

func NewAddAuthorizerPayload(pubKey, url string) AddAuthorizerPayload {
	return &addAuthorizerPayload{
		PublicKey: pubKey,
		URL:       url,
	}
}

type addAuthorizerPayload struct {
	PublicKey         string                      `json:"public_key"`
	URL               string                      `json:"url"`
	StakePoolSettings AuthorizerStakePoolSettings `json:"stake_pool_settings"` // Used to initially create stake pool
}

func (a *addAuthorizerPayload) SetStakePoolSettings(delegateWallet string, minStake int64, maxStake int64, numDelegates int, serviceCharge float64) {
	a.StakePoolSettings = AuthorizerStakePoolSettings{
		DelegateWallet: delegateWallet,
		MinStake:       minStake,
		MaxStake:       maxStake,
		NumDelegates:   numDelegates,
		ServiceCharge:  serviceCharge,
	}
}

type AuthorizerStakePoolSettings struct {
	DelegateWallet string  `json:"delegate_wallet"`
	MinStake       int64   `json:"min_stake"`
	MaxStake       int64   `json:"max_stake"`
	NumDelegates   int     `json:"num_delegates"`
	ServiceCharge  float64 `json:"service_charge"`
}

type AuthorizerConfig struct {
	Fee int64 `json:"fee"`
}

type VestingDest struct {
	ID     string `json:"id"`     // destination ID
	Amount int64  `json:"amount"` // amount to vest for the destination
}

type VestingAddRequest interface {
	AddDestinations(id string, amount int64)
}

func NewVestingAddRequest(desc string, startTime int64, duration int64) VestingAddRequest {
	return &vestingAddRequest{
		Description: desc,
		StartTime:   startTime,
		Duration:    duration,
	}
}

type vestingAddRequest struct {
	Description  string         `json:"description"`  // allow empty
	StartTime    int64          `json:"start_time"`   //
	Duration     int64          `json:"duration"`     //
	Destinations []*VestingDest `json:"destinations"` //
}

func (vr *vestingAddRequest) AddDestinations(id string, amount int64) {
	vr.Destinations = append(vr.Destinations, &VestingDest{ID: id, Amount: amount})
}

type InputMap interface {
	AddField(key, value string)
}

type inputMap struct {
	Fields map[string]string `json:"fields"`
}

func NewInputMap() InputMap {
	return &inputMap{
		Fields: make(map[string]string),
	}
}

func (im *inputMap) AddField(key, value string) {
	im.Fields[key] = value
}

func parseCoinStr(vs string) (uint64, error) {
	if vs == "" {
		return 0, nil
	}

	v, err := strconv.ParseUint(vs, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid token value: %v, err: %v", vs, err)
	}

	return v, nil
}

// NewTransaction allocation new generic transaction object for any operation
func NewTransaction(cb TransactionCallback, txnFee string, nonce int64) (TransactionScheme, error) {
	v, err := parseCoinStr(txnFee)
	if err != nil {
		return nil, err
	}

	err = CheckConfig()
	if err != nil {
		return nil, err
	}
	if _config.isSplitWallet {
		if _config.authUrl == "" {
			return nil, errors.New("", "auth url not set")
		}
		logging.Info("New transaction interface with auth")
		return newTransactionWithAuth(cb, v, nonce)
	}
	logging.Info("New transaction interface")
	t, err := newTransaction(cb, v, nonce)
	return t, err
}

func (t *Transaction) ExecuteSmartContract(address, methodName string, input string, val string) error {
	v, err := parseCoinStr(val)
	if err != nil {
		return err
	}

	err = t.createSmartContractTxn(address, methodName, json.RawMessage(input), v)
	if err != nil {
		return err
	}
	go func() {
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) SetTransactionFee(txnFee string) error {
	fee, err := parseCoinStr(txnFee)
	if err != nil {
		return err
	}

	return t.setTransactionFee(fee)
}

func (t *Transaction) setTransactionFee(fee uint64) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction fee.")
	}
	t.txn.TransactionFee = fee
	return nil
}

func (t *Transaction) Send(toClientID string, val string, desc string) error {
	v, err := parseCoinStr(val)
	if err != nil {
		return err
	}

	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSend
		t.txn.ToClientID = toClientID
		t.txn.Value = v
		t.txn.TransactionData = string(txnData)
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) SendWithSignatureHash(toClientID string, val string, desc string, sig string, CreationDate int64, hash string) error {
	v, err := parseCoinStr(val)
	if err != nil {
		return err
	}

	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSend
		t.txn.ToClientID = toClientID
		t.txn.Value = v
		t.txn.Hash = hash
		t.txn.TransactionData = string(txnData)
		t.txn.Signature = sig
		t.txn.CreationDate = CreationDate
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) VestingAdd(ar VestingAddRequest, value string) (
	err error) {

	v, err := parseCoinStr(value)
	if err != nil {
		return err
	}

	err = t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_ADD, ar, v)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCLock(nodeID string, lock string) (err error) {
	v, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	var mscl MinerSCLock
	mscl.ID = nodeID

	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, &mscl, v)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCCollectReward(providerId, poolId string, providerType int) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		PoolId:       poolId,
		ProviderType: providerType,
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

func (t *Transaction) StorageSCCollectReward(providerId, poolId string, providerType int) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		PoolId:       poolId,
		ProviderType: providerType,
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
func (t *Transaction) FinalizeAllocation(allocID string, fee string) (
	err error) {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

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
	t.setTransactionFee(v)
	go func() { t.setNonceAndSubmit() }()
	return
}

// CancelAllocation transaction.
func (t *Transaction) CancelAllocation(allocID string, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type cancelRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CANCEL_ALLOCATION, &cancelRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	t.setTransactionFee(v)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// CreateAllocation transaction.
func (t *Transaction) CreateAllocation(car *CreateAllocationRequest, lock, fee string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	fv, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_ALLOCATION, car.toCreateAllocationSCInput(), lv)
	if err != nil {
		logging.Error(err)
		return err
	}
	t.setTransactionFee(fv)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// CreateReadPool for current user.
func (t *Transaction) CreateReadPool(fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_READ_POOL, nil, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	t.setTransactionFee(v)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// ReadPoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) ReadPoolLock(allocID, blobberID string,
	duration int64, lock, fee string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	fv, err := parseCoinStr(fee)
	if err != nil {
		return err
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
		transaction.STORAGESC_READ_POOL_LOCK, &lr, lv)
	if err != nil {
		logging.Error(err)
		return err
	}
	t.setTransactionFee(fv)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// ReadPoolUnlock for current user and given pool.
func (t *Transaction) ReadPoolUnlock(poolID string, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_UNLOCK, &unlockRequest{
			PoolID: poolID,
		}, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	t.setTransactionFee(v)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// StakePoolLock used to lock tokens in a stake pool of a blobber.
func (t *Transaction) StakePoolLock(blobberID string, lock, fee string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	fv, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_LOCK, &spr, lv)
	if err != nil {
		logging.Error(err)
		return err
	}
	t.setTransactionFee(fv)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// StakePoolUnlock by blobberID and poolID.
func (t *Transaction) StakePoolUnlock(blobberID, poolID string, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

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
		return err
	}
	t.setTransactionFee(v)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// UpdateBlobberSettings update settings of a blobber.
func (t *Transaction) UpdateBlobberSettings(b Blobber, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, b, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	t.setTransactionFee(v)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// UpdateAllocation transaction.
func (t *Transaction) UpdateAllocation(allocID string, sizeDiff int64,
	expirationDiff int64, lock, fee string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	fv, err := parseCoinStr(fee)
	if err != nil {
		return err
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
		transaction.STORAGESC_UPDATE_ALLOCATION, &uar, lv)
	if err != nil {
		logging.Error(err)
		return err
	}
	t.setTransactionFee(fv)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// WritePoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) WritePoolLock(allocID, lock, fee string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	fv, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	var lr = struct {
		AllocationID string        `json:"allocation_id"`
	} {
		AllocationID: allocID,
	}

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_LOCK, &lr, lv)
	if err != nil {
		logging.Error(err)
		return err
	}
	t.setTransactionFee(fv)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// WritePoolUnlock for current user and given pool.
func (t *Transaction) WritePoolUnlock(allocID string, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	var ur = struct {
		AllocationID string `json:"allocation_id"`
	} {
		AllocationID: allocID,
	}

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_UNLOCK, &ur, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	t.setTransactionFee(v)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

func (t *Transaction) VestingUpdateConfig(vscc InputMap) (err error) {

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

func (t *Transaction) FaucetUpdateConfig(ip InputMap) (err error) {

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

func (t *Transaction) MinerScUpdateConfig(ip InputMap) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerScUpdateGlobals(ip InputMap) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_GLOBALS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) StorageScUpdateConfig(ip InputMap) (err error) {
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) ZCNSCUpdateGlobalConfig(ip InputMap) (err error) {
	err = t.createSmartContractTxn(ZCNSCSmartContractAddress,
		transaction.ZCNSC_UPDATE_GLOBAL_CONFIG, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go t.setNonceAndSubmit()
	return
}

func (t *Transaction) GetVerifyConfirmationStatus() int {
	return int(t.verifyConfirmationStatus)
}

type MinerSCMinerInfo interface {
	GetID() string
}

func NewMinerSCMinerInfo(id string, delegateWallet string,
	minStake int64, maxStake int64, numDelegates int, serviceCharge float64) MinerSCMinerInfo {
	return &minerSCMinerInfo{
		simpleMiner: simpleMiner{ID: id},
		minerSCDelegatePool: minerSCDelegatePool{
			Settings: StakePoolSettings{
				DelegateWallet: delegateWallet,
				MinStake:       minStake,
				MaxStake:       maxStake,
				NumDelegates:   numDelegates,
				ServiceCharge:  serviceCharge,
			},
		},
	}
}

type minerSCMinerInfo struct {
	simpleMiner         `json:"simple_miner"`
	minerSCDelegatePool `json:"stake_pool"`
}

func (mi *minerSCMinerInfo) GetID() string {
	return mi.ID
}

type minerSCDelegatePool struct {
	Settings StakePoolSettings `json:"settings"`
}

type simpleMiner struct {
	ID string `json:"id"`
}

func (t *Transaction) MinerSCMinerSettings(info MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_SETTINGS, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCSharderSettings(info MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_SETTINGS, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCDeleteMiner(info MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_DELETE, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCDeleteSharder(info MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_DELETE, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

type AuthorizerNode interface {
	GetID() string
}

func NewAuthorizerNode(id string, fee int64) AuthorizerNode {
	return &authorizerNode{
		ID:     id,
		Config: &AuthorizerConfig{Fee: fee},
	}
}

type authorizerNode struct {
	ID     string            `json:"id"`
	Config *AuthorizerConfig `json:"config"`
}

func (a *authorizerNode) GetID() string {
	return a.ID
}

func (t *Transaction) ZCNSCUpdateAuthorizerConfig(ip AuthorizerNode) (err error) {
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

	tq, err := newTransactionQuery(_config.chain.Sharders)
	if err != nil {
		logging.Error(err)
		return err
	}

	go func() {

		for {

			tq.Reset()
			// Get transaction confirmationBlock from a random sharder
			confirmBlockHeader, confirmationBlock, lfbBlockHeader, err := tq.getFastConfirmation(t.txnHash, nil)

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
					confirmBlockHeader, confirmationBlock, lfbBlockHeader, err = tq.getConsensusConfirmation(getMinShardersVerify(), t.txnHash, nil)
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
					t.completeVerifyWithConStatus(StatusSuccess, Success, string(output), nil)
				case "2":
					txOutput := tr["transaction_output"]
					t.completeVerifyWithConStatus(StatusSuccess, ChargeableError, string(txOutput), nil)
				default:
					t.completeVerify(StatusError, string(output), nil)
				}
				return
			}
		}
	}()
	return nil
}

func (t *Transaction) ZCNSCAddAuthorizer(ip AddAuthorizerPayload) (err error) {
	err = t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_ADD_AUTHORIZER, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go t.setNonceAndSubmit()
	return
}

// ConvertToValue converts ZCN tokens to value
func ConvertToValue(token float64) string {
	return strconv.FormatUint(uint64(token*float64(TOKEN_UNIT)), 10)
}

func makeTimeoutContext(tm RequestTimeout) (context.Context, func()) {

	if tm != nil && tm.Get() > 0 {
		return context.WithTimeout(context.Background(), time.Millisecond*time.Duration(tm.Get()))

	}
	return context.Background(), func() {}

}

func GetLatestFinalized(numSharders int, timeout RequestTimeout) (b *BlockHeader, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	ctx, cancel := makeTimeoutContext(timeout)
	defer cancel()

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

func GetLatestFinalizedMagicBlock(numSharders int, timeout RequestTimeout) ([]byte, error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	ctx, cancel := makeTimeoutContext(timeout)
	defer cancel()

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders, GET_LATEST_FINALIZED_MAGIC_BLOCK, result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
		m              *block.MagicBlock
		err            error
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

	if m != nil {
		return json.Marshal(m)
	}

	return nil, err
}

// GetChainStats gets chain stats with time out
// timeout in milliseconds
func GetChainStats(timeout RequestTimeout) ([]byte, error) {
	var result = make(chan *util.GetResponse, 1)
	defer close(result)

	ctx, cancel := makeTimeoutContext(timeout)
	defer cancel()

	var (
		b   *block.ChainStats
		err error
	)

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

	return []byte(rsp.Body), nil
}

type BlockHeader struct {
	Version               string `json:"version,omitempty"`
	CreationDate          int64  `json:"creation_date,omitempty"`
	Hash                  string `json:"hash,omitempty"`
	MinerID               string `json:"miner_id,omitempty"`
	Round                 int64  `json:"round,omitempty"`
	RoundRandomSeed       int64  `json:"round_random_seed,omitempty"`
	MerkleTreeRoot        string `json:"merkle_tree_root,omitempty"`
	StateHash             string `json:"state_hash,omitempty"`
	ReceiptMerkleTreeRoot string `json:"receipt_merkle_tree_root,omitempty"`
	NumTxns               int64  `json:"num_txns,omitempty"`
}

type Block struct {
	MinerID           string `json:"miner_id"`
	Round             int64  `json:"round"`
	RoundRandomSeed   int64  `json:"round_random_seed"`
	RoundTimeoutCount int    `json:"round_timeout_count"`

	Hash            string  `json:"hash"`
	Signature       string  `json:"signature"`
	ChainID         string  `json:"chain_id"`
	ChainWeight     float64 `json:"chain_weight"`
	RunningTxnCount int64   `json:"running_txn_count"`

	Version      string `json:"version"`
	CreationDate int64  `json:"creation_date"`

	MagicBlockHash string `json:"magic_block_hash"`
	PrevHash       string `json:"prev_hash"`

	ClientStateHash string `json:"state_hash"`

	// unexported fields
	header *BlockHeader         `json:"-"`
	txns   []*TransactionMobile `json:"transactions,omitempty"`
}

func (b *Block) GetHeader() *BlockHeader {
	return b.header
}

type IterTxnFunc func(idx int, txn *TransactionMobile)

type Transactions struct {
	txns []*TransactionMobile
}

func (tm *Transactions) Len() int {
	return len(tm.txns)
}

func (tm *Transactions) Get(idx int) (*TransactionMobile, error) {
	if idx < 0 && idx >= len(tm.txns) {
		return nil, stderrors.New("index out of bounds")
	}

	return tm.txns[idx], nil
}

func (b *Block) GetTxns() *Transactions {
	return &Transactions{
		txns: b.txns,
	}
}

func toMobileBlock(b *block.Block) *Block {
	lb := &Block{
		header: &BlockHeader{
			Version:               b.Header.Version,
			CreationDate:          b.Header.CreationDate,
			Hash:                  b.Header.Hash,
			MinerID:               b.Header.MinerID,
			Round:                 b.Header.Round,
			RoundRandomSeed:       b.Header.RoundRandomSeed,
			MerkleTreeRoot:        b.Header.MerkleTreeRoot,
			StateHash:             b.Header.StateHash,
			ReceiptMerkleTreeRoot: b.Header.ReceiptMerkleTreeRoot,
			NumTxns:               b.Header.NumTxns,
		},
		MinerID:           string(b.MinerID),
		Round:             b.Round,
		RoundRandomSeed:   b.RoundRandomSeed,
		RoundTimeoutCount: b.RoundTimeoutCount,

		Hash:            string(b.Hash),
		Signature:       b.Signature,
		ChainID:         string(b.ChainID),
		ChainWeight:     b.ChainWeight,
		RunningTxnCount: b.RunningTxnCount,

		Version:      b.Version,
		CreationDate: int64(b.CreationDate),

		MagicBlockHash: b.MagicBlockHash,
		PrevHash:       b.PrevHash,

		ClientStateHash: string(b.ClientStateHash),
	}

	lb.txns = make([]*TransactionMobile, len(b.Txns))
	for i, txn := range b.Txns {
		lb.txns[i] = &TransactionMobile{
			Hash:              txn.Hash,
			Version:           txn.Version,
			ClientID:          txn.ClientID,
			PublicKey:         txn.PublicKey,
			ToClientID:        txn.ToClientID,
			ChainID:           txn.ChainID,
			TransactionData:   txn.TransactionData,
			Signature:         txn.Signature,
			CreationDate:      txn.CreationDate,
			TransactionType:   txn.TransactionType,
			TransactionOutput: txn.TransactionOutput,
			TransactionNonce:  txn.TransactionNonce,
			OutputHash:        txn.OutputHash,
			Status:            txn.Status,
			Value:             strconv.FormatUint(txn.Value, 10),
			TransactionFee:    strconv.FormatUint(txn.TransactionFee, 10),
		}
	}

	return lb
}

//TransactionMobile entity that encapsulates the transaction related data and meta data
type TransactionMobile struct {
	Hash              string `json:"hash,omitempty"`
	Version           string `json:"version,omitempty"`
	ClientID          string `json:"client_id,omitempty"`
	PublicKey         string `json:"public_key,omitempty"`
	ToClientID        string `json:"to_client_id,omitempty"`
	ChainID           string `json:"chain_id,omitempty"`
	TransactionData   string `json:"transaction_data"`
	Value             string `json:"transaction_value"`
	Signature         string `json:"signature,omitempty"`
	CreationDate      int64  `json:"creation_date,omitempty"`
	TransactionType   int    `json:"transaction_type"`
	TransactionOutput string `json:"transaction_output,omitempty"`
	TransactionFee    string `json:"transaction_fee"`
	TransactionNonce  int64  `json:"transaction_nonce"`
	OutputHash        string `json:"txn_output_hash"`
	Status            int    `json:"transaction_status"`
}

// RequestTimeout will be used for setting requests with timeout
type RequestTimeout interface {
	Set(int64)  // milliseconds
	Get() int64 // milliseconds
}

type timeoutCtx struct {
	millisecond int64
}

func NewRequestTimeout(timeout int64) RequestTimeout {
	return &timeoutCtx{millisecond: timeout}
}

func (t *timeoutCtx) Set(tm int64) {
	t.millisecond = tm
}

func (t *timeoutCtx) Get() int64 {
	return t.millisecond
}

func GetBlockByRound(numSharders int, round int64, timeout RequestTimeout) (b *Block, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	ctx, cancel := makeTimeoutContext(timeout)
	defer cancel()

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

		b = toMobileBlock(respo.Block)

		b.header = &BlockHeader{
			Version:               respo.Header.Version,
			CreationDate:          respo.Header.CreationDate,
			Hash:                  respo.Header.Hash,
			MinerID:               respo.Header.MinerID,
			Round:                 respo.Header.Round,
			RoundRandomSeed:       respo.Header.RoundRandomSeed,
			MerkleTreeRoot:        respo.Header.MerkleTreeRoot,
			StateHash:             respo.Header.StateHash,
			ReceiptMerkleTreeRoot: respo.Header.ReceiptMerkleTreeRoot,
			NumTxns:               respo.Header.NumTxns,
		}

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

func GetMagicBlockByNumber(numSharders int, number int64, timeout RequestTimeout) ([]byte, error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	ctx, cancel := makeTimeoutContext(timeout)
	defer cancel()

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders,
		fmt.Sprintf("%smagic_block_number=%d", GET_MAGIC_BLOCK_INFO, number),
		result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
		ret            []byte
		err            error
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

		ret = []byte(rsp.Body)
		var h = encryption.FastHash([]byte(respo.MagicBlock.Hash))
		if roundConsensus[h]++; roundConsensus[h] > maxConsensus {
			maxConsensus = roundConsensus[h]
		}
	}

	if maxConsensus == 0 {
		return nil, errors.New("", "magic block info not found")
	}

	if err != nil {
		return nil, err
	}

	return ret, nil
}
