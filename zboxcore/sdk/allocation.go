package sdk

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/0chain/errors"
	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/mitchellh/go-homedir"
)

var (
	noBLOBBERS     = errors.New("", "No Blobbers set in this allocation")
	notInitialized = errors.New("sdk_not_initialized", "Please call InitStorageSDK Init and use GetAllocation to get the allocation object")
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

const (
	CanUploadMask = uint16(1)  // 0000 0001
	CanDeleteMask = uint16(2)  // 0000 0010
	CanUpdateMask = uint16(4)  // 0000 0100
	CanMoveMask   = uint16(8)  // 0000 1000
	CanCopyMask   = uint16(16) // 0001 0000
	CanRenameMask = uint16(32) // 0010 0000
)

// Expected success rate is calculated (NumDataShards)*100/(NumDataShards+NumParityShards)
// Additional success percentage on top of expected success rate
const additionalSuccessRate = (10)

var GetFileInfo = func(localpath string) (os.FileInfo, error) {
	return sys.Files.Stat(localpath)
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
	Min uint64 `json:"min"`
	Max uint64 `json:"max"`
}

// IsValid price range.
func (pr *PriceRange) IsValid() bool {
	return pr.Min <= pr.Max
}

// Terms represents Blobber terms. A Blobber can update its terms,
// but any existing offer will use terms of offer signing time.
type Terms struct {
	ReadPrice        common.Balance `json:"read_price"`  // tokens / GB
	WritePrice       common.Balance `json:"write_price"` // tokens / GB
	MinLockDemand    float64        `json:"min_lock_demand"`
	MaxOfferDuration time.Duration  `json:"max_offer_duration"`
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
	WritePool      common.Balance            `json:"write_pool"`
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
	FileOptions             uint16           `json:"file_options"`
	ThirdPartyExtendable    bool             `json:"third_party_extendable"`
	Curators                []string         `json:"curators"`

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

	// conseususes
	consensusThreshold int
	fullconsensus      int
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
	a.uploadChan = make(chan *UploadRequest, 10)
	a.downloadChan = make(chan *DownloadRequest, 10)
	a.repairChan = make(chan *RepairRequest, 1)
	a.ctx, a.ctxCancelF = context.WithCancel(context.Background())
	a.uploadProgressMap = make(map[string]*UploadRequest)
	a.downloadProgressMap = make(map[string]*DownloadRequest)
	a.mutex = &sync.Mutex{}
	a.fullconsensus, a.consensusThreshold = a.getConsensuses()
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
	for {
		select {
		case <-ctx.Done():
			l.Logger.Info("Upload cancelled by the parent")
			return
		case uploadReq := <-a.uploadChan:

			l.Logger.Info(fmt.Sprintf("received a upload request for %v %v\n", uploadReq.filepath, uploadReq.remotefilepath))
			go uploadReq.processUpload(ctx, a)
		case downloadReq := <-a.downloadChan:

			l.Logger.Info(fmt.Sprintf("received a download request for %v\n", downloadReq.remotefilepath))
			go downloadReq.processDownload(ctx)
		case repairReq := <-a.repairChan:

			l.Logger.Info(fmt.Sprintf("received a repair request for %v\n", repairReq.listDir.Path))
			go repairReq.processRepair(ctx, a)
		}
	}
}

func (a *Allocation) CanUpload() bool {
	return (a.FileOptions & CanUploadMask) > 0
}

func (a *Allocation) CanDelete() bool {
	return (a.FileOptions & CanDeleteMask) > 0
}

func (a *Allocation) CanUpdate() bool {
	return (a.FileOptions & CanUpdateMask) > 0
}

func (a *Allocation) CanMove() bool {
	return (a.FileOptions & CanMoveMask) > 0
}

func (a *Allocation) CanCopy() bool {
	return (a.FileOptions & CanCopyMask) > 0
}

func (a *Allocation) CanRename() bool {
	return (a.FileOptions & CanRenameMask) > 0
}

// UpdateFile [Deprecated]please use CreateChunkedUpload
func (a *Allocation) UpdateFile(workdir, localpath string, remotepath string,
	status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, true, false, "", false)
}

// UploadFile [Deprecated]please use CreateChunkedUpload
func (a *Allocation) UploadFile(workdir, localpath string, remotepath string,
	status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, false, false, "", false)
}

