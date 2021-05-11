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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

var (
	noBLOBBERS     = errors.New("No Blobbers set in this allocation")
	notInitialized = common.NewError("sdk_not_initialized", "Please call InitStorageSDK Init and use GetAllocation to get the allocation object")
)

var GetFileInfo = func(localpath string) (os.FileInfo, error) {
	return os.Stat(localpath)
}

type BlobberAllocationStats struct {
	BlobberID        string
	BlobberURL       string
	ID               string `json:"ID"`
	Tx               string `json:"Tx"`
	TotalSize        int64  `json:"TotalSize"`
	UsedSize         int    `json:"UsedSize"`
	OwnerID          string `json:"OwnerID"`
	OwnerPublicKey   string `json:"OwnerPublicKey"`
	Expiration       int    `json:"Expiration"`
	AllocationRoot   string `json:"AllocationRoot"`
	BlobberSize      int    `json:"BlobberSize"`
	BlobberSizeUsed  int    `json:"BlobberSizeUsed"`
	LatestRedeemedWM string `json:"LatestRedeemedWM"`
	IsRedeemRequired bool   `json:"IsRedeemRequired"`
	CleanedUp        bool   `json:"CleanedUp"`
	Finalized        bool   `json:"Finalized"`
	Terms            []struct {
		ID           int    `json:"ID"`
		BlobberID    string `json:"BlobberID"`
		AllocationID string `json:"AllocationID"`
		ReadPrice    int    `json:"ReadPrice"`
		WritePrice   int    `json:"WritePrice"`
	} `json:"Terms"`
	PayerID string `json:"PayerID"`
}

type ConsolidatedFileMeta struct {
	Name            string
	Type            string
	Path            string
	LookupHash      string
	Hash            string
	MimeType        string
	Size            int64
	ActualFileSize  int64
	ActualNumBlocks int64
	EncryptedKey    string
	CommitMetaTxns  []fileref.CommitMetaTxn
	Collaborators   []fileref.Collaborator
	Attributes      fileref.Attributes
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

// PriceRange represents a price range allowed by user to filter blobbers.
type PriceRange struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

// IsValid price range.
func (pr *PriceRange) IsValid() bool {
	return 0 <= pr.Min && pr.Min <= pr.Max
}

// Terms represents Blobber terms. A Blobber can update its terms,
// but any existing offer will use terms of offer signing time.
type Terms struct {
	ReadPrice               common.Balance `json:"read_price"`  // tokens / GB
	WritePrice              common.Balance `json:"write_price"` // tokens / GB
	MinLockDemand           float64        `json:"min_lock_demand"`
	MaxOfferDuration        time.Duration  `json:"max_offer_duration"`
	ChallengeCompletionTime time.Duration  `json:"challenge_completion_time"`
}

type BlobberAllocation struct {
	BlobberID       string         `json:"blobber_id"`
	Size            int64          `json:"size"`
	Terms           Terms          `json:"terms"`
	MinLockDemand   common.Balance `json:"min_lock_demand"`
	Spent           common.Balance `json:"spent"`
	Penalty         common.Balance `json:"penalty"`
	ReadReward      common.Balance `json:"read_reward"`
	Returned        common.Balance `json:"returned"`
	ChallengeReward common.Balance `json:"challenge_reward"`
	FinalReward     common.Balance `json:"final_reward"`
}

type Allocation struct {
	ID             string                    `json:"id"`
	Tx             string                    `json:"tx"`
	DataShards     int                       `json:"data_shards"`
	ParityShards   int                       `json:"parity_shards"`
	Size           int64                     `json:"size"`
	Expiration     int64                     `json:"expiration_date"`
	Owner          string                    `json:"owner_id"`
	OwnerPublicKey string                    `json:"owner_public_key"`
	Payer          string                    `json:"payer_id"`
	Blobbers       []*blockchain.StorageNode `json:"blobbers"`
	Stats          *AllocationStats          `json:"stats"`
	TimeUnit       time.Duration             `json:"time_unit"`

	// BlobberDetails contains real terms used for the allocation.
	// If the allocation has updated, then terms calculated using
	// weighted average values.
	BlobberDetails []*BlobberAllocation `json:"blobber_details"`

	// ReadPriceRange is requested reading prices range.
	ReadPriceRange PriceRange `json:"read_price_range"`
	// WritePriceRange is requested writing prices range.
	WritePriceRange PriceRange `json:"write_price_range"`

	ChallengeCompletionTime time.Duration    `json:"challenge_completion_time"`
	StartTime               common.Timestamp `json:"start_time"`
	Finalized               bool             `json:"finalized,omitempty"`
	Canceled                bool             `json:"canceled,omitempty"`
	MovedToChallenge        common.Balance   `json:"moved_to_challenge,omitempty"`
	MovedBack               common.Balance   `json:"moved_back,omitempty"`
	MovedToValidators       common.Balance   `json:"moved_to_validators,omitempty"`

	numBlockDownloads       int
	uploadChan              chan *UploadRequest
	downloadChan            chan *DownloadRequest
	repairChan              chan *RepairRequest
	ctx                     context.Context
	ctxCancelF              context.CancelFunc
	mutex                   *sync.Mutex
	uploadProgressMap       map[string]*UploadRequest
	downloadProgressMap     map[string]*DownloadRequest
	repairRequestInProgress *RepairRequest
	initialized             bool
}

func (a *Allocation) GetStats() *AllocationStats {
	return a.Stats
}

func (a *Allocation) GetBlobberStats() map[string]*BlobberAllocationStats {
	numList := len(a.Blobbers)
	wg := &sync.WaitGroup{}
	wg.Add(numList)
	rspCh := make(chan *BlobberAllocationStats, numList)
	for _, blobber := range a.Blobbers {
		go getAllocationDataFromBlobber(blobber, a.Tx, rspCh, wg)
	}
	wg.Wait()
	result := make(map[string]*BlobberAllocationStats, len(a.Blobbers))
	for i := 0; i < numList; i++ {
		resp := <-rspCh
		result[resp.BlobberURL] = resp
	}
	return result
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
	a.repairChan = make(chan *RepairRequest, 1)
	a.ctx, a.ctxCancelF = context.WithCancel(context.Background())
	a.uploadProgressMap = make(map[string]*UploadRequest)
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
			Logger.Info(fmt.Sprintf("received a upload request for %v %v\n", uploadReq.filepath, uploadReq.remotefilepath))
			go uploadReq.processUpload(ctx, a)
		case downloadReq := <-a.downloadChan:
			Logger.Info(fmt.Sprintf("received a download request for %v\n", downloadReq.remotefilepath))
			go downloadReq.processDownload(ctx)
		case repairReq := <-a.repairChan:
			Logger.Info(fmt.Sprintf("received a repair request for %v\n", repairReq.listDir.Path))
			go repairReq.processRepair(ctx, a)
		}
	}
}

