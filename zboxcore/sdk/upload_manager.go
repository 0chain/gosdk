package sdk

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"golang.org/x/crypto/sha3"
)

var (
	uploadManagers      = make(map[string]*UploadManager)
	uploadManagersMutex sync.Mutex
)

//GetOrCreateUploadManager get or create a UploadManager for blobber
func GetOrCreateUploadManager(req *UploadRequest) *UploadManager {
	uploadManagersMutex.Lock()
	defer uploadManagersMutex.Unlock()

	um, ok := uploadManagers[req.filepath]

	//create a new UploadManager if file is new or changed
	if !ok || (um.size != req.filemeta.Size || um.thumbnailSize != req.filemeta.ThumbnailSize) {
		um = &UploadManager{
			localPath: req.filepath,
			size:      req.filemeta.Size,
			fileBytes: make([]byte, 0, req.filemeta.Size),

			localThumbnailPath: req.thumbnailpath,
			thumbnailSize:      req.filemeta.ThumbnailSize,
			thumbnailBytes:     make([]byte, 0, req.filemeta.ThumbnailSize),
			blobbers:           make(map[string]*UploadProgress),
		}
		uploadManagers[req.filepath] = um
	}

	return um
}

//RemoveUploadManager remove UploadManager for filepath
func RemoveUploadManager(filepath string) {
	uploadManagersMutex.Lock()
	defer uploadManagersMutex.Unlock()

	delete(uploadManagers, filepath)
}

//UploadManager a upload manager for retry and resum uploading
type UploadManager struct {
	sync.RWMutex

	isLoading bool
	isLoaded  bool //content is full-loaded

	localPath string //local file
	size      int64  //size of local file
	fileBytes []byte

	localThumbnailPath string //local thumbnail
	thumbnailSize      int64  //size of local thumbnail
	thumbnailBytes     []byte

	//file meta data
	Filename      string `json:"filename"`
	Path          string `json:"filepath"`
	Hash          string `json:"content_hash,omitempty"`
	ThumbnailHash string `json:"thumbnail_content_hash,omitempty"`

	MerkleRoot          string             `json:"merkle_root,omitempty"`
	ActualHash          string             `json:"actual_hash"`
	ActualSize          int64              `json:"actual_size"`
	ActualThumbnailSize int64              `json:"actual_thumb_size"`
	ActualThumbnailHash string             `json:"actual_thumb_hash"`
	MimeType            string             `json:"mimetype"`
	CustomMeta          string             `json:"custom_meta,omitempty"`
	EncryptedKey        string             `json:"encrypted_key,omitempty"`
	Attributes          fileref.Attributes `json:"attributes,omitempty"`
	ThumbnailSize       int64
	ShardSize           int64

	blobbers map[string]*UploadProgress
}

