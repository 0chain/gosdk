package sdk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"go.uber.org/zap"
)

type RepairRequest struct {
	listDir           *ListResult
	isRepairCanceled  bool
	localRootPath     string
	statusCB          StatusCallback
	completedCallback func()
	filesRepaired     int
	wg                *sync.WaitGroup
	allocation        *Allocation
	repairPath        string
}

type RepairStatusCB struct {
	wg       *sync.WaitGroup
	success  bool
	err      error
	statusCB StatusCallback
}

var RepairBlocks = 100

func (cb *RepairStatusCB) Started(allocationId, filePath string, op int, totalBytes int) {
	cb.statusCB.Started(allocationId, filePath, op, totalBytes)
}

func (cb *RepairStatusCB) InProgress(allocationId, filePath string, op int, completedBytes int, data []byte) {
	cb.statusCB.InProgress(allocationId, filePath, op, completedBytes, data)
}

func (cb *RepairStatusCB) RepairCompleted(filesRepaired int) {
	cb.statusCB.RepairCompleted(filesRepaired)
}

func (cb *RepairStatusCB) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	cb.statusCB.Completed(allocationId, filePath, filename, mimetype, size, op)
	cb.success = true
	cb.wg.Done()
}

func (cb *RepairStatusCB) Error(allocationID string, filePath string, op int, err error) {
	cb.statusCB.Error(allocationID, filePath, op, err)
	cb.success = false
	cb.err = err
	cb.wg.Done()
}

func (r *RepairRequest) processRepair(ctx context.Context, a *Allocation) {
	if r.completedCallback != nil {
		defer r.completedCallback()
	}

	if r.checkForCancel(a) {
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	SetNumBlockDownloads(RepairBlocks)
	currBatchSize := BatchSize
	BatchSize = BatchSize / 2
	defer func() {
		BatchSize = currBatchSize
	}()
	if !singleClientMode {
		SetSingleClietnMode(true)
		defer SetSingleClietnMode(false)
	}
	r.allocation = a
	if a.StorageVersion == StorageV2 {
		r.iterateDirV2(ctx)
	} else {
		r.iterateDir(a, r.listDir)
	}
	if r.statusCB != nil {
		r.statusCB.RepairCompleted(r.filesRepaired)
	}
}

// holds result of repair size
type RepairSize struct {
	// upload size in bytes
	UploadSize uint64 `json:"upload_size"`
	// download size in bytes
	DownloadSize uint64 `json:"download_size"`
}

// gets size to repair for remote dir.
func (r *RepairRequest) Size(ctx context.Context, dir *ListResult) (RepairSize, error) {
	var rs RepairSize
	var err error
	switch dir.Type {
	case fileref.DIRECTORY:
		if len(dir.Children) == 0 {
			// fetch dir
			dir, err = r.allocation.ListDir(dir.Path, WithListRequestForRepair(true), WithListRequestPageLimit(-1))
			if err != nil {
				return rs, err
			}
		}
		for _, subDir := range dir.Children {
			subDirSz, err := r.Size(ctx, subDir)
			if err != nil {
				return rs, err
			}
			rs.UploadSize += subDirSz.UploadSize
			rs.DownloadSize += subDirSz.DownloadSize
		}
	case fileref.FILE:
		// this returns repair operations required
		repairOps := r.repairFile(r.allocation, dir)
		if repairOps == nil {
			err = fmt.Errorf("fetch repairOps failed")
			return rs, err
		}
		for _, repairOp := range repairOps {
			if repairOp.OperationType == constants.FileOperationInsert {
				rs.UploadSize += uint64(repairOp.Mask.CountOnes()) * uint64(getShardSize(repairOp.FileMeta.ActualSize, r.allocation.DataShards, repairOp.EncryptedKey != ""))
				rs.DownloadSize += uint64(repairOp.FileMeta.ActualSize)
			}
		}
	}
	return rs, err
}

func (r *RepairRequest) iterateDir(a *Allocation, dir *ListResult) []OperationRequest {
	ops := make([]OperationRequest, 0)
	switch dir.Type {
	case fileref.DIRECTORY:
		if len(dir.Children) == 0 {
			var err error
			dir, err = a.ListDir(dir.Path, WithListRequestForRepair(true), WithListRequestPageLimit(-1))
			if err != nil {
				l.Logger.Error("Failed to get listDir for path ", zap.Any("path", dir.Path), zap.Error(err))
				return nil
			}
		}
		if len(dir.Children) == 0 {
			if dir.deleteMask.CountOnes() > 0 {
				l.Logger.Info("Deleting minority shards for the path :", zap.Any("path", dir.Path))
				consensus := dir.deleteMask.CountOnes()
				if consensus < a.DataShards {

					err := a.deleteFile(dir.Path, 0, consensus, dir.deleteMask)
					if err != nil {
						l.Logger.Error("repair_file_failed", zap.Error(err))
						if r.statusCB != nil {
							r.statusCB.Error(a.ID, dir.Path, OpRepair, err)
						}
						return nil
					}
					r.filesRepaired++
				} else if consensus < len(a.Blobbers) {
					createMask := dir.deleteMask.Not().And(zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1))
					err := a.createDir(dir.Path, 0, createMask.CountOnes(), createMask)
					if err != nil {
						l.Logger.Error("repair_file_failed", zap.Error(err))
						if r.statusCB != nil {
							r.statusCB.Error(a.ID, dir.Path, OpRepair, err)
						}
						return nil
					}
					r.filesRepaired++
				}
			}
		}
		for _, childDir := range dir.Children {
			if r.checkForCancel(a) {
				return nil
			}
			ops = append(ops, r.iterateDir(a, childDir)...)
			if len(ops) >= RepairBatchSize/2 {
				r.repairOperation(a, ops)
				ops = nil
			}
		}
		if len(ops) > 0 {
			r.repairOperation(a, ops)
			ops = nil
		}
	case fileref.FILE:
		// this returns op object and mask
		repairOps := r.repairFile(a, dir)
		if repairOps != nil {
			ops = append(ops, repairOps...)
		}

	default:
		l.Logger.Info("Invalid directory type", zap.Any("type", dir.Type))
	}
	return ops
}