func (a *Allocation) CreateDir(remotePath string) error {
	if !a.isInitialized() {
		return notInitialized
	}

	if remotePath == "" {
		return errors.New("invalid_name", "Invalid name for dir")
	}

	if !path.IsAbs(remotePath) {
		return errors.New("invalid_path", "Path is not absolute")
	}

	remotePath = zboxutil.RemoteClean(remotePath)
	req := DirRequest{
		allocationID: a.ID,
		allocationTx: a.Tx,
		blobbers:     a.Blobbers,
		mu:           &sync.Mutex{},
		dirMask:      zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1),
		connectionID: zboxutil.NewConnectionId(),
		ctx:          a.ctx,
		remotePath:   remotePath,
		wg:           &sync.WaitGroup{},
		Consensus: Consensus{
			consensusThresh: a.consensusThreshold,
			fullconsensus:   a.fullconsensus,
		},
	}

	err := req.ProcessDir(a)
	return err
}

func (a *Allocation) RepairFile(localpath string, remotepath string,
	status StatusCallback) error {

	idr, _ := homedir.Dir()
	return a.StartChunkedUpload(idr, localpath, remotepath, status, false, true,
		"", false)
}

// UpdateFileWithThumbnail [Deprecated]please use CreateChunkedUpload
func (a *Allocation) UpdateFileWithThumbnail(workdir, localpath string, remotepath string,
	thumbnailpath string, status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, true, false,
		thumbnailpath, false)
}

// UploadFileWithThumbnail [Deprecated]please use CreateChunkedUpload
func (a *Allocation) UploadFileWithThumbnail(workdir string, localpath string,
	remotepath string, thumbnailpath string,
	status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, false, false,
		thumbnailpath, false)
}

// EncryptAndUpdateFile [Deprecated]please use CreateChunkedUpload
func (a *Allocation) EncryptAndUpdateFile(workdir string, localpath string, remotepath string,
	status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, true, false, "", true)
}

// EncryptAndUploadFile [Deprecated]please use CreateChunkedUpload
func (a *Allocation) EncryptAndUploadFile(workdir string, localpath string, remotepath string,
	status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, false, false, "", true)
}

// EncryptAndUpdateFileWithThumbnail [Deprecated]please use CreateChunkedUpload
func (a *Allocation) EncryptAndUpdateFileWithThumbnail(workdir string, localpath string,
	remotepath string, thumbnailpath string, status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, true, false,
		thumbnailpath, true)
}

// EncryptAndUploadFileWithThumbnail [Deprecated]please use CreateChunkedUpload
func (a *Allocation) EncryptAndUploadFileWithThumbnail(
	workdir string,
	localpath string,
	remotepath string,
	thumbnailpath string,

	status StatusCallback,
) error {

	return a.StartChunkedUpload(workdir,
		localpath,
		remotepath,
		status,
		false,
		false,
		thumbnailpath,
		true,
	)
}

func (a *Allocation) StartChunkedUpload(workdir, localPath string,
	remotePath string,
	status StatusCallback,
	isUpdate bool,
	isRepair bool,
	thumbnailPath string,
	encryption bool,

) error {

	if !a.isInitialized() {
		return notInitialized
	}

	if (!isUpdate && !a.CanUpload()) || (isUpdate && !a.CanUpdate()) {
		return constants.ErrFileOptionNotPermitted
	}

	fileReader, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	fileInfo, err := fileReader.Stat()
	if err != nil {
		return err
	}

	mimeType, err := zboxutil.GetFileContentType(fileReader)
	if err != nil {
		return err
	}

	remotePath = zboxutil.RemoteClean(remotePath)
	isabs := zboxutil.IsRemoteAbs(remotePath)
	if !isabs {
		err = thrown.New("invalid_path", "Path should be valid and absolute")
		return err
	}
	remotePath = zboxutil.GetFullRemotePath(localPath, remotePath)

	_, fileName := filepath.Split(remotePath)

	fileMeta := FileMeta{
		Path:       localPath,
		ActualSize: fileInfo.Size(),
		MimeType:   mimeType,
		RemoteName: fileName,
		RemotePath: remotePath,
	}

	options := []ChunkedUploadOption{
		WithEncrypt(encryption),
		WithStatusCallback(status),
	}

	if thumbnailPath != "" {
		buf, err := sys.Files.ReadFile(thumbnailPath)
		if err != nil {
			return err
		}

		options = append(options, WithThumbnail(buf))
	}

	ChunkedUpload, err := CreateChunkedUpload(workdir,
		a, fileMeta, fileReader,
		isUpdate, isRepair,
		options...)
	if err != nil {
		return err
	}

	return ChunkedUpload.Start()
}

