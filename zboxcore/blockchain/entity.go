package blockchain

import "encoding/json"

type ChainConfig struct {
	Sharders          []string
	Miners            []string
	PreferredBlobbers []string
	ChainID           string
	MaxTxnQuery       int
	QuerySleepTime    int
}

type StorageNode struct {
	ID      string `json:"id"`
	Baseurl string `json:"url"`
}

func PopulateNodes(nodesjson string) ([]string, error) {
	sharders := make([]string, 0)
	err := json.Unmarshal([]byte(nodesjson), &sharders)
	return sharders, err
}

var chain *ChainConfig

func init() {
	chain = &ChainConfig{
		MaxTxnQuery:    5,
		QuerySleepTime: 5,
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
