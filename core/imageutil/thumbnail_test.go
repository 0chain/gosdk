package imageutil_test

import (
	"encoding/json"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/0chain/gosdk/core/imageutil"
	"github.com/stretchr/testify/require"
)

var (
	inpJpeg = filepath.Join("resources", "input.jpeg")
	inpPng  = filepath.Join("resources", "input.png")
	inpGif  = filepath.Join("resources", "input.gif")
	inpWebp = filepath.Join("resources", "input.webp")
)

func TestThumbnail(t *testing.T) {

	type inp struct {
		filePath string
		width    int
		height   int
		options  imageutil.Option
	}

	inpData := []inp{
		{
			// jpeg file
			filePath: inpJpeg,
			width:    100,
			height:   200,
		},
		{
			// png file
			filePath: inpPng,
			width:    200,
			height:   300,
		},
		{
			// gif file
			filePath: inpGif,
			width:    100,
			height:   200,
		},
		{
			// webp file
			filePath: inpWebp,
			width:    200,
			height:   200,
		},
		{
			// With options
			filePath: inpWebp,
			width:    500,
			height:   500,
			options: imageutil.Option{
				IFormat: "webp",
			},
		},
	}

	for _, i := range inpData {
		t.Logf("image: %v", i.filePath)
		buf, err := os.ReadFile(i.filePath)
		require.Nilf(t, err, "err reading file %s : %v", i.filePath, err)
		options, err := json.Marshal(i.options)
		require.Nilf(t, err, "err marshal options %v: %v", i.options, err)
		res, err := imageutil.Thumbnail(buf, i.width, i.height, string(options))
		require.Nilf(t, err, "err generating thumbnail: %v", err)
		require.NotEmpty(t, res.ThumbnailImg, "resulting thumbnail shouldn't be empty")
	}
}