func (a *Allocation) UpdateFile(localpath string, remotepath string,
	attrs fileref.Attributes, status StatusCallback) error {

	return a.uploadOrUpdateFile(localpath, remotepath, status, true, "", false,
		false, attrs)
}

func (a *Allocation) UploadFile(localpath string, remotepath string,
	attrs fileref.Attributes, status StatusCallback) error {

	return a.uploadOrUpdateFile(localpath, remotepath, status, false, "", false,
		false, attrs)
}

func (a *Allocation) RepairFile(localpath string, remotepath string,
	status StatusCallback) error {

	return a.uploadOrUpdateFile(localpath, remotepath, status, false, "",
		false, true, fileref.Attributes{})
}

func (a *Allocation) UpdateFileWithThumbnail(localpath string, remotepath string,
	thumbnailpath string, attrs fileref.Attributes, status StatusCallback) error {

	return a.uploadOrUpdateFile(localpath, remotepath, status, true,
		thumbnailpath, false, false, attrs)
}

func (a *Allocation) UploadFileWithThumbnail(localpath string,
	remotepath string, thumbnailpath string, attrs fileref.Attributes,
	status StatusCallback) error {

	return a.uploadOrUpdateFile(localpath, remotepath, status, false,
		thumbnailpath, false, false, attrs)
}

func (a *Allocation) EncryptAndUpdateFile(localpath string, remotepath string,
	attrs fileref.Attributes, status StatusCallback) error {

	return a.uploadOrUpdateFile(localpath, remotepath, status, true, "", true,
		false, attrs)
}

func (a *Allocation) EncryptAndUploadFile(localpath string, remotepath string,
	attrs fileref.Attributes, status StatusCallback) error {

	return a.uploadOrUpdateFile(localpath, remotepath, status, false, "", true,
		false, attrs)
}

func (a *Allocation) EncryptAndUpdateFileWithThumbnail(localpath string,
	remotepath string, thumbnailpath string, attrs fileref.Attributes, status StatusCallback) error {

	return a.uploadOrUpdateFile(localpath, remotepath, status, true,
		thumbnailpath, true, false, attrs)
}

func (a *Allocation) EncryptAndUploadFileWithThumbnail(localpath string,
	remotepath string, thumbnailpath string, attrs fileref.Attributes,
	status StatusCallback) error {

	return a.uploadOrUpdateFile(localpath, remotepath, status, false,
		thumbnailpath, true, false, attrs)
}

