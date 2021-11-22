package config

type (
	// ChainConfig represents config interface used for setup wallet.
	ChainConfig interface {
		// LogDir returns directory to store logs.
		LogDir() string

		// LogLvl returns level of logs.
		LogLvl() string

		// BlockWorker returns address of dns server.
		BlockWorker() string

		// SignatureScheme returns signature scheme.
		SignatureScheme() string
	}
)
