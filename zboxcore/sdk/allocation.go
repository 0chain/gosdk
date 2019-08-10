package sdk

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/0chain/gosdk/zboxcore/fileref"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

var (
	noBLOBBERS     = errors.New("No Blobbers set in this allocation")
	notInitialized = common.NewError("sdk_not_initialized", "Please call InitStorageSDK Init and use GetAllocation to get the allocation object")
)

type ConsolidatedFileMeta struct {
	Name          string
	Type          string
	Path          string
	PathHash      string
	LookupHash    string
	Hash          string
	MimeType      string
	Size          int64
	ThumbnailSize int64
	ThumbnailHash string
}

type AllocationStats struct {
	UsedSize                  int64  `json:"used_size"`
	NumWrites                 int64  `json:"num_of_writes"`
	NumReads                  int64  `json:"num_of_reads"`
	TotalChallenges           int64  `json:"total_challenges"`
	OpenChallenges            int64  `json:"num_open_challenges"`
	SuccessChallenges         int64  `json:"num_success_challenges"`
	FailedChallenges          int64  `json:"num_failed_challenges"`
	LastestClosedChallengeTxn string `json:"latest_closed_challenge"`
}

type Allocation struct {
	ID           string                    `json:"id"`
	DataShards   int                       `json:"data_shards"`
	ParityShards int                       `json:"parity_shards"`
	Size         int64                     `json:"size"`
	Expiration   int64                     `json:"expiration_date"`
	Blobbers     []*blockchain.StorageNode `json:"blobbers"`
	Stats        *AllocationStats          `json:"stats"`

	uploadChan          chan *UploadRequest
	downloadChan        chan *DownloadRequest
	ctx                 context.Context
	ctxCancelF          context.CancelFunc
	mutex               *sync.Mutex
	downloadProgressMap map[string]*DownloadRequest
	initialized         bool
}

// For sync app
const (
	Upload   = "Upload"
	Download = "Download"
	Update   = "Update"
	Delete   = "Delete"
)

type FileList struct {
	Path string
	Size int64
	Hash string
}

type FileDiff struct {
	Op   string `json:"operation"`
	Path string `json:"path"`
}

func (a *Allocation) GetStats() *AllocationStats {
	return a.Stats
}

func (a *Allocation) InitAllocation() {
	// if a.uploadChan != nil {
	// 	close(a.uploadChan)
	// }
	// if a.downloadChan != nil {
	// 	close(a.downloadChan)
	// }
	// if a.ctx != nil {
	// 	a.ctx.Done()
	// }
	// for _, v := range a.downloadProgressMap {
	// 	v.isDownloadCanceled = true
	// }
	a.uploadChan = make(chan *UploadRequest, 10)
	a.downloadChan = make(chan *DownloadRequest, 10)
	a.ctx, a.ctxCancelF = context.WithCancel(context.Background())
	a.downloadProgressMap = make(map[string]*DownloadRequest)
	a.mutex = &sync.Mutex{}
	a.startWorker(a.ctx)
	InitCommitWorker(a.Blobbers)
	InitBlockDownloader(a.Blobbers)
	a.initialized = true
}

func (a *Allocation) isInitialized() bool {
	return a.initialized && sdkInitialized
}

func (a *Allocation) startWorker(ctx context.Context) {
	go a.dispatchWork(ctx)
}

func (a *Allocation) dispatchWork(ctx context.Context) {
	for true {
		select {
		case <-ctx.Done():
			Logger.Info("Upload cancelled by the parent")
			return
		case uploadReq := <-a.uploadChan:
			fmt.Printf("received a upload request for %v %v\n", uploadReq.filepath, uploadReq.remotefilepath)
			go uploadReq.processUpload(ctx, a)
		case downloadReq := <-a.downloadChan:
			fmt.Printf("received a download request for %v\n", downloadReq.remotefilepath)
			go downloadReq.processDownload(ctx, a)
		}
	}
}

func (a *Allocation) UpdateFile(localpath string, remotepath string, status StatusCallback) error {
	return a.uploadOrUpdateFile(localpath, remotepath, status, true, "")
}

