package sdk

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/0chain/gosdk/core/sys"
)

// #EXTM3U
// #EXT-X-VERSION:3
// #EXT-X-TARGETDURATION:5
// #EXT-X-MEDIA-SEQUENCE:17
// #EXTINF:5.000000,
// tv17.ts
// #EXTINF:5.000000,
// tv18.ts
// #EXTINF:5.000000,
// tv19.ts
// #EXTINF:5.000000,
// tv20.ts
// #EXTINF:5.000000,
// tv21.ts

// MediaPlaylist queue-based m3u8 playlist
type MediaPlaylist struct {
	dir   string
	delay int

	writer M3u8Writer

	// wait wait to play
	wait []string

	next chan string
	seq  int
}

// NewMediaPlaylist create media playlist(.m3u8)
func NewMediaPlaylist(delay int, dir string, writer M3u8Writer) *MediaPlaylist {
	m3u8 := &MediaPlaylist{
		dir:    dir,
		delay:  delay,
		wait:   make([]string, 0, 5),
		next:   make(chan string, 100),
		seq:    0,
		writer: writer,
	}

	m3u8.writer.Write([]byte("#EXTM3U\n"))                                           //nolint
	m3u8.writer.Write([]byte("#EXT-X-VERSION:3\n"))                                  //nolint
	m3u8.writer.Write([]byte("#EXT-X-TARGETDURATION:" + strconv.Itoa(delay) + "\n")) //nolint
	m3u8.writer.Sync()                                                               //nolint
	go m3u8.Play()

	return m3u8
}

// Append append new item into playlist
func (m *MediaPlaylist) Append(item string) {
	m.next <- item
}

// Play start to play the contents of the playlist with 1 second buffer between each item
func (m *MediaPlaylist) Play() {

	for {

		item := <-m.next

		_, err := sys.Files.Stat(filepath.Join(m.dir, item))

		if err == nil {
			if len(m.wait) < 5 {
				m.seq = 1
				m.wait = append(m.wait, "."+string(os.PathSeparator)+item)
			} else {
				m.seq++
				m.wait = append(m.wait[1:], "."+string(os.PathSeparator)+item)
			}

			m.flush()
		}

		sys.Sleep(1 * time.Second)

	}

}

// flush try flush new ts file into playlist. return true if list is full
func (m *MediaPlaylist) flush() {

	m.writer.Truncate(0)       //nolint
	m.writer.Seek(0, 0)        //nolint
	m.writer.Write(m.Encode()) //nolint
	m.writer.Sync()            //nolint

}

// Encode encode m3u8
func (m *MediaPlaylist) Encode() []byte {
	var buf bytes.Buffer

	duration := strconv.Itoa(m.delay)

	buf.WriteString("#EXTM3U\n")
	buf.WriteString("#EXT-X-VERSION:3\n")
	buf.WriteString("#EXT-X-TARGETDURATION:" + duration + "\n")
	buf.WriteString("#EXT-X-MEDIA-SEQUENCE:" + strconv.Itoa(m.seq) + "\n")

	for _, it := range m.wait {
		buf.WriteString("#EXTINF:" + duration + ",\n")
		buf.WriteString(it + "\n")
	}

	return buf.Bytes()
}

// String implement Stringer
func (m *MediaPlaylist) String() string {
	return string(m.Encode())
}

// M3u8Writer m3u8 writer
type M3u8Writer interface {
	io.WriteSeeker
	Truncate(size int64) error
	Sync() error
}
