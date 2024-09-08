package client

import (
	"errors"
	"strings"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/zcncrypto"
)

var (
	client Client
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
func SetWallet(isSplit bool, w zcncrypto.Wallet) {
	client.wallet = &w
	client.wallet.IsSplit = isSplit
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

func SetNonce(n int64) error {
	client.nonce = n
	return nil
}

func SetTxnFee(f uint64) error {
	client.txnFee = f
	return nil
}

func SetSignatureScheme(signatureScheme string) error {
	if signatureScheme != constants.BLS0CHAIN.String() && signatureScheme != constants.ED25519.String() {
		return errors.New("invalid/unsupported signature scheme")
	}
	client.signatureScheme = signatureScheme
	return nil
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
	return client.wallet.ClientID
}

func GetWallet() *zcncrypto.Wallet {
	return client.wallet
}

func GetClient() *zcncrypto.Wallet {
	return client.wallet
}
