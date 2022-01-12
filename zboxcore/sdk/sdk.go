package sdk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/logger"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const STORAGE_SCADDRESS = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"

var sdkNotInitialized = errors.New("sdk_not_initialized", "SDK is not initialised")

const (
	OpUpload   int = 0
	OpDownload int = 1
	OpRepair   int = 2
	OpUpdate   int = 3
)

type StatusCallback interface {
	Started(allocationId, filePath string, op int, totalBytes int)
	InProgress(allocationId, filePath string, op int, completedBytes int, data []byte)
	Error(allocationID string, filePath string, op int, err error)
	Completed(allocationId, filePath string, filename string, mimetype string, size int, op int)
	CommitMetaCompleted(request, response string, txn *transaction.Transaction, err error)
	RepairCompleted(filesRepaired int)
}

var numBlockDownloads = 10
var sdkInitialized = false
var networkWorkerTimerInHours = 1

// GetVersion - returns version string
func GetVersion() string {
	return version.VERSIONSTR
}

// SetLogLevel set the log level.
// lvl - 0 disabled; higher number (upto 4) more verbosity
func SetLogLevel(lvl int) {
	Logger.SetLevel(lvl)
}

// SetLogFile
// logFile - Log file
// verbose - true - console output; false - no console output
func SetLogFile(logFile string, verbose bool) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	Logger.SetLogFile(f, verbose)
	Logger.Info("******* Storage SDK Version: ", version.VERSIONSTR, " *******")
}

func GetLogger() *logger.Logger {
	return &Logger
}

// InitStorageSDK init storage sdk with walletJSON
//   {
//		"client_id":"322d1dadec182effbcbdeef77d84f",
//		"client_key":"3b6d02a22ec82d4d9aa1402917ca2",
//		"keys":[{
//			"public_key":"3b6d02a22ec82d4d9aa1402917ca268",
//			"private_key":"25f2e1355d3864de01aba0bfec3702"
//			}],
//		"mnemonics":"double wink spin mushroom thing notable trumpet chapter",
//		"version":"1.0",
//		"date_created":"2021-08-18T08:34:39+08:00"
//	 }
func InitStorageSDK(walletJSON string, blockWorker, chainID, signatureScheme string, preferredBlobbers []string) error {

	err := client.PopulateClient(walletJSON, signatureScheme)
	if err != nil {
		return err
	}

	blockchain.SetChainID(chainID)
	blockchain.SetPreferredBlobbers(preferredBlobbers)
	blockchain.SetBlockWorker(blockWorker)

	err = UpdateNetworkDetails()
	if err != nil {
		return err
	}

	go UpdateNetworkDetailsWorker(context.Background())
	sdkInitialized = true
	return nil
}

func GetNetwork() *Network {
	return &Network{
		Miners:   blockchain.GetMiners(),
		Sharders: blockchain.GetSharders(),
	}
}

func SetMaxTxnQuery(num int) {
	blockchain.SetMaxTxnQuery(num)

	cfg, _ := conf.GetClientConfig()
	if cfg != nil {
		cfg.MaxTxnQuery = num
	}

}

func SetQuerySleepTime(time int) {
	blockchain.SetQuerySleepTime(time)

	cfg, _ := conf.GetClientConfig()
	if cfg != nil {
		cfg.QuerySleepTime = time
	}

}

func SetMinSubmit(num int) {
	blockchain.SetMinSubmit(num)
}
func SetMinConfirmation(num int) {
	blockchain.SetMinConfirmation(num)
}

func SetNetwork(miners []string, sharders []string) {
	blockchain.SetMiners(miners)
	blockchain.SetSharders(sharders)
}

//
// read pool
//

func CreateReadPool() (err error) {
	if !sdkInitialized {
		return sdkNotInitialized
	}
	_, _, err = smartContractTxn(transaction.SmartContractTxnData{
		Name: transaction.STORAGESC_CREATE_READ_POOL,
	})
	return
}

type BlobberPoolStat struct {
	BlobberID common.Key     `json:"blobber_id"`
	Balance   common.Balance `json:"balance"`
}

