package sdk

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

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
	versionMap        map[int64]zboxutil.Uint128
	resMap            map[int64]*getRes
}

type RepairStatusCB struct {
	wg       *sync.WaitGroup
	success  bool
	err      error
	statusCB StatusCallback
}

type getRes struct {
	oTR             *ObjectTreeResult
	err             error
	listCompleted   bool
	idx             int
	lastMatchedPath string
	lastIndex       int
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
	currentSize := MultiOpBatchSize
	SetMultiOpBatchSize(multiOpRepairBatchSize)
	defer SetMultiOpBatchSize(currentSize)
	r.allocation = a
	r.iterateDir(ctx)

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
	// switch dir.Type {
	// case fileref.DIRECTORY:
	// 	if len(dir.Children) == 0 {
	// 		// fetch dir
	// 		dir, err = r.allocation.ListDir(dir.Path, WithListRequestForRepair(true), WithListRequestPageLimit(-1))
	// 		if err != nil {
	// 			return rs, err
	// 		}
	// 	}
	// 	for _, subDir := range dir.Children {
	// 		subDirSz, err := r.Size(ctx, subDir)
	// 		if err != nil {
	// 			return rs, err
	// 		}
	// 		rs.UploadSize += subDirSz.UploadSize
	// 		rs.DownloadSize += subDirSz.DownloadSize
	// 	}
	// case fileref.FILE:
	// 	// this returns repair operations required
	// 	repairOps := r.repairFile(r.allocation, dir)
	// 	if repairOps == nil {
	// 		err = fmt.Errorf("fetch repairOps failed")
	// 		return rs, err
	// 	}
	// 	for _, repairOp := range repairOps {
	// 		if repairOp.OperationType == constants.FileOperationInsert {
	// 			rs.UploadSize += uint64(repairOp.Mask.CountOnes()) * uint64(getShardSize(repairOp.FileMeta.ActualSize, r.allocation.DataShards, repairOp.EncryptedKey != ""))
	// 			rs.DownloadSize += uint64(repairOp.FileMeta.ActualSize)
	// 		}
	// 	}
	// }
	return rs, err
}

func (r *RepairRequest) iterateDir(ctx context.Context) {
	r.versionMap = make(map[int64]zboxutil.Uint128)
	latestVersion := r.allocation.allocationVersion
	for idx, blobber := range r.allocation.Blobbers {
		r.versionMap[blobber.AllocationVersion] = r.versionMap[blobber.AllocationVersion].Or(zboxutil.NewUint128(1).Lsh(uint64(idx)))
	}
	if r.versionMap[latestVersion].CountOnes() < r.allocation.DataShards {
		l.Logger.Error("No consensus on latest allocation version: ", latestVersion)
		if r.statusCB != nil {
			r.statusCB.Error(r.allocation.ID, "/", OpRepair, errors.New("no consensus on latest allocation version"))
		}
		return
	}
	if len(r.versionMap) == 1 {
		return
	}

	r.resMap = make(map[int64]*getRes)
	err := r.repairObjects(ctx, latestVersion, fileref.FILE)
	if err != nil {
		return
	}
	l.Logger.Debug("repair file completed")
	err = r.repairObjects(ctx, latestVersion, fileref.DIRECTORY)
	if err != nil {
		return
	}
	l.Logger.Debug("repair directory completed")

	r.allocation.CheckAllocStatus() //nolint:errcheck
	if r.allocation.allocationVersion != latestVersion {
		l.Logger.Error("Allocation version changed during repair operation")
		if r.statusCB != nil {
			r.statusCB.Error(r.allocation.ID, r.repairPath, OpRepair, errors.New("allocation version changed during repair operation"))
		}
		return
	}
	wg := &sync.WaitGroup{}
	commiReqs := make([]*CommitRequest, 0)
	//send final commit to remaining blobbers
	for version, mask := range r.versionMap {
		if version == latestVersion {
			continue
		}
		currMask := mask
		var pos uint64
		for i := currMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			wg.Add(1)
			pos = uint64(i.TrailingZeros())
			commitReq := &CommitRequest{
				allocationID: r.allocation.ID,
				allocationTx: r.allocation.Tx,
				blobber:      r.allocation.Blobbers[pos],
				connectionID: zboxutil.NewConnectionId(),
				wg:           wg,
				timestamp:    time.Now().Unix(),
				blobberInd:   pos,
				version:      latestVersion,
			}
			go AddCommitRequest(commitReq)
			commiReqs = append(commiReqs, commitReq)
		}
	}
	wg.Wait()
	for _, commitReq := range commiReqs {
		if commitReq.result != nil && !commitReq.result.Success {
			l.Logger.Error("Failed to commit repair operation: ", commitReq.result.ErrorMessage)
			if r.statusCB != nil {
				r.statusCB.Error(r.allocation.ID, r.repairPath, OpRepair, errors.New(commitReq.result.ErrorMessage))
			}
			return
		}
	}
}

