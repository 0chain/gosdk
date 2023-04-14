package sdk

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	// "github.com/0chain/errors"
	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/google/uuid"
)

type AllocationDetails struct {
	allocationObj *Allocation
	allocationID string
	allocationTx string
}
type Operationer interface {
	Process(allocDetails AllocationDetails,connectionID string, blobbers []*blockchain.StorageNode) ([]fileref.RefEntity, error)
	buildChange(refs []fileref.RefEntity, uid uuid.UUID ) []allocationchange.AllocationChange
}



type MultiOperation struct {
	connectionID  string
	operations    []Operationer
	AllocationDetails
	blobbers      []*blockchain.StorageNode
	ctx           context.Context
	ctxCncl       context.CancelFunc
	operationMask zboxutil.Uint128
	maskMU        *sync.Mutex
	Consensus

	changes [][]allocationchange.AllocationChange
}

func (mo *MultiOperation) Process() error {
	l.Logger.Info("MultiOperation Process start");
	wg := &sync.WaitGroup{}
	mo.changes = make([][]allocationchange.AllocationChange, len(mo.operations))
	l.Logger.Info("len of mo.oper: ", len(mo.operations))
	ctx, ctxCncl := context.WithCancel(context.Background())
	defer ctxCncl()
	errs := make(chan error, 1)

	for idx, op := range mo.operations {
		wg.Add(1)
		allocDetails := AllocationDetails{allocationObj: mo.allocationObj, allocationID: mo.allocationID, allocationTx: mo.allocationTx}
		// Here make it goroutine
		go func(op Operationer, idx int) {
			defer wg.Done()

			// Check for other goroutines signal
			select {
			case <-ctx.Done():
				return
			default:
			}

			refs, err := op.Process(allocDetails, mo.connectionID, mo.blobbers) // Process with each blobber

			if err != nil {
				l.Logger.Error(err);

				select {
				case errs <- thrown.New("", err.Error()):
				default:
				}
				ctxCncl()

				return 
			}
			uid := util.GetNewUUID()
			changes := op.buildChange(refs, uid)

			// change := op.buildChange(refs, uid)
			mo.changes[idx] = changes
			// mo.changes = append(mo.changes, change)
		}(op, idx)
	}
	wg.Wait()
	if ctx.Err() != nil {
		return <-errs
	}
	l.Logger.Info("Individual Operation process done");

	// Take transpose of mo.change because it will be easier to iterate mo if it contains blobber changes
	// in row instead of column. Currently mo.change[0] contains allocationChange for operation 1 and so on.
	// But we want mo.changes[0] to have allocationChange for blobber 1 and mo.changes[1] to have allocationChange for 
	// blobber 2 and so on. 
	mo.changes = zboxutil.Transpose(mo.changes)

	// var commitIdMeta map[string]string // be careful as it would create two uuid for same path
	// var rootRef *fileref.Ref
	// Get the fileref from the blobber


	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), mo.allocationObj)
	if err != nil {
		return fmt.Errorf("Operation failed: %s", err.Error())
	}

	l.Logger.Info("Trying to lock write marker.....");
	err = writeMarkerMutex.Lock(mo.ctx, &mo.operationMask, mo.maskMU,
		mo.blobbers, &mo.Consensus, 0, time.Minute, mo.connectionID)
	if err != nil {
		return fmt.Errorf("Operation failed: %s", err.Error())
	}
	l.Logger.Info("WriteMarker locked");
	defer writeMarkerMutex.Unlock(mo.ctx, mo.operationMask, mo.blobbers, time.Minute, mo.connectionID) //nolint: errcheck

	mo.Consensus.Reset()
	activeBlobbers := mo.operationMask.CountOnes()
	commitReqs := make([]*CommitRequest, activeBlobbers)

	wg.Add(activeBlobbers)
	var pos uint64 = 0
	var cntr = 0
	for i := mo.operationMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		commitReq := &CommitRequest{
			allocationID: mo.allocationID,
			allocationTx: mo.allocationTx,
			blobber:      mo.blobbers[pos],
			connectionID: mo.connectionID,
			wg:           wg,
		}
		for _, change := range mo.changes[pos] {
			commitReq.changes = append(commitReq.changes, change)
		}
		commitReqs[cntr] = commitReq;
		l.Logger.Info("Commit request sending to blobber ", commitReq.blobber.Baseurl);
		// Here this should be goroutine
		AddCommitRequest(commitReq)
		cntr++;
	}
	wg.Wait()

	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Info("Commit success", commitReq.blobber.Baseurl)
				mo.consensus++
			} else {
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !mo.isConsensusOk() {
		return thrown.New("consensus_not_met",
			fmt.Sprintf("Commit failed. Required consensus %d, got %d",
				mo.Consensus.consensusThresh, mo.Consensus.consensus))
	}

	return nil;

}
type OperationRequest struct {
	OperationType string
	LocalPath string
	RemotePath string
	DestName string  // Required only for rename operation
	DestPath string // Required for copy operation

	// Required for uploads
	Workdir string
	FileMeta FileMeta 
	FileReader io.Reader 
	Opts []ChunkedUploadOption

}

