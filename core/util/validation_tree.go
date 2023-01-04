package util

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"math"
	"sync"
)

type ValidationTree struct {
	writeLock      *sync.Mutex
	writeCount     int
	dataSize       int64
	writtenSize    int64
	leafIndex      int
	leaves         [][]byte
	isFinal        bool
	h              hash.Hash
	validationRoot []byte
}

func (v *ValidationTree) GetLeaves() [][]byte {
	return v.leaves
}

func (v *ValidationTree) GetValidationRoot() []byte {
	if len(v.validationRoot) > 0 {
		return v.validationRoot
	}
	v.calculateRoot()
	return v.validationRoot
}

func (v *ValidationTree) Write(b []byte) (int, error) {
	v.writeLock.Lock()
	defer v.writeLock.Unlock()

	if v.writtenSize+int64(len(b)) > v.dataSize {
		return 0, fmt.Errorf("data size overflow. expected %d, got %d", v.dataSize, v.writtenSize+int64(len(b)))
	}

	byteLen := len(b)
	shouldContinue := true
	for i, j := 0, MaxMerkleLeavesSize-v.writeCount; shouldContinue; i, j = j, j+MaxMerkleLeavesSize {
		if j > byteLen {
			j = byteLen
			shouldContinue = false
		}

		n, _ := v.h.Write(b[i:j])
		v.writeCount += n
		if v.writeCount == MaxMerkleLeavesSize {
			v.leaves[v.leafIndex] = v.h.Sum(nil)
			v.leafIndex++
			v.writeCount = 0
			v.h.Reset()
		}
	}
	v.writtenSize += int64(byteLen)
	return byteLen, nil
}

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

func (v *ValidationTree) Finalize() error {
	v.writeLock.Lock()
	defer v.writeLock.Unlock()

	if v.isFinal {
		return errors.New("already finalized")
	}
	if v.writtenSize != v.dataSize {
		return fmt.Errorf("invalid size. Expected %d got %d", v.dataSize, v.writtenSize)
	}

	if v.writeCount > 0 {
		v.leaves[v.leafIndex] = v.h.Sum(nil)
	}
	return nil
}

func NewValidationTree(dataSize int64) *ValidationTree {
	v := ValidationTree{
		writeLock: &sync.Mutex{},
		dataSize:  dataSize,
	}
	return &v
}
