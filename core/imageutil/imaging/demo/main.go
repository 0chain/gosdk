package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/ccitt"
	_ "golang.org/x/image/riff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vector"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"
	_ "github.com/gen2brain/heic"

	"github.com/0chain/gosdk/core/imageutil/imaging"
)

type img struct {
	// location to download image
	location string
	// location type can be either remote or local
	locationType string
	// expected width of thumbnail
	width int
	// expected height of thumbnail
	height int
	// ResampleFilter
	resample imaging.ResampleFilter
}

var (
	permissions = 0644
	images      = []img{
		{
			location:     "https://go.dev/blog/go-brand/Go-Logo/JPG/Go-Logo_Aqua.jpg",
			locationType: "remote",
			width:        100,
			height:       100,
		},
		{
			location:     "https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Yellow.png",
			locationType: "remote",
			width:        200,
			height:       200,
		},
		{
			location:     filepath.Join("..", "..", "resources", "input.webp"),
			locationType: "local",
			width:        100,
			height:       200,
		},
		{
			location:     filepath.Join("..", "..", "resources", "input.heic"),
			locationType: "local",
			width:        200,
			height:       200,
		},
	}
)

func main() {

	for i, input := range images {
		log.Printf("Location: %s", input.location)
		var b []byte
		switch input.locationType {
		case "remote":
			log.Printf("downloading image...")
			resp, err := http.Get(input.location)
			if err != nil {
				log.Printf("err downloading url: %v", err)
				continue
			}
			defer resp.Body.Close()
			b, err = io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("err reading response body: %v", err)
				continue
			}
		case "local":
			log.Printf("reading file...")
			bytes, err := os.ReadFile(input.location)
			if err != nil {
				log.Printf("err reading file: %v", err)
				continue
			}
			b = bytes
		default:
			log.Printf("unsupported location type: %v", input.locationType)
		}

		log.Printf("creating Thumbnail...")
		var resJpeg []byte
		resJpeg, err := imaging.Thumbnail(b, input.width, input.height, input.resample)
		if err != nil {
			log.Printf("err creating thumbnail: %v", err)
			continue
		}

		log.Printf("decoding image format...")
		_, format, err := image.Decode(bytes.NewReader(b))
		if err != nil {
			log.Printf("err decoding response body: %v", err)
			continue
		}

		folderPath := fmt.Sprintf("images/%d/", i)
		if _, err := os.Stat(folderPath); err == nil {
			// folderPath exists
			err = os.RemoveAll(folderPath)
			if err != nil {
				log.Printf("err deleting images folder: %v", err)
				continue
			}
		}
		err = os.MkdirAll(folderPath, fs.FileMode(permissions))
		if err != nil {
			log.Printf("err creating images folder: %v", err)
			continue
		}

		log.Printf("creating input file...")
		ipath := fmt.Sprintf("images/%d/input.%s", i, format)
		ifile, err := os.Create(ipath)
		if err != nil {
			log.Printf("err creating input file: %v", err)
			continue
		}
		defer ifile.Close()
		_, err = io.Copy(ifile, bytes.NewReader(b))
		if err != nil {
			log.Printf("err copying input file: %v", err)
			continue
		}

		log.Printf("creating outfile file")
		opath := fmt.Sprintf("images/%d/output.jpeg", i)
		err = os.WriteFile(opath, resJpeg, fs.FileMode(permissions))
		if err != nil {
			log.Printf("err creating outfile file: %v", err)
		}
	}
	log.Println("completed")
}