func (a *Allocation) uploadOrUpdateFile(localpath string, remotepath string,
	status StatusCallback, isUpdate bool, thumbnailpath string, encryption bool,
	isRepair bool, attrs fileref.Attributes) error {

	if !a.isInitialized() {
		return notInitialized
	}

	fileInfo, err := GetFileInfo(localpath)
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

	remotepath = zboxutil.RemoteClean(remotepath)
	isabs := zboxutil.IsRemoteAbs(remotepath)
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
	uploadReq.filemeta.Attributes = attrs
	uploadReq.remaining = uploadReq.filemeta.Size
	uploadReq.thumbRemaining = uploadReq.filemeta.ThumbnailSize
	uploadReq.isUpdate = isUpdate
	uploadReq.isRepair = isRepair
	uploadReq.connectionID = zboxutil.NewConnectionId()
	uploadReq.statusCallback = status
	uploadReq.datashards = a.DataShards
	uploadReq.parityshards = a.ParityShards
	uploadReq.setUploadMask(len(a.Blobbers))
	uploadReq.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	uploadReq.fullconsensus = float32(a.DataShards + a.ParityShards)
	uploadReq.isEncrypted = encryption
	uploadReq.completedCallback = func(filepath string) {
		a.mutex.Lock()
		defer a.mutex.Unlock()
		delete(a.uploadProgressMap, filepath)
	}

	if uploadReq.isRepair {
		found, repairRequired, fileRef, err := a.RepairRequired(remotepath)
		if err != nil {
			return err
		}

		if !repairRequired {
			return fmt.Errorf("Repair not required")
		}

		file, _ := ioutil.ReadFile(localpath)
		hash := sha1.New()
		hash.Write(file)
		contentHash := hex.EncodeToString(hash.Sum(nil))
		if contentHash != fileRef.ActualFileHash {
			return fmt.Errorf("Content hash doesn't match")
		}

		uploadReq.filemeta.Hash = fileRef.ActualFileHash
		uploadReq.uploadMask = found.Not().And(uploadReq.uploadMask)
		uploadReq.fullconsensus = float32(uploadReq.uploadMask.Add64(1).TrailingZeros())
	}

	if !uploadReq.IsFullConsensusSupported() {
		return fmt.Errorf("allocation requires [%v] blobbers, which is greater than the maximum permitted number of [%v]. reduce number of data or parity shards and try again", uploadReq.fullconsensus, uploadReq.GetMaxBlobbersSupported())
	}

	go func() {
		a.uploadChan <- uploadReq
		a.mutex.Lock()
		defer a.mutex.Unlock()
		a.uploadProgressMap[localpath] = uploadReq
	}()
	return nil
}

func (a *Allocation) RepairRequired(remotepath string) (zboxutil.Uint128, bool, *fileref.FileRef, error) {
	if !a.isInitialized() {
		return zboxutil.Uint128{}, false, nil, notInitialized
	}

	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.blobbers = a.Blobbers
	listReq.fullconsensus = float32(a.DataShards + a.ParityShards)
	listReq.consensusThresh = 100 / listReq.fullconsensus
	listReq.ctx = a.ctx
	listReq.remotefilepath = remotepath
	found, fileRef, _ := listReq.getFileConsensusFromBlobbers()
	if fileRef == nil {
		return found, false, fileRef, fmt.Errorf("File not found for the given remotepath")
	}

	uploadMask := zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)

	return found, !found.Equals(uploadMask), fileRef, nil
}

func (a *Allocation) DownloadFile(localPath string, remotePath string, status StatusCallback) error {
	return a.downloadFile(localPath, remotePath, DOWNLOAD_CONTENT_FULL, 1, 0, numBlockDownloads, status)
}

func (a *Allocation) DownloadFileByBlock(localPath string, remotePath string, startBlock int64, endBlock int64, numBlocks int, status StatusCallback) error {
	return a.downloadFile(localPath, remotePath, DOWNLOAD_CONTENT_FULL, startBlock, endBlock, numBlocks, status)
}

func (a *Allocation) DownloadThumbnail(localPath string, remotePath string, status StatusCallback) error {
	return a.downloadFile(localPath, remotePath, DOWNLOAD_CONTENT_THUMB, 1, 0, numBlockDownloads, status)
}

func (a *Allocation) downloadFile(localPath string, remotePath string, contentMode string,
	startBlock int64, endBlock int64, numBlocks int,
	status StatusCallback) error {
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
	downloadReq.allocationTx = a.Tx
	downloadReq.ctx = a.ctx
	downloadReq.localpath = localPath
	downloadReq.remotefilepath = remotePath
	downloadReq.statusCallback = status
	downloadReq.downloadMask = zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
	downloadReq.blobbers = a.Blobbers
	downloadReq.datashards = a.DataShards
	downloadReq.parityshards = a.ParityShards
	downloadReq.startBlock = startBlock - 1
	downloadReq.endBlock = endBlock
	downloadReq.numBlocks = int64(numBlocks)
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
	listReq.allocationTx = a.Tx
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
	consensusThresh := (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	fullconsensus := float32(a.DataShards + a.ParityShards)
	return a.listDir(path, consensusThresh, fullconsensus)
}

func (a *Allocation) listDir(path string, consensusThresh, fullconsensus float32) (*ListResult, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}
	if len(path) == 0 {
		return nil, common.NewError("invalid_path", "Invalid path for the list")
	}
	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return nil, common.NewError("invalid_path", "Path should be valid and absolute")
	}
	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.blobbers = a.Blobbers
	listReq.consensusThresh = consensusThresh
	listReq.fullconsensus = fullconsensus
	listReq.ctx = a.ctx
	listReq.remotefilepath = path
	ref := listReq.GetListFromBlobbers()
	if ref != nil {
		return ref, nil
	}
	return nil, common.NewError("list_request_failed", "Failed to get list response from the blobbers")
}

