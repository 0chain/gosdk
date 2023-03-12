package sdk

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
)

// downloadHahser verify hash for downloading
type downloadHasher struct {
	hasher       hash.Hash
	streamHasher hash.Hash
	chunkIndex   int
	shardSize    int
	buf          []byte
}

// createDownloadHasher create a DownloadHasher instance
func createDownloadHasher(chunkSize int, dataShards int, encryptOnUpload bool) *downloadHasher {

	if encryptOnUpload {
		chunkSize -= EncryptionHeaderSize + EncryptedDataPaddingSize
	}

	shardSize := chunkSize * dataShards

	return &downloadHasher{
		hasher:       sha256.New(),
		streamHasher: sha256.New(),
		shardSize:    shardSize,
		buf:          make([]byte, 0, shardSize),
		chunkIndex:   0,
	}
}

// Write write bytes for hash
func (dh *downloadHasher) Write(p []byte) (n int, err error) {

	for _, v := range p {
		if len(dh.buf) == dh.shardSize {
			dh.streamHasher.Write(dh.buf)
			dh.chunkIndex++
			dh.buf = make([]byte, 0, dh.shardSize)
		}

		dh.buf = append(dh.buf, v)

	}

	return dh.hasher.Write(p)
}

// GetHash get sha1 hash for old upload
func (dh *downloadHasher) GetHash() string {
	return hex.EncodeToString(dh.hasher.Sum(nil))
}

// GetMerkleRoot get merkle root hash for new upload
func (dh *downloadHasher) GetMerkleRoot() string {
	if len(dh.buf) > 0 {
		dh.streamHasher.Write(dh.buf)
		dh.chunkIndex++
		dh.buf = make([]byte, 0, dh.shardSize)
	}

	return hex.EncodeToString(dh.streamHasher.Sum(nil))
}
