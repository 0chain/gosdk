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
	"strings"
	"time"

	"github.com/0chain/errors"
	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/block"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
)

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
	UpdateBlobberSettings(blobber *Blobber, fee string) error
	UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock, fee string) error
	WritePoolLock(allocID string, blobberID string, duration int64, lock, fee string) error
	WritePoolUnlock(poolID string, fee string) error

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
	Blobbers       string // blobber urls combined with ','
	ReadPriceMin   int64
	ReadPriceMax   int64
	WritePriceMin  int64
	WritePriceMax  int64
}

func (car *CreateAllocationRequest) toCreateAllocationSCInput() *createAllocationRequest {
	return &createAllocationRequest{
		DataShards:      car.DataShards,
		ParityShards:    car.ParityShards,
		Size:            car.Size,
		Expiration:      car.Expiration,
		Owner:           car.Owner,
		OwnerPublicKey:  car.OwnerPublicKey,
		Blobbers:        strings.Split(car.Blobbers, ","),
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

type Blobber struct {
	ID                string            `json:"id"`
	BaseURL           string            `json:"url"`
	Terms             Terms             `json:"terms"`
	Capacity          int64             `json:"capacity"`
	Allocated         int64             `json:"allocated"`
	LastHealthCheck   int64             `json:"last_health_check"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

type AddAuthorizerPayload struct {
	PublicKey         string                      `json:"public_key"`
	URL               string                      `json:"url"`
	StakePoolSettings AuthorizerStakePoolSettings `json:"stake_pool_settings"` // Used to initially create stake pool
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
	AddDestinations(dest *VestingDest)
}

func NewVestingAddRequest(desc string, startTime int64, duration int64) VestingAddRequest {
	return &vestingAddRequest{
		description: desc,
		startTime:   startTime,
		duration:    duration,
	}
}

type vestingAddRequest struct {
	description  string         `json:"description"`  // allow empty
	startTime    int64          `json:"start_time"`   //
	duration     int64          `json:"duration"`     //
	destinations []*VestingDest `json:"destinations"` //
}

func (vr *vestingAddRequest) AddDestinations(dest *VestingDest) {
	vr.destinations = append(vr.destinations, dest)
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
	v, err := strconv.ParseUint(vs, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid token value: %v", vs)
	}

	if v/uint64(TOKEN_UNIT) == 0 {
		return 0, fmt.Errorf("token value must be multiple value of 1e10")
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
		Logger.Info("New transaction interface with auth")
		return newTransactionWithAuth(cb, v, nonce)
	}
	Logger.Info("New transaction interface")
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
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
		return err
	}
	t.setTransactionFee(v)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// UpdateBlobberSettings update settings of a blobber.
func (t *Transaction) UpdateBlobberSettings(b *Blobber, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, b, 0)
	if err != nil {
		Logger.Error(err)
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
		Logger.Error(err)
		return err
	}
	t.setTransactionFee(fv)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// WritePoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) WritePoolLock(allocID, blobberID string, duration int64,
	lock, fee string) error {
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
		transaction.STORAGESC_WRITE_POOL_LOCK, &lr, lv)
	if err != nil {
		Logger.Error(err)
		return err
	}
	t.setTransactionFee(fv)
	go func() { t.setNonceAndSubmit() }()
	return nil
}

// WritePoolUnlock for current user and given pool.
func (t *Transaction) WritePoolUnlock(poolID string, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_UNLOCK, &unlockRequest{
			PoolID: poolID,
		}, 0)
	if err != nil {
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
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
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerScUpdateGlobals(ip InputMap) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_GLOBALS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) StorageScUpdateConfig(ip InputMap) (err error) {
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) ZCNSCUpdateGlobalConfig(ip InputMap) (err error) {
	err = t.createSmartContractTxn(ZCNSCSmartContractAddress,
		transaction.ZCNSC_UPDATE_GLOBAL_CONFIG, ip, 0)
	if err != nil {
		Logger.Error(err)
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
	StakingPoolSettings() StakePoolSettings
}

func NewMinerSCMinerInfo(id string, settings StakePoolSettings) MinerSCMinerInfo {
	return &minerSCMinerInfo{
		simpleMiner: simpleMiner{ID: id},
		minerSCDelegatePool: minerSCDelegatePool{
			Settings: settings,
		},
	}
}

type minerSCDelegatePool struct {
	Settings StakePoolSettings `json:"settings"`
}

type simpleMiner struct {
	ID string `json:"id"`
}

type minerSCMinerInfo struct {
	simpleMiner         `json:"simple_miner"`
	minerSCDelegatePool `json:"stake_pool"`
}

func (mi *minerSCMinerInfo) GetID() string {
	return mi.ID
}

func (mi *minerSCMinerInfo) StakingPoolSettings() StakePoolSettings {
	return mi.Settings
}

func (t *Transaction) MinerSCMinerSettings(info MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_SETTINGS, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCSharderSettings(info MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_SETTINGS, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCDeleteMiner(info MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_DELETE, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCDeleteSharder(info MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_DELETE, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
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
		Logger.Error(err)
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
					Logger.Info(err, " now: ", now)
				} else {
					Logger.Info(err, " now: ", now, ", LFB creation time:", lfbBlockHeader.CreationDate)
				}

				// transaction is done or expired. it means random sharder might be outdated, try to query it from s/S sharders to confirm it
				if util.MaxInt64(lfbBlockHeader.getCreationDate(now), now) >= (t.txn.CreationDate + int64(defaultTxnExpirationSeconds)) {
					Logger.Info("falling back to ", getMinShardersVerify(), " of ", len(_config.chain.Sharders), " Sharders")
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

// ConvertToValue converts ZCN tokens to value
func ConvertToValue(token float64) string {
	return strconv.FormatUint(uint64(token*float64(TOKEN_UNIT)), 10)
}

type ReqTimeout struct {
	Milliseconds int64
}

func makeTimeoutContext(tm *ReqTimeout) (context.Context, func()) {
	var (
		ctx    context.Context
		cancel func()
	)

	if tm != nil && tm.Milliseconds > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*time.Duration(tm.Milliseconds))
	} else {
		ctx = context.Background()
	}

	return ctx, cancel
}

func GetLatestFinalized(numSharders int, tm *ReqTimeout) (b *BlockHeader, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	ctx, cancel := makeTimeoutContext(tm)
	defer cancel()

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders, GET_LATEST_FINALIZED, result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	for i := 0; i < numSharders; i++ {
		var rsp = <-result

		Logger.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			Logger.Error(rsp.Body)
			continue
		}

		if err = json.Unmarshal([]byte(rsp.Body), &b); err != nil {
			Logger.Error("block parse error: ", err)
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

func GetLatestFinalizedMagicBlock(numSharders int, tm *ReqTimeout) (m *block.MagicBlock, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	ctx, cancel := makeTimeoutContext(tm)
	defer cancel()

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

		Logger.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			Logger.Error(rsp.Body)
			continue
		}

		var respo respObj
		if err = json.Unmarshal([]byte(rsp.Body), &respo); err != nil {
			Logger.Error(" magic block parse error: ", err)
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

func GetChainStats(tm *ReqTimeout) (b *block.ChainStats, err error) {
	var result = make(chan *util.GetResponse, 1)
	defer close(result)

	ctx, cancel := makeTimeoutContext(tm)
	defer cancel()

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

// ForEachTxns iterates over all block.Txns as gomobine does not support slices
// for most of the data struct, so to get the transactions in a block, the caller
// needs to define the
func (b *Block) ForEachTxns(tf IterTxnFunc) {
	for i, t := range b.txns {
		tf(i, t)
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

//Transaction entity that encapsulates the transaction related data and meta data
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

func GetBlockByRound(numSharders int, round int64, tm *ReqTimeout) (b *Block, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	ctx, cancel := makeTimeoutContext(tm)
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

		Logger.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			Logger.Error(rsp.Body)
			continue
		}

		var respo respObj
		if err = json.Unmarshal([]byte(rsp.Body), &respo); err != nil {
			Logger.Error("block parse error: ", err)
			err = nil
			continue
		}

		if respo.Block == nil {
			Logger.Debug(rsp.Url, "no block in response:", rsp.Body)
			continue
		}

		if respo.Header == nil {
			Logger.Debug(rsp.Url, "no block header in response:", rsp.Body)
			continue
		}

		if respo.Header.Hash != string(respo.Block.Hash) {
			Logger.Debug(rsp.Url, "header and block hash mismatch:", rsp.Body)
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

func GetMagicBlockByNumber(numSharders int, number int64, tm *ReqTimeout) (m *block.MagicBlock, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	ctx, cancel := makeTimeoutContext(tm)
	defer cancel()

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

		Logger.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			Logger.Error(rsp.Body)
			continue
		}

		var respo respObj
		if err = json.Unmarshal([]byte(rsp.Body), &respo); err != nil {
			Logger.Error(" magic block parse error: ", err)
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

// FromAll query transaction from all sharders whatever it is selected or offline in previous queires, and return consensus result
func (tq *TransactionQuery) FromAll(query string, handle QueryResultHandle, tm *ReqTimeout) error {
	if tq == nil || tq.max == 0 {
		return ErrNoAvailableSharders
	}

	ctx, cancel := makeTimeoutContext(tm)
	defer cancel()

	urls := make([]string, 0, tq.max)
	for _, host := range tq.sharders {
		urls = append(urls, tq.buildUrl(host, query))
	}

	r := resty.New()
	r.DoGet(ctx, urls...).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			res := QueryResult{
				Content:    respBody,
				Error:      err,
				StatusCode: http.StatusBadRequest,
			}

			if resp != nil {
				res.StatusCode = resp.StatusCode

				Logger.Debug(req.URL.String() + " " + resp.Status)
				Logger.Debug(string(respBody))
			} else {
				Logger.Debug(req.URL.String())

			}

			if handle != nil {
				if handle(res) {

					cf()
				}
			}

			return nil
		})

	r.Wait()

	return nil
}

// FromAny query transaction from any sharder that is not selected in previous queires. use any used sharder if there is not any unused sharder
func (tq *TransactionQuery) FromAny(query string, tm *ReqTimeout) (QueryResult, error) {
	res := QueryResult{
		StatusCode: http.StatusBadRequest,
	}

	ctx, cancel := makeTimeoutContext(tm)
	defer cancel()

	err := tq.validate(1)

	if err != nil {
		return res, err
	}

	host, err := tq.randOne(ctx)

	if err != nil {
		return res, err
	}

	r := resty.New()
	requestUrl := tq.buildUrl(host, query)

	Logger.Debug("GET", requestUrl)

	r.DoGet(ctx, requestUrl).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			res.Error = err
			if err != nil {
				return err
			}

			res.Content = respBody
			Logger.Debug(string(respBody))

			if resp != nil {
				res.StatusCode = resp.StatusCode
			}

			return nil
		})

	errs := r.Wait()

	if len(errs) > 0 {
		return res, errs[0]
	}

	return res, nil

}

func (tq *TransactionQuery) GetInfo(query string, tm *ReqTimeout) (*QueryResult, error) {

	consensuses := make(map[int]int)
	var maxConsensus int
	var consensusesResp QueryResult
	// {host}{query}

	err := tq.FromAll(query,
		func(qr QueryResult) bool {
			//ignore response if it is network error
			if qr.StatusCode >= 500 {
				return false
			}

			consensuses[qr.StatusCode]++
			if consensuses[qr.StatusCode] >= maxConsensus {
				maxConsensus = consensuses[qr.StatusCode]
				consensusesResp = qr
			}

			return false

		}, tm)

	if err != nil {
		return nil, err
	}

	if maxConsensus == 0 {
		return nil, errors.New("zcn: query not found")
	}

	rate := float32(maxConsensus*100) / float32(tq.max)
	if rate < consensusThresh {
		return nil, ErrInvalidConsensus
	}

	if consensusesResp.StatusCode != http.StatusOK {
		return nil, stderrors.New(string(consensusesResp.Content))
	}

	return &consensusesResp, nil
}

func (tq *TransactionQuery) getConsensusConfirmation(numSharders int, txnHash string, tm *ReqTimeout) (*blockHeader, map[string]json.RawMessage, *blockHeader, error) {
	maxConfirmation := int(0)
	txnConfirmations := make(map[string]int)
	var confirmationBlockHeader *blockHeader
	var confirmationBlock map[string]json.RawMessage
	var lfbBlockHeader *blockHeader
	maxLfbBlockHeader := int(0)
	lfbBlockHeaders := make(map[string]int)

	// {host}/v1/transaction/get/confirmation?hash={txnHash}&content=lfb
	err := tq.FromAll(tq.buildUrl("", TXN_VERIFY_URL, txnHash, "&content=lfb"),
		func(qr QueryResult) bool {
			if qr.StatusCode != http.StatusOK {
				return false
			}

			var cfmBlock map[string]json.RawMessage
			err := json.Unmarshal([]byte(qr.Content), &cfmBlock)
			if err != nil {
				Logger.Error("txn confirmation parse error", err)
				return false
			}

			// parse `confirmation` section as block header
			cfmBlockHeader, err := getBlockHeaderFromTransactionConfirmation(txnHash, cfmBlock)
			if err != nil {
				Logger.Error("txn confirmation parse header error", err)

				// parse `latest_finalized_block` section
				if lfbRaw, ok := cfmBlock["latest_finalized_block"]; ok {
					var lfb blockHeader
					err := json.Unmarshal([]byte(lfbRaw), &lfb)
					if err != nil {
						Logger.Error("round info parse error.", err)
						return false
					}

					lfbBlockHeaders[lfb.Hash]++
					if lfbBlockHeaders[lfb.Hash] > maxLfbBlockHeader {
						maxLfbBlockHeader = lfbBlockHeaders[lfb.Hash]
						lfbBlockHeader = &lfb
					}
				}

				return false
			}

			txnConfirmations[cfmBlockHeader.Hash]++
			if txnConfirmations[cfmBlockHeader.Hash] > maxConfirmation {
				maxConfirmation = txnConfirmations[cfmBlockHeader.Hash]

				if maxConfirmation >= numSharders {
					confirmationBlockHeader = cfmBlockHeader
					confirmationBlock = cfmBlock

					// it is consensus by enough sharders, and latest_finalized_block is valid
					// return true to cancel other requests
					return true
				}
			}

			return false

		}, tm)

	if err != nil {
		return nil, nil, lfbBlockHeader, err
	}

	if maxConfirmation == 0 {
		return nil, nil, lfbBlockHeader, errors.New("zcn: transaction not found")
	}

	if maxConfirmation < numSharders {
		return nil, nil, lfbBlockHeader, ErrInvalidConsensus
	}

	return confirmationBlockHeader, confirmationBlock, lfbBlockHeader, nil
}

// getFastConfirmation get txn confirmation from a random online sharder
func (tq *TransactionQuery) getFastConfirmation(ctx context.Context, txnHash string) (*blockHeader, map[string]json.RawMessage, *blockHeader, error) {
	var confirmationBlockHeader *blockHeader
	var confirmationBlock map[string]json.RawMessage
	var lfbBlockHeader blockHeader

	// {host}/v1/transaction/get/confirmation?hash={txnHash}&content=lfb
	result, err := tq.FromAny(ctx, tq.buildUrl("", TXN_VERIFY_URL, txnHash, "&content=lfb"))
	if err != nil {
		return nil, nil, nil, err
	}

	if result.StatusCode == http.StatusOK {

		err = json.Unmarshal(result.Content, &confirmationBlock)
		if err != nil {
			Logger.Error("txn confirmation parse error", err)
			return nil, nil, nil, err
		}

		// parse `confirmation` section as block header
		confirmationBlockHeader, err = getBlockHeaderFromTransactionConfirmation(txnHash, confirmationBlock)
		if err == nil {
			return confirmationBlockHeader, confirmationBlock, nil, nil
		}

		Logger.Error("txn confirmation parse header error", err)

		// parse `latest_finalized_block` section
		lfbRaw, ok := confirmationBlock["latest_finalized_block"]
		if !ok {
			return confirmationBlockHeader, confirmationBlock, nil, err
		}

		err = json.Unmarshal([]byte(lfbRaw), &lfbBlockHeader)
		if err == nil {
			return confirmationBlockHeader, confirmationBlock, &lfbBlockHeader, ErrTransactionNotConfirmed
		}

		Logger.Error("round info parse error.", err)
		return nil, nil, nil, err

	}

	return nil, nil, nil, thrown.Throw(ErrTransactionNotFound, strconv.Itoa(result.StatusCode))
}
