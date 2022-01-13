//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
	"sync"

	"github.com/0chain/gosdk/core/common"
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

	todoQueue   chan *sdk.ListResult
	reloadQueue chan *sdk.ListResult
	latestTodo  *sdk.ListResult

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
	p.todoQueue = make(chan *sdk.ListResult, 100)
	p.reloadQueue = make(chan *sdk.ListResult, p.prefetchQty)
	p.downloadedQueue = make(chan []byte, p.prefetchQty)

	go p.reloadList()
	go p.startDownload()
	go p.nextTodo()
}

func (p *Player) stop() {
	p.cancel()
}

func (p *Player) download(it *sdk.ListResult) {
	//Download(p.allocationID, remotePath string, authTicket string, lookupHash string, downloadThumbnailOnly bool, rxPay bool, autoCommit bool)
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

	defer sdk.FS.Remove(localPath) //nolint

	PrintInfo("downloading [", it.Path, "]")
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

	PrintInfo("downloaded [", it.Path, "]")
	fs, _ := sdk.FS.Open(localPath)

	mf, _ := fs.(*common.MemFile)

	//AppendVideo(mf.Buffer.Bytes())

	p.downloadedQueue <- mf.Buffer.Bytes()

}

func (p *Player) startDownload() {
	for {
		select {
		case <-p.ctx.Done():
			PrintInfo("cancelled download")
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
		PrintInfo("reload list")
		p.reloadQueue <- p.latestTodo
	}
}

func (p *Player) reloadList() {
	for {
		select {
		case <-p.ctx.Done():
			PrintInfo("canceled reloadList")
			return
		case latestTodo := <-p.reloadQueue:

			list, err := p.loadList()

			if err != nil {
				PrintError(err.Error())
				continue
			}

			sort.Sort(SortedListResult(list))
			PrintInfo("got list ", len(list))

			for _, it := range list {
				if latestTodo == nil || (len(it.Name) > len(latestTodo.Name) || it.Name > latestTodo.Name) {
					PrintInfo("found [", it.Path, "]")
					p.latestTodo = it
					p.todoQueue <- it
					latestTodo = it
				}
			}

		}
	}

}

func (p *Player) loadList() ([]*sdk.ListResult, error) {

	if p.isViewer {
		//get list from authticket
		ref, err := p.allocationObj.ListDirFromAuthTicket(p.authTicket, p.lookupHash)
		if err != nil {
			return nil, err
		}

		return ref.Children, nil
	}

	//get list from remote allocations's path
	ref, err := p.allocationObj.ListDir(p.remotePath)
	if err != nil {
		return nil, err
	}

	return ref.Children, nil
}

func (p *Player) getNextSegment() []byte {
	return <-p.downloadedQueue
}

// SortedListResult sort files order by time
type SortedListResult []*sdk.ListResult

func (a SortedListResult) Len() int {
	return len(a)
}
func (a SortedListResult) Less(i, j int) bool {

	l := a[i]
	r := a[j]

	if len(l.Name) < len(r.Name) {
		return true
	}

	if len(l.Name) > len(r.Name) {
		return false
	}

	return l.Name < r.Name
}
func (a SortedListResult) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
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
