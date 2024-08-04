// Proxy for the core logger package.
package logger

import "github.com/0chain/gosdk/core/logger"

var defaultLogLevel = logger.DEBUG
var Logger logger.Logger

func init() {
	Logger.Init(defaultLogLevel, "0box-sdk")
}
