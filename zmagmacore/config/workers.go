package config

type (
	// BalanceWorker represents worker options described in "workers.balance" section of the config yaml file.
	BalanceWorker struct {
		WaitResponseTimeout int64 `yaml:"wait_response_timeout"` // in seconds
		ScrapingTime        int64 `yaml:"scraping_time"`         // in seconds
	}
)
