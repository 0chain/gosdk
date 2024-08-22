//go:build js && wasm
// +build js,wasm

package main

import "errors"

// Player is the interface for a file player
type Player interface {
	Start() error
	Stop()

	GetNext() []byte
}

var currentPlayer Player

// play starts playing a playable file or stream
//   - allocationID is the allocation id
//   - remotePath is the remote path of the file or stream
//   - authTicket is the auth ticket, in case of accessing as a shared file
//   - lookupHash is the lookup hash for the file
//   - isLive is the flag to indicate if the file is live or not
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

// stop stops the current player
func stop() error {
	if currentPlayer != nil {
		currentPlayer.Stop()
	}

	currentPlayer = nil

	return nil
}

// getNextSegment gets the next segment of the current player
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