func (a *Allocation) GetFileMeta(path string) (*ConsolidatedFileMeta, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}

	result := &ConsolidatedFileMeta{}
	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
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
		result.Size = ref.ActualFileSize
		result.EncryptedKey = ref.EncryptedKey
		result.CommitMetaTxns = ref.CommitMetaTxns
		result.Collaborators = ref.Collaborators
		result.Attributes = ref.Attributes
		result.ActualFileSize = ref.Size
		result.ActualNumBlocks = ref.NumBlocks
		return result, nil
	}
	return nil, common.NewError("file_meta_error", "Error getting the file meta data from blobbers")
}

func (a *Allocation) GetFileMetaFromAuthTicket(authTicket string, lookupHash string) (*ConsolidatedFileMeta, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}

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
	listReq.allocationTx = a.Tx
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
		result.Size = ref.ActualFileSize
		result.CommitMetaTxns = ref.CommitMetaTxns
		result.ActualFileSize = ref.Size
		result.ActualNumBlocks = ref.NumBlocks
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
	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return nil, common.NewError("invalid_path", "Path should be valid and absolute")
	}
	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
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
	consensusThresh := (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	fullconsensus := float32(a.DataShards + a.ParityShards)
	return a.deleteFile(path, consensusThresh, fullconsensus)
}

func (a *Allocation) deleteFile(path string, threshConsensus, fullConsensus float32) error {
	if !a.isInitialized() {
		return notInitialized
	}

	if len(path) == 0 {
		return common.NewError("invalid_path", "Invalid path for the list")
	}
	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return common.NewError("invalid_path", "Path should be valid and absolute")
	}

	req := &DeleteRequest{}
	req.blobbers = a.Blobbers
	req.allocationID = a.ID
	req.allocationTx = a.Tx
	req.consensusThresh = threshConsensus
	req.fullconsensus = fullConsensus
	req.ctx = a.ctx
	req.remotefilepath = path
	req.deleteMask = 0
	req.listMask = 0
	req.connectionID = zboxutil.NewConnectionId()
	return req.ProcessDelete()
}

func (a *Allocation) RenameObject(path string, destName string) error {
	if !a.isInitialized() {
		return notInitialized
	}

	if len(path) == 0 {
		return common.NewError("invalid_path", "Invalid path for the list")
	}
	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return common.NewError("invalid_path", "Path should be valid and absolute")
	}

	req := &RenameRequest{}
	req.blobbers = a.Blobbers
	req.allocationID = a.ID
	req.allocationTx = a.Tx
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

func (a *Allocation) UpdateObjectAttributes(path string,
	attrs fileref.Attributes) (err error) {

	if !a.isInitialized() {
		return notInitialized
	}

	if len(path) == 0 {
		return common.NewError("update_attrs", "Invalid path for the list")
	}

	path = zboxutil.RemoteClean(path)
	var isabs = zboxutil.IsRemoteAbs(path)
	if !isabs {
		return common.NewError("update_attrs",
			"Path should be valid and absolute")
	}

	var attrsb []byte
	if attrsb, err = json.Marshal(attrs); err != nil {
		panic(err)
	}

	var ar AttributesRequest

	ar.blobbers = a.Blobbers
	ar.allocationID = a.ID
	ar.allocationTx = a.Tx
	ar.Attributes = attrs
	ar.attributes = string(attrsb)
	ar.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	ar.fullconsensus = float32(a.DataShards + a.ParityShards)
	ar.ctx = a.ctx
	ar.remotefilepath = path
	ar.attributesMask = 0
	ar.connectionID = zboxutil.NewConnectionId()

	return ar.ProcessAttributes()
}

func (a *Allocation) MoveObject(path string, destPath string) error {
	err := a.CopyObject(path, destPath)
	if err != nil {
		return err
	}
	return a.DeleteFile(path)
}

func (a *Allocation) CopyObject(path string, destPath string) error {
	if !a.isInitialized() {
		return notInitialized
	}

	if len(path) == 0 || len(destPath) == 0 {
		return common.NewError("invalid_path", "Invalid path for copy")
	}
	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return common.NewError("invalid_path", "Path should be valid and absolute")
	}

	req := &CopyRequest{}
	req.blobbers = a.Blobbers
	req.allocationID = a.ID
	req.allocationTx = a.Tx
	req.destPath = destPath
	req.consensusThresh = (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards)
	req.fullconsensus = float32(a.DataShards + a.ParityShards)
	req.ctx = a.ctx
	req.remotefilepath = path
	req.copyMask = 0
	req.connectionID = zboxutil.NewConnectionId()
	return req.ProcessCopy()
}

