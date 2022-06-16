package zcncore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/block"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/sys"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// compiler time check
var (
	_ TransactionScheme = (*Transaction)(nil)
	_ TransactionScheme = (*TransactionWithAuth)(nil)
)

var (
	errNetwork          = errors.New("", "network error. host not reachable")
	errUserRejected     = errors.New("", "rejected by user")
	errAuthVerifyFailed = errors.New("", "verification failed for auth response")
	errAuthTimeout      = errors.New("", "auth timed out")
	errAddSignature     = errors.New("", "error adding signature")
)

// TransactionCallback needs to be implemented by the caller for transaction related APIs
type TransactionCallback interface {
	OnTransactionComplete(t *Transaction, status int)
	OnVerifyComplete(t *Transaction, status int)
	OnAuthComplete(t *Transaction, status int)
}

/*Confirmation - a data structure that provides the confirmation that a transaction is included into the block chain */
type confirmation struct {
	Version               string                   `json:"version"`
	Hash                  string                   `json:"hash"`
	BlockHash             string                   `json:"block_hash"`
	PreviousBlockHash     string                   `json:"previous_block_hash"`
	Transaction           *transaction.Transaction `json:"txn,omitempty"`
	CreationDate          int64                    `json:"creation_date,omitempty"`
	MinerID               string                   `json:"miner_id"`
	Round                 int64                    `json:"round"`
	Status                int                      `json:"transaction_status" msgpack:"sot"`
	RoundRandomSeed       int64                    `json:"round_random_seed"`
	StateChangesCount     int                      `json:"state_changes_count"`
	MerkleTreeRoot        string                   `json:"merkle_tree_root"`
	MerkleTreePath        *util.MTPath             `json:"merkle_tree_path"`
	ReceiptMerkleTreeRoot string                   `json:"receipt_merkle_tree_root"`
	ReceiptMerkleTreePath *util.MTPath             `json:"receipt_merkle_tree_path"`
}

type blockHeader struct {
	Version               string `json:"version,omitempty"`
	CreationDate          int64  `json:"creation_date,omitempty"`
	Hash                  string `json:"hash,omitempty"`
	MinerId               string `json:"miner_id,omitempty"`
	Round                 int64  `json:"round,omitempty"`
	RoundRandomSeed       int64  `json:"round_random_seed,omitempty"`
	StateChangesCount     int    `json:"state_changes_count"`
	MerkleTreeRoot        string `json:"merkle_tree_root,omitempty"`
	StateHash             string `json:"state_hash,omitempty"`
	ReceiptMerkleTreeRoot string `json:"receipt_merkle_tree_root,omitempty"`
	NumTxns               int64  `json:"num_txns,omitempty"`
}

func (bh *blockHeader) getCreationDate(defaultTime int64) int64 {
	if bh == nil {
		return defaultTime
	}

	return bh.CreationDate
}

type Transaction struct {
	txn                      *transaction.Transaction
	txnOut                   string
	txnHash                  string
	txnStatus                int
	txnError                 error
	txnCb                    TransactionCallback
	verifyStatus             int
	verifyConfirmationStatus ConfirmationStatus
	verifyOut                string
	verifyError              error
}

type SendTxnData struct {
	Note string `json:"note"`
}

// TransactionScheme implements few methods for block chain.
//
// Note: to be buildable on MacOSX all arguments should have names.
type TransactionScheme interface {
	// SetTransactionCallback implements storing the callback
	// used to call after the transaction or verification is completed
	SetTransactionCallback(cb TransactionCallback) error
	// Send implements sending token to a given clientid
	Send(toClientID string, val int64, desc string) error
	// StoreData implements store the data to blockchain
	StoreData(data string) error
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input interface{}, val int64) error
	// ExecuteFaucetSCWallet implements the `Faucet Smart contract` for a given wallet
	ExecuteFaucetSCWallet(walletStr string, methodName string, input []byte) error
	// GetTransactionHash implements retrieval of hash of the submitted transaction
	GetTransactionHash() string
	//RegisterMultiSig registers a group wallet and subwallets with MultisigSC
	RegisterMultiSig(walletstr, mswallet string) error
	// SetTransactionHash implements verify a previous transaction status
	SetTransactionHash(hash string) error
	// SetTransactionFee implements method to set the transaction fee
	SetTransactionFee(txnFee int64) error
	// SetTransactionNonce implements method to set the transaction nonce
	SetTransactionNonce(txnNonce int64) error
	// Verify implements verify the transaction
	Verify() error
	// GetVerifyConfirmationStatus implements the verification status from sharders
	GetVerifyConfirmationStatus() ConfirmationStatus
	// GetVerifyOutput implements the verification output from sharders
	GetVerifyOutput() string
	// GetTransactionError implements error string in case of transaction failure
	GetTransactionError() string
	// GetVerifyError implements error string in case of verify failure error
	GetVerifyError() string
	// GetTransactionNonce returns nonce
	GetTransactionNonce() int64

	// Output of transaction.
	Output() []byte

	// Hash Transaction status regardless of status
	Hash() string

	// Vesting SC

	VestingTrigger(poolID string) error
	VestingStop(sr *VestingStopRequest) error
	VestingUnlock(poolID string) error
	VestingAdd(ar *VestingAddRequest, value int64) error
	VestingDelete(poolID string) error
	VestingUpdateConfig(*InputMap) error

	// Miner SC

	MinerSCCollectReward(string, string, Provider) error
	MinerSCMinerSettings(*MinerSCMinerInfo) error
	MinerSCSharderSettings(*MinerSCMinerInfo) error
	MinerSCLock(minerID string, lock int64) error
	MinerSCUnlock(minerID, poolID string) error
	MinerScUpdateConfig(*InputMap) error
	MinerScUpdateGlobals(*InputMap) error
	MinerSCDeleteMiner(*MinerSCMinerInfo) error
	MinerSCDeleteSharder(*MinerSCMinerInfo) error

	// Storage SC

	StorageSCCollectReward(string, string, Provider) error
	FinalizeAllocation(allocID string, fee int64) error
	CancelAllocation(allocID string, fee int64) error
	CreateAllocation(car *CreateAllocationRequest, lock, fee int64) error //
	CreateReadPool(fee int64) error
	ReadPoolLock(allocID string, blobberID string, duration int64, lock, fee int64) error
	ReadPoolUnlock(poolID string, fee int64) error
	StakePoolLock(blobberID string, lock, fee int64) error
	StakePoolUnlock(blobberID string, poolID string, fee int64) error
	UpdateBlobberSettings(blobber *Blobber, fee int64) error
	UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock, fee int64) error
	WritePoolLock(allocID string, blobberID string, duration int64, lock, fee int64) error
	WritePoolUnlock(poolID string, fee int64) error
	StorageScUpdateConfig(*InputMap) error

	// Faucet

	FaucetUpdateConfig(*InputMap) error

	// ZCNSC Common transactions

	// ZCNSCUpdateGlobalConfig updates global config
	ZCNSCUpdateGlobalConfig(*InputMap) error
	// ZCNSCUpdateAuthorizerConfig updates authorizer config by ID
	ZCNSCUpdateAuthorizerConfig(*AuthorizerNode) error
	// ZCNSCAddAuthorizer adds authorizer
	ZCNSCAddAuthorizer(*AddAuthorizerPayload) error
}

