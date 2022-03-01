package fileref

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/0chain/gosdk/core/encryption"
)

// Hashnode ref node in hash tree
type Hashnode struct {
	// hash data
	AllocationID   string          `json:"allocation_id,omitempty"`
	Type           string          `json:"type,omitempty"`
	Name           string          `json:"name,omitempty"`
	Path           string          `json:"path,omitempty"`
	ContentHash    string          `json:"content_hash,omitempty"`
	MerkleRoot     string          `json:"merkle_root,omitempty"`
	ActualFileHash string          `json:"actual_file_hash,omitempty"`
	Attributes     json.RawMessage `json:"attributes,omitempty"`
	ChunkSize      int64           `json:"chunk_size,omitempty"`
	Size           int64           `json:"size,omitempty"`
	ActualFileSize int64           `json:"actual_file_size,omitempty"`

	Children   []*Hashnode `json:"children,omitempty"`
	lookupHash string      `json:"-"`
}

func (n *Hashnode) AddChild(c *Hashnode) {
	if n.Children == nil {
		n.Children = make([]*Hashnode, 0, 10)
	}

	n.Children = append(n.Children, c)
}

// GetLookupHash get lookuphash
func (n *Hashnode) GetLookupHash() string {
	if n.lookupHash == "" {
		n.lookupHash = encryption.Hash(n.AllocationID + ":" + n.Path)
	}
	return n.lookupHash
}

// GetHashCode get hash code
func (n *Hashnode) GetHashCode() string {
	// dir
	if n.Type == DIRECTORY {
		if len(n.Children) == 0 {
			return ""
		}

		sort.SliceStable(n.Children, func(i, j int) bool {
			return strings.Compare(n.Children[i].GetLookupHash(), n.Children[j].GetLookupHash()) == -1
		})

		childHashes := make([]string, len(n.Children))

		var size int64

		for i, child := range n.Children {
			childHashes[i] = child.GetHashCode()
			size += child.Size
		}

		n.Size = size

		return encryption.Hash(strings.Join(childHashes, ":"))

	}

	//file
	if len(n.Attributes) == 0 {
		n.Attributes = json.RawMessage("{}")
	}

	attrs, _ := json.Marshal(n.Attributes)

	hashArray := []string{
		n.AllocationID,
		n.Type,
		n.Name,
		n.Path,
		strconv.FormatInt(n.Size, 10),
		n.ContentHash,
		n.MerkleRoot,
		strconv.FormatInt(n.ActualFileSize, 10),
		n.ActualFileHash,
		string(attrs),
		strconv.FormatInt(n.ChunkSize, 10),
	}

	return encryption.Hash(strings.Join(hashArray, ":"))

}
