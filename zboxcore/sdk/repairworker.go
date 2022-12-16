package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
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
	fileID            int64
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

type FileOp int

const (
	WMUpload FileOp = iota
	WMUpdate
	WMRename
	WMCopy
	WMMove
	WMDelete
	WMNewDir
)

type WriteMarker struct {
	ClientID      string `json:"client_id"`
	BlobberID     string `json:"blobber_id"`
	AllocationID  string `json:"allocation_id"`
	TransactionID string `json:"transaction_id"`

	AllocationRoot         string `json:"allocation_root"`
	PreviousAllocationRoot string `json:"previous_allocation_root"`
	FileMetaRoot           string `json:"file_meta_root"`
	Size                   int64  `json:"size"`
	Timestamp              int64  `json:"timestamp"`
	Signature              string `json:"signature"`
	BlockNumber            int64  `json:"block_number"` //used in alloc_written_size

	FileID    int64  `json:"file_id"`
	Operation FileOp `json:"operation"`
}

// return a fileID for a given remotePath
// func getFileIDFromPath(remotePath string) (int, error) {

// }

func GetWriteMarkers(allocationID string, FileID string) []*WriteMarker {
	fmt.Println("getting write makers for FileID", FileID)
	var b []byte
	// todo: paginate
	b, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/getWriteMarkers",
		map[string]string{"allocation_id": allocationID, "file_id": FileID}, nil)
	fmt.Println("sent request")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	wms := make([]*WriteMarker, 0)
	err = json.Unmarshal(b, &wms)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println("wms return")
	return wms
}

func (r *RepairRequest) processRepair(ctx context.Context, a *Allocation) {
	if r.completedCallback != nil {
		defer r.completedCallback()
	}

	if r.checkForCancel(a) {
		return
	}

	_, err := r.FindRepairsNeeded(a)
	if err != nil {
		r.statusCB.Error(a.ID, r.localRootPath, 0, err)
	}

	// for _, repair := range repairs {
	// 	r.repairFile(a, r.localRootPath)
	// }

	// r.iterateDir(a, r.listDir)

	if r.statusCB != nil {
		r.statusCB.RepairCompleted(r.filesRepaired)
	}
}

