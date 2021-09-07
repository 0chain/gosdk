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
	errAuthVerifyFailed = errors.New("", "verfication failed for auth response")
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
	RoundRandomSeed       int64  `json:"round_random_seed,omitempy"`
	MerkleTreeRoot        string `json:"merkle_tree_root,omitempty"`
	StateHash             string `json:"state_hash,omitempty"`
	ReceiptMerkleTreeRoot string `json:"receipt_merkle_tree_root,omitempty"`
	NumTxns               int64  `json:"num_txns,omitempty"`
}

type Transaction struct {
	txn          *transaction.Transaction
	txnOut       string
	txnHash      string
	txnStatus    int
	txnError     error
	txnCb        TransactionCallback
	verifyStatus int
	verifyOut    string
	verifyError  error
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
	// ExecuteSmartContract impements the Faucet Smart contract
	ExecuteSmartContract(address, methodName, jsoninput string, val int64) error
	// ExecuteFaucetSCWallet impements the Faucet Smart contract for a given wallet
	ExecuteFaucetSCWallet(walletStr string, methodName string, input []byte) error
	// GetTransactionHash implements retrieval of hash of the submitted transaction
	GetTransactionHash() string
	// LockTokens implements the lock token.
	LockTokens(val int64, durationHr int64, durationMin int) error
	// UnlockTokens implements unlocking of earlier locked tokens.
	UnlockTokens(poolID string) error
	//RegisterMultiSig registers a group wallet and subwallets with MultisigSC
	RegisterMultiSig(walletstr, mswallet string) error
	// SetTransactionHash implements verify a previous transation status
	SetTransactionHash(hash string) error
	// SetTransactionFee implements method to set the transaction fee
	SetTransactionFee(txnFee int64) error
	// Verify implements verify the transaction
	Verify() error
	// GetVerifyOutput implements the verifcation output from sharders
	GetVerifyOutput() string
	// GetTransactionError implements error string incase of transaction failure
	GetTransactionError() string
	// GetVerifyError implements error string incase of verify failure error
	GetVerifyError() string

	// Output of transaction.
	Output() []byte

	// vesting SC
	VestingTrigger(poolID string) error
	VestingStop(sr *VestingStopRequest) error
	VestingUnlock(poolID string) error
	VestingAdd(ar *VestingAddRequest, value int64) error
	VestingDelete(poolID string) error
	VestingUpdateConfig(*InputMap) error

	// Miner SC

	MinerSCMinerSettings(*MinerSCMinerInfo) error
	MinerSCSharderSettings(*MinerSCMinerInfo) error
	MinerSCLock(minerID string, lock int64) error
	MienrSCUnlock(minerID, poolID string) error
	MinerScUpdateConfig(*InputMap) error
	MinerScUpdateGlobals(*InputMap) error
	MinerSCDeleteMiner(*MinerSCMinerInfo) error
	MinerSCDeleteSharder(*MinerSCMinerInfo) error


	// Storage SC

	FinalizeAllocation(allocID string, fee int64) error
	CancelAllocation(allocID string, fee int64) error
	CreateAllocation(car *CreateAllocationRequest, lock, fee int64) error //
	CreateReadPool(fee int64) error
	ReadPoolLock(allocID string, blobberID string, duration int64, lock, fee int64) error
	ReadPoolUnlock(poolID string, fee int64) error
	StakePoolLock(blobberID string, lock, fee int64) error
	StakePoolUnlock(blobberID string, poolID string, fee int64) error
	StakePoolPayInterests(blobberID string, fee int64) error
	UpdateBlobberSettings(blobber *Blobber, fee int64) error
	UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock, fee int64) error
	WritePoolLock(allocID string, blobberID string, duration int64, lock, fee int64) error
	WritePoolUnlock(poolID string, fee int64) error
	StorageScUpdateConfig(*InputMap) error

	// Interest pool SC
	InterestPoolUpdateConfig(*InputMap) error

	// Faucet
	FaucetUpdateConfig(*InputMap) error
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

func (t *Transaction) completeTxn(status int, out string, err error) {
	t.txnStatus = status
	t.txnOut = out
	t.txnError = err
	if t.txnCb != nil {
		t.txnCb.OnTransactionComplete(t, t.txnStatus)
	}
}

func (t *Transaction) completeVerify(status int, out string, err error) {
	t.verifyStatus = status
	t.verifyOut = out
	t.verifyError = err
	if t.txnCb != nil {
		t.txnCb.OnVerifyComplete(t, t.verifyStatus)
	}
}

