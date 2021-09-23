// package blobber  wrap blobber's apis as sdk client
package blobber

import (
	"github.com/0chain/gosdk/sdks"
)

// Blobber blobber sdk client instance
type Blobber struct {
	BaseURL string
	*sdks.ZBox
}

func New(zbox *sdks.ZBox, baseURL string) *Blobber {
	b := &Blobber{
		BaseURL: baseURL,
		ZBox:    zbox,
	}

	return b
}
