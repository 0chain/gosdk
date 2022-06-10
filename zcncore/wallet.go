package zcncore

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/tokenrate"
	"github.com/0chain/gosdk/core/transaction"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type ChainConfig struct {
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

var defaultLogLevel = logger.DEBUG
var Logger logger.Logger

const (
	REGISTER_CLIENT                  = `/v1/client/put`
	GET_CLIENT                       = `/v1/client/get`
	PUT_TRANSACTION                  = `/v1/transaction/put`
	TXN_VERIFY_URL                   = `/v1/transaction/get/confirmation?hash=`
	GET_BALANCE                      = `/v1/client/get/balance?client_id=`
	GET_BLOCK_INFO                   = `/v1/block/get?`
	GET_MAGIC_BLOCK_INFO             = `/v1/block/magic/get?`
	GET_LATEST_FINALIZED             = `/v1/block/get/latest_finalized`
	GET_LATEST_FINALIZED_MAGIC_BLOCK = `/v1/block/get/latest_finalized_magic_block`
	GET_CHAIN_STATS                  = `/v1/chain/get/stats`

	// vesting SC

	VESTINGSC_PFX = `/v1/screst/` + VestingSmartContractAddress

	GET_VESTING_CONFIG       = VESTINGSC_PFX + `/vesting-config`
	GET_VESTING_POOL_INFO    = VESTINGSC_PFX + `/getPoolInfo`
	GET_VESTING_CLIENT_POOLS = VESTINGSC_PFX + `/getClientPools`

	// faucet sc

	FAUCETSC_PFX        = `/v1/screst/` + FaucetSmartContractAddress
	GET_FAUCETSC_CONFIG = FAUCETSC_PFX + `/faucet-config`

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
	STORAGESC_GET_BLOBBERS             = STORAGESC_PFX + "/getblobbers"
	STORAGESC_GET_BLOBBER              = STORAGESC_PFX + "/getBlobber"
	STORAGESC_GET_WRITE_POOL_INFO      = STORAGESC_PFX + "/getWritePoolStat"
	STORAGE_GET_TOTAL_STORED_DATA      = STORAGESC_PFX + "/total-stored-data"
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
const consensusThresh = float32(25.0)

const (
	defaultMinSubmit               = int(50)
	defaultMinConfirmation         = int(50)
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

type ConfirmationStatus int

const (
	Undefined ConfirmationStatus = iota
	Success
	ChargeableError
)

const TOKEN_UNIT int64 = 1e10

const (
	OpGetTokenLockConfig int = iota
	OpGetLockedTokens
	OpGetUserPools
	OpGetUserPoolDetail
	// storage SC ops
	OpStorageSCGetConfig
	OpStorageSCGetChallengePoolInfo
	OpStorageSCGetAllocation
	OpStorageSCGetAllocations
	OpStorageSCGetReadPoolInfo
	OpStorageSCGetStakePoolInfo
	OpStorageSCGetBlobbers
	OpStorageSCGetBlobber
	OpStorageSCGetWritePoolInfo
	OpZCNSCGetGlobalConfig
	OpZCNSCGetAuthorizer
	OpZCNSCGetAuthorizerNodes
)

// WalletCallback needs to be implmented for wallet creation.
type WalletCallback interface {
	OnWalletCreateComplete(status int, wallet string, err string)
}

// GetBalanceCallback needs to be implemented by the caller of GetBalance() to get the status
type GetBalanceCallback interface {
	OnBalanceAvailable(status int, value int64, info string)
}

// GetNonceCallback needs to be implemented by the caller of GetNonce() to get the status
type GetNonceCallback interface {
	OnNonceAvailable(status int, nonce int64, info string)
}

type GetNonceCallbackStub struct {
}

func (g *GetNonceCallbackStub) OnNonceAvailable(status int, nonce int64, info string) {
}

// GetInfoCallback needs to be implemented by the caller of GetLockTokenConfig() and GetLockedTokens()
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

type localConfig struct {
	chain         ChainConfig
	wallet        zcncrypto.Wallet
	authUrl       string
	isConfigured  bool
	isValidWallet bool
	isSplitWallet bool
}

// Singleton
var _config localConfig

func init() {
	Logger.Init(defaultLogLevel, "0chain-core-sdk")
}
func checkSdkInit() error {
	if !_config.isConfigured || len(_config.chain.Miners) < 1 || len(_config.chain.Sharders) < 1 {
		return errors.New("", "SDK not initialized")
	}
	return nil
}
func checkWalletConfig() error {
	if !_config.isValidWallet || _config.wallet.ClientID == "" {
		Logger.Error("wallet info not found. returning error.")
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
	Logger.Info("Minimum miners used for submit :", minMiners)
	return minMiners
}

func GetMinShardersVerify() int {
	return getMinShardersVerify()
}

func getMinShardersVerify() int {
	minSharders := util.MaxInt(calculateMinRequired(float64(_config.chain.MinConfirmation), float64(len(_config.chain.Sharders))/100), 1)
	Logger.Info("Minimum sharders used for verify :", minSharders)
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
// lvl - 0 disabled; higher number (upto 4) more verbosity
func SetLogLevel(lvl int) {
	Logger.SetLevel(lvl)
}

// SetLogFile - sets file path to write log
// verbose - true - console output; false - no console output
func SetLogFile(logFile string, verbose bool) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	Logger.SetLogFile(f, verbose)
	Logger.Info("******* Wallet SDK Version:", version.VERSIONSTR, " ******* (SetLogFile)")
}

func GetLogger() *logger.Logger {
	return &Logger
}

// CloseLog closes log file
func CloseLog() {
	Logger.Close()
}

// Init inializes the SDK with miner, sharder and signature scheme provided in
// configuration provided in JSON format
// It is used for 0proxy, 0box, 0explorer, andorid, ios : walletJSON is ChainConfig
//	 {
//      "chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe",
//		"signature_scheme" : "bls0chain",
//		"block_worker" : "http://localhost/dns",
// 		"min_submit" : 50,
//		"min_confirmation" : 50,
//		"confirmation_chain_length" : 3,
//		"num_keys" : 1,
//		"eth_node" : "https://ropsten.infura.io/v3/xxxxxxxxxxxxxxx"
//	 }
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

		go UpdateNetworkDetailsWorker(context.Background())

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
		}

		conf.InitClientConfig(cfg)
	}
	Logger.Info("******* Wallet SDK Version:", version.VERSIONSTR, " ******* (Init)")
	return err
}

func WithEthereumNode(uri string) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.EthNode = uri
		return nil
	}
}