//Load load local files into memory first.
//TODO: it can be updated with lazy load/stream mode for memory performance
func (um *UploadManager) Load(req *UploadRequest, a *Allocation, file *fileref.FileRef, uploadCh chan []byte, uploadThumbCh chan []byte) {

	fileBytes := make([]byte, 0, len(um.fileBytes))
	thumbnailBytes := make([]byte, 0, len(um.thumbnailBytes))

	shardSize := (req.filemeta.Size + int64(a.DataShards) - 1) / int64(a.DataShards)
	chunkSizeWithHeader := int64(fileref.CHUNK_SIZE)
	if req.isEncrypted {
		chunkSizeWithHeader -= 16
		chunkSizeWithHeader -= 2 * 1024
	}
	chunksPerShard := (shardSize + chunkSizeWithHeader - 1) / chunkSizeWithHeader
	if req.isEncrypted {
		shardSize += chunksPerShard * (16 + (2 * 1024))
	}
	thumbnailSize := int64(0)
	remaining := shardSize

	fileMerkleRoot := ""
	fileContentHash := ""
	thumbContentHash := ""

	if um.isLoaded {
		//TODO: it hits performance, allocationObj.UploadFile should be refactored
		// Just read the data to unblock
		for remaining > 0 {
			dataBytes := <-uploadCh
			remaining = remaining - int64(len(dataBytes))
		}
		_ = <-uploadCh
		return
	}

	// Setup file hash compute
	h := sha1.New()
	//merkleHash := sha3.New256()
	hWr := io.MultiWriter(h)
	merkleHashes := make([]hash.Hash, 1024)
	merkleLeaves := make([]util.Hashable, 1024)
	for idx := range merkleHashes {
		merkleHashes[idx] = sha3.New256()
	}
	// Read the data
	for remaining > 0 {
		dataBytes, ok := <-uploadCh
		if !ok {
			return
		}
		fileBytes = append(fileBytes, dataBytes...)

		hWr.Write(dataBytes)
		merkleChunkSize := 64
		for i := 0; i < len(dataBytes); i += merkleChunkSize {
			end := i + merkleChunkSize
			if end > len(dataBytes) {
				end = len(dataBytes)
			}
			offset := i / merkleChunkSize
			merkleHashes[offset].Write(dataBytes[i:end])
		}

		remaining = remaining - int64(len(dataBytes))

	}
	for idx := range merkleHashes {
		merkleLeaves[idx] = util.NewStringHashable(hex.EncodeToString(merkleHashes[idx].Sum(nil)))
	}
	var mt util.MerkleTreeI = &util.MerkleTree{}
	mt.ComputeTree(merkleLeaves)
	if !req.isRepair {
		//Read last push "done" from channel that is pushed completePush
		// Wait for file hash to be ready
		// Logger.Debug("Waiting for file hash....")
		_ = <-uploadCh
		//Logger.Debug("File Hash ready", obj.file.Hash)
	}
	fileContentHash = hex.EncodeToString(h.Sum(nil))
	fileMerkleRoot = mt.GetRoot()

	if len(req.thumbnailpath) > 0 {
		thumbnailSize = (req.filemeta.ThumbnailSize + int64(a.DataShards) - 1) / int64(a.DataShards)
		chunkSizeWithHeader := int64(fileref.CHUNK_SIZE)
		if req.isEncrypted {
			chunkSizeWithHeader -= 16
			chunkSizeWithHeader -= 2 * 1024
		}
		chunksPerShard := (thumbnailSize + chunkSizeWithHeader - 1) / chunkSizeWithHeader
		if req.isEncrypted {
			thumbnailSize += chunksPerShard * (16 + (2 * 1024))
		}
		remaining := thumbnailSize

		// Setup file hash compute
		h := sha1.New()
		hWr := io.MultiWriter(h)
		// Read the data
		for remaining > 0 {
			dataBytes, ok := <-uploadThumbCh
			if !ok {
				return
			}
			thumbnailBytes = append(thumbnailBytes, dataBytes...)
			hWr.Write(dataBytes)
			remaining = remaining - int64(len(dataBytes))
		}
		if !req.isRepair {
			// Wait for file hash to be ready
			// Logger.Debug("Waiting for file hash....")
			_ = <-uploadThumbCh
			// Logger.Debug("File Hash ready", obj.file.Hash)
		}
		thumbContentHash = hex.EncodeToString(h.Sum(nil))
	}

	um.Lock()
	//TODO: bytes are repeated too many times, it is not bad. Almost of channels should be removed
	if !um.isLoaded {
		um.fileBytes = fileBytes
		um.thumbnailBytes = thumbnailBytes

		um.Filename = file.Name
		um.Path = file.Path
		um.ActualHash = req.filemeta.Hash
		um.ActualSize = req.filemeta.Size
		um.ActualThumbnailHash = req.filemeta.ThumbnailHash
		um.ActualThumbnailSize = req.filemeta.ThumbnailSize
		um.MimeType = req.filemeta.MimeType
		um.Attributes = req.filemeta.Attributes
		um.Hash = fileContentHash
		um.ThumbnailHash = thumbContentHash
		um.ThumbnailSize = thumbnailSize
		um.MerkleRoot = fileMerkleRoot
		um.ShardSize = shardSize

		if req.isEncrypted {
			um.EncryptedKey = req.encscheme.GetEncryptedKey()
		}

		um.isLoaded = true
	}

	um.Unlock()
}

