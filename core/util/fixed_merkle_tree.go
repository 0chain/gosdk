package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"

	"github.com/0chain/errors"
)

// FixedMerkleTree A trusted mekerl tree for outsourcing attack protection. see section 1.8 on whitepager
// see detail on https://github.com/0chain/blobber/wiki/Protocols#what-is-fixedmerkletree
type FixedMerkleTree struct {
	// ChunkSize size of chunk
	ChunkSize int `json:"chunk_size,omitempty"`
	// Leaves a leaf is a CompactMerkleTree for 1/1024 shard data
	Leaves []hash.Hash `json:"-"`
}

// NewFixedMerkleTree create a FixedMerkleTree with specify hash method
func NewFixedMerkleTree(chunkSize int) *FixedMerkleTree {

	t := &FixedMerkleTree{
		ChunkSize: chunkSize,
	}
	t.initLeaves()

	return t

}

func (fmt *FixedMerkleTree) initLeaves() {
	fmt.Leaves = make([]hash.Hash, 1024)
	for n := 0; n < 1024; n++ {
		fmt.Leaves[n] = sha256.New()
	}
}

func (fmt *FixedMerkleTree) Write(buf []byte, chunkIndex int) error {
	//split chunk into 1024 parts for challenge hash
	merkleChunkSize := fmt.ChunkSize / 1024

	// chunksize is less than 1024
	if merkleChunkSize == 0 {
		merkleChunkSize = 1
	}

	total := len(buf)
	offset := 0
	for i := 0; i < total; i += merkleChunkSize {
		end := i + merkleChunkSize
		if end > len(buf) {
			end = len(buf)
		}

		if len(fmt.Leaves) != 1024 {
			fmt.initLeaves()
		}

		_, err := fmt.Leaves[offset].Write(buf[i:end])
		if err != nil {
			return err
		}

		offset++
		if offset >= 1024 {
			offset = 0
		}
	}

	return nil
}

// GetMerkleRoot get merkle tree
func (fmt *FixedMerkleTree) GetMerkleTree() MerkleTreeI {
	merkleLeaves := make([]Hashable, 1024)

	for idx, leaf := range fmt.Leaves {

		merkleLeaves[idx] = NewStringHashable(hex.EncodeToString(leaf.Sum(nil)))
	}
	var mt MerkleTreeI = &MerkleTree{}

	mt.ComputeTree(merkleLeaves)

	return mt
}

// GetMerkleRoot get merkle root
func (fmt *FixedMerkleTree) GetMerkleRoot() string {
	return fmt.GetMerkleTree().GetRoot()
}

// Reload reset and reload leaves from io.Reader
func (fmt *FixedMerkleTree) Reload(reader io.Reader) error {

	fmt.initLeaves()

	bytesBuf := bytes.NewBuffer(make([]byte, 0, fmt.ChunkSize))
	for i := 0; ; i++ {
		written, err := io.CopyN(bytesBuf, reader, int64(fmt.ChunkSize))

		if written > 0 {
			err = fmt.Write(bytesBuf.Bytes(), i)
			bytesBuf.Reset()

			if err != nil {
				return err
			}

		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

	}

	return nil
}
