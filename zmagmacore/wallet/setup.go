package wallet

import (
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/zcncore"
)

// SetupZCNSDK runs zcncore.SetLogFile, zcncore.SetLogLevel and zcncore.InitZCNSDK using provided Config.
//
// If an error occurs during execution, the program terminates with code 2 and the error will be written in os.Stderr.
//
// SetupZCNSDK should be used only once while application is starting.
func SetupZCNSDK(cfg Config) error {
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
