package zcncore

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	stdErrors "errors"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/tokenrate"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	openssl "github.com/Luzifer/go-openssl/v3"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	GET_CLIENT                       = `/v1/client/get`
	PUT_TRANSACTION                  = `/v1/transaction/put`
	TXN_VERIFY_URL                   = `/v1/transaction/get/confirmation?hash=`
	GET_BLOCK_INFO                   = `/v1/block/get?`
	GET_MAGIC_BLOCK_INFO             = `/v1/block/magic/get?`
	GET_LATEST_FINALIZED             = `/v1/block/get/latest_finalized`
	GET_LATEST_FINALIZED_MAGIC_BLOCK = `/v1/block/get/latest_finalized_magic_block`
	GET_FEE_STATS                    = `/v1/block/get/fee_stats`
	GET_CHAIN_STATS                  = `/v1/chain/get/stats`
	// vesting SC

	VESTINGSC_PFX = `/v1/screst/` + VestingSmartContractAddress

	GET_VESTING_CONFIG       = VESTINGSC_PFX + `/vesting-config`
	GET_VESTING_POOL_INFO    = VESTINGSC_PFX + `/getPoolInfo`
	GET_VESTING_CLIENT_POOLS = VESTINGSC_PFX + `/getClientPools`

	// faucet sc

	FAUCETSC_PFX        = `/v1/screst/` + FaucetSmartContractAddress
	GET_FAUCETSC_CONFIG = FAUCETSC_PFX + `/faucet-config`

	// zcn sc
	ZCNSC_PFX                      = `/v1/screst/` + ZCNSCSmartContractAddress
	GET_MINT_NONCE                 = ZCNSC_PFX + `/v1/mint_nonce`
	GET_NOT_PROCESSED_BURN_TICKETS = ZCNSC_PFX + `/v1/not_processed_burn_tickets`
	GET_AUTHORIZER                 = ZCNSC_PFX + `/getAuthorizer`

	// miner SC

	MINERSC_PFX          = `/v1/screst/` + MinerSmartContractAddress
	GET_MINERSC_NODE     = MINERSC_PFX + "/nodeStat"
	GET_MINERSC_POOL     = MINERSC_PFX + "/nodePoolStat"
	GET_MINERSC_CONFIG   = MINERSC_PFX + "/configs"
	GET_MINERSC_GLOBALS  = MINERSC_PFX + "/globalSettings"
	GET_MINERSC_USER     = MINERSC_PFX + "/getUserPools"
	GET_MINERSC_MINERS   = MINERSC_PFX + "/getMinerList"
	GET_MINERSC_SHARDERS = MINERSC_PFX + "/getSharderList"
	GET_MINERSC_EVENTS   = MINERSC_PFX + "/getEvents"

	// storage SC

	STORAGESC_PFX = "/v1/screst/" + StorageSmartContractAddress

	STORAGESC_GET_SC_CONFIG            = STORAGESC_PFX + "/storage-config"
	STORAGESC_GET_CHALLENGE_POOL_INFO  = STORAGESC_PFX + "/getChallengePoolStat"
	STORAGESC_GET_ALLOCATION           = STORAGESC_PFX + "/allocation"
	STORAGESC_GET_ALLOCATIONS          = STORAGESC_PFX + "/allocations"
	STORAGESC_GET_READ_POOL_INFO       = STORAGESC_PFX + "/getReadPoolStat"
	STORAGESC_GET_STAKE_POOL_INFO      = STORAGESC_PFX + "/getStakePoolStat"
	STORAGESC_GET_STAKE_POOL_USER_INFO = STORAGESC_PFX + "/getUserStakePoolStat"
	STORAGESC_GET_USER_LOCKED_TOTAL    = STORAGESC_PFX + "/getUserLockedTotal"
	STORAGESC_GET_BLOBBERS             = STORAGESC_PFX + "/getblobbers"
	STORAGESC_GET_BLOBBER              = STORAGESC_PFX + "/getBlobber"
	STORAGESC_GET_VALIDATOR            = STORAGESC_PFX + "/get_validator"
	STORAGESC_GET_TRANSACTIONS         = STORAGESC_PFX + "/transactions"

	STORAGE_GET_SNAPSHOT            = STORAGESC_PFX + "/replicate-snapshots"
	STORAGE_GET_BLOBBER_SNAPSHOT    = STORAGESC_PFX + "/replicate-blobber-aggregates"
	STORAGE_GET_MINER_SNAPSHOT      = STORAGESC_PFX + "/replicate-miner-aggregates"
	STORAGE_GET_SHARDER_SNAPSHOT    = STORAGESC_PFX + "/replicate-sharder-aggregates"
	STORAGE_GET_AUTHORIZER_SNAPSHOT = STORAGESC_PFX + "/replicate-authorizer-aggregates"
	STORAGE_GET_VALIDATOR_SNAPSHOT  = STORAGESC_PFX + "/replicate-validator-aggregates"
	STORAGE_GET_USER_SNAPSHOT       = STORAGESC_PFX + "/replicate-user-aggregates"
)

const (
	StorageSmartContractAddress  = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7`
	VestingSmartContractAddress  = `2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead`
	FaucetSmartContractAddress   = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d3`
	MultiSigSmartContractAddress = `27b5ef7120252b79f9dd9c05505dd28f328c80f6863ee446daede08a84d651a7`
	MinerSmartContractAddress    = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9`
	ZCNSCSmartContractAddress    = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712e0`
	MultiSigRegisterFuncName     = "register"
	MultiSigVoteFuncName         = "vote"
)

// In percentage
const consensusThresh = 25

const (
	defaultMinSubmit               = int(10)
	defaultMinConfirmation         = int(10)
	defaultConfirmationChainLength = int(3)
	defaultTxnExpirationSeconds    = 60
	defaultWaitSeconds             = 3 * time.Second
)

