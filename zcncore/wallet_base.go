package zcncore

import (
	"encoding/hex"
	"fmt"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/util"
	"strings"
	"time"

	"errors"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/logger"
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

	GET_MINT_NONCE = `/v1/mint_nonce`

	// miner SC

	GET_MINERSC_NODE     = "/nodeStat"
	GET_MINERSC_POOL     = "/nodePoolStat"
	GET_MINERSC_USER     = "/getUserPools"
	GET_MINERSC_MINERS   = "/getMinerList"
	GET_MINERSC_SHARDERS = "/getSharderList"
)

const (
	StorageSmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7`
	FaucetSmartContractAddress  = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d3`
	MinerSmartContractAddress   = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9`
	ZCNSCSmartContractAddress   = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712e0`
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
	// OnSetupComplete This call back gives the status of the Two factor authenticator(zauth) setup.
	OnSetupComplete(status int, err string)
}

func init() {
	logging.Init(defaultLogLevel, "0chain-core-sdk")
}

func checkSdkInit() error {
	_, err := client.GetNode()
	if err != nil {
		return err
	}
	return nil
}

func checkWalletConfig() error {
	if !client.IsWalletSet() {
		return errors.New("wallet info not found. set wallet info")
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

// SetLogLevel set the log level.
// lvl - 0 disabled; higher number (upto 4) more verbosity
func SetLogLevel(lvl int) {
	logging.SetLevel(lvl)
}

// SetLogFile - sets file path to write log
// verbose - true - console output; false - no console output
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

// CreateWalletOffline creates the wallet for the config signature scheme.
func CreateWalletOffline() (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme("bls0chain")
	wallet, err := sigScheme.GenerateKeys()
	if err != nil {
		return "", errors.New("failed to generate keys: " + err.Error())
	}
	w, err := wallet.Marshal()
	if err != nil {
		return "", errors.New("wallet encoding failed: " + err.Error())
	}
	return w, nil
}

// RecoverOfflineWallet recovers the previously generated wallet using the mnemonic.
func RecoverOfflineWallet(mnemonic string) (string, error) {
	if !zcncrypto.IsMnemonicValid(mnemonic) {
		return "", errors.New("Invalid mnemonic")
	}
	sigScheme := zcncrypto.NewSignatureScheme(client.SignatureScheme())
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
		return errors.New("Invalid mnemonic")
	}
	go func() {
		sigScheme := zcncrypto.NewSignatureScheme(client.SignatureScheme())
		_, err := sigScheme.RecoverKeys(mnemonic)
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", err.Error())
			return
		}
	}()
	return nil
}

// SplitKeys Split keys from the primary master key
func SplitKeys(privateKey string, numSplits int) (string, error) {
	if client.SignatureScheme() != constants.BLS0CHAIN.String() {
		return "", errors.New("signature key doesn't support split key")
	}
	sigScheme := zcncrypto.NewSignatureScheme(client.SignatureScheme())
	err := sigScheme.SetPrivateKey(privateKey)
	if err != nil {
		return "", errors.New("set private key failed." + err.Error())
	}
	w, err := sigScheme.SplitKeys(numSplits)
	if err != nil {
		return "", errors.New("split key failed." + err.Error())
	}
	wStr, err := w.Marshal()
	if err != nil {
		return "", errors.New("wallet encoding failed." + err.Error())
	}
	return wStr, nil
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

// ConvertToValue converts ZCN tokens to SAS tokens
// # Inputs
//   - token: ZCN tokens
func ConvertToValue(token float64) uint64 {
	return uint64(token * common.TokenUnit)
}

func SignWithKey(privateKey, hash string) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme("bls0chain")
	err := sigScheme.SetPrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}

var AddSignature = func(privateKey, signature string, hash string) (string, error) {
	var (
		ss  = zcncrypto.NewSignatureScheme(client.SignatureScheme())
		err error
	)

	err = ss.SetPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	return ss.Add(signature, hash)
}

// ConvertToToken converts the SAS tokens to ZCN tokens
//   - token: SAS tokens amount
func ConvertToToken(token int64) float64 {
	return float64(token) / float64(common.TokenUnit)
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
