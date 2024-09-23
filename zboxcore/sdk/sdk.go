package sdk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

	"github.com/0chain/common/core/currency"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/node"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/0chain/gosdk/core/common"
	enc "github.com/0chain/gosdk/core/encryption"
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
const MINERSC_SCADDRESS = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9"
const ZCNSC_SCADDRESS = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712e0"

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

var (
	numBlockDownloads         = 100
	sdkInitialized            = false
	networkWorkerTimerInHours = 1
	singleClientMode          = false
	shouldVerifyHash          = true
)

func SetSingleClietnMode(mode bool) {
	singleClientMode = mode
}

func SetShouldVerifyHash(verify bool) {
	shouldVerifyHash = verify
}

func SetSaveProgress(save bool) {
	shouldSaveProgress = save
}

// GetVersion - returns version string
func GetVersion() string {
	return version.VERSIONSTR
}

// SetLogLevel set the log level.
//   - lvl: 0 disabled; higher number (upto 4) more verbosity
func SetLogLevel(lvl int) {
	l.Logger.SetLevel(lvl)
}

// SetLogFile set the log file and verbosity levels
//   - logFile: Log file
//   - verbose: true - console output; false - no console output
func SetLogFile(logFile string, verbose bool) {
	var ioWriter = &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100, // MB
		MaxBackups: 5,   // number of backups
		MaxAge:     28,  //days
		LocalTime:  false,
		Compress:   false, // disabled by default
	}

	l.Logger.SetLogFile(ioWriter, verbose)
	l.Logger.Info("******* Storage SDK Version: ", version.VERSIONSTR, " *******")
}

// GetLogger retrieves logger instance
func GetLogger() *logger.Logger {
	return &l.Logger
}

// InitStorageSDK Initialize the storage SDK
//
//   - walletJSON: Client's wallet JSON
//   - blockWorker: Block worker URL (block worker refers to 0DNS)
//   - chainID: ID of the blokcchain network
//   - signatureScheme: Signature scheme that will be used for signing transactions
//   - preferredBlobbers: List of preferred blobbers to use when creating an allocation. This is usually configured by the client in the configuration files
//   - nonce: Initial nonce value for the transactions
//   - fee: Preferred value for the transaction fee, just the first value is taken
func InitStorageSDK(walletJSON string,
	blockWorker, chainID, signatureScheme string,
	preferredBlobbers []string,
	nonce int64,
	fee ...uint64) error {
	err := client.PopulateClient(walletJSON, signatureScheme)
	if err != nil {
		return err
	}

	blockchain.SetChainID(chainID)
	blockchain.SetBlockWorker(blockWorker)

	err = InitNetworkDetails()
	if err != nil {
		return err
	}

	client.SetClientNonce(nonce)
	if len(fee) > 0 {
		client.SetTxnFee(fee[0])
	}

	go UpdateNetworkDetailsWorker(context.Background())
	sdkInitialized = true
	return nil
}

// GetNetwork retrieves the network details
func GetNetwork() *Network {
	return &Network{
		Miners:   blockchain.GetMiners(),
		Sharders: blockchain.GetAllSharders(),
	}
}

// SetMaxTxnQuery set the maximum number of transactions to query
func SetMaxTxnQuery(num int) {
	blockchain.SetMaxTxnQuery(num)

	cfg, _ := conf.GetClientConfig()
	if cfg != nil {
		cfg.MaxTxnQuery = num
	}

}

// SetQuerySleepTime set the sleep time between queries
func SetQuerySleepTime(time int) {
	blockchain.SetQuerySleepTime(time)

	cfg, _ := conf.GetClientConfig()
	if cfg != nil {
		cfg.QuerySleepTime = time
	}

}

// SetMinSubmit set the minimum number of miners to submit the transaction
func SetMinSubmit(num int) {
	blockchain.SetMinSubmit(num)
}

