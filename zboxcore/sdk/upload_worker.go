package sdk

import (
	"fmt"
	"io"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/google/uuid"
)



type UploadOperation struct {
	workdir string
	fileMeta FileMeta
	fileReader io.Reader
	opts []ChunkedUploadOption
	refs []fileref.FileRef
	isUpdate bool
	

}
func (uo *UploadOperation) Process(allocDetails AllocationDetails,connectionID string, blobbers []*blockchain.StorageNode) ([]fileref.RefEntity, error){
	cu, err := CreateChunkedUpload(uo.workdir, allocDetails.allocationObj, uo.fileMeta, uo.fileReader, uo.isUpdate, false, connectionID, uo.opts...);
	if err != nil {
		return nil, err
	}
	err = cu.process()
	
	if err != nil {
		l.Logger.Info("temporaray start is giving error: ", err);
		return nil, nil;
	}


	if err != nil {
		return nil, err
	}


	var pos uint64
	numList := len(cu.blobbers)
	objectTreeRefsEntity := make([]fileref.RefEntity, numList)
	uo.refs = make([]fileref.FileRef, numList)
	for i := cu.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		objectTreeRefsEntity[pos] = cu.blobbers[pos].fileRef;
		uo.refs[pos] = *cu.blobbers[pos].fileRef
		uo.refs[pos].NumBlocks =  int64(cu.progress.ChunkIndex + 1)
		uo.refs[pos].ChunkSize = cu.chunkSize
	}

	l.Logger.Info("Completed the upload")
	// return objectTreeRefsEntity, nil 
	return objectTreeRefsEntity, nil
}

func (uo *UploadOperation) buildChange(dummyRefs []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange {
	changes := make([]allocationchange.AllocationChange, len(uo.refs))
	for idx, ref := range uo.refs {
		ref := ref
		if uo.isUpdate {
			change := &allocationchange.UpdateFileChange{}
			change.NewFile = &ref
			change.NumBlocks = ref.NumBlocks
			change.Operation = constants.FileOperationUpdate
			change.Size = ref.Size
			changes[idx] = change
			continue;
		}
		newChange := &allocationchange.NewFileChange{}
		newChange.File = &ref
		newChange.NumBlocks = ref.NumBlocks
		
		newChange.Operation = constants.FileOperationInsert
		newChange.Size = ref.Size
		newChange.Uuid = uid
		changes[idx] = newChange
	}
	fmt.Println("The length of uo.refs is ", len(uo.refs));
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
		return  thrown.Throw(constants.ErrFileOptionNotPermitted, "file_option_not_permitted ")
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
			return  thrown.New("chunk_upload", err.Error())
		}
		if len(otr.Refs) != 1 {
			return thrown.New("chunk_upload", fmt.Sprintf("Expected refs 1, got %d", len(otr.Refs)))
		}
	}
	return nil
	
}
