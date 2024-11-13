//go:build mobile
// +build mobile

package transaction

import (
	"encoding/json"
	"strconv"
)

// Transaction represents entity that encapsulates the transaction related data and metadata.
type Transaction struct {
	Hash              string `json:"hash,omitempty"`
	Version           string `json:"version,omitempty"`
	ClientID          string `json:"client_id,omitempty"`
	PublicKey         string `json:"public_key,omitempty"`
	ToClientID        string `json:"to_client_id,omitempty"`
	ChainID           string `json:"chain_id,omitempty"`
	TransactionData   string `json:"transaction_data"`
	Value             string `json:"transaction_value"`
	Signature         string `json:"signature,omitempty"`
	CreationDate      int64  `json:"creation_date,omitempty"`
	TransactionType   int    `json:"transaction_type"`
	TransactionOutput string `json:"transaction_output,omitempty"`
	TransactionFee    string `json:"transaction_fee"`
	TransactionNonce  int64  `json:"transaction_nonce"`
	OutputHash        string `json:"txn_output_hash"`
	Status            int    `json:"transaction_status"`
}

// TransactionWrapper represents wrapper for mobile transaction entity.
type TransactionWrapper struct {
	Hash              string `json:"hash,omitempty"`
	Version           string `json:"version,omitempty"`
	ClientID          string `json:"client_id,omitempty"`
	PublicKey         string `json:"public_key,omitempty"`
	ToClientID        string `json:"to_client_id,omitempty"`
	ChainID           string `json:"chain_id,omitempty"`
	TransactionData   string `json:"transaction_data"`
	Value             uint64 `json:"transaction_value"`
	Signature         string `json:"signature,omitempty"`
	CreationDate      int64  `json:"creation_date,omitempty"`
	TransactionType   int    `json:"transaction_type"`
	TransactionOutput string `json:"transaction_output,omitempty"`
	TransactionFee    uint64 `json:"transaction_fee"`
	TransactionNonce  int64  `json:"transaction_nonce"`
	OutputHash        string `json:"txn_output_hash"`
	Status            int    `json:"transaction_status"`
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	valueRaw, err := strconv.ParseUint(t.Value, 0, 64)
	if err != nil {
		return nil, err
	}

	transactionFeeRaw, err := strconv.ParseUint(t.TransactionFee, 0, 64)
	if err != nil {
		return nil, err
	}

	wrapper := TransactionWrapper{
		Hash:              t.Hash,
		Version:           t.Version,
		ClientID:          t.ClientID,
		PublicKey:         t.PublicKey,
		ToClientID:        t.ToClientID,
		ChainID:           t.ChainID,
		TransactionData:   t.TransactionData,
		Value:             valueRaw,
		Signature:         t.Signature,
		CreationDate:      t.CreationDate,
		TransactionType:   t.TransactionType,
		TransactionOutput: t.TransactionOutput,
		TransactionFee:    transactionFeeRaw,
		TransactionNonce:  t.TransactionNonce,
		OutputHash:        t.OutputHash,
		Status:            t.Status,
	}

	return json.Marshal(wrapper)
}
