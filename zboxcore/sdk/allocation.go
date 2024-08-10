package sdk

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0chain/common/core/currency"
	"github.com/0chain/errors"
	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/pathutil"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/mitchellh/go-homedir"
	"go.uber.org/zap"
)

var (
	noBLOBBERS             = errors.New("", "No Blobbers set in this allocation")
	notInitialized         = errors.New("sdk_not_initialized", "Please call InitStorageSDK Init and use GetAllocation to get the allocation object")
	IsWasm                 = false
	MultiOpBatchSize       = 50
	RepairBatchSize        = 50
	Workdir                string
	multiOpRepairBatchSize = 10
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

const (
	emptyFileDataHash = "d41d8cd98f00b204e9800998ecf8427e"
	getRefPageLimit   = 1000
)

// Expected success rate is calculated (NumDataShards)*100/(NumDataShards+NumParityShards)

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
	NumBlocks       int64
	ActualFileSize  int64
	ActualNumBlocks int64
	EncryptedKey    string

	ActualThumbnailSize int64
	ActualThumbnailHash string

	Collaborators []fileref.Collaborator
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
	MaxOfferDuration time.Duration  `json:"max_offer_duration"`
}

// UpdateTerms represents Blobber terms during update blobber calls.
// A Blobber can update its terms, but any existing offer will use terms of offer signing time.
type UpdateTerms struct {
	ReadPrice        *common.Balance `json:"read_price,omitempty"`  // tokens / GB
	WritePrice       *common.Balance `json:"write_price,omitempty"` // tokens / GB
	MaxOfferDuration *time.Duration  `json:"max_offer_duration,omitempty"`
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
	WritePriceRange         PriceRange       `json:"write_price_range"`
	MinLockDemand           float64          `json:"min_lock_demand"`
	ChallengeCompletionTime time.Duration    `json:"challenge_completion_time"`
	StartTime               common.Timestamp `json:"start_time"`
	Finalized               bool             `json:"finalized,omitempty"`
	Canceled                bool             `json:"canceled,omitempty"`
	MovedToChallenge        common.Balance   `json:"moved_to_challenge,omitempty"`
	MovedBack               common.Balance   `json:"moved_back,omitempty"`
	MovedToValidators       common.Balance   `json:"moved_to_validators,omitempty"`
	FileOptions             uint16           `json:"file_options"`
	ThirdPartyExtendable    bool             `json:"third_party_extendable"`

	numBlockDownloads       int
	downloadChan            chan *DownloadRequest
	repairChan              chan *RepairRequest
	ctx                     context.Context
	ctxCancelF              context.CancelFunc
	mutex                   *sync.Mutex
	commitMutex             *sync.Mutex
	downloadProgressMap     map[string]*DownloadRequest
	downloadRequests        []*DownloadRequest
	repairRequestInProgress *RepairRequest
	initialized             bool
	checkStatus             bool
	readFree                bool
	// conseususes
	consensusThreshold int
	fullconsensus      int
	allocationVersion  int64
}

type OperationRequest struct {
	OperationType  string
	LocalPath      string
	RemotePath     string
	DestName       string // Required only for rename operation
	DestPath       string // Required for copy and move operation
	IsUpdate       bool
	IsRepair       bool // Required for repair operation
	IsWebstreaming bool
	EncryptedKey   string

	// Required for uploads
	Workdir         string
	FileMeta        FileMeta
	FileReader      io.Reader
	Mask            *zboxutil.Uint128 // Required for delete repair operation
	DownloadFile    bool              // Required for upload repair operation
	StreamUpload    bool              // Required for streaming file when actualSize is not available
	CancelCauseFunc context.CancelCauseFunc
	Opts            []ChunkedUploadOption
	CopyDirOnly     bool
}

func GetReadPriceRange() (PriceRange, error) {
	return getPriceRange("max_read_price")
}
func GetWritePriceRange() (PriceRange, error) {
	return getPriceRange("max_write_price")
}

func SetMultiOpBatchSize(size int) {
	MultiOpBatchSize = size
}

func SetWasm() {
	IsWasm = true
	BatchSize = 4
	extraCount = 0
	RepairBatchSize = 20
}

func (a *Allocation) SetCheckStatus(checkStatus bool) {
	a.checkStatus = checkStatus
}

func getPriceRange(name string) (PriceRange, error) {
	conf, err := GetStorageSCConfig()
	if err != nil {
		return PriceRange{}, err
	}
	f := conf.Fields[name]
	fStr, ok := f.(string)
	if !ok {
		return PriceRange{}, fmt.Errorf("type is wrong")
	}
	mrp, err := strconv.ParseFloat(fStr, 64)
	if err != nil {
		return PriceRange{}, err
	}
	coin, err := currency.ParseZCN(mrp)
	if err != nil {
		return PriceRange{}, err
	}
	max, err := coin.Int64()
	if err != nil {
		return PriceRange{}, err
	}
	return PriceRange{0, uint64(max)}, err

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
		go getAllocationDataFromBlobber(blobber, a.ID, a.Tx, rspCh, wg)
	}
	wg.Wait()
	result := make(map[string]*BlobberAllocationStats, len(a.Blobbers))
	for i := 0; i < numList; i++ {
		resp := <-rspCh
		result[resp.BlobberURL] = resp
	}
	return result
}

var downloadWorkerCount = 6

func SetDownloadWorkerCount(count int) {
	downloadWorkerCount = count
}

func (a *Allocation) InitAllocation() {
	a.downloadChan = make(chan *DownloadRequest, 100)
	a.repairChan = make(chan *RepairRequest, 1)
	a.ctx, a.ctxCancelF = context.WithCancel(context.Background())
	a.downloadProgressMap = make(map[string]*DownloadRequest)
	a.downloadRequests = make([]*DownloadRequest, 0, 100)
	a.mutex = &sync.Mutex{}
	a.commitMutex = &sync.Mutex{}
	a.fullconsensus, a.consensusThreshold = a.getConsensuses()
	a.readFree = true
	if a.ReadPriceRange.Max > 0 {
		for _, blobberDetail := range a.BlobberDetails {
			if blobberDetail.Terms.ReadPrice > 0 {
				a.readFree = false
				break
			}
		}
	}
	a.startWorker(a.ctx)
	InitCommitWorker(a.Blobbers)
	InitBlockDownloader(a.Blobbers, downloadWorkerCount)
	a.CheckAllocStatus() //nolint:errcheck
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
		case downloadReq := <-a.downloadChan:
			l.Logger.Info(fmt.Sprintf("received a download request for %v\n", downloadReq.remotefilepath))
			go func() {
				downloadReq.processDownload()
			}()
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

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, true, false, "", false, false)
}

// UploadFile [Deprecated]please use CreateChunkedUpload
func (a *Allocation) UploadFile(workdir, localpath string, remotepath string,
	status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, false, false, "", false, false)
}

func (a *Allocation) RepairFile(file sys.File, remotepath string, statusCallback StatusCallback, mask zboxutil.Uint128, ref ORef) *OperationRequest {
	idr, _ := homedir.Dir()
	if Workdir != "" {
		idr = Workdir
	}
	fileMeta := FileMeta{
		ActualSize: ref.ActualFileSize,
		MimeType:   ref.MimeType,
		RemoteName: ref.Name,
		RemotePath: remotepath,
		CustomMeta: ref.CustomMeta,
	}
	var opts []ChunkedUploadOption
	if ref.EncryptedKey != "" {
		opts = []ChunkedUploadOption{
			WithMask(mask),
			WithEncrypt(true),
			WithStatusCallback(statusCallback),
			WithEncryptedPoint(ref.EncryptedKeyPoint),
			WithChunkNumber(100),
		}
	} else {
		opts = []ChunkedUploadOption{
			WithMask(mask),
			WithStatusCallback(statusCallback),
			WithChunkNumber(100),
		}
	}
	op := &OperationRequest{
		OperationType: constants.FileOperationInsert,
		IsRepair:      true,
		RemotePath:    remotepath,
		Workdir:       idr,
		FileMeta:      fileMeta,
		Opts:          opts,
		FileReader:    file,
		Mask:          &mask,
		EncryptedKey:  ref.EncryptedKey,
	}
	if ref.ActualFileHash == emptyFileDataHash {
		op.FileMeta.ActualSize = 0
	}
	return op
}