func Sign(hash string) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(_config.wallet.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}

func SignWith0Wallet(hash string, w *zcncrypto.Wallet) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(w.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}

func signFn(hash string) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	sigScheme.SetPrivateKey(_config.wallet.Keys[0].PrivateKey)
	return sigScheme.Sign(hash)
}

func signWithWallet(hash string, wi interface{}) (string, error) {
	w, ok := wi.(*zcncrypto.Wallet)

	if !ok {
		fmt.Printf("Error in casting to wallet")
		return "", errors.New("", "error in casting to wallet")
	}
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	sigScheme.SetPrivateKey(w.Keys[0].PrivateKey)
	return sigScheme.Sign(hash)
}

func txnTypeString(t int) string {
	switch t {
	case transaction.TxnTypeSend:
		return "send"
	case transaction.TxnTypeLockIn:
		return "lock-in"
	case transaction.TxnTypeData:
		return "data"
	case transaction.TxnTypeSmartContract:
		return "smart contract"
	default:
		return "unknown"
	}
}

func (t *Transaction) Output() []byte {
	return []byte(t.txnOut)
}

func (t *Transaction) Hash() string {
	return t.txn.Hash
}

func (t *Transaction) completeTxn(status int, out string, err error) {
	t.txnStatus = status
	t.txnOut = out
	t.txnError = err
	if t.txnCb != nil {
		t.txnCb.OnTransactionComplete(t, t.txnStatus)
	}
}

func (t *Transaction) completeVerify(status int, out string, err error) {
	t.completeVerifyWithConStatus(status, 0, out, err)
}

func (t *Transaction) completeVerifyWithConStatus(status int, conStatus ConfirmationStatus, out string, err error) {
	t.verifyStatus = status
	t.verifyConfirmationStatus = conStatus
	t.verifyOut = out
	t.verifyError = err
	if status == StatusError {
		transaction.Cache.Evict(t.txn.ClientID)
	}
	if t.txnCb != nil {
		t.txnCb.OnVerifyComplete(t, t.verifyStatus)
	}
}

type getNonceCallBack struct {
	nonceCh chan int64
	err     error
}

func (g getNonceCallBack) OnNonceAvailable(status int, nonce int64, info string) {
	if status != StatusSuccess {
		g.err = errors.New("get_nonce", "failed respond nonce")
	}

	g.nonceCh <- nonce
}

func (t *Transaction) setNonceAndSubmit() {
	t.setNonce()
	t.submitTxn()
}

func (t *Transaction) setNonce() {
	nonce := t.txn.TransactionNonce
	if nonce < 1 {
		nonce = transaction.Cache.GetNextNonce(t.txn.ClientID)
	} else {
		transaction.Cache.Set(t.txn.ClientID, nonce)
	}
	t.txn.TransactionNonce = nonce
}

func (t *Transaction) submitTxn() {
	// Clear the status, in case transaction object reused
	t.txnStatus = StatusUnknown
	t.txnOut = ""
	t.txnError = nil

	// If Signature is not passed compute signature
	if t.txn.Signature == "" {
		err := t.txn.ComputeHashAndSign(signFn)
		if err != nil {
			t.completeTxn(StatusError, "", err)
			transaction.Cache.Evict(t.txn.ClientID)
			return
		}
	}

	result := make(chan *util.PostResponse, len(_config.chain.Miners))
	defer close(result)
	var tSuccessRsp string
	var tFailureRsp string
	randomMiners := util.GetRandom(_config.chain.Miners, getMinMinersSubmit())
	for _, miner := range randomMiners {
		go func(minerurl string) {
			url := minerurl + PUT_TRANSACTION
			Logger.Info("Submitting ", txnTypeString(t.txn.TransactionType), " transaction to ", minerurl, " with JSON ", string(t.txn.DebugJSON()))
			req, err := util.NewHTTPPostRequest(url, t.txn)
			if err != nil {
				Logger.Error(minerurl, " new post request failed. ", err.Error())
				return
			}
			res, err := req.Post()
			if err != nil {
				Logger.Error(minerurl, " submit transaction error. ", err.Error())
			}
			result <- res
		}(miner)
	}
	consensus := float32(0)
	for range randomMiners {
		rsp := <-result
		Logger.Debug(rsp.Url, rsp.Status)
		if rsp.StatusCode == http.StatusOK {
			consensus++
			tSuccessRsp = rsp.Body
		} else {
			Logger.Error(rsp.Body)
			tFailureRsp = rsp.Body
		}

	}
	rate := consensus * 100 / float32(len(randomMiners))
	if rate < consensusThresh {
		t.completeTxn(StatusError, "", fmt.Errorf("submit transaction failed. %s", tFailureRsp))
		transaction.Cache.Evict(t.txn.ClientID)
		return
	}
	sys.Sleep(3 * time.Second)
	t.completeTxn(StatusSuccess, tSuccessRsp, nil)
}

