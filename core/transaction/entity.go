// Provides low-level functions and types to work with the native smart contract transactions.
package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/sys"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/util"
	lru "github.com/hashicorp/golang-lru"
)

var Logger logger.Logger

const STORAGE_SCADDRESS = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
const MINERSC_SCADDRESS = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9"
const ZCNSC_SCADDRESS = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712e0"

const TXN_SUBMIT_URL = "v1/transaction/put"
const TXN_VERIFY_URL = "v1/transaction/get/confirmation?hash="
const BLOCK_BY_ROUND_URL = "v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/block?round="

const (
	TxnSuccess         = 1 // Indicates the transaction is successful in updating the state or smart contract
	TxnChargeableError = 2 // Indicates the transaction is successful in updating the state or smart contract
	TxnFail            = 3 // Indicates a transaction has failed to update the state or smart contract
)

// TxnReceipt - a transaction receipt is a processed transaction that contains the output
type TxnReceipt struct {
	Transaction *Transaction
}

// SmartContractTxnData data structure to hold the smart contract transaction data
type SmartContractTxnData struct {
	Name      string      `json:"name"`
	InputArgs interface{} `json:"input"`
}

type StorageAllocation struct {
	ID             string  `json:"id"`
	DataShards     int     `json:"data_shards"`
	ParityShards   int     `json:"parity_shards"`
	Size           int64   `json:"size"`
	Expiration     int64   `json:"expiration_date"`
	Owner          string  `json:"owner_id"`
	OwnerPublicKey string  `json:"owner_public_key"`
	ReadRatio      *Ratio  `json:"read_ratio"`
	WriteRatio     *Ratio  `json:"write_ratio"`
	MinLockDemand  float64 `json:"min_lock_demand"`
}
type Ratio struct {
	ZCN  int64 `json:"zcn"`
	Size int64 `json:"size"`
}
type RoundBlockHeader struct {
	Version               string `json:"version"`
	CreationDate          int64  `json:"creation_date"`
	Hash                  string `json:"block_hash"`
	PreviousBlockHash     string `json:"previous_block_hash"`
	MinerID               string `json:"miner_id"`
	Round                 int64  `json:"round"`
	RoundRandomSeed       int64  `json:"round_random_seed"`
	MerkleTreeRoot        string `json:"merkle_tree_root"`
	StateChangesCount     int    `json:"state_changes_count"`
	StateHash             string `json:"state_hash"`
	ReceiptMerkleTreeRoot string `json:"receipt_merkle_tree_root"`
	NumberOfTxns          int64  `json:"num_txns"`
}

type Block struct {
	Hash                  string `json:"hash" gorm:"uniqueIndex:idx_bhash"`
	Version               string `json:"version"`
	CreationDate          int64  `json:"creation_date" gorm:"index:idx_bcreation_date"`
	Round                 int64  `json:"round" gorm:"index:idx_bround"`
	MinerID               string `json:"miner_id"`
	RoundRandomSeed       int64  `json:"round_random_seed"`
	MerkleTreeRoot        string `json:"merkle_tree_root"`
	StateHash             string `json:"state_hash"`
	ReceiptMerkleTreeRoot string `json:"receipt_merkle_tree_root"`
	NumTxns               int    `json:"num_txns"`
	MagicBlockHash        string `json:"magic_block_hash"`
	PrevHash              string `json:"prev_hash"`
	Signature             string `json:"signature"`
	ChainId               string `json:"chain_id"`
	StateChangesCount     int    `json:"state_changes_count"`
	RunningTxnCount       string `json:"running_txn_count"`
	RoundTimeoutCount     int    `json:"round_timeout_count"`
}

