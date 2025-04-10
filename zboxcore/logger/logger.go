// Proxy for the core logger package.
package logger

import (
	"github.com/0chain/gosdk/core/logger"
)

// Logger global logger instance
var Logger = logger.GetLogger()
