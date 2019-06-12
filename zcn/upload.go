package zcn

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/bits"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"0chain.net/clientsdk/encryption"
	"0chain.net/clientsdk/util"
	"golang.org/x/crypto/sha3"
)

type uploadFormData struct {
	ConnectionID string `json:"connection_id"`
	Filename     string `json:"filename"`
	Path         string `json:"filepath"`
	Hash         string `json:"content_hash,omitempty"`
	MerkleRoot   string `json:"merkle_root,omitempty"`
	ActualHash   string `json:"actual_hash"`
	ActualSize   int64  `json:"actual_size"`
	CustomMeta   string `json:"custom_meta,omitempty"`
}

type uploadResult struct {
	Filename   string `json:"filename"`
	ShardSize  int64  `json:"size"`
	Hash       string `json:"content_hash,omitempty"`
	MerkleRoot string `json:"merkle_root,omitempty"`
}

func (obj *Allocation) prepareUpload(blobber *util.Blobber, uploadCh chan []byte, wg *sync.WaitGroup) {
	bodyReader, bodyWriter := io.Pipe()
	formWriter := multipart.NewWriter(bodyWriter)
	req, _ := util.NewUploadRequest(blobber.UrlRoot, obj.allocationId, obj.client, bodyReader, obj.isUploadUpdate)
	//timeout := time.Duration(int64(math.Max(10, float64(obj.file.Size)/(CHUNK_SIZE*float64(len(obj.blobbers)/2)))))
	//ctx, cncl := context.WithTimeout(context.Background(), (time.Second * timeout))
	ctx, cncl := context.WithCancel(context.Background())
	req.Header.Add("Content-Type", formWriter.FormDataContentType())
	var formData uploadFormData
	shardSize := (obj.file.Size + int64(obj.encoder.iDataShards) - 1) / int64(obj.encoder.iDataShards)
	remaining := shardSize
	go func() {
		fileField, err := formWriter.CreateFormFile("uploadFile", obj.file.Name)
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
			dataBytes := <-uploadCh
			fileField.Write(dataBytes)
			hWr.Write(dataBytes)
			merkleLeaves = append(merkleLeaves, util.NewStringHashable(hex.EncodeToString(merkleHash.Sum(nil))))
			merkleHash.Reset()
			remaining = remaining - int64(len(dataBytes))
		}
		var mt util.MerkleTreeI = &util.MerkleTree{}
		mt.ComputeTree(merkleLeaves)
		if !obj.isUploadRepair {
			// Wait for file hash to be ready
			// Logger.Debug("Waiting for file hash....")
			_ = <-uploadCh
			// Logger.Debug("File Hash ready", obj.file.Hash)
		}
		formData = uploadFormData{
			ConnectionID: fmt.Sprintf("%d", blobber.ConnObj.ConnectionId),
			Filename:     obj.file.Name,
			Path:         obj.file.Path,
			ActualHash:   obj.file.ActualHash,
			ActualSize:   obj.file.Size,
			Hash:         hex.EncodeToString(h.Sum(nil)),
			MerkleRoot:   mt.GetRoot(),
		}
		// Logger.Debug("FileFormData:", formData)
		var metaData []byte
		metaData, err = json.Marshal(formData)
		// Logger.Debug("Upload with",string(metaData))
		if err == nil {
			_ = formWriter.WriteField("uploadMeta", string(metaData))
			bodyWriter.CloseWithError(formWriter.Close())
		}
	}()
	_ = httpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Upload : ", err)
			return err
		}
		defer resp.Body.Close()
		// fmt.Println(blobber.UrlRoot, "Resp Status:", resp.StatusCode, (float32(1) / float32(len(obj.blobbers)) * 100))

		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Error("Error: Resp ", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			Logger.Error(blobber.UrlRoot, " Upload error response: ", resp.StatusCode, string(resp_body))
			return err
		} else {
			var r uploadResult
			err := json.Unmarshal(resp_body, &r)
			if err != nil {
				Logger.Error(blobber.UrlRoot, " Upload response parse error: ", err)
				return err
			}
			if r.Filename != formData.Filename || r.ShardSize != shardSize ||
				r.Hash != formData.Hash || r.MerkleRoot != formData.MerkleRoot {
				err = fmt.Errorf(blobber.UrlRoot, "Unexpected upload response data", string(resp_body))
				Logger.Error(err)
				return err
			}
			obj.consensus++
			Logger.Info(blobber.UrlRoot, formData.Path, " uploaded")
			// Update connection object
			fileHashData := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v:%v:%v", obj.allocationId, "f", formData.Filename, formData.Path, shardSize, formData.Hash, formData.MerkleRoot, formData.ActualSize, formData.ActualHash)
			// Insert file into connection object appropriate path
			fileHash := encryption.Hash(fileHashData)
			if obj.isUploadUpdate {
				err = blobber.ConnObj.UpdateFile(obj.file.Path, fileHash, shardSize)
			} else {
				err = blobber.ConnObj.AddFile(obj.file.Path, fileHash, shardSize)
			}
			if err != nil {
				Logger.Error("Error adding/updating path to blobber: ", err)
			}
		}
		return nil
	})
	wg.Done()
}

