package zcncore

import (
	"context"
	"encoding/json"
	"errors"
	stderrors "errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/util"
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
type QueryResultHandle func(result QueryResult) (status, consensus bool)

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
	logging.Info("zcn: check health ", requestUrl)
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
			tq.selected = make(map[string]bool)
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

func (tq *TransactionQuery) FromConsensus(ctx context.Context, query string, handle QueryResultHandle) error {

	min := getMinShardersVerify()
	next := make(chan string, 1)
	result := make(chan QueryResult, min)

	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			case sharder := <-next:
				go func(sharder string) {

					err := tq.checkHealth(ctx, sharder)
					if err != nil {
						result <- QueryResult{
							Error:      err,
							StatusCode: http.StatusServiceUnavailable,
						}
						return
					}

					url := tq.buildUrl(sharder, query)
					r := resty.New(resty.WithTimeout(15 * time.Second))
					r.DoGet(ctx, url).
						Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
							res := QueryResult{
								Content:    respBody,
								Error:      err,
								StatusCode: http.StatusBadRequest,
							}

							if resp != nil {
								res.StatusCode = resp.StatusCode

								logging.Debug(req.URL.String() + " " + resp.Status)
								logging.Debug(string(respBody))
							} else {
								logging.Debug(req.URL.String())
							}

							result <- res

							return nil
						})

					r.Wait()

				}(sharder)
			}
		}
	}()

	// send requests to min sharders first
	for i := 0; i < min; i++ {
		next <- tq.sharders[i]
	}

	sentCount := min
	for {
		select {
		case <-ctx.Done():
		case res := <-result:
			status, consensus := handle(res)
			if status {
				sentCount++
				if consensus {
					return nil
				}

				if sentCount >= tq.max {
					return nil
				}

				// query success, but doesn't reach consensus, waiting for other sharder's response
				continue
			}

			if sentCount+1 >= tq.max {
				return nil
			}

			// query failed: 5xx,timeout, push an additional sharder to query
			next <- tq.sharders[sentCount+1]
		}
	}

	return nil
}

// FromAll query transaction from all sharders whatever it is selected or offline in previous queires, and return consensus result
func (tq *TransactionQuery) FromAll(ctx context.Context, query string, handle QueryResultHandle) error {
	if tq == nil || tq.max == 0 {
		return ErrNoAvailableSharders
	}

	urls := make([]string, 0, tq.max)
	for _, host := range tq.sharders {
		urls = append(urls, tq.buildUrl(host, query))
	}

	r := resty.New()
	r.DoGet(ctx, urls...).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			res := QueryResult{
				Content:    respBody,
				Error:      err,
				StatusCode: http.StatusBadRequest,
			}

			if resp != nil {
				res.StatusCode = resp.StatusCode

				logging.Debug(req.URL.String() + " " + resp.Status)
				logging.Debug(string(respBody))
			} else {
				logging.Debug(req.URL.String())

			}

			if handle != nil {
				if _, consensus := handle(res); consensus {

					cf()
				}
			}

			return nil
		})

	r.Wait()

	return nil
}

// FromAny query transaction from any sharder that is not selected in previous queires. use any used sharder if there is not any unused sharder
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

	logging.Debug("GET", requestUrl)

	r.DoGet(ctx, requestUrl).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			res.Error = err
			if err != nil {
				return err
			}

			res.Content = respBody
			logging.Debug(string(respBody))

			if resp != nil {
				res.StatusCode = resp.StatusCode
			}

			return nil
		})

	errs := r.Wait()

	if len(errs) > 0 {
		return res, errs[0]
	}

	return res, nil

}

func (tq *TransactionQuery) getConsensusConfirmation(ctx context.Context, numSharders int, txnHash string) (*blockHeader, map[string]json.RawMessage, *blockHeader, error) {
	maxConfirmation := int(0)
	txnConfirmations := make(map[string]int)
	var confirmationBlockHeader *blockHeader
	var confirmationBlock map[string]json.RawMessage
	var lfbBlockHeader *blockHeader
	maxLfbBlockHeader := int(0)
	lfbBlockHeaders := make(map[string]int)

	// {host}/v1/transaction/get/confirmation?hash={txnHash}&content=lfb
	err := tq.FromAll(ctx,
		tq.buildUrl("", TXN_VERIFY_URL, txnHash, "&content=lfb"),
		func(qr QueryResult) (bool, bool) {
			if qr.StatusCode != http.StatusOK {
				return false, false
			}

			var cfmBlock map[string]json.RawMessage
			err := json.Unmarshal([]byte(qr.Content), &cfmBlock)
			if err != nil {
				logging.Error("txn confirmation parse error", err)
				return false, false
			}

			// parse `confirmation` section as block header
			cfmBlockHeader, err := getBlockHeaderFromTransactionConfirmation(txnHash, cfmBlock)
			if err != nil {
				logging.Error("txn confirmation parse header error", err)

				// parse `latest_finalized_block` section
				if lfbRaw, ok := cfmBlock["latest_finalized_block"]; ok {
					var lfb blockHeader
					err := json.Unmarshal([]byte(lfbRaw), &lfb)
					if err != nil {
						logging.Error("round info parse error.", err)
						return false, false
					}

					lfbBlockHeaders[lfb.Hash]++
					if lfbBlockHeaders[lfb.Hash] > maxLfbBlockHeader {
						maxLfbBlockHeader = lfbBlockHeaders[lfb.Hash]
						lfbBlockHeader = &lfb
					}
				}

				return false, false
			}

			txnConfirmations[cfmBlockHeader.Hash]++
			if txnConfirmations[cfmBlockHeader.Hash] > maxConfirmation {
				maxConfirmation = txnConfirmations[cfmBlockHeader.Hash]

				if maxConfirmation >= numSharders {
					confirmationBlockHeader = cfmBlockHeader
					confirmationBlock = cfmBlock

					// it is consensus by enough sharders, and latest_finalized_block is valid
					// return true to cancel other requests
					return true, true
				}
			}

			return true, false

		})

	if err != nil {
		return nil, nil, lfbBlockHeader, err
	}

	if maxConfirmation == 0 {
		return nil, nil, lfbBlockHeader, stderrors.New("zcn: transaction not found")
	}

	if maxConfirmation < numSharders {
		return nil, nil, lfbBlockHeader, ErrInvalidConsensus
	}

	return confirmationBlockHeader, confirmationBlock, lfbBlockHeader, nil
}

