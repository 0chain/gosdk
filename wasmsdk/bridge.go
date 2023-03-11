package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/0chain/gosdk/zcnbridge"
	"github.com/0chain/gosdk/zcnbridge/errors"
	"github.com/0chain/gosdk/zcncore"
)

//type BridgeSDKConfig struct {
//	LogLevel         *string
//	LogPath          *string
//	ConfigBridgeFile *string
//	ConfigChainFile  *string
//	ConfigDir        *string
//	Development      *bool
//}

var bridge *zcnbridge.BridgeClient

func initBridge(
	ethereumAddress string,
	bridgeAddress string,
	authorizersAddress string,
	wzcnAddress string,
	ethereumNodeURL string,
	gasLimit uint64,
	value int64,
	consensusThreshold float64) error {
	// Create bridge client configuration
	//zcnbridge.CreateInitialClientConfig(
	//	*cfg.ConfigBridgeFile,
	//	*cfg.ConfigDir,
	//	"0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947",
	//	"0xF26B52df8c6D9b9C20bfD7819Bed75a75258c7dB",
	//	"0x930E1BE76461587969Cb7eB9BFe61166b1E70244",
	//	"https://ropsten.infura.io/v3/22cb2849f5f74b8599f3dc2a23085bd4",
	//	"password",
	//	300000,
	//	0,
	//	75.0,
	//)
	//
	cfg := zcnbridge.BridgeClientYaml{
		Password:           "",
		EthereumAddress:    ethereumAddress,
		BridgeAddress:      bridgeAddress,
		AuthorizersAddress: authorizersAddress,
		WzcnAddress:        wzcnAddress,
		EthereumNodeURL:    ethereumNodeURL,
		GasLimit:           gasLimit,
		Value:              value,
		ConsensusThreshold: consensusThreshold,
	}

	wallet := zcncore.GetWalletRaw()
	if len(wallet.ClientID) == 0 {
		return errors.New("wallet_error", "wallet is not set")
	}

	bridge = zcnbridge.CreateBridgeClientWithConfig(cfg, &wallet)

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

// Returns all not processed WZCN burn tickets burned for client id given as a param
func getNotProcessedWZCNBurnTickets() string {
	var cb zcncore.GetMintNonceCallbackStub
	if err := zcncore.GetMintNonce(&cb); err != nil {
		return errors.Wrap("getNotProcessedWZCNBurnTickets", "failed to retreive ZCN processed mint nonces", err).Error()
	}

	burnTickets, err := bridge.GetNotProcessedWZCNBurnTickets(context.Background(), cb.Value)
	if err != nil {
		return errors.Wrap("getNotProcessedWZCNBurnTickets", "failed to retreive WZCN burn tickets", err).Error()
	}

	var result []byte
	result, err = json.Marshal(burnTickets)
	if err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to marshal WZCN burn tickets", err).Error()
	}

	return string(result)
}

// Returns all not processed ZCN burn tickets burned for a certain ethereum address
func getNotProcessedZCNBurnTickets() string {
	userNonce, err := bridge.GetUserNonceMinted(context.Background(), bridge.EthereumAddress)
	if err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to retreive user nonce", err).Error()
	}

	var cb zcncore.GetNotProcessedZCNBurnTicketsCallbackStub
	err = zcncore.GetNotProcessedZCNBurnTickets(bridge.EthereumAddress, userNonce.Int64(), &cb)
	if err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to retreive ZCN burn tickets", err).Error()
	}

	var result []byte
	result, err = json.Marshal(cb.Value)
	if err != nil {
		return errors.Wrap("getNotProcessedZCNBurnTickets", "failed to marshal ZCN burn tickets", err).Error()
	}

	return string(result)
}