func (obj *Allocation) setupUpload() {
	// Allocate the data channels.
	numUploads := bits.OnesCount32(obj.uploadMask)
	obj.uploadDataCh = make([]chan []byte, numUploads)
	for i := range obj.uploadDataCh {
		obj.uploadDataCh[i] = make(chan []byte)
	}
	// Setup file hash compute
	if !obj.isUploadRepair {
		obj.file.FileHash = sha1.New()
		obj.file.FileHashWr = io.MultiWriter(obj.file.FileHash)
	}

	obj.wg.Add(numUploads)
	// Clear success
	obj.consensus = 0
	// Start upload for each blobber
	c, pos := 0, 0
	for i := obj.uploadMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		go obj.prepareUpload(&obj.blobbers[pos], obj.uploadDataCh[c], &obj.wg)
		c++
	}
}

func getFullRemotePath(localPath, remotePath string) string {
	if remotePath == "" || strings.HasSuffix(remotePath, "/") {
		remotePath = strings.TrimRight(remotePath, "/")
		_, fileName := filepath.Split(localPath)
		remotePath = fmt.Sprintf("%s/%s", remotePath, fileName)
	}
	return remotePath
}

func (obj *Allocation) UploadOrRepairFile(localPath, remotePath string, statusCb StatusCallback) error {
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("Local file error: %s", err.Error())
	}
	remotePath = getFullRemotePath(localPath, remotePath)
	if !obj.isUploadRepair && !obj.isUploadUpdate {
		fInfo := util.GetFileInfo(&obj.dirTree, remotePath)
		if fInfo != nil {
			return fmt.Errorf("File at path already exists")
		}
	}
	var inFile *os.File
	inFile, err = os.Open(localPath)
	if err != nil {
		return fmt.Errorf("Open file failed: %s", err.Error())
	}
	defer inFile.Close()
	obj.pauseBgSync()
	defer obj.resumeBgSync()
	var fileName string
	_, fileName = filepath.Split(remotePath)
	obj.file.Name = fileName
	obj.file.Size = fileInfo.Size()
	obj.file.Path = remotePath
	obj.file.Type = filepath.Ext(fileName)
	obj.file.Remaining = obj.file.Size
	obj.setupUpload()
	statusCb.Started(obj.allocationId, remotePath, OpUpload, int(obj.file.Size))

	size := obj.file.Size
	// Calculate number of bytes per shard.
	perShard := (size + int64(obj.encoder.iDataShards) - 1) / int64(obj.encoder.iDataShards)
	// Pad data to Shards*perShard.
	padding := make([]byte, (int64(obj.encoder.iDataShards)*perShard)-size)
	dataReader := io.MultiReader(inFile, bytes.NewBuffer(padding))
	chunksPerShard := (perShard + int64(CHUNK_SIZE) - 1) / CHUNK_SIZE
	Logger.Debug("Size:", size, " perShard:", perShard, " chunks/shard:", chunksPerShard)
	sent := int(0)
	for ctr := int64(0); ctr < chunksPerShard; ctr++ {
		remaining := int64(math.Min(float64(perShard-(ctr*CHUNK_SIZE)), CHUNK_SIZE))
		b1 := make([]byte, remaining*int64(obj.encoder.iDataShards))
		_, err = dataReader.Read(b1)
		if err != nil {
			return fmt.Errorf("Read failed: %s", err.Error())
		}
		err = obj.push(b1)
		if err != nil {
			return fmt.Errorf("Push error: %s", err.Error())
		}
		sent = sent + int(remaining*int64(obj.encoder.iDataShards))
		statusCb.InProgress(obj.allocationId, remotePath, OpUpload, sent)
	}
	err = obj.completePush()
	if err != nil {
		return fmt.Errorf("Upload failed: %s", err.Error())
	}
	inFile.Seek(0, 0)
	mimetype, err := util.GetFileContentType(inFile)
	statusCb.Completed(obj.allocationId, remotePath, fileName, mimetype, int(size), OpUpload)
	return nil
}

