package transaction

import (
	"context"
)

// VerifyTransaction verifies including in blockchain transaction with provided hash.
//
// If execution completed with no error, returns Transaction with provided hash.
func VerifyTransaction(ctx context.Context, txnHash string, clientID, publicKey string) (*Transaction, error) {
	txn, err := NewTransactionEntity(clientID, publicKey)
	if err != nil {
		return nil, err
	}

	txn.Hash = txnHash
	err = txn.Verify(ctx)
	if err != nil {
		return nil, err
	}
	return txn, nil
}
