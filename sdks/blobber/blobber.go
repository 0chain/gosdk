// package blobber  wrap blobber's apis as sdk client
package blobber

import (
	"github.com/0chain/gosdk/sdks"
)

// Blobber blobber sdk client instance
type Blobber struct {
	BaseURLs []string
	*sdks.ZBox
}

func New(zbox *sdks.ZBox, baseURLs ...string) *Blobber {
	b := &Blobber{
		BaseURLs: baseURLs,
		ZBox:     zbox,
	}

	return b
}