func WithChainID(id string) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.ChainID = id
		return nil
	}
}

func WithMinSubmit(m int) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.MinSubmit = m
		return nil
	}
}

func WithMinConfirmation(m int) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.MinConfirmation = m
		return nil
	}
}

func WithConfirmationChainLength(m int) func(c *ChainConfig) error {
	return func(c *ChainConfig) error {
		c.ConfirmationChainLength = m
		return nil
	}
}

// InitSignatureScheme initializes signature scheme only.
func InitSignatureScheme(scheme string) {
	_config.chain.SignatureScheme = scheme
}

// InitZCNSDK initializes the SDK with miner, sharder and signature scheme provided.
func InitZCNSDK(blockWorker string, signscheme string, configs ...func(*ChainConfig) error) error {
	if signscheme != "ed25519" && signscheme != "bls0chain" {
		return errors.New("", "invalid/unsupported signature scheme")
	}
	_config.chain.BlockWorker = blockWorker
	_config.chain.SignatureScheme = signscheme

	err := UpdateNetworkDetails()
	if err != nil {
		log.Println("UpdateNetworkDetails:", err)
		return err
	}

	go UpdateNetworkDetailsWorker(context.Background())

	for _, conf := range configs {
		err := conf(&_config.chain)
		if err != nil {
			return errors.Wrap(err, "invalid/unsupported options.")
		}
	}
	assertConfig()
	_config.isConfigured = true
	Logger.Info("******* Wallet SDK Version:", version.VERSIONSTR, " ******* (InitZCNSDK)")

	cfg := &conf.Config{
		BlockWorker:             _config.chain.BlockWorker,
		MinSubmit:               _config.chain.MinSubmit,
		MinConfirmation:         _config.chain.MinConfirmation,
		ConfirmationChainLength: _config.chain.ConfirmationChainLength,
		SignatureScheme:         _config.chain.SignatureScheme,
		ChainID:                 _config.chain.ChainID,
		EthereumNode:            _config.chain.EthNode,
	}

	conf.InitClientConfig(cfg)

	return nil
}

