package crypto

import (
	"bufio"
	"encoding/hex"
	"io"
	"os"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zmagmacore/errors"
)

// ReadKeysFile reads file existing in keysFile dir and parses public and private keys from file.
func ReadKeysFile(keysFile string) (publicKey, privateKey []byte, err error) {
	const errCode = "read_keys"

	reader, err := os.Open(keysFile)
	if err != nil {
		return nil, nil, errors.Wrap(errCode, "error while open keys file", err)
	}

	publicKeyHex, privateKeyHex := readKeys(reader)
	err = reader.Close()
	if err != nil {
		return nil, nil, errors.Wrap(errCode, "error while close keys file", err)
	}
	publicKey, err = hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, nil, errors.Wrap(errCode, "error while decoding public key", err)
	}
	privateKey, err = hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, nil, errors.Wrap(errCode, "error while decoding private key", err)
	}

	return publicKey, privateKey, nil
}

// readKeys reads a publicKey and a privateKey from a io.Reader passed in args.
// They are assumed to be in two separate lines one followed by the other.
func readKeys(reader io.Reader) (publicKey string, privateKey string) {
	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	publicKey = scanner.Text()
	scanner.Scan()
	privateKey = scanner.Text()
	scanner.Scan()

	return publicKey, privateKey
}

// Verify verifies passed signature of the passed hash with passed public key using the signature scheme.
func Verify(publicKey, signature, hash, scheme string) (bool, error) {
	signScheme := zcncrypto.NewSignatureScheme(scheme)
	if signScheme != nil {
		err := signScheme.SetPublicKey(publicKey)
		if err != nil {
			return false, err
		}
		return signScheme.Verify(signature, hash)
	}

	return false, errors.New("invalid_signature_scheme", "invalid signature scheme")
}
