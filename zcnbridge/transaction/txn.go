package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zcnbridge/errors"
	ctime "github.com/0chain/gosdk/zcnbridge/time"
	"github.com/0chain/gosdk/zcncore"
)

var (
	_ zcncore.TransactionCallback = (*callback)(nil)
)

type (
	// TransactionProvider ...
	TransactionProvider interface {
		NewTransactionEntity(txnFee uint64) (Transaction, error)
	}

	// transactionProvider ...
	transactionProvider struct{}

	// Transaction interface describes transaction entity.
	Transaction interface {
		ExecuteSmartContract(ctx context.Context, address, funcName string, input interface{}, val uint64) (string, error)
		Verify(ctx context.Context) error
		GetScheme() zcncore.TransactionScheme
		GetCallback() TransactionCallbackAwaitable
		GetTransactionOutput() string
		GetHash() string
		SetHash(string)
	}

	// TransactionEntity entity that encapsulates the transaction related data and metadata.
	transactionEntity struct {
		Hash              string `json:"hash,omitempty"`
		Version           string `json:"version,omitempty"`
		TransactionOutput string `json:"transaction_output,omitempty"`
		scheme            zcncore.TransactionScheme
		callBack          TransactionCallbackAwaitable
	}
)

type (
	verifyOutput struct {
		Confirmation confirmation `json:"confirmation"`
	}

	// confirmation represents the acceptance that a transaction is included into the blockchain.
	confirmation struct {
		Version               string          `json:"version"`
		Hash                  string          `json:"hash"`
		BlockHash             string          `json:"block_hash"`
		PreviousBlockHash     string          `json:"previous_block_hash"`
		Transaction           Transaction     `json:"txn,omitempty"`
		CreationDate          ctime.Timestamp `json:"creation_date"`
		MinerID               string          `json:"miner_id"`
		Round                 int64           `json:"round"`
		Status                int             `json:"transaction_status"`
		RoundRandomSeed       int64           `json:"round_random_seed"`
		MerkleTreeRoot        string          `json:"merkle_tree_root"`
		MerkleTreePath        *util.MTPath    `json:"merkle_tree_path"`
		ReceiptMerkleTreeRoot string          `json:"receipt_merkle_tree_root"`
		ReceiptMerkleTreePath *util.MTPath    `json:"receipt_merkle_tree_path"`
	}
)

func NewTransactionProvider() TransactionProvider {
	return &transactionProvider{}
}

func (t *transactionProvider) NewTransactionEntity(txnFee uint64) (Transaction, error) {
	return NewTransactionEntity(txnFee)
}

// NewTransactionEntity creates Transaction with initialized fields.
// Sets version, client ID, creation date, public key and creates internal zcncore.TransactionScheme.
func NewTransactionEntity(txnFee uint64) (Transaction, error) {
	txn := &transactionEntity{
		callBack: NewStatus().(*callback),
	}
	zcntxn, err := zcncore.NewTransaction(txn.callBack, txnFee, 0)
	if err != nil {
		return nil, err
	}

	txn.scheme = zcntxn

	return txn, nil
}

// ExecuteSmartContract executes function of smart contract with provided address.
//
// Returns hash of executed transaction.
func (t *transactionEntity) ExecuteSmartContract(ctx context.Context, address, funcName string, input interface{},
	val uint64) (string, error) {
	const errCode = "transaction_send"

	tran, err := t.scheme.ExecuteSmartContract(address, funcName, input, val)
	t.Hash = tran.Hash

	if err != nil {
		msg := fmt.Sprintf("error while sending txn: %v", err)
		return "", errors.New(errCode, msg)
	}

	if err := t.callBack.WaitCompleteCall(ctx); err != nil {
		msg := fmt.Sprintf("error while sending txn: %v", err)
		return "", errors.New(errCode, msg)
	}

	if len(t.scheme.GetTransactionError()) > 0 {
		return "", errors.New(errCode, t.scheme.GetTransactionError())
	}

	return t.scheme.Hash(), nil
}

func (t *transactionEntity) Verify(ctx context.Context) error {
	const errCode = "transaction_verify"

	err := t.scheme.Verify()
	if err != nil {
		msg := fmt.Sprintf("error while verifying txn: %v; txn hash: %s", err, t.scheme.GetTransactionHash())
		return errors.New(errCode, msg)
	}

	if err := t.callBack.WaitVerifyCall(ctx); err != nil {
		msg := fmt.Sprintf("error while verifying txn: %v; txn hash: %s", err, t.scheme.GetTransactionHash())
		return errors.New(errCode, msg)
	}

	switch t.scheme.GetVerifyConfirmationStatus() {
	case zcncore.ChargeableError:
		return errors.New(errCode, strings.Trim(t.scheme.GetVerifyOutput(), "\""))
	case zcncore.Success:
		fmt.Println("Executed smart contract successfully with txn: ", t.scheme.GetTransactionHash())
	default:
		msg := fmt.Sprint("\nExecute smart contract failed. Unknown status code: " +
			strconv.Itoa(int(t.scheme.GetVerifyConfirmationStatus())))
		return errors.New(errCode, msg)
	}

	vo := new(verifyOutput)
	if err := json.Unmarshal([]byte(t.scheme.GetVerifyOutput()), vo); err != nil {
		return errors.New(errCode, "error while unmarshalling confirmation: "+err.Error()+", json: "+t.scheme.GetVerifyOutput())
	}

	if vo.Confirmation.Transaction != nil {
		t.Hash = vo.Confirmation.Transaction.GetHash()
		t.TransactionOutput = vo.Confirmation.Transaction.GetTransactionOutput()
	} else {
		return errors.New(errCode, "got invalid confirmation (missing transaction)")
	}

	return nil
}

// GetSheme returns transaction scheme
func (t *transactionEntity) GetScheme() zcncore.TransactionScheme {
	return t.scheme
}

// GetHash returns transaction hash
func (t *transactionEntity) GetHash() string {
	return t.Hash
}

// SetHash sets transaction hash
func (t *transactionEntity) SetHash(hash string) {
	t.Hash = hash
}

// GetTransactionOutput returns transaction output
func (t *transactionEntity) GetTransactionOutput() string {
	return t.TransactionOutput
}

func (t *transactionEntity) GetCallback() TransactionCallbackAwaitable {
	return t.callBack
}

// GetVersion returns transaction version
func (t *transactionEntity) GetVersion() string {
	return t.Version
}

// Verify checks including of transaction in the blockchain.
func Verify(ctx context.Context, hash string) (Transaction, error) {
	t, err := NewTransactionEntity(0)
	if err != nil {
		return nil, err
	}

	scheme := t.GetScheme()

	if err := scheme.SetTransactionHash(hash); err != nil {
		return nil, err
	}

	err = t.Verify(ctx)

	return t, err
}
