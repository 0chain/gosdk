package config

type (
	// Handler represents config options for handlers.
	Handler struct {
		RateLimit float64    `yaml:"rate_limit"` // per second
		Log       LogHandler `yaml:"log"`
	}

	// LogHandler represents config options described in "handler.log" section of the config yaml file.
	LogHandler struct {
		BufLength int64 `yaml:"buf_length"` // in kilobytes
	}
)
