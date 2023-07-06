package sdk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/0chain/common/core/currency"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/sys"
	"go.uber.org/zap"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const STORAGE_SCADDRESS = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"

var sdkNotInitialized = errors.New("sdk_not_initialized", "SDK is not initialised")
var allocationNotFound = errors.New("couldnt_find_allocation", "Couldn't find the allocation required for update")

const (
	OpUpload            int = 0
	OpDownload          int = 1
	OpRepair            int = 2
	OpUpdate            int = 3
	opThumbnailDownload int = 4
)

type StatusCallback interface {
	Started(allocationId, filePath string, op int, totalBytes int)
	InProgress(allocationId, filePath string, op int, completedBytes int, data []byte)
	Error(allocationID string, filePath string, op int, err error)
	Completed(allocationId, filePath string, filename string, mimetype string, size int, op int)
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
	l.Logger.SetLevel(lvl)
}

// SetLogFile
// logFile - Log file
// verbose - true - console output; false - no console output
func SetLogFile(logFile string, verbose bool) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	l.Logger.SetLogFile(f, verbose)
	l.Logger.Info("******* Storage SDK Version: ", version.VERSIONSTR, " *******")
}

func GetLogger() *logger.Logger {
	return &l.Logger
}

func InitStorageSDK(walletJSON string,
	blockWorker, chainID, signatureScheme string,
	preferredBlobbers []string,
	nonce int64,
	fee ...uint64) error {
	err := client.PopulateClient(walletJSON, signatureScheme)
	if err != nil {
		return err
	}

	client.SetClientNonce(nonce)
	if len(fee) > 0 {
		client.SetTxnFee(fee[0])
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
	transaction.InitCache(sharders)
}

//
// read pool
//

func CreateReadPool() (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}
	hash, _, nonce, _, err = smartContractTxn(transaction.SmartContractTxnData{
		Name: transaction.STORAGESC_CREATE_READ_POOL,
	})
	return
}

type BackPool struct {
	ID      string         `json:"id"`
	Balance common.Balance `json:"balance"`
}

//
// read pool
//

type ReadPool struct {
	Balance common.Balance `json:"balance"`
}

// GetReadPoolInfo for given client, or, if the given clientID is empty,
// for current client of the sdk.
func GetReadPoolInfo(clientID string) (info *ReadPool, err error) {
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

	info = new(ReadPool)
	if err = json.Unmarshal(b, info); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return
}

// ReadPoolLock locks given number of tokes for given duration in read pool.
func ReadPoolLock(tokens, fee uint64) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_READ_POOL_LOCK,
		InputArgs: nil,
	}
	hash, _, nonce, _, err = smartContractTxnValueFee(sn, tokens, fee)
	return
}

// ReadPoolUnlock unlocks tokens in expired read pool
func ReadPoolUnlock(fee uint64) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_READ_POOL_UNLOCK,
		InputArgs: nil,
	}
	hash, _, nonce, _, err = smartContractTxnValueFee(sn, 0, fee)
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
	ID         common.Key     `json:"id"`          // blobber ID
	Balance    common.Balance `json:"balance"`     // current balance
	DelegateID common.Key     `json:"delegate_id"` // wallet
	Rewards    common.Balance `json:"rewards"`     // current
	UnStake    bool           `json:"unstake"`     // want to unstake

	TotalReward  common.Balance   `json:"total_reward"`
	TotalPenalty common.Balance   `json:"total_penalty"`
	Status       string           `json:"status"`
	RoundCreated int64            `json:"round_created"`
	StakedAt     common.Timestamp `json:"staked_at"`
}

// StakePool full info.
type StakePoolInfo struct {
	ID         common.Key     `json:"pool_id"` // pool ID
	Balance    common.Balance `json:"balance"` // total balance
	StakeTotal common.Balance `json:"stake_total"`
	// delegate pools
	Delegate []StakePoolDelegatePoolInfo `json:"delegate"`
	// rewards
	Rewards common.Balance `json:"rewards"`

	// Settings of the stake pool
	Settings blockchain.StakePoolSettings `json:"settings"`
}

