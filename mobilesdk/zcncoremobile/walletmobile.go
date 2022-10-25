package zcncoremobile

import (
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zcncore"
)

// IsMnemonicValid is an utility function to check the mnemonic valid
func IsMnemonicValid(mnemonic string) bool {
	return zcncrypto.IsMnemonicValid(mnemonic)
}

// ConvertToToken converts the value to ZCN tokens
func ConvertToToken(value int64) float64 {
	return float64(value) / float64(zcncore.TOKEN_UNIT)
}

// GetVersion - returns version string
func GetVersion() string {
	return version.VERSIONSTR
}

// SetLogLevel set the log level.
// lvl - 0 disabled; higher number (upto 4) more verbosity
func SetLogLevel(lvl int) {
	zcncore.Logger.SetLevel(lvl)
}

func Init(chainConfigJSON string) error {
	return zcncore.Init(chainConfigJSON)
}
