package main

import (
	"context"
	"time"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zcnbridge"
	"github.com/0chain/gosdk/zcnbridge/errors"
)

//type BridgeSDKConfig struct {
//	LogLevel         *string
//	LogPath          *string
//	ConfigBridgeFile *string
//	ConfigChainFile  *string
//	ConfigDir        *string
//	Development      *bool
//}

var client *zcnbridge.BridgeClient

func initBridge(cfg zcnbridge.BridgeClientYaml, wallet zcncrypto.Wallet) {
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
	//yaml := zcnbridge.BridgeClientYaml{
	//	Password:           "",
	//	EthereumAddress:    "",
	//	BridgeAddress:      "",
	//	AuthorizersAddress: "",
	//	WzcnAddress:        "",
	//	EthereumNodeURL:    "",
	//	GasLimit:           0,
	//	Value:              0,
	//	ConsensusThreshold: 0,
	//}
	//
	//*zcncrypto.Wallet{
	//	ClientID:    "",
	//	ClientKey:   "",
	//	Keys:        nil,
	//	Mnemonic:    "",
	//	Version:     "",
	//	DateCreated: "",
	//	Nonce:       0,
	//}

	client = zcnbridge.CreateBridgeClientWithConfig(cfg, &wallet)
}

func mintZCN(burnTrxHash string, timeout int64) (string, error) {

	// ASK authorizers for burn tickets to mint in WZCN
	mintPayload, err := client.QueryZChainMintPayload(burnTrxHash)
	if err != nil {
		return "", errors.Wrap("mint", "failed to QueryZChainMintPayload", err)
	}

	c, cancel := context.WithTimeout(context.Background(), time.Duration(timeout*time.Second.Nanoseconds()))
	defer cancel()

	mintTrx, err := client.MintZCN(c, mintPayload)
	if err != nil {
		return "", errors.Wrap("mint", "failed to MintZCN for txn "+mintTrx.Hash, err)
	}

	return mintTrx.Hash, nil
}