func (a *Allocation) UploadFile(localpath string, remotepath string, status StatusCallback) error {
	return a.uploadOrUpdateFile(localpath, remotepath, status, false, "")
}

func (a *Allocation) UpdateFileWithThumbnail(localpath string, remotepath string, thumbnailpath string, status StatusCallback) error {
	return a.uploadOrUpdateFile(localpath, remotepath, status, true, thumbnailpath)
}

func (a *Allocation) UploadFileWithThumbnail(localpath string, remotepath string, thumbnailpath string, status StatusCallback) error {
	return a.uploadOrUpdateFile(localpath, remotepath, status, false, thumbnailpath)
}

func (a *Allocation) uploadOrUpdateFile(localpath string, remotepath string, status StatusCallback, isUpdate bool, thumbnailpath string) error {
	if !a.isInitialized() {
		return notInitialized
	}
	fileInfo, err := os.Stat(localpath)
	if err != nil {
		return fmt.Errorf("Local file error: %s", err.Error())
	}
	thumbnailSize := int64(0)
	if len(thumbnailpath) > 0 {
		fileInfo, err := os.Stat(thumbnailpath)
		if err != nil {
			thumbnailSize = 0
			thumbnailpath = ""
		} else {
			thumbnailSize = fileInfo.Size()
		}

	}

	remotepath = filepath.Clean(remotepath)
	isabs := filepath.IsAbs(remotepath)
	if !isabs {
		return common.NewError("invalid_path", "Path should be valid and absolute")
	}
	remotepath = zboxutil.GetFullRemotePath(localpath, remotepath)
	var fileName string
	_, fileName = filepath.Split(remotepath)
	uploadReq := &UploadRequest{}
	uploadReq.remotefilepath = remotepath
	uploadReq.thumbnailpath = thumbnailpath
	uploadReq.filepath = localpath
	uploadReq.filemeta = &UploadFileMeta{}
	uploadReq.filemeta.Name = fileName
	uploadReq.filemeta.Size = fileInfo.Size()
	uploadReq.filemeta.Path = remotepath
	uploadReq.filemeta.ThumbnailSize = thumbnailSize
	uploadReq.remaining = uploadReq.filemeta.Size
	uploadReq.thumbRemaining = uploadReq.filemeta.ThumbnailSize
	uploadReq.isRepair = false
	uploadReq.isUpdate = isUpdate
	uploadReq.connectionID = zboxutil.NewConnectionId()
	uploadReq.statusCallback = status
	uploadReq.datashards = a.DataShards
	uploadReq.parityshards = a.ParityShards
	uploadReq.uploadMask = ((1 << uint32(len(a.Blobbers))) - 1)
	uploadReq.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	uploadReq.fullconsensus = float32(a.DataShards + a.ParityShards)
	go func() {
		a.uploadChan <- uploadReq
	}()
	return nil
}
func (a *Allocation) DownloadFile(localPath string, remotePath string, status StatusCallback) error {
	return a.downloadFile(localPath, remotePath, DOWNLOAD_CONTENT_FULL, status)
}

func (a *Allocation) DownloadThumbnail(localPath string, remotePath string, status StatusCallback) error {
	return a.downloadFile(localPath, remotePath, DOWNLOAD_CONTENT_THUMB, status)
}

