package zcncore

import (
	"sync"

	"github.com/0chain/gosdk/core/logger"
)

// Singleton
// TODO: Remove these variable and Use zchain.Gcontainer
var (
	_config localConfig
	logging logger.Logger
	stableMiners []string
	mGuard sync.Mutex
)