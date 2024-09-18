package zcnbridge

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"sync"

	coreClient "github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zcnbridge/errors"
	"github.com/0chain/gosdk/zcnbridge/ethereum"
	h "github.com/0chain/gosdk/zcnbridge/http"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/0chain/gosdk/zcnbridge/zcnsc"
	"go.uber.org/zap"
)

type (
	// authorizerResponse is HTTP client response event
	authorizerResponse struct {
		// 	AuthorizerID is authorizer where the job was performed
		AuthorizerID string
		// event is server job event
		event JobResult
		// error describes an error occurred during event processing on client side during the call to server
		error
	}

	requestHandler struct {
		path        string
		values      map[string]string
		bodyDecoder func([]byte) (JobResult, error)
	}

	responseChannelType chan *authorizerResponse
	eventsChannelType   chan []JobResult
)

var (
	client *http.Client
)

// QueryEthereumMintPayload gets burn ticket and creates mint payload to be minted in the Ethereum chain
// zchainBurnHash - Ethereum burn transaction hash
func (b *BridgeClient) QueryEthereumMintPayload(zchainBurnHash string) (*ethereum.MintPayload, error) {
	client = h.CleanClient()
	authorizers, err := getAuthorizers(true)

	if err != nil || len(authorizers) == 0 {
		return nil, errors.Wrap("get_authorizers", "failed to get authorizers", err)
	}

	var (
		totalWorkers = len(authorizers)
		values       = map[string]string{
			"hash": zchainBurnHash,
		}
	)

	handler := &requestHandler{
		path:   wallet.BurnNativeTicketPath,
		values: values,
		bodyDecoder: func(body []byte) (JobResult, error) {
			ev := &ProofZCNBurn{}
			err := json.Unmarshal(body, ev)
			return ev, err
		},
	}

	thresh := b.ConsensusThreshold
	results := queryAllAuthorizers(authorizers, handler)
	numSuccess := len(results)
	quorum := math.Ceil((float64(numSuccess) * 100) / float64(totalWorkers))

	if numSuccess > 0 && quorum >= thresh {
		burnTicket, ok := results[0].(*ProofZCNBurn)
		if !ok {
			return nil, errors.Wrap("type_cast", "failed to convert to *proofEthereumBurn", err)
		}

		var sigs []*ethereum.AuthorizerSignature
		for _, result := range results {
			ticket := result.(*ProofZCNBurn)
			sig := &ethereum.AuthorizerSignature{
				ID:        ticket.GetAuthorizerID(),
				Signature: ticket.Signature,
			}
			sigs = append(sigs, sig)
		}

		payload := &ethereum.MintPayload{
			ZCNTxnID:   burnTicket.TxnID,
			Amount:     burnTicket.Amount,
			To:         burnTicket.To,
			Nonce:      burnTicket.Nonce,
			Signatures: sigs,
		}

		return payload, nil
	}

	text := fmt.Sprintf("failed to reach the quorum. #Success: %d from #Total: %d", numSuccess, totalWorkers)
	return nil, errors.New("get_burn_ticket", text)
}

// QueryEthereumBurnEvents gets ethereum burn events
func (b *BridgeClient) QueryEthereumBurnEvents(startNonce string) ([]*ethereum.BurnEvent, error) {
	client = h.CleanClient()
	authorizers, err := getAuthorizers(true)

	if err != nil || len(authorizers) == 0 {
		return nil, errors.Wrap("get_authorizers", "failed to get authorizers", err)
	}

	var (
		totalWorkers = len(authorizers)
		values       = map[string]string{
			"clientid":        coreClient.Id(),
			"ethereumaddress": b.EthereumAddress,
			"startnonce":      startNonce,
		}
	)

	handler := &requestHandler{
		path:   wallet.BurnWzcnBurnEventsPath,
		values: values,
		bodyDecoder: func(body []byte) (JobResult, error) {
			ev := &EthereumBurnEvents{}
			err := json.Unmarshal(body, ev)
			return ev, err
		},
	}

	thresh := b.ConsensusThreshold
	results := queryAllAuthorizers(authorizers, handler)
	numSuccess := len(results)
	quorum := math.Ceil((float64(numSuccess) * 100) / float64(totalWorkers))

	if numSuccess > 0 && quorum >= thresh {
		burnEvents, ok := results[0].(*EthereumBurnEvents)
		if !ok {
			return nil, errors.Wrap("type_cast", "failed to convert to *ethereumBurnEvents", err)
		}

		result := make([]*ethereum.BurnEvent, 0)

		for _, burnEvent := range burnEvents.BurnEvents {
			result = append(result, &ethereum.BurnEvent{
				Nonce:           burnEvent.Nonce,
				Amount:          burnEvent.Amount,
				TransactionHash: burnEvent.TransactionHash,
			})
		}

		return result, nil
	}

	text := fmt.Sprintf("failed to reach the quorum. #Success: %d from #Total: %d", numSuccess, totalWorkers)
	return nil, errors.New("get_burn_events", text)
}

