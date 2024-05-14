package imageutil_test

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/0chain/gosdk/core/imageutil"
	"github.com/stretchr/testify/require"
)

var (
	avifPath = filepath.Join("test_data", "input.avif")
	bmpPath = filepath.Join("test_data", "input.bmp")
	// ddsPath = filepath.Join("test_data", "input.dds")
	exrPath = filepath.Join("test_data", "input.exr")
	gifPath = filepath.Join("test_data", "input.gif")
	// hdrPath = filepath.Join("test_data", "input.hdr")
	heicPath = filepath.Join("test_data", "input.heic")
	// heifPath = filepath.Join("test_data", "input.heif")
	icoPath = filepath.Join("test_data", "input.ico")
	jfifPath = filepath.Join("test_data", "input.jfif")
	jpePath = filepath.Join("test_data", "input.jpe")
	jpegPath = filepath.Join("test_data", "input.jpeg")
	jpgPath = filepath.Join("test_data", "input.jpg")
	jpsPath = filepath.Join("test_data", "input.jps")
	pngPath = filepath.Join("test_data", "input.png")
	pnmPath = filepath.Join("test_data", "input.pnm")
	svgPath = filepath.Join("test_data", "input.svg")
	// tgaPath = filepath.Join("test_data", "input.tga")
	tiffPath = filepath.Join("test_data", "input.tiff")
	webpPath = filepath.Join("test_data", "input.webp")
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
			// avif file
			filePath: avifPath, width: 100, height: 200,
		},
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
			// heic file
			filePath: heicPath, width: 200, height: 200,
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
			// jpe file
			filePath: jpePath, width: 200, height: 200,
		},
		{
			// jpe file
			filePath: jpePath, width: 200, height: 200,
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
			// svg file
			filePath: svgPath, width: 200, height: 200,
		},
		{
			// tiff file
			filePath: tiffPath, width: 200, height: 200,
		},
		{
			// webp file
			filePath: webpPath, width: 200, height: 200,
		},
		{
			// with options
			filePath: heicPath, width: 500, height: 500,
			options: imageutil.Option{
				IFormat: "heic",
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
		require.NotEmpty(t, res.Format, "resulting Format shouldn't be empty")
		switch res.Format {
			case "jpeg" :
				_, format, err := image.Decode(bytes.NewReader(res.ThumbnailImg))
				require.Nilf(t, err, "err decoding image: %v", err)
				require.Equal(t, "jpeg", format, "format mismatch; result format: jpeg, image format: %v", format)
			default:
				t.Errorf("unknown format: %v", res.Format)	
		}
	}
}