// uploadOrUpdateFile [Deprecated]please use CreateChunkedUpload
func (a *Allocation) uploadOrUpdateFile(localpath string,
	remotepath string,
	status StatusCallback,
	isUpdate bool,
	thumbnailpath string,
	encryption bool,
	isRepair bool,

) error {

	if !a.isInitialized() {
		return notInitialized
	}

	fileInfo, err := GetFileInfo(localpath)
	if err != nil {
		return errors.Wrap(err, "Local file error")
	}
	thumbnailSize := int64(0)
	if len(thumbnailpath) > 0 {
		fileInfo, err := sys.Files.Stat(thumbnailpath)
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
		return errors.New("invalid_path", "Path should be valid and absolute")
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
	uploadReq.isUpdate = isUpdate
	uploadReq.isRepair = isRepair
	uploadReq.connectionID = zboxutil.NewConnectionId()
	uploadReq.statusCallback = status
	uploadReq.datashards = a.DataShards
	uploadReq.parityshards = a.ParityShards
	uploadReq.setUploadMask(len(a.Blobbers))
	uploadReq.fullconsensus = a.fullconsensus
	uploadReq.consensusThresh = a.consensusThreshold
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
			return errors.New("", "Repair not required")
		}

		file, _ := ioutil.ReadFile(localpath)
		hash := sha256.New()
		hash.Write(file)
		contentHash := hex.EncodeToString(hash.Sum(nil))
		print(contentHash)
		if contentHash != fileRef.ActualFileHash {
			return errors.New("", "Content hash doesn't match")
		}

		uploadReq.filemeta.Hash = fileRef.ActualFileHash
		uploadReq.uploadMask = found.Not().And(uploadReq.uploadMask)
		uploadReq.fullconsensus = uploadReq.uploadMask.Add64(1).TrailingZeros()
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
	listReq.fullconsensus = a.fullconsensus
	listReq.consensusThresh = a.consensusThreshold
	listReq.ctx = a.ctx
	listReq.remotefilepath = remotepath
	found, fileRef, _ := listReq.getFileConsensusFromBlobbers()
	if fileRef == nil {
		return found, false, fileRef, errors.New("", "File not found for the given remotepath")
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
	if stat, err := sys.Files.Stat(localPath); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("Local path is not a directory '%s'", localPath)
		}
		localPath = strings.TrimRight(localPath, "/")
		_, rFile := filepath.Split(remotePath)
		localPath = fmt.Sprintf("%s/%s", localPath, rFile)
		if _, err := sys.Files.Stat(localPath); err == nil {
			return fmt.Errorf("Local file already exists '%s'", localPath)
		}
	}
	dir, _ := filepath.Split(localPath)
	if dir != "" {
		if err := sys.Files.MkdirAll(dir, 0744); err != nil {
			return err
		}
	}

	if len(a.Blobbers) == 0 {
		return noBLOBBERS
	}

	downloadReq := &DownloadRequest{}
	downloadReq.maskMu = &sync.Mutex{}
	downloadReq.allocationID = a.ID
	downloadReq.allocationTx = a.Tx
	downloadReq.allocOwnerID = a.Owner
	downloadReq.ctx, downloadReq.ctxCncl = context.WithCancel(a.ctx)
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
	downloadReq.fullconsensus = a.fullconsensus
	downloadReq.consensusThresh = a.consensusThreshold
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
		return nil, errors.New("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	at := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return nil, errors.New("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	if len(at.FilePathHash) == 0 || len(lookupHash) == 0 {
		return nil, errors.New("invalid_path", "Invalid path for the list")
	}

	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.blobbers = a.Blobbers
	listReq.fullconsensus = a.fullconsensus
	listReq.consensusThresh = a.consensusThreshold
	listReq.ctx = a.ctx
	listReq.remotefilepathhash = lookupHash
	listReq.authToken = at
	ref, err := listReq.GetListFromBlobbers()

	if err != nil {
		return nil, err
	}

	if ref != nil {
		return ref, nil
	}
	return nil, errors.New("list_request_failed", "Failed to get list response from the blobbers")
}

