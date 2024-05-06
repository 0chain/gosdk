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
	avifPath = filepath.Join("resources", "input.avif")
	bmpPath = filepath.Join("resources", "input.bmp")
	// ddsPath = filepath.Join("resources", "input.dds")
	exrPath = filepath.Join("resources", "input.exr")
	gifPath = filepath.Join("resources", "input.gif")
	// hdrPath = filepath.Join("resources", "input.hdr")
	heicPath = filepath.Join("resources", "input.heic")
	heifPath = filepath.Join("resources", "input.heif")
	icoPath = filepath.Join("resources", "input.ico")
	jfifPath = filepath.Join("resources", "input.jfif")
	jpePath = filepath.Join("resources", "input.jpe")
	jpegPath = filepath.Join("resources", "input.jpeg")
	jpgPath = filepath.Join("resources", "input.jpg")
	jpsPath = filepath.Join("resources", "input.jps")
	pngPath = filepath.Join("resources", "input.png")
	pnmPath = filepath.Join("resources", "input.pnm")
	// tgaPath = filepath.Join("resources", "input.tga")
	tiffPath = filepath.Join("resources", "input.tiff")
	webpPath = filepath.Join("resources", "input.webp")
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
			// heif file
			filePath: heifPath, width: 200, height: 200,
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
			// heif file
			filePath: heifPath, width: 200, height: 200,
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
		{
			name: "sample_640*426.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_640*426.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_1280*853.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1920*1280.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_1920*1280.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_5184*3456.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_5184*3456.jpeg"),
			width: 120, height: 90,
		},

		{
			name: "sample_640*426.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_640*426.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_1280*853.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_1920*1280.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_1920*1280.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_5184*3456.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_5184*3456.tiff"),
			width: 120, height: 90,
		},
	}

	for _, iData := range inpData {
		b.Run(iData.name, func(b *testing.B) {
			buf, err := os.ReadFile(iData.filePath)
			require.Nilf(b, err, "err reading file %s : %v", iData.filePath, err)
			options, err := json.Marshal(iData.options)
			require.Nilf(b, err, "err marshal options %v: %v", iData.options, err)
			b.ResetTimer()
			for i:=0; i<b.N; i++ {
				_, err = imageutil.Thumbnail(buf, iData.width, iData.height, string(options))
				require.Nilf(b, err, "convert failed with err : %v", err)
			}
		})
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
		{
			name: "sample_640*426.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_640*426.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_1280*853.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1920*1280.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_1920*1280.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_5184*3456.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_5184*3456.jpeg"),
			width: 120, height: 90,
		},

		{
			name: "sample_640*426.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_640*426.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_1280*853.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_1920*1280.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_1920*1280.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_5184*3456.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_5184*3456.tiff"),
			width: 120, height: 90,
		},
	}

	imageRs, err := imageutil.NewImageRs()
	require.Nilf(b, err, "err instantiating image-rs")
	for _, iData := range inpData {
		b.Run(iData.name, func(b *testing.B) {
			buf, err := os.ReadFile(iData.filePath)
			require.Nilf(b, err, "err reading file %s : %v", iData.filePath, err)
			b.ResetTimer()
			for i:=0; i<b.N; i++ {
				_, err = imageRs.Convert(buf, iData.width, iData.height, imageutil.ConvertOptions{})
				require.Nilf(b, err, "convert failed with err : %v", err)
			}
		})
	}
}

func BenchmarkGoNativeDecodeConvert(b *testing.B) {
	type inp struct {
		name 	string
		filePath string
		width    int
		height   int
	}

	inpData := []inp{
		{
			name: "sample_640*426.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_640*426.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_1280*853.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_1920*1280.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_1920*1280.jpeg"),
			width: 120, height: 90,
		},
		{
			name: "sample_5184*3456.jpeg", filePath: filepath.Join("resources_large", "jpeg", "sample_5184*3456.jpeg"),
			width: 120, height: 90,
		},

		{
			name: "sample_640*426.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_640*426.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_1280*853.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_1280*853.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_1920*1280.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_1920*1280.tiff"),
			width: 120, height: 90,
		},
		{
			name: "sample_5184*3456.tiff", filePath: filepath.Join("resources_large", "tiff", "sample_5184*3456.tiff"),
			width: 120, height: 90,
		},
	}

	gonative, err := imageutil.NewGoNativeDecode()
	require.Nilf(b, err, "err instantiating go-native-decode")
	for _, iData := range inpData {
		b.Run(iData.name, func(b *testing.B) {
			buf, err := os.ReadFile(iData.filePath)
			require.Nilf(b, err, "err reading file %s : %v", iData.filePath, err)
			b.ResetTimer()
			for i:=0; i<b.N; i++ {
				_, err = gonative.Convert(buf, iData.width, iData.height, imageutil.ConvertOptions{})
				require.Nilf(b, err, "convert failed with err : %v", err)
			}
		})
	}
}