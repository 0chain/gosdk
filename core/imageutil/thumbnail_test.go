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
	bmpPath = filepath.Join("test_data", "input.bmp")
	// ddsPath = filepath.Join("test_data", "input.dds")
	exrPath = filepath.Join("test_data", "input.exr")
	gifPath = filepath.Join("test_data", "input.gif")
	// hdrPath = filepath.Join("test_data", "input.hdr")
	icoPath = filepath.Join("test_data", "input.ico")
	jfifPath = filepath.Join("test_data", "input.jfif")
	jpePath = filepath.Join("test_data", "input.jpe")
	jpegPath = filepath.Join("test_data", "input.jpeg")
	jpgPath = filepath.Join("test_data", "input.jpg")
	jpsPath = filepath.Join("test_data", "input.jps")
	pngPath = filepath.Join("test_data", "input.png")
	pnmPath = filepath.Join("test_data", "input.pnm")
	// tgaPath = filepath.Join("test_data", "input.tga")
	tiffPath = filepath.Join("test_data", "input.tiff")
	webpPath = filepath.Join("test_data", "input.webp")
)

func TestThumbnail(t *testing.T) {

	type inp struct {
		filePath string
		width    int
		height   int
	}

	inpData := []inp {
		{
			// bmp file
			filePath: bmpPath, width: 200, height: 300,
		},
		{
			// exr file
			filePath: exrPath, width: 100, height: 200,
		},
		{
			// gif file
			filePath: gifPath, width: 200, height: 200,
		},
		{
			// ico file
			filePath: icoPath, width: 200, height: 200,
		},
		{
			// jfif file
			filePath: jfifPath, width: 200, height: 200,
		},
		{
			// jpe file
			filePath: jpePath, width: 200, height: 200,
		},
		{
			// jpeg file
			filePath: jpegPath, width: 200, height: 200,
		},
		{
			// jpg file
			filePath: jpgPath, width: 200, height: 200,
		},
		{
			// jps file
			filePath: jpsPath, width: 200, height: 200,
		},
		{
			// png file
			filePath: pngPath, width: 200, height: 200,
		},
		{
			// pnm file
			filePath: pnmPath, width: 200, height: 200,
		},
		{
			// tiff file
			filePath: tiffPath, width: 200, height: 200,
		},
		{
			// webp file
			filePath: webpPath, width: 200, height: 200,
		},
	}

	for _, i := range inpData {
		t.Logf("image: %v", i.filePath)
		buf, err := os.ReadFile(i.filePath)
		require.Nilf(t, err, "err reading file %s : %v", i.filePath, err)
		res, err := imageutil.Thumbnail(buf, i.width, i.height)
		require.Nilf(t, err, "err generating thumbnail: %v", err)
		require.NotEmpty(t, res, "resulting thumbnail shouldn't be empty")
		_, format, err := image.Decode(bytes.NewReader(res))
		require.Nilf(t, err, "err decoding image: %v", err)
		require.Equal(t, "jpeg", format, "format mismatch; result format: jpeg, image format: %v", format)
	}
}

func BenchmarkThumbnail(b *testing.B) {
	type inp struct {
		name 	string
		filePath string
		width    int
		height   int
	}

	inpData := []inp{
		// bmp
		{
			name: "sample_640x426.bmp", filePath: filepath.Join("benchmark_data", "sample_640x426.bmp"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280x853.bmp", filePath: filepath.Join("benchmark_data", "sample_1280x853.bmp"),
			width: 120, height: 90,
		},
		// exr
		{
			name: "sample_640x426.exr", filePath: filepath.Join("benchmark_data", "sample_640x426.exr"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280x853.exr", filePath: filepath.Join("benchmark_data", "sample_1280x853.exr"),
			width: 120, height: 90,
		},
		// gif
		{
			name: "sample_640x426.gif", filePath: filepath.Join("benchmark_data", "sample_640x426.gif"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280x853.gif", filePath: filepath.Join("benchmark_data", "sample_1280x853.gif"),
			width: 120, height: 90,
		},
		// ico
		{
			name: "sample_640x426.ico", filePath: filepath.Join("benchmark_data", "sample_640x426.ico"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280x853.ico", filePath: filepath.Join("benchmark_data", "sample_1280x853.ico"),
			width: 120, height: 90,
		},
		// jfif
		{
			name: "sample.jfif", filePath: filepath.Join("benchmark_data", "sample.jfif"),
			width: 120, height: 90,
		},
		// jpe
		{
			name: "sample_640x426.jpe", filePath: filepath.Join("benchmark_data", "sample_640x426.jpe"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280x853.jpe", filePath: filepath.Join("benchmark_data", "sample_1280x853.jpe"),
			width: 120, height: 90,
		},
		// jpeg
		{
			name: "sample_640x426.jpeg", filePath: filepath.Join("benchmark_data", "sample_640x426.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280x853.jpeg", filePath: filepath.Join("benchmark_data", "sample_1280x853.jpeg"),
			width: 120, height: 90,
		},
		// jpg
		{
			name: "sample_640x426.jpg", filePath: filepath.Join("benchmark_data", "sample_640x426.jpg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280x853.jpg", filePath: filepath.Join("benchmark_data", "sample_1280x853.jpg"),
			width: 120, height: 90,
		},
		// jps
		{
			name: "sample_640x426.jps", filePath: filepath.Join("benchmark_data", "sample_640x426.jps"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280x853.jps", filePath: filepath.Join("benchmark_data", "sample_1280x853.jps"),
			width: 120, height: 90,
		},
		// png
		{
			name: "sample_640x426.png", filePath: filepath.Join("benchmark_data", "sample_640x426.png"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280x853.png", filePath: filepath.Join("benchmark_data", "sample_1280x853.png"),
			width: 120, height: 90,
		},
		// pnm
		{
			name: "sample_640x426.pnm", filePath: filepath.Join("benchmark_data", "sample_640x426.pnm"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280x853.pnm", filePath: filepath.Join("benchmark_data", "sample_1280x853.pnm"),
			width: 120, height: 90,
		},
		// tiff
		{
			name: "sample_640x426.tiff", filePath: filepath.Join("benchmark_data", "sample_640x426.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280x853.tiff", filePath: filepath.Join("benchmark_data", "sample_1280x853.tiff"),
			width: 120, height: 90,
		},
		// webp
		{
			name: "sample.webp", filePath: filepath.Join("benchmark_data", "sample.webp"),
			width: 120, height: 90,
		},
	}

	for _, iData := range inpData {
		buf, err := os.ReadFile(iData.filePath)
		require.Nilf(b, err, "err reading file %s : %v", iData.filePath, err)
		var res []byte
		b.Run(iData.name, func(b *testing.B) {
			for i:=0; i<b.N; i++ {
				res, err = imageutil.Thumbnail(buf, iData.width, iData.height)
				require.Nilf(b, err, "convert failed with err : %v", err)
			}
		})
		b.Logf("file size: %v in KB\n", float32(len(buf)) / float32(1024))
		b.Logf("thumbnail size: %v in KB\n", float32(len(res)) / float32(1024))
	}
}