//Create get or create a UploadProgress for a blobber
func (um *UploadManager) Create(blobber *blockchain.StorageNode, connectionID string) *UploadProgress {
	um.Lock()
	defer um.Unlock()

	progress, ok := um.blobbers[blobber.ID]

	if !ok {
		progress = &UploadProgress{
			UploadOffset: 0,
			UploadLenght: um.size,
			Blobber:      *blobber,
			ConnectionID: connectionID,
		}

		um.blobbers[blobber.ID] = progress
	}

	return progress
}

//Done remove UploadProgress
func (um *UploadManager) Done(blobber blockchain.StorageNode) {
	um.Lock()
	defer um.Unlock()

	delete(um.blobbers, blobber.ID)
}

//Start start or resume upload
func (um *UploadManager) Start(up *UploadProgress, req *UploadRequest, a *Allocation, file *fileref.FileRef) {
	//UploadProgress must be multi-thread safe.
	up.Lock()
	defer up.Unlock()

	//first requse, upload first chunk and thunmbnail if it has
	max := int64(len(um.fileBytes))

	isFinal := false
	sent := 0
	for up.UploadOffset < max {

		offset := up.UploadOffset + CHUNK_SIZE

		if offset > max {
			offset = max
			isFinal = true
		}

		fileBytes := um.fileBytes[up.UploadOffset:offset]

		if up.UploadOffset == 0 {
			if err := up.UploadChunk(um, isFinal, up.UploadOffset, fileBytes, um.thumbnailBytes, req, a, file); err != nil {
				return
			}
		} else {
			if err := up.UploadChunk(um, isFinal, up.UploadOffset, fileBytes, nil, req, a, file); err != nil {
				return
			}
		}
		sent = sent + len(fileBytes)
		if req.statusCallback != nil {
			req.statusCallback.InProgress(a.ID, req.remotefilepath, OpUpload, sent*(a.DataShards+a.ParityShards), nil)
		}

		up.UploadOffset = offset

	}

	req.consensus++

	um.Done(up.Blobber)

	// if up.UploadOffset < um.size {

	// }

	// up.UploadOffset

}

// //Cancel cancel a upload.
// func (um *UploadManager) Cancel() {

// }

//UploadProgress upload stats for blobber
type UploadProgress struct {
	sync.Mutex
	Blobber blockchain.StorageNode

	UploadOffset int64
	UploadLenght int64
	ConnectionID string
}

