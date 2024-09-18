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

type SignFunc func(hash string, clientId ...string) (string, error)

// Client maintains client's information
type Client struct {
	wallet          *zcncrypto.Wallet
	wallets         map[string]*zcncrypto.Wallet
	useMultiWallets bool
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
		sign: func(hash string, clientId ...string) (string, error) {
			if len(clientId) > 0 {
				w, ok := client.wallets[clientId[0]]
				if !ok {
					return "", errors.New("invalid client id")
				}
				return sys.Sign(hash, client.signatureScheme, GetClientSysKeys(w.ClientID))
			}
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

func GetClientSysKeys(clientId ...string) []sys.KeyPair {
	wallet := client.wallet
	if len(clientId) > 0 {
		w, ok := client.wallets[clientId[0]]
		if !ok {
			return nil
		}
		wallet = w
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

// SetSplitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
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

func Sign(hash string, clientId ...string) (string, error) {
	if len(clientId) > 0 {
		w, ok := client.wallets[clientId[0]]
		if !ok {
			return "", errors.New("invalid client id")
		}
		return client.sign(hash, w.ClientID)
	}
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

func Id() string {
	return client.wallet.ClientID
}

func GetClient() *zcncrypto.Wallet {
	return client.wallet
}
