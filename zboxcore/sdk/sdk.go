package sdk

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/0chain/gosdk/zboxcore/marker"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const STORAGE_SCADDRESS = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"

const (
	OpUpload   int = 0
	OpDownload int = 1
	OpRepair   int = 2
	OpUpdate   int = 3
)

type StatusCallback interface {
	Started(allocationId, filePath string, op int, totalBytes int)
	InProgress(allocationId, filePath string, op int, completedBytes int)
	Error(allocationID string, filePath string, op int, err error)
	Completed(allocationId, filePath string, filename string, mimetype string, size int, op int)
}

var numBlockDownloads = 10
var sdkInitialized = false

// GetVersion - returns version string
func GetVersion() string {
	return version.VERSIONSTR
}

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

func InitStorageSDK(clientJson string, miners []string, sharders []string, chainID string, signatureScheme string, preferredBlobbers []string) error {
	err := client.PopulateClient(clientJson, signatureScheme)
	if err != nil {
		return err
	}
	blockchain.SetMiners(miners)
	blockchain.SetSharders(sharders)
	blockchain.SetPreferredBlobbers(preferredBlobbers)
	blockchain.SetChainID(chainID)
	sdkInitialized = true
	return nil
}

func SetMaxTxnQuery(num int) {
	blockchain.SetMaxTxnQuery(num)
}

func SetQuerySleepTime(time int) {
	blockchain.SetQuerySleepTime(time)
}

//
// read pool
//

func CreateReadPool() (err error) {
	_, err = smartContractTxn(transaction.SmartContractTxnData{
		Name: transaction.NEW_READ_POOL,
	})
	return
}

