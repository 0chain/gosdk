//go:build mobile
// +build mobile

package zcncore

import (
	"context"
	"encoding/json"

	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/conf"
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

// func updateNetworkDetailsWorker(ctx context.Context) {
// 	ticker := time.NewTicker(time.Duration(networkWorkerTimerInHours) * time.Hour)
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			logging.Info("Network stopped by user")
// 			return
// 		case <-ticker.C:
// 			err := UpdateNetworkDetails()
// 			if err != nil {
// 				logging.Error("Update network detail worker fail", zap.Error(err))
// 				return
// 			}
// 			logging.Info("Successfully updated network details")
// 			return
// 		}
// 	}
// }

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
	n := NewNetwork()
	n.net.Miners = network.Miners
	n.net.Sharders = network.Sharders
	return n, nil
}

//Deprecated: Use client.Node instance to get its network details
func GetNetwork() *Network {
	nodeClient, err := client.GetNode()
	if err != nil {
		panic(err)
	}
	n := NewNetwork()
	n.net.Miners = nodeClient.Network().Miners
	n.net.Sharders = nodeClient.Network().Sharders
	return n
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
