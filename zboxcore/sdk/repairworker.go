package sdk

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"go.uber.org/zap"
)

type RepairRequest struct {
	listDir           *ListResult
	isRepairCanceled  bool
	localImagePath    string
	localFilePath     string
	completedCallback func()
}

type RepairStatusCB struct {
	wg      *sync.WaitGroup
	success bool
	err     error
}

func (cb *RepairStatusCB) CommitMetaCompleted(request, response string, err error) {}

func (cb *RepairStatusCB) Started(allocationId, filePath string, op int, totalBytes int) {}

func (cb *RepairStatusCB) InProgress(allocationId, filePath string, op int, completedBytes int) {}

func (cb *RepairStatusCB) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	cb.success = true
	cb.wg.Done()
}

func (cb *RepairStatusCB) Error(allocationID string, filePath string, op int, err error) {
	cb.success = false
	cb.err = err
	cb.wg.Done()
}

func (r *RepairRequest) processRepair(ctx context.Context, a *Allocation) {
	a.UpdateRepairStatus(true)
	defer a.UpdateRepairStatus(false)

	if r.completedCallback != nil {
		defer r.completedCallback()
	}

	if r.isRepairCanceled {
		Logger.Info("Repair Cancelled by the user")
		return
	}
	r.iterateDir(a, r.listDir)
	return
}

func (r *RepairRequest) iterateDir(a *Allocation, dir *ListResult) {
	if dir.Type == fileref.DIRECTORY && len(dir.Children) > 0 {
		for _, childDir := range dir.Children {
			if r.isRepairCanceled {
				Logger.Info("Repair Cancelled by the user")
				return
			}
			r.iterateDir(a, childDir)
		}
	} else if dir.Type == fileref.FILE {
		r.repairFile(a, dir)
	} else {
		Logger.Error("Invalid directory type", zap.Any("type", dir.Type))
	}
	return
}

func (r *RepairRequest) repairFile(a *Allocation, file *ListResult) {
	if r.isRepairCanceled {
		Logger.Info("Repair Cancelled by the user")
		return
	}

	Logger.Info("Repairing file for the path :", zap.Any("path", file.Path))
	_, repairRequired, err := a.RepairRequired(file.Path)
	if err != nil {
		Logger.Error("repair_required_failed", zap.Error(err))
		return
	}

	if repairRequired {
		var wg sync.WaitGroup
		statusCB := &RepairStatusCB{wg: &wg}

		localPath := r.getLocalPath(file)

		if !checkFileExists(localPath) {
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
		}

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
		Logger.Info("Repair file success", zap.Any("localpath", localPath), zap.Any("remotepath", file.Path))
	}

	return
}

func (r *RepairRequest) getLocalPath(file *ListResult) string {
	if strings.Contains(file.MimeType, "image") {
		return r.localImagePath + file.Name
	}
	return r.localFilePath + file.Name
}

func checkFileExists(localPath string) bool {
	info, err := os.Stat(localPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
