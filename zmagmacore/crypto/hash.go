// DEPRECATED: This package is deprecated and will be removed in a future release.
package crypto

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

const HashLength = 32

type HashBytes [HashLength]byte

// Hash computes hash of the given data using RawHash and returns result as hex decoded string.
func Hash(data interface{}) string {
	return hex.EncodeToString(RawHash(data))
}

// RawHash computes SHA3-256 hash depending on data type and returns the hash bytes.
//
// RawHash panics if data type is unknown.
//
// Known types:
//
// - []byte
//
// - HashBytes
//
// - string
func RawHash(data interface{}) []byte {
	var databuf []byte
	switch dataImpl := data.(type) {
	case []byte:
		databuf = dataImpl
	case HashBytes:
		databuf = dataImpl[:]
	case string:
		databuf = []byte(dataImpl)
	default:
		panic("unknown type")
	}
	hash := sha3.New256()
	hash.Write(databuf)
	var buf []byte
	return hash.Sum(buf)
}
