// pakcage conf provide config helpers for ~/.zcn/config.yaml, ～/.zcn/network.yaml and ~/.zcn/wallet.json

package conf

import (
	"errors"
)

var (
	// ErrMssingConfig config file is missing
	ErrMssingConfig = errors.New("[conf]missing config file")
	// ErrInvalidValue invalid value in config
	ErrInvalidValue = errors.New("[conf]invalid value")
	// ErrBadParsing fail to parse config via spf13/viper
	ErrBadParsing = errors.New("[conf]bad parsing")
)

// Reader a config reader
type Reader interface {
	GetString(key string) string
	GetInt(key string) int
	GetStringSlice(key string) []string
}
