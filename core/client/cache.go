package client

import (
	"go.uber.org/zap"
	"log"
	"sync"
)

var Cache *NonceCache
var once sync.Once

type NonceCache struct {
	cache    map[string]int64
	guard    sync.Mutex
	sharders *NodeHolder
}

func InitCache(sharders *NodeHolder) {
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
		bal, err := GetBalance(clientId)
		if err != nil || bal == nil {
			nc.cache[clientId] = 0
		} else {
			nc.cache[clientId] = bal.Nonce
		}

		log.Println("NonceCache.GetNextNonce: ", clientId, nc.cache[clientId], zap.Any("bal", bal), zap.Error(err))
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
