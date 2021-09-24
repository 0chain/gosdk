package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// User represents config used for registration of user.
type User struct {
	ID         string `yaml:"id"`
	ConsumerID string `yaml:"consumer_id"`
}

// Read reads config yaml file from path.
func (u *User) Read(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(f *os.File) { _ = f.Close() }(f)

	decoder := yaml.NewDecoder(f)

	return decoder.Decode(u)
}
