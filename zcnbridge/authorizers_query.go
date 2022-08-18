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
		values      u.Values
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
	authorizers, err := getAuthorizers()

	if err != nil || len(authorizers) == 0 {
		return nil, errors.Wrap("get_authorizers", "failed to get authorizers", err)
	}

	var (
		totalWorkers = len(authorizers)
		values       = u.Values{
			"hash": []string{zchainBurnHash},
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

	if numSuccess > 0 && quorum >= thresh && len(results) > 1 {
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
			Nonce:      burnTicket.Nonce,
			Signatures: sigs,
		}

		return payload, nil
	}

	text := fmt.Sprintf("failed to reach the quorum. #Success: %d from #Total: %d", numSuccess, totalWorkers)
	return nil, errors.New("get_burn_ticket", text)
}

// QueryZChainMintPayload gets burn ticket and creates mint payload to be minted in the ZChain
// ethBurnHash - Ethereum burn transaction hash
func (b *BridgeClient) QueryZChainMintPayload(ethBurnHash string) (*zcnsc.MintPayload, error) {
	client = h.CleanClient()
	authorizers, err := getAuthorizers()

	if err != nil || len(authorizers) == 0 {
		return nil, errors.Wrap("get_authorizers", "failed to get authorizers", err)
	}

	var (
		totalWorkers = len(authorizers)
		values       = u.Values{
			"hash":     []string{ethBurnHash},
			"address":  []string{wallet.ZCNSCSmartContractAddress},
			"clientid": []string{b.ClientID()},
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

	if numSuccess > 0 && quorum >= thresh && len(results) > 1 {
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
			Nonce:             burnTicket.Amount,
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
	log.Logger.Info("Query from authorizer", zap.String("ID", au.ID), zap.String("URL", au.URL))
	ticketURL := strings.TrimSuffix("https://dev.0chain.net/authorizer01/", "/") + request.path
	resp, body := readResponse(client.PostForm(ticketURL, request.values))
	resp.AuthorizerID = au.ID

	if resp.error != nil {
		log.Logger.Error(
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
