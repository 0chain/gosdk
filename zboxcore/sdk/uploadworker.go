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
	"math/bits"
	"mime/multipart"
	"net/http"
	"os"
	"sync"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/encoder"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"golang.org/x/crypto/sha3"
)

// Expected success rate is calculated (NumDataShards)*100/(NumDataShards+NumParityShards)
// Additional success percentage on top of expected success rate
const additionalSuccessRate = (10)

type FileMeta struct {
	Name          string
	Path          string
	Hash          string
	MimeType      string
	Size          int64
	ThumbnailSize int64
	ThumbnailHash string
}

type uploadFormData struct {
	ConnectionID        string `json:"connection_id"`
	Filename            string `json:"filename"`
	Path                string `json:"filepath"`
	Hash                string `json:"content_hash,omitempty"`
	ThumbnailHash       string `json:"thumbnail_content_hash,omitempty"`
	MerkleRoot          string `json:"merkle_root,omitempty"`
	ActualHash          string `json:"actual_hash"`
	ActualSize          int64  `json:"actual_size"`
	ActualThumbnailSize int64  `json:"actual_thumb_size"`
	ActualThumbnailHash string `json:"actual_thumb_hash"`
	MimeType            string `json:"mimetype"`
	CustomMeta          string `json:"custom_meta,omitempty"`
}

type uploadResult struct {
	Filename   string `json:"filename"`
	ShardSize  int64  `json:"size"`
	Hash       string `json:"content_hash,omitempty"`
	MerkleRoot string `json:"merkle_root,omitempty"`
}

type UploadRequest struct {
	filepath        string
	thumbnailpath   string
	remotefilepath  string
	statusCallback  StatusCallback
	fileHash        hash.Hash
	fileHashWr      io.Writer
	thumbnailHash   hash.Hash
	thumbnailHashWr io.Writer
	file            []*fileref.FileRef
	filemeta        *FileMeta
	remaining       int64
	thumbRemaining  int64
	wg              *sync.WaitGroup
	uploadDataCh    []chan []byte
	uploadThumbCh   []chan []byte
	isRepair        bool
	isUpdate        bool
	connectionID    string
	datashards      int
	parityshards    int
	uploadMask      uint32
	Consensus
}

func (req *UploadRequest) prepareUpload(a *Allocation, blobber *blockchain.StorageNode, file *fileref.FileRef, uploadCh chan []byte, uploadThumbCh chan []byte, wg *sync.WaitGroup) {
	bodyReader, bodyWriter := io.Pipe()
	formWriter := multipart.NewWriter(bodyWriter)
	httpreq, _ := zboxutil.NewUploadRequest(blobber.Baseurl, a.ID, bodyReader, req.isUpdate)
	//timeout := time.Duration(int64(math.Max(10, float64(obj.file.Size)/(CHUNK_SIZE*float64(len(obj.blobbers)/2)))))
	//ctx, cncl := context.WithTimeout(context.Background(), (time.Second * timeout))

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	var formData uploadFormData
	shardSize := (req.filemeta.Size + int64(a.DataShards) - 1) / int64(a.DataShards)
	thumbnailSize := int64(0)
	remaining := shardSize
	go func() {
		fileMerkleRoot := ""
		fileContentHash := ""
		thumbContentHash := ""
		internalWG := &sync.WaitGroup{}
		internalWG.Add(2)
		go func() {
			defer internalWG.Done()
			fileField, err := formWriter.CreateFormFile("uploadFile", file.Name)
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
			merkleHash := sha3.New256()
			hWr := io.MultiWriter(h, merkleHash)
			merkleLeaves := make([]util.Hashable, 0)
			// Read the data
			for remaining > 0 {
				dataBytes, ok := <-uploadCh
				if !ok {
					return
				}
				fileField.Write(dataBytes)
				hWr.Write(dataBytes)
				merkleLeaves = append(merkleLeaves, util.NewStringHashable(hex.EncodeToString(merkleHash.Sum(nil))))
				merkleHash.Reset()
				remaining = remaining - int64(len(dataBytes))
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
		}()

		go func() {
			defer internalWG.Done()
			if len(req.thumbnailpath) == 0 {
				return
			}

			thumbnailSize = (req.filemeta.ThumbnailSize + int64(a.DataShards) - 1) / int64(a.DataShards)
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
		}()
		internalWG.Wait()

		formData = uploadFormData{
			ConnectionID:        req.connectionID,
			Filename:            file.Name,
			Path:                file.Path,
			ActualHash:          req.filemeta.Hash,
			ActualSize:          req.filemeta.Size,
			ActualThumbnailHash: req.filemeta.ThumbnailHash,
			ActualThumbnailSize: req.filemeta.ThumbnailSize,
			MimeType:            req.filemeta.MimeType,
			Hash:                fileContentHash,
			ThumbnailHash:       thumbContentHash,
			MerkleRoot:          fileMerkleRoot,
		}
		_ = formWriter.WriteField("connection_id", req.connectionID)
		// Logger.Debug("FileFormData:", formData)
		var metaData []byte
		metaData, err := json.Marshal(formData)
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
			return err
		}
		defer resp.Body.Close()

		respbody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Error("Error: Resp ", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			Logger.Error(blobber.Baseurl, " Upload error response: ", resp.StatusCode, string(respbody))
			return err
		}
		var r uploadResult
		err = json.Unmarshal(respbody, &r)
		if err != nil {
			Logger.Error(blobber.Baseurl, " Upload response parse error: ", err)
			return err
		}
		if r.Filename != formData.Filename || r.ShardSize != shardSize ||
			r.Hash != formData.Hash || r.MerkleRoot != formData.MerkleRoot {
			err = fmt.Errorf(blobber.Baseurl, "Unexpected upload response data", string(respbody))
			Logger.Error(err)
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
		file.CalculateHash()
		return nil
	})
	wg.Done()
}

