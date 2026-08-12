//go:build js && wasm
// +build js,wasm

package main

import "errors"

var currentPlayerMW Player

// play starts playing a playable file or stream
//   - allocationID is the allocation id
//   - remotePath is the remote path of the file or stream
//   - authTicket is the auth ticket, in case of accessing as a shared file
//   - lookupHash is the lookup hash for the file
//   - isLive is the flag to indicate if the file is live or not
//
// playMW starts playing a playable file or stream and accepts a key for multi-wallet signing.
func playMW(allocationID, remotePath, authTicket, lookupHash string, isLive bool, key string) error {
	var err error

	if currentPlayerMW != nil {
		currentPlayerMW.Stop()
		currentPlayerMW = nil
	}

	if isLive {
		currentPlayerMW, err = createStreamPalyer(allocationID, remotePath, authTicket, lookupHash, key)
		if err != nil {
			return err
		}

	} else {
		currentPlayerMW, err = createFilePalyer(allocationID, remotePath, authTicket, lookupHash, key)
		if err != nil {
			return err
		}
	}

	return currentPlayerMW.Start()

}

// stop stops the current player
func stopMW() error {
	if currentPlayerMW != nil {
		currentPlayerMW.Stop()
	}

	currentPlayerMW = nil

	return nil
}

// getNextSegment gets the next segment of the current player
func getNextSegmentMW() ([]byte, error) {
	if currentPlayerMW == nil {
		return nil, errors.New("No player is available")
	}

	return currentPlayerMW.GetNext(), nil
}
