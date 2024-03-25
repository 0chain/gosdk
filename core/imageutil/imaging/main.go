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

	"github.com/0chain/gosdk/core/imageutil"
)

type img struct {
	// url to download image
	url string
	// expected width of thumbnail
	width int
	// expected height of thumbnail
	height int
	// ResampleFilter
	resample imageutil.ResampleFilter
}

var (
	permissions = 0644
	images      = []img{
		{
			url:    "https://go.dev/blog/go-brand/Go-Logo/JPG/Go-Logo_Aqua.jpg",
			width:  100,
			height: 100,
		},
		{
			url:    "https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Yellow.png",
			width:  200,
			height: 200,
		},
		{
			url:    "https://go.dev/blog/go-brand/Go-Logo/SVG/Go-Logo_Blue.svg",
			width:  100,
			height: 200,
		},
	}
)

func main() {

	for i, input := range images {
		log.Printf("URL: %s", input.url)
		log.Printf("downloading image...")
		resp, err := http.Get(input.url)
		if err != nil {
			log.Printf("err downloading url: %v", err)
			continue
		}
		defer resp.Body.Close()

		log.Printf("creating Thumbnail...")
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("err reading response body: %v", err)
			continue
		}
		var resPng []byte
		resPng, err = imageutil.ThumbnailImaging(b, input.width, input.height, input.resample)
		if err != nil {
			log.Printf("err creating thumbnail: %v", err)
			continue
		}

		log.Printf("decoding image format...")
		reader := bytes.NewReader(b)
		_, format, err := image.Decode(reader)
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
		reader = bytes.NewReader(b)
		_, err = io.Copy(ifile, reader)
		if err != nil {
			log.Printf("err copying input file: %v", err)
			continue
		}

		log.Printf("creating outfile file")
		opath := fmt.Sprintf("images/%d/output.png", i)
		err = os.WriteFile(opath, resPng, fs.FileMode(permissions))
		if err != nil {
			log.Printf("err creating outfile file: %v", err)
		}
	}
	log.Println("completed")
}