func (a *Allocation) ListDir(path string) (*ListResult, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}
	if len(path) == 0 {
		return nil, errors.New("invalid_path", "Invalid path for the list")
	}
	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return nil, errors.New("invalid_path", "Path should be valid and absolute")
	}
	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.blobbers = a.Blobbers
	listReq.fullconsensus = a.fullconsensus
	listReq.consensusThresh = a.consensusThreshold
	listReq.ctx = a.ctx
	listReq.remotefilepath = path
	ref, err := listReq.GetListFromBlobbers()
	if err != nil {
		return nil, err
	}

	if ref != nil {
		return ref, nil
	}
	return nil, errors.New("list_request_failed", "Failed to get list response from the blobbers")
}

// This function will retrieve paginated objectTree and will handle concensus; Required tree should be made in application side.
// TODO use allocation context
func (a *Allocation) GetRefs(path, offsetPath, updatedDate, offsetDate, fileType, refType string, level, pageLimit int) (*ObjectTreeResult, error) {
	if len(path) == 0 || !zboxutil.IsRemoteAbs(path) {
		return nil, errors.New("invalid_path", fmt.Sprintf("Absolute path required. Path provided: %v", path))
	}
	if !a.isInitialized() {
		return nil, notInitialized
	}
	oTreeReq := &ObjectTreeRequest{
		allocationID:   a.ID,
		allocationTx:   a.Tx,
		blobbers:       a.Blobbers,
		remotefilepath: path,
		pageLimit:      pageLimit,
		level:          level,
		offsetPath:     offsetPath,
		updatedDate:    updatedDate,
		offsetDate:     offsetDate,
		fileType:       fileType,
		refType:        refType,
		wg:             &sync.WaitGroup{},
		ctx:            a.ctx,
	}
	oTreeReq.fullconsensus = a.fullconsensus
	oTreeReq.consensusThresh = a.consensusThreshold

	return oTreeReq.GetRefs()
}

func (a *Allocation) GetRecentlyAddedRefs(page int, fromDate int64, pageLimit int) (*RecentlyAddedRefResult, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}

	if page < 1 {
		return nil, errors.New("invalid_params",
			fmt.Sprintf("page value should be greater than or equal to 1."+
				"Got page: %d", page))
	}

	offset := int64(page-1) * int64(pageLimit)
	req := &RecentlyAddedRefRequest{
		allocationID: a.ID,
		allocationTx: a.Tx,
		blobbers:     a.Blobbers,
		offset:       offset,
		fromDate:     fromDate,
		ctx:          a.ctx,
		wg:           &sync.WaitGroup{},
		pageLimit:    pageLimit,
		Consensus: Consensus{
			fullconsensus:   a.fullconsensus,
			consensusThresh: a.consensusThreshold,
		},
	}
	return req.GetRecentlyAddedRefs()
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
	listReq.fullconsensus = a.fullconsensus
	listReq.consensusThresh = a.consensusThreshold
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
		result.ActualFileSize = ref.Size
		result.ActualNumBlocks = ref.NumBlocks
		return result, nil
	}
	return nil, errors.New("file_meta_error", "Error getting the file meta data from blobbers")
}

func (a *Allocation) GetFileMetaFromAuthTicket(authTicket string, lookupHash string) (*ConsolidatedFileMeta, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}

	result := &ConsolidatedFileMeta{}
	sEnc, err := base64.StdEncoding.DecodeString(authTicket)
	if err != nil {
		return nil, errors.New("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	at := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return nil, errors.New("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	if len(at.FilePathHash) == 0 || len(lookupHash) == 0 {
		return nil, errors.New("invalid_path", "Invalid path for the list")
	}

	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.blobbers = a.Blobbers
	listReq.fullconsensus = a.fullconsensus
	listReq.consensusThresh = a.consensusThreshold
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
	return nil, errors.New("file_meta_error", "Error getting the file meta data from blobbers")
}

func (a *Allocation) GetFileStats(path string) (map[string]*FileStats, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}
	if len(path) == 0 {
		return nil, errors.New("invalid_path", "Invalid path for the list")
	}
	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return nil, errors.New("invalid_path", "Path should be valid and absolute")
	}
	listReq := &ListRequest{}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.blobbers = a.Blobbers
	listReq.fullconsensus = a.fullconsensus
	listReq.consensusThresh = a.consensusThreshold
	listReq.ctx = a.ctx
	listReq.remotefilepath = path
	ref := listReq.getFileStatsFromBlobbers()
	if ref != nil {
		return ref, nil
	}
	return nil, errors.New("file_stats_request_failed", "Failed to get file stats response from the blobbers")
}