type AllocationPoolStat struct {
	ID           string             `json:"id"`
	Balance      common.Balance     `json:"balance"`
	ExpireAt     common.Timestamp   `json:"expire_at"`
	AllocationID common.Key         `json:"allocation_id"`
	Blobbers     []*BlobberPoolStat `json:"blobbers"`
	Locked       bool               `json:"locked"`
}

type BackPool struct {
	ID      string         `json:"id"`
	Balance common.Balance `json:"balance"`
}

// AllocationPoolsStat represents read or write pool statistic.
type AllocationPoolStats struct {
	Pools []*AllocationPoolStat `json:"pools"`
	Back  *BackPool             `json:"back,omitempty"`
}

func (aps *AllocationPoolStats) AllocFilter(allocID string) {
	if allocID == "" {
		return
	}
	var i int
	for _, pi := range aps.Pools {
		if pi.AllocationID != common.Key(allocID) {
			continue
		}
		aps.Pools[i], i = pi, i+1
	}
	aps.Pools = aps.Pools[:i]
}

// GetReadPoolInfo for given client, or, if the given clientID is empty,
// for current client of the sdk.
func GetReadPoolInfo(clientID string) (info *AllocationPoolStats, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}

	if clientID == "" {
		clientID = client.GetClientID()
	}

	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getReadPoolStat",
		map[string]string{"client_id": clientID}, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting read pool info")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response")
	}

	info = new(AllocationPoolStats)
	if err = json.Unmarshal(b, info); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return
}

// ReadPoolLock locks given number of tokes for given duration in read pool.
func ReadPoolLock(dur time.Duration, allocID, blobberID string,
	tokens, fee int64) (err error) {
	if !sdkInitialized {
		return sdkNotInitialized
	}

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var req lockRequest
	req.Duration = dur
	req.AllocationID = allocID
	req.BlobberID = blobberID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_READ_POOL_LOCK,
		InputArgs: &req,
	}
	_, _, err = smartContractTxnValueFee(sn, tokens, fee)
	return
}

// ReadPoolUnlock unlocks tokens in expired read pool
func ReadPoolUnlock(poolID string, fee int64) (err error) {
	if !sdkInitialized {
		return sdkNotInitialized
	}

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}

	var req unlockRequest
	req.PoolID = poolID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_READ_POOL_UNLOCK,
		InputArgs: &req,
	}
	_, _, err = smartContractTxnValueFee(sn, 0, fee)
	return
}

//
// stake pool
//

// StakePoolOfferInfo represents stake pool offer information.
type StakePoolOfferInfo struct {
	Lock         common.Balance   `json:"lock"`
	Expire       common.Timestamp `json:"expire"`
	AllocationID common.Key       `json:"allocation_id"`
	IsExpired    bool             `json:"is_expired"`
}

// StakePoolRewardsInfo represents stake pool rewards.
type StakePoolRewardsInfo struct {
	Charge    common.Balance `json:"charge"`    // total for all time
	Blobber   common.Balance `json:"blobber"`   // total for all time
	Validator common.Balance `json:"validator"` // total for all time
}

// StakePoolDelegatePoolInfo represents delegate pool of a stake pool info.
type StakePoolDelegatePoolInfo struct {
	ID               common.Key     `json:"id"`                // pool ID
	Balance          common.Balance `json:"balance"`           // current balance
	DelegateID       common.Key     `json:"delegate_id"`       // wallet
	Rewards          common.Balance `json:"rewards"`           // total for all time
	Interests        common.Balance `json:"interests"`         // total for all time
	Penalty          common.Balance `json:"penalty"`           // total for all time
	PendingInterests common.Balance `json:"pending_interests"` // total for all time
	// Unstake > 0, then the pool wants to unstake. And the Unstake is maximal
	// time it can't be unstaked.
	Unstake common.Timestamp `json:"unstake"`
}