func TestImageRsConvert(t *testing.T) {

	type inp struct {
		filePath string
		width    int
		height   int
		options  imageutil.ConvertOptions
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

	imageRs, err := imageutil.NewImageRs()
	require.Nilf(t, err, "err instantiating image-rs")

	for _, i := range inpData {
		t.Logf("image: %v", i.filePath)
		buf, err := os.ReadFile(i.filePath)
		require.Nilf(t, err, "err reading file %s : %v", i.filePath, err)
		res, err := imageRs.Convert(buf, i.width, i.height, i.options)
		require.Nilf(t, err, "err generating thumbnail: %v", err)
		require.NotEmpty(t, res.ThumbnailImg, "resulting thumbnail shouldn't be empty")
		require.Equal(t, "jpeg", res.Format, "resulting Format should be in jpeg")
		_, format, err := image.Decode(bytes.NewReader(res.ThumbnailImg))
		require.Nilf(t, err, "err decoding image: %v", err)
		require.Equal(t, "jpeg", format, "format mismatch; result format: jpeg, image format: %v", format)
	}
}

func TestGoNativeDecodeConvert(t *testing.T) {

	type inp struct {
		filePath string
		width    int
		height   int
		options  imageutil.ConvertOptions
	}

	inpData := []inp {
		{
			// avif file
			filePath: avifPath, width: 100, height: 200,
		},
		{
			// bmp file
			filePath: bmpPath, width: 200, height: 300,
		},
		{
			// gif file
			filePath: gifPath, width: 200, height: 200,
		},
		{
			// heic file
			filePath: heicPath, width: 200, height: 200,
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
			// jpe file
			filePath: jpePath, width: 200, height: 200,
		},
		{
			// jpe file
			filePath: jpePath, width: 200, height: 200,
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
			// svg file
			filePath: svgPath, width: 200, height: 200,
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

	gonative, err := imageutil.NewGoNativeDecode()
	require.Nilf(t, err, "err instantiating go-native-decode")

	for _, i := range inpData {
		t.Logf("image: %v", i.filePath)
		buf, err := os.ReadFile(i.filePath)
		require.Nilf(t, err, "err reading file %s : %v", i.filePath, err)
		res, err := gonative.Convert(buf, i.width, i.height, i.options)
		require.Nilf(t, err, "err generating thumbnail: %v", err)
		require.NotEmpty(t, res.ThumbnailImg, "resulting thumbnail shouldn't be empty")
		require.Equal(t, "jpeg", res.Format, "resulting Format should be in jpeg")
		_, format, err := image.Decode(bytes.NewReader(res.ThumbnailImg))
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
		options  imageutil.Option
	}

	inpData := []inp{
		// avif
		{
			name: "sample.avif", filePath: filepath.Join("benchmark_data", "sample.avif"),
			width: 120, height: 90,
		},
		// bmp
		{
			name: "sample_640*426.bmp", filePath: filepath.Join("benchmark_data", "sample_640*426.bmp"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.bmp", filePath: filepath.Join("benchmark_data", "sample_1280*853.bmp"),
			width: 120, height: 90,
		},
		// exr
		{
			name: "sample_640*426.exr", filePath: filepath.Join("benchmark_data", "sample_640*426.exr"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.exr", filePath: filepath.Join("benchmark_data", "sample_1280*853.exr"),
			width: 120, height: 90,
		},
		// gif
		{
			name: "sample_640*426.gif", filePath: filepath.Join("benchmark_data", "sample_640*426.gif"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.gif", filePath: filepath.Join("benchmark_data", "sample_1280*853.gif"),
			width: 120, height: 90,
		},
		// heic
		{
			name: "sample.heic", filePath: filepath.Join("benchmark_data", "sample.heic"),
			width: 120, height: 90,
		},
		// ico
		{
			name: "sample_640*426.ico", filePath: filepath.Join("benchmark_data", "sample_640*426.ico"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.ico", filePath: filepath.Join("benchmark_data", "sample_1280*853.ico"),
			width: 120, height: 90,
		},
		// jfif
		{
			name: "sample.jfif", filePath: filepath.Join("benchmark_data", "sample.jfif"),
			width: 120, height: 90,
		},
		// jpe
		{
			name: "sample_640*426.jpe", filePath: filepath.Join("benchmark_data", "sample_640*426.jpe"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.jpe", filePath: filepath.Join("benchmark_data", "sample_1280*853.jpe"),
			width: 120, height: 90,
		},
		// jpeg
		{
			name: "sample_640*426.jpeg", filePath: filepath.Join("benchmark_data", "sample_640*426.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.jpeg", filePath: filepath.Join("benchmark_data", "sample_1280*853.jpeg"),
			width: 120, height: 90,
		},
		// jpg
		{
			name: "sample_640*426.jpg", filePath: filepath.Join("benchmark_data", "sample_640*426.jpg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.jpg", filePath: filepath.Join("benchmark_data", "sample_1280*853.jpg"),
			width: 120, height: 90,
		},
		// jps
		{
			name: "sample_640*426.jps", filePath: filepath.Join("benchmark_data", "sample_640*426.jps"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.jps", filePath: filepath.Join("benchmark_data", "sample_1280*853.jps"),
			width: 120, height: 90,
		},
		// png
		{
			name: "sample_640*426.png", filePath: filepath.Join("benchmark_data", "sample_640*426.png"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.png", filePath: filepath.Join("benchmark_data", "sample_1280*853.png"),
			width: 120, height: 90,
		},
		// pnm
		{
			name: "sample_640*426.pnm", filePath: filepath.Join("benchmark_data", "sample_640*426.pnm"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.pnm", filePath: filepath.Join("benchmark_data", "sample_1280*853.pnm"),
			width: 120, height: 90,
		},
		// svg
		{
			name: "sample.svg", filePath: filepath.Join("benchmark_data", "sample.svg"),
			width: 120, height: 90,
		},
		// tiff
		{
			name: "sample_640*426.tiff", filePath: filepath.Join("benchmark_data", "sample_640*426.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.tiff", filePath: filepath.Join("benchmark_data", "sample_1280*853.tiff"),
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
		options, err := json.Marshal(iData.options)
		require.Nilf(b, err, "err marshal options %v: %v", iData.options, err)
		var res imageutil.ConvertRes
		b.Run(iData.name, func(b *testing.B) {
			for i:=0; i<b.N; i++ {
				res, err = imageutil.Thumbnail(buf, iData.width, iData.height, string(options))
				require.Nilf(b, err, "convert failed with err : %v", err)
			}
		})
		b.Logf("file size: %v in KB\n", float32(len(buf)) / float32(1024))
		b.Logf("thumbnail size: %v in KB\n", float32(len(res.ThumbnailImg)) / float32(1024))
		b.Logf("converter: %v\n", res.Converter)
	}
}

func BenchmarkImageRsConvert(b *testing.B) {
	type inp struct {
		name 	string
		filePath string
		width    int
		height   int
	}

	inpData := []inp{
		// bmp
		{
			name: "sample_640*426.bmp", filePath: filepath.Join("benchmark_data", "sample_640*426.bmp"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.bmp", filePath: filepath.Join("benchmark_data", "sample_1280*853.bmp"),
			width: 120, height: 90,
		},
		// exr
		{
			name: "sample_640*426.exr", filePath: filepath.Join("benchmark_data", "sample_640*426.exr"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.exr", filePath: filepath.Join("benchmark_data", "sample_1280*853.exr"),
			width: 120, height: 90,
		},
		// gif
		{
			name: "sample_640*426.gif", filePath: filepath.Join("benchmark_data", "sample_640*426.gif"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.gif", filePath: filepath.Join("benchmark_data", "sample_1280*853.gif"),
			width: 120, height: 90,
		},
		// ico
		{
			name: "sample_640*426.ico", filePath: filepath.Join("benchmark_data", "sample_640*426.ico"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.ico", filePath: filepath.Join("benchmark_data", "sample_1280*853.ico"),
			width: 120, height: 90,
		},
		// jfif
		{
			name: "sample.jfif", filePath: filepath.Join("benchmark_data", "sample.jfif"),
			width: 120, height: 90,
		},
		// jpe
		{
			name: "sample_640*426.jpe", filePath: filepath.Join("benchmark_data", "sample_640*426.jpe"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.jpe", filePath: filepath.Join("benchmark_data", "sample_1280*853.jpe"),
			width: 120, height: 90,
		},
		// jpeg
		{
			name: "sample_640*426.jpeg", filePath: filepath.Join("benchmark_data", "sample_640*426.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.jpeg", filePath: filepath.Join("benchmark_data", "sample_1280*853.jpeg"),
			width: 120, height: 90,
		},
		// jpg
		{
			name: "sample_640*426.jpg", filePath: filepath.Join("benchmark_data", "sample_640*426.jpg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.jpg", filePath: filepath.Join("benchmark_data", "sample_1280*853.jpg"),
			width: 120, height: 90,
		},
		// jps
		{
			name: "sample_640*426.jps", filePath: filepath.Join("benchmark_data", "sample_640*426.jps"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.jps", filePath: filepath.Join("benchmark_data", "sample_1280*853.jps"),
			width: 120, height: 90,
		},
		// png
		{
			name: "sample_640*426.png", filePath: filepath.Join("benchmark_data", "sample_640*426.png"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.png", filePath: filepath.Join("benchmark_data", "sample_1280*853.png"),
			width: 120, height: 90,
		},
		// pnm
		{
			name: "sample_640*426.pnm", filePath: filepath.Join("benchmark_data", "sample_640*426.pnm"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.pnm", filePath: filepath.Join("benchmark_data", "sample_1280*853.pnm"),
			width: 120, height: 90,
		},
		// tiff
		{
			name: "sample_640*426.tiff", filePath: filepath.Join("benchmark_data", "sample_640*426.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.tiff", filePath: filepath.Join("benchmark_data", "sample_1280*853.tiff"),
			width: 120, height: 90,
		},
		// webp
		{
			name: "sample.webp", filePath: filepath.Join("benchmark_data", "sample.webp"),
			width: 120, height: 90,
		},
	}

	imageRs, err := imageutil.NewImageRs()
	require.Nilf(b, err, "err instantiating image-rs")
	for _, iData := range inpData {
		buf, err := os.ReadFile(iData.filePath)
		require.Nilf(b, err, "err reading file %s : %v", iData.filePath, err)
		b.Run(iData.name, func(b *testing.B) {
			for i:=0; i<b.N; i++ {
				_, err = imageRs.Convert(buf, iData.width, iData.height, imageutil.ConvertOptions{})
				require.Nilf(b, err, "convert failed with err : %v", err)
			}
		})
	}
}