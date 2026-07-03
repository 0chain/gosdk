//go:build js && wasm
// +build js,wasm

package sdk

import (
	thrown "github.com/0chain/errors"
)

// processParallel (HashVersion 2) is not supported on the wasm build —
// CreateChunkedUpload forces hashVersion back to 1 when IsWasm, so this stub
// is never reached; it exists only to keep the shared process() compiling.
func (su *ChunkedUpload) processParallel() error {
	return thrown.New("upload_failed", "hashVersion 2 upload is not supported in wasm")
}
