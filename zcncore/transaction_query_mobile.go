//go:build mobile
// +build mobile

package zcncore

import (
	"context"
	"net/http"

	"github.com/0chain/gosdk/core/resty"
)

// fromAll query transaction from all sharders whatever it is selected or offline in previous queires, and return consensus result
func (tq *TransactionQuery) fromAll(query string, handle QueryResultHandle, timeout RequestTimeout) error {
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
func (tq *TransactionQuery) fromAny(query string, timeout RequestTimeout) (QueryResult, error) {
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
