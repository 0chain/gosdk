package wasm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"syscall/js"
	"testing"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zcncore"
	"github.com/stretchr/testify/assert"
)

// Setting up mock server, miner and sharder
func setup(t *testing.T) {
	sharder = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
			},
		),
	)

	miner = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
			},
		),
	)

	server = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				n := zcncore.Network{Miners: []string{miner.URL + "/miner01"}, Sharders: []string{sharder.URL + "/sharder01"}}
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
}

func testSetWalletInfo(t *testing.T) {
	setWalletInfo := js.FuncOf(SetWalletInfo)
	defer setWalletInfo.Release()

	wi := setWalletInfo.Invoke(walletConfig, js.Global().Call("eval", "true"))

	assert.Equal(t, true, wi.IsNull())
}

func TestAllConfig(t *testing.T) {
	Logger.Info("Setting Up All Configuration")
	t.Run("Set Wallet Info", func(t *testing.T) { testSetWalletInfo(t) })
	t.Run("Setup Mock Server Miner And Sharder", func(t *testing.T) { setup(t) })
	defer server.Close()

	var chainConfig = fmt.Sprintf("{\"chain_id\":\"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe\",\"block_worker\":%#v,\"miners\":[%#v],\"sharders\":[%#v],\"signature_scheme\":\"bls0chain\",\"min_submit\":50,\"min_confirmation\":50,\"confirmation_chain_length\":3,\"eth_node\":\"\"}", server.URL+"/dns", miner.URL+"/miner01", sharder.URL+"/sharder01")

	var initConfig = fmt.Sprintf("{\"port\":31082,\"chain_id\":\"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe\",\"deployment_mode\":0,\"signature_scheme\":\"bls0chain\",\"block_worker\":\"%s\",\"cleanup_worker\":10}", server.URL+"/dns")

	t.Run("Initialize Config", func(t *testing.T) {
		initCfg := js.FuncOf(InitializeConfig)
		defer initCfg.Release()
		res := initCfg.Invoke(initConfig)

		assert.Equal(t, res.IsNull(), true)
	})

	t.Run("Test InitStorageSDK", func(t *testing.T) {
		initStorageSDK := js.FuncOf(InitStorageSDK)
		defer initStorageSDK.Release()

		result, err := await(initStorageSDK.Invoke(storageConfig, chainConfig))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].Truthy())
		assert.Equal(t, blockchain.GetBlockWorker(), server.URL+"/dns")
	})
}