// GetStakePoolInfo for given client, or, if the given clientID is empty,
// for current client of the sdk.
func GetStakePoolInfo(providerType ProviderType, providerID string) (info *StakePoolInfo, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}

	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getStakePoolStat",
		map[string]string{"provider_type": strconv.Itoa(int(providerType)), "provider_id": providerID}, nil)
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
func GetStakePoolUserInfo(clientID string, offset, limit int) (info *StakePoolUserInfo, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	if clientID == "" {
		clientID = client.GetClientID()
	}

	var b []byte
	params := map[string]string{
		"client_id": clientID,
		"offset":    strconv.FormatInt(int64(offset), 10),
		"limit":     strconv.FormatInt(int64(limit), 10),
	}
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS,
		"/getUserStakePoolStat", params, nil)
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
	ProviderType ProviderType `json:"provider_type,omitempty"`
	ProviderID   string       `json:"provider_id,omitempty"`
}

// StakePoolLock locks tokens lack in stake pool
func StakePoolLock(providerType ProviderType, providerID string, value, fee uint64) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	if providerType == 0 {
		return "", 0, errors.New("stake_pool_lock", "provider is required")
	}

	if providerID == "" {
		return "", 0, errors.New("stake_pool_lock", "provider_id is required")
	}

	spr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerID,
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_STAKE_POOL_LOCK,
		InputArgs: &spr,
	}
	hash, _, nonce, _, err = smartContractTxnValueFee(sn, value, fee)
	return
}

// stakePoolLock is stake pool unlock response in case where tokens
// can't be unlocked due to opened offers.
type stakePoolLock struct {
	Client       string       `json:"client"`
	ProviderId   string       `json:"provider_id"`
	ProviderType ProviderType `json:"provider_type"`
	Amount       int64        `json:"amount"`
}

// StakePoolUnlock unlocks a stake pool tokens. If tokens can't be unlocked due
// to opened offers, then it returns time where the tokens can be unlocked,
// marking the pool as 'want to unlock' to avoid its usage in offers in the
// future. The time is maximal time that can be lesser in some cases. To
// unlock tokens can't be unlocked now, wait the time and unlock them (call
// this function again).
func StakePoolUnlock(providerType ProviderType, providerID string, fee uint64) (unstake int64, nonce int64, err error) {
	if !sdkInitialized {
		return 0, 0, sdkNotInitialized
	}

	if providerType == 0 {
		return 0, 0, errors.New("stake_pool_lock", "provider is required")
	}

	if providerID == "" {
		return 0, 0, errors.New("stake_pool_lock", "provider_id is required")
	}

	spr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerID,
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_STAKE_POOL_UNLOCK,
		InputArgs: &spr,
	}

	var out string
	if _, out, nonce, _, err = smartContractTxnValueFee(sn, 0, fee); err != nil {
		return // an error
	}

	var spuu stakePoolLock
	if err = json.Unmarshal([]byte(out), &spuu); err != nil {
		return
	}

	return spuu.Amount, nonce, nil
}

//
// write pool
//

// WritePoolLock locks given number of tokes for given duration in read pool.
func WritePoolLock(allocID string, tokens, fee uint64) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	type lockRequest struct {
		AllocationID string `json:"allocation_id"`
	}

	var req lockRequest
	req.AllocationID = allocID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_WRITE_POOL_LOCK,
		InputArgs: &req,
	}
	hash, _, nonce, _, err = smartContractTxnValueFee(sn, tokens, fee)
	return
}

// WritePoolUnlock unlocks tokens in expired read pool
func WritePoolUnlock(allocID string, fee uint64) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	type unlockRequest struct {
		AllocationID string `json:"allocation_id"`
	}

	var req unlockRequest
	req.AllocationID = allocID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_WRITE_POOL_UNLOCK,
		InputArgs: &req,
	}
	hash, _, nonce, _, err = smartContractTxnValueFee(sn, 0, fee)
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
	b, err = zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/storage-config", nil,
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
	ID                       common.Key                   `json:"id"`
	BaseURL                  string                       `json:"url"`
	Terms                    Terms                        `json:"terms"`
	Capacity                 common.Size                  `json:"capacity"`
	Allocated                common.Size                  `json:"allocated"`
	LastHealthCheck          common.Timestamp             `json:"last_health_check"`
	PublicKey                string                       `json:"-"`
	StakePoolSettings        blockchain.StakePoolSettings `json:"stake_pool_settings"`
	TotalStake               int64                        `json:"total_stake"`
	UsedAllocation           int64                        `json:"used_allocation"`
	TotalOffers              int64                        `json:"total_offers"`
	TotalServiceCharge       int64                        `json:"total_service_charge"`
	UncollectedServiceCharge int64                        `json:"uncollected_service_charge"`
	IsKilled                 bool                         `json:"is_killed"`
	IsShutdown               bool                         `json:"is_shutdown"`
	NotAvailable             bool                         `json:"not_available"`
}

