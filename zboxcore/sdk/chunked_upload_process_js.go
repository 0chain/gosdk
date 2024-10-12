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
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/hack-pad/go-webworkers/worker"
	"github.com/hack-pad/safejs"
	"github.com/hitenjain14/fasthttp"
	"github.com/valyala/bytebufferpool"
	"golang.org/x/sync/errgroup"
)

var (
	hasherMap map[string]workerProcess
)

type workerProcess struct {
	wg     *sync.WaitGroup
	hasher Hasher
}

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
	OnlyHash          bool
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

	select {
	case <-su.ctx.Done():
		return context.Cause(su.ctx)
	default:
	}

	fileMetaJSON, err := json.Marshal(su.fileMeta)
	if err != nil {
		return err
	}

	var (
		pos          uint64
		successCount int
	)

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
		OnlyHash:          chunkEndIndex <= su.progress.ChunkIndex,
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
		select {
		case <-su.ctx.Done():
			return context.Cause(su.ctx)
		case su.listenChan <- struct{}{}:
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
		worker := jsbridge.GetWorker(blobber.blobber.ID)
		if worker == nil {
			logger.Logger.Error("worker not found for blobber: ", blobber.blobber.Baseurl)
			continue
		}
		if len(thumbnailShards) > 0 {
			thumbnailChunkData = thumbnailShards[pos]
		}
		obj := js.Global().Get("Object").New()
		jsbridge.SetMsgType(&obj, "upload")
		obj.Set("fileMeta", fileMetaUint8)
		if formInfo.OnlyHash && su.progress.UploadMask.And(zboxutil.NewUint128(1).Lsh(pos)).Equals64(0) {
			//check if pos is set in upload mask in progress
			formInfo.OnlyHash = false
			noHashFormInfoJSON, _ := json.Marshal(formInfo)
			noHashFormInfoUint8 := js.Global().Get("Uint8Array").New(len(noHashFormInfoJSON))
			js.CopyBytesToJS(noHashFormInfoUint8, noHashFormInfoJSON)
			obj.Set("formInfo", noHashFormInfoUint8)
			formInfo.OnlyHash = true //reset to true
		} else {
			obj.Set("formInfo", formInfoUint8)
		}

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
		err = worker.PostMessage(safejs.Safe(obj), []safejs.Value{safejs.Safe(fileshardUint8.Get("buffer"))})
		if err == nil {
			successCount++
		} else {
			logger.Logger.Error("error posting message to worker: ", err)
			su.uploadMask = su.uploadMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
		}
		if isFinal {
			blobber.fileRef.ChunkSize = su.chunkSize
			blobber.fileRef.Size = su.shardUploadedSize
			blobber.fileRef.Path = su.fileMeta.RemotePath
			blobber.fileRef.ActualFileHash = su.fileMeta.ActualHash
			blobber.fileRef.ActualFileSize = su.fileMeta.ActualSize
			blobber.fileRef.EncryptedKey = su.encryptedKey
		}
	}

	if successCount < su.consensus.consensusThresh {
		su.removeProgress()
		return thrown.New("upload_failed", "Upload failed. Error posting message to worker")
	}
	fileShards = nil
	if isFinal {
		su.uploadWG.Wait()
		select {
		case <-su.ctx.Done():
			return context.Cause(su.ctx)
		default:
		}
		su.progress.UploadLength = su.fileMeta.ActualSize
		if su.statusCallback != nil {
			su.statusCallback.InProgress(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, int(su.progress.UploadLength), nil)
		}
	}

	return nil
}

type FinalWorkerResult struct {
	DataHash             string
	ThumbnailContentHash string
}

