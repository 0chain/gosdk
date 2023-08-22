//go:build !mobile
// +build !mobile

package zcncore

import (
	"context"
	"encoding/json"
	"errors"
	stderrors "errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
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
	ErrNoAvailableMiners       = errors.New("zcn: no available miners")
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
	sync.RWMutex
	max                int
	sharders           []string
	miners             []string
	numShardersToBatch int

	selected map[string]interface{}
	offline  map[string]interface{}
}

func NewTransactionQuery(sharders []string, miners []string) (*TransactionQuery, error) {

	if len(sharders) == 0 {
		return nil, ErrNoAvailableSharders
	}

	tq := &TransactionQuery{
		max:                len(sharders),
		sharders:           sharders,
		numShardersToBatch: 3,
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

// checkSharderHealth checks the health of a sharder (denoted by host) and returns if it is healthy
// or ErrNoOnlineSharders if no sharders are healthy/up at the moment.
func (tq *TransactionQuery) checkSharderHealth(ctx context.Context, host string) error {
	tq.RLock()
	_, ok := tq.offline[host]
	tq.RUnlock()
	if ok {
		return ErrSharderOffline
	}

	// check health
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
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
		if errors.Is(errs[0], context.DeadlineExceeded) {
			return context.DeadlineExceeded
		}
		tq.Lock()
		tq.offline[host] = true
		tq.Unlock()
		return ErrSharderOffline
	}
	return nil
}

// getRandomSharder returns a random healthy sharder
func (tq *TransactionQuery) getRandomSharder(ctx context.Context) (string, error) {
	if tq.sharders == nil || len(tq.sharders) == 0 {
		return "", ErrNoAvailableMiners
	}

	shuffledSharders := util.Shuffle(tq.sharders)

	return shuffledSharders[0], nil
}

//nolint:unused
func (tq *TransactionQuery) getRandomSharderWithHealthcheck(ctx context.Context) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	shuffledSharders := util.Shuffle(tq.sharders)
	for i := 0; i < len(shuffledSharders); i += tq.numShardersToBatch {
		var mu sync.Mutex
		done := false
		errCh := make(chan error, tq.numShardersToBatch)
		successCh := make(chan string)
		last := i + tq.numShardersToBatch - 1

		if last > len(shuffledSharders)-1 {
			last = len(shuffledSharders) - 1
		}
		numShardersOffline := 0
		for j := i; j <= last; j++ {
			sharder := shuffledSharders[j]
			go func(sharder string) {
				err := tq.checkSharderHealth(ctx, sharder)
				if err != nil {
					errCh <- err
				} else {
					mu.Lock()
					if !done {
						successCh <- sharder
						done = true
					}
					mu.Unlock()
				}
			}(sharder)
		}
	innerLoop:
		for {
			select {
			case e := <-errCh:
				switch e {
				case ErrSharderOffline:
					tq.RLock()
					if len(tq.offline) >= tq.max {
						tq.RUnlock()
						return "", ErrNoOnlineSharders
					}
					tq.RUnlock()
					numShardersOffline++
					if numShardersOffline >= tq.numShardersToBatch {
						break innerLoop
					}
				case context.DeadlineExceeded:
					return "", e
				}
			case s := <-successCh:
				return s, nil
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					return "", context.DeadlineExceeded
				}
			}
		}
	}
	return "", ErrNoOnlineSharders
}

//getRandomMiner returns a random miner
func (tq *TransactionQuery) getRandomMiner(ctx context.Context) (string, error) {

	if tq.miners == nil || len(tq.miners) == 0 {
		return "", ErrNoAvailableMiners
	}

	shuffledMiners := util.Shuffle(tq.miners)

	return shuffledMiners[0], nil
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
				if handle(res) {

					cf()
				}
			}

			return nil
		})

	r.Wait()

	return nil
}