const (
	StatusSuccess      int = 0
	StatusNetworkError int = 1
	// TODO: Change to specific error
	StatusError            int = 2
	StatusRejectedByUser   int = 3
	StatusInvalidSignature int = 4
	StatusAuthError        int = 5
	StatusAuthVerifyFailed int = 6
	StatusAuthTimeout      int = 7
	StatusUnknown          int = -1
)

var defaultLogLevel = logger.DEBUG
var logging logger.Logger

// GetLogger returns the logger instance
func GetLogger() *logger.Logger {
	return &logging
}

// CloseLog closes log file
func CloseLog() {
	logging.Close()
}

const TOKEN_UNIT int64 = 1e10

const (
	OpGetTokenLockConfig int = iota
	OpGetLockedTokens
	OpGetUserPools
	OpGetUserPoolDetail
	OpGetNotProcessedBurnTickets
	OpGetMintNonce
	// storage SC ops
	// OpStorageSCGetConfig Get global storage SC config
	OpStorageSCGetConfig

	// OpStorageSCGetChallengePoolInfo Get challenge pool info
	OpStorageSCGetChallengePoolInfo

	// OpStorageSCGetAllocation Get allocation info
	OpStorageSCGetAllocation

	// OpStorageSCGetAllocations Get all allocations
	OpStorageSCGetAllocations

	// OpStorageSCGetReadPoolInfo Get read pool info
	OpStorageSCGetReadPoolInfo

	// OpStorageSCGetStakePoolInfo Get stake pool info
	OpStorageSCGetStakePoolInfo

	// OpStorageSCGetStakePoolUserInfo Get blobbers
	OpStorageSCGetBlobbers

	// OpStorageSCGetBlobber Get blobber information
	OpStorageSCGetBlobber

	// OpStorageSCGetValidator Get transaction info
	OpStorageSCGetTransactions

	// OpStorageSCGetSnapshots Get global snapshots
	OpStorageSCGetSnapshots

	// OpStorageSCGetBlobberSnapshots Get blobber snapshots
	OpStorageSCGetBlobberSnapshots

	// OpStorageSCGetMinerSnapshots Get miner snapshots
	OpStorageSCGetMinerSnapshots

	// OpStorageSCGetSharderSnapshots Get sharder snapshots
	OpStorageSCGetSharderSnapshots

	// OpStorageSCGetAuthorizerSnapshots Get authorizer snapshots
	OpStorageSCGetAuthorizerSnapshots

	// OpStorageSCGetValidatorSnapshots Get validator snapshots
	OpStorageSCGetValidatorSnapshots

	// OpStorageSCGetUserSnapshots Get user snapshots
	OpStorageSCGetUserSnapshots

	// OpStorageSCGetUserLockedTotal Get global configuration
	OpZCNSCGetGlobalConfig

	// OpZCNSCGetMintNonce Get authorizer information
	OpZCNSCGetAuthorizer

	// OpZCNSCGetAuthorizerNodes Get authorizer nodes
	OpZCNSCGetAuthorizerNodes
)

// WalletCallback needs to be implemented for wallet creation.
type WalletCallback interface {
	OnWalletCreateComplete(status int, wallet string, err string)
}

// GetBalanceCallback needs to be implemented by the caller of GetBalance() to get the status
type GetBalanceCallback interface {
	OnBalanceAvailable(status int, value int64, info string)
}

// BurnTicket represents the burn ticket of native ZCN tokens used by the bridge protocol to mint ERC20 tokens
type BurnTicket struct {
	Hash   string `json:"hash"`
	Amount int64  `json:"amount"`
	Nonce  int64  `json:"nonce"`
}

// GetNonceCallback needs to be implemented by the caller of GetNonce() to get the status
type GetNonceCallback interface {
	OnNonceAvailable(status int, nonce int64, info string)
}

type GetNonceCallbackStub struct {
	status int
	nonce  int64
	info   string
}

func (g *GetNonceCallbackStub) OnNonceAvailable(status int, nonce int64, info string) {
	g.status = status
	g.nonce = nonce
	g.info = info
}

// GetInfoCallback represents the functions that will be called when the response of a GET request to the sharders is available
type GetInfoCallback interface {
	// OnInfoAvailable will be called when GetLockTokenConfig is complete
	// if status == StatusSuccess then info is valid
	// is status != StatusSuccess then err will give the reason
	OnInfoAvailable(op int, status int, info string, err string)
}

// AuthCallback needs to be implemented by the caller SetupAuth()
type AuthCallback interface {
	// This call back gives the status of the Two factor authenticator(zauth) setup.
	OnSetupComplete(status int, err string)
}

// Singleton
var _config localConfig
var miners []string
var mGuard sync.Mutex

func init() {
	logging.Init(defaultLogLevel, "0chain-core-sdk")
}

func GetStableMiners() []string {
	mGuard.Lock()
	defer mGuard.Unlock()
	if len(miners) == 0 {
		miners = util.GetRandom(_config.chain.Miners, getMinMinersSubmit())
	}

	return miners
}
func ResetStableMiners() {
	mGuard.Lock()
	defer mGuard.Unlock()
	miners = util.GetRandom(_config.chain.Miners, getMinMinersSubmit())
}

func checkSdkInit() error {
	if !_config.isConfigured || len(_config.chain.Miners) < 1 || len(_config.chain.Sharders) < 1 {
		return errors.New("", "SDK not initialized")
	}
	return nil
}
func checkWalletConfig() error {
	if !_config.isValidWallet || _config.wallet.ClientID == "" {
		logging.Error("wallet info not found. returning error.")
		return errors.New("", "wallet info not found. set wallet info")
	}
	return nil
}
func CheckConfig() error {

	err := checkSdkInit()
	if err != nil {
		return err
	}
	err = checkWalletConfig()
	if err != nil {
		return err
	}
	return nil
}

