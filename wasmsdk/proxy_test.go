// go:build test
// +build test

package main

import (
	"fmt"
	"syscall/js"
	"testing"

	"github.com/0chain/gosdk/bls"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/stretchr/testify/assert"

	"github.com/0chain/gosdk/wasmsdk/httpwasm"
)

// var server *httptest.Server
// var sharder *httpwasm.Server
// var miner *httpwasm.Server

func TestAllConfig(t *testing.T) {
	Logger.Info("Setting Up All Configuration")

	sharder := httpwasm.NewSharderServer()
	defer sharder.Close()

	miner := httpwasm.NewMinerServer()
	defer miner.Close()

	blockchain.SetMiners([]string{miner.URL + "/miner01"})
	blockchain.SetSharders([]string{miner.URL + "/sharder01"})

	server := httpwasm.NewDefaultServer()
	defer server.Close()

	var chainConfig = fmt.Sprintf("{\"chain_id\":\"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe\",\"block_worker\":%#v,\"miners\":[%#v],\"sharders\":[%#v],\"signature_scheme\":\"bls0chain\",\"min_submit\":50,\"min_confirmation\":50,\"confirmation_chain_length\":3,\"eth_node\":\"\"}", server.URL+"/dns", miner.URL+"/miner01", sharder.URL+"/sharder01")

	var initConfig = fmt.Sprintf("{\"port\":31082,\"chain_id\":\"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe\",\"deployment_mode\":0,\"signature_scheme\":\"bls0chain\",\"block_worker\":\"%s\",\"cleanup_worker\":10,\"preferred_blobers\":[]}", server.URL+"/dns")

	var storageConfig = fmt.Sprintf("{\"wallet\":%s,\"signature_scheme\":\"bls0chain\"}", walletConfig)

	t.Run("Initialize Config", func(t *testing.T) {
		initCfg := js.FuncOf(InitializeConfig)
		defer initCfg.Release()
		res := initCfg.Invoke(initConfig)

		assert.Equal(t, res.IsNull(), true)
	})

	t.Run("Test InitZCNSDK", func(t *testing.T) {
		assert.NotEqual(t, 0, Configuration.BlockWorker, Configuration.ChainID, Configuration.SignatureScheme)

		initZCNSDK := js.FuncOf(InitZCNSDK)
		defer initZCNSDK.Release()

		result, err := await(initZCNSDK.Invoke(Configuration.BlockWorker, Configuration.SignatureScheme))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].Bool())
	})

	t.Run("Test InitStorageSDK", func(t *testing.T) {
		initStorageSDK := js.FuncOf(InitStorageSDK)
		defer initStorageSDK.Release()

		result, err := await(initStorageSDK.Invoke(storageConfig, chainConfig))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].Bool())

		assert.Equal(t, blockchain.GetBlockWorker(), server.URL+"/dns")
	})

	t.Run("Test SetWalletInfo", func(t *testing.T) {
		setWalletInfo := js.FuncOf(SetWalletInfo)
		defer setWalletInfo.Release()

		assert.Equal(t, true, setWalletInfo.Invoke(walletConfig, js.Global().Call("eval", "true")).IsNull())
	})
}

var verifyPublickey = `041eeb1b4eb9b2456799d8e2a566877e83bc5d76ff38b964bd4b7796f6a6ccae6f1966a4d91d362669fafa3d95526b132a6341e3dfff6447e0e76a07b3a7cfa6e8034574266b382b8e5174477ab8a32a49a57eda74895578031cd2d41fd0aef446046d6e633f5eb68a93013dfac1420bf7a1e1bf7a87476024478e97a1cc115de9`
var signPrivatekey = `18c09c2639d7c8b3f26b273cdbfddf330c4f86c2ac3030a6b9a8533dc0c91f5e`
var data = `TEST`

func TestSSSignAndVerify(t *testing.T) {
	signScheme := zcncrypto.NewSignatureScheme("bls0chain")
	signScheme.SetPrivateKey(signPrivatekey)
	hash := zcncrypto.Sha3Sum256(data)

	fmt.Println("hash", hash)
	fmt.Println("privkey", signScheme.GetPrivateKey())

	var sk bls.SecretKey
	sk.DeserializeHexStr(signScheme.GetPrivateKey())
	pk := sk.GetPublicKey()
	fmt.Println("pubkey", pk.ToString())

	signature, err := signScheme.Sign(hash)

	fmt.Println("signature", signature)

	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	verifyScheme := zcncrypto.NewSignatureScheme("bls0chain")
	verifyScheme.SetPublicKey(verifyPublickey)
	if ok, err := verifyScheme.Verify(signature, hash); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	}
}
