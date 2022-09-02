package sdk

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
)

type NewDownloadRequest struct {
	DownloadRequest
}

func (req *NewDownloadRequest) downloadBlock(startBlock int64, totalBlock int) ([]byte, error) {
	return nil, nil
}

func (req *NewDownloadRequest) processDownload(ctx context.Context) {
	if req.completedCallback != nil {
		defer req.completedCallback(req.remotefilepath, req.remotefilepathhash)
	}

	remotePathCB := req.remotefilepath
	if remotePathCB == "" {
		remotePathCB = req.remotefilepathhash
	}
	fRef, err := req.getFileRef(remotePathCB)
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}

	size, chunksPerShard, actualPerShard, err := req.calculateShardsParams(fRef, remotePathCB)
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}

	logger.Logger.Info(
		fmt.Sprintf("Downloading file with size: %d from start block: %d and end block: %d. "+
			"Actual size per blobber: %d", size, req.startBlock, req.endBlock, actualPerShard),
	)

	f, err := req.getFileHandler(remotePathCB)
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	defer f.Close()

	var isFullDownload bool
	fileHasher := createDownloadHasher(req.chunkSize, req.datashards, fRef.EncryptedKey != "")
	var mW io.Writer
	if req.startBlock == 0 && req.endBlock == chunksPerShard {
		isFullDownload = true
		mW = io.MultiWriter(fileHasher, f)
	} else {
		mW = io.MultiWriter(f)
	}

	var downloaded int
	startBlock, endBlock, numBlocks := req.startBlock, req.endBlock, req.numBlocks

	for startBlock < endBlock {
		if startBlock+numBlocks > endBlock {
			numBlocks = endBlock - startBlock
		}
		logger.Logger.Info("Downloading block ", startBlock+1, " - ", startBlock+numBlocks)

		data, err := req.downloadBlock(startBlock+1, int(numBlocks))
		if err != nil {
			req.errorCB(errors.Wrap(err, fmt.Sprintf("Download failed for block %d. ", startBlock+1)), remotePathCB)
			return
		}
		if req.isDownloadCanceled {
			req.errorCB(errors.New("download_abort", "Download aborted by user"), remotePathCB)
			return
		}

		n := int64(math.Min(float64(size), float64(len(data))))
		_, err = mW.Write(data[:n])

		if err != nil {
			req.errorCB(errors.Wrap(err, "Write file failed"), remotePathCB)
			return
		}
		downloaded = downloaded + int(n)
		size = size - n

		if req.statusCallback != nil {
			req.statusCallback.InProgress(req.allocationID, remotePathCB, OpDownload, downloaded, data)
		}
		if (startBlock + numBlocks) > endBlock {
			startBlock += endBlock - startBlock
		} else {
			startBlock += numBlocks
		}
	}

	if isFullDownload {
		err := req.checkContentHash(fRef, *fileHasher, remotePathCB)
		if err != nil {
			logger.Logger.Error(err)
			return
		}
	}

	f.Sync()
	if req.statusCallback != nil {
		req.statusCallback.Completed(
			req.allocationID, remotePathCB, fRef.Name, "", int(fRef.ActualFileSize), OpDownload)
	}
}

func (req *NewDownloadRequest) errorCB(err error, remotePathCB string) {
	sys.Files.Remove(req.localpath) //nolint: errcheck
	if req.statusCallback != nil {
		req.statusCallback.Error(
			req.allocationID, remotePathCB, OpDownload, err)
	}
	return
}

func (req *NewDownloadRequest) checkContentHash(
	fRef *fileref.FileRef, fileHasher downloadHasher, remotepathCB string) (err error) {

	merkleRoot := fileHasher.GetMerkleRoot()
	expectedHash := fRef.ActualFileHash
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		expectedHash = fRef.ActualThumbnailHash
	}

	if expectedHash != merkleRoot {
		err = errors.New("merkle_root_mismatch", "File content didn't match with uploaded file")
		req.errorCB(err, remotepathCB)
		return
	}
	return nil
}

func (req *NewDownloadRequest) getFileHandler(remotePathCB string) (sys.File, error) {
	f, err := sys.Files.OpenFile(req.localpath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		req.errorCB(errors.Wrap(err, "Can't create local file"), remotePathCB)
		return nil, err
	}
	return f, nil
}

func (req *NewDownloadRequest) calculateShardsParams(
	fRef *fileref.FileRef, remotePathCB string) (
	size, chunksPerShard, actualPerShard int64, err error) {

	size = fRef.ActualFileSize
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		size = fRef.ActualThumbnailSize
	}
	req.encryptedKey = fRef.EncryptedKey
	req.chunkSize = int(fRef.ChunkSize)

	// fRef.ActualFileSize is size of file that does not include encryption bytes.
	// that is why, actualPerShard will have different value for encrypted file.
	effectivePerShard := (size + int64(req.datashards) - 1) / int64(req.datashards)
	effectiveChunkSize := fRef.ChunkSize
	if fRef.EncryptedKey != "" {
		effectiveChunkSize -= EncryptionHeaderSize + EncryptedDataPaddingSize
	}

	chunksPerShard = (effectivePerShard + effectiveChunkSize - 1) / effectiveChunkSize
	actualPerShard = chunksPerShard * fRef.ChunkSize
	if req.endBlock == 0 {
		req.endBlock = chunksPerShard
	}

	if req.startBlock >= req.endBlock {
		err = errors.New("invalid_block_num", "start block should be less than end block")
		req.errorCB(err, remotePathCB)
		return 0, 0, 0, err
	}

	return
}

func (req *NewDownloadRequest) getFileRef(remotePathCB string) (fRef *fileref.FileRef, err error) {
	listReq := &ListRequest{
		remotefilepath:     req.remotefilepath,
		remotefilepathhash: req.remotefilepathhash,
		allocationID:       req.allocationID,
		allocationTx:       req.allocationTx,
		blobbers:           req.blobbers,
		authToken:          req.authTicket,
		Consensus: Consensus{
			mu:              &sync.RWMutex{},
			fullconsensus:   req.fullconsensus,
			consensusThresh: req.consensusThresh,
		},
		ctx: req.ctx,
	}

	req.downloadMask, fRef, _ = listReq.getFileConsensusFromBlobbers()
	if req.downloadMask.Equals64(0) || fRef == nil {
		err = errors.New("consensus_not_met", "No minimum consensus for file meta data of file")
		req.errorCB(err, remotePathCB)
		return
	}

	if fRef.Type == fileref.DIRECTORY {
		err = errors.New("invalid_operation", "cannot downoad directory")
		req.errorCB(err, remotePathCB)
		return
	}

	// the ChunkSize value can't be less than 0kb
	if fRef.ChunkSize <= 0 {
		err = errors.New("invalid_chunk_size", "File ChunkSize value is not permitted")
		req.errorCB(err, remotePathCB)
		return
	}

	return
}