func (t *Transaction) submitTxn() {
	// Clear the status, incase transaction object reused
	t.txnStatus = StatusUnknown
	t.txnOut = ""
	t.txnError = nil

	// If Signature is not passed compute signature
	if t.txn.Signature == "" {
		err := t.txn.ComputeHashAndSign(signFn)
		if err != nil {
			t.completeTxn(StatusError, "", err)
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
			Logger.Info("Submitting ", txnTypeString(t.txn.TransactionType), " transaction to ", minerurl)
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
			return
		}(miner)
	}
	consensus := float32(0)
	for range randomMiners {
		select {
		case rsp := <-result:
			Logger.Debug(rsp.Url, rsp.Status)
			if rsp.StatusCode == http.StatusOK {
				consensus++
				tSuccessRsp = rsp.Body
			} else {
				Logger.Error(rsp.Body)
				tFailureRsp = rsp.Body
			}
		}
	}
	rate := consensus * 100 / float32(len(randomMiners))
	if rate < consensusThresh {
		t.completeTxn(StatusError, "", fmt.Errorf("submit transaction failed. %s", tFailureRsp))
		return
	}
	time.Sleep(3 * time.Second)
	t.completeTxn(StatusSuccess, tSuccessRsp, nil)
}

func newTransaction(cb TransactionCallback, txnFee int64) (*Transaction, error) {
	t := &Transaction{}
	t.txn = transaction.NewTransactionEntity(_config.wallet.ClientID, _config.chain.ChainID, _config.wallet.ClientKey)
	t.txnStatus, t.verifyStatus = StatusUnknown, StatusUnknown
	t.txnCb = cb
	t.txn.TransactionFee = txnFee
	return t, nil
}

// NewTransaction allocation new generic transaction object for any operation
func NewTransaction(cb TransactionCallback, txnFee int64) (TransactionScheme, error) {
	err := checkConfig()
	if err != nil {
		return nil, err
	}
	if _config.isSplitWallet {
		if _config.authUrl == "" {
			return nil, errors.New("", "auth url not set")
		}
		Logger.Info("New transaction interface with auth")
		return newTransactionWithAuth(cb, txnFee)
	}
	Logger.Info("New transaction interface")
	return newTransaction(cb, txnFee)
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

func (t *Transaction) Send(toClientID string, val int64, desc string) error {
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSend
		t.txn.ToClientID = toClientID
		t.txn.Value = val
		t.txn.TransactionData = desc
		t.submitTxn()
	}()
	return nil
}

func (t *Transaction) SendWithSignatureHash(toClientID string, val int64, desc string, sig string, CreationDate int64, hash string) error {
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSend
		t.txn.ToClientID = toClientID
		t.txn.Value = val
		t.txn.Hash = hash
		t.txn.TransactionData = desc
		t.txn.Signature = sig
		t.txn.CreationDate = CreationDate
		t.submitTxn()
	}()
	return nil
}

