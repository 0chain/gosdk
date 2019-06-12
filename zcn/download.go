package zcn

import (
	"bytes"
	"context"
	"crypto/sha1"
	b64 "encoding/base64"
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
	"time"

	"0chain.net/clientsdk/util"
)

type downloadBlockResponse struct {
	reader io.ReadCloser
	idx    int
	err    error
}

type downloadBlock struct {
	Data []byte `json:"data"`
}

func (obj *Allocation) getLatestReadMarker(blobber *util.Blobber) (int64, error) {
	defer obj.wg.Done()
	req, err := util.NewLatestReadMarkerRequest(blobber.UrlRoot, obj.client)
	if err != nil {
		return -1, fmt.Errorf("New RM error: %s", err.Error())
	}
	var rm util.ReadMarker
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	err = httpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		// fmt.Println(blobber.UrlRoot, "Resp Status:", resp.StatusCode)
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Error: Resp : %s", err.Error())
		}
		if resp.StatusCode == http.StatusOK {
			err = json.Unmarshal(resp_body, &rm)
			if err != nil {
				return fmt.Errorf("RM response parse error: %s", err.Error())
			}
			return nil
		} else {
			return fmt.Errorf("%s Response Error: %s", blobber.UrlRoot, string(resp_body))
		}
	})
	if err == nil {
		Logger.Debug(blobber.UrlRoot, " Latest read marker count", rm.ReadCounter)
		blobber.ReadCounter = rm.ReadCounter
		return rm.ReadCounter, nil
	}
	errString := err.Error()
	if strings.Contains(errString, "entity_not_found") {
		Logger.Debug(blobber.UrlRoot, " No read marker found")
		blobber.ReadCounter = 0
		return 0, nil
	}
	Logger.Error("Latest RM error: ", err)
	return -1, err
}

func (obj *Allocation) downloadBlobberBlock(blobber *util.Blobber, blobberIdx int, path string, blockNum int64, rspCh chan<- *downloadBlockResponse, isPathHash bool, authTicket *authTicket) {
	defer obj.wg.Done()
	rm := util.NewReadMarker()
	rm.ClientID = obj.client.Id
	rm.ClientPublicKey = obj.client.PublicKey
	rm.BlobberID = blobber.Id
	rm.AllocationID = obj.allocationId
	rm.OwnerID = obj.client.Id
	rm.Timestamp = util.Now()
	rm.ReadCounter = blobber.ReadCounter + 1
	err := rm.Sign(obj.client.PrivateKey)
	if err != nil {
		rspCh <- &downloadBlockResponse{reader: nil, idx: blobberIdx, err: fmt.Errorf("Error: Signing readmarker failed: %s", err.Error())}
		return
	}
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	rmData, err := json.Marshal(rm)
	if err != nil {
		rspCh <- &downloadBlockResponse{reader: nil, idx: blobberIdx, err: fmt.Errorf("Error creating readmarker: %s", err.Error())}
		return
	}
	if isPathHash {
		formWriter.WriteField("path_hash", path)
	} else {
		formWriter.WriteField("path", path)
	}

	formWriter.WriteField("block_num", fmt.Sprintf("%d", blockNum))
	formWriter.WriteField("read_marker", string(rmData))
	if authTicket != nil {
		authTicketBytes, _ := json.Marshal(authTicket)
		formWriter.WriteField("auth_token", string(authTicketBytes))
	}
	formWriter.Close()
	req, err := util.NewDownloadRequest(blobber.UrlRoot, obj.allocationId, obj.client, body)
	if err != nil {
		rspCh <- &downloadBlockResponse{reader: nil, idx: blobberIdx, err: fmt.Errorf("Error creating download request: %s", err.Error())}
		return
	}
	req.Header.Add("Content-Type", formWriter.FormDataContentType())
	// TODO: Fix the timeout
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	_ = httpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			rspCh <- &downloadBlockResponse{reader: nil, idx: blobberIdx, err: err}
			return err
		}
		if resp.StatusCode == http.StatusOK {
			obj.consensus++
			blobber.ReadCounter++
			rspCh <- &downloadBlockResponse{reader: resp.Body, idx: blobberIdx, err: nil}
		} else {
			defer resp.Body.Close()
			resp_body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				rspCh <- &downloadBlockResponse{reader: nil, idx: blobberIdx, err: err}
				return err
			}
			err = fmt.Errorf("Response Error: %s", string(resp_body))
			rspCh <- &downloadBlockResponse{reader: nil, idx: blobberIdx, err: err}
			return err
		}
		return nil
	})
}

