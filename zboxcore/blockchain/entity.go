package blockchain

import (
	"encoding/json"
	"math"
	"sync/atomic"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/node"
)

type ChainConfig struct {
	BlockWorker     string
	Sharders        []string
	Miners          []string
	MinSubmit       int
	MinConfirmation int
	ChainID         string
	MaxTxnQuery     int
	QuerySleepTime  int
}

// StakePoolSettings information.
type StakePoolSettings struct {
	DelegateWallet string         `json:"delegate_wallet"`
	MinStake       common.Balance `json:"min_stake"`
	MaxStake       common.Balance `json:"max_stake"`
	NumDelegates   int            `json:"num_delegates"`
	ServiceCharge  float64        `json:"service_charge"`
}

// UpdateStakePoolSettings information.
type UpdateStakePoolSettings struct {
	DelegateWallet *string         `json:"delegate_wallet,omitempty"`
	MinStake       *common.Balance `json:"min_stake,omitempty"`
	MaxStake       *common.Balance `json:"max_stake,omitempty"`
	NumDelegates   *int            `json:"num_delegates,omitempty"`
	ServiceCharge  *float64        `json:"service_charge,omitempty"`
}

type ValidationNode struct {
	ID                string            `json:"id"`
	BaseURL           string            `json:"url"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

type UpdateValidationNode struct {
	ID                string                   `json:"id"`
	BaseURL           *string                  `json:"url"`
	StakePoolSettings *UpdateStakePoolSettings `json:"stake_pool_settings"`
}

type StorageNode struct {
	ID      string `json:"id"`
	Baseurl string `json:"url"`

	skip uint64 `json:"-"` // skip on error
}

func (sn *StorageNode) SetSkip(t bool) {
	var val uint64
	if t {
		val = 1
	}
	atomic.StoreUint64(&sn.skip, val)
}

func (sn *StorageNode) IsSkip() bool {
	return atomic.LoadUint64(&sn.skip) > 0
}

func PopulateNodes(nodesjson string) ([]string, error) {
	sharders := make([]string, 0)
	err := json.Unmarshal([]byte(nodesjson), &sharders)
	return sharders, err
}

var chain *ChainConfig
var Sharders *node.NodeHolder

func init() {
	chain = &ChainConfig{
		MaxTxnQuery:     5,
		QuerySleepTime:  5,
		MinSubmit:       50,
		MinConfirmation: 50,
	}
}

func GetChainID() string {
	return chain.ChainID
}

func PopulateChain(minerjson string, sharderjson string) error {
	var err error
	chain.Miners, err = PopulateNodes(minerjson)
	if err != nil {
		return err
	}
	sharders, err := PopulateNodes(sharderjson)
	if err != nil {
		return err
	}
	SetSharders(sharders)
	return nil
}

func GetBlockWorker() string {
	return chain.BlockWorker
}

func GetAllSharders() []string {
	return Sharders.All()
}
func GetSharders() []string {
	return Sharders.Healthy()
}

func GetMiners() []string {
	return chain.Miners
}

func GetMaxTxnQuery() int {
	return chain.MaxTxnQuery
}

func GetQuerySleepTime() int {
	return chain.QuerySleepTime
}

func GetMinSubmit() int {
	return chain.MinSubmit
}

func GetMinConfirmation() int {
	return chain.MinConfirmation
}

func SetBlockWorker(blockWorker string) {
	chain.BlockWorker = blockWorker
}

func GetReqConsensus(n, minReqNum, reqPercent int) int {
	reqNum := int(math.Max(float64(minReqNum), float64(n*reqPercent/100)))
	if reqNum > n {
		reqNum = n
	}
	return reqNum
}

func SetSharders(sharderArray []string) {
	consensus := conf.DefaultSharderConsensous
	config, err := conf.GetClientConfig()
	if err == nil && config != nil {
		consensus = GetReqConsensus(len(sharderArray), consensus, config.MinConfirmation)
	}
	Sharders = node.NewHolder(sharderArray, consensus)
}

func SetMiners(minerArray []string) {
	chain.Miners = minerArray
}

func SetChainID(id string) {
	chain.ChainID = id
}

func SetMaxTxnQuery(num int) {
	chain.MaxTxnQuery = num
}

func SetQuerySleepTime(time int) {
	chain.QuerySleepTime = time
}

func SetMinSubmit(minSubmit int) {
	chain.MinSubmit = minSubmit
}

func SetMinConfirmation(minConfirmation int) {
	chain.MinConfirmation = minConfirmation
}
