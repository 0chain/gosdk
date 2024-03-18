//go:build !mobile
// +build !mobile

package zcncore

import (
	"encoding/json"
	"fmt"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/zcncrypto"
)

//MSVote -- this should mimic the type Vote defined in MultiSig SC
type MSVote struct {
	ProposalID string `json:"proposal_id"`

	// Client ID in transfer is that of the multi-sig wallet, not the signer.
	Transfer MSTransfer `json:"transfer"`

	Signature string `json:"signature"`
}

//MSTransfer - a data structure to hold state transfer from one client to another
type MSTransfer struct {
	ClientID   string `json:"from"`
	ToClientID string `json:"to"`
	Amount     uint64 `json:"amount"`
}

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
	Id              int                         `json:"id"`
	SignatureScheme string                      `json:"signature_scheme"`
	GroupClientID   string                      `json:"group_client_id"`
	GroupKey        zcncrypto.SignatureScheme   `json:"group_key"`
	SignerClientIDs []string                    `json:"sig_client_ids"`
	SignerKeys      []zcncrypto.SignatureScheme `json:"signer_keys"`
	T               int                         `json:"threshold"`
	N               int                         `json:"num_subkeys"`
}

func (msw *MSWallet) UnmarshalJSON(data []byte) error {
	m := &struct {
		Id              int         `json:"id"`
		SignatureScheme string      `json:"signature_scheme"`
		GroupClientID   string      `json:"group_client_id"`
		SignerClientIDs []string    `json:"sig_client_ids"`
		T               int         `json:"threshold"`
		N               int         `json:"num_subkeys"`
		GroupKey        interface{} `json:"group_key"`
		SignerKeys      interface{} `json:"signer_keys"`
	}{}

	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	msw.Id = m.Id
	msw.SignatureScheme = m.SignatureScheme
	msw.GroupClientID = m.GroupClientID
	msw.SignerClientIDs = m.SignerClientIDs
	msw.T = m.T
	msw.N = m.N

	if m.GroupKey != nil {
		groupKeyBuf, err := json.Marshal(m.GroupKey)
		if err != nil {
			return err
		}

		ss := zcncrypto.NewSignatureScheme(m.SignatureScheme)

		if err := json.Unmarshal(groupKeyBuf, &ss); err != nil {
			return err
		}

		msw.GroupKey = ss
	}

	signerKeys, err := zcncrypto.UnmarshalSignatureSchemes(m.SignatureScheme, m.SignerKeys)
	if err != nil {
		return err
	}
	msw.SignerKeys = signerKeys

	return nil
}

// Marshal returns json string
func (msw *MSWallet) Marshal() (string, error) {
	msws, err := json.Marshal(msw)
	if err != nil {
		return "", errors.New("", "Invalid Wallet")
	}
	return string(msws), nil
}

//GetMultisigPayload given a multisig wallet as a string, makes a multisig wallet payload to register
func GetMultisigPayload(mswstr string) (interface{}, error) {
	var msw MSWallet
	err := json.Unmarshal([]byte(mswstr), &msw)

	if err != nil {
		fmt.Printf("Error while creating multisig wallet from input:\n%v", mswstr)
		return "", err
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

	return msscw, nil
}

//GetMultisigVotePayload given a multisig vote as a string, makes a multisig vote payload to register
func GetMultisigVotePayload(msvstr string) (interface{}, error) {
	var msv MSVote
	err := json.Unmarshal([]byte(msvstr), &msv)

	if err != nil {
		fmt.Printf("Error while creating multisig wallet from input:\n%v", msvstr)
		return nil, err
	}

	//Marshalling and unmarshalling validates the string. Do any additional veirfication here.

	return msv, nil

}

// CreateMSVote create a vote for multisig
func CreateMSVote(proposal, grpClientID, signerWalletstr, toClientID string, token uint64) (string, error) {
	if proposal == "" || grpClientID == "" || toClientID == "" || signerWalletstr == "" {
		return "", errors.New("", "proposal or groupClient or signer wallet or toClientID cannot be empty")
	}

	if token < 1 {
		return "", errors.New("", "Token cannot be less than 1")
	}

	signerWallet, err := getWallet(signerWalletstr)
	if err != nil {
		fmt.Printf("Error while parsing the signer wallet. %v", err)
		return "", err
	}

	//Note: Is this honored by multisig sc?
	transfer := MSTransfer{
		ClientID:   grpClientID,
		ToClientID: toClientID,
		Amount:     token,
	}

	buff, _ := json.Marshal(transfer)
	hash := encryption.Hash(buff)

	sigScheme := zcncrypto.NewSignatureScheme(signatureScheme)
	if err := sigScheme.SetPrivateKey(signerWallet.Keys[0].PrivateKey); err != nil {
		return "", err
	}

	sig, err := sigScheme.Sign(hash)
	if err != nil {
		return "", err
	}

	vote := MSVote{
		Transfer:   transfer,
		ProposalID: proposal,
		Signature:  sig,
	}

	vbytes, err := json.Marshal(vote)
	if err != nil {
		fmt.Printf("error in marshalling vote %v", vote)
		return "", err
	}
	return string(vbytes), nil
}
