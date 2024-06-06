//go:build js && wasm
// +build js,wasm

package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"syscall/js"
	"time"

	"github.com/0chain/errors"
	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/hack-pad/safejs"
	"github.com/hitenjain14/fasthttp"
	"golang.org/x/sync/errgroup"
)

var (
	hasherMap = make(map[string]Hasher)
)

type ChunkedUploadFormInfo struct {
	ConnectionID      string
	ChunkSize         int64
	ChunkStartIndex   int
	ChunkEndIndex     int
	IsFinal           bool
	EncryptedKey      string
	EncryptedKeyPoint string
	ShardSize         int64
	HttpMethod        string
	AllocationID      string
	AllocationTx      string
}

// createUploadProgress create a new UploadProgress
func (su *ChunkedUpload) createUploadProgress(connectionId string) {
	if su.progress.ChunkSize <= 0 {
		su.progress = UploadProgress{
			ConnectionID:      connectionId,
			ChunkIndex:        -1,
			ChunkSize:         su.chunkSize,
			EncryptOnUpload:   su.encryptOnUpload,
			EncryptedKeyPoint: su.encryptedKeyPoint,
			ActualSize:        su.fileMeta.ActualSize,
			ChunkNumber:       su.chunkNumber,
		}
	}
	su.progress.Blobbers = make([]*UploadBlobberStatus, su.allocationObj.DataShards+su.allocationObj.ParityShards)

	for i := 0; i < len(su.progress.Blobbers); i++ {
		su.progress.Blobbers[i] = &UploadBlobberStatus{}
	}

	su.progress.ID = su.progressID()
	su.saveProgress()
}

// processUpload process upload fragment to its blobber
func (su *ChunkedUpload) processUpload(chunkStartIndex, chunkEndIndex int,
	fileShards []blobberShards, thumbnailShards blobberShards,
	isFinal bool, uploadLength int64) error {
	if len(fileShards) == 0 {
		return thrown.New("upload_failed", "Upload failed. No data to upload")
	}

	fileMetaJSON, err := json.Marshal(su.fileMeta)
	if err != nil {
		return err
	}

	var (
		pos          uint64
		successCount int
	)
	workers := jsbridge.GetWorkers()
	if len(workers) == 0 || len(workers) < len(su.blobbers) {
		return thrown.New("upload_failed", "Upload failed. No workers to upload")
	}

	formInfo := ChunkedUploadFormInfo{
		ConnectionID:      su.progress.ConnectionID,
		ChunkSize:         su.chunkSize,
		ChunkStartIndex:   chunkStartIndex,
		ChunkEndIndex:     chunkEndIndex,
		IsFinal:           isFinal,
		EncryptedKey:      su.encryptedKey,
		EncryptedKeyPoint: su.progress.EncryptedKeyPoint,
		ShardSize:         su.shardSize,
		HttpMethod:        su.httpMethod,
		AllocationID:      su.allocationObj.ID,
		AllocationTx:      su.allocationObj.Tx,
	}
	formInfoJSON, err := json.Marshal(formInfo)
	if err != nil {
		return err
	}

	//convert json objects to uint8 arrays
	fileMetaUint8 := js.Global().Get("Uint8Array").New(len(fileMetaJSON))
	js.CopyBytesToJS(fileMetaUint8, fileMetaJSON)
	formInfoUint8 := js.Global().Get("Uint8Array").New(len(formInfoJSON))
	js.CopyBytesToJS(formInfoUint8, formInfoJSON)

	if chunkStartIndex > 0 {
		err = su.listen(false)
		if err != nil {
			return err
		}
		// index := chunkStartIndex - 1
		// go su.updateProgress(index)
		uploadLength := su.allocationObj.GetChunkReadSize(su.encryptOnUpload) * int64(su.chunkNumber)
		su.progress.UploadLength += uploadLength
		if su.progress.UploadLength > su.fileMeta.ActualSize {
			su.progress.UploadLength = su.fileMeta.ActualSize
		}
		if su.statusCallback != nil {
			su.statusCallback.InProgress(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, int(su.progress.UploadLength), nil)
		}
	}

	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		blobber := su.blobbers[pos]
		blobber.progress.UploadLength += uploadLength
		var thumbnailChunkData []byte

		if len(thumbnailShards) > 0 {
			thumbnailChunkData = thumbnailShards[pos]
		}
		obj := js.Global().Get("Object").New()
		obj.Set("fileMeta", fileMetaUint8)
		obj.Set("formInfo", formInfoUint8)

		if len(thumbnailChunkData) > 0 {
			thumbnailChunkDataUint8 := js.Global().Get("Uint8Array").New(len(thumbnailChunkData))
			js.CopyBytesToJS(thumbnailChunkDataUint8, thumbnailChunkData)
			obj.Set("thumbnailChunkData", thumbnailChunkDataUint8)
			blobber.fileRef.ThumbnailSize = int64(len(thumbnailChunkData))
			blobber.fileRef.ActualThumbnailSize = su.fileMeta.ActualThumbnailSize
			blobber.fileRef.ActualThumbnailHash = su.fileMeta.ActualThumbnailHash
		}

		dataLen := int64(len(fileShards[pos])-1)*int64(len(fileShards[pos][0])) + int64(len(fileShards[pos][len(fileShards[pos])-1]))

		fileshardUint8 := js.Global().Get("Uint8Array").New(dataLen)
		offset := 0
		for _, shard := range fileShards[pos] {
			js.CopyBytesToJS(fileshardUint8.Call("subarray", offset, offset+len(shard)), shard)
			offset += len(shard)
		}
		obj.Set("fileShard", fileshardUint8)
		err = workers[pos].PostMessage(safejs.Safe(obj), []safejs.Value{safejs.Safe(fileshardUint8.Get("buffer"))})
		if err == nil {
			successCount++
		}
	}

	if successCount < su.consensus.consensusThresh {
		su.removeProgress()
		return thrown.New("upload_failed", "Upload failed. Error posting message to worker")
	}

	if isFinal {
		err = su.listen(true)
		if err != nil {
			return err
		}
		// index := chunkEndIndex
		// go su.updateProgress(index)
		su.progress.UploadLength = su.fileMeta.ActualSize
		if su.statusCallback != nil {
			su.statusCallback.InProgress(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, int(su.progress.UploadLength), nil)
		}
	}

	return nil
}