func assertConfig() {
	if _config.chain.MinSubmit <= 0 {
		_config.chain.MinSubmit = defaultMinSubmit
	}
	if _config.chain.MinConfirmation <= 0 {
		_config.chain.MinConfirmation = defaultMinConfirmation
	}
	if _config.chain.ConfirmationChainLength <= 0 {
		_config.chain.ConfirmationChainLength = defaultConfirmationChainLength
	}
}
func getMinMinersSubmit() int {
	minMiners := util.MaxInt(calculateMinRequired(float64(_config.chain.MinSubmit), float64(len(_config.chain.Miners))/100), 1)
	logging.Info("Minimum miners used for submit :", minMiners)
	return minMiners
}

func GetMinShardersVerify() int {
	return getMinShardersVerify()
}

func getMinShardersVerify() int {
	minSharders := util.MaxInt(calculateMinRequired(float64(_config.chain.MinConfirmation), float64(len(Sharders.Healthy()))/100), 1)
	logging.Info("Minimum sharders used for verify :", minSharders)
	return minSharders
}
func getMinRequiredChainLength() int64 {
	return int64(_config.chain.ConfirmationChainLength)
}

func calculateMinRequired(minRequired, percent float64) int {
	return int(math.Ceil(minRequired * percent))
}

// GetVersion - returns version string
func GetVersion() string {
	return version.VERSIONSTR
}

// SetLogLevel set the log level.
//   - lvl: 0 disabled; higher number (upto 4) more verbosity
func SetLogLevel(lvl int) {
	logging.SetLevel(lvl)
}

// SetLogFile - sets file path to write log
//   - logFile: log file path
//   - verbose: true - console output; false - no console output
func SetLogFile(logFile string, verbose bool) {
	ioWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100, // MB
		MaxBackups: 5,   // number of backups
		MaxAge:     28,  //days
		LocalTime:  false,
		Compress:   false, // disabled by default
	}
	logging.SetLogFile(ioWriter, verbose)
	logging.Info("******* Wallet SDK Version:", version.VERSIONSTR, " ******* (SetLogFile)")
}

// Init initialize the SDK with miner, sharder and signature scheme provided in configuration provided in JSON format
// # Inputs
//   - chainConfigJSON: json format of zcn config
//     {
//     "block_worker": "https://dev.0chain.net/dns",
//     "signature_scheme": "bls0chain",
//     "min_submit": 50,
//     "min_confirmation": 50,
//     "confirmation_chain_length": 3,
//     "max_txn_query": 5,
//     "query_sleep_time": 5,
//     "preferred_blobbers": ["https://dev.0chain.net/blobber02","https://dev.0chain.net/blobber03"],
//     "chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe",
//     "ethereum_node":"https://ropsten.infura.io/v3/xxxxxxxxxxxxxxx",
//     "zbox_host":"https://0box.dev.0chain.net",
//     "zbox_app_type":"vult",
//     "sharder_consensous": 2,
//     }
func Init(chainConfigJSON string) error {
	err := json.Unmarshal([]byte(chainConfigJSON), &_config.chain)
	if err == nil {
		// Check signature scheme is supported
		if _config.chain.SignatureScheme != "ed25519" && _config.chain.SignatureScheme != "bls0chain" {
			return errors.New("", "invalid/unsupported signature scheme")
		}

		err = UpdateNetworkDetails()
		if err != nil {
			return err
		}

		go updateNetworkDetailsWorker(context.Background())

		assertConfig()
		_config.isConfigured = true

		cfg := &conf.Config{
			BlockWorker:             _config.chain.BlockWorker,
			MinSubmit:               _config.chain.MinSubmit,
			MinConfirmation:         _config.chain.MinConfirmation,
			ConfirmationChainLength: _config.chain.ConfirmationChainLength,
			SignatureScheme:         _config.chain.SignatureScheme,
			ChainID:                 _config.chain.ChainID,
			EthereumNode:            _config.chain.EthNode,
			SharderConsensous:       _config.chain.SharderConsensous,
		}

		conf.InitClientConfig(cfg)
	}
	logging.Info("0chain: test logging")
	logging.Info("******* Wallet SDK Version:", version.VERSIONSTR, " ******* (Init) Test")
	return err
}

// InitSignatureScheme initializes signature scheme only.
//   - scheme: signature scheme
func InitSignatureScheme(scheme string) {
	_config.chain.SignatureScheme = scheme
}

// CreateWalletOffline creates the wallet for the config signature scheme.
func CreateWalletOffline() (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	wallet, err := sigScheme.GenerateKeys()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate keys")
	}
	w, err := wallet.Marshal()
	if err != nil {
		return "", errors.Wrap(err, "wallet encoding failed")
	}
	return w, nil
}

// RecoverOfflineWallet recovers the previously generated wallet using the mnemonic.
//   - mnemonic: mnemonics to recover
func RecoverOfflineWallet(mnemonic string) (string, error) {
	if !zcncrypto.IsMnemonicValid(mnemonic) {
		return "", errors.New("", "Invalid mnemonic")
	}

	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	wallet, err := sigScheme.RecoverKeys(mnemonic)
	if err != nil {
		return "", err
	}

	walletString, err := wallet.Marshal()
	if err != nil {
		return "", err
	}

	return walletString, nil
}

