package wallet

type (
	// Config represents config interface used for setup wallet.
	Config interface {
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
