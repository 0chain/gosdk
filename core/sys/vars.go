// Package sys provides platform-independent interfaces to support webassembly runtime
package sys

import (
	"time"
)

var (
	//Files file system implementation on sdk. DiskFS doesn't work on webassembly. it should be initialized with common.NewMemFS()
	Files FS = NewDiskFS()

	//Sleep  pauses the current goroutine for at least the duration.
	//  time.Sleep will stop webassembly main thread. it should be bridged to javascript method on webassembly sdk
	Sleep = time.Sleep

	// Sign sign method. it should be initialized on different platform.
	Sign         SignFunc
	SignWithAuth SignFunc

	// Verify verify method. it should be initialized on different platform.
	Verify VerifyFunc

	// Verify verify method. it should be initialized on different platform.
	VerifyWith VerifyWithFunc

	Authorize AuthorizeFunc

	AuthCommon AuthorizeFunc
)

// SetAuthorize sets the authorize callback function
func SetAuthorize(auth AuthorizeFunc) {
	Authorize = auth
}

func SetAuthCommon(auth AuthorizeFunc) {
	AuthCommon = auth
}
