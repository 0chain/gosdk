package zcnbridge

import (
	"encoding/json"
	"errors"
)

type (
	// JobStatus = Ethereum transaction status
	JobStatus uint
	// JobResult = Authorizer task result, it wraps actual result of the query inside authorizer
	JobResult interface {
		// Error = Status of Authorizer job on authorizer server
		Error() error
		// Data returns the actual result
		Data() interface{}
		// SetAuthorizerID Assigns authorizer ID to the Job
		SetAuthorizerID(ID string)
		// GetAuthorizerID returns authorizer ID
		GetAuthorizerID() string
	}
	// JobError result of internal request wrapped in authorizer job
	JobError struct {
		error
	}

	// EthereumBurnEvents represents burn events returned by authorizers
	EthereumBurnEvents struct {
		AuthorizerID string `json:"authorizer_id,omitempty"`
		BurnEvents   []struct {
			Nonce           int64  `json:"nonce"`
			Amount          int64  `json:"amount"`
			TransactionHash string `json:"transaction_hash"`
		} `json:"burn_events"`
	}

	// ProofEthereumBurn Authorizer returns this type for Ethereum transaction
	ProofEthereumBurn struct {
		TxnID             string `json:"ethereum_txn_id"`
		Nonce             int64  `json:"nonce"`
		Amount            int64  `json:"amount"`
		ReceivingClientID string `json:"receiving_client_id"` // 0ZCN address
		Signature         string `json:"signature"`
	}

	// ProofZCNBurn Authorizer returns this type for ZCN transaction
	ProofZCNBurn struct {
		AuthorizerID string `json:"authorizer_id,omitempty"`
		TxnID        string `json:"0chain_txn_id"`
		To           string `json:"to"`
		Nonce        int64  `json:"nonce"`
		Amount       int64  `json:"amount"`
		Signature    []byte `json:"signature"`
	}
)

func (e *JobError) UnmarshalJSON(buf []byte) error {
	e.error = errors.New(string(buf))
	return nil
}

func (e *JobError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Error())
}

// WZCNBurnEvent returned from burn ticket handler of: /v1/ether/burnticket/get
type WZCNBurnEvent struct {
	// 	AuthorizerID Authorizer ID
	AuthorizerID string `json:"authorizer_id,omitempty"`
	// BurnTicket Returns burn ticket
	BurnTicket *ProofEthereumBurn `json:"ticket,omitempty"`
	// Err gives error of job on server side
	Err *JobError `json:"err,omitempty"`
	// Status gives job status on server side (authoriser)
	Status JobStatus `json:"status,omitempty"`
}

func (r *WZCNBurnEvent) GetAuthorizerID() string {
	return r.AuthorizerID
}

func (r *WZCNBurnEvent) SetAuthorizerID(id string) {
	r.AuthorizerID = id
}

func (r *WZCNBurnEvent) Error() error {
	return r.Err
}

func (r *WZCNBurnEvent) Data() interface{} {
	return r.BurnTicket
}

func (r *EthereumBurnEvents) GetAuthorizerID() string {
	return r.AuthorizerID
}

func (r *EthereumBurnEvents) SetAuthorizerID(id string) {
	r.AuthorizerID = id
}

func (r *EthereumBurnEvents) Error() error {
	return nil
}

func (r *EthereumBurnEvents) Data() interface{} {
	return r
}

func (r *ProofZCNBurn) GetAuthorizerID() string {
	return r.AuthorizerID
}

func (r *ProofZCNBurn) SetAuthorizerID(id string) {
	r.AuthorizerID = id
}

func (r *ProofZCNBurn) Error() error {
	return nil
}

func (r *ProofZCNBurn) Data() interface{} {
	return r
}
