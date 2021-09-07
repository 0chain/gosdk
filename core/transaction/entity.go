package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/util"
)

const TXN_SUBMIT_URL = "v1/transaction/put"
const TXN_VERIFY_URL = "v1/transaction/get/confirmation?hash="

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
	VESTING_TRIGGER         = "trigger"
	VESTING_STOP            = "stop"
	VESTING_UNLOCK          = "unlock"
	VESTING_ADD             = "add"
	VESTING_DELETE          = "delete"
	VESTING_UPDATE_SETTINGS = "vestingsc-update-settings"

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
	STORAGESC_REMOVE_CURATOR           = "remove_curator"
	STORAGESC_CURATOR_TRANSFER         = "curator_transfer_allocation"
	STORAGESC_UPDATE_SETTINGS          = "update_settings"

	MINERSC_LOCK             = "addToDelegatePool"
	MINERSC_UNLOCK           = "deleteFromDelegatePool"
	MINERSC_MINER_SETTINGS   = "update_miner_settings"
	MINERSC_SHARDER_SETTINGS = "update_sharder_settings"
	MINERSC_UPDATE_SETTINGS  = "update_settings"
	MINERSC_UPDATE_GLOBALS   = "update_globals"
	MINERSC_MINER_DELETE     = "delete_miner"
	MINERSC_SHARDER_DELETE   = "delete_sharder"

	// Faucet SC
	FAUCETSC_UPDATE_SETTINGS = "update-settings"

	// Interest pool SC
	INTERESTPOOLSC_UPDATE_SETTINGS = "updateVariables"
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

// VerifyTransaction query transaction status from sharders, and verify it by mininal confirmation
func VerifyTransaction(txnHash string, sharders []string) (*Transaction, error) {
	if cfg == nil {
		return nil, ErrConfigIsNotInitialized
	}

	numSharders := len(sharders)

	if numSharders == 0 {
		return nil, ErrNoAvailableSharder
	}

	minNumConfirmation := int(math.Ceil(float64(cfg.MinConfirmation*numSharders) / 100))

	rand := util.NewRand(numSharders)

	selectedSharders := make([]string, 0, minNumConfirmation+1)

	// random pick minNumConfirmation+1 first
	for i := 0; i <= minNumConfirmation; i++ {
		n, err := rand.Next()

		if err != nil {
			break
		}

		selectedSharders = append(selectedSharders, sharders[n])
	}

	numSuccess := 0

	var retTxn *Transaction

	//leave first item for ErrTooLessConfirmation
	var msgList = make([]string, 1, numSharders)

	urls := make([]string, 0, len(selectedSharders))

	for _, sharder := range selectedSharders {
		urls = append(urls, fmt.Sprintf("%v/%v%v", sharder, TXN_VERIFY_URL, txnHash))
	}

	header := map[string]string{
		"Content-Type":                "application/json; charset=utf-8",
		"Access-Control-Allow-Origin": "*",
	}

	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: resty.DefaultDialTimeout,
		}).Dial,
		TLSHandshakeTimeout: resty.DefaultDialTimeout,
	}
	r := resty.New(transport, func(req *http.Request, resp *http.Response, cf context.CancelFunc, err error) error {
		url := req.URL.String()

		if err != nil { //network issue
			msgList = append(msgList, err.Error())
			return err
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil { //network issue
			msgList = append(msgList, url+": "+err.Error())
			return err
		}

		if resp.StatusCode != 200 {
			msgList = append(msgList, url+": ["+strconv.Itoa(resp.StatusCode)+"] "+string(body))
			return errors.Throw(ErrInvalidRequest, strconv.Itoa(resp.StatusCode)+": "+resp.Status)
		}

		var objmap map[string]json.RawMessage
		err = json.Unmarshal(body, &objmap)
		if err != nil {
			msgList = append(msgList, "json: "+string(body))
			return err
		}
		txnRawJSON, ok := objmap["txn"]

		// txn data is found, success
		if ok {
			txn := &Transaction{}
			err = json.Unmarshal(txnRawJSON, txn)
			if err != nil {
				msgList = append(msgList, "json: "+string(txnRawJSON))
				return err
			}
			if len(txn.Signature) > 0 {
				retTxn = txn
			}
			numSuccess++

		} else {
			// txn data is not found, but get block_hash, success
			if _, ok := objmap["block_hash"]; ok {
				numSuccess++
			} else {
				// txn and block_hash
				msgList = append(msgList, fmt.Sprintf("Sharder does not have the block summary with url: %s, contents: %s", url, string(body)))
			}

		}

		return nil
	},
		resty.WithTimeout(resty.DefaultRequestTimeout),
		resty.WithRetry(resty.DefaultRetry),
		resty.WithHeader(header))

	for {
		r.DoGet(context.TODO(), urls...)

		r.Wait()

		if numSuccess >= minNumConfirmation {
			break
		}

		// pick more one sharder to query transaction
		n, err := rand.Next()

		if errors.Is(err, util.ErrNoItem) {
			break
		}

		urls = []string{fmt.Sprintf("%v/%v%v", sharders[n], TXN_VERIFY_URL, txnHash)}

	}

	if numSuccess > 0 && numSuccess >= minNumConfirmation {

		if retTxn == nil {
			return nil, errors.Throw(ErrNoTxnDetail, strings.Join(msgList, "\r\n"))
		}

		return retTxn, nil
	}

	msgList[0] = fmt.Sprintf("min_confirmation is %v%%, but got %v/%v sharders", cfg.MinConfirmation, numSuccess, numSharders)

	return nil, errors.Throw(ErrTooLessConfirmation, strings.Join(msgList, "\r\n"))

}
