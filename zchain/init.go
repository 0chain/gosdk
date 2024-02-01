package zchain

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/node"
	"github.com/0chain/gosdk/core/util"
	"go.uber.org/zap"
)

type ChainConfig struct {
	ChainID                 string   
	BlockWorker             string   
	Miners                  []string 
	Sharders                []string 
	SignatureScheme         string   
	MinSubmit               int      
	MinConfirmation         int      
	ConfirmationChainLength int      
	EthNode                 string   
	SharderConsensous       int    
	MaxTxnQuery     		int
	QuerySleepTime  		int
}

// Maintains SDK configuration.
// Initialized through [InitZChain] function.
type Config struct {
	chain *ChainConfig
	logLvl int
}

func (c *Config) getMinMinersSubmit() int {
	return util.MaxInt(1, int(math.Ceil(float64(c.chain.MinSubmit) * float64(len(c.chain.Miners)) / 100)))
}

// Container to hold global data.
// Initialized through [InitZChain] function.
type GlobalContainer struct {
	stableMiners 	[]string
	sharders 		*node.NodeHolder
	config 			*Config

	mguard 			sync.RWMutex
}

func (gc *GlobalContainer) GetStableMiners() []string {
	gc.mguard.RLock()
	defer gc.mguard.Unlock()
	return gc.stableMiners
}

func (gc *GlobalContainer) ResetStableMiners() {
	gc.mguard.Lock()
	defer gc.mguard.Unlock()
	gc.stableMiners = util.GetRandom(gc.config.chain.Miners, gc.config.getMinMinersSubmit())
}

var (
	Gcontainer 	*GlobalContainer
	logging 	logger.Logger
)

type SignScheme string

const (
	ED25519 SignScheme = "ed25519"
	BLS0CHAIN SignScheme = "bls0chain"
)

type OptionKey int

const (
	ChainId OptionKey = iota
	MinSubmit
	MinConfirmation
	ConfirmationChainLength
	EthNode
	SharderConsensous

	LoggingLevel
)

// default options value
const (
	defaultMinSubmit               	= 	int(10)
	defaultMinConfirmation         	= 	int(10)
	defaultConfirmationChainLength 	= 	int(3)
	defaultMaxTxnQuery             	= 	int(5)
	defaultQuerySleepTime          	= 	int(5)
	defaultSharderConsensous		= 	int(3)
	defaultLogLevel                	= 	logger.DEBUG
) 

func init() {
	logging.Init(logger.DEBUG, "0chain-config")
}

func InitZChain(ctx context.Context, blockWorker string, signscheme SignScheme, options map[OptionKey]interface{}) error {
	// get miners, sharders
	miners, sharders, err := getNetworkDetails(ctx, blockWorker)
	if err != nil {
		logging.Error("Failed to get network details ", zap.Error(err))
		return err
	}

	// init config
	config := &Config{
		chain: &ChainConfig{
			BlockWorker: blockWorker,
			SignatureScheme: string(signscheme),
			Miners: miners,
			Sharders: sharders,
			MinSubmit: defaultMinSubmit,
			MinConfirmation: defaultMinConfirmation,
			ConfirmationChainLength: defaultConfirmationChainLength,
			MaxTxnQuery: defaultMaxTxnQuery,
			QuerySleepTime: defaultQuerySleepTime,
			SharderConsensous: util.MinInt(defaultSharderConsensous, len(sharders)),
		},
		logLvl: defaultLogLevel,
	}

	// override default values
	for optionKey, optionValue := range options {
		switch optionKey {
			case ChainId:
				chainId, isTypeString := optionValue.(string)
				if !isTypeString {
					return errors.New("option ChainId is not of string type")
				}
				config.chain.ChainID = chainId
			case MinSubmit:
				minSubmit, isTypeInt := optionValue.(int)
				if !isTypeInt {
					return errors.New("option MinSubmit is not of int type")
				}
				config.chain.MinSubmit = minSubmit
			case MinConfirmation:
				minConfirmation, isTypeInt := optionValue.(int)
				if !isTypeInt {
					return errors.New("option MinConfirmation is not of int type") 
				}
				config.chain.MinConfirmation = minConfirmation
			case ConfirmationChainLength:
				confirmationChainLength, isTypeInt := optionValue.(int)
				if !isTypeInt {
					return errors.New("option ConfirmationChainLength is not of int type") 
				}
				config.chain.ConfirmationChainLength = confirmationChainLength
			case EthNode:
				ethNode, isTypeString := optionValue.(string)
				if !isTypeString {
					return errors.New("option EthNode is not of string type")
				}
				config.chain.EthNode = ethNode
			case SharderConsensous:
				sharderConsensous, isTypeInt := optionValue.(int)
				if !isTypeInt {
					return errors.New("option SharderConsensous is not of int type")
				}
				config.chain.SharderConsensous = sharderConsensous
			case LoggingLevel:
				loggingLevel, isTypeInt := optionValue.(int)
				if !isTypeInt {
					return errors.New("option LoggingLevel is not of int type")
				}
				logging.SetLevel(loggingLevel)
		}
	}

	// init GlobalContainer
	Gcontainer = &GlobalContainer {
		stableMiners: util.GetRandom(miners, config.getMinMinersSubmit()),
		sharders: node.NewHolder(config.chain.Sharders, config.chain.SharderConsensous),
		config: config,
	}
	
	// update miners, sharders periodically
	go func() {
		ticker := time.NewTicker(time.Duration(1) * time.Hour)
		defer ticker.Stop()
		for {
			select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					
			}
		}
	}()

	// init modules
	node.InitCache(Gcontainer.sharders)
	return nil
}

func getNetworkDetails(ctx context.Context, blockWorker string) ([]string, []string, error) {
	networkUrl := blockWorker + "/network"
	networkGetCtx, networkGetCancelCtx := context.WithTimeoutCause(ctx, 60 * time.Second, errors.New("timeout connecting network: " + networkUrl))
	defer networkGetCancelCtx()
	req, err := util.NewHTTPGetRequestContext(networkGetCtx, networkUrl)
	if err != nil {
		return nil, nil, errors.New("Unable to create new http request with error: " + err.Error())
	}
	res, err := req.Get()
	if err != nil {
		return nil, nil, errors.New("Unable to get http request with error: " + err.Error())
	}
	if res.StatusCode != http.StatusOK {
		return nil, nil, errors.New("Unable to get http request with status Ok: " + res.Status)
	}
	type responseBody struct {
		Miners   []string `json:"miners"`
		Sharders []string `json:"sharders"`
	}
	var respBody responseBody
	err = json.Unmarshal([]byte(res.Body), &respBody)
	if err != nil {
		return nil, nil, errors.New("Error unmarshaling response :" + res.Body)
	}
	return respBody.Miners, respBody.Sharders, nil

}