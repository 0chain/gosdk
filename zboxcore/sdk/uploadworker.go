package sdk

import (
	"hash"
	"io"
	"sync"

	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type UploadFileMeta struct {
	// Name remote file name
	Name string
	// Path remote path
	Path string
	// Hash hash of entire source file
	Hash     string
	MimeType string
	// Size total bytes of entire source file
	Size int64

	// ThumbnailSize total bytes of entire thumbnail
	ThumbnailSize int64
	// ThumbnailHash hash code of entire thumbnail
	ThumbnailHash string
}

type UploadResult struct {
	Filename   string `json:"filename"`
	ShardSize  int64  `json:"size"`
	Hash       string `json:"content_hash,omitempty"`
	MerkleRoot string `json:"merkle_root,omitempty"`
}

type UploadRequest struct {
	filepath          string
	thumbnailpath     string
	remotefilepath    string
	statusCallback    StatusCallback
	fileHash          hash.Hash
	fileHashWr        io.Writer
	thumbnailHash     hash.Hash
	thumbnailHashWr   io.Writer
	file              []*fileref.FileRef
	filemeta          *UploadFileMeta
	remaining         int64
	thumbRemaining    int64
	wg                *sync.WaitGroup
	uploadDataCh      []chan []byte
	uploadThumbCh     []chan []byte
	isRepair          bool
	isUpdate          bool
	connectionID      string
	datashards        int
	parityshards      int
	uploadMask        zboxutil.Uint128
	isEncrypted       bool
	encscheme         encryption.EncryptionScheme
	isUploadCanceled  bool
	completedCallback func(filepath string)
	err               error
	Consensus
}
