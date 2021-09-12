package wasm

import (
	"fmt"
	"testing"

	"github.com/0chain/gosdk/bls"
	"github.com/0chain/gosdk/core/zcncrypto"
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
