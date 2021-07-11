package sdk

import (
	"encoding/base64"
	"encoding/json"
	"sync"
	"time"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc"
	"github.com/0chain/gosdk/core/clients/blobberClient"
	"github.com/0chain/gosdk/core/common/errors"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/commitmeta"
	. "github.com/0chain/gosdk/zboxcore/logger"
)

type CommitMetaRequest struct {
	commitmeta.CommitMetaData
	status    StatusCallback
	a         *Allocation
	authToken string
	wg        *sync.WaitGroup
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
		err = errors.New("transaction_validation_failed", "Failed to get the transaction confirmation")
		req.status.CommitMetaCompleted(commitMetaDataString, "", err)
		return
	}

	if ok := req.updateCommitMetaTxnToBlobbers(t.Hash); ok {
		Logger.Info("Updated commitMetaTxnID to all blobbers")
	} else {
		Logger.Info("Failed to update commitMetaTxnID to all blobbers")
	}

	commitMetaResponse := &commitmeta.CommitMetaResponse{
		TxnID:    t.Hash,
		MetaData: req.CommitMetaData.MetaData,
	}

	Logger.Info("Marshaling commitMetaResponse to bytes")
	commitMetaReponseBytes, err := json.Marshal(commitMetaResponse)
	if err != nil {
		Logger.Error("Failed to marshal commitMetaResponse to bytes")
		req.status.CommitMetaCompleted(commitMetaDataString, "", err)
	}

	Logger.Info("Converting commitMetaResponse bytes to string")
	commitMetaResponseString := string(commitMetaReponseBytes)

	Logger.Info("Commit complete, Calling CommitMetaCompleted callback")
	req.status.CommitMetaCompleted(commitMetaDataString, commitMetaResponseString, nil)

	Logger.Info("All process done, Calling return")
	return
}

func (req *CommitMetaRequest) updateCommitMetaTxnToBlobbers(txnHash string) bool {
	numList := len(req.a.Blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	rspCh := make(chan bool, numList)
	for i := 0; i < numList; i++ {
		go req.updatCommitMetaTxnToBlobber(req.a.Blobbers[i], i, txnHash, rspCh)
	}
	req.wg.Wait()
	count := 0
	for i := 0; i < numList; i++ {
		resp := <-rspCh
		if resp {
			count++
		}
	}
	return count == numList
}

func (req *CommitMetaRequest) updatCommitMetaTxnToBlobber(blobber *blockchain.StorageNode, blobberIdx int, txnHash string, rspCh chan<- bool) {
	defer req.wg.Done()

	var authToken string
	if len(req.authToken) > 0 {
		sEnc, err := base64.StdEncoding.DecodeString(req.authToken)
		if err != nil {
			Logger.Error("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
			return
		}

		authToken = string(sEnc)
	}

	_, err := blobberClient.CommitMetaTxn(blobber.Baseurl, &blobbergrpc.CommitMetaTxnRequest{
		Path:       req.MetaData.Path,
		PathHash:   req.MetaData.LookupHash,
		AuthToken:  authToken,
		Allocation: req.a.Tx,
		TxnId:      txnHash,
	})
	if err != nil {
		Logger.Error("Update CommitMetaTxn: ", err)
		rspCh <- false
		return
	}

	rspCh <- true
	return
}
