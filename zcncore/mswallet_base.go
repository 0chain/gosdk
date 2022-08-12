//go:build !mobile
// +build !mobile

package zcncore

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/zcncrypto"
)

//MSVoteCallback callback definition multisig Vote function
type MSVoteCallback interface {
	OnVoteComplete(status int, proposal string, err string)
}

// CreateMSWallet returns multisig wallet information
func CreateMSWallet(t, n int) (string, string, []string, error) {
	if t < 1 || t > n {
		return "", "", nil, errors.New("bls0_generate_threshold_key_shares", fmt.Sprintf("Given threshold (%d) is less than 1 or greater than numsigners (%d)", t, n))
	}

	id := 0
	if _config.chain.SignatureScheme != "bls0chain" {
		return "", "", nil, errors.New("", "encryption scheme for this blockchain is not bls0chain")

	}

	groupKey := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	wallet, err := groupKey.GenerateKeys()
	if err != nil {
		return "", "", nil, err
	}

	logging.Info(fmt.Sprintf("Wallet id: %s", wallet.ClientKey))

	groupClientID := GetClientID(groupKey.GetPublicKey())
	//Code modified to directly use BLS0ChainThresholdScheme
	signerKeys, err := zcncrypto.GenerateThresholdKeyShares(t, n, groupKey)

	if err != nil {
		return "", "", nil, errors.Wrap(err, "Err in generateThresholdKeyShares")
	}
	var signerClientIDs []string
	for _, key := range signerKeys {
		signerClientIDs = append(signerClientIDs, GetClientID(key.GetPublicKey()))
	}

	msw := MSWallet{
		Id:              id,
		SignatureScheme: _config.chain.SignatureScheme,
		GroupClientID:   groupClientID,
		GroupKey:        groupKey,
		SignerClientIDs: signerClientIDs,
		SignerKeys:      signerKeys,
		T:               t,
		N:               n,
	}

	wallets, errw := getWallets(msw)

	if errw != nil {
		return "", "", nil, errw

	}
	smsw, er := msw.Marshal()
	if er != nil {
		return "", "", nil, er
	}
	return smsw, groupClientID, wallets, nil

}

//RegisterWallet registers multisig related wallets
func RegisterWallet(walletString string, cb WalletCallback) {
	var w zcncrypto.Wallet
	err := json.Unmarshal([]byte(walletString), &w)

	if err != nil {
		cb.OnWalletCreateComplete(StatusError, walletString, err.Error())
	}

	//We do not want to send private key to blockchain
	w.Keys[0].PrivateKey = ""
	err = RegisterToMiners(&w, cb)
	if err != nil {
		cb.OnWalletCreateComplete(StatusError, "", err.Error())
	}
}

func getWallets(msw MSWallet) ([]string, error) {

	wallets := make([]string, 0, msw.N+1)

	b0ss := msw.GroupKey

	grw, err := makeWallet(b0ss.GetPrivateKey(), b0ss.GetPublicKey(), b0ss.GetMnemonic())

	if err != nil {
		return nil, err
	}

	wallets = append(wallets, grw)

	for _, signer := range msw.SignerKeys {
		w, err := makeWallet(signer.GetPrivateKey(), signer.GetPublicKey(), "")
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, w)
	}
	return wallets, nil
}

func makeWallet(privateKey, publicKey, mnemonic string) (string, error) {
	w := &zcncrypto.Wallet{}
	w.Keys = make([]zcncrypto.KeyPair, 1)
	w.Keys[0].PrivateKey = privateKey
	w.Keys[0].PublicKey = publicKey
	w.ClientID = GetClientID(publicKey) //VerifyThis
	w.ClientKey = publicKey
	w.Mnemonic = mnemonic
	w.Version = zcncrypto.CryptoVersion
	w.DateCreated = time.Now().Format(time.RFC3339)

	return w.Marshal()
}

// GetClientID -- computes Client ID from publickey
func GetClientID(pkey string) string {
	publicKeyBytes, err := hex.DecodeString(pkey)
	if err != nil {
		panic(err)
	}

	return encryption.Hash(publicKeyBytes)
}

func GetClientWalletID() string {
	return _config.wallet.ClientID
}
