package zcncore

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/logger"
	"go.uber.org/zap"

	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/version"
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

// TransactionScheme implements few methods for block chain.
//
// Note: to be buildable on MacOSX all arguments should have names.
type TransactionScheme interface {
	TransactionCommon
	// SetTransactionCallback implements storing the callback
	// used to call after the transaction or verification is completed
	SetTransactionCallback(cb TransactionCallback) error
	// StoreData implements store the data to blockchain
	StoreData(data string) error
	// ExecuteFaucetSCWallet implements the `Faucet Smart contract` for a given wallet
	ExecuteFaucetSCWallet(walletStr string, methodName string, input []byte) error
	// GetTransactionHash implements retrieval of hash of the submitted transaction
	GetTransactionHash() string
	// SetTransactionHash implements verify a previous transaction status
	SetTransactionHash(hash string) error
	// SetTransactionNonce implements method to set the transaction nonce
	SetTransactionNonce(txnNonce int64) error
	// Verify implements verify the transaction
	Verify() error
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
	VestingDelete(poolID string) error

	// Miner SC
}

// TransactionCallback needs to be implemented by the caller for transaction related APIs
type TransactionCallback interface {
	OnTransactionComplete(t *Transaction, status int)
	OnVerifyComplete(t *Transaction, status int)
	OnAuthComplete(t *Transaction, status int)
}

type localConfig struct {
	chain         ChainConfig
	wallet        zcncrypto.Wallet
	authUrl       string
	isConfigured  bool
	isValidWallet bool
	isSplitWallet bool
}

type ChainConfig struct {
	ChainID                 string   `json:"chain_id,omitempty"`
	BlockWorker             string   `json:"block_worker"`
	Miners                  []string `json:"miners"`
	Sharders                []string `json:"sharders"`
	SignatureScheme         string   `json:"signature_scheme"`
	MinSubmit               int      `json:"min_submit"`
	MinConfirmation         int      `json:"min_confirmation"`
	ConfirmationChainLength int      `json:"confirmation_chain_length"`
	EthNode                 string   `json:"eth_node"`
}

// InitZCNSDK initializes the SDK with miner, sharder and signature scheme provided.
func InitZCNSDK(blockWorker string, signscheme string, configs ...func(*ChainConfig) error) error {
	if signscheme != "ed25519" && signscheme != "bls0chain" {
		return errors.New("", "invalid/unsupported signature scheme")
	}
	_config.chain.BlockWorker = blockWorker
	_config.chain.SignatureScheme = signscheme

	err := UpdateNetworkDetails()
	if err != nil {
		fmt.Println("UpdateNetworkDetails:", err)
		return err
	}

	go updateNetworkDetailsWorker(context.Background())

	for _, conf := range configs {
		err := conf(&_config.chain)
		if err != nil {
			return errors.Wrap(err, "invalid/unsupported options.")
		}
	}
	assertConfig()
	_config.isConfigured = true
	logging.Info("******* Wallet SDK Version:", version.VERSIONSTR, " ******* (InitZCNSDK)")

	cfg := &conf.Config{
		BlockWorker:             _config.chain.BlockWorker,
		MinSubmit:               _config.chain.MinSubmit,
		MinConfirmation:         _config.chain.MinConfirmation,
		ConfirmationChainLength: _config.chain.ConfirmationChainLength,
		SignatureScheme:         _config.chain.SignatureScheme,
		ChainID:                 _config.chain.ChainID,
		EthereumNode:            _config.chain.EthNode,
	}

	conf.InitClientConfig(cfg)

	return nil
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
	verifyConfirmationStatus int
	verifyOut                string
	verifyError              error
}

type SendTxnData struct {
	Note string `json:"note"`
}

func Sign(hash string) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(_config.wallet.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}

var SignFn = func(hash string) (string, error) {
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

func (t *Transaction) completeVerifyWithConStatus(status int, conStatus int, out string, err error) {
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
		err := t.txn.ComputeHashAndSign(SignFn)
		if err != nil {
			t.completeTxn(StatusError, "", err)
			transaction.Cache.Evict(t.txn.ClientID)
			return
		}
	}

	var (
		randomMiners = util.GetRandom(_config.chain.Miners, getMinMinersSubmit())
		minersN      = len(randomMiners)
		failedCount  int32
		failC        = make(chan struct{})
		resultC      = make(chan *util.PostResponse, minersN)
	)

	for _, miner := range randomMiners {
		go func(minerurl string) {
			url := minerurl + PUT_TRANSACTION
			logging.Info("Submitting ", txnTypeString(t.txn.TransactionType), " transaction to ", minerurl, " with JSON ", string(t.txn.DebugJSON()))
			req, err := util.NewHTTPPostRequest(url, t.txn)
			if err != nil {
				logging.Error(minerurl, " new post request failed. ", err.Error())

				if int(atomic.AddInt32(&failedCount, 1)) == minersN {
					close(failC)
				}
				return
			}

			res, err := req.Post()
			if err != nil {
				logging.Error(minerurl, " submit transaction error. ", err.Error())
				if int(atomic.AddInt32(&failedCount, 1)) == minersN {
					close(failC)
				}
				return
			}

			resultC <- res
		}(miner)
	}

	select {
	case <-failC:
		logging.Error("failed to submit transaction")
		t.completeTxn(StatusError, "", fmt.Errorf("failed to submit transaction to all miners"))
		return
	case ret := <-resultC:
		logging.Debug("finish txn submitting, ", ret.Url, ", Status: ", ret.Status, ", output:", ret.Body)
		if ret.StatusCode == http.StatusOK {
			t.completeTxn(StatusSuccess, ret.Body, nil)
		} else {
			t.completeTxn(StatusError, "", fmt.Errorf("submit transaction failed. %s", ret.Body))
			transaction.Cache.Evict(t.txn.ClientID)
		}
	}
}

