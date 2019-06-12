package zcn

import (
	"context"
	"errors"
	"fmt"
	"math/bits"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"0chain/gosdk/logger"
	"0chain/gosdk/util"
)

var defaultLogLevel = logger.NONE
var Logger logger.Logger

const CHUNK_SIZE = (64 * 1024)

// Expected success rate is calculated (NumDataShards)*100/(NumDataShards+NumParityShards)
// Additional success percentage on top of expected success rate
const additionalSuccessRate = (10)

var (
	noBLOBBERS = errors.New("No Blobbers set in this allocation")
)

const (
	OpUpload   int = 0
	OpDownload int = 1
	OpRepair   int = 2
)

type bgSyncNotif int

const (
	bgSyncPause  bgSyncNotif = 0
	bgSyncResume bgSyncNotif = 1
	bgSyncStop   bgSyncNotif = 2
)

const (
	fileMetaBlobberCount string = "BLOBBERCOUNT"
)

type allocationImpl interface {
	SetConfig(clientJson, dirTreeJson, blobbersJson string, iDataShards, iParityShard int) error
	GetDirTree() string
	AddDir(path string) error
	ListDir(path string) string
	GetBlobbers() string
	UploadFile(localPath, remotePath string, statusCb StatusCallback) error
	UpdateFile(localPath, remotePath string, statusCb StatusCallback) error
	RepairFile(localPath, remotePath string, statusCb StatusCallback) error
	Commit() error
	DownloadFile(remotePath, localPath string, statusCb StatusCallback) error
	DownloadCancel()
	DeleteFile() error
	GetShareAuthToken(remotePath string, clientID string) string
	DownloadFileFromShareLink(localPath string, authTokenB64 string, statusCb StatusCallback) error
	GetFileStats(remotePath string) string
}

type StatusCallback interface {
	Started(allocationId, filePath string, op int, totalBytes int)
	InProgress(allocationId, filePath string, op int, completedBytes int)
	Completed(allocationId, filePath string, filename string, mimetype string, size int, op int)
}

type bgSync struct {
	pollIdx  int
	isPaused bool
	ticker   *time.Ticker
	c        chan bgSyncNotif
	cRet     chan bool
	wg       sync.WaitGroup
}

type Allocation struct {
	allocationId          string
	consensusThresh       float32
	consensus             float32
	encoder               *streamEncoder
	blobbers              []util.Blobber
	client                util.ClientConfig
	dirTree               util.FileDirInfo
	file                  util.FileConfig
	uploadDataCh          []chan []byte
	uploadMask            uint32
	downloadMask          uint32
	isUploadRepair        bool
	isUploadUpdate        bool
	isUploadCommitPending bool
	isRepairCommitPending bool
	isDownloadCanceled    bool
	wg                    sync.WaitGroup
	bg                    bgSync
}

