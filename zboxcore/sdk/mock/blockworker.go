package mock

import (
	"encoding/json"
	"net/http"
	"testing"
)

const (
	NETWORK_PATH = "/network"
)

type blockWorker struct {
	Miners   []string
	Sharders []string
	mapHandler map[string]http.Handler
}

var bw *blockWorker

func (bw *blockWorker) getMapHandler() map[string]http.Handler {
	if bw.mapHandler == nil {
		bw.mapHandler = map[string]http.Handler{
			NETWORK_PATH: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				miners := bw.Miners
				if miners == nil {
					miners = []string{}
				}
				sharders := bw.Sharders
				if sharders == nil {
					sharders = []string{}
				}
				var netWorkResp = struct {
					Miners   []string `json:"miners"`
					Sharders []string `json:"sharders"`
				}{
					Miners:   miners,
					Sharders: sharders,
				}

				b, _ := json.Marshal(&netWorkResp)

				w.WriteHeader(200)
				w.Write(b)
			}),
		}
	}
	return bw.mapHandler
}

func SetBlockWorkerHandler(t *testing.T, path string, handler http.HandlerFunc) {
	if s == nil {
		t.Error("block worker http server is not initialized")
	}
	bw.mapHandler[path] = handler
}

func NewBlockWorkerHTTPServer(t *testing.T,miners, sharders []string) (url string, close func()) {
	bw = &blockWorker{Miners: miners, Sharders: sharders}
	return NewHTTPServer(t, bw.getMapHandler())
}
