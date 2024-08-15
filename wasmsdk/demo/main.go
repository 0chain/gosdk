// Demo for the wasm sdk- do not use.
package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zcncore"
	"github.com/uptrace/bunrouter"
)

func main() {

	zcncore.InitSignatureScheme("bls0chain")
	
	ctx, cf := context.WithCancel(context.Background())

	router := bunrouter.New()

	router.GET("/shutdown", func(w http.ResponseWriter, req bunrouter.Request) error {
		log.Println("wasm:shutdown")
		cf()
		return nil
	})

	// create a new wallet
	router.POST("/wallet", func(w http.ResponseWriter, req bunrouter.Request) error {

		wallet, err := zcncore.CreateWalletOffline()

		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(wallet))

		return nil
	})

	fileServer := http.FileServer(http.Dir("./"))

	router.GET("/*path", bunrouter.HTTPHandler(fileServer))

	go func() {
		log.Println("wasm: listening on http://127.0.0.1:8080")
		log.Println(http.ListenAndServe(":8080", router))

	}()

	<-ctx.Done()

}

type statusBar struct {
	walletString string
	wg           *sync.WaitGroup
	success      bool
	errMsg       string
	balance      common.Balance
	nonce        int64
	wallets      []string
	clientID     string
}
