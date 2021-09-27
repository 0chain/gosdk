//go:build wasm
// +build wasm

package wasm

import (
	"fmt"
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"

	"net/http/httptest"
)

var validMnemonic = "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"

var walletConfig = "{\"client_id\":\"9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85\",\"client_key\":\"40cd10039913ceabacf05a7c60e1ad69bb2964987bc50f77495e514dc451f907c3d8ebcdab20eedde9c8f39b9a1d66609a637352f318552fb69d4b3672516d1a\",\"keys\":[{\"public_key\":\"40cd10039913ceabacf05a7c60e1ad69bb2964987bc50f77495e514dc451f907c3d8ebcdab20eedde9c8f39b9a1d66609a637352f318552fb69d4b3672516d1a\",\"private_key\":\"a3a88aad5d89cec28c6e37c2925560ce160ac14d2cdcf4a4654b2bb358fe7514\"}],\"mnemonics\":\"inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown\",\"version\":\"1.0\",\"date_created\":\"2021-05-21 17:32:29.484657 +0545 +0545 m=+0.072791323\"}"

var storageConfig = fmt.Sprintf("{\"wallet\":%s,\"signature_scheme\":\"bls0chain\"}", walletConfig)

var server *httptest.Server
var miner *httptest.Server
var sharder *httptest.Server

func TestWasmWallet(t *testing.T) {
	Logger.Info("Testing WASM version of Zcncore.Wallet")

	TestAllConfig(t)

	t.Run("Test Network", func(t *testing.T) {
		getNetworkJSON := js.FuncOf(GetNetworkJSON)
		defer getNetworkJSON.Release()

		network := fmt.Sprintf("{\"miners\":[%#v],\"sharders\":[%#v]}", miner.URL+"/miner01", sharder.URL+"/sharder01")

		networkJson := getNetworkJSON.Invoke().String()

		assert.Equal(t, network, networkJson)

		var miner_dummy = js.Global().Call("eval", fmt.Sprintf("({minerString: %#v})", miner.URL+"/miner02"))

		var sharders_dummy = js.Global().Call("eval", fmt.Sprintf("({shardersArray: [%#v, %#v]})", sharder.URL+"/sharder02", sharder.URL+"/sharder03"))

		setNetwork := js.FuncOf(SetWalletNetwork)
		defer setNetwork.Release()

		miners := miner_dummy.Get("minerString")
		sharders := sharders_dummy.Get("shardersArray")

		assert.Empty(t, setNetwork.Invoke(miners, sharders).Truthy())

		newNetwork := fmt.Sprintf("{\"miners\":[%#v],\"sharders\":[%#v,%#v]}", miner.URL+"/miner02", sharder.URL+"/sharder02", sharder.URL+"/sharder03")

		newNetworkJson := getNetworkJSON.Invoke().String()

		assert.Equal(t, newNetwork, newNetworkJson)
	})

	t.Run("Test Mnemonic", func(t *testing.T) {
		isMnemonicValid := js.FuncOf(IsMnemonicValid)
		defer isMnemonicValid.Release()

		assert.Equal(t, true, isMnemonicValid.Invoke(validMnemonic).Bool())
	})

	t.Run("Test Get Version", func(t *testing.T) {
		getVersion := js.FuncOf(GetVersion)
		defer getVersion.Release()

		assert.Equal(t, "v1.3.0", getVersion.Invoke().String())
	})

	t.Run("Test Set Auth URL", func(t *testing.T) {
		setAuthUrl := js.FuncOf(SetAuthUrl)
		defer setAuthUrl.Release()

		assert.Equal(t, true, setAuthUrl.Invoke("miner/miner").IsNull())
	})

	t.Run("Test Split Keys", func(t *testing.T) {
		splitKeys := js.FuncOf(SplitKeys)
		defer splitKeys.Release()

		assert.NotEmpty(t, splitKeys.Invoke("a3a88aad5d89cec28c6e37c2925560ce160ac14d2cdcf4a4654b2bb358fe7514", "2").String())
	})

	t.Run("Test Conversion", func(t *testing.T) {
		token := "100"
		ctv := js.FuncOf(ConvertToValue)
		defer ctv.Release()

		assert.Equal(t, 1000000000000, ctv.Invoke(token).Int())

		val := ctv.Invoke(token).Int()
		ctt := js.FuncOf(ConvertToToken)
		defer ctt.Release()

		assert.Equal(t, float64(100), ctt.Invoke(fmt.Sprintf("%d", val)).Float())
	})

	t.Run("Test Encryption", func(t *testing.T) {
		key := "0123456789abcdef"
		var message string = "Lorem ipsum dolor sit amet"

		enc := js.FuncOf(Encrypt)
		defer enc.Release()

		emsg := enc.Invoke(key, message)

		dec := js.FuncOf(Decrypt)
		defer dec.Release()

		dmsg := dec.Invoke(key, emsg.String())

		assert.Equal(t, message, dmsg.String(), "The two message should be the same.")
	})

	t.Run("Test Get Min Sharder Verify", func(t *testing.T) {
		getMinSharders := js.FuncOf(GetMinShardersVerify)
		defer getMinSharders.Release()
		result, err := await(getMinSharders.Invoke())

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, 1, result[0].Int())
	})
}
