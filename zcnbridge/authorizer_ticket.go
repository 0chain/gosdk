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

// authorizerBurnEvent The output of an authorizer job
type authorizerBurnEvent struct {
	// 	AuthorizerID Authorizer ID
	AuthorizerID string `json:"authorizer_id"`
	// BurnTicket Returns burn ticket
	BurnTicket *proofOfBurn `json:"ticket,omitempty"`
	// Err gives error of job on server side
	Err *JobError `json:"err,omitempty"`
	// Status gives job status on server side
	Status JobStatus `json:"status,omitempty"`
}

func (r *authorizerBurnEvent) Error() error {
	return r.Err
}

func (r *authorizerBurnEvent) Data() interface{} {
	return r.BurnTicket
}