func newTransaction(cb TransactionCallback, txnFee int64, nonce int64) (*Transaction, error) {
	t := &Transaction{}
	t.txn = transaction.NewTransactionEntity(_config.wallet.ClientID, _config.chain.ChainID, _config.wallet.ClientKey, nonce)
	t.txnStatus, t.verifyStatus = StatusUnknown, StatusUnknown
	t.txnCb = cb
	t.txn.TransactionFee = txnFee
	t.txn.TransactionNonce = nonce
	return t, nil
}

// NewTransaction allocation new generic transaction object for any operation
func NewTransaction(cb TransactionCallback, txnFee int64, nonce int64) (TransactionScheme, error) {
	err := CheckConfig()
	if err != nil {
		return nil, err
	}
	if _config.isSplitWallet {
		if _config.authUrl == "" {
			return nil, errors.New("", "auth url not set")
		}
		Logger.Info("New transaction interface with auth")
		return newTransactionWithAuth(cb, txnFee, nonce)
	}
	Logger.Info("New transaction interface")
	t, err := newTransaction(cb, txnFee, nonce)
	return t, err
}

func (t *Transaction) SetTransactionCallback(cb TransactionCallback) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction hash.")
	}
	t.txnCb = cb
	return nil
}

func (t *Transaction) SetTransactionFee(txnFee int64) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction fee.")
	}
	t.txn.TransactionFee = txnFee
	return nil
}
func (t *Transaction) SetTransactionNonce(txnNonce int64) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction fee.")
	}
	t.txn.TransactionNonce = txnNonce
	return nil
}

func (t *Transaction) Send(toClientID string, val int64, desc string) error {
	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSend
		t.txn.ToClientID = toClientID
		t.txn.Value = val
		t.txn.TransactionData = string(txnData)
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) SendWithSignatureHash(toClientID string, val int64, desc string, sig string, CreationDate int64, hash string) error {
	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSend
		t.txn.ToClientID = toClientID
		t.txn.Value = val
		t.txn.Hash = hash
		t.txn.TransactionData = string(txnData)
		t.txn.Signature = sig
		t.txn.CreationDate = CreationDate
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) StoreData(data string) error {
	go func() {
		t.txn.TransactionType = transaction.TxnTypeData
		t.txn.TransactionData = data
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) createSmartContractTxn(address, methodName string, input interface{}, value int64) error {
	sn := transaction.SmartContractTxnData{Name: methodName, InputArgs: input}
	snBytes, err := json.Marshal(sn)
	if err != nil {
		return errors.Wrap(err, "create smart contract failed due to invalid data.")
	}
	t.txn.TransactionType = transaction.TxnTypeSmartContract
	t.txn.ToClientID = address
	t.txn.TransactionData = string(snBytes)
	t.txn.Value = value
	return nil
}

func (t *Transaction) createFaucetSCWallet(walletStr string, methodName string, input []byte) (*zcncrypto.Wallet, error) {
	w, err := GetWallet(walletStr)
	if err != nil {
		fmt.Printf("Error while parsing the wallet. %v\n", err)
		return nil, err
	}
	err = t.createSmartContractTxn(FaucetSmartContractAddress, methodName, input, 0)
	if err != nil {
		return nil, err
	}
	return w, nil
}

// ExecuteFaucetSCWallet implements the Faucet Smart contract for a given wallet
func (t *Transaction) ExecuteFaucetSCWallet(walletStr string, methodName string, input []byte) error {
	w, err := t.createFaucetSCWallet(walletStr, methodName, input)
	if err != nil {
		return err
	}
	go func() {
		nonce := t.txn.TransactionNonce
		if nonce < 1 {
			nonce = transaction.Cache.GetNextNonce(t.txn.ClientID)
		} else {
			transaction.Cache.Set(t.txn.ClientID, nonce)
		}
		t.txn.TransactionNonce = nonce
		t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
		fmt.Printf("submitted transaction\n")
		t.submitTxn()
	}()
	return nil
}

func (t *Transaction) ExecuteSmartContract(address, methodName string, input interface{}, val int64) error {
	err := t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) SetTransactionHash(hash string) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction hash.")
	}
	t.txnHash = hash
	return nil
}

func (t *Transaction) GetTransactionHash() string {
	if t.txnHash != "" {
		return t.txnHash
	}
	if t.txnStatus != StatusSuccess {
		return ""
	}
	var txnout map[string]json.RawMessage
	err := json.Unmarshal([]byte(t.txnOut), &txnout)
	if err != nil {
		fmt.Println("Error in parsing", err)
	}
	var entity map[string]interface{}
	err = json.Unmarshal(txnout["entity"], &entity)
	if err != nil {
		Logger.Error("json unmarshal error on GetTransactionHash()")
		return t.txnHash
	}
	if hash, ok := entity["hash"].(string); ok {
		t.txnHash = hash
	}
	return t.txnHash
}

func queryFromSharders(numSharders int, query string,
	result chan *util.GetResponse) {

	queryFromShardersContext(context.Background(), numSharders, query, result)
}

func queryFromShardersContext(ctx context.Context, numSharders int,
	query string, result chan *util.GetResponse) {

	for _, sharder := range util.Shuffle(_config.chain.Sharders) {
		go func(sharderurl string) {
			Logger.Info("Query from ", sharderurl+query)
			url := fmt.Sprintf("%v%v", sharderurl, query)
			req, err := util.NewHTTPGetRequestContext(ctx, url)
			if err != nil {
				Logger.Error(sharderurl, " new get request failed. ", err.Error())
				return
			}
			res, err := req.Get()
			if err != nil {
				Logger.Error(sharderurl, " get error. ", err.Error())
			}
			result <- res
		}(sharder)
	}
}

