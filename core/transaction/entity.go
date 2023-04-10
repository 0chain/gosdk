package transaction

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/util"
)

const TXN_SUBMIT_URL = "v1/transaction/put"
const TXN_VERIFY_URL = "v1/transaction/get/confirmation?hash="
const BLOCK_BY_ROUND_URL = "v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/block?round="

const (
	TxnSuccess         = 1 // Indicates the transaction is successful in updating the state or smart contract
	TxnChargeableError = 2 // Indicates the transaction is successful in updating the state or smart contract
	TxnFail            = 3 // Indicates a transaction has failed to update the state or smart contract
)

// Transaction entity that encapsulates the transaction related data and meta data
type Transaction struct {
	Hash              string `json:"hash,omitempty"`
	Version           string `json:"version,omitempty"`
	ClientID          string `json:"client_id,omitempty"`
	PublicKey         string `json:"public_key,omitempty"`
	ToClientID        string `json:"to_client_id,omitempty"`
	ChainID           string `json:"chain_id,omitempty"`
	TransactionData   string `json:"transaction_data"`
	Value             uint64 `json:"transaction_value"`
	Signature         string `json:"signature,omitempty"`
	CreationDate      int64  `json:"creation_date,omitempty"`
	TransactionType   int    `json:"transaction_type"`
	TransactionOutput string `json:"transaction_output,omitempty"`
	TransactionFee    uint64 `json:"transaction_fee"`
	TransactionNonce  int64  `json:"transaction_nonce"`
	OutputHash        string `json:"txn_output_hash"`
	Status            int    `json:"transaction_status"`
}

// TxnReceipt - a transaction receipt is a processed transaction that contains the output
type TxnReceipt struct {
	Transaction *Transaction
}

type SmartContractTxnData struct {
	Name      string      `json:"name"`
	InputArgs interface{} `json:"input"`
}

type StorageAllocation struct {
	ID             string `json:"id"`
	DataShards     int    `json:"data_shards"`
	ParityShards   int    `json:"parity_shards"`
	Size           int64  `json:"size"`
	Expiration     int64  `json:"expiration_date"`
	Owner          string `json:"owner_id"`
	OwnerPublicKey string `json:"owner_public_key"`
	ReadRatio      *Ratio `json:"read_ratio"`
	WriteRatio     *Ratio `json:"write_ratio"`
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
	FREE_UPDATE_ALLOCATION    = "free_update_allocation"
	LOCK_TOKEN                = "lock"
	UNLOCK_TOKEN              = "unlock"

	ADD_FREE_ALLOCATION_ASSIGNER = "add_free_storage_assigner"

	// Vesting SC
	VESTING_TRIGGER         = "trigger"
	VESTING_STOP            = "stop"
	VESTING_UNLOCK          = "unlock"
	VESTING_ADD             = "add"
	VESTING_DELETE          = "delete"
	VESTING_UPDATE_SETTINGS = "vestingsc-update-settings"

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
	STORAGESC_COLLECT_REWARD            = "collect_reward"
	STORAGESC_KILL_BLOBBER              = "kill_blobber"
	STORAGESC_KILL_VALIDATOR            = "kill_validator"
	STORAGESC_SHUTDOWN_BLOBBER          = "shutdown_blobber"
	STORAGESC_SHUTDOWN_VALIDATOR        = "shutdown_validator"

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
)

type SignFunc = func(msg string) (string, error)
type VerifyFunc = func(signature, msgHash, publicKey string) (bool, error)
type SignWithWallet = func(msg string, wallet interface{}) (string, error)

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

func (t *Transaction) VerifyTransaction(verifyHandler VerifyFunc) (bool, error) {
	// Store the hash
	hash := t.Hash
	t.ComputeHashData()
	if t.Hash != hash {
		return false, errors.New("verify_transaction", fmt.Sprintf(`{"error":"hash_mismatch", "expected":"%v", "actual":%v"}`, t.Hash, hash))
	}
	return verifyHandler(t.Signature, t.Hash, t.PublicKey)
}

func SendTransactionSync(txn *Transaction, miners []string) {
	wg := sync.WaitGroup{}
	wg.Add(len(miners))
	for _, miner := range miners {
		url := fmt.Sprintf("%v/%v", miner, TXN_SUBMIT_URL)
		go sendTransactionToURL(url, txn, &wg)
	}
	wg.Wait()
}

func sendTransactionToURL(url string, txn *Transaction, wg *sync.WaitGroup) ([]byte, error) {
	if wg != nil {
		defer wg.Done()
	}
	postReq, err := util.NewHTTPPostRequest(url, txn)
	if err != nil {
		//Logger.Error("Error in serializing the transaction", txn, err.Error())
		return nil, err
	}
	postResponse, err := postReq.Post()
	if postResponse.StatusCode >= 200 && postResponse.StatusCode <= 299 {
		return []byte(postResponse.Body), nil
	}
	return nil, errors.Wrap(err, errors.New("transaction_send_error", postResponse.Body))
}