// SetMinConfirmation set the minimum number of miners to confirm the transaction
func SetMinConfirmation(num int) {
	blockchain.SetMinConfirmation(num)
}

// SetNetwork set the network details, given the miners and sharders urls
//   - miners: list of miner urls
//   - sharders: list of sharder urls
func SetNetwork(miners []string, sharders []string) {
	blockchain.SetMiners(miners)
	blockchain.SetSharders(sharders)
	node.InitCache(blockchain.Sharders)
}

// CreateReadPool creates a read pool for the SDK client.
// Read pool is used to lock tokens for read operations.
// Currently, all read operations are free 🚀.
func CreateReadPool() (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}
	hash, _, nonce, _, err = storageSmartContractTxn(transaction.SmartContractTxnData{
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
//   - clientID: client ID
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

// StakePool information of stake pool of a provider.
type StakePoolInfo struct {
	ID         common.Key     `json:"pool_id"` // pool ID
	Balance    common.Balance `json:"balance"` // total balance
	StakeTotal common.Balance `json:"stake_total"`
	// delegate pools
	Delegate []StakePoolDelegatePoolInfo `json:"delegate"`
	// rewards
	Rewards common.Balance `json:"rewards"`
	// total rewards
	TotalRewards common.Balance `json:"total_rewards"`
	// Settings of the stake pool
	Settings blockchain.StakePoolSettings `json:"settings"`
}

// GetStakePoolInfo retrieve stake pool info for the current client configured to the sdk, given provider type and provider ID.
//   - providerType: provider type
//   - providerID: provider ID
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

// GetStakePoolUserInfo obtains blobbers/validators delegate pools statistic for a user.
// If given clientID is empty string, then current client used.
//   - clientID: client ID
//   - offset: offset
//   - limit: limit
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

// stakePoolLock is stake pool unlock response in case where tokens
// can't be unlocked due to opened offers.
type stakePoolLock struct {
	Client       string       `json:"client"`
	ProviderId   string       `json:"provider_id"`
	ProviderType ProviderType `json:"provider_type"`
	Amount       int64        `json:"amount"`
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

// GetChallengePoolInfo retrieve challenge pool info for given allocation.
//   - allocID: allocation ID
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

// GetMptData retrieves mpt key data.
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

// GetStorageSCConfig retrieves storage SC configurations.
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

// Blobber type represents blobber information.
type Blobber struct {
	// ID of the blobber
	ID common.Key `json:"id"`

	// BaseURL of the blobber
	BaseURL string `json:"url"`

	// Terms of the blobber
	Terms Terms `json:"terms"`

	// Capacity of the blobber
	Capacity common.Size `json:"capacity"`

	// Allocated size of the blobber
	Allocated common.Size `json:"allocated"`

	// LastHealthCheck of the blobber
	LastHealthCheck common.Timestamp `json:"last_health_check"`

	// PublicKey of the blobber
	PublicKey string `json:"-"`

	// StakePoolSettings settings of the blobber staking
	StakePoolSettings blockchain.StakePoolSettings `json:"stake_pool_settings"`

	// TotalStake of the blobber in SAS
	TotalStake int64 `json:"total_stake"`

	// UsedAllocation of the blobber in SAS
	UsedAllocation int64 `json:"used_allocation"`

	// TotalOffers of the blobber in SAS
	TotalOffers int64 `json:"total_offers"`

	// TotalServiceCharge of the blobber in SAS
	TotalServiceCharge int64 `json:"total_service_charge"`

	// UncollectedServiceCharge of the blobber in SAS
	UncollectedServiceCharge int64 `json:"uncollected_service_charge"`

	// IsKilled flag of the blobber, if true then the blobber is killed
	IsKilled bool `json:"is_killed"`

	// IsShutdown flag of the blobber, if true then the blobber is shutdown
	IsShutdown bool `json:"is_shutdown"`

	// NotAvailable flag of the blobber, if true then the blobber is not available
	NotAvailable bool `json:"not_available"`

	// IsRestricted flag of the blobber, if true then the blobber is restricted
	IsRestricted bool `json:"is_restricted"`
}

// UpdateBlobber is used during update blobber settings calls.
// Note the types are of pointer types with omitempty json property.
// This is done to correctly identify which properties are actually changing.
type UpdateBlobber struct {
	ID                       common.Key                          `json:"id"`
	BaseURL                  *string                             `json:"url,omitempty"`
	Terms                    *UpdateTerms                        `json:"terms,omitempty"`
	Capacity                 *common.Size                        `json:"capacity,omitempty"`
	Allocated                *common.Size                        `json:"allocated,omitempty"`
	LastHealthCheck          *common.Timestamp                   `json:"last_health_check,omitempty"`
	StakePoolSettings        *blockchain.UpdateStakePoolSettings `json:"stake_pool_settings,omitempty"`
	TotalStake               *int64                              `json:"total_stake,omitempty"`
	UsedAllocation           *int64                              `json:"used_allocation,omitempty"`
	TotalOffers              *int64                              `json:"total_offers,omitempty"`
	TotalServiceCharge       *int64                              `json:"total_service_charge,omitempty"`
	UncollectedServiceCharge *int64                              `json:"uncollected_service_charge,omitempty"`
	IsKilled                 *bool                               `json:"is_killed,omitempty"`
	IsShutdown               *bool                               `json:"is_shutdown,omitempty"`
	NotAvailable             *bool                               `json:"not_available,omitempty"`
	IsRestricted             *bool                               `json:"is_restricted,omitempty"`
}

// ResetBlobberStatsDto represents blobber stats reset request.
type ResetBlobberStatsDto struct {
	BlobberID     string `json:"blobber_id"`
	PrevAllocated int64  `json:"prev_allocated"`
	PrevSavedData int64  `json:"prev_saved_data"`
	NewAllocated  int64  `json:"new_allocated"`
	NewSavedData  int64  `json:"new_saved_data"`
}

// Validator represents validator information.
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

// UpdateValidator is used during update validator settings calls.
// Note the types are of pointer types with omitempty json property.
// This is done to correctly identify which properties are actually changing.
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

// ConvertToValidationNode converts UpdateValidator request to blockchain.UpdateValidationNode.
func (v *UpdateValidator) ConvertToValidationNode() *blockchain.UpdateValidationNode {
	blockValidator := &blockchain.UpdateValidationNode{
		ID:      string(v.ID),
		BaseURL: v.BaseURL,
	}

	sp := &blockchain.UpdateStakePoolSettings{
		DelegateWallet: v.DelegateWallet,
		NumDelegates:   v.NumDelegates,
		ServiceCharge:  v.ServiceCharge,
	}

	if v.DelegateWallet != nil ||
		v.MinStake != nil ||
		v.MaxStake != nil ||
		v.NumDelegates != nil ||
		v.ServiceCharge != nil {
		blockValidator.StakePoolSettings = sp
	}

	return blockValidator
}

func getBlobbersInternal(active, stakable bool, limit, offset int) (bs []*Blobber, err error) {
	type nodes struct {
		Nodes []*Blobber
	}

	url := fmt.Sprintf("/getblobbers?active=%s&limit=%d&offset=%d&stakable=%s",
		strconv.FormatBool(active),
		limit,
		offset,
		strconv.FormatBool(stakable),
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

// GetBlobbers returns list of blobbers.
//   - active: if true then only active blobbers are returned
//   - stakable: if true then only stakable blobbers are returned
func GetBlobbers(active, stakable bool) (bs []*Blobber, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}

	limit, offset := 20, 0

	blobbers, err := getBlobbersInternal(active, stakable, limit, offset)
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
		blobbers, err = getBlobbersInternal(active, stakable, limit, offset)
		if err != nil {
			return blobbers, err
		}
		blobbersSl = append(blobbersSl, blobbers...)

	}
	return blobbersSl, nil
}

// GetBlobber retrieve blobber by id.
//   - blobberID: the id of blobber
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

// GetValidator retrieve validator instance by id.
//   - validatorID: the id of validator
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

// GetValidators returns list of validators.
//   - stakable: if true then only stakable validators are returned
func GetValidators(stakable bool) (validators []*Validator, err error) {
	if !sdkInitialized {
		return nil, sdkNotInitialized
	}
	var b []byte
	b, err = zboxutil.MakeSCRestAPICall(
		STORAGE_SCADDRESS,
		"/validators",
		map[string]string{
			"stakable": strconv.FormatBool(stakable),
		},
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

// GetClientEncryptedPublicKey - get the client's public key
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

// GetAllocationFromAuthTicket - get allocation from given auth ticket hash.
// AuthTicket is used to access free allocations, and it's generated by the Free Storage Assigner.
//   - authTicket: the auth ticket hash
//
// returns the allocation instance and error if any
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

// GetAllocation - get allocation from given allocation id
//
//   - allocationID: the allocation id
//
// returns the allocation instance and error if any
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
		return nil, errors.New("allocation_decode_error", "Error decoding the allocation: "+err.Error()+" "+string(allocationBytes))
	}
	hashdata := allocationObj.Tx
	sig, ok := zboxutil.SignCache.Get(hashdata)
	if !ok {
		sig, err = client.Sign(enc.Hash(hashdata))
		zboxutil.SignCache.Add(hashdata, sig)
		if err != nil {
			return nil, err
		}
	}

	allocationObj.sig = sig
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
	allocation.IsEnterprise = updatedAllocationObj.IsEnterprise
	return nil
}

// SetNumBlockDownloads - set the number of block downloads, needs to be between 1 and 500 (inclusive). Default is 20.
//   - num: the number of block downloads
func SetNumBlockDownloads(num int) {
	if num > 0 && num <= 500 {
		numBlockDownloads = num
	}
}

// GetAllocations - get all allocations for the current client
//
// returns the list of allocations and error if any
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

// GetAllocationsForClient - get all allocations for given client id
//
//   - clientID: the client id
//
// returns the list of allocations and error if any
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

// FileOptionsParameters is used to specify the file options parameters for an allocation, which control the usage permissions of the files in the allocation.
type FileOptionsParameters struct {
	ForbidUpload FileOptionParam
	ForbidDelete FileOptionParam
	ForbidUpdate FileOptionParam
	ForbidMove   FileOptionParam
	ForbidCopy   FileOptionParam
	ForbidRename FileOptionParam
}

// CreateAllocationOptions is used to specify the options for creating a new allocation.
type CreateAllocationOptions struct {
	DataShards           int
	ParityShards         int
	Size                 int64
	ReadPrice            PriceRange
	WritePrice           PriceRange
	Lock                 uint64
	BlobberIds           []string
	BlobberAuthTickets   []string
	ThirdPartyExtendable bool
	IsEnterprise         bool
	FileOptionsParams    *FileOptionsParameters
	Force                bool
}

// CreateAllocationWith creates a new allocation with the given options for the current client using the SDK.
// Similar ro CreateAllocationForOwner but uses an options struct instead of individual parameters.
//   - options is the options struct instance for creating the allocation.
//
// returns the hash of the new_allocation_request transaction, the nonce of the transaction, the transaction object and an error if any.
func CreateAllocationWith(options CreateAllocationOptions) (
	string, int64, *transaction.Transaction, error) {

	return CreateAllocationForOwner(client.GetClientID(),
		client.GetClientPublicKey(), options.DataShards, options.ParityShards,
		options.Size, options.ReadPrice, options.WritePrice, options.Lock,
		options.BlobberIds, options.BlobberAuthTickets, options.ThirdPartyExtendable, options.IsEnterprise, options.Force, options.FileOptionsParams)
}

// GetAllocationBlobbers returns a list of blobber ids that can be used for a new allocation.
//
//   - datashards is the number of data shards for the allocation.
//   - parityshards is the number of parity shards for the allocation.
//   - size is the size of the allocation.
//   - readPrice is the read price range for the allocation (Reads in Züs are free!).
//   - writePrice is the write price range for the allocation.
//   - force is a flag indicating whether to force the allocation to be created.
//
// returns the list of blobber ids and an error if any.
func GetAllocationBlobbers(
	datashards, parityshards int,
	size int64,
	isRestricted int,
	readPrice, writePrice PriceRange,
	force ...bool,
) ([]string, error) {
	var allocationRequest = map[string]interface{}{
		"data_shards":       datashards,
		"parity_shards":     parityshards,
		"size":              size,
		"read_price_range":  readPrice,
		"write_price_range": writePrice,
		"is_restricted":     isRestricted,
	}

	allocationData, _ := json.Marshal(allocationRequest)

	params := make(map[string]string)
	params["allocation_data"] = string(allocationData)
	if len(force) > 0 && force[0] {
		params["force"] = strconv.FormatBool(force[0])
	}

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
	size int64,
	readPrice, writePrice PriceRange,
	preferredBlobberIds, blobberAuthTickets []string, force bool,
) (map[string]interface{}, error) {
	for _, authTicket := range blobberAuthTickets {
		if len(authTicket) > 0 {
			return map[string]interface{}{
				"data_shards":          datashards,
				"parity_shards":        parityshards,
				"size":                 size,
				"blobbers":             preferredBlobberIds,
				"blobber_auth_tickets": blobberAuthTickets,
				"read_price_range":     readPrice,
				"write_price_range":    writePrice,
			}, nil
		}
	}

	allocBlobberIDs, err := GetAllocationBlobbers(
		datashards, parityshards, size, 2, readPrice, writePrice, force,
	)
	if err != nil {
		return nil, err
	}

	blobbers := append(preferredBlobberIds, allocBlobberIDs...)

	// filter duplicates
	ids := make(map[string]bool)
	uniqueBlobbers := []string{}
	uniqueBlobberAuthTickets := []string{}

	for _, b := range blobbers {
		if !ids[b] {
			uniqueBlobbers = append(uniqueBlobbers, b)
			uniqueBlobberAuthTickets = append(uniqueBlobberAuthTickets, "")
			ids[b] = true
		}
	}

	return map[string]interface{}{
		"data_shards":          datashards,
		"parity_shards":        parityshards,
		"size":                 size,
		"blobbers":             uniqueBlobbers,
		"blobber_auth_tickets": uniqueBlobberAuthTickets,
		"read_price_range":     readPrice,
		"write_price_range":    writePrice,
	}, nil
}

// GetBlobberIds returns a list of blobber ids that can be used for a new allocation.
//
//   - blobberUrls is a list of blobber urls.
//
// returns a list of blobber ids that can be used for the new allocation and an error if any.
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

// GetFreeAllocationBlobbers returns a list of blobber ids that can be used for a new free allocation.
//
//   - request is the request data for the free allocation.
//
// returns a list of blobber ids that can be used for the new free allocation and an error if any.
func GetFreeAllocationBlobbers(request map[string]interface{}) ([]string, error) {
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

// AddFreeStorageAssigner adds a new free storage assigner (txn: `storagesc.add_free_allocation_assigner`).
// The free storage assigner is used to create free allocations. Can only be called by chain owner.
//
//   - name is the name of the assigner.
//   - publicKey is the public key of the assigner.
//   - individualLimit is the individual limit of the assigner for a single free allocation request
//   - totalLimit is the total limit of the assigner for all free allocation requests.
//
// returns the hash of the transaction, the nonce of the transaction and an error if any.
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
	hash, _, n, _, err := storageSmartContractTxn(sn)

	return hash, n, err
}

// FinalizeAllocation sends a finalize request for an allocation (txn: `storagesc.finalize_allocation`)
//
//   - allocID is the id of the allocation.
//
// returns the hash of the transaction, the nonce of the transaction and an error if any.
func FinalizeAllocation(allocID string) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_FINALIZE_ALLOCATION,
		InputArgs: map[string]interface{}{"allocation_id": allocID},
	}
	hash, _, nonce, _, err = storageSmartContractTxn(sn)
	return
}

// CancelAllocation sends a cancel request for an allocation (txn: `storagesc.cancel_allocation`)
//
//   - allocID is the id of the allocation.
//
// returns the hash of the transaction, the nonce of the transaction and an error if any.
func CancelAllocation(allocID string) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_CANCEL_ALLOCATION,
		InputArgs: map[string]interface{}{"allocation_id": allocID},
	}
	hash, _, nonce, _, err = storageSmartContractTxn(sn)
	return
}

