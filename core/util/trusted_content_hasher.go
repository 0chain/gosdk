package util

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
)

// TrustedConentHasher A trusted mekerl tree for outsourcing attack protection. see section 1.8 on whitepager
type TrustedConentHasher struct {
	ChunkSize int
	leaves    []*StreamMerkleHasher
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

		if len(tch.leaves) == 0 {
			tch.leaves = make([]*StreamMerkleHasher, 1024)
			for n := 0; n < 1024; n++ {
				tch.leaves[n] = NewStreamMerkleHasher(nil)
			}
		}

		tch.leaves[offset].Push(hex.EncodeToString(h.Sum(nil)), chunkIndex)
	}
}

// GetMerkleRoot get merkle root
func (tch *TrustedConentHasher) GetMerkleRoot() string {
	merkleLeaves := make([]Hashable, 1024)

	for idx, leaf := range tch.leaves {

		merkleLeaves[idx] = NewStringHashable(leaf.GetMerkleRoot())
	}
	var mt MerkleTreeI = &MerkleTree{}

	mt.ComputeTree(merkleLeaves)

	return mt.GetRoot()
}

// UnmarshalJSON  implments json.Unmarshaler
func (tch *TrustedConentHasher) UnmarshalJSON(b []byte) error {
	var leaves []*StreamMerkleHasher

	err := json.Unmarshal(b, &leaves)
	if err != nil {
		return err
	}

	tch.leaves = leaves

	return nil
}

// MarshalJSON implements json.Marshaler
func (tch TrustedConentHasher) MarshalJSON() ([]byte, error) {
	if len(tch.leaves) == 0 {
		tch.leaves = make([]*StreamMerkleHasher, 1024)
		for n := 0; n < 1024; n++ {
			tch.leaves[n] = NewStreamMerkleHasher(nil)
		}
	}

	return json.Marshal(tch.leaves)
}
