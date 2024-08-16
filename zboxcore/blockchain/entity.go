// Methods and types for blockchain entities and interactions.
package blockchain

import (
	"encoding/json"
	"math"
	"sync"
	"sync/atomic"

	"github.com/0chain/gosdk/core/util"

	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/node"
)

var miners []string
var mGuard sync.Mutex

func getMinMinersSubmit() int {
	minMiners := util.MaxInt(calculateMinRequired(float64(chain.MinSubmit), float64(len(chain.Miners))/100), 1)
	return minMiners
}

func calculateMinRequired(minRequired, percent float64) int {
	return int(math.Ceil(minRequired * percent))
}

// GetStableMiners get stable miners
func GetStableMiners() []string {
	mGuard.Lock()
	defer mGuard.Unlock()
	if len(miners) == 0 {
		miners = util.GetRandom(chain.Miners, getMinMinersSubmit())
	}

	return miners
}

// ResetStableMiners reset stable miners to random miners
func ResetStableMiners() {
	mGuard.Lock()
	defer mGuard.Unlock()
	miners = util.GetRandom(chain.Miners, getMinMinersSubmit())
}

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
	DelegateWallet string  `json:"delegate_wallet"`
	NumDelegates   int     `json:"num_delegates"`
	ServiceCharge  float64 `json:"service_charge"`
}

// UpdateStakePoolSettings represent stake pool information of a provider node.
type UpdateStakePoolSettings struct {
	DelegateWallet *string  `json:"delegate_wallet,omitempty"`
	NumDelegates   *int     `json:"num_delegates,omitempty"`
	ServiceCharge  *float64 `json:"service_charge,omitempty"`
}

// ValidationNode represents a validation node (miner)
type ValidationNode struct {
	ID                string            `json:"id"`
	BaseURL           string            `json:"url"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

// UpdateValidationNode represents a validation node (miner) update
type UpdateValidationNode struct {
	ID                string                   `json:"id"`
	BaseURL           *string                  `json:"url"`
	StakePoolSettings *UpdateStakePoolSettings `json:"stake_pool_settings"`
}

// StorageNode represents a storage node (blobber)
type StorageNode struct {
	ID                string `json:"id"`
	Baseurl           string `json:"url"`
	AllocationVersion int64  `json:"-"`

	skip uint64 `json:"-"` // skip on error
}

// SetSkip set skip, whether to skip this node in operations or not
//   - t is the boolean value
func (sn *StorageNode) SetSkip(t bool) {
	var val uint64
	if t {
		val = 1
	}
	atomic.StoreUint64(&sn.skip, val)
}

// IsSkip check if skip
func (sn *StorageNode) IsSkip() bool {
	return atomic.LoadUint64(&sn.skip) > 0
}

// PopulateNodes populate nodes from json string
//   - nodesjson is the json string
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
		MinSubmit:       10,
		MinConfirmation: 10,
	}
}

// GetChainConfig get chain config
func GetChainID() string {
	return chain.ChainID
}

// PopulateChain populate chain from json string
//   - minerjson is the array of miner urls, serialized as json
//   - sharderjson is the array of sharder urls, serialized as json
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

// GetBlockWorker get block worker
func GetBlockWorker() string {
	return chain.BlockWorker
}

// GetSharders get sharders
func GetAllSharders() []string {
	return Sharders.All()
}

// GetSharders get healthy sharders
func GetSharders() []string {
	return Sharders.Healthy()
}

// GetMiners get miners
func GetMiners() []string {
	return chain.Miners
}

// GetMaxTxnQuery get max transaction query
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

func SetSharders(sharderArray []string) {
	consensus := conf.DefaultSharderConsensous
	config, err := conf.GetClientConfig()
	if err == nil && config != nil {
		consensus = config.SharderConsensous
	}
	if len(sharderArray) < consensus {
		consensus = len(sharderArray)
	}
	Sharders = node.NewHolder(sharderArray, consensus)
}

func SetMiners(minerArray []string) {
	chain.Miners = minerArray
}

func SetChainID(id string) {
	chain.ChainID = id
}

// SetMaxTxnQuery set max transaction query, maximum number of trials to query a transaction confirmation from sharders.
//   - num is the number of transaction query
func SetMaxTxnQuery(num int) {
	chain.MaxTxnQuery = num
}

// SetQuerySleepTime set query sleep time, number of seconds to sleep between each transaction query.
//   - time is the sleep time
func SetQuerySleepTime(time int) {
	chain.QuerySleepTime = time
}

// SetMinSubmit set minimum submit, minimum number of miners to submit a transaction
//   - minSubmit is the minimum submit
func SetMinSubmit(minSubmit int) {
	chain.MinSubmit = minSubmit
}

// SetMinConfirmation set minimum confirmation, minimum number of miners to confirm a transaction
//   - minConfirmation is the minimum confirmation
func SetMinConfirmation(minConfirmation int) {
	chain.MinConfirmation = minConfirmation
}