func getBlockHeaderFromTransactionConfirmation(txnHash string, cfmBlock map[string]json.RawMessage) (*blockHeader, error) {
	block := &blockHeader{}
	if cfmBytes, ok := cfmBlock["confirmation"]; ok {
		var cfm confirmation
		err := json.Unmarshal(cfmBytes, &cfm)
		if err != nil {
			return nil, errors.Wrap(err, "txn confirmation parse error.")
		}
		if cfm.Transaction == nil {
			return nil, fmt.Errorf("missing transaction %s in block confirmation", txnHash)
		}
		if txnHash != cfm.Transaction.Hash {
			return nil, fmt.Errorf("invalid transaction hash. Expected: %s. Received: %s", txnHash, cfm.Transaction.Hash)
		}
		if !util.VerifyMerklePath(cfm.Transaction.Hash, cfm.MerkleTreePath, cfm.MerkleTreeRoot) {
			return nil, errors.New("", "txn merkle validation failed.")
		}
		txnRcpt := transaction.NewTransactionReceipt(cfm.Transaction)
		if !util.VerifyMerklePath(txnRcpt.GetHash(), cfm.ReceiptMerkleTreePath, cfm.ReceiptMerkleTreeRoot) {
			return nil, errors.New("", "txn receipt cmerkle validation failed.")
		}
		prevBlockHash := cfm.PreviousBlockHash
		block.MinerId = cfm.MinerID
		block.Hash = cfm.BlockHash
		block.CreationDate = cfm.CreationDate
		block.Round = cfm.Round
		block.RoundRandomSeed = cfm.RoundRandomSeed
		block.StateChangesCount = cfm.StateChangesCount
		block.MerkleTreeRoot = cfm.MerkleTreeRoot
		block.ReceiptMerkleTreeRoot = cfm.ReceiptMerkleTreeRoot
		// Verify the block
		if isBlockExtends(prevBlockHash, block) {
			return block, nil
		}

		return nil, errors.New("", "block hash verification failed in confirmation")
	}
	return nil, errors.New("", "txn confirmation not found.")
}

func getTransactionConfirmation(numSharders int, txnHash string) (*blockHeader, map[string]json.RawMessage, *blockHeader, error) {
	result := make(chan *util.GetResponse)
	defer close(result)

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromSharders(numSharders, fmt.Sprintf("%v%v&content=lfb", TXN_VERIFY_URL, txnHash), result)

	maxConfirmation := int(0)
	txnConfirmations := make(map[string]int)
	var blockHdr *blockHeader
	var lfb blockHeader
	var confirmation map[string]json.RawMessage
	for i := 0; i < numSharders; i++ {
		rsp := <-result
		Logger.Debug(rsp.Url + " " + rsp.Status)
		Logger.Debug(rsp.Body)
		if rsp.StatusCode == http.StatusOK {
			var cfmLfb map[string]json.RawMessage
			err := json.Unmarshal([]byte(rsp.Body), &cfmLfb)
			if err != nil {
				Logger.Error("txn confirmation parse error", err)
				continue
			}
			bH, err := getBlockHeaderFromTransactionConfirmation(txnHash, cfmLfb)
			if err != nil {
				Logger.Error(err)
			}
			if err == nil {
				txnConfirmations[bH.Hash]++
				if txnConfirmations[bH.Hash] > maxConfirmation {
					maxConfirmation = txnConfirmations[bH.Hash]
					blockHdr = bH
					confirmation = cfmLfb
				}
			} else if lfbRaw, ok := cfmLfb["latest_finalized_block"]; ok {
				err := json.Unmarshal([]byte(lfbRaw), &lfb)
				if err != nil {
					Logger.Error("round info parse error.", err)
					continue
				}
			}
		}

	}
	if maxConfirmation == 0 {
		return nil, confirmation, &lfb, errors.New("", "transaction not found")
	}
	return blockHdr, confirmation, &lfb, nil
}

func GetLatestFinalized(ctx context.Context, numSharders int) (b *block.Header, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders, GET_LATEST_FINALIZED, result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	for i := 0; i < numSharders; i++ {
		var rsp = <-result

		Logger.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			Logger.Error(rsp.Body)
			continue
		}

		if err = json.Unmarshal([]byte(rsp.Body), &b); err != nil {
			Logger.Error("block parse error: ", err)
			err = nil
			continue
		}

		var h = encryption.FastHash([]byte(b.Hash))
		if roundConsensus[h]++; roundConsensus[h] > maxConsensus {
			maxConsensus = roundConsensus[h]
		}
	}

	if maxConsensus == 0 {
		return nil, errors.New("", "block info not found")
	}

	return
}

func GetLatestFinalizedMagicBlock(ctx context.Context, numSharders int) (m *block.MagicBlock, err error) {
	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders, GET_LATEST_FINALIZED_MAGIC_BLOCK, result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	type respObj struct {
		MagicBlock *block.MagicBlock `json:"magic_block"`
	}

	for i := 0; i < numSharders; i++ {
		var rsp = <-result

		Logger.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			Logger.Error(rsp.Body)
			continue
		}

		var respo respObj
		if err = json.Unmarshal([]byte(rsp.Body), &respo); err != nil {
			Logger.Error(" magic block parse error: ", err)
			err = nil
			continue
		}

		m = respo.MagicBlock
		var h = encryption.FastHash([]byte(respo.MagicBlock.Hash))
		if roundConsensus[h]++; roundConsensus[h] > maxConsensus {
			maxConsensus = roundConsensus[h]
		}
	}

	if maxConsensus == 0 {
		return nil, errors.New("", "magic block info not found")
	}

	return
}

func GetChainStats(ctx context.Context) (b *block.ChainStats, err error) {
	var result = make(chan *util.GetResponse, 1)
	defer close(result)

	var numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders, GET_CHAIN_STATS, result)
	var rsp *util.GetResponse
	for i := 0; i < numSharders; i++ {
		var x = <-result
		if x.StatusCode != http.StatusOK {
			continue
		}
		rsp = x
	}

	if rsp == nil {
		return nil, errors.New("http_request_failed", "Request failed with status not 200")
	}

	if err = json.Unmarshal([]byte(rsp.Body), &b); err != nil {
		return nil, err
	}
	return
}

