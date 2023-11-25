package util

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeebo/blake3"
)

const (
	HashSize = 32
)

func TestValidationTreeWrite(t *testing.T) {
	dataSizes := []int64{
		MaxMerkleLeavesSize,
		MaxMerkleLeavesSize - 24*KB,
		MaxMerkleLeavesSize * 2,
		MaxMerkleLeavesSize * 3,
		MaxMerkleLeavesSize*10 - 1,
	}

	for _, s := range dataSizes {
		data := make([]byte, s)
		n, err := rand.Read(data)
		require.NoError(t, err)
		require.EqualValues(t, s, n)

		root := calculateValidationMerkleRoot(data)

		vt := NewValidationTree(s)
		diff := 1
		i := len(data) - diff

		_, err = vt.Write(data[0:i])
		require.NoError(t, err)
		vt.calculateRoot()

		require.False(t, bytes.Equal(root, vt.validationRoot))

		_, err = vt.Write(data[i:])
		require.NoError(t, err)

		err = vt.Finalize()
		require.NoError(t, err)

		vt.calculateRoot()
		require.True(t, bytes.Equal(root, vt.validationRoot))

		require.Error(t, vt.Finalize())
	}
}

func TestValidationTreeCalculateDepth(t *testing.T) {
	in := map[int]int{
		1:   1,
		2:   2,
		3:   3,
		4:   3,
		10:  5,
		100: 8,
	}

	for k, d := range in {
		v := ValidationTree{leaves: make([][]byte, k)}
		require.Equal(t, v.CalculateDepth(), d)
	}
}

func TestMerklePathVerificationForValidationTree(t *testing.T) {

	type input struct {
		dataSize int64
		startInd int
		endInd   int
	}

	tests := []*input{
		{
			dataSize: 24 * KB,
			startInd: 0,
			endInd:   0,
		},
		{
			dataSize: 340 * KB,
			startInd: 1,
			endInd:   3,
		},
		{
			dataSize: 640 * KB,
			startInd: 1,
			endInd:   4,
		},
		{
			dataSize: 640*KB + 1,
			startInd: 1,
			endInd:   5,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Data size: %d KB, startInd: %d, endInd:%d",
			test.dataSize/KB,
			test.startInd,
			test.endInd,
		), func(t *testing.T) {

			b := make([]byte, test.dataSize)
			n, err := rand.Read(b)

			require.NoError(t, err)
			require.EqualValues(t, test.dataSize, n)

			root, nodes, indexes, data, err := calculateValidationRootAndNodes(b, test.startInd, test.endInd)
			require.NoError(t, err)

			t.Logf("nodes len: %d; index len: %d, indexes: %v", len(nodes), len(indexes), indexes)
			vp := MerklePathForMultiLeafVerification{
				RootHash: root,
				Nodes:    nodes,
				Index:    indexes,
				DataSize: test.dataSize,
			}

			err = vp.VerifyMultipleBlocks(data)
			require.NoError(t, err)

			err = vp.VerifyMultipleBlocks(data[1:])
			require.Error(t, err)
		})

	}
}

func calculateValidationMerkleRoot(data []byte) []byte {
	hashes := make([][]byte, 0)
	for i := 0; i < len(data); i += MaxMerkleLeavesSize {
		j := i + MaxMerkleLeavesSize
		if j > len(data) {
			j = len(data)
		}
		h := blake3.New()
		_, _ = h.Write(data[i:j])
		hashes = append(hashes, h.Sum(nil))
	}

	if len(hashes) == 1 {
		return hashes[0]
	}
	for len(hashes) != 1 {
		newHashes := make([][]byte, 0)
		if len(hashes)%2 == 0 {
			for i := 0; i < len(hashes); i += 2 {
				h := blake3.New()
				_, _ = h.Write(hashes[i])
				_, _ = h.Write(hashes[i+1])
				newHashes = append(newHashes, h.Sum(nil))
			}
		} else {
			for i := 0; i < len(hashes)-1; i += 2 {
				h := blake3.New()
				_, _ = h.Write(hashes[i])
				_, _ = h.Write(hashes[i+1])
				newHashes = append(newHashes, h.Sum(nil))
			}
			h := blake3.New()
			_, _ = h.Write(hashes[len(hashes)-1])
			newHashes = append(newHashes, h.Sum(nil))
		}

		hashes = newHashes
	}
	return hashes[0]
}