const (
	NEW_ALLOCATION_REQUEST    = "new_allocation_request"
	NEW_FREE_ALLOCATION       = "free_allocation_request"
	UPDATE_ALLOCATION_REQUEST = "update_allocation_request"
	LOCK_TOKEN                = "lock"
	UNLOCK_TOKEN              = "unlock"

	ADD_FREE_ALLOCATION_ASSIGNER = "add_free_storage_assigner"

	// Storage SC
	STORAGESC_FINALIZE_ALLOCATION       = "finalize_allocation"
	STORAGESC_CANCEL_ALLOCATION         = "cancel_allocation"
	STORAGESC_CREATE_ALLOCATION         = "new_allocation_request"
	STORAGESC_CREATE_READ_POOL          = "new_read_pool"
	STORAGESC_READ_POOL_LOCK            = "read_pool_lock"
	STORAGESC_READ_POOL_UNLOCK          = "read_pool_unlock"
	STORAGESC_STAKE_POOL_LOCK           = "stake_pool_lock"
	STORAGESC_STAKE_POOL_UNLOCK         = "stake_pool_unlock"
	STORAGESC_UPDATE_BLOBBER_SETTINGS   = "update_blobber_settings"
	STORAGESC_UPDATE_VALIDATOR_SETTINGS = "update_validator_settings"
	STORAGESC_UPDATE_ALLOCATION         = "update_allocation_request"
	STORAGESC_WRITE_POOL_LOCK           = "write_pool_lock"
	STORAGESC_WRITE_POOL_UNLOCK         = "write_pool_unlock"
	STORAGESC_UPDATE_SETTINGS           = "update_settings"
	ADD_HARDFORK                        = "add_hardfork"
	STORAGESC_COLLECT_REWARD            = "collect_reward"
	STORAGESC_KILL_BLOBBER              = "kill_blobber"
	STORAGESC_KILL_VALIDATOR            = "kill_validator"
	STORAGESC_SHUTDOWN_BLOBBER          = "shutdown_blobber"
	STORAGESC_SHUTDOWN_VALIDATOR        = "shutdown_validator"
	STORAGESC_RESET_BLOBBER_STATS       = "reset_blobber_stats"
	STORAGESC_RESET_ALLOCATION_STATS    = "reset_allocation_stats"

	MINERSC_LOCK             = "addToDelegatePool"
	MINERSC_UNLOCK           = "deleteFromDelegatePool"
	MINERSC_MINER_SETTINGS   = "update_miner_settings"
	MINERSC_SHARDER_SETTINGS = "update_sharder_settings"
	MINERSC_UPDATE_SETTINGS  = "update_settings"
	MINERSC_UPDATE_GLOBALS   = "update_globals"
	MINERSC_MINER_DELETE     = "delete_miner"
	MINERSC_SHARDER_DELETE   = "delete_sharder"
	MINERSC_COLLECT_REWARD   = "collect_reward"
	MINERSC_KILL_MINER       = "kill_miner"
	MINERSC_KILL_SHARDER     = "kill_sharder"

	// Faucet SC
	FAUCETSC_UPDATE_SETTINGS = "update-settings"

	// ZCNSC smart contract

	ZCNSC_UPDATE_GLOBAL_CONFIG     = "update-global-config"
	ZCNSC_UPDATE_AUTHORIZER_CONFIG = "update-authorizer-config"
	ZCNSC_ADD_AUTHORIZER           = "add-authorizer"
	ZCNSC_AUTHORIZER_HEALTH_CHECK  = "authorizer-health-check"
	ZCNSC_DELETE_AUTHORIZER        = "delete-authorizer"
	ZCNSC_COLLECT_REWARD           = "collect-rewards"
	ZCNSC_LOCK                     = "add-to-delegate-pool"
	ZCNSC_UNLOCK                   = "delete-from-delegate-pool"

	ESTIMATE_TRANSACTION_COST = `/v1/estimate_txn_fee`
	FEES_TABLE                = `/v1/fees_table`
)

type SignFunc = func(msg string, clientId ...string) (string, error)
type VerifyFunc = func(publicKey, signature, msgHash string) (bool, error)
type SignWithWallet = func(msg string, wallet interface{}) (string, error)

var cache *lru.Cache

func init() {
	var err error
	cache, err = lru.New(100)
	if err != nil {
		fmt.Println("caching Initilization failed, err:", err)
	}
}

