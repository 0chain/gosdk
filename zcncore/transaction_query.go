package zcncore

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"strings"
	"time"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/resty"
)

var (
	ErrNoAvailableSharders    = errors.New("zcn: no available sharders")
	ErrNoEnoughSharders       = errors.New("zcn: sharders is not enough")
	ErrNoEnoughOnlineSharders = errors.New("zcn: online sharders is not enough")
	ErrInvalidNumSharder      = errors.New("zcn: number of sharders is invalid")
	ErrNoOnlineSharders       = errors.New("zcn: no any online sharder")
	ErrSharderOffline         = errors.New("zcn: sharder is offline")
	ErrNoConsensus            = errors.New("zcn: no valid consensus")
)

const (
	SharderEndpointHealthCheck = "/_health_check"
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

func (tq *TransactionQuery) Reset() {
	tq.selected = make(map[string]interface{})
	tq.offline = make(map[string]interface{})
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

// FromAny query transaction form any sharder that is not selected in previous queires. use any used sharder if there is not any unused sharder
func (tq *TransactionQuery) FromAny(ctx context.Context, query string) (QueryResult, error) {

	res := QueryResult{
		StatusCode: http.StatusBadRequest,
	}

	err := tq.validate(1)

	if err != nil {
		return res, err
	}

	host, err := tq.randOne(ctx)

	if err != nil {
		return res, err
	}

	r := resty.New()
	requestUrl := tq.buildUrl(host, query)

	Logger.Debug("GET", requestUrl)

	r.DoGet(ctx, requestUrl).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			Logger.Debug(requestUrl + " " + resp.Status)

			res.Error = err
			if err != nil {
				return err
			}

			res.Content = respBody
			res.StatusCode = resp.StatusCode

			Logger.Debug(string(respBody))

			return nil
		})

	errs := r.Wait()

	if len(errs) > 0 {
		return res, errs[0]
	}

	return res, nil

}

// FromConsensus query transaction from all sharders whatever it is selected or offline in previous queires, and return consensus result
func (tq *TransactionQuery) FromConsensus(ctx context.Context, query string, handle QueryResultHandle) (*QueryResult, error) {
	if tq == nil || tq.max == 0 {
		return nil, ErrNoAvailableSharders
	}

	urls := make([]string, 0, tq.max)
	for _, host := range tq.sharders {
		urls = append(urls, tq.buildUrl(host, query))
	}

	var result *QueryResult

	r := resty.New()
	r.DoGet(ctx, urls...).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			res := QueryResult{
				Content:    respBody,
				Error:      err,
				StatusCode: http.StatusBadRequest,
			}
			Logger.Debug(req.URL.String() + " " + resp.Status)

			if resp != nil {
				res.StatusCode = resp.StatusCode
				Logger.Debug(string(respBody))
			}

			if handle != nil {
				if handle(res) {
					result = &res
					cf()
				}
			}

			return nil
		})

	r.Wait()

	if result == nil {
		return nil, ErrNoConsensus
	}

	return result, nil
}

// GetFastConfirmation get txn confirmation from a random online sharder
func (tq *TransactionQuery) GetFastConfirmation(ctx context.Context, txnHash string) (*blockHeader, map[string]json.RawMessage, *blockHeader, error) {

	// {host}/v1/transaction/get/confirmation?hash={txnHash}&content=lfb
	result, err := tq.FromAny(ctx, tq.buildUrl("", TXN_VERIFY_URL, txnHash, "&content=lfb"))
	if err != nil {
		return nil, nil, nil, err
	}

	if result.StatusCode == http.StatusOK {
		var confirmation map[string]json.RawMessage
		err := json.Unmarshal(result.Content, &confirmation)
		if err != nil {
			Logger.Error("txn confirmation parse error", err)
			return nil, nil, nil, err
		}
		bH, err := getBlockHeaderFromTransactionConfirmation(txnHash, confirmation)

		if err == nil {
			return bH, confirmation, nil, nil
		}

		if lfbRaw, ok := confirmation["latest_finalized_block"]; ok {
			var lfb blockHeader
			err := json.Unmarshal([]byte(lfbRaw), &lfb)
			if err == nil {
				return nil, confirmation, &lfb, nil
			}

			Logger.Error("round info parse error.", err)
			return nil, confirmation, nil, err
		}

		Logger.Error(err)
	}

	return nil, nil, nil, errors.New("zcn: transaction not found ")
}

func (tq *TransactionQuery) GetConfirmation(ctx context.Context, numSharders int, txnHash string) (*blockHeader, error) {

	maxConfirmation := int(0)
	txnConfirmations := make(map[string]int)
	var blockHdr *blockHeader

	// {host}/v1/transaction/get/confirmation?hash={txnHash}&content=lfb
	_, err := tq.FromConsensus(ctx,
		tq.buildUrl("", TXN_VERIFY_URL, txnHash, "&content=lfb"),
		func(qr QueryResult) bool {
			if qr.StatusCode == http.StatusOK {
				var cfmLfb map[string]json.RawMessage
				err := json.Unmarshal([]byte(qr.Content), &cfmLfb)
				if err != nil {
					Logger.Error("txn confirmation parse error", err)
					return false
				}

				// {
				//	"confirmation":{}
				//}
				bH, err := getBlockHeaderFromTransactionConfirmation(txnHash, cfmLfb)

				// confirmation section found
				if err == nil {
					txnConfirmations[bH.Hash]++
					if txnConfirmations[bH.Hash] > maxConfirmation {
						maxConfirmation = txnConfirmations[bH.Hash]
						blockHdr = bH
					}

					// it is consensus by required sharders
					if maxConfirmation >= numSharders {
						// return true to cancel other requests
						return true
					}
				}

			}

			return false
		})

	if err != nil {
		return nil, err
	}

	return blockHdr, nil
}
