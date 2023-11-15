package encryption

import (
	"crypto/sha1"
	"encoding/hex"

	"github.com/zeebo/blake3"
	"golang.org/x/crypto/sha3"
)

const HASH_LENGTH = 32

type HashBytes [HASH_LENGTH]byte

/*Hash - hash the given data and return the hash as hex string */
func Hash(data interface{}) string {
	return hex.EncodeToString(RawHash(data))
}

/*RawHash - Logic to hash the text and return the hash bytes */
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

func BlakeHash(data interface{}) []byte {
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
	hash := blake3.New()
	_, _ = hash.Write(databuf)
	return hash.Sum(nil)
}

/*FastHash - sha1 hash the given data and return the hash as hex string */
func FastHash(data interface{}) string {
	return hex.EncodeToString(RawFastHash(data))
}

/*RawFastHash - Logic to sha1 hash the text and return the hash bytes */
func RawFastHash(data interface{}) []byte {
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
	hash := sha1.New()
	hash.Write(databuf)
	var buf []byte
	return hash.Sum(buf)
}
