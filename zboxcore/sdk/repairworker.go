package sdk

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
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
}

type RepairStatusCB struct {
	wg       *sync.WaitGroup
	success  bool
	err      error
	statusCB StatusCallback
}

func (cb *RepairStatusCB) CommitMetaCompleted(request, response string, err error) {
	cb.statusCB.CommitMetaCompleted(request, response, err)
}

func (cb *RepairStatusCB) Started(allocationId, filePath string, op int, totalBytes int) {
	cb.statusCB.Started(allocationId, filePath, op, totalBytes)
}

func (cb *RepairStatusCB) InProgress(allocationId, filePath string, op int, completedBytes int) {
	cb.statusCB.InProgress(allocationId, filePath, op, completedBytes)
}

func (cb *RepairStatusCB) RepairCompleted(filesRepaired int) {
	cb.statusCB.RepairCompleted(filesRepaired)
}

func (cb *RepairStatusCB) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	cb.statusCB.Completed(allocationId, filePath, filename, mimetype, size, op)
	if op == OpDownload || op == OpCommit {
		cb.success = true
		cb.wg.Done()
	}
}

func (cb *RepairStatusCB) Error(allocationID string, filePath string, op int, err error) {
	cb.statusCB.Error(allocationID, filePath, op, err)
	cb.success = false
	cb.err = err
	cb.wg.Done()
}

func (r *RepairRequest) processRepair(ctx context.Context, a *Allocation) {
	r.updateRepairStatusToBlobbers(a, "true")
	defer r.updateRepairStatusToBlobbers(a, "false")

	if r.completedCallback != nil {
		defer r.completedCallback()
	}

	if r.checkForCancel() {
		return
	}

	r.iterateDir(a, r.listDir)

	if r.statusCB != nil {
		r.statusCB.RepairCompleted(r.filesRepaired)
	}

	return
}

func (r *RepairRequest) iterateDir(a *Allocation, dir *ListResult) {
	switch dir.Type {
	case fileref.DIRECTORY:
		if len(dir.Children) == 0 {
			var err error
			dir, err = a.ListDir(dir.Path)
			if err != nil {
				Logger.Error("Failed to get listDir for path ", zap.Any("path", dir.Path), zap.Error(err))
				return
			}
		}
		for _, childDir := range dir.Children {
			if r.checkForCancel() {
				return
			}
			r.iterateDir(a, childDir)
		}

	case fileref.FILE:
		r.repairFile(a, dir)

	default:
		Logger.Info("Invalid directory type", zap.Any("type", dir.Type))
	}

	return
}

func (r *RepairRequest) repairFile(a *Allocation, file *ListResult) {
	if r.checkForCancel() {
		return
	}

	Logger.Info("Checking file for the path :", zap.Any("path", file.Path))
	_, repairRequired, _, err := a.RepairRequired(file.Path)
	if err != nil {
		Logger.Error("repair_required_failed", zap.Error(err))
		return
	}

	if repairRequired {
		Logger.Info("Repair required for the path :", zap.Any("path", file.Path))
		var wg sync.WaitGroup
		statusCB := &RepairStatusCB{
			wg:       &wg,
			statusCB: r.statusCB,
		}

		localPath := r.getLocalPath(file)

		if !checkFileExists(localPath) {
			if r.checkForCancel() {
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
		}

		if r.checkForCancel() {
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
		Logger.Info("Repair file success", zap.Any("localpath", localPath), zap.Any("remotepath", file.Path))
		r.filesRepaired++
	}

	return
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

func (r *RepairRequest) checkForCancel() bool {
	if r.isRepairCanceled {
		Logger.Info("Repair Cancelled by the user")
		if r.statusCB != nil {
			r.statusCB.RepairCompleted(r.filesRepaired)
		}
		return true
	}
	return false
}

func (r *RepairRequest) updateRepairStatusToBlobbers(a *Allocation, status string) {
	numList := len(a.Blobbers)
	r.wg = &sync.WaitGroup{}
	r.wg.Add(numList)
	rspCh := make(chan bool, numList)
	for i := 0; i < numList; i++ {
		go r.updateRepairStatusToBlobber(a.Blobbers[i], i, status, rspCh, a)
	}
	r.wg.Wait()
	count := 0
	for i := 0; i < numList; i++ {
		resp := <-rspCh
		if resp {
			count++
		}
	}
	if count == numList {
		Logger.Info("Repair status updated to all blobbers : ", status)
	} else {
		Logger.Error("Failed to update repair status to all blobber : ", status)
	}
	return
}

func (r *RepairRequest) updateRepairStatusToBlobber(blobber *blockchain.StorageNode, blobberIdx int,
	status string, rspCh chan<- bool, a *Allocation) {

	defer r.wg.Done()
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	formWriter.WriteField("repair", status)

	formWriter.Close()
	httpreq, err := zboxutil.NewUpdateRepairStatusRequest(blobber.Baseurl, a.Tx, body)
	if err != nil {
		Logger.Error("Update repair status request error: ", err.Error())
		return
	}

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(a.ctx, (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Update repair status : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			rspCh <- true
			return nil
		}
		rspCh <- false
		return err
	})
}
