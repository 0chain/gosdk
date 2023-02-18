package zbox

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/transaction"

	"github.com/0chain/gosdk/zboxcore/sdk"
)

//StreamingService - holder for streaming service
type StreamingService struct {
	allocation *Allocation
	downloader *M3u8Downloader
}

//StreamingService - implementation of streaming service
type StreamingImpl interface {
	GetFirstSegment(localPath, remotePath, tmpPath string, delay, maxSegments int) (string, error)
	PlayStreaming(localPath, remotePath, authTicket, lookupHash, initSegment string, delay int, statusCb StatusCallbackWrapped) error
	Stop() error
	GetCurrentManifest() string
}

//CreateStreamingService - creating streaming service instance
func CreateStreamingService(allocation *Allocation) StreamingImpl {
	return &StreamingService{
		allocation: allocation,
	}
}

// GetFirstSegment - getting the amount of segments in maxSegments for very first playback
func (s *StreamingService) GetFirstSegment(localPath, remotePath, tmpPath string, delay, maxSegments int) (string, error) {
	if len(remotePath) == 0 {
		return "", errors.New("Error: remotepath / authticket flag is missing")
	}

	if len(localPath) == 0 {
		return "", errors.New("Error: localpath is missing")
	}

	dir := path.Dir(localPath)
	file, err := os.Create(localPath)

	if err != nil {
		return "", err
	}

	downloader := &M3u8Downloader{
		localDir:      dir,
		localPath:     localPath,
		remotePath:    remotePath,
		authTicket:    "",
		allocationID:  s.allocation.ID,
		rxPay:         false,
		downloadQueue: make(chan MediaItem, 100),
		playlist:      NewMediaPlaylist(delay, file),
		done:          make(chan error, 1),
	}

	downloader.allocationObj = s.allocation.sdkAllocation

	listResult, err := s.allocation.sdkAllocation.ListDir(downloader.remotePath)
	list := listResult.Children
	if err != nil {
		return "", err
	}

	downloader.Lock()
	n := len(downloader.items)
	max := maxSegments
	latestItem := ""

	sort.Slice(list, func(i, j int) bool {
		return GetNumber(list[i].Name) < GetNumber(list[j].Name)
	})

	if n < max {
		for i := n; i < max; i++ {
			if i > len(list) {
				continue
			}

			item := MediaItem{
				Name: list[i].Name,
				Path: list[i].Path,
			}
			downloader.items = append(downloader.items, item)
			downloader.playlist.Append(item.Name)
			downloader.playlist.Wait = append(downloader.playlist.Wait, item.Name)

			downloader.localDir = tmpPath

			if _, err = downloader.download(item); err == nil {
				fmt.Println(err)
			}
			latestItem = item.Name
			downloader.addToDownload(item)
		}
	}
	downloader.Unlock()

	downloader.playlist.Writer.Truncate(0)
	downloader.playlist.Writer.Seek(0, 0)
	downloader.playlist.Writer.Write(downloader.playlist.Encode())
	downloader.playlist.Writer.Sync()

	return latestItem, nil
}

// PlayStreaming - start streaming playback
func (s *StreamingService) PlayStreaming(localPath, remotePath, authTicket, lookupHash, initSegment string, delay int, statusCb StatusCallbackWrapped) error {
	downloader, err := createM3u8Downloader(localPath, remotePath, authTicket, s.allocation.ID, lookupHash, initSegment, false, delay)
	if err != nil {
		return err
	}

	downloader.status = statusCb
	s.downloader = downloader

	go downloader.start()
	return nil
}

// TODO
func (s *StreamingService) Stop() error {
	return nil
}

func (s *StreamingService) GetCurrentManifest() string {
	return string(s.downloader.playlist.Encode())
}

// M3u8Downloader download files from blobber's dir, and build them into a local m3u8 playlist
type M3u8Downloader struct {
	sync.RWMutex

	localDir     string
	localPath    string
	remotePath   string
	authTicket   string
	allocationID string
	rxPay        bool

	allocationObj *sdk.Allocation

	lookupHash    string
	items         []MediaItem
	downloadQueue chan MediaItem
	playlist      *MediaPlaylist
	done          chan error
	status        StatusCallbackWrapped
	initSegment   string
}

