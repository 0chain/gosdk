package client

import (
	"errors"
	"strings"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/zcncrypto"
)

// maintains client's data 
var (
	wallet 	*zcncrypto.Wallet
	splitKeyWallet	bool
	authUrl string
	nonce	int64
	fee 	uint64
)

func init() {
	wallet = &zcncrypto.Wallet{}
}

// SetWallet should be set before any transaction or client specific APIs
func SetWallet(w zcncrypto.Wallet) error {
	wallet = &w
	return nil
}

// splitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
func SetSplitKeyWallet(isSplitKeyWallet bool) error {
	cfg, err := conf.GetClientConfig()
	if err != nil {
		return err
	}
	if cfg.SignatureScheme == constants.BLS0CHAIN.String() {
		splitKeyWallet = isSplitKeyWallet
	}
	return nil
}

// SetAuthUrl will be called by app to set zauth URL to SDK
func SetAuthUrl(url string) error {
	if !splitKeyWallet {
		return errors.New("wallet type is not split key")
	}
	if url == "" {
		return errors.New("invalid auth url")
	}
	authUrl = strings.TrimRight(url, "/")
	return nil
}

func SetNonce(n int64) error {
	nonce = n
	return nil
}

func SetFee(f uint64) error {
	fee = f
	return nil
}

func Wallet() *zcncrypto.Wallet {
	return wallet
}

func SplitKeyWallet() bool {
	return splitKeyWallet
}

func AuthUrl() string {
	return authUrl
}

func Nonce() int64 {
	return nonce
}

func Fee() uint64 {
	return fee
}

func IsWalletSet() bool {
	return wallet == nil || wallet.ClientID != ""
}
