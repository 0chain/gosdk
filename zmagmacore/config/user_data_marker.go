package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type (
	// UserDataMarker represents config used for registration UserDataMarker
	UserDataMarker struct {
		UserID     string    `yaml:"user_id"`
		ProviderID string    `yaml:"provider_id"`
		SessionID  string    `yaml:"session_id"`
		DataUsage  DataUsage `yaml:"data_usage"`
		Qos        QoS       `yaml:"qos"`
	}

	// DataUsage represents config session data sage implementation.
	DataUsage struct {
		DownloadBytes uint64 `yaml:"download_bytes"`
		UploadBytes   uint64 `yaml:"upload_bytes"`
		SessionID     string `yaml:"session_id"`
		SessionTime   uint32 `yaml:"session_time"`
	}
)

// Read reads config yaml file from path.
func (u *UserDataMarker) Read(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(f *os.File) { _ = f.Close() }(f)

	decoder := yaml.NewDecoder(f)

	return decoder.Decode(u)
}