func (a *Allocation) GetAuthTicketForShare(path string, filename string, referenceType string, refereeClientID string) (string, error) {
	return a.GetAuthTicket(path, filename, referenceType, refereeClientID, "")
}

func (a *Allocation) GetAuthTicket(path string, filename string, referenceType string, refereeClientID string, refereeEncryptionPublicKey string) (string, error) {
	if !a.isInitialized() {
		return "", notInitialized
	}
	if len(path) == 0 {
		return "", common.NewError("invalid_path", "Invalid path for the list")
	}
	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return "", common.NewError("invalid_path", "Path should be valid and absolute")
	}

	shareReq := &ShareRequest{}
	shareReq.allocationID = a.ID
	shareReq.allocationTx = a.Tx
	shareReq.blobbers = a.Blobbers
	shareReq.ctx = a.ctx
	shareReq.remotefilepath = path
	shareReq.remotefilename = filename
	if referenceType == fileref.DIRECTORY {
		shareReq.refType = fileref.DIRECTORY
	} else {
		shareReq.refType = fileref.FILE
	}
	if len(refereeEncryptionPublicKey) > 0 {
		authTicket, err := shareReq.GetAuthTicketForEncryptedFile(refereeClientID, refereeEncryptionPublicKey)
		if err != nil {
			return "", err
		}
		return authTicket, nil

	}
	authTicket, err := shareReq.GetAuthTicket(refereeClientID)
	if err != nil {
		return "", err
	}
	return authTicket, nil
}

func (a *Allocation) CancelUpload(localpath string) error {
	if uploadReq, ok := a.uploadProgressMap[localpath]; ok {
		uploadReq.isUploadCanceled = true
		return nil
	}
	return common.NewError("local_path_not_found", "Invalid path. No upload in progress for the path "+localpath)
}

func (a *Allocation) CancelDownload(remotepath string) error {
	if downloadReq, ok := a.downloadProgressMap[remotepath]; ok {
		downloadReq.isDownloadCanceled = true
		return nil
	}
	return common.NewError("remote_path_not_found", "Invalid path. No download in progress for the path "+remotepath)
}

func (a *Allocation) DownloadThumbnailFromAuthTicket(localPath string,
	authTicket string, remoteLookupHash string, remoteFilename string,
	rxPay bool, status StatusCallback) error {

	return a.downloadFromAuthTicket(localPath, authTicket, remoteLookupHash,
		1, 0, numBlockDownloads, remoteFilename, DOWNLOAD_CONTENT_THUMB,
		rxPay, status)
}

func (a *Allocation) DownloadFromAuthTicket(localPath string, authTicket string,
	remoteLookupHash string, remoteFilename string, rxPay bool,
	status StatusCallback) error {

	return a.downloadFromAuthTicket(localPath, authTicket, remoteLookupHash,
		1, 0, numBlockDownloads, remoteFilename, DOWNLOAD_CONTENT_FULL,
		rxPay, status)
}

func (a *Allocation) DownloadFromAuthTicketByBlocks(localPath string,
	authTicket string, startBlock int64, endBlock int64, numBlocks int,
	remoteLookupHash string, remoteFilename string, rxPay bool,
	status StatusCallback) error {

	return a.downloadFromAuthTicket(localPath, authTicket, remoteLookupHash,
		startBlock, endBlock, numBlocks, remoteFilename, DOWNLOAD_CONTENT_FULL,
		rxPay, status)
}