// UpdateFileWithThumbnail [Deprecated]please use CreateChunkedUpload
func (a *Allocation) UpdateFileWithThumbnail(workdir, localpath string, remotepath string,
	thumbnailpath string, status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, true, false,
		thumbnailpath, false, false)
}

// UploadFileWithThumbnail [Deprecated]please use CreateChunkedUpload
func (a *Allocation) UploadFileWithThumbnail(workdir string, localpath string,
	remotepath string, thumbnailpath string,
	status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, false, false,
		thumbnailpath, false, false)
}

// EncryptAndUpdateFile [Deprecated]please use CreateChunkedUpload
func (a *Allocation) EncryptAndUpdateFile(workdir string, localpath string, remotepath string,
	status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, true, false, "", true, false)
}

// EncryptAndUploadFile [Deprecated]please use CreateChunkedUpload
func (a *Allocation) EncryptAndUploadFile(workdir string, localpath string, remotepath string,
	status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, false, false, "", true, false)
}

// EncryptAndUpdateFileWithThumbnail [Deprecated]please use CreateChunkedUpload
func (a *Allocation) EncryptAndUpdateFileWithThumbnail(workdir string, localpath string,
	remotepath string, thumbnailpath string, status StatusCallback) error {

	return a.StartChunkedUpload(workdir, localpath, remotepath, status, true, false,
		thumbnailpath, true, false)
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
		false,
	)
}

func (a *Allocation) StartMultiUpload(workdir string, localPaths []string, fileNames []string, thumbnailPaths []string, encrypts []bool, chunkNumbers []int, remotePaths []string, isUpdate []bool, isWebstreaming []bool, status StatusCallback) error {
	if len(localPaths) != len(thumbnailPaths) {
		return errors.New("invalid_value", "length of localpaths and thumbnailpaths must be equal")
	}
	if len(localPaths) != len(encrypts) {
		return errors.New("invalid_value", "length of encrypt not equal to number of files")
	}
	if !a.isInitialized() {
		return notInitialized
	}

	if !a.CanUpload() {
		return constants.ErrFileOptionNotPermitted
	}

	totalOperations := len(localPaths)
	if totalOperations == 0 {
		return nil
	}
	operationRequests := make([]OperationRequest, totalOperations)
	for idx, localPath := range localPaths {
		remotePath := zboxutil.RemoteClean(remotePaths[idx])
		isabs := zboxutil.IsRemoteAbs(remotePath)
		if !isabs {
			err := thrown.New("invalid_path", "Path should be valid and absolute")
			return err
		}
		fileReader, err := os.Open(localPath)
		if err != nil {
			return err
		}
		defer fileReader.Close()
		thumbnailPath := thumbnailPaths[idx]
		fileName := fileNames[idx]
		chunkNumber := chunkNumbers[idx]
		if fileName == "" {
			return thrown.New("invalid_param", "filename can't be empty")
		}
		encrypt := encrypts[idx]

		fileInfo, err := fileReader.Stat()
		if err != nil {
			return err
		}

		mimeType, err := zboxutil.GetFileContentType(path.Ext(fileName), fileReader)
		if err != nil {
			return err
		}

		if !strings.HasSuffix(remotePath, "/") {
			remotePath = remotePath + "/"
		}
		fullRemotePath := zboxutil.GetFullRemotePath(localPath, remotePath)
		fullRemotePathWithoutName, _ := pathutil.Split(fullRemotePath)
		fullRemotePath = fullRemotePathWithoutName + "/" + fileName

		fileMeta := FileMeta{
			Path:       localPath,
			ActualSize: fileInfo.Size(),
			MimeType:   mimeType,
			RemoteName: fileName,
			RemotePath: fullRemotePath,
		}
		options := []ChunkedUploadOption{
			WithStatusCallback(status),
			WithEncrypt(encrypt),
		}
		if chunkNumber != 0 {
			options = append(options, WithChunkNumber(chunkNumber))
		}
		if thumbnailPath != "" {
			buf, err := sys.Files.ReadFile(thumbnailPath)
			if err != nil {
				return err
			}

			options = append(options, WithThumbnail(buf))
		}
		operationRequests[idx] = OperationRequest{
			FileMeta:      fileMeta,
			FileReader:    fileReader,
			OperationType: constants.FileOperationInsert,
			Opts:          options,
			Workdir:       workdir,
			RemotePath:    fileMeta.RemotePath,
		}

		if isUpdate[idx] {
			operationRequests[idx].OperationType = constants.FileOperationUpdate
		}
		if isWebstreaming[idx] {
			operationRequests[idx].IsWebstreaming = true
		}

	}
	err := a.DoMultiOperation(operationRequests)
	if err != nil {
		logger.Logger.Error("Error in multi upload ", err.Error())
		return err
	}
	return nil
}

func (a *Allocation) StartChunkedUpload(workdir, localPath string,
	remotePath string,
	status StatusCallback,
	isUpdate bool,
	isRepair bool,
	thumbnailPath string,
	encryption bool,
	webStreaming bool,
	uploadOpts ...ChunkedUploadOption,
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

	remotePath = zboxutil.RemoteClean(remotePath)
	isabs := zboxutil.IsRemoteAbs(remotePath)
	if !isabs {
		err = thrown.New("invalid_path", "Path should be valid and absolute")
		return err
	}
	remotePath = zboxutil.GetFullRemotePath(localPath, remotePath)

	_, fileName := pathutil.Split(remotePath)

	mimeType, err := zboxutil.GetFileContentType(path.Ext(fileName), fileReader)
	if err != nil {
		return err
	}

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
	options = append(options, uploadOpts...)

	if thumbnailPath != "" {
		buf, err := sys.Files.ReadFile(thumbnailPath)
		if err != nil {
			return err
		}

		options = append(options, WithThumbnail(buf))
	}

	connectionId := zboxutil.NewConnectionId()
	now := time.Now()
	ChunkedUpload, err := CreateChunkedUpload(a.ctx, workdir,
		a, fileMeta, fileReader,
		isUpdate, isRepair, webStreaming, connectionId,
		options...)
	if err != nil {
		return err
	}
	elapsedCreateChunkedUpload := time.Since(now)
	logger.Logger.Info("[StartChunkedUpload]", zap.String("allocation_id", a.ID),
		zap.Duration("CreateChunkedUpload", elapsedCreateChunkedUpload))

	return ChunkedUpload.Start()
}

func (a *Allocation) GetCurrentVersion() (bool, error) {
	//get versions from blobbers

	wg := &sync.WaitGroup{}
	markerChan := make(chan *RollbackBlobber, len(a.Blobbers))
	var errCnt int32
	for _, blobber := range a.Blobbers {

		wg.Add(1)
		go func(blobber *blockchain.StorageNode) {

			defer wg.Done()
			wr, err := GetWritemarker(a.ID, a.Tx, blobber.ID, blobber.Baseurl)
			if err != nil {
				atomic.AddInt32(&errCnt, 1)
				logger.Logger.Error("error during getWritemarke", zap.Error(err))
			}
			if wr == nil {
				markerChan <- nil
			} else {
				markerChan <- &RollbackBlobber{
					blobber:      blobber,
					lvm:          wr,
					commitResult: &CommitResult{},
				}
			}
		}(blobber)

	}

	wg.Wait()
	close(markerChan)

	versionMap := make(map[int64][]*RollbackBlobber)

	for rb := range markerChan {

		if rb == nil || rb.lvm == nil {
			continue
		}

		if _, ok := versionMap[rb.lvm.VersionMarker.Version]; !ok {
			versionMap[rb.lvm.VersionMarker.Version] = make([]*RollbackBlobber, 0)
		}

		versionMap[rb.lvm.VersionMarker.Version] = append(versionMap[rb.lvm.VersionMarker.Version], rb)

		if len(versionMap) > 2 {
			return false, fmt.Errorf("more than 2 versions found")
		}

	}
	// TODO: check how many blobbers can be down
	if errCnt > 0 {
		return false, fmt.Errorf("error in getting writemarker from %v blobbers", errCnt)
	}

	if len(versionMap) == 0 {
		return true, nil
	}

	// TODO:return if len(versionMap) == 1

	var prevVersion int64
	var latestVersion int64

	for version := range versionMap {
		if prevVersion == 0 {
			prevVersion = version
		} else {
			latestVersion = version
		}
	}

	if prevVersion > latestVersion {
		prevVersion, latestVersion = latestVersion, prevVersion
	}

	// TODO: Check if allocation can be repaired

	success := true

	// rollback to prev version
	for _, rb := range versionMap[latestVersion] {

		wg.Add(1)
		go func(rb *RollbackBlobber) {
			defer wg.Done()
			err := rb.processRollback(context.TODO(), a.ID)
			if err != nil {
				success = false
			}
		}(rb)
	}

	wg.Wait()

	if !success {
		return false, fmt.Errorf("error in rollback")
	}

	return success, nil
}