func (r *RepairRequest) uploadFileOp(file ORef, opMask zboxutil.Uint128) OperationRequest {
	var wg sync.WaitGroup
	wg.Add(1)
	statusCB := &RepairStatusCB{
		wg:       &wg,
		statusCB: r.statusCB,
	}
	memFile := &sys.MemChanFile{
		Buffer:         make(chan []byte, 100),
		ChunkWriteSize: int(r.allocation.GetChunkReadSize(file.EncryptedKey != "")),
	}
	op := r.allocation.RepairFile(memFile, file.Path, statusCB, opMask, file)
	if op.FileMeta.ActualSize > 0 {
		op.DownloadFile = true
	}
	return *op
}

func (r *RepairRequest) repairObjects(ctx context.Context, latestVersion int64, fileType string) error {
	clear(r.resMap)
	for {
		r.getRefsWithVersion(ctx, fileType)
		if r.resMap[latestVersion].err != nil {
			l.Logger.Error("Failed to get refs for repair path: ", r.repairPath, r.resMap[latestVersion].err)
			if r.statusCB != nil {
				r.statusCB.Error(r.allocation.ID, r.repairPath, OpRepair, r.resMap[latestVersion].err)
			}
			return r.resMap[latestVersion].err
		}
		for version, res := range r.resMap {
			if res.err != nil {
				l.Logger.Error("Failed to get refs for repair path: ", r.repairPath, res.err)
				delete(r.versionMap, version)
				delete(r.resMap, version)
			}
		}
		if len(r.versionMap) == 1 {
			l.Logger.Error("all other versions failed to get refs for repair path: ", r.repairPath)
			err := errors.New("all other versions failed to get refs for repair path")
			if r.statusCB != nil {
				r.statusCB.Error(r.allocation.ID, r.repairPath, OpRepair, err)
			}
			return err
		}
		ops := make([]OperationRequest, 0)
		for idx, file := range r.resMap[latestVersion].oTR.Refs {
			if r.checkForCancel(r.allocation) {
				return errors.New("repair cancelled")
			}

			for version, res := range r.resMap {
				if version == latestVersion {
					continue
				}

				if file.Path <= res.lastMatchedPath {
					continue
				}

				if res.idx == len(res.oTR.Refs) {
					if res.listCompleted {
						// upload the file
						opMask := r.versionMap[version]
						if fileType == fileref.FILE {
							op := r.uploadFileOp(file, opMask)
							ops = append(ops, op)
						} else {
							op := OperationRequest{
								OperationType: constants.FileOperationCreateDir,
								RemotePath:    file.Path,
							}
							ops = append(ops, op)
						}
						res.lastMatchedPath = file.Path
						res.lastIndex = idx
					}
					continue
				}

				currentPath := res.oTR.Refs[res.idx].Path
				if currentPath == file.Path {
					if file.Type == fileref.FILE && file.ActualFileHash != res.oTR.Refs[res.idx].ActualFileHash {
						// upload the file
						opMask := r.versionMap[version]
						op := r.uploadFileOp(file, opMask)
						ops = append(ops, op)
					}
					res.lastMatchedPath = currentPath
					res.idx++
					res.lastIndex = idx
				} else if currentPath < file.Path {
					for res.idx < len(res.oTR.Refs) && res.oTR.Refs[res.idx].Path < file.Path {
						//delete the file
						opMask := r.versionMap[version]
						op := OperationRequest{
							OperationType: constants.FileOperationDelete,
							RemotePath:    res.oTR.Refs[res.idx].Path,
							Mask:          &opMask,
						}
						res.idx++
						ops = append(ops, op)
					}
					if res.idx < len(res.oTR.Refs) && res.oTR.Refs[res.idx].Path > file.Path {
						opMask := r.versionMap[version]
						// upload the file
						if fileType == fileref.FILE {
							op := r.uploadFileOp(file, opMask)
							ops = append(ops, op)
						} else {
							op := OperationRequest{
								OperationType: constants.FileOperationCreateDir,
								RemotePath:    file.Path,
							}
							ops = append(ops, op)
						}
						res.lastMatchedPath = file.Path
						res.idx++
						res.lastIndex = idx
					}
				} else {
					//TODO: take a union of mask so we don't have duplicate upload operations
					// upload the file
					opMask := r.versionMap[version]
					if fileType == fileref.FILE {
						op := r.uploadFileOp(file, opMask)
						ops = append(ops, op)
					} else {
						op := OperationRequest{
							OperationType: constants.FileOperationCreateDir,
							RemotePath:    file.Path,
						}
						ops = append(ops, op)
					}
					res.lastMatchedPath = file.Path
					res.lastIndex = idx
				}
			}
		}

		if r.resMap[latestVersion].listCompleted {
			for version, res := range r.resMap {
				if version == latestVersion {
					continue
				}
				for res.idx < len(res.oTR.Refs) {
					//delete the file
					l.Logger.Debug("Deleting file: ", res.oTR.Refs[res.idx].Path)
					opMask := r.versionMap[version]
					op := OperationRequest{
						OperationType: constants.FileOperationDelete,
						RemotePath:    res.oTR.Refs[res.idx].Path,
						Mask:          &opMask,
					}
					res.idx++
					ops = append(ops, op)
				}
				if res.listCompleted {
					for i := res.lastIndex + 1; i < len(r.resMap[latestVersion].oTR.Refs); i++ {
						// upload the file
						l.Logger.Debug("Uploading file: ", r.resMap[latestVersion].oTR.Refs[i].Path)
						opMask := r.versionMap[version]
						if fileType == fileref.FILE {
							op := r.uploadFileOp(r.resMap[latestVersion].oTR.Refs[i], opMask)
							ops = append(ops, op)
						} else {
							op := OperationRequest{
								OperationType: constants.FileOperationCreateDir,
								RemotePath:    r.resMap[latestVersion].oTR.Refs[i].Path,
							}
							ops = append(ops, op)
						}
					}
				}
			}
		} else {
			minLastIndex := len(r.resMap[latestVersion].oTR.Refs)
			for version, res := range r.resMap {
				if version == latestVersion {
					continue
				}
				if !res.listCompleted {
					if res.lastIndex < minLastIndex {
						minLastIndex = res.lastIndex
					}
					if res.idx < len(res.oTR.Refs) {
						res.oTR.Refs = res.oTR.Refs[res.idx:]
					} else {
						res.oTR.Refs = nil
					}
				}
			}
			if minLastIndex+1 < len(r.resMap[latestVersion].oTR.Refs) {
				r.resMap[latestVersion].oTR.Refs = r.resMap[latestVersion].oTR.Refs[minLastIndex+1:]
			} else {
				r.resMap[latestVersion].oTR.Refs = nil
			}
		}

		err := r.repairOperation(ops, latestVersion, r.resMap[latestVersion].oTR.OffsetPath)
		if err != nil {
			l.Logger.Error("Failed to repair files: ", err)
			return err
		}

		completedLists := 0
		for _, res := range r.resMap {
			if res.listCompleted {
				completedLists++
			}
		}
		if completedLists == len(r.versionMap) {
			break
		}
	}

	return nil
}

