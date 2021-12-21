package stubs

import (
	"github.com/0chain/gosdk/zcnbridge/ethereum"
	"github.com/0chain/gosdk/zcnbridge/zcnsc"
)

// GetAuthorizerZCNResponseStub Authorizers returned ticket from Ethereum side to
// mint tokens in ZCN
func GetAuthorizerZCNResponseStub(hash string) (*zcnsc.MintPayload, error) {
	// 1. To mint to ZCN client, the client must have the wallet registered in ZCN chain
	// 2. Create mint payload
	// 3. Call mint on zcnsc contract
	return nil, nil
}

// GetAuthorizerEthereumResponseStub Authorizers returned ticket from ZCN side to
// mint tokens in Ethereum
func GetAuthorizerEthereumResponseStub(hash string) (*ethereum.MintPayload, error) {
	// 1. Client has ethereum account
	// 2. Create mint payload
	// 3. Call mint function on the bridge solidity contract
	return nil, nil
}