func (a *Allocation) DeleteFile(path string) error {
	return a.deleteFile(path, a.consensusThreshold, a.fullconsensus)
}

func (a *Allocation) deleteFile(path string, threshConsensus, fullConsensus int) error {
	if !a.isInitialized() {
		return notInitialized
	}

	if !a.CanDelete() {
		return constants.ErrFileOptionNotPermitted
	}

	if len(path) == 0 {
		return errors.New("invalid_path", "Invalid path for the list")
	}
	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return errors.New("invalid_path", "Path should be valid and absolute")
	}

	req := &DeleteRequest{}
	req.allocationObj = a
	req.blobbers = a.Blobbers
	req.allocationID = a.ID
	req.allocationTx = a.Tx
	req.consensus.Init(threshConsensus, fullConsensus)
	req.ctx = a.ctx
	req.remotefilepath = path
	req.connectionID = zboxutil.NewConnectionId()
	req.deleteMask = zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
	req.maskMu = &sync.Mutex{}
	err := req.ProcessDelete()
	return err
}

func (a *Allocation) RenameObject(path string, destName string) error {
	if !a.isInitialized() {
		return notInitialized
	}

	if !a.CanRename() {
		return constants.ErrFileOptionNotPermitted
	}

	if path == "" {
		return errors.New("invalid_path", "Invalid path for the list")
	}

	if path == "/" {
		return errors.New("invalid_operation", "cannot rename root path")
	}

	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return errors.New("invalid_path", "Path should be valid and absolute")
	}

	err := ValidateRemoteFileName(destName)
	if err != nil {
		return err
	}

	req := &RenameRequest{}
	req.allocationObj = a
	req.blobbers = a.Blobbers
	req.allocationID = a.ID
	req.allocationTx = a.Tx
	req.newName = destName
	req.consensus.fullconsensus = a.fullconsensus
	req.consensus.consensusThresh = a.consensusThreshold
	req.ctx = a.ctx
	req.remotefilepath = path
	req.renameMask = zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
	req.maskMU = &sync.Mutex{}
	req.connectionID = zboxutil.NewConnectionId()
	return req.ProcessRename()
}

func (a *Allocation) MoveObject(srcPath string, destPath string) error {
	if !a.isInitialized() {
		return notInitialized
	}

	if !a.CanMove() {
		return constants.ErrFileOptionNotPermitted
	}

	if len(srcPath) == 0 || len(destPath) == 0 {
		return errors.New("invalid_path", "Invalid path for copy")
	}
	srcPath = zboxutil.RemoteClean(srcPath)
	isabs := zboxutil.IsRemoteAbs(srcPath)
	if !isabs {
		return errors.New("invalid_path", "Path should be valid and absolute")
	}

	err := ValidateRemoteFileName(destPath)
	if err != nil {
		return err
	}

	req := &MoveRequest{}
	req.allocationObj = a
	req.blobbers = a.Blobbers
	req.allocationID = a.ID
	req.allocationTx = a.Tx
	if destPath != "/" {
		destPath = strings.TrimSuffix(destPath, "/")
	}
	req.destPath = destPath
	req.fullconsensus = a.fullconsensus
	req.consensusThresh = a.consensusThreshold
	req.ctx = a.ctx
	req.remotefilepath = srcPath
	req.moveMask = zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
	req.maskMU = &sync.Mutex{}
	req.connectionID = zboxutil.NewConnectionId()
	return req.ProcessMove()
}

func (a *Allocation) CopyObject(path string, destPath string) error {
	if !a.isInitialized() {
		return notInitialized
	}

	if !a.CanCopy() {
		return constants.ErrFileOptionNotPermitted
	}

	if len(path) == 0 || len(destPath) == 0 {
		return errors.New("invalid_path", "Invalid path for copy")
	}
	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return errors.New("invalid_path", "Path should be valid and absolute")
	}

	err := ValidateRemoteFileName(destPath)
	if err != nil {
		return err
	}

	req := &CopyRequest{}
	req.allocationObj = a
	req.blobbers = a.Blobbers
	req.allocationID = a.ID
	req.allocationTx = a.Tx
	if destPath != "/" {
		destPath = strings.TrimSuffix(destPath, "/")
	}
	req.destPath = destPath
	req.fullconsensus = a.fullconsensus
	req.consensusThresh = a.consensusThreshold
	req.ctx = a.ctx
	req.remotefilepath = path
	req.copyMask = zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
	req.maskMU = &sync.Mutex{}
	req.connectionID = zboxutil.NewConnectionId()
	return req.ProcessCopy()
}

