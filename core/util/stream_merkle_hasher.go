package util

import "errors"

var (
	// ErrLeafExists leaf has been computed, it can be skipped now
	ErrLeafExists = errors.New("merkle: leaf exists, it can be skipped")
	// ErrLeafNoSequenced leaf MUST be pushed one by one
	ErrLeafNoSequenced = errors.New("merkle: leaf must be pushed with sequence")
)

// StreamMerkleHasher it is a stateful algorithm. It takes data in (leaf nodes), hashes it, and computes as many parent hashes as it can.
// 	- /0chain/go/gosdk/docs/merkle/streaming-merkle-hasher.txt
type StreamMerkleHasher struct {
	Tree  []string                        `json:"tree"`  //node tree with computed as many parent hashes as it can
	Hash  func(left, right string) string `json:"-"`     //it should be set once hasher is created
	Count int                             `json:"count"` //how many leaves has been pushed
}

// Push add leaf hash and update the the Merkle tree.
func (hasher *StreamMerkleHasher) Push(leaf string, index int) error {

	if index < hasher.Count {
		return ErrLeafExists
	}

	if index > hasher.Count {
		return ErrLeafNoSequenced
	}

	rightHash := leaf

	for i, node := range hasher.Tree {
		if node == "" { // If we find an empty spot in the nodes, we put the hash there and quit.
			hasher.Tree[i] = rightHash
			hasher.Count++
			return nil
		}
		// Otherwise, hash the old hash with the new hash.
		leftHash := hasher.Tree[i]
		rightHash = hasher.Hash(leftHash, rightHash)
		// We no longer need to keep the old hash at this level in memory.
		hasher.Tree[i] = ""
	}

	if hasher.Tree == nil {
		hasher.Tree = make([]string, 0, 10)
	}

	//no valid left hash found, so make it as a new leaf hash
	hasher.Tree = append(hasher.Tree, rightHash)
	hasher.Count++
	return nil

}

// GetMerkleRoot calculate the Merkle root when all leave has been added,
// For the last, lowest-level hash, we hash it with itself.
// From there, the nodes are hashed to the top level
// to calculate the Merkle root.
func (hasher *StreamMerkleHasher) GetMerkleRoot() string {

	rightHash := ""

	// Fill in missing nodes.
	for i := range hasher.Tree {

		leftHash := hasher.Tree[i]
		if i == len(hasher.Tree) && rightHash == "" {
			// Perfectly balanced Merkle tree.
			return leftHash
		}
		if leftHash == "" && rightHash == "" {
			// Both leaves are null (subsumed by a higher node hash)
			continue
		} else if rightHash == "" {
			// If there is no right hash (at this level or lower in the tree),
			// Hash the left hash with itself.
			rightHash = hasher.Hash(leftHash, leftHash)
		} else if leftHash == "" {
			// Similarly, if there is no left half,
			// hash the right with itself.
			rightHash = hasher.Hash(rightHash, rightHash)
		} else {
			// Otherwise, the hash at this level will be the right hash
			// for higher levels in the tree.
			rightHash = hasher.Hash(leftHash, rightHash)
		}
	}

	return rightHash
}
