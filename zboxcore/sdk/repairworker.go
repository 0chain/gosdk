package sdk

import (
	"context"
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
}

type RepairStatusCB struct {
	wg       *sync.WaitGroup
	success  bool
	err      error
	statusCB StatusCallback
}

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
	SetNumBlockDownloads(100)
	r.iterateDir(a, r.listDir)

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
				memFile := &sys.MemChanFile{
					Buffer:         make(chan []byte, 100),
					ChunkWriteSize: int(a.GetChunkReadSize(ref.EncryptedKey != "")),
				}
				op = a.RepairFile(memFile, file.Path, statusCB, found, ref)
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
	if r.isRepairCanceled {
		l.Logger.Info("Repair Cancelled by the user")
		if r.statusCB != nil {
			r.statusCB.RepairCompleted(r.filesRepaired)
		}
		return true
	}
	return false
}