func (su *ChunkedUpload) listen(allEventChan []eventChanWorker, respChan chan error) {
	su.consensus.Reset()

	var (
		pos      uint64
		errCount int32
		wg       sync.WaitGroup
		wgErrors = make(chan error, len(su.blobbers))
		isFinal  bool
	)

	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		wg.Add(1)
		go func(pos uint64) {
			var uploadSuccess bool
			defer func() {
				if !uploadSuccess {
					su.maskMu.Lock()
					su.uploadMask = su.uploadMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
					su.maskMu.Unlock()
				}
				wg.Done()
			}()
			blobber := su.blobbers[pos]

			eventChan := allEventChan[pos]
			if eventChan.C == nil {
				errC := atomic.AddInt32(&errCount, 1)
				if errC >= int32(su.consensus.consensusThresh) {
					wgErrors <- thrown.New("upload_failed", "Upload failed. Worker event channel not found")
				}
				return
			}
			for {
				event, ok := <-eventChan.C
				if !ok {
					logger.Logger.Error("chan closed from: ", blobber.blobber.Baseurl)
					errC := atomic.AddInt32(&errCount, 1)
					if errC >= int32(su.consensus.consensusThresh) {
						if su.ctx.Err() != nil {
							wgErrors <- context.Cause(su.ctx)
						} else {
							wgErrors <- thrown.New("upload_failed", "Upload failed. Worker event channel closed")
						}
					}
					return
				}
				msgType, data, err := jsbridge.GetMsgType(event)
				if err != nil {
					errC := atomic.AddInt32(&errCount, 1)
					if errC >= int32(su.consensus.consensusThresh) {
						wgErrors <- errors.Wrap(err, "could not get msgType")
					}
					return
				}

				switch msgType {
				case "auth":
					if err := su.processWebWorkerAuthRequest(data, eventChan); err != nil {
						errC := atomic.AddInt32(&errCount, 1)
						if errC >= int32(su.consensus.consensusThresh) {
							wgErrors <- err
						}
						return
					}
				case "upload":
					//get error message
					//get final result
					var err error
					isFinal, err = su.processWebWorkerUpload(data, blobber, pos)
					if err != nil {
						errC := atomic.AddInt32(&errCount, 1)
						if errC >= int32(su.consensus.consensusThresh) {
							wgErrors <- err
						}
					} else {
						uploadSuccess = true
					}
					return
				default:
					logger.Logger.Error("unknown msg type: ", msgType)
				}
			}

		}(pos)
	}

	wg.Wait()
	close(wgErrors)
	for err := range wgErrors {
		logger.Logger.Error("error from worker: ", err)
		su.ctxCncl(thrown.New("upload_failed", fmt.Sprintf("Upload failed. %s", err)))
		respChan <- err
	}

	if !su.consensus.isConsensusOk() {
		logger.Logger.Error("consensus not met")
		err := thrown.New("consensus_not_met", fmt.Sprintf("Upload failed File not found for path %s. Required consensus atleast %d, got %d",
			su.fileMeta.RemotePath, su.consensus.consensusThresh, su.consensus.getConsensus()))
		su.ctxCncl(err)
		respChan <- err
	}
	for chunkEndIndex, mask := range su.processMap {
		if mask.CountOnes() >= su.consensus.consensusThresh {
			su.updateProgress(chunkEndIndex, mask)
			delete(su.processMap, chunkEndIndex)
		}
	}

	if isFinal {
		close(respChan)
	} else {
		respChan <- nil
	}
}

func (su *ChunkedUpload) processWebWorkerUpload(data *safejs.Value, blobber *ChunkedUploadBlobber, pos uint64) (bool, error) {
	var isFinal bool
	success, err := data.Get("success")
	if err != nil {
		return false, errors.Wrap(err, "could not get 'success' field")
	}
	res, _ := success.Bool()
	if !res {
		errMsg, err := data.Get("error")
		if err != nil {
			return false, errors.Wrap(err, "could not get 'error' field")
		}

		errMsgStr, _ := errMsg.String()
		return false, fmt.Errorf("%s", errMsgStr)
	}

	chunkEndIndexObj, _ := data.Get("chunkEndIndex")
	chunkEndIndex, _ := chunkEndIndexObj.Int()
	su.updateChunkProgress(chunkEndIndex, pos)
	finalRequestObject, _ := data.Get("isFinal")
	finalRequest, _ := finalRequestObject.Bool()
	if finalRequest {
		finalResult, err := data.Get("finalResult")
		if err != nil {
			logger.Logger.Error("errorGettingFinalResult")
			return false, errors.Wrap(err, "could not get 'finalResult' field")
		}

		len, err := finalResult.Length()
		if err != nil {
			logger.Logger.Error("errorGettingFinalResultLength")
			return false, errors.Wrap(err, "could not get 'finalResult' Length")
		}

		resBuf := make([]byte, len)
		safejs.CopyBytesToGo(resBuf, finalResult)
		var finalResultObj FinalWorkerResult
		err = json.Unmarshal(resBuf, &finalResultObj)
		if err != nil {
			logger.Logger.Error("errorGettingFinalResultUnmarshal")
			return false, errors.Wrap(err, "could not unmarshal 'finalResult' obj")
		}
		blobber.fileRef.ThumbnailHash = finalResultObj.ThumbnailContentHash
		isFinal = true
	}

	su.consensus.Done()
	return isFinal, nil
}

