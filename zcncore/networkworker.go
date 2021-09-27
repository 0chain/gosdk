package zcncore

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/util"
	"go.uber.org/zap"
)

const NETWORK_ENDPOINT = "/network"

var networkWorkerTimerInHours = 1

type Network struct {
	Miners   []string `json:"miners"`
	Sharders []string `json:"sharders"`
}

func UpdateNetworkDetailsWorker(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(networkWorkerTimerInHours) * time.Hour)
	for {
		select {
		case <-ctx.Done():
			Logger.Info("Network stopped by user")
			return
		case <-ticker.C:
			err := UpdateNetworkDetails()
			if err != nil {
				Logger.Error("Update network detail worker fail", zap.Error(err))
				return
			}
			Logger.Info("Successfully updated network details")
			return
		}
	}
}

func UpdateNetworkDetails() error {
	networkDetails, err := GetNetworkDetails()
	if err != nil {
		Logger.Error("Failed to update network details ", zap.Error(err))
		return err
	}

	shouldUpdate := UpdateRequired(networkDetails)
	if shouldUpdate {
		_config.isConfigured = false
		_config.chain.Miners = networkDetails.Miners
		_config.chain.Sharders = networkDetails.Sharders
		conf.InitChainNetwork(&conf.Network{
			Sharders: networkDetails.Sharders,
			Miners:   networkDetails.Miners,
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

	minerSame := reflect.DeepEqual(miners, networkDetails.Miners)
	sharderSame := reflect.DeepEqual(sharders, networkDetails.Sharders)

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

	var networkResponse Network
	err = json.Unmarshal([]byte(res.Body), &networkResponse)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling response :")
	}
	return &networkResponse, nil

}