// RecoverWallet recovers the previously generated wallet using the mnemonic.
// It also registers the wallet again to block chain.
func RecoverWallet(mnemonic string, statusCb WalletCallback) error {
	if !zcncrypto.IsMnemonicValid(mnemonic) {
		return errors.New("", "Invalid mnemonic")
	}
	go func() {
		sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
		_, err := sigScheme.RecoverKeys(mnemonic)
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", err.Error())
			return
		}
	}()
	return nil
}

// Split keys from the primary master key
func SplitKeys(privateKey string, numSplits int) (string, error) {
	w, err := SplitKeysWallet(privateKey, numSplits)
	if err != nil {
		return "", errors.Wrap(err, "split key failed.")
	}
	wStr, err := w.Marshal()
	if err != nil {
		return "", errors.Wrap(err, "wallet encoding failed.")
	}
	return wStr, nil
}

func SplitKeysWallet(privateKey string, numSplits int) (*zcncrypto.Wallet, error) {
	if _config.chain.SignatureScheme != "bls0chain" {
		return nil, errors.New("", "signature key doesn't support split key")
	}
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "set private key failed")
	}
	w, err := sigScheme.SplitKeys(numSplits)
	if err != nil {
		return nil, errors.Wrap(err, "split key failed.")
	}

	w.IsSplit = true

	return w, nil
}

type GetClientResponse struct {
	ID           string `json:"id"`
	Version      string `json:"version"`
	CreationDate int    `json:"creation_date"`
	PublicKey    string `json:"public_key"`
}

func GetClientDetails(clientID string) (*GetClientResponse, error) {
	minerurl := util.GetRandom(_config.chain.Miners, 1)[0]
	url := minerurl + GET_CLIENT
	url = fmt.Sprintf("%v?id=%v", url, clientID)
	req, err := util.NewHTTPGetRequest(url)
	if err != nil {
		logging.Error(minerurl, "new get request failed. ", err.Error())
		return nil, err
	}
	res, err := req.Get()
	if err != nil {
		logging.Error(minerurl, "send error. ", err.Error())
		return nil, err
	}

	var clientDetails GetClientResponse
	err = json.Unmarshal([]byte(res.Body), &clientDetails)
	if err != nil {
		return nil, err
	}

	return &clientDetails, nil
}

// IsMnemonicValid is a utility function to check the mnemonic valid
//   - mnemonic: mnemonics to check
func IsMnemonicValid(mnemonic string) bool {
	return zcncrypto.IsMnemonicValid(mnemonic)
}

// SetWallet should be set before any transaction or client specific APIs
// splitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
func SetWallet(w zcncrypto.Wallet, splitKeyWallet bool) error {
	_config.wallet = w

	if _config.chain.SignatureScheme == "bls0chain" {
		_config.isSplitWallet = splitKeyWallet
	}
	_config.isValidWallet = true

	return nil
}

func GetWalletRaw() zcncrypto.Wallet {
	return _config.wallet
}

// SetWalletInfo should be set before any transaction or client specific APIs.
// "splitKeyWallet" parameter is valid only if SignatureScheme is "BLS0Chain"
//   - jsonWallet: json format of wallet
//   - splitKeyWallet: if wallet keys is split
func SetWalletInfo(jsonWallet string, splitKeyWallet bool) error {
	err := json.Unmarshal([]byte(jsonWallet), &_config.wallet)
	if err != nil {
		return err
	}

	if _config.chain.SignatureScheme == "bls0chain" {
		_config.isSplitWallet = splitKeyWallet
	}
	_config.isValidWallet = true

	return nil
}

// SetAuthUrl will be called by app to set zauth URL to SDK.
// # Inputs
//   - url: the url of zAuth server
func SetAuthUrl(url string) error {
	if !_config.isSplitWallet {
		return errors.New("", "wallet type is not split key")
	}
	if url == "" {
		return errors.New("", "invalid auth url")
	}
	_config.authUrl = strings.TrimRight(url, "/")
	return nil
}

func getWalletBalance(clientId string) (common.Balance, int64, error) {
	err := checkSdkInit()
	if err != nil {
		return 0, 0, err
	}

	cb := &walletCallback{}
	cb.Add(1)

	go func() {
		value, info, err := getBalanceFromSharders(clientId)
		if err != nil && strings.TrimSpace(info) != `{"error":"value not present"}` {
			cb.OnBalanceAvailable(StatusError, value, info)
			cb.err = err
			return
		}
		cb.OnBalanceAvailable(StatusSuccess, value, info)
	}()

	cb.Wait()

	var clientState struct {
		Nonce int64 `json:"nonce"`
	}
	err = json.Unmarshal([]byte(cb.info), &clientState)
	if err != nil {
		return 0, 0, err
	}

	return cb.balance, clientState.Nonce, cb.err
}

// GetBalance retrieve wallet balance from sharders
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetBalance(cb GetBalanceCallback) error {
	err := CheckConfig()
	if err != nil {
		return err
	}
	go func() {
		value, info, err := getBalanceFromSharders(_config.wallet.ClientID)
		if err != nil {
			logging.Error(err)
			cb.OnBalanceAvailable(StatusError, 0, info)
			return
		}
		cb.OnBalanceAvailable(StatusSuccess, value, info)
	}()
	return nil
}

// GetMintNonce retrieve the client's latest mint nonce from sharders
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetMintNonce(cb GetInfoCallback) error {
	err := CheckConfig()
	if err != nil {
		return err
	}

	go GetInfoFromSharders(withParams(GET_MINT_NONCE, Params{
		"client_id": _config.wallet.ClientID,
	}), OpGetMintNonce, cb)
	return nil
}

