package conf

// ZConfig settings from ~/.zcn/config.yaml
// block_worker: http://198.18.0.98:9091
// signature_scheme: bls0chain
// min_submit: 50
// min_confirmation: 50
// confirmation_chain_length: 3
// max_txn_query: 5
// query_sleep_time: 5
// # # OPTIONAL - Uncomment to use/ Add more if you want
// # preferred_blobbers:
// #   - http://one.devnet-0chain.net:31051
// #   - http://one.devnet-0chain.net:31052
// #   - http://one.devnet-0chain.net:31053
type ZConfig struct {
	// BlockWorker the url of 0dns's network api
	BlockWorker string
	// PreferredBlobbers preferred blobbers on new allocation
	PreferredBlobbers []string

	// MinSubmit mininal submit from blobber
	MinSubmit int
	// MinConfirmation mininal confirmation from sharders
	MinConfirmation int
	// CconfirmationChainLength minial confirmation chain length
	ConfirmationChainLength int

	// additional settings depending network latency
	// MaxTxnQuery maximum transcation query from sharders
	MaxTxnQuery int
	// QuerySleepTime sleep time before transcation query
	QuerySleepTime int
}