// ProviderType is the type of the provider.
type ProviderType int

const (
	ProviderMiner ProviderType = iota + 1
	ProviderSharder
	ProviderBlobber
	ProviderValidator
	ProviderAuthorizer
)

// KillProvider kills a blobber or a validator (txn: `storagesc.kill_blobber` or `storagesc.kill_validator`)
//   - providerId is the id of the provider.
//   - providerType` is the type of the provider, either 3 for `ProviderBlobber` or 4 for `ProviderValidator.
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
	hash, _, n, _, err := storageSmartContractTxn(sn)
	return hash, n, err
}

// ShutdownProvider shuts down a blobber or a validator (txn: `storagesc.shutdown_blobber` or `storagesc.shutdown_validator`)
//   - providerId is the id of the provider.
//   - providerType` is the type of the provider, either 3 for `ProviderBlobber` or 4 for `ProviderValidator.
func ShutdownProvider(providerType ProviderType, providerID string) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var input = map[string]interface{}{
		"provider_id": providerID,
	}

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
	hash, _, n, _, err := storageSmartContractTxn(sn)
	return hash, n, err
}

// CollectRewards collects the rewards for a provider (txn: `storagesc.collect_reward`)
//   - providerId is the id of the provider.
//   - providerType is the type of the provider.
func CollectRewards(providerId string, providerType ProviderType) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var input = map[string]interface{}{
		"provider_id":   providerId,
		"provider_type": providerType,
	}

	var sn = transaction.SmartContractTxnData{
		InputArgs: input,
	}

	var scAddress string
	switch providerType {
	case ProviderBlobber, ProviderValidator:
		scAddress = STORAGE_SCADDRESS
		sn.Name = transaction.STORAGESC_COLLECT_REWARD
	case ProviderMiner, ProviderSharder:
		scAddress = MINERSC_SCADDRESS
		sn.Name = transaction.MINERSC_COLLECT_REWARD
	// case ProviderAuthorizer:
	// 	scAddress = ZCNSC_SCADDRESS
	// 	sn.Name = transaction.ZCNSC_COLLECT_REWARD
	default:
		return "", 0, fmt.Errorf("collect rewards provider type %v not implimented", providerType)
	}

	hash, _, n, _, err := smartContractTxn(scAddress, sn)
	return hash, n, err
}

