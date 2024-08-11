// DEPRECATED: This package is deprecated and will be removed in a future release.
package config

type (
	// ServerChain represents config describes blockchain options and signature scheme options.
	ServerChain struct {
		ID              string `yaml:"id"`
		OwnerID         string `yaml:"owner_id"`
		BlockWorker     string `yaml:"block_worker"`
		SignatureScheme string `yaml:"signature_scheme"`
	}
)