func GetBlockByRound(ctx context.Context, numSharders int, round int64) (b *block.Block, err error) {

	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders,
		fmt.Sprintf("%sround=%d&content=full,header", GET_BLOCK_INFO, round),
		result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	type respObj struct {
		Block  *block.Block  `json:"block"`
		Header *block.Header `json:"header"`
	}

	for i := 0; i < numSharders; i++ {
		var rsp = <-result

		Logger.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			Logger.Error(rsp.Body)
			continue
		}

		var respo respObj
		if err = json.Unmarshal([]byte(rsp.Body), &respo); err != nil {
			Logger.Error("block parse error: ", err)
			err = nil
			continue
		}

		if respo.Block == nil {
			Logger.Debug(rsp.Url, "no block in response:", rsp.Body)
			continue
		}

		if respo.Header == nil {
			Logger.Debug(rsp.Url, "no block header in response:", rsp.Body)
			continue
		}

		if respo.Header.Hash != string(respo.Block.Hash) {
			Logger.Debug(rsp.Url, "header and block hash mismatch:", rsp.Body)
			continue
		}

		b = respo.Block
		b.Header = respo.Header

		var h = encryption.FastHash([]byte(b.Hash))
		if roundConsensus[h]++; roundConsensus[h] > maxConsensus {
			maxConsensus = roundConsensus[h]
		}
	}

	if maxConsensus == 0 {
		return nil, errors.New("", "round info not found")
	}

	return
}

func GetMagicBlockByNumber(ctx context.Context, numSharders int, number int64) (m *block.MagicBlock, err error) {

	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromShardersContext(ctx, numSharders,
		fmt.Sprintf("%smagic_block_number=%d", GET_MAGIC_BLOCK_INFO, number),
		result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	type respObj struct {
		MagicBlock *block.MagicBlock `json:"magic_block"`
	}

	for i := 0; i < numSharders; i++ {
		var rsp = <-result

		Logger.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			Logger.Error(rsp.Body)
			continue
		}

		var respo respObj
		if err = json.Unmarshal([]byte(rsp.Body), &respo); err != nil {
			Logger.Error(" magic block parse error: ", err)
			err = nil
			continue
		}

		m = respo.MagicBlock
		var h = encryption.FastHash([]byte(respo.MagicBlock.Hash))
		if roundConsensus[h]++; roundConsensus[h] > maxConsensus {
			maxConsensus = roundConsensus[h]
		}
	}

	if maxConsensus == 0 {
		return nil, errors.New("", "magic block info not found")
	}

	return
}

func getBlockInfoByRound(numSharders int, round int64, content string) (*blockHeader, error) {
	result := make(chan *util.GetResponse)
	defer close(result)
	numSharders = len(_config.chain.Sharders) // overwrite, use all
	queryFromSharders(numSharders, fmt.Sprintf("%vround=%v&content=%v", GET_BLOCK_INFO, round, content), result)
	maxConsensus := int(0)
	roundConsensus := make(map[string]int)
	var blkHdr blockHeader
	for i := 0; i < numSharders; i++ {
		rsp := <-result
		Logger.Debug(rsp.Url, rsp.Status)
		if rsp.StatusCode == http.StatusOK {
			var objmap map[string]json.RawMessage
			err := json.Unmarshal([]byte(rsp.Body), &objmap)
			if err != nil {
				Logger.Error("round info parse error. ", err)
				continue
			}
			if header, ok := objmap["header"]; ok {
				err := json.Unmarshal([]byte(header), &objmap)
				if err != nil {
					Logger.Error("round info parse error. ", err)
					continue
				}
				if hash, ok := objmap["hash"]; ok {
					h := encryption.FastHash([]byte(hash))
					roundConsensus[h]++
					if roundConsensus[h] > maxConsensus {
						maxConsensus = roundConsensus[h]
						err := json.Unmarshal([]byte(header), &blkHdr)
						if err != nil {
							Logger.Error("round info parse error. ", err)
							continue
						}
					}
				}
			} else {
				Logger.Debug(rsp.Url, "no round confirmation. Resp:", rsp.Body)
			}
		} else {
			Logger.Error(rsp.Body)
		}

	}
	if maxConsensus == 0 {
		return nil, errors.New("", "round info not found.")
	}
	return &blkHdr, nil
}

func isBlockExtends(prevHash string, block *blockHeader) bool {
	data := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v:%v", block.MinerId, prevHash, block.CreationDate, block.Round,
		block.RoundRandomSeed, block.StateChangesCount, block.MerkleTreeRoot, block.ReceiptMerkleTreeRoot)
	h := encryption.Hash(data)
	return block.Hash == h
}

