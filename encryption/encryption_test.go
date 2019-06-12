package encryption

import (
	"fmt"
	"testing"
)

var data = "0chain.net rocks"
var expectedHash = "6cb51770083ba34e046bc6c953f9f05b64e16a0956d4e496758b97c9cf5687d5"

func TestHash(t *testing.T) {
	if Hash(data) != expectedHash {
		fmt.Printf("invalid hash\n")
	} else {
		fmt.Printf("hash successful\n")
	}
}

func TestGenerateKeys(t *testing.T) {
	publicKey, privateKey, err := GenerateKeys()
	fmt.Printf("keys: %v,%v, %v\n", privateKey, publicKey, err)
}

func BenchmarkGenerateKeys(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateKeys()
	}
}

func TestSignAndVerify(t *testing.T) {
	// publicKey, privateKey, err := GenerateKeys()
	_, err := Sign("68e375b1e8018087e045ff9825826506d0a0f89e5863fd9be1f6156fadfde070150df93a0d6b5c5a865bce8aa0160290b3516ece0690d9fff3d4c8b29cc1aafd", "c9000e8c49f582e42b6197ff59161cc44314c078e59040aab88c4b5f10aaaea9")
	if err != nil {
		t.Fatalf("error signing: %v\n", err.Error())
	}
	if ok, err := Verify("150df93a0d6b5c5a865bce8aa0160290b3516ece0690d9fff3d4c8b29cc1aafd", "881a3f4e0c6abeb624b04ad39864d30ca3647d239f4954bc471aab22888eb1854e234ac1fb94fca50e9201b8960a60a714002906c7cde34c69a906629a395104", "c9000e8c49f582e42b6197ff59161cc44314c078e59040aab88c4b5f10aaaea9"); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	} else {
		fmt.Printf("Signing Verification successful\n")
	}
}

func BenchmarkSign(b *testing.B) {
	_, privateKey, err := GenerateKeys()
	if err == nil {
		return
	}
	for i := 0; i < b.N; i++ {
		Sign(privateKey, expectedHash)
	}
}

func BenchmarkVerify(b *testing.B) {
	publicKey, privateKey, err := GenerateKeys()
	signature, err := Sign(privateKey, expectedHash)
	if err != nil {
		return
	}
	for i := 0; i < b.N; i++ {
		Verify(publicKey, signature, expectedHash)
	}
}
