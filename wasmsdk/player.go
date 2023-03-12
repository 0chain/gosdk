//go:build js && wasm
// +build js,wasm

package main

import "errors"

type Player interface {
	Start() error
	Stop()

	GetNext() []byte
}

var currentPlayer Player

func play(allocationID, remotePath, authTicket, lookupHash string, isLive bool) error {
	var err error

	if currentPlayer != nil {
		currentPlayer.Stop()
		currentPlayer = nil
	}

	if isLive {
		currentPlayer, err = createStreamPalyer(allocationID, remotePath, authTicket, lookupHash)
		if err != nil {
			return err
		}

	} else {
		currentPlayer, err = createFilePalyer(allocationID, remotePath, authTicket, lookupHash)
		if err != nil {
			return err
		}
	}

	return currentPlayer.Start()

}

func stop() error {
	if currentPlayer != nil {
		currentPlayer.Stop()
	}

	currentPlayer = nil

	return nil
}

func getNextSegment() ([]byte, error) {
	if currentPlayer == nil {
		return nil, errors.New("No player is available")
	}

	return currentPlayer.GetNext(), nil
}

func withRecover(send func()) (success bool) {
	defer func() {
		if recover() != nil {
			//recover panic from `send on closed channel`
			success = false
		}
	}()

	send()

	return true
}