func NewTransactionEntity(clientID string, chainID string, publicKey string, nonce int64) *Transaction {
	txn := &Transaction{}
	txn.Version = "1.0"
	txn.ClientID = clientID
	txn.CreationDate = int64(common.Now())
	txn.ChainID = chainID
	txn.PublicKey = publicKey
	txn.TransactionNonce = nonce
	return txn
}

func (t *Transaction) ComputeHashAndSignWithWallet(signHandler SignWithWallet, signingWallet interface{}) error {
	t.ComputeHashData()
	var err error
	t.Signature, err = signHandler(t.Hash, signingWallet)
	if err != nil {
		return err
	}
	return nil
}

func (t *Transaction) ComputeHashAndSign(signHandler SignFunc) error {
	t.ComputeHashData()
	var err error
	t.Signature, err = signHandler(t.Hash)
	if err != nil {
		return err
	}
	return nil
}

func (t *Transaction) ComputeHashData() {
	hashdata := fmt.Sprintf("%v:%v:%v:%v:%v:%v", t.CreationDate, t.TransactionNonce, t.ClientID,
		t.ToClientID, t.Value, encryption.Hash(t.TransactionData))
	t.Hash = encryption.Hash(hashdata)
}

func (t *Transaction) DebugJSON() []byte {
	jsonByte, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		panic(err) // This JSONify function only supposed to be debug-only anyway.
	}
	return jsonByte
}

// GetHash - implement interface
func (rh *TxnReceipt) GetHash() string {
	return rh.Transaction.OutputHash
}

/*GetHashBytes - implement Hashable interface */
func (rh *TxnReceipt) GetHashBytes() []byte {
	return util.HashStringToBytes(rh.Transaction.OutputHash)
}

// NewTransactionReceipt - create a new transaction receipt
func NewTransactionReceipt(t *Transaction) *TxnReceipt {
	return &TxnReceipt{Transaction: t}
}

// VerifySigWith verify the signature with the given public key and handler
func (t *Transaction) VerifySigWith(pubkey string, verifyHandler VerifyFunc) (bool, error) {
	// Store the hash
	hash := t.Hash
	t.ComputeHashData()
	if t.Hash != hash {
		return false, errors.New("verify_transaction", fmt.Sprintf(`{"error":"hash_mismatch", "expected":"%v", "actual":%v"}`, t.Hash, hash))
	}
	return verifyHandler(pubkey, t.Signature, t.Hash)
}

// SendTransactionSync sends transactions to all miners in parallel and returns as soon as minSubmit is reached.
func SendTransactionSync(txn *Transaction, miners []string, minSubmit int) error {
	// Context with 1-minute timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	successCh := make(chan struct{}, len(miners))
	failCh := make(chan error, len(miners))

	// Send transactions in parallel
	for _, miner := range miners {
		url := fmt.Sprintf("%v/%v", miner, TXN_SUBMIT_URL)
		go func(url string) {
			err := sendTransactionToURL(ctx, url, txn)
			if err != nil {
				failCh <- err
			} else {
				successCh <- struct{}{}
			}
		}(url)
	}

	// Track successful responses
	successCount := 0
	for {
		select {
		case <-ctx.Done(): // If the context times out
			return fmt.Errorf("operation timed out: %v", ctx.Err())
		case <-successCh:
			successCount++
			if (successCount*100)/len(miners) >= minSubmit {
				cancel() // Cancel remaining requests
				return nil
			}
		case err := <-failCh:
			// Log the error (optional)
			fmt.Printf("Transaction failed: %v\n", err)
		}
	}
}

// sendTransactionToURL sends a transaction to a given URL and respects cancellation.
func sendTransactionToURL(ctx context.Context, url string, txn *Transaction) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil) // Assume `txn` is serialized inside
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}
	return fmt.Errorf("transaction failed with status: %d", resp.StatusCode)
}

type cachedObject struct {
	Expiration time.Duration
	Value      interface{}
}

