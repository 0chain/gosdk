package zcncore

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"net/http"
	"strconv"
	"time"
)

var (
	errNetwork          = errors.New("network error. host not reachable")
	errUserRejected     = errors.New("rejected by user")
	errAuthVerifyFailed = errors.New("verfication failed for auth response")
	errAuthTimeout      = errors.New("auth timed out")
	errAddSignature     = errors.New("error adding signature")
)
// TransactionCallback needs to be implemented by the caller for transaction related APIs
type TransactionCallback interface {
	OnTransactionComplete(t *Transaction, status int)
	OnVerifyComplete(t *Transaction, status int)
	OnAuthComplete(t *Transaction, status int)
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

// TransactionScheme implements few methods for block chain
type TransactionScheme interface {
	// SetTransactionCallback implements storing the callback
	// used to call after the transaction or verification is completed
	SetTransactionCallback(cb TransactionCallback) error
	// Send implements sending token to a given clientid
	Send(toClientID string, val int64, desc string) error
	// StoreData implements store the data to blockchain
	StoreData(data string) error
	// ExecuteSmartContract impements the Faucet Smart contract
	ExecuteSmartContract(address, methodName, input string, val int64) error
	// GetTransactionHash implements retrieval of hash of the submitted transaction
	GetTransactionHash() string
	// LockTokens implements the lock token.
	LockTokens(val int64, durationHr int64, durationMin int) error
	// UnlockTokens implements unlocking of earlier locked tokens.
	UnlockTokens(poolID string) error
	// Stake implementes token to be stake on clientID
	Stake(clientID string, val int64) error
	// DeleteStake implements deleteing staked tokens
	DeleteStake(clientID, poolID string) error
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
}

func signFn(hash string) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	sigScheme.SetPrivateKey(_config.wallet.Keys[0].PrivateKey)
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
	return ""
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
			Logger.Info("Submitting", txnTypeString(t.txn.TransactionType), "transaction to", minerurl)
			req, err := util.NewHTTPPostRequest(url, t.txn)
			if err != nil {
				Logger.Error(minerurl, "new post request failed. ", err.Error())
				return
			}
			res, err := req.Post()
			if err != nil {
				Logger.Error(minerurl, "submit transaction error. ", err.Error())
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
			return nil, fmt.Errorf("auth url not set")
		}
		Logger.Info("New transaction interface with auth")
		return newTransactionWithAuth(cb, txnFee)
	}
	Logger.Info("New transaction interface")
	return newTransaction(cb, txnFee)
}
func (t *Transaction) SetTransactionCallback(cb TransactionCallback) error {
	if t.txnStatus != StatusUnknown {
		return fmt.Errorf("transaction already exists. cannot set transaction hash.")
	}
	t.txnCb = cb
	return nil
}
func (t *Transaction) SetTransactionFee(txnFee int64) error {
	if t.txnStatus != StatusUnknown {
		return fmt.Errorf("transaction already exists. cannot set transaction fee.")
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
		return fmt.Errorf("create smart contract failed due to invalid data. %s", err.Error())
	}
		t.txn.TransactionType = transaction.TxnTypeSmartContract
	t.txn.ToClientID = address
		t.txn.TransactionData = string(snBytes)
	t.txn.Value = value
	return nil
}
func (t *Transaction) ExecuteSmartContract(address, methodName, input string, val int64) error {
	err := t.createSmartContractTxn(address, methodName, input, val)
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
		return fmt.Errorf("transaction already exists. cannot set transaction hash.")
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

func queryFromSharders(numSharders int, query string, result chan *util.GetResponse) {
	randomShaders := util.GetRandom(_config.chain.Sharders, numSharders)
	for _, sharder := range randomShaders {
		go func(sharderurl string) {
			Logger.Info("Query from", sharderurl+query)
			url := fmt.Sprintf("%v%v", sharderurl, query)
			req, err := util.NewHTTPGetRequest(url)
			if err != nil {
				Logger.Error(sharderurl, "new get request failed. ", err.Error())
				return
			}
			res, err := req.Get()
			if err != nil {
				Logger.Error(sharderurl, "get error. ", err.Error())
			}
			result <- res
			return
		}(sharder)
	}
}
func getTransactionConfirmation(numSharders int, txnHash string) (map[string]json.RawMessage, string, *blockHeader, error) {
	result := make(chan *util.GetResponse)
	defer close(result)
	queryFromSharders(numSharders, fmt.Sprintf("%v%v&content=lfb", TXN_VERIFY_URL, txnHash), result)
	maxConfirmation := int(0)
	txnConfirmations := make(map[string]int)
	var confirmedTxn map[string]json.RawMessage
	var blockHash string
	var lfb blockHeader
	for i := 0; i < numSharders; i++ {
		select {
		case rsp := <-result:
			Logger.Debug(rsp.Url, rsp.Status)
			Logger.Error(rsp.Body)
			if rsp.StatusCode == http.StatusOK {
				var cfmLfb map[string]json.RawMessage
				err := json.Unmarshal([]byte(rsp.Body), &cfmLfb)
				if err != nil {
					Logger.Error("txn confirmation parse error", err)
					continue
				}
				if cfm, ok := cfmLfb["confirmation"]; ok {
					var objmap map[string]json.RawMessage
					err := json.Unmarshal([]byte(cfm), &objmap)
					if err != nil {
						Logger.Error("txn confirmation parse error", err)
						continue
					}
					if _, ok := objmap["txn"]; ok {
						h := encryption.FastHash([]byte(objmap["txn"]))
						txnConfirmations[h]++
						if txnConfirmations[h] > maxConfirmation {
							maxConfirmation = txnConfirmations[h]
							confirmedTxn = objmap
							if bh, ok := objmap["block_hash"]; ok {
								blockHash, _ = strconv.Unquote(string(bh))
							}
						}
					} else {
						Logger.Debug(rsp.Url, "No transaction confirmation")
					}
				} else if lfbRaw, ok := cfmLfb["latest_finalized_block"]; ok {
					err := json.Unmarshal([]byte(lfbRaw), &lfb)
					if err != nil {
						Logger.Error("round info parse error", err)
						continue
					}
				}
			} else {
				Logger.Error(rsp.Body)
			}
		}
	}
	if confirmedTxn == nil {
		return nil, "", &lfb, fmt.Errorf("transaction not found")
	}
	return confirmedTxn, blockHash, &lfb, nil
}
func parseBlockRound(txn map[string]json.RawMessage) int64 {
	if r, ok := txn["round"]; ok {
		round, err := strconv.ParseInt(string(r), 10, 64)
		if err != nil {
			Logger.Error("invalid round number")
			return 0
		}
		return round
	}
	return 0
}
func getBlockInfoByRound(numSharders int, round int64, content string) (*blockHeader, error) {
	result := make(chan *util.GetResponse)
	defer close(result)
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
					Logger.Error("round info parse error", err)
					continue
				}
				if header, ok := objmap["header"]; ok {
					err := json.Unmarshal([]byte(header), &objmap)
					if err != nil {
						Logger.Error("round info parse error", err)
						continue
					}
					if hash, ok := objmap["hash"]; ok {
						h := encryption.FastHash([]byte(hash))
						roundConsensus[h]++
						if roundConsensus[h] > maxConsensus {
							maxConsensus = roundConsensus[h]
							err := json.Unmarshal([]byte(header), &blkHdr)
							if err != nil {
								Logger.Error("round info parse error", err)
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
		return nil, fmt.Errorf("round info not found")
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
func validateChain(conf map[string]json.RawMessage, confirmBlockhash string) bool {
	confirmRound := parseBlockRound(conf)
	Logger.Debug("Confirmation round: ", confirmRound)
	currentBlockHash := confirmBlockhash
	round := confirmRound + 1
	for {
		nextBlock, err := getBlockInfoByRound(1, round, "header")
		if err != nil {
			Logger.Info(err, "after a second falling thru to ", getMinShardersVerify(), "of ", len(_config.chain.Sharders), "Sharders")
			time.Sleep(1 * time.Second)
			nextBlock, err = getBlockInfoByRound(getMinShardersVerify(), round, "header")
			if err != nil {
				Logger.Error("err", "block chain stalled. waiting", defaultWaitSeconds, "...")
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
		return fmt.Errorf("invalid transaction. cannot be verified")
	}
	if t.txnHash == "" && t.txnStatus == StatusSuccess {
		h := t.GetTransactionHash()
		if h == "" {
			return fmt.Errorf("invalid transaction. cannot be verified")
		}
	}
	// If transaction is verify only start from current time
	if t.txn.CreationDate == 0 {
		t.txn.CreationDate = common.Now()
	}

				go func() {
		for {
			// Get transaction confirmation from random sharder
			confirmation, blockHash, lfb, err := getTransactionConfirmation(1, t.txnHash)
							if err != nil {
				tn := common.Now()
				Logger.Info(err, "now:", tn, "LFB creation time:", lfb.CreationDate)
				if util.MaxInt64(lfb.CreationDate, tn) < (t.txn.CreationDate + int64(defaultTxnExpirationSeconds)) {
					Logger.Info("falling back to", getMinShardersVerify(), "of ", len(_config.chain.Sharders), "Sharders")
					confirmation, blockHash, lfb, err = getTransactionConfirmation(getMinShardersVerify(), t.txnHash)
							if err != nil {
						if t.isTransactionExpired(lfb.CreationDate, tn) {
							t.completeVerify(StatusError, "", fmt.Errorf(`{"error": "verify transaction failed"`))
									return
						}
							continue
					}
				} else {
					if t.isTransactionExpired(lfb.CreationDate, tn) {
						t.completeVerify(StatusError, "", fmt.Errorf(`{"error": "verify transaction failed"`))
				return
		}
						continue
					}
						}
			valid := validateChain(confirmation, blockHash)
			if valid {
				output, err := json.Marshal(confirmation)
				if err != nil {
					t.completeVerify(StatusError, "", fmt.Errorf(`{"error": "transaction confirmation json marshal error"`))
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
func (t *Transaction) createStakeTxn(clientID string, val int64) error {
	input := make(map[string]interface{})
	input["id"] = clientID
	return t.createSmartContractTxn(StakeSmartContractAddress, transaction.STAKE, input, val)
}

func (t *Transaction) Stake(clientID string, val int64) error {
	err := t.createStakeTxn(clientID, val)
	if err != nil {
		return err
	}
	go func() {
		t.submitTxn()
	}()
	return nil
}

func (t *Transaction) createDeleteStakeTxn(clientID, poolID string) error {
	input := make(map[string]interface{})
	input["id"] = clientID
	input["pool_id"] = poolID
	return t.createSmartContractTxn(StakeSmartContractAddress, transaction.DELETE_STAKE, input, 0)
}

func (t *Transaction) DeleteStake(clientID, poolID string) error {
	err := t.createDeleteStakeTxn(clientID, poolID)
	if err != nil {
		return err
	}
	go func() {
		t.submitTxn()
	}()
	return nil
}
