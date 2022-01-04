//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"errors"
	"sort"

	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

type Player struct {
	allocationID string
	remotePath   string
	authTicket   string
	lookupHash   string

	isOwner       bool
	allocationObj *sdk.Allocation
	authTicketObj *marker.AuthTicket

	todo        chan *sdk.ListResult
	reload      chan *sdk.ListResult
	latestTodo  *sdk.ListResult
	ctx         context.Context
	cancel      context.CancelFunc
	prefetchQty int
}

var currentPlayer *Player

func (p *Player) start() {
	if p.cancel != nil {
		p.cancel()
	}

	p.ctx, p.cancel = context.WithCancel(context.TODO())
	p.todo = make(chan *sdk.ListResult, 100)
	p.reload = make(chan *sdk.ListResult, p.prefetchQty)

	go p.reloadList()
	go p.startDownload()
	go p.nextTodo()
}

func (p *Player) stop() {
	p.cancel()
}

func (p *Player) download(it *sdk.ListResult) {

}

func (p *Player) startDownload() {
	for {
		select {
		case <-p.ctx.Done():
			PrintInfo("cancelled download")
			break
		case it := <-p.todo:
			p.download(it)
			PrintInfo("downloaded [", it.Name, "]")
			go p.nextTodo()
		}

	}
}

func (p *Player) nextTodo() {
	if len(p.todo) < p.prefetchQty {
		PrintInfo("reload list")
		p.reload <- p.latestTodo
	}
}

func (p *Player) reloadList() {
	for {
		select {
		case <-p.ctx.Done():
			PrintInfo("canceled reloadList")
			return
		case latestTodo := <-p.reload:

			list, err := p.loadList()

			if err != nil {
				PrintError(err.Error())
				continue
			}

			sort.Sort(SortedListResult(list))
			PrintInfo("got list ", len(list))

			for _, it := range list {
				if latestTodo == nil || (len(it.Name) > len(latestTodo.Name) || it.Name > latestTodo.Name) {
					PrintInfo("found [", it.Name, "]")
					p.todo <- it
					latestTodo = it
					p.latestTodo = it
				}
			}

		}
	}

}

func (p *Player) loadList() ([]*sdk.ListResult, error) {
	//get list from remoete allocations's path
	if p.isOwner {

		ref, err := p.allocationObj.ListDir(p.remotePath)
		if err != nil {
			return nil, err
		}

		return ref.Children, nil
	}

	//get list from authticket
	ref, err := p.allocationObj.ListDirFromAuthTicket(p.authTicket, p.lookupHash)
	if err != nil {
		return nil, err
	}

	return ref.Children, nil
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
	player.prefetchQty = 5

	//player is owner
	if len(remotePath) > 0 {
		if len(allocationID) == 0 {
			return nil, RequiredArg("allocationID")
		}

		allocationObj, err := sdk.GetAllocation(allocationID)
		if err != nil {
			PrintError("Error fetching the allocation", err)
			return nil, err
		}

		player.isOwner = true
		player.allocationObj = allocationObj

		return player, nil

	}

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

	player.isOwner = false
	player.allocationObj = allocationObj
	player.authTicketObj = at

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
