package wasm

import (
	// "encoding/json"
	"fmt"
	"strings"
	"syscall/js"
	"testing"

	"github.com/0chain/gosdk/zboxcore/blockchain"
)

var storageConfig = fmt.Sprintf("{\"wallet\":%s,\"signature_scheme\":\"bls0chain\"}", walletConfig)

var chainConfig = "{\"chain_id\":\"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe\",\"block_worker\":\"http://127.0.0.1:1/dns\",\"miners\":[\"http://127.0.0.1:1/miner01\"],\"sharders\":[\"http://127.0.0.1:1/sharder01\"],\"signature_scheme\":\"bls0chain\",\"min_submit\":50,\"min_confirmation\":50,\"confirmation_chain_length\":3,\"eth_node\":\"https://ropsten.infura.io/v3/f0a254d8d18b4749bd8540da63b3292b\",\"num_keys\":\"1\"}"

func TestInitializeConfig(t *testing.T) {
	setup(t)
	setNetwork := js.FuncOf(ZBOXSetNetwork)
	defer setNetwork.Release()

	setNetwork.Invoke(miner.URL+"/miner01", sharder.URL+"/sharder01")

	if got := blockchain.GetMiners(); len(got) != 1 {
		t.Errorf("got %#v, want 1", strings.Join(got, ","))
	}

	if got := blockchain.GetSharders(); len(got) != 1 {
		t.Errorf("got %#v, want 1", strings.Join(got, ","))
	}

	var initConfig = fmt.Sprintf("{\"port\":31082,\"chain_id\":\"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe\",\"deployment_mode\":0,\"signature_scheme\":\"bls0chain\",\"block_worker\":\"%s\",\"cleanup_worker\":10}", server.URL+"/dns")

	initCfg := js.FuncOf(InitializeConfig)
	defer initCfg.Release()

	if got := initCfg.Invoke(initConfig); !got.IsNull() {
		t.Errorf("got %#v, want nil", got.Get("error").String())
	}
}

func TestInitStorageSDK(t *testing.T) {
	TestInitializeConfig(t)

	initStorageSDK := js.FuncOf(InitStorageSDK)
	defer initStorageSDK.Release()

	initStorageSDK.Invoke(storageConfig, chainConfig)

	if got := blockchain.GetBlockWorker(); got == "" {
		t.Errorf("got %#v, want %#v", got, "")
	}
}

func TestZBOXNetwork(t *testing.T) {
	TestInitStorageSDK(t)
	defer server.Close()

	getNetwork := js.FuncOf(GetNetwork)
	defer getNetwork.Release()

	if got := getNetwork.Invoke().String(); got != fmt.Sprintf("{\"miners\":[%#v],\"sharders\":[%#v]}", miner.URL+"/miner01", sharder.URL+"/sharder01") {
		t.Errorf("got %#v, want %#v", got, fmt.Sprintf("{\"miners\":[%#v],\"sharders\":[%#v]}", miner.URL+"/miner01", sharder.URL+"/sharder01"))
	}

	setNetwork := js.FuncOf(ZBOXSetNetwork)
	defer setNetwork.Release()

	setNetwork.Invoke(miner.URL+"/miner03", sharder.URL+"/sharder03")

	// We call getNetwork again to test ZBOXSetNetwork function
	if got := getNetwork.Invoke().String(); got != fmt.Sprintf("{\"miners\":[%#v],\"sharders\":[%#v]}", miner.URL+"/miner03", sharder.URL+"/sharder03") {
		t.Errorf("got %#v, want %#v", got, fmt.Sprintf("{\"miners\":[%#v],\"sharders\":[%#v]}", miner.URL+"/miner03", sharder.URL+"/sharder03"))
	}
}

func TestSetMaxTxnQuery(t *testing.T) {
	TestInitStorageSDK(t)
	defer server.Close()

	if got := blockchain.GetMaxTxnQuery(); got != 5 {
		t.Errorf("got %#v, want %#v", got, 5)
	}

	setMaxTxnQuery := js.FuncOf(SetMaxTxnQuery)
	defer setMaxTxnQuery.Release()

	setMaxTxnQuery.Invoke("100")

	if got := blockchain.GetMaxTxnQuery(); got == 5 {
		t.Errorf("got %#v, want %#v", got, 100)
	}
}

func TestSetQuerySleepTime(t *testing.T) {
	TestInitStorageSDK(t)
	defer server.Close()

	if got := blockchain.GetQuerySleepTime(); got != 5 {
		t.Errorf("got %#v, want %#v", got, 5)
	}

	setQuerySleepTime := js.FuncOf(SetQuerySleepTime)
	defer setQuerySleepTime.Release()

	setQuerySleepTime.Invoke("100")

	if got := blockchain.GetQuerySleepTime(); got == 5 {
		t.Errorf("got %#v, want %#v", got, 100)
	}
}

func TestSetMinSubmit(t *testing.T) {
	TestInitStorageSDK(t)
	defer server.Close()

	if got := blockchain.GetMinSubmit(); got != 50 {
		t.Errorf("got %#v, want %#v", got, 50)
	}

	setMinSubmit := js.FuncOf(SetMinSubmit)
	defer setMinSubmit.Release()

	setMinSubmit.Invoke("100")

	if got := blockchain.GetMinSubmit(); got == 50 {
		t.Errorf("got %#v, want %#v", got, 100)
	}
}

func TestSetMinConfirmation(t *testing.T) {
	TestInitStorageSDK(t)
	defer server.Close()

	if got := blockchain.GetMinConfirmation(); got != 50 {
		t.Errorf("got %#v, want %#v", got, 50)
	}

	setMinConfirmation := js.FuncOf(SetMinConfirmation)
	defer setMinConfirmation.Release()

	setMinConfirmation.Invoke("100")

	if got := blockchain.GetMinConfirmation(); got == 50 {
		t.Errorf("got %#v, want %#v", got, 100)
	}
}

func TestZBOXSetNetwork(t *testing.T) {
	TestInitStorageSDK(t)
	defer server.Close()

	setNetwork := js.FuncOf(SetNetwork)
	defer setNetwork.Release()

	setNetwork.Invoke(miner.URL+"/miner03", sharder.URL+"/sharder03")

}

func TestCreateReadPool(t *testing.T) {
	TestInitializeConfig(t)

	initStorageSDK := js.FuncOf(InitStorageSDK)
	defer initStorageSDK.Release()

	initStorageSDK.Invoke(storageConfig, chainConfig)

	createReadPool := js.FuncOf(CreateReadPool)
	defer createReadPool.Release()

	if got := createReadPool.Invoke(); !got.Get("Promise").IsUndefined() {
		t.Errorf("CreateReadPool failed")
	}
}
