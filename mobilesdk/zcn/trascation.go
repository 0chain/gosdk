//go:build mobile
// +build mobile

package zcn

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/transaction"
)

// EstimateFee estimates transaction fee
// jsonTxn:
//
//	type TransactionPutRequest struct {
//		Hash              string `json:"hash"`
//		Signature         string `json:"signature"`
//		PublicKey         string `json:"public_key,omitempty"`
//		Version           string `json:"version"`
//		ClientId          string `json:"client_id"`
//		ToClientId        string `json:"to_client_id"`
//		TransactionData   string `json:"transaction_data"`
//		TransactionValue  int64  `json:"transaction_value"`
//		CreationDate      int64  `json:"creation_date"`
//		TransactionFee    int64  `json:"transaction_fee"`
//		TransactionType   int    `json:"transaction_type"`
//		TransactionOutput string `json:"transaction_output,omitempty"`
//		TxnOutputHash     string `json:"txn_output_hash"`
//		TransactionNonce  int    `json:"transaction_nonce"`
//	}
func EstimateFee(jsonTxn string, miners []string, reqPercent float32) (int64, error) {
	txn := &transaction.Transaction{}
	err := json.Unmarshal([]byte(jsonTxn), txn)
	if err != nil {
		return 0, err
	}
	fee, err := transaction.EstimateFee(txn, miners, reqPercent)
	if err != nil {
		return 0, err
	}

	return int64(fee), nil
}
