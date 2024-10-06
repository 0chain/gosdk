//go:build mobile
// +build mobile

package sdk

import (
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/imageutil"
)

// GetLookupHash get lookup hash with allocation id and path.
// Lookuphash is a hashed value of the augmentation of allocation id and remote path.
// It is used to identify the file in the blobbers.
//
//   - allocationID : allocation id
//   - remotePath : remote path
func GetLookupHash(allocationID string, remotePath string) string {
	return encryption.Hash(allocationID + ":" + remotePath)
}

func Thumbnail(buf []byte, width, height int) ([]byte, error) {	
	return imageutil.Thumbnail(buf, width, height)
}