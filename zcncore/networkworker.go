//go:build !mobile
// +build !mobile

package zcncore

import (
	"context"
	"encoding/json"

	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/conf"
)

const NETWORK_ENDPOINT = "/network"

type Network struct {
	Miners   []string `json:"miners"`
	Sharders []string `json:"sharders"`
}

//Deprecated: Get client.Node instance to check whether network update is required and update network accordingly
func UpdateNetworkDetails() error {
	nodeClient, err := client.GetNode()
	if err != nil {
		return err
	}
	shouldUpdate, network, err := nodeClient.ShouldUpdateNetwork()
	if err != nil {
		logging.Error("error on ShouldUpdateNetwork check: ", err)
		return err
	}
	if shouldUpdate {
		logging.Info("Updating network")
		if err = nodeClient.UpdateNetwork(network); err != nil {
			logging.Error("error on updating network: ", err)
			return err
		}
		logging.Info("network updated successfully")
	}
	return nil
}

//Deprecated: Get client.Node instance to check whether network update is required 
func UpdateRequired(networkDetails *Network) bool {
	nodeClient, err := client.GetNode()
	if err != nil {
		panic(err)
	}
	shouldUpdate, _, err := nodeClient.ShouldUpdateNetwork()
	if err != nil {
		logging.Error("error on ShouldUpdateNetwork check: ", err)
		panic(err)
	}
	return shouldUpdate
}

//Deprecated: Use client.GetNetwork() function
func GetNetworkDetails() (*Network, error) {
	cfg, err := conf.GetClientConfig()
	if err != nil {
		return nil, err
	}
	network, err := client.GetNetwork(context.Background(), cfg.BlockWorker)
	if err != nil {
		return nil, err
	}
	return &Network{
		Miners: network.Miners,
		Sharders: network.Sharders,
	}, nil
}

//Deprecated: Use client.Node instance to get its network details
func GetNetwork() *Network {
	nodeClient, err := client.GetNode()
	if err != nil {
		panic(err)
	}
	return &Network{
		Miners:   nodeClient.Network().Miners,
		Sharders: nodeClient.Network().Sharders,
	}
}

//Deprecated: Use client.Node instance UpdateNetwork() method 
func SetNetwork(miners []string, sharders []string) {
	nodeClient, err := client.GetNode()
	if err != nil {
		panic(err)
	}
	network, err := conf.NewNetwork(miners, sharders)
	if err != nil {
		panic(err)
	}
	err = nodeClient.UpdateNetwork(network)
	if err != nil {
		logging.Error("error updating network: ", err)
		panic(err)
	}
	logging.Info("network updated successfully")
}

//Deprecated: Use client.GetNetwork() function 
func GetNetworkJSON() string {
	network := GetNetwork()
	networkBytes, _ := json.Marshal(network)
	return string(networkBytes)
}
