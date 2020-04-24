package zcncore

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/core/zcncrypto"
)

type ChainConfig struct {
	ChainID                 string   `json:"chain_id,omitempty"`
	Miners                  []string `json:"miners"`
	Sharders                []string `json:"sharders"`
	SignatureScheme         string   `json:"signaturescheme"`
	MinSubmit               int      `json:"min_submit"`
	MinConfirmation         int      `json:"min_confirmation"`
	ConfirmationChainLength int      `json:"confirmation_chain_length"`
}

var defaultLogLevel = logger.DEBUG
var Logger logger.Logger

const (
	REGISTER_CLIENT          = `/v1/client/put`
	PUT_TRANSACTION          = `/v1/transaction/put`
	TXN_VERIFY_URL           = `/v1/transaction/get/confirmation?hash=`
	GET_BALANCE              = `/v1/client/get/balance?client_id=`
	GET_LOCK_CONFIG          = `/v1/screst/` + InterestPoolSmartContractAddress + `/getLockConfig`
	GET_LOCKED_TOKENS        = `/v1/screst/` + InterestPoolSmartContractAddress + `/getPoolsStats?client_id=`
	GET_BLOCK_INFO           = `/v1/block/get?`
	GET_LATEST_FINALIZED     = `/v1/block/get/latest_finalized`
	GET_CHAIN_STATS          = `/v1/chain/get/stats`
	GET_USER_POOLS           = `/v1/screst/` + MinerSmartContractAddress + `/getUserPools?client_id=`
	GET_USER_POOL_DETAIL     = `/v1/screst/` + MinerSmartContractAddress + `/getPoolsStats?`
	GET_VESTING_CONFIG       = `/v1/screst/` + VestingSmartContractAddress + `/getConfig`
	GET_VESTING_POOL_INFO    = `/v1/screst/` + VestingSmartContractAddress + `/getPoolInfo`
	GET_VESTING_CLIENT_POOLS = `/v1/screst/` + VestingSmartContractAddress + `/getClientPools`

	// TORM (sfxdx): remove from zwallet
	GET_BLOBBERS            = `/v1/screst/` + StorageSmartContractAddress + `/getblobbers`
	GET_READ_POOL_STATS     = `/v1/screst/` + StorageSmartContractAddress + `/getReadPoolsStats?client_id=`
	GET_WRITE_POOL_STAT     = `/v1/screst/` + StorageSmartContractAddress + `/getWritePoolStat?allocation_id=`
	GET_STAKE_POOL_STAT     = `/v1/screst/` + StorageSmartContractAddress + `/getStakePoolStat?blobber_id=`
	GET_CHALLENGE_POOL_STAT = `/v1/screst/` + StorageSmartContractAddress + `/getChallengePoolStat?allocation_id=`
	GET_STORAGE_SC_CONFIG   = `/v1/screst/` + StorageSmartContractAddress + `/getConfig`
)

const (
	// TORM (sfxdx) remove from zwallet
	StorageSmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7`

	VestingSmartContractAddress      = `2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead`
	FaucetSmartContractAddress       = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d3`
	InterestPoolSmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9`
	MultiSigSmartContractAddress     = `27b5ef7120252b79f9dd9c05505dd28f328c80f6863ee446daede08a84d651a7`
	MinerSmartContractAddress        = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d1`
	MultiSigRegisterFuncName         = "register"
	MultiSigVoteFuncName             = "vote"
)

// In percentage
const consensusThresh = float32(25.0)

