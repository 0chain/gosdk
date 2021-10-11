package wasm

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"testing"

	"github.com/0chain/gosdk/wasm/httpwasm"
	"github.com/0chain/gosdk/zcncore"
	"github.com/stretchr/testify/assert"
)

var validMnemonic = "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"

var walletConfig = "{\"client_id\":\"9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85\",\"client_key\":\"40cd10039913ceabacf05a7c60e1ad69bb2964987bc50f77495e514dc451f907c3d8ebcdab20eedde9c8f39b9a1d66609a637352f318552fb69d4b3672516d1a\",\"keys\":[{\"public_key\":\"041eeb1b4eb9b2456799d8e2a566877e83bc5d76ff38b964bd4b7796f6a6ccae6f1966a4d91d362669fafa3d95526b132a6341e3dfff6447e0e76a07b3a7cfa6e8034574266b382b8e5174477ab8a32a49a57eda74895578031cd2d41fd0aef446046d6e633f5eb68a93013dfac1420bf7a1e1bf7a87476024478e97a1cc115de9\",\"private_key\":\"18c09c2639d7c8b3f26b273cdbfddf330c4f86c2ac3030a6b9a8533dc0c91f5e\"}],\"mnemonics\":\"inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown\",\"version\":\"1.0\",\"date_created\":\"2021-05-21 17:32:29.484657 +0545 +0545 m=+0.072791323\"}"

func TestWasmWallet(t *testing.T) {
	Logger.Info("Testing WASM version of Zcncore.Wallet")

	sharder := httpwasm.NewSharderServer()
	defer sharder.Close()
	miner := httpwasm.NewMinerServer()
	defer miner.Close()

	t.Run("Test Network", func(t *testing.T) {
		var miner_dummy = js.Global().Call("eval", fmt.Sprintf("({minerString: %#v})", miner.URL+"/miner02"))

		var sharders_dummy = js.Global().Call("eval", fmt.Sprintf("({shardersArray: [%#v, %#v]})", sharder.URL+"/sharder02", sharder.URL+"/sharder03"))

		setNetwork := js.FuncOf(SetWalletNetwork)
		defer setNetwork.Release()

		miners := miner_dummy.Get("minerString")
		sharders := sharders_dummy.Get("shardersArray")

		assert.Empty(t, setNetwork.Invoke(miners, sharders).Truthy())

		newNetwork := fmt.Sprintf("{\"miners\":[%#v],\"sharders\":[%#v,%#v]}", miner.URL+"/miner02", sharder.URL+"/sharder02", sharder.URL+"/sharder03")

		getNetworkJSON := js.FuncOf(GetNetworkJSON)
		defer getNetworkJSON.Release()
		newNetworkJson := getNetworkJSON.Invoke().String()

		assert.Equal(t, newNetwork, newNetworkJson)
	})

	t.Run("Test GetVersion", func(t *testing.T) {
		getVersion := js.FuncOf(GetVersion)
		defer getVersion.Release()

		assert.Equal(t, "v1.2.87", getVersion.Invoke().String())
	})

	t.Run("Test SetAuthURL", func(t *testing.T) {
		setAuthUrl := js.FuncOf(SetAuthUrl)
		defer setAuthUrl.Release()

		assert.Equal(t, true, setAuthUrl.Invoke("miner/miner").IsNull())
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

	t.Run("Test GetMinSharderVerify", func(t *testing.T) {
		getMinSharders := js.FuncOf(GetMinShardersVerify)
		defer getMinSharders.Release()
		result, err := await(getMinSharders.Invoke())

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, 1, result[0].Int())
	})

	t.Run("Test CreateWallet", func(t *testing.T) {
		createWallet := js.FuncOf(CreateWallet)
		defer createWallet.Release()

		result, err := await(createWallet.Invoke())

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, result[0].Get("status").Int())
		assert.Empty(t, result[0].Get("err").String())
		assert.Equal(t, true, result[0].Get("wallet").Truthy())

	})

	t.Run("Test RecoverWallet", func(t *testing.T) {
		recoverWallet := js.FuncOf(RecoverWallet)
		defer recoverWallet.Release()

		jsMnemonic := js.Global().Call("eval", fmt.Sprintf(`({mnemonic: %#v})`, validMnemonic))
		mnemonic := jsMnemonic.Get("mnemonic")

		result, err := await(recoverWallet.Invoke(mnemonic))

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, result[0].Get("status").Int())
		assert.Empty(t, result[0].Get("err").String())
		assert.Equal(t, true, result[0].Get("wallet").Truthy())
	})

	t.Run("Test SplitKeys", func(t *testing.T) {
		splitKeys := js.FuncOf(SplitKeys)
		defer splitKeys.Release()

		assert.NotEmpty(t, splitKeys.Invoke("a3a88aad5d89cec28c6e37c2925560ce160ac14d2cdcf4a4654b2bb358fe7514", "2").String())
	})

	t.Run("Test GetClientDetails", func(t *testing.T) {
		getClientDetails := js.FuncOf(GetClientDetails)
		defer getClientDetails.Release()

		jsClientDetails := js.Global().Call("eval", `({clientId: "9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85"})`)
		clientID := jsClientDetails.Get("clientId")

		result, err := await(getClientDetails.Invoke(clientID))

		aps := result[0].String()

		var clientResponse zcncore.GetClientResponse

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(aps), &clientResponse))
		assert.Equal(t, clientID.String(), clientResponse.ID)
	})

	t.Run("Test IsMnemonicValid", func(t *testing.T) {
		isMnemonicValid := js.FuncOf(IsMnemonicValid)
		defer isMnemonicValid.Release()

		assert.Equal(t, true, isMnemonicValid.Invoke(validMnemonic).Bool())
	})
}
