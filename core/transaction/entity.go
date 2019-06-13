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
	TransactionData   string `json:"transaction_data,omitempty"`
	Value             int64  `json:"transaction_value,omitempty"`
	Signature         string `json:"signature,omitempty"`
	CreationDate      int64  `json:"creation_date,omitempty"`
	TransactionType   int    `json:"transaction_type,omitempty"`
	TransactionOutput string `json:"transaction_output,omitempty"`
	OutputHash        string `json:"txn_output_hash"`
}

type SmartContractTxnData struct {
	Name      string      `json:"name"`
	InputArgs interface{} `json:"input"`
}

const NEW_ALLOCATION_REQUEST = "new_allocation_request"
const LOCK_TOKEN = "lock"
const UNLOCK_TOKEN = "unlock"

type SignFunc = func(msg string) (string, error)

func NewTransactionEntity(clientID string, chainID string, publicKey string) *Transaction {
	txn := &Transaction{}
	txn.Version = "1.0"
	txn.ClientID = clientID
	txn.CreationDate = common.Now()
	txn.ChainID = chainID
	txn.PublicKey = publicKey
	return txn
}

func (t *Transaction) ComputeHashAndSign(signHandler SignFunc) error {
	hashdata := fmt.Sprintf("%v:%v:%v:%v:%v", t.CreationDate, t.ClientID,
		t.ToClientID, t.Value, encryption.Hash(t.TransactionData))
	t.Hash = encryption.Hash(hashdata)
	var err error
	t.Signature, err = signHandler(t.Hash)
	if err != nil {
		return err
	}
	return nil
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
