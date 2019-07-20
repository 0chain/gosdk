package zcncore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/core/zcncrypto"
)

type ChainConfig struct {
	ChainID         string   `json:"chain_id,omitempty"`
	Miners          []string `json:"miners"`
	Sharders        []string `json:"sharders"`
	SignatureScheme string   `json:"signaturescheme"`
	MinConfirmation         int      `json:"min_confirmation"`
	ConfirmationChainLength int      `json:"confirmation_chain_length"`
}

var defaultLogLevel = logger.DEBUG
var Logger logger.Logger

var REGISTER_CLIENT = `/v1/client/put`
var PUT_TRANSACTION = `/v1/transaction/put`
var TXN_VERIFY_URL = `/v1/transaction/get/confirmation?hash=`
var GET_BALANCE = `/v1/client/get/balance?client_id=`
var GET_LOCK_CONFIG = `/v1/scstate/get?sc_address=`
var GET_LOCKED_TOKENS = `/v1/screst/` + InterestPoolSmartContractAddress + `/getPoolsStats?client_id=`
var GET_BLOCK_INFO = `/v1/block/get?`
var GET_USER_POOLS = `/v1/screst/` + StakeSmartContractAddress + `/getUserPools?client_id=`
var GET_USER_POOL_DETAIL = `/v1/screst/` + StakeSmartContractAddress + `/getPoolsStats?`

const StorageSmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7`
const FaucetSmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d3`
const InterestPoolSmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9`
const StakeSmartContractAddress = `CF9C03CD22C9C7B116EED04E4A909F95ABEC17E98FE631D6AC94D5D8420C5B20`

// In percentage
const consensusThresh = float32(25.0)

const defaultMinConfirmation = int(25)
const defaultConfirmationChainLength = int(5)
const defaultTxnExpirationSeconds = 15
const defaultWaitSeconds = (3 * time.Second)
const (
	StatusSuccess      int = 0
	StatusNetworkError int = 1
	// TODO: Change to specific error
	StatusError   int = 2
	StatusRejectedByUser   int = 3
	StatusInvalidSignature int = 4
	StatusAuthError        int = 5
	StatusAuthVerifyFailed int = 6
	StatusAuthTimeout      int = 7
	StatusUnknown int = -1
)

const TOKEN_UNIT = int64(10000000000)

const (
	OpGetTokenLockConfig int = 0
	OpGetLockedTokens    int = 1
	OpGetUserPools       int = 2
	OpGetUserPoolDetail  int = 3
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
	if _config.chain.MinConfirmation <= 0 {
		_config.chain.MinConfirmation = defaultMinConfirmation
	}
	if _config.chain.ConfirmationChainLength <= 0 {
		_config.chain.ConfirmationChainLength = defaultConfirmationChainLength
	}
}
func getMinMinersSubmit() int {
	return util.MaxInt((_config.chain.MinConfirmation * len(_config.chain.Miners) / 100), 1)
}
func getMinShardersVerify() int {
	return util.MaxInt((_config.chain.MinConfirmation * len(_config.chain.Sharders) / 100), 1)
}
func getMinRequiredChainLength() int64 {
	return int64(_config.chain.ConfirmationChainLength)
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

// InitZCNSDK initializes the SDK with miner, sharder and signature scheme provided.
func InitZCNSDK(miners []string, sharders []string, signscheme string) error {
	_config.chain.Miners = miners
	_config.chain.Sharders = sharders
	_config.chain.SignatureScheme = signscheme
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
			regData := map[string]string{"id": wallet.ClientID, "public_key": wallet.ClientKey}
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
		urlSuffix := fmt.Sprintf("%v%v&key=%v", GET_LOCK_CONFIG,
			InterestPoolSmartContractAddress, InterestPoolSmartContractAddress)
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
// GetZcnUSDInfo returns USD value for ZCN token from coinmarketcap.com
func GetZcnUSDInfo(cb GetUSDInfoCallback) error {
	go func() {
		req, err := util.NewHTTPGetRequest("https://api.coinmarketcap.com/v2/ticker/2882/")
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
