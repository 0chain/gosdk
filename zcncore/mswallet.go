package zcncore

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/zcncrypto"
)

// MultisigSCWallet --this should mimic MultisigWallet definition in MultiSig SC
type MultisigSCWallet struct {
	ClientID        string `json:"client_id"`
	SignatureScheme string `json:"signature_scheme"`
	PublicKey       string `json:"public_key"`

	SignerThresholdIDs []string `json:"signer_threshold_ids"`
	SignerPublicKeys   []string `json:"signer_public_keys"`

	NumRequired int `json:"num_required"`
}

// MSWallet Client data necessary for a multi-sig wallet.
type MSWallet struct {
	Id              int                                  `json:"id"`
	SignatureScheme string                               `json:"signature_scheme"`
	GroupClientID   string                               `json:"group_client_id"`
	GroupKey        *zcncrypto.BLS0ChainScheme           `json:"group_key"`
	SignerClientIDs []string                             `json:"sig_client_ids"`
	SignerKeys      []zcncrypto.BLS0ChainThresholdScheme `json:"signer_keys"`
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

	groupKey := zcncrypto.NewBLS0ChainScheme()
	wallet, err := groupKey.GenerateKeys(1)
	if err != nil {
		cb.OnMultiSigWalletCreated(StatusError, "", nil, fmt.Sprintf("%s", err.Error()))
		return nil
	}

	Logger.Info(fmt.Sprintf("Wallet id: %s", wallet.ClientKey))

	groupClientID := GetClientID(groupKey.GetPublicKey())
	//Code modified to directly use BLS0ChainThresholdScheme
	signerKeys, err := zcncrypto.BLS0GenerateThresholdKeyShares(t, n, groupKey)

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

//RegisterWallet registers multisig related wallets
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

	b0ss := msw.GroupKey

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

//GetMultisigPayload given a multisig wallet as a string, makes a multisig wallet payload to register
func GetMultisigPayload(mswstr string) ([]byte, error) {
	var msw MSWallet
	err := json.Unmarshal([]byte(mswstr), &msw)

	if err != nil {
		fmt.Printf("Error while creating multisig wallet from input:\n%v", mswstr)
		return nil, err
	}
	var signerThresholdIDs []string
	var signerPublicKeys []string

	for _, scheme := range msw.SignerKeys {
		signerThresholdIDs = append(signerThresholdIDs, scheme.GetID())
		signerPublicKeys = append(signerPublicKeys, scheme.GetPublicKey())
	}

	msscw := MultisigSCWallet{
		ClientID:        msw.GroupClientID,
		SignatureScheme: msw.SignatureScheme,
		PublicKey:       msw.GroupKey.GetPublicKey(),

		SignerThresholdIDs: signerThresholdIDs,
		SignerPublicKeys:   signerPublicKeys,

		NumRequired: msw.T,
	}

	msscwBytes, err := json.Marshal(msscw)
	if err != nil {
		fmt.Printf("\nerror in converting msscw to bytes:%v\n", err)
	}
	return msscwBytes, nil

}
