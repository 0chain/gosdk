package logger

import "0chain.net/clientsdk/core/logger"

var defaultLogLevel = logger.DEBUG
var Logger logger.Logger

func init() {
	Logger.Init(defaultLogLevel, "0box-sdk")
}