// GetNotProcessedZCNBurnTickets retrieve burn tickets that are not compensated by minting
//   - ethereumAddress: ethereum address for the issuer of the burn tickets
//   - startNonce: start nonce for the burn tickets
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetNotProcessedZCNBurnTickets(ethereumAddress, startNonce string, cb GetInfoCallback) error {
	err := CheckConfig()
	if err != nil {
		return err
	}

	go GetInfoFromSharders(withParams(GET_NOT_PROCESSED_BURN_TICKETS, Params{
		"ethereum_address": ethereumAddress,
		"nonce":            startNonce,
	}), OpGetNotProcessedBurnTickets, cb)

	return nil
}

// GetNonce retrieve wallet nonce from sharders
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetNonce(cb GetNonceCallback) error {
	if cb == nil {
		cb = &GetNonceCallbackStub{}
	}

	err := CheckConfig()
	if err != nil {
		return err
	}

	go func() {
		value, info, err := getNonceFromSharders(_config.wallet.ClientID)
		if err != nil {
			logging.Error(err)
			cb.OnNonceAvailable(StatusError, 0, info)
			return
		}

		cb.OnNonceAvailable(StatusSuccess, value, info)
	}()

	return nil
}

// GetWalletBalance retrieve wallet nonce from sharders
//   - clientID: client id
func GetWalletNonce(clientID string) (int64, error) {
	cb := &GetNonceCallbackStub{}

	err := CheckConfig()
	if err != nil {
		return 0, err
	}
	wait := &sync.WaitGroup{}
	wait.Add(1)
	go func() {
		defer wait.Done()
		value, info, err := getNonceFromSharders(clientID)
		if err != nil {
			logging.Error(err)
			cb.OnNonceAvailable(StatusError, 0, info)
			return
		}
		cb.OnNonceAvailable(StatusSuccess, value, info)
	}()

	wait.Wait()

	if cb.status == StatusSuccess {
		return cb.nonce, nil
	}

	return 0, stdErrors.New(cb.info)
}

// GetBalanceWallet retreives wallet balance from sharders
//   - walletStr: wallet string
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetBalanceWallet(walletStr string, cb GetBalanceCallback) error {
	w, err := getWallet(walletStr)
	if err != nil {
		fmt.Printf("Error while parsing the wallet. %v\n", err)
		return err
	}

	go func() {
		value, info, err := getBalanceFromSharders(w.ClientID)
		if err != nil {
			logging.Error(err)
			cb.OnBalanceAvailable(StatusError, 0, info)
			return
		}
		cb.OnBalanceAvailable(StatusSuccess, value, info)
	}()
	return nil
}

func getBalanceFromSharders(clientID string) (int64, string, error) {
	return Sharders.GetBalanceFieldFromSharders(clientID, "balance")
}

func getNonceFromSharders(clientID string) (int64, string, error) {
	return Sharders.GetBalanceFieldFromSharders(clientID, "nonce")
}

// ConvertToToken converts the SAS tokens to ZCN tokens
//   - token: SAS tokens amount
func ConvertToToken(token int64) float64 {
	return float64(token) / float64(common.TokenUnit)
}

// ConvertTokenToUSD converts the ZCN tokens to USD amount
//   - token: ZCN tokens amount
func ConvertTokenToUSD(token float64) (float64, error) {
	zcnRate, err := getTokenUSDRate()
	if err != nil {
		return 0, err
	}
	return token * zcnRate, nil
}

// ConvertUSDToToken converts the USD amount to ZCN tokens
//   - usd: USD amount
func ConvertUSDToToken(usd float64) (float64, error) {
	zcnRate, err := getTokenUSDRate()
	if err != nil {
		return 0, err
	}
	return usd * (1 / zcnRate), nil
}

func getTokenUSDRate() (float64, error) {
	return tokenrate.GetUSD(context.TODO(), "zcn")
}

// getWallet get a wallet object from a wallet string
func getWallet(walletStr string) (*zcncrypto.Wallet, error) {
	var w zcncrypto.Wallet
	err := json.Unmarshal([]byte(walletStr), &w)
	if err != nil {
		fmt.Printf("error while parsing wallet string.\n%v\n", err)
		return nil, err
	}

	return &w, nil
}

// GetWalletClientID extract wallet client id from wallet string
//   - walletStr: wallet string to get client id
func GetWalletClientID(walletStr string) (string, error) {
	w, err := getWallet(walletStr)
	if err != nil {
		return "", err
	}
	return w.ClientID, nil
}

// GetZcnUSDInfo returns USD value for ZCN token by tokenrate
func GetZcnUSDInfo() (float64, error) {
	return tokenrate.GetUSD(context.TODO(), "zcn")
}

// SetupAuth prepare auth app with clientid, key and a set of public, private key and local publickey
// which is running on PC/Mac.
func SetupAuth(authHost, clientID, clientKey, publicKey, privateKey, localPublicKey string, cb AuthCallback) error {
	go func() {
		authHost = strings.TrimRight(authHost, "/")
		data := map[string]string{"client_id": clientID, "client_key": clientKey, "public_key": publicKey, "private_key": privateKey, "peer_public_key": localPublicKey}
		req, err := util.NewHTTPPostRequest(authHost+"/setup", data)
		if err != nil {
			logging.Error("new post request failed. ", err.Error())
			return
		}
		res, err := req.Post()
		if err != nil {
			logging.Error(authHost+"send error. ", err.Error())
		}
		if res.StatusCode != http.StatusOK {
			cb.OnSetupComplete(StatusError, res.Body)
			return
		}
		cb.OnSetupComplete(StatusSuccess, "")
	}()
	return nil
}

// GetIdForUrl retrieve the ID of the network node (miner/sharder) given its url.
//   - url: url of the node.
func GetIdForUrl(url string) string {
	url = strings.TrimRight(url, "/")
	url = fmt.Sprintf("%v/_nh/whoami", url)
	req, err := util.NewHTTPGetRequest(url)
	if err != nil {
		logging.Error(url, "new get request failed. ", err.Error())
		return ""
	}
	res, err := req.Get()
	if err != nil {
		logging.Error(url, "get error. ", err.Error())
		return ""
	}

	s := strings.Split(res.Body, ",")
	if len(s) >= 3 {
		return s[3]
	}
	return ""
}

