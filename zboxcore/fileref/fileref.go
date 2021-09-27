package fileref

import (
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
)

const CHUNK_SIZE = 64 * 1024

const (
	FILE      = "f"
	DIRECTORY = "d"
)

type CommitMetaTxn struct {
	RefID     int64  `json:"ref_id"`
	TxnID     string `json:"txn_id"`
	CreatedAt string `json:"created_at"`
}

type Collaborator struct {
	RefID     int64  `json:"ref_id"`
	ClientID  string `json:"client_id"`
	CreatedAt string `json:"created_at"`
}

// The Attributes represents file attributes.
type Attributes struct {
	// The WhoPaysForReads represents reading payer. It can be allocation owner
	// or a 3rd party user. It affects read operations only. It requires
	// blobbers to be trusted.
	WhoPaysForReads common.WhoPays `json:"who_pays_for_reads,omitempty"`

	// add more file / directory attributes by needs with
	// 'omitempty' json tag to avoid hash difference for
	// equal values
}

// IsZero returns true, if the Attributes is zero.
func (a *Attributes) IsZero() bool {
	return (*a) == (Attributes{})
}

// Validate the Attributes.
func (a *Attributes) Validate() (err error) {
	if err = a.WhoPaysForReads.Validate(); err != nil {
		return errors.Wrap(err, "invalid who_pays_for_reads field")
	}
	return
}

type FileRef struct {
	Ref                 `json:",squash"`
	CustomMeta          string          `json:"custom_meta"`
	ContentHash         string          `json:"content_hash"`
	MerkleRoot          string          `json:"merkle_root"`
	ThumbnailSize       int64           `json:"thumbnail_size"`
	ThumbnailHash       string          `json:"thumbnail_hash"`
	ActualFileSize      int64           `json:"actual_file_size"`
	ActualFileHash      string          `json:"actual_file_hash"`
	ActualThumbnailSize int64           `json:"actual_thumbnail_size"`
	ActualThumbnailHash string          `json:"actual_thumbnail_hash"`
	MimeType            string          `json:"mimetype"`
	EncryptedKey        string          `json:"encrypted_key"`
	CommitMetaTxns      []CommitMetaTxn `json:"commit_meta_txns"`
	Collaborators       []Collaborator  `json:"collaborators"`
	Attributes          Attributes      `json:"attributes"`
}

type RefEntity interface {
	GetNumBlocks() int64
	GetSize() int64
	GetHash() string
	CalculateHash() string
	GetType() string
	GetPathHash() string
	GetLookupHash() string
	GetPath() string
	GetName() string
	GetAttributes() Attributes
	GetCreatedAt() string
	GetUpdatedAt() string
}

type Ref struct {
	Type           string     `json:"type"`
	AllocationID   string     `json:"allocation_id"`
	Name           string     `json:"name"`
	Path           string     `json:"path"`
	Size           int64      `json:"size"`
	ActualSize     int64      `json:"actual_file_size"`
	Hash           string     `json:"hash"`
	ChunkSize      int64      `json:"chunk_size"`
	NumBlocks      int64      `json:"num_of_blocks"`
	PathHash       string     `json:"path_hash"`
	LookupHash     string     `json:"lookup_hash"`
	Attributes     Attributes `json:"attributes"`
	childrenLoaded bool
	Children       []RefEntity `json:"-"`
	CreatedAt      string      `json:"created_at"`
	UpdatedAt      string      `json:"updated_at"`
}

func GetReferenceLookup(allocationID string, path string) string {
	return encryption.Hash(allocationID + ":" + path)
}

func (r *Ref) CalculateHash() string {
	if len(r.Children) == 0 && !r.childrenLoaded {
		return r.Hash
	}
	sort.SliceStable(r.Children, func(i, j int) bool {
		return strings.Compare(GetReferenceLookup(r.AllocationID, r.Children[i].GetPath()), GetReferenceLookup(r.AllocationID, r.Children[j].GetPath())) == -1
	})
	for _, childRef := range r.Children {
		childRef.CalculateHash()
	}
	childHashes := make([]string, len(r.Children))
	childPathHashes := make([]string, len(r.Children))
	var refNumBlocks int64
	var size int64
	for index, childRef := range r.Children {
		childHashes[index] = childRef.GetHash()
		childPathHashes[index] = childRef.GetPathHash()
		refNumBlocks += childRef.GetNumBlocks()
		size += childRef.GetSize()
	}
	// fmt.Println("ref name and path, hash :" + r.Name + " " + r.Path + " " + r.Hash)
	// fmt.Println("ref hash data: " + strings.Join(childHashes, ":"))
	r.Hash = encryption.Hash(strings.Join(childHashes, ":"))
	// fmt.Println("ref hash : " + r.Hash)

	r.NumBlocks = refNumBlocks
	r.Size = size

	//fmt.Println("Ref Path hash: " + strings.Join(childPathHashes, ":"))
	r.PathHash = encryption.Hash(strings.Join(childPathHashes, ":"))

	return r.Hash
}