// StakePoolSettings information.
type StakePoolSettings struct {
	// DelegateWallet for pool owner.
	DelegateWallet string `json:"delegate_wallet"`
	// MinStake allowed.
	MinStake common.Balance `json:"min_stake"`
	// MaxStake allowed.
	MaxStake common.Balance `json:"max_stake"`
	// NumDelegates maximum allowed.
	NumDelegates int `json:"num_delegates"`
	// ServiceCharge is blobber service charge.
	ServiceCharge float64 `json:"service_charge"`
}

// StakePool full info.
type StakePoolInfo struct {
	ID      common.Key     `json:"pool_id"` // blobber ID
	Balance common.Balance `json:"balance"` // total stake
	Unstake common.Balance `json:"unstake"` // total unstake amount

	Free       common.Size    `json:"free"`        // free staked space
	Capacity   common.Size    `json:"capacity"`    // blobber bid
	WritePrice common.Balance `json:"write_price"` // its write price

	Offers      []*StakePoolOfferInfo `json:"offers"`       //
	OffersTotal common.Balance        `json:"offers_total"` //
	// delegate pools
	Delegate []*StakePoolDelegatePoolInfo `json:"delegate"`
	Earnings common.Balance               `json:"interests"` // total for all
	Penalty  common.Balance               `json:"penalty"`   // total for all
	// rewards
	Rewards StakePoolRewardsInfo `json:"rewards"`
	// settings
	Settings StakePoolSettings `json:"settings"`
}

// GetStakePoolInfo for given client, or, if the given clientID is empty,
// for current client of the sdk.
func GetStakePoolInfo(blobberID string) (info *StakePoolInfo, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	if blobberID == "" {
		blobberID = client.GetClientID()
	}

	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getStakePoolStat",
		map[string]string{"blobber_id": blobberID}, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting stake pool info:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response")
	}

	info = new(StakePoolInfo)
	if err = json.Unmarshal(b, info); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return
}

// StakePoolUserInfo represents user stake pools statistic.
type StakePoolUserInfo struct {
	Pools map[common.Key][]*StakePoolDelegatePoolInfo `json:"pools"`
}

// GetStakePoolUserInfo obtains blobbers/validators delegate pools statistic
// for a user. If given clientID is empty string, then current client used.
func GetStakePoolUserInfo(clientID string) (info *StakePoolUserInfo, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	if clientID == "" {
		clientID = client.GetClientID()
	}

	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS,
		"/getUserStakePoolStat", map[string]string{"client_id": clientID}, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting stake pool user info:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response")
	}

	info = new(StakePoolUserInfo)
	if err = json.Unmarshal(b, info); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return
}

type stakePoolRequest struct {
	BlobberID string `json:"blobber_id,omitempty"`
	PoolID    string `json:"pool_id,omitempty"`
}

// StakePoolLock locks tokens lack in stake pool
func StakePoolLock(blobberID string, value, fee int64) (poolID string, err error) {
	if !sdkInitialized {
		return poolID, sdkNotInitialized
	}
	if blobberID == "" {
		blobberID = client.GetClientID()
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_STAKE_POOL_LOCK,
		InputArgs: &spr,
	}
	poolID, _, err = smartContractTxnValueFee(sn, value, fee)
	return
}

// StakePoolUnlockUnstake is stake pool unlock response in case where tokens
// can't be unlocked due to opened offers. In this case it returns the maximal
// time to wait to be able to unlock the tokens. The real time can be lesser if
// someone cancels an allocation, or someone else stake more tokens, etc.
type StakePoolUnlockUnstake struct {
	Unstake common.Timestamp `json:"unstake"`
}

// StakePoolUnlock unlocks a stake pool tokens. If tokens can't be unlocked due
// to opened offers, then it returns time where the tokens can be unlocked,
// marking the pool as 'want to unlock' to avoid its usage in offers in the
// future. The time is maximal time that can be lesser in some cases. To
// unlock tokens can't be unlocked now, wait the time and unlock them (call
// this function again).
func StakePoolUnlock(blobberID, poolID string, fee int64) (
	unstake common.Timestamp, err error) {

	if !sdkInitialized {
		return 0, sdkNotInitialized
	}
	if blobberID == "" {
		blobberID = client.GetClientID()
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID
	spr.PoolID = poolID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_STAKE_POOL_UNLOCK,
		InputArgs: &spr,
	}

	var out string
	if _, out, err = smartContractTxnValueFee(sn, 0, fee); err != nil {
		return // an error
	}

	var spuu StakePoolUnlockUnstake
	if err = json.Unmarshal([]byte(out), &spuu); err != nil {
		return
	}

	return spuu.Unstake, nil
}