func GetNetwork() *Network {
	return &Network{
		Miners:   _config.chain.Miners,
		Sharders: _config.chain.Sharders,
	}
}

func SetNetwork(miners []string, sharders []string) {
	_config.chain.Miners = miners
	_config.chain.Sharders = sharders

	transaction.InitCache(sharders)

	conf.InitChainNetwork(&conf.Network{
		Miners:   miners,
		Sharders: sharders,
	})
}

func GetNetworkJSON() string {
	network := GetNetwork()
	networkBytes, _ := json.Marshal(network)
	return string(networkBytes)
}

// CreateWallet creates the wallet for to configure signature scheme.
// It also registers the wallet again to blockchain.
func CreateWallet(statusCb WalletCallback) error {
	if len(_config.chain.Miners) < 1 || len(_config.chain.Sharders) < 1 {
		return errors.New("", "SDK not initialized")
	}
	go func() {
		sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
		wallet, err := sigScheme.GenerateKeys()
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", err.Error())
			return
		}
		err = RegisterToMiners(wallet, statusCb)
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", err.Error())
			return
		}
	}()
	return nil
}

// RecoverOfflineWallet recovers the previously generated wallet using the mnemonic.
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
		wallet, err := sigScheme.RecoverKeys(mnemonic)
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", err.Error())
			return
		}

		err = RegisterToMiners(wallet, statusCb)
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", err.Error())
			return
		}

	}()
	return nil
}

// Split keys from the primary master key
func SplitKeys(privateKey string, numSplits int) (string, error) {
	if _config.chain.SignatureScheme != "bls0chain" {
		return "", errors.New("", "signature key doesn't support split key")
	}
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(privateKey)
	if err != nil {
		return "", errors.Wrap(err, "set private key failed")
	}
	w, err := sigScheme.SplitKeys(numSplits)
	if err != nil {
		return "", errors.Wrap(err, "split key failed.")
	}
	wStr, err := w.Marshal()
	if err != nil {
		return "", errors.Wrap(err, "wallet encoding failed.")
	}
	return wStr, nil
}

// RegisterToMiners can be used to register the wallet.
func RegisterToMiners(wallet *zcncrypto.Wallet, statusCb WalletCallback) error {
	result := make(chan *util.PostResponse)
	defer close(result)
	for _, miner := range _config.chain.Miners {
		go func(minerurl string) {
			url := minerurl + REGISTER_CLIENT
			Logger.Info(url)
			regData := map[string]string{
				"id":         wallet.ClientID,
				"public_key": wallet.ClientKey,
			}
			req, err := util.NewHTTPPostRequest(url, regData)
			if err != nil {
				Logger.Error(minerurl, "new post request failed. ", err.Error())
				return
			}
			res, err := req.Post()
			if err != nil {
				Logger.Error(minerurl, "send error. ", err.Error())
			}
			result <- res
		}(miner)
	}
	consensus := float32(0)
	for range _config.chain.Miners {
		rsp := <-result
		Logger.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode == http.StatusOK {
			consensus++
		} else {
			Logger.Debug(rsp.Body)
		}

	}
	rate := consensus * 100 / float32(len(_config.chain.Miners))
	if rate < consensusThresh {
		statusCb.OnWalletCreateComplete(StatusError, "", "rate is less than consensus")
		return fmt.Errorf("Register consensus not met. Consensus: %f, Expected: %f", rate, consensusThresh)
	}
	w, err := wallet.Marshal()
	if err != nil {
		statusCb.OnWalletCreateComplete(StatusError, w, err.Error())
		return errors.Wrap(err, "wallet encoding failed")
	}
	statusCb.OnWalletCreateComplete(StatusSuccess, w, "")
	return nil
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
		Logger.Error(minerurl, "new get request failed. ", err.Error())
		return nil, err
	}
	res, err := req.Get()
	if err != nil {
		Logger.Error(minerurl, "send error. ", err.Error())
		return nil, err
	}

	var clientDetails GetClientResponse
	err = json.Unmarshal([]byte(res.Body), &clientDetails)
	if err != nil {
		return nil, err
	}

	return &clientDetails, nil
}

