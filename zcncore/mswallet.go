package zcncore

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/zcncrypto"
)

// MSWallet Client data necessary for a multi-sig wallet.
type MSWallet struct {
	Id              int                                  `json:"id"`
	SignatureScheme string                               `json:"signature_scheme"`
	GroupClientID   string                               `json:"group_client_id"`
	GroupKey        zcncrypto.SignatureScheme            `json:"group_key"`
	SignerClientIDs []string                             `json:"sig_client_ids"`
	SignerKeys      []zcncrypto.ThresholdSignatureScheme `json:"signer_keys"`
	T               int                                  `json:"threshold"`
	N               int                                  `json:"num_subkeys"`
}

// Marshal returns json string
func (msw *MSWallet) Marshal() (string, error) {
	msws, err := json.Marshal(msw)
	if err != nil {
		return "", fmt.Errorf("Invalid Wallet")
	}
	return string(msws), nil
}

//MSWalletCallback callback definition that the callee is waiting on
type MSWalletCallback interface {
	OnMultiSigWalletCreated(status int, wallet string, wallets []string, err string)
}

// CreateMSWallet returns multisig wallet information
func CreateMSWallet(cb MSWalletCallback) error {
	Logger.Info("here in createMSWallet")
	id := 0 //Do we need this?
	t := 2  //number of keys (not percentage) required for token transfer
	n := 3  //total number of subkeys genereated
	if _config.chain.SignatureScheme != "bls0chain" {
		cb.OnMultiSigWalletCreated(StatusError, "", nil, "Encryption scheme for this blockchain is not bls0chain.")
		return nil
	}
	groupKey := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	wallet, err := groupKey.GenerateKeys(1)
	if err != nil {
		cb.OnMultiSigWalletCreated(StatusError, "", nil, fmt.Sprintf("%s", err.Error()))
		return nil
	}

	Logger.Info(fmt.Sprintf("Wallet id: %s", wallet.ClientKey))

	groupClientID := GetClientID(groupKey.GetPublicKey())
	signerKeys, err := zcncrypto.GenerateThresholdKeyShares(_config.chain.SignatureScheme, t, n, groupKey)

	if err != nil {
		cb.OnMultiSigWalletCreated(StatusError, "", nil, fmt.Sprintf("Err in generateThresholdKeyShares %s", err.Error()))
		return nil
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
	//registerMSWallets(msw, cb)

	wallets, errw := getWallets(msw)

	if errw != nil {
		cb.OnMultiSigWalletCreated(StatusError, "", nil, "Err in making wallets")
		return nil
	}
	smsw, er := msw.Marshal()
	if er != nil {
		cb.OnMultiSigWalletCreated(StatusError, "", nil, fmt.Sprintf("%s", er.Error()))
	} else {
		cb.OnMultiSigWalletCreated(StatusSuccess, smsw, wallets, "")
	}

	return nil
}

func RegisterWallet(walletString string, cb WalletCallback) {
	var w zcncrypto.Wallet
	err := json.Unmarshal([]byte(walletString), &w)

	if err != nil {
		cb.OnWalletCreateComplete(StatusError, walletString, fmt.Sprintf("%s", err.Error()))
	}

	//We do not want to send private key to blockchain
	w.Keys[0].PrivateKey = ""
	err = RegisterToMiners(&w, cb)
	if err != nil {
		cb.OnWalletCreateComplete(StatusError, "", fmt.Sprintf("%s", err.Error()))
	}

}
func getWallets(msw MSWallet) ([]string, error) {

	wallets := make([]string, 0, (msw.N + 1))

	b0ss, ok := msw.GroupKey.(*zcncrypto.BLS0ChainScheme)
	if !ok {
		return nil, errors.New("Err in making groupWallet")
	}

	grw, err := makeWallet(b0ss.PrivateKey, b0ss.PublicKey, b0ss.Mnemonic)

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
	w.DateCreated = time.Now().String()

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