type FinalWorkerResult struct {
	FixedMerkleRoot      string
	ValidationRoot       string
	ThumbnailContentHash string
}

func (su *ChunkedUpload) listen(isFinal bool) error {
	su.consensus.Reset()
	workers := jsbridge.GetWorkers()
	ctx, cancel := context.WithTimeout(su.ctx, su.uploadTimeOut)
	defer cancel()

	var (
		pos      uint64
		errCount int32
		wg       sync.WaitGroup
		wgErrors = make(chan error, len(su.blobbers))
	)

	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		wg.Add(1)
		var err error
		go func(pos uint64) {
			defer func() {
				if err != nil {
					su.maskMu.Lock()
					su.uploadMask = su.uploadMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
					su.maskMu.Unlock()
				}
				wg.Done()
			}()
			worker := workers[pos]
			blobber := su.blobbers[pos]
			eventChan, err := worker.Listen(ctx)
			if err != nil {
				errC := atomic.AddInt32(&errCount, 1)
				if errC >= int32(su.consensus.consensusThresh) {
					cancel()
					wgErrors <- err
				}
				return
			}
			event, ok := <-eventChan
			if !ok {
				logger.Logger.Error("chan closed from: ", worker.Name)
				errC := atomic.AddInt32(&errCount, 1)
				if errC >= int32(su.consensus.consensusThresh) {
					wgErrors <- thrown.New("upload_failed", "Upload failed. Worker event channel closed")
				}
				return
			}
			data, err := event.Data()
			if err != nil {
				errC := atomic.AddInt32(&errCount, 1)
				if errC >= int32(su.consensus.consensusThresh) {
					wgErrors <- thrown.New("upload_failed", "Upload failed. Error getting worker data")
				}
				return
			}
			success, err := data.Get("success")
			if err != nil {
				errC := atomic.AddInt32(&errCount, 1)
				if errC >= int32(su.consensus.consensusThresh) {
					wgErrors <- thrown.New("upload_failed", "Upload failed. Error getting worker data")
				}
				return
			}
			res, _ := success.Bool()
			if !res {
				//get error message
				errMsg, err := data.Get("error")
				if err != nil {
					errC := atomic.AddInt32(&errCount, 1)
					if errC >= int32(su.consensus.consensusThresh) {
						wgErrors <- thrown.New("upload_failed", "Upload failed. Error getting worker data")
					}
					return
				}
				errMsgStr, _ := errMsg.String()
				logger.Logger.Error("error from worker: ", errMsgStr)
				errC := atomic.AddInt32(&errCount, 1)
				if errC >= int32(su.consensus.consensusThresh) {
					wgErrors <- thrown.New("upload_failed", fmt.Sprintf("Upload failed. %s", errMsgStr))
				}
			}
			if isFinal {
				//get final result
				finalResult, err := data.Get("finalResult")
				if err != nil {
					logger.Logger.Error("errorGettingFinalResult")
					errC := atomic.AddInt32(&errCount, 1)
					if errC >= int32(su.consensus.consensusThresh) {
						wgErrors <- thrown.New("upload_failed", "Upload failed. Error getting worker data")
					}
					return
				}
				len, err := finalResult.Length()
				if err != nil {
					logger.Logger.Error("errorGettingFinalResultLength")
					errC := atomic.AddInt32(&errCount, 1)
					if errC >= int32(su.consensus.consensusThresh) {
						wgErrors <- thrown.New("upload_failed", "Upload failed. Error getting worker data")
					}
					return
				}
				resBuf := make([]byte, len)
				safejs.CopyBytesToGo(resBuf, finalResult)
				var finalResultObj FinalWorkerResult
				err = json.Unmarshal(resBuf, &finalResultObj)
				if err != nil {
					logger.Logger.Error("errorGettingFinalResultUnmarshal")
					errC := atomic.AddInt32(&errCount, 1)
					if errC >= int32(su.consensus.consensusThresh) {
						wgErrors <- thrown.New("upload_failed", "Upload failed. Error getting worker data")
					}
					return
				}
				blobber.fileRef.FixedMerkleRoot = finalResultObj.FixedMerkleRoot
				blobber.fileRef.ValidationRoot = finalResultObj.ValidationRoot
				blobber.fileRef.ThumbnailHash = finalResultObj.ThumbnailContentHash
				blobber.fileRef.ChunkSize = su.chunkSize
				blobber.fileRef.Size = su.shardUploadedSize
				blobber.fileRef.Path = su.fileMeta.RemotePath
				blobber.fileRef.ActualFileHash = su.fileMeta.ActualHash
				blobber.fileRef.ActualFileSize = su.fileMeta.ActualSize
				blobber.fileRef.EncryptedKey = su.encryptedKey
				blobber.fileRef.CalculateHash()
			}
			su.consensus.Done()

		}(pos)

	}
	wg.Wait()
	close(wgErrors)
	for err := range wgErrors {
		su.ctxCncl(thrown.New("upload_failed", fmt.Sprintf("Upload failed. %s", err)))
		return err
	}

	if !su.consensus.isConsensusOk() {
		err := thrown.New("consensus_not_met", fmt.Sprintf("Upload failed File not found for path %s. Required consensus atleast %d, got %d",
			su.fileMeta.RemotePath, su.consensus.consensusThresh, su.consensus.getConsensus()))
		su.ctxCncl(err)
		return err
	}
	return nil
}