func (t *Transaction) StoreData(data string) error {
	go func() {
		t.txn.TransactionType = transaction.TxnTypeData
		t.txn.TransactionData = data
		t.submitTxn()
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

// ExecuteFaucetSCWallet impements the Faucet Smart contract for a given wallet
func (t *Transaction) ExecuteFaucetSCWallet(walletStr string, methodName string, input []byte) error {
	w, err := t.createFaucetSCWallet(walletStr, methodName, input)
	if err != nil {
		return err
	}
	go func() {
		t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
		fmt.Printf("submitted transaction\n")
		t.submitTxn()
	}()
	return nil
}

func (t *Transaction) ExecuteSmartContract(address, methodName, jsoninput string, val int64) error {
	scData := make(map[string]interface{})
	json.Unmarshal([]byte(jsoninput), &scData)
	err := t.createSmartContractTxn(address, methodName, scData, val)
	if err != nil {
		return err
	}
	go func() {
		t.submitTxn()
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
			return
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
		block.MerkleTreeRoot = cfm.MerkleTreeRoot
		block.ReceiptMerkleTreeRoot = cfm.ReceiptMerkleTreeRoot
		// Verify the block
		if isBlockExtends(prevBlockHash, block) {
			return block, nil
		} else {
			return nil, errors.New("", "block hash verification failed in confirmation")
		}
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
		select {
		case rsp := <-result:
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
		select {
		case rsp := <-result:
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
	}
	if maxConsensus == 0 {
		return nil, errors.New("", "round info not found.")
	}
	return &blkHdr, nil
}

func isBlockExtends(prevHash string, block *blockHeader) bool {
	data := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v", block.MinerId, prevHash, block.CreationDate, block.Round,
		block.RoundRandomSeed, block.MerkleTreeRoot, block.ReceiptMerkleTreeRoot)
	h := encryption.Hash(data)
	if block.Hash == h {
		return true
	}
	return false
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
			time.Sleep(1 * time.Second)
			nextBlock, err = getBlockInfoByRound(getMinShardersVerify(), round, "header")
			if err != nil {
				Logger.Error(err, " block chain stalled. waiting", defaultWaitSeconds, "...")
				time.Sleep(defaultWaitSeconds)
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
	time.Sleep(defaultWaitSeconds)
	return false
}
func (t *Transaction) Verify() error {
	if t.txnHash == "" && t.txnStatus == StatusUnknown {
		return errors.New("", "invalid transaction. cannot be verified.")
	}
	if t.txnHash == "" && t.txnStatus == StatusSuccess {
		h := t.GetTransactionHash()
		if h == "" {
			return errors.New("", "invalid transaction. cannot be verified.")
		}
	}
	// If transaction is verify only start from current time
	if t.txn.CreationDate == 0 {
		t.txn.CreationDate = int64(common.Now())
	}

	go func() {
		for {
			// Get transaction confirmation from random sharder
			confirmBlock, confirmation, lfb, err := getTransactionConfirmation(1, t.txnHash)
			if err != nil {
				tn := int64(common.Now())
				Logger.Info(err, " now: ", tn, ", LFB creation time:", lfb.CreationDate)
				if util.MaxInt64(lfb.CreationDate, tn) < (t.txn.CreationDate + int64(defaultTxnExpirationSeconds)) {
					Logger.Info("falling back to ", getMinShardersVerify(), " of ", len(_config.chain.Sharders), " Sharders")
					confirmBlock, confirmation, lfb, err = getTransactionConfirmation(getMinShardersVerify(), t.txnHash)
					if err != nil {
						if t.isTransactionExpired(lfb.CreationDate, tn) {
							t.completeVerify(StatusError, "", errors.New("", `{"error": "verify transaction failed"}`))
							return
						}
						continue
					}
				} else {
					if t.isTransactionExpired(lfb.CreationDate, tn) {
						t.completeVerify(StatusError, "", errors.New("", `{"error": "verify transaction failed"}`))
						return
					}
					continue
				}
			}
			valid := validateChain(confirmBlock)
			if valid {
				output, err := json.Marshal(confirmation)
				if err != nil {
					t.completeVerify(StatusError, "", errors.New("", `{"error": "transaction confirmation json marshal error"`))
					return
				}
				t.completeVerify(StatusSuccess, string(output), nil)
				return
			}
		}
	}()
	return nil
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
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
	return
}

func (t *Transaction) VestingUnlock(poolID string) (err error) {

	err = t.vestingPoolTxn(transaction.VESTING_UNLOCK, poolID, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
	return
}

func (t *Transaction) VestingDelete(poolID string) (err error) {

	err = t.vestingPoolTxn(transaction.VESTING_DELETE, poolID, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.submitTxn() }()
	return
}

func (t *Transaction) VestingUpdateConfig(vscc *InputMap) (err error) {

	err = t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_UPDATE_SETTINGS, vscc, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.submitTxn() }()
	return
}

// interest pool smart contract

func (t *Transaction) InterestPoolUpdateConfig(ip *InputMap) (err error) {

	err = t.createSmartContractTxn(InterestPoolSmartContractAddress,
		transaction.INTERESTPOOLSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
	return
}

func (t *Transaction) MinerScUpdateGlobals(ip *InputMap) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_GLOBALS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.submitTxn() }()
	return
}

type MinerSCPoolStats struct {
	DelegateID   string         `json:"delegate_id"`
	High         common.Balance `json:"high"`
	Low          common.Balance `json:"low"`
	InterestRate float64        `json:"interest_rate"`
	TotalPaid    common.Balance `json:"total_paid"`
	NumRounds    int64          `json:"number_rounds"`
}

type MinerSCPool struct {
	ID      string         `json:"id"`
	Balance common.Balance `json:"balance"`
}

type MinerSCDelegatePool struct {
	MinerSCPoolStats `json:"stats"`
	MinerSCPool      `json:"pool"`
}

type MinerSCMinerStat struct {
	// for miner (totals)
	BlockReward      common.Balance `json:"block_reward,omitempty"`
	ServiceCharge    common.Balance `json:"service_charge,omitempty"`
	UsersFee         common.Balance `json:"users_fee,omitempty"`
	BlockShardersFee common.Balance `json:"block_sharders_fee,omitempty"`
	// for sharder (total)
	SharderRewards common.Balance `json:"sharder_rewards,omitempty"`
}

type MinerSCMinerInfo struct {
	*SimpleMinerSCMinerInfo `json:"simple_miner"`
	Pending                 map[string]MinerSCDelegatePool `json:"pending"`
	Active                  map[string]MinerSCDelegatePool `json:"active"`
	Deleting                map[string]MinerSCDelegatePool `json:"deleting"`
}

type SimpleMinerSCMinerInfo struct {
	ID      string `json:"id"`
	BaseURL string `json:"url"`

	DelegateWallet    string         `json:"delegate_wallet"`     //
	ServiceCharge     float64        `json:"service_charge"`      // %
	NumberOfDelegates int            `json:"number_of_delegates"` //
	MinStake          common.Balance `json:"min_stake"`           //
	MaxStake          common.Balance `json:"max_stake"`           //

	Stat MinerSCMinerStat `json:"stat"`
}

func (t *Transaction) MinerSCMinerSettings(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_SETTINGS, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.submitTxn() }()
	return
}

func (t *Transaction) MinerSCSharderSettings(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_SETTINGS, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.submitTxn() }()
	return
}

func (t *Transaction) MinerSCDeleteMiner(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_DELETE, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.submitTxn() }()
	return
}

