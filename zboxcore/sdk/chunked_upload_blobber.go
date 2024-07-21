package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/0chain/errors"
	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/hitenjain14/fasthttp"
	"golang.org/x/sync/errgroup"
)

// ChunkedUploadBlobber client of blobber's upload
type ChunkedUploadBlobber struct {
	writeMarkerMutex *WriteMarkerMutex
	blobber          *blockchain.StorageNode
	fileRef          *fileref.FileRef
	progress         *UploadBlobberStatus

	commitChanges []allocationchange.AllocationChange
	commitResult  *CommitResult
}

func (sb *ChunkedUploadBlobber) sendUploadRequest(
	ctx context.Context, su *ChunkedUpload,
	isFinal bool,
	encryptedKey string, dataBuffers []*bytes.Buffer,
	formData ChunkedUploadFormMetadata, contentSlice []string,
	pos uint64, consensus *Consensus) (err error) {

	defer func() {

		if err != nil {
			su.maskMu.Lock()
			su.uploadMask = su.uploadMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
			su.maskMu.Unlock()
		}
	}()

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
			consensus.Done()
		}
	}

	eg, _ := errgroup.WithContext(ctx)

	for dataInd := 0; dataInd < len(dataBuffers); dataInd++ {
		ind := dataInd
		eg.Go(func() error {
			var (
				shouldContinue bool
			)
			var req *fasthttp.Request
			for i := 0; i < 3; i++ {
				req, err = zboxutil.NewFastUploadRequest(
					sb.blobber.Baseurl, su.allocationObj.ID, su.allocationObj.Tx, dataBuffers[ind].Bytes(), su.httpMethod)
				if err != nil {
					return err
				}

				req.Header.Add("Content-Type", contentSlice[ind])
				err, shouldContinue = func() (err error, shouldContinue bool) {
					resp := fasthttp.AcquireResponse()
					defer fasthttp.ReleaseResponse(resp)
					err = zboxutil.FastHttpClient.DoTimeout(req, resp, su.uploadTimeOut)
					fasthttp.ReleaseRequest(req)
					if err != nil {
						logger.Logger.Error("Upload : ", err)
						if errors.Is(err, fasthttp.ErrConnectionClosed) || errors.Is(err, syscall.EPIPE) {
							return err, true
						}
						return fmt.Errorf("Error while doing reqeust. Error %s", err), false
					}

					if resp.StatusCode() == http.StatusOK {
						return
					}

					respbody := resp.Body()
					if resp.StatusCode() == http.StatusTooManyRequests {
						logger.Logger.Error("Got too many request error")
						var r int
						r, err = zboxutil.GetFastRateLimitValue(resp)
						if err != nil {
							logger.Logger.Error(err)
							return
						}
						time.Sleep(time.Duration(r) * time.Second)
						shouldContinue = true
						return
					}

					msg := string(respbody)
					logger.Logger.Error(sb.blobber.Baseurl,
						" Upload error response: ", resp.StatusCode(),
						"err message: ", msg)
					err = errors.Throw(constants.ErrBadRequest, msg)
					return
				}()

				if shouldContinue {
					continue
				}

				if err != nil {
					return err
				}

				break
			}
			return err
		})
	}
	err = eg.Wait()
	if err != nil {
		return err
	}
	consensus.Done()

	if formData.ThumbnailBytesLen > 0 {

		sb.fileRef.ThumbnailSize = int64(formData.ThumbnailBytesLen)
		sb.fileRef.ThumbnailHash = formData.ThumbnailContentHash

		sb.fileRef.ActualThumbnailSize = su.fileMeta.ActualThumbnailSize
		sb.fileRef.ActualThumbnailHash = su.fileMeta.ActualThumbnailHash
	}

	// fixed fileRef in last chunk on stream
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

