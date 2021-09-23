package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type (
	// Consumer represents config used for registration of node.
	Consumer struct {
		ID    string `yaml:"id"`
		ExtID string `yaml:"ext_id"`
		Host  string `yaml:"host"`
	}
)

// Read reads config yaml file from path.
func (c *Consumer) Read(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(f *os.File) { _ = f.Close() }(f)

	decoder := yaml.NewDecoder(f)

	return decoder.Decode(c)
}