func createM3u8Downloader(localPath, remotePath, authTicket, allocationID, lookupHash, initSegment string, rxPay bool, delay int) (*M3u8Downloader, error) {
	if len(remotePath) == 0 && (len(authTicket) == 0) {
		return nil, errors.New("Error: remotepath / authticket flag is missing")
	}

	if len(localPath) == 0 {
		return nil, errors.New("Error: localpath is missing")
	}
	dir := path.Dir(localPath)

	file, err := os.Create(localPath)

	if err != nil {
		return nil, err
	}

	downloader := &M3u8Downloader{
		localDir:      dir,
		localPath:     localPath,
		remotePath:    remotePath,
		authTicket:    authTicket,
		allocationID:  allocationID,
		rxPay:         rxPay,
		downloadQueue: make(chan MediaItem, 100),
		playlist:      NewMediaPlaylist(delay, file),
		done:          make(chan error, 1),
		initSegment:   initSegment,
	}

	if len(remotePath) > 0 {
		if len(allocationID) == 0 { // check if the flag "path" is set
			return nil, errors.New("Error: allocation flag is missing") // If not, we'll let the user know
		}

		allocationObj, err := sdk.GetAllocation(allocationID)

		if err != nil {
			return nil, fmt.Errorf("Error fetching the allocation: %s", err)
		}

		downloader.allocationObj = allocationObj

	} else if len(authTicket) > 0 {
		allocationObj, err := sdk.GetAllocationFromAuthTicket(authTicket)
		if err != nil {
			return nil, fmt.Errorf("Error fetching the allocation: %s", err)
		}

		downloader.allocationObj = allocationObj

		at := sdk.InitAuthTicket(authTicket)
		isDir, err := at.IsDir()
		if isDir && len(lookupHash) == 0 {
			lookupHash, err = at.GetLookupHash()
			if err != nil {
				return nil, fmt.Errorf("Error getting the lookuphash from authticket: %s", err)
			}

			downloader.lookupHash = lookupHash
		}
		if !isDir {
			return nil, fmt.Errorf("Invalid operation. Auth ticket is not for a directory: %s", err)
		}
	}

	return downloader, nil
}

// Start start to download ,and build playlist
func (d *M3u8Downloader) start() error {
	d.status.Started(d.allocationID, d.localPath, 0, 0)

	go d.autoRefreshList()
	go d.autoDownload()
	go d.playlist.Play()

	err := <-d.done

	return err
}

func (d *M3u8Downloader) autoRefreshList() {
	for {
		list, err := d.getList()
		if err != nil {
			continue
		}

		d.Lock()
		n := len(d.items)
		max := len(list)

		initId := n
		if len(d.initSegment) > 0 {
			initId = sort.Search(len(list), func(i int) bool {
				return list[i].Name > d.initSegment
			})
			if initId != len(list) { // only if found
				if n == max || n > initId {
					initId = n
				}
			}
		}

		if initId < max {
			for i := initId; i < max; i++ {
				item := MediaItem{
					Name: list[i].Name,
					Path: list[i].Path,
				}

				d.items = append(d.items, item)
				d.addToDownload(item)
			}
		}
		d.Unlock()

		time.Sleep(1 * time.Second)
	}
}

func (d *M3u8Downloader) autoDownload() {
	for {
		item := <-d.downloadQueue

		localPath := d.localDir + string(os.PathSeparator) + item.Name
		_, err := os.Stat(localPath)
		if err == nil {
			d.playlist.Append(item.Name)
			continue
		}

		for i := 0; i < 3; i++ {
			if path, err := d.download(item); err == nil {
				d.playlist.Append(item.Name)
				d.status.InProgress(d.allocationID, path, 1, len(d.items), nil)
				break
			}
		}
	}
}

func (d *M3u8Downloader) addToDownload(item MediaItem) {
	d.downloadQueue <- item
}

func (d *M3u8Downloader) download(item MediaItem) (string, error) {
	wg := &sync.WaitGroup{}
	statusBar := &StatusBarMocked{wg: wg}
	wg.Add(1)

	localPath := d.localDir + string(os.PathSeparator) + item.Name
	remotePath := item.Path

	if len(d.remotePath) > 0 {
		err := d.allocationObj.DownloadFile(localPath, remotePath, false, statusBar)
		if err != nil {
			return "", err
		}
		wg.Wait()
	}

	if !statusBar.success {
		return "", statusBar.err
	}

	return localPath, nil
}

func (d *M3u8Downloader) getList() ([]*sdk.ListResult, error) {
	var list []*sdk.ListResult

	if len(d.remotePath) > 0 {
		ref, err := d.allocationObj.ListDir(d.remotePath)
		if err != nil {
			return nil, err
		}

		list = ref.Children
	}

	if len(d.authTicket) > 0 {
		ref, err := d.allocationObj.ListDirFromAuthTicket(d.authTicket, d.lookupHash)
		if err != nil {
			return nil, err
		}

		list = ref.Children
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("files not found")
	}

	sort.Slice(list, func(i, j int) bool {
		return GetNumber(list[i].Name) < GetNumber(list[j].Name)
	})

	return list, nil
}

// MediaItem is .ts file
type MediaItem struct {
	Name string
	Path string
}

type StatusBarMocked struct {
	wg      *sync.WaitGroup
	success bool
	err     error
}

func (s *StatusBarMocked) Error(allocationID string, filePath string, op int, err error) {
	s.success = false
	s.err = err
	s.wg.Done()
}

func (s *StatusBarMocked) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	s.success = true
	s.wg.Done()
}

func (s *StatusBarMocked) CommitMetaCompleted(request, response string, txn *transaction.Transaction, err error) {
}

func (s *StatusBarMocked) RepairCompleted(filesRepaired int) {}

func (s *StatusBarMocked) Started(allocationId, filePath string, op int, totalBytes int) {}

func (s *StatusBarMocked) InProgress(allocationId, filePath string, op int, completedBytes int, data []byte) {
}
