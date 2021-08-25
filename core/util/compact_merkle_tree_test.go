package util

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompactMerkleTreeWithEvenLeaves(t *testing.T) {

	/*
					  0──┐
					    [0+1]─────┐
					  1──┘        │
					        [[0+1]+[2+3]]─────────┐
					  2──┐        │               │
					    [2+3]─────┘               │
					  3──┘                        │
				                    [[[0+1]+[2+3]]+[[4+5]+[4+5]]]
				  	  4──┐                        │
					    [4+5]─────┐               │
					  5──┘        │               │
		                    [[4+5]+[4+5]]─────────┘
			                      │
			                 ─────┘


	*/

	hasher := CompactMerkleTree{Hash: func(left, right string) string {
		return "[" + left + "+" + right + "]"
	}}

	for i := 0; i < 6; i++ {
		require.NotEqual(t, ErrLeafNoSequenced, hasher.AddLeaf(strconv.Itoa(i), i))
	}

	require.Equal(t, "[[[0+1]+[2+3]]+[[4+5]+[4+5]]]", hasher.GetMerkleRoot(), "MerkleRoot with even leaves MUST equal")

}

func TestCompactMerkleTreeWithOddLeaves(t *testing.T) {

	/*
					  0──┐
					   [0+1]─────┐
					  1──┘       │
					       [[0+1]+[2+3]]───────┐
					  2──┐       │             │
					   [2+3]─────┘             │
					  3──┘                     │
				                 [[[0+1]+[2+3]]+[[4+5]+[6+6]]]
				  	  4──┐                     │
					   [4+5]─────┐             │
					  5──┘       │             │
		                   [[4+5]+[6+6]]───────┘
					  6──┐       │
		               [6+6]─────┘
					   ──┘

	*/

	hasher := CompactMerkleTree{Hash: func(left, right string) string {
		return "[" + left + "+" + right + "]"
	}}

	for i := 0; i < 7; i++ {
		require.NotEqual(t, ErrLeafNoSequenced, hasher.AddLeaf(strconv.Itoa(i), i))
	}

	require.Equal(t, "[[[0+1]+[2+3]]+[[4+5]+[6+6]]]", hasher.GetMerkleRoot(), "MerkleRoot with odd leaves MUST equal")
}

func TestCompactMerkleTreeWithStateful(t *testing.T) {

	/*
					  0──┐
					   [0+1]─────┐
					  1──┘       │
					       [[0+1]+[2+3]]───────┐
					  2──┐       │             │
					   [2+3]─────┘             │
					  3──┘                     │
				                 [[[0+1]+[2+3]]+[[4+5]+[6+6]]]
				  	  4──┐                     │
					   [4+5]─────┐             │
					  5──┘       │             │
		                   [[4+5]+[6+6]]───────┘
					  6──┐       │
		               [6+6]─────┘
					   ──┘

	*/

	hasher := CompactMerkleTree{Hash: func(left, right string) string {
		return "[" + left + "+" + right + "]"
	}}

	for i := 0; i < 7; i++ {

		require.NotEqual(t, ErrLeafNoSequenced, hasher.AddLeaf(strconv.Itoa(i), i))
		//try to push a leaf twice, merkle root should work properly
		require.NotEqual(t, ErrLeafNoSequenced, hasher.AddLeaf(strconv.Itoa(i), i))
	}

	require.Equal(t, "[[[0+1]+[2+3]]+[[4+5]+[6+6]]]", hasher.GetMerkleRoot(), "MerkleRoot with odd leaves MUST equal")
}

func TestCompactMerkleTreeWithNoSequenced(t *testing.T) {

	/*
					  0──┐
					   [0+1]─────┐
					  1──┘       │
					       [[0+1]+[2+3]]───────┐
					  2──┐       │             │
					   [2+3]─────┘             │
					  3──┘                     │
				                 [[[0+1]+[2+3]]+[[4+5]+[6+6]]]
				  	  4──┐                     │
					   [4+5]─────┐             │
					  5──┘       │             │
		                   [[4+5]+[6+6]]───────┘
					  6──┐       │
		               [6+6]─────┘
					   ──┘

	*/

	hasher := CompactMerkleTree{Hash: func(left, right string) string {
		return "[" + left + "+" + right + "]"
	}}

	for i := 0; i < 7; i++ {

		require.NotEqual(t, ErrLeafNoSequenced, hasher.AddLeaf(strconv.Itoa(i), i))
	}

	require.Equal(t, ErrLeafNoSequenced, hasher.AddLeaf("10", 10))
}
