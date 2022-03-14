package transaction

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zcnbridge/errors"
	ctime "github.com/0chain/gosdk/zcnbridge/time"
	"github.com/0chain/gosdk/zcncore"
)

type (
	// Transaction entity that encapsulates the transaction related data and metadata.
	Transaction struct {
		Hash              string `json:"hash,omitempty"`
		Version           string `json:"version,omitempty"`
		TransactionOutput string `json:"transaction_output,omitempty"`
		scheme            zcncore.TransactionScheme
		callBack          *callback
	}
)

// NewTransactionEntity creates Transaction with initialized fields.
// Sets version, client ID, creation date, public key and creates internal zcncore.TransactionScheme.
func NewTransactionEntity() (*Transaction, error) {
	txn := &Transaction{
		callBack: newCallBack(),
	}
	zcntxn, err := zcncore.NewTransaction(txn.callBack, 0, 0)
	if err != nil {
		return nil, err
	}
	txn.scheme = zcntxn

	return txn, nil
}

// ExecuteSmartContract executes function of smart contract with provided address.
//
// Returns hash of executed transaction.
func (t *Transaction) ExecuteSmartContract(ctx context.Context, address, funcName, input string, val int64) (string, error) {
	const errCode = "transaction_send"

	err := t.scheme.ExecuteSmartContract(address, funcName, input, val)
	if err != nil {
		msg := fmt.Sprintf("error while sending txn: %v", err)
		return "", errors.New(errCode, msg)
	}

	if err := t.callBack.waitCompleteCall(ctx); err != nil {
		msg := fmt.Sprintf("error while sending txn: %v", err)
		return "", errors.New(errCode, msg)
	}

	t.Hash = t.scheme.GetTransactionHash()
	if len(t.scheme.GetTransactionError()) > 0 {
		return "", errors.New(errCode, t.scheme.GetTransactionError())
	}

	return t.Hash, nil
}

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
		Transaction           *Transaction    `json:"txn,omitempty"`
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

// Verify checks including of transaction in the blockchain.
func (t *Transaction) Verify(ctx context.Context) error {
	const errCode = "transaction_verify"

	if err := t.scheme.SetTransactionHash(t.Hash); err != nil {
		return err
	}

	err := t.scheme.Verify()
	if err != nil {
		msg := fmt.Sprintf("error while verifying txn: %v; txn hash: %s", err, t.scheme.GetTransactionHash())
		return errors.New(errCode, msg)
	}

	if err := t.callBack.waitVerifyCall(ctx); err != nil {
		msg := fmt.Sprintf("error while verifying txn: %v; txn hash: %s", err, t.scheme.GetTransactionHash())
		return errors.New(errCode, msg)
	}

	vo := new(verifyOutput)
	if err := json.Unmarshal([]byte(t.scheme.GetVerifyOutput()), vo); err != nil {
		return errors.New(errCode, "error while unmarshalling confirmation: "+err.Error()+", json: "+t.scheme.GetVerifyOutput())
	}

	if vo.Confirmation.Transaction != nil {
		t.TransactionOutput = vo.Confirmation.Transaction.TransactionOutput
	} else {
		return errors.New(errCode, "got invalid confirmation (missing transaction)")
	}

	return nil
}
