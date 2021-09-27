package sdk

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encoder"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"golang.org/x/crypto/sha3"
)

func isSetTestEnv(name string) bool {
	return os.Getenv("INTEGRATION_TESTS_"+name) != ""
}

// Expected success rate is calculated (NumDataShards)*100/(NumDataShards+NumParityShards)
// Additional success percentage on top of expected success rate
const additionalSuccessRate = (10)

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

	// Attributes file attributes in blockchain
	Attributes fileref.Attributes
}

type uploadFormData struct {
	ConnectionID string `json:"connection_id"`
	// Filename remote file name
	Filename string `json:"filename"`
	// Path remote path
	Path string `json:"filepath"`

	// Hash hash of shard data (encoded, encrypted)
	Hash string `json:"content_hash,omitempty"`
	// Hash hash of shard thumbnail (encoded, encrypted)
	ThumbnailHash string `json:"thumbnail_content_hash,omitempty"`

	// MerkleRoot merkle's root hash of shard data (encoded, encrypted)
	MerkleRoot string `json:"merkle_root,omitempty"`

	// ActualHash hash of orignial file (unencoded, unencrypted)
	ActualHash string `json:"actual_hash"`
	// ActualSize total bytes of orignial file (unencoded, unencrypted)
	ActualSize int64 `json:"actual_size"`
	// ActualThumbnailSize total bytes of orignial thumbnail (unencoded, unencrypted)
	ActualThumbnailSize int64 `json:"actual_thumb_size"`
	// ActualThumbnailHash hash of orignial thumbnail (unencoded, unencrypted)
	ActualThumbnailHash string `json:"actual_thumb_hash"`

	MimeType     string             `json:"mimetype"`
	CustomMeta   string             `json:"custom_meta,omitempty"`
	EncryptedKey string             `json:"encrypted_key,omitempty"`
	Attributes   fileref.Attributes `json:"attributes,omitempty"`
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

func (req *UploadRequest) setUploadMask(numBlobbers int) {
	req.uploadMask = zboxutil.NewUint128(1).Lsh(uint64(numBlobbers)).Sub64(1)
}

func (req *UploadRequest) prepareUpload(
	a *Allocation,
	blobber *blockchain.StorageNode,
	file *fileref.FileRef,
	uploadCh chan []byte,
	uploadThumbCh chan []byte,
	wg *sync.WaitGroup,
) {
	bodyReader, bodyWriter := io.Pipe()
	formWriter := multipart.NewWriter(bodyWriter)
	httpreq, _ := zboxutil.NewUploadRequest(blobber.Baseurl, a.Tx, bodyReader, req.isUpdate)
	//timeout := time.Duration(int64(math.Max(10, float64(obj.file.Size)/(CHUNK_SIZE*float64(len(obj.blobbers)/2)))))
	//ctx, cncl := context.WithTimeout(context.Background(), (time.Second * timeout))

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	var formData uploadFormData
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
	sent := 0
	go func() {
		fileMerkleRoot := ""
		fileContentHash := ""
		thumbContentHash := ""
		var (
			fileField io.Writer
			err       error
		)
		fileField, err = formWriter.CreateFormFile("uploadFile", file.Name)
		if err != nil {
			Logger.Error("Create form failed: ", err)
			bodyWriter.CloseWithError(err)
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
			fileField.Write(dataBytes)
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
			sent = sent + len(dataBytes)
			if req.statusCallback != nil {
				req.statusCallback.InProgress(a.ID, req.remotefilepath, OpUpload, sent*(a.DataShards+a.ParityShards), nil)
			}
		}
		for idx := range merkleHashes {
			merkleLeaves[idx] = util.NewStringHashable(hex.EncodeToString(merkleHashes[idx].Sum(nil)))
		}
		var mt util.MerkleTreeI = &util.MerkleTree{}
		mt.ComputeTree(merkleLeaves)
		if !req.isRepair {
			// Wait for file hash to be ready
			// Logger.Debug("Waiting for file hash....")
			_ = <-uploadCh
			// Logger.Debug("File Hash ready", obj.file.Hash)
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

			fileField, err := formWriter.CreateFormFile("uploadThumbnailFile", file.Name+".thumb")
			if err != nil {
				Logger.Error("Create form failed: ", err)
				return
			}
			// Setup file hash compute
			h := sha1.New()

			hWr := io.MultiWriter(h)
			// Read the data
			for remaining > 0 {
				dataBytes, ok := <-uploadThumbCh
				if !ok {
					return
				}
				fileField.Write(dataBytes)
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

		formData = uploadFormData{
			ConnectionID:        req.connectionID,
			Filename:            file.Name,
			Path:                file.Path,
			ActualHash:          req.filemeta.Hash,
			ActualSize:          req.filemeta.Size,
			ActualThumbnailHash: req.filemeta.ThumbnailHash,
			ActualThumbnailSize: req.filemeta.ThumbnailSize,
			MimeType:            req.filemeta.MimeType,
			Attributes:          req.filemeta.Attributes,
			Hash:                fileContentHash,
			ThumbnailHash:       thumbContentHash,
			MerkleRoot:          fileMerkleRoot,
		}
		if req.isEncrypted {
			formData.EncryptedKey = req.encscheme.GetEncryptedKey()
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
	_ = zboxutil.HttpDo(a.ctx, a.ctxCancelF, httpreq, func(resp *http.Response, err error) error {
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
			Logger.Error(blobber.Baseurl, " Upload error response: ", resp.StatusCode, string(respbody))
			req.err = errors.New("", string(respbody))
			return err
		}
		var r UploadResult
		err = json.Unmarshal(respbody, &r)
		if err != nil {
			Logger.Error(blobber.Baseurl, " Upload response parse error: ", err)
			req.err = err
			return err
		}
		if r.Filename != formData.Filename || r.ShardSize != shardSize ||
			r.Hash != formData.Hash || r.MerkleRoot != formData.MerkleRoot {
			err = fmt.Errorf(blobber.Baseurl, "Unexpected upload response data", string(respbody))
			Logger.Error(err)
			req.err = err
			return err
		}
		req.consensus++
		Logger.Info(blobber.Baseurl, formData.Path, " uploaded")
		file.MerkleRoot = formData.MerkleRoot
		file.ContentHash = formData.Hash
		file.ThumbnailHash = formData.ThumbnailHash
		file.ThumbnailSize = thumbnailSize
		file.Size = shardSize
		file.Path = formData.Path
		file.ActualFileHash = formData.ActualHash
		file.ActualFileSize = formData.ActualSize
		file.ActualThumbnailHash = formData.ActualThumbnailHash
		file.ActualThumbnailSize = formData.ActualThumbnailSize
		file.EncryptedKey = formData.EncryptedKey
		file.CalculateHash()
		return nil
	})
	wg.Done()
}

// setups upload for each blobber with same file
func (req *UploadRequest) setupUpload(a *Allocation) error {
	numUploads := req.uploadMask.CountOnes()
	req.uploadDataCh = make([]chan []byte, numUploads)
	req.uploadThumbCh = make([]chan []byte, numUploads)
	req.file = make([]*fileref.FileRef, numUploads)

	for i := range req.uploadDataCh {
		req.uploadDataCh[i] = make(chan []byte)
		req.uploadThumbCh[i] = make(chan []byte)
		req.file[i] = &fileref.FileRef{}
		req.file[i].Name = req.filemeta.Name
		req.file[i].Path = req.remotefilepath
		req.file[i].Type = fileref.FILE
		req.file[i].AllocationID = a.ID
		req.file[i].Attributes = req.filemeta.Attributes
	}

	if !req.isRepair {
		req.fileHash = sha1.New()
		req.fileHashWr = io.MultiWriter(req.fileHash)
		req.thumbnailHash = sha1.New()
		req.thumbnailHashWr = io.MultiWriter(req.thumbnailHash)
	}
	if req.isEncrypted {
		req.encscheme = encryption.NewEncryptionScheme()
		mnemonic := client.GetClient().Mnemonic
		_, err := req.encscheme.Initialize(mnemonic)
		if err != nil {
			return err
		}
		req.encscheme.InitForEncryption("filetype:audio")
	}

	req.wg = &sync.WaitGroup{}
	req.wg.Add(numUploads)
	req.consensus = 0

	// Start upload for each blobber
	var c, pos uint64 = 0, 0
	for i := req.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		go req.prepareUpload(a, a.Blobbers[pos], req.file[c], req.uploadDataCh[c], req.uploadThumbCh[c], req.wg)
		c++
	}
	return nil
}

// this will push the data to the blobber, encrypt if the using the encrypted upload mode
func (req *UploadRequest) pushData(data []byte) error {
	//TODO: Check for optimization
	n := int64(math.Min(float64(req.remaining), float64(len(data))))
	if !req.isRepair {
		req.fileHashWr.Write(data[:n])
	}
	req.remaining = req.remaining - n
	erasureencoder, err := encoder.NewEncoder(req.datashards, req.parityshards)
	if err != nil {
		return err
	}
	shards, err := erasureencoder.Encode(data)
	if err != nil {
		Logger.Error("Erasure coding failed.", err.Error())
		return err
	}
	var c, pos uint64 = 0, 0
	if req.isEncrypted {
		for i := req.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			encMsg, err := req.encscheme.Encrypt(shards[pos])
			if err != nil {
				Logger.Error("Encryption failed.", err.Error())
				return err
			}
			header := make([]byte, 2*1024)
			copy(header[:], encMsg.MessageChecksum+","+encMsg.OverallChecksum)
			shards[pos] = append(header, encMsg.EncryptedData...)
			c++
		}

		c, pos = 0, 0
	}
	for i := req.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		req.uploadDataCh[c] <- shards[pos]
		c++
	}

	return nil
}

func (req *UploadRequest) completePush() error {
	if !req.isRepair {
		req.filemeta.Hash = hex.EncodeToString(req.fileHash.Sum(nil))
		//fmt.Println("req.filemeta.Hash=" + req.filemeta.Hash)
		var c, pos uint64 = 0, 0
		for i := req.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			req.uploadDataCh[c] <- []byte("done")
			c++
		}
	}
	req.wg.Wait()
	if !req.isConsensusOk() {
		return fmt.Errorf("Upload failed: Consensus_rate:%f, expected:%f", req.getConsensusRate(), req.getConsensusRequiredForOk())
	}
	return nil
}

