package zcncore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
}

var defaultLogLevel = logger.DEBUG
var Logger logger.Logger

var LATEST_FINALIZED_BLOCK = `/v1/block/get/latest_finalized`
var REGISTER_CLIENT = `/v1/client/put`
var PUT_TRANSACTION = `/v1/transaction/put`
var TXN_VERIFY_URL = `/v1/transaction/get/confirmation?hash=`
var GET_BALANCE = `/v1/client/get/balance?client_id=`
var GET_LOCK_CONFIG = `/v1/scstate/get?sc_address=`
var GET_LOCKED_TOKENS = `/v1/screst/` + InterestPoolSmartContractAddress + `/getPoolsStats?client_id=`

const StorageSmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7`
const FaucetSmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d3`
const InterestPoolSmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9`
const MultiSigSmartContractAddress = `27b5ef7120252b79f9dd9c05505dd28f328c80f6863ee446daede08a84d651a7`
const MultiSigRegisterFuncName = "register"
const MultiSigVoteFuncName = "vote"

// In percentage
const consensusThresh = float32(25.0)

const (
	StatusSuccess      int = 0
	StatusNetworkError int = 1
	// TODO: Change to specific error
	StatusError   int = 2
	StatusUnknown int = -1
)

const TOKEN_UNIT = int64(10000000000)

const (
	OpGetTokenLockConfig int = 0
	OpGetLockedTokens    int = 1
)

// WalletCallback needs to be implmented for wallet creation.
type WalletCallback interface {
	OnWalletCreateComplete(status int, wallet string, err string)
}

// GetBalanceCallback needs to be implemented by the caller of GetBalance() to get the status
type GetBalanceCallback interface {
	OnBalanceAvailable(status int, value int64)
}

// GetInfoCallback needs to be implemented by the caller of GetLockTokenConfig() and GetLockedTokens()
type GetInfoCallback interface {
	// OnInfoAvailable will be called when GetLockTokenConfig is complete
	// if status == StatusSuccess then info is valid
	// is status != StatusSuccess then err will give the reason
	OnInfoAvailable(op int, status int, info string, err string)
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
	isConfigured  bool
	isValidWallet bool
}

// Singleton
var _config localConfig

func init() {
	Logger.Init(defaultLogLevel, "0chain-core-sdk")
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
	Logger.Info("*******  Wallet SDK Version:", version.VERSIONSTR, " *******")
	return nil
}

// CreateWallet creates the a wallet for the configure signature scheme.
// It also registers the wallet again to block chain.
func CreateWallet(numKeys int, statusCb WalletCallback) error {
	if len(_config.chain.Miners) < 1 || len(_config.chain.Sharders) < 1 {
		return fmt.Errorf("SDK not initialized")
	}
	go func() {
		sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
		wallet, err := sigScheme.GenerateKeys(numKeys)
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
func RecoverWallet(mnemonic string, numKeys int, statusCb WalletCallback) error {
	if numKeys < 1 {
		return fmt.Errorf("Invalid number of keys")
	}
	if zcncrypto.IsMnemonicValid(mnemonic) != true {
		return fmt.Errorf("Invalid mnemonic")
	}
	go func() {
		sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
		wallet, err := sigScheme.RecoverKeys(mnemonic, numKeys)
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
			// Logger.Debug(rsp.Body)
			if rsp.StatusCode == http.StatusOK {
				consensus++
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

// SetWalletInfo should be set before any transaction or GetBalance APIs
func SetWalletInfo(w string) error {
	err := json.Unmarshal([]byte(w), &_config.wallet)
	if err == nil {
		_config.isValidWallet = true
	}
	return err
}

func checkWalletConfig() error {
	if len(_config.chain.Miners) < 1 || len(_config.chain.Sharders) < 1 {
		return fmt.Errorf("SDK not initialized")
	}
	if _config.wallet.ClientID == "" {
		Logger.Error("wallet info not found. returning error.")
		return fmt.Errorf("wallet info not found. set wallet info.")
	}
	return nil
}

// GetBalance retreives wallet balance from sharders
func GetBalance(cb GetBalanceCallback) error {
	err := checkWalletConfig()
	if err != nil {
		return err
	}
	go func() {
		value, err := getBalanceFromSharders(_config.wallet.ClientID)
		if err != nil {
			Logger.Error(err)
			cb.OnBalanceAvailable(StatusError, 0)
			return
		}
		cb.OnBalanceAvailable(StatusSuccess, value)
	}()
	return nil
}

func getBalanceFromSharders(clientID string) (int64, error) {
	result := make(chan *util.GetResponse)
	defer close(result)
	for _, sharder := range _config.chain.Sharders {
		go func(sharderurl string) {
			Logger.Debug("Getting balance from sharder:", sharderurl)
			url := fmt.Sprintf("%v%v%v", sharder, GET_BALANCE, clientID)
			req, err := util.NewHTTPGetRequest(url)
			if err != nil {
				Logger.Error(sharder, "new get request failed. ", err.Error())
				return
			}
			res, err := req.Get()
			if err != nil {
				Logger.Error(sharder, "get error. ", err.Error())
			}
			result <- res
			return
		}(sharder)
	}
	consensus := float32(0)
	balMap := make(map[int64]float32)
	winBalance := int64(0)
	for range _config.chain.Sharders {
		select {
		case rsp := <-result:
			Logger.Debug(rsp.Url, rsp.Status)
			if rsp.StatusCode != http.StatusOK {
				Logger.Error(rsp.Body)
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
				}
			}
		}
	}
	rate := consensus * 100 / float32(len(_config.chain.Sharders))
	if rate < consensusThresh {
		return 0, fmt.Errorf("get balance failed. consensus not reached")
	}
	return winBalance, nil
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
	for _, sharder := range _config.chain.Sharders {
		go func(sharderurl string) {
			Logger.Debug("Getting info from sharder:", sharderurl)
			url := fmt.Sprintf("%v%v", sharder, urlSuffix)
			req, err := util.NewHTTPGetRequest(url)
			if err != nil {
				Logger.Error(sharder, "new get request failed. ", err.Error())
				return
			}
			res, err := req.Get()
			if err != nil {
				Logger.Error(sharder, "get error. ", err.Error())
			}
			result <- res
			return
		}(sharder)
	}
	consensus := float32(0)
	var tSuccessRsp string
	var tFailureRsp string
	for range _config.chain.Sharders {
		select {
		case rsp := <-result:
			Logger.Debug(rsp.Url, rsp.Status)
			if rsp.StatusCode != http.StatusOK {
				Logger.Error(rsp.Body)
				tFailureRsp = rsp.Body
				continue
			}
			// TODO: Any other validation for consensus
			consensus++
			tSuccessRsp = rsp.Body
		}
	}
	rate := consensus * 100 / float32(len(_config.chain.Sharders))
	if rate < consensusThresh {
		cb.OnInfoAvailable(op, StatusError, "", fmt.Sprintf("consensus not reached. %v", tFailureRsp))
		return
	}
	cb.OnInfoAvailable(op, StatusSuccess, tSuccessRsp, "")
}

// GetLockConfig returns the lock token configuration information such as interest rate from blockchain
func GetLockConfig(cb GetInfoCallback) error {
	if len(_config.chain.Miners) < 1 || len(_config.chain.Sharders) < 1 {
		return fmt.Errorf("SDK not initialized")
	}
	err := checkWalletConfig()
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
	if len(_config.chain.Miners) < 1 || len(_config.chain.Sharders) < 1 {
		return fmt.Errorf("SDK not initialized")
	}
	err := checkWalletConfig()
	if err != nil {
		return err
	}
	go func() {
		urlSuffix := fmt.Sprintf("%v%v", GET_LOCKED_TOKENS, _config.wallet.ClientID)
		getInfoFromSharders(urlSuffix, OpGetLockedTokens, cb)
	}()
	return nil
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
