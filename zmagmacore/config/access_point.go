package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	// AccessPoint represents config used for access points of node.
	AccessPoint struct {
		ID            string `yaml:"id"`
		Terms         Terms  `yaml:"terms"`
		MinStake      bool   `yaml:"min_stake"`
		ProviderExtID string `yaml:"provider_ext_id"`
	}

	// Terms represents config of access point terms.
	Terms struct {
		Price           float32        `yaml:"price"`             // tokens per Megabyte
		PriceAutoUpdate float32        `yaml:"price_auto_update"` // price change on auto update
		MinCost         float32        `yaml:"min_cost"`          // minimal cost for a session
		Volume          int64          `yaml:"volume"`            // bytes per a session
		QoS             *QoS           `yaml:"qos"`               // qos of service
		QoSAutoUpdate   *QoSAutoUpdate `yaml:"qos_auto_update"`   // qos change on auto update
		ProlongDuration time.Duration  `yaml:"prolong_duration"`  // duration in seconds to prolong the terms
		ExpiredAt       time.Duration  `yaml:"expired_at"`        // time that will be added to the current time
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
func (p *AccessPoint) Read(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(f *os.File) { _ = f.Close() }(f)

	decoder := yaml.NewDecoder(f)

	return decoder.Decode(p)
}
