//go:build wasm
// +build wasm

package wasm

import (
	"fmt"
	"syscall/js"
	"testing"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/stretchr/testify/assert"
)

func TestWasmSDK(t *testing.T) {
	Logger.Info("Testing WASM SDK")

	TestAllConfig(t)

	var miner_dummy = js.Global().Call("eval", fmt.Sprintf("({minerString: %#v})", miner.URL+"/miner02"))

	var sharders_dummy = js.Global().Call("eval", fmt.Sprintf("({shardersArray: [%#v, %#v]})", sharder.URL+"/sharder02", sharder.URL+"/sharder03"))

	t.Run("Test SDK SetNetwork", func(t *testing.T) {
		setNetwork := js.FuncOf(SetNetwork)
		defer setNetwork.Release()

		miners := miner_dummy.Get("minerString")
		sharders := sharders_dummy.Get("shardersArray")

		assert.Empty(t, setNetwork.Invoke(miners, sharders).Truthy())
		assert.Equal(t, blockchain.GetMiners()[0], miner.URL+"/miner02")
		assert.Equal(t, blockchain.GetSharders()[0], sharder.URL+"/sharder02")
		assert.Equal(t, blockchain.GetSharders()[1], sharder.URL+"/sharder03")
	})

	t.Run("Test Get Network", func(t *testing.T) {
		getNetwork := js.FuncOf(GetNetwork)
		defer getNetwork.Release()
		res := getNetwork.Invoke()

		assert.Equal(t, res.Get("miners").String(), miner.URL+"/miner02")
		assert.Equal(t, res.Get("sharders").String(), sharder.URL+"/sharder02,"+sharder.URL+"/sharder03")
	})

	t.Run("Test Set Max Txn Query", func(t *testing.T) {
		assert.Equal(t, 5, blockchain.GetMaxTxnQuery())

		setMaxTxnQuery := js.FuncOf(SetMaxTxnQuery)
		defer setMaxTxnQuery.Release()

		assert.Empty(t, setMaxTxnQuery.Invoke("1").Truthy())
		assert.Equal(t, 1, blockchain.GetMaxTxnQuery())
	})

	t.Run("Test Set Query Sleep Time", func(t *testing.T) {
		assert.Equal(t, 5, blockchain.GetQuerySleepTime())

		setQuerySleepTime := js.FuncOf(SetQuerySleepTime)
		defer setQuerySleepTime.Release()

		assert.Empty(t, setQuerySleepTime.Invoke("1").Truthy())
		assert.Equal(t, 1, blockchain.GetQuerySleepTime())
	})

	t.Run("Test Set Min Submit", func(t *testing.T) {
		assert.Equal(t, 50, blockchain.GetMinSubmit())

		setMinSubmit := js.FuncOf(SetMinSubmit)
		defer setMinSubmit.Release()

		assert.Empty(t, setMinSubmit.Invoke("2").Truthy())
		assert.Equal(t, 2, blockchain.GetMinSubmit())
	})

	t.Run("Test Set Min Confirmation", func(t *testing.T) {
		assert.Equal(t, 50, blockchain.GetMinConfirmation())

		setMinConfirmation := js.FuncOf(SetMinConfirmation)
		defer setMinConfirmation.Release()

		assert.Empty(t, setMinConfirmation.Invoke("2").Truthy())
		assert.Equal(t, 2, blockchain.GetMinConfirmation())
	})
}