func (r *RepairRequest) iterateDir(a *Allocation, dir *ListResult) {
	switch dir.Type {
	case fileref.DIRECTORY:
		if len(dir.Children) == 0 {
			var err error
			dir, err = a.ListDir(dir.Path)
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

type repair struct {
	op   FileOp
	mask bool
}

// getMajorityWriteMarker returns writeMarker which is majorly applied among blobbers
func getMajorityWriteMarker(fileMetaRootMap map[string][]*WriteMarker) (majorityWm *WriteMarker) {
	count := 0
	for _, wms := range fileMetaRootMap {
		if len(wms) > count {
			count = len(wms)
			majorityWm = wms[0]
		}
	}
	return
}

// getBlobbersNeedingRepair return the boolean mask associated with blobbers in allocations
func getBlobbersNeedingRepair(a *Allocation, majorityWm *WriteMarker, allocBlobbersMap map[string][]*WriteMarker, depth int) (zboxutil.Uint128, bool) {
	found := zboxutil.NewUint128(0)
	repairMask := zboxutil.NewUint128(0)
	for i, blobber := range a.Blobbers {
		wms := allocBlobbersMap[blobber.ID]
		if depth < len(wms) {
			wm := wms[depth]
			if majorityWm.FileMetaRoot == wm.FileMetaRoot {
				shift := zboxutil.NewUint128(1).Lsh(uint64(i))
				found = found.Or(shift)
				continue
			}
		}
		logger.Logger.Debug("blobber needing repair", blobber.ID)
		shift := zboxutil.NewUint128(1).Lsh(uint64(i))
		repairMask = repairMask.Or(shift)
	}

	mask := zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
	return repairMask, !found.Equals(mask)
}

func (r *RepairRequest) FindRepairsNeeded(a *Allocation) ([]repair, error) {

	wms := GetWriteMarkers(a.ID, strconv.FormatInt(r.fileID, 10))

	// for repair, we only consider the blobbers which are part of the current allocation. So, remove writemarkers associated with non-existent blobbers
	allocBlobbersMap := map[string][]*WriteMarker{}

	for _, allocBlobber := range a.Blobbers {
		allocBlobbersMap[allocBlobber.ID] = []*WriteMarker{}
	}

	for _, wm := range wms {
		if _, ok := allocBlobbersMap[wm.BlobberID]; ok {
			allocBlobbersMap[wm.BlobberID] = append(allocBlobbersMap[wm.BlobberID], wm)
		}
	}

	maxLen := 0
	// sort wms based on timestamp
	for blobberID, wm := range allocBlobbersMap {
		sort.Slice(wm, func(i, j int) bool {
			return wm[i].Timestamp < wm[j].Timestamp
		})
		if len(wm) > maxLen {
			maxLen = len(wm)
		}
		allocBlobbersMap[blobberID] = wm
	}

	// todo: validate the chaining of writemarker with PreviousAllocationRoot.

	for blobberID, wms := range allocBlobbersMap {
		fmt.Println("For BlobberID", blobberID)
		for _, wm := range wms {
			fmt.Println(wm.BlobberID, wm.FileMetaRoot, wm.Operation, wm.Timestamp)
		}
	}

	// Now that the write markers are sorted by timestamps, lets check if the FileMetaRoots
	// are in equal. At any place, If they are not equal, find the majority writeMarker and apply
	// that write marker on no complying blobbers

	// at each depth there is going to be one file operation
	for depth := 0; depth < maxLen; depth++ {

		logger.Logger.Debug("depth", depth)
		fileMetaRootMap := map[string][]*WriteMarker{} // used to identify the majority writemarker
		for _, blobber := range a.Blobbers {
			wms := allocBlobbersMap[blobber.ID]
			fmt.Println("wms", wms)
			if depth < maxLen && depth < len(wms) {
				wm := wms[depth]
				fileMetaRootMap[wm.FileMetaRoot] = append(fileMetaRootMap[wm.FileMetaRoot], wm)
			}
		}

		// get Majority Write Marker
		majorityWM := getMajorityWriteMarker(fileMetaRootMap)

		// get the blobbers in the allocation that needed repair for the above write marker operation
		// the blobbers in allocation who doesn't have the FileMetaRoot of the Majority WriteMarker needs repair
		repairMask, repairRequired := getBlobbersNeedingRepair(a, majorityWM, allocBlobbersMap, depth)

		if !repairRequired {
			logger.Logger.Debug("repair not required at this depth")
			continue
		}

		if majorityWM.Operation == WMDelete {
			path := zboxutil.RemoteClean("/14.png")
			req := &DeleteRequest{}
			req.allocationObj = a
			req.blobbers = a.Blobbers
			req.allocationID = a.ID
			req.allocationTx = a.Tx
			req.consensus.mu = &sync.RWMutex{}
			req.consensus.Init(1, repairMask.CountOnes())
			req.ctx = a.ctx
			req.remotefilepath = path
			req.connectionID = zboxutil.NewConnectionId()
			// use full allocation mask
			req.deleteMask = zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
			req.maskMu = &sync.Mutex{}
			req.ProcessDelete()
			break
		}
	}
	return []repair{}, nil
}

func (r *RepairRequest) repairFile(a *Allocation, file *ListResult) {
	if r.checkForCancel(a) {
		return
	}
	l.Logger.Info("Checking file for the path :", zap.Any("path", file.Path))
	found, repairRequired, _, err := a.RepairRequired(file.Path)
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

			localPath := r.getLocalPath(file)

			if !checkFileExists(localPath) {
				if r.checkForCancel(a) {
					return
				}
				l.Logger.Info("Downloading file for the path :", zap.Any("path", file.Path))
				wg.Add(1)
				err = a.DownloadFile(localPath, file.Path, statusCB)
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
			} else {
				l.Logger.Info("FILE EXISTS", zap.Any("bool", true))
			}

			if r.checkForCancel(a) {
				return
			}

			l.Logger.Info("Repairing file for the path :", zap.Any("path", file.Path))
			wg.Add(1)
			err = a.RepairFile(localPath, file.Path, statusCB)
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
		} else {
			l.Logger.Info("Repair by delete", zap.Any("path", file.Path))
			consensus := found.CountOnes()
			err := a.deleteFile(file.Path, consensus, consensus)
			if err != nil {
				l.Logger.Error("repair_file_failed", zap.Error(err))
				return
			}
		}
		l.Logger.Info("Repair file success", zap.Any("remotepath", file.Path))
		r.filesRepaired++
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
