package main

import (
	"context"
	"encoding/json"
	"path"
	"strconv"
	"time"

	"github.com/0chain/gosdk/zcnbridge"
	"github.com/0chain/gosdk/zcnbridge/errors"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcnbridge/transaction"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/0chain/gosdk/zcncore"
	"github.com/ethereum/go-ethereum/ethclient"
)

var bridge *zcnbridge.BridgeClient

func initBridge(
	ethereumAddress string,
	bridgeAddress string,
	authorizersAddress string,
	tokenAddress string,
	ethereumNodeURL string,
	gasLimit uint64,
	value int64,
	consensusThreshold float64) error {
	if len(zcncore.GetWalletRaw().ClientID) == 0 {
		return errors.New("wallet_error", "wallet is not set")
	}

	ethereumClient, err := ethclient.Dial(ethereumNodeURL)
	if err != nil {
		return errors.New("wallet_error", err.Error())
	}

	transactionProvider := transaction.NewTransactionProvider()

	keyStore := zcnbridge.NewKeyStore(
		path.Join(".", zcnbridge.EthereumWalletStorageDir))

	bridge = zcnbridge.NewBridgeClient(
		bridgeAddress,
		tokenAddress,
		authorizersAddress,
		ethereumAddress,
		ethereumNodeURL,
		"",
		gasLimit,
		consensusThreshold,
		zcnbridge.BancorAPIURL,
		ethereumClient,
		transactionProvider,
		keyStore,
	)

	return nil
}

// Burns ZCN tokens and returns a hash of the burn transaction
func burnZCN(amount, txnfee uint64) string { //nolint
	if bridge == nil {
		return errors.New("burnZCN", "bridge is not initialized").Error()
	}

	tx, err := bridge.BurnZCN(context.Background(), amount, txnfee)
	if err != nil {
		return errors.Wrap("burnZCN", "failed to burn ZCN tokens", err).Error()
	}

	return tx.GetHash()
}

// Mints ZCN tokens and returns a hash of the mint transaction
func mintZCN(burnTrxHash string, timeout int) string { //nolint
	mintPayload, err := bridge.QueryZChainMintPayload(burnTrxHash)
	if err != nil {
		return errors.Wrap("mintZCN", "failed to QueryZChainMintPayload", err).Error()
	}

	c, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	hash, err := bridge.MintZCN(c, mintPayload)
	if err != nil {
		return errors.Wrap("mintZCN", "failed to MintZCN for txn "+hash, err).Error()
	}

	return hash
}

// Returns a payload used to perform minting of WZCN tokens
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

// Returns all not processed WZCN burn events for the given client id param
func getNotProcessedWZCNBurnEvents() string {
	var mintNonce int64
	cb := wallet.NewZCNStatus(&mintNonce)

	cb.Begin()

	if err := zcncore.GetMintNonce(cb); err != nil {
		return errors.Wrap("getNotProcessedWZCNBurnEvents", "failed to retreive last ZCN processed mint nonce", err).Error()
	}

	if err := cb.Wait(); err != nil {
		return errors.Wrap("getNotProcessedWZCNBurnEvents", "failed to retreive last ZCN processed mint nonce", err).Error()
	}

	if !cb.Success {
		return errors.New("getNotProcessedWZCNBurnEvents", "failed to retreive last ZCN processed mint nonce").Error()
	}

	log.Logger.Debug("MintNonce = " + strconv.Itoa(int(mintNonce)))
	burnEvents, err := bridge.QueryEthereumBurnEvents(strconv.Itoa(int(mintNonce)))
	if err != nil {
		return errors.Wrap("getNotProcessedWZCNBurnEvents", "failed to retreive WZCN burn events", err).Error()
	}

	var result []byte
	result, err = json.Marshal(burnEvents)
	if err != nil {
		return errors.Wrap("getNotProcessedWZCNBurnEvents", "failed to marshal WZCN burn events", err).Error()
	}

	return string(result)
}

// Returns all not processed ZCN burn tickets burned for a certain ethereum address
func getNotProcessedZCNBurnTickets() string {
	userNonce, err := bridge.GetUserNonceMinted(context.Background(), bridge.EthereumAddress)
	if err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to retreive user nonce", err).Error()
	}

	var burnTickets []zcncore.BurnTicket
	cb := wallet.NewZCNStatus(&burnTickets)
	cb.Begin()

	err = zcncore.GetNotProcessedZCNBurnTickets(bridge.EthereumAddress, userNonce.String(), cb)
	if err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to retreive ZCN burn tickets", err).Error()
	}

	if err := cb.Wait(); err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to retreive ZCN burn tickets", err).Error()
	}

	if !cb.Success {
		return errors.New("getNotProcessedZCNBurnTickets", "failed to retreive ZCN burn tickets").Error()
	}

	var result []byte
	result, err = json.Marshal(burnTickets)
	if err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to marshal ZCN burn tickets", err).Error()
	}

	return string(result)
}

// estimateGasAmount performs gas amount estimation for the given transaction.
func estimateGasAmount(from, to string, value int64) string { // nolint:golint,unused
	estimateGasAmountResponse, err := bridge.EstimateGasAmount(context.Background(), from, to, value)
	if err != nil {
		return errors.Wrap("estimateGasPrice", "failed to estimate gas price", err).Error()
	}

	var result []byte
	result, err = json.Marshal(estimateGasAmountResponse)
	if err != nil {
		return errors.Wrap("estimateGasAmount", "failed to marshal gas amount estimation result", err).Error()
	}

	return string(result)
}

// estimateGasPrice performs gas estimation for the given transaction using Alchemy enhanced API returning
// approximate final gas fee.
func estimateGasPrice(from, to string, value int64) string { // nolint:golint,unused
	estimateGasPriceResponse, err := bridge.EstimateGasPrice(context.Background(), from, to, value)
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