// StakePoolPayInterests unlocks a stake pool rewards.
func StakePoolPayInterests(bloberID string) (err error) {
	if !sdkInitialized {
		return sdkNotInitialized
	}
	if bloberID == "" {
		bloberID = client.GetClientID()
	}

	var spr stakePoolRequest
	spr.BlobberID = bloberID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_STAKE_POOL_PAY_INTERESTS,
		InputArgs: &spr,
	}
	_, _, err = smartContractTxnValueFee(sn, 0, 0)
	return
}

//
// write pool
//

// GetWritePoolInfo for given client, or, if the given clientID is empty,
// for current client of the sdk.
func GetWritePoolInfo(clientID string) (info *AllocationPoolStats, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	if clientID == "" {
		clientID = client.GetClientID()
	}

	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getWritePoolStat",
		map[string]string{"client_id": clientID}, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting read pool info:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response")
	}

	info = new(AllocationPoolStats)
	if err = json.Unmarshal(b, info); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return
}

// WritePoolLock locks given number of tokes for given duration in read pool.
func WritePoolLock(dur time.Duration, allocID, blobberID string,
	tokens, fee int64) (err error) {
	if !sdkInitialized {
		return sdkNotInitialized
	}

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var req lockRequest
	req.Duration = dur
	req.AllocationID = allocID
	req.BlobberID = blobberID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_WRITE_POOL_LOCK,
		InputArgs: &req,
	}
	_, _, err = smartContractTxnValueFee(sn, tokens, fee)
	return
}

// WritePoolUnlock unlocks tokens in expired read pool
func WritePoolUnlock(poolID string, fee int64) (err error) {
	if !sdkInitialized {
		return sdkNotInitialized
	}

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}

	var req unlockRequest
	req.PoolID = poolID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_WRITE_POOL_UNLOCK,
		InputArgs: &req,
	}
	_, _, err = smartContractTxnValueFee(sn, 0, fee)
	return
}

//
// challenge pool
//

// ChallengePoolInfo represents a challenge pool stat.
type ChallengePoolInfo struct {
	ID         string           `json:"id"`
	Balance    common.Balance   `json:"balance"`
	StartTime  common.Timestamp `json:"start_time"`
	Expiration common.Timestamp `json:"expiration"`
	Finalized  bool             `json:"finalized"`
}

// GetChallengePoolInfo for given allocation.
func GetChallengePoolInfo(allocID string) (info *ChallengePoolInfo, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}

	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS,
		"/getChallengePoolStat", map[string]string{"allocation_id": allocID},
		nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting challenge pool info:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response")
	}

	info = new(ChallengePoolInfo)
	if err = json.Unmarshal(b, info); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return
}

func GetMptData(key string) ([]byte, error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}

	var b []byte
	b, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS,
		"/get_mpt_key", map[string]string{"key": key},
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting mpt key data:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response")
	}

	return b, nil
}

//
// storage SC configurations and blobbers
//

type InputMap struct {
	Fields map[string]interface{} `json:"fields"`
}

func GetStorageSCConfig() (conf *InputMap, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}

	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getConfig", nil,
		nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting storage SC configs:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response")
	}

	conf = new(InputMap)
	conf.Fields = make(map[string]interface{})
	if err = json.Unmarshal(b, conf); err != nil {
		return nil, errors.Wrap(err, "rror decoding response:")
	}

	return
}

