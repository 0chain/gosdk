package zcnbridge

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	bridge "github.com/0chain/gosdk/zcnbridge/http"
	"github.com/0chain/gosdk/zmagmacore/errors"
)

const (
	Retrying        = 5
	RequestDuration = time.Second * 120
	PollInterval    = time.Second * 5
)

type (
	// jobResult is HTTP client response result
	jobResult struct {
		// result is server job result
		result *authorizerBurnEvent
		// error describes an error occurred during result processing on client side
		error
	}

	responseChannelType chan *jobResult
	resultsChannelType  chan []*authorizerBurnEvent
)

var (
	client *retryablehttp.Client
)

func CreateMintPayload(ctx context.Context, hash, address, clientID string, requiredQuorum int) (*MintPayload, error) {
	client = bridge.NewRetryableClient(Retrying)
	authorizers, err := GetAuthorizers()

	if err != nil || len(authorizers.NodeMap) == 0 {
		return nil, errors.Wrap("get_authorizers", "failed to get authorizers", err)
	}

	var (
		totalWorkers    = len(authorizers.NodeMap)
		resultsChannel  = make(resultsChannelType)
		responseChannel = make(responseChannelType, totalWorkers)
		values          = url.Values{
			"hash":     []string{hash},
			"address":  []string{address},
			"clientid": []string{clientID},
		}
	)

	var wg sync.WaitGroup

	go handleResults(responseChannel, resultsChannel, &wg)
	defer close(resultsChannel)

	for _, node := range authorizers.NodeMap {
		wg.Add(1)
		go getResultFromAuthoriser(ctx, node, values, responseChannel)
	}

	wg.Wait()
	close(responseChannel)
	results := <-resultsChannel

	numSuccess := len(results)

	quorum := int(math.Ceil((float64(numSuccess) * 100) / float64(totalWorkers)))

	if numSuccess > 0 && quorum >= requiredQuorum && len(resultsChannel) > 1 {
		burnTicket := results[0].BurnTicket

		var sigs []*AuthorizerSignature
		for _, result := range results {
			sig := &AuthorizerSignature{
				ID:        result.AuthorizerID,
				Signature: result.BurnTicket.Signature,
			}
			sigs = append(sigs, sig)
		}

		payload := &MintPayload{
			EthereumTxnID:     burnTicket.TxnID,
			Amount:            burnTicket.Amount,
			Nonce:             burnTicket.Amount,
			Signatures:        sigs,
			ReceivingClientID: burnTicket.ReceivingClientID,
		}

		return payload, nil
	}

	text := fmt.Sprintf("failed to reach the quorum. #Success: %d from #Total: %d", numSuccess, totalWorkers)
	return nil, errors.New("get_burn_ticket", text)
}

func handleResults(responseChannel responseChannelType, resultsChannel resultsChannelType, wg *sync.WaitGroup) {
	var results []*authorizerBurnEvent
	for result := range responseChannel {
		if result.error == nil {
			results = append(results, result.result)
		}
		wg.Done()
	}
	resultsChannel <- results
}

func getResultFromAuthoriser(ctx context.Context, node *AuthorizerNode, values url.Values, responseChannel responseChannelType) {
	currContext, cancel := context.WithTimeout(ctx, RequestDuration)
	defer cancel()

	// TODO: Specify error in each branch

	for {
		select {
		case <-currContext.Done():
			return
		default:
			response, err := client.PostForm(node.URL, values)
			if res, ok := processResponse(ctx, response, err); ok {
				res.result.AuthorizerID = node.ID
				responseChannel <- res
				return
			}
			time.Sleep(PollInterval)
		}
	}
}

func processResponse(ctx context.Context, response *http.Response, err error) (*jobResult, bool) {
	if err != nil {
		err = errors.Wrap("authorizer_post_request", "failed to call the authorizer", err)
	}

	if response.StatusCode >= 400 {
		err = errors.Wrap("authorizer_post_request", fmt.Sprintf("error %d", response.StatusCode), err)
	}

	res := &jobResult{
		error: err,
	}

	// TODO: Unmarshal body

	return res, err == nil
}
