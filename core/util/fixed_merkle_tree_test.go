package util

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixedMerkleTreeWithChunkSize(t *testing.T) {

	tests := []struct {
		Name       string
		MerkleTree FixedMerkleTree
		Data       []byte
	}{
		{
			Name:       "ChunkSize = 1024",
			MerkleTree: *NewFixedMerkleTree(1024),
			Data:       GenerateRandomBytes(1024),
		},
		{
			Name:       "ChunkSize > 1024",
			MerkleTree: *NewFixedMerkleTree(1025),
			Data:       GenerateRandomBytes(1025),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			require := require.New(t)
			require.Nil(test.MerkleTree.Write(test.Data, 0))
		})
	}

}

func TestFixedMerkleTreeChunksizeLessThan1024(t *testing.T) {

}

func TestFixedMerkleTreeChunksizeGreaterThan1024(t *testing.T) {

}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil
	}

	return b
}
