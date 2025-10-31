package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/zcncrypto"
)

var (
	client         Client
	sdkInitialized bool

	Sign SignFunc
	// SignByMultiWallet SignByMultiWalletFunc
	sigC = make(chan struct{}, 1)
)

type SignFunc func(hash string, keys ...string) (string, error)

// type SignByMultiWalletFunc func(hash string, pubkey string) (string, error)

// maintains client's information
type Client struct {
	wallet          *zcncrypto.Wallet
	wallets         map[string]*zcncrypto.Wallet
	signatureScheme string
	splitKeyWallet  bool
	authUrl         string
	nonce           int64
	txnFee          uint64
	sign            SignFunc
	wg              map[string]*sync.WaitGroup
	walletCount     map[string]int // maintains count of wallets in the WaitGroup by pubkey
	mu              sync.RWMutex   // allow concurrent readers
}

type InitSdkOptions struct {
	WalletJSON              string
	BlockWorker             string
	ChainID                 string
	SignatureScheme         string
	Nonce                   int64
	IsSplitWallet           bool
	AddWallet               bool
	TxnFee                  *int
	MinConfirmation         *int
	ConfirmationChainLength *int
	MinSubmit               *int
	SharderConsensous       *int
	ZboxHost                string
	ZboxAppType             string
}

func init() {
	sys.Sign = signHash
	sys.SignWithAuth = signHashWithAuth

	// prime the sign channel
	sigC <- struct{}{}

	// default Sign implementation (uses client.wallet or wallets map)
	Sign = func(hash string, keys ...string) (string, error) {
		client.mu.RLock()
		defer client.mu.RUnlock()
		wallet := client.wallet
		if len(keys) > 0 && keys[0] != "" {
			if client.wallets != nil {
				if w, ok := client.wallets[keys[0]]; ok && w != nil {
					wallet = w
				}
			} else {
				return "", errors.New("no wallets available for signing by key: " + keys[0])
			}
		}

		if !wallet.IsSplit {
			return sys.Sign(hash, client.signatureScheme, GetClientSysKeys(keys...))
		}

		// split-key signing via auth
		<-sigC
		sig, err := sys.SignWithAuth(hash, client.signatureScheme, GetClientSysKeys(keys...), wallet.Keys[0].PublicKey)
		sigC <- struct{}{}
		return sig, err
	}

	sys.Verify = verifySignature
	sys.VerifyWith = verifySignatureWith
	sys.VerifyEd25519With = verifyEd25519With

	client.wg = make(map[string]*sync.WaitGroup)
	client.walletCount = make(map[string]int)
}

func GetClient() *zcncrypto.Wallet {
	return client.wallet
}

var SignFn = func(hash string) (string, error) {
	ss := zcncrypto.NewSignatureScheme(client.signatureScheme)

	err := ss.SetPrivateKey(client.wallet.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}

	return ss.Sign(hash)
}

// InitSDK Initialize the storage SDK
//
//   - walletJSON: Client's wallet JSON
//   - blockWorker: Block worker URL (block worker refers to 0DNS)
//   - chainID: ID of the blokcchain network
//   - signatureScheme: Signature scheme that will be used for signing transactions
//   - preferredBlobbers: List of preferred blobbers to use when creating an allocation. This is usually configured by the client in the configuration files
//   - nonce: Initial nonce value for the transactions
//   - fee: Preferred value for the transaction fee, just the first value is taken
func InitSDK(walletJSON string,
	blockWorker, chainID, signatureScheme string,
	nonce int64, addWallet bool,
	options ...int) error {

	if addWallet {
		wallet := zcncrypto.Wallet{}
		err := json.Unmarshal([]byte(walletJSON), &wallet)
		if err != nil {
			return err
		}

		SetWallet(wallet)
		SetSignatureScheme(signatureScheme)
		SetNonce(nonce)
		if len(options) > 0 {
			SetTxnFee(uint64(options[0]))
		}
	}

	var minConfirmation, minSubmit, confirmationChainLength, sharderConsensous int
	if len(options) > 1 {
		minConfirmation = options[1]
	}
	if len(options) > 2 {
		minSubmit = options[2]
	}
	if len(options) > 3 {
		confirmationChainLength = options[3]
	}
	if len(options) > 4 {
		sharderConsensous = options[4]
	}

	err := Init(context.Background(), conf.Config{
		BlockWorker:             blockWorker,
		SignatureScheme:         signatureScheme,
		ChainID:                 chainID,
		MinConfirmation:         minConfirmation,
		MinSubmit:               minSubmit,
		ConfirmationChainLength: confirmationChainLength,
		SharderConsensous:       sharderConsensous,
	})
	if err != nil {
		return err
	}
	SetSdkInitialized(true)
	return nil
}