func (a *Allocation) downloadFile(localPath string, remotePath string, contentMode string, status StatusCallback) error {
	if !a.isInitialized() {
		return notInitialized
	}
	if stat, err := os.Stat(localPath); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("Local path is not a directory '%s'", localPath)
		}
		localPath = strings.TrimRight(localPath, "/")
		_, rFile := filepath.Split(remotePath)
		localPath = fmt.Sprintf("%s/%s", localPath, rFile)
		if _, err := os.Stat(localPath); err == nil {
			return fmt.Errorf("Local file already exists '%s'", localPath)
		}
	}
	lPath, _ := filepath.Split(localPath)
	os.MkdirAll(lPath, os.ModePerm)

	if len(a.Blobbers) <= 1 {
		return noBLOBBERS
	}

	downloadReq := &DownloadRequest{}
	downloadReq.allocationID = a.ID
	downloadReq.ctx, _ = context.WithCancel(a.ctx)
	downloadReq.localpath = localPath
	downloadReq.remotefilepath = remotePath
	downloadReq.statusCallback = status
	downloadReq.downloadMask = ((1 << uint32(len(a.Blobbers))) - 1)
	downloadReq.blobbers = a.Blobbers
	downloadReq.datashards = a.DataShards
	downloadReq.parityshards = a.ParityShards
	downloadReq.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	downloadReq.fullconsensus = float32(a.DataShards + a.ParityShards)
	downloadReq.completedCallback = func(remotepath string, remotepathhash string) {
		a.mutex.Lock()
		defer a.mutex.Unlock()
		delete(a.downloadProgressMap, remotepath)
	}
	downloadReq.contentMode = contentMode
	go func() {
		a.downloadChan <- downloadReq
		a.mutex.Lock()
		defer a.mutex.Unlock()
		a.downloadProgressMap[remotePath] = downloadReq
	}()
	return nil
}

func (a *Allocation) ListDirFromAuthTicket(authTicket string, lookupHash string) (*ListResult, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}
	sEnc, err := base64.StdEncoding.DecodeString(authTicket)
	if err != nil {
		return nil, common.NewError("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	at := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return nil, common.NewError("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	if len(at.FilePathHash) == 0 || len(lookupHash) == 0 {
		return nil, common.NewError("invalid_path", "Invalid path for the list")
	}

	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.blobbers = a.Blobbers
	listReq.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	listReq.fullconsensus = float32(a.DataShards + a.ParityShards)
	listReq.ctx = a.ctx
	listReq.remotefilepathhash = lookupHash
	listReq.authToken = at
	ref := listReq.GetListFromBlobbers()
	if ref != nil {
		return ref, nil
	}
	return nil, common.NewError("list_request_failed", "Failed to get list response from the blobbers")
}

func (a *Allocation) ListDir(path string) (*ListResult, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}
	if len(path) == 0 {
		return nil, common.NewError("invalid_path", "Invalid path for the list")
	}
	path = filepath.Clean(path)
	isabs := filepath.IsAbs(path)
	if !isabs {
		return nil, common.NewError("invalid_path", "Path should be valid and absolute")
	}
	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.blobbers = a.Blobbers
	listReq.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	listReq.fullconsensus = float32(a.DataShards + a.ParityShards)
	listReq.ctx = a.ctx
	listReq.remotefilepath = path
	ref := listReq.GetListFromBlobbers()
	if ref != nil {
		return ref, nil
	}
	return nil, common.NewError("list_request_failed", "Failed to get list response from the blobbers")
}

func (a *Allocation) GetFileMeta(path string) (*ConsolidatedFileMeta, error) {
	result := &ConsolidatedFileMeta{}
	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.blobbers = a.Blobbers
	listReq.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	listReq.fullconsensus = float32(a.DataShards + a.ParityShards)
	listReq.ctx = a.ctx
	listReq.remotefilepath = path
	_, ref, _ := listReq.getFileConsensusFromBlobbers()
	if ref != nil {
		result.Type = ref.Type
		result.Name = ref.Name
		result.Hash = ref.ActualFileHash
		result.LookupHash = ref.LookupHash
		result.MimeType = ref.MimeType
		result.Path = ref.Path
		result.PathHash = ref.PathHash
		result.Size = ref.ActualFileSize
		result.ThumbnailHash = ref.ActualThumbnailHash
		result.ThumbnailSize = ref.ActualThumbnailSize
		return result, nil
	}
	return nil, common.NewError("file_meta_error", "Error getting the file meta data from blobbers")
}

