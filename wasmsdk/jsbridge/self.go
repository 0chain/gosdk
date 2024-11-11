//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"github.com/hack-pad/go-webworkers/worker"
)

var (
	selfWorker *worker.GlobalSelf
)

func NewSelfWorker() (*worker.GlobalSelf, error) {
	worker, err := worker.Self()
	if worker != nil {
		selfWorker = worker
	}
	return selfWorker, err
}

func GetSelfWorker() *worker.GlobalSelf {
	return selfWorker
}