func ProcessEventData(data safejs.Value) {
	fileMeta, formInfo, fileShards, thumbnailChunkData, err := parseEventData(data)
	if err != nil {
		selfPostMessage(false, err.Error(), nil)
		return
	}
	hasher, ok := hasherMap[fileMeta.RemotePath]
	if !ok {
		hasher = CreateHasher(formInfo.ShardSize)
		hasherMap[fileMeta.RemotePath] = hasher
	}
	if formInfo.IsFinal {
		defer delete(hasherMap, fileMeta.RemotePath)
	}
	formBuilder := CreateChunkedUploadFormBuilder()
	blobberData, err := formBuilder.Build(fileMeta, hasher, formInfo.ConnectionID, formInfo.ChunkSize, formInfo.ChunkStartIndex, formInfo.ChunkEndIndex, formInfo.IsFinal, formInfo.EncryptedKey, formInfo.EncryptedKeyPoint,
		fileShards, thumbnailChunkData, formInfo.ShardSize)
	if err != nil {
		selfPostMessage(false, err.Error(), nil)
		return
	}
	blobberURL := os.Getenv("BLOBBER_URL")
	err = sendUploadRequest(blobberData.dataBuffers, blobberData.contentSlice, blobberURL, formInfo.AllocationID, formInfo.AllocationTx, formInfo.HttpMethod)
	if err != nil {
		selfPostMessage(false, err.Error(), nil)
		return
	}
	if formInfo.IsFinal {
		finalResult := &FinalWorkerResult{
			FixedMerkleRoot:      blobberData.formData.FixedMerkleRoot,
			ValidationRoot:       blobberData.formData.ValidationRoot,
			ThumbnailContentHash: blobberData.formData.ThumbnailContentHash,
		}
		selfPostMessage(true, "", finalResult)
	} else {
		selfPostMessage(true, "", nil)
	}

}

