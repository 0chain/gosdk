package zcnbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/0chain/gosdk/zcnbridge/errors"
	bridge "github.com/0chain/gosdk/zcnbridge/http"
	"github.com/0chain/gosdk/zcnbridge/node"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	Retrying        = 5
	RequestDuration = time.Second * 120
	PollInterval    = time.Second * 5
)

type (
	// jobResult is HTTP client response burnEvent
	jobResult struct {
		// burnEvent is server job burnEvent
		burnEvent *AuthorizerBurnEvent
		// error describes an error occurred during burnEvent processing on client side
		error
	}

	jobResultChannelType  chan *jobResult
	burnEventsChannelType chan []*AuthorizerBurnEvent
)

var (
	client *retryablehttp.Client
)

// CreateMintPayload gets burn ticket and creates mint payload to be minted in the chain
func CreateMintPayload(ctx context.Context, hash string) (*MintPayload, error) {
	client = bridge.NewRetryableClient(Retrying)
	authorizers, err := GetAuthorizers()

	if err != nil || len(authorizers.NodeMap) == 0 {
		return nil, errors.Wrap("get_authorizers", "failed to get authorizers", err)
	}

	var (
		totalWorkers      = len(authorizers.NodeMap)
		burnEventsChannel = make(burnEventsChannelType)
		jobsChannel       = make(jobResultChannelType, totalWorkers)
		values            = url.Values{
			"hash":     []string{hash},
			"address":  []string{wallet.ZCNSCSmartContractAddress},
			"clientid": []string{node.ID()},
		}
	)

	var wg sync.WaitGroup

	go handleResults(jobsChannel, burnEventsChannel, &wg)
	defer close(burnEventsChannel)

	for _, authorizer := range authorizers.NodeMap {
		wg.Add(1)
		go getResultFromAuthoriser(ctx, authorizer, values, jobsChannel)
	}

	wg.Wait()
	close(jobsChannel)
	results := <-burnEventsChannel

	numSuccess := len(results)

	quorum := math.Ceil((float64(numSuccess) * 100) / float64(totalWorkers))

	if numSuccess > 0 && quorum >= wallet.ConsensusThresh && len(burnEventsChannel) > 1 {
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

func handleResults(jobResults jobResultChannelType, burnEvents burnEventsChannelType, wg *sync.WaitGroup) {
	var events []*AuthorizerBurnEvent
	for result := range jobResults {
		if result.error == nil {
			events = append(events, result.burnEvent)
		}
		wg.Done()
	}
	burnEvents <- events
}

func getResultFromAuthoriser(ctx context.Context, node *AuthorizerNode, values url.Values, responseChannel jobResultChannelType) {
	currContext, cancel := context.WithTimeout(ctx, RequestDuration)
	defer cancel()

	var (
		job               *jobResult
		response          *http.Response
		lastError         error
		ok                bool
		burnWZCNTicketURL = strings.TrimSuffix(node.URL, "/") + wallet.BurnWzcnTicketPath
	)

	for {
		select {
		case <-currContext.Done():
			job, _ = processResponse(nil, currContext.Err())
			job.burnEvent.AuthorizerID = node.ID
			responseChannel <- job
			return
		default:
			response, lastError = client.PostForm(burnWZCNTicketURL, values)
			if job, ok = processResponse(response, lastError); ok {
				job.burnEvent.AuthorizerID = node.ID
				responseChannel <- job
				return
			}
			time.Sleep(PollInterval)
		}
	}
}

func processResponse(response *http.Response, err error) (*jobResult, bool) {
	var (
		res = &jobResult{}
		ev  = &AuthorizerBurnEvent{}
	)

	if err != nil {
		err = errors.Wrap("authorizer_post_process", "failed to call the authorizer", err)
	}

	if response == nil {
		res.error = err
		return res, false
	}

	if response.StatusCode >= 400 {
		err = errors.Wrap("authorizer_post_process", fmt.Sprintf("error %d", response.StatusCode), err)
	}

	body, e := ioutil.ReadAll(response.Body)
	if e != nil || len(body) == 0 {
		res.error = errors.Wrap("authorizer_post_process", "failed to read body", e)
		return res, false
	}

	e = json.Unmarshal(body, ev)
	if e != nil {
		res.error = errors.Wrap("decode_message_body", "failed to decode message body", e)
		return res, false
	}

	res.error = err
	res.burnEvent = ev

	return res, err == nil
}