func (su *ChunkedUpload) processWebWorkerAuthRequest(data *safejs.Value, eventChan eventChanWorker) error {
	authMsg, err := jsbridge.ParseEventDataField(data, "msg")
	if err != nil {
		return errors.Wrap(err, "could not parse 'msg' field")
	}

	rsp, err := sys.AuthCommon(string(authMsg))
	if err != nil {
		return errors.Wrap(err, "chunk upload authCommon failed")
	}

	if err := jsbridge.PostMessage(jsbridge.GetWorker(eventChan.workerID), jsbridge.MsgTypeAuthRsp,
		map[string]string{
			"data": rsp,
		}); err != nil {
		return errors.Wrap(err, "chunk upload postMessage failed")
	}

	return nil
}

func ProcessEventData(data safejs.Value) {
	fileMeta, formInfo, fileShards, thumbnailChunkData, err := parseEventData(data)
	var remotePath string
	if fileMeta != nil {
		remotePath = fileMeta.RemotePath
	}
	if err != nil {
		selfPostMessage(false, false, err.Error(), remotePath, 0, nil)
		return
	}
	wp, ok := hasherMap[fileMeta.RemotePath]
	if !ok {
		wp = workerProcess{
			hasher: CreateHasher(formInfo.ShardSize),
			wg:     &sync.WaitGroup{},
		}
		hasherMap[fileMeta.RemotePath] = wp
	} else if formInfo.ChunkStartIndex == 0 {
		wp.hasher = CreateHasher(formInfo.ShardSize)
		hasherMap[fileMeta.RemotePath] = wp
	}
	if formInfo.IsFinal {
		defer delete(hasherMap, fileMeta.RemotePath)
	}
	formBuilder := CreateChunkedUploadFormBuilder()
	uploadData, err := formBuilder.Build(fileMeta, wp.hasher, formInfo.ConnectionID, formInfo.ChunkSize, formInfo.ChunkStartIndex, formInfo.ChunkEndIndex, formInfo.IsFinal, formInfo.EncryptedKey, formInfo.EncryptedKeyPoint,
		fileShards, thumbnailChunkData, formInfo.ShardSize)
	if err != nil {
		selfPostMessage(false, false, err.Error(), remotePath, formInfo.ChunkEndIndex, nil)
		return
	}
	if formInfo.OnlyHash {
		if formInfo.IsFinal {
			finalResult := &FinalWorkerResult{
				DataHash:             uploadData.formData.DataHash,
				ThumbnailContentHash: uploadData.formData.ThumbnailContentHash,
			}
			selfPostMessage(true, true, "", remotePath, formInfo.ChunkEndIndex, finalResult)
		} else {
			selfPostMessage(true, false, "", remotePath, formInfo.ChunkEndIndex, nil)
		}
		return
	}
	blobberURL := os.Getenv("BLOBBER_URL")
	if !formInfo.IsFinal {
		wp.wg.Add(1)
	}
	go func(blobberData blobberData, remotePath string, wg *sync.WaitGroup) {
		if formInfo.IsFinal && len(blobberData.dataBuffers) > 1 {
			err = sendUploadRequest(blobberData.dataBuffers[:len(blobberData.dataBuffers)-1], blobberData.contentSlice[:len(blobberData.contentSlice)-1], blobberURL, formInfo.AllocationID, formInfo.AllocationTx, formInfo.HttpMethod)
			if err != nil {
				selfPostMessage(false, true, err.Error(), remotePath, formInfo.ChunkEndIndex, nil)
				return
			}
			wg.Wait()
			err = sendUploadRequest(blobberData.dataBuffers[len(blobberData.dataBuffers)-1:], blobberData.contentSlice[len(blobberData.contentSlice)-1:], blobberURL, formInfo.AllocationID, formInfo.AllocationTx, formInfo.HttpMethod)
			if err != nil {
				selfPostMessage(false, true, err.Error(), remotePath, formInfo.ChunkEndIndex, nil)
				return
			}
		} else {
			if formInfo.IsFinal {
				wg.Wait()
			} else {
				defer wg.Done()
			}
			err = sendUploadRequest(blobberData.dataBuffers, blobberData.contentSlice, blobberURL, formInfo.AllocationID, formInfo.AllocationTx, formInfo.HttpMethod)
			if err != nil {
				selfPostMessage(false, formInfo.IsFinal, err.Error(), remotePath, formInfo.ChunkEndIndex, nil)
				return
			}
		}
		if formInfo.IsFinal {
			finalResult := &FinalWorkerResult{
				DataHash:             blobberData.formData.DataHash,
				ThumbnailContentHash: blobberData.formData.ThumbnailContentHash,
			}
			selfPostMessage(true, true, "", remotePath, formInfo.ChunkEndIndex, finalResult)
		} else {
			selfPostMessage(true, false, "", remotePath, formInfo.ChunkEndIndex, nil)
		}
	}(uploadData, remotePath, wp.wg)

}

