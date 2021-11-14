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
	}
	JobError struct {
		error
	}

	proofOfBurn struct {
		TxnID             string `json:"ethereum_txn_id"`
		Amount            int64  `json:"amount"`
		ReceivingClientID string `json:"receiving_client_id"` // 0ZCN address
		Nonce             int64  `json:"nonce"`
		Signature         string `json:"signature"`
	}
)

func (e *JobError) UnmarshalJSON(buf []byte) error {
	e.error = errors.New(string(buf))
	return nil
}

func (e *JobError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Error())
}

// AuthorizerBurnEvent returned from burn ticket handler of: /v1/ether/burnticket/get
type AuthorizerBurnEvent struct {
	// 	AuthorizerID Authorizer ID
	AuthorizerID string `json:"authorizer_id,omitempty"`
	// BurnTicket Returns burn ticket
	BurnTicket *proofOfBurn `json:"ticket,omitempty"`
	// Err gives error of job on server side
	Err *JobError `json:"err,omitempty"`
	// Status gives job status on server side (authoriser)
	Status JobStatus `json:"status,omitempty"`
}

func (r *AuthorizerBurnEvent) Error() error {
	return r.Err
}

func (r *AuthorizerBurnEvent) Data() interface{} {
	return r.BurnTicket
}
