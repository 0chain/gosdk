// Provides helper methods and data structures to work with images.
package imageutil

import (
	"bytes"
	"fmt"
	"github.com/disintegration/gift"
	"image"
	"image/gif"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/ccitt"
	_ "golang.org/x/image/riff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vector"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

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
func CreateThumbnail(buf []byte, width, height int) ([]byte, error) {
	// Decode the image from the buffer
	img, format, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	// Resize the image using Lanczos resampling

	g := gift.New(
		gift.Resize(width, height, gift.LanczosResampling),
	)
	thumbnail := image.NewRGBA(g.Bounds(img.Bounds()))
	g.Draw(thumbnail, img)
	//thumbnail := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	// Create a buffer to hold the new resized image
	var outBuffer bytes.Buffer

	// Encode the image back into the original format
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&outBuffer, thumbnail, nil)
	case "png":
		err = png.Encode(&outBuffer, thumbnail)
	case "gif":
		err = gif.Encode(&outBuffer, thumbnail, nil)
	case "bmp", "ccitt", "riff", "tiff", "vector", "vp8", "vp8l", "webp":
		err = jpeg.Encode(&outBuffer, thumbnail, nil) // Use JPEG as fallback since WebP/GIF encoding is less common in Go
	default:
		err = fmt.Errorf("unsupported image format")
	}

	if err != nil {
		return nil, err
	}

	// Return the resized image buffer
	return outBuffer.Bytes(), nil
}
