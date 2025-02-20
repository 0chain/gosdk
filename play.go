package main

import (
	"github.com/0chain/gosdk/core/client"
	"sync"
	"time"
)

func main() {
	clientId := "90cbab94239ef9c74ec83f8e616fb921a5dd8b36214b1b514e5010b449e5f162"

	err := client.InitSDK("", "https://dev.zus.network/dns", "0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe", "bls0chain", 0, false)
	if err != nil {
		println(err.Error())
		return
	}

	size := 1000
	resChan := make(chan *client.GetBalanceResponse, size)

	var wg sync.WaitGroup
	for i := 0; i < size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := client.GetBalance(clientId)
			if err != nil {
				println(i, err.Error())
				return
			}
			resChan <- res
		}()
	}

	wg.Wait()

	for i := 0; i < size; i++ {
		res := <-resChan
		println(i, res.Balance)
	}

}

// Timestamp represents Unix time (e.g. in seconds)
type Timestamp int64

// Now - current datetime
func Now() Timestamp {
	return Timestamp(time.Now().Unix())
}
