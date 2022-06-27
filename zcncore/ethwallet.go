//go:build !mobile
// +build !mobile

package zcncore

func GetEthClient() (*ethclient.Client, error) {
	return getEthClient()
}
