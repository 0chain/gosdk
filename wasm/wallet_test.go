package wasm_test

import (
	// "fmt"
	// "math"
	"encoding/json"
	"syscall/js"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/wasm"

	"github.com/0chain/gosdk/zcncore"
)

func TestGetNetworkJSON(t *testing.T) {
	// setup wallet
	w, err := zcncrypto.NewBLS0ChainScheme().GenerateKeys()
	if err != nil {
		t.Fatal(err)
	}
	wBlob, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	if err := zcncore.SetWalletInfo(string(wBlob), true); err != nil {
		t.Fatal(err)
	}

	// setup servers
	sharderServ := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
			},
		),
	)
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				n := zcncore.Network{Miners: []string{"miner 1"}, Sharders: []string{sharderServ.URL}}
				blob, err := json.Marshal(n)
				if err != nil {
					t.Fatal(err)
				}

				if _, err := w.Write(blob); err != nil {
					t.Fatal(err)
				}
			},
		),
	)

	if err := zcncore.InitZCNSDK(server.URL, "bls0chain"); err != nil {
		t.Fatal(err)
	}

	getNetworkJSON := js.FuncOf(wasm.GetNetworkJSON)
	defer getNetworkJSON.Release()
	if got := getNetworkJSON.Invoke().String(); got != "{\"miners\":[\"miner 1\"],\"sharders\":[\"http://127.0.0.1:1\"]}" {
		t.Errorf("got %#v, want %#v", got, "{\"miners\":[\"miner 1\"],\"sharders\":[\"http://127.0.0.1:1\"]}")
	}

	getVersion := js.FuncOf(wasm.GetVersion)
	if got := getVersion.Invoke().String(); got != "v1.3.0" {
		t.Errorf("got %#v, want %#v", got, "v1.3.0")
	}

	// setNetworkJSON := js.FuncOf(wasm.SetNetwork)
	// defer setNetworkJSON.Release()
	// newNetwork := []js.Value{
	// 	js.ValueOf("miner 2"),
	// 	js.ValueOf("http://127.0.0.1:1")}

	// setNetworkJSON.Invoke(newNetwork)
	// if got := getNetworkJSON.Invoke().String(); got != "{\"miners\":[\"miner 2\"],\"sharders\":[\"http://127.0.0.1:1\"]}" {
	// 	t.Errorf("got %#v, want %#v", got, "{\"miners\":[\"miner 2\"],\"sharders\":[\"http://127.0.0.1:1\"]}")
	// }
}