func (a *Allocation) RepairRequired(remotepath string) (zboxutil.Uint128, zboxutil.Uint128, bool, *fileref.FileRef, error) {
	if !a.isInitialized() {
		return zboxutil.Uint128{}, zboxutil.Uint128{}, false, nil, notInitialized
	}

	listReq := &ListRequest{Consensus: Consensus{RWMutex: &sync.RWMutex{}}}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.blobbers = a.Blobbers
	listReq.fullconsensus = a.fullconsensus
	listReq.consensusThresh = a.DataShards
	listReq.ctx = a.ctx
	listReq.remotefilepath = remotepath
	found, deleteMask, fileRef, _ := listReq.getFileConsensusFromBlobbers()
	if fileRef == nil {
		var repairErr error
		if deleteMask.Equals(zboxutil.NewUint128(0)) {
			repairErr = errors.New("", "File not found for the given remotepath")
		}
		return found, deleteMask, false, fileRef, repairErr
	}

	uploadMask := zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)

	return found, deleteMask, !found.Equals(uploadMask), fileRef, nil
}

func (a *Allocation) DoMultiOperation(operations []OperationRequest, opts ...MultiOperationOption) error {
	if len(operations) == 0 {
		return nil
	}
	if !a.isInitialized() {
		return notInitialized
	}
	connectionID := zboxutil.NewConnectionId()
	var mo MultiOperation
	for i := 0; i < len(operations); {
		// resetting multi operation and previous paths for every batch
		mo.allocationObj = a
		mo.operationMask = zboxutil.NewUint128(0)
		mo.maskMU = &sync.Mutex{}
		mo.connectionID = connectionID
		mo.ctx, mo.ctxCncl = context.WithCancelCause(a.ctx)
		mo.Consensus = Consensus{
			RWMutex:         &sync.RWMutex{},
			consensusThresh: a.consensusThreshold,
			fullconsensus:   a.fullconsensus,
		}
		for _, opt := range opts {
			opt(&mo)
		}
		previousPaths := make(map[string]bool)
		connectionErrors := make([]error, len(mo.allocationObj.Blobbers))

		var wg sync.WaitGroup
		for blobberIdx := range mo.allocationObj.Blobbers {
			wg.Add(1)
			go func(pos int) {
				defer wg.Done()
				err := mo.createConnectionObj(pos)
				if err != nil {
					l.Logger.Error(err.Error())
					connectionErrors[pos] = err
				}
			}(blobberIdx)
		}
		wg.Wait()
		// Check consensus
		if mo.operationMask.CountOnes() < mo.consensusThresh {
			l.Logger.Error("Multioperation: create connection failed. Required consensus not met",
				zap.Int("consensusThresh", mo.consensusThresh),
				zap.Int("operationMask", mo.operationMask.CountOnes()),
				zap.Any("connectionErrors", connectionErrors))

			majorErr := zboxutil.MajorError(connectionErrors)
			if majorErr != nil {
				return errors.New("consensus_not_met",
					fmt.Sprintf("Multioperation: create connection failed. Required consensus %d got %d. Major error: %s",
						mo.consensusThresh, mo.operationMask.CountOnes(), majorErr.Error()))
			}
			return errors.New("consensus_not_met",
				fmt.Sprintf("Multioperation: create connection failed. Required consensus %d got %d",
					mo.consensusThresh, mo.operationMask.CountOnes()))
		}

		for ; i < len(operations); i++ {
			if len(mo.operations) >= MultiOpBatchSize {
				// max batch size reached, commit
				connectionID = zboxutil.NewConnectionId()
				break
			}
			op := operations[i]
			op.RemotePath = strings.TrimSpace(op.RemotePath)
			if op.FileMeta.RemotePath != "" {
				op.FileMeta.RemotePath = strings.TrimSpace(op.FileMeta.RemotePath)
				op.FileMeta.RemoteName = strings.TrimSpace(op.FileMeta.RemoteName)
			}
			remotePath := op.RemotePath
			parentPaths := GenerateParentPaths(remotePath)

			if _, ok := previousPaths[remotePath]; ok {
				// conflict found, commit
				connectionID = zboxutil.NewConnectionId()
				break
			}

			var (
				operation       Operationer
				err             error
				newConnectionID string
			)

			switch op.OperationType {
			case constants.FileOperationRename:
				operation = NewRenameOperation(op.RemotePath, op.DestName, mo.operationMask, mo.maskMU, mo.consensusThresh, mo.fullconsensus, mo.ctx)

			case constants.FileOperationCopy:
				operation = NewCopyOperation(op.RemotePath, op.DestPath, mo.operationMask, mo.maskMU, mo.consensusThresh, mo.fullconsensus, op.CopyDirOnly, mo.ctx)

			case constants.FileOperationMove:
				operation = NewMoveOperation(op.RemotePath, op.DestPath, mo.operationMask, mo.maskMU, mo.consensusThresh, mo.fullconsensus, mo.ctx)

			case constants.FileOperationInsert:
				cancelLock.Lock()
				CancelOpCtx[op.FileMeta.RemotePath] = mo.ctxCncl
				cancelLock.Unlock()
				operation, newConnectionID, err = NewUploadOperation(mo.ctx, op.Workdir, mo.allocationObj, mo.connectionID, op.FileMeta, op.FileReader, false, op.IsWebstreaming, op.IsRepair, op.DownloadFile, op.StreamUpload, op.Opts...)

			case constants.FileOperationDelete:
				if op.Mask != nil {
					operation = NewDeleteOperation(op.RemotePath, *op.Mask, mo.maskMU, op.Mask.CountOnes(), mo.fullconsensus, mo.ctx)
				} else {
					operation = NewDeleteOperation(op.RemotePath, mo.operationMask, mo.maskMU, mo.consensusThresh, mo.fullconsensus, mo.ctx)
				}

			case constants.FileOperationUpdate:
				cancelLock.Lock()
				CancelOpCtx[op.FileMeta.RemotePath] = mo.ctxCncl
				cancelLock.Unlock()
				operation, newConnectionID, err = NewUploadOperation(mo.ctx, op.Workdir, mo.allocationObj, mo.connectionID, op.FileMeta, op.FileReader, true, op.IsWebstreaming, op.IsRepair, op.DownloadFile, op.StreamUpload, op.Opts...)

			case constants.FileOperationCreateDir:
				operation = NewDirOperation(op.RemotePath, op.FileMeta.CustomMeta, mo.operationMask, mo.maskMU, mo.consensusThresh, mo.fullconsensus, mo.ctx)

			default:
				return errors.New("invalid_operation", "Operation is not valid")
			}
			if err != nil {
				return err
			}

			if newConnectionID != "" && newConnectionID != connectionID {
				connectionID = newConnectionID
				break
			}
			err = operation.Verify(a)
			if err != nil {
				return err
			}

			for path := range parentPaths {
				previousPaths[path] = true
			}

			mo.operations = append(mo.operations, operation)
		}

		if len(mo.operations) > 0 {
			err := mo.Process()
			if err != nil {
				logger.Logger.Error("Error in multi operation", zap.Error(err))
				return err
			}

			mo.operations = nil
		}
	}
	return nil
}

func GenerateParentPaths(path string) map[string]bool {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	parentPaths := make(map[string]bool)

	for i := range parts {
		parentPaths["/"+strings.Join(parts[:i+1], "/")] = true
	}
	return parentPaths
}

func (a *Allocation) DownloadFileToFileHandler(
	fileHandler sys.File,
	remotePath string,
	verifyDownload bool,
	status StatusCallback,
	isFinal bool,
	downloadReqOpts ...DownloadRequestOption,
) error {
	return a.addAndGenerateDownloadRequest(fileHandler, remotePath, DOWNLOAD_CONTENT_FULL, 1, 0,
		numBlockDownloads, verifyDownload, status, isFinal, "", downloadReqOpts...)
}