// QueryZChainMintPayload gets burn ticket and creates mint payload to be minted in the ZChain
// ethBurnHash - Ethereum burn transaction hash
func (b *BridgeClient) QueryZChainMintPayload(ethBurnHash string) (*zcnsc.MintPayload, error) {
	client = h.CleanClient()
	authorizers, err := getAuthorizers(true)
	log.Logger.Info("Got authorizers", zap.Int("amount", len(authorizers)))

	if err != nil || len(authorizers) == 0 {
		return nil, errors.Wrap("get_authorizers", "failed to get authorizers", err)
	}

	var (
		totalWorkers = len(authorizers)
		values       = map[string]string{
			"hash":     ethBurnHash,
			"clientid": coreClient.Id(),
		}
	)

	handler := &requestHandler{
		path:   wallet.BurnWzcnTicketPath,
		values: values,
		bodyDecoder: func(body []byte) (JobResult, error) {
			ev := &WZCNBurnEvent{}
			err := json.Unmarshal(body, ev)
			return ev, err
		},
	}

	thresh := b.ConsensusThreshold
	results := queryAllAuthorizers(authorizers, handler)
	numSuccess := len(results)
	quorum := math.Ceil((float64(numSuccess) * 100) / float64(totalWorkers))

	if numSuccess > 0 && quorum >= thresh {
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
			Amount:            common.Balance(burnTicket.Amount),
			Nonce:             burnTicket.Nonce,
			Signatures:        sigs,
			ReceivingClientID: burnTicket.ReceivingClientID,
		}

		return payload, nil
	}

	text := fmt.Sprintf("failed to reach the quorum. #Success: %d from #Total: %d", numSuccess, totalWorkers)
	return nil, errors.New("get_burn_ticket", text)
}

func queryAllAuthorizers(authorizers []*AuthorizerNode, handler *requestHandler) []JobResult {
	var (
		totalWorkers    = len(authorizers)
		eventsChannel   = make(eventsChannelType)
		responseChannel = make(responseChannelType, totalWorkers)
	)
	defer close(eventsChannel)

	var wg sync.WaitGroup

	for _, authorizer := range authorizers {
		wg.Add(1)
		go queryAuthorizer(authorizer, handler, responseChannel)
	}

	go handleResponse(responseChannel, eventsChannel, &wg)

	wg.Wait()
	close(responseChannel)
	results := <-eventsChannel

	return results
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

func queryAuthorizer(au *AuthorizerNode, request *requestHandler, responseChannel responseChannelType) {
	Logger.Info("Query from authorizer", zap.String("ID", au.ID), zap.String("URL", au.URL))
	ticketURL := strings.TrimSuffix(au.URL, "/") + request.path

	req, err := http.NewRequest("GET", ticketURL, nil)
	if err != nil {
		log.Logger.Error("failed to create request", zap.Error(err))
		return
	}

	q := req.URL.Query()
	for k, v := range request.values {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	Logger.Info(req.URL.String())
	resp, body := readResponse(client.Do(req))
	resp.AuthorizerID = au.ID

	if resp.error != nil {
		Logger.Error(
			"failed to process response",
			zap.Error(resp.error),
			zap.String("node.id", au.ID),
			zap.String("node.url", au.URL),
		)
	}

	event, errEvent := request.bodyDecoder(body)
	event.SetAuthorizerID(au.ID)

	if errEvent != nil {
		err := errors.Wrap("decode_message_body", "failed to decode message body", errEvent)
		log.Logger.Error(
			"failed to decode event body",
			zap.Error(err),
			zap.String("node.id", au.ID),
			zap.String("node.url", au.URL),
			zap.String("body", string(body)),
		)
	}

	resp.event = event

	responseChannel <- resp
}

func readResponse(response *http.Response, err error) (res *authorizerResponse, body []byte) {
	res = &authorizerResponse{}
	if err != nil {
		err = errors.Wrap("authorizer_post_process", "failed to call the authorizer", err)
		Logger.Error("request response error", zap.Error(err))
	}

	if response == nil {
		res.error = err
		Logger.Error("response is empty", zap.Error(err))
		return res, nil
	}

	if response.StatusCode >= 400 {
		err = errors.Wrap("authorizer_post_process", fmt.Sprintf("error %d", response.StatusCode), err)
		Logger.Error("request response status", zap.Error(err))
	}

	body, er := ioutil.ReadAll(response.Body)
	log.Logger.Debug("response", zap.String("response", string(body)))
	defer response.Body.Close()

	if er != nil || len(body) == 0 {
		var errstrings []string
		er = errors.Wrap("authorizer_post_process", "failed to read body", er)
		if err != nil {
			errstrings = append(errstrings, err.Error())
		}
		errstrings = append(errstrings, er.Error())
		err = fmt.Errorf(strings.Join(errstrings, ":"))
	}

	res.error = err

	return res, body
}
