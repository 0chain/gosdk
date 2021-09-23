package sdk

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
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

// MediaPlaylist m3u8 encoder and decoder
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

	m3u8.writer.Write([]byte("#EXTM3U\n"))
	m3u8.writer.Write([]byte("#EXT-X-VERSION:3\n"))
	m3u8.writer.Write([]byte("#EXT-X-TARGETDURATION:" + strconv.Itoa(delay) + "\n"))
	m3u8.writer.Sync()
	go m3u8.Play()

	return m3u8
}

// Append append new item
func (m *MediaPlaylist) Append(item string) {
	m.next <- item
}

// Play start to push item into playlist
func (m *MediaPlaylist) Play() {

	for {

		item := <-m.next

		_, err := os.Stat(filepath.Join(m.dir, item))

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

		time.Sleep(1 * time.Second)

	}

}

// flush try flush new ts file into playlist. return true if list is full
func (m *MediaPlaylist) flush() {

	m.writer.Truncate(0)
	m.writer.Seek(0, 0)
	m.writer.Write(m.Encode())
	m.writer.Sync()

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
