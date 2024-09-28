package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/0chain/gosdk/core/conf"
	"strings"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/zcncrypto"
)

var (
	client         Client
	sdkInitialized bool
)

type SignFunc func(hash string) (string, error)

// maintains client's information
type Client struct {
	wallet          *zcncrypto.Wallet
	signatureScheme string
	splitKeyWallet  bool
	authUrl         string
	nonce           int64
	txnFee          uint64
	sign            SignFunc
}

func init() {
	sys.Sign = signHash
	client = Client{
		wallet: &zcncrypto.Wallet{},
		sign: func(hash string) (string, error) {
			return sys.Sign(hash, client.signatureScheme, GetClientSysKeys())
		},
	}
	sys.Verify = verifySignature
	sys.VerifyWith = verifySignatureWith
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

func GetClientSysKeys() []sys.KeyPair {
	var keys []sys.KeyPair
	for _, kv := range client.wallet.Keys {
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

func Sign(hash string) (string, error) {
	return client.sign(hash)
}

func IsWalletSet() bool {
	return client.wallet.ClientID != ""
}

func PublicKey() string {
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

func ClientID() string {
	if client.wallet.ClientID == "" {
		fmt.Println("ClientID is empty")
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
	preferredBlobbers []string,
	nonce int64, isSplitWallet, addWallet bool,
	fee ...uint64) error {

	if addWallet {
		wallet := zcncrypto.Wallet{}
		err := json.Unmarshal([]byte(walletJSON), &wallet)
		if err != nil {
			return err
		}

		SetWallet(wallet)
		SetSignatureScheme(signatureScheme)
		SetNonce(nonce)
		if len(fee) > 0 {
			SetTxnFee(fee[0])
		}
	}

	err := Init(context.Background(), conf.Config{
		BlockWorker:       blockWorker,
		SignatureScheme:   signatureScheme,
		ChainID:           chainID,
		PreferredBlobbers: preferredBlobbers,
		MaxTxnQuery:       5,
		QuerySleepTime:    5,
		MinSubmit:         10,
		MinConfirmation:   10,
		IsSplitWallet:     isSplitWallet,
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