func newTransaction(cb TransactionCallback, txnFee uint64, nonce int64) (*Transaction, error) {
	t := &Transaction{}
	t.txn = transaction.NewTransactionEntity(_config.wallet.ClientID, _config.chain.ChainID, _config.wallet.ClientKey, nonce)
	t.txnStatus, t.verifyStatus = StatusUnknown, StatusUnknown
	t.txnCb = cb
	t.txn.TransactionNonce = nonce
	t.txn.TransactionFee = txnFee
	return t, nil
}

func (t *Transaction) SetTransactionCallback(cb TransactionCallback) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction hash.")
	}
	t.txnCb = cb
	return nil
}

func (t *Transaction) SetTransactionNonce(txnNonce int64) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction fee.")
	}
	t.txn.TransactionNonce = txnNonce
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

type txnFeeOption struct {
	// stop estimate txn fee, usually if txn fee was 0, the createSmartContractTxn method would
	// estimate the txn fee by calling API from 0chain network. With this option, we could force
	// the txn to have zero fee for those exempt transactions.
	noEstimateFee bool
}

// FeeOption represents txn fee related option type
type FeeOption func(*txnFeeOption)

// WithNoEstimateFee would prevent txn fee estimation from remote
func WithNoEstimateFee() FeeOption {
	return func(o *txnFeeOption) {
		o.noEstimateFee = true
	}
}

func (t *Transaction) createSmartContractTxn(address, methodName string, input interface{}, value uint64, opts ...FeeOption) error {
	sn := transaction.SmartContractTxnData{Name: methodName, InputArgs: input}
	snBytes, err := json.Marshal(sn)
	if err != nil {
		return errors.Wrap(err, "create smart contract failed due to invalid data")
	}

	t.txn.TransactionType = transaction.TxnTypeSmartContract
	t.txn.ToClientID = address
	t.txn.TransactionData = string(snBytes)
	t.txn.Value = value

	if t.txn.TransactionFee > 0 {
		return nil
	}

	tf := &txnFeeOption{}
	for _, opt := range opts {
		opt(tf)
	}

	if tf.noEstimateFee {
		return nil
	}

	// TODO: check if transaction is exempt to avoid unnecessary fee estimation
	minFee, err := transaction.EstimateFee(t.txn, _config.chain.Miners, 0.2)
	if err != nil {
		logger.Logger.Error("failed estimate txn fee",
			zap.Any("txn", t.txn.Hash),
			zap.Error(err))
		return err
	}

	t.txn.TransactionFee = minFee

	return nil
}

func (t *Transaction) createFaucetSCWallet(walletStr string, methodName string, input []byte) (*zcncrypto.Wallet, error) {
	w, err := getWallet(walletStr)
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
		logging.Error("json unmarshal error on GetTransactionHash()")
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

	for _, sharder := range util.Shuffle(_config.chain.Sharders)[:numSharders] {
		go func(sharderurl string) {
			logging.Info("Query from ", sharderurl+query)
			url := fmt.Sprintf("%v%v", sharderurl, query)
			req, err := util.NewHTTPGetRequestContext(ctx, url)
			if err != nil {
				logging.Error(sharderurl, " new get request failed. ", err.Error())
				return
			}
			res, err := req.Get()
			if err != nil {
				logging.Error(sharderurl, " get error. ", err.Error())
			}
			result <- res
		}(sharder)
	}
}

