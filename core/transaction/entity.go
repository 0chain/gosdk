package transaction

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/util"
)

const TXN_SUBMIT_URL = "v1/transaction/put"
const TXN_VERIFY_URL = "v1/transaction/get/confirmation?hash="

var ErrNoTxnDetail = common.NewError("missing_transaction_detail", "No transaction detail was found on any of the sharders")

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
const NEW_ALLOCATION_REQUEST = "new_allocation_request"
const LOCK_TOKEN = "lock"
const UNLOCK_TOKEN = "unlock"
const STAKE = "addToDelegatePool"
const DELETE_STAKE = "deleteFromDelegatePool"

type SignFunc = func(msg string) (string, error)
type VerifyFunc = func(signature, msgHash, publicKey string) (bool, error)

func NewTransactionEntity(clientID string, chainID string, publicKey string) *Transaction {
	txn := &Transaction{}
	txn.Version = "1.0"
	txn.ClientID = clientID
	txn.CreationDate = common.Now()
	txn.ChainID = chainID
	txn.PublicKey = publicKey
	return txn
}

func (t *Transaction) ComputeHash() {
	hashdata := fmt.Sprintf("%v:%v:%v:%v:%v", t.CreationDate, t.ClientID,
		t.ToClientID, t.Value, encryption.Hash(t.TransactionData))
	t.Hash = encryption.Hash(hashdata)
}
func (t *Transaction) ComputeHashAndSign(signHandler SignFunc) error {
	t.ComputeHash()
	var err error
	t.Signature, err = signHandler(t.Hash)
	if err != nil {
		return err
	}
	return nil
}
func (t *Transaction) VerifyTransaction(verifyHandler VerifyFunc) (bool, error) {
	// Store the hash
	hash := t.Hash
	t.ComputeHash()
	if t.Hash != hash {
		return false, fmt.Errorf(`{"error":"hash_mismatch", "expected":"%v", "actual":%v"}`, t.Hash, hash)
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
	return nil, common.NewError("transaction_send_error", postResponse.Body)
}

func VerifyTransaction(txnHash string, sharders []string) (*Transaction, error) {
	numSharders := len(sharders)
	numSuccess := 0
	var retTxn *Transaction
	for _, sharder := range sharders {
		url := fmt.Sprintf("%v/%v%v", sharder, TXN_VERIFY_URL, txnHash)
		req, err := util.NewHTTPGetRequest(url)
		response, err := req.Get()
		if err != nil {
			//Logger.Error("Error getting transaction confirmation", err.Error())
			numSharders--
		} else {
			if response.StatusCode != 200 {
				continue
			}
			contents := response.Body
			var objmap map[string]json.RawMessage
			err = json.Unmarshal([]byte(contents), &objmap)
			if err != nil {
				//Logger.Error("Error unmarshalling response", err.Error())
				continue
			}
			if _, ok := objmap["txn"]; !ok {
				//Logger.Info("Not transaction information. Only block summary.", url, contents)
				if _, ok := objmap["block_hash"]; ok {
					numSuccess++
					continue
				}
				//Logger.Info("Sharder does not have the block summary", url, contents)
				continue
			}
			txn := &Transaction{}
			err = json.Unmarshal(objmap["txn"], txn)
			if err != nil {
				//Logger.Error("Error unmarshalling to get transaction response", err.Error())
			}
			if len(txn.Signature) > 0 {
				retTxn = txn
			}

			numSuccess++
		}
	}
	if numSharders == 0 || float64(numSuccess*1.0/numSharders) > float64(0.5) {
		if retTxn != nil {
			return retTxn, nil
		}
		return nil, ErrNoTxnDetail
	}
	return nil, common.NewError("transaction_not_found", "Transaction was not found on any of the sharders")
}