func retriveFromTable(table map[string]map[string]int64, txnName, toAddress string) (uint64, error) {
	var fees uint64
	if val, ok := table[toAddress]; ok {
		fees = uint64(val[txnName])
	} else {
		if txnName == "transfer" {
			fees = uint64(table["transfer"]["transfer"])
		} else {
			return 0, fmt.Errorf("failed to get fees for txn %s", txnName)
		}
	}
	return fees, nil
}

// EstimateFee estimates transaction fee
func EstimateFee(txn *Transaction, miners []string, reqPercent ...float32) (uint64, error) {
	const minReqNum = 3
	var reqN int

	if len(reqPercent) > 0 {
		reqN = int(reqPercent[0] * float32(len(miners)))
	}

	txData := txn.TransactionData

	var sn SmartContractTxnData
	err := json.Unmarshal([]byte(txData), &sn)
	if err != nil {
		return 0, err
	}

	txnName := sn.Name
	txnName = strings.ToLower(txnName)
	toAddress := txn.ToClientID

	reqN = util.MaxInt(minReqNum, reqN)
	reqN = util.MinInt(reqN, len(miners))
	randomMiners := util.Shuffle(miners)[:reqN]

	// Retrieve the object from the cache
	cached, ok := cache.Get(FEES_TABLE)
	if ok {
		cachedObj, ok := cached.(*cachedObject)
		if ok {
			table := cachedObj.Value.(map[string]map[string]int64)
			fees, err := retriveFromTable(table, txnName, toAddress)
			if err != nil {
				return 0, err
			}
			return fees, nil
		}
	}

	table, err := GetFeesTable(randomMiners, reqPercent...)
	if err != nil {
		return 0, err
	}

	fees, err := retriveFromTable(table, txnName, toAddress)
	if err != nil {
		return 0, err
	}

	cache.Add(FEES_TABLE, &cachedObject{
		Expiration: 30 * time.Hour,
		Value:      table,
	})

	return fees, nil
}

// GetFeesTable get fee tables
func GetFeesTable(miners []string, reqPercent ...float32) (map[string]map[string]int64, error) {
	const minReqNum = 3
	var reqN int

	if len(reqPercent) > 0 {
		reqN = int(reqPercent[0] * float32(len(miners)))
	}

	reqN = util.MaxInt(minReqNum, reqN)
	reqN = util.MinInt(reqN, len(miners))
	randomMiners := util.Shuffle(miners)[:reqN]

	var (
		feesC = make(chan string, reqN)
		errC  = make(chan error, reqN)
	)

	wg := &sync.WaitGroup{}
	wg.Add(len(randomMiners))

	for _, miner := range randomMiners {
		go func(minerUrl string) {
			defer wg.Done()

			url := minerUrl + FEES_TABLE
			req, err := util.NewHTTPGetRequest(url)
			if err != nil {
				errC <- fmt.Errorf("create request failed, url: %s, err: %v", url, err)
				return
			}

			res, err := req.Get()
			if err != nil {
				errC <- fmt.Errorf("request failed, url: %s, err: %v", url, err)
				return
			}

			if res.StatusCode == http.StatusOK {
				feesC <- res.Body
				return
			}

			feesC <- ""

		}(miner)
	}

	// wait for requests to complete
	wg.Wait()
	close(feesC)
	close(errC)

	feesCount := make(map[string]int, reqN)
	for f := range feesC {
		feesCount[f]++
	}

	if len(feesCount) > 0 {
		var (
			max  int
			fees string
		)

		for f, count := range feesCount {
			if f != "" && count > max {
				max = count
				fees = f
			}
		}

		feesTable := make(map[string]map[string]int64)
		err := json.Unmarshal([]byte(fees), &feesTable)
		if err != nil {
			return nil, errors.New("failed to get fees table", err.Error())
		}

		return feesTable, nil
	}

	errs := make([]string, 0, reqN)
	for err := range errC {
		errs = append(errs, err.Error())
	}

	return nil, errors.New("failed to get fees table", strings.Join(errs, ","))
}

func SmartContractTxn(scAddress string, sn SmartContractTxnData, clients ...string) (
	hash, out string, nonce int64, txn *Transaction, err error) {
	return SmartContractTxnValue(scAddress, sn, 0, clients...)
}