func (tq *TransactionQuery) GetInfo(ctx context.Context, query string) (*QueryResult, error) {

	consensuses := make(map[int]int)
	var maxConsensus int
	var consensusesResp QueryResult
	// {host}{query}
	err := tq.FromAll(ctx, query,
		func(qr QueryResult) bool {
			//ignore response if it is network error
			if qr.StatusCode >= 500 {
				return false
			}

			consensuses[qr.StatusCode]++
			if consensuses[qr.StatusCode] > maxConsensus {
				maxConsensus = consensuses[qr.StatusCode]
				consensusesResp = qr
			}

			// If number of 200's is equal to number of some other status codes, use 200's.
			if qr.StatusCode == http.StatusOK && consensuses[qr.StatusCode] == maxConsensus {
				maxConsensus = consensuses[qr.StatusCode]
				consensusesResp = qr
			}

			return false

		})

	if err != nil {
		return nil, err
	}

	if maxConsensus == 0 {
		return nil, stderrors.New("zcn: query not found")
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

// FromAny queries transaction from any sharder that is not selected in previous queries.
// use any used sharder if there is not any unused sharder
func (tq *TransactionQuery) FromAny(ctx context.Context, query string, provider Provider) (QueryResult, error) {

	res := QueryResult{
		StatusCode: http.StatusBadRequest,
	}

	err := tq.validate(1)

	if err != nil {
		return res, err
	}

	var host string

	// host, err := tq.getRandomSharder(ctx)

	switch provider {
	case ProviderMiner:
		host, err = tq.getRandomMiner(ctx)
	case ProviderSharder:
		host, err = tq.getRandomSharder(ctx)
	}

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
		func(qr QueryResult) bool {
			if qr.StatusCode != http.StatusOK {
				return false
			}

			var cfmBlock map[string]json.RawMessage
			err := json.Unmarshal([]byte(qr.Content), &cfmBlock)
			if err != nil {
				logging.Error("txn confirmation parse error", err)
				return false
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
						return false
					}

					lfbBlockHeaders[lfb.Hash]++
					if lfbBlockHeaders[lfb.Hash] > maxLfbBlockHeader {
						maxLfbBlockHeader = lfbBlockHeaders[lfb.Hash]
						lfbBlockHeader = &lfb
					}
				}

				return false
			}

			txnConfirmations[cfmBlockHeader.Hash]++
			if txnConfirmations[cfmBlockHeader.Hash] > maxConfirmation {
				maxConfirmation = txnConfirmations[cfmBlockHeader.Hash]

				if maxConfirmation >= numSharders {
					confirmationBlockHeader = cfmBlockHeader
					confirmationBlock = cfmBlock

					// it is consensus by enough sharders, and latest_finalized_block is valid
					// return true to cancel other requests
					return true
				}
			}

			return false

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
	result, err := tq.FromAny(ctx, tq.buildUrl("", TXN_VERIFY_URL, txnHash, "&content=lfb"), ProviderSharder)
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

func GetInfoFromSharders(urlSuffix string, op int, cb GetInfoCallback) {

	tq, err := NewTransactionQuery(util.Shuffle(_config.chain.Sharders), []string{})
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

	logging.Info("sharder url suffix", urlSuffix)
	

	tq, err := NewTransactionQuery(util.Shuffle(_config.chain.Sharders), []string{})
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	logging.Info("transaction query", tq)
	logging.Info("sharders", _config.chain.Sharders)

	qr, err := tq.FromAny(context.TODO(), urlSuffix, ProviderSharder)
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	cb.OnInfoAvailable(op, StatusSuccess, string(qr.Content), "")
}

func GetInfoFromAnyMiner(urlSuffix string, op int, cb getInfoCallback) {

	tq, err := NewTransactionQuery([]string{}, util.Shuffle(_config.chain.Miners))

	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}
	qr, err := tq.FromAny(context.TODO(), urlSuffix, ProviderMiner)

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
