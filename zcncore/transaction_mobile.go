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
	"github.com/0chain/gosdk/core/node"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
)

const (
	Undefined int = iota
	Success

	// ChargeableError is an error that still charges the user for the transaction.
	ChargeableError
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

type stakePoolRequest struct {
	ProviderType int    `json:"provider_type,omitempty"`
	ProviderID   string `json:"provider_id,omitempty"`
}

type TransactionCommon interface {
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input interface{}, val uint64, feeOpts ...FeeOption) (*transaction.Transaction, error)

	// Send implements sending token to a given clientid
	Send(toClientID string, val uint64, desc string) error

	VestingAdd(ar VestingAddRequest, value string) error

	MinerSCLock(providerId string, providerType int, lock string) error
	MinerSCUnlock(providerId string, providerType int) error
	MinerSCCollectReward(providerId string, providerType int) error
	StorageSCCollectReward(providerId string, providerType int) error

	FinalizeAllocation(allocID string) error
	CancelAllocation(allocID string) error
	CreateAllocation(car *CreateAllocationRequest, lock string) error //
	CreateReadPool() error
	ReadPoolLock(allocID string, blobberID string, duration int64, lock string) error
	ReadPoolUnlock() error
	StakePoolLock(providerId string, providerType int, lock string) error
	StakePoolUnlock(providerId string, providerType int) error
	UpdateBlobberSettings(blobber Blobber) error
	UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock string) error
	WritePoolLock(allocID string, lock string) error
	WritePoolUnlock(allocID string) error

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
	NumDelegates   int     `json:"num_delegates"`
	ServiceCharge  float64 `json:"service_charge"`
}

type Terms struct {
	ReadPrice        int64 `json:"read_price"`  // tokens / GB
	WritePrice       int64 `json:"write_price"` // tokens / GB
	MaxOfferDuration int64 `json:"max_offer_duration"`
}

type Blobber interface {
	SetTerms(readPrice int64, writePrice int64, minLockDemand float64, maxOfferDuration int64)
	SetStakePoolSettings(delegateWallet string, numDelegates int, serviceCharge float64)
	SetAvailable(bool)
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
	NotAvailable      bool              `json:"not_available"`
}

func (b *blobber) SetStakePoolSettings(delegateWallet string, numDelegates int, serviceCharge float64) {
	b.StakePoolSettings = StakePoolSettings{
		DelegateWallet: delegateWallet,
		NumDelegates:   numDelegates,
		ServiceCharge:  serviceCharge,
	}
}

func (b *blobber) SetTerms(readPrice int64, writePrice int64, minLockDemand float64, maxOfferDuration int64) {
	b.Terms = Terms{
		ReadPrice:        readPrice,
		WritePrice:       writePrice,
		MaxOfferDuration: maxOfferDuration,
	}
}

func (b *blobber) SetAvailable(availability bool) {
	b.NotAvailable = availability
}

