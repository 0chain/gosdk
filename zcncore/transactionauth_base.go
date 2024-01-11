package zcncore

import (
	"encoding/json"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/node"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/zcncrypto"
)

type TransactionWithAuth struct {
	t *Transaction
}

func (ta *TransactionWithAuth) Hash() string {
	return ta.t.txnHash
}

func (ta *TransactionWithAuth) SetTransactionNonce(txnNonce int64) error {
	return ta.t.SetTransactionNonce(txnNonce)
}

func newTransactionWithAuth(cb TransactionCallback, txnFee uint64, nonce int64) (*TransactionWithAuth, error) {
	ta := &TransactionWithAuth{}
	var err error
	ta.t, err = newTransaction(cb, txnFee, nonce)
	return ta, err
}

func (ta *TransactionWithAuth) getAuthorize() (*transaction.Transaction, error) {
	ta.t.txn.PublicKey = _config.wallet.ClientKey
	err := ta.t.txn.ComputeHashAndSign(SignFn)
	if err != nil {
		return nil, errors.Wrap(err, "signing error.")
	}

	jsonByte, err := json.Marshal(ta.t.txn)
	if err != nil {
		return nil, err
	}

	if sys.Authorize == nil {
		return nil, errors.New("not_initialized", "no authorize func is set, define it in native code and set in sys")
	}
	authorize, err := sys.Authorize(string(jsonByte))
	if err != nil {
		return nil, err
	}

	var txnResp transaction.Transaction
	err = json.Unmarshal([]byte(authorize), &txnResp)
	if err != nil {
		return nil, errors.Wrap(err, "invalid json on auth response.")
	}
	logging.Debug(txnResp)
	// Verify the signature on the result
	ok, err := txnResp.VerifyTransaction(sys.VerifyWith)
	if err != nil {
		logging.Error("verification failed for txn from auth", err.Error())
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
	ta.t.completeTxn(status, out, err) //nolint
}

func verifyFn(signature, msgHash, publicKey string) (bool, error) {
	v := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := v.SetPublicKey(publicKey)
	if err != nil {
		return false, err
	}

	ok, err := v.Verify(signature, msgHash)
	if err != nil || !ok {
		return false, errors.New("", `{"error": "signature_mismatch"}`)
	}
	return true, nil
}

func (ta *TransactionWithAuth) sign(otherSig string) error {
	ta.t.txn.ComputeHashData()

	sig, err := AddSignature(_config.wallet.Keys[0].PrivateKey, otherSig, ta.t.txn.Hash)
	if err != nil {
		return err
	}
	ta.t.txn.Signature = sig
	return nil
}

func (ta *TransactionWithAuth) submitTxn() {
	nonce := ta.t.txn.TransactionNonce
	if nonce < 1 {
		nonce = node.Cache.GetNextNonce(ta.t.txn.ClientID)
	} else {
		node.Cache.Set(ta.t.txn.ClientID, nonce)
	}
	ta.t.txn.TransactionNonce = nonce
	authTxn, err := ta.getAuthorize()
	if err != nil {
		logging.Error("get auth error for send.", err.Error())
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
		nonce := ta.t.txn.TransactionNonce
		if nonce < 1 {
			nonce = node.Cache.GetNextNonce(ta.t.txn.ClientID)
		} else {
			node.Cache.Set(ta.t.txn.ClientID, nonce)
		}
		ta.t.txn.TransactionNonce = nonce
		err = ta.t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
		if err != nil {
			return
		}
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) SetTransactionCallback(cb TransactionCallback) error {
	return ta.t.SetTransactionCallback(cb)
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

func (ta *TransactionWithAuth) Output() []byte {
	return []byte(ta.t.txnOut)
}

// GetTransactionNonce returns nonce
func (ta *TransactionWithAuth) GetTransactionNonce() int64 {
	return ta.t.txn.TransactionNonce
}

// ========================================================================== //
//                                vesting pool                                //
// ========================================================================== //

func (ta *TransactionWithAuth) VestingTrigger(poolID string) (err error) {
	err = ta.t.vestingPoolTxn(transaction.VESTING_TRIGGER, poolID, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) VestingStop(sr *VestingStopRequest) (err error) {
	err = ta.t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_STOP, sr, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) VestingUnlock(poolID string) (err error) {

	err = ta.t.vestingPoolTxn(transaction.VESTING_UNLOCK, poolID, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) VestingDelete(poolID string) (err error) {
	err = ta.t.vestingPoolTxn(transaction.VESTING_DELETE, poolID, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

//
// miner sc
//

// RegisterMultiSig register a multisig wallet with the SC.
func (ta *TransactionWithAuth) RegisterMultiSig(walletstr string, mswallet string) error {
	return errors.New("", "not implemented")
}

//
// Storage SC
//
