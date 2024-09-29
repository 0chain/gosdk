package util

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"math"
	"sync"

	"github.com/minio/sha256-simd"
)

const (
	// Left tree node chile
	Left = iota

	// Right tree node child
	Right
)

const (
	START_LENGTH = 64
	ADD_LENGTH   = 320
)

// ValidationTree is a merkle tree that is used to validate the data
type ValidationTree struct {
	writeLock      sync.Mutex
	writeCount     int
	dataSize       int64
	writtenSize    int64
	leafIndex      int
	leaves         [][]byte
	isFinalized    bool
	h              hash.Hash
	validationRoot []byte
}

// GetLeaves returns the leaves of the validation tree
func (v *ValidationTree) GetLeaves() [][]byte {
	return v.leaves
}

// SetLeaves sets the leaves of the validation tree.
// 		- leaves: leaves of the validation tree, each leaf is in byte format
func (v *ValidationTree) SetLeaves(leaves [][]byte) {
	v.leaves = leaves
}

// GetDataSize returns the data size of the validation tree
func (v *ValidationTree) GetDataSize() int64 {
	return v.dataSize
}

// GetValidationRoot returns the validation root of the validation tree
func (v *ValidationTree) GetValidationRoot() []byte {
	if len(v.validationRoot) > 0 {
		return v.validationRoot
	}
	v.calculateRoot()
	return v.validationRoot
}

// Write writes the data to the validation tree
func (v *ValidationTree) Write(b []byte) (int, error) {
	v.writeLock.Lock()
	defer v.writeLock.Unlock()

	if v.isFinalized {
		return 0, fmt.Errorf("tree is already finalized")
	}

	if len(b) == 0 {
		return 0, nil
	}

	if v.dataSize > 0 && v.writtenSize+int64(len(b)) > v.dataSize {
		return 0, fmt.Errorf("data size overflow. expected %d, got %d", v.dataSize, v.writtenSize+int64(len(b)))
	}

	byteLen := len(b)
	shouldContinue := true
	// j is initialized to MaxMerkleLeavesSize - writeCount so as to make up MaxMerkleLeavesSize with previously
	// read bytes. If previously it had written MaxMerkleLeavesSize - 1, then j will be initialized to 1 so
	// in first iteration it will only read 1 byte and write it to v.h after which hash of v.h will be calculated
	// and stored in v.Leaves and v.h will be reset.
	for i, j := 0, MaxMerkleLeavesSize-v.writeCount; shouldContinue; i, j = j, j+MaxMerkleLeavesSize {
		if j > byteLen {
			j = byteLen
			shouldContinue = false
		}

		n, _ := v.h.Write(b[i:j])
		v.writeCount += n // update write count
		if v.writeCount == MaxMerkleLeavesSize {
			if v.leafIndex >= len(v.leaves) {
				// increase leaves size
				leaves := make([][]byte, len(v.leaves)+ADD_LENGTH)
				copy(leaves, v.leaves)
				v.leaves = leaves
			}
			v.leaves[v.leafIndex] = v.h.Sum(nil)
			v.leafIndex++
			v.writeCount = 0 // reset writeCount
			v.h.Reset()      // reset hasher
		}
	}
	v.writtenSize += int64(byteLen)
	return byteLen, nil
}

// CalculateDepth calculates the depth of the validation tree
func (v *ValidationTree) CalculateDepth() int {
	return int(math.Ceil(math.Log2(float64(len(v.leaves))))) + 1
}

