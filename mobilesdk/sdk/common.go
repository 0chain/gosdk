//go:build mobile
// +build mobile

package sdk

import (
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/imageutil"
)

// GetLookupHash get lookup hash with allocation id and path
// ## Inputs
//   - allocationID
//   - remotePath
//
// ## Outputs
//   - lookup_hash
func GetLookupHash(allocationID string, remotePath string) string {
	return encryption.Hash(allocationID + ":" + remotePath)
}

func Thumbnail(buf []byte, width, height int, crop string) ([]byte, error) {	
	return imageutil.ThumbnailVips(buf, width, height, imageutil.Crop(crop))
}