//UploadChunk upload a chunk
func (up *UploadProgress) UploadChunk(um *UploadManager, isFinal bool, offset int64, fileBytes, thumbnailBytes []byte, req *UploadRequest, a *Allocation, file *fileref.FileRef) error {
	bodyReader, bodyWriter := io.Pipe()
	formWriter := multipart.NewWriter(bodyWriter)
	httpreq, _ := zboxutil.NewUploadRequest(up.Blobber.Baseurl, a.Tx, bodyReader, req.isUpdate)

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	var formData uploadFormData

	go func() {

		fileField, err := formWriter.CreateFormFile("uploadFile", file.Name)
		if err != nil {
			Logger.Error("Create form failed: ", err)
			bodyWriter.CloseWithError(err)

			return
		}

		h1 := sha1.New()
		hWr1 := io.MultiWriter(h1)

		fileField.Write(fileBytes)
		hWr1.Write(fileBytes)

		fileContentHash := hex.EncodeToString(h1.Sum(nil))

		if req.statusCallback != nil {
			req.statusCallback.InProgress(a.ID, req.remotefilepath, OpUpload, len(fileBytes)*(a.DataShards+a.ParityShards), nil)
		}

		var thumbContentHash string
		if len(thumbnailBytes) > 0 {

			fileField, err := formWriter.CreateFormFile("uploadThumbnailFile", file.Name+".thumb")
			if err != nil {
				Logger.Error("Create form failed: ", err)
				return
			}

			h2 := sha1.New()
			hWr2 := io.MultiWriter(h2)

			fileField.Write(thumbnailBytes)
			hWr2.Write(thumbnailBytes)

			thumbContentHash = hex.EncodeToString(h2.Sum(nil))
		}

		formData = uploadFormData{
			IsResumable:  true,
			UploadOffset: offset,
			UploadLength: um.ActualSize,

			ConnectionID:        up.ConnectionID,        //use old connectionid to resume upload
			Filename:            um.Filename,            //  file.Name,
			Path:                um.Path,                //  file.Path,
			ActualHash:          um.ActualHash,          //  req.filemeta.Hash,
			ActualSize:          um.ActualSize,          //  req.filemeta.Size,
			ActualThumbnailHash: um.ActualThumbnailHash, // req.filemeta.ThumbnailHash,
			ActualThumbnailSize: um.ActualThumbnailSize, // req.filemeta.ThumbnailSize,
			MimeType:            um.MimeType,            // req.filemeta.MimeType,
			Attributes:          um.Attributes,          //  req.filemeta.Attributes,
			Hash:                fileContentHash,
			ThumbnailHash:       thumbContentHash,
			MerkleRoot:          um.MerkleRoot,
		}

		if isFinal { //last chunk, use hash of whole file instead
			formData.Hash = um.Hash
			formData.IsFinal = true
		}

		if req.isEncrypted {
			formData.EncryptedKey = um.EncryptedKey
		}
		_ = formWriter.WriteField("connection_id", req.connectionID)
		var metaData []byte
		metaData, err = json.Marshal(formData)
		// Logger.Debug("Upload with",string(metaData))
		if err == nil {
			if req.isUpdate {
				_ = formWriter.WriteField("updateMeta", string(metaData))
			} else {
				_ = formWriter.WriteField("uploadMeta", string(metaData))
			}

			bodyWriter.CloseWithError(formWriter.Close())
		}
	}()

	return zboxutil.HttpDo(a.ctx, a.ctxCancelF, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Upload : ", err)
			req.err = err
			return err
		}
		defer resp.Body.Close()

		respbody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Error("Error: Resp ", err)
			req.err = err
			return err
		}
		if resp.StatusCode != http.StatusOK {
			Logger.Error(up.Blobber.Baseurl, " Upload error response: ", resp.StatusCode, string(respbody))
			req.err = fmt.Errorf(string(respbody))
			return err
		}
		var r uploadResult
		err = json.Unmarshal(respbody, &r)
		if err != nil {
			Logger.Error(up.Blobber.Baseurl, " Upload response parse error: ", err)
			req.err = err
			return err
		}
		if r.Filename != formData.Filename || r.ShardSize != um.ShardSize ||
			r.Hash != formData.Hash || r.MerkleRoot != formData.MerkleRoot {
			err = fmt.Errorf(up.Blobber.Baseurl, "Unexpected upload response data", string(respbody))
			Logger.Error(err)
			req.err = err
			return err
		}

		Logger.Info(up.Blobber.Baseurl, formData.Path, " uploaded")
		file.MerkleRoot = formData.MerkleRoot
		file.ContentHash = formData.Hash
		file.ThumbnailHash = formData.ThumbnailHash
		file.ThumbnailSize = um.ThumbnailSize
		file.Size = um.ShardSize
		file.Path = formData.Path
		file.ActualFileHash = formData.ActualHash
		file.ActualFileSize = formData.ActualSize
		file.ActualThumbnailHash = formData.ActualThumbnailHash
		file.ActualThumbnailSize = formData.ActualThumbnailSize
		file.EncryptedKey = formData.EncryptedKey
		file.CalculateHash()
		return nil
	})
}