func (r *RepairRequest) getRefsWithVersion(ctx context.Context, fileType string) {
	mpLock := sync.Mutex{}
	resWg := sync.WaitGroup{}
	for version, mask := range r.versionMap {
		resWg.Add(1)
		currVersion := version
		currMask := mask
		go func(version int64, mask zboxutil.Uint128) {
			defer resWg.Done()
			var offsetPath string
			mpLock.Lock()
			if r.resMap[version] != nil {
				offsetPath = r.resMap[version].oTR.OffsetPath
			}
			mpLock.Unlock()
			l.Logger.Debug("Getting refs for repair path: ", r.repairPath, " version: ", version, " mask: ", mask.CountOnes())
			res, err := r.allocation.GetRefs(r.repairPath, offsetPath, "", "", fileType, fileref.REGULAR, 0, getRefPageLimit, WithSingleBlobber(true), WithObjectMask(mask), WithObjectContext(ctx))
			if err != nil {
				if code, _ := zboxutil.GetErrorMessageCode(err.Error()); code == INVALID_PATH {
					err = nil
					res = &ObjectTreeResult{Refs: make([]ORef, 0)}
				}
			}
			mpLock.Lock()
			currRes := r.resMap[version]
			if currRes != nil {
				if res != nil {
					currRes.oTR.Refs = append(currRes.oTR.Refs, res.Refs...)
					if len(res.Refs) < getRefPageLimit {
						currRes.listCompleted = true
					}
				}
				currRes.idx = 0
				currRes.lastIndex = -1
			} else {
				getResult := &getRes{oTR: res, err: err, lastIndex: -1}
				if res != nil && len(res.Refs) < getRefPageLimit {
					getResult.listCompleted = true
				}
				r.resMap[version] = getResult
			}
			mpLock.Unlock()
		}(currVersion, currMask)
	}
	resWg.Wait()
}

func (r *RepairRequest) repairOperation(ops []OperationRequest, latestVersion int64, repairOffsetPath string) error {
	err := r.allocation.DoMultiOperation(ops, WithRepair(latestVersion, repairOffsetPath))
	if err != nil {
		l.Logger.Error("repair_file_failed", zap.Error(err))
		status := r.statusCB != nil
		for _, op := range ops {
			if op.DownloadFile {
				_ = r.allocation.CancelDownload(op.RemotePath)
			}
			if status {
				r.statusCB.Error(r.allocation.ID, op.RemotePath, OpRepair, err)
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
	return err
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
