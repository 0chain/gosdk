package sdk

import (
	"errors"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/zcncrypto"
	_ "github.com/0chain/gosdk/zboxcore/client" //import it to initialize sys.Sign
)

var ErrInvalidSignatureScheme = errors.New("invalid_signature_scheme")

// SignRequest sign data with private key and scheme
func SignRequest(privateKey, signatureScheme string, data string) (string, error) {
	hash := encryption.Hash(data)
	return sys.Sign(hash, signatureScheme, []sys.KeyPair{{
		PrivateKey: privateKey,
	}})
}

// VerifySignature verify signature with public key, schema and data
func VerifySignature(publicKey, signatureScheme string, data string, signature string) (bool, error) {

	hash := encryption.Hash(data)

	signScheme := zcncrypto.NewSignatureScheme(signatureScheme)
	if signScheme != nil {
		err := signScheme.SetPublicKey(publicKey)
		if err != nil {
			return false, err
		}
		return signScheme.Verify(signature, hash)
	}
	return false, ErrInvalidSignatureScheme
}