type UpdateBlobber struct {
	ID                       common.Key                    `json:"id"`
	BaseURL                  *string                       `json:"url,omitempty"`
	Terms                    *Terms                        `json:"terms,omitempty"`
	Capacity                 *common.Size                  `json:"capacity,omitempty"`
	Allocated                *common.Size                  `json:"allocated,omitempty"`
	LastHealthCheck          *common.Timestamp             `json:"last_health_check,omitempty"`
	StakePoolSettings        *blockchain.StakePoolSettings `json:"stake_pool_settings,omitempty"`
	TotalStake               *int64                        `json:"total_stake,omitempty"`
	UsedAllocation           *int64                        `json:"used_allocation,omitempty"`
	TotalOffers              *int64                        `json:"total_offers,omitempty"`
	TotalServiceCharge       *int64                        `json:"total_service_charge,omitempty"`
	UncollectedServiceCharge *int64                        `json:"uncollected_service_charge,omitempty"`
	IsKilled                 *bool                         `json:"is_killed,omitempty"`
	IsShutdown               *bool                         `json:"is_shutdown,omitempty"`
	NotAvailable             *bool                         `json:"not_available,omitempty"`
}

type Validator struct {
	ID                       common.Key       `json:"validator_id"`
	BaseURL                  string           `json:"url"`
	PublicKey                string           `json:"-"`
	DelegateWallet           string           `json:"delegate_wallet"`
	MinStake                 common.Balance   `json:"min_stake"`
	MaxStake                 common.Balance   `json:"max_stake"`
	NumDelegates             int              `json:"num_delegates"`
	ServiceCharge            float64          `json:"service_charge"`
	StakeTotal               int64            `json:"stake_total"`
	TotalServiceCharge       int64            `json:"total_service_charge"`
	UncollectedServiceCharge int64            `json:"uncollected_service_charge"`
	LastHealthCheck          common.Timestamp `json:"last_health_check"`
	IsKilled                 bool             `json:"is_killed"`
	IsShutdown               bool             `json:"is_shutdown"`
}

type UpdateValidator struct {
	ID                       common.Key        `json:"validator_id"`
	BaseURL                  *string           `json:"url,omitempty"`
	DelegateWallet           *string           `json:"delegate_wallet,omitempty"`
	MinStake                 *common.Balance   `json:"min_stake,omitempty"`
	MaxStake                 *common.Balance   `json:"max_stake,omitempty"`
	NumDelegates             *int              `json:"num_delegates,omitempty"`
	ServiceCharge            *float64          `json:"service_charge,omitempty"`
	StakeTotal               *int64            `json:"stake_total,omitempty"`
	TotalServiceCharge       *int64            `json:"total_service_charge,omitempty"`
	UncollectedServiceCharge *int64            `json:"uncollected_service_charge,omitempty"`
	LastHealthCheck          *common.Timestamp `json:"last_health_check,omitempty"`
	IsKilled                 *bool             `json:"is_killed,omitempty"`
	IsShutdown               *bool             `json:"is_shutdown,omitempty"`
}

func (v *Validator) ConvertToValidationNode() *blockchain.ValidationNode {
	return &blockchain.ValidationNode{
		ID:      string(v.ID),
		BaseURL: v.BaseURL,
		StakePoolSettings: blockchain.StakePoolSettings{
			DelegateWallet: v.DelegateWallet,
			MinStake:       v.MinStake,
			MaxStake:       v.MaxStake,
			NumDelegates:   v.NumDelegates,
			ServiceCharge:  v.ServiceCharge,
		},
	}
}

func (v *UpdateValidator) ConvertToValidationNode() *blockchain.ValidationNode {
	validationNode := &blockchain.ValidationNode{ID: string(v.ID)}

	if v.BaseURL != nil {
		validationNode.BaseURL = *v.BaseURL
	}

	if v.DelegateWallet != nil {
		validationNode.StakePoolSettings.DelegateWallet = *v.DelegateWallet
	}

	if v.MinStake != nil {
		validationNode.StakePoolSettings.MinStake = *v.MinStake
	}

	if v.MaxStake != nil {
		validationNode.StakePoolSettings.MaxStake = *v.MaxStake
	}

	if v.NumDelegates != nil {
		validationNode.StakePoolSettings.NumDelegates = *v.NumDelegates
	}

	if v.ServiceCharge != nil {
		validationNode.StakePoolSettings.ServiceCharge = *v.ServiceCharge
	}

	return validationNode
}

