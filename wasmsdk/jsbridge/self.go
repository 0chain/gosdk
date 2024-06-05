//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"github.com/hack-pad/go-webworkers/worker"
)

func NewSelfWorker() (*worker.GlobalSelf, error) {
	return worker.Self()
}