func (obj *Allocation) UploadFile(localPath, remotePath string, statusCb StatusCallback) error {
	if obj.isRepairCommitPending {
		return fmt.Errorf("Upload not allowed. Repair commit pending")
	}
	// Upload to all blobbers
	obj.uploadMask = ((1 << uint32(len(obj.blobbers))) - 1)
	err := obj.UploadOrRepairFile(localPath, remotePath, statusCb)
	if err == nil {
		// Add the file to directory with size and actual content hash
		fl, err := util.InsertFile(&obj.dirTree, obj.file.Path, obj.file.ActualHash, obj.file.Size)
		if err != nil {
			return fmt.Errorf("Adding filepath to dirtree failed %s", err.Error())
		}
		fl.Meta[fileMetaBlobberCount] = obj.consensus
		obj.isUploadCommitPending = true
	}
	return err
}

func (obj *Allocation) UpdateFile(localPath, remotePath string, statusCb StatusCallback) error {
	if obj.isRepairCommitPending {
		return fmt.Errorf("Update not allowed. Repair commit pending")
	}
	remotePath = getFullRemotePath(localPath, remotePath)
	fl := util.GetFileInfo(&obj.dirTree, remotePath)
	if fl == nil {
		return fmt.Errorf("File not found")
	}
	// Mark as new update
	obj.isUploadUpdate = true
	clearFlag := func() { obj.isUploadUpdate = false }
	defer clearFlag()
	// Update to all blobbers
	obj.uploadMask = ((1 << uint32(len(obj.blobbers))) - 1)
	err := obj.UploadOrRepairFile(localPath, remotePath, statusCb)
	if err == nil {
		fl.Hash = obj.file.ActualHash
		fl.Size = obj.file.Size
		fl.Meta[fileMetaBlobberCount] = obj.consensus
		obj.isUploadCommitPending = true
	}
	return err
}

func (obj *Allocation) RepairFile(localPath, remotePath string, statusCb StatusCallback) error {
	remotePath = getFullRemotePath(localPath, remotePath)
	fl := util.GetFileInfo(&obj.dirTree, remotePath)
	if fl == nil {
		return fmt.Errorf("Remote filepath not found")
	}
	if obj.isUploadCommitPending {
		return fmt.Errorf("Repair not allowed. Upload commit pending.")
	}
	// Upload to only blobbers dont have file
	found := obj.getFileConsensusFromBlobbers(remotePath)
	allMask := uint32((1 << uint32(len(obj.blobbers))) - 1)
	if found == allMask {
		return fmt.Errorf("No repair required")
	}
	obj.uploadMask = (^found & allMask)
	// No need to compute has again
	obj.file.ActualHash = fl.Hash
	// Mark as repair upload
	obj.isUploadRepair = true
	clearFlag := func() { obj.isUploadRepair = false }
	defer clearFlag()
	err := obj.UploadOrRepairFile(localPath, remotePath, statusCb)
	if err == nil {
		fl.Meta[fileMetaBlobberCount] = float32(fl.Meta[fileMetaBlobberCount].(float64)) + obj.consensus
		obj.isRepairCommitPending = true
	}
	return err
}