func CreateMultiOperation(allocObj *Allocation, operations []OperationRequest) (*MultiOperation, error) {
	var multiOperation MultiOperation;
	multiOperation.allocationObj = allocObj;
	multiOperation.allocationID = allocObj.ID
	multiOperation.allocationTx = allocObj.Tx;
	multiOperation.blobbers = allocObj.Blobbers
	multiOperation.operationMask = zboxutil.NewUint128(1).Lsh(uint64(len(allocObj.Blobbers))).Sub64(1)
	multiOperation.maskMU = &sync.Mutex{}
	multiOperation.ctx, multiOperation.ctxCncl = context.WithCancel(allocObj.ctx)
	multiOperation.Consensus = Consensus{
		consensusThresh: allocObj.consensusThreshold,
		fullconsensus:   allocObj.fullconsensus,
	}
	multiOperation.connectionID = zboxutil.NewConnectionId()

	for _, operation := range(operations) {
		switch operation.OperationType {
		case constants.FileOperationRename: 
			destName := operation.DestName
			remotePath := operation.RemotePath
			if destName == "" {
				return nil, thrown.New("missing parameter", "Rename operation must have destName")
			}
			if remotePath == "" {
				return nil, thrown.New("missing parameter", "Rename operation must have remotePath")
			}
			renameOp := &RenameOperation{newName: destName, remotefilepath: remotePath}
			renameOp.renameMask = multiOperation.operationMask
			renameOp.maskMU = multiOperation.maskMU
			renameOp.consensus.consensusThresh = multiOperation.consensusThresh
			renameOp.consensus.fullconsensus = multiOperation.fullconsensus
			renameOp.ctx, renameOp.ctxCncl =  context.WithCancel(allocObj.ctx)
			
			// multiOperation = append(multiOperation, renameOp);
			multiOperation.operations = append(multiOperation.operations, renameOp)

		// case constants.FileOperationCopy: 
		default: 
			return nil, thrown.New("invalid_operation", "Operation is not valid");

		}
	}
	return &multiOperation, nil;
}



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
	return objectTreeRefsEntity, nil
}

func (uo *UploadOperation) buildChange(dummyRefs []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange {
	changes := make([]allocationchange.AllocationChange, len(uo.refs))
	for idx, ref := range uo.refs {
		
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





















// type DownloadOperation struct {
// 	*DownloadRequest
// }

// func (do *DownloadOperation) Process() {
// 	// do.downloadBlock(0)
// 	// do.processDownload()
// }
// func (do *DownloadOperation) buildChange(ref *fileref.FileRef, uid uuid.UUID) allocationchange.AllocationChange {
// 	newChange := &allocationchange.DeleteFileChange{}
// 	newChange.ObjectTree = ref
// 	newChange.NumBlocks = ref.GetNumBlocks()
// 	newChange.Operation = constants.FileOperationDelete
// 	newChange.Size = ref.GetSize()
// 	return newChange
// }