func (t *Transaction) MinerSCDeleteSharder(info *MinerSCMinerInfo) (err error) {
	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_DELETE, info, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.submitTxn() }()
	return
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
	go func() { t.submitTxn() }()
	return
}

type MinerSCUnlock struct {
	ID     string `json:"id"`
	PoolID string `json:"pool_id"`
}

func (t *Transaction) MienrSCUnlock(nodeID, poolID string) (err error) {
	var mscul MinerSCUnlock
	mscul.ID = nodeID
	mscul.PoolID = poolID

	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UNLOCK, &mscul, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.submitTxn() }()
	return
}

//
// interest pool
//

func (t *Transaction) createLockTokensTxn(val int64, durationHr int64, durationMin int) error {
	lockInput := make(map[string]interface{})
	lockInput["duration"] = fmt.Sprintf("%dh%dm", durationHr, durationMin)
	err := t.createSmartContractTxn(InterestPoolSmartContractAddress, transaction.LOCK_TOKEN, lockInput, val)
	return err
}

func (t *Transaction) LockTokens(val int64, durationHr int64, durationMin int) error {
	err := t.createLockTokensTxn(val, durationHr, durationMin)
	if err != nil {
		Logger.Error(err)
		return err
	}
	go func() {
		t.submitTxn()
	}()
	return nil
}

func (t *Transaction) createUnlockTokensTxn(poolID string) error {
	unlockInput := make(map[string]interface{})
	unlockInput["pool_id"] = poolID
	return t.createSmartContractTxn(InterestPoolSmartContractAddress, transaction.UNLOCK_TOKEN, unlockInput, 0)
}

func (t *Transaction) UnlockTokens(poolID string) error {
	err := t.createUnlockTokensTxn(poolID)
	if err != nil {
		Logger.Error(err)
		return err
	}
	go func() {
		t.submitTxn()
	}()
	return nil
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
	t.txn = transaction.NewTransactionEntity(w.ClientID, _config.chain.ChainID, w.ClientKey)
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

func (t *Transaction) StorageScUpdateConfig(ip *InputMap) (err error) {
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
	return
}

// PriceRange represents a price range allowed by user to filter blobbers.
type PriceRange struct {
	Min common.Balance `json:"min"`
	Max common.Balance `json:"max"`
}

// CreateAllocationRequest is information to create allocation.
type CreateAllocationRequest struct {
	DataShards                 int              `json:"data_shards"`
	ParityShards               int              `json:"parity_shards"`
	Size                       common.Size      `json:"size"`
	Expiration                 common.Timestamp `json:"expiration_date"`
	Owner                      string           `json:"owner_id"`
	OwnerPublicKey             string           `json:"owner_public_key"`
	PreferredBlobbers          []string         `json:"preferred_blobbers"`
	ReadPriceRange             PriceRange       `json:"read_price_range"`
	WritePriceRange            PriceRange       `json:"write_price_range"`
	MaxChallengeCompletionTime time.Duration    `json:"max_challenge_completion_time"`
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
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
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

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_UNLOCK, &spr, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.submitTxn() }()
	return
}

// StakePoolPayInterests trigger interests payments.
func (t *Transaction) StakePoolPayInterests(blobberID string, fee int64) (
	err error) {

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_PAY_INTERESTS, &spr, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.submitTxn() }()
	return
}

type StakePoolSettings struct {
	DelegateWallet string         `json:"delegate_wallet"`
	MinStake       common.Balance `json:"min_stake"`
	MaxStake       common.Balance `json:"max_stake"`
	NumDelegates   int            `json:"num_delegates"`
}

type Terms struct {
	ReadPrice               common.Balance `json:"read_price"`  // tokens / GB
	WritePrice              common.Balance `json:"write_price"` // tokens / GB
	MinLockDemand           float64        `json:"min_lock_demand"`
	MaxOfferDuration        time.Duration  `json:"max_offer_duration"`
	ChallengeCompletionTime time.Duration  `json:"challenge_completion_time"`
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

// UpdateBlobberSettings update settings of a blobber.
func (t *Transaction) UpdateBlobberSettings(b *Blobber, fee int64) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, b, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
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
	go func() { t.submitTxn() }()
	return
}