func (r *Ref) GetHash() string {
	return r.Hash
}

func (r *Ref) GetType() string {
	return r.Type
}

func (r *Ref) GetNumBlocks() int64 {
	return r.NumBlocks
}

func (r *Ref) GetSize() int64 {
	return r.Size
}

func (r *Ref) GetPathHash() string {
	return r.PathHash
}

func (r *Ref) GetLookupHash() string {
	return r.LookupHash
}

func (r *Ref) GetPath() string {
	return r.Path
}

func (r *Ref) GetName() string {
	return r.Name
}

func (r *Ref) GetAttributes() Attributes {
	return r.Attributes
}

func (r *Ref) GetCreatedAt() string {
	return r.CreatedAt
}

func (r *Ref) GetUpdatedAt() string {
	return r.UpdatedAt
}

func (r *Ref) AddChild(child RefEntity) {
	if r.Children == nil {
		r.Children = make([]RefEntity, 0)
	}
	r.Children = append(r.Children, child)
	sort.SliceStable(r.Children, func(i, j int) bool {
		return strings.Compare(GetReferenceLookup(r.AllocationID, r.Children[i].GetPath()), GetReferenceLookup(r.AllocationID, r.Children[j].GetPath())) == -1
	})
	r.childrenLoaded = true
}

func (r *Ref) RemoveChild(idx int) {
	if idx < 0 {
		return
	}
	r.Children = append(r.Children[:idx], r.Children[idx+1:]...)
	sort.SliceStable(r.Children, func(i, j int) bool {
		return strings.Compare(GetReferenceLookup(r.AllocationID, r.Children[i].GetPath()), GetReferenceLookup(r.AllocationID, r.Children[j].GetPath())) == -1
	})
}

func (fr *FileRef) GetHashData() string {
	hashArray := make([]string, 0)
	hashArray = append(hashArray, fr.AllocationID)
	hashArray = append(hashArray, fr.Type)
	hashArray = append(hashArray, fr.Name)
	hashArray = append(hashArray, fr.Path)
	hashArray = append(hashArray, strconv.FormatInt(fr.Size, 10))
	hashArray = append(hashArray, fr.ContentHash)
	hashArray = append(hashArray, fr.MerkleRoot)
	hashArray = append(hashArray, strconv.FormatInt(fr.ActualFileSize, 10))
	hashArray = append(hashArray, fr.ActualFileHash)
	var attrs, _ = json.Marshal(&fr.Attributes)
	hashArray = append(hashArray, string(attrs))
	hashArray = append(hashArray, strconv.FormatInt(fr.ChunkSize, 10))
	return strings.Join(hashArray, ":")
}

func (fr *FileRef) GetHash() string {
	return fr.Hash
}

func (fr *FileRef) CalculateHash() string {
	// fmt.Println("fileref name , path, hash", fr.Name, fr.Path, fr.Hash)
	// fmt.Println("Fileref hash data: " + fr.GetHashData())
	fr.Hash = encryption.Hash(fr.GetHashData())
	// fmt.Println("Fileref hash : " + fr.Hash)
	fr.NumBlocks = int64(math.Ceil(float64(fr.Size*1.0) / CHUNK_SIZE))
	fr.PathHash = GetReferenceLookup(fr.AllocationID, fr.Path)
	return fr.Hash
}

func (fr *FileRef) GetType() string {
	return fr.Type
}

func (fr *FileRef) GetNumBlocks() int64 {
	return fr.NumBlocks
}

func (fr *FileRef) GetSize() int64 {
	return fr.Size
}

func (fr *FileRef) GetPathHash() string {
	return fr.PathHash
}

func (fr *FileRef) GetLookupHash() string {
	return fr.LookupHash
}

func (fr *FileRef) GetPath() string {
	return fr.Path
}
func (fr *FileRef) GetName() string {
	return fr.Name
}

func (fr *FileRef) GetAttributes() Attributes {
	return fr.Attributes
}

func (fr *FileRef) GetCreatedAt() string {
	return fr.CreatedAt
}

func (fr *FileRef) GetUpdatedAt() string {
	return fr.UpdatedAt
}
