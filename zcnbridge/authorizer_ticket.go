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
)

func (e *JobError) UnmarshalJSON(buf []byte) error {
	e.error = errors.New(string(buf))
	return nil
}

func (e *JobError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Error())
}

// AuthorizerBurnEvent returned from burn ticket handler Example: /v1/ether/burnticket/get
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