// ReadPoolStat is number of tokens and locking status for the tokens.
type ReadPoolStat struct {
	ID        string        `json:"pool_id"`
	StartTime int64         `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	TimeLeft  time.Duration `json:"time_left"`
	Locked    bool          `json:"locked"`
	Balance   int64         `json:"balance"`
}

// ReadPoolInfo is set of read pool locks statistic.
type ReadPoolInfo struct {
	Stats []*ReadPoolStat `json:"stats"`
}

// GetReadPoolInfo for given client, or, if the given clientID is empty,
// for current client of the sdk.
func GetReadPoolInfo(clientID string) (info *ReadPoolInfo, err error) {

	if clientID == "" {
		clientID = client.GetClientID()
	}

	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getReadPoolsStats",
		map[string]string{"client_id": clientID}, nil)
	if err != nil {
		return nil, fmt.Errorf("error requesting read pool info: %v", err)
	}

	info = new(ReadPoolInfo)
	if err = json.Unmarshal(b, info); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return
}

// ReadPoolLock locks given number of tokes for given duration in read pool.
func ReadPoolLock(dur time.Duration, tokens, fee int64) (err error) {

	type lockRequest struct {
		Duration time.Duration `json:"duration"`
	}

	var req lockRequest
	req.Duration = dur

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.READ_POOL_LOCK,
		InputArgs: &req,
	}
	_, err = smartContractTxnValueFee(sn, tokens, fee)
	return
}

// ReadPoolUnlock unlocks tokens in expired read pool
func ReadPoolUnlock(poolID string, fee int64) (err error) {

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}

	var req unlockRequest
	req.PoolID = poolID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.READ_POOL_UNLOCK,
		InputArgs: &req,
	}
	_, err = smartContractTxnValueFee(sn, 0, fee)
	return
}

//
// stake pool
//

type StakePoolOfferStat struct {
	Lock         int64  `json:"lock"`   // balance
	Expire       int64  `json:"expire"` // time, seconds
	AllocationID string `json:"allocation_id"`
	IsExpired    bool   `json:"is_expired"`
}

type StakePoolInfo struct {
	ID            string                `json:"pool_id"`  // id
	Locked        int64                 `json:"locked"`   // balance
	Unlocked      int64                 `json:"unlocked"` // balance
	Offers        []*StakePoolOfferStat `json:"offers"`
	OffersTotal   int64                 `json:"offers_total"`   // balance
	RequiredStake int64                 `json:"required_stake"` // balance
}

// GetStakePoolInfo for given client, or, if the given clientID is empty,
// for current client of the sdk.
func GetStakePoolInfo(blobberID string) (info *StakePoolInfo, err error) {

	if blobberID == "" {
		blobberID = client.GetClientID()
	}

	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getStakePoolStat",
		map[string]string{"blobber_id": blobberID}, nil)
	if err != nil {
		return nil, fmt.Errorf("error requesting stake pool info: %v", err)
	}

	info = new(StakePoolInfo)
	if err = json.Unmarshal(b, info); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return
}

// StakePoolUnlock unlocks tokens in stake pool
func StakePoolUnlock(fee int64) (err error) {
	var sn = transaction.SmartContractTxnData{
		Name: transaction.STAKE_POOL_UNLOCK,
	}
	_, err = smartContractTxnValueFee(sn, 0, fee)
	return
}

//
// write pool
//

type WritePoolInfo struct {
	ID        string        `json:"pool_id"`    //
	StartTime int64         `json:"start_time"` // time, seconds
	Duration  time.Duration `json:"duration"`   //
	TimeLeft  time.Duration `json:"time_left"`  //
	Locked    bool          `json:"locked"`     //
	Balance   int64         `json:"balance"`    // balance
}

// GetWritePoolInfo for given client, or, if the given clientID is empty,
// for current client of the sdk.
func GetWritePoolInfo(allocID string) (info *WritePoolInfo, err error) {
	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getWritePoolStat",
		map[string]string{"allocation_id": allocID}, nil)
	if err != nil {
		return nil, fmt.Errorf("error requesting write pool info: %v", err)
	}

	info = new(WritePoolInfo)
	if err = json.Unmarshal(b, info); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return
}

// WritePoolLock unlocks tokens in expired write pool
func WritePoolLock(allocID string, tokens, fee int64) (err error) {

	type unlockRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	var req unlockRequest
	req.AllocationID = allocID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.WRITE_POOL_LOCK,
		InputArgs: &req,
	}
	_, err = smartContractTxnValueFee(sn, tokens, fee)
	return
}

//
// challenge pool
//

type ChallengePoolInfo struct {
	ID        string        `json:"pool_id"`    //
	StartTime int64         `json:"start_time"` // time, seconds
	Duration  time.Duration `json:"duration"`   //
	TimeLeft  time.Duration `json:"time_left"`  //
	Locked    bool          `json:"locked"`     //
	Balance   int64         `json:"balance"`    // balance
}

// GetChallengePoolInfo for given allocation.
func GetChallengePoolInfo(allocID string) (info *ChallengePoolInfo, err error) {
	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS,
		"/getChallengePoolStat", map[string]string{"allocation_id": allocID},
		nil)
	if err != nil {
		return nil, fmt.Errorf("error requesting challenge pool info: %v", err)
	}

	info = new(ChallengePoolInfo)
	if err = json.Unmarshal(b, info); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return
}

//
// storage SC configurations and blobbers
//

type StorageSCReadPoolConfig struct {
	MinLock       int64         `json:"min_lock"`        // balance, tokens
	MinLockPeriod time.Duration `json:"min_lock_period"` //
	MaxLockPeriod time.Duration `json:"max_lock_period"` //
}

type StorageSCWritePoolConfig struct {
	MinLock int64 `json:"min_lock"` // balance, tokens
}

type StorageSCConfig struct {
	ChallengeEnabled           bool                      `json:"challenge_enabled"`
	ChallengeRatePerMBMin      time.Duration             `json:"challenge_rate_per_mb_min"`
	MinAllocSize               int64                     `json:"min_alloc_size"` // size, bytes
	MinAllocDuration           time.Duration             `json:"min_alloc_duration"`
	MaxChallengeCompletionTime time.Duration             `json:"max_challenge_completion_time"`
	MinOfferDuration           time.Duration             `json:"min_offer_duration"`
	MinBlobberCapacity         int64                     `json:"min_blobber_capacity"`
	ReadPool                   *StorageSCReadPoolConfig  `json:"readpool"`
	WritePool                  *StorageSCWritePoolConfig `json:"writepool"`
	ValidatorReward            float64                   `json:"validator_reward"`
	BlobberSlash               float64                   `json:"blobber_slash"`
}

func GetStorageSCConfig() (conf *StorageSCConfig, err error) {
	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getConfig", nil,
		nil)
	if err != nil {
		return nil, fmt.Errorf("error requesting storage SC configs: %v", err)
	}

	conf = new(StorageSCConfig)
	if err = json.Unmarshal(b, conf); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if conf.ReadPool == nil || conf.WritePool == nil {
		return nil, errors.New("invalid confg: missing read/write pool configs")
	}
	return
}

// Terms represents Blobber terms. A Blobber can update its terms,
// but any existing offer will use terms of offer signing time.
type Terms struct {
	ReadPrice               int64         `json:"read_price"`                // tokens / read
	WritePrice              int64         `json:"write_price"`               // tokens / GB
	MinLockDemand           float64       `json:"min_lock_demand"`           //
	MaxOfferDuration        time.Duration `json:"max_offer_duration"`        //
	ChallengeCompletionTime time.Duration `json:"challenge_completion_time"` //
}

type Blobber struct {
	ID              string `json:"id"`
	BaseURL         string `json:"url"`
	Terms           Terms  `json:"terms"`             // terms
	Capacity        int64  `json:"capacity"`          // total blobber capacity
	Used            int64  `json:"used"`              // allocated capacity
	LastHealthCheck int64  `json:"last_health_check"` // time, seconds
	PublicKey       string `json:"-"`
}

func GetBlobbers() (bs []*Blobber, err error) {
	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getblobbers", nil,
		nil)
	if err != nil {
		return nil, fmt.Errorf("error requesting blobbers: %v", err)
	}

	type nodes struct {
		Nodes []*Blobber
	}

	var wrap nodes

	if err = json.Unmarshal(b, &wrap); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return wrap.Nodes, nil
}

//
// ---
//

func GetClientEncryptedPublicKey() (string, error) {
	if !sdkInitialized {
		return "", common.NewError("sdk_not_initialized", "SDK is not initialised")
	}
	encScheme := encryption.NewEncryptionScheme()
	err := encScheme.Initialize(client.GetClient().Mnemonic)
	if err != nil {
		return "", err
	}
	return encScheme.GetPublicKey()
}

func GetAllocationFromAuthTicket(authTicket string) (*Allocation, error) {
	sEnc, err := base64.StdEncoding.DecodeString(authTicket)
	if err != nil {
		return nil, common.NewError("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	at := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return nil, common.NewError("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	return GetAllocation(at.AllocationID)
}

func GetAllocation(allocationID string) (*Allocation, error) {
	params := make(map[string]string)
	params["allocation"] = allocationID
	allocationBytes, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/allocation", params, nil)
	if err != nil {
		return nil, common.NewError("allocation_fetch_error", "Error fetching the allocation."+err.Error())
	}
	allocationObj := &Allocation{}
	err = json.Unmarshal(allocationBytes, allocationObj)
	if err != nil {
		return nil, common.NewError("allocation_decode_error", "Error decoding the allocation."+err.Error())
	}
	allocationObj.numBlockDownloads = numBlockDownloads
	allocationObj.InitAllocation()
	return allocationObj, nil
}

func SetNumBlockDownloads(num int) {
	if num > 0 && num <= 100 {
		numBlockDownloads = num
	}
	return
}

func GetAllocations() ([]*Allocation, error) {
	return GetAllocationsForClient(client.GetClientID())
}

func GetAllocationsForClient(clientID string) ([]*Allocation, error) {
	params := make(map[string]string)
	params["client"] = clientID
	allocationsBytes, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/allocations", params, nil)
	if err != nil {
		return nil, common.NewError("allocations_fetch_error", "Error fetching the allocations."+err.Error())
	}
	allocations := make([]*Allocation, 0)
	err = json.Unmarshal(allocationsBytes, &allocations)
	if err != nil {
		return nil, common.NewError("allocations_decode_error", "Error decoding the allocations."+err.Error())
	}
	return allocations, nil
}

func CreateAllocation(datashards, parityshards int, size, expiry int64,
	readPrice, writePrice PriceRange, lock int64) (
	string, error) {

	return CreateAllocationForOwner(client.GetClientID(),
		client.GetClientPublicKey(), datashards, parityshards,
		size, expiry, readPrice, writePrice, lock,
		blockchain.GetPreferredBlobbers())
}

func CreateAllocationForOwner(owner, ownerpublickey string,
	datashards, parityshards int,
	size, expiry int64, readPrice, writePrice PriceRange, lock int64,
	preferredBlobbers []string) (string, error) {

	var allocationRequest = map[string]interface{}{
		"data_shards":        datashards,
		"parity_shards":      parityshards,
		"size":               size,
		"owner_id":           owner,
		"owner_public_key":   ownerpublickey,
		"expiration_date":    expiry,
		"preferred_blobbers": preferredBlobbers,
		"read_price_range":   readPrice,
		"write_price_range":  writePrice,
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.NEW_ALLOCATION_REQUEST,
		InputArgs: allocationRequest,
	}
	return smartContractTxnValue(sn, lock)
}

func UpdateAllocation(size int64, expiry int64, allocationID string) (string, error) {
	updateAllocationRequest := make(map[string]interface{})
	updateAllocationRequest["owner_id"] = client.GetClientID()
	updateAllocationRequest["id"] = allocationID
	updateAllocationRequest["size"] = size
	updateAllocationRequest["expiration_date"] = expiry

	sn := transaction.SmartContractTxnData{
		Name:      transaction.UPDATE_ALLOCATION_REQUEST,
		InputArgs: updateAllocationRequest,
	}
	return smartContractTxn(sn)
}

func FinalizeAllocation(allocationID string) (string, error) {
	var req = make(map[string]interface{})
	req["allocation_id"] = allocationID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.FINALIZE_ALLOCATION,
		InputArgs: req,
	}
	return smartContractTxn(sn)
}

func smartContractTxn(sn transaction.SmartContractTxnData) (string, error) {
	return smartContractTxnValue(sn, 0)
}

func smartContractTxnValue(sn transaction.SmartContractTxnData, value int64) (string, error) {
	return smartContractTxnValueFee(sn, value, 0)
}

func smartContractTxnValueFee(sn transaction.SmartContractTxnData, value, fee int64) (string, error) {
	requestBytes, err := json.Marshal(sn)
	if err != nil {
		return "", err
	}
	txn := transaction.NewTransactionEntity(client.GetClientID(), blockchain.GetChainID(), client.GetClientPublicKey())
	txn.TransactionData = string(requestBytes)
	txn.ToClientID = STORAGE_SCADDRESS
	txn.Value = value
	txn.TransactionFee = fee
	txn.TransactionType = transaction.TxnTypeSmartContract
	err = txn.ComputeHashAndSign(client.Sign)
	if err != nil {
		return "", err
	}
	transaction.SendTransactionSync(txn, blockchain.GetMiners())
	querySleepTime := time.Duration(blockchain.GetQuerySleepTime()) * time.Second
	time.Sleep(querySleepTime)
	retries := 0
	var t *transaction.Transaction
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
		return "", err
	}
	if t == nil {
		return "", common.NewError("transaction_validation_failed", "Failed to get the transaction confirmation")
	}

	return t.Hash, nil
}

func CommitToFabric(metaTxnData, fabricConfigJSON string) (string, error) {
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
		return "", common.NewError("fabric_config_decode_error", "Unable to decode fabric config json")
	}

	// Clear if any existing args passed
	fabricConfig.Body.Args = fabricConfig.Body.Args[:0]

	fabricConfig.Body.Args = append(fabricConfig.Body.Args, metaTxnData)

	fabricData, err := json.Marshal(fabricConfig.Body)
	if err != nil {
		return "", common.NewError("fabric_config_encode_error", "Unable to encode fabric config body")
	}

	req, ctx, cncl, err := zboxutil.NewHTTPRequest(http.MethodPost, fabricConfig.URL, fabricData)
	if err != nil {
		return "", common.NewError("fabric_commit_error", "Unable to create new http request with error "+err.Error())
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
			return fmt.Errorf("Error reading response : %s", err.Error())
		}
		Logger.Debug("Fabric commit result:", string(respBody))
		if resp.StatusCode == http.StatusOK {
			fabricResponse = string(respBody)
			return nil
		}
		return fmt.Errorf("Fabric commit status not OK, Status : %v", resp.StatusCode)
	})
	return fabricResponse, err
}
