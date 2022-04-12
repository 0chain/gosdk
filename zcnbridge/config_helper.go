package zcnbridge

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/spf13/viper"
)

func initBridgeConfig(sdkConfig *BridgeSDKConfig) *viper.Viper {
	cfg := readConfig(sdkConfig, func() string {
		return *sdkConfig.ConfigBridgeFile
	})

	log.Logger.Info(fmt.Sprintf("Bridge config has been initialized from %s", cfg.ConfigFileUsed()))

	return cfg
}

func readConfig(sdkConfig *BridgeSDKConfig, getConfigName func() string) *viper.Viper {
	cfg := viper.New()
	cfg.AddConfigPath(*sdkConfig.ConfigDir)
	cfg.SetConfigName(getConfigName())
	cfg.SetConfigType("yaml")
	err := cfg.ReadInConfig()
	if err != nil {
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "???"
			line = 0
		}
		f := filepath.Base(file)
		header := fmt.Sprintf("[ERROR] %s:%d: ", f, line)
		ExitWithError(header+"Can't read config: ", err)
	}
	return cfg
}

func GetConfigDir() string {
	var configDir string
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	configDir = home + "/.zcn"
	return configDir
}

func ExitWithError(v ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
