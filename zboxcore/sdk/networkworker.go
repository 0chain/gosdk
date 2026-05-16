package sdk

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/0chain/gosdk/core/node"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"go.uber.org/zap"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const NETWORK_ENDPOINT = "/network"

type Network struct {
	Miners   []string `json:"miners"`
	Sharders []string `json:"sharders"`
}

func UpdateNetworkDetailsWorker(ctx context.Context) {
	// Exponential backoff from 30s while chain is unreachable, capped at
	// the long refresh interval. After a successful refresh, keep
	// polling at the long interval so we still pick up network changes.
	delay := 30 * time.Second
	maxDelay := time.Duration(networkWorkerTimerInHours) * time.Hour
	timer := time.NewTimer(delay)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			l.Logger.Info("Network stopped by user")
			return
		case <-timer.C:
			if err := UpdateNetworkDetails(); err != nil {
				l.Logger.Error("Update network details failed; will retry", zap.Error(err))
				if delay < maxDelay {
					delay *= 2
					if delay > maxDelay {
						delay = maxDelay
					}
				}
				timer.Reset(delay)
				continue
			}
			l.Logger.Info("Successfully updated network details")
			delay = maxDelay
			timer.Reset(maxDelay)
		}
	}
}

func UpdateNetworkDetails() error {
	networkDetails, err := GetNetworkDetails()
	if err != nil {
		l.Logger.Error("Failed to update network details ", zap.Error(err))
		return err
	}

	shouldUpdate := UpdateRequired(networkDetails)
	if shouldUpdate {
		forceUpdateNetworkDetails(networkDetails)
	}
	return nil
}

func InitNetworkDetails() error {
	networkDetails, err := GetNetworkDetails()
	if err != nil {
		l.Logger.Error("Failed to update network details ", zap.Error(err))
		return err
	}
	forceUpdateNetworkDetails(networkDetails)
	return nil
}

func forceUpdateNetworkDetails(networkDetails *Network) {
	sdkInitialized = false
	blockchain.SetMiners(networkDetails.Miners)
	blockchain.SetSharders(networkDetails.Sharders)
	node.InitCache(blockchain.Sharders)
	conf.InitChainNetwork(&conf.Network{
		Sharders: networkDetails.Sharders,
		Miners:   networkDetails.Miners,
	})
	sdkInitialized = true
}

func UpdateRequired(networkDetails *Network) bool {
	miners := blockchain.GetMiners()
	sharders := blockchain.GetAllSharders()
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
		return nil, errors.New("get_network_details_error", "Unable to create new http request with error "+err.Error())
	}

	var networkResponse Network
	err = zboxutil.HttpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Get network error : ", err)
			return err
		}

		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "Error reading response : ")
		}

		l.Logger.Debug("Get network result:", string(respBody))
		if resp.StatusCode == http.StatusOK {
			err = json.Unmarshal(respBody, &networkResponse)
			if err != nil {
				return errors.Wrap(err, "Error unmarshaling response :")
			}
			return nil
		}
		return errors.New(strconv.Itoa(resp.StatusCode), "Get network details status not OK")

	})
	return &networkResponse, err
}
