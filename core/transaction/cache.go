package transaction

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/util"
)

const GET_BALANCE = `/v1/client/get/balance?client_id=`
const consensusThresh = float32(25.0)

//var defaultLogLevel = logger.DEBUG
//var Logger logger.Logger
var Cache *NonceCache
var once sync.Once

type NonceCache struct {
	cache    map[string]int64
	guard    sync.Mutex
	sharders []string
}

func InitCache(sharders []string) {
	Cache.sharders = sharders
}

func init() {
	//Logger.Init(defaultLogLevel, "0chain-core-cache")
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
		nonce, _, err := nc.getNonceFromSharders(clientId)
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

func queryFromSharders(sharders []string, query string,
	result chan *util.GetResponse) {

	queryFromShardersContext(context.Background(), sharders, query, result)
}

func (nc *NonceCache) getNonceFromSharders(clientID string) (int64, string, error) {
	return getBalanceFieldFromSharders(clientID, "nonce", nc.sharders)
}

func getBalanceFieldFromSharders(clientID, name string, sharders []string) (int64, string, error) {
	result := make(chan *util.GetResponse, len(sharders))
	// getMinShardersVerify
	var numSharders = len(sharders) // overwrite, use all
	queryFromSharders(sharders, fmt.Sprintf("%v%v", GET_BALANCE, clientID), result)
	consensus := float32(0)
	balMap := make(map[int64]float32)
	nonce := int64(0)
	var winInfo string
	waitTimeC := time.After(10 * time.Second)
	for i := 0; i < numSharders; i++ {
		select {
		case <- waitTimeC:
			return 0, "", stdErrors.New("get balance failed. consensus not reached")
		case rsp := <-result:
				if rsp.StatusCode != http.StatusOK {
					continue
				}

				var objmap map[string]json.RawMessage
				err := json.Unmarshal([]byte(rsp.Body), &objmap)
				if err != nil {
					continue
				}
				if v, ok := objmap[name]; ok {
					bal, err := strconv.ParseInt(string(v), 10, 64)
					if err != nil {
						continue
					}
					balMap[bal]++
					if balMap[bal] > consensus {
						consensus = balMap[bal]
						nonce = bal
						winInfo = rsp.Body

						rate := consensus * 100 / float32(len(sharders))
						if rate >= consensusThresh {
							return nonce, winInfo, nil
						}
					}
				}
		}
	}

	return 0, "", stdErrors.New("get balance failed, consensus not reached")
}

func queryFromShardersContext(ctx context.Context, sharders []string,
	query string, result chan *util.GetResponse) {

	for _, sharder := range util.Shuffle(sharders) {
		go func(sharderurl string) {
			//Logger.Info("Query from ", sharderurl+query)
			url := fmt.Sprintf("%v%v", sharderurl, query)
			req, err := util.NewHTTPGetRequestContext(ctx, url)
			if err != nil {
				//Logger.Error(sharderurl, " new get request failed. ", err.Error())
				return
			}
			res, err := req.Get()
			if err != nil {
				//Logger.Error(sharderurl, " get error. ", err.Error())
			}
			result <- res
		}(sharder)
	}
}
