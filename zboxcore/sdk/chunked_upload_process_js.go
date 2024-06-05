//go:build js && wasm
// +build js,wasm

package sdk

import "github.com/0chain/gosdk/wasmsdk/jsbridge"

func getWorkers() []*jsbridge.WasmWebWorker {
	return jsbridge.GetWorkers()
}

// processUpload process upload fragment to its blobber
func (su *ChunkedUpload) processUpload(chunkStartIndex, chunkEndIndex int,
	fileShards []blobberShards, thumbnailShards blobberShards,
	isFinal bool, uploadLength int64) error {

	return nil
}
