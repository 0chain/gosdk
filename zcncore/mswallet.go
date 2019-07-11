package zcncore

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

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
	OnMultiSigWalletCreated(status int, wallet string, err string)
}

// CreateMSWallet returns multisig wallet information
func CreateMSWallet(cb MSWalletCallback) error {
	Logger.Info("here in createMSWallet")
	id := 0 //Do we need this?
	t := 2  //number of keys (not percentage) required for token transfer
	n := 3  //total number of subkeys genereated
	if _config.chain.SignatureScheme != "bls0chain" {
		cb.OnMultiSigWalletCreated(StatusError, "", "Encryption scheme for this blockchain is not bls0chain.")
		return nil
	}
	groupKey := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	wallet, err := groupKey.GenerateKeys(1)
	if err != nil {
		cb.OnMultiSigWalletCreated(StatusError, "", fmt.Sprintf("%s", err.Error()))
		return nil
	}

	Logger.Info(fmt.Sprintf("Wallet id: %s", wallet.ClientKey))

	groupClientID := clientIDForKey(groupKey)
	signerKeys, err := zcncrypto.GenerateThresholdKeyShares(_config.chain.SignatureScheme, t, n, groupKey)

	if err != nil {
		cb.OnMultiSigWalletCreated(StatusError, "", fmt.Sprintf("Err in generateThresholdKeyShares %s", err.Error()))
		return nil
	}
	var signerClientIDs []string
	for _, key := range signerKeys {
		signerClientIDs = append(signerClientIDs, clientIDForKey(key))
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

	smsw, er := msw.Marshal()
	if er != nil {
		cb.OnMultiSigWalletCreated(StatusError, "", fmt.Sprintf("%s", err.Error()))
	} else {
		cb.OnMultiSigWalletCreated(StatusSuccess, smsw, "")
	}

	return nil
}

func clientIDForKey(key zcncrypto.SignatureScheme) string {
	publicKeyBytes, err := hex.DecodeString(key.GetPublicKey())
	if err != nil {
		panic(err)
	}

	return encryption.Hash(publicKeyBytes)
}
