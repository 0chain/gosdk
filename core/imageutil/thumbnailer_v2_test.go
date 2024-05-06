package imageutil_test

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/0chain/gosdk/core/imageutil"
	"github.com/stretchr/testify/require"
)

var (
	inpJpeg    = filepath.Join("resources", "input.jpeg")
	inpPng     = filepath.Join("resources", "input.png")
	inpSvg     = filepath.Join("resources", "input.svg")
)

func TestThumbnailVips(t *testing.T) {

	type inp struct {
		filePath string
		width    int
		height   int
		crop     imageutil.Crop
	}

	inpData := []inp{
		{
			//jpeg file
			filePath: inpJpeg,
			width:    100,
			height:   200,
			crop:     imageutil.All,
		},
		{
			//png file
			filePath: inpPng,
			width:    200,
			height:   300,
			crop:     imageutil.Attention,
		},
		{
			//svg file
			filePath: inpSvg,
			width:    100,
			height:   200,
			crop:     imageutil.Centre,
		},
		{
			//empty crop value
			filePath: inpJpeg,
			width:    100,
			height:   100,
		},
	}

	for _, i := range inpData {
		t.Logf("image: %v", i.filePath)
		buf, err := os.ReadFile(i.filePath)
		require.Nilf(t, err, "err reading file %s : %v", i.filePath, err)
		resJpeg, err := imageutil.ThumbnailVips(buf, i.width, i.height, i.crop)
		require.Nilf(t, err, "err generating thumbnail: %v", err)
		_, format, err := image.Decode(bytes.NewReader(resJpeg))
		t.Logf("image format: %v", format)
		require.Nilf(t, err, "err decoding image: %v", err)
		require.Equal(t, "jpeg", format, "thumbnail should be in jpeg format")
	}
}