func (a *Allocation) DownloadByBlocksToFileHandler(
	fileHandler sys.File,
	remotePath string,
	startBlock, endBlock int64,
	numBlocks int,
	verifyDownload bool,
	status StatusCallback,
	isFinal bool,
	downloadReqOpts ...DownloadRequestOption,
) error {
	return a.addAndGenerateDownloadRequest(fileHandler, remotePath, DOWNLOAD_CONTENT_FULL, startBlock, endBlock,
		numBlocks, verifyDownload, status, isFinal, "", downloadReqOpts...)
}

func (a *Allocation) DownloadThumbnailToFileHandler(
	fileHandler sys.File,
	remotePath string,
	verifyDownload bool,
	status StatusCallback,
	isFinal bool,
	downloadReqOpts ...DownloadRequestOption,
) error {
	return a.addAndGenerateDownloadRequest(fileHandler, remotePath, DOWNLOAD_CONTENT_THUMB, 1, 0,
		numBlockDownloads, verifyDownload, status, isFinal, "", downloadReqOpts...)
}

func (a *Allocation) DownloadFile(localPath string, remotePath string, verifyDownload bool, status StatusCallback, isFinal bool, downloadReqOpts ...DownloadRequestOption) error {
	f, localFilePath, toKeep, err := a.prepareAndOpenLocalFile(localPath, remotePath)
	if err != nil {
		return err
	}
	downloadReqOpts = append(downloadReqOpts, WithFileCallback(func() {
		f.Close() //nolint: errcheck
	}))
	err = a.addAndGenerateDownloadRequest(f, remotePath, DOWNLOAD_CONTENT_FULL, 1, 0,
		numBlockDownloads, verifyDownload, status, isFinal, localFilePath, downloadReqOpts...)
	if err != nil {
		if !toKeep {
			os.Remove(localFilePath) //nolint: errcheck
		}
		f.Close() //nolint: errcheck
		return err
	}
	return nil
}

// TODO: Use a map to store the download request and use flag isFinal to start the download, calculate readCount in parallel if possible
func (a *Allocation) DownloadFileByBlock(
	localPath string, remotePath string, startBlock int64, endBlock int64,
	numBlocks int, verifyDownload bool, status StatusCallback, isFinal bool, downloadReqOpts ...DownloadRequestOption) error {
	f, localFilePath, toKeep, err := a.prepareAndOpenLocalFile(localPath, remotePath)
	if err != nil {
		return err
	}
	downloadReqOpts = append(downloadReqOpts, WithFileCallback(func() {
		f.Close() //nolint: errcheck
	}))
	err = a.addAndGenerateDownloadRequest(f, remotePath, DOWNLOAD_CONTENT_FULL, startBlock, endBlock,
		numBlockDownloads, verifyDownload, status, isFinal, localFilePath, downloadReqOpts...)
	if err != nil {
		if !toKeep {
			os.Remove(localFilePath) //nolint: errcheck
		}
		f.Close() //nolint: errcheck
		return err
	}
	return nil
}

func (a *Allocation) DownloadThumbnail(localPath string, remotePath string, verifyDownload bool, status StatusCallback, isFinal bool) error {
	f, localFilePath, toKeep, err := a.prepareAndOpenLocalFile(localPath, remotePath)
	if err != nil {
		return err
	}

	err = a.addAndGenerateDownloadRequest(f, remotePath, DOWNLOAD_CONTENT_THUMB, 1, 0,
		numBlockDownloads, verifyDownload, status, isFinal, localFilePath, WithFileCallback(func() {
			f.Close() //nolint: errcheck
		}))
	if err != nil {
		if !toKeep {
			os.Remove(localFilePath) //nolint: errcheck
		}
		f.Close() //nolint: errcheck
		return err
	}
	return nil
}

func (a *Allocation) generateDownloadRequest(
	fileHandler sys.File,
	remotePath string,
	contentMode string,
	startBlock, endBlock int64,
	numBlocks int,
	verifyDownload bool,
	status StatusCallback,
	connectionID string,
	localFilePath string,
) (*DownloadRequest, error) {
	if len(a.Blobbers) == 0 {
		return nil, noBLOBBERS
	}

	downloadReq := &DownloadRequest{Consensus: Consensus{RWMutex: &sync.RWMutex{}}}
	downloadReq.maskMu = &sync.Mutex{}
	downloadReq.allocationID = a.ID
	downloadReq.allocationTx = a.Tx
	downloadReq.allocOwnerID = a.Owner
	downloadReq.allocOwnerPubKey = a.OwnerPublicKey
	downloadReq.ctx, downloadReq.ctxCncl = context.WithCancel(a.ctx)
	downloadReq.fileHandler = fileHandler
	downloadReq.localFilePath = localFilePath
	downloadReq.remotefilepath = remotePath
	downloadReq.statusCallback = status
	downloadReq.downloadMask = zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
	downloadReq.blobbers = a.Blobbers
	downloadReq.datashards = a.DataShards
	downloadReq.parityshards = a.ParityShards
	downloadReq.startBlock = startBlock - 1
	downloadReq.endBlock = endBlock
	downloadReq.numBlocks = int64(numBlocks)
	downloadReq.shouldVerify = verifyDownload
	downloadReq.fullconsensus = a.fullconsensus
	downloadReq.consensusThresh = a.DataShards
	downloadReq.completedCallback = func(remotepath string, remotepathhash string) {
		a.mutex.Lock()
		defer a.mutex.Unlock()
		delete(a.downloadProgressMap, remotepath)
	}
	// downloadReq.fileCallback = func() {
	// 	if downloadReq.fileHandler != nil {
	// 		downloadReq.fileHandler.Close() //nolint: errcheck
	// 	}
	// }
	downloadReq.contentMode = contentMode
	downloadReq.connectionID = connectionID
	downloadReq.downloadQueue = make(downloadQueue, len(a.Blobbers))
	for i := 0; i < len(a.Blobbers); i++ {
		downloadReq.downloadQueue[i].timeTaken = 1000000
	}

	return downloadReq, nil
}

func (a *Allocation) addAndGenerateDownloadRequest(
	fileHandler sys.File,
	remotePath, contentMode string,
	startBlock, endBlock int64,
	numBlocks int,
	verifyDownload bool,
	status StatusCallback,
	isFinal bool,
	localFilePath string,
	downloadReqOpts ...DownloadRequestOption,
) error {
	downloadReq, err := a.generateDownloadRequest(
		fileHandler, remotePath, contentMode, startBlock, endBlock,
		numBlocks, verifyDownload, status, "", localFilePath)
	if err != nil {
		return err
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if len(a.downloadRequests) > 0 {
		downloadReq.connectionID = a.downloadRequests[0].connectionID
	} else {
		downloadReq.connectionID = zboxutil.NewConnectionId()
	}
	for _, opt := range downloadReqOpts {
		opt(downloadReq)
	}
	downloadReq.workdir = filepath.Join(downloadReq.workdir, ".zcn")
	a.downloadProgressMap[remotePath] = downloadReq
	a.downloadRequests = append(a.downloadRequests, downloadReq)
	if isFinal {
		downloadOps := a.downloadRequests
		a.downloadRequests = nil
		go func() {
			a.processReadMarker(downloadOps)
		}()
	}
	return nil
}

func (a *Allocation) processReadMarker(drs []*DownloadRequest) {
	blobberMap := make(map[uint64]int64)
	mpLock := sync.Mutex{}
	wg := sync.WaitGroup{}
	now := time.Now()

	for _, dr := range drs {
		wg.Add(1)
		go func(dr *DownloadRequest) {
			defer wg.Done()
			if a.readFree {
				dr.freeRead = true
			}
			dr.processDownloadRequest()
			var pos uint64
			if !dr.skip {
				for i := dr.downloadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
					pos = uint64(i.TrailingZeros())
					mpLock.Lock()
					blobberMap[pos] += dr.blocksPerShard
					mpLock.Unlock()
				}
			}
		}(dr)
	}
	wg.Wait()
	elapsedProcessDownloadRequest := time.Since(now)

	// Do not send readmarkers for free reads
	if a.readFree {
		for _, dr := range drs {
			if dr.skip {
				continue
			}
			go func(dr *DownloadRequest) {
				a.downloadChan <- dr
			}(dr)
		}
		l.Logger.Debug("[processReadMarker]", zap.String("allocation_id", a.ID),
			zap.Int("num of download requests", len(drs)),
			zap.Duration("processDownloadRequest", elapsedProcessDownloadRequest))
		return
	}

	successMask := zboxutil.NewUint128(0)
	var redeemError error

	for pos, totalBlocks := range blobberMap {
		if totalBlocks == 0 {
			continue
		}
		wg.Add(1)
		go func(pos uint64, totalBlocks int64) {
			blobber := drs[0].blobbers[pos]
			err := drs[0].submitReadMarker(blobber, totalBlocks)
			if err == nil {
				successMask = successMask.Or(zboxutil.NewUint128(1).Lsh(pos))
			} else {
				redeemError = err
			}
			wg.Done()
		}(pos, totalBlocks)
	}
	wg.Wait()
	elapsedSubmitReadmarker := time.Since(now) - elapsedProcessDownloadRequest

	l.Logger.Info("[processReadMarker]", zap.String("allocation_id", a.ID),
		zap.Int("num of download requests", len(drs)),
		zap.Duration("processDownloadRequest", elapsedProcessDownloadRequest),
		zap.Duration("submitReadmarker", elapsedSubmitReadmarker))
	for _, dr := range drs {
		if dr.skip {
			continue
		}
		dr.downloadMask = successMask.And(dr.downloadMask)
		if dr.consensusThresh > dr.downloadMask.CountOnes() {
			if redeemError == nil {
				redeemError = errors.New("read_marker_failed", "Failed to submit read marker to the blobbers")
			}
			dr.errorCB(redeemError, dr.remotefilepath)
			continue
		}
		go func(dr *DownloadRequest) {
			a.downloadChan <- dr
		}(dr)
	}
}