func (a *Allocation) GetFileMetaFromAuthTicket(authTicket string, lookupHash string) (*ConsolidatedFileMeta, error) {
	result := &ConsolidatedFileMeta{}
	sEnc, err := base64.StdEncoding.DecodeString(authTicket)
	if err != nil {
		return nil, common.NewError("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	at := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return nil, common.NewError("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	if len(at.FilePathHash) == 0 || len(lookupHash) == 0 {
		return nil, common.NewError("invalid_path", "Invalid path for the list")
	}

	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.blobbers = a.Blobbers
	listReq.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	listReq.fullconsensus = float32(a.DataShards + a.ParityShards)
	listReq.ctx = a.ctx
	listReq.remotefilepathhash = lookupHash
	listReq.authToken = at
	_, ref, _ := listReq.getFileConsensusFromBlobbers()
	if ref != nil {
		result.Type = ref.Type
		result.Name = ref.Name
		result.Hash = ref.ActualFileHash
		result.LookupHash = ref.LookupHash
		result.MimeType = ref.MimeType
		result.Path = ref.Path
		result.PathHash = ref.PathHash
		result.Size = ref.ActualFileSize
		result.ThumbnailHash = ref.ActualThumbnailHash
		result.ThumbnailSize = ref.ActualThumbnailSize
		return result, nil
	}
	return nil, common.NewError("file_meta_error", "Error getting the file meta data from blobbers")
}

func (a *Allocation) GetFileStats(path string) (map[string]*FileStats, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}
	if len(path) == 0 {
		return nil, common.NewError("invalid_path", "Invalid path for the list")
	}
	path = filepath.Clean(path)
	isabs := filepath.IsAbs(path)
	if !isabs {
		return nil, common.NewError("invalid_path", "Path should be valid and absolute")
	}
	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.blobbers = a.Blobbers
	listReq.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	listReq.fullconsensus = float32(a.DataShards + a.ParityShards)
	listReq.ctx = a.ctx
	listReq.remotefilepath = path
	ref := listReq.getFileStatsFromBlobbers()
	if ref != nil {
		return ref, nil
	}
	return nil, common.NewError("file_stats_request_failed", "Failed to get file stats response from the blobbers")
}

func (a *Allocation) DeleteFile(path string) error {
	if !a.isInitialized() {
		return notInitialized
	}
	if len(path) == 0 {
		return common.NewError("invalid_path", "Invalid path for the list")
	}
	path = filepath.Clean(path)
	isabs := filepath.IsAbs(path)
	if !isabs {
		return common.NewError("invalid_path", "Path should be valid and absolute")
	}

	req := &DeleteRequest{}
	req.blobbers = a.Blobbers
	req.allocationID = a.ID
	req.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	req.fullconsensus = float32(a.DataShards + a.ParityShards)
	req.ctx = a.ctx
	req.remotefilepath = path
	req.deleteMask = 0
	req.listMask = 0
	req.connectionID = zboxutil.NewConnectionId()
	err := req.ProcessDelete()
	return err
}

func (a *Allocation) RenameObject(path string, destName string) error {
	if !a.isInitialized() {
		return notInitialized
	}
	if len(path) == 0 {
		return common.NewError("invalid_path", "Invalid path for the list")
	}
	path = filepath.Clean(path)
	isabs := filepath.IsAbs(path)
	if !isabs {
		return common.NewError("invalid_path", "Path should be valid and absolute")
	}

	req := &RenameRequest{}
	req.blobbers = a.Blobbers
	req.allocationID = a.ID
	req.newName = destName
	req.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	req.fullconsensus = float32(a.DataShards + a.ParityShards)
	req.ctx = a.ctx
	req.remotefilepath = path
	req.renameMask = 0
	req.connectionID = zboxutil.NewConnectionId()
	err := req.ProcessRename()
	return err
}

func (a *Allocation) CopyObject(path string, destPath string) error {
	if !a.isInitialized() {
		return notInitialized
	}
	if len(path) == 0 || len(destPath) == 0 {
		return common.NewError("invalid_path", "Invalid path for copy")
	}
	path = filepath.Clean(path)
	isabs := filepath.IsAbs(path)
	if !isabs {
		return common.NewError("invalid_path", "Path should be valid and absolute")
	}

	req := &CopyRequest{}
	req.blobbers = a.Blobbers
	req.allocationID = a.ID
	req.destPath = destPath
	req.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	req.fullconsensus = float32(a.DataShards + a.ParityShards)
	req.ctx = a.ctx
	req.remotefilepath = path
	req.copyMask = 0
	req.connectionID = zboxutil.NewConnectionId()
	err := req.ProcessCopy()
	return err
}

func (a *Allocation) GetAuthTicketForShare(path string, filename string, referenceType string, refereeClientID string) (string, error) {
	if !a.isInitialized() {
		return "", notInitialized
	}
	if len(path) == 0 {
		return "", common.NewError("invalid_path", "Invalid path for the list")
	}
	path = filepath.Clean(path)
	isabs := filepath.IsAbs(path)
	if !isabs {
		return "", common.NewError("invalid_path", "Path should be valid and absolute")
	}

	shareReq := &ShareRequest{}
	shareReq.allocationID = a.ID
	shareReq.blobbers = a.Blobbers
	shareReq.ctx = a.ctx
	shareReq.remotefilepath = path
	shareReq.remotefilename = filename
	if referenceType == fileref.DIRECTORY {
		shareReq.refType = fileref.DIRECTORY
	} else {
		shareReq.refType = fileref.FILE
	}
	authTicket, err := shareReq.GetAuthTicket(refereeClientID)
	if err != nil {
		return "", err
	}

	return authTicket, nil
}

func (a *Allocation) CancelDownload(remotepath string) error {
	if downloadReq, ok := a.downloadProgressMap[remotepath]; ok {
		downloadReq.isDownloadCanceled = true
		return nil
	}
	return common.NewError("remote_path_not_found", "Invalid path. Do download in progress for the path "+remotepath)
}

func (a *Allocation) DownloadThumbnailFromAuthTicket(localPath string, authTicket string, remoteLookupHash string, remoteFilename string, status StatusCallback) error {
	return a.downloadFromAuthTicket(localPath, authTicket, remoteLookupHash, remoteFilename, DOWNLOAD_CONTENT_THUMB, status)
}

func (a *Allocation) DownloadFromAuthTicket(localPath string, authTicket string, remoteLookupHash string, remoteFilename string, status StatusCallback) error {
	return a.downloadFromAuthTicket(localPath, authTicket, remoteLookupHash, remoteFilename, DOWNLOAD_CONTENT_FULL, status)
}

func (a *Allocation) downloadFromAuthTicket(localPath string, authTicket string, remoteLookupHash string, remoteFilename string, contentMode string, status StatusCallback) error {
	if !a.isInitialized() {
		return notInitialized
	}
	sEnc, err := base64.StdEncoding.DecodeString(authTicket)
	if err != nil {
		return common.NewError("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	at := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return common.NewError("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	if stat, err := os.Stat(localPath); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("Local path is not a directory '%s'", localPath)
		}
		localPath = strings.TrimRight(localPath, "/")
		_, rFile := filepath.Split(remoteFilename)
		localPath = fmt.Sprintf("%s/%s", localPath, rFile)
		if _, err := os.Stat(localPath); err == nil {
			return fmt.Errorf("Local file already exists '%s'", localPath)
		}
	}
	if len(a.Blobbers) <= 1 {
		return noBLOBBERS
	}

	downloadReq := &DownloadRequest{}
	downloadReq.allocationID = a.ID
	downloadReq.ctx, _ = context.WithCancel(a.ctx)
	downloadReq.localpath = localPath
	downloadReq.remotefilepathhash = remoteLookupHash
	downloadReq.authTicket = at
	downloadReq.statusCallback = status
	downloadReq.downloadMask = ((1 << uint32(len(a.Blobbers))) - 1)
	downloadReq.blobbers = a.Blobbers
	downloadReq.datashards = a.DataShards
	downloadReq.parityshards = a.ParityShards
	downloadReq.contentMode = contentMode
	downloadReq.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	downloadReq.fullconsensus = float32(a.DataShards + a.ParityShards)
	downloadReq.completedCallback = func(remotepath string, remotepathHash string) {
		a.mutex.Lock()
		defer a.mutex.Unlock()
		delete(a.downloadProgressMap, remotepathHash)
	}
	go func() {
		a.downloadChan <- downloadReq
		a.mutex.Lock()
		defer a.mutex.Unlock()
		a.downloadProgressMap[remoteLookupHash] = downloadReq
	}()
	return nil
}

func (a *Allocation) getRemoteFileDir(dirList []string, fileList *[]FileList) ([]string, error) {
	childDirList := make([]string, 0)
	for _, dir := range dirList {
		ref, err := a.ListDir(dir)
		if err != nil {
			return []string{}, err
		}
		for _, child := range ref.Children {
			if child.Type == fileref.FILE {
				*fileList = append(*fileList, FileList{Path: child.Path, Size: child.Size, Hash: child.Hash})
			} else {
				childDirList = append(childDirList, child.Path)
			}
		}
	}
	return childDirList, nil
}

func (a *Allocation) getRemoteFileList() ([]FileList, error) {
	var remoteList []FileList
	dirs := []string{"/"}
	var err error
	for {
		dirs, err = a.getRemoteFileDir(dirs, &remoteList)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		if len(dirs) == 0 {
			break
		}
	}
	return remoteList, err
}

func calcFileHash(filePath string) string {
	fp, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	h := sha1.New()
	if _, err := io.Copy(h, fp); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func addLocalFileList(root string, fileList *[]FileList, filter map[string]bool) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		// Filter out
		if _, ok := filter[info.Name()]; ok {
			return nil
		}
		if !info.IsDir() {
			*fileList = append(*fileList, FileList{Path: "/" + strings.TrimLeft(path, root), Size: info.Size(), Hash: calcFileHash(path)})
		}
		return nil
	}
}

