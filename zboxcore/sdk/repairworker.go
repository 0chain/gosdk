package sdk

import (
	"context"
	"os"
	"sync"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
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
}

type RepairStatusCB struct {
	wg       *sync.WaitGroup
	success  bool
	err      error
	statusCB StatusCallback
}

func (cb *RepairStatusCB) CommitMetaCompleted(request, response string, txn *transaction.Transaction, err error) {
	cb.statusCB.CommitMetaCompleted(request, response, txn, err)
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

	r.iterateDir(a, r.listDir)

	if r.statusCB != nil {
		r.statusCB.RepairCompleted(r.filesRepaired)
	}
}

func (r *RepairRequest) iterateDir(a *Allocation, dir *ListResult) {
	switch dir.Type {
	case fileref.DIRECTORY:
		if len(dir.Children) == 0 {
			var err error
			fullconsensus := float32(a.DataShards + a.ParityShards)
			consensusThresh := 100 / fullconsensus
			dir, err = a.listDir(dir.Path, consensusThresh, fullconsensus)
			if err != nil {
				Logger.Error("Failed to get listDir for path ", zap.Any("path", dir.Path), zap.Error(err))
				return
			}
		}
		for _, childDir := range dir.Children {
			if r.checkForCancel(a) {
				return
			}
			r.iterateDir(a, childDir)
		}

	case fileref.FILE:
		r.repairFile(a, dir)

	default:
		Logger.Info("Invalid directory type", zap.Any("type", dir.Type))
	}
}

func (r *RepairRequest) repairFile(a *Allocation, file *ListResult) {
	if r.checkForCancel(a) {
		return
	}
	Logger.Info("Checking file for the path :", zap.Any("path", file.Path))
	found, repairRequired, _, err := a.RepairRequired(file.Path)
	if err != nil {
		Logger.Error("repair_required_failed", zap.Error(err))
		return
	}

	if repairRequired {
		Logger.Info("Repair required for the path :", zap.Any("path", file.Path))
		if found.CountOnes() >= a.DataShards {
			Logger.Info("Repair by upload", zap.Any("path", file.Path))
			var wg sync.WaitGroup
			statusCB := &RepairStatusCB{
				wg:       &wg,
				statusCB: r.statusCB,
			}

			localPath := r.getLocalPath(file)

			if !checkFileExists(localPath) {
				if r.checkForCancel(a) {
					return
				}
				Logger.Info("Downloading file for the path :", zap.Any("path", file.Path))
				wg.Add(1)
				err = a.DownloadFile(localPath, file.Path, statusCB)
				if err != nil {
					Logger.Error("download_file_failed", zap.Error(err))
					return
				}
				wg.Wait()
				if !statusCB.success {
					Logger.Error("Failed to download file for repair, Status call back success failed",
						zap.Any("localpath", localPath), zap.Any("remotepath", file.Path))
					return
				}
				Logger.Info("Download file success for repair", zap.Any("localpath", localPath), zap.Any("remotepath", file.Path))
				statusCB.success = false
			} else {
				Logger.Info("FILE EXISTS", zap.Any("bool", true))
			}

			if r.checkForCancel(a) {
				return
			}

			Logger.Info("Repairing file for the path :", zap.Any("path", file.Path))
			wg.Add(1)
			err = a.RepairFile(localPath, file.Path, statusCB)
			if err != nil {
				Logger.Error("repair_file_failed", zap.Error(err))
				return
			}
			wg.Wait()
			if !statusCB.success {
				Logger.Error("Failed to repair file, Status call back success failed",
					zap.Any("localpath", localPath), zap.Any("remotepath", file.Path))
				return
			}
		} else {
			Logger.Info("Repair by delete", zap.Any("path", file.Path))
			consensus := float32(found.CountOnes())
			err := a.deleteFile(file.Path, consensus, consensus)
			if err != nil {
				Logger.Error("repair_file_failed", zap.Error(err))
				return
			}
		}
		Logger.Info("Repair file success", zap.Any("remotepath", file.Path))
		r.filesRepaired++
	}

}

func (r *RepairRequest) getLocalPath(file *ListResult) string {
	return r.localRootPath + file.Path
}

func checkFileExists(localPath string) bool {
	info, err := os.Stat(localPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (r *RepairRequest) checkForCancel(a *Allocation) bool {
	if r.isRepairCanceled {
		Logger.Info("Repair Cancelled by the user")
		if r.statusCB != nil {
			r.statusCB.RepairCompleted(r.filesRepaired)
		}
		return true
	}
	return false
}