func (a *Allocation) GetAuthTicketForShare(
	path, filename, referenceType, refereeClientID string) (string, error) {

	now := time.Now()
	return a.GetAuthTicket(path, filename, referenceType, refereeClientID, "", 0, &now)
}

func (a *Allocation) RevokeShare(path string, refereeClientID string) error {
	success := make(chan int, len(a.Blobbers))
	notFound := make(chan int, len(a.Blobbers))
	wg := &sync.WaitGroup{}
	for idx := range a.Blobbers {
		baseUrl := a.Blobbers[idx].Baseurl
		query := &url.Values{}
		query.Add("path", path)
		query.Add("refereeClientID", refereeClientID)

		httpreq, err := zboxutil.NewRevokeShareRequest(baseUrl, a.Tx, query)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := zboxutil.HttpDo(a.ctx, a.ctxCancelF, httpreq, func(resp *http.Response, err error) error {
				if err != nil {
					l.Logger.Error("Revoke share : ", err)
					return err
				}
				defer resp.Body.Close()

				respbody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					l.Logger.Error("Error: Resp ", err)
					return err
				}
				if resp.StatusCode != http.StatusOK {
					l.Logger.Error(baseUrl, " Revoke share error response: ", resp.StatusCode, string(respbody))
					return fmt.Errorf(string(respbody))
				}
				data := map[string]interface{}{}
				err = json.Unmarshal(respbody, &data)
				if err != nil {
					return err
				}
				if data["status"].(float64) == http.StatusNotFound {
					notFound <- 1
				}
				return nil
			})
			if err == nil {
				success <- 1
			}
		}()
	}
	wg.Wait()
	if len(success) == len(a.Blobbers) {
		if len(notFound) == len(a.Blobbers) {
			return errors.New("", "share not found")
		}
		return nil
	}
	return errors.New("", "consensus not reached")
}

func (a *Allocation) GetAuthTicket(path, filename string,
	referenceType, refereeClientID, refereeEncryptionPublicKey string, expiration int64, availableAfter *time.Time) (string, error) {

	if !a.isInitialized() {
		return "", notInitialized
	}

	if path == "" {
		return "", errors.New("invalid_path", "Invalid path for the list")
	}

	path = zboxutil.RemoteClean(path)
	isabs := zboxutil.IsRemoteAbs(path)
	if !isabs {
		return "", errors.New("invalid_path", "Path should be valid and absolute")
	}

	shareReq := &ShareRequest{
		expirationSeconds: expiration,
		allocationID:      a.ID,
		allocationTx:      a.Tx,
		blobbers:          a.Blobbers,
		ctx:               a.ctx,
		remotefilepath:    path,
		remotefilename:    filename,
	}

	if referenceType == fileref.DIRECTORY {
		shareReq.refType = fileref.DIRECTORY
	} else {
		shareReq.refType = fileref.FILE
	}

	aTicket, err := shareReq.getAuthTicket(refereeClientID, refereeEncryptionPublicKey)
	if err != nil {
		return "", err
	}

	atBytes, err := json.Marshal(aTicket)
	if err != nil {
		return "", err
	}

	if err := a.UploadAuthTicketToBlobber(string(atBytes), refereeEncryptionPublicKey, availableAfter); err != nil {
		return "", err
	}

	aTicket.ReEncryptionKey = ""
	if err := aTicket.Sign(); err != nil {
		return "", err
	}

	atBytes, err = json.Marshal(aTicket)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(atBytes), nil
}