const (
	defaultMinSubmit               = int(50)
	defaultMinConfirmation         = int(50)
	defaultConfirmationChainLength = int(3)
	defaultTxnExpirationSeconds    = 60
	defaultWaitSeconds             = (3 * time.Second)
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

const TOKEN_UNIT = int64(10000000000)

const (
	OpGetTokenLockConfig int = iota
	OpGetLockedTokens
	OpGetUserPools
	OpGetUserPoolDetail
	OpGetBlobbers
	OpGetReadPoolsStats
	OpGetWritePoolStat
	OpGetStakePoolStat
	OpGetChallengePoolStat
	OpGetStorageSCConfig
)

// WalletCallback needs to be implmented for wallet creation.
type WalletCallback interface {
	OnWalletCreateComplete(status int, wallet string, err string)
}

// GetBalanceCallback needs to be implemented by the caller of GetBalance() to get the status
type GetBalanceCallback interface {
	OnBalanceAvailable(status int, value int64, info string)
}

// GetInfoCallback needs to be implemented by the caller of GetLockTokenConfig() and GetLockedTokens()
type GetInfoCallback interface {
	// OnInfoAvailable will be called when GetLockTokenConfig is complete
	// if status == StatusSuccess then info is valid
	// is status != StatusSuccess then err will give the reason
	OnInfoAvailable(op int, status int, info string, err string)
}

// GetUSDInfoCallback needs to be implemented by the caller of GetZcnUSDInfo()
type GetUSDInfoCallback interface {
	// This will be called when GetZcnUSDInfo completes.
	// if status == StatusSuccess then info is valid
	// is status != StatusSuccess then err will give the reason
	OnUSDInfoAvailable(status int, info string, err string)
}

// AuthCallback needs to be implemented by the caller SetupAuth()
type AuthCallback interface {
	// This call back gives the status of the Two factor authenticator(zauth) setup.
	OnSetupComplete(status int, err string)
}

type regInfo struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
}

type httpResponse struct {
	status string
	body   []byte
	err    error
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
		return fmt.Errorf("SDK not initialized")
	}
	return nil
}
func checkWalletConfig() error {
	if !_config.isValidWallet || _config.wallet.ClientID == "" {
		Logger.Error("wallet info not found. returning error.")
		return fmt.Errorf("wallet info not found. set wallet info.")
	}
	return nil
}
func checkConfig() error {
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
	Logger.Info("******* Wallet SDK Version:", version.VERSIONSTR, " *******")
}

// CloseLog closes log file
func CloseLog() {
	Logger.Close()
}

// Init inializes the SDK with miner, sharder and signature scheme provided in
// configuration provided in JSON format
func Init(c string) error {
	err := json.Unmarshal([]byte(c), &_config.chain)
	if err == nil {
		// Check signature scheme is supported
		if _config.chain.SignatureScheme != "ed25519" && _config.chain.SignatureScheme != "bls0chain" {
			return fmt.Errorf("invalid/unsupported signature scheme")
		}
		assertConfig()
		_config.isConfigured = true
	}
	Logger.Info("*******  Wallet SDK Version:", version.VERSIONSTR, " *******")
	return err
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

// InitZCNSDK initializes the SDK with miner, sharder and signature scheme provided.
func InitZCNSDK(miners []string, sharders []string, signscheme string, configs ...func(*ChainConfig) error) error {
	if signscheme != "ed25519" && signscheme != "bls0chain" {
		return fmt.Errorf("invalid/unsupported signature scheme")
	}
	_config.chain.Miners = miners
	_config.chain.Sharders = sharders
	_config.chain.SignatureScheme = signscheme
	for _, conf := range configs {
		err := conf(&_config.chain)
		if err != nil {
			return fmt.Errorf("invalid/unsupported options. %s", err)
		}
	}
	assertConfig()
	_config.isConfigured = true
	Logger.Info("*******  Wallet SDK Version:", version.VERSIONSTR, " *******")
	return nil
}

// CreateWallet creates the a wallet for the configure signature scheme.
// It also registers the wallet again to block chain.
func CreateWallet(statusCb WalletCallback) error {
	if len(_config.chain.Miners) < 1 || len(_config.chain.Sharders) < 1 {
		return fmt.Errorf("SDK not initialized")
	}
	go func() {
		sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
		wallet, err := sigScheme.GenerateKeys()
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", fmt.Sprintf("%s", err.Error()))
			return
		}
		err = RegisterToMiners(wallet, statusCb)
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", fmt.Sprintf("%s", err.Error()))
			return
		}
	}()
	return nil
}

