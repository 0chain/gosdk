// Proxy for the core logger package.
package logger

import "github.com/0chain/gosdk_common/core/logger"

var defaultLogLevel = logger.DEBUG

// Logger global logger instance
var Logger logger.Logger

func init() {
	Logger.Init(defaultLogLevel, "0box-sdk")
}
