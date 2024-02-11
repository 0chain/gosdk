package client

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/node"
	"github.com/0chain/gosdk/core/util"
	"go.uber.org/zap"
)

var (
	logging    logger.Logger
	nodeClient *Node
)

func init() {
	logging.Init(logger.DEBUG, "0chain-core")
}

// Maintains central states of SDK (client's context, network).
// Initialized through [Init] function.
// Use client.GetNode() to get its instance after Init is called.
type Node struct {
	stableMiners []string
	sharders     *node.NodeHolder
	network 	*conf.Network
	clientCtx context.Context

	networkGuard sync.RWMutex
}

// Returns stable miner urls.
// Length of stable miners is depedent on config's MinSubmit and number of miners in network.
func (n *Node) GetStableMiners() []string {
	n.networkGuard.RLock()
	defer n.networkGuard.RUnlock()
	return n.stableMiners
}

// ResetStableMiners resets stable miners as a random permutation of network miners.
// Length of stable miners is depedent on config's MinSubmit and number of miners in network.
func (n *Node) ResetStableMiners() {
	n.networkGuard.Lock()
	defer n.networkGuard.Unlock()
	cfg, _ := conf.GetClientConfig()
	reqMiners := util.MaxInt(1, int(math.Ceil(float64(cfg.MinSubmit)*float64(len(n.network.Miners))/100)))
	n.stableMiners = util.GetRandom(n.network.Miners, reqMiners)
}

// Returns minimum sharders used for verification
func (n *Node) GetMinShardersVerify() int {
	n.networkGuard.RLock()
	defer n.networkGuard.RUnlock()
	cfg, _ := conf.GetClientConfig()
	minSharders := util.MaxInt(1, int(math.Ceil(float64(cfg.MinConfirmation)*float64(len(n.sharders.Healthy()))/100)))
	logging.Info("Minimum sharders used for verify :", minSharders)
	return minSharders
}

// Returns NodeHolder instance
func (n *Node) Sharders() *node.NodeHolder {
	n.networkGuard.RLock()
	defer n.networkGuard.RUnlock()
	return n.sharders
}

// Returns network configuration
func (n *Node) Network() *conf.Network {
	n.networkGuard.RLock()
	defer n.networkGuard.RUnlock()
	return n.network
}

// Gets network details and return it as second value.
// First value is true iff current network details doesn't match existing network details.
// Use node.UpdateNetwork() method to set the new network.
func (n *Node) ShouldUpdateNetwork() (bool, *conf.Network, error) {
	cfg, err := conf.GetClientConfig()
	if err != nil {
		return false, nil, err
	}
	network, err := GetNetwork(n.clientCtx, cfg.BlockWorker)
	if err != nil {
		logging.Error("Failed to get network details ", zap.Error(err))
		return false, nil, err
	}
	n.networkGuard.RLock()
	defer n.networkGuard.RUnlock()
	if reflect.DeepEqual(network, n.network) {
		return false, network, nil
	}
	return true, network, nil
}

// Use node.UpdateNetwork() method to set the new network. 
func (n *Node) UpdateNetwork(network *conf.Network) error {
	n.networkGuard.Lock()
	defer n.networkGuard.Unlock()
	cfg, err := conf.GetClientConfig()
	if err != nil {
		return err
	}
	n.network = network
	n.sharders = node.NewHolder(n.network.Sharders, util.MinInt(len(n.network.Sharders), util.MaxInt(cfg.SharderConsensous, conf.DefaultSharderConsensous)))
	node.InitCache(n.sharders)
	return nil
}

// Initializes SDK. 
func Init(ctx context.Context, cfg conf.Config) error {
	// validate
	err := validate(&cfg)
	if err != nil {
		return err
	}

	// set default value for options if unset
	setOptionsDefaultValue(&cfg)

	network, err := GetNetwork(ctx, cfg.BlockWorker)
	if err != nil {
		logging.Error("Failed to get network details ", zap.Error(err))
		return err
	}

	reqMiners := util.MaxInt(1, int(math.Ceil(float64(cfg.MinSubmit)*float64(len(network.Miners))/100)))
	sharders := node.NewHolder(network.Sharders, util.MinInt(len(network.Sharders), util.MaxInt(cfg.SharderConsensous, conf.DefaultSharderConsensous)))
	nodeClient = &Node{
		stableMiners: util.GetRandom(network.Miners, reqMiners),
		sharders:     sharders,
		network: network,
		clientCtx: ctx,
	}

	//init packages
	conf.InitClientConfig(&cfg)
	node.InitCache(nodeClient.sharders)

	// update Network periodically
	go func() {
		ticker := time.NewTicker(time.Duration(1) * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				shouldUpdate, network, err := nodeClient.ShouldUpdateNetwork()
				if err != nil {
					logging.Error("error on ShouldUpdateNetwork check: ", err)
					continue
				}
				if shouldUpdate {
					if err = nodeClient.UpdateNetwork(network); err != nil {
						logging.Error("error on updating network: ", err)
					}
				}
			}
		}
	}()

	return nil
}

// Returns Node instance. If this function is called before Init(), error is returned.
func GetNode() (*Node, error) {
	if nodeClient != nil {
		return nodeClient, nil
	}
	return nil, errors.New("0chain-sdk is not initialized")
}

// GetNetwork gets current network details from 0chain network.
func GetNetwork(ctx context.Context, blockWorker string) (*conf.Network, error) {
	networkUrl := blockWorker + "/network"
	networkGetCtx, networkGetCancelCtx := context.WithTimeout(ctx, 60*time.Second)
	defer networkGetCancelCtx()
	req, err := util.NewHTTPGetRequestContext(networkGetCtx, networkUrl)
	if err != nil {
		return nil, errors.New("Unable to create new http request with error: " + err.Error())
	}
	res, err := req.Get()
	if err != nil {
		return nil, errors.New("Unable to get http request with error: " + err.Error())
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("Unable to get http request with status Ok: " + res.Status)
	}
	type network struct {
		Miners   []string `json:"miners"`
		Sharders []string `json:"sharders"`
	}
	var n network
	err = json.Unmarshal([]byte(res.Body), &n)
	if err != nil {
		return nil, errors.New("Error unmarshaling response :" + res.Body)
	}
	return conf.NewNetwork(n.Miners, n.Sharders)
}

func validate(cfg *conf.Config) error {
	if cfg.BlockWorker == "" {
		return errors.New("chain BlockWorker can't be empty")
	}
	if cfg.SignatureScheme != string(constants.BLS0CHAIN) && cfg.SignatureScheme != string(constants.ED25519) {
		return errors.New("invalid/unsupported signature scheme")
	}
	return nil
}

func setOptionsDefaultValue(cfg *conf.Config) {
	if cfg.MinSubmit <= 0 {
		cfg.MinSubmit = conf.DefaultMinSubmit
	}
	if cfg.MinConfirmation <= 0 {
		cfg.MinConfirmation = conf.DefaultMinConfirmation
	}
	if cfg.ConfirmationChainLength <= 0 {
		cfg.ConfirmationChainLength = conf.DefaultConfirmationChainLength
	}
	if cfg.MaxTxnQuery <= 0 {
		cfg.MaxTxnQuery = conf.DefaultMaxTxnQuery
	}
	if cfg.QuerySleepTime <= 0 {
		cfg.QuerySleepTime = conf.DefaultMaxTxnQuery
	}
	if cfg.SharderConsensous <= 0 {
		cfg.SharderConsensous = conf.DefaultSharderConsensous
	}
}
