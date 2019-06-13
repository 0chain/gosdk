package blockchain

import "encoding/json"

type ChainConfig struct {
	Sharders []string
	Miners   []string
	ChainID  string
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
	chain = &ChainConfig{}
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

func SetSharders(sharderArray []string) {
	chain.Sharders = sharderArray
}

func SetMiners(minerArray []string) {
	chain.Miners = minerArray
}

func SetChainID(id string) {
	chain.ChainID = id
}