func (obj *Allocation) downloadBlock(path string, blockNum int64, blockSize int64, isPathHash bool, authTicket *authTicket) ([]byte, error) {
	obj.consensus = 0
	numDownloads := bits.OnesCount32(obj.downloadMask)
	obj.wg.Add(numDownloads)
	rspCh := make(chan *downloadBlockResponse, numDownloads)
	// Download from only specific blobbers
	c, pos := 0, 0
	for i := obj.downloadMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		go obj.downloadBlobberBlock(&obj.blobbers[pos], pos, path, blockNum, rspCh, isPathHash, authTicket)
		c++
	}
	//obj.wg.Wait()
	shards := make([][]byte, len(obj.blobbers))
	var decodeLen int
	success := 0
	for i := 0; i < numDownloads; i++ {
		result := <-rspCh
		if result.err != nil {
			Logger.Error("Download block : ", obj.blobbers[result.idx].UrlRoot, result.err)
		} else if result.reader != nil {
			defer result.reader.Close()
			response, err := ioutil.ReadAll(result.reader)
			if err != nil {
				return []byte{}, fmt.Errorf("[%d] Read error:%s\n", result.idx, err.Error())
			}
			var rspData downloadBlock
			err = json.Unmarshal(response, &rspData)
			if err != nil {
				return []byte{}, fmt.Errorf("[%d] Json decode error:%s\n", result.idx, err.Error())
			}
			shards[result.idx] = rspData.Data
			// All share should have equal length
			decodeLen = len(shards[result.idx])
			// fmt.Printf("[%d]:%s Size:%d\n", i, obj.blobbers[result.idx].UrlRoot, len(shards[result.idx]))
			success++
			if success >= obj.encoder.iDataShards {
				go func(respChan chan *downloadBlockResponse, num int) {
					if num <= 0 {
						return
					}
					result := <-rspCh
					for i := 0; i < num; i++ {
						if result.reader != nil {
							defer result.reader.Close()
						}
					}
				}(rspCh, numDownloads-success)
				break
			}
		}
	}

	data, err := obj.encoder.decode(shards, decodeLen)
	if err != nil {
		return []byte{}, fmt.Errorf("Block decode error %s", err.Error())
	}
	return data, nil
}

func (obj *Allocation) downloadFileContent(fileInfo *util.FileDirInfo, localPath, remotePath string, isPathHash bool, authTicket *authTicket, statusCb StatusCallback) error {
	obj.pauseBgSync()
	defer obj.resumeBgSync()
	size := fileInfo.Size
	// TODO: Handle latest readmarker failures.
	obj.wg.Add(len(obj.blobbers))
	for i := 0; i < len(obj.blobbers); i++ {
		go obj.getLatestReadMarker(&obj.blobbers[i])
	}
	obj.wg.Wait()
	if isPathHash {
		obj.downloadMask = ((1 << uint32(len(obj.blobbers))) - 1)
	} else {
		// Only download from the Blobbers passes the consensus
		obj.downloadMask = obj.getFileConsensusFromBlobbers(remotePath)
		if obj.downloadMask == 0 {
			return fmt.Errorf("No minimum consensus for download")
		}
	}
	// Calculate number of bytes per shard.
	perShard := (size + int64(obj.encoder.iDataShards) - 1) / int64(obj.encoder.iDataShards)
	chunksPerShard := (perShard + int64(CHUNK_SIZE) - 1) / CHUNK_SIZE
	wrFile, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Can't create local file %s", err.Error())
	}
	defer wrFile.Close()
	obj.isDownloadCanceled = false
	statusCb.Started(obj.allocationId, remotePath, OpDownload, int(size))
	Logger.Debug("Download Size:", size, " Shard:", perShard, " chunks/shard:", chunksPerShard)
	downloaded := int(0)
	fH := sha1.New()
	mW := io.MultiWriter(fH, wrFile)
	for cnt := int64(0); cnt < chunksPerShard; cnt++ {
		blockSize := int64(math.Min(float64(perShard-(cnt*CHUNK_SIZE)), CHUNK_SIZE))
		data, err := obj.downloadBlock(remotePath, cnt+1, blockSize, isPathHash, authTicket)
		if err != nil {
			os.Remove(localPath)
			return fmt.Errorf("Download failed for block %d. Error : %s", cnt+1, err.Error())
		}
		if obj.isDownloadCanceled {
			obj.isDownloadCanceled = false
			os.Remove(localPath)
			return fmt.Errorf("Download aborted by user")
		}
		n := int64(math.Min(float64(size), float64(len(data))))
		_, err = mW.Write(data[:n])
		if err != nil {
			os.Remove(localPath)
			return fmt.Errorf("Write file failed : %s", err.Error())
		}
		downloaded = downloaded + int(n)
		size = size - n
		statusCb.InProgress(obj.allocationId, remotePath, OpDownload, downloaded)
	}
	calcHash := hex.EncodeToString(fH.Sum(nil))
	if calcHash != fileInfo.Hash {
		os.Remove(localPath)
		return fmt.Errorf("File content didn't match with uploaded file")
	}
	wrFile.Sync()
	wrFile.Close()
	wrFile, _ = os.Open(localPath)
	defer wrFile.Close()
	wrFile.Seek(0, 0)
	mimetype, _ := util.GetFileContentType(wrFile)
	statusCb.Completed(obj.allocationId, remotePath, fileInfo.Name, mimetype, int(size), OpDownload)
	return nil
}

