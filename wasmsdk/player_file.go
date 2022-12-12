//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

type FilePlayer struct {
	allocationID string
	remotePath   string
	authTicket   string
	lookupHash   string
	numBlocks    int

	isViewer      bool
	allocationObj *sdk.Allocation
	authTicketObj *marker.AuthTicket
	playlistFile  *sdk.PlaylistFile

	downloadedChunks chan []byte
	ctx              context.Context
	cancel           context.CancelFunc
	prefetchQty      int
}

func (p *FilePlayer) Start() error {
	if p.cancel != nil {
		p.cancel()
	}

	p.ctx, p.cancel = context.WithCancel(context.TODO())

	file, err := p.loadPlaylistFile()
	if err != nil {
		return err
	}

	p.playlistFile = file

	p.downloadedChunks = make(chan []byte, p.prefetchQty)

	go p.startDownload()

	return nil
}

func (p *FilePlayer) Stop() {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
}

func (p *FilePlayer) download(startBlock int64) {
	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg}
	wg.Add(1)

	endBlock := startBlock + int64(p.numBlocks) - 1

	if endBlock > p.playlistFile.NumBlocks {
		endBlock = p.playlistFile.NumBlocks
	}

	fileName := strconv.FormatInt(startBlock, 10) + "-" + strconv.FormatInt(endBlock, 10) + "-" + p.playlistFile.Name
	localPath := filepath.Join(p.allocationID, fileName)

	downloader, err := sdk.CreateDownloader(p.allocationID, localPath, p.remotePath,
		sdk.WithAllocation(p.allocationObj),
		sdk.WithAuthticket(p.authTicket, p.lookupHash),
		sdk.WithBlocks(startBlock, endBlock, p.numBlocks))

	if err != nil {
		PrintError(err.Error())
		return
	}

	defer sys.Files.Remove(localPath) //nolint

	PrintInfo("playlist: downloading blocks[", p.playlistFile.Name, ":", startBlock, "-", endBlock, "]")
	err = downloader.Start(statusBar)

	if err == nil {
		wg.Wait()
	} else {
		PrintError("Download failed.", err.Error())
		return
	}
	if !statusBar.success {
		PrintError("Download failed: unknown error")
		return
	}

	PrintInfo("playlist: downloaded blocks[", p.playlistFile.Name, ":", startBlock, "-", endBlock, "]")
	fs, _ := sys.Files.Open(localPath)

	mf, _ := fs.(*sys.MemFile)

	withRecover(func() {
		if p.downloadedChunks != nil {
			p.downloadedChunks <- mf.Buffer.Bytes()
		}
	})
}

func (p *FilePlayer) startDownload() {
	if p.playlistFile.NumBlocks < 1 {
		PrintError("playlist: numBlocks is invalid")
		return
	}
	var startBlock int64 = 1
	for {
		select {
		case <-p.ctx.Done():
			PrintInfo("playlist: download is cancelled")
			return
		default:
			p.download(startBlock)

			startBlock += int64(p.numBlocks)

			if startBlock > p.playlistFile.NumBlocks {

				go func() {
					// trigger js to close stream
					p.downloadedChunks <- nil
				}()
				return
			}

		}
	}

}

func (p *FilePlayer) loadPlaylistFile() (*sdk.PlaylistFile, error) {

	if p.isViewer {
		//get playlist file from auth ticket
		return sdk.GetPlaylistFileByAuthTicket(p.ctx, p.allocationObj, p.authTicket, p.lookupHash)
	}

	//get playlist file from remote allocations's path
	return sdk.GetPlaylistFile(p.ctx, p.allocationObj, p.remotePath)
}

func (p *FilePlayer) GetNext() []byte {
	b, ok := <-p.downloadedChunks
	if ok {
		return b
	}

	return nil
}

// createFilePalyer create player for remotePath
func createFilePalyer(allocationID, remotePath, authTicket, lookupHash string) (*FilePlayer, error) {

	player := &FilePlayer{}
	player.prefetchQty = 3
	player.remotePath = remotePath
	player.authTicket = authTicket
	player.lookupHash = lookupHash
	player.numBlocks = 10

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
