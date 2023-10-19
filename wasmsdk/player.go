//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"
	"fmt"
)

type Player interface {
	Start() error
	Stop()

	GetNext() []byte
	VideoSeek(pos int64) []byte
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
	fmt.Println("### cal get next segment")
	if currentPlayer == nil {
		return nil, errors.New("No player is available")
	}

	return currentPlayer.GetNext(), nil
}

func videoSeek(pos int64) ([]byte, error) {
	fmt.Println("### cal video seek")
	if currentPlayer == nil {
		return nil, errors.New("No player is available")
	}

	return currentPlayer.VideoSeek(pos), nil
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
