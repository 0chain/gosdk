package conf

// ZNetwork settings from ~/.zcn/network.yaml
type ZNetwork struct {
	// Sharders sharder list of blockchain
	Sharders []string
	// Miners miner list of blockchain
	Miners []string
}