type Validator interface {
	SetStakePoolSettings(delegateWallet string, numDelegates int, serviceCharge float64)
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

func (v *validator) SetStakePoolSettings(delegateWallet string, numDelegates int, serviceCharge float64) {
	v.StakePoolSettings = StakePoolSettings{
		DelegateWallet: delegateWallet,
		NumDelegates:   numDelegates,
		ServiceCharge:  serviceCharge,
	}
}

// AddAuthorizerPayload is the interface gathering the functions to add a new authorizer.
type AddAuthorizerPayload interface {
	// SetStakePoolSettings sets the stake pool settings for the authorizer.
	SetStakePoolSettings(delegateWallet string, numDelegates int, serviceCharge float64)
}

// NewAddAuthorizerPayload creates a new AddAuthorizerPayload concrete instance.
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

// SetStakePoolSettings sets the stake pool settings for the authorizer.
func (a *addAuthorizerPayload) SetStakePoolSettings(delegateWallet string, numDelegates int, serviceCharge float64) {
	a.StakePoolSettings = AuthorizerStakePoolSettings{
		DelegateWallet: delegateWallet,
		NumDelegates:   numDelegates,
		ServiceCharge:  serviceCharge,
	}
}

type AuthorizerHealthCheckPayload struct {
	ID string `json:"id"` // authorizer ID
}

// AuthorizerStakePoolSettings represents configuration of an authorizer stake pool.
type AuthorizerStakePoolSettings struct {
	DelegateWallet string  `json:"delegate_wallet"`
	NumDelegates   int     `json:"num_delegates"`
	ServiceCharge  float64 `json:"service_charge"`
}

// AuthorizerConfig represents configuration of an authorizer node.
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

// InputMap represents an interface of functions to add fields to a map.
type InputMap interface {
	// AddField adds a field to the map.
	// 		- key: field key
	// 		- value: field value
	AddField(key, value string)
}

type inputMap struct {
	Fields map[string]string `json:"fields"`
}

// NewInputMap creates a new InputMap concrete instance.
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

// ExecuteSmartContract prepare and send a smart contract transaction to the blockchain
func (t *Transaction) ExecuteSmartContract(address, methodName string, input interface{}, val uint64, feeOpts ...FeeOption) (*transaction.Transaction, error) {
	// t.createSmartContractTxn(address, methodName, input, val, opts...)
	err := t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return nil, err
	}
	go func() {
		t.setNonceAndSubmit()
	}()

	return t.txn, nil
}

func (t *Transaction) setTransactionFee(fee uint64) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction fee.")
	}
	t.txn.TransactionFee = fee
	return nil
}

// Send to send a transaction to a given clientID
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

// SendWithSignatureHash to send a transaction to a given clientID with a signature hash
//   - toClientID: client ID in the To field of the transaction
//   - val: amount of tokens to send
//   - desc: description of the transaction
//   - sig: signature hash
//   - CreationDate: creation date of the transaction
//   - hash: hash of the transaction
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

func (t *Transaction) MinerSCLock(providerId string, providerType int, lock string) error {

	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	pr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerId,
	}
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, pr, lv)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return err
}