// RecoverWallet recovers the previously generated wallet using the mnemonic.
// It also registers the wallet again to block chain.
func RecoverWallet(mnemonic string, statusCb WalletCallback) error {
	if zcncrypto.IsMnemonicValid(mnemonic) != true {
		return fmt.Errorf("Invalid mnemonic")
	}
	go func() {
		sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
		wallet, err := sigScheme.RecoverKeys(mnemonic)
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", fmt.Sprintf("%s", err.Error()))
			return
		}
		err = RegisterToMiners(wallet, statusCb)
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", fmt.Sprintf("%s", err.Error()))
			return
		}
	}()
	return nil
}

// Split keys from the primary master key
func SplitKeys(privateKey string, numSplits int) (string, error) {
	if _config.chain.SignatureScheme != "bls0chain" {
		return "", fmt.Errorf("signature key doesn't support split key")
	}
	sigScheme := zcncrypto.NewBLS0ChainScheme()
	err := sigScheme.SetPrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("set private key failed - %s", err.Error())
	}
	w, err := sigScheme.SplitKeys(numSplits)
	if err != nil {
		return "", fmt.Errorf("split key failed. %s", err.Error())
	}
	wStr, err := w.Marshal()
	if err != nil {
		return "", fmt.Errorf("wallet encoding failed. %s", err.Error())
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
			return
		}(miner)
	}
	consensus := float32(0)
	for range _config.chain.Miners {
		select {
		case rsp := <-result:
			Logger.Debug(rsp.Url, rsp.Status)

			if rsp.StatusCode == http.StatusOK {
				consensus++
			} else {
				Logger.Debug(rsp.Body)
			}
		}
	}
	rate := consensus * 100 / float32(len(_config.chain.Miners))
	if rate < consensusThresh {
		return fmt.Errorf("Register consensus not met. Consensus: %f, Expected: %f", rate, consensusThresh)
	}
	w, err := wallet.Marshal()
	if err != nil {
		return fmt.Errorf("wallet encoding failed - %s", err.Error())
	}
	time.Sleep(3 * time.Second)
	statusCb.OnWalletCreateComplete(StatusSuccess, w, "")
	return nil
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
		return fmt.Errorf("wallet type is not split key")
	}
	if url == "" {
		return fmt.Errorf("invalid auth url")
	}
	_config.authUrl = strings.TrimRight(url, "/")
	return nil
}

