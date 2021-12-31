//go:build js && wasm
// +build js,wasm

package main

import (
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
}

func (p *Player) getList() {

}

func CreatePalyer(allocationID, remotePath, authTicket, lookupHash string) error {
	player := &Player{}

	//player is owner
	if len(remotePath) > 0 {
		if len(allocationID) == 0 {
			return RequiredArg("allocationID")
		}

		allocationObj, err := sdk.GetAllocation(allocationID)
		if err != nil {
			PrintError("Error fetching the allocation", err)
			return err
		}

		player.isOwner = true
		player.allocationObj = allocationObj

		// ref, err := allocationObj.ListDir(remotePath)
		// if err != nil {
		// 	return err
		// }

		// ref.Children, nil

	}

	//player is viewer via shared authticket

	at, err := sdk.InitAuthTicket(authTicket).Unmarshall()

	if err != nil {
		PrintError(err)
		return err
	}

	allocationObj, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return err
	}

	player.isOwner = false
	player.allocationObj = allocationObj
	player.authTicketObj = at

	// //get list from authticket
	// ref, err := allocationObj.ListDirFromAuthTicket(authTicket, lookupHash)
	// if err != nil {
	// 	return nil, err
	// }

	return nil

}

func Play(allocationID, remotePath, authTicket, lookupHash string) error {

	return nil
}

func Stop() error {
	return nil
}