func (r *RepairRequest) repairFile(a *Allocation, file *ListResult) []OperationRequest {
	ops := make([]OperationRequest, 0)
	if r.checkForCancel(a) {
		return nil
	}
	l.Logger.Info("Checking file for the path :", zap.Any("path", file.Path))
	found, deleteMask, repairRequired, ref, err := a.RepairRequired(file.Path)
	if err != nil {
		l.Logger.Error("repair_required_failed", zap.Error(err))
		return nil
	}
	if repairRequired {
		l.Logger.Info("Repair required for the path :", zap.Any("path", file.Path))
		if found.CountOnes() >= a.DataShards {
			l.Logger.Info("Repair by upload", zap.Any("path", file.Path))
			var wg sync.WaitGroup
			statusCB := &RepairStatusCB{
				wg:       &wg,
				statusCB: r.statusCB,
			}

			if deleteMask.CountOnes() > 0 {
				l.Logger.Info("Deleting minority shards for the path :", zap.Any("path", file.Path))
				op := OperationRequest{
					OperationType: constants.FileOperationDelete,
					RemotePath:    file.Path,
					Mask:          &deleteMask,
				}
				ops = append(ops, op)
			}
			wg.Add(1)
			localPath := r.getLocalPath(file)
			var op *OperationRequest
			if !checkFileExists(localPath) {
				if r.checkForCancel(a) {
					return nil
				}
				pipeFile := sys.NewPipeFile()
				op = a.RepairFile(pipeFile, file.Path, statusCB, found, ref)
				if op.FileMeta.ActualSize > 0 {
					op.DownloadFile = true
				}
			} else {
				f, err := sys.Files.Open(localPath)
				if err != nil {
					l.Logger.Error("repair_file_failed", zap.Error(err))
					return nil
				}
				op = a.RepairFile(f, file.Path, statusCB, found, ref)
			}
			ops = append(ops, *op)
			if r.checkForCancel(a) {
				return nil
			}
		} else {
			l.Logger.Info("Repair by delete", zap.Any("path", file.Path))
			op := OperationRequest{
				OperationType: constants.FileOperationDelete,
				RemotePath:    file.Path,
				Mask:          &found,
			}
			ops = append(ops, op)
		}
	} else if deleteMask.CountOnes() > 0 {
		l.Logger.Info("Deleting minority shards for the path :", zap.Any("path", file.Path))
		op := OperationRequest{
			OperationType: constants.FileOperationDelete,
			RemotePath:    file.Path,
			Mask:          &deleteMask,
		}
		ops = append(ops, op)
	}
	return ops
}

