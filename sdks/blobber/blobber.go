package blobber

import (
	"github.com/0chain/gosdk/sdks"
)

// Blobber blobber sdk client instance
type Blobber struct {
	BaseURLs []string
	*sdks.ZBox
}

// New create an sdk client instance given its configuration
//   - zbox zbox sdk client instance
//   - baseURLs base urls of the blobber
func New(zbox *sdks.ZBox, baseURLs ...string) *Blobber {
	b := &Blobber{
		BaseURLs: baseURLs,
		ZBox:     zbox,
	}

	return b
}
