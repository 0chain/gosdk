package util

import (
	"encoding/hex"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/minio/sha256-simd"
)

/*MerkleTreeI - a merkle tree interface required for constructing and providing verification */
type MerkleTreeI interface {
	//API to create a tree from leaf nodes
	ComputeTree(hashes []Hashable)
	GetRoot() string
	GetTree() []string

	//API to load an existing tree
	SetTree(leavesCount int, tree []string) error

	// API for verification when the leaf node is known
	GetPath(hash Hashable) *MTPath               // Server needs to provide this
	VerifyPath(hash Hashable, path *MTPath) bool //This is only required by a client but useful for testing

	/* API for random verification when the leaf node is uknown
	(verification of the data to hash used as leaf node is outside this API) */
	GetPathByIndex(idx int) *MTPath
}

/*MTPath - The merkle tree path*/
type MTPath struct {
	Nodes     []string `json:"nodes"`
	LeafIndex int      `json:"leaf_index"`
}

/*Hash - the hashing used for the merkle tree construction */
func Hash(text string) string {
	return encryption.Hash(text)
}

func MHashBytes(h1, h2 []byte) []byte {
	buf := make([]byte, len(h1)+len(h2))
	copy(buf, h1)
	copy(buf[len(h1):], h2)
	hash := sha256.New()
	_, _ = hash.Write(buf)
	return hash.Sum(nil)
}

/*MHash - merkle hashing of a pair of child hashes */
func MHash(h1 string, h2 string) string {
	return Hash(h1 + h2)
}

// DecodeAndMHash will decode hex-encoded string to []byte format.
// This function should only be used with hex-encoded string otherwise the result will
// be obsolute.
func DecodeAndMHash(h1, h2 string) string {
	b1, _ := hex.DecodeString(h1)

	b2, _ := hex.DecodeString(h2)

	b3 := MHashBytes(b1, b2)
	return hex.EncodeToString(b3)
}

type StringHashable struct {
	Hash string
}

func NewStringHashable(hash string) *StringHashable {
	return &StringHashable{Hash: hash}
}

func (sh *StringHashable) GetHash() string {
	return sh.Hash
}
func (sh *StringHashable) GetHashBytes() []byte {
	return []byte(sh.Hash)
}

func (StringHashable) Write(_ []byte) (int, error) {
	return 0, nil
}
