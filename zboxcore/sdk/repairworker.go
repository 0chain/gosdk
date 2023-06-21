package sdk

import (
	"context"
	"sync"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
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
			dir, err = a.ListDir(dir.Path, true)
			if err != nil {
				l.Logger.Error("Failed to get listDir for path ", zap.Any("path", dir.Path), zap.Error(err))
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
		l.Logger.Info("Invalid directory type", zap.Any("type", dir.Type))
	}
}

func (r *RepairRequest) repairFile(a *Allocation, file *ListResult) {
	if r.checkForCancel(a) {
		return
	}
	l.Logger.Info("Checking file for the path :", zap.Any("path", file.Path))
	found, deleteMask, repairRequired, _, err := a.RepairRequired(file.Path)
	if err != nil {
		l.Logger.Error("repair_required_failed", zap.Error(err))
		return
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
				consensus := deleteMask.CountOnes()
				err := a.deleteFile(file.Path, 0, consensus, deleteMask)
				if err != nil {
					l.Logger.Error("delete_file_failed", zap.Error(err))
					return
				}
			}

			localPath := r.getLocalPath(file)
			toDelete := false
			if !checkFileExists(localPath) {
				if r.checkForCancel(a) {
					return
				}
				l.Logger.Info("Downloading file for the path :", zap.Any("path", file.Path))
				wg.Add(1)
				err = a.DownloadFile(localPath, file.Path, true, statusCB, true)
				if err != nil {
					l.Logger.Error("download_file_failed", zap.Error(err))
					return
				}
				wg.Wait()
				if !statusCB.success {
					l.Logger.Error("Failed to download file for repair, Status call back success failed",
						zap.Any("localpath", localPath), zap.Any("remotepath", file.Path))
					return
				}
				l.Logger.Info("Download file success for repair", zap.Any("localpath", localPath), zap.Any("remotepath", file.Path))
				statusCB.success = false
				toDelete = true
			} else {
				l.Logger.Info("FILE EXISTS", zap.Any("bool", true))
			}

			if r.checkForCancel(a) {
				return
			}

			l.Logger.Info("Repairing file for the path :", zap.Any("path", file.Path))
			wg.Add(1)
			err = a.RepairFile(localPath, file.Path, statusCB, found)
			if err != nil {
				l.Logger.Error("repair_file_failed", zap.Error(err))
				return
			}
			wg.Wait()
			if !statusCB.success {
				l.Logger.Error("Failed to repair file, Status call back success failed",
					zap.Any("localpath", localPath), zap.Any("remotepath", file.Path))
				return
			}
			if toDelete {
				_ = sys.Files.Remove(localPath)
			}
		} else {
			l.Logger.Info("Repair by delete", zap.Any("path", file.Path))
			consensus := found.CountOnes()
			err := a.deleteFile(file.Path, 1, consensus, found)
			if err != nil {
				l.Logger.Error("repair_file_failed", zap.Error(err))
				return
			}
		}
		l.Logger.Info("Repair file success", zap.Any("remotepath", file.Path))
		r.filesRepaired++
	} else if deleteMask.CountOnes() > 0 {
		l.Logger.Info("Deleting minority shards for the path :", zap.Any("path", file.Path))
		consensus := deleteMask.CountOnes()
		err := a.deleteFile(file.Path, 0, consensus, deleteMask)
		if err != nil {
			l.Logger.Error("repair_file_failed", zap.Error(err))
			return
		}
	}

}

func (r *RepairRequest) getLocalPath(file *ListResult) string {
	return r.localRootPath + file.Path
}

func checkFileExists(localPath string) bool {
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
