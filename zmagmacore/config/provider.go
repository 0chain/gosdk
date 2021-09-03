package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type (
	// Provider represents configs of the providers' node.
	Provider struct {
		ID       string `yaml:"id"`
		ExtID    string `yaml:"ext_id"`
		Host     string `yaml:"host"`
		MinStake int64  `yaml:"min_stake"`
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
