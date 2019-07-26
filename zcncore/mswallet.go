package zcncore

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

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

//MSTransfer - a data structure to hold state transfer from one client to another
type MSTransfer struct {
	ClientID   string `json:"from"`
	ToClientID string `json:"to"`
	Amount     int64  `json:"amount"`
}

// Marshal returns json string
func (msw *MSWallet) Marshal() (string, error) {
	msws, err := json.Marshal(msw)
	if err != nil {
		return "", fmt.Errorf("Invalid Wallet")
	}
	return string(msws), nil
}

//MSVoteCallback callback definition multisig Vote function
type MSVoteCallback interface {
	OnVoteComplete(status int, proposal string, err string)
}

// CreateMSWallet returns multisig wallet information
func CreateMSWallet(t, n int) (string, string, []string, error) {
	id := 0
	if _config.chain.SignatureScheme != "bls0chain" {
		return "", "", nil, fmt.Errorf("encryption scheme for this blockchain is not bls0chain")

	}

	groupKey := zcncrypto.NewBLS0ChainScheme()
	wallet, err := groupKey.GenerateKeys()
	if err != nil {
		return "", "", nil, fmt.Errorf("%s", err.Error())
	}

	Logger.Info(fmt.Sprintf("Wallet id: %s", wallet.ClientKey))

	groupClientID := GetClientID(groupKey.GetPublicKey())
	//Code modified to directly use BLS0ChainThresholdScheme
	signerKeys, err := zcncrypto.BLS0GenerateThresholdKeyShares(t, n, groupKey)

	if err != nil {
		return "", "", nil, fmt.Errorf("Err in generateThresholdKeyShares %s", err.Error())
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
		cb.OnWalletCreateComplete(StatusError, walletString, fmt.Sprintf("%s", err.Error()))
	}

	//We do not want to send private key to blockchain
	w.Keys[0].PrivateKey = ""
	err = RegisterToMiners(&w, cb)
	if err != nil {
		cb.OnWalletCreateComplete(StatusError, "", fmt.Sprintf("%s", err.Error()))
	}

}

//CreateMSVote create a vote for multisig
func CreateMSVote(proposal, grpClientID, signerWalletstr, toClientID string, token int64) (string, error) {

	if proposal == "" || grpClientID == "" || toClientID == "" || signerWalletstr == "" {
		return "", fmt.Errorf("proposal or groupClient or signer wallet or toClientID cannot be empty")
	}

	if token < 1 {
		return "", fmt.Errorf("Token cannot be less than 1")
	}

	signerWallet, err := GetWallet(signerWalletstr)
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

	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	sigScheme.SetPrivateKey(signerWallet.Keys[0].PrivateKey)
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
		signerThresholdIDs = append(signerThresholdIDs, scheme.Ids)
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
