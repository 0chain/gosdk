//go:build mobile
// +build mobile

package zcncore

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/zcncrypto"
)

type MultisigSCWallet interface {
	GetClientID() string
	GetSignatureScheme() string
	GetPublicKey() string
	GetNumRequired() int
	GetSignerThresholdIDs() Stringers
	GetSignerPublicKeys() Stringers
}

// Stringers wraps the methods for accessing string slice
type Stringers interface {
	Len() int                  // return the number of string slice
	Get(i int) (string, error) // get string of given index
}

// stringSlice implements the Stringers interface
type stringSlice []string

func (ss stringSlice) Len() int {
	return len(ss)
}

func (ss stringSlice) Get(i int) (string, error) {
	if i < 0 || i >= len(ss) {
		return "", errors.New("index out of bounds")
	}
	return ss[i], nil
}

//GetMultisigPayload given a multisig wallet as a string, makes a multisig wallet payload to register
func GetMultisigPayload(mswstr string) (MultisigSCWallet, error) {
	var msw msWallet
	err := json.Unmarshal([]byte(mswstr), &msw)
	if err != nil {
		return nil, err
	}

	var signerThresholdIDs []string
	var signerPublicKeys []string

	for _, scheme := range msw.SignerKeys {
		signerThresholdIDs = append(signerThresholdIDs, scheme.GetID())
		signerPublicKeys = append(signerPublicKeys, scheme.GetPublicKey())
	}

	return &multisigSCWallet{
		ClientID:        msw.GroupClientID,
		SignatureScheme: msw.SignatureScheme,
		PublicKey:       msw.GroupKey.GetPublicKey(),

		SignerThresholdIDs: signerThresholdIDs,
		SignerPublicKeys:   signerPublicKeys,

		NumRequired: msw.T,
	}, nil
}

type multisigSCWallet struct {
	ClientID        string `json:"client_id"`
	SignatureScheme string `json:"signature_scheme"`
	PublicKey       string `json:"public_key"`

	SignerThresholdIDs []string `json:"signer_threshold_ids"`
	SignerPublicKeys   []string `json:"signer_public_keys"`

	NumRequired int `json:"num_required"`
}

func (m *multisigSCWallet) GetClientID() string {
	return m.ClientID
}

func (m *multisigSCWallet) GetSignatureScheme() string {
	return m.SignatureScheme
}

func (m *multisigSCWallet) GetPublicKey() string {
	return m.PublicKey
}

func (m *multisigSCWallet) GetSignerThresholdIDs() Stringers {
	return stringSlice(m.SignerThresholdIDs)
}

func (m *multisigSCWallet) GetSignerPublicKeys() Stringers {
	return stringSlice(m.SignerPublicKeys)
}

func (m *multisigSCWallet) GetNumRequired() int {
	return m.NumRequired
}

type msWallet struct {
	Id              int                         `json:"id"`
	SignatureScheme string                      `json:"signature_scheme"`
	GroupClientID   string                      `json:"group_client_id"`
	GroupKey        zcncrypto.SignatureScheme   `json:"group_key"`
	SignerClientIDs []string                    `json:"sig_client_ids"`
	SignerKeys      []zcncrypto.SignatureScheme `json:"signer_keys"`
	T               int                         `json:"threshold"`
	N               int                         `json:"num_subkeys"`
}

func (msw *msWallet) UnmarshalJSON(data []byte) error {
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
func (msw *msWallet) Marshal() (string, error) {
	msws, err := json.Marshal(msw)
	if err != nil {
		return "", errors.New("invalid wallet")
	}
	return string(msws), nil
}

type MSVote interface {
	GetProposalID() string
	GetSignature() string
	GetTransferClientID() string
	GetTransferToClientID() string
	GetTransferAmount() string
}

type msVote struct {
	ProposalID string `json:"proposal_id"`

	// Client ID in transfer is that of the multi-sig wallet, not the signer.
	Transfer msTransfer `json:"transfer"`

	Signature string `json:"signature"`
}

func (m *msVote) GetProposalID() string {
	return m.ProposalID
}

func (m *msVote) GetTransferClientID() string {
	return m.Transfer.ClientID
}

func (m *msVote) GetTransferToClientID() string {
	return m.Transfer.ToClientID
}

func (m *msVote) GetTransferAmount() string {
	return strconv.FormatUint(m.Transfer.Amount, 10)
}

func (m *msVote) GetSignature() string {
	return m.Signature
}

//msTransfer - a data structure to hold state transfer from one client to another
type msTransfer struct {
	ClientID   string `json:"from"`
	ToClientID string `json:"to"`
	Amount     uint64 `json:"amount"`
}

//GetMultisigVotePayload given a multisig vote as a string, makes a multisig vote payload to register
func GetMultisigVotePayload(msvstr string) (MSVote, error) {
	var msv msVote
	err := json.Unmarshal([]byte(msvstr), &msv)
	if err != nil {
		return nil, err
	}

	return &msv, nil
}

// CreateMSVote create a vote for multisig
func CreateMSVote(proposal, grpClientID, signerWalletstr, toClientID string, tokenStr string) (string, error) {
	if proposal == "" || grpClientID == "" || toClientID == "" || signerWalletstr == "" {
		return "", errors.New("proposal or groupClient or signer wallet or toClientID cannot be empty")
	}

	token, err := strconv.ParseUint(tokenStr, 10, 64)
	if err != nil {
		return "", err
	}

	if token < 1 {
		return "", errors.New("token cannot be less than 1")
	}

	signerWallet, err := getWallet(signerWalletstr)
	if err != nil {
		return "", err
	}

	//Note: Is this honored by multisig sc?
	transfer := msTransfer{
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

	vote := msVote{
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