func queryFromMinersContext(ctx context.Context, numMiners int, query string, result chan *util.GetResponse) {

	randomMiners := util.Shuffle(_config.chain.Miners)[:numMiners]
	for _, miner := range randomMiners {
		go func(minerurl string) {
			logging.Info("Query from ", minerurl+query)
			url := fmt.Sprintf("%v%v", minerurl, query)
			req, err := util.NewHTTPGetRequestContext(ctx, url)
			if err != nil {
				logging.Error(minerurl, " new get request failed. ", err.Error())
				return
			}
			res, err := req.Get()
			if err != nil {
				logging.Error(minerurl, " get error. ", err.Error())
			}
			result <- res
		}(miner)
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
		logging.Debug(rsp.Url + " " + rsp.Status)
		logging.Debug(rsp.Body)
		if rsp.StatusCode == http.StatusOK {
			var cfmLfb map[string]json.RawMessage
			err := json.Unmarshal([]byte(rsp.Body), &cfmLfb)
			if err != nil {
				logging.Error("txn confirmation parse error", err)
				continue
			}
			bH, err := getBlockHeaderFromTransactionConfirmation(txnHash, cfmLfb)
			if err != nil {
				logging.Error(err)
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
					logging.Error("round info parse error.", err)
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

func getBlockInfoByRound(numSharders int, round int64, content string) (*blockHeader, error) {
	numSharders = len(_config.chain.Sharders) // overwrite, use all
	resultC := make(chan *util.GetResponse, numSharders)
	queryFromSharders(numSharders, fmt.Sprintf("%vround=%v&content=%v", GET_BLOCK_INFO, round, content), resultC)
	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
		waitTime       = time.NewTimer(10 * time.Second)
		failedCount    int
	)

	type blockRound struct {
		Header blockHeader `json:"header"`
	}

	for i := 0; i < numSharders; i++ {
		select {
		case <-waitTime.C:
			return nil, stdErrors.New("failed to get block info by round with consensus, timeout")
		case rsp := <-resultC:
			logging.Debug(rsp.Url, rsp.Status)
			if failedCount*100/numSharders > 100-consensusThresh {
				return nil, stdErrors.New("failed to get block info by round with consensus, too many failures")
			}

			if rsp.StatusCode != http.StatusOK {
				logging.Debug(rsp.Url, "no round confirmation. Resp:", rsp.Body)
				failedCount++
				continue
			}

			var br blockRound
			err := json.Unmarshal([]byte(rsp.Body), &br)
			if err != nil {
				logging.Error("round info parse error. ", err)
				failedCount++
				continue
			}

			if len(br.Header.Hash) == 0 {
				failedCount++
				continue
			}

			h := br.Header.Hash
			roundConsensus[h]++
			if roundConsensus[h] > maxConsensus {
				maxConsensus = roundConsensus[h]
				if maxConsensus*100/numSharders >= consensusThresh {
					return &br.Header, nil
				}
			}
		}
	}

	return nil, stdErrors.New("failed to get block info by round with consensus")
}

func isBlockExtends(prevHash string, block *blockHeader) bool {
	data := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v:%v", block.MinerId, prevHash, block.CreationDate, block.Round,
		block.RoundRandomSeed, block.StateChangesCount, block.MerkleTreeRoot, block.ReceiptMerkleTreeRoot)
	h := encryption.Hash(data)
	return block.Hash == h
}

func validateChain(confirmBlock *blockHeader) bool {
	confirmRound := confirmBlock.Round
	logging.Debug("Confirmation round: ", confirmRound)
	currentBlockHash := confirmBlock.Hash
	round := confirmRound + 1
	for {
		nextBlock, err := getBlockInfoByRound(1, round, "header")
		if err != nil {
			logging.Info(err, " after a second falling thru to ", getMinShardersVerify(), "of ", len(_config.chain.Sharders), "Sharders")
			sys.Sleep(1 * time.Second)
			nextBlock, err = getBlockInfoByRound(getMinShardersVerify(), round, "header")
			if err != nil {
				logging.Error(err, " block chain stalled. waiting", defaultWaitSeconds, "...")
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
	value uint64) error {

	return t.createSmartContractTxn(VestingSmartContractAddress,
		function, vestingRequest{PoolID: common.Key(poolID)}, value)
}

func (t *Transaction) VestingTrigger(poolID string) (err error) {

	err = t.vestingPoolTxn(transaction.VESTING_TRIGGER, poolID, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

type VestingStopRequest struct {
	PoolID      string `json:"pool_id"`
	Destination string `json:"destination"`
}

func (t *Transaction) VestingStop(sr *VestingStopRequest) (err error) {

	err = t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_STOP, sr, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) VestingUnlock(poolID string) (err error) {

	err = t.vestingPoolTxn(transaction.VESTING_UNLOCK, poolID, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) VestingDelete(poolID string) (err error) {

	err = t.vestingPoolTxn(transaction.VESTING_DELETE, poolID, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

type scCollectReward struct {
	ProviderId   string `json:"provider_id"`
	ProviderType int    `json:"provider_type"`
}

type MinerSCLock struct {
	ID string `json:"id"`
}

type MinerSCUnlock struct {
	ID string `json:"id"`
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
		logging.Error("Unmarshal of transaction data to fileMeta failed, Maybe not a commit meta txn :", t.Hash)
		return false, nil
	}

	return metaOperation.MetaData.Hash == metaTxnData.MetaData.Hash, nil
}

//
// Storage SC transactions
//
