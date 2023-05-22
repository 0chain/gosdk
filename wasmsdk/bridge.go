package main

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/0chain/gosdk/zcnbridge"
	"github.com/0chain/gosdk/zcnbridge/errors"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/0chain/gosdk/zcncore"
)

var bridge *zcnbridge.BridgeClient

func initBridge(
	ethereumAddress string,
	bridgeAddress string,
	authorizersAddress string,
	TokenAddress string,
	ethereumNodeURL string,
	gasLimit uint64,
	value int64,
	consensusThreshold float64) error {
	wallet := zcncore.GetWalletRaw()
	if len(wallet.ClientID) == 0 {
		return errors.New("wallet_error", "wallet is not set")
	}

	bridge = &zcnbridge.BridgeClient{
		EthereumAddress:    ethereumAddress,
		BridgeAddress:      bridgeAddress,
		AuthorizersAddress: authorizersAddress,
		TokenAddress:       TokenAddress,
		Password:           "",
		EthereumNodeURL:    ethereumNodeURL,
		Homedir:            ".",
		GasLimit:           gasLimit,
		ConsensusThreshold: consensusThreshold,
	}

	return nil
}

// Burns ZCN tokens and returns a hash of the burn transaction
func burnZCN(amount uint64) string {
	if bridge == nil {
		return errors.New("burnZCN", "bridge is not initialized").Error()
	}

	tx, err := bridge.BurnZCN(context.Background(), amount)
	if err != nil {
		return errors.Wrap("burnZCN", "failed to burn ZCN tokens", err).Error()
	}

	return tx.Hash
}

// Mints ZCN tokens and returns a hash of the mint transaction
func mintZCN(burnTrxHash string, timeout int) string {

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