func (obj *Allocation) push(data []byte) error {
	//TODO: Check for optimization
	n := int64(math.Min(float64(obj.file.Remaining), float64(len(data))))
	if !obj.isUploadRepair {
		obj.file.FileHashWr.Write(data[:n])
	}
	obj.file.Remaining = obj.file.Remaining - n
	shards, err := obj.encoder.encode(data)
	if err != nil {
		fmt.Println("Erasure coding failed", err)
		return err
	}
	c, pos := 0, 0
	for i := obj.uploadMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		obj.uploadDataCh[c] <- shards[pos]
		c++
	}
	return nil
}

func (obj *Allocation) completePush() error {
	if !obj.isUploadRepair {
		obj.file.ActualHash = hex.EncodeToString(obj.file.FileHash.Sum(nil))
		// Start upload for each blobber
		c, pos := 0, 0
		for i := obj.uploadMask; i != 0; i &= ^(1 << uint32(pos)) {
			pos = bits.TrailingZeros32(i)
			obj.uploadDataCh[c] <- []byte("done")
			c++
		}
	}
	obj.wg.Wait()
	if !obj.isConsensusOk() {
		return fmt.Errorf("Upload failed: Consensus_rate:%f, expected:%f", obj.getConsensusRate(), obj.getConsensusRequiredForOk())
	}
	return nil
}

func (obj *Allocation) commitBlobber(blobber *util.Blobber) {
	defer obj.wg.Done()
	wm := util.NewWriteMarker()
	timestamp := util.Now()
	wm.AllocationRoot = blobber.ConnObj.GetAllocationRoot(timestamp)
	wm.PreviousAllocationRoot = blobber.DirTree.Hash
	wm.AllocationID = obj.allocationId
	wm.Size = blobber.ConnObj.GetSize()
	wm.BlobberID = blobber.Id
	wm.Timestamp = timestamp
	wm.ClientID = obj.client.Id
	err := wm.Sign(obj.client.PrivateKey)
	if err != nil {
		Logger.Error("Signing writemarker failed: ", err)
		return
	}
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	wmData, err := json.Marshal(wm)
	if err != nil {
		Logger.Error("Creating writemarker failed: ", err)
		return
	}
	formWriter.WriteField("connection_id", fmt.Sprintf("%d", blobber.ConnObj.ConnectionId))
	formWriter.WriteField("write_marker", string(wmData))
	formWriter.Close()

	req, err := util.NewCommitRequest(blobber.UrlRoot, obj.allocationId, obj.client, body)
	if err != nil {
		Logger.Error("Error creating commit req: ", err)
		return
	}
	req.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 60))
	_ = httpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Commit: ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			obj.consensus++
			Logger.Info(blobber.UrlRoot, blobber.ConnObj.GetCommitData(), " committed")
			// Store the connection object dirtree to blobber
			blobber.DirTree = blobber.ConnObj.DirTree
		} else {
			Logger.Error("Commit response: ", resp.StatusCode)
		}

		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Error("Response read: ", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			Logger.Error(blobber.UrlRoot, " Commit response:", string(resp_body))
		}
		return nil
	})
	// Both success and failure case reset the connection object
	blobber.ConnObj.Reset()
}

func (obj *Allocation) Commit() error {
	obj.pauseBgSync()
	defer obj.resumeBgSync()
	obj.consensus = 0
	obj.wg.Add(len(obj.blobbers))
	for i := 0; i < len(obj.blobbers); i++ {
		go obj.commitBlobber(&obj.blobbers[i])
	}
	obj.wg.Wait()
	if !obj.isConsensusOk() {
		return fmt.Errorf("Commit failed: Consensus_rate:%2f, expected:%2f", obj.getConsensusRate(), obj.consensusThresh)
	}
	obj.isUploadCommitPending, obj.isRepairCommitPending = false, false
	return nil
}
