package util

import (
	"crypto/sha1"
	"encoding/hex"
)

// TrustedConentHasher A trusted mekerl tree for outsourcing attack protection. see section 1.8 on whitepager
type TrustedConentHasher struct {
	// ChunkSize size of chunk
	ChunkSize int `json:"chunk_size,omitempty"`
	// Leaves a leaf is a StreamMerkleHash of 1/1024 shard
	Leaves []*StreamMerkleHasher `json:"leaves,omitempty"`
}

func (tch *TrustedConentHasher) Write(buf []byte, chunkIndex int) {
	merkleChunkSize := tch.ChunkSize / 1024
	total := len(buf)
	for i := 0; i < total; i += merkleChunkSize {
		end := i + merkleChunkSize
		if end > len(buf) {
			end = len(buf)
		}
		offset := i / merkleChunkSize

		h := sha1.New()
		h.Write(buf[i:end])

		if len(tch.Leaves) != 1024 {
			tch.Leaves = make([]*StreamMerkleHasher, 1024)
			for n := 0; n < 1024; n++ {
				tch.Leaves[n] = NewStreamMerkleHasher(nil)
			}
		}

		tch.Leaves[offset].Push(hex.EncodeToString(h.Sum(nil)), chunkIndex)
	}
}

// GetMerkleRoot get merkle root
func (tch *TrustedConentHasher) GetMerkleRoot() string {
	merkleLeaves := make([]Hashable, 1024)

	for idx, leaf := range tch.Leaves {

		merkleLeaves[idx] = NewStringHashable(leaf.GetMerkleRoot())
	}
	var mt MerkleTreeI = &MerkleTree{}

	mt.ComputeTree(merkleLeaves)

	return mt.GetRoot()
}
