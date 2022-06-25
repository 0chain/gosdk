//go:build mobile
// +build mobile

package zcncore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/block"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
)

type TransactionCommon interface {
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input string, val string) error

	// Send implements sending token to a given clientid
	Send(toClientID string, val string, desc string) error
	// SetTransactionFee implements method to set the transaction fee
	SetTransactionFee(txnFee string) error

	VestingAdd(ar *VestingAddRequest, value string) error

	MinerSCLock(minerID string, lock string) error

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
}

// PriceRange represents a price range allowed by user to filter blobbers.
type PriceRange struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

// CreateAllocationRequest is information to create allocation.
type CreateAllocationRequest struct {
	DataShards      int        `json:"data_shards"`
	ParityShards    int        `json:"parity_shards"`
	Size            int64      `json:"size"`
	Expiration      int64      `json:"expiration_date"`
	Owner           string     `json:"owner_id"`
	OwnerPublicKey  string     `json:"owner_public_key"`
	Blobbers        []string   `json:"blobbers"`
	ReadPriceRange  PriceRange `json:"read_price_range"`
	WritePriceRange PriceRange `json:"write_price_range"`
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
func (t *Transaction) CreateAllocation(car *CreateAllocationRequest,
	lock, fee string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	fv, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_ALLOCATION, car, lv)
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
		return er
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

// ConvertToValue converts ZCN tokens to value
func ConvertToValue(token float64) string {
	return strconv.FormatUint(uint64(token*float64(TOKEN_UNIT)), 10)
}

type ReqTimeout struct {
	Milliseconds int64
}

func GetLatestFinalized(numSharders int, tm *ReqTimeout) (b *block.Header, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	var (
		ctx    context.Context
		cancel func()
	)

	if tm != nil && tm.Milliseconds > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*tm.Milliseconds)
	} else {
		ctx, cancel = context.Background()
	}
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

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders, GET_LATEST_FINALIZED_MAGIC_BLOCK, result)

	var (
		ctx    context.Context
		cancel func()
	)

	if tm != nil && tm.Milliseconds > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*tm.Milliseconds)
	} else {
		ctx, cancel = context.Background()
	}
	defer cancel()

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

	var (
		ctx    context.Context
		cancel func()
	)

	if tm != nil && tm.Milliseconds > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*tm.Milliseconds)
	} else {
		ctx, cancel = context.Background()
	}
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

func GetBlockByRound(numSharders int, round int64, tm *ReqTimeout) (b *block.Block, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	var (
		ctx    context.Context
		cancel func()
	)

	if tm != nil && tm.Milliseconds > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*tm.Milliseconds)
	} else {
		ctx, cancel = context.Background()
	}
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

func GetMagicBlockByNumber(numSharders int, number int64, tm *ReqTimeout) (m *block.MagicBlock, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	var (
		ctx    context.Context
		cancel func()
	)

	if tm != nil && tm.Milliseconds > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*tm.Milliseconds)
	} else {
		ctx, cancel = context.Background()
	}
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