// IsMnemonicValid is an utility function to check the mnemonic valid
func IsMnemonicValid(mnemonic string) bool {
	return zcncrypto.IsMnemonicValid(mnemonic)
}

// SetWalletInfo should be set before any transaction or client specific APIs
// splitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
func SetWalletInfo(w string, splitKeyWallet bool) error {
	err := json.Unmarshal([]byte(w), &_config.wallet)
	if err == nil {
		if _config.chain.SignatureScheme == "bls0chain" {
			_config.isSplitWallet = splitKeyWallet
		}
		_config.isValidWallet = true
	}
	return err
}

// SetAuthUrl will be called by app to set zauth URL to SDK.
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

// GetBalance retreives wallet balance from sharders
func GetBalance(cb GetBalanceCallback) error {
	err := CheckConfig()
	if err != nil {
		return err
	}
	go func() {
		value, info, err := getBalanceFromSharders(_config.wallet.ClientID)
		if err != nil {
			Logger.Error(err)
			cb.OnBalanceAvailable(StatusError, 0, info)
			return
		}
		cb.OnBalanceAvailable(StatusSuccess, value, info)
	}()
	return nil
}

// GetBalance retreives wallet nonce from sharders
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
			Logger.Error(err)
			cb.OnNonceAvailable(StatusError, 0, info)
			return
		}
		cb.OnNonceAvailable(StatusSuccess, value, info)
	}()
	return nil
}

// GetBalance retreives wallet balance from sharders
func GetBalanceWallet(walletStr string, cb GetBalanceCallback) error {

	w, err := GetWallet(walletStr)
	if err != nil {
		fmt.Printf("Error while parsing the wallet. %v\n", err)
		return err
	}

	go func() {
		value, info, err := getBalanceFromSharders(w.ClientID)
		if err != nil {
			Logger.Error(err)
			cb.OnBalanceAvailable(StatusError, 0, info)
			return
		}
		cb.OnBalanceAvailable(StatusSuccess, value, info)
	}()
	return nil
}

func getBalanceFromSharders(clientID string) (int64, string, error) {
	return getBalanceFieldFromSharders(clientID, "balance")
}

func getNonceFromSharders(clientID string) (int64, string, error) {
	return getBalanceFieldFromSharders(clientID, "nonce")
}

func getBalanceFieldFromSharders(clientID, name string) (int64, string, error) {
	result := make(chan *util.GetResponse)
	defer close(result)
	// getMinShardersVerify
	var numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromSharders(numSharders, fmt.Sprintf("%v%v", GET_BALANCE, clientID), result)
	consensus := float32(0)
	balMap := make(map[int64]float32)
	winBalance := int64(0)
	var winInfo string
	var winError string
	for i := 0; i < numSharders; i++ {
		rsp := <-result
		Logger.Debug(rsp.Url, rsp.Status)
		if rsp.StatusCode != http.StatusOK {
			Logger.Error(rsp.Body)
			winError = rsp.Body
			continue
		}
		Logger.Debug(rsp.Body)
		var objmap map[string]json.RawMessage
		err := json.Unmarshal([]byte(rsp.Body), &objmap)
		if err != nil {
			continue
		}
		if v, ok := objmap[name]; ok {
			bal, err := strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				continue
			}
			balMap[bal]++
			if balMap[bal] > consensus {
				consensus = balMap[bal]
				winBalance = bal
				winInfo = rsp.Body
			}
		}

	}
	rate := consensus * 100 / float32(len(_config.chain.Sharders))
	if rate < consensusThresh {
		return 0, winError, errors.New("", "get balance failed. consensus not reached")
	}
	return winBalance, winInfo, nil
}

// ConvertToToken converts the value to ZCN tokens
func ConvertToToken(value int64) float64 {
	return float64(value) / float64(TOKEN_UNIT)
}

