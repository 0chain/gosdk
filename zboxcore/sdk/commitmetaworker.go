package sdk

import (
	"encoding/json"
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	. "github.com/0chain/gosdk/zboxcore/logger"
)

type CommitMetaData struct {
	CrudType string
	MetaData *ConsolidatedFileMeta
}

type CommitMetaRequest struct {
	CommitMetaData
	status StatusCallback
}

type CommitMetaResponse struct {
	TxnID    string
	MetaData *ConsolidatedFileMeta
}

func (req *CommitMetaRequest) processCommitMetaRequest() {
	commitMetaDataBytes, err := json.Marshal(req.CommitMetaData)
	if err != nil {
		req.status.CommitMetaCompleted("", "", err)
		return
	}
	commitMetaDataString := string(commitMetaDataBytes)

	txn := transaction.NewTransactionEntity(client.GetClientID(), blockchain.GetChainID(), client.GetClientPublicKey())
	txn.TransactionData = commitMetaDataString
	txn.TransactionType = transaction.TxnTypeData
	err = txn.ComputeHashAndSign(client.Sign)
	if err != nil {
		req.status.CommitMetaCompleted(commitMetaDataString, "", err)
		return
	}

	transaction.SendTransactionSync(txn, blockchain.GetMiners())
	querySleepTime := time.Duration(blockchain.GetQuerySleepTime()) * time.Second
	time.Sleep(querySleepTime)
	retries := 0
	var t *transaction.Transaction
	for retries < blockchain.GetMaxTxnQuery() {
		t, err = transaction.VerifyTransaction(txn.Hash, blockchain.GetSharders())
		if err == nil {
			break
		}
		retries++
		time.Sleep(querySleepTime)
	}

	if err != nil {
		Logger.Error("Error verifying the commit transaction", err.Error(), txn.Hash)
		req.status.CommitMetaCompleted(commitMetaDataString, "", err)
		return
	}
	if t == nil {
		err = common.NewError("transaction_validation_failed", "Failed to get the transaction confirmation")
		req.status.CommitMetaCompleted(commitMetaDataString, "", err)
		return
	}

	commitMetaResponse := &CommitMetaResponse{
		TxnID:    t.Hash,
		MetaData: req.CommitMetaData.MetaData,
	}
	commitMetaReponseBytes, err := json.Marshal(commitMetaResponse)
	if err != nil {
		req.status.CommitMetaCompleted(commitMetaDataString, "", err)
	}
	commitMetaResponseString := string(commitMetaReponseBytes)
	req.status.CommitMetaCompleted(commitMetaDataString, commitMetaResponseString, nil)
	return
}