func (a *Allocation) prepareAndOpenLocalFile(localPath string, remotePath string) (*os.File, string, bool, error) {
	var toKeep bool

	if !a.isInitialized() {
		return nil, "", toKeep, notInitialized
	}

	var localFilePath string

	// If the localPath has a file extension, treat it as a file. Otherwise, treat it as a directory.
	if filepath.Ext(localPath) != "" {
		localFilePath = localPath
	} else {
		localFileName := filepath.Base(remotePath)
		localFilePath = filepath.Join(localPath, localFileName)
	}

	// Create necessary directories if they do not exist
	dir := filepath.Dir(localFilePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0744); err != nil {
			return nil, "", toKeep, err
		}
	}

	var f *os.File
	info, err := os.Stat(localFilePath)
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.OpenFile(localFilePath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return nil, "", toKeep, errors.Wrap(err, "Can't create local file")
		}
	} else {
		f, err = os.OpenFile(localFilePath, os.O_WRONLY, 0644)
		if err != nil {
			return nil, "", toKeep, errors.Wrap(err, "Can't open local file in append mode")
		}
		if info.Size() > 0 {
			toKeep = true
		}
	}

	return f, localFilePath, toKeep, nil
}

func (a *Allocation) ListDirFromAuthTicket(authTicket string, lookupHash string, opts ...ListRequestOptions) (*ListResult, error) {
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

	listReq := &ListRequest{Consensus: Consensus{RWMutex: &sync.RWMutex{}}}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.blobbers = a.Blobbers
	listReq.fullconsensus = a.fullconsensus
	listReq.consensusThresh = a.consensusThreshold
	listReq.ctx = a.ctx
	listReq.remotefilepathhash = lookupHash
	listReq.authToken = at
	for _, opt := range opts {
		opt(listReq)
	}
	ref, err := listReq.GetListFromBlobbers()

	if err != nil {
		return nil, err
	}

	if ref != nil {
		return ref, nil
	}
	return nil, errors.New("list_request_failed", "Failed to get list response from the blobbers")
}

func (a *Allocation) ListDir(path string, opts ...ListRequestOptions) (*ListResult, error) {
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
	listReq := &ListRequest{Consensus: Consensus{RWMutex: &sync.RWMutex{}}}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.blobbers = a.Blobbers
	listReq.fullconsensus = a.fullconsensus
	listReq.consensusThresh = a.DataShards
	listReq.ctx = a.ctx
	listReq.remotefilepath = path
	for _, opt := range opts {
		opt(listReq)
	}
	ref, err := listReq.GetListFromBlobbers()
	if err != nil {
		return nil, err
	}

	if ref != nil {
		return ref, nil
	}
	return nil, errors.New("list_request_failed", "Failed to get list response from the blobbers")
}

func (a *Allocation) getRefs(path, pathHash, authToken, offsetPath, updatedDate, offsetDate, fileType, refType string, level, pageLimit int, opts ...ObjectTreeRequestOption) (*ObjectTreeResult, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}

	oTreeReq := &ObjectTreeRequest{
		allocationID:   a.ID,
		allocationTx:   a.Tx,
		blobbers:       a.Blobbers,
		authToken:      authToken,
		pathHash:       pathHash,
		remotefilepath: path,
		pageLimit:      pageLimit,
		level:          level,
		offsetPath:     offsetPath,
		updatedDate:    updatedDate,
		offsetDate:     offsetDate,
		fileType:       fileType,
		refType:        refType,
		ctx:            a.ctx,
		reqMask:        zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1),
	}
	oTreeReq.fullconsensus = a.fullconsensus
	oTreeReq.consensusThresh = a.DataShards
	for _, opt := range opts {
		opt(oTreeReq)
	}
	//TODO: Have a mask on allocation object to track blobbers which are on latest version
	// if singleClientMode {
	// 	oTreeReq.singleBlobber = true
	// }
	return oTreeReq.GetRefs()
}

func (a *Allocation) getDownloadMaskForBlobber(blobberID string) (zboxutil.Uint128, []*blockchain.StorageNode, error) {

	x := zboxutil.NewUint128(1)
	blobberIdx := 0
	found := false
	for idx, b := range a.Blobbers {
		if b.ID == blobberID {
			found = true
			blobberIdx = idx
		}
	}

	if !found {
		return x, nil, fmt.Errorf("no blobber found with the given ID")
	}

	return x, a.Blobbers[blobberIdx : blobberIdx+1], nil
}

func (a *Allocation) DownloadFromBlobber(blobberID, localPath, remotePath string, status StatusCallback, opts ...DownloadRequestOption) error {

	mask, blobbers, err := a.getDownloadMaskForBlobber(blobberID)
	if err != nil {
		l.Logger.Error(err)
		return err
	}

	verifyDownload := false // should be set to false

	f, localFilePath, toKeep, err := a.prepareAndOpenLocalFile(localPath, remotePath)
	if err != nil {
		return err
	}
	downloadReq, err := a.generateDownloadRequest(f, remotePath, DOWNLOAD_CONTENT_FULL, 1, 0, numBlockDownloads, verifyDownload,
		status, zboxutil.NewConnectionId(), localFilePath)
	if err != nil {
		if !toKeep {
			os.Remove(localFilePath) //nolint: errcheck
		}
		f.Close() //nolint: errcheck
		return err
	}

	downloadReq.downloadMask = mask
	downloadReq.blobbers = blobbers
	downloadReq.fullconsensus = 1
	downloadReq.consensusThresh = 1
	opts = append(opts, WithFileCallback(func() {
		f.Close() //nolint: errcheck
	}))
	for _, opt := range opts {
		opt(downloadReq)
	}

	fRef, err := downloadReq.getFileRef()
	if err != nil {
		l.Logger.Error(err.Error())
		downloadReq.errorCB(fmt.Errorf("Error while getting file ref. Error: %v", err), remotePath)
		return err
	}
	downloadReq.numBlocks = fRef.NumBlocks

	a.processReadMarker([]*DownloadRequest{downloadReq})
	if downloadReq.skip {
		return errors.New("download_request_failed", "Failed to get download response from the blobbers")
	}
	return nil
}

