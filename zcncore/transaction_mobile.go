//go:build mobile
// +build mobile

package zcncore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/block"
	"github.com/0chain/gosdk/core/encryption"
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

	VestingAdd(ar *VestingAddRequest, value string) error

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

type VestingAddRequest struct {
	Description  string         `json:"description"`  // allow empty
	StartTime    int64          `json:"start_time"`   //
	Duration     int64          `json:"duration"`     //
	Destinations []*VestingDest `json:"destinations"` //
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

func (t *Transaction) VestingAdd(ar *VestingAddRequest, value string) (
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
		header:            b.Header,
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
