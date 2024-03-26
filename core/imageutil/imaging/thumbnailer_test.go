package imaging_test

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/0chain/gosdk/core/imageutil/imaging"
	"github.com/stretchr/testify/require"
)

var (
	inpJpeg = filepath.Join("..", "resources", "input.jpeg")
	inpPng  = filepath.Join("..", "resources", "input.png")
	inpGif  = filepath.Join("..", "resources", "input.gif")
)

func TestThumbnail(t *testing.T) {

	type inp struct {
		filePath string
		width    int
		height   int
		resample imaging.ResampleFilter
	}

	inpData := []inp{
		{
			//jpeg file
			filePath: inpJpeg,
			width:    100,
			height:   200,
			resample: imaging.Lanczos,
		},
		{
			//png file
			filePath: inpPng,
			width:    200,
			height:   300,
			resample: imaging.Lanczos,
		},
		{
			//gif file
			filePath: inpGif,
			width:    100,
			height:   200,
			resample: imaging.Lanczos,
		},
		{
			//empty resample value
			filePath: inpJpeg,
			width:    100,
			height:   100,
		},
	}

	for _, i := range inpData {
		t.Logf("image: %v", i.filePath)
		buf, err := os.ReadFile(i.filePath)
		require.Nilf(t, err, "err reading file %s : %v", i.filePath, err)
		resJpeg, err := imaging.Thumbnail(buf, i.width, i.height, i.resample)
		require.Nilf(t, err, "err generating thumbnail: %v", err)
		_, format, err := image.Decode(bytes.NewReader(resJpeg))
		t.Logf("image format: %v", format)
		require.Nilf(t, err, "err decoding image: %v", err)
		require.Equal(t, "jpeg", format, "thumbnail should be in jpeg format")
	}
}