//
// vesting pool
//

type Params map[string]string

func (p Params) Query() string {
	if len(p) == 0 {
		return ""
	}
	var params = make(url.Values)
	for k, v := range p {
		params[k] = []string{v}
	}
	return "?" + params.Encode()
}

//
// miner SC
//

// GetMiners obtains list of all active miners.
//   - cb: info callback instance, carries the response of the GET request to the sharders
//   - limit: how many miners should be fetched
//   - offset: how many miners should be skipped
//   - active: retrieve only active miners
//   - stakable: retreive only stakable miners
func GetMiners(cb GetInfoCallback, limit, offset int, active bool, stakable bool) {
	getMinersInternal(cb, active, stakable, limit, offset)
}

func getMinersInternal(cb GetInfoCallback, active, stakable bool, limit, offset int) {
	if err := CheckConfig(); err != nil {
		return
	}

	var url = withParams(GET_MINERSC_MINERS, Params{
		"active":   strconv.FormatBool(active),
		"stakable": strconv.FormatBool(stakable),
		"offset":   strconv.FormatInt(int64(offset), 10),
		"limit":    strconv.FormatInt(int64(limit), 10),
	})

	go GetInfoFromSharders(url, 0, cb)
}

// GetSharders obtains a list of sharders given the following parameters.
//   - cb: info callback instance, carries the response of the GET request to the sharders
//   - limit: how many sharders should be fetched
//   - offset: how many sharders should be skipped
//   - active: retrieve only active sharders
//   - stakable: retrieve only sharders that can be staked
func GetSharders(cb GetInfoCallback, limit, offset int, active, stakable bool) {
	getShardersInternal(cb, active, stakable, limit, offset)
}

func getShardersInternal(cb GetInfoCallback, active, stakable bool, limit, offset int) {
	if err := CheckConfig(); err != nil {
		return
	}

	var url = withParams(GET_MINERSC_SHARDERS, Params{
		"active":   strconv.FormatBool(active),
		"stakable": strconv.FormatBool(stakable),
		"offset":   strconv.FormatInt(int64(offset), 10),
		"limit":    strconv.FormatInt(int64(limit), 10),
	})

	go GetInfoFromSharders(url, 0, cb)
}

func withParams(uri string, params Params) string {
	return uri + params.Query()
}

// GetMinerSCNodeInfo get miner information from sharders
//   - id: the id of miner
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetMinerSCNodeInfo(id string, cb GetInfoCallback) (err error) {

	if err = CheckConfig(); err != nil {
		return
	}

	go GetInfoFromSharders(withParams(GET_MINERSC_NODE, Params{
		"id": id,
	}), 0, cb)
	return
}

// GetMinerSCNodePool get miner smart contract node pool
//   - id: the id of miner
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetMinerSCNodePool(id string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go GetInfoFromSharders(withParams(GET_MINERSC_POOL, Params{
		"id":      id,
		"pool_id": _config.wallet.ClientID,
	}), 0, cb)

	return
}

// GetMinerSCUserInfo retrieve user stake pools for the providers related to the Miner SC (miners/sharders).
//   - clientID: user's wallet id
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetMinerSCUserInfo(clientID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	if clientID == "" {
		clientID = _config.wallet.ClientID
	}
	go GetInfoFromSharders(withParams(GET_MINERSC_USER, Params{
		"client_id": clientID,
	}), 0, cb)

	return
}

// GetMinerSCConfig get miner SC configuration
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetMinerSCConfig(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go GetInfoFromSharders(GET_MINERSC_CONFIG, 0, cb)
	return
}

// GetMinerSCGlobals get miner SC globals
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetMinerSCGlobals(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go GetInfoFromSharders(GET_MINERSC_GLOBALS, 0, cb)
	return
}

//
// Storage SC
//

// GetStorageSCConfig obtains Storage SC configurations.
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetStorageSCConfig(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go GetInfoFromSharders(STORAGESC_GET_SC_CONFIG, OpStorageSCGetConfig, cb)
	return
}

// GetChallengePoolInfo obtains challenge pool information for an allocation.
func GetChallengePoolInfo(allocID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGESC_GET_CHALLENGE_POOL_INFO, Params{
		"allocation_id": allocID,
	})
	go GetInfoFromSharders(url, OpStorageSCGetChallengePoolInfo, cb)
	return
}

// GetAllocation obtains allocation information.
func GetAllocation(allocID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGESC_GET_ALLOCATION, Params{
		"allocation": allocID,
	})
	go GetInfoFromSharders(url, OpStorageSCGetAllocation, cb)
	return
}

// GetAllocations obtains list of allocations of a user.
func GetAllocations(clientID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	if clientID == "" {
		clientID = _config.wallet.ClientID
	}
	var url = withParams(STORAGESC_GET_ALLOCATIONS, Params{
		"client": clientID,
	})
	go GetInfoFromSharders(url, OpStorageSCGetAllocations, cb)
	return
}

// GetSnapshots obtains list of global snapshots, given an initial round and a limit.
// Global snapshots are historical records of some aggregate data related
// to the network (like total staked amount and total reward amount).
//   - round: round number to start fetching snapshots
//   - limit: how many snapshots should be fetched
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetSnapshots(round int64, limit int64, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGE_GET_SNAPSHOT, Params{
		"round": strconv.FormatInt(round, 10),
		"limit": strconv.FormatInt(limit, 10),
	})
	go GetInfoFromAnySharder(url, OpStorageSCGetSnapshots, cb)
	return
}