type Blobber struct {
	ID                common.Key        `json:"id"`
	BaseURL           string            `json:"url"`
	Terms             Terms             `json:"terms"`
	Capacity          common.Size       `json:"capacity"`
	Used              common.Size       `json:"used"`
	LastHealthCheck   common.Timestamp  `json:"last_health_check"`
	PublicKey         string            `json:"-"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

func GetBlobbers() (bs []*Blobber, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}

	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getblobbers", nil,
		nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting blobbers:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response")
	}

	type nodes struct {
		Nodes []*Blobber
	}

	var wrap nodes

	if err = json.Unmarshal(b, &wrap); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return wrap.Nodes, nil
}

// GetBlobber instance.
func GetBlobber(blobberID string) (blob *Blobber, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(
		STORAGE_SCADDRESS,
		"/getBlobber",
		map[string]string{"blobber_id": blobberID},
		nil)
	if err != nil {
		return nil, errors.Wrap(err, "requesting blobber:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response from sharders")
	}
	blob = new(Blobber)
	if err = json.Unmarshal(b, blob); err != nil {
		return nil, errors.Wrap(err, "decoding response:")
	}
	return
}

//
// ---
//

func GetClientEncryptedPublicKey() (string, error) {
	if !sdkInitialized {
		return "", sdkNotInitialized
	}
	encScheme := encryption.NewEncryptionScheme()
	_, err := encScheme.Initialize(client.GetClient().Mnemonic)
	if err != nil {
		return "", err
	}
	return encScheme.GetPublicKey()
}

func GetAllocationFromAuthTicket(authTicket string) (*Allocation, error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	sEnc, err := base64.StdEncoding.DecodeString(authTicket)
	if err != nil {
		return nil, errors.New("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	at := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return nil, errors.New("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	return GetAllocation(at.AllocationID)
}

func GetAllocation(allocationID string) (*Allocation, error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	params := make(map[string]string)
	params["allocation"] = allocationID
	allocationBytes, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/allocation", params, nil)
	if err != nil {
		return nil, errors.New("allocation_fetch_error", "Error fetching the allocation."+err.Error())
	}
	allocationObj := &Allocation{}
	err = json.Unmarshal(allocationBytes, allocationObj)
	if err != nil {
		return nil, errors.New("allocation_decode_error", "Error decoding the allocation."+err.Error())
	}
	allocationObj.numBlockDownloads = numBlockDownloads
	allocationObj.InitAllocation()
	return allocationObj, nil
}

func GetAllocationUpdates(allocation *Allocation) error {
	if allocation == nil {
		return errors.New("allocation_not_initialized", "")
	}

	params := make(map[string]string)
	params["allocation"] = allocation.ID
	allocationBytes, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/allocation", params, nil)
	if err != nil {
		return errors.New("allocation_fetch_error", "Error fetching the allocation."+err.Error())
	}

	updatedAllocationObj := new(Allocation)
	if err := json.Unmarshal(allocationBytes, updatedAllocationObj); err != nil {
		return errors.New("allocation_decode_error", "Error decoding the allocation."+err.Error())
	}

	allocation.DataShards = updatedAllocationObj.DataShards
	allocation.ParityShards = updatedAllocationObj.ParityShards
	allocation.Size = updatedAllocationObj.Size
	allocation.Expiration = updatedAllocationObj.Expiration
	allocation.Payer = updatedAllocationObj.Payer
	allocation.Blobbers = updatedAllocationObj.Blobbers
	allocation.Stats = updatedAllocationObj.Stats
	allocation.TimeUnit = updatedAllocationObj.TimeUnit
	allocation.IsImmutable = updatedAllocationObj.IsImmutable
	allocation.BlobberDetails = updatedAllocationObj.BlobberDetails
	allocation.ReadPriceRange = updatedAllocationObj.ReadPriceRange
	allocation.WritePriceRange = updatedAllocationObj.WritePriceRange
	allocation.ChallengeCompletionTime = updatedAllocationObj.ChallengeCompletionTime
	allocation.StartTime = updatedAllocationObj.StartTime
	allocation.Finalized = updatedAllocationObj.Finalized
	allocation.Canceled = updatedAllocationObj.Canceled
	allocation.MovedToChallenge = updatedAllocationObj.MovedToChallenge
	allocation.MovedBack = updatedAllocationObj.MovedBack
	allocation.MovedToValidators = updatedAllocationObj.MovedToValidators
	allocation.Curators = updatedAllocationObj.Curators
	return nil
}

func SetNumBlockDownloads(num int) {
	if num > 0 && num <= 100 {
		numBlockDownloads = num
	}
}

func GetAllocations() ([]*Allocation, error) {
	return GetAllocationsForClient(client.GetClientID())
}

func GetAllocationsForClient(clientID string) ([]*Allocation, error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	params := make(map[string]string)
	params["client"] = clientID
	allocationsBytes, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/allocations", params, nil)
	if err != nil {
		return nil, errors.New("allocations_fetch_error", "Error fetching the allocations."+err.Error())
	}
	allocations := make([]*Allocation, 0)
	err = json.Unmarshal(allocationsBytes, &allocations)
	if err != nil {
		return nil, errors.New("allocations_decode_error", "Error decoding the allocations."+err.Error())
	}
	return allocations, nil
}

func CreateAllocationWithBlobbers(datashards, parityshards int, size, expiry int64,
	readPrice, writePrice PriceRange, mcct time.Duration, lock int64, blobbers []string) (
	string, error) {

	return CreateAllocationForOwner(client.GetClientID(),
		client.GetClientPublicKey(), datashards, parityshards,
		size, expiry, readPrice, writePrice, mcct, lock,
		blobbers)
}

func CreateAllocation(datashards, parityshards int, size, expiry int64,
	readPrice, writePrice PriceRange, mcct time.Duration, lock int64) (
	string, error) {

	return CreateAllocationForOwner(client.GetClientID(),
		client.GetClientPublicKey(), datashards, parityshards,
		size, expiry, readPrice, writePrice, mcct, lock,
		blockchain.GetPreferredBlobbers())
}

func CreateAllocationForOwner(owner, ownerpublickey string,
	datashards, parityshards int, size, expiry int64,
	readPrice, writePrice PriceRange, mcct time.Duration,
	lock int64, preferredBlobbers []string) (hash string, err error) {

	if !sdkInitialized {
		return "", sdkNotInitialized
	}

	var allocationRequest = map[string]interface{}{
		"data_shards":                   datashards,
		"parity_shards":                 parityshards,
		"size":                          size,
		"owner_id":                      owner,
		"owner_public_key":              ownerpublickey,
		"expiration_date":               expiry,
		"preferred_blobbers":            preferredBlobbers,
		"read_price_range":              readPrice,
		"write_price_range":             writePrice,
		"max_challenge_completion_time": mcct,
		"diversify_blobbers":            false,
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.NEW_ALLOCATION_REQUEST,
		InputArgs: allocationRequest,
	}
	hash, _, err = smartContractTxnValue(sn, lock)
	return
}

func AddFreeStorageAssigner(name, publicKey string, individualLimit, totalLimit float64) error {
	if !sdkInitialized {
		return sdkNotInitialized
	}

	var input = map[string]interface{}{
		"name":             name,
		"public_key":       publicKey,
		"individual_limit": individualLimit,
		"total_limit":      totalLimit,
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.ADD_FREE_ALLOCATION_ASSIGNER,
		InputArgs: input,
	}
	_, _, err := smartContractTxn(sn)

	return err
}

func CreateFreeAllocation(marker string, value int64) (string, error) {
	if !sdkInitialized {
		return "", sdkNotInitialized
	}

	var input = map[string]interface{}{
		"recipient_public_key": client.GetClientPublicKey(),
		"marker":               marker,
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.NEW_FREE_ALLOCATION,
		InputArgs: input,
	}
	hash, _, err := smartContractTxnValue(sn, value)
	return hash, err
}

func UpdateAllocation(size int64, expiry int64, allocationID string,
	lock int64, setImmutable, updateTerms bool) (hash string, err error) {

	if !sdkInitialized {
		return "", sdkNotInitialized
	}

	updateAllocationRequest := make(map[string]interface{})
	updateAllocationRequest["owner_id"] = client.GetClientID()
	updateAllocationRequest["id"] = allocationID
	updateAllocationRequest["size"] = size
	updateAllocationRequest["expiration_date"] = expiry
	updateAllocationRequest["set_immutable"] = setImmutable
	updateAllocationRequest["update_terms"] = updateTerms

	sn := transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_ALLOCATION,
		InputArgs: updateAllocationRequest,
	}
	hash, _, err = smartContractTxnValue(sn, lock)
	return
}

func CreateFreeUpdateAllocation(marker, allocationId string, value int64) (string, error) {
	if !sdkInitialized {
		return "", sdkNotInitialized
	}

	var input = map[string]interface{}{
		"allocation_id": allocationId,
		"marker":        marker,
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.FREE_UPDATE_ALLOCATION,
		InputArgs: input,
	}
	hash, _, err := smartContractTxnValue(sn, value)
	return hash, err
}

func FinalizeAllocation(allocID string) (hash string, err error) {
	if !sdkInitialized {
		return "", sdkNotInitialized
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_FINALIZE_ALLOCATION,
		InputArgs: map[string]interface{}{"allocation_id": allocID},
	}
	hash, _, err = smartContractTxn(sn)
	return
}

func CancelAllocation(allocID string) (hash string, err error) {
	if !sdkInitialized {
		return "", sdkNotInitialized
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_CANCEL_ALLOCATION,
		InputArgs: map[string]interface{}{"allocation_id": allocID},
	}
	hash, _, err = smartContractTxn(sn)
	return
}

func RemoveCurator(curatorId, allocationId string) (string, error) {
	if !sdkInitialized {
		return "", sdkNotInitialized
	}

	var allocationRequest = map[string]interface{}{
		"curator_id":    curatorId,
		"allocation_id": allocationId,
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_REMOVE_CURATOR,
		InputArgs: allocationRequest,
	}
	hash, _, err := smartContractTxn(sn)
	return hash, err
}

func AddCurator(curatorId, allocationId string) (string, error) {
	if !sdkInitialized {
		return "", sdkNotInitialized
	}

	var allocationRequest = map[string]interface{}{
		"curator_id":    curatorId,
		"allocation_id": allocationId,
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_ADD_CURATOR,
		InputArgs: allocationRequest,
	}
	hash, _, err := smartContractTxn(sn)
	return hash, err
}

func CuratorTransferAllocation(allocationId, newOwner, newOwnerPublicKey string) (string, error) {
	if !sdkInitialized {
		return "", sdkNotInitialized
	}

	var allocationRequest = map[string]interface{}{
		"allocation_id":        allocationId,
		"new_owner_id":         newOwner,
		"new_owner_public_key": newOwnerPublicKey,
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_CURATOR_TRANSFER,
		InputArgs: allocationRequest,
	}
	hash, _, err := smartContractTxn(sn)
	return hash, err
}

func UpdateBlobberSettings(blob *Blobber) (resp string, err error) {
	if !sdkInitialized {
		return "", sdkNotInitialized
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS,
		InputArgs: blob,
	}
	resp, _, err = smartContractTxn(sn)
	return
}

func smartContractTxn(sn transaction.SmartContractTxnData) (
	hash, out string, err error) {

	return smartContractTxnValue(sn, 0)
}

func smartContractTxnValue(sn transaction.SmartContractTxnData, value int64) (
	hash, out string, err error) {

	return smartContractTxnValueFee(sn, value, 0)
}

func smartContractTxnValueFee(sn transaction.SmartContractTxnData,
	value, fee int64) (hash, out string, err error) {

	var requestBytes []byte
	if requestBytes, err = json.Marshal(sn); err != nil {
		return
	}

	var txn = transaction.NewTransactionEntity(client.GetClientID(),
		blockchain.GetChainID(), client.GetClientPublicKey())

	txn.TransactionData = string(requestBytes)
	txn.ToClientID = STORAGE_SCADDRESS
	txn.Value = value
	txn.TransactionFee = fee
	txn.TransactionType = transaction.TxnTypeSmartContract

	if err = txn.ComputeHashAndSign(client.Sign); err != nil {
		return
	}

	transaction.SendTransactionSync(txn, blockchain.GetMiners())

	var (
		querySleepTime = time.Duration(blockchain.GetQuerySleepTime()) * time.Second
		retries        = 0
		t              *transaction.Transaction
	)

	time.Sleep(querySleepTime)

	for retries < blockchain.GetMaxTxnQuery() {
		t, err = transaction.VerifyTransaction(txn.Hash, blockchain.GetSharders())
		if err == nil {
			break
		}
		retries++
		time.Sleep(querySleepTime)
	}

	if err != nil {
		Logger.Error("Error verifying the transaction", err.Error(), txn.Hash)
		return
	}

	if t == nil {
		return "", "", errors.New("transaction_validation_failed",
			"Failed to get the transaction confirmation")
	}

	if t.Status == transaction.TxnFail {
		return t.Hash, t.TransactionOutput, errors.New("", t.TransactionOutput)
	}

	return t.Hash, t.TransactionOutput, nil
}

func CommitToFabric(metaTxnData, fabricConfigJSON string) (string, error) {
	if !sdkInitialized {
		return "", sdkNotInitialized
	}
	var fabricConfig struct {
		URL  string `json:"url"`
		Body struct {
			Channel          string   `json:"channel"`
			ChaincodeName    string   `json:"chaincode_name"`
			ChaincodeVersion string   `json:"chaincode_version"`
			Method           string   `json:"method"`
			Args             []string `json:"args"`
		} `json:"body"`
		Auth struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"auth"`
	}

	err := json.Unmarshal([]byte(fabricConfigJSON), &fabricConfig)
	if err != nil {
		return "", errors.New("fabric_config_decode_error", "Unable to decode fabric config json")
	}

	// Clear if any existing args passed
	fabricConfig.Body.Args = fabricConfig.Body.Args[:0]

	fabricConfig.Body.Args = append(fabricConfig.Body.Args, metaTxnData)

	fabricData, err := json.Marshal(fabricConfig.Body)
	if err != nil {
		return "", errors.New("fabric_config_encode_error", "Unable to encode fabric config body")
	}

	req, ctx, cncl, err := zboxutil.NewHTTPRequest(http.MethodPost, fabricConfig.URL, fabricData)
	if err != nil {
		return "", errors.New("fabric_commit_error", "Unable to create new http request with error "+err.Error())
	}

	// Set basic auth
	req.SetBasicAuth(fabricConfig.Auth.Username, fabricConfig.Auth.Password)

	var fabricResponse string
	err = zboxutil.HttpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Fabric commit error : ", err)
			return err
		}
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "Error reading response :")
		}
		Logger.Debug("Fabric commit result:", string(respBody))
		if resp.StatusCode == http.StatusOK {
			fabricResponse = string(respBody)
			return nil
		}
		return errors.New(strconv.Itoa(resp.StatusCode), "Fabric commit status not OK!")
	})
	return fabricResponse, err
}