func SmartContractTxnValue(scAddress string, sn SmartContractTxnData, value uint64, clients ...string) (
	hash, out string, nonce int64, txn *Transaction, err error) {

	return SmartContractTxnValueFeeWithRetry(scAddress, sn, value, client.TxnFee(), clients...)
}

func SmartContractTxnValueFeeWithRetry(scAddress string, sn SmartContractTxnData,
	value, fee uint64, clients ...string) (hash, out string, nonce int64, t *Transaction, err error) {
	hash, out, nonce, t, err = SmartContractTxnValueFee(scAddress, sn, value, fee, clients...)

	if err != nil && strings.Contains(err.Error(), "invalid transaction nonce") {
		return SmartContractTxnValueFee(scAddress, sn, value, fee, clients...)
	}
	return
}

func SmartContractTxnValueFee(scAddress string, sn SmartContractTxnData,
	value, fee uint64, clients ...string) (hash, out string, nonce int64, t *Transaction, err error) {

	clientId := client.Id(clients...)
	if len(clients) > 0 && clients[0] != "" {
		clientId = clients[0]
	}

	var requestBytes []byte
	if requestBytes, err = json.Marshal(sn); err != nil {
		return
	}

	cfg, err := conf.GetClientConfig()
	if err != nil {
		return
	}

	nodeClient, err := client.GetNode()
	if err != nil {
		return
	}

	txn := NewTransactionEntity(client.Id(clientId),
		cfg.ChainID, client.PublicKey(clientId), nonce)

	txn.TransactionData = string(requestBytes)
	txn.ToClientID = scAddress
	txn.Value = value
	txn.TransactionFee = fee
	txn.TransactionType = TxnTypeSmartContract

	if len(clients) > 0 {
		txn.ClientID = clients[0]
	}
	if len(clients) > 1 {
		txn.ToClientID = clients[1]
		txn.TransactionType = TxnTypeSend
	}

	// adjust fees if not set
	if fee == 0 {
		fee, err = EstimateFee(txn, nodeClient.Network().Miners, 0.2)
		if err != nil {
			Logger.Error("failed to estimate txn fee",
				zap.Error(err),
				zap.Any("txn", txn))
			return
		}
		txn.TransactionFee = fee
	}

	if txn.TransactionNonce == 0 {
		txn.TransactionNonce = client.Cache.GetNextNonce(txn.ClientID)
	}

	if err = txn.ComputeHashAndSign(client.Sign); err != nil {
		return
	}

	msg := fmt.Sprintf("executing transaction '%s' with hash %s ", sn.Name, txn.Hash)
	Logger.Info(msg)
	Logger.Info("estimated txn fee: ", txn.TransactionFee)

	err = SendTransactionSync(txn, nodeClient.GetStableMiners(), cfg.MinSubmit)
	if err != nil {
		Logger.Info("transaction submission failed", zap.Error(err))
		client.Cache.Evict(txn.ClientID)
		nodeClient.ResetStableMiners()
		return
	}

	var (
		querySleepTime = time.Duration(cfg.QuerySleepTime) * time.Second
		retries        = 0
	)

	sys.Sleep(querySleepTime)

	for retries < cfg.MaxTxnQuery {
		t, err = VerifyTransaction(txn.Hash)
		if err == nil {
			break
		}
		retries++
		sys.Sleep(querySleepTime)
	}

	if err != nil {
		Logger.Error("Error verifying the transaction", err.Error(), txn.Hash)
		client.Cache.Evict(txn.ClientID)
		return
	}

	if t == nil {
		return "", "", 0, txn, errors.New("transaction_validation_failed",
			"Failed to get the transaction confirmation")
	}

	if t.Status == TxnFail {
		return t.Hash, t.TransactionOutput, 0, t, errors.New("", t.TransactionOutput)
	}

	if t.Status == TxnChargeableError {
		return t.Hash, t.TransactionOutput, t.TransactionNonce, t, errors.New("", t.TransactionOutput)
	}

	return t.Hash, t.TransactionOutput, t.TransactionNonce, t, nil
}