// ConvertToValue converts ZCN tokens to value
func ConvertToValue(token float64) int64 {
	return int64(token * float64(TOKEN_UNIT))
}

func ConvertTokenToUSD(token float64) (float64, error) {
	zcnRate, err := getTokenUSDRate()
	if err != nil {
		return 0, err
	}
	return token * zcnRate, nil
}

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

func getInfoFromSharders(urlSuffix string, op int, cb GetInfoCallback) {

	tq, err := NewTransactionQuery(util.Shuffle(_config.chain.Sharders))
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	qr, err := tq.GetInfo(context.TODO(), urlSuffix)
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	cb.OnInfoAvailable(op, StatusSuccess, string(qr.Content), "")
}

//GetWallet get a wallet object from a wallet string
func GetWallet(walletStr string) (*zcncrypto.Wallet, error) {

	var w zcncrypto.Wallet

	err := json.Unmarshal([]byte(walletStr), &w)

	if err != nil {
		fmt.Printf("error while parsing wallet string.\n%v\n", err)
		return nil, err
	}

	return &w, nil

}

//GetWalletClientID -- given a walletstr return ClientID
func GetWalletClientID(walletStr string) (string, error) {
	w, err := GetWallet(walletStr)
	if err != nil {
		return "", err
	}
	return w.ClientID, nil
}

// GetZcnUSDInfo returns USD value for ZCN token from coinmarketcap.com
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
			Logger.Error("new post request failed. ", err.Error())
			return
		}
		res, err := req.Post()
		if err != nil {
			Logger.Error(authHost+"send error. ", err.Error())
		}
		if res.StatusCode != http.StatusOK {
			cb.OnSetupComplete(StatusError, res.Body)
			return
		}
		cb.OnSetupComplete(StatusSuccess, "")
	}()
	return nil
}

