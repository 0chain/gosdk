package config

type (
	// ServerChain represents config options described in "server_chain" section of the config yaml file.
	// ServerChain must be a field of Config struct
	ServerChain struct {
		ID              string `yaml:"id"`
		OwnerID         string `yaml:"owner_id"`
		BlockWorker     string `yaml:"block_worker"`
		SignatureScheme string `yaml:"signature_scheme"`
	}
)
