package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type (
	// Provider represents configs of the providers node.
	Provider struct {
		ID    string        `yaml:"id"`
		ExtID string        `yaml:"ext_id"`
		Host  string        `yaml:"host"`
		Terms ProviderTerms `yaml:"terms"`
	}

	// ProviderTerms represents config of provider and services terms.
	ProviderTerms struct {
		Price           float32        `yaml:"price"`             // tokens per Megabyte
		PriceAutoUpdate float32        `yaml:"price_auto_update"` // price change on auto update
		MinCost         float32        `yaml:"min_cost"`          // minimal cost for a session
		Volume          int64          `yaml:"volume"`            // bytes per a session
		QoS             *QoS           `yaml:"qos"`               // qos of service
		QoSAutoUpdate   *QoSAutoUpdate `yaml:"qos_auto_update"`   // qos change on auto update
		ProlongDuration int64          `yaml:"prolong_duration"`  // duration in seconds to prolong the terms
		ExpiredAt       int64          `yaml:"expired_at"`        // timestamp till a session valid
	}

	// QoSAutoUpdate represents config of qos terms on auto update.
	QoSAutoUpdate struct {
		DownloadMbps float32 `yaml:"download_mbps"`
		UploadMbps   float32 `yaml:"upload_mbps"`
	}

	// QoS represents config of qos.
	QoS struct {
		DownloadMbps float32 `yaml:"download_mbps"`
		UploadMbps   float32 `yaml:"upload_mbps"`
	}
)

// Read reads config yaml file from path.
func (p *Provider) Read(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(f *os.File) { _ = f.Close() }(f)

	decoder := yaml.NewDecoder(f)

	return decoder.Decode(p)
}
