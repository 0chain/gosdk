package fileref

import (
	"strconv"
	"strings"
	"testing"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/stretchr/testify/require"
)

func TestHashnode_GetHashcode(t *testing.T) {
	tests := []struct {
		name           string
		exceptedRef    blobberRef
		actualHashnode Hashnode
	}{
		{
			name: "Empty root should be same",
			exceptedRef: blobberRef{
				Type: DIRECTORY,
			},
			actualHashnode: Hashnode{
				Type: DIRECTORY,
			},
		},
		{
			name: "Nested nodes should be same",
			exceptedRef: blobberRef{
				AllocationID:   "nested_nodes",
				Type:           DIRECTORY,
				Name:           "/",
				Path:           "/",
				ContentHash:    "content_hash",
				MerkleRoot:     "merkle_root",
				ActualFileHash: "actual_file_hash",
				ChunkSize:      1024,
				Size:           10240,
				ActualFileSize: 10240,
				Children: []*blobberRef{
					{
						AllocationID:   "nested_nodes",
						Type:           DIRECTORY,
						Name:           "sub1",
						Path:           "/sub1",
						ContentHash:    "content_hash",
						MerkleRoot:     "merkle_root",
						ActualFileHash: "actual_file_hash",
						ChunkSize:      1024,
						Size:           10240,
						ActualFileSize: 10240,
						childrenLoaded: true,
						LookupHash:     GetReferenceLookup("nested_nodes", "/sub1"),
						Children: []*blobberRef{
							{
								AllocationID:   "nested_nodes",
								Type:           FILE,
								Name:           "file1",
								Path:           "/sub1/file1",
								ContentHash:    "content_hash",
								MerkleRoot:     "merkle_root",
								ActualFileHash: "actual_file_hash",
								ChunkSize:      1024,
								Size:           10240,
								ActualFileSize: 10240,
								LookupHash:     GetReferenceLookup("nested_nodes", "/sub1/file1"),
							},
						},
					},
					{
						AllocationID:   "nested_nodes",
						Type:           DIRECTORY,
						Name:           "emptydir",
						Path:           "/emptydir",
						ContentHash:    "content_hash",
						MerkleRoot:     "merkle_root",
						ActualFileHash: "actual_file_hash",
						ChunkSize:      0,
						Size:           0,
						ActualFileSize: 0,
						LookupHash:     GetReferenceLookup("nested_nodes", "/emptydir"),
					},
				},
			},
			actualHashnode: Hashnode{
				AllocationID:   "nested_nodes",
				Type:           DIRECTORY,
				Name:           "/",
				Path:           "/",
				ContentHash:    "content_hash",
				MerkleRoot:     "merkle_root",
				ActualFileHash: "actual_file_hash",
				ChunkSize:      1024,
				Size:           10240,
				ActualFileSize: 10240,
				Children: []*Hashnode{
					{
						AllocationID:   "nested_nodes",
						Type:           DIRECTORY,
						Name:           "sub1",
						Path:           "/sub1",
						ContentHash:    "content_hash",
						MerkleRoot:     "merkle_root",
						ActualFileHash: "actual_file_hash",
						ChunkSize:      1024,
						Size:           10240,
						ActualFileSize: 10240,
						Children: []*Hashnode{
							{
								AllocationID:   "nested_nodes",
								Type:           FILE,
								Name:           "file1",
								Path:           "/sub1/file1",
								ContentHash:    "content_hash",
								MerkleRoot:     "merkle_root",
								ActualFileHash: "actual_file_hash",
								ChunkSize:      1024,
								Size:           10240,
								ActualFileSize: 10240,
							},
						},
					},
					{
						AllocationID:   "nested_nodes",
						Type:           DIRECTORY,
						Name:           "emptydir",
						Path:           "/emptydir",
						ContentHash:    "content_hash",
						MerkleRoot:     "merkle_root",
						ActualFileHash: "actual_file_hash",
						ChunkSize:      0,
						Size:           0,
						ActualFileSize: 0,
					},
				},
			},
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(test *testing.T) {
			require.Equal(test, it.exceptedRef.CalculateHash(), it.actualHashnode.GetHashCode())
		})
	}
}

// blobberRef copied from https://github.com/0chain/blobber/blob/staging/code/go/0chain.net/blobbercore/reference/ref.go for unit tests
type blobberRef struct {
	ID           int64  `gorm:"column:id;primary_key"`
	Type         string `gorm:"column:type" dirlist:"type" filelist:"type"`
	AllocationID string `gorm:"column:allocation_id"`
	LookupHash   string `gorm:"column:lookup_hash" dirlist:"lookup_hash" filelist:"lookup_hash"`
	Name         string `gorm:"column:name" dirlist:"name" filelist:"name"`
	Path         string `gorm:"column:path" dirlist:"path" filelist:"path"`
	Hash         string `gorm:"column:hash" dirlist:"hash" filelist:"hash"`

	ParentPath string `gorm:"column:parent_path"`

	ContentHash    string `gorm:"column:content_hash" filelist:"content_hash"`
	Size           int64  `gorm:"column:size" dirlist:"size" filelist:"size"`
	MerkleRoot     string `gorm:"column:merkle_root" filelist:"merkle_root"`
	ActualFileSize int64  `gorm:"column:actual_file_size" filelist:"actual_file_size"`
	ActualFileHash string `gorm:"column:actual_file_hash" filelist:"actual_file_hash"`

	Children       []*blobberRef `gorm:"-"`
	childrenLoaded bool

	ChunkSize int64 `gorm:"column:chunk_size" dirlist:"chunk_size" filelist:"chunk_size"`
}

func (fr *blobberRef) GetFileHashData() string {
	hashArray := make([]string, 0, 11)
	hashArray = append(hashArray,
		fr.AllocationID,
		fr.Type,
		fr.Name,
		fr.Path,
		strconv.FormatInt(fr.Size, 10),
		fr.ContentHash,
		fr.MerkleRoot,
		strconv.FormatInt(fr.ActualFileSize, 10),
		fr.ActualFileHash,
		strconv.FormatInt(fr.ChunkSize, 10),
	)

	return strings.Join(hashArray, ":")
}

func (fr *blobberRef) CalculateFileHash() string {
	fr.Hash = encryption.Hash(fr.GetFileHashData())
	fr.LookupHash = GetReferenceLookup(fr.AllocationID, fr.Path)

	return fr.Hash
}

func (r *blobberRef) CalculateDirHash() string {
	// empty directory, return hash directly
	if len(r.Children) == 0 && !r.childrenLoaded {
		return r.Hash
	}
	for _, childRef := range r.Children {
		childRef.CalculateHash()
	}
	childHashes := make([]string, len(r.Children))

	var size int64
	for index, childRef := range r.Children {
		childHashes[index] = childRef.Hash

		size += childRef.Size
	}

	r.Hash = encryption.Hash(strings.Join(childHashes, ":"))

	r.Size = size

	r.LookupHash = GetReferenceLookup(r.AllocationID, r.Path)

	return r.Hash
}

func (r *blobberRef) CalculateHash() string {
	if r.Type == DIRECTORY {
		return r.CalculateDirHash()
	}
	return r.CalculateFileHash()
}