// GetRefsWithAuthTicket get refs that are children of shared remote path.
func (a *Allocation) GetRefsWithAuthTicket(authToken, offsetPath, updatedDate, offsetDate, fileType, refType string, level, pageLimit int) (*ObjectTreeResult, error) {
	if authToken == "" {
		return nil, errors.New("empty_auth_token", "auth token cannot be empty")
	}
	sEnc, err := base64.StdEncoding.DecodeString(authToken)
	if err != nil {
		return nil, errors.New("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}

	authTicket := new(marker.AuthTicket)
	if err := json.Unmarshal(sEnc, authTicket); err != nil {
		return nil, errors.New("json_unmarshall_error", err.Error())
	}

	at, _ := json.Marshal(authTicket)
	return a.getRefs("", authTicket.FilePathHash, string(at), offsetPath, updatedDate, offsetDate, fileType, refType, level, pageLimit)
}

// This function will retrieve paginated objectTree and will handle concensus; Required tree should be made in application side.
func (a *Allocation) GetRefs(path, offsetPath, updatedDate, offsetDate, fileType, refType string, level, pageLimit int, opts ...ObjectTreeRequestOption) (*ObjectTreeResult, error) {
	if len(path) == 0 || !zboxutil.IsRemoteAbs(path) {
		return nil, errors.New("invalid_path", fmt.Sprintf("Absolute path required. Path provided: %v", path))
	}

	return a.getRefs(path, "", "", offsetPath, updatedDate, offsetDate, fileType, refType, level, pageLimit, opts...)
}

func (a *Allocation) GetRefsFromLookupHash(pathHash, offsetPath, updatedDate, offsetDate, fileType, refType string, level, pageLimit int) (*ObjectTreeResult, error) {
	if pathHash == "" {
		return nil, errors.New("invalid_lookup_hash", "lookup hash cannot be empty")
	}

	return a.getRefs("", pathHash, "", offsetPath, updatedDate, offsetDate, fileType, refType, level, pageLimit)

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
			RWMutex:         &sync.RWMutex{},
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
	listReq := &ListRequest{Consensus: Consensus{RWMutex: &sync.RWMutex{}}}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.blobbers = a.Blobbers
	listReq.fullconsensus = a.fullconsensus
	listReq.consensusThresh = a.consensusThreshold
	listReq.ctx = a.ctx
	listReq.remotefilepath = path
	_, _, ref, _ := listReq.getFileConsensusFromBlobbers()
	if ref != nil {
		result.Type = ref.Type
		result.Name = ref.Name
		result.Hash = ref.ActualFileHash
		result.LookupHash = ref.LookupHash
		result.MimeType = ref.MimeType
		result.Path = ref.Path
		result.Size = ref.Size
		result.NumBlocks = ref.NumBlocks
		result.EncryptedKey = ref.EncryptedKey
		result.Collaborators = ref.Collaborators
		result.ActualFileSize = ref.ActualFileSize
		result.ActualThumbnailHash = ref.ActualThumbnailHash
		result.ActualThumbnailSize = ref.ActualThumbnailSize
		if result.ActualFileSize > 0 {
			result.ActualNumBlocks = (ref.ActualFileSize + CHUNK_SIZE - 1) / CHUNK_SIZE
		}
		return result, nil
	}
	return nil, errors.New("file_meta_error", "Error getting the file meta data from blobbers")
}

func (a *Allocation) GetChunkReadSize(encrypt bool) int64 {
	chunkDataSize := int64(DefaultChunkSize)
	if encrypt {
		chunkDataSize -= (EncryptedDataPaddingSize + EncryptionHeaderSize)
	}
	return chunkDataSize * int64(a.DataShards)
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

	listReq := &ListRequest{Consensus: Consensus{RWMutex: &sync.RWMutex{}}}
	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.blobbers = a.Blobbers
	listReq.fullconsensus = a.fullconsensus
	listReq.consensusThresh = a.consensusThreshold
	listReq.ctx = a.ctx
	listReq.remotefilepathhash = lookupHash
	listReq.authToken = at
	_, _, ref, _ := listReq.getFileConsensusFromBlobbers()
	if ref != nil {
		result.Type = ref.Type
		result.Name = ref.Name
		result.Hash = ref.ActualFileHash
		result.LookupHash = ref.LookupHash
		result.MimeType = ref.MimeType
		result.Path = ref.Path
		result.Size = ref.Size
		result.NumBlocks = ref.NumBlocks
		result.ActualFileSize = ref.ActualFileSize
		result.ActualThumbnailHash = ref.ActualThumbnailHash
		result.ActualThumbnailSize = ref.ActualThumbnailSize
		if result.ActualFileSize > 0 {
			result.ActualNumBlocks = (result.ActualFileSize + CHUNK_SIZE - 1) / CHUNK_SIZE
		}
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
	listReq := &ListRequest{Consensus: Consensus{RWMutex: &sync.RWMutex{}}}
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
	return a.deleteFile(path, a.consensusThreshold, a.fullconsensus, zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1))
}

func (a *Allocation) deleteFile(path string, threshConsensus, fullConsensus int, mask zboxutil.Uint128) error {
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

	req := &DeleteRequest{consensus: Consensus{RWMutex: &sync.RWMutex{}}}
	req.allocationObj = a
	req.blobbers = a.Blobbers
	req.allocationID = a.ID
	req.allocationTx = a.Tx
	req.consensus.Init(threshConsensus, fullConsensus)
	req.ctx, req.ctxCncl = context.WithCancel(a.ctx)
	req.remotefilepath = path
	req.connectionID = zboxutil.NewConnectionId()
	req.deleteMask = mask
	req.maskMu = &sync.Mutex{}
	req.timestamp = int64(common.Now())
	err := req.ProcessDelete()
	return err
}

func (a *Allocation) createDir(remotePath string, threshConsensus, fullConsensus int, mask zboxutil.Uint128) error {
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
	timestamp := int64(common.Now())
	req := DirRequest{
		allocationObj: a,
		allocationID:  a.ID,
		allocationTx:  a.Tx,
		blobbers:      a.Blobbers,
		mu:            &sync.Mutex{},
		dirMask:       mask,
		connectionID:  zboxutil.NewConnectionId(),
		remotePath:    remotePath,
		wg:            &sync.WaitGroup{},
		timestamp:     timestamp,
		Consensus: Consensus{
			RWMutex:         &sync.RWMutex{},
			consensusThresh: threshConsensus,
			fullconsensus:   fullConsensus,
		},
		alreadyExists: make(map[uint64]bool),
	}
	req.ctx, req.ctxCncl = context.WithCancel(a.ctx)

	err := req.ProcessDir(a)
	return err
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

		httpreq, err := zboxutil.NewRevokeShareRequest(baseUrl, a.ID, a.Tx, query)
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

var ErrInvalidPrivateShare = errors.New("invalid_private_share", "private sharing is only available for encrypted file")

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

	if referenceType == fileref.FILE && refereeClientID != "" {
		fileMeta, err := a.GetFileMeta(path)
		if err != nil {
			return "", err
		}

		// private sharing is only available for encrypted file
		if fileMeta.EncryptedKey == "" {
			return "", ErrInvalidPrivateShare
		}
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
		httpreq, err := zboxutil.NewShareRequest(url, a.ID, a.Tx, body)
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
		RWMutex:         &sync.RWMutex{},
		consensus:       len(success),
		consensusThresh: a.DataShards,
		fullconsensus:   a.fullconsensus,
	}
	if !consensus.isConsensusOk() {
		return errors.New("", "consensus not reached")
	}
	return nil
}

func (a *Allocation) CancelDownload(remotepath string) error {
	if downloadReq, ok := a.downloadProgressMap[remotepath]; ok {
		downloadReq.isDownloadCanceled = true
		downloadReq.ctxCncl()
		return nil
	}
	return errors.New("remote_path_not_found", "Invalid path. No download in progress for the path "+remotepath)
}

func (a *Allocation) DownloadFromReader(
	remotePath, localPath, lookupHash, authTicket, contentMode string,
	verifyDownload bool,
	blocksPerMarker uint,
) error {

	finfo, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	if !finfo.IsDir() {
		return errors.New("invalid_path", "local path must be directory")
	}

	r, err := a.GetAllocationFileReader(remotePath, lookupHash, authTicket, contentMode, verifyDownload, blocksPerMarker)
	if err != nil {
		return err
	}

	sd := r.(*StreamDownload)

	fileName := filepath.Base(sd.remotefilepath)
	var localFPath string
	if contentMode == DOWNLOAD_CONTENT_THUMB {
		localFPath = filepath.Join(localPath, fileName, ".thumb")
	} else {
		localFPath = filepath.Join(localPath, fileName)
	}

	finfo, err = os.Stat(localFPath)

	var f *os.File
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(localFPath)
	} else {
		_, err = r.Seek(finfo.Size(), io.SeekStart)
		if err != nil {
			return err
		}
		f, err = os.OpenFile(localFPath, os.O_WRONLY|os.O_APPEND, 0644)
	}

	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 1024*KB)
	for {
		n, err := r.Read(buf)
		if err != nil && errors.Is(err, io.EOF) {
			_, err = f.Write(buf[:n])
			if err != nil {
				return err
			}
			break
		}
		_, err = f.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	return nil
}

// GetStreamDownloader will check file ref existence and returns an instance that provides
// io.ReadSeekerCloser interface
func (a *Allocation) GetAllocationFileReader(
	remotePath, lookupHash, authTicket, contentMode string,
	verifyDownload bool,
	blocksPerMarker uint,
) (io.ReadSeekCloser, error) {

	if !a.isInitialized() {
		return nil, notInitialized
	}
	//Remove content mode option
	remotePath = filepath.Clean(remotePath)
	var res *ObjectTreeResult
	var err error
	switch {
	case authTicket != "":
		res, err = a.GetRefsWithAuthTicket(authTicket, "", "", "", "", "regular", 0, 1)
	case remotePath != "":
		res, err = a.GetRefs(remotePath, "", "", "", "", "regular", 0, 1)
	case lookupHash != "":
		res, err = a.GetRefsFromLookupHash(lookupHash, "", "", "", "", "regular", 0, 1) //
	default:
		return nil, errors.New("invalid_path", "remote path or authticket is required")
	}

	if err != nil {
		return nil, err
	}

	if len(res.Refs) == 0 {
		return nil, errors.New("file_does_not_exist", "")
	}
	ref := &res.Refs[0]
	if ref.Type != fileref.FILE {
		return nil, errors.New("operation_not_supported", "downloading other than file is not supported")
	}

	if blocksPerMarker == 0 {
		blocksPerMarker = uint(numBlockDownloads)
	}

	sdo := &StreamDownloadOption{
		ContentMode:     contentMode,
		AuthTicket:      authTicket,
		VerifyDownload:  verifyDownload,
		BlocksPerMarker: blocksPerMarker,
	}

	return GetDStorageFileReader(a, ref, sdo)
}

func (a *Allocation) DownloadFileToFileHandlerFromAuthTicket(
	fileHandler sys.File,
	authTicket string,
	remoteLookupHash string,
	remoteFilename string,
	verifyDownload bool,
	status StatusCallback,
	isFinal bool,
	downloadReqOpts ...DownloadRequestOption,
) error {
	return a.downloadFromAuthTicket(fileHandler, authTicket, remoteLookupHash, 1, 0, numBlockDownloads,
		remoteFilename, DOWNLOAD_CONTENT_FULL, verifyDownload, status, isFinal, "", downloadReqOpts...)
}

func (a *Allocation) DownloadByBlocksToFileHandlerFromAuthTicket(
	fileHandler sys.File,
	authTicket string,
	remoteLookupHash string,
	startBlock, endBlock int64,
	numBlocks int,
	remoteFilename string,
	verifyDownload bool,
	status StatusCallback,
	isFinal bool,
	downloadReqOpts ...DownloadRequestOption,
) error {
	return a.downloadFromAuthTicket(fileHandler, authTicket, remoteLookupHash, startBlock, endBlock, numBlocks,
		remoteFilename, DOWNLOAD_CONTENT_FULL, verifyDownload, status, isFinal, "", downloadReqOpts...)
}

func (a *Allocation) DownloadThumbnailToFileHandlerFromAuthTicket(
	fileHandler sys.File,
	authTicket string,
	remoteLookupHash string,
	remoteFilename string,
	verifyDownload bool,
	status StatusCallback,
	isFinal bool,
) error {
	return a.downloadFromAuthTicket(fileHandler, authTicket, remoteLookupHash, 1, 0, numBlockDownloads,
		remoteFilename, DOWNLOAD_CONTENT_THUMB, verifyDownload, status, isFinal, "")
}

func (a *Allocation) DownloadThumbnailFromAuthTicket(
	localPath string,
	authTicket string,
	remoteLookupHash string,
	remoteFilename string,
	verifyDownload bool,
	status StatusCallback,
	isFinal bool,
	downloadReqOpts ...DownloadRequestOption,
) error {
	f, localFilePath, toKeep, err := a.prepareAndOpenLocalFile(localPath, remoteFilename)
	if err != nil {
		return err
	}
	downloadReqOpts = append(downloadReqOpts, WithFileCallback(func() {
		f.Close() //nolint: errcheck
	}))
	err = a.downloadFromAuthTicket(f, authTicket, remoteLookupHash, 1, 0, numBlockDownloads, remoteFilename,
		DOWNLOAD_CONTENT_THUMB, verifyDownload, status, isFinal, localFilePath, downloadReqOpts...)
	if err != nil {
		if !toKeep {
			os.Remove(localFilePath) //nolint: errcheck
		}
		f.Close() //nolint: errcheck
		return err
	}
	return nil
}

func (a *Allocation) DownloadFromAuthTicket(localPath string, authTicket string,
	remoteLookupHash string, remoteFilename string, verifyDownload bool, status StatusCallback, isFinal bool, downloadReqOpts ...DownloadRequestOption) error {
	f, localFilePath, toKeep, err := a.prepareAndOpenLocalFile(localPath, remoteFilename)
	if err != nil {
		return err
	}
	downloadReqOpts = append(downloadReqOpts, WithFileCallback(func() {
		f.Close() //nolint: errcheck
	}))
	err = a.downloadFromAuthTicket(f, authTicket, remoteLookupHash, 1, 0, numBlockDownloads, remoteFilename,
		DOWNLOAD_CONTENT_FULL, verifyDownload, status, isFinal, localFilePath, downloadReqOpts...)
	if err != nil {
		if !toKeep {
			os.Remove(localFilePath) //nolint: errcheck
		}
		f.Close() //nolint: errcheck
		return err
	}
	return nil
}

func (a *Allocation) DownloadFromAuthTicketByBlocks(localPath string,
	authTicket string, startBlock int64, endBlock int64, numBlocks int,
	remoteLookupHash string, remoteFilename string, verifyDownload bool,
	status StatusCallback, isFinal bool, downloadReqOpts ...DownloadRequestOption) error {

	f, localFilePath, toKeep, err := a.prepareAndOpenLocalFile(localPath, remoteFilename)
	if err != nil {
		return err
	}
	downloadReqOpts = append(downloadReqOpts, WithFileCallback(func() {
		f.Close() //nolint: errcheck
	}))
	err = a.downloadFromAuthTicket(f, authTicket, remoteLookupHash, startBlock, endBlock, numBlockDownloads, remoteFilename,
		DOWNLOAD_CONTENT_FULL, verifyDownload, status, isFinal, localFilePath, downloadReqOpts...)
	if err != nil {
		if !toKeep {
			os.Remove(localFilePath) //nolint: errcheck
		}
		f.Close() //nolint: errcheck
		return err
	}
	return nil
}

func (a *Allocation) downloadFromAuthTicket(fileHandler sys.File, authTicket string,
	remoteLookupHash string, startBlock int64, endBlock int64, numBlocks int,
	remoteFilename string, contentMode string, verifyDownload bool,
	status StatusCallback, isFinal bool, localFilePath string, downlaodReqOpts ...DownloadRequestOption) error {

	sEnc, err := base64.StdEncoding.DecodeString(authTicket)
	if err != nil {
		return errors.New("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	at := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return errors.New("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}

	if len(a.Blobbers) == 0 {
		return noBLOBBERS
	}

	downloadReq := &DownloadRequest{Consensus: Consensus{RWMutex: &sync.RWMutex{}}}
	downloadReq.maskMu = &sync.Mutex{}
	downloadReq.allocationID = a.ID
	downloadReq.allocationTx = a.Tx
	downloadReq.allocOwnerID = a.Owner
	downloadReq.allocOwnerPubKey = a.OwnerPublicKey
	downloadReq.ctx, downloadReq.ctxCncl = context.WithCancel(a.ctx)
	downloadReq.fileHandler = fileHandler
	downloadReq.localFilePath = localFilePath
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
	downloadReq.shouldVerify = verifyDownload
	downloadReq.fullconsensus = a.fullconsensus
	downloadReq.consensusThresh = a.consensusThreshold
	downloadReq.downloadQueue = make(downloadQueue, len(a.Blobbers))
	for i := 0; i < len(a.Blobbers); i++ {
		downloadReq.downloadQueue[i].timeTaken = 1000000
	}
	downloadReq.connectionID = zboxutil.NewConnectionId()
	downloadReq.completedCallback = func(remotepath string, remotepathHash string) {
		a.mutex.Lock()
		defer a.mutex.Unlock()
		delete(a.downloadProgressMap, remotepathHash)
	}
	downloadReq.fileCallback = func() {
		if downloadReq.fileHandler != nil {
			downloadReq.fileHandler.Close() //nolint: errcheck
		}
	}
	for _, opt := range downlaodReqOpts {
		opt(downloadReq)
	}
	a.mutex.Lock()
	a.downloadProgressMap[remoteLookupHash] = downloadReq
	if len(a.downloadRequests) > 0 {
		downloadReq.connectionID = a.downloadRequests[0].connectionID
	}
	a.downloadRequests = append(a.downloadRequests, downloadReq)
	if isFinal {
		downloadOps := a.downloadRequests
		a.downloadRequests = nil
		go func() {
			a.processReadMarker(downloadOps)
		}()
	}
	a.mutex.Unlock()
	return nil
}

func (a *Allocation) StartRepair(localRootPath, pathToRepair string, statusCB StatusCallback) error {
	if !a.isInitialized() {
		return notInitialized
	}

	listDir, err := a.ListDir(pathToRepair,
		WithListRequestForRepair(true),
		WithListRequestPageLimit(-1),
	)
	if err != nil {
		return err
	}
	a.CheckAllocStatus() //nolint:errcheck

	repairReq := &RepairRequest{
		listDir:       listDir,
		localRootPath: localRootPath,
		statusCB:      statusCB,
		repairPath:    pathToRepair,
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

// RepairAlloc repairs all the files in allocation
func (a *Allocation) RepairAlloc(statusCB StatusCallback) (err error) {
	var dir string
	if IsWasm {
		dir = "/tmp"
	} else {
		dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	return a.StartRepair(dir, "/", statusCB)
}

// Gets the size in bytes to repair allocation
func (a *Allocation) RepairSize(remotePath string) (RepairSize, error) {
	if !a.isInitialized() {
		return RepairSize{}, notInitialized
	}

	dir, err := a.ListDir(remotePath,
		WithListRequestForRepair(true),
		WithListRequestPageLimit(-1),
	)
	if err != nil {
		return RepairSize{}, err
	}

	repairReq := RepairRequest{
		allocation: a,
	}
	return repairReq.Size(context.Background(), dir)
}

func (a *Allocation) CancelUpload(remotePath string) error {
	cancelLock.Lock()
	cancelFunc, ok := CancelOpCtx[remotePath]
	cancelLock.Unlock()
	if !ok {
		return errors.New("remote_path_not_found", "Invalid path. No upload in progress for the path "+remotePath)
	} else {
		cancelFunc(fmt.Errorf("upload canceled by user"))
	}
	return nil
}

func (a *Allocation) PauseUpload(remotePath string) error {
	cancelLock.Lock()
	cancelFunc, ok := CancelOpCtx[remotePath]
	cancelLock.Unlock()
	if !ok {
		logger.Logger.Error("PauseUpload: remote path not found", remotePath)
		return errors.New("remote_path_not_found", "Invalid path. No upload in progress for the path "+remotePath)
	} else {
		logger.Logger.Info("PauseUpload: remote path found", remotePath)
		cancelFunc(ErrPauseUpload)
	}
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
		writePrice, err := v.Terms.WritePrice.ToToken()
		if err != nil {
			return 0, 0, err
		}
		if writePrice > maxWritePrice {
			maxWritePrice = writePrice
		}
		readPrice, err := v.Terms.ReadPrice.ToToken()
		if err != nil {
			return 0, 0, err
		}
		if readPrice > maxReadPrice {
			maxReadPrice = readPrice
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
		writePrice, err := v.Terms.WritePrice.ToToken()
		if err != nil {
			return 0, 0, err
		}
		if writePrice < minWritePrice || minWritePrice < 0 {
			minWritePrice = writePrice
		}
		readPrice, err := v.Terms.ReadPrice.ToToken()
		if err != nil {
			return 0, 0, err
		}
		if readPrice < minReadPrice || minReadPrice < 0 {
			minReadPrice = readPrice
		}
	}

	return minWritePrice, minReadPrice, nil
}

func (a *Allocation) GetMaxStorageCostFromBlobbers(size int64, blobbers []*BlobberAllocation) (float64, error) {
	var cost common.Balance // total price for size / duration

	for _, d := range blobbers {
		var err error
		cost, err = common.AddBalance(cost, a.uploadCostForBlobber(float64(d.Terms.WritePrice), size,
			a.DataShards, a.ParityShards))
		if err != nil {
			return 0.0, err
		}
	}

	return cost.ToToken()
}

func (a *Allocation) GetMaxStorageCost(size int64) (float64, error) {
	var cost common.Balance // total price for size / duration

	for _, d := range a.BlobberDetails {
		fmt.Printf("write price for blobber %f datashards %d parity %d\n",
			float64(d.Terms.WritePrice), a.DataShards, a.ParityShards)

		var err error
		cost, err = common.AddBalance(cost, a.uploadCostForBlobber(float64(d.Terms.WritePrice), size,
			a.DataShards, a.ParityShards))
		if err != nil {
			return 0.0, err
		}
	}
	fmt.Printf("Total cost %d\n", cost)
	return cost.ToToken()
}

func (a *Allocation) GetMinStorageCost(size int64) (common.Balance, error) {
	minW, _, err := a.GetMinWriteRead()
	if err != nil {
		return 0, err
	}

	return a.uploadCostForBlobber(minW, size, a.DataShards, a.ParityShards), nil
}

func (a *Allocation) uploadCostForBlobber(price float64, size int64, data, parity int) (
	cost common.Balance) {

	if data == 0 || parity == 0 {
		return 0.0
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

func (a *Allocation) SetConsensusThreshold() {
	a.consensusThreshold = a.DataShards
}

func (a *Allocation) UpdateWithRepair(
	size int64,
	extend bool,
	lock uint64,
	addBlobberId, addBlobberAuthTicket, removeBlobberId string,
	setThirdPartyExtendable bool, fileOptionsParams *FileOptionsParameters,
	statusCB StatusCallback,
) (string, error) {
	updatedAlloc, hash, isRepairRequired, err := a.UpdateWithStatus(size, extend, lock, addBlobberId, addBlobberAuthTicket, removeBlobberId, setThirdPartyExtendable, fileOptionsParams, statusCB)
	if err != nil {
		return hash, err
	}

	if isRepairRequired {
		if err := updatedAlloc.RepairAlloc(statusCB); err != nil {
			return hash, err
		}
	}

	return hash, nil
}

func (a *Allocation) UpdateWithStatus(
	size int64,
	extend bool,
	lock uint64,
	addBlobberId, addBlobberAuthTicket, removeBlobberId string,
	setThirdPartyExtendable bool, fileOptionsParams *FileOptionsParameters,
	statusCB StatusCallback,
) (*Allocation, string, bool, error) {
	var (
		alloc            *Allocation
		isRepairRequired bool
	)
	if lock > math.MaxInt64 {
		return alloc, "", isRepairRequired, errors.New("invalid_lock", "int64 overflow on lock value")
	}

	l.Logger.Info("Updating allocation")
	hash, _, err := UpdateAllocation(size, extend, a.ID, lock, addBlobberId, addBlobberAuthTicket, removeBlobberId, setThirdPartyExtendable, fileOptionsParams)
	if err != nil {
		return alloc, "", isRepairRequired, err
	}
	l.Logger.Info(fmt.Sprintf("allocation updated with hash: %s", hash))

	if addBlobberId != "" {
		l.Logger.Info("waiting for a minute for the blobber to be added to network")

		deadline := time.Now().Add(1 * time.Minute)
		for time.Now().Before(deadline) {
			alloc, err = GetAllocation(a.ID)
			if err != nil {
				l.Logger.Error("failed to get allocation")
				return alloc, hash, isRepairRequired, err
			}

			for _, blobber := range alloc.Blobbers {
				if addBlobberId == blobber.ID {
					l.Logger.Info("allocation updated successfully")
					a = alloc
					goto repair
				}
			}
			time.Sleep(1 * time.Second)
		}
		return alloc, "", isRepairRequired, errors.New("", "new blobber not found in the updated allocation")
	}

repair:
	l.Logger.Info("starting repair")

	shouldRepair := false
	if addBlobberId != "" {
		shouldRepair = true
	}

	if shouldRepair {
		isRepairRequired = true
	}

	return alloc, hash, isRepairRequired, nil
}