func validateChain(confirmBlock *blockHeader) bool {
	confirmRound := confirmBlock.Round
	Logger.Debug("Confirmation round: ", confirmRound)
	currentBlockHash := confirmBlock.Hash
	round := confirmRound + 1
	for {
		nextBlock, err := getBlockInfoByRound(1, round, "header")
		if err != nil {
			Logger.Info(err, " after a second falling thru to ", getMinShardersVerify(), "of ", len(_config.chain.Sharders), "Sharders")
			sys.Sleep(1 * time.Second)
			nextBlock, err = getBlockInfoByRound(getMinShardersVerify(), round, "header")
			if err != nil {
				Logger.Error(err, " block chain stalled. waiting", defaultWaitSeconds, "...")
				sys.Sleep(defaultWaitSeconds)
				continue
			}
		}
		if isBlockExtends(currentBlockHash, nextBlock) {
			currentBlockHash = nextBlock.Hash
			round++
		}
		if (round > confirmRound) && (round-confirmRound < getMinRequiredChainLength()) {
			continue
		}
		if round < confirmRound {
			return false
		}
		// Validation success
		break
	}
	return true
}
func (t *Transaction) isTransactionExpired(lfbCreationTime, currentTime int64) bool {
	// latest finalized block zero implies no response. use currentTime as lfb
	if lfbCreationTime == 0 {
		lfbCreationTime = currentTime
	}
	if util.MinInt64(lfbCreationTime, currentTime) > (t.txn.CreationDate + int64(defaultTxnExpirationSeconds)) {
		return true
	}
	// Wait for next retry
	sys.Sleep(defaultWaitSeconds)
	return false
}
func (t *Transaction) Verify() error {
	if t.txnHash == "" && t.txnStatus == StatusUnknown {
		return errors.New("", "invalid transaction. cannot be verified.")
	}
	if t.txnHash == "" && t.txnStatus == StatusSuccess {
		h := t.GetTransactionHash()
		if h == "" {
			transaction.Cache.Evict(t.txn.ClientID)
			return errors.New("", "invalid transaction. cannot be verified.")
		}
	}
	// If transaction is verify only start from current time
	if t.txn.CreationDate == 0 {
		t.txn.CreationDate = int64(common.Now())
	}

	tq, err := NewTransactionQuery(_config.chain.Sharders)
	if err != nil {
		Logger.Error(err)
		return err
	}

	go func() {

		for {

			tq.Reset()
			// Get transaction confirmationBlock from a random sharder
			confirmBlockHeader, confirmationBlock, lfbBlockHeader, err := tq.GetFastConfirmation(context.TODO(), t.txnHash)

			if err != nil {
				now := int64(common.Now())

				// maybe it is a network or server error
				if lfbBlockHeader == nil {
					Logger.Info(err, " now: ", now)
				} else {
					Logger.Info(err, " now: ", now, ", LFB creation time:", lfbBlockHeader.CreationDate)
				}

				// transaction is done or expired. it means random sharder might be outdated, try to query it from s/S sharders to confirm it
				if util.MaxInt64(lfbBlockHeader.getCreationDate(now), now) >= (t.txn.CreationDate + int64(defaultTxnExpirationSeconds)) {
					Logger.Info("falling back to ", getMinShardersVerify(), " of ", len(_config.chain.Sharders), " Sharders")
					confirmBlockHeader, confirmationBlock, lfbBlockHeader, err = tq.GetConsensusConfirmation(context.TODO(), getMinShardersVerify(), t.txnHash)
				}

				// txn not found in fast confirmation/consensus confirmation
				if err != nil {

					if lfbBlockHeader == nil {
						// no any valid lfb on all sharders. maybe they are network/server errors. try it again
						continue
					}

					// it is expired
					if t.isTransactionExpired(lfbBlockHeader.getCreationDate(now), now) {
						t.completeVerify(StatusError, "", errors.New("", `{"error": "verify transaction failed"}`))
						return
					}
					continue
				}

			}

			valid := validateChain(confirmBlockHeader)
			if valid {
				output, err := json.Marshal(confirmationBlock)
				if err != nil {
					t.completeVerify(StatusError, "", errors.New("", `{"error": "transaction confirmation json marshal error"`))
					return
				}
				confJson := confirmationBlock["confirmation"]

				var conf map[string]json.RawMessage
				if err := json.Unmarshal(confJson, &conf); err != nil {
					return
				}
				txnJson := conf["txn"]

				var tr map[string]json.RawMessage
				if err := json.Unmarshal(txnJson, &tr); err != nil {
					return
				}

				txStatus := tr["transaction_status"]
				switch string(txStatus) {
				case "1":
					t.completeVerifyWithConStatus(StatusSuccess, Success, string(output), nil)
				case "2":
					txOutput := tr["transaction_output"]
					t.completeVerifyWithConStatus(StatusSuccess, ChargeableError, string(txOutput), nil)
				default:
					t.completeVerify(StatusError, string(output), nil)
				}
				return
			}
		}
	}()
	return nil
}

func (t *Transaction) GetVerifyConfirmationStatus() ConfirmationStatus {
	return t.verifyConfirmationStatus
}

func (t *Transaction) GetVerifyOutput() string {
	if t.verifyStatus == StatusSuccess {
		return t.verifyOut
	}
	return ""
}

func (t *Transaction) GetTransactionError() string {
	if t.txnStatus != StatusSuccess {
		return t.txnError.Error()
	}
	return ""
}

func (t *Transaction) GetVerifyError() string {
	if t.verifyStatus != StatusSuccess {
		return t.verifyError.Error()
	}
	return ""
}

// GetTransactionNonce returns nonce
func (t *Transaction) GetTransactionNonce() int64 {
	return t.txn.TransactionNonce
}

// ========================================================================== //
//                               vesting pool                                 //
// ========================================================================== //

type vestingRequest struct {
	PoolID common.Key `json:"pool_id"`
}

func (t *Transaction) vestingPoolTxn(function string, poolID string,
	value int64) error {

	return t.createSmartContractTxn(VestingSmartContractAddress,
		function, vestingRequest{PoolID: common.Key(poolID)}, int64(value))
}

