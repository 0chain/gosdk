package blockchain

import (
	"encoding/json"
	"sync/atomic"
)

type ChainConfig struct {
	BlockWorker       string
	Sharders          []string
	Miners            []string
	PreferredBlobbers []string
	MinSubmit         int
	MinConfirmation   int
	ChainID           string
	MaxTxnQuery       int
	QuerySleepTime    int
}

type StorageNode struct {
	ID      string `json:"id"`
	Baseurl string `json:"url"`

	skip uint64 // skip on error
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
	chain.Sharders, err = PopulateNodes(sharderjson)
	if err != nil {
		return err
	}
	return nil
}

func GetBlockWorker() string {
	return chain.BlockWorker
}

func GetSharders() []string {
	return chain.Sharders
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

func GetPreferredBlobbers() []string {
	return chain.PreferredBlobbers
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
	chain.Sharders = sharderArray
}

func SetMiners(minerArray []string) {
	chain.Miners = minerArray
}

func SetPreferredBlobbers(preferredBlobberArray []string) {
	chain.PreferredBlobbers = preferredBlobberArray
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
