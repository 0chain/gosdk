package main

import "github.com/0chain/gosdk/core/transaction"

var commitMetaTxnChan = make(chan MetadataCommitResult, 10)

type MetadataCommitResult struct {
	Error error
	Txn   *transaction.Transaction
}

func setLastMetadataCommitTxn(txn *transaction.Transaction, err error) {
	go func() {
		commitMetaTxnChan <- MetadataCommitResult{
			Error: err,
			Txn:   txn,
		}
	}()
}

func getLastMetadataCommitTxn() (*transaction.Transaction, error) {
	r := <-commitMetaTxnChan

	return r.Txn, r.Error
}
