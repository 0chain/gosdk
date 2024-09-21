package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/zcnbridge"
	"github.com/0chain/gosdk/zcnbridge/errors"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcncore"
	"github.com/ethereum/go-ethereum/ethclient"
	"path"
	"strconv"
)

var bridge *zcnbridge.BridgeClient

// initBridge initializes the bridge client
//   - ethereumAddress: ethereum address of the wallet owner
//   - bridgeAddress: address of the bridge contract on the Ethereum network
//   - authorizersAddress: address of the authorizers contract on the Ethereum network
//   - tokenAddress: address of the token contract on the Ethereum network
//   - ethereumNodeURL: URL of the Ethereum node
//   - gasLimit: gas limit for the transactions
//   - value: value to be sent with the transaction (unused)
//   - consensusThreshold: consensus threshold for the transactions
func initBridge(
	ethereumAddress string,
	bridgeAddress string,
	authorizersAddress string,
	tokenAddress string,
	ethereumNodeURL string,
	gasLimit uint64,
	value int64,
	consensusThreshold float64) error {
	if len(client.Wallet().ClientID) == 0 {
		return errors.New("wallet_error", "wallet is not set")
	}

	ethereumClient, err := ethclient.Dial(ethereumNodeURL)
	if err != nil {
		return errors.New("wallet_error", err.Error())
	}

	keyStore := zcnbridge.NewKeyStore(
		path.Join(".", zcnbridge.EthereumWalletStorageDir))

	bridge = zcnbridge.NewBridgeClient(
		bridgeAddress,
		tokenAddress,
		authorizersAddress,
		"",
		ethereumAddress,
		ethereumNodeURL,
		"",
		gasLimit,
		consensusThreshold,
		ethereumClient,
		keyStore,
	)

	return nil
}

// burnZCN Burns ZCN tokens and returns a hash of the burn transaction
//   - amount: amount of ZCN tokens to burn
//   - txnfee: transaction fee
func burnZCN(amount uint64) string { //nolint
	if bridge == nil {
		return errors.New("burnZCN", "bridge is not initialized").Error()
	}

	hash, _, err := bridge.BurnZCN(amount)
	if err != nil {
		return errors.Wrap("burnZCN", "failed to burn ZCN tokens", err).Error()
	}

	return hash
}

// mintZCN Mints ZCN tokens and returns a hash of the mint transaction
//   - burnTrxHash: hash of the burn transaction
//   - timeout: timeout in seconds
func mintZCN(burnTrxHash string, timeout int) string { //nolint
	mintPayload,

		err := bridge.QueryZChainMintPayload(burnTrxHash)
	if err != nil {
		return errors.Wrap("mintZCN", "failed to QueryZChainMintPayload", err).Error()
	}

	hash, err := bridge.MintZCN(mintPayload)
	if err != nil {
		return errors.Wrap("mintZCN", "failed to MintZCN for txn "+hash, err).Error()
	}

	return hash
}

// getMintWZCNPayload returns the mint payload for the given burn transaction hash
//   - burnTrxHash: hash of the burn transaction
func getMintWZCNPayload(burnTrxHash string) string {
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
func getNotProcessedWZCNBurnEvents() string {
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
	burnEvents, err := bridge.QueryEthereumBurnEvents(strconv.Itoa(int(mintNonce)))
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
func getNotProcessedZCNBurnTickets() string {
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
func estimateBurnWZCNGasAmount(from, to, amountTokens string) string { // nolint:golint,unused
	estimateBurnWZCNGasAmountResponse, err := bridge.EstimateBurnWZCNGasAmount(
		context.Background(), from, to, amountTokens)
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
func estimateMintWZCNGasAmount(from, to, zcnTransaction, amountToken string, nonce int64, signaturesRaw []string) string { // nolint:golint,unused
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
func estimateGasPrice() string { // nolint:golint,unused
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
