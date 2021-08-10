package conf

// Network settings from ~/.zcn/network.yaml
type Network struct {
	// Sharders sharder list of blockchain
	Sharders []string
	// Miners miner list of blockchain
	Miners []string
}
