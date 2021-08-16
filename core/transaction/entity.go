package transaction

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/common/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/util"
)

const TXN_SUBMIT_URL = "v1/transaction/put"
const TXN_VERIFY_URL = "v1/transaction/get/confirmation?hash="

var ErrNoTxnDetail = errors.New("missing_transaction_detail", "No transaction detail was found on any of the sharders")

//Transaction entity that encapsulates the transaction related data and meta data
type Transaction struct {
	Hash              string `json:"hash,omitempty"`
	Version           string `json:"version,omitempty"`
	ClientID          string `json:"client_id,omitempty"`
	PublicKey         string `json:"public_key,omitempty"`
	ToClientID        string `json:"to_client_id,omitempty"`
	ChainID           string `json:"chain_id,omitempty"`
	TransactionData   string `json:"transaction_data"`
	Value             int64  `json:"transaction_value"`
	Signature         string `json:"signature,omitempty"`
	CreationDate      int64  `json:"creation_date,omitempty"`
	TransactionType   int    `json:"transaction_type"`
	TransactionOutput string `json:"transaction_output,omitempty"`
	TransactionFee    int64  `json:"transaction_fee"`
	OutputHash        string `json:"txn_output_hash"`
}

//TxnReceipt - a transaction receipt is a processed transaction that contains the output
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
	CreationData          int64  `json:"creation_date"`
	Hash                  string `json:"hash"`
	MinerID               string `json:"miner_id"`
	Round                 int64  `json:"round"`
	RoundRandomSeed       int64  `json:"round_random_seed"`
	MerkleTreeRoot        string `json:"merkle_tree_root"`
	StateHash             string `json:"state_hash"`
	ReceiptMerkleTreeRoot string `json:"receipt_merkle_tree_root"`
	NumberOfTxns          int64  `json:"num_txns"`
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
	VESTING_TRIGGER       = "trigger"
	VESTING_STOP          = "stop"
	VESTING_UNLOCK        = "unlock"
	VESTING_ADD           = "add"
	VESTING_DELETE        = "delete"
	VESTING_UPDATE_CONFIG = "update_config"

	// Storage SC
	STORAGESC_FINALIZE_ALLOCATION      = "finalize_allocation"
	STORAGESC_CANCEL_ALLOCATION        = "cancel_allocation"
	STORAGESC_CREATE_ALLOCATION        = "new_allocation_request"
	STORAGESC_CREATE_READ_POOL         = "new_read_pool"
	STORAGESC_READ_POOL_LOCK           = "read_pool_lock"
	STORAGESC_READ_POOL_UNLOCK         = "read_pool_unlock"
	STORAGESC_STAKE_POOL_LOCK          = "stake_pool_lock"
	STORAGESC_STAKE_POOL_UNLOCK        = "stake_pool_unlock"
	STORAGESC_STAKE_POOL_PAY_INTERESTS = "stake_pool_pay_interests"
	STORAGESC_UPDATE_BLOBBER_SETTINGS  = "update_blobber_settings"
	STORAGESC_UPDATE_ALLOCATION        = "update_allocation_request"
	STORAGESC_WRITE_POOL_LOCK          = "write_pool_lock"
	STORAGESC_WRITE_POOL_UNLOCK        = "write_pool_unlock"
	STORAGESC_ADD_CURATOR              = "add_curator"
	STORAGESC_CURATOR_TRANSFER         = "curator_transfer_allocation"

	// Miner SC
	MINERSC_LOCK             = "addToDelegatePool"
	MINERSC_UNLOCK           = "deleteFromDelegatePool"
	MINERSC_MINER_SETTINGS   = "update_miner_settings"
	MINERSC_SHARDER_SETTINGS = "update_sharder_settings"

	// Faucet SC
	FAUCETSC_UPDATE_SETTINGS = "faucetsc-update-settings"
)

type SignFunc = func(msg string) (string, error)
type VerifyFunc = func(signature, msgHash, publicKey string) (bool, error)
type SignWithWallet = func(msg string, wallet interface{}) (string, error)

func NewTransactionEntity(clientID string, chainID string, publicKey string) *Transaction {
	txn := &Transaction{}
	txn.Version = "1.0"
	txn.ClientID = clientID
	txn.CreationDate = int64(common.Now())
	txn.ChainID = chainID
	txn.PublicKey = publicKey
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
	hashdata := fmt.Sprintf("%v:%v:%v:%v:%v", t.CreationDate, t.ClientID,
		t.ToClientID, t.Value, encryption.Hash(t.TransactionData))
	t.Hash = encryption.Hash(hashdata)
}

//GetHash - implement interface
func (rh *TxnReceipt) GetHash() string {
	return rh.Transaction.OutputHash
}

/*GetHashBytes - implement Hashable interface */
func (rh *TxnReceipt) GetHashBytes() []byte {
	return util.HashStringToBytes(rh.Transaction.OutputHash)
}

//NewTransactionReceipt - create a new transaction receipt
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

func VerifyTransaction(txnHash string, sharders []string) (*Transaction, error) {
	numSharders := len(sharders)
	numSuccess := 0
	var retTxn *Transaction
	var customError error
	for _, sharder := range sharders {
		url := fmt.Sprintf("%v/%v%v", sharder, TXN_VERIFY_URL, txnHash)
		req, err := util.NewHTTPGetRequest(url)
		if err != nil {
			customError = errors.Wrap(customError, err)
			numSharders--
			continue
		}
		response, err := req.Get()
		if err != nil {
			customError = errors.Wrap(customError, err)
			numSharders--
			continue
		} else {
			if response.StatusCode != 200 {
				customError = errors.Wrap(customError, err)
				continue
			}
			contents := response.Body
			var objmap map[string]json.RawMessage
			err = json.Unmarshal([]byte(contents), &objmap)
			if err != nil {
				customError = errors.Wrap(customError, err)
				continue
			}
			if _, ok := objmap["txn"]; !ok {
				if _, ok := objmap["block_hash"]; ok {
					numSuccess++
				} else {
					customError = errors.Wrap(customError, fmt.Sprintf("Sharder does not have the block summary with url: %s, contents: %s", url, contents))
				}
				continue
			}
			txn := &Transaction{}
			err = json.Unmarshal(objmap["txn"], txn)
			if err != nil {
				customError = errors.Wrap(customError, err)
				continue
			}
			if len(txn.Signature) > 0 {
				retTxn = txn
			}
			numSuccess++
		}
	}
	if numSharders == 0 || float64(numSuccess*1.0/numSharders) > 0.5 {
		if retTxn != nil {
			return retTxn, nil
		}
		return nil, errors.Wrap(customError, ErrNoTxnDetail)
	}
	return nil, errors.Wrap(customError, errors.New("transaction_not_found", "Transaction was not found on any of the sharders"))
}
