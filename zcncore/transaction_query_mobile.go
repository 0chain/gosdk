//go:build mobile
// +build mobile

package zcncore

import (
	"context"
	"encoding/json"
	"errors"
	stderrors "errors"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/client"
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
	SharderEndpointHealthCheck = "/_health_check"
)

type QueryResult struct {
	Content    []byte
	StatusCode int
	Error      error
}

// queryResultHandle handle query response, return true if it is a consensus-result
type queryResultHandle func(result QueryResult) bool

type transactionQuery struct {
	max      int
	sharders []string

	selected map[string]interface{}
	offline  map[string]interface{}
}

func (tq *transactionQuery) Reset() {
	tq.selected = make(map[string]interface{})
	tq.offline = make(map[string]interface{})
}

// validate validate data and input
func (tq *transactionQuery) validate(num int) error {
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
func (tq *transactionQuery) buildUrl(host string, parts ...string) string {
	var sb strings.Builder

	sb.WriteString(strings.TrimSuffix(host, "/"))

	for _, it := range parts {
		sb.WriteString(it)
	}

	return sb.String()
}

// checkHealth check health
func (tq *transactionQuery) checkHealth(ctx context.Context, host string) error {

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
func (tq *transactionQuery) randOne(ctx context.Context) (string, error) {

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

func newTransactionQuery(sharders []string) (*transactionQuery, error) {

	if len(sharders) == 0 {
		return nil, ErrNoAvailableSharders
	}

	tq := &transactionQuery{
		max:      len(sharders),
		sharders: sharders,
	}
	tq.selected = make(map[string]interface{})
	tq.offline = make(map[string]interface{})

	return tq, nil
}

// fromAll query transaction from all sharders whatever it is selected or offline in previous queires, and return consensus result
func (tq *transactionQuery) fromAll(query string, handle queryResultHandle, timeout RequestTimeout) error {
	if tq == nil || tq.max == 0 {
		return ErrNoAvailableSharders
	}

	ctx, cancel := makeTimeoutContext(timeout)
	defer cancel()

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

// fromAny query transaction from any sharder that is not selected in previous queires. use any used sharder if there is not any unused sharder
func (tq *transactionQuery) fromAny(query string, timeout RequestTimeout) (QueryResult, error) {
	res := QueryResult{
		StatusCode: http.StatusBadRequest,
	}

	ctx, cancel := makeTimeoutContext(timeout)
	defer cancel()

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

func (tq *transactionQuery) getInfo(query string, timeout RequestTimeout) (*QueryResult, error) {

	consensuses := make(map[int]int)
	var maxConsensus int
	var consensusesResp QueryResult
	// {host}{query}

	err := tq.fromAll(query,
		func(qr QueryResult) bool {
			//ignore response if it is network error
			if qr.StatusCode >= 500 {
				return false
			}

			consensuses[qr.StatusCode]++
			if consensuses[qr.StatusCode] >= maxConsensus {
				maxConsensus = consensuses[qr.StatusCode]
				consensusesResp = qr
			}

			return false

		}, timeout)

	if err != nil {
		return nil, err
	}

	if maxConsensus == 0 {
		return nil, stderrors.New("zcn: query not found")
	}

	rate := float32(maxConsensus*100) / float32(tq.max)
	if rate < consensusThresh {
		return nil, ErrInvalidConsensus
	}

	if consensusesResp.StatusCode != http.StatusOK {
		return nil, stderrors.New(string(consensusesResp.Content))
	}

	return &consensusesResp, nil
}

func (tq *transactionQuery) getConsensusConfirmation(numSharders int, txnHash string, timeout RequestTimeout) (*blockHeader, map[string]json.RawMessage, *blockHeader, error) {
	var maxConfirmation int
	txnConfirmations := make(map[string]int)
	var confirmationBlockHeader *blockHeader
	var confirmationBlock map[string]json.RawMessage
	var lfbBlockHeader *blockHeader
	maxLfbBlockHeader := int(0)
	lfbBlockHeaders := make(map[string]int)

	// {host}/v1/transaction/get/confirmation?hash={txnHash}&content=lfb
	err := tq.fromAll(tq.buildUrl("", TXN_VERIFY_URL, txnHash, "&content=lfb"),
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

		}, timeout)

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
func (tq *transactionQuery) getFastConfirmation(txnHash string, timeout RequestTimeout) (*blockHeader, map[string]json.RawMessage, *blockHeader, error) {
	var confirmationBlockHeader *blockHeader
	var confirmationBlock map[string]json.RawMessage
	var lfbBlockHeader blockHeader

	// {host}/v1/transaction/get/confirmation?hash={txnHash}&content=lfb
	result, err := tq.fromAny(tq.buildUrl("", TXN_VERIFY_URL, txnHash, "&content=lfb"), timeout)
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
	nodeClient, err := client.GetNode()
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	tq, err := newTransactionQuery(util.Shuffle(nodeClient.Sharders().Healthy()))
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	qr, err := tq.getInfo(urlSuffix, nil)
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	cb.OnInfoAvailable(op, StatusSuccess, string(qr.Content), "")
}

func GetInfoFromAnySharder(urlSuffix string, op int, cb GetInfoCallback) {
	nodeClient, err := client.GetNode()
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	tq, err := newTransactionQuery(util.Shuffle(nodeClient.Sharders().Healthy()))
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	qr, err := tq.fromAny(urlSuffix, nil)
	if err != nil {
		cb.OnInfoAvailable(op, StatusError, "", err.Error())
		return
	}

	cb.OnInfoAvailable(op, StatusSuccess, string(qr.Content), "")
}