func (v *ValidationTree) calculateRoot() {
	totalLeaves := len(v.leaves)
	depth := v.CalculateDepth()
	nodes := make([][]byte, totalLeaves)
	copy(nodes, v.leaves)
	h := sha256.New()

	for i := 0; i < depth; i++ {
		if len(nodes) == 1 {
			break
		}
		newNodes := make([][]byte, 0)
		if len(nodes)%2 == 0 {
			for j := 0; j < len(nodes); j += 2 {
				h.Reset()
				h.Write(nodes[j])
				h.Write(nodes[j+1])
				newNodes = append(newNodes, h.Sum(nil))
			}
		} else {
			for j := 0; j < len(nodes)-1; j += 2 {
				h.Reset()
				h.Write(nodes[j])
				h.Write(nodes[j+1])
				newNodes = append(newNodes, h.Sum(nil))
			}
			h.Reset()
			h.Write(nodes[len(nodes)-1])
			newNodes = append(newNodes, h.Sum(nil))
		}
		nodes = newNodes
	}

	v.validationRoot = nodes[0]
}

// Finalize finalizes the validation tree, set isFinalized to true and calculate the root
func (v *ValidationTree) Finalize() error {
	v.writeLock.Lock()
	defer v.writeLock.Unlock()

	if v.isFinalized {
		return errors.New("already finalized")
	}
	if v.dataSize > 0 && v.writtenSize != v.dataSize {
		return fmt.Errorf("invalid size. Expected %d got %d", v.dataSize, v.writtenSize)
	}

	v.isFinalized = true

	if v.writeCount > 0 {
		if v.leafIndex == len(v.leaves) {
			// increase leaves size
			leaves := make([][]byte, len(v.leaves)+1)
			copy(leaves, v.leaves)
			v.leaves = leaves
		}
		v.leaves[v.leafIndex] = v.h.Sum(nil)
	} else {
		v.leafIndex--
	}
	if v.leafIndex < len(v.leaves) {
		v.leaves = v.leaves[:v.leafIndex+1]
	}
	return nil
}

// NewValidationTree creates a new validation tree
//   - dataSize is the size of the data
func NewValidationTree(dataSize int64) *ValidationTree {
	totalLeaves := (dataSize + MaxMerkleLeavesSize - 1) / MaxMerkleLeavesSize
	if totalLeaves == 0 {
		totalLeaves = START_LENGTH
	}
	return &ValidationTree{
		dataSize: dataSize,
		h:        sha256.New(),
		leaves:   make([][]byte, totalLeaves),
	}
}

// MerklePathForMultiLeafVerification is used to verify multiple blocks with single instance of
// merkle path. Usually client would request with counter incremented by 10. So if the block size
// is 64KB and counter is incremented by 10 then client is requesting 640 KB of data. Blobber can then
// provide sinlge merkle path instead of sending 10 merkle paths.
type MerklePathForMultiLeafVerification struct {
	// RootHash that was signed by the client
	RootHash []byte
	// Nodes contains a slice for each merkle node level. Each slice contains hash that will
	// be concatenated with the calculated hash from the level below.
	// It is used together with field Index [][]int
	// Length of Nodes will be according to number of blocks requested. If whole data is requested then
	// blobber will send nil for Nodes i.e. length of Nodes will become zero.
	Nodes [][][]byte `json:"nodes"`
	// Index slice that determines whether to concatenate hash to left or right.
	// It should have maximum of length 2 and minimum of 0. It is used together with field Nodes [][][]byte
	Index [][]int `json:"index"`
	// DataSize is size of data received by the blobber for the respective file.
	// It is not same as actual file size
	DataSize      int64
	totalLeaves   int
	requiredDepth int
}

