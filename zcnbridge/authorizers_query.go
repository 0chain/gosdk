package zcnbridge

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

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

const (
	// Represents max amount of retries, when timeout status was received
	maxCallbackRetries = 5

	// Represents callback retry delay for http calls, when timeout status was received
	callbackRetryDelay = time.Second * 5
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
func (b *BridgeClient) QueryEthereumBurnEvents(startNonce string, keys ...string) ([]*ethereum.BurnEvent, error) {
	client = h.CleanClient()
	authorizers, err := getAuthorizers(true)

	if err != nil || len(authorizers) == 0 {
		return nil, errors.Wrap("get_authorizers", "failed to get authorizers", err)
	}

	var (
		totalWorkers = len(authorizers)
		values       = map[string]string{
			"clientid":        coreClient.Id(keys...),
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
func (b *BridgeClient) QueryZChainMintPayload(ethBurnHash string, keys ...string) (*zcnsc.MintPayload, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Add exponential backoff delay
			delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			log.Logger.Info("Retrying QueryZChainMintPayload",
				zap.String("hash", ethBurnHash),
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay))
			time.Sleep(delay)
		}

		payload, err := b.queryZChainMintPayloadOnce(ethBurnHash, keys...)
		if err == nil {
			return payload, nil
		}
		lastErr = err
		log.Logger.Warn("Failed to get mint payload", zap.Error(err), zap.Int("attempt", attempt+1))
	}

	return nil, errors.Wrap("query_zchain_mint_payload", "max retries exceeded", lastErr)
}

// queryZChainMintPayloadOnce performs a single attempt to query mint payload
func (b *BridgeClient) queryZChainMintPayloadOnce(ethBurnHash string, keys ...string) (*zcnsc.MintPayload, error) {
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
			"clientid": coreClient.Id(keys...),
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
			ReceivingClientID: strings.TrimPrefix(burnTicket.ReceivingClientID, "0x"),
		}

		return payload, nil
	} else {
		return b.QueryZChainMintPayload(ethBurnHash)
	}
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

	resp := &authorizerResponse{AuthorizerID: au.ID}

	req, err := http.NewRequest("GET", ticketURL, nil)
	if err != nil {
		log.Logger.Error("failed to create request", zap.Error(err))
		resp.error = errors.Wrap("request_creation", "failed to create request", err)
		responseChannel <- resp
		return
	}

	q := req.URL.Query()
	for k, v := range request.values {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	Logger.Info(req.URL.String())

	respData, body := readResponse(func() (*http.Response, error) { return client.Do(req) })
	resp.error = respData.error

	if resp.error != nil {
		Logger.Error(
			"failed to process response",
			zap.Error(resp.error),
			zap.String("node.id", au.ID),
			zap.String("node.url", au.URL),
		)
		responseChannel <- resp
		return
	}

	// Only attempt to decode if we have a body
	if len(body) == 0 {
		resp.error = errors.New("empty_response", "empty response body received")
		Logger.Error("empty response body",
			zap.String("node.id", au.ID),
			zap.String("node.url", au.URL),
		)
		responseChannel <- resp
		return
	}

	event, errEvent := request.bodyDecoder(body)
	if errEvent != nil {
		resp.error = errors.Wrap("decode_message_body", "failed to decode message body", errEvent)
		log.Logger.Error(
			"failed to decode event body",
			zap.Error(resp.error),
			zap.String("node.id", au.ID),
			zap.String("node.url", au.URL),
			zap.String("body", string(body)),
		)
		responseChannel <- resp
		return
	}

	// Only set ID if event is not nil
	if event != nil {
		event.SetAuthorizerID(au.ID)
		resp.event = event
	} else {
		resp.error = errors.New("nil_event", "decoded event is nil")
	}

	responseChannel <- resp
}

func readResponse(responseCallback func() (*http.Response, error)) (res *authorizerResponse, body []byte) {
	res = &authorizerResponse{}

	var (
		retryTicker  *time.Ticker
		retryCounter int

		response *http.Response
		err      error
	)

	for {
		if retryTicker != nil {
			select {
			case <-retryTicker.C:
			default:
			}
		}

		response, err = responseCallback()
		if response == nil {
			res.error = err
			Logger.Error("response is empty", zap.Error(err))
			return res, nil
		}

		if err != nil {
			err = errors.Wrap("authorizer_post_process", "failed to call the authorizer", err)
			Logger.Error("request response error", zap.Error(err))
		}

		if response.StatusCode == 408 {
			if retryTicker == nil {
				retryTicker = time.NewTicker(callbackRetryDelay)
			}

			if retryCounter >= maxCallbackRetries {
				err = errors.Wrap("authorizer_post_process", fmt.Sprintf("error %d", response.StatusCode), err)
				Logger.Error("request response status", zap.Error(err))

				break
			}

			retryCounter++

			continue
		}

		if response.StatusCode >= 400 {
			err = errors.Wrap("authorizer_post_process", fmt.Sprintf("error %d", response.StatusCode), err)
			Logger.Error("request response status", zap.Error(err))

			break
		}

		break
	}

	if retryTicker != nil {
		retryTicker.Stop()
	}

	body, er := io.ReadAll(response.Body)
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
