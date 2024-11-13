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
	"github.com/0chain/gosdk/core/node"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/fileref"

	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
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
	SharderConsensous       int      `json:"sharder_consensous"`
	IsSplitWallet           bool     `json:"is_split_wallet"`
}

var Sharders *node.NodeHolder

// InitZCNSDK initializes the SDK given block worker and signature scheme provided.
//   - blockWorker: block worker, which is the url for the DNS service for locating miners and sharders
//   - signscheme: signature scheme to be used for signing the transactions
//   - configs: configuration options
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
	_config.isSplitWallet = _config.chain.IsSplitWallet
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
		SharderConsensous:       _config.chain.SharderConsensous,
	}

	conf.InitClientConfig(cfg)

	return nil
}

func IsSplitWallet() bool {
	return _config.isSplitWallet
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

// Transaction data structure that provides the transaction details
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

func SignWithKey(privateKey, hash string) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme("bls0chain")
	err := sigScheme.SetPrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}

func VerifyWithKey(pubKey, signature, hash string) (bool, error) {
	sigScheme := zcncrypto.NewSignatureScheme("bls0chain")
	err := sigScheme.SetPublicKey(pubKey)
	if err != nil {
		return false, err
	}
	return sigScheme.Verify(signature, hash)
}

var SignFn = func(hash string) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(_config.wallet.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}

var AddSignature = func(privateKey, signature string, hash string) (string, error) {
	var (
		ss  = zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
		err error
	)

	err = ss.SetPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	return ss.Add(signature, hash)
}

func signWithWallet(hash string, wi interface{}) (string, error) {
	w, ok := wi.(*zcncrypto.Wallet)

	if !ok {
		fmt.Printf("Error in casting to wallet")
		return "", errors.New("", "error in casting to wallet")
	}
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(w.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}
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

// Output implements the output of transaction
func (t *Transaction) Output() []byte {
	return []byte(t.txnOut)
}

// Hash implements the hash of transaction
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
		node.Cache.Evict(t.txn.ClientID)
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
		g.err = errors.New("get_nonce", "failed respond nonce") //nolint
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
		nonce = node.Cache.GetNextNonce(t.txn.ClientID)
	} else {
		node.Cache.Set(t.txn.ClientID, nonce)
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
			node.Cache.Evict(t.txn.ClientID)
			return
		}
	}

	var (
		randomMiners = GetStableMiners()
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

			if res.StatusCode != http.StatusOK {
				logging.Error(minerurl, " submit transaction failed with status code ", res.StatusCode)
				if int(atomic.AddInt32(&failedCount, 1)) == minersN {
					resultC <- res
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
		node.Cache.Evict(t.txn.ClientID)
		ResetStableMiners()
		return
	case ret := <-resultC:
		logging.Debug("finish txn submitting, ", ret.Url, ", Status: ", ret.Status, ", output:", ret.Body)
		if ret.StatusCode == http.StatusOK {
			t.completeTxn(StatusSuccess, ret.Body, nil)
		} else {
			t.completeTxn(StatusError, "", fmt.Errorf("submit transaction failed. %s", ret.Body))
			node.Cache.Evict(t.txn.ClientID)
			ResetStableMiners()
		}
	}
}

// SetTransactionCallback implements storing the callback
func (t *Transaction) SetTransactionCallback(cb TransactionCallback) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction hash.")
	}
	t.txnCb = cb
	return nil
}

// SetTransactionNonce implements method to set the transaction nonce
func (t *Transaction) SetTransactionNonce(txnNonce int64) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction fee.")
	}
	t.txn.TransactionNonce = txnNonce
	return nil
}

// StoreData implements store the data to blockchain
func (t *Transaction) StoreData(data string) error {
	go func() {
		t.txn.TransactionType = transaction.TxnTypeData
		t.txn.TransactionData = data
		t.setNonceAndSubmit()
	}()
	return nil
}

