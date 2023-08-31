package sdk

import (
	"io"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/google/uuid"
)

type UploadOperation struct {
	workdir        string
	fileMeta       FileMeta
	fileReader     io.Reader
	opts           []ChunkedUploadOption
	refs           []*fileref.FileRef
	isUpdate       bool
	isWebstreaming bool
	statusCallback StatusCallback
	opCode         int
	progressInfo   ProgressInfo
	chunkedUpload  *ChunkedUpload
}

type ProgressInfo struct {
	ID             string
	ProgressStorer ChunkedUploadProgressStorer
}

func (uo *UploadOperation) Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error) {
	l.Logger.Info("Connection ID: ", connectionID)
	err := uo.chunkedUpload.process()
	if err != nil {
		uo.chunkedUpload.ctxCncl()
		return nil, uo.chunkedUpload.uploadMask, err
	}

	var pos uint64
	numList := len(uo.chunkedUpload.blobbers)
	uo.refs = make([]*fileref.FileRef, numList)
	for i := uo.chunkedUpload.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		uo.refs[pos] = uo.chunkedUpload.blobbers[pos].fileRef
		uo.refs[pos].ChunkSize = uo.chunkedUpload.chunkSize
	}

	l.Logger.Info("Completed the upload")
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
	if allocationObj == nil {
		return thrown.Throw(constants.ErrInvalidParameter, "allocationObj")
	}

	if !uo.isUpdate && !allocationObj.CanUpload() || uo.isUpdate && !allocationObj.CanUpdate() {
		return thrown.Throw(constants.ErrFileOptionNotPermitted, "file_option_not_permitted ")
	}

	err := ValidateRemoteFileName(uo.fileMeta.RemoteName)
	if err != nil {
		return err
	}
	spaceLeft := allocationObj.Size
	if allocationObj.Stats != nil {
		spaceLeft -= allocationObj.Stats.UsedSize
	}

	if uo.isUpdate {
		f, err := allocationObj.GetFileMeta(uo.fileMeta.RemotePath)
		if err != nil {
			return err
		}
		spaceLeft += f.ActualFileSize
	}
	if uo.fileMeta.ActualSize > spaceLeft {
		return ErrNoEnoughSpaceLeftInAllocation
	}
	return nil
}

func (uo *UploadOperation) Completed(allocObj *Allocation) {
	if uo.statusCallback != nil {
		uo.statusCallback.Completed(allocObj.ID, uo.fileMeta.RemotePath, uo.fileMeta.RemoteName, uo.fileMeta.MimeType, int(uo.fileMeta.ActualSize), uo.opCode)
	}
}

func (uo *UploadOperation) Error(allocObj *Allocation, consensus int, err error) {
	if uo.statusCallback != nil {
		uo.statusCallback.Error(allocObj.ID, uo.fileMeta.RemotePath, uo.opCode, err)
	}
}

func NewUploadOperation(workdir string, allocObj *Allocation, connectionID string, fileMeta FileMeta, fileReader io.Reader, isUpdate, isWebstreaming bool, opts ...ChunkedUploadOption) (*UploadOperation, string, error) {
	uo := &UploadOperation{
		workdir:        workdir,
		fileMeta:       fileMeta,
		fileReader:     fileReader,
		opts:           opts,
		isUpdate:       isUpdate,
		isWebstreaming: isWebstreaming,
	}

	cu, err := CreateChunkedUpload(uo.workdir, allocObj, uo.fileMeta, uo.fileReader, uo.isUpdate, false, uo.isWebstreaming, connectionID, uo.opts...)
	if err != nil {
		return nil, "", err
	}

	uo.chunkedUpload = cu
	uo.progressInfo = ProgressInfo{
		ID:             cu.progress.ID,
		ProgressStorer: cu.progressStorer,
	}
	uo.statusCallback = cu.statusCallback
	uo.opCode = cu.opCode
	uo.fileMeta = cu.fileMeta
	uo.fileReader = cu.fileReader

	return uo, cu.progress.ConnectionID, nil
}
