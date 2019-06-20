package zcncore

import (
	"encoding/json"
	"fmt"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	"net/http"
	"time"
)

// TransactionCallback needs to be implemented by the caller for transaction related APIs
type TransactionCallback interface {
	OnTransactionComplete(t *Transaction, status int)
	OnVerifyComplete(t *Transaction, status int)
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
	// SendWithSignature implements sending token.
	// signature will be passed by application where multi signature is involved.
	SendWithSignature(toClientID string, val int64, desc string, sig string) error
	// StoreData implements store the data to blockchain
	StoreData(data string) error
	// ExecuteFaucetSC impements the Faucet Smart contract
	ExecuteFaucetSC(methodName string, input []byte) error
	// GetTransactionHash implements retrieval of hash of the submitted transaction
	GetTransactionHash() string
	// LockTokens implements the lock token.
	LockTokens(val int64, durationHr int64, durationMin int) error
	// UnlockTokens implements unlocking of earlier locked tokens.
	UnlockTokens(poolID string) error
	// SetTransactionHash implements verify a previous transation status
	SetTransactionHash(hash string) error
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
	for _, miner := range _config.chain.Miners {
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
	for range _config.chain.Miners {
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
	rate := consensus * 100 / float32(len(_config.chain.Miners))
	if rate < consensusThresh {
		t.completeTxn(StatusError, "", fmt.Errorf("submit transaction failed. %s", tFailureRsp))
		return
	}
	time.Sleep(3 * time.Second)
	t.completeTxn(StatusSuccess, tSuccessRsp, nil)
}

// NewTransaction allocation new generic transaction object for any operation
func NewTransaction(cb TransactionCallback) (*Transaction, error) {
	if _config.wallet.ClientID == "" {
		return nil, fmt.Errorf("wallet info not found. set wallet info.")
	}
	t := &Transaction{}
	t.txn = transaction.NewTransactionEntity(_config.wallet.ClientID, _config.chain.ChainID, _config.wallet.ClientKey)
	t.txnStatus, t.verifyStatus = StatusUnknown, StatusUnknown
	t.txnCb = cb
	return t, nil
}

func (t *Transaction) SetTransactionCallback(cb TransactionCallback) error {
	if t.txnStatus != StatusUnknown {
		return fmt.Errorf("transaction already exists. cannot set transaction hash.")
	}
	t.txnCb = cb
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

func (t *Transaction) SendWithSignature(toClientID string, val int64, desc string, sig string) error {
	t.txn.Signature = sig
	t.Send(toClientID, val, desc)
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

func (t *Transaction) ExecuteFaucetSC(methodName string, input []byte) error {
	sn := transaction.SmartContractTxnData{Name: methodName, InputArgs: input}
	snBytes, err := json.Marshal(sn)
	if err != nil {
		return fmt.Errorf("execute faucet failed due to invalid data. %s", err.Error())
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSmartContract
		t.txn.ToClientID = FaucetSmartContractAddress
		t.txn.TransactionData = string(snBytes)
		t.txn.Value = 0
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
	go func() {
		result := make(chan *util.GetResponse)
		var tSuccessRsp string
		var tFailureRsp string
		defer close(result)
		for _, sharder := range _config.chain.Sharders {
			go func(sharderurl string) {
				Logger.Info("Verify transaction hash: ", t.txnHash, " from", sharderurl)
				url := fmt.Sprintf("%v%v%v", sharderurl, TXN_VERIFY_URL, t.txnHash)
				ticker := time.NewTicker(time.Second)
				defer ticker.Stop()
				var res *util.GetResponse
				done := make(chan bool)
				go func() {
					lburl := fmt.Sprintf("%v%v", sharderurl, LATEST_FINALIZED_BLOCK)
					lbticker := time.NewTicker(time.Second)
					defer lbticker.Stop()
					for true {
						select {
						case <-lbticker.C:
							req, err := util.NewHTTPGetRequest(lburl)
							if err != nil {
								Logger.Error(sharderurl, "new get request failed. ", err.Error())
							}
							res, err := req.Get()
							if err != nil {
								Logger.Error(sharderurl, "get error. ", err.Error())
							}
							var objmap map[string]json.RawMessage
							err = json.Unmarshal([]byte(res.Body), &objmap)
							if err != nil {
								Logger.Debug("error getting latest finalized block: ", err)
								break
							}
							if date, ok := objmap["creation_date"]; ok {
								var dateTimeStamp int64
								err = json.Unmarshal(date, &dateTimeStamp)
								if dateTimeStamp-t.txn.CreationDate > 10 {
									done <- true
									return
								}
							}
						}
					}
				}()
				accepted := false
				for !accepted {
					select {
					case <-done:
						accepted = true
					case <-ticker.C:
						req, err := util.NewHTTPGetRequest(url)
						if err != nil {
							Logger.Error(sharderurl, "new get request failed. ", err.Error())
						}
						res, err = req.Get()
						if err != nil {
							Logger.Error(sharderurl, "get error. ", err.Error())
						}
						var objmap map[string]json.RawMessage
						err = json.Unmarshal([]byte(res.Body), &objmap)
						if err != nil {
							continue
						}
						_, accepted = objmap["version"]
					}
				}
				result <- res
				return
			}(sharder)
		}
		consensus := float32(0)
		for range _config.chain.Sharders {
			select {
			case rsp := <-result:
				Logger.Debug(rsp.Url, rsp.Status)
				if rsp.StatusCode == http.StatusOK {
					var objmap map[string]json.RawMessage
					err := json.Unmarshal([]byte(rsp.Body), &objmap)
					if err != nil {
						continue
					}
					if _, ok := objmap["txn"]; !ok {
						Logger.Debug("no transaction information. only block summary.", rsp.Url, rsp.Body)
						if _, ok := objmap["block_hash"]; ok {
							consensus++
							continue
						}
						Logger.Debug("sharder does not have the block summary", rsp.Url, rsp.Body)
						continue
					}
					tSuccessRsp = rsp.Body
					consensus++
				} else {
					Logger.Error(rsp.Body)
				}
			}
		}
		rate := consensus * 100 / float32(len(_config.chain.Sharders))
		if rate < consensusThresh {
			t.completeVerify(StatusError, "", fmt.Errorf("verify transaction failed. %s", tFailureRsp))
		} else {
			t.completeVerify(StatusSuccess, tSuccessRsp, nil)
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

func (t *Transaction) LockTokens(val int64, durationHr int64, durationMin int) error {
	lockInput := make(map[string]interface{})
	lockInput["duration"] = fmt.Sprintf("%dh%dm", durationHr, durationMin)
	sn := transaction.SmartContractTxnData{Name: transaction.LOCK_TOKEN, InputArgs: lockInput}
	snBytes, err := json.Marshal(sn)
	if err != nil {
		return fmt.Errorf("lock token failed due to invalid data. %s", err.Error())
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSmartContract
		t.txn.ToClientID = InterestPoolSmartContractAddress
		t.txn.TransactionData = string(snBytes)
		t.txn.Value = val
		t.submitTxn()
	}()
	return nil
}

func (t *Transaction) UnlockTokens(poolID string) error {
	unlockInput := make(map[string]interface{})
	unlockInput["pool_id"] = poolID
	sn := transaction.SmartContractTxnData{Name: transaction.UNLOCK_TOKEN, InputArgs: unlockInput}
	snBytes, err := json.Marshal(sn)
	if err != nil {
		return fmt.Errorf("unlock token failed due to invalid data. %s", err.Error())
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSmartContract
		t.txn.ToClientID = InterestPoolSmartContractAddress
		t.txn.TransactionData = string(snBytes)
		t.txn.Value = 0
		t.submitTxn()
	}()
	return nil
}