func httpDo(ctx context.Context, cncl context.CancelFunc, req *http.Request, f func(*http.Response, error) error) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          1000,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   5,
	}

	client := &http.Client{Transport: transport}
	c := make(chan error, 1)
	go func() { c <- f(client.Do(req.WithContext(ctx))) }()
	// TODO: Check cncl context required in any case
	// defer cncl()
	select {
	case <-ctx.Done():
		transport.CancelRequest(req)
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func (obj *Allocation) check4AllocationRootChange(t time.Time) {
	incrIdx := func() { obj.bg.pollIdx = (obj.bg.pollIdx + 1) % len(obj.blobbers) }
	defer incrIdx()
	for i := 0; i < len(obj.blobbers); i++ {
		blobber := &obj.blobbers[obj.bg.pollIdx]
		Logger.Debug("Polling ", blobber.UrlRoot)
		lsData, err := obj.getDir(blobber, i, "/")
		if err == nil {
			if blobber.DirTree.Hash != lsData.AllocationRoot {
				Logger.Debug("Resync blobbers...")
				obj.syncAllBlobbers()
				Logger.Debug("Resync Completed!!")
			}
			break
		}
		incrIdx()
	}
	Logger.Debug("    Polling completed!!")
}

func (obj *Allocation) startBgSync() {
	Logger.Debug("Starting BG Sync ...")
	go func() {
		for {
			select {
			case notif := <-obj.bg.c:
				switch notif {
				case bgSyncPause:
					Logger.Debug("BG sync pause...")
					obj.bg.isPaused = true
				case bgSyncResume:
					Logger.Debug("BG sync resume...")
					obj.bg.isPaused = false
				case bgSyncStop:
					Logger.Debug("Bg sync stop...")
					obj.bg.isPaused = false
					obj.bg.ticker.Stop()
				}
				obj.bg.cRet <- true
				if notif == bgSyncStop {
					return
				}
			case t := <-obj.bg.ticker.C:
				if obj.bg.isPaused == false {
					obj.check4AllocationRootChange(t)
				}
			}
		}
	}()
}

func (obj *Allocation) pauseBgSync() {
	obj.bg.c <- bgSyncPause
	<-obj.bg.cRet
	Logger.Debug("BG Sync paused")
}

func (obj *Allocation) resumeBgSync() {
	obj.bg.c <- bgSyncResume
	<-obj.bg.cRet
	Logger.Debug("BG Sync resumed")
}

func (obj *Allocation) stopBgSync() {
	obj.bg.c <- bgSyncStop
	<-obj.bg.cRet
	Logger.Debug("BG Sync stopped")
}

func (obj *Allocation) initBgSync() {
	obj.bg.ticker = time.NewTicker(30 * time.Second)
	obj.bg.c = make(chan bgSyncNotif)
	obj.bg.cRet = make(chan bool)
	obj.bg.pollIdx = 0
	obj.check4AllocationRootChange(time.Now())
	// Start the BG Sync routine
	obj.startBgSync()
}

func (obj *Allocation) getConsensusRate() float32 {
	if obj.isUploadRepair {
		return (obj.consensus * 100) / float32(bits.OnesCount32(obj.uploadMask))
	} else {
		return (obj.consensus * 100) / float32(len(obj.blobbers))
	}
}

func (obj *Allocation) getConsensusRequiredForOk() float32 {
	return (obj.consensusThresh + additionalSuccessRate)
}

func (obj *Allocation) isConsensusOk() bool {
	return (obj.getConsensusRate() >= obj.getConsensusRequiredForOk())
}

func (obj *Allocation) isConsensusMin() bool {
	return (obj.getConsensusRate() >= obj.consensusThresh)
}

func init() {
	Logger.Init(defaultLogLevel)
}

// Returns the SDO version string
func GetVersion() string {
	return strVERSION
}

// lvl - 0 disabled; higher number (upto 4) more verbosity
func SetLogLevel(lvl int) {
	Logger.SetLevel(lvl)
}

// logFile - Log file
// verbose - true - console output; false - no console output
func SetLogFile(logFile string, verbose bool) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	Logger.SetLogFile(f, verbose)
	Logger.Info("******* SDK Version: ", strVERSION, " *******")
}

// Close log file
func CloseLog() {
	Logger.Close()
}

func CreateInstance(allocationId string) (*Allocation, error) {
	Logger.Info("**** New instance: ", allocationId, "****")
	return &Allocation{allocationId: allocationId, wg: sync.WaitGroup{}}, nil
}

func (obj *Allocation) SetConfig(clientJson, dirTreeJson, blobbersJson string, iDataShards, iParityShard int) error {
	var err error
	obj.client, err = util.GetClientConfig(clientJson)
	if err != nil {
		return err
	}
	// Parse dir tree if provided (or) create / directory
	if dirTreeJson != "" && dirTreeJson != "{}" {
		obj.dirTree, err = util.GetDirTreeFromJson(dirTreeJson)
		if err != nil {
			return err
		}
	} else {
		obj.dirTree = util.NewDirTree()
	}
	obj.blobbers, err = util.GetBlobbers(blobbersJson)
	if err != nil {
		return err
	}
	if iDataShards+iParityShard != len(obj.blobbers) {
		err = fmt.Errorf("Number of blobbers doesn't match with requested data and parity shards")
		return err
	}
	// Calculate minimum success threshold
	obj.consensusThresh = (float32(iDataShards) * 100) / float32(iDataShards+iParityShard)

	for i := 0; i < len(obj.blobbers); i++ {
		obj.blobbers[i].ConnObj.Reset()
		obj.blobbers[i].ConnObj.DirTree = obj.blobbers[i].DirTree
	}
	obj.encoder, err = newEncoder(iDataShards, iParityShard)
	if err != nil {
		return err
	}
	// Init the BG Sync routine
	obj.initBgSync()
	return nil
}

func (obj *Allocation) GetBlobbers() string {
	var s strings.Builder
	fmt.Fprintf(&s, "[\n")
	for i, blobber := range obj.blobbers {
		str := util.GetBlobberJson(&blobber)
		s.WriteString(str)
		if i != len(obj.blobbers)-1 {
			s.WriteString(",\n")
		}
	}
	fmt.Fprintf(&s, "\n]")
	return s.String()
}

func (obj *Allocation) GetDirTree() string {
	return util.GetJsonFromDirTree(&obj.dirTree)
}

func (obj *Allocation) AddDir(path string) error {
	_ = util.AddDir(&obj.dirTree, path)
	return nil
}

func (obj *Allocation) ListDir(path string) string {
	var s strings.Builder
	fmt.Fprintf(&s, "[\n")
	children := util.ListDir(&obj.dirTree, path)
	for i, child := range children {
		str := util.GetJsonFromDirTree(&child)
		s.WriteString(str)
		if i != len(children)-1 {
			s.WriteString(",\n")
		}
	}
	fmt.Fprintf(&s, "\n]")
	return s.String()
}