/*
VerifyMultipleBlocks will verify merkle path for continuous data which is multiple of 64KB blocks

There can be at most 2 hashes in the input for each depth i.e. of the format below:
h1, data1, data2, data3, data4, h2
Note that data1, data2, data3,... should be continuous data

i#3                h14
i#2       h12             h13
i#1    h7      h8      h9    h10
i#0  h0, h1, h2, h3, h4, h5, h6

Consider there are 7 leaves(0...6) as shown above. Now if client wants data from
1-3 then blobber needs to provide:

1. One node from i#0, [h0]; data1 will generate h1,data2 will generate h2 and so on...
2. Zero node from i#1; h0 and h1 will generate h7 and h2 and h3 will generate h8
3. One node from i#2, h[13]; h7 and h8 will generate h12. Now to get h14, we need h13
which will be provided by blobber

i#5                                                  h37
i#4                                 h35                                       h36
i#3                h32                               h33                      h34
i#2        h27             h28             h29                 h30            h31
i#1    h18     h19     h20     h21     h22      h23       h24       h25,      h26
i#0  h0, h1, h2, h3, h4, h5, h6, h7, h8, h9, h10, h11, h12, h13, h14, h15, h16, h17

Consider there are 16 leaves(0..15) with total data = 16*64KB as shown above.
If client wants data from 3-10 then blobber needs to provide:
1. Two nodes from i#0, [h2, h11]
2. One node from i#1, [h16]
3. One node from i#2, [h27]

If client had required data from 2-9 then blobber would have to provide:
1. Zero nodes from i#0
2. Two nodes from i#1, [h16, h21]
3. One node from i#2, [h27]
*/
func (m *MerklePathForMultiLeafVerification) VerifyMultipleBlocks(data []byte) error {

	hashes := make([][]byte, 0)
	h := sha256.New()
	// Calculate hashes from the data responded from the blobber.
	for i := 0; i < len(data); i += MaxMerkleLeavesSize {
		endIndex := i + MaxMerkleLeavesSize
		if endIndex > len(data) {
			endIndex = len(data)
		}
		h.Reset()
		h.Write(data[i:endIndex])
		hashes = append(hashes, h.Sum(nil))
	}

	if m.requiredDepth == 0 {
		m.calculateRequiredLevels()
	}
	for i := 0; i < m.requiredDepth-1; i++ {
		if len(m.Nodes) > i {
			if len(m.Index[i]) == 2 { // has both nodes to append for
				hashes = append([][]byte{m.Nodes[i][0]}, hashes...)
				hashes = append(hashes, m.Nodes[i][1])
			} else if len(m.Index[i]) == 1 { // hash single node to append for
				if m.Index[i][0] == Right { // append to right
					hashes = append(hashes, m.Nodes[i][0])
				} else {
					hashes = append([][]byte{m.Nodes[i][0]}, hashes...)
				}
			}
		}

		hashes = m.calculateIntermediateHashes(hashes)

	}

	if len(hashes) == 0 {
		return fmt.Errorf("no hashes to verify, data is empty")
	}

	if !bytes.Equal(m.RootHash, hashes[0]) {
		return fmt.Errorf("calculated root %s; expected %s",
			hex.EncodeToString(hashes[0]),
			hex.EncodeToString(m.RootHash))
	}
	return nil
}

func (m *MerklePathForMultiLeafVerification) calculateIntermediateHashes(hashes [][]byte) [][]byte {
	newHashes := make([][]byte, 0)
	h := sha256.New()
	if len(hashes)%2 == 0 {
		for i := 0; i < len(hashes); i += 2 {
			h.Reset()
			h.Write(hashes[i])
			h.Write(hashes[i+1])
			newHashes = append(newHashes, h.Sum(nil))
		}
	} else {
		for i := 0; i < len(hashes)-1; i += 2 {
			h.Reset()
			h.Write(hashes[i])
			h.Write(hashes[i+1])
			newHashes = append(newHashes, h.Sum(nil))
		}
		h.Reset()
		h.Write(hashes[len(hashes)-1])
		newHashes = append(newHashes, h.Sum(nil))
	}
	return newHashes
}

func (m *MerklePathForMultiLeafVerification) calculateTotalLeaves() {
	m.totalLeaves = int((m.DataSize + MaxMerkleLeavesSize - 1) / MaxMerkleLeavesSize)
}

func (m *MerklePathForMultiLeafVerification) calculateRequiredLevels() {
	if m.totalLeaves == 0 {
		m.calculateTotalLeaves()
	}
	m.requiredDepth = int(math.Ceil(math.Log2(float64(m.totalLeaves)))) + 1 // Add root hash to be a level
}
