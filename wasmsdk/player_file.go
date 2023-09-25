//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"fmt"

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
	endBlock := startBlock + int64(p.numBlocks) - 1

	if endBlock > p.playlistFile.NumBlocks {
		endBlock = p.playlistFile.NumBlocks
	}
	fmt.Println("start:", startBlock, "end:", endBlock, "numBlocks:", p.numBlocks, "total:", p.playlistFile.NumBlocks)

	data, err := downloadBlocks(p.allocationID, p.remotePath, p.authTicket, p.lookupHash, p.numBlocks, startBlock, endBlock, "", true)
	if err != nil {
		PrintError(err.Error())
		return
	}
	withRecover(func() {
		if p.downloadedChunks != nil {
			p.downloadedChunks <- data
		}
	})
}

func (p *FilePlayer) startDownload() {
	fmt.Println("start download")
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
			fmt.Println("download start:", startBlock)
			p.download(startBlock)

			startBlock += int64(p.numBlocks)
			fmt.Println("download end, new start:", startBlock)

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

	d, err := p.allocationObj.ListDir(p.remotePath)
	if err != nil {
		fmt.Println("could not list dir:", p.remotePath)
		return nil, err
	}
	f := d.Children[0]
	fmt.Printf("dir: %+v\n", f)
	return &sdk.PlaylistFile{
		Name:       f.Name,
		Path:       f.Path,
		LookupHash: f.LookupHash,
		NumBlocks:  f.NumBlocks,
		Size:       f.Size,
		MimeType:   f.MimeType,
		Type:       f.Type,
	}, nil

	//get playlist file from remote allocations's path
	// return sdk.GetPlaylistFile(p.ctx, p.allocationObj, p.remotePath)
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
	player.numBlocks = 100
	player.allocationID = allocationID

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