func InitHasherMap() {
	hasherMap = make(map[string]workerProcess)
}

func selfPostMessage(success, isFinal bool, errMsg, remotePath string, chunkEndIndex int, finalResult *FinalWorkerResult) {
	obj := js.Global().Get("Object").New()
	obj.Set("success", success)
	obj.Set("error", errMsg)
	obj.Set("isFinal", isFinal)
	obj.Set("chunkEndIndex", chunkEndIndex)
	obj.Set("remotePath", remotePath)
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

	// msgType is upload
	jsbridge.SetMsgType(&obj, jsbridge.MsgTypeUpload)

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
						if errors.Is(err, fasthttp.ErrConnectionClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, fasthttp.ErrDialTimeout) {
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
					if i == 2 {
						if err != nil {
							logger.Logger.Error("Retry limit exceeded for upload: ", err)
						}
						return errors.Throw(constants.ErrBadRequest, "Retry limit exceeded for upload")
					}
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
	return eg.Wait()
}

type eventChanWorker struct {
	C        <-chan worker.MessageEvent
	workerID string
}

func (su *ChunkedUpload) startProcessor() {
	su.listenChan = make(chan struct{}, su.uploadWorkers)
	su.processMap = make(map[int]zboxutil.Uint128)
	su.uploadWG.Add(1)
	go func() {
		respChan := make(chan error, 1)
		allEventChan := make([]eventChanWorker, len(su.blobbers))
		var pos uint64
		for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			blobber := su.blobbers[pos]
			webWorker := jsbridge.GetWorker(blobber.blobber.ID)
			if webWorker != nil {
				eventChan := make(chan worker.MessageEvent, su.uploadWorkers)
				err := webWorker.SubscribeToEvents(su.fileMeta.RemotePath, eventChan)
				if err != nil {
					logger.Logger.Error("error subscribing to events: ", err)
					su.ctxCncl(thrown.New("upload_failed", "Upload failed. Error subscribing to events"))
					return
				}
				defer webWorker.UnsubscribeToEvents(su.fileMeta.RemotePath)
				allEventChan[pos] = eventChanWorker{
					C:        eventChan,
					workerID: blobber.blobber.ID,
				}
			}
		}
		defer su.uploadWG.Done()
		for {
			go su.listen(allEventChan, respChan)
			select {
			case <-su.ctx.Done():
				return
			case err, ok := <-respChan:
				if !ok || err != nil {
					return
				}
				<-su.listenChan
			}
		}
	}()
}

func (su *ChunkedUpload) updateChunkProgress(chunkEndIndex int, pos uint64) {
	su.processMapLock.Lock()
	su.processMap[chunkEndIndex] = su.processMap[chunkEndIndex].Or(zboxutil.NewUint128(1).Lsh(pos))
	su.processMapLock.Unlock()
}
