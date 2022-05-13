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
)

const (
	SharderEndpointHealthCheck = "/_health_check"
)

type QueryResult struct {
	Content    []byte
	StatusCode int
	Error      error
}

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

	result := QueryResult{
		StatusCode: http.StatusBadRequest,
	}

	err := tq.validate(1)

	if err != nil {
		return result, err
	}

	host, err := tq.randOne(ctx)

	if err != nil {
		return result, err
	}

	r := resty.New()
	requestUrl := tq.buildUrl(host, query)

	Logger.Debug("GET", requestUrl)

	r.DoGet(ctx, requestUrl).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			Logger.Debug(requestUrl + " " + resp.Status)

			result.Error = err
			if err != nil {
				return err
			}

			result.Content = respBody
			result.StatusCode = resp.StatusCode

			Logger.Debug(string(respBody))

			return nil
		})

	errs := r.Wait()

	if len(errs) > 0 {
		return result, errs[0]
	}

	return result, nil

}

// From query transaction from s/S sharders whatever it is selected in previous queires
func (tq *TransactionQuery) From(ctx context.Context, numSharders int, query string) ([]QueryResult, error) {
	err := tq.validate(numSharders)

	if err != nil {
		return nil, err
	}

	urls := make([]string, 0, numSharders)
	selected := make(map[string]interface{})
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {

		if len(selected) >= tq.max {
			return nil, ErrNoOnlineSharders
		}

		i := randGen.Intn(len(tq.sharders))
		host := tq.sharders[i]

		_, ok := selected[host]

		// it was selected, try next
		if ok {
			continue
		}

		selected[host] = true

		err := tq.checkHealth(ctx, host)

		if err != nil {
			if errors.Is(err, ErrNoOnlineSharders) {
				return nil, err
			}

			// it is offline, try next one
			continue
		}

		urls = append(urls, tq.buildUrl(host, query))

		if len(urls) >= numSharders {
			break
		}
	}

	results := make([]QueryResult, 0, numSharders)

	r := resty.New()
	r.DoGet(ctx, urls...).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			result := QueryResult{
				Content:    respBody,
				Error:      err,
				StatusCode: http.StatusBadRequest,
			}

			if resp != nil {
				result.StatusCode = resp.StatusCode
			}

			results = append(results, result)

			return nil
		})

	r.Wait()

	return results, nil
}

// FromAll query transaction from all sharders whatever it is selected or offline in previous queires
func (tq *TransactionQuery) FromAll(ctx context.Context, query string) ([]QueryResult, error) {
	if tq == nil || tq.max == 0 {
		return nil, ErrNoAvailableSharders
	}

	urls := make([]string, 0, tq.max)
	for _, host := range tq.sharders {
		urls = append(urls, tq.buildUrl(host, query))
	}

	results := make([]QueryResult, 0, tq.max)

	r := resty.New()
	r.DoGet(ctx, urls...).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			result := QueryResult{
				Content:    respBody,
				Error:      err,
				StatusCode: http.StatusBadRequest,
			}
			Logger.Debug(req.URL.String() + " " + resp.Status)

			if resp != nil {
				result.StatusCode = resp.StatusCode
				Logger.Debug(string(respBody))
			}

			results = append(results, result)

			return nil
		})

	r.Wait()

	return results, nil
}

// GetFastConfirmation get txn confirmation from a random online sharder
func (tq *TransactionQuery) GetFastConfirmation(ctx context.Context, txnHash string) (*blockHeader, map[string]json.RawMessage, *blockHeader, error) {

	// {host}/v1/transaction/get/confirmation?hash={txnHash}&&content=lfb
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

func (tq *TransactionQuery) GetConfirmation(ctx context.Context, numSharders int, txnHash string) (*blockHeader, map[string]json.RawMessage, *blockHeader, error) {

	// {host}/v1/transaction/get/confirmation?hash={txnHash}&&content=lfb
	results, err := tq.From(ctx, numSharders, tq.buildUrl("", TXN_VERIFY_URL, txnHash, "&content=lfb"))
	if err != nil {
		return nil, nil, nil, err
	}

	maxConfirmation := int(0)
	txnConfirmations := make(map[string]int)
	var blockHdr *blockHeader
	var lfb blockHeader
	var confirmation map[string]json.RawMessage
	for _, result := range results {

		if result.StatusCode == http.StatusOK {
			var cfmLfb map[string]json.RawMessage
			err := json.Unmarshal([]byte(result.Content), &cfmLfb)
			if err != nil {
				Logger.Error("txn confirmation parse error", err)
				continue
			}
			bH, err := getBlockHeaderFromTransactionConfirmation(txnHash, cfmLfb)

			if err == nil {
				txnConfirmations[bH.Hash]++
				if txnConfirmations[bH.Hash] > maxConfirmation {
					maxConfirmation = txnConfirmations[bH.Hash]
					blockHdr = bH
					confirmation = cfmLfb
				}

				// check next result
				continue
			}

			lfbRaw, ok := cfmLfb["latest_finalized_block"]
			if !ok {
				// don't have `confirmation` or `latest_finalized_block` in block, logging it for debugging
				Logger.Error(err)
				continue
			}
			err = json.Unmarshal([]byte(lfbRaw), &lfb)
			if err != nil {
				Logger.Error("round info parse error.", err)
			}

		}
	}
	if maxConfirmation == 0 {
		return nil, confirmation, &lfb, errors.New("transaction not found")
	}
	return blockHdr, confirmation, &lfb, nil
}