func (a *Allocation) downloadFromAuthTicket(localPath string, authTicket string,
	remoteLookupHash string, startBlock int64, endBlock int64, numBlocks int,
	remoteFilename string, contentMode string, rxPay bool,
	status StatusCallback) error {

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
	downloadReq.allocationTx = a.Tx
	downloadReq.ctx = a.ctx
	downloadReq.localpath = localPath
	downloadReq.remotefilepathhash = remoteLookupHash
	downloadReq.authTicket = at
	downloadReq.statusCallback = status
	downloadReq.downloadMask = zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
	downloadReq.blobbers = a.Blobbers
	downloadReq.datashards = a.DataShards
	downloadReq.parityshards = a.ParityShards
	downloadReq.contentMode = contentMode
	downloadReq.startBlock = startBlock - 1
	downloadReq.endBlock = endBlock
	downloadReq.numBlocks = int64(numBlocks)
	downloadReq.rxPay = rxPay
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

func (a *Allocation) CommitMetaTransaction(path, crudOperation, authTicket, lookupHash string, fileMeta *ConsolidatedFileMeta, status StatusCallback) (err error) {
	if !a.isInitialized() {
		return notInitialized
	}

	if fileMeta == nil {
		if len(path) > 0 {
			fileMeta, err = a.GetFileMeta(path)
			if err != nil {
				return err
			}
		} else if len(authTicket) > 0 {
			fileMeta, err = a.GetFileMetaFromAuthTicket(authTicket, lookupHash)
			if err != nil {
				return err
			}
		}
	}

	req := &CommitMetaRequest{
		CommitMetaData: CommitMetaData{
			CrudType: crudOperation,
			MetaData: fileMeta,
		},
		status:    status,
		a:         a,
		authToken: authTicket,
	}
	go req.processCommitMetaRequest()
	return nil
}

func (a *Allocation) StartRepair(localRootPath, pathToRepair string, statusCB StatusCallback) error {
	if !a.isInitialized() {
		return notInitialized
	}

	fullconsensus := float32(a.DataShards + a.ParityShards)
	consensusThresh := 100 / fullconsensus
	listDir, err := a.listDir(pathToRepair, consensusThresh, fullconsensus)
	if err != nil {
		return err
	}

	repairReq := &RepairRequest{
		listDir:       listDir,
		localRootPath: localRootPath,
		statusCB:      statusCB,
	}

	repairReq.completedCallback = func() {
		a.mutex.Lock()
		defer a.mutex.Unlock()
		a.repairRequestInProgress = nil
	}

	go func() {
		a.repairChan <- repairReq
		a.mutex.Lock()
		defer a.mutex.Unlock()
		a.repairRequestInProgress = repairReq
	}()
	return nil
}

func (a *Allocation) CancelRepair() error {
	if a.repairRequestInProgress != nil {
		a.repairRequestInProgress.isRepairCanceled = true
		return nil
	}
	return common.NewError("invalid_cancel_repair_request", "No repair in progress for the allocation")
}

type CommitFolderData struct {
	OpType    string
	PreValue  string
	CurrValue string
}

type CommitFolderResponse struct {
	TxnID string
	Data  *CommitFolderData
}

func (a *Allocation) CommitFolderChange(operation, preValue, currValue string) (string, error) {
	if !a.isInitialized() {
		return "", notInitialized
	}

	data := &CommitFolderData{
		OpType:    operation,
		PreValue:  preValue,
		CurrValue: currValue,
	}

	commitFolderDataBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	commitFolderDataString := string(commitFolderDataBytes)

	txn := transaction.NewTransactionEntity(client.GetClientID(), blockchain.GetChainID(), client.GetClientPublicKey())
	txn.TransactionData = commitFolderDataString
	txn.TransactionType = transaction.TxnTypeData
	err = txn.ComputeHashAndSign(client.Sign)
	if err != nil {
		return "", err
	}

	transaction.SendTransactionSync(txn, blockchain.GetMiners())
	querySleepTime := time.Duration(blockchain.GetQuerySleepTime()) * time.Second
	time.Sleep(querySleepTime)
	retries := 0
	var t *transaction.Transaction
	for retries < blockchain.GetMaxTxnQuery() {
		t, err = transaction.VerifyTransaction(txn.Hash, blockchain.GetSharders())
		if err == nil {
			break
		}
		retries++
		time.Sleep(querySleepTime)
	}

	if err != nil {
		Logger.Error("Error verifying the commit transaction", err.Error(), txn.Hash)
		return "", err
	}
	if t == nil {
		err = common.NewError("transaction_validation_failed", "Failed to get the transaction confirmation")
		return "", err
	}

	commitFolderResponse := &CommitFolderResponse{
		TxnID: t.Hash,
		Data:  data,
	}
	commitFolderReponseBytes, _ := json.Marshal(commitFolderResponse)

	commitFolderResponseString := string(commitFolderReponseBytes)
	return commitFolderResponseString, nil
}

func (a *Allocation) AddCollaborator(filePath, collaboratorID string) error {
	if !a.isInitialized() {
		return notInitialized
	}

	req := &CollaboratorRequest{
		path:           filePath,
		collaboratorID: collaboratorID,
		a:              a,
	}

	if req.UpdateCollaboratorToBlobbers() {
		return nil
	}
	return common.NewError("add_collaborator_failed", "Failed to add collaborator on all blobbers.")
}

func (a *Allocation) RemoveCollaborator(filePath, collaboratorID string) error {
	if !a.isInitialized() {
		return notInitialized
	}

	req := &CollaboratorRequest{
		path:           filePath,
		collaboratorID: collaboratorID,
		a:              a,
	}

	if req.RemoveCollaboratorFromBlobbers() {
		return nil
	}
	return common.NewError("remove_collaborator_failed", "Failed to remove collaborator on all blobbers.")
}

// For sync app
const (
	Upload      = "Upload"
	Download    = "Download"
	Update      = "Update"
	Delete      = "Delete"
	Conflict    = "Conflict"
	LocalDelete = "LocalDelete"
)

type fileInfo struct {
	Size int64  `json:"size"`
	Hash string `json:"hash"`
	Type string `json:"type"`
}

type FileDiff struct {
	Op   string `json:"operation"`
	Path string `json:"path"`
	Type string `json:"type"`
}

func (a *Allocation) getRemoteFilesAndDirs(dirList []string, fMap map[string]fileInfo, exclMap map[string]int) ([]string, error) {
	childDirList := make([]string, 0)
	for _, dir := range dirList {
		ref, err := a.ListDir(dir)
		if err != nil {
			return []string{}, err
		}
		for _, child := range ref.Children {
			if _, ok := exclMap[child.Path]; ok {
				continue
			}
			fMap[child.Path] = fileInfo{Size: child.Size, Hash: child.Hash, Type: child.Type}
			if child.Type == fileref.DIRECTORY {
				childDirList = append(childDirList, child.Path)
			}
		}
	}
	return childDirList, nil
}

func (a *Allocation) GetRemoteFileMap(exclMap map[string]int) (map[string]fileInfo, error) {
	// 1. Iteratively get dir and files seperately till no more dirs left
	remoteList := make(map[string]fileInfo)
	dirs := []string{"/"}
	var err error
	for {
		dirs, err = a.getRemoteFilesAndDirs(dirs, remoteList, exclMap)
		if err != nil {
			Logger.Error(err.Error())
			break
		}
		if len(dirs) == 0 {
			break
		}
	}
	Logger.Debug("Remote List: ", remoteList)
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

func getRemoteExcludeMap(exclPath []string) map[string]int {
	exclMap := make(map[string]int)
	for idx, path := range exclPath {
		exclMap[strings.TrimRight(path, "/")] = idx
	}
	return exclMap
}

func addLocalFileList(root string, fMap map[string]fileInfo, dirList *[]string, filter map[string]bool, exclMap map[string]int) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			Logger.Error("Local file list error for path", path, err.Error())
			return err
		}
		// Filter out
		if _, ok := filter[info.Name()]; ok {
			return nil
		}
		lPath, err := filepath.Rel(root, path)
		if err != nil {
			Logger.Error("getting relative path failed", err)
			return err
		}
		lPath = "/" + lPath
		// Exclude
		if _, ok := exclMap[lPath]; ok {
			return nil
		}
		// Add to list
		if info.IsDir() {
			*dirList = append(*dirList, lPath)
		} else {
			fMap[lPath] = fileInfo{Size: info.Size(), Hash: calcFileHash(path), Type: fileref.FILE}
		}
		return nil
	}
}