func (a *Allocation) UploadAuthTicketToBlobber(authTicket string, clientEncPubKey string, availableAfter *time.Time) error {
	success := make(chan int, len(a.Blobbers))
	wg := &sync.WaitGroup{}
	for idx := range a.Blobbers {
		url := a.Blobbers[idx].Baseurl
		body := new(bytes.Buffer)
		formWriter := multipart.NewWriter(body)
		if err := formWriter.WriteField("encryption_public_key", clientEncPubKey); err != nil {
			return err
		}
		if err := formWriter.WriteField("auth_ticket", authTicket); err != nil {
			return err
		}
		if availableAfter != nil {
			if err := formWriter.WriteField("available_after", strconv.FormatInt(availableAfter.Unix(), 10)); err != nil {
				return err
			}
		}

		if err := formWriter.Close(); err != nil {
			return err
		}
		httpreq, err := zboxutil.NewShareRequest(url, a.Tx, body)
		if err != nil {
			return err
		}
		httpreq.Header.Set("Content-Type", formWriter.FormDataContentType())

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := zboxutil.HttpDo(a.ctx, a.ctxCancelF, httpreq, func(resp *http.Response, err error) error {
				if err != nil {
					l.Logger.Error("Insert share info : ", err)
					return err
				}
				defer resp.Body.Close()

				respbody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					l.Logger.Error("Error: Resp ", err)
					return err
				}
				if resp.StatusCode != http.StatusOK {
					l.Logger.Error(url, " Insert share info error response: ", resp.StatusCode, string(respbody))
					return fmt.Errorf(string(respbody))
				}
				return nil
			})
			if err == nil {
				success <- 1
			}
		}()
	}
	wg.Wait()
	consensus := Consensus{
		consensus:       len(success),
		consensusThresh: a.consensusThreshold,
		fullconsensus:   a.fullconsensus,
	}
	if !consensus.isConsensusOk() {
		return errors.New("", "consensus not reached")
	}
	return nil
}

func (a *Allocation) CancelUpload(localpath string) error {
	if uploadReq, ok := a.uploadProgressMap[localpath]; ok {
		uploadReq.isUploadCanceled = true
		return nil
	}
	return errors.New("local_path_not_found", "Invalid path. No upload in progress for the path "+localpath)
}

func (a *Allocation) CancelDownload(remotepath string) error {
	if downloadReq, ok := a.downloadProgressMap[remotepath]; ok {
		downloadReq.isDownloadCanceled = true
		downloadReq.ctxCncl()
		return nil
	}
	return errors.New("remote_path_not_found", "Invalid path. No download in progress for the path "+remotepath)
}

func (a *Allocation) DownloadThumbnailFromAuthTicket(localPath string,
	authTicket string, remoteLookupHash string, remoteFilename string,
	status StatusCallback) error {

	return a.downloadFromAuthTicket(localPath, authTicket, remoteLookupHash,
		1, 0, numBlockDownloads, remoteFilename, DOWNLOAD_CONTENT_THUMB,
		status)
}

func (a *Allocation) DownloadFromAuthTicket(localPath string, authTicket string,
	remoteLookupHash string, remoteFilename string, status StatusCallback) error {

	return a.downloadFromAuthTicket(localPath, authTicket, remoteLookupHash,
		1, 0, numBlockDownloads, remoteFilename, DOWNLOAD_CONTENT_FULL,
		status)
}

func (a *Allocation) DownloadFromAuthTicketByBlocks(localPath string,
	authTicket string, startBlock int64, endBlock int64, numBlocks int,
	remoteLookupHash string, remoteFilename string,
	status StatusCallback) error {

	return a.downloadFromAuthTicket(localPath, authTicket, remoteLookupHash,
		startBlock, endBlock, numBlocks, remoteFilename, DOWNLOAD_CONTENT_FULL,
		status)
}

func (a *Allocation) downloadFromAuthTicket(localPath string, authTicket string,
	remoteLookupHash string, startBlock int64, endBlock int64, numBlocks int,
	remoteFilename string, contentMode string,
	status StatusCallback) error {

	if !a.isInitialized() {
		return notInitialized
	}
	sEnc, err := base64.StdEncoding.DecodeString(authTicket)
	if err != nil {
		return errors.New("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	at := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return errors.New("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}

	if stat, err := sys.Files.Stat(localPath); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("Local path is not a directory '%s'", localPath)
		}
		localPath = strings.TrimRight(localPath, "/")
		_, rFile := filepath.Split(remoteFilename)
		localPath = fmt.Sprintf("%s/%s", localPath, rFile)
		if _, err := sys.Files.Stat(localPath); err == nil {
			return fmt.Errorf("Local file already exists '%s'", localPath)
		}
	}
	if len(a.Blobbers) == 0 {
		return noBLOBBERS
	}

	downloadReq := &DownloadRequest{}
	downloadReq.maskMu = &sync.Mutex{}
	downloadReq.allocationID = a.ID
	downloadReq.allocationTx = a.Tx
	downloadReq.allocOwnerID = a.Owner
	downloadReq.ctx, downloadReq.ctxCncl = context.WithCancel(a.ctx)
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
	downloadReq.fullconsensus = a.fullconsensus
	downloadReq.consensusThresh = a.consensusThreshold
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