func getBlobbersInternal(active bool, limit, offset int) (bs []*Blobber, err error) {
	type nodes struct {
		Nodes []*Blobber
	}

	url := fmt.Sprintf("/getblobbers?active=%s&limit=%d&offset=%d",
		strconv.FormatBool(active),
		limit,
		offset,
	)
	b, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, url, nil, nil)
	var wrap nodes
	if err != nil {
		return nil, errors.Wrap(err, "error requesting blobbers:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response")
	}

	if err = json.Unmarshal(b, &wrap); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return wrap.Nodes, nil
}

func GetBlobbers(active bool) (bs []*Blobber, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}

	limit, offset := 20, 0

	blobbers, err := getBlobbersInternal(active, limit, offset)
	if err != nil {
		return nil, err
	}

	var blobbersSl []*Blobber
	blobbersSl = append(blobbersSl, blobbers...)
	for {
		// if the len of output returned is less than the limit it means this is the last round of pagination
		if len(blobbers) < limit {
			break
		}

		// get the next set of blobbers
		offset += 20
		blobbers, err = getBlobbersInternal(active, limit, offset)
		if err != nil {
			return blobbers, err
		}
		blobbersSl = append(blobbersSl, blobbers...)

	}
	return blobbersSl, nil
}

// GetBlobber instance.
//
//	# Inputs
//	-	blobberID: the id of blobber
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

// GetValidator instance.
func GetValidator(validatorID string) (validator *Validator, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(
		STORAGE_SCADDRESS,
		"/get_validator",
		map[string]string{"validator_id": validatorID},
		nil)
	if err != nil {
		return nil, errors.Wrap(err, "requesting validator:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response from sharders")
	}
	validator = new(Validator)
	if err = json.Unmarshal(b, validator); err != nil {
		return nil, errors.Wrap(err, "decoding response:")
	}
	return
}

