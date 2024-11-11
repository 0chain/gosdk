//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

type StreamPlayer struct {
	sync.RWMutex
	allocationID string
	remotePath   string
	authTicket   string
	lookupHash   string

	isViewer      bool
	allocationObj *sdk.Allocation
	authTicketObj *marker.AuthTicket

	waitingToDownloadFiles      chan sdk.PlaylistFile
	latestWaitingToDownloadFile sdk.PlaylistFile

	downloadedFiles chan []byte
	ctx             context.Context
	cancel          context.CancelFunc
	prefetchQty     int

	timer jsbridge.Timer
}

func (p *StreamPlayer) Start() error {
	if p.cancel != nil {
		p.cancel()
	}

	p.ctx, p.cancel = context.WithCancel(context.TODO())
	p.waitingToDownloadFiles = make(chan sdk.PlaylistFile, p.prefetchQty)
	p.downloadedFiles = make(chan []byte, p.prefetchQty)

	p.timer = *jsbridge.NewTimer(5*time.Second, p.reloadList)
	p.timer.Start()

	go p.reloadList()
	go p.startDownload()

	return nil
}

func (p *StreamPlayer) Stop() {
	if p.cancel != nil {
		p.timer.Stop()
		p.cancel()
		p.cancel = nil
	}

	close(p.downloadedFiles)
}

func (p *StreamPlayer) download(it sdk.PlaylistFile) {
	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg, totalBytesMap: make(map[string]int)}
	wg.Add(1)

	fileName := it.Name
	localPath := filepath.Join(p.allocationID, fileName)

	fs, _ := sys.Files.Open(localPath)
	mf, _ := fs.(*sys.MemFile)

	downloader, err := sdk.CreateDownloader(p.allocationID, localPath, it.Path,
		sdk.WithAllocation(p.allocationObj),
		sdk.WithAuthticket(p.authTicket, p.lookupHash),
		sdk.WithFileHandler(mf))

	if err != nil {
		PrintError(err.Error())
		return
	}

	defer sys.Files.Remove(localPath) //nolint

	PrintInfo("playlist: downloading [", it.Path, "]")
	err = downloader.Start(statusBar, true)

	if err == nil {
		wg.Wait()
	} else {
		PrintError("playlist: download failed.", err.Error())
		return
	}
	if !statusBar.success {
		PrintError("playlist: download failed: unknown error")
		return
	}

	PrintInfo("playlist: downloaded [", it.Path, "]")

	withRecover(func() {
		if p.downloadedFiles != nil {
			p.downloadedFiles <- mf.Buffer
		}
	})
}

func (p *StreamPlayer) startDownload() {
	for {
		select {
		case <-p.ctx.Done():
			PrintInfo("playlist: download is cancelled")
			close(p.waitingToDownloadFiles)
			return
		case it, ok := <-p.waitingToDownloadFiles:
			if ok {
				if strings.HasSuffix(it.Name, ".ts") {
					p.download(it)
				}
			}
		}
	}
}

func (p *StreamPlayer) reloadList() {

	// `waiting to download files` buffer is too less, try to load latest list from remote
	if len(p.waitingToDownloadFiles) < p.prefetchQty {

		list, err := p.loadList()

		if err != nil {
			PrintError(err.Error())
			return
		}

		PrintInfo("playlist: ", len(list))

		for _, it := range list {
			PrintInfo("playlist: +", it.Path)

			if !withRecover(func() {
				if p.waitingToDownloadFiles != nil {
					p.waitingToDownloadFiles <- it
				}
			}) {
				// player is stopped
				return
			}

			p.Lock()
			p.latestWaitingToDownloadFile = it
			p.Unlock()
		}
	}
}

func (p *StreamPlayer) loadList() ([]sdk.PlaylistFile, error) {
	lookupHash := ""

	p.RLock()
	if p.latestWaitingToDownloadFile.Name != "" {
		lookupHash = p.latestWaitingToDownloadFile.LookupHash
	}
	p.RUnlock()

	if p.isViewer {
		//get list from authticket
		return sdk.GetPlaylistByAuthTicket(p.ctx, p.allocationObj, p.authTicket, p.lookupHash, lookupHash)
	}

	d, err := p.allocationObj.ListDir(p.remotePath)
	if err != nil {
		return nil, err
	}
	fmt.Printf("dir: %+v\n", d)
	return []sdk.PlaylistFile{
		sdk.PlaylistFile{
			Name:       d.Name,
			Path:       d.Path,
			LookupHash: d.LookupHash,
			NumBlocks:  d.ActualNumBlocks,
			Size:       d.Size,
			MimeType:   d.MimeType,
			Type:       d.Type,
		},
	}, nil

	// return []sdk.PlaylistFile{}, nil
	//get list from remote allocations's path
	// return sdk.GetPlaylist(p.ctx, p.allocationObj, p.remotePath, lookupHash)
}

func (p *StreamPlayer) GetNext() []byte {
	b, ok := <-p.downloadedFiles
	if ok {
		return b
	}

	return nil
}

// createStreamPalyer create player for remotePath
func createStreamPalyer(allocationID, remotePath, authTicket, lookupHash string) (*StreamPlayer, error) {

	player := &StreamPlayer{}
	player.prefetchQty = 3
	player.remotePath = remotePath
	player.authTicket = authTicket
	player.lookupHash = lookupHash

	//player is viewer
	if len(authTicket) > 0 {
		//player is viewer via shared authticket
		at, err := sdk.InitAuthTicket(authTicket).Unmarshall()

		if err != nil {
			PrintError(err)
			return nil, err
		}

		allocationObj, err := sdk.GetAllocationFromAuthTicket(authTicket)
		if err != nil {
			PrintError("Error fetching the allocation", err)
			return nil, err
		}

		player.isViewer = true
		player.allocationObj = allocationObj
		player.authTicketObj = at
		player.lookupHash = at.FilePathHash

		return player, nil

	}

	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	player.isViewer = false
	player.allocationObj = allocationObj

	return player, nil
}
