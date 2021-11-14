package zcnbridge

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	u "net/url"
	"strings"
	"sync"

	"github.com/0chain/gosdk/zcnbridge/errors"
	bridge "github.com/0chain/gosdk/zcnbridge/http"
	"github.com/0chain/gosdk/zcnbridge/node"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/hashicorp/go-retryablehttp"
)

type (
	// authorizerResponse is HTTP client response event
	authorizerResponse struct {
		// 	AuthorizerID is authorizer where the job was performed
		AuthorizerID string
		// event is server job event
		event JobResult
		// error describes an error occurred during event processing on client side
		error
	}

	responseChannelType chan *authorizerResponse
	eventsChannelType   chan []JobResult
)

var (
	client *retryablehttp.Client
)

// CreateMintPayload gets burn ticket and creates mint payload to be minted in the chain
func CreateMintPayload(hash string) (*MintPayload, error) {
	client = bridge.NewRetryableClient()
	authorizers, err := GetAuthorizers()

	if err != nil || len(authorizers.NodeMap) == 0 {
		return nil, errors.Wrap("get_authorizers", "failed to get authorizers", err)
	}

	var (
		totalWorkers = len(authorizers.NodeMap)
		values       = u.Values{
			"hash":     []string{hash},
			"address":  []string{wallet.ZCNSCSmartContractAddress},
			"clientid": []string{node.ID()},
		}
	)

	results := queryAllAuthorizers(authorizers, wallet.BurnWzcnTicketPath, values)

	numSuccess := len(results)

	quorum := math.Ceil((float64(numSuccess) * 100) / float64(totalWorkers))

	if numSuccess > 0 && quorum >= wallet.ConsensusThresh && len(results) > 1 {
		burnTicket, ok := results[0].Data().(*proofEthereumBurn)
		if !ok {
			return nil, errors.Wrap("type_cast", "failed to convert to *proofEthereumBurn", err)
		}

		var sigs []*AuthorizerSignature
		for _, result := range results {
			ticket := result.Data().(*proofEthereumBurn)
			sig := &AuthorizerSignature{
				ID:        result.GetAuthorizerID(),
				Signature: ticket.Signature,
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

func queryAllAuthorizers(authorizers *AuthorizerNodes, path string, values u.Values) []JobResult {
	var (
		totalWorkers    = len(authorizers.NodeMap)
		eventsChannel   = make(eventsChannelType)
		responseChannel = make(responseChannelType, totalWorkers)
	)

	var wg sync.WaitGroup

	go handleResponse(responseChannel, eventsChannel, &wg)
	defer close(eventsChannel)

	for _, authorizer := range authorizers.NodeMap {
		wg.Add(1)
		go queryAuthoriser(authorizer, path, values, responseChannel)
	}

	wg.Wait()
	close(responseChannel)
	return <-eventsChannel
}

func handleResponse(responseChannel responseChannelType, eventsChannel eventsChannelType, wg *sync.WaitGroup) {
	var events []JobResult
	for job := range responseChannel {
		if job.error == nil {
			event := job.event
			event.SetAuthorizerID(job.AuthorizerID)
			events = append(events, event)
		}
		wg.Done()
	}
	eventsChannel <- events
}

func queryAuthoriser(node *AuthorizerNode, path string, values u.Values, responseChannel responseChannelType) {
	var (
		ticketURL = strings.TrimSuffix(node.URL, "/") + path
	)

	if job, ok := processResponse(client.PostForm(ticketURL, values)); ok {
		job.AuthorizerID = node.ID
		responseChannel <- job
	}
}

func processResponse(response *http.Response, err error) (*authorizerResponse, bool) {
	var (
		res = &authorizerResponse{}
		ev  = &WZCNBurnEvent{}
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
	res.event = ev

	return res, err == nil
}