func (t *Transaction) MinerSCUnlock(providerId string, providerType int) error {
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

func (t *Transaction) MinerSCCollectReward(providerId string, providerType int) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
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

func (t *Transaction) StorageSCCollectReward(providerId string, providerType int) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
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
func (t *Transaction) FinalizeAllocation(allocID string) (err error) {
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
func (t *Transaction) CancelAllocation(allocID string) error {
	type cancelRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err := t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CANCEL_ALLOCATION, &cancelRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// CreateAllocation transaction.
func (t *Transaction) CreateAllocation(car *CreateAllocationRequest, lock string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_ALLOCATION, car.toCreateAllocationSCInput(), lv)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// CreateReadPool for current user.
func (t *Transaction) CreateReadPool() error {
	err := t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_READ_POOL, nil, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// ReadPoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) ReadPoolLock(allocID, blobberID string,
	duration int64, lock string) error {
	lv, err := parseCoinStr(lock)
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
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// ReadPoolUnlock for current user and given pool.
func (t *Transaction) ReadPoolUnlock() error {
	err := t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_UNLOCK, nil, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// StakePoolLock used to lock tokens in a stake pool of a blobber.
func (t *Transaction) StakePoolLock(providerId string, providerType int, lock string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	spr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerId,
	}

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_LOCK, &spr, lv)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// StakePoolUnlock by blobberID
func (t *Transaction) StakePoolUnlock(providerId string, providerType int) error {
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
func (t *Transaction) UpdateBlobberSettings(b Blobber) error {
	err := t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, b, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// UpdateAllocation transaction.
func (t *Transaction) UpdateAllocation(allocID string, sizeDiff int64,
	expirationDiff int64, lock string) error {
	lv, err := parseCoinStr(lock)
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
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// WritePoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) WritePoolLock(allocID, lock string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	var lr = struct {
		AllocationID string `json:"allocation_id"`
	}{
		AllocationID: allocID,
	}

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_LOCK, &lr, lv)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// WritePoolUnlock for current user and given pool.
func (t *Transaction) WritePoolUnlock(allocID string) error {
	var ur = struct {
		AllocationID string `json:"allocation_id"`
	}{
		AllocationID: allocID,
	}

	err := t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_UNLOCK, &ur, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
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

// MinerSCMinerInfo interface for miner info functions on miner smart contract.
type MinerSCMinerInfo interface {
	// GetID returns the ID of the miner
	GetID() string
}

// NewMinerSCMinerInfo creates a new miner info.
//   - id: miner ID
//   - delegateWallet: delegate wallet
//   - minStake: minimum stake
//   - maxStake: maximum stake
//   - numDelegates: number of delegates
//   - serviceCharge: service charge
func NewMinerSCMinerInfo(id string, delegateWallet string,
	minStake int64, maxStake int64, numDelegates int, serviceCharge float64) MinerSCMinerInfo {
	return &minerSCMinerInfo{
		simpleMiner: simpleMiner{ID: id},
		minerSCDelegatePool: minerSCDelegatePool{
			Settings: StakePoolSettings{
				DelegateWallet: delegateWallet,
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

// AuthorizerNode interface for authorizer node functions.
type AuthorizerNode interface {
	// GetID returns the ID of the authorizer node.
	GetID() string
}

// NewAuthorizerNode creates a new authorizer node.
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
			node.Cache.Evict(t.txn.ClientID)
			return errors.New("", "invalid transaction. cannot be verified.")
		}
	}
	// If transaction is verify only start from current time
	if t.txn.CreationDate == 0 {
		t.txn.CreationDate = int64(common.Now())
	}

	tq, err := newTransactionQuery(Sharders.Healthy())
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
					logging.Info("falling back to ", getMinShardersVerify(), " of ", len(_config.chain.Sharders), " Sharders", len(Sharders.Healthy()), "Healthy sharders")
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

// EstimateFee estimates transaction fee
func (t *Transaction) EstimateFee(reqPercent float32) (int64, error) {
	fee, err := transaction.EstimateFee(t.txn, _config.chain.Miners, reqPercent)
	return int64(fee), err
}

// ConvertTokenToSAS converts ZCN tokens to SAS tokens
// # Inputs
//   - token: ZCN tokens
func ConvertTokenToSAS(token float64) uint64 {
	return uint64(token * common.TokenUnit)
}

// ConvertToValue converts ZCN tokens to SAS tokens with string format
//   - token: ZCN tokens
func ConvertToValue(token float64) string {
	return strconv.FormatUint(ConvertTokenToSAS(token), 10)
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
func GetLatestFinalizedMagicBlock(numSharders int, timeout RequestTimeout) ([]byte, error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	ctx, cancel := makeTimeoutContext(timeout)
	defer cancel()

	numSharders = len(Sharders.Healthy()) // overwrite, use all
	Sharders.QueryFromShardersContext(ctx, numSharders, GET_LATEST_FINALIZED_MAGIC_BLOCK, result)

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

	return []byte(rsp.Body), nil
}

func GetFeeStats(timeout RequestTimeout) ([]byte, error) {

	var numMiners = 4

	if numMiners > len(_config.chain.Miners) {
		numMiners = len(_config.chain.Miners)
	}

	var result = make(chan *util.GetResponse, numMiners)

	ctx, cancel := makeTimeoutContext(timeout)
	defer cancel()

	var (
		b   *block.FeeStats
		err error
	)

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
			return nil, ctx.Err()
		}
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

// TransactionMobile entity that encapsulates the transaction related data and meta data
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

	numSharders = len(Sharders.Healthy()) // overwrite, use all
	Sharders.QueryFromShardersContext(ctx, numSharders,
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

	numSharders = len(Sharders.Healthy()) // overwrite, use all
	Sharders.QueryFromShardersContext(ctx, numSharders,
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

// GetFeesTable get fee tables
func GetFeesTable(reqPercent float32) (string, error) {

	fees, err := transaction.GetFeesTable(_config.chain.Miners, reqPercent)
	if err != nil {
		return "", err
	}

	js, err := json.Marshal(fees)
	if err != nil {
		return "", err
	}

	return string(js), nil
}