// TransferAllocation transfers the ownership of an allocation to a new owner. (txn: `storagesc.update_allocation_request`)
//
//   - allocationId is the id of the allocation.
//   - newOwner is the client id of the new owner.
//   - newOwnerPublicKey is the public key of the new owner.
//
// returns the hash of the transaction, the nonce of the transaction and an error if any.
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
	hash, _, n, _, err := storageSmartContractTxn(sn)
	return hash, n, err
}

// UpdateBlobberSettings updates the settings of a blobber (txn: `storagesc.update_blobber_settings`)
//   - blob is the update blobber request inputs.
func UpdateBlobberSettings(blob *UpdateBlobber) (resp string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}
	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS,
		InputArgs: blob,
	}
	resp, _, nonce, _, err = storageSmartContractTxn(sn)
	return
}

// UpdateValidatorSettings updates the settings of a validator (txn: `storagesc.update_validator_settings`)
//   - v is the update validator request inputs.
func UpdateValidatorSettings(v *UpdateValidator) (resp string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_VALIDATOR_SETTINGS,
		InputArgs: v.ConvertToValidationNode(),
	}
	resp, _, nonce, _, err = storageSmartContractTxn(sn)
	return
}

// ResetBlobberStats resets the stats of a blobber (txn: `storagesc.reset_blobber_stats`)
//   - rbs is the reset blobber stats dto, contains the blobber id and its stats.
func ResetBlobberStats(rbs *ResetBlobberStatsDto) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_RESET_BLOBBER_STATS,
		InputArgs: rbs,
	}
	hash, _, n, _, err := storageSmartContractTxn(sn)
	return hash, n, err
}

