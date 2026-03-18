package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/0chain/gosdk/zcnbridge/errors"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcncore"
	"strconv"
)


// burnZCN Burns ZCN tokens and returns a hash of the burn transaction
//   - amount: amount of ZCN tokens to burn
//   - txnfee: transaction fee
func burnZCNMW(amount uint64, key string) string { //nolint
	if bridge == nil {
		return errors.New("burnZCN", "bridge is not initialized").Error()
	}

	hash, _, err := bridge.BurnZCN(amount, key)
	if err != nil {
		return errors.Wrap("burnZCN", "failed to burn ZCN tokens", err).Error()
	}

	return hash
}

// mintZCN Mints ZCN tokens and returns a hash of the mint transaction
//   - burnTrxHash: hash of the burn transaction
//   - timeout: timeout in seconds
func mintZCNMW(burnTrxHash string, timeout int, key string) string { //nolint
	mintPayload,

		err := bridge.QueryZChainMintPayload(burnTrxHash, key)
	if err != nil {
		return errors.Wrap("mintZCN", "failed to QueryZChainMintPayload", err).Error()
	}

	hash, err := bridge.MintZCN(mintPayload, key)
	if err != nil {
		return errors.Wrap("mintZCN", "failed to MintZCN for txn "+hash, err).Error()
	}

	return hash
}

// getMintWZCNPayload returns the mint payload for the given burn transaction hash
//   - burnTrxHash: hash of the burn transaction
func getMintWZCNPayloadMW(burnTrxHash string, key string) string { //nolint:unused
	mintPayload, err := bridge.QueryEthereumMintPayload(burnTrxHash)
	if err != nil {
		return errors.Wrap("getMintWZCNPayload", "failed to query ethereum mint payload", err).Error()
	}
	var result []byte
	result, err = json.Marshal(mintPayload)
	if err != nil {
		return errors.Wrap("getMintWZCNPayload", "failed to query ethereum mint payload", err).Error()
	}

	return string(result)
}

// getNotProcessedWZCNBurnEvents returns all not processed WZCN burn events from the Ethereum network
func getNotProcessedWZCNBurnEventsMW(key string) string { //nolint:unused
	var (
		mintNonce int64
		res       []byte
		err       error
	)
	if res, err = zcncore.GetMintNonce(); err != nil {
		return errors.Wrap("getNotProcessedWZCNBurnEvents", "failed to retreive last ZCN processed mint nonce", err).Error()
	}

	if err = json.Unmarshal(res, &mintNonce); err != nil {
		return errors.Wrap("getNotProcessedWZCNBurnEvents", "failed to unmarshal last ZCN processed mint nonce", err).Error()
	}

	log.Logger.Debug("MintNonce = " + strconv.Itoa(int(mintNonce)))
	burnEvents, err := bridge.QueryEthereumBurnEvents(strconv.Itoa(int(mintNonce)), key)
	if err != nil {
		return errors.Wrap("getNotProcessedWZCNBurnEvents", "failed to retrieve WZCN burn events", err).Error()
	}

	var result []byte
	result, err = json.Marshal(burnEvents)
	if err != nil {
		return errors.Wrap("getNotProcessedWZCNBurnEvents", "failed to marshal WZCN burn events", err).Error()
	}

	return string(result)
}

// getNotProcessedZCNBurnTickets Returns all not processed ZCN burn tickets burned for a certain ethereum address
func getNotProcessedZCNBurnTicketsMW(key string) string { //nolint:unused
	userNonce, err := bridge.GetUserNonceMinted(context.Background(), bridge.EthereumAddress)
	if err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to retreive user nonce", err).Error()
	}

	var (
		res         []byte
		burnTickets []zcncore.BurnTicket
	)

	res, err = zcncore.GetNotProcessedZCNBurnTickets(bridge.EthereumAddress, userNonce.String())
	if err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to retreive ZCN burn tickets", err).Error()
	}

	if err = json.Unmarshal(res, &burnTickets); err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to unmarshal ZCN burn tickets", err).Error()
	}

	var result []byte
	result, err = json.Marshal(burnTickets)
	if err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to marshal ZCN burn tickets", err).Error()
	}

	return string(result)
}

// estimateBurnWZCNGasAmount performs gas amount estimation for the given burn wzcn transaction.
//   - from: address of the sender
//   - to: address of the receiver
//   - amountTokens: amount of tokens to burn (as a string)
func estimateBurnWZCNGasAmountMW(from, to, amountTokens, key string) string { // nolint:golint,unused
	estimateBurnWZCNGasAmountResponse, err := bridge.EstimateBurnWZCNGasAmount(
		context.Background(), from, to, amountTokens, key)
	if err != nil {
		return errors.Wrap("estimateBurnWZCNGasAmount", "failed to estimate gas amount", err).Error()
	}

	var result []byte
	result, err = json.Marshal(estimateBurnWZCNGasAmountResponse)
	if err != nil {
		return errors.Wrap("estimateBurnWZCNGasAmount", "failed to marshal gas amount estimation result", err).Error()
	}

	return string(result)
}

// estimateMintWZCNGasAmount performs gas amount estimation for the given mint wzcn transaction.
//   - from: address of the sender
//   - to: address of the receiver
//   - zcnTransaction: hash of the ZCN transaction
//   - amountToken: amount of tokens to mint (as a string)
//   - nonce: nonce of the transaction
//   - signaturesRaw: encoded format (base-64) of the burn signatures received from the authorizers.
func estimateMintWZCNGasAmountMW(from, to, zcnTransaction, amountToken string, nonce int64, signaturesRaw []string, key string) string { // nolint:golint,unused
	var signaturesBytes [][]byte

	var (
		signatureBytes []byte
		err            error
	)

	for _, signature := range signaturesRaw {
		signatureBytes, err = base64.StdEncoding.DecodeString(signature)
		if err != nil {
			return errors.Wrap("estimateMintWZCNGasAmount", "failed to convert raw signature into bytes", err).Error()
		}

		signaturesBytes = append(signaturesBytes, signatureBytes)
	}

	estimateMintWZCNGasAmountResponse, err := bridge.EstimateMintWZCNGasAmount(
		context.Background(), from, to, zcnTransaction, amountToken, nonce, signaturesBytes)
	if err != nil {
		return errors.Wrap("estimateMintWZCNGasAmount", "failed to estimate gas amount", err).Error()
	}

	var result []byte
	result, err = json.Marshal(estimateMintWZCNGasAmountResponse)
	if err != nil {
		return errors.Wrap("estimateMintWZCNGasAmount", "failed to marshal gas amount estimation result", err).Error()
	}

	return string(result)
}

// estimateGasPrice performs gas estimation for the given transaction using Alchemy enhanced API returning
// approximate final gas fee.
func estimateGasPriceMW(key string) string { // nolint:golint,unused
	estimateGasPriceResponse, err := bridge.EstimateGasPrice(context.Background())
	if err != nil {
		return errors.Wrap("estimateGasPrice", "failed to estimate gas price", err).Error()
	}

	var result []byte
	result, err = json.Marshal(estimateGasPriceResponse)
	if err != nil {
		return errors.Wrap("estimateGasPrice", "failed to marshal gas price estimation result", err).Error()
	}

	return string(result)
}