type TxnFeeOption struct {
	// stop estimate txn fee, usually if txn fee was 0, the createSmartContractTxn method would
	// estimate the txn fee by calling API from 0chain network. With this option, we could force
	// the txn to have zero fee for those exempt transactions.
	noEstimateFee bool
}

// FeeOption represents txn fee related option type
type FeeOption func(*TxnFeeOption)

// WithNoEstimateFee would prevent txn fee estimation from remote
func WithNoEstimateFee() FeeOption {
	return func(o *TxnFeeOption) {
		o.noEstimateFee = true
	}
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
			nonce = node.Cache.GetNextNonce(t.txn.ClientID)
		} else {
			node.Cache.Set(t.txn.ClientID, nonce)
		}
		t.txn.TransactionNonce = nonce
		err = t.txn.ComputeHashAndSignWithWallet(signWithWallet, w)
		if err != nil {
			return
		}
		fmt.Printf("submitted transaction\n")
		t.submitTxn()
	}()
	return nil
}

// SetTransactionHash implements verify a previous transaction status
//   - hash: transaction hash
func (t *Transaction) SetTransactionHash(hash string) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction hash.")
	}
	t.txnHash = hash
	return nil
}

// GetTransactionHash implements retrieval of hash of the submitted transaction
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

func getBlockInfoByRound(round int64, content string) (*blockHeader, error) {
	numSharders := len(Sharders.Healthy()) // overwrite, use all
	resultC := make(chan *util.GetResponse, numSharders)
	Sharders.QueryFromSharders(numSharders, fmt.Sprintf("%vround=%v&content=%v", GET_BLOCK_INFO, round, content), resultC)
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
			if rsp == nil {
				logging.Error("nil response")
				continue
			}
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
		nextBlock, err := getBlockInfoByRound(round, "header")
		if err != nil {
			logging.Info(err, " after a second falling thru to ", getMinShardersVerify(), "of ", len(_config.chain.Sharders), "Sharders", len(Sharders.Healthy()), "Healthy sharders")
			sys.Sleep(1 * time.Second)
			nextBlock, err = getBlockInfoByRound(round, "header")
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

// GetVerifyOutput implements the verification output from sharders
func (t *Transaction) GetVerifyOutput() string {
	if t.verifyStatus == StatusSuccess {
		return t.verifyOut
	}
	return ""
}

// GetTransactionError implements error string in case of transaction failure
func (t *Transaction) GetTransactionError() string {
	if t.txnStatus != StatusSuccess {
		return t.txnError.Error()
	}
	return ""
}

// GetVerifyError implements error string in case of verify failure error
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

type VestingStopRequest struct {
	PoolID      string `json:"pool_id"`
	Destination string `json:"destination"`
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

type CommitMetaData struct {
	CrudType string
	MetaData *ConsolidatedFileMeta
}

type CommitMetaResponse struct {
	TxnID    string
	MetaData *ConsolidatedFileMeta
}

type ConsolidatedFileMeta struct {
	Name            string
	Type            string
	Path            string
	LookupHash      string
	Hash            string
	MimeType        string
	Size            int64
	NumBlocks       int64
	ActualFileSize  int64
	ActualNumBlocks int64
	EncryptedKey    string

	ActualThumbnailSize int64
	ActualThumbnailHash string

	Collaborators []fileref.Collaborator
}

func VerifyContentHash(metaTxnDataJSON string) (bool, error) {
	var metaTxnData CommitMetaResponse
	err := json.Unmarshal([]byte(metaTxnDataJSON), &metaTxnData)
	if err != nil {
		return false, errors.New("metaTxnData_decode_error", "Unable to decode metaTxnData json")
	}

	t, err := transaction.VerifyTransaction(metaTxnData.TxnID, blockchain.GetSharders())
	if err != nil {
		return false, errors.New("fetch_txm_details", "Unable to fetch txn details")
	}

	var metaOperation CommitMetaData
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
