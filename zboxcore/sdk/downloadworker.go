package sdk

import (
	"strings"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/bits"
	"os"
	"bytes"
	"sync"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/encoder"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
)

const (
	DOWNLOAD_CONTENT_FULL  = "full"
	DOWNLOAD_CONTENT_THUMB = "thumbnail"
)

type DownloadRequest struct {
	allocationID       string
	blobbers           []*blockchain.StorageNode
	datashards         int
	parityshards       int
	remotefilepath     string
	remotefilepathhash string
	localpath          string
	statusCallback     StatusCallback
	ctx                context.Context
	authTicket         *marker.AuthTicket
	wg                 *sync.WaitGroup
	downloadMask       uint32
	encryptedKey 	   string
	isDownloadCanceled bool
	completedCallback  func(remotepath string, remotepathhash string)
	contentMode        string
	Consensus
}

func (req *DownloadRequest) downloadBlock(blockNum int64, blockSize int64) ([]byte, error) {
	req.consensus = 0
	numDownloads := bits.OnesCount32(req.downloadMask)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numDownloads)
	rspCh := make(chan *downloadBlock, numDownloads)
	// Download from only specific blobbers
	c, pos := 0, 0
	for i := req.downloadMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		blockDownloadReq := &BlockDownloadRequest{}
		blockDownloadReq.allocationID = req.allocationID
		blockDownloadReq.authTicket = req.authTicket
		blockDownloadReq.blobber = req.blobbers[pos]
		blockDownloadReq.blobberIdx = pos
		blockDownloadReq.blockNum = blockNum
		blockDownloadReq.contentMode = req.contentMode
		blockDownloadReq.result = rspCh
		blockDownloadReq.wg = req.wg
		blockDownloadReq.ctx = req.ctx
		blockDownloadReq.remotefilepath = req.remotefilepath
		blockDownloadReq.remotefilepathhash = req.remotefilepathhash
		go AddBlockDownloadReq(blockDownloadReq)
		//go obj.downloadBlobberBlock(&obj.blobbers[pos], pos, path, blockNum, rspCh, isPathHash, authTicket)
		c++
	}
	//req.wg.Wait()
	shards := make([][]byte, len(req.blobbers))
	var decodeLen int
	success := 0
	var encscheme encryption.EncryptionScheme
	if len(req.encryptedKey) > 0 {
		encscheme = encryption.NewEncryptionScheme()
		encscheme.Initialize(client.GetClient().Mnemonic)
		encscheme.InitForDecryption("filetype:audio", req.encryptedKey)
	}
	

	for i := 0; i < numDownloads; i++ {
		result := <-rspCh
		if !result.Success {
			Logger.Error("Download block : ", req.blobbers[result.idx].Baseurl, result.err)
		} else {
			if len(req.encryptedKey) > 0 {
				headerBytes := result.Data[:(2 * 1024)]
				headerBytes = bytes.Trim(headerBytes, "\x00")
				headerString := string(headerBytes)
				encMsg := &encryption.EncryptedMessage{}
				encMsg.EncryptedData = result.Data[(2 * 1024):]
				fmt.Println("headerString = ", headerString)
				headerChecksums := strings.Split(headerString, ",")
				if len(headerChecksums) != 2 {
					Logger.Error("Block has invalid header", req.blobbers[result.idx].Baseurl)
					continue
				}
				encMsg.MessageChecksum, encMsg.OverallChecksum = headerChecksums[0], headerChecksums[1]
				encMsg.EncryptedKey = encscheme.GetEncryptedKey()
				decryptedBytes, err := encscheme.Decrypt(encMsg)
				if err != nil {
					Logger.Error("Block decryption failed", req.blobbers[result.idx].Baseurl, err)
					continue
				}
				shards[result.idx] = decryptedBytes
			} else {
				shards[result.idx] = result.Data
			}
			// All share should have equal length
			decodeLen = len(shards[result.idx])
			// fmt.Printf("[%d]:%s Size:%d\n", i, req.blobbers[result.idx].Baseurl, len(shards[result.idx]))
			success++
			if success >= req.datashards {
				go func(respChan chan *downloadBlock, num int) {
					for num > 0 {
						<-rspCh
						num--
					}
					return
				}(rspCh, numDownloads-success)
				break
			}
		}
	}

	erasureencoder, err := encoder.NewEncoder(req.datashards, req.parityshards)
	data, err := erasureencoder.Decode(shards, decodeLen)
	if err != nil {
		return []byte{}, fmt.Errorf("Block decode error %s", err.Error())
	}
	return data, nil
}

