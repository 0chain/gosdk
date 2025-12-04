//go:build mobile
// +build mobile

package sdk

import (
	"github.com/0chain/gosdk/core/imageutil"
)

// CreateThumbnail create thumbnail of an image buffer. It supports
//   - png
//   - jpeg
//   - gif
//   - bmp
//   - ccitt
//   - riff
//   - tiff
//   - vector
//   - vp8
//   - vp8l
//   - webp
//
// ## Inputs
//   - buf: image buffer as byte array
//   - width: thumbnail width
//   - height: thumbnail height
//
// ## Outputs
//   - thumbnail image buffer as byte array
//   - error
func CreateThumbnail(buf []byte, width, height int) ([]byte, error) {
	return imageutil.CreateThumbnail(buf, width, height)
}
