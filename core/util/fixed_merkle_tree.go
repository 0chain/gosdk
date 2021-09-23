package util

import (
	"bytes"
	"io"

	"github.com/0chain/errors"
)

// FixedMerkleTree A trusted mekerl tree for outsourcing attack protection. see section 1.8 on whitepager
// see detail on https://github.com/0chain/blobber/wiki/Protocols#what-is-fixedmerkletree
type FixedMerkleTree struct {
	// ChunkSize size of chunk
	ChunkSize int `json:"chunk_size,omitempty"`
	// Leaves a leaf is a CompactMerkleTree for 1/1024 shard data
	Leaves []*CompactMerkleTree `json:"leaves,omitempty"`
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
	fmt.Leaves = make([]*CompactMerkleTree, 1024)
	for n := 0; n < 1024; n++ {
		fmt.Leaves[n] = NewCompactMerkleTree(nil)
	}
}

func (fmt *FixedMerkleTree) Write(buf []byte, chunkIndex int) error {
	merkleChunkSize := fmt.ChunkSize / 1024
	total := len(buf)
	for i := 0; i < total; i += merkleChunkSize {
		end := i + merkleChunkSize
		if end > len(buf) {
			end = len(buf)
		}
		offset := i / merkleChunkSize

		if len(fmt.Leaves) != 1024 {
			fmt.initLeaves()
		}

		err := fmt.Leaves[offset].AddDataBlocks(buf[i:end], chunkIndex)
		if errors.Is(err, ErrLeafNoSequenced) {
			return err
		}
	}

	return nil
}

// GetMerkleRoot get merkle tree
func (fmt *FixedMerkleTree) GetMerkleTree() MerkleTreeI {
	merkleLeaves := make([]Hashable, 1024)

	for idx, leaf := range fmt.Leaves {

		merkleLeaves[idx] = NewStringHashable(leaf.GetMerkleRoot())
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
