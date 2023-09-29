package transaction

import (
	"sync"

	"github.com/0chain/gosdk/core/util"
)

var Cache *NonceCache
var once sync.Once

type NonceCache struct {
	cache    map[string]int64
	guard    sync.Mutex
	sharders *util.NodeHolder
}

func InitCache(sharders *util.NodeHolder) {
	Cache.sharders = sharders
}

func init() {
	once.Do(func() {
		Cache = &NonceCache{
			cache: make(map[string]int64),
		}
	})
}

func (nc *NonceCache) GetNextNonce(clientId string) int64 {
	nc.guard.Lock()
	defer nc.guard.Unlock()
	if _, ok := nc.cache[clientId]; !ok {
		nonce, _, err := nc.sharders.GetNonceFromSharders(clientId)
		if err != nil {
			nonce = 0
		}
		nc.cache[clientId] = nonce
	}

	nc.cache[clientId] += 1
	return nc.cache[clientId]
}

func (nc *NonceCache) Set(clientId string, nonce int64) {
	nc.guard.Lock()
	defer nc.guard.Unlock()
	nc.cache[clientId] = nonce
}

func (nc *NonceCache) Evict(clientId string) {
	nc.guard.Lock()
	defer nc.guard.Unlock()
	delete(nc.cache, clientId)
}
