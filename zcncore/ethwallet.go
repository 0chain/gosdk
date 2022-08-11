//go:build !mobile
// +build !mobile

package zcncore

import "github.com/ethereum/go-ethereum/ethclient"

func GetEthClient() (*ethclient.Client, error) {
	return getEthClient()
}