// List all validators
func GetValidators() (validators []*Validator, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(
		STORAGE_SCADDRESS,
		"/validators",
		nil,
		nil)
	if err != nil {
		return nil, errors.Wrap(err, "requesting validator list")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response from sharders")
	}
	if err = json.Unmarshal(b, &validators); err != nil {
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
	allocation.FileOptions = updatedAllocationObj.FileOptions
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

func getAllocationsInternal(clientID string, limit, offset int) ([]*Allocation, error) {
	params := make(map[string]string)
	params["client"] = clientID
	params["limit"] = fmt.Sprint(limit)
	params["offset"] = fmt.Sprint(offset)
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

// get paginated results
func GetAllocationsForClient(clientID string) ([]*Allocation, error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	limit, offset := 20, 0

	allocations, err := getAllocationsInternal(clientID, limit, offset)
	if err != nil {
		return nil, err
	}

	var allocationsFin []*Allocation
	allocationsFin = append(allocationsFin, allocations...)
	for {
		// if the len of output returned is less than the limit it means this is the last round of pagination
		if len(allocations) < limit {
			break
		}

		// get the next set of blobbers
		offset += 20
		allocations, err = getAllocationsInternal(clientID, limit, offset)
		if err != nil {
			return allocations, err
		}
		allocationsFin = append(allocationsFin, allocations...)

	}
	return allocationsFin, nil
}

type FileOptionParam struct {
	Changed bool
	Value   bool
}

type FileOptionsParameters struct {
	ForbidUpload FileOptionParam
	ForbidDelete FileOptionParam
	ForbidUpdate FileOptionParam
	ForbidMove   FileOptionParam
	ForbidCopy   FileOptionParam
	ForbidRename FileOptionParam
}

type CreateAllocationOptions struct {
	DataShards           int
	ParityShards         int
	Size                 int64
	Expiry               int64
	ReadPrice            PriceRange
	WritePrice           PriceRange
	Lock                 uint64
	BlobberIds           []string
	ThirdPartyExtendable bool
	FileOptionsParams    *FileOptionsParameters
}

func CreateAllocationWith(options CreateAllocationOptions) (
	string, int64, *transaction.Transaction, error) {

	if len(options.BlobberIds) > 0 {
		return CreateAllocationForOwner(client.GetClientID(),
			client.GetClientPublicKey(), options.DataShards, options.ParityShards,
			options.Size, options.Expiry, options.ReadPrice, options.WritePrice, options.Lock,
			options.BlobberIds, options.ThirdPartyExtendable, options.FileOptionsParams)
	}

	return CreateAllocation(options.DataShards, options.ParityShards,
		options.Size, options.Expiry, options.ReadPrice, options.WritePrice, options.Lock,
		options.ThirdPartyExtendable, options.FileOptionsParams)

}

func CreateAllocation(datashards, parityshards int, size, expiry int64,
	readPrice, writePrice PriceRange, lock uint64, thirdPartyExtendable bool, fileOptionsParams *FileOptionsParameters) (
	string, int64, *transaction.Transaction, error) {

	if lock > math.MaxInt64 {
		return "", 0, nil, errors.New("invalid_lock", "int64 overflow on lock value")
	}

	preferredBlobberIds, err := GetBlobberIds(blockchain.GetPreferredBlobbers())
	if err != nil {
		return "", 0, nil, errors.New("failed_get_blobber_ids", "failed to get preferred blobber ids: "+err.Error())
	}
	return CreateAllocationForOwner(client.GetClientID(),
		client.GetClientPublicKey(), datashards, parityshards,
		size, expiry, readPrice, writePrice, lock,
		preferredBlobberIds, thirdPartyExtendable, fileOptionsParams)
}

func CreateAllocationForOwner(
	owner, ownerpublickey string,
	datashards, parityshards int, size, expiry int64,
	readPrice, writePrice PriceRange,
	lock uint64, preferredBlobberIds []string, thirdPartyExtendable bool, fileOptionsParams *FileOptionsParameters,
) (hash string, nonce int64, txn *transaction.Transaction, err error) {

	if lock > math.MaxInt64 {
		return "", 0, nil, errors.New("invalid_lock", "int64 overflow on lock value")
	}

	allocationRequest, err := getNewAllocationBlobbers(
		datashards, parityshards, size, expiry, readPrice, writePrice, preferredBlobberIds)
	if err != nil {
		return "", 0, nil, errors.New("failed_get_allocation_blobbers", "failed to get blobbers for allocation: "+err.Error())
	}

	if !sdkInitialized {
		return "", 0, nil, sdkNotInitialized
	}

	allocationRequest["owner_id"] = owner
	allocationRequest["owner_public_key"] = ownerpublickey
	allocationRequest["third_party_extendable"] = thirdPartyExtendable
	allocationRequest["file_options_changed"], allocationRequest["file_options"] = calculateAllocationFileOptions(63 /*0011 1111*/, fileOptionsParams)

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.NEW_ALLOCATION_REQUEST,
		InputArgs: allocationRequest,
	}
	hash, _, nonce, txn, err = smartContractTxnValue(sn, lock)
	return
}

func GetAllocationBlobbers(
	datashards, parityshards int,
	size, expiry int64,
	readPrice, writePrice PriceRange,
) ([]string, error) {
	var allocationRequest = map[string]interface{}{
		"data_shards":       datashards,
		"parity_shards":     parityshards,
		"size":              size,
		"expiration_date":   expiry,
		"read_price_range":  readPrice,
		"write_price_range": writePrice,
	}

	allocationData, _ := json.Marshal(allocationRequest)

	params := make(map[string]string)
	params["allocation_data"] = string(allocationData)

	allocBlobber, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/alloc_blobbers", params, nil)
	if err != nil {
		return nil, err
	}
	var allocBlobberIDs []string

	err = json.Unmarshal(allocBlobber, &allocBlobberIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal blobber IDs")
	}

	return allocBlobberIDs, nil
}

func getNewAllocationBlobbers(
	datashards, parityshards int,
	size, expiry int64,
	readPrice, writePrice PriceRange,
	preferredBlobberIds []string,
) (map[string]interface{}, error) {
	allocBlobberIDs, err := GetAllocationBlobbers(
		datashards, parityshards, size, expiry, readPrice, writePrice,
	)
	if err != nil {
		return nil, err
	}

	blobbers := append(preferredBlobberIds, allocBlobberIDs...)

	// filter duplicates
	ids := make(map[string]bool)
	uniqueBlobbers := []string{}
	for _, b := range blobbers {
		if !ids[b] {
			uniqueBlobbers = append(uniqueBlobbers, b)
			ids[b] = true
		}
	}

	return map[string]interface{}{
		"data_shards":       datashards,
		"parity_shards":     parityshards,
		"size":              size,
		"expiration_date":   expiry,
		"blobbers":          uniqueBlobbers,
		"read_price_range":  readPrice,
		"write_price_range": writePrice,
	}, nil
}

func GetBlobberIds(blobberUrls []string) ([]string, error) {

	if len(blobberUrls) == 0 {
		return nil, nil
	}

	urlsStr, err := json.Marshal(blobberUrls)
	if err != nil {
		return nil, err
	}

	params := make(map[string]string)
	params["blobber_urls"] = string(urlsStr)
	idsStr, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/blobber_ids", params, nil)
	if err != nil {
		return nil, err
	}

	var blobberIDs []string
	err = json.Unmarshal(idsStr, &blobberIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal preferred blobber IDs")
	}

	return blobberIDs, nil
}

func getFreeAllocationBlobbers(request map[string]interface{}) ([]string, error) {
	data, _ := json.Marshal(request)

	params := make(map[string]string)
	params["free_allocation_data"] = string(data)

	allocBlobber, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/free_alloc_blobbers", params, nil)
	if err != nil {
		return nil, err
	}
	var allocBlobberIDs []string

	err = json.Unmarshal(allocBlobber, &allocBlobberIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal blobber IDs")
	}

	return allocBlobberIDs, nil
}

func AddFreeStorageAssigner(name, publicKey string, individualLimit, totalLimit float64) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
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
	hash, _, n, _, err := smartContractTxn(sn)

	return hash, n, err
}

func CreateFreeAllocation(marker string, value uint64) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	recipientPublicKey := client.GetClientPublicKey()

	var input = map[string]interface{}{
		"recipient_public_key": recipientPublicKey,
		"marker":               marker,
	}

	blobbers, err := getFreeAllocationBlobbers(input)
	if err != nil {
		return "", 0, err
	}

	input["blobbers"] = blobbers

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.NEW_FREE_ALLOCATION,
		InputArgs: input,
	}
	hash, _, n, _, err := smartContractTxnValue(sn, value)
	return hash, n, err
}

