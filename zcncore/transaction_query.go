package zcncore

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"strings"
	"time"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/resty"
)

var (
	ErrNoAvailableSharders     = errors.New("zcn: no available sharders")
	ErrNoEnoughSharders        = errors.New("zcn: sharders is not enough")
	ErrNoEnoughOnlineSharders  = errors.New("zcn: online sharders is not enough")
	ErrInvalidNumSharder       = errors.New("zcn: number of sharders is invalid")
	ErrNoOnlineSharders        = errors.New("zcn: no any online sharder")
	ErrSharderOffline          = errors.New("zcn: sharder is offline")
	ErrInvalidConsensus        = errors.New("zcn: invalid consensus")
	ErrTransactionNotFound     = errors.New("zcn: transaction not found")
	ErrTransactionNotConfirmed = errors.New("zcn: transaction not confirmed")
)

const (
	SharderEndpointHealthCheck = "/v1/healthcheck"
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

	selected map[string]interface{}
	offline  map[string]interface{}
}

func NewTransactionQuery(sharders []string) (*TransactionQuery, error) {

	if len(sharders) == 0 {
		return nil, ErrNoAvailableSharders
	}

	tq := &TransactionQuery{
		max:      len(sharders),
		sharders: sharders,
	}
	tq.selected = make(map[string]interface{})
	tq.offline = make(map[string]interface{})

	return tq, nil
}

func (tq *TransactionQuery) Reset() {
	tq.selected = make(map[string]interface{})
	tq.offline = make(map[string]interface{})
}

// validate validate data and input
func (tq *TransactionQuery) validate(num int) error {
	if tq == nil || tq.max == 0 {
		return ErrNoAvailableSharders
	}

	if num < 1 {
		return ErrInvalidNumSharder
	}

	if num > tq.max {
		return ErrNoEnoughSharders
	}

	if num > (tq.max - len(tq.offline)) {
		return ErrNoEnoughOnlineSharders
	}

	return nil

}

// buildUrl build url with host and parts
func (tq *TransactionQuery) buildUrl(host string, parts ...string) string {
	var sb strings.Builder

	sb.WriteString(strings.TrimSuffix(host, "/"))

	for _, it := range parts {
		sb.WriteString(it)
	}

	return sb.String()
}

// checkHealth check health
func (tq *TransactionQuery) checkHealth(ctx context.Context, host string) error {

	_, ok := tq.offline[host]
	if ok {
		return ErrSharderOffline
	}

	// check health
	r := resty.New()
	requestUrl := tq.buildUrl(host, SharderEndpointHealthCheck)
	Logger.Info("zcn: check health ", requestUrl)
	r.DoGet(ctx, requestUrl)
	r.Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
		if err != nil {
			return err
		}

		// 5xx: it is a server error, not client error
		if resp.StatusCode >= http.StatusInternalServerError {
			return thrown.Throw(ErrSharderOffline, resp.Status)
		}

		return nil
	})
	errs := r.Wait()

	if len(errs) > 0 {
		tq.offline[host] = true

		if len(tq.offline) >= tq.max {
			return ErrNoOnlineSharders
		}
	}

	return nil
}

// randOne random one health sharder
func (tq *TransactionQuery) randOne(ctx context.Context) (string, error) {

	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {

		// reset selected if all sharders were selected
		if len(tq.selected) >= tq.max {
			tq.selected = make(map[string]interface{})
		}

		i := randGen.Intn(len(tq.sharders))
		host := tq.sharders[i]

		_, ok := tq.selected[host]

		// it was selected, try next
		if ok {
			continue
		}

		tq.selected[host] = true

		err := tq.checkHealth(ctx, host)

		if err != nil {
			if errors.Is(err, ErrNoOnlineSharders) {
				return "", err
			}

			// it is offline, try next one
			continue
		}

		return host, nil
	}
}