func InitSDKWithWebApp(params InitSdkOptions) error {
	if params.MinConfirmation != nil && params.MinSubmit != nil && params.ConfirmationChainLength != nil && params.SharderConsensous != nil {
		err := InitSDK(params.WalletJSON, params.BlockWorker, params.ChainID, params.SignatureScheme, params.Nonce, params.AddWallet, *params.MinConfirmation, *params.MinSubmit, *params.ConfirmationChainLength, *params.SharderConsensous)
		if err != nil {
			return err
		}
	} else {
		err := InitSDK(params.WalletJSON, params.BlockWorker, params.ChainID, params.SignatureScheme, params.Nonce, params.AddWallet)
		if err != nil {
			return err
		}
	}
	conf.SetZboxAppConfigs(params.ZboxHost, params.ZboxAppType)
	SetIsAppFlow(true)
	return nil
}


func IsSDKInitialized() bool {
	return sdkInitialized
}

func SetSdkInitialized(val bool) {
	sdkInitialized = val
}

// Sign helpers that use sys package
func signHashWithAuth(hash, signatureScheme string, keys []sys.KeyPair, key ...string) (string, error) {
	sig, err := sys.Sign(hash, signatureScheme, keys)
	if err != nil {
		return "", fmt.Errorf("failed to sign with split key: %v", err)
	}

	var clientID string
	if len(key) > 0 && key[0] != "" {
		wallet := GetWalletByKey(key[0])
		if wallet == nil {
			return "", fmt.Errorf("wallet not found for pubkey: %s", key[0])
		}
		clientID = wallet.ClientID
	} else {
		clientID = client.wallet.ClientID
	}

	data, err := json.Marshal(AuthMessage{
		Hash:      hash,
		Signature: sig,
		ClientID:  clientID,
	})
	if err != nil {
		return "", err
	}

	if sys.AuthCommon == nil {
		return "", errors.New("authCommon is not set")
	}

	rsp, err := sys.AuthCommon(string(data), key...)
	if err != nil {
		return "", err
	}

	var sigpk struct {
		Sig string `json:"sig"`
	}

	if err = json.Unmarshal([]byte(rsp), &sigpk); err != nil {
		return "", err
	}

	return sigpk.Sig, nil
}

func signHash(hash string, signatureScheme string, keys []sys.KeyPair) (string, error) {
	retSignature := ""
	for _, kv := range keys {
		ss := zcncrypto.NewSignatureScheme(signatureScheme)

		err := ss.SetPrivateKey(kv.PrivateKey)
		if err != nil {
			return "", err
		}

		if len(retSignature) == 0 {
			retSignature, err = ss.Sign(hash)
		} else {
			retSignature, err = ss.Add(retSignature, hash)
		}
		if err != nil {
			return "", err
		}
	}
	return retSignature, nil
}

func verifySignature(signature string, msg string) (bool, error) {
	ss := zcncrypto.NewSignatureScheme(client.signatureScheme)
	if err := ss.SetPublicKey(client.wallet.ClientKey); err != nil {
		return false, err
	}

	return ss.Verify(signature, msg)
}

func verifySignatureWith(pubKey, signature, hash string) (bool, error) {
	sch := zcncrypto.NewSignatureScheme(client.signatureScheme)
	err := sch.SetPublicKey(pubKey)
	if err != nil {
		return false, err
	}
	return sch.Verify(signature, hash)
}

func verifyEd25519With(pubKey, signature, hash string) (bool, error) {
	sch := zcncrypto.NewSignatureScheme(constants.ED25519.String())
	err := sch.SetPublicKey(pubKey)
	if err != nil {
		return false, err
	}
	return sch.Verify(signature, hash)
}

// GetClientSysKeys reads wallets -> use RLock
func GetClientSysKeys(keys ...string) []sys.KeyPair {
	client.mu.RLock()
	defer client.mu.RUnlock()
	wallet := client.wallet
	if len(keys) > 0 && keys[0] != "" && client.wallets != nil {
		if w, ok := client.wallets[keys[0]]; ok && w != nil {
			wallet = w
		}
	}

	var sysKeys []sys.KeyPair
	for _, kv := range wallet.Keys {
		sysKeys = append(sysKeys, sys.KeyPair{
			PrivateKey: kv.PrivateKey,
			PublicKey:  kv.PublicKey,
		})
	}
	return sysKeys
}

// SetWallet should be set before any transaction or client specific APIs
func SetWallet(w zcncrypto.Wallet) {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.wallet = &w
	if client.wallets == nil {
		client.wallets = make(map[string]*zcncrypto.Wallet)
	}
	client.wallets[w.ClientID] = &w
}

// GetWalletByKey gets a wallet by client pubkey.
func GetWalletByKey(key string) *zcncrypto.Wallet {
	client.mu.RLock()
	defer client.mu.RUnlock()
	if client.wallets == nil {
		return nil
	}
	return client.wallets[key]
}

func GetWallet() *zcncrypto.Wallet {
	return client.wallet
}

