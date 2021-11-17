package authorizer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	u "net/url"
	"strings"
	"sync"

	"github.com/0chain/gosdk/zcnbridge/ethereum"

	"github.com/0chain/gosdk/zcnbridge/zcnsc"

	"github.com/0chain/gosdk/zcnbridge/log"
	"go.uber.org/zap"

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

	requestHandler struct {
		path    string
		values  u.Values
		decoder func([]byte) (JobResult, error)
	}

	responseChannelType chan *authorizerResponse
	eventsChannelType   chan []JobResult
)

var (
	client *retryablehttp.Client
)

func CreateWZCNMintPayload(hash string) (*ethereum.MintPayload, error) {
	client = bridge.NewRetryableClient()
	authorizers, err := GetAuthorizers()

	if err != nil || len(authorizers.NodeMap) == 0 {
		return nil, errors.Wrap("get_authorizers", "failed to get authorizers", err)
	}

	var (
		totalWorkers = len(authorizers.NodeMap)
		values       = u.Values{
			"hash": []string{hash},
		}
	)

	handler := &requestHandler{
		path:   wallet.BurnNativeTicketPath,
		values: values,
		decoder: func(body []byte) (JobResult, error) {
			ev := &ZCNBurnEvent{}
			err := json.Unmarshal(body, ev)
			return ev, err
		},
	}

	results := queryAllAuthorizers(authorizers, handler)
	numSuccess := len(results)
	quorum := math.Ceil((float64(numSuccess) * 100) / float64(totalWorkers))

	if numSuccess > 0 && quorum >= wallet.ConsensusThresh && len(results) > 1 {
		burnTicket, ok := results[0].Data().(*proofZCNBurn)
		if !ok {
			return nil, errors.Wrap("type_cast", "failed to convert to *proofEthereumBurn", err)
		}

		var sigs []*ethereum.AuthorizerSignature
		for _, result := range results {
			ticket := result.Data().(*proofZCNBurn)
			sig := &ethereum.AuthorizerSignature{
				ID:        result.GetAuthorizerID(),
				Signature: ticket.Signature,
			}
			sigs = append(sigs, sig)
		}

		payload := &ethereum.MintPayload{
			ZCNTxnID:   burnTicket.TxnID,
			Amount:     burnTicket.Amount,
			Nonce:      burnTicket.Nonce,
			Signatures: sigs,
		}

		return payload, nil
	}

	text := fmt.Sprintf("failed to reach the quorum. #Success: %d from #Total: %d", numSuccess, totalWorkers)
	return nil, errors.New("get_burn_ticket", text)
}

// CreateZCNMintPayload gets burn ticket and creates mint payload to be minted in the chain
func CreateZCNMintPayload(hash string) (*zcnsc.MintPayload, error) {
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

	handler := &requestHandler{
		path:   wallet.BurnWzcnTicketPath,
		values: values,
		decoder: func(body []byte) (JobResult, error) {
			ev := &WZCNBurnEvent{}
			err := json.Unmarshal(body, ev)
			return ev, err
		},
	}

	results := queryAllAuthorizers(authorizers, handler)
	numSuccess := len(results)
	quorum := math.Ceil((float64(numSuccess) * 100) / float64(totalWorkers))

	if numSuccess > 0 && quorum >= wallet.ConsensusThresh && len(results) > 1 {
		burnTicket, ok := results[0].Data().(*ProofEthereumBurn)
		if !ok {
			return nil, errors.Wrap("type_cast", "failed to convert to *proofEthereumBurn", err)
		}

		var sigs []*zcnsc.AuthorizerSignature
		for _, result := range results {
			ticket := result.Data().(*ProofEthereumBurn)
			sig := &zcnsc.AuthorizerSignature{
				ID:        result.GetAuthorizerID(),
				Signature: ticket.Signature,
			}
			sigs = append(sigs, sig)
		}

		payload := &zcnsc.MintPayload{
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

func queryAllAuthorizers(authorizers *Nodes, handler *requestHandler) []JobResult {
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
		go queryAuthoriser(authorizer, handler, responseChannel)
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

func queryAuthoriser(node *Node, request *requestHandler, responseChannel responseChannelType) {
	var (
		ticketURL = strings.TrimSuffix(node.URL, "/") + request.path
	)

	job, body := processResponse(client.PostForm(ticketURL, request.values))
	if job.error != nil {
		log.Logger.Error(
			"failed to process response",
			zap.Error(job.error),
			zap.String("node.id", node.ID),
			zap.String("node.url", node.URL),
		)
		return
	}

	event, err := request.decoder(body)
	if err != nil {
		err := errors.Wrap("decode_message_body", "failed to decode message body", err)
		log.Logger.Error(
			"failed to decode event body",
			zap.Error(err),
			zap.String("node.id", node.ID),
			zap.String("node.url", node.URL),
		)
		return
	}

	job.AuthorizerID = node.ID
	job.event = event
	responseChannel <- job
}

func processResponse(response *http.Response, err error) (*authorizerResponse, []byte) {
	var (
		res = &authorizerResponse{}
	)

	if err != nil {
		err = errors.Wrap("authorizer_post_process", "failed to call the authorizer", err)
	}

	if response == nil {
		res.error = err
		return res, nil
	}

	if response.StatusCode >= 400 {
		err = errors.Wrap("authorizer_post_process", fmt.Sprintf("error %d", response.StatusCode), err)
	}

	body, er := ioutil.ReadAll(response.Body)
	if er != nil || len(body) == 0 {
		var errstrings []string
		er = errors.Wrap("authorizer_post_process", "failed to read body", er)
		errstrings = append(errstrings, err.Error())
		errstrings = append(errstrings, er.Error())
		err = fmt.Errorf(strings.Join(errstrings, "\n"))
	}

	res.error = err

	return res, body
}
