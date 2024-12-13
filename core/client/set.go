package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/0chain/gosdk/core/conf"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/zcncrypto"
)

var (
	client         Client
	sdkInitialized bool

	Sign SignFunc
	sigC = make(chan struct{}, 1)
)

type SignFunc func(hash string, clients ...string) (string, error)

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
}

func init() {
	sys.Sign = signHash
	sys.SignWithAuth = signHash

	sigC <- struct{}{}

	// initialize SignFunc as default implementation
	Sign = func(hash string, clients ...string) (string, error) {
		wallet := client.wallet

		if len(clients) > 0 && clients[0] != "" && client.wallets[clients[0]] != nil {
			wallet = client.wallets[clients[0]]
		}

		if !wallet.IsSplit {
			return sys.Sign(hash, client.signatureScheme, GetClientSysKeys(clients...))
		}

		// get sign lock
		<-sigC
		fmt.Println("Sign: with sys.SignWithAuth:", sys.SignWithAuth, "sysKeys:", GetClientSysKeys(clients...))
		sig, err := sys.SignWithAuth(hash, client.signatureScheme, GetClientSysKeys(clients...))
		sigC <- struct{}{}
		return sig, err
	}

	sys.Verify = verifySignature
	sys.VerifyWith = verifySignatureWith
	sys.VerifyEd25519With = verifyEd25519With
}

var SignFn = func(hash string) (string, error) {
	ss := zcncrypto.NewSignatureScheme(client.signatureScheme)

	err := ss.SetPrivateKey(client.wallet.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}

	return ss.Sign(hash)
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

func GetClientSysKeys(clients ...string) []sys.KeyPair {
	var wallet *zcncrypto.Wallet
	if len(clients) > 0 && clients[0] != "" && client.wallets[clients[0]] != nil {
		wallet = client.wallets[clients[0]]
	} else {
		wallet = client.wallet
	}

	var keys []sys.KeyPair
	for _, kv := range wallet.Keys {
		keys = append(keys, sys.KeyPair{
			PrivateKey: kv.PrivateKey,
			PublicKey:  kv.PublicKey,
		})
	}

	return keys
}

// SetWallet should be set before any transaction or client specific APIs
func SetWallet(w zcncrypto.Wallet) {
	client.wallet = &w
	if client.wallets == nil {
		client.wallets = make(map[string]*zcncrypto.Wallet)
	}
	client.wallets[w.ClientID] = &w
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

func PublicKey(clients ...string) string {
	if len(clients) > 0 && clients[0] != "" && client.wallets[clients[0]] != nil {
		if client.wallets[clients[0]] == nil {
			fmt.Println("Public key is empty")
			return ""
		}
		return client.wallets[clients[0]].ClientKey
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

func Id(clients ...string) string {
	if len(clients) > 0 && clients[0] != "" && client.wallets[clients[0]] != nil {
		if client.wallets[clients[0]] == nil {
			fmt.Println("Id is empty : ", clients[0])
			return ""
		}
		return client.wallets[clients[0]].ClientID
	}
	return client.wallet.ClientID
}

func GetWallet() *zcncrypto.Wallet {
	return client.wallet
}

func GetClient() *zcncrypto.Wallet {
	return client.wallet
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

func IsSDKInitialized() bool {
	return sdkInitialized
}

func SetSdkInitialized(val bool) {
	sdkInitialized = val
}

func PopulateClient(walletJSON, signatureScheme string) (zcncrypto.Wallet, error) {
	wallet := zcncrypto.Wallet{}
	err := json.Unmarshal([]byte(walletJSON), &wallet)
	if err != nil {
		return wallet, err
	}

	SetWallet(wallet)
	SetSignatureScheme(signatureScheme)
	return wallet, nil
}

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
