package zcnbridge

import (
	"encoding/json"
	"errors"
)

type (
	JobStatus uint
	JobResult interface {
		Error() error
		Data() interface{}
		SetAuthorizerID(ID string)
		GetAuthorizerID() string
	}
	JobError struct {
		error
	}

	proofEthereumBurn struct {
		TxnID             string `json:"ethereum_txn_id"`
		Nonce             int64  `json:"nonce"`
		Amount            int64  `json:"amount"`
		ReceivingClientID string `json:"receiving_client_id"` // 0ZCN address
		Signature         string `json:"signature"`
	}

	proofZCNBurn struct {
		TxnID           string `json:"0chain_txn_id"`
		Nonce           int64  `json:"nonce"`
		Amount          int64  `json:"amount"`
		EthereumAddress string `json:"ethereum_address"`
		Signature       string `json:"signatures"`
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
	BurnTicket *proofEthereumBurn `json:"ticket,omitempty"`
	// Err gives error of job on server side
	Err *JobError `json:"err,omitempty"`
	// Status gives job status on server side (authoriser)
	Status JobStatus `json:"status,omitempty"`
}

func (r *WZCNBurnEvent) GetAuthorizerID() string {
	return r.AuthorizerID
}

func (r *WZCNBurnEvent) SetAuthorizerID(ID string) {
	r.AuthorizerID = ID
}

func (r *WZCNBurnEvent) Error() error {
	return r.Err
}

func (r *WZCNBurnEvent) Data() interface{} {
	return r.BurnTicket
}

// ZCNBurnEvent ZCN burn ticket
type ZCNBurnEvent struct {
	// 	AuthorizerID Authorizer ID
	AuthorizerID string `json:"authorizer_id,omitempty"`
	// BurnTicket Returns burn ticket
	BurnTicket *proofZCNBurn `json:"ticket,omitempty"`
	// Err gives error of job on server side
	Err *JobError `json:"err,omitempty"`
}

func (r *ZCNBurnEvent) GetAuthorizerID() string {
	return r.AuthorizerID
}

func (r *ZCNBurnEvent) SetAuthorizerID(ID string) {
	r.AuthorizerID = ID
}

func (r *ZCNBurnEvent) Error() error {
	return r.Err
}

func (r *ZCNBurnEvent) Data() interface{} {
	return r.BurnTicket
}