func getLocalFileMap(rootPath string, filters []string, exclMap map[string]int) (map[string]fileInfo, error) {
	localMap := make(map[string]fileInfo)
	var dirList []string
	filterMap := make(map[string]bool)
	for _, f := range filters {
		filterMap[f] = true
	}
	err := filepath.Walk(rootPath, addLocalFileList(rootPath, localMap, &dirList, filterMap, exclMap))
	if err != nil {
		return nil, err
	}
	// Add the dirs at the end of the list for dir deletion after all file deletion
	for _, d := range dirList {
		localMap[d] = fileInfo{Type: fileref.DIRECTORY}
	}
	Logger.Debug("Local List: ", localMap)
	return localMap, err
}

func isParentFolderExists(lFDiff []FileDiff, path string) bool {
	subdirs := strings.Split(path, "/")
	p := "/"
	for _, dir := range subdirs {
		p = filepath.Join(p, dir)
		for _, f := range lFDiff {
			if f.Path == p {
				return true
			}
		}
	}
	return false
}

func findDelta(rMap map[string]fileInfo, lMap map[string]fileInfo, prevMap map[string]fileInfo, localRootPath string) []FileDiff {
	var lFDiff []FileDiff

	// Create a remote hash map and find modifications
	rMod := make(map[string]fileInfo)
	for rFile, rInfo := range rMap {
		if pm, ok := prevMap[rFile]; ok {
			// Remote file existed in previous sync also
			if pm.Hash != rInfo.Hash {
				// File modified in remote
				rMod[rFile] = rInfo
			}
		}
	}

	// Create a local hash map and find modification
	lMod := make(map[string]fileInfo)
	for lFile, lInfo := range lMap {
		if pm, ok := prevMap[lFile]; ok {
			// Local file existed in previous sync also
			if pm.Hash != lInfo.Hash {
				// File modified in local
				lMod[lFile] = lInfo
			}
		}
	}

	// Iterate remote list and get diff
	rDelMap := make(map[string]string)
	for rPath, _ := range rMap {
		op := Download
		bRemoteModified := false
		bLocalModified := false
		if _, ok := rMod[rPath]; ok {
			bRemoteModified = true
		}
		if _, ok := lMod[rPath]; ok {
			bLocalModified = true
			delete(lMap, rPath)
		}
		if bRemoteModified && bLocalModified {
			op = Conflict
		} else if bLocalModified {
			op = Update
		} else if bRemoteModified {
			// Remote modified, local not change
			op = Download
		} else if _, ok := lMap[rPath]; ok {
			// No conflicts and file exists locally
			delete(lMap, rPath)
			continue
		} else if _, ok := prevMap[rPath]; ok {
			op = Delete
			// Remote allows delete directory skip individual file deletion
			rDelMap[rPath] = "d"
			rDir, _ := filepath.Split(rPath)
			rDir = strings.TrimRight(rDir, "/")
			if _, ok := rDelMap[rDir]; ok {
				continue
			}
		}
		lFDiff = append(lFDiff, FileDiff{Path: rPath, Op: op, Type: rMap[rPath].Type})
	}

	// Upload all local files
	for lPath, _ := range lMap {
		op := Upload
		if _, ok := lMod[lPath]; ok { // duplicate check for local modified
			op = Update
		} else if _, ok := prevMap[lPath]; ok {
			op = LocalDelete
		}
		if op != LocalDelete {
			// Skip if it is a directory
			lAbsPath := filepath.Join(localRootPath, lPath)
			fInfo, err := os.Stat(lAbsPath)
			if err != nil {
				continue
			}
			if fInfo.IsDir() {
				continue
			}
		}
		lFDiff = append(lFDiff, FileDiff{Path: lPath, Op: op, Type: lMap[lPath].Type})
	}

	// If there are differences, remove childs if the parent folder is deleted
	if len(lFDiff) > 0 {
		sort.SliceStable(lFDiff, func(i, j int) bool { return lFDiff[i].Path < lFDiff[j].Path })
		Logger.Debug("Sorted diff: ", lFDiff)
		var newlFDiff []FileDiff
		for _, f := range lFDiff {
			if f.Op == LocalDelete || f.Op == Delete {
				if isParentFolderExists(newlFDiff, f.Path) == false {
					newlFDiff = append(newlFDiff, f)
				}
			} else {
				// Add only files for other Op
				if f.Type == fileref.FILE {
					newlFDiff = append(newlFDiff, f)
				}
			}
		}
		return newlFDiff
	}
	return lFDiff
}

