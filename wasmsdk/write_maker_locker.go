package main

import "sync"

type writeMarkerLocker struct {
	sync.Mutex
}

func (wml *writeMarkerLocker) Lock() error {
	//wml.Lock()

	return nil
}

func (wml *writeMarkerLocker) Unlock() {
	//wml.Unlock()
}
