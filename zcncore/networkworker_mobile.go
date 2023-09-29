//go:build mobile
// +build mobile

package zcncore

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/0chain/gosdk/core/transaction"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/util"
	"go.uber.org/zap"
)

const NETWORK_ENDPOINT = "/network"

var networkWorkerTimerInHours = 1

type Network struct {
	net network
}

func NewNetwork() *Network {
	return &Network{}
}

func (net *Network) AddMiner(miner string) {
	net.net.Miners = append(net.net.Miners, miner)
}

func (net *Network) AddSharder(sharder string) {
	net.net.Sharders = append(net.net.Sharders, sharder)
}

type network struct {
	Miners   []string `json:"miners"`
	Sharders []string `json:"sharders"`
}

func updateNetworkDetailsWorker(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(networkWorkerTimerInHours) * time.Hour)
	for {
		select {
		case <-ctx.Done():
			logging.Info("Network stopped by user")
			return
		case <-ticker.C:
			err := UpdateNetworkDetails()
			if err != nil {
				logging.Error("Update network detail worker fail", zap.Error(err))
				return
			}
			logging.Info("Successfully updated network details")
			return
		}
	}
}

func UpdateNetworkDetails() error {
	networkDetails, err := GetNetworkDetails()
	if err != nil {
		logging.Error("Failed to update network details ", zap.Error(err))
		return err
	}

	shouldUpdate := UpdateRequired(networkDetails)
	if shouldUpdate {
		_config.isConfigured = false
		_config.chain.Miners = networkDetails.net.Miners
		_config.chain.Sharders = networkDetails.net.Sharders
		Sharders = util.NewHolder(networkDetails.net.Sharders, len(networkDetails.net.Sharders))
		transaction.InitCache(Sharders)
		conf.InitChainNetwork(&conf.Network{
			Sharders: networkDetails.net.Sharders,
			Miners:   networkDetails.net.Miners,
		})
		_config.isConfigured = true
	}
	return nil
}

func UpdateRequired(networkDetails *Network) bool {
	miners := _config.chain.Miners
	sharders := _config.chain.Sharders
	if len(miners) == 0 || len(sharders) == 0 {
		return true
	}

	minerSame := reflect.DeepEqual(miners, networkDetails.net.Miners)
	sharderSame := reflect.DeepEqual(sharders, networkDetails.net.Sharders)

	if minerSame && sharderSame {
		return false
	}
	return true
}

func GetNetworkDetails() (*Network, error) {
	req, err := util.NewHTTPGetRequest(_config.chain.BlockWorker + NETWORK_ENDPOINT)
	if err != nil {
		return nil, errors.New("get_network_details_error", "Unable to create new http request with error "+err.Error())
	}

	res, err := req.Get()
	if err != nil {
		return nil, errors.New("get_network_details_error", "Unable to get http request with error "+err.Error())
	}

	var networkResponse network
	err = json.Unmarshal([]byte(res.Body), &networkResponse)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling response :")
	}
	return &Network{net: networkResponse}, nil
}

func GetNetwork() *Network {
	return &Network{
		net: network{
			Miners:   _config.chain.Miners,
			Sharders: _config.chain.Sharders,
		},
	}
}

func SetNetwork(net *Network) {
	_config.chain.Miners = net.net.Miners
	_config.chain.Sharders = net.net.Sharders
	Sharders = util.NewHolder(_config.chain.Sharders, len(_config.chain.Sharders))

	transaction.InitCache(Sharders)

	conf.InitChainNetwork(&conf.Network{
		Miners:   net.net.Miners,
		Sharders: net.net.Sharders,
	})
}

func GetNetworkJSON() string {
	network := GetNetwork()
	networkBytes, _ := json.Marshal(network)
	return string(networkBytes)
}