// GetBlobberSnapshots obtains list of allocations of a blobber.
// Blobber snapshots are historical records of the blobber instance to track its change over time and serve graph requests,
// which are requests that need multiple data points, distributed over an interval of time, usually to plot them on a
// graph.
//   - round: round number
//   - limit: how many blobber snapshots should be fetched
//   - offset: how many blobber snapshots should be skipped
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetBlobberSnapshots(round int64, limit int64, offset int64, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGE_GET_BLOBBER_SNAPSHOT, Params{
		"round":  strconv.FormatInt(round, 10),
		"limit":  strconv.FormatInt(limit, 10),
		"offset": strconv.FormatInt(offset, 10),
	})
	go GetInfoFromAnySharder(url, OpStorageSCGetBlobberSnapshots, cb)
	return
}

// GetMinerSnapshots obtains a list of miner snapshots starting from a specific round.
// Miner snapshots are historical records of the miner instance to track its change over time and serve graph requests,
// which are requests that need multiple data points, distributed over an interval of time, usually to plot them on a
// graph.
//   - round: round number to start fetching snapshots
//   - limit: how many miner snapshots should be fetched
//   - offset: how many miner snapshots should be skipped
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetMinerSnapshots(round int64, limit int64, offset int64, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGE_GET_MINER_SNAPSHOT, Params{
		"round":  strconv.FormatInt(round, 10),
		"limit":  strconv.FormatInt(limit, 10),
		"offset": strconv.FormatInt(offset, 10),
	})
	go GetInfoFromAnySharder(url, OpStorageSCGetMinerSnapshots, cb)
	return
}

// GetSharderSnapshots obtains a list of sharder snapshots starting from a specific round.
// Sharder snapshots are historical records of the sharder instance to track its change over time and serve graph requests,
// which are requests that need multiple data points, distributed over an interval of time, usually to plot them on a
// graph.
//   - round: round number to start fetching snapshots
//   - limit: how many sharder snapshots should be fetched
//   - offset: how many sharder snapshots should be skipped
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetSharderSnapshots(round int64, limit int64, offset int64, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGE_GET_SHARDER_SNAPSHOT, Params{
		"round":  strconv.FormatInt(round, 10),
		"limit":  strconv.FormatInt(limit, 10),
		"offset": strconv.FormatInt(offset, 10),
	})
	go GetInfoFromAnySharder(url, OpStorageSCGetSharderSnapshots, cb)
	return
}

// GetValidatorSnapshots obtains list of validator snapshots from the sharders.
// Validator snapshots are historical records of the validator instance to track its change over time and serve graph requests,
// which are requests that need multiple data points, distributed over an interval of time, usually to plot them on a
// graph.
//   - round: round number to start fetching snapshots
//   - limit: how many validator snapshots should be fetched
//   - offset: how many validator snapshots should be skipped
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetValidatorSnapshots(round int64, limit int64, offset int64, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGE_GET_VALIDATOR_SNAPSHOT, Params{
		"round":  strconv.FormatInt(round, 10),
		"limit":  strconv.FormatInt(limit, 10),
		"offset": strconv.FormatInt(offset, 10),
	})
	go GetInfoFromAnySharder(url, OpStorageSCGetValidatorSnapshots, cb)
	return
}

// GetAuthorizerSnapshots obtains list of authorizers snapshots from the sharders.
// Authorizer snapshots are historical records of the authorizer instance to track its change over time and serve graph requests,
// which are requests that need multiple data points, distributed over an interval of time, usually to plot them on a
// graph.
//   - round: round number to start fetching snapshots
//   - limit: how many authorizer snapshots should be fetched
//   - offset: how many authorizer snapshots should be skipped
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetAuthorizerSnapshots(round int64, limit int64, offset int64, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGE_GET_AUTHORIZER_SNAPSHOT, Params{
		"round":  strconv.FormatInt(round, 10),
		"limit":  strconv.FormatInt(limit, 10),
		"offset": strconv.FormatInt(offset, 10),
	})
	go GetInfoFromAnySharder(url, OpStorageSCGetAuthorizerSnapshots, cb)
	return
}

// GetUserSnapshots replicates user snapshots from the sharders
// User snapshots are historical records of the client data to track its change over time and serve graph requests,
// which are requests that need multiple data points, distributed over an interval of time, usually to plot them on a
// graph.
//   - round: round number to start fetching snapshots
//   - limit: how many user snapshots should be fetched
//   - offset: how many user snapshots should be skipped
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetUserSnapshots(round int64, limit int64, offset int64, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGE_GET_USER_SNAPSHOT, Params{
		"round":  strconv.FormatInt(round, 10),
		"limit":  strconv.FormatInt(limit, 10),
		"offset": strconv.FormatInt(offset, 10),
	})
	go GetInfoFromAnySharder(url, OpStorageSCGetUserSnapshots, cb)
	return
}

// GetReadPoolInfo obtains information about read pool of a user.
func GetReadPoolInfo(clientID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	if clientID == "" {
		clientID = _config.wallet.ClientID
	}
	var url = withParams(STORAGESC_GET_READ_POOL_INFO, Params{
		"client_id": clientID,
	})
	go GetInfoFromSharders(url, OpStorageSCGetReadPoolInfo, cb)
	return
}

// GetStakePoolInfo obtains information about stake pool of a blobber and
// related validator.
func GetStakePoolInfo(blobberID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGESC_GET_STAKE_POOL_INFO, Params{
		"blobber_id": blobberID,
	})
	go GetInfoFromSharders(url, OpStorageSCGetStakePoolInfo, cb)
	return
}

