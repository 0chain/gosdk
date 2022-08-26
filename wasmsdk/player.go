//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

type Player struct {
	allocationID string
	remotePath   string
	authTicket   string
	lookupHash   string

	isViewer      bool
	allocationObj *sdk.Allocation
	authTicketObj *marker.AuthTicket

	todoQueue   chan sdk.PlaylistFile
	reloadQueue chan sdk.PlaylistFile
	latestTodo  sdk.PlaylistFile

	downloadedQueue chan []byte
	ctx             context.Context
	cancel          context.CancelFunc
	prefetchQty     int
}

var currentPlayer *Player

func (p *Player) start() {
	if p.cancel != nil {
		p.cancel()
	}

	p.ctx, p.cancel = context.WithCancel(context.TODO())
	p.todoQueue = make(chan sdk.PlaylistFile, 100)
	p.reloadQueue = make(chan sdk.PlaylistFile, p.prefetchQty)
	p.downloadedQueue = make(chan []byte, p.prefetchQty)

	go p.reloadList()
	go p.startDownload()
	go p.nextTodo()
}

func (p *Player) stop() {
	p.cancel()
}

func (p *Player) download(it sdk.PlaylistFile) {
	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg}
	wg.Add(1)

	fileName := it.Name
	localPath := filepath.Join(p.allocationID, fileName)

	downloader, err := sdk.CreateDownloader(p.allocationID, localPath, it.Path,
		sdk.WithAllocation(p.allocationObj),
		sdk.WithAuthticket(p.authTicket, p.lookupHash))

	if err != nil {
		PrintError(err.Error())
		return
	}

	defer sys.Files.Remove(localPath) //nolint

	PrintInfo("playlist: downloading [", it.Path, "]")
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

	PrintInfo("playlist: downloaded [", it.Path, "]")
	fs, _ := sys.Files.Open(localPath)

	mf, _ := fs.(*sys.MemFile)

	//AppendVideo(mf.Buffer.Bytes())

	p.downloadedQueue <- mf.Buffer.Bytes()

}

func (p *Player) startDownload() {
	for {
		select {
		case <-p.ctx.Done():
			PrintInfo("playlist: download is cancelled")
			close(p.todoQueue)
			close(p.reloadQueue)
			return
		case it := <-p.todoQueue:
			p.download(it)

			go p.nextTodo()
		}

	}
}

func (p *Player) nextTodo() {
	if len(p.todoQueue) < p.prefetchQty {
		PrintInfo("playlist: reload")

		p.reloadQueue <- p.latestTodo

	}
}

func (p *Player) reloadList() {
	for {
		select {
		case <-p.ctx.Done():
			PrintInfo("playlist: reload is canceled")
			return
		case <-p.reloadQueue:

			list, err := p.loadList()

			if len(list) == 0 {
				sys.Sleep(3 * time.Second)
				go p.nextTodo()
				continue
			}

			if err != nil {
				PrintError(err.Error())
				continue
			}

			PrintInfo("playlist: ", len(list))

			for _, it := range list {
				PrintInfo("playlist: +", it.Path)
				p.latestTodo = it
				p.todoQueue <- it
			}
		}
	}

}

func (p *Player) loadList() ([]sdk.PlaylistFile, error) {
	lookupHash := ""

	if p.latestTodo.Name != "" {
		lookupHash = p.latestTodo.LookupHash
	}

	if p.isViewer {
		//get list from authticket
		return sdk.GetPlaylistByAuthTicket(p.ctx, p.allocationObj, p.authTicket, p.lookupHash, lookupHash)
	}

	//get list from remote allocations's path
	return sdk.GetPlaylist(p.ctx, p.allocationObj, p.remotePath, lookupHash)
}

func (p *Player) getNextSegment() []byte {
	return <-p.downloadedQueue
}

// createPalyer create player for remotePath
func createPalyer(allocationID, remotePath, authTicket, lookupHash string) (*Player, error) {

	player := &Player{}
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

func Play(allocationID, remotePath, authTicket, lookupHash string) error {

	player, err := createPalyer(allocationID, remotePath, authTicket, lookupHash)
	if err != nil {
		return err
	}

	if currentPlayer != nil {
		return errors.New("please stop current player first")
	}

	currentPlayer = player
	go currentPlayer.start()

	return nil
}

func Stop() error {

	if currentPlayer == nil {
		return errors.New("No player is available")
	}

	currentPlayer.stop()
	currentPlayer = nil

	return nil
}

func GetNextSegment() ([]byte, error) {
	if currentPlayer == nil {
		return nil, errors.New("No player is available")
	}

	return currentPlayer.getNextSegment(), nil
}
