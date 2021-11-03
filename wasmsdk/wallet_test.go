// go:build test
// +build test

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"testing"
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/wasmsdk/httpwasm"
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

	t.Run("Test GetBalance", func(t *testing.T) {
		getBalance := js.FuncOf(GetBalance)
		defer getBalance.Release()

		result, err := await(getBalance.Invoke())

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, 1000, result[0].Get("value").Int())
		assert.Equal(t, `{"balance":1000}`, result[0].Get("info").String())
	})

	t.Run("Test GetBalanceWallet", func(t *testing.T) {
		getBalanceWallet := js.FuncOf(GetBalanceWallet)
		defer getBalanceWallet.Release()

		jsWallet := js.Global().Call("eval", fmt.Sprintf(`({wallet: %#v})`, walletConfig))
		wallet := jsWallet.Get("wallet")

		result, err := await(getBalanceWallet.Invoke(wallet))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, 1000, result[0].Get("value").Int())
		assert.Equal(t, `{"balance":1000}`, result[0].Get("info").String())
	})

	t.Run("Test GetLockConfig", func(t *testing.T) {
		getLockConfig := js.FuncOf(GetLockConfig)
		defer getLockConfig.Release()

		result, err := await(getLockConfig.Invoke())

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, 0, result[0].Get("op").Int())
		assert.Equal(t, `{"balance":1000}`, result[0].Get("info").String())
	})

	t.Run("Test GetLockedTokens", func(t *testing.T) {
		getLockedTokens := js.FuncOf(GetLockedTokens)
		defer getLockedTokens.Release()

		result, err := await(getLockedTokens.Invoke())

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, 1, result[0].Get("op").Int())
		assert.Equal(t, `{"balance":1000}`, result[0].Get("info").String())
	})

	t.Run("Test GetWallet", func(t *testing.T) {
		getWallet := js.FuncOf(GetWallet)
		defer getWallet.Release()

		jsWallet := js.Global().Call("eval", fmt.Sprintf(`({wallet: %#v})`, walletConfig))
		wallet := jsWallet.Get("wallet")

		result := getWallet.Invoke(wallet)

		var w zcncrypto.Wallet
		res := result.String()

		assert.Empty(t, json.Unmarshal([]byte(res), &w))
		assert.Equal(t, "40cd10039913ceabacf05a7c60e1ad69bb2964987bc50f77495e514dc451f907c3d8ebcdab20eedde9c8f39b9a1d66609a637352f318552fb69d4b3672516d1a", w.ClientKey)
	})

	t.Run("Test GetWalletClientID", func(t *testing.T) {
		getWalletClientID := js.FuncOf(GetWalletClientID)
		defer getWalletClientID.Release()

		jsWallet := js.Global().Call("eval", fmt.Sprintf(`({wallet: %#v})`, walletConfig))
		wallet := jsWallet.Get("wallet")

		result := getWalletClientID.Invoke(wallet).String()

		assert.Equal(t, "9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85", result)
	})

	t.Run("Test SetupAuth", func(t *testing.T) {
		setupAuth := js.FuncOf(SetupAuth)
		defer setupAuth.Release()

		server := httpwasm.NewDefaultServer()
		defer server.Close()

		jsAuth := js.Global().Call("eval", fmt.Sprintf(`({auth_host: %#v,client_id: "9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85", client_key: "40cd10039913ceabacf05a7c60e1ad69bb2964987bc50f77495e514dc451f907c3d8ebcdab20eedde9c8f39b9a1d66609a637352f318552fb69d4b3672516d1a", public_key: "041eeb1b4eb9b2456799d8e2a566877e83bc5d76ff38b964bd4b7796f6a6ccae6f1966a4d91d362669fafa3d95526b132a6341e3dfff6447e0e76a07b3a7cfa6e8034574266b382b8e5174477ab8a32a49a57eda74895578031cd2d41fd0aef446046d6e633f5eb68a93013dfac1420bf7a1e1bf7a87476024478e97a1cc115de9", private_key: "18c09c2639d7c8b3f26b273cdbfddf330c4f86c2ac3030a6b9a8533dc0c91f5e", peer_public_key: "041eeb1b4eb9b2456799d8e2a566877e83bc5d76ff38b964bd4b7796f6a6ccae6f1966a4d91d362669fafa3d95526b132a6341e3dfff6447e0e76a07b3a7cfa6e8034574266b382b8e5174477ab8a32a49a57eda74895578031cd2d41fd0aef446046d6e633f5eb68a93013dfac1420bf7a1e1bf7a87476024478e97a1cc115de9"})`, server.URL+"/"))

		authHost := jsAuth.Get("auth_host")
		clientID := jsAuth.Get("client_id")
		clientKey := jsAuth.Get("client_key")
		publicKey := jsAuth.Get("public_key")
		privateKey := jsAuth.Get("private_key")
		peerPublicKey := jsAuth.Get("peer_public_key")

		result, err := await(setupAuth.Invoke(authHost, clientID, clientKey, publicKey, privateKey, peerPublicKey))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, 0, result[0].Get("status").Int())
		assert.Empty(t, result[0].Get("err").String())
	})

	t.Run("Test GetIdForUrl", func(t *testing.T) {
		getIdForUrl := js.FuncOf(GetIdForUrl)
		defer getIdForUrl.Release()

		server := httpwasm.NewDefaultServer()
		defer server.Close()

		jsAuth := js.Global().Call("eval", fmt.Sprintf(`({url: %#v})`, server.URL+"/"))

		url := jsAuth.Get("url")

		result, _ := await(getIdForUrl.Invoke(url))

		assert.Equal(t, "9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85", result[0].String())
	})

	t.Run("Test GetVestingPoolInfo", func(t *testing.T) {
		getVestingPoolInfo := js.FuncOf(GetVestingPoolInfo)
		defer getVestingPoolInfo.Release()

		jsPoolID := js.Global().Call("eval", fmt.Sprintf(`({poolID: %#v})`, httpwasm.GetMockId(1)))
		PoolID := jsPoolID.Get("poolID")

		result, err := await(getVestingPoolInfo.Invoke(PoolID))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, 0, result[0].Get("op").Int())
		assert.Equal(t, httpwasm.GetMockId(1), result[0].Get("info").String())
	})

	t.Run("Test GetVestingClientList", func(t *testing.T) {
		getVestingClientList := js.FuncOf(GetVestingClientList)
		defer getVestingClientList.Release()
		jsClientID := js.Global().Call("eval", fmt.Sprintf(`({clientID: %#v})`, httpwasm.GetMockId(0)))
		clientID := jsClientID.Get("clientID")

		result, err := await(getVestingClientList.Invoke(clientID))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, 0, result[0].Get("op").Int())
		assert.Equal(t, httpwasm.GetMockId(0), result[0].Get("info").String())
	})

	t.Run("Test GetVestingSCConfig", func(t *testing.T) {
		getVestingSCConfig := js.FuncOf(GetVestingSCConfig)
		defer getVestingSCConfig.Release()

		scconfig := zcncore.VestingSCConfig{
			MinLock:              common.Balance(2000),
			MinDuration:          time.Duration(time.Hour),
			MaxDuration:          time.Duration(time.Hour * 48),
			MaxDestinations:      20,
			MaxDescriptionLength: 100,
		}

		result, err := await(getVestingSCConfig.Invoke())
		sc := result[0].Get("info").String()

		var resp zcncore.VestingSCConfig

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(sc), &resp))
		assert.Equal(t, 0, result[0].Get("op").Int())
		assert.Equal(t, scconfig, resp)
	})

	t.Run("Test GetMiners", func(t *testing.T) {
		getMiners := js.FuncOf(GetMiners)
		defer getMiners.Release()

		result, err := await(getMiners.Invoke())
		minerArr := result[0].Get("info").String()

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, 0, result[0].Get("op").Int())
		assert.Equal(t, "[\"127.0.0.1:1/miner01\",\"127.0.0.1:1/miner02\"]", minerArr)
	})

	t.Run("Test GetSharders", func(t *testing.T) {
		getSharders := js.FuncOf(GetSharders)
		defer getSharders.Release()

		result, err := await(getSharders.Invoke())
		sharderArr := result[0].Get("info").String()

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, 0, result[0].Get("op").Int())
		assert.Equal(t, "[\"127.0.0.1:1/sharder01\",\"127.0.0.1:1/sharder02\"]", sharderArr)
	})

	t.Run("Test GetMinerSCNodeInfo", func(t *testing.T) {
		getMinerSCNodeInfo := js.FuncOf(GetMinerSCNodeInfo)
		defer getMinerSCNodeInfo.Release()

		jsID := js.Global().Call("eval", fmt.Sprintf(`({id: %#v})`, httpwasm.GetMockId(100)))
		id := jsID.Get("id")

		result, err := await(getMinerSCNodeInfo.Invoke(id))
		msc := result[0].Get("info").String()

		var resp zcncore.MinerSCNodes

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(msc), &resp))
		assert.Equal(t, 0, result[0].Get("op").Int())
		assert.Equal(t, httpwasm.GetMockId(100), resp.Nodes[1].Miner.ID)
	})
}
