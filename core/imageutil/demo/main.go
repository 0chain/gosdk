package main

import (
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/0chain/gosdk/core/imageutil"
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
}

var (
	permissions = 0644
	images      = []img{
		{
			location:     filepath.Join("..", "resources", "input.gif"),
			locationType: "local",
			width:        500,
			height:       500,
		},
		{
			location:     filepath.Join("..", "resources", "input.png"),
			locationType: "local",
			width:        500,
			height:       500,
		},
		{
			location:     filepath.Join("..", "resources", "input.jpeg"),
			locationType: "local",
			width:        500,
			height:       500,
		},
		{
			location:     filepath.Join("..", "resources", "input.webp"),
			locationType: "local",
			width:        500,
			height:       500,
		},
		{
			location:     filepath.Join("..", "resources", "input.heic"),
			locationType: "local",
			width:        500,
			height:       500,
		},
		{
			location:     filepath.Join("..", "resources", "input.avif"),
			locationType: "local",
			width:        500,
			height:       500,
		},
	}
)

func main() {

	folderPath := "images"
	if _, err := os.Stat(folderPath); err == nil {
		// folderPath exists
		err = os.RemoveAll(folderPath)
		if err != nil {
			log.Panicf("err deleting images folder: %v", err)
		}
	}
	err := os.MkdirAll(folderPath, fs.FileMode(permissions))
	if err != nil {
		log.Panicf("err creating images folder: %v", err)
	}

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
		var res imageutil.ConvertRes
		res, err := imageutil.Thumbnail(b, input.width, input.height, "{}")
		if err != nil {
			log.Printf("err creating thumbnail: %v", err)
			continue
		}

		log.Printf("res_arr size: %v, first 30 bytes: %v\n", len(res.ThumbnailImg), res.ThumbnailImg[0:30])

		log.Printf("creating outfile file")
		opath := filepath.Join(folderPath, fmt.Sprintf("output%d.jpeg", i))
		err = os.WriteFile(opath, res.ThumbnailImg, fs.FileMode(permissions))
		if err != nil {
			log.Printf("err creating outfile file: %v", err)
		}

	}
	log.Println("Done!")
}
