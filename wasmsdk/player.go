//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"net/http"
)

func Play(allocationID, remotePath, authTicket, lookupHash string) error {

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

	}

	//player is viewer via shared authticket

	return nil
}

func Stop() error {
	return nil
}
