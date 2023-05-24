package util

import (
	"math/rand"
	"testing"

	"fmt"

	"github.com/stretchr/testify/require"
)

const (
	KB = 1024
)

func TestFixedMerkleTreeWrite(t *testing.T) {
	for i := 0; i < 100; i++ {
		var n int64
		for {
			n = rand.Int63n(KB * KB)
			if n != 0 {
				break
			}
		}

		t.Run(fmt.Sprintf("Fmt test with dataSize: %d", n), func(t *testing.T) {

			b := make([]byte, n)
			rand.Read(b) //nolint

			leaves := make([]Hashable, FixedMerkleLeaves)
			for i := 0; i < len(leaves); i++ {
				leaves[i] = getNewLeaf()
			}

			for i := 0; i < len(b); i += MaxMerkleLeavesSize {
				leafCount := 0
				endInd := i + MaxMerkleLeavesSize
				if endInd > len(b) {
					endInd = len(b)
				}

				d := b[i:endInd]
				for j := 0; j < len(d); j += MerkleChunkSize {
					endInd := j + MerkleChunkSize
					if endInd > len(d) {
						endInd = len(d)
					}

					_, err := leaves[leafCount].Write(d[j:endInd])
					require.NoError(t, err)
					leafCount++
				}
			}

			mt := MerkleTree{}
			mt.ComputeTree(leaves)

			root := mt.GetRoot()

			ft := NewFixedMerkleTree()
			_, err := ft.Write(b)
			require.NoError(t, err)
			err = ft.Finalize()
			require.NoError(t, err)

			root1 := ft.GetMerkleRoot()
			require.Equal(t, root, root1)
		})
	}
}