func (a *Allocation) GetAllocationDiff(lastSyncCachePath string, localRootPath string, localFileFilters []string, remoteExcludePath []string) ([]FileDiff, error) {
	var lFdiff []FileDiff
	prevRemoteFileMap := make(map[string]fileInfo)
	// 1. Validate localSycnCachePath
	if len(lastSyncCachePath) > 0 {
		// Validate cache path
		fileInfo, err := os.Stat(lastSyncCachePath)
		if err == nil {
			if fileInfo.IsDir() {
				return lFdiff, fmt.Errorf("invalid file cache. %v", err)
			}
			content, err := ioutil.ReadFile(lastSyncCachePath)
			if err != nil {
				return lFdiff, fmt.Errorf("can't read cache file.")
			}
			err = json.Unmarshal(content, &prevRemoteFileMap)
			if err != nil {
				return lFdiff, fmt.Errorf("invalid cache content.")
			}
		}
	}

	// 2. Build a map for exclude path
	exclMap := getRemoteExcludeMap(remoteExcludePath)

	// 3. Get flat file list from remote
	remoteFileMap, err := a.GetRemoteFileMap(exclMap)
	if err != nil {
		return lFdiff, fmt.Errorf("error getting list dir from remote. %v", err)
	}

	// 4. Get flat file list on the local filesystem
	localRootPath = strings.TrimRight(localRootPath, "/")
	localFileList, err := getLocalFileMap(localRootPath, localFileFilters, exclMap)
	if err != nil {
		return lFdiff, fmt.Errorf("error getting list dir from local. %v", err)
	}

	// 5. Get the file diff with operation
	lFdiff = findDelta(remoteFileMap, localFileList, prevRemoteFileMap, localRootPath)
	Logger.Debug("Diff: ", lFdiff)
	return lFdiff, nil
}

// SaveRemoteSnapShot - Saves the remote current information to the given file
// This file can be passed to GetAllocationDiff to exactly find the previous sync state to current.
func (a *Allocation) SaveRemoteSnapshot(pathToSave string, remoteExcludePath []string) error {
	bIsFileExists := false
	// Validate path
	fileInfo, err := os.Stat(pathToSave)
	if err == nil {
		if fileInfo.IsDir() {
			return fmt.Errorf("invalid file path to save. %v", err)
		}
		bIsFileExists = true
	}

	// Get flat file list from remote
	exclMap := getRemoteExcludeMap(remoteExcludePath)
	remoteFileList, err := a.GetRemoteFileMap(exclMap)
	if err != nil {
		return fmt.Errorf("error getting list dir from remote. %v", err)
	}

	// Now we got the list from remote, delete the file if exists
	if bIsFileExists {
		err = os.Remove(pathToSave)
		if err != nil {
			return fmt.Errorf("error deleting previous cache. %v", err)
		}
	}
	by, err := json.Marshal(remoteFileList)
	if err != nil {
		return fmt.Errorf("failed to convert JSON. %v", err)
	}
	err = ioutil.WriteFile(pathToSave, by, 0644)
	if err != nil {
		return fmt.Errorf("error saving file. %v", err)
	}
	// Successfully saved
	return nil
}