func ResetAllocationStats(allocationId string) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_RESET_ALLOCATION_STATS,
		InputArgs: allocationId,
	}
	hash, _, n, _, err := storageSmartContractTxn(sn)
	return hash, n, err
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

// GetAllocationMinLock calculates and returns the minimum lock demand for creating a new allocation, which represents the cost of the creation process.
//   - datashards is the number of data shards for the allocation.
//   - parityshards is the number of parity shards for the allocation.
//   - size is the size of the allocation.
//   - writePrice is the write price range for the allocation.
//
// returns the minimum lock demand for the creation process and an error if any.
func GetAllocationMinLock(
	datashards, parityshards int,
	size int64,
	writePrice PriceRange,
) (int64, error) {
	baSize := int64(math.Ceil(float64(size) / float64(datashards)))
	totalSize := baSize * int64(datashards+parityshards)

	sizeInGB := float64(totalSize) / GB

	cost := sizeInGB * float64(writePrice.Max)
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

// GetUpdateAllocationMinLock returns the minimum lock demand for updating an allocation, which represents the cost of the update operation.
//
//   - allocationID is the id of the allocation.
//   - size is the new size of the allocation.
//   - extend is a flag indicating whether to extend the expiry of the allocation.
//   - addBlobberId is the id of the blobber to add to the allocation.
//   - removeBlobberId is the id of the blobber to remove from the allocation.
//
// returns the minimum lock demand for the update operation and an error if any.
func GetUpdateAllocationMinLock(
	allocationID string,
	size int64,
	extend bool,
	addBlobberId,
	removeBlobberId string) (int64, error) {
	updateAllocationRequest := make(map[string]interface{})
	updateAllocationRequest["owner_id"] = client.GetClientID()
	updateAllocationRequest["owner_public_key"] = ""
	updateAllocationRequest["id"] = allocationID
	updateAllocationRequest["size"] = size
	updateAllocationRequest["extend"] = extend
	updateAllocationRequest["add_blobber_id"] = addBlobberId
	updateAllocationRequest["remove_blobber_id"] = removeBlobberId

	data, err := json.Marshal(updateAllocationRequest)
	if err != nil {
		return 0, errors.Wrap(err, "failed to encode request into json")
	}

	params := make(map[string]string)
	params["data"] = string(data)

	responseBytes, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/allocation-update-min-lock", params, nil)
	if err != nil {
		return 0, errors.Wrap(err, "failed to request allocation update min lock")
	}

	var response = make(map[string]int64)
	if err = json.Unmarshal(responseBytes, &response); err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf("failed to decode response: %s", string(responseBytes)))
	}

	v, ok := response["min_lock_demand"]
	if !ok {
		return 0, errors.New("", "min_lock_demand not found in response")
	}
	return v, nil
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
