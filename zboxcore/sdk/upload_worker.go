package sdk

import (
	"fmt"
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
	refs           []fileref.FileRef
	isUpdate       bool
	statusCallback StatusCallback
	opCode         int
}

func (uo *UploadOperation) Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error) {
	cu, err := CreateChunkedUpload(uo.workdir, allocObj, uo.fileMeta, uo.fileReader, uo.isUpdate, false, connectionID, uo.opts...)
	uo.statusCallback = cu.statusCallback
	uo.opCode = cu.opCode
	if err != nil {
		return nil, cu.uploadMask, err
	}
	err = cu.process()
	if err != nil {
		cu.ctxCncl()
		return nil, cu.uploadMask, err
	}

	var pos uint64
	numList := len(cu.blobbers)
	uo.refs = make([]fileref.FileRef, numList)
	for i := cu.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		uo.refs[pos] = *cu.blobbers[pos].fileRef
		uo.refs[pos].ChunkSize = cu.chunkSize
	}

	l.Logger.Info("Completed the upload")
	return nil, cu.uploadMask, nil
}

func (uo *UploadOperation) buildChange(_ []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange {
	changes := make([]allocationchange.AllocationChange, len(uo.refs))
	for idx, ref := range uo.refs {
		ref := ref
		if uo.isUpdate {
			change := &allocationchange.UpdateFileChange{}
			change.NewFile = &ref
			change.Operation = constants.FileOperationUpdate
			change.Size = ref.Size
			changes[idx] = change
			continue
		}
		newChange := &allocationchange.NewFileChange{}
		newChange.File = &ref

		newChange.Operation = constants.FileOperationInsert
		newChange.Size = ref.Size
		newChange.Uuid = uid
		changes[idx] = newChange
	}
	return changes

}

func (uo *UploadOperation) build(workdir string, fileMeta FileMeta, fileReader io.Reader, isUpdate bool, opts ...ChunkedUploadOption) {
	uo.workdir = workdir
	uo.fileMeta = fileMeta
	uo.fileReader = fileReader
	uo.opts = opts
	uo.isUpdate = isUpdate
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
	if uo.isUpdate {
		otr, err := allocationObj.GetRefs(uo.fileMeta.RemotePath, "", "", "", fileref.FILE, "regular", 0, 1)
		if err != nil {
			l.Logger.Error(err)
			return thrown.New("chunk_upload", err.Error())
		}
		if len(otr.Refs) != 1 {
			return thrown.New("chunk_upload", fmt.Sprintf("Expected refs 1, got %d", len(otr.Refs)))
		}
	}
	return nil
}

func (uo *UploadOperation) Completed(allocObj *Allocation) {
	if uo.statusCallback != nil {
		uo.statusCallback.Completed(allocObj.ID, uo.fileMeta.Path, uo.fileMeta.RemoteName, uo.fileMeta.MimeType, int(uo.fileMeta.ActualSize), uo.opCode)
	}
}

func (uo *UploadOperation) Error(allocObj *Allocation, consensus int, err error) {
	if consensus != 0 {
		l.Logger.Info("Commit consensus failed, Deleting remote file....")
		allocObj.deleteFile(uo.fileMeta.RemotePath, consensus, consensus) //nolint
	}
	if uo.statusCallback != nil {
		uo.statusCallback.Error(allocObj.ID, uo.fileMeta.RemotePath, uo.opCode, err)
	}
}
