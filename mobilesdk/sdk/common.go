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

func Thumbnail(buf []byte, width, height int, options string) (imageutil.ConvertRes, error) {	
	return imageutil.Thumbnail(buf, width, height, options)
}