func GetAllocationMinLock(datashards, parityshards int, size, expiry int64,
	readPrice, writePrice PriceRange, mcct time.Duration) (int64, error) {
	if !sdkInitialized {
		return 0, sdkNotInitialized
	}

	var allocationRequestData = map[string]interface{}{
		"data_shards":                   datashards,
		"parity_shards":                 parityshards,
		"size":                          size,
		"owner_id":                      client.GetClientID(),
		"owner_public_key":              client.GetClientPublicKey(),
		"expiration_date":               expiry,
		"preferred_blobbers":            blockchain.GetPreferredBlobbers(),
		"read_price_range":              readPrice,
		"write_price_range":             writePrice,
		"max_challenge_completion_time": mcct,
		"diversify_blobbers":            false,
	}
	allocationData, _ := json.Marshal(allocationRequestData)

	params := make(map[string]string)
	params["allocation_data"] = string(allocationData)
	allocationsBytes, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/allocation_min_lock", params, nil)
	if err != nil {
		return 0, errors.New("allocation_min_lock_fetch_error", "Error fetching the allocation min lock."+err.Error())
	}

	var response = make(map[string]int64)
	err = json.Unmarshal(allocationsBytes, &response)
	if err != nil {
		return 0, errors.New("allocation_min_lock_decode_error", "Error decoding the response."+err.Error())
	}
	return response["min_lock_demand"], nil
}
