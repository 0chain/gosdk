package wallet

import (
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/zcnbridge/config"
	"github.com/0chain/gosdk/zcnbridge/crypto"
	"github.com/0chain/gosdk/zcncore"
)

func Setup() (*Wallet, error) {
	err := setupZCNSDK(config.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup ZCNSDK")
	}

	publicKey, privateKey, err := crypto.ReadKeysFile(*config.Client.KeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize wallet keys")
	}

	wallet := CreateWallet(publicKey, privateKey)
	err = wallet.RegisterToMiners()
	if err != nil {
		return nil, errors.Wrap(err, "failed to register to miners")
	}

	return wallet, nil
}

// setupZCNSDK runs zcncore.SetLogFile, zcncore.SetLogLevel and zcncore.InitZCNSDK using provided Config.
//
// If an error occurs during execution, the program terminates with code 2 and the error will be written in os.Stderr.
//
// setupZCNSDK should be used only once while application is starting.
func setupZCNSDK(cfg Config) error {
	var logName = cfg.LogDir() + "/zsdk.log"
	zcncore.SetLogFile(logName, false)
	zcncore.SetLogLevel(logLevelFromStr(cfg.LogLvl()))
	return zcncore.InitZCNSDK(cfg.BlockWorker(), cfg.SignatureScheme())
}

// logLevelFromStr converts string log level to gosdk logger level int value.
func logLevelFromStr(level string) int {
	switch level {
	case "none":
		return logger.NONE
	case "fatal":
		return logger.FATAL
	case "error":
		return logger.ERROR
	case "info":
		return logger.INFO
	case "debug":
		return logger.DEBUG

	default:
		return -1
	}
}
