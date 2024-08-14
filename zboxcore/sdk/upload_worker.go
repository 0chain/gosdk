package sdk

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UploadOperation struct {
	refs          []*fileref.FileRef
	opCode        int
	chunkedUpload *ChunkedUpload
	isUpdate      bool
	isDownload    bool
}

var ErrPauseUpload = errors.New("upload paused by user")

func (uo *UploadOperation) Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error) {
	if uo.isDownload {
		if f, ok := uo.chunkedUpload.fileReader.(*sys.MemChanFile); ok {
			err := allocObj.DownloadFileToFileHandler(f, uo.chunkedUpload.fileMeta.RemotePath, false, nil, true, WithFileCallback(func() {
				f.Close() //nolint:errcheck
			}))
			if err != nil {
				l.Logger.Error("DownloadFileToFileHandler Failed", zap.String("path", uo.chunkedUpload.fileMeta.RemotePath), zap.Error(err))
				return nil, uo.chunkedUpload.uploadMask, err
			}
		}
	}
	err := uo.chunkedUpload.process()
	if err != nil {
		l.Logger.Error("UploadOperation Failed", zap.String("name", uo.chunkedUpload.fileMeta.RemoteName), zap.Error(err))
		return nil, uo.chunkedUpload.uploadMask, err
	}
	var pos uint64
	numList := len(uo.chunkedUpload.blobbers)
	uo.refs = make([]*fileref.FileRef, numList)
	for i := uo.chunkedUpload.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		uo.refs[pos] = uo.chunkedUpload.blobbers[pos].fileRef
		uo.refs[pos].ChunkSize = uo.chunkedUpload.chunkSize
		remotePath := uo.chunkedUpload.fileMeta.RemotePath
		allocationID := allocObj.ID
		if singleClientMode {
			lookuphash := fileref.GetReferenceLookup(allocationID, remotePath)
			cacheKey := fileref.GetCacheKey(lookuphash, uo.chunkedUpload.blobbers[pos].blobber.ID)
			fileref.DeleteFileRef(cacheKey)
		}
	}
	l.Logger.Debug("UploadOperation Success", zap.String("name", uo.chunkedUpload.fileMeta.RemoteName))
	return nil, uo.chunkedUpload.uploadMask, nil
}

func (uo *UploadOperation) buildChange(_ []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange {
	changes := make([]allocationchange.AllocationChange, len(uo.refs))
	for idx, ref := range uo.refs {
		if ref == nil {
			change := &allocationchange.EmptyFileChange{}
			changes[idx] = change
			continue
		}
		if uo.isUpdate {
			change := &allocationchange.UpdateFileChange{}
			change.NewFile = ref
			change.NumBlocks = ref.NumBlocks
			change.Operation = constants.FileOperationUpdate
			change.Size = ref.Size
			changes[idx] = change
			continue
		}
		newChange := &allocationchange.NewFileChange{}
		newChange.File = ref
		newChange.NumBlocks = ref.NumBlocks
		newChange.Operation = constants.FileOperationInsert
		newChange.Size = ref.Size
		newChange.Uuid = uid
		changes[idx] = newChange
	}
	return changes

}

func (uo *UploadOperation) Verify(allocationObj *Allocation) error {
	return nil
}

func (uo *UploadOperation) Completed(allocObj *Allocation) {
	if uo.chunkedUpload.progressStorer != nil {
		uo.chunkedUpload.removeProgress()
	}
	cancelLock.Lock()
	delete(CancelOpCtx, uo.chunkedUpload.fileMeta.RemotePath)
	cancelLock.Unlock()
	if uo.chunkedUpload.statusCallback != nil {
		uo.chunkedUpload.statusCallback.Completed(allocObj.ID, uo.chunkedUpload.fileMeta.RemotePath, uo.chunkedUpload.fileMeta.RemoteName, uo.chunkedUpload.fileMeta.MimeType, int(uo.chunkedUpload.fileMeta.ActualSize), uo.opCode)
	}
}

func (uo *UploadOperation) Error(allocObj *Allocation, consensus int, err error) {
	if uo.chunkedUpload.progressStorer != nil && !strings.Contains(err.Error(), "context") && !errors.Is(err, ErrPauseUpload) {
		uo.chunkedUpload.removeProgress()
	}
	cancelLock.Lock()
	delete(CancelOpCtx, uo.chunkedUpload.fileMeta.RemotePath)
	cancelLock.Unlock()
	if uo.chunkedUpload.statusCallback != nil {
		uo.chunkedUpload.statusCallback.Error(allocObj.ID, uo.chunkedUpload.fileMeta.RemotePath, uo.opCode, err)
	}
}

func NewUploadOperation(ctx context.Context, workdir string, allocObj *Allocation, connectionID string, fileMeta FileMeta, fileReader io.Reader, isUpdate, isWebstreaming, isRepair, isMemoryDownload, isStreamUpload bool, opts ...ChunkedUploadOption) (*UploadOperation, string, error) {
	uo := &UploadOperation{}
	if fileMeta.ActualSize == 0 && !isStreamUpload {
		byteReader := bytes.NewReader([]byte(
			emptyFileDataHash))
		fileReader = byteReader
		opts = append(opts, WithActualHash(emptyFileDataHash))
		fileMeta.ActualSize = int64(len(emptyFileDataHash))
	}

	cu, err := CreateChunkedUpload(ctx, workdir, allocObj, fileMeta, fileReader, isUpdate, isRepair, isWebstreaming, connectionID, opts...)
	if err != nil {
		return nil, "", err
	}

	uo.chunkedUpload = cu
	uo.opCode = cu.opCode
	uo.isUpdate = isUpdate
	uo.isDownload = isMemoryDownload
	return uo, cu.progress.ConnectionID, nil
}