func getLocalFileList(rootPath string, filters []string) ([]FileList, error) {
	var localList []FileList
	filterMap := make(map[string]bool)
	for _, f := range filters {
		filterMap[f] = true
	}
	err := filepath.Walk(rootPath, addLocalFileList(rootPath, &localList, filterMap))
	return localList, err
}

func findDelta(remote []FileList, local []FileList) []FileDiff {
	fMap := make(map[string]string)
	// Create a remote hash map
	for _, rFile := range remote {
		fMap[rFile.Path] = rFile.Hash
	}
	uMap := make(map[string]string)
	// Compare it with local hash map and get union
	for _, lFile := range local {
		if hash, ok := fMap[lFile.Path]; ok {
			uMap[lFile.Path] = hash
		}
	}

	var lFDiff []FileDiff
	// Find files to be download or update
	for _, rFile := range remote {
		op := Download
		if hash, ok := uMap[rFile.Path]; ok {
			if hash == rFile.Hash {
				continue
			}
			op = Update
		}
		lFDiff = append(lFDiff, FileDiff{Path: rFile.Path, Op: op})
	}

	for _, lFile := range local {
		op := Upload
		if hash, ok := uMap[lFile.Path]; ok {
			if hash == lFile.Hash {
				continue
			}
			op = Update
		}
		// TODO: Delete
		lFDiff = append(lFDiff, FileDiff{Path: lFile.Path, Op: op})
	}
	return lFDiff
}

func (a *Allocation) GetAllocationDiff(lastKnownAllocationRoot string, localRootPath string, localFileFilters []string) ([]FileDiff, error) {
	var lFdiff []FileDiff
	// Get flat file list from remote
	remoteFileList, err := a.getRemoteFileList()
	if err != nil {
		return lFdiff, fmt.Errorf("error getting list dir from remote. %v", err)
	}

	// Get flat file list on the local filesystem
	localFileList, err := getLocalFileList(localRootPath, localFileFilters)
	if err != nil {
		return lFdiff, fmt.Errorf("error getting list dir from local. %v", err)
	}

	// Get the file diff with operation
	lFdiff = findDelta(remoteFileList, localFileList)

	return lFdiff, nil
}
