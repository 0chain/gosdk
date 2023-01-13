package zcncore

import (
	"context"
	"fmt"

	"github.com/0chain/gosdk/core/util"
)

type QueryResult struct {
	Content    []byte
	StatusCode int
	Error      error
}

// QueryResultHandle handle query response, return true if it is a consensus-result
type QueryResultHandle func(result QueryResult) bool

type TransactionQuery struct {
	max      int
	sharders []string

	selected map[string]bool
	offline  map[string]bool
}

func NewTransactionQuery(sharders []string) (*TransactionQuery, error) {

	if len(sharders) == 0 {
		return nil, ErrNoAvailableSharders
	}

	tq := &TransactionQuery{
		max:      len(sharders),
		sharders: sharders,
	}
	tq.selected = make(map[string]bool)
	tq.offline = make(map[string]bool)

	return tq, nil
}

func (tq *TransactionQuery) Reset() {
	tq.selected = make(map[string]bool)
	tq.offline = make(map[string]bool)
}

func MakeSCRestAPICall(scAddress string, relativePath string, params map[string]string) ([]byte, error) {

	path := fmt.Sprintf("/v1/screst/%v/%v", scAddress, relativePath)
	query := withParams(path, Params(params))

	tq, err := NewTransactionQuery(util.Shuffle(_config.chain.Sharders))
	if err != nil {
		return nil, err
	}

	qr, err := tq.GetInfo(context.TODO(), query)
	if err != nil {
		return nil, err
	}

	return qr.Content, nil
}