func (req *UploadRequest) processUpload(ctx context.Context, a *Allocation) {
	if req.completedCallback != nil {
		defer req.completedCallback(req.filepath)
	}

	var inFile *os.File
	inFile, err := os.Open(req.filepath)
	if err != nil && req.statusCallback != nil {
		req.statusCallback.Error(a.ID, req.filepath, OpUpload, errors.New("open_file_failed", err.Error()))
		return
	}
	defer inFile.Close()
	mimetype, err := zboxutil.GetFileContentType(inFile)
	if err != nil && req.statusCallback != nil {
		req.statusCallback.Error(a.ID, req.filepath, OpUpload, errors.New("mime_type_error", err.Error()))
		return
	}
	req.filemeta.MimeType = mimetype
	err = req.setupUpload(a)
	if err != nil && req.statusCallback != nil {
		req.statusCallback.Error(a.ID, req.filepath, OpUpload, errors.New("setup_upload_failed", err.Error()))
		return
	}
	size := req.filemeta.Size
	// Calculate number of bytes per shard.
	perShard := (size + int64(a.DataShards) - 1) / int64(a.DataShards)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	if len(req.thumbnailpath) > 0 {
		wg.Add(1)
		go req.processThumbnail(a, wg)
	}
	go func() {
		defer wg.Done()
		// Pad data to Shards*perShard.
		padding := make([]byte, (int64(a.DataShards)*perShard)-size)
		dataReader := io.MultiReader(inFile, bytes.NewBuffer(padding))
		chunkSizeWithHeader := int64(fileref.CHUNK_SIZE)
		if req.isEncrypted {
			chunkSizeWithHeader -= 16
			chunkSizeWithHeader -= 2 * 1024
		}
		chunksPerShard := (perShard + chunkSizeWithHeader - 1) / chunkSizeWithHeader
		Logger.Info("Size:", size, " perShard:", perShard, " chunks/shard:", chunksPerShard)
		req.isUploadCanceled = false
		if req.statusCallback != nil {
			req.statusCallback.Started(a.ID, req.remotefilepath, OpUpload, int(perShard)*(a.DataShards+a.ParityShards))
		}

		for ctr := int64(0); ctr < chunksPerShard; ctr++ {
			remaining := int64(math.Min(float64(perShard-(ctr*chunkSizeWithHeader)), float64(chunkSizeWithHeader)))
			b1 := make([]byte, remaining*int64(a.DataShards))
			_, err = dataReader.Read(b1)
			if err != nil && req.statusCallback != nil {
				req.statusCallback.Error(a.ID, req.filepath, OpUpload, errors.New("read_failed", err.Error()))
				return
			}
			if req.isUploadCanceled {
				req.isUploadCanceled = false
				if !req.isUpdate && !req.isRepair {
					go a.DeleteFile(req.remotefilepath)
				}
				if req.statusCallback != nil {
					req.statusCallback.Error(a.ID, req.filepath, OpUpload, errors.New("user_aborted", "Upload aborted by user"))
				}
				return
			}
			err = req.pushData(b1)
			if err != nil {
				req.statusCallback.Error(a.ID, req.filepath, OpUpload, errors.New("push_error", err.Error()))
				return
			}

		}
		err = req.completePush()
		if err != nil && req.statusCallback != nil {
			req.statusCallback.Error(a.ID, req.remotefilepath, OpUpload, err)
			return
		}
	}()
	wg.Wait()
	Logger.Info("Completed the upload. Submitting for commit")

	for _, ch := range req.uploadDataCh {
		close(ch)
	}

	for _, ch := range req.uploadThumbCh {
		close(ch)
	}
	Logger.Info("Closed all the channels. Submitting for commit")
	req.consensus = 0
	wg = &sync.WaitGroup{}
	ones := req.uploadMask.CountOnes()
	wg.Add(ones)
	commitReqs := make([]*CommitRequest, ones)
	var c, pos uint64 = 0, 0
	for i := req.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		//go req.prepareUpload(a, a.Blobbers[pos], req.file[c], req.uploadDataCh[c], req.wg)
		commitReq := &CommitRequest{}
		commitReq.allocationID = a.ID
		commitReq.allocationTx = a.Tx
		commitReq.blobber = a.Blobbers[pos]
		if req.isUpdate {
			newChange := &allocationchange.UpdateFileChange{}
			newChange.NewFile = req.file[c]
			newChange.NumBlocks = req.file[c].NumBlocks
			newChange.Operation = constants.FileOperationUpdate
			newChange.Size = req.file[c].Size
			newChange.NewFile.Attributes = req.file[c].Attributes
			commitReq.changes = append(commitReq.changes, newChange)
		} else {
			newChange := &allocationchange.NewFileChange{}
			newChange.File = req.file[c]
			newChange.NumBlocks = req.file[c].NumBlocks
			newChange.Operation = constants.FileOperationInsert
			newChange.Size = req.file[c].Size
			newChange.File.Attributes = req.file[c].Attributes
			commitReq.changes = append(commitReq.changes, newChange)
		}

		commitReq.connectionID = req.connectionID
		commitReq.wg = wg
		commitReqs[c] = commitReq
		go AddCommitRequest(commitReq)
		c++
	}
	wg.Wait()

	retries := 0
	req.consensus = 0
	for retries < 3 && !req.isConsensusOk() {
		req.consensus = 0
		failedCommits := make([]*CommitRequest, 0)
		for _, commitReq := range commitReqs {
			if commitReq.result != nil {
				if commitReq.result.Success {
					Logger.Info("Commit success", commitReq.blobber.Baseurl, "Retries ", retries)
					req.consensus++
				} else {
					failedCommits = append(failedCommits, commitReq)
					Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage, "Retries ", retries)
				}
			} else {
				failedCommits = append(failedCommits, commitReq)
				Logger.Info("Commit result not set", commitReq.blobber.Baseurl, "Retries ", retries)
			}
		}
		if !req.isConsensusOk() {
			wg := &sync.WaitGroup{}
			wg.Add(len(failedCommits))
			for _, failedCommit := range failedCommits {
				failedCommit.wg = wg
				go AddCommitRequest(failedCommit)
			}
			wg.Wait()
		}
		retries++
	}
	// for _, commitReq := range commitReqs {
	// 	if commitReq.result != nil {
	// 		if commitReq.result.Success {
	// 			Logger.Info("Commit success", commitReq.blobber.Baseurl, "Retries ", retries)
	// 			req.consensus++
	// 		} else {
	// 			Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage, "Retries ", retries)
	// 		}
	// 	} else {
	// 		Logger.Info("Commit result not set", commitReq.blobber.Baseurl, "Retries ", retries)
	// 	}
	// }

	if !req.isConsensusOk() {
		if req.consensus != 0 {
			Logger.Info("Commit consensus failed, Deleting remote file....")
			a.deleteFile(req.remotefilepath, req.consensus, req.consensus)
		}
		if req.statusCallback != nil {
			req.statusCallback.Error(a.ID, req.remotefilepath, OpUpload, errors.New("commit_consensus_failed", "Upload failed as there was no commit consensus"))
			return
		}
	}

	if req.statusCallback != nil {
		sizeInCallback := int64(float32(perShard) * req.consensus)
		OpID := OpUpload
		if req.isUpdate {
			OpID = OpUpdate
		}
		req.statusCallback.Completed(a.ID, req.remotefilepath, req.filemeta.Name, req.filemeta.MimeType, int(sizeInCallback), OpID)
	}

	return
}

func (req *UploadRequest) IsFullConsensusSupported() bool {
	var maxBlobbers = req.GetMaxBlobbersSupported()

	return maxBlobbers >= int(req.fullconsensus)
}

func (req *UploadRequest) GetMaxBlobbersSupported() int {
	return req.uploadMask.CountOnes()
}