func GetIdForUrl(url string) string {
	url = strings.TrimRight(url, "/")
	url = fmt.Sprintf("%v/_nh/whoami", url)
	req, err := util.NewHTTPGetRequest(url)
	if err != nil {
		Logger.Error(url, "new get request failed. ", err.Error())
		return ""
	}
	res, err := req.Get()
	if err != nil {
		Logger.Error(url, "get error. ", err.Error())
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

type VestingDestInfo struct {
	ID     common.Key       `json:"id"`     // identifier
	Wanted common.Balance   `json:"wanted"` // wanted amount for entire period
	Earned common.Balance   `json:"earned"` // can unlock
	Vested common.Balance   `json:"vested"` // already vested
	Last   common.Timestamp `json:"last"`   // last time unlocked
}

type VestingPoolInfo struct {
	ID           common.Key         `json:"pool_id"`      // pool ID
	Balance      common.Balance     `json:"balance"`      // real pool balance
	Left         common.Balance     `json:"left"`         // owner can unlock
	Description  string             `json:"description"`  // description
	StartTime    common.Timestamp   `json:"start_time"`   // from
	ExpireAt     common.Timestamp   `json:"expire_at"`    // until
	Destinations []*VestingDestInfo `json:"destinations"` // receivers
	ClientID     common.Key         `json:"client_id"`    // owner
}

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

func WithParams(uri string, params Params) string {
	return uri + params.Query()
}

func GetVestingPoolInfo(poolID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	getInfoFromSharders(WithParams(GET_VESTING_POOL_INFO, Params{
		"pool_id": poolID,
	}), 0, cb)
	return
}

type VestingClientList struct {
	Pools []common.Key `json:"pools"`
}

func GetVestingClientList(clientID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	if clientID == "" {
		clientID = _config.wallet.ClientID // if not blank
	}
	go getInfoFromSharders(WithParams(GET_VESTING_CLIENT_POOLS, Params{
		"client_id": clientID,
	}), 0, cb)
	return
}

type VestingSCConfig struct {
	MinLock              common.Balance `json:"min_lock"`
	MinDuration          time.Duration  `json:"min_duration"`
	MaxDuration          time.Duration  `json:"max_duration"`
	MaxDestinations      int            `json:"max_destinations"`
	MaxDescriptionLength int            `json:"max_description_length"`
}

type InputMap struct {
	Fields map[string]string `json:"fields"`
}

func GetVestingSCConfig(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go getInfoFromSharders(GET_VESTING_CONFIG, 0, cb)
	return
}

// faucet

func GetFaucetSCConfig(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go getInfoFromSharders(GET_FAUCETSC_CONFIG, 0, cb)
	return
}

//
// miner SC
//

type Miner struct {
	ID         string      `json:"id"`
	N2NHost    string      `json:"n2n_host"`
	Host       string      `json:"host"`
	Port       int         `json:"port"`
	PublicKey  string      `json:"public_key"`
	ShortName  string      `json:"short_name"`
	BuildTag   string      `json:"build_tag"`
	TotalStake int64       `json:"total_stake"`
	Stat       interface{} `json:"stat"`
}

type DelegatePool struct {
	Balance      int64  `json:"balance"`
	Reward       int64  `json:"reward"`
	Status       int    `json:"status"`
	RoundCreated int64  `json:"round_created"` // used for cool down
	DelegateID   string `json:"delegate_id"`
}

type StakePool struct {
	Pools    map[string]*DelegatePool `json:"pools"`
	Reward   int64                    `json:"rewards"`
	Settings StakePoolSettings        `json:"settings"`
	Minter   int                      `json:"minter"`
}

type Node struct {
	Miner     Miner `json:"simple_miner"`
	StakePool `json:"stake_pool"`
}

type MinerSCNodes struct {
	Nodes []Node `json:"Nodes"`
}

// GetMiners obtains list of all active miners.
func GetMiners(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = GET_MINERSC_MINERS
	go getInfoFromSharders(url, 0, cb)
	return
}

// GetSharders obtains list of all active sharders.
func GetSharders(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = GET_MINERSC_SHARDERS
	go getInfoFromSharders(url, 0, cb)
	return
}

func GetEvents(cb GetInfoCallback, filters map[string]string) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go getInfoFromSharders(WithParams(GET_MINERSC_EVENTS, Params{
		"block_number": filters["block_number"],
		"tx_hash":      filters["tx_hash"],
		"type":         filters["type"],
		"tag":          filters["tag"],
	}), 0, cb)
	return
}

func GetMinerSCNodeInfo(id string, cb GetInfoCallback) (err error) {

	if err = CheckConfig(); err != nil {
		return
	}

	go getInfoFromSharders(WithParams(GET_MINERSC_NODE, Params{
		"id": id,
	}), 0, cb)
	return
}

func GetMinerSCNodePool(id, poolID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go getInfoFromSharders(WithParams(GET_MINERSC_POOL, Params{
		"id":      id,
		"pool_id": poolID,
	}), 0, cb)

	return
}

type MinerSCDelegatePoolInfo struct {
	ID         common.Key     `json:"id"`
	Balance    common.Balance `json:"balance"`
	Reward     common.Balance `json:"reward"`      // uncollected reread
	RewardPaid common.Balance `json:"reward_paid"` // total reward all time
	Status     string         `json:"status"`
}

type MinerSCUserPoolsInfo struct {
	Pools map[string][]*MinerSCDelegatePoolInfo `json:"pools"`
}

func GetMinerSCUserInfo(clientID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	if clientID == "" {
		clientID = _config.wallet.ClientID
	}
	go getInfoFromSharders(WithParams(GET_MINERSC_USER, Params{
		"client_id": clientID,
	}), 0, cb)

	return
}

func GetMinerSCConfig(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go getInfoFromSharders(GET_MINERSC_CONFIG, 0, cb)
	return
}

func GetMinerSCGlobals(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go getInfoFromSharders(GET_MINERSC_GLOBALS, 0, cb)
	return
}

//
// Storage SC
//

// GetStorageSCConfig obtains Storage SC configurations.
func GetStorageSCConfig(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go getInfoFromSharders(STORAGESC_GET_SC_CONFIG, OpStorageSCGetConfig, cb)
	return
}

// GetChallengePoolInfo obtains challenge pool information for an allocation.
func GetChallengePoolInfo(allocID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = WithParams(STORAGESC_GET_CHALLENGE_POOL_INFO, Params{
		"allocation_id": allocID,
	})
	go getInfoFromSharders(url, OpStorageSCGetChallengePoolInfo, cb)
	return
}

// GetAllocation obtains allocation information.
func GetAllocation(allocID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = WithParams(STORAGESC_GET_ALLOCATION, Params{
		"allocation": allocID,
	})
	go getInfoFromSharders(url, OpStorageSCGetAllocation, cb)
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
	var url = WithParams(STORAGESC_GET_ALLOCATIONS, Params{
		"client": clientID,
	})
	go getInfoFromSharders(url, OpStorageSCGetAllocations, cb)
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
	var url = WithParams(STORAGESC_GET_READ_POOL_INFO, Params{
		"client_id": clientID,
	})
	go getInfoFromSharders(url, OpStorageSCGetReadPoolInfo, cb)
	return
}

// GetStakePoolInfo obtains information about stake pool of a blobber and
// related validator.
func GetStakePoolInfo(blobberID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = WithParams(STORAGESC_GET_STAKE_POOL_INFO, Params{
		"blobber_id": blobberID,
	})
	go getInfoFromSharders(url, OpStorageSCGetStakePoolInfo, cb)
	return
}

// GetStakePoolUserInfo for a user.
func GetStakePoolUserInfo(clientID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	if clientID == "" {
		clientID = _config.wallet.ClientID
	}
	var url = WithParams(STORAGESC_GET_STAKE_POOL_USER_INFO, Params{
		"client_id": clientID,
	})
	go getInfoFromSharders(url, OpStorageSCGetStakePoolInfo, cb)
	return
}

// GetBlobbers obtains list of all active blobbers.
func GetBlobbers(cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = STORAGESC_GET_BLOBBERS

	go getInfoFromSharders(url, OpStorageSCGetBlobbers, cb)
	return
}

// GetBlobber obtains blobber information.
func GetBlobber(blobberID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	var url = WithParams(STORAGESC_GET_BLOBBER, Params{
		"blobber_id": blobberID,
	})
	go getInfoFromSharders(url, OpStorageSCGetBlobber, cb)
	return
}

// GetWritePoolInfo obtains information about all write pools of a user.
// If given clientID is empty, then current user used.
func GetWritePoolInfo(clientID string, cb GetInfoCallback) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	if clientID == "" {
		clientID = _config.wallet.ClientID
	}
	var url = WithParams(STORAGESC_GET_WRITE_POOL_INFO, Params{
		"client_id": clientID,
	})
	go getInfoFromSharders(url, OpStorageSCGetWritePoolInfo, cb)
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

func Decrypt(key, text string) (string, error) {
	keyBytes := []byte(key)
	textBytes, _ := hex.DecodeString(text)
	response, err := zboxutil.Decrypt(keyBytes, textBytes)
	if err != nil {
		return "", err
	}
	return string(response), nil
}

type NonceCache struct {
	cache map[string]int64
	guard sync.Mutex
}

func NewNonceCache() *NonceCache {
	return &NonceCache{cache: make(map[string]int64)}
}

func (nc *NonceCache) GetNextNonce(clientId string) int64 {
	nc.guard.Lock()
	defer nc.guard.Unlock()
	if _, ok := nc.cache[clientId]; !ok {
		back := &getNonceCallBack{
			nonceCh: make(chan int64),
			err:     nil,
		}
		if err := GetNonce(back); err != nil {
			return 0
		}

		timeout, _ := context.WithTimeout(context.Background(), time.Second)
		select {
		case n := <-back.nonceCh:
			if back.err != nil {
				return 0
			}
			nc.cache[clientId] = n
		case <-timeout.Done():
			return 0
		}
	}

	nc.cache[clientId] += 1
	return nc.cache[clientId]
}

func (nc *NonceCache) Set(clientId string, nonce int64) {
	nc.guard.Lock()
	defer nc.guard.Unlock()
	nc.cache[clientId] = nonce
}

func (nc *NonceCache) Evict(clientId string) {
	nc.guard.Lock()
	defer nc.guard.Unlock()
	delete(nc.cache, clientId)
}