// GetBalance retreives wallet balance from sharders
func GetBalance(cb GetBalanceCallback) error {
	err := checkConfig()
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
	result := make(chan *util.GetResponse)
	defer close(result)
	queryFromSharders(getMinShardersVerify(), fmt.Sprintf("%v%v", GET_BALANCE, clientID), result)
	consensus := float32(0)
	balMap := make(map[int64]float32)
	winBalance := int64(0)
	var winInfo string
	var winError string
	for i := 0; i < getMinShardersVerify(); i++ {
		select {
		case rsp := <-result:
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
			if v, ok := objmap["balance"]; ok {
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
	}
	rate := consensus * 100 / float32(len(_config.chain.Sharders))
	if rate < consensusThresh {
		return 0, winError, fmt.Errorf("get balance failed. consensus not reached")
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

func getInfoFromSharders(urlSuffix string, op int, cb GetInfoCallback) {
	result := make(chan *util.GetResponse)
	defer close(result)
	queryFromSharders(getMinShardersVerify(), urlSuffix, result)
	consensus := float32(0)
	resultMap := make(map[int]float32)
	var winresult *util.GetResponse
	for i := 0; i < getMinShardersVerify(); i++ {
		select {
		case rsp := <-result:
			Logger.Debug(rsp.Url, rsp.Status)
			resultMap[rsp.StatusCode]++
			if resultMap[rsp.StatusCode] > consensus {
				consensus = resultMap[rsp.StatusCode]
				winresult = rsp
			}
		}
	}
	rate := consensus * 100 / float32(len(_config.chain.Sharders))
	if rate < consensusThresh {
		newerr := fmt.Sprintf(`{"code": "consensus_failed", "error": "consensus failed on sharders.", "server_error": "%v"}`, winresult.Body)
		cb.OnInfoAvailable(op, StatusError, "", newerr)
		return
	}
	if winresult.StatusCode != http.StatusOK {
		cb.OnInfoAvailable(op, StatusError, "", winresult.Body)
	} else {
		cb.OnInfoAvailable(op, StatusSuccess, winresult.Body, "")
	}
}

// GetLockConfig returns the lock token configuration information such as interest rate from blockchain
func GetLockConfig(cb GetInfoCallback) error {
	err := checkSdkInit()
	if err != nil {
		return err
	}
	go func() {
		// urlSuffix := fmt.Sprintf("%v%v&key=%v", GET_LOCK_CONFIG,
		// 	InterestPoolSmartContractAddress, InterestPoolSmartContractAddress)
		urlSuffix := GET_LOCK_CONFIG
		getInfoFromSharders(urlSuffix, OpGetTokenLockConfig, cb)
	}()
	return nil
}

// GetLockedTokens returns the ealier locked token pool stats
func GetLockedTokens(cb GetInfoCallback) error {
	err := checkConfig()
	if err != nil {
		return err
	}
	go func() {
		urlSuffix := fmt.Sprintf("%v%v", GET_LOCKED_TOKENS, _config.wallet.ClientID)
		getInfoFromSharders(urlSuffix, OpGetLockedTokens, cb)
	}()
	return nil
}

// read pool stats

// GetReadPoolsStats returns statistic of locked tokens in read pool.
func GetReadPoolsStats(cb GetInfoCallback) (err error) {
	if err = checkConfig(); err != nil {
		return
	}
	go func() {
		var url = fmt.Sprintf("%v%v", GET_READ_POOL_STATS,
			_config.wallet.ClientID)
		getInfoFromSharders(url, OpGetReadPoolsStats, cb)
	}()
	return
}

// write pool stat

// GetWritePoolStat returns statistic of locked tokens in a write pool.
func GetWritePoolStat(cb GetInfoCallback, allocID string) (err error) {
	if err = checkConfig(); err != nil {
		return
	}
	go func() {
		var url = fmt.Sprintf("%v%v", GET_WRITE_POOL_STAT, allocID)
		getInfoFromSharders(url, OpGetWritePoolStat, cb)
	}()
	return
}

// stake pool stat

// GetStakePoolStat returns statistic of locked tokens in a stake pool.
func GetStakePoolStat(cb GetInfoCallback, blobberID string) (err error) {
	if err = checkConfig(); err != nil {
		return
	}
	go func() {
		var url = fmt.Sprintf("%v%v", GET_STAKE_POOL_STAT, blobberID)
		getInfoFromSharders(url, OpGetStakePoolStat, cb)
	}()
	return
}

// GetChallengePoolStat returns statistic of tokens in a challenge pool.
func GetChallengePoolStat(cb GetInfoCallback, allocID string) (err error) {
	if err = checkConfig(); err != nil {
		return
	}
	go func() {
		var url = fmt.Sprintf("%v%v", GET_CHALLENGE_POOL_STAT, allocID)
		getInfoFromSharders(url, OpGetChallengePoolStat, cb)
	}()
	return
}

// storage SC configurations

// GetStorageSCConfig returns current configurations of storage SC.
func GetStorageSCConfig(cb GetInfoCallback) (err error) {
	if err = checkConfig(); err != nil {
		return
	}
	go func() {
		getInfoFromSharders(GET_STORAGE_SC_CONFIG, OpGetStorageSCConfig, cb)
	}()
	return
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
func GetZcnUSDInfo(cb GetUSDInfoCallback) error {
	go func() {
		req, err := util.NewHTTPGetRequest("https://api.coingecko.com/api/v3/coins/0chain?localization=false")
		if err != nil {
			Logger.Error("new get request failed." + err.Error())
			cb.OnUSDInfoAvailable(StatusError, "", "new get request failed."+err.Error())
			return
		}
		res, err := req.Get()
		if err != nil {
			Logger.Error("get error. ", err.Error())
			cb.OnUSDInfoAvailable(StatusError, "", "get error"+err.Error())
			return
		}
		if res.StatusCode != http.StatusOK {
			cb.OnUSDInfoAvailable(StatusError, "", fmt.Sprintf("%s: %s", res.Status, res.Body))
			return
		}
		cb.OnUSDInfoAvailable(StatusSuccess, res.Body, "")
	}()
	return nil
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

func GetUserPools(cb GetInfoCallback) error {
	err := checkConfig()
	if err != nil {
		return err
	}
	go func() {
		urlSuffix := fmt.Sprintf("%v%v", GET_USER_POOLS, _config.wallet.ClientID)
		getInfoFromSharders(urlSuffix, OpGetUserPools, cb)
	}()
	return nil
}

func GetUserPoolDetails(clientID, poolID string, cb GetInfoCallback) error {
	err := checkConfig()
	if err != nil {
		return err
	}
	go func() {
		urlSuffix := fmt.Sprintf("%vminer_id=%v&pool_id=%v", GET_USER_POOL_DETAIL, clientID, poolID)
		getInfoFromSharders(urlSuffix, OpGetUserPoolDetail, cb)
	}()
	return nil
}

func GetBlobbers(cb GetInfoCallback) error {
	err := checkSdkInit()
	if err != nil {
		return err
	}
	go func() {
		urlSuffix := GET_BLOBBERS
		getInfoFromSharders(urlSuffix, OpGetBlobbers, cb)
	}()
	return nil
}

// on json info available

type OnJSONInfoCb struct {
	value interface{}
	err   error
	got   chan struct{}
}

func (ojsonic *OnJSONInfoCb) OnInfoAvailable(op int, status int,
	info string, errMsg string) {

	defer close(ojsonic.got)

	if status != StatusSuccess {
		ojsonic.err = errors.New(errMsg)
		return
	}
	var err error
	if err = json.Unmarshal([]byte(info), ojsonic.value); err != nil {
		ojsonic.err = fmt.Errorf("decoding response: %v", err)
	}
}

// Wait for info.
func (ojsonic *OnJSONInfoCb) Wait() (err error) {
	<-ojsonic.got
	return ojsonic.err
}

func NewJSONInfoCB(val interface{}) (cb *OnJSONInfoCb) {
	cb = new(OnJSONInfoCb)
	cb.value = val
	cb.got = make(chan struct{})
	return
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

func withParams(uri string, params url.Values) string {
	return uri + "?" + params.Encode()
}

func GetVestingPoolInfo(poolID common.Key) (vpi *VestingPoolInfo, err error) {
	if err = checkSdkInit(); err != nil {
		return
	}
	vpi = new(VestingPoolInfo)
	var cb = NewJSONInfoCB(vpi)
	go getInfoFromSharders(withParams(GET_VESTING_POOL_INFO, url.Values{
		"pool_id": []string{string(poolID)},
	}), 0, cb)
	err = cb.Wait()
	return
}

type VestingClientList struct {
	Pools []common.Key `json:"pools"`
}

func GetVestingClientList(clientID common.Key) (
	vcl *VestingClientList, err error) {

	if err = checkSdkInit(); err != nil {
		return
	}
	if clientID == "" {
		clientID = common.Key(_config.wallet.ClientID) // if not blank
	}

	vcl = new(VestingClientList)
	var cb = NewJSONInfoCB(vcl)
	go getInfoFromSharders(withParams(GET_VESTING_CLIENT_POOLS, url.Values{
		"client_id": []string{string(clientID)},
	}), 0, cb)
	err = cb.Wait()
	return
}

type VestingSCConfig struct {
	MinLock              common.Balance `json:"min_lock"`
	MinDuration          time.Duration  `json:"min_duration"`
	MaxDuration          time.Duration  `json:"max_duration"`
	MaxDestinations      int            `json:"max_destinations"`
	MaxDescriptionLength int            `json:"max_description_length"`
}

func (vscc *VestingSCConfig) IsZero() bool {
	return (*vscc) == (VestingSCConfig{})
}

func GetVestingSCConfig() (vscc *VestingSCConfig, err error) {

	if err = checkSdkInit(); err != nil {
		return
	}
	vscc = new(VestingSCConfig)
	var cb = NewJSONInfoCB(vscc)
	go getInfoFromSharders(GET_VESTING_CONFIG, 0, cb)
	if err = cb.Wait(); err != nil {
		return
	}
	if vscc.IsZero() {
		return nil, errors.New("empty response from sharders")
	}
	return
}