func selfPostMessage(success bool, errMsg string, finalResult *FinalWorkerResult) {
	obj := js.Global().Get("Object").New()
	obj.Set("success", success)
	obj.Set("error", errMsg)
	if finalResult != nil {
		finalResultJSON, err := json.Marshal(finalResult)
		if err != nil {
			obj.Set("finalResult", nil)
		} else {
			finalResultUint8 := js.Global().Get("Uint8Array").New(len(finalResultJSON))
			js.CopyBytesToJS(finalResultUint8, finalResultJSON)
			obj.Set("finalResult", finalResultUint8)
		}
	}
	self := jsbridge.GetSelfWorker()
	self.PostMessage(safejs.Safe(obj), nil) //nolint:errcheck

}

func parseEventData(data safejs.Value) (*FileMeta, *ChunkedUploadFormInfo, [][]byte, []byte, error) {

	fileMetaUint8, err := data.Get("fileMeta")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	formInfoUint8, err := data.Get("formInfo")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	fileShardUint8, err := data.Get("fileShard")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	//get fileMetaUint8 length
	fileMetaLen, err := fileMetaUint8.Length()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	fileMetaBytes := make([]byte, fileMetaLen)
	safejs.CopyBytesToGo(fileMetaBytes, fileMetaUint8)
	fileMeta := &FileMeta{}
	err = json.Unmarshal(fileMetaBytes, fileMeta)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	// get formInfoUint8 length
	formInfoLen, err := formInfoUint8.Length()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	formInfoBytes := make([]byte, formInfoLen)
	safejs.CopyBytesToGo(formInfoBytes, formInfoUint8)
	formInfo := &ChunkedUploadFormInfo{}
	err = json.Unmarshal(formInfoBytes, formInfo)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	chunkSize := formInfo.ChunkSize
	if chunkSize == 0 {
		chunkSize = CHUNK_SIZE
	}
	// get fileShardUint8 length
	fileShardLen, err := fileShardUint8.Length()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	buf := make([]byte, fileShardLen)
	safejs.CopyBytesToGo(buf, fileShardUint8)
	fileShards := splitData(buf, int(chunkSize))

	thumbnailChunkDataUint8, err := data.Get("thumbnailChunkData")
	if err != nil {
		return fileMeta, formInfo, fileShards, nil, nil
	}
	thumbnailChunkDataLen, err := thumbnailChunkDataUint8.Length()
	if err != nil {
		return fileMeta, formInfo, fileShards, nil, nil
	}
	thumbnailChunkData := make([]byte, thumbnailChunkDataLen)
	safejs.CopyBytesToGo(thumbnailChunkData, thumbnailChunkDataUint8)
	return fileMeta, formInfo, fileShards, thumbnailChunkData, nil
}

func sendUploadRequest(dataBuffers []*bytes.Buffer, contentSlice []string, blobberURL, allocationID, allocationTx, httpMethod string) (err error) {
	eg, _ := errgroup.WithContext(context.TODO())
	for dataInd := 0; dataInd < len(dataBuffers); dataInd++ {
		ind := dataInd
		eg.Go(func() error {
			var (
				shouldContinue bool
			)
			var req *fasthttp.Request
			for i := 0; i < 3; i++ {
				req, err = zboxutil.NewFastUploadRequest(
					blobberURL, allocationID, allocationTx, dataBuffers[ind].Bytes(), httpMethod)
				if err != nil {
					return err
				}

				req.Header.Add("Content-Type", contentSlice[ind])
				err, shouldContinue = func() (err error, shouldContinue bool) {
					resp := fasthttp.AcquireResponse()
					defer fasthttp.ReleaseResponse(resp)
					err = zboxutil.FastHttpClient.DoTimeout(req, resp, DefaultUploadTimeOut)
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
					logger.Logger.Error(blobberURL,
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
	return eg.Wait()
}
