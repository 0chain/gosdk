// Provides helper methods and data structures to work with images.
package imageutil

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
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
	// image.Decode requires that you import the right image package.
	// Ignored return value is image format name.
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	// I've hard-coded a crop rectangle, start (0,0), end (100, 100).
	img, err = cropImage(img, image.Rect(0, 0, width, height))
	if err != nil {
		return nil, err
	}

	fd := &bytes.Buffer{}

	err = png.Encode(fd, img)
	if err != nil {
		return nil, err
	}

	return fd.Bytes(), nil
}

// cropImage takes an image and crops it to the specified rectangle.
func cropImage(img image.Image, crop image.Rectangle) (image.Image, error) {

	// img is an Image interface. This checks if the underlying value has a
	// method called SubImage. If it does, then we can use SubImage to crop the
	// image.
	simg, ok := img.(SubImager)
	if !ok {
		return nil, fmt.Errorf("image does not support cropping")
	}

	return simg.SubImage(crop), nil
}