func calculateValidationRootAndNodes(b []byte, startInd, endInd int) (
	root []byte, nodes [][][]byte, indexes [][]int, data []byte, err error,
) {

	totalLeaves := int(math.Ceil(float64(len(b)) / float64(MaxMerkleLeavesSize)))
	depth := int(math.Ceil(math.Log2(float64(totalLeaves)))) + 1

	if endInd >= totalLeaves {
		endInd = totalLeaves - 1
	}

	hashes := make([][]byte, 0)
	nodesData := make([]byte, 0)
	h := blake3.New()
	for i := 0; i < len(b); i += MaxMerkleLeavesSize {
		j := i + MaxMerkleLeavesSize
		if j > len(b) {
			j = len(b)
		}

		_, _ = h.Write(b[i:j])
		leafHash := h.Sum(nil)
		hashes = append(hashes, leafHash)
		h.Reset()
	}

	if len(hashes) == 1 {
		return hashes[0], nil, nil, b, nil
	}

	for len(hashes) != 1 {
		newHashes := make([][]byte, 0)
		if len(hashes)%2 == 0 {
			for i := 0; i < len(hashes); i += 2 {
				h := blake3.New()
				_, _ = h.Write(hashes[i])
				_, _ = h.Write(hashes[i+1])
				nodesData = append(nodesData, hashes[i]...)
				nodesData = append(nodesData, hashes[i+1]...)
				newHashes = append(newHashes, h.Sum(nil))
			}
		} else {
			for i := 0; i < len(hashes)-1; i += 2 {
				h := blake3.New()
				_, _ = h.Write(hashes[i])
				_, _ = h.Write(hashes[i+1])
				nodesData = append(nodesData, hashes[i]...)
				nodesData = append(nodesData, hashes[i+1]...)
				newHashes = append(newHashes, h.Sum(nil))
			}
			h := blake3.New()
			_, _ = h.Write(hashes[len(hashes)-1])
			nodesData = append(nodesData, hashes[len(hashes)-1]...)
			newHashes = append(newHashes, h.Sum(nil))
		}

		hashes = newHashes
	}

	nodes, indexes, err = getMerkleProofOfMultipleIndexes(nodesData, totalLeaves, depth, startInd, endInd)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	startOffset := startInd * 64 * KB
	endOffset := startOffset + (endInd-startInd+1)*64*KB
	if endOffset > len(b) {
		endOffset = len(b)
	}

	return hashes[0], nodes, indexes, b[startOffset:endOffset], nil
}

func getMerkleProofOfMultipleIndexes(nodesData []byte, totalLeaves, depth, startInd, endInd int) (
	[][][]byte, [][]int, error) {

	if endInd >= totalLeaves {
		endInd = totalLeaves - 1
	}

	if endInd < startInd {
		return nil, nil, errors.New("end index cannot be lesser than start index")
	}

	offsets, leftRightIndexes := getFileOffsetsAndNodeIndexes(totalLeaves, depth, startInd, endInd)

	offsetInd := 0
	nodeHashes := make([][][]byte, len(leftRightIndexes))
	for i, indexes := range leftRightIndexes {
		for range indexes {
			b := make([]byte, HashSize)
			off := offsets[offsetInd]
			n := copy(b, nodesData[off:off+HashSize])
			if n != HashSize {
				return nil, nil, errors.New("invalid hash length")
			}
			nodeHashes[i] = append(nodeHashes[i], b)
			offsetInd++
		}
	}
	return nodeHashes, leftRightIndexes, nil
}

func getFileOffsetsAndNodeIndexes(totalLeaves, depth, startInd, endInd int) ([]int, [][]int) {

	nodeIndexes, leftRightIndexes := getNodeIndexes(totalLeaves, depth, startInd, endInd)
	offsets := make([]int, 0)
	totalNodes := 0
	curNodesTot := totalLeaves
	for i := 0; i < len(nodeIndexes); i++ {
		for _, ind := range nodeIndexes[i] {
			offsetInd := ind + totalNodes
			offsets = append(offsets, offsetInd*HashSize)
		}
		totalNodes += curNodesTot
		curNodesTot = (curNodesTot + 1) / 2
	}

	return offsets, leftRightIndexes
}

func getNodeIndexes(totalLeaves, depth, startInd, endInd int) ([][]int, [][]int) {

	indexes := make([][]int, 0)
	leftRightIndexes := make([][]int, 0)
	totalNodes := totalLeaves
	for i := depth - 1; i >= 0; i-- {
		if startInd == 0 && endInd == totalNodes-1 {
			break
		}

		nodeOffsets := make([]int, 0)
		lftRtInd := make([]int, 0)
		if startInd&1 == 1 {
			nodeOffsets = append(nodeOffsets, startInd-1)
			lftRtInd = append(lftRtInd, Left)
		}

		if endInd != totalNodes-1 && endInd&1 == 0 {
			nodeOffsets = append(nodeOffsets, endInd+1)
			lftRtInd = append(lftRtInd, Right)
		}

		indexes = append(indexes, nodeOffsets)
		leftRightIndexes = append(leftRightIndexes, lftRtInd)
		startInd = startInd / 2
		endInd = endInd / 2
		totalNodes = (totalNodes + 1) / 2
	}
	return indexes, leftRightIndexes
}