func (sb *ChunkedUploadBlobber) processCommit(ctx context.Context, su *ChunkedUpload, pos uint64, timestamp int64) (err error) {
	defer func() {
		if err != nil {

			su.maskMu.Lock()
			su.uploadMask = su.uploadMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
			su.maskMu.Unlock()
		}
	}()

	rootRef, latestWM, size, fileIDMeta, err := sb.processWriteMarker(ctx, su)

	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	wm := &marker.WriteMarker{}
	wm.AllocationRoot = rootRef.Hash
	if latestWM != nil {
		wm.PreviousAllocationRoot = latestWM.AllocationRoot
	} else {
		wm.PreviousAllocationRoot = ""
	}

	wm.FileMetaRoot = rootRef.FileMetaHash
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

	fileIDMetaData, err := json.Marshal(fileIDMeta)
	if err != nil {
		logger.Logger.Error("Error marshalling file ID Meta: ", err)
		return err
	}

	err = formWriter.WriteField("file_id_meta", string(fileIDMetaData))
	if err != nil {
		return err
	}

	err = formWriter.WriteField("connection_id", su.progress.ConnectionID)
	if err != nil {
		return err
	}

	err = formWriter.WriteField("write_marker", string(wmData))
	if err != nil {
		return err
	}

	formWriter.Close()

	req, err := zboxutil.NewCommitRequest(sb.blobber.Baseurl, su.allocationObj.ID, su.allocationObj.Tx, body)
	if err != nil {
		logger.Logger.Error("Error creating commit req: ", err)
		return err
	}
	req.Header.Add("Content-Type", formWriter.FormDataContentType())

	logger.Logger.Info("Committing to blobber. " + sb.blobber.Baseurl)

	var (
		resp           *http.Response
		shouldContinue bool
	)

	for retries := 0; retries < 3; retries++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			reqCtx, ctxCncl := context.WithTimeout(ctx, su.commitTimeOut)
			resp, err = su.client.Do(req.WithContext(reqCtx))
			defer ctxCncl()

			if err != nil {
				logger.Logger.Error("Commit: ", err)
				return
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}

			var respBody []byte
			if resp.StatusCode == http.StatusOK {
				logger.Logger.Info(sb.blobber.Baseurl, su.progress.ConnectionID, " committed")
				su.consensus.Done()
				return
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				logger.Logger.Info(sb.blobber.Baseurl, su.progress.ConnectionID,
					" got too many request error. Retrying")

				var r int
				r, err = zboxutil.GetRateLimitValue(resp)
				if err != nil {
					logger.Logger.Error(err)
					return
				}

				time.Sleep(time.Duration(r) * time.Second)
				shouldContinue = true
				return
			}

			respBody, err = io.ReadAll(resp.Body)
			if err != nil {
				logger.Logger.Error("Response read: ", err)
				return
			}

			if strings.Contains(string(respBody), "pending_markers:") {
				logger.Logger.Info("Commit pending for blobber ",
					sb.blobber.Baseurl, "with connection id: ", su.progress.ConnectionID, " Retrying again")
				time.Sleep(5 * time.Second)
				shouldContinue = true
				return
			}

			err = thrown.New("commit_error",
				fmt.Sprintf("Got error response %s with status %d", respBody, resp.StatusCode))
			return
		}()
		if shouldContinue {
			continue
		}
		return
	}
	return thrown.New("commit_error", fmt.Sprintf("Commit failed with response status %d", resp.StatusCode))
}

func (sb *ChunkedUploadBlobber) processWriteMarker(
	ctx context.Context, su *ChunkedUpload) (
	*fileref.Ref, *marker.WriteMarker, int64, map[string]string, error) {

	logger.Logger.Info("received a commit request")
	paths := make([]string, 0)
	for _, change := range sb.commitChanges {
		paths = append(paths, change.GetAffectedPath()...)
	}

	var lR ReferencePathResult
	req, err := zboxutil.NewReferencePathRequest(sb.blobber.Baseurl, su.allocationObj.ID, su.allocationObj.Tx, paths)
	if err != nil || len(paths) == 0 {
		logger.Logger.Error("Creating ref path req", err)
		return nil, nil, 0, nil, err
	}

	resp, err := su.client.Do(req)

	if err != nil {
		logger.Logger.Error("Ref path error:", err)
		return nil, nil, 0, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Logger.Error("Ref path response : ", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Logger.Error("Ref path: Resp", err)
		return nil, nil, 0, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, 0, nil, fmt.Errorf("Reference path error response: Status: %d - %s ", resp.StatusCode, string(body))
	}

	err = json.Unmarshal(body, &lR)
	if err != nil {
		logger.Logger.Error("Reference path json decode error: ", err)
		return nil, nil, 0, nil, err
	}

	rootRef, err := lR.GetDirTree(su.allocationObj.ID)
	if err != nil {
		return nil, nil, 0, nil, err
	}

	if lR.LatestWM != nil {
		rootRef.CalculateHash()
		prevAllocationRoot := rootRef.Hash
		if prevAllocationRoot != lR.LatestWM.AllocationRoot {
			logger.Logger.Info("Allocation root from latest writemarker mismatch. Expected: " +
				prevAllocationRoot + " got: " + lR.LatestWM.AllocationRoot)
			return nil, nil, 0, nil, fmt.Errorf(
				"calculated allocation root mismatch from blobber %s. Expected: %s, Got: %s",
				sb.blobber.Baseurl, prevAllocationRoot, lR.LatestWM.AllocationRoot)
		}
	}

	var size int64
	fileIDMeta := make(map[string]string)
	for _, change := range sb.commitChanges {
		err = change.ProcessChange(rootRef, fileIDMeta)
		if err != nil {
			logger.Logger.Error(err)
			return nil, nil, 0, nil, err
		}
		size += change.GetSize()
	}
	rootRef.CalculateHash()
	return rootRef, lR.LatestWM, size, fileIDMeta, nil
}