// AddWallet adds a new wallet to the sdk.
func AddWallet(wallet zcncrypto.Wallet) {
	pubkey := wallet.Keys[0].PublicKey

	client.mu.Lock()
	defer client.mu.Unlock()

	if client.wallets == nil {
		client.wallets = make(map[string]*zcncrypto.Wallet)
	}
	if client.wg == nil {
		client.wg = make(map[string]*sync.WaitGroup)
	}
	if client.walletCount == nil {
		client.walletCount = make(map[string]int)
	}

	// if wallet already present, just increment counter
	if _, exists := client.wallets[pubkey]; exists {
		client.walletCount[pubkey]++
		// ensure wg exists
		if client.wg[pubkey] == nil {
			client.wg[pubkey] = &sync.WaitGroup{}
		}
		client.wg[pubkey].Add(1)
		return
	}

	// add new wallet
	client.wallets[pubkey] = &wallet
	if client.wg[pubkey] == nil {
		client.wg[pubkey] = &sync.WaitGroup{}
	}
	client.wg[pubkey].Add(1)
	client.walletCount[pubkey]++
}

// RemoveWallet removes a wallet from the sdk.
func RemoveWallet(pubkey string) {
	client.mu.Lock()
	defer client.mu.Unlock()

	if pubkey == "" || client.walletCount == nil {
		return
	}

	// only decrement if count > 0
	if client.walletCount[pubkey] > 0 {
		client.walletCount[pubkey]--
		// call Done on wg only if it exists
		if wg, ok := client.wg[pubkey]; ok && wg != nil {
			wg.Done()
		}
	}

	// if count reaches zero, clean up maps
	if client.walletCount[pubkey] == 0 {
		delete(client.wallets, pubkey)
		delete(client.wg, pubkey)
		delete(client.walletCount, pubkey)
	}
}

// SetWalletMode sets current wallet split key mode.
func SetWalletMode(mode bool) {
	if client.wallet != nil {
		client.wallet.IsSplit = mode
	}
}

// splitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
func SetSplitKeyWallet(isSplitKeyWallet bool) error {
	if client.signatureScheme == constants.BLS0CHAIN.String() {
		client.splitKeyWallet = isSplitKeyWallet
	}
	return nil
}

// SetAuthUrl will be called by app to set zauth URL to SDK
func SetAuthUrl(url string) error {
	if !client.splitKeyWallet {
		return errors.New("wallet type is not split key")
	}
	if url == "" {
		return errors.New("invalid auth url")
	}
	client.authUrl = strings.TrimRight(url, "/")
	return nil
}

func SetNonce(n int64) {
	client.nonce = n
}

func SetTxnFee(f uint64) {
	client.txnFee = f
}

func SetSignatureScheme(signatureScheme string) {
	if signatureScheme != constants.BLS0CHAIN.String() && signatureScheme != constants.ED25519.String() {
		panic("invalid/unsupported signature scheme")
	}
	client.signatureScheme = signatureScheme
}

func Wallet() *zcncrypto.Wallet {
	return client.wallet
}

func SignatureScheme() string {
	return client.signatureScheme
}

func SplitKeyWallet() bool {
	return client.splitKeyWallet
}

func AuthUrl() string {
	return client.authUrl
}

func Nonce() int64 {
	return client.nonce
}

func TxnFee() uint64 {
	return client.txnFee
}

func IsWalletSet() bool {
	return client.wallet.ClientID != ""
}

// PublicKey lookup uses read lock
func PublicKey(keys ...string) string {
	if len(keys) > 0 && keys[0] != "" {
		client.mu.RLock()
		if client.wallets != nil {
			if w, ok := client.wallets[keys[0]]; ok && w != nil {
				k := w.Keys[0].PublicKey
				client.mu.RUnlock()
				return k
			}
		}
		client.mu.RUnlock()
	}

	return client.wallet.ClientKey
}

func Mnemonic() string {
	return client.wallet.Mnemonic
}

func PrivateKey() string {
	for _, kv := range client.wallet.Keys {
		return kv.PrivateKey
	}
	return ""
}

func Id(keys ...string) string {
	if len(keys) > 0 && keys[0] != "" {
		client.mu.RLock()
		if client.wallets != nil {
			if w, ok := client.wallets[keys[0]]; ok && w != nil {
				id := w.ClientID
				client.mu.RUnlock()
				return id
			}
		}
		client.mu.RUnlock()
	}
	return client.wallet.ClientID
}

// VerifySignature ...
func VerifySignature(signature string, msg string) (bool, error) {
	ss := zcncrypto.NewSignatureScheme(client.signatureScheme)
	if err := ss.SetPublicKey(PublicKey()); err != nil {
		return false, err
	}
	return ss.Verify(signature, msg)
}

func VerifySignatureWith(pubKey, signature, hash string) (bool, error) {
	sch := zcncrypto.NewSignatureScheme(client.signatureScheme)
	err := sch.SetPublicKey(pubKey)
	if err != nil {
		return false, err
	}
	return sch.Verify(signature, hash)
}