func UpdateAllocation(
	size, expiry int64,
	allocationID string,
	lock uint64,
	updateTerms bool,
	addBlobberId, removeBlobberId string,
	setThirdPartyExtendable bool, fileOptionsParams *FileOptionsParameters,
) (hash string, nonce int64, err error) {

	if lock > math.MaxInt64 {
		return "", 0, errors.New("invalid_lock", "int64 overflow on lock value")
	}

	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	alloc, err := GetAllocation(allocationID)
	if err != nil {
		return "", 0, allocationNotFound
	}

	updateAllocationRequest := make(map[string]interface{})
	updateAllocationRequest["owner_id"] = client.GetClientID()
	updateAllocationRequest["owner_public_key"] = ""
	updateAllocationRequest["id"] = allocationID
	updateAllocationRequest["size"] = size
	updateAllocationRequest["expiration_date"] = expiry
	updateAllocationRequest["update_terms"] = updateTerms
	updateAllocationRequest["add_blobber_id"] = addBlobberId
	updateAllocationRequest["remove_blobber_id"] = removeBlobberId
	updateAllocationRequest["set_third_party_extendable"] = setThirdPartyExtendable
	updateAllocationRequest["file_options_changed"], updateAllocationRequest["file_options"] = calculateAllocationFileOptions(alloc.FileOptions, fileOptionsParams)

	sn := transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_ALLOCATION,
		InputArgs: updateAllocationRequest,
	}
	hash, _, nonce, _, err = smartContractTxnValue(sn, lock)
	return
}

func CreateFreeUpdateAllocation(marker, allocationId string, value uint64) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var input = map[string]interface{}{
		"allocation_id": allocationId,
		"marker":        marker,
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.FREE_UPDATE_ALLOCATION,
		InputArgs: input,
	}
	hash, _, n, _, err := smartContractTxnValue(sn, value)
	return hash, n, err
}

func FinalizeAllocation(allocID string) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_FINALIZE_ALLOCATION,
		InputArgs: map[string]interface{}{"allocation_id": allocID},
	}
	hash, _, nonce, _, err = smartContractTxn(sn)
	return
}

func CancelAllocation(allocID string) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_CANCEL_ALLOCATION,
		InputArgs: map[string]interface{}{"allocation_id": allocID},
	}
	hash, _, nonce, _, err = smartContractTxn(sn)
	return
}

type ProviderType int

const (
	ProviderMiner ProviderType = iota + 1
	ProviderSharder
	ProviderBlobber
	ProviderValidator
	ProviderAuthorizer
)

func KillProvider(providerId string, providerType ProviderType) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var input = map[string]interface{}{
		"provider_id": providerId,
	}
	var sn = transaction.SmartContractTxnData{
		InputArgs: input,
	}
	switch providerType {
	case ProviderBlobber:
		sn.Name = transaction.STORAGESC_KILL_BLOBBER
	case ProviderValidator:
		sn.Name = transaction.STORAGESC_KILL_VALIDATOR
	default:
		return "", 0, fmt.Errorf("kill provider type %v not implimented", providerType)
	}
	hash, _, n, _, err := smartContractTxn(sn)
	return hash, n, err
}

