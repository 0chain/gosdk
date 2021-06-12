package util

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamMerkleHasherWithEvenLeaves(t *testing.T) {

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

	hasher := StreamMerkleHasher{Hash: func(left, right string) string {
		return "[" + left + "+" + right + "]"
	}}

	for i := 0; i < 6; i++ {
		require.NotEqual(t, ErrLeafNoSequenced, hasher.Push(strconv.Itoa(i), i))
	}

	require.Equal(t, "[[[0+1]+[2+3]]+[[4+5]+[4+5]]]", hasher.GetMerkleRoot(), "MerkleRoot with even leaves MUST equal")

}

func TestStreamMerkleHasherWithOddLeaves(t *testing.T) {

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

	hasher := StreamMerkleHasher{Hash: func(left, right string) string {
		return "[" + left + "+" + right + "]"
	}}

	for i := 0; i < 7; i++ {
		require.NotEqual(t, ErrLeafNoSequenced, hasher.Push(strconv.Itoa(i), i))
	}

	require.Equal(t, "[[[0+1]+[2+3]]+[[4+5]+[6+6]]]", hasher.GetMerkleRoot(), "MerkleRoot with odd leaves MUST equal")
}

func TestStreamMerkleHasherWithStateful(t *testing.T) {

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

	hasher := StreamMerkleHasher{Hash: func(left, right string) string {
		return "[" + left + "+" + right + "]"
	}}

	for i := 0; i < 7; i++ {

		require.NotEqual(t, ErrLeafNoSequenced, hasher.Push(strconv.Itoa(i), i))
		//try to push a leaf twice, merkle root should work properly
		require.NotEqual(t, ErrLeafNoSequenced, hasher.Push(strconv.Itoa(i), i))
	}

	require.Equal(t, "[[[0+1]+[2+3]]+[[4+5]+[6+6]]]", hasher.GetMerkleRoot(), "MerkleRoot with odd leaves MUST equal")
}

func TestStreamMerkleHasherWithNoSequenced(t *testing.T) {

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

	hasher := StreamMerkleHasher{Hash: func(left, right string) string {
		return "[" + left + "+" + right + "]"
	}}

	for i := 0; i < 7; i++ {

		require.NotEqual(t, ErrLeafNoSequenced, hasher.Push(strconv.Itoa(i), i))
	}

	require.Equal(t, ErrLeafNoSequenced, hasher.Push("10", 10))
}