func (obj *Allocation) DownloadFileFromShareLink(localPath string, authTokenB64 string, statusCb StatusCallback) error {
	if _, err := os.Stat(localPath); err == nil {
		return fmt.Errorf("Local file already exists '%s'", localPath)
	}

	sDec, err := b64.StdEncoding.DecodeString(authTokenB64)
	if err != nil {
		return fmt.Errorf("Error decoding the encoded auth ticket")
	}
	at := &authTicket{}
	err = json.Unmarshal(sDec, at)
	if err != nil {
		return fmt.Errorf("Error decoding json for auth ticket")
	}
	if len(obj.blobbers) <= 1 {
		return noBLOBBERS
	}

	fileInfos := make([]*util.FileDirInfo, len(obj.blobbers))
	obj.pauseBgSync()
	defer obj.resumeBgSync()
	// TODO: Handle latest readmarker failures.
	obj.wg.Add(len(obj.blobbers))
	for i := 0; i < len(obj.blobbers); i++ {
		go func(idx int) {
			fileInfos[idx], err = obj.getFileMetaInfoFromBlobber(&obj.blobbers[idx], at.FilePathHash, at, true)
			if err != nil {
				Logger.Error("Error in getting the file meta response.", err)
			}
		}(i)
	}
	obj.wg.Wait()
	fileInfoHashCount := make(map[string]float32)
	obj.consensus = 0
	majorityFileInfo := &util.FileDirInfo{}
	for _, fileinfo := range fileInfos {
		//fmt.Println(fileinfo)
		if fileinfo != nil {
			fileInfoHashCount[fileinfo.GetInfoHash()]++
			if fileInfoHashCount[fileinfo.GetInfoHash()] > obj.consensus {
				majorityFileInfo = fileinfo
				obj.consensus = fileInfoHashCount[fileinfo.GetInfoHash()]
			}
		}
	}

	if !obj.isConsensusMin() {
		return fmt.Errorf("Cannot get concensus on the file info from blobbers")
	}

	return obj.downloadFileContent(majorityFileInfo, localPath, at.FilePathHash, true, at, statusCb)
}

func (obj *Allocation) DownloadFile(remotePath, localPath string, statusCb StatusCallback) error {
	if stat, err := os.Stat(localPath); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("Local file already exists '%s'", localPath)
		} else {
			localPath = strings.TrimRight(localPath, "/")
			_, rFile := filepath.Split(remotePath)
			localPath = fmt.Sprintf("%s/%s", localPath, rFile)
			if _, err := os.Stat(localPath); err == nil {
				return fmt.Errorf("Local file already exists '%s'", localPath)
			}
		}
	}
	if len(obj.blobbers) <= 1 {
		return noBLOBBERS
	}
	fileInfo := util.GetFileInfo(&obj.dirTree, remotePath)
	if fileInfo == nil {
		// TODO: Use list API from blobber to confirm
		return fmt.Errorf("Remote file doesn't exists")
	}
	return obj.downloadFileContent(fileInfo, localPath, remotePath, false, nil, statusCb)
}

// Cancel ongoing download
func (obj *Allocation) DownloadCancel() {
	obj.isDownloadCanceled = true
}