func (a *Allocation) StartRepair(localRootPath, pathToRepair string, statusCB StatusCallback) error {
	if !a.isInitialized() {
		return notInitialized
	}

	listDir, err := a.ListDir(pathToRepair)
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
	return errors.New("invalid_cancel_repair_request", "No repair in progress for the allocation")
}

func (a *Allocation) GetMaxWriteReadFromBlobbers(blobbers []*BlobberAllocation) (maxW float64, maxR float64, err error) {
	if !a.isInitialized() {
		return 0, 0, notInitialized
	}

	if len(blobbers) == 0 {
		return 0, 0, noBLOBBERS
	}

	maxWritePrice, maxReadPrice := 0.0, 0.0
	for _, v := range blobbers {
		if v.Terms.WritePrice.ToToken() > maxWritePrice {
			maxWritePrice = v.Terms.WritePrice.ToToken()
		}
		if v.Terms.ReadPrice.ToToken() > maxReadPrice {
			maxReadPrice = v.Terms.ReadPrice.ToToken()
		}
	}

	return maxWritePrice, maxReadPrice, nil
}

func (a *Allocation) GetMaxWriteRead() (maxW float64, maxR float64, err error) {
	return a.GetMaxWriteReadFromBlobbers(a.BlobberDetails)
}

func (a *Allocation) GetMinWriteRead() (minW float64, minR float64, err error) {
	if !a.isInitialized() {
		return 0, 0, notInitialized
	}

	blobbersCopy := a.BlobberDetails
	if len(blobbersCopy) == 0 {
		return 0, 0, noBLOBBERS
	}

	minWritePrice, minReadPrice := -1.0, -1.0
	for _, v := range blobbersCopy {
		if v.Terms.WritePrice.ToToken() < minWritePrice || minWritePrice < 0 {
			minWritePrice = v.Terms.WritePrice.ToToken()
		}
		if v.Terms.ReadPrice.ToToken() < minReadPrice || minReadPrice < 0 {
			minReadPrice = v.Terms.ReadPrice.ToToken()
		}
	}

	return minWritePrice, minReadPrice, nil
}

func (a *Allocation) GetMaxStorageCostFromBlobbers(size int64, blobbers []*BlobberAllocation) (float64, error) {
	var cost common.Balance // total price for size / duration

	for _, d := range blobbers {
		cost += a.uploadCostForBlobber(float64(d.Terms.WritePrice), size,
			a.DataShards, a.ParityShards)
	}

	return cost.ToToken(), nil
}

func (a *Allocation) GetMaxStorageCost(size int64) (float64, error) {
	var cost common.Balance // total price for size / duration

	for _, d := range a.BlobberDetails {
		fmt.Printf("write price for blobber %f datashards %d parity %d\n",
			float64(d.Terms.WritePrice), a.DataShards, a.ParityShards)

		cost += a.uploadCostForBlobber(float64(d.Terms.WritePrice), size,
			a.DataShards, a.ParityShards)

		fmt.Printf("Total cost %d\n", cost)
	}

	return cost.ToToken(), nil
}

func (a *Allocation) GetMinStorageCost(size int64) (common.Balance, error) {
	minW, _, err := a.GetMinWriteRead()
	if err != nil {
		return -1, err
	}

	return a.uploadCostForBlobber(minW, size, a.DataShards, a.ParityShards), nil
}

func (a *Allocation) uploadCostForBlobber(price float64, size int64, data, parity int) (
	cost common.Balance) {

	if data == 0 || parity == 0 {
		return -1.0
	}

	var ps = (size + int64(data) - 1) / int64(data)
	ps = ps * int64(data+parity)

	return common.Balance(price * a.sizeInGB(ps))
}

func (a *Allocation) sizeInGB(size int64) float64 {
	return float64(size) / GB
}

func (a *Allocation) getConsensuses() (fullConsensus, consensusThreshold int) {
	if a.DataShards == 0 {
		return 0, 0
	}

	if a.ParityShards == 0 {
		return a.DataShards, a.DataShards
	}

	return a.DataShards + a.ParityShards, a.DataShards + 1
}
