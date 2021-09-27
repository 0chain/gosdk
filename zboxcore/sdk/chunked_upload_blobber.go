package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

// ChunkedUploadBobbler client of blobber's upload
type ChunkedUploadBobbler struct {
	*FLock
	blobber  *blockchain.StorageNode
	fileRef  *fileref.FileRef
	progress *UploadBlobberStatus

	commitChanges []allocationchange.AllocationChange
	commitResult  *CommitResult
}

func (sb *ChunkedUploadBobbler) sendUploadRequest(ctx context.Context, su *ChunkedUpload, chunkIndex int, isFinal bool, encryptedKey string, body *bytes.Buffer, formData ChunkedUploadFormMetadata) error {

	if formData.FileBytesLen == 0 {
		//fixed fileRef in last chunk on stream. io.EOF with nil bytes
		if isFinal {
			sb.fileRef.ChunkSize = su.chunkSize
			sb.fileRef.Size = su.shardUploadedSize
			sb.fileRef.Path = su.fileMeta.RemotePath
			sb.fileRef.ActualFileHash = su.fileMeta.ActualHash
			sb.fileRef.ActualFileSize = su.fileMeta.ActualSize

			sb.fileRef.EncryptedKey = encryptedKey
			sb.fileRef.CalculateHash()
		}

		return nil
	}

	req, err := zboxutil.NewUploadRequestWithMethod(sb.blobber.Baseurl, su.allocationObj.Tx, body, su.httpMethod)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", formData.ContentType)

	resp, err := su.client.Do(req.WithContext(ctx))

	if err != nil {
		logger.Logger.Error("Upload : ", err)
		return err
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Logger.Error("Error: Resp ", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		logger.Logger.Error(sb.blobber.Baseurl, " Upload error response: ", resp.StatusCode, string(respbody))
		return err
	}
	var r UploadResult
	err = json.Unmarshal(respbody, &r)
	if err != nil {
		logger.Logger.Error(sb.blobber.Baseurl, " Upload response parse error: ", err)
		return err
	}
	if r.Filename != su.fileMeta.RemoteName || r.Hash != formData.ChunkHash {
		err = fmt.Errorf("%s Unexpected upload response data %s %s %s", sb.blobber.Baseurl, su.fileMeta.RemoteName, formData.ChunkHash, string(respbody))
		logger.Logger.Error(err)
		return err
	}

	logger.Logger.Info(sb.blobber.Baseurl, su.fileMeta.RemotePath, " uploaded")

	su.consensus.Done()

	//fixed fileRef
	if err == nil {

		//fixed thumbnail info in first chunk if it has thumbnail
		if formData.ThumbnailBytesLen > 0 {

			sb.fileRef.ThumbnailSize = int64(formData.ThumbnailBytesLen)
			sb.fileRef.ThumbnailHash = formData.ThumbnailContentHash

			sb.fileRef.ActualThumbnailSize = su.fileMeta.ActualThumbnailSize
			sb.fileRef.ActualThumbnailHash = su.fileMeta.ActualThumbnailHash
		}

		//fixed fileRef in last chunk on stream
		if isFinal {
			sb.fileRef.MerkleRoot = formData.ChallengeHash
			sb.fileRef.ContentHash = formData.ContentHash

			sb.fileRef.ChunkSize = su.chunkSize
			sb.fileRef.Size = su.shardUploadedSize
			sb.fileRef.Path = su.fileMeta.RemotePath
			sb.fileRef.ActualFileHash = su.fileMeta.ActualHash
			sb.fileRef.ActualFileSize = su.fileMeta.ActualSize

			sb.fileRef.EncryptedKey = encryptedKey
			sb.fileRef.CalculateHash()
		}
	}

	return err

}

func (sb *ChunkedUploadBobbler) processCommit(ctx context.Context, su *ChunkedUpload) error {

	err := sb.Lock()
	if err != nil {
		return err
	}

	defer sb.Unlock()

	rootRef, latestWM, size, err := sb.processWriteMarker(ctx, su)

	if err != nil {
		return err
	}

	wm := &marker.WriteMarker{}
	timestamp := int64(common.Now())
	wm.AllocationRoot = encryption.Hash(rootRef.Hash + ":" + strconv.FormatInt(timestamp, 10))
	if latestWM != nil {
		wm.PreviousAllocationRoot = latestWM.AllocationRoot
	} else {
		wm.PreviousAllocationRoot = ""
	}

	wm.AllocationID = su.allocationObj.ID
	wm.Size = size
	wm.BlobberID = sb.blobber.ID

	wm.Timestamp = timestamp
	wm.ClientID = client.GetClientID()
	err = wm.Sign()
	if err != nil {
		logger.Logger.Error("Signing writemarker failed: ", err)
		return err
	}
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	wmData, err := json.Marshal(wm)
	if err != nil {
		logger.Logger.Error("Creating writemarker failed: ", err)
		return err
	}
	formWriter.WriteField("connection_id", su.progress.ConnectionID)
	formWriter.WriteField("write_marker", string(wmData))

	formWriter.Close()

	req, err := zboxutil.NewCommitRequest(sb.blobber.Baseurl, su.allocationObj.Tx, body)
	if err != nil {
		logger.Logger.Error("Error creating commit req: ", err)
		return err
	}
	req.Header.Add("Content-Type", formWriter.FormDataContentType())

	logger.Logger.Info("Committing to blobber." + sb.blobber.Baseurl)

	//for retries := 0; retries < 3; retries++ {

	resp, err := su.client.Do(req.WithContext(ctx))

	if err != nil {
		logger.Logger.Error("Commit: ", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		logger.Logger.Info(sb.blobber.Baseurl, su.progress.ConnectionID, " committed")
	} else {
		logger.Logger.Error("Commit response: ", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Logger.Error("Response read: ", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		logger.Logger.Error(sb.blobber.Baseurl, " Commit response:", string(respBody))
		return thrown.New("commit_error", string(respBody))
	}

	if err == nil {
		su.consensus.Done()
		return nil
	}
	//}

	return nil
}

func (sb *ChunkedUploadBobbler) processWriteMarker(ctx context.Context, su *ChunkedUpload) (*fileref.Ref, *marker.WriteMarker, int64, error) {
	logger.Logger.Info("received a commit request")
	paths := make([]string, 0)
	for _, change := range sb.commitChanges {
		paths = append(paths, change.GetAffectedPath())
	}

	var lR ReferencePathResult
	req, err := zboxutil.NewReferencePathRequest(sb.blobber.Baseurl, su.allocationObj.Tx, paths)
	if err != nil || len(paths) == 0 {
		logger.Logger.Error("Creating ref path req", err)
		return nil, nil, 0, err
	}

	resp, err := su.client.Do(req)

	if err != nil {
		logger.Logger.Error("Ref path error:", err)
		return nil, nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Logger.Error("Ref path response : ", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Logger.Error("Ref path: Resp", err)
		return nil, nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, 0, fmt.Errorf("Reference path error response: Status: %d - %s ", resp.StatusCode, string(body))
	}

	err = json.Unmarshal(body, &lR)
	if err != nil {
		logger.Logger.Error("Reference path json decode error: ", err)
		return nil, nil, 0, err
	}

	//process the commit request for the blobber here
	if err != nil {

		return nil, nil, 0, err
	}
	rootRef, err := lR.GetDirTree(su.allocationObj.ID)
	if lR.LatestWM != nil {

		rootRef.CalculateHash()
		prevAllocationRoot := encryption.Hash(rootRef.Hash + ":" + strconv.FormatInt(lR.LatestWM.Timestamp, 10))
		// TODO: it is a concurrent change conflict on database.  check concurrent write for allocation
		if prevAllocationRoot != lR.LatestWM.AllocationRoot {
			logger.Logger.Info("Allocation root from latest writemarker mismatch. Expected: " + prevAllocationRoot + " got: " + lR.LatestWM.AllocationRoot)
		}
	}
	if err != nil {

		return nil, nil, 0, err
	}
	size := int64(0)
	for _, change := range sb.commitChanges {
		err = change.ProcessChange(rootRef)
		if err != nil {
			break
		}
		size += change.GetSize()
	}
	if err != nil {

		return nil, nil, 0, err
	}

	return rootRef, lR.LatestWM, size, nil
}