func (req *DownloadRequest) processDownload(ctx context.Context, a *Allocation) {
	remotePathCallback := req.remotefilepath
	if len(req.remotefilepath) == 0 {
		remotePathCallback = req.remotefilepathhash
	}
	if req.completedCallback != nil {
		defer req.completedCallback(req.remotefilepath, req.remotefilepathhash)
	}

	// Only download from the Blobbers passes the consensus
	var fileRef *fileref.FileRef
	listReq := &ListRequest{remotefilepath: req.remotefilepath, remotefilepathhash: req.remotefilepathhash, allocationID: req.allocationID, blobbers: req.blobbers, ctx: req.ctx}
	listReq.authToken = req.authTicket
	req.downloadMask, fileRef, _ = listReq.getFileConsensusFromBlobbers()
	if req.downloadMask == 0 || fileRef == nil {
		if req.statusCallback != nil {
			req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, fmt.Errorf("No minimum consensus for file meta data of file"))
		}
		return
	}

	size := fileRef.ActualFileSize
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		size = fileRef.ActualThumbnailSize
	}
	req.encryptedKey = fileRef.EncryptedKey
	// Calculate number of bytes per shard.
	perShard := (size + int64(req.datashards) - 1) / int64(req.datashards)
	chunkSizeWithHeader := int64(fileref.CHUNK_SIZE)
	if len(fileRef.EncryptedKey) > 0 {
		chunkSizeWithHeader -= 16
		chunkSizeWithHeader -= 2 * 1024
	}
	chunksPerShard := (perShard + chunkSizeWithHeader - 1) / chunkSizeWithHeader
	if len(fileRef.EncryptedKey) > 0 {
		perShard += chunksPerShard * (16 + (2 *1024))
	}

	wrFile, err := os.OpenFile(req.localpath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		if req.statusCallback != nil {
			Logger.Error(err.Error())
			req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, fmt.Errorf("Can't create local file %s", err.Error()))
		}
		return
	}
	defer wrFile.Close()
	req.isDownloadCanceled = false
	if req.statusCallback != nil {
		req.statusCallback.Started(req.allocationID, remotePathCallback, OpDownload, int(size))
	}

	Logger.Info("Download Size:", size, " Shard:", perShard, " chunks/shard:", chunksPerShard)
	downloaded := int(0)
	fH := sha1.New()
	mW := io.MultiWriter(fH, wrFile)
	for cnt := int64(0); cnt < chunksPerShard; cnt++ {
		blockSize := int64(math.Min(float64(perShard-(cnt*fileref.CHUNK_SIZE)), fileref.CHUNK_SIZE))
		data, err := req.downloadBlock(cnt+1, blockSize)
		if err != nil {
			os.Remove(req.localpath)
			if req.statusCallback != nil {
				req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, fmt.Errorf("Download failed for block %d. Error : %s", cnt+1, err.Error()))
			}
			return
		}
		if req.isDownloadCanceled {
			req.isDownloadCanceled = false
			os.Remove(req.localpath)
			if req.statusCallback != nil {
				req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, fmt.Errorf("Download aborted by user"))
			}
			return
		}
		n := int64(math.Min(float64(size), float64(len(data))))
		_, err = mW.Write(data[:n])
		if err != nil {
			os.Remove(req.localpath)
			if req.statusCallback != nil {
				req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, fmt.Errorf("Write file failed : %s", err.Error()))
			}
			return
		}
		downloaded = downloaded + int(n)
		size = size - n
		if req.statusCallback != nil {
			req.statusCallback.InProgress(req.allocationID, remotePathCallback, OpDownload, downloaded)
		}

	}
	calcHash := hex.EncodeToString(fH.Sum(nil))
	expectedHash := fileRef.ActualFileHash
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		expectedHash = fileRef.ActualThumbnailHash
	}
	if calcHash != expectedHash {
		os.Remove(req.localpath)
		if req.statusCallback != nil {
			req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, fmt.Errorf("File content didn't match with uploaded file"))
		}
		return
	}
	wrFile.Sync()
	wrFile.Close()
	wrFile, _ = os.Open(req.localpath)
	defer wrFile.Close()
	wrFile.Seek(0, 0)
	mimetype, _ := zboxutil.GetFileContentType(wrFile)
	if req.statusCallback != nil {
		req.statusCallback.Completed(req.allocationID, remotePathCallback, fileRef.Name, mimetype, int(fileRef.ActualFileSize), OpDownload)
	}
	return
}
