package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/hitenjain14/fasthttp"
	"github.com/valyala/bytebufferpool"
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
	uploadMetaSlice []string,
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

	// Raw-upload path threads connection_id via URL query + uploadMeta JSON
	// via the X-Upload-Meta header (populated when GOSDK_USE_RAW_UPLOAD=1).
	// On the multipart path uploadMetaSlice entries are empty so the branch
	// below is a no-op.

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

				// Raw-upload path: add ?connection_id query + X-Upload-Meta header.
				// The blobber's GetField/TryParseForm short-circuits when X-Upload-Meta
				// is present (no multipart parse), and the new RawUploadFileCommand
				// reads connection_id from URL query + metadata from the header.
				if ind < len(uploadMetaSlice) && uploadMetaSlice[ind] != "" {
					uri := string(req.URI().FullURI())
					sep := "?"
					if strings.Contains(uri, "?") {
						sep = "&"
					}
					req.SetRequestURI(uri + sep + "connection_id=" + su.progress.ConnectionID)
					req.Header.Set("X-Upload-Meta", uploadMetaSlice[ind])
				}
				err, shouldContinue = func() (err error, shouldContinue bool) {
					resp := fasthttp.AcquireResponse()
					defer fasthttp.ReleaseResponse(resp)
					// INSTRUMENTATION (May 28): per-request HTTP POST timing — names
					// each blobber upload's wall-time + bytes + status. The drain
					// (uploadWG.Wait) is the bottleneck for large; this shows whether
					// the per-request cost is TLS handshake (high) or pooled (low).
					bytesN := len(dataBuffers[ind].Bytes())
					tReq := time.Now()
					err = zboxutil.FastHttpClient.DoTimeout(req, resp, su.uploadTimeOut)
					reqMs := time.Since(tReq).Milliseconds()
					// INSTRUMENTATION: add connection_id to per-request log so we
					// can compute dispatch-rate per file in post-processing —
					// time between successive batch sends for one connection_id
					// tells us if the gateway is the throttle (blobber-side
					// shows lock_ms = 0 across the board).
					_ = su.progress.ConnectionID
					status := 0
					if err == nil {
						status = resp.StatusCode()
					}
					logger.Logger.Info(fmt.Sprintf("[upload-req] conn=%s blobber=%s bytes=%d ms=%d status=%d attempt=%d", su.progress.ConnectionID, sb.blobber.Baseurl, bytesN, reqMs, status, i))
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
				buff := &bytebufferpool.ByteBuffer{
					B: dataBuffers[ind].Bytes(),
				}
				formDataPool.Put(buff)

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

	commitReq := &CommitRequest{
		blobber:      sb.blobber,
		allocationID: su.allocationObj.ID,
		allocationTx: su.allocationObj.Tx,
		connectionID: su.progress.ConnectionID,
		timestamp:    timestamp,
		version:      sb.blobber.AllocationVersion + 1,
	}
	if err = commitReq.commitBlobber(); err != nil {
		return err
	}
	su.consensus.Done()
	return nil
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
	req, err := zboxutil.NewReferencePathRequest(sb.blobber.Baseurl, su.allocationObj.ID, su.allocationObj.Tx, su.allocationObj.sig, paths)
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