func ShutdownProvider(providerType ProviderType) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var input = map[string]interface{}{}
	var sn = transaction.SmartContractTxnData{
		InputArgs: input,
	}
	switch providerType {
	case ProviderBlobber:
		sn.Name = transaction.STORAGESC_SHUTDOWN_BLOBBER
	case ProviderValidator:
		sn.Name = transaction.STORAGESC_SHUTDOWN_VALIDATOR
	default:
		return "", 0, fmt.Errorf("shutdown provider type %v not implimented", providerType)
	}
	hash, _, n, _, err := smartContractTxn(sn)
	return hash, n, err
}

func CollectRewards(providerId string, providerType ProviderType) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var input = map[string]interface{}{
		"provider_id":   providerId,
		"provider_type": providerType,
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_COLLECT_REWARD,
		InputArgs: input,
	}
	hash, _, n, _, err := smartContractTxn(sn)
	return hash, n, err
}

func TransferAllocation(allocationId, newOwner, newOwnerPublicKey string) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	alloc, err := GetAllocation(allocationId)
	if err != nil {
		return "", 0, allocationNotFound
	}

	var allocationRequest = map[string]interface{}{
		"id":                         allocationId,
		"owner_id":                   newOwner,
		"owner_public_key":           newOwnerPublicKey,
		"size":                       0,
		"expiration_date":            0,
		"update_terms":               false,
		"add_blobber_id":             "",
		"remove_blobber_id":          "",
		"set_third_party_extendable": alloc.ThirdPartyExtendable,
		"file_options_changed":       false,
		"file_options":               alloc.FileOptions,
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_ALLOCATION,
		InputArgs: allocationRequest,
	}
	hash, _, n, _, err := smartContractTxn(sn)
	return hash, n, err
}

func UpdateBlobberSettings(blob *UpdateBlobber) (resp string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS,
		InputArgs: blob,
	}
	resp, _, nonce, _, err = smartContractTxn(sn)
	return
}

func UpdateValidatorSettings(v *UpdateValidator) (resp string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_VALIDATOR_SETTINGS,
		InputArgs: v.ConvertToValidationNode(),
	}
	resp, _, nonce, _, err = smartContractTxn(sn)
	return
}

func SmartContractTxn(sn transaction.SmartContractTxnData) (
	hash, out string, nonce int64, txn *transaction.Transaction, err error) {

	return smartContractTxnValue(sn, 0)
}

func smartContractTxn(sn transaction.SmartContractTxnData) (
	hash, out string, nonce int64, txn *transaction.Transaction, err error) {

	return smartContractTxnValue(sn, 0)
}

func smartContractTxnValue(sn transaction.SmartContractTxnData, value uint64) (
	hash, out string, nonce int64, txn *transaction.Transaction, err error) {

	// Fee is set during sdk initialization.
	return smartContractTxnValueFee(sn, value, client.TxnFee())
}