// getFastConfirmation get txn confirmation from a random online sharder
func (tq *TransactionQuery) getFastConfirmation(ctx context.Context, txnHash string) (*blockHeader, map[string]json.RawMessage, *blockHeader, error) {
	var confirmationBlockHeader *blockHeader
	var confirmationBlock map[string]json.RawMessage
	var lfbBlockHeader blockHeader

	// {host}/v1/transaction/get/confirmation?hash={txnHash}&content=lfb
	result, err := tq.FromAny(ctx, tq.buildUrl("", TXN_VERIFY_URL, txnHash, "&content=lfb"))
	if err != nil {
		return nil, nil, nil, err
	}

	if result.StatusCode == http.StatusOK {

		err = json.Unmarshal(result.Content, &confirmationBlock)
		if err != nil {
			logging.Error("txn confirmation parse error", err)
			return nil, nil, nil, err
		}

		// parse `confirmation` section as block header
		confirmationBlockHeader, err = getBlockHeaderFromTransactionConfirmation(txnHash, confirmationBlock)
		if err == nil {
			return confirmationBlockHeader, confirmationBlock, nil, nil
		}

		logging.Error("txn confirmation parse header error", err)

		// parse `latest_finalized_block` section
		lfbRaw, ok := confirmationBlock["latest_finalized_block"]
		if !ok {
			return confirmationBlockHeader, confirmationBlock, nil, err
		}

		err = json.Unmarshal([]byte(lfbRaw), &lfbBlockHeader)
		if err == nil {
			return confirmationBlockHeader, confirmationBlock, &lfbBlockHeader, ErrTransactionNotConfirmed
		}

		logging.Error("round info parse error.", err)
		return nil, nil, nil, err

	}

	return nil, nil, nil, thrown.Throw(ErrTransactionNotFound, strconv.Itoa(result.StatusCode))
}

func (tq *TransactionQuery) GetInfo(ctx context.Context, query string) (*QueryResult, error) {

	consensuses := make(map[int]int)
	var maxConsensus int
	var consensusesResp QueryResult
	// {host}{query}
	err := tq.FromConsensus(ctx, query,
		func(qr QueryResult) (bool, bool) {
			//ignore response if it is network error
			if qr.StatusCode >= 500 {
				return false, false
			}

			consensuses[qr.StatusCode]++
			if consensuses[qr.StatusCode] > maxConsensus {
				maxConsensus = consensuses[qr.StatusCode]
				consensusesResp = qr
			}

			// If number of 200's is equal to number of some other status codes, use 200's response. it fixes the random issue about 50% failed and 50% success on 2 sharders
			if qr.StatusCode == http.StatusOK {

				if consensuses[http.StatusOK] == maxConsensus {
					maxConsensus = consensuses[qr.StatusCode]
					consensusesResp = qr
				}

				// consensus has been reached, don't waiting for other requests
				if maxConsensus*100/tq.max >= consensusThresh {
					return true, true
				}

				// query success, but doesn't reach consensus
				return true, false
			}

			// query fails
			return false, false

		})

	if err != nil {
		return nil, err
	}

	if maxConsensus == 0 {
		return nil, stderrors.New("zcn: object not found")
	}

	rate := maxConsensus * 100 / tq.max
	if rate < consensusThresh {
		return nil, ErrInvalidConsensus
	}

	if consensusesResp.StatusCode != http.StatusOK {
		return nil, stderrors.New(string(consensusesResp.Content))
	}

	return &consensusesResp, nil
}

func MakeSCRestAPICall(scAddress string, relativePath string, params map[string]string) ([]byte, error) {

	path := fmt.Sprintf("/v1/screst/%v%v", scAddress, relativePath)
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

func GetInfoFromSharders(urlSuffix string, op int, cb GetInfoCallback) {

	tq, err := NewTransactionQuery(util.Shuffle(_config.chain.Sharders))
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	qr, err := tq.GetInfo(context.TODO(), urlSuffix)
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	cb.OnInfoAvailable(op, StatusSuccess, string(qr.Content), "")
}

func GetInfoFromAnySharder(urlSuffix string, op int, cb GetInfoCallback) {

	tq, err := NewTransactionQuery(util.Shuffle(_config.chain.Sharders))
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	qr, err := tq.FromAny(context.TODO(), urlSuffix)
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	cb.OnInfoAvailable(op, StatusSuccess, string(qr.Content), "")
}

func GetEvents(cb GetInfoCallback, filters map[string]string) (err error) {
	if err = CheckConfig(); err != nil {
		return
	}
	go GetInfoFromSharders(WithParams(GET_MINERSC_EVENTS, Params{
		"block_number": filters["block_number"],
		"tx_hash":      filters["tx_hash"],
		"type":         filters["type"],
		"tag":          filters["tag"],
	}), 0, cb)
	return
}

func WithParams(uri string, params Params) string {
	return withParams(uri, params)
}
