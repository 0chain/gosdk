package zcncore

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
)

type TransactionWithAuth struct {
	t *Transaction
}

func newTransactionWithAuth(cb TransactionCallback, txnFee int64) (*TransactionWithAuth, error) {
	ta := &TransactionWithAuth{}
	var err error
	ta.t, err = newTransaction(cb, txnFee)
	return ta, err
}

func (ta *TransactionWithAuth) getAuthorize() (*transaction.Transaction, error) {
	ta.t.txn.PublicKey = _config.wallet.Keys[0].PublicKey
	err := ta.t.txn.ComputeHashAndSign(signFn)
	if err != nil {
		return nil, fmt.Errorf("signing error. %v", err.Error())
	}
	req, err := util.NewHTTPPostRequest(_config.authUrl+"/transaction", ta.t.txn)
	if err != nil {
		return nil, fmt.Errorf("new post request failed for auth %v", err.Error())
	}
	res, err := req.Post()
	if err != nil {
		return nil, errNetwork
	}
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return nil, errUserRejected
		}
		return nil, fmt.Errorf("auth error: %v. %v", res.Status, res.Body)
	}
	var txnResp transaction.Transaction
	err = json.Unmarshal([]byte(res.Body), &txnResp)
	if err != nil {
		return nil, fmt.Errorf("invalid json on auth response. %v", err)
	}
	Logger.Debug(txnResp)
	// Verify the signature on the result
	ok, err := txnResp.VerifyTransaction(verifyFn)
	if err != nil {
		Logger.Error("verification failed for txn from auth", err.Error())
		return nil, errAuthVerifyFailed
	}
	if !ok {
		ta.completeTxn(StatusAuthVerifyFailed, "", errAuthVerifyFailed)
		return nil, errAuthVerifyFailed
	}
	return &txnResp, nil
}

func (ta *TransactionWithAuth) completeTxn(status int, out string, err error) {
	// do error code translation
	if status != StatusSuccess {
		switch err {
		case errNetwork:
			status = StatusNetworkError
		case errUserRejected:
			status = StatusRejectedByUser
		case errAuthVerifyFailed:
			status = StatusAuthVerifyFailed
		case errAuthTimeout:
			status = StatusAuthTimeout
		}
	}
	ta.t.completeTxn(status, out, err)
}

func (ta *TransactionWithAuth) SetTransactionCallback(cb TransactionCallback) error {
	return ta.t.SetTransactionCallback(cb)
}

func (ta *TransactionWithAuth) SetTransactionFee(txnFee int64) error {
	return ta.t.SetTransactionFee(txnFee)
}

func verifyFn(signature, msgHash, publicKey string) (bool, error) {
	v := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	v.SetPublicKey(publicKey)
	ok, err := v.Verify(signature, msgHash)
	if err != nil || ok == false {
		return false, fmt.Errorf(`{"error": "signature_mismatch"}`)
	}
	return true, nil
}

func (ta *TransactionWithAuth) sign(otherSig string) error {
	ta.t.txn.ComputeHashData()
	sig := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	sig.SetPrivateKey(_config.wallet.Keys[0].PrivateKey)
	var err error
	ta.t.txn.Signature, err = sig.Add(otherSig, ta.t.txn.Hash)
	return err
}

func (ta *TransactionWithAuth) submitTxn() {
	authTxn, err := ta.getAuthorize()
	if err != nil {
		Logger.Error("get auth error for send.", err.Error())
		ta.completeTxn(StatusAuthError, "", err)
		return
	}
	// Authorized by user. Give callback to app.
	if ta.t.txnCb != nil {
		ta.t.txnCb.OnAuthComplete(ta.t, StatusSuccess)
	}
	// Use the timestamp from auth and sign
	ta.t.txn.CreationDate = authTxn.CreationDate
	err = ta.sign(authTxn.Signature)
	if err != nil {
		ta.completeTxn(StatusError, "", errAddSignature)
	}
	ta.t.submitTxn()
}

func (ta *TransactionWithAuth) Send(toClientID string, val int64, desc string) error {
	go func() {
		ta.t.txn.TransactionType = transaction.TxnTypeSend
		ta.t.txn.ToClientID = toClientID
		ta.t.txn.Value = val
		ta.t.txn.TransactionData = desc
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) StoreData(data string) error {
	go func() {
		ta.t.txn.TransactionType = transaction.TxnTypeData
		ta.t.txn.TransactionData = data
		ta.submitTxn()
	}()
	return nil
}

// ExecuteFaucetSCWallet impements the Faucet Smart contract for a given wallet
func (ta *TransactionWithAuth) ExecuteFaucetSCWallet(walletStr string, methodName string, input []byte) error {
	w, err := ta.t.createFaucetSCWallet(walletStr, methodName, input)
	if err != nil {
		return err
	}
	go func() {
		ta.t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) ExecuteSmartContract(address, methodName, input string, val int64) error {
	err := ta.t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) SetTransactionHash(hash string) error {
	return ta.t.SetTransactionHash(hash)
}

func (ta *TransactionWithAuth) GetTransactionHash() string {
	return ta.t.GetTransactionHash()
}

func (ta *TransactionWithAuth) Verify() error {
	return ta.t.Verify()
}

func (ta *TransactionWithAuth) GetVerifyOutput() string {
	return ta.t.GetVerifyOutput()
}

func (ta *TransactionWithAuth) GetTransactionError() string {
	return ta.t.GetTransactionError()
}

func (ta *TransactionWithAuth) GetVerifyError() string {
	return ta.t.GetVerifyError()
}

func (ta *TransactionWithAuth) LockTokens(val int64, durationHr int64, durationMin int) error {
	err := ta.t.createLockTokensTxn(val, durationHr, durationMin)
	if err != nil {
		Logger.Error(err)
		return err
	}
	go func() {
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) UnlockTokens(poolID string) error {
	err := ta.t.createUnlockTokensTxn(poolID)
	if err != nil {
		Logger.Error(err)
		return err
	}
	go func() {
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) Stake(clientID string, val int64) error {
	err := ta.t.createStakeTxn(clientID, val)
	if err != nil {
		return err
	}
	go func() {
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) DeleteStake(clientID, poolID string) error {
	err := ta.t.createDeleteStakeTxn(clientID, poolID)
	if err != nil {
		return err
	}
	go func() {
		ta.submitTxn()
	}()
	return nil
}

//RegisterMultiSig register a multisig wallet with the SC.
func (ta *TransactionWithAuth) RegisterMultiSig(walletstr string, mswallet string) error {
	return fmt.Errorf("not implemented")
}
