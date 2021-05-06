package main

import (
	"fmt"
	"testing"
	"github.com/0chain/gosdk/core/version"

	"github.com/0chain/gosdk/bls"
	// "github.com/0chain/gosdk/miracl"
	// "github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/zcncrypto"

	"github.com/0chain/gosdk/zboxcore/sdk"
	"syscall/js"
)

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

// Basic wasm function.
func addFunction(this js.Value, p []js.Value) interface{} {
	sum := p[0].Int() + p[1].Int()
	return js.ValueOf(sum)
}

// Ported from `code/go/0proxy.io/zproxycore/handler/wallet.go`
func GetClientEncryptedPublicKey(this js.Value, p []js.Value) interface{} {
	initSDK(p[0].String())
	key, err := sdk.GetClientEncryptedPublicKey()
	if err != nil {
		return js.ValueOf("get_public_encryption_key_failed: " + err.Error())
	}
	return js.ValueOf(key)
}

// Ported from `code/go/0proxy.io/zproxycore/handler/util.go`
func initSDK(clientJSON string) error {
	return sdk.InitStorageSDK(clientJSON,
		Configuration.BlockWorker,
		Configuration.ChainID,
		Configuration.SignatureScheme,
		nil)
}

// Ported from `code/go/0proxy.io/zproxycore/zproxy/main.go`
// TODO: should be passing in JSON. Better than a long arg list.
func initializeConfig(this js.Value, p []js.Value) interface{} {
	Configuration.ChainID = p[0].String()
	Configuration.SignatureScheme = p[1].String()
	Configuration.Port = p[2].Int()
	Configuration.BlockWorker = p[3].String()
	Configuration.CleanUpWorkerMinutes = p[4].Int()
	return nil
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/core/config/config.go`
//-----------------------------------------------------------------------------

/*Config - all the config options passed from the command line*/
type Config struct {
	Port                 int
	ChainID              string
	DeploymentMode       byte
	SignatureScheme      string
	BlockWorker          string
	CleanUpWorkerMinutes int
}

/*Configuration of the system */
var Configuration Config

//-----------------------------------------------------------------------------

func main() {
	// Ported over a basic unit test to make sure it runs in the browser.
	// TestSSSignAndVerify(new(testing.T))

	fmt.Printf("0CHAIN - GOSDK (version=%v)\n", version.VERSIONSTR)

	c := make(chan struct{}, 0)
	js.Global().Set("add", js.FuncOf(addFunction))
	js.Global().Set("GetClientEncryptedPublicKey", js.FuncOf(GetClientEncryptedPublicKey))
	js.Global().Set("initializeConfig", js.FuncOf(initializeConfig))
	<-c
}