// GetStakePoolUserInfo for a user.
// # Inputs
//   - clientID: the id of wallet
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetStakePoolUserInfo(clientID string, offset, limit int, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	if clientID == "" {
		clientID = _config.wallet.ClientID
	}

	var url = withParams(STORAGESC_GET_STAKE_POOL_USER_INFO, Params{
		"client_id": clientID,
		"offset":    strconv.FormatInt(int64(offset), 10),
		"limit":     strconv.FormatInt(int64(limit), 10),
	})
	go GetInfoFromSharders(url, OpStorageSCGetStakePoolInfo, cb)
	return
}

// GetStakeableBlobbers obtains list of all active blobbers that can be staked (i.e. still number of delegations < max_delegations)
// # Inputs
//   - cb: info callback instance, carries the response of the GET request to the sharders
//   - limit: how many blobbers should be fetched
//   - offset: how many blobbers should be skipped
//   - active: only fetch active blobbers
func GetStakableBlobbers(cb GetInfoCallback, limit, offset int, active bool) {
	getBlobbersInternal(cb, active, limit, offset, true)
}

// GetBlobbers obtains list of all active blobbers.
//   - cb: info callback instance, carries the response of the GET request to the sharders
//   - limit: how many blobbers should be fetched
//   - offset: how many blobbers should be skipped
//   - active: only fetch active blobbers
func GetBlobbers(cb GetInfoCallback, limit, offset int, active bool) {
	getBlobbersInternal(cb, active, limit, offset, false)
}

func getBlobbersInternal(cb GetInfoCallback, active bool, limit, offset int, stakable bool) {
	if err := CheckConfig(); err != nil {
		return
	}

	var url = withParams(STORAGESC_GET_BLOBBERS, Params{
		"active":   strconv.FormatBool(active),
		"offset":   strconv.FormatInt(int64(offset), 10),
		"limit":    strconv.FormatInt(int64(limit), 10),
		"stakable": strconv.FormatBool(stakable),
	})

	go GetInfoFromSharders(url, OpStorageSCGetBlobbers, cb)
}

// GetBlobber obtains blobber information.
//   - blobberID: blobber id
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetBlobber(blobberID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGESC_GET_BLOBBER, Params{
		"blobber_id": blobberID,
	})
	go GetInfoFromSharders(url, OpStorageSCGetBlobber, cb)
	return
}

// GetValidator obtains validator information.
//   - validatorID: validator id
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetValidator(validatorID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(STORAGESC_GET_VALIDATOR, Params{
		"validator_id": validatorID,
	})
	go GetInfoFromSharders(url, OpStorageSCGetBlobber, cb)
	return
}

// GetAuthorizer obtains authorizer information from the sharders.
//   - authorizerID: authorizer id
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetAuthorizer(authorizerID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(GET_AUTHORIZER, Params{
		"id": authorizerID,
	})
	go GetInfoFromSharders(url, OpStorageSCGetBlobber, cb)
	return
}

// GetMinerSharder obtains miner sharder information from the sharders.
//   - id: miner sharder id
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetMinerSharder(id string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = withParams(GET_MINERSC_NODE, Params{
		"id": id,
	})
	go GetInfoFromSharders(url, OpStorageSCGetBlobber, cb)
	return
}

// GetTransactions query transactions from sharders
// # Inputs
//   - toClient:   	receiver
//   - fromClient: 	sender
//   - block_hash: 	block hash
//   - sort:				desc or asc
//   - limit: 			how many transactions should be fetched
//   - offset:			how many transactions should be skipped
//   - cb: 					callback to get result
func GetTransactions(toClient, fromClient, block_hash, sort string, limit, offset int, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}

	params := Params{}
	if toClient != "" {
		params["to_client_id"] = toClient
	}
	if fromClient != "" {
		params["client_id"] = fromClient
	}
	if block_hash != "" {
		params["block_hash"] = block_hash
	}
	if sort != "" {
		params["sort"] = sort
	}
	if limit != 0 {
		l := strconv.Itoa(limit)
		params["limit"] = l
	}
	if offset != 0 {
		o := strconv.Itoa(offset)
		params["offset"] = o
	}

	var u = withParams(STORAGESC_GET_TRANSACTIONS, params)
	go GetInfoFromSharders(u, OpStorageSCGetTransactions, cb)
	return
}

func Encrypt(key, text string) (string, error) {
	keyBytes := []byte(key)
	textBytes := []byte(text)
	response, err := zboxutil.Encrypt(keyBytes, textBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(response), nil
}

// Decrypt decrypts encrypted text using the key.
//   - key: key to use for decryption
//   - text: text to decrypt
func Decrypt(key, text string) (string, error) {
	keyBytes := []byte(key)
	textBytes, _ := hex.DecodeString(text)
	response, err := zboxutil.Decrypt(keyBytes, textBytes)
	if err != nil {
		return "", err
	}
	return string(response), nil
}

func CryptoJsEncrypt(passphrase, message string) (string, error) {
	o := openssl.New()

	enc, err := o.EncryptBytes(passphrase, []byte(message), openssl.DigestMD5Sum)
	if err != nil {
		return "", err
	}

	return string(enc), nil
}

func CryptoJsDecrypt(passphrase, encryptedMessage string) (string, error) {
	o := openssl.New()
	dec, err := o.DecryptBytes(passphrase, []byte(encryptedMessage), openssl.DigestMD5Sum)
	if err != nil {
		return "", err
	}

	return string(dec), nil
}

// GetPublicEncryptionKey returns the public encryption key for the given mnemonic
func GetPublicEncryptionKey(mnemonic string) (string, error) {
	encScheme := encryption.NewEncryptionScheme()
	_, err := encScheme.Initialize(mnemonic)
	if err != nil {
		return "", err
	}
	return encScheme.GetPublicKey()
}