func (req *UploadRequest) setupUpload(a *Allocation) {
	numUploads := bits.OnesCount32(req.uploadMask)
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
	}

	if !req.isRepair {
		req.fileHash = sha1.New()
		req.fileHashWr = io.MultiWriter(req.fileHash)
		req.thumbnailHash = sha1.New()
		req.thumbnailHashWr = io.MultiWriter(req.thumbnailHash)
	}
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numUploads)
	req.consensus = 0

	// Start upload for each blobber
	c, pos := 0, 0
	for i := req.uploadMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		go req.prepareUpload(a, a.Blobbers[pos], req.file[c], req.uploadDataCh[c], req.uploadThumbCh[c], req.wg)
		c++
	}
}

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
	c, pos := 0, 0
	for i := req.uploadMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		req.uploadDataCh[c] <- shards[pos]
		c++
	}
	return nil
}

func (req *UploadRequest) completePush() error {
	if !req.isRepair {
		req.filemeta.Hash = hex.EncodeToString(req.fileHash.Sum(nil))
		c, pos := 0, 0
		for i := req.uploadMask; i != 0; i &= ^(1 << uint32(pos)) {
			pos = bits.TrailingZeros32(i)
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
	var inFile *os.File
	inFile, err := os.Open(req.filepath)
	if err != nil && req.statusCallback != nil {
		req.statusCallback.Error(a.ID, req.filepath, OpUpload, fmt.Errorf("Open file failed: %s", err.Error()))
		return
	}
	defer inFile.Close()
	mimetype, err := zboxutil.GetFileContentType(inFile)
	if err != nil && req.statusCallback != nil {
		req.statusCallback.Error(a.ID, req.filepath, OpUpload, fmt.Errorf("Error detecting the mimetype: %s", err.Error()))
		return
	}
	req.filemeta.MimeType = mimetype
	req.setupUpload(a)
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
		chunksPerShard := (perShard + int64(fileref.CHUNK_SIZE) - 1) / fileref.CHUNK_SIZE
		Logger.Debug("Size:", size, " perShard:", perShard, " chunks/shard:", chunksPerShard)
		if req.statusCallback != nil {
			req.statusCallback.Started(a.ID, req.remotefilepath, OpUpload, int(perShard)*(a.DataShards+a.ParityShards))
		}

		sent := int(0)
		for ctr := int64(0); ctr < chunksPerShard; ctr++ {
			remaining := int64(math.Min(float64(perShard-(ctr*fileref.CHUNK_SIZE)), fileref.CHUNK_SIZE))
			b1 := make([]byte, remaining*int64(a.DataShards))
			_, err = dataReader.Read(b1)
			if err != nil && req.statusCallback != nil {
				req.statusCallback.Error(a.ID, req.filepath, OpUpload, fmt.Errorf("Read failed: %s", err.Error()))
				return
			}
			err = req.pushData(b1)
			if err != nil {
				req.statusCallback.Error(a.ID, req.filepath, OpUpload, fmt.Errorf("Push error: %s", err.Error()))
				return
			}
			sent = sent + int(remaining*int64(a.DataShards+a.ParityShards))
			if req.statusCallback != nil {
				req.statusCallback.InProgress(a.ID, req.remotefilepath, OpUpload, sent)
			}

		}
		err = req.completePush()
		if err != nil && req.statusCallback != nil {
			req.statusCallback.Error(a.ID, req.remotefilepath, OpUpload, fmt.Errorf("Upload failed: %s", err.Error()))
			return
		}
	}()
	wg.Wait()

	for _, ch := range req.uploadDataCh {
		close(ch)
	}

	for _, ch := range req.uploadThumbCh {
		close(ch)
	}

	req.consensus = 0
	wg = &sync.WaitGroup{}
	wg.Add(bits.OnesCount32(req.uploadMask))
	commitReqs := make([]*CommitRequest, bits.OnesCount32(req.uploadMask))
	c, pos := 0, 0
	for i := req.uploadMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		//go req.prepareUpload(a, a.Blobbers[pos], req.file[c], req.uploadDataCh[c], req.wg)
		commitReq := &CommitRequest{}
		commitReq.allocationID = a.ID
		commitReq.blobber = a.Blobbers[pos]
		if req.isUpdate {
			newChange := &allocationchange.UpdateFileChange{}
			newChange.NewFile = req.file[c]
			newChange.NumBlocks = req.file[c].NumBlocks
			newChange.Operation = allocationchange.UPDATE_OPERATION
			newChange.Size = req.file[c].Size
			commitReq.changes = append(commitReq.changes, newChange)
		} else {
			newChange := &allocationchange.NewFileChange{}
			newChange.File = req.file[c]
			newChange.NumBlocks = req.file[c].NumBlocks
			newChange.Operation = allocationchange.INSERT_OPERATION
			newChange.Size = req.file[c].Size
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
	for retries < 1 && !req.isConsensusOk() {
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
		if req.statusCallback != nil {
			req.statusCallback.Error(a.ID, req.remotefilepath, OpUpload, fmt.Errorf("Upload failed: Commit consensus failed"))
			return
		}
	}

	if req.statusCallback != nil {
		sizeInCallback := int64(float32(perShard) * req.consensus)
		req.statusCallback.Completed(a.ID, req.remotefilepath, req.filemeta.Name, req.filemeta.MimeType, int(sizeInCallback), OpUpload)
	}
	return
}
