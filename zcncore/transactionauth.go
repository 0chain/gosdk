package zcncore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
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
	ta.t.txn.PublicKey = _config.wallet.Keys[0].PublicKey
	err := ta.t.txn.ComputeHashAndSign(SignFn)
	if err != nil {
		return nil, errors.Wrap(err, "signing error.")
	}
	req, err := util.NewHTTPPostRequest(_config.authUrl+"/transaction", ta.t.txn)
	if err != nil {
		return nil, errors.Wrap(err, "new post request failed for auth")
	}
	res, err := req.Post()
	if err != nil {
		return nil, errNetwork
	}
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return nil, errUserRejected
		}
		return nil, errors.New(strconv.Itoa(res.StatusCode), fmt.Sprintf("auth error: %v. %v", res.Status, res.Body))
	}
	var txnResp transaction.Transaction
	err = json.Unmarshal([]byte(res.Body), &txnResp)
	if err != nil {
		return nil, errors.Wrap(err, "invalid json on auth response.")
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

func (ta *TransactionWithAuth) SetTransactionFee(txnFee uint64) error {
	return ta.t.SetTransactionFee(txnFee)
}

func verifyFn(signature, msgHash, publicKey string) (bool, error) {
	v := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	v.SetPublicKey(publicKey)
	ok, err := v.Verify(signature, msgHash)
	if err != nil || !ok {
		return false, errors.New("", `{"error": "signature_mismatch"}`)
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
	nonce := ta.t.txn.TransactionNonce
	if nonce < 1 {
		nonce = transaction.Cache.GetNextNonce(ta.t.txn.ClientID)
	} else {
		transaction.Cache.Set(ta.t.txn.ClientID, nonce)
	}
	ta.t.txn.TransactionNonce = nonce

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

func (ta *TransactionWithAuth) Send(toClientID string, val uint64, desc string) error {
	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		ta.t.txn.TransactionType = transaction.TxnTypeSend
		ta.t.txn.ToClientID = toClientID
		ta.t.txn.Value = val
		ta.t.txn.TransactionData = string(txnData)
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
		nonce := ta.t.txn.TransactionNonce
		if nonce < 1 {
			nonce = transaction.Cache.GetNextNonce(ta.t.txn.ClientID)
		} else {
			transaction.Cache.Set(ta.t.txn.ClientID, nonce)
		}
		ta.t.txn.TransactionNonce = nonce
		ta.t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
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

func (ta *TransactionWithAuth) GetVerifyConfirmationStatus() ConfirmationStatus {
	return ta.t.GetVerifyConfirmationStatus()
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
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) VestingStop(sr *VestingStopRequest) (err error) {
	err = ta.t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_STOP, sr, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) VestingUnlock(poolID string) (err error) {

	err = ta.t.vestingPoolTxn(transaction.VESTING_UNLOCK, poolID, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) VestingAdd(ar *VestingAddRequest,
	value uint64) (err error) {

	err = ta.t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_ADD, ar, value)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) VestingDelete(poolID string) (err error) {
	err = ta.t.vestingPoolTxn(transaction.VESTING_DELETE, poolID, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) VestingUpdateConfig(
	ip *InputMap,
) (err error) {
	err = ta.t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// faucet smart contract

func (ta *TransactionWithAuth) FaucetUpdateConfig(
	ip *InputMap,
) (err error) {
	err = ta.t.createSmartContractTxn(FaucetSmartContractAddress,
		transaction.FAUCETSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

//
// miner sc
//

func (ta *TransactionWithAuth) MinerSCMinerSettings(info *MinerSCMinerInfo) (
	err error) {

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_SETTINGS, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCSharderSettings(info *MinerSCMinerInfo) (
	err error) {

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_SETTINGS, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCDeleteMiner(info *MinerSCMinerInfo) (
	err error) {

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_DELETE, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCDeleteSharder(info *MinerSCMinerInfo) (
	err error) {

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_DELETE, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCCollectReward(providerId, poolId string, providerType Provider) error {
	pr := &SCCollectReward{
		ProviderId:   providerId,
		PoolId:       poolId,
		ProviderType: providerType,
	}
	err := ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_COLLECT_REWARD, pr, 0)
	if err != nil {
		Logger.Error(err)
		return err
	}
	go ta.submitTxn()
	return err
}

func (ta *TransactionWithAuth) MinerSCLock(minerID string, lock uint64) (err error) {

	var mscl MinerSCLock
	mscl.ID = minerID

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, &mscl, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCUnlock(nodeID, poolID string) (
	err error) {

	var mscul MinerSCUnlock
	mscul.ID = nodeID
	mscul.PoolID = poolID

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UNLOCK, &mscul, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerScUpdateConfig(
	ip *InputMap,
) (err error) {
	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerScUpdateGlobals(
	ip *InputMap,
) (err error) {
	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_GLOBALS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

//RegisterMultiSig register a multisig wallet with the SC.
func (ta *TransactionWithAuth) RegisterMultiSig(walletstr string, mswallet string) error {
	return errors.New("", "not implemented")
}

//
// Storage SC
//

func (ta *TransactionWithAuth) StorageSCCollectReward(providerId, poolId string, providerType Provider) error {
	pr := &SCCollectReward{
		ProviderId:   providerId,
		PoolId:       poolId,
		ProviderType: providerType,
	}
	err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_COLLECT_REWARD, pr, 0)
	if err != nil {
		Logger.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return err
}

func (ta *TransactionWithAuth) StorageScUpdateConfig(
	ip *InputMap,
) (err error) {
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// FinalizeAllocation transaction.
func (ta *TransactionWithAuth) FinalizeAllocation(allocID string, fee uint64) (
	err error) {

	type finiRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_FINALIZE_ALLOCATION, &finiRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// CancelAllocation transaction.
func (ta *TransactionWithAuth) CancelAllocation(allocID string, fee uint64) (
	err error) {

	type cancelRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CANCEL_ALLOCATION, &cancelRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// CreateAllocation transaction.
func (ta *TransactionWithAuth) CreateAllocation(car *CreateAllocationRequest,
	lock uint64, fee uint64) (err error) {

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_ALLOCATION, car, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// CreateReadPool for current user.
func (ta *TransactionWithAuth) CreateReadPool(fee uint64) (err error) {

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_READ_POOL, nil, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// ReadPoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (ta *TransactionWithAuth) ReadPoolLock(allocID, blobberID string,
	duration int64, lock, fee uint64) (err error) {

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_LOCK, &lr, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// ReadPoolUnlock for current user and given pool.
func (ta *TransactionWithAuth) ReadPoolUnlock(poolID string, fee uint64) (
	err error) {

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_UNLOCK, &unlockRequest{
			PoolID: poolID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// StakePoolLock used to lock tokens in a stake pool of a blobber.
func (ta *TransactionWithAuth) StakePoolLock(blobberID string,
	lock, fee uint64) (err error) {

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_LOCK, &spr, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// StakePoolUnlock by blobberID and poolID.
func (ta *TransactionWithAuth) StakePoolUnlock(blobberID, poolID string,
	fee uint64) (err error) {

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
		PoolID    string `json:"pool_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID
	spr.PoolID = poolID

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_UNLOCK, &spr, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// UpdateBlobberSettings update settings of a blobber.
func (ta *TransactionWithAuth) UpdateBlobberSettings(blob *Blobber, fee uint64) (
	err error) {

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, blob, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// UpdateValidatorSettings update settings of a validator.
func (ta *TransactionWithAuth) UpdateValidatorSettings(v *Validator, fee uint64) (
	err error) {

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_VALIDATOR_SETTINGS, v, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// UpdateAllocation transaction.
func (ta *TransactionWithAuth) UpdateAllocation(allocID string, sizeDiff int64,
	expirationDiff int64, lock, fee uint64) (err error) {

	type updateAllocationRequest struct {
		ID         string `json:"id"`              // allocation id
		Size       int64  `json:"size"`            // difference
		Expiration int64  `json:"expiration_date"` // difference
	}

	var uar updateAllocationRequest
	uar.ID = allocID
	uar.Size = sizeDiff
	uar.Expiration = expirationDiff

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_ALLOCATION, &uar, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// WritePoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (ta *TransactionWithAuth) WritePoolLock(allocID string, lock, fee uint64) (err error) {

	type lockRequest struct {
		AllocationID string `json:"allocation_id"`
	}

	var lr lockRequest
	lr.AllocationID = allocID

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_LOCK, &lr, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

// WritePoolUnlock for current user and given pool.
func (ta *TransactionWithAuth) WritePoolUnlock(allocID string, fee uint64) (
	err error) {

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_UNLOCK, &unlockRequest{
			PoolID: allocID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	ta.t.SetTransactionFee(fee)
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) ZCNSCUpdateGlobalConfig(ip *InputMap) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_GLOBAL_CONFIG, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) ZCNSCUpdateAuthorizerConfig(ip *AuthorizerNode) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_AUTHORIZER_CONFIG, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) ZCNSCAddAuthorizer(ip *AddAuthorizerPayload) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_ADD_AUTHORIZER, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go ta.submitTxn()
	return
}