func smartContractTxnValueFee(sn transaction.SmartContractTxnData,
	value, fee uint64) (hash, out string, nonce int64, t *transaction.Transaction, err error) {

	var requestBytes []byte
	if requestBytes, err = json.Marshal(sn); err != nil {
		return
	}

	//nonce = client.GetClient().Nonce
	//if nonce != 0 {
	//	nonce++
	//}
	txn := transaction.NewTransactionEntity(client.GetClientID(),
		blockchain.GetChainID(), client.GetClientPublicKey(), nonce)

	txn.TransactionData = string(requestBytes)
	txn.ToClientID = STORAGE_SCADDRESS
	txn.Value = value
	txn.TransactionFee = fee
	txn.TransactionType = transaction.TxnTypeSmartContract

	// adjust fees if not set
	if fee == 0 {
		fee, err = transaction.EstimateFee(txn, blockchain.GetMiners(), 0.2)
		if err != nil {
			l.Logger.Error("failed to estimate txn fee",
				zap.Error(err),
				zap.Any("txn", txn))
			return
		}
		txn.TransactionFee = fee
	}

	if txn.TransactionNonce == 0 {
		txn.TransactionNonce = transaction.Cache.GetNextNonce(txn.ClientID)
	}

	if err = txn.ComputeHashAndSign(client.Sign); err != nil {
		return
	}

	msg := fmt.Sprintf("executing transaction '%s' with hash %s ", sn.Name, txn.Hash)
	l.Logger.Info(msg)
	l.Logger.Info("estimated txn fee: ", txn.TransactionFee)

	transaction.SendTransactionSync(txn, blockchain.GetMiners())

	var (
		querySleepTime = time.Duration(blockchain.GetQuerySleepTime()) * time.Second
		retries        = 0
	)

	sys.Sleep(querySleepTime)

	for retries < blockchain.GetMaxTxnQuery() {
		t, err = transaction.VerifyTransaction(txn.Hash, blockchain.GetSharders())
		if err == nil {
			break
		}
		retries++
		sys.Sleep(querySleepTime)
	}

	if err != nil {
		l.Logger.Error("Error verifying the transaction", err.Error(), txn.Hash)
		transaction.Cache.Evict(txn.ClientID)
		return
	}

	if t == nil {
		return "", "", 0, txn, errors.New("transaction_validation_failed",
			"Failed to get the transaction confirmation")
	}

	if t.Status == transaction.TxnFail {
		return t.Hash, t.TransactionOutput, 0, t, errors.New("", t.TransactionOutput)
	}

	if t.Status == transaction.TxnChargeableError {
		return t.Hash, t.TransactionOutput, t.TransactionNonce, t, errors.New("", t.TransactionOutput)
	}

	return t.Hash, t.TransactionOutput, t.TransactionNonce, t, nil
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
			l.Logger.Error("Fabric commit error : ", err)
			return err
		}
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "Error reading response :")
		}
		l.Logger.Debug("Fabric commit result:", string(respBody))
		if resp.StatusCode == http.StatusOK {
			fabricResponse = string(respBody)
			return nil
		}
		return errors.New(strconv.Itoa(resp.StatusCode), "Fabric commit status not OK!")
	})
	return fabricResponse, err
}

// expire in milliseconds
func GetAllocationMinLock(
	datashards, parityshards int,
	size, expiry int64,
	readPrice, writePrice PriceRange,
) (int64, error) {
	baSize := int64(math.Ceil(float64(size) / float64(datashards)))
	totalSize := baSize * int64(datashards+parityshards)
	config, err := GetStorageSCConfig()
	if err != nil {
		return 0, err
	}
	t := config.Fields["time_unit"]
	timeunitStr, ok := t.(string)
	if !ok {
		return 0, fmt.Errorf("bad time_unit type")
	}
	timeunit, err := time.ParseDuration(timeunitStr)
	if err != nil {
		return 0, fmt.Errorf("bad time_unit format")
	}

	duration := expiry / timeunit.Milliseconds()
	if expiry%timeunit.Milliseconds() != 0 {
		duration++
	}

	sizeInGB := float64(totalSize) / GB
	cost := float64(duration) * (sizeInGB*float64(writePrice.Max) + sizeInGB*float64(readPrice.Max))
	coin, err := currency.Float64ToCoin(cost)
	if err != nil {
		return 0, err
	}
	i, err := coin.Int64()
	if err != nil {
		return 0, err
	}
	return i, nil
}

// calculateAllocationFileOptions calculates the FileOptions 16-bit mask given the user input
func calculateAllocationFileOptions(initial uint16, fop *FileOptionsParameters) (bool, uint16) {
	if fop == nil {
		return false, initial
	}

	mask := initial

	if fop.ForbidUpload.Changed {
		mask = updateMaskBit(mask, 0, !fop.ForbidUpload.Value)
	}

	if fop.ForbidDelete.Changed {
		mask = updateMaskBit(mask, 1, !fop.ForbidDelete.Value)
	}

	if fop.ForbidUpdate.Changed {
		mask = updateMaskBit(mask, 2, !fop.ForbidUpdate.Value)
	}

	if fop.ForbidMove.Changed {
		mask = updateMaskBit(mask, 3, !fop.ForbidMove.Value)
	}

	if fop.ForbidCopy.Changed {
		mask = updateMaskBit(mask, 4, !fop.ForbidCopy.Value)
	}

	if fop.ForbidRename.Changed {
		mask = updateMaskBit(mask, 5, !fop.ForbidRename.Value)
	}

	return mask != initial, mask
}

// updateMaskBit Set/Clear (based on `value`) bit value of the bit of `mask` at `index` (starting with LSB as 0) and return the updated mask
func updateMaskBit(mask uint16, index uint8, value bool) uint16 {
	if value {
		return mask | uint16(1<<index)
	} else {
		return mask & ^uint16(1<<index)
	}
}
