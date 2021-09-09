package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	. "github.com/0chain/gosdk/zboxcore/logger"
	"go.uber.org/zap"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const NETWORK_ENDPOINT = "/network"

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
		sdkInitialized = false
		blockchain.SetMiners(networkDetails.Miners)
		blockchain.SetSharders(networkDetails.Sharders)
		sdkInitialized = true
	}
	return nil
}

func UpdateRequired(networkDetails *Network) bool {
	miners := blockchain.GetMiners()
	sharders := blockchain.GetSharders()
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
	req, ctx, cncl, err := zboxutil.NewHTTPRequest(http.MethodGet, blockchain.GetBlockWorker()+NETWORK_ENDPOINT, nil)
	if err != nil {
		return nil, common.NewError("get_network_details_error", "Unable to create new http request with error "+err.Error())
	}

	var networkResponse Network
	err = zboxutil.HttpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Get network error : ", err)
			return err
		}

		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Error reading response : %s", err.Error())
		}

		Logger.Debug("Get network result:", string(respBody))
		if resp.StatusCode == http.StatusOK {
			err = json.Unmarshal(respBody, &networkResponse)
			if err != nil {
				return fmt.Errorf("Error unmarshaling response : %s", err.Error())
			}
			return nil
		}
		return fmt.Errorf("Get network details status not OK, Status : %v", resp.StatusCode)
	})
	return &networkResponse, err
}