func (t *Transaction) VestingTrigger(poolID string) (err error) {

	err = t.vestingPoolTxn(transaction.VESTING_TRIGGER, poolID, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

type VestingStopRequest struct {
	PoolID      common.Key `json:"pool_id"`
	Destination common.Key `json:"destination"`
}

func (t *Transaction) VestingStop(sr *VestingStopRequest) (err error) {

	err = t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_STOP, sr, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) VestingUnlock(poolID string) (err error) {

	err = t.vestingPoolTxn(transaction.VESTING_UNLOCK, poolID, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

type VestingDest struct {
	ID     common.Key     `json:"id"`     // destination ID
	Amount common.Balance `json:"amount"` // amount to vest for the destination
}

type VestingAddRequest struct {
	Description  string           `json:"description"`  // allow empty
	StartTime    common.Timestamp `json:"start_time"`   //
	Duration     time.Duration    `json:"duration"`     //
	Destinations []*VestingDest   `json:"destinations"` //
}

func (t *Transaction) VestingAdd(ar *VestingAddRequest, value int64) (
	err error) {

	err = t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_ADD, ar, value)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) VestingDelete(poolID string) (err error) {

	err = t.vestingPoolTxn(transaction.VESTING_DELETE, poolID, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) VestingUpdateConfig(vscc *InputMap) (err error) {

	err = t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_UPDATE_SETTINGS, vscc, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// faucet smart contract

func (t *Transaction) FaucetUpdateConfig(ip *InputMap) (err error) {

	err = t.createSmartContractTxn(FaucetSmartContractAddress,
		transaction.FAUCETSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

//
// miner SC
//

func (t *Transaction) MinerScUpdateConfig(ip *InputMap) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerScUpdateGlobals(ip *InputMap) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_GLOBALS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

type MinerSCDelegatePool struct {
	Settings StakePoolSettings `json:"settings"`
}

type SimpleMiner struct {
	ID string `json:"id"`
}

type MinerSCMinerInfo struct {
	SimpleMiner         `json:"simple_miner"`
	MinerSCDelegatePool `json:"stake_pool"`
}

func (t *Transaction) MinerSCMinerSettings(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_SETTINGS, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCSharderSettings(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_SETTINGS, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCDeleteMiner(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_DELETE, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCDeleteSharder(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_DELETE, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

type Provider int

const (
	ProviderMiner Provider = iota
	ProviderSharder
	ProviderBlobber
	ProviderValidator
	ProviderAuthorizer
)

type SCCollectReward struct {
	ProviderId   string   `json:"provider_id"`
	PoolId       string   `json:"pool_id"`
	ProviderType Provider `json:"provider_type"`
}

func (t *Transaction) MinerSCCollectReward(providerId, poolId string, providerType Provider) error {
	pr := &SCCollectReward{
		ProviderId:   providerId,
		PoolId:       poolId,
		ProviderType: providerType,
	}
	err := t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_COLLECT_REWARD, pr, 0)
	if err != nil {
		Logger.Error(err)
		return err
	}
	go func() { t.setNonceAndSubmit() }()
	return err
}

type MinerSCLock struct {
	ID string `json:"id"`
}

func (t *Transaction) MinerSCLock(nodeID string, lock int64) (err error) {

	var mscl MinerSCLock
	mscl.ID = nodeID

	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, &mscl, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

type MinerSCUnlock struct {
	ID     string `json:"id"`
	PoolID string `json:"pool_id"`
}

func (t *Transaction) MinerSCUnlock(nodeID, poolID string) (err error) {
	var mscul MinerSCUnlock
	mscul.ID = nodeID
	mscul.PoolID = poolID

	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UNLOCK, &mscul, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

//RegisterMultiSig register a multisig wallet with the SC.
func (t *Transaction) RegisterMultiSig(walletstr string, mswallet string) error {
	w, err := GetWallet(walletstr)
	if err != nil {
		fmt.Printf("Error while parsing the wallet. %v\n", err)
		return err
	}

	msw, err := GetMultisigPayload(mswallet)
	if err != nil {
		fmt.Printf("\nError in registering. %v\n", err)
		return err
	}
	sn := transaction.SmartContractTxnData{Name: MultiSigRegisterFuncName, InputArgs: msw}
	snBytes, err := json.Marshal(sn)
	if err != nil {
		return errors.Wrap(err, "execute multisig register failed due to invalid data.")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSmartContract
		t.txn.ToClientID = MultiSigSmartContractAddress
		t.txn.TransactionData = string(snBytes)
		t.txn.Value = 0
		nonce := t.txn.TransactionNonce
		if nonce < 1 {
			nonce = transaction.Cache.GetNextNonce(t.txn.ClientID)
		} else {
			transaction.Cache.Set(t.txn.ClientID, nonce)
		}
		t.txn.TransactionNonce = nonce

		t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
		t.submitTxn()
	}()
	return nil
}

// NewMSTransaction new transaction object for multisig operation
func NewMSTransaction(walletstr string, cb TransactionCallback) (*Transaction, error) {
	w, err := GetWallet(walletstr)
	if err != nil {
		fmt.Printf("Error while parsing the wallet. %v", err)
		return nil, err
	}
	t := &Transaction{}
	t.txn = transaction.NewTransactionEntity(w.ClientID, _config.chain.ChainID, w.ClientKey, w.Nonce)
	t.txnStatus, t.verifyStatus = StatusUnknown, StatusUnknown
	t.txnCb = cb
	return t, nil
}

//RegisterVote register a multisig wallet with the SC.
func (t *Transaction) RegisterVote(signerwalletstr string, msvstr string) error {

	w, err := GetWallet(signerwalletstr)
	if err != nil {
		fmt.Printf("Error while parsing the wallet. %v", err)
		return err
	}

	msv, err := GetMultisigVotePayload(msvstr)

	if err != nil {
		fmt.Printf("\nError in voting. %v\n", err)
		return err
	}
	sn := transaction.SmartContractTxnData{Name: MultiSigVoteFuncName, InputArgs: msv}
	snBytes, err := json.Marshal(sn)
	if err != nil {
		return errors.Wrap(err, "execute multisig vote failed due to invalid data.")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSmartContract
		t.txn.ToClientID = MultiSigSmartContractAddress
		t.txn.TransactionData = string(snBytes)
		t.txn.Value = 0
		nonce := t.txn.TransactionNonce
		if nonce < 1 {
			nonce = transaction.Cache.GetNextNonce(t.txn.ClientID)
		} else {
			transaction.Cache.Set(t.txn.ClientID, nonce)
		}
		t.txn.TransactionNonce = nonce
		t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
		t.submitTxn()
	}()
	return nil
}

func VerifyContentHash(metaTxnDataJSON string) (bool, error) {
	var metaTxnData sdk.CommitMetaResponse
	err := json.Unmarshal([]byte(metaTxnDataJSON), &metaTxnData)
	if err != nil {
		return false, errors.New("metaTxnData_decode_error", "Unable to decode metaTxnData json")
	}

	t, err := transaction.VerifyTransaction(metaTxnData.TxnID, blockchain.GetSharders())
	if err != nil {
		return false, errors.New("fetch_txm_details", "Unable to fetch txn details")
	}

	var metaOperation sdk.CommitMetaData
	err = json.Unmarshal([]byte(t.TransactionData), &metaOperation)
	if err != nil {
		Logger.Error("Unmarshal of transaction data to fileMeta failed, Maybe not a commit meta txn :", t.Hash)
		return false, nil
	}

	return metaOperation.MetaData.Hash == metaTxnData.MetaData.Hash, nil
}

//
// Storage SC transactions
//

func (t *Transaction) StorageSCCollectReward(providerId, poolId string, providerType Provider) error {
	pr := &SCCollectReward{
		ProviderId:   providerId,
		PoolId:       poolId,
		ProviderType: providerType,
	}
	err := t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_COLLECT_REWARD, pr, 0)
	if err != nil {
		Logger.Error(err)
		return err
	}
	go t.setNonceAndSubmit()
	return err
}

func (t *Transaction) StorageScUpdateConfig(ip *InputMap) (err error) {
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// FinalizeAllocation transaction.
func (t *Transaction) FinalizeAllocation(allocID string, fee int64) (
	err error) {

	type finiRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_FINALIZE_ALLOCATION, &finiRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// CancelAllocation transaction.
func (t *Transaction) CancelAllocation(allocID string, fee int64) (
	err error) {

	type cancelRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CANCEL_ALLOCATION, &cancelRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// PriceRange represents a price range allowed by user to filter blobbers.
type PriceRange struct {
	Min common.Balance `json:"min"`
	Max common.Balance `json:"max"`
}

// CreateAllocationRequest is information to create allocation.
type CreateAllocationRequest struct {
	DataShards      int              `json:"data_shards"`
	ParityShards    int              `json:"parity_shards"`
	Size            common.Size      `json:"size"`
	Expiration      common.Timestamp `json:"expiration_date"`
	Owner           string           `json:"owner_id"`
	OwnerPublicKey  string           `json:"owner_public_key"`
	Blobbers        []string         `json:"blobbers"`
	ReadPriceRange  PriceRange       `json:"read_price_range"`
	WritePriceRange PriceRange       `json:"write_price_range"`
}

// CreateAllocation transaction.
func (t *Transaction) CreateAllocation(car *CreateAllocationRequest,
	lock, fee int64) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_ALLOCATION, car, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// CreateReadPool for current user.
func (t *Transaction) CreateReadPool(fee int64) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_READ_POOL, nil, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// ReadPoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) ReadPoolLock(allocID, blobberID string,
	duration int64, lock, fee int64) (err error) {

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_LOCK, &lr, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// ReadPoolUnlock for current user and given pool.
func (t *Transaction) ReadPoolUnlock(poolID string, fee int64) (err error) {
	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_UNLOCK, &unlockRequest{
			PoolID: poolID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// StakePoolLock used to lock tokens in a stake pool of a blobber.
func (t *Transaction) StakePoolLock(blobberID string, lock, fee int64) (
	err error) {

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_LOCK, &spr, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// StakePoolUnlock by blobberID and poolID.
func (t *Transaction) StakePoolUnlock(blobberID, poolID string,
	fee int64) (err error) {

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
		PoolID    string `json:"pool_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID
	spr.PoolID = poolID

	err = t.createSmartContractTxn(StorageSmartContractAddress, transaction.STORAGESC_STAKE_POOL_UNLOCK, &spr, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

type StakePoolSettings struct {
	DelegateWallet string         `json:"delegate_wallet"`
	MinStake       common.Balance `json:"min_stake"`
	MaxStake       common.Balance `json:"max_stake"`
	NumDelegates   int            `json:"num_delegates"`
	ServiceCharge  float64        `json:"service_charge"`
}

type Terms struct {
	ReadPrice        common.Balance `json:"read_price"`  // tokens / GB
	WritePrice       common.Balance `json:"write_price"` // tokens / GB
	MinLockDemand    float64        `json:"min_lock_demand"`
	MaxOfferDuration time.Duration  `json:"max_offer_duration"`
}

type Blobber struct {
	ID                common.Key        `json:"id"`
	BaseURL           string            `json:"url"`
	Terms             Terms             `json:"terms"`
	Capacity          common.Size       `json:"capacity"`
	Used              common.Size       `json:"used"`
	LastHealthCheck   common.Timestamp  `json:"last_health_check"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

type AuthorizerStakePoolSettings struct {
	DelegateWallet string         `json:"delegate_wallet"`
	MinStake       common.Balance `json:"min_stake"`
	MaxStake       common.Balance `json:"max_stake"`
	NumDelegates   int            `json:"num_delegates"`
	ServiceCharge  float64        `json:"service_charge"`
}

type AddAuthorizerPayload struct {
	PublicKey         string                      `json:"public_key"`
	URL               string                      `json:"url"`
	StakePoolSettings AuthorizerStakePoolSettings `json:"stake_pool_settings"` // Used to initially create stake pool
}

type AuthorizerConfig struct {
	Fee common.Balance `json:"fee"`
}

type AuthorizerNode struct {
	ID     string            `json:"id"`
	Config *AuthorizerConfig `json:"config"`
}

// UpdateBlobberSettings update settings of a blobber.
func (t *Transaction) UpdateBlobberSettings(b *Blobber, fee int64) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, b, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// UpdateAllocation transaction.
func (t *Transaction) UpdateAllocation(allocID string, sizeDiff int64,
	expirationDiff int64, lock, fee int64) (err error) {

	type updateAllocationRequest struct {
		ID         string `json:"id"`              // allocation id
		Size       int64  `json:"size"`            // difference
		Expiration int64  `json:"expiration_date"` // difference
	}

	var uar updateAllocationRequest
	uar.ID = allocID
	uar.Size = sizeDiff
	uar.Expiration = expirationDiff

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_ALLOCATION, &uar, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// WritePoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) WritePoolLock(allocID, blobberID string, duration int64,
	lock, fee int64) (err error) {

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_LOCK, &lr, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// WritePoolUnlock for current user and given pool.
func (t *Transaction) WritePoolUnlock(poolID string, fee int64) (
	err error) {

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_UNLOCK, &unlockRequest{
			PoolID: poolID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

//
// ZCNSC transactions
//

func (t *Transaction) ZCNSCUpdateGlobalConfig(ip *InputMap) (err error) {
	err = t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_GLOBAL_CONFIG, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go t.setNonceAndSubmit()
	return
}

func (t *Transaction) ZCNSCUpdateAuthorizerConfig(ip *AuthorizerNode) (err error) {
	err = t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_AUTHORIZER_CONFIG, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go t.setNonceAndSubmit()
	return
}

func (t *Transaction) ZCNSCAddAuthorizer(ip *AddAuthorizerPayload) (err error) {
	err = t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_ADD_AUTHORIZER, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go t.setNonceAndSubmit()
	return
}
