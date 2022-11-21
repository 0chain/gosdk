package main

import (
	"context"
	"github.com/0chain/gosdk/zcnbridge/ethereum"
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

func mintZCN(burnTrxHash string, timeout int) string {

	// ASK authorizers for burn tickets to mint in ZCN
	mintPayload, err := bridge.QueryZChainMintPayload(burnTrxHash)
	if err != nil {
		return errors.Wrap("mint", "failed to QueryZChainMintPayload", err).Error()
	}

	c, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	hash, err := bridge.MintZCN(c, mintPayload)
	if err != nil {
		return errors.Wrap("mint", "failed to MintZCN for txn "+hash, err).Error()
	}

	return hash
}

func getMintWZCNPayload(burnTrxHash string) (*ethereum.MintPayload, error) {
	mintPayload, err := bridge.QueryEthereumMintPayload(burnTrxHash)
	if err != nil {
		return nil, errors.Wrap("mint", "failed to QueryZChainMintPayload", err)
	}
	return mintPayload, nil
}