func (r *RepairRequest) repairOperation(a *Allocation, ops []OperationRequest) {
	err := a.DoMultiOperation(ops, WithRepair())
	if err != nil {
		l.Logger.Error("repair_file_failed", zap.Error(err))
		status := r.statusCB != nil
		for _, op := range ops {
			if op.DownloadFile {
				_ = a.CancelDownload(op.RemotePath)
			}
			if status {
				r.statusCB.Error(a.ID, op.RemotePath, OpRepair, err)
			}
		}
	} else {
		r.filesRepaired += len(ops)
	}
	for _, op := range ops {
		if op.FileReader != nil && !op.DownloadFile {
			if f, ok := op.FileReader.(io.Closer); ok {
				f.Close()
			}
		}
	}
}

func (r *RepairRequest) getLocalPath(file *ListResult) string {
	return r.localRootPath + file.Path
}

func checkFileExists(localPath string) bool {
	if IsWasm {
		return false
	}
	info, err := sys.Files.Stat(localPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (r *RepairRequest) checkForCancel(a *Allocation) bool {
	return r.isRepairCanceled
}

type diffRef struct {
	tgtRef  ORef
	tgtChan <-chan ORef
	tgtEOF  bool
	mask    zboxutil.Uint128
}

func (r *RepairRequest) iterateDirV2(ctx context.Context) {
	versionMap := make(map[string]*diffRef)
	r.allocation.CheckAllocStatus() //nolint:errcheck
	latestRoot := r.allocation.allocationRoot
	for idx, blobber := range r.allocation.Blobbers {
		if versionMap[blobber.AllocationRoot] == nil {
			versionMap[blobber.AllocationRoot] = &diffRef{}
		}
		versionMap[blobber.AllocationRoot].mask = versionMap[blobber.AllocationRoot].mask.Or(zboxutil.NewUint128(1).Lsh(uint64(idx)))
	}
	if versionMap[latestRoot].mask.CountOnes() < r.allocation.DataShards {
		l.Logger.Error("No consensus on latest allocation root: ", latestRoot)
		if r.statusCB != nil {
			r.statusCB.Error(r.allocation.ID, r.repairPath, OpRepair, errors.New("no consensus on latest allocation root"))
		}
		return
	}
	if len(versionMap) == 1 {
		return
	}
	// get the src list channel
	srcChan := r.allocation.ListObjects(ctx, r.repairPath, "", "", "", fileref.FILE, fileref.REGULAR, 0, getRefPageLimit, WithSingleBlobber(true), WithObjectMask(versionMap[latestRoot].mask), WithObjectContext(ctx))

	for root, diff := range versionMap {
		if root == latestRoot {
			continue
		}
		diff.tgtChan = r.allocation.ListObjects(ctx, r.repairPath, "", "", "", fileref.FILE, fileref.REGULAR, 0, getRefPageLimit, WithSingleBlobber(true), WithObjectMask(diff.mask), WithObjectContext(ctx))
		diff.tgtRef, diff.tgtEOF = <-diff.tgtChan
	}
	var (
		toNextRef = true
		srcRef    ORef
		srcEOF    = true
		ops       []OperationRequest
	)
	for {
		if r.checkForCancel(r.allocation) {
			return
		}
		if toNextRef {
			if !srcEOF {
				break
			}
			srcRef, srcEOF = <-srcChan
			if srcRef.Err != nil {
				l.Logger.Error("Failed to get source file reference ", srcRef.Err.Error())
				if r.statusCB != nil {
					r.statusCB.Error(r.allocation.ID, r.repairPath, OpRepair, srcRef.Err)
				}
				return
			}
		}
		l.Logger.Debug("Checking file for the path :", srcRef.Path)
		toNextRef = true
		var (
			uploadMask zboxutil.Uint128
			deleteMask zboxutil.Uint128
		)
		for root, diff := range versionMap {
			if root == latestRoot {
				continue
			}

			// check if both target and source are at EOF
			if !srcEOF && !diff.tgtEOF {
				continue
			}
			// if target is at EOF, upload the src file
			if !diff.tgtEOF {
				uploadMask = uploadMask.Or(diff.mask)
				continue
			}
			// if source is at EOF, delete the target file
			if !srcEOF {
				delMask := diff.mask
				op := OperationRequest{
					OperationType: constants.FileOperationDelete,
					RemotePath:    diff.tgtRef.Path,
					Mask:          &delMask,
				}
				ops = append(ops, op)
				diff.tgtRef, diff.tgtEOF = <-diff.tgtChan
				toNextRef = false
				continue
			}
			if diff.tgtRef.Err != nil {
				l.Logger.Error("Failed to get target file reference ", diff.tgtRef.Err.Error())
				if r.statusCB != nil {
					r.statusCB.Error(r.allocation.ID, r.repairPath, OpRepair, diff.tgtRef.Err)
				}
				continue
			}
			// if both source and target are at same path
			if diff.tgtRef.Path == srcRef.Path {
				// if both source and target are at same path and hash is different
				if diff.tgtRef.ActualFileHash != srcRef.ActualFileHash {
					deleteMask = deleteMask.Or(diff.mask)
					uploadMask = uploadMask.Or(diff.mask)
				}
				diff.tgtRef, diff.tgtEOF = <-diff.tgtChan
			} else if diff.tgtRef.Path < srcRef.Path {
				deleteMask = deleteMask.Or(diff.mask)
				toNextRef = false
				diff.tgtRef, diff.tgtEOF = <-diff.tgtChan
			}
		}
		if deleteMask.CountOnes() > 0 {
			op := OperationRequest{
				OperationType: constants.FileOperationDelete,
				RemotePath:    srcRef.Path,
				Mask:          &deleteMask,
			}
			ops = append(ops, op)
		}
		if uploadMask.CountOnes() > 0 {
			op := r.uploadFileOp(srcRef, uploadMask)
			ops = append(ops, op)
		}
		if len(ops) >= RepairBatchSize {
			r.repairOperation(r.allocation, ops)
			ops = nil
		}
	}
	if len(ops) > 0 {
		r.repairOperation(r.allocation, ops)
		ops = nil
	}

}

func (r *RepairRequest) uploadFileOp(file ORef, opMask zboxutil.Uint128) OperationRequest {
	var wg sync.WaitGroup
	wg.Add(1)
	statusCB := &RepairStatusCB{
		wg:       &wg,
		statusCB: r.statusCB,
	}

	ref := &fileref.FileRef{
		ActualFileSize: file.ActualFileSize,
		MimeType:       file.MimeType,
		CustomMeta:     file.CustomMeta,
		Ref: fileref.Ref{
			Name: file.Name,
		},
		EncryptedKey:      file.EncryptedKey,
		EncryptedKeyPoint: file.EncryptedKeyPoint,
	}
	pipeFile := sys.NewPipeFile()
	op := r.allocation.RepairFile(pipeFile, file.Path, statusCB, opMask, ref)
	if op.FileMeta.ActualSize > 0 {
		op.DownloadFile = true
	}
	return *op
}
