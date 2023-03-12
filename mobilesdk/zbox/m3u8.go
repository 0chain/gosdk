package zbox

import (
	"bytes"
	"io"
	"sort"
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
	delay int

	Writer M3u8Writer

	Wait []string

	next chan string

	Seq int
}

// M3u8Writer m3u8 writer
type M3u8Writer interface {
	io.WriteSeeker
	Truncate(size int64) error
	Sync() error
}

// NewMediaPlaylist create media playlist(.m3u8)
func NewMediaPlaylist(delay int, writer M3u8Writer) *MediaPlaylist {
	m3u8 := &MediaPlaylist{
		delay:  delay,
		Wait:   make([]string, 0, 100),
		next:   make(chan string, 100),
		Seq:    0,
		Writer: writer,
	}

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

		found := sort.Search(len(m.Wait), func(i int) bool {
			return m.Wait[i] == item
		})
		if found != len(m.Wait) { // means found
			continue
		}

		m.Wait = append(m.Wait, item)
		m.flush()

		time.Sleep(1 * time.Second)
	}
}

// flush try flush new ts file into playlist
func (m *MediaPlaylist) flush() {
	if len(m.Wait) == 0 {
		return
	}

	m.Writer.Truncate(0)
	m.Writer.Seek(0, 0)
	m.Writer.Write(m.Encode())
	m.Writer.Sync()
}

// Encode encode m3u8
func (m *MediaPlaylist) Encode() []byte {
	var buf bytes.Buffer

	if len(m.Wait) == 0 {
		return buf.Bytes()
	}

	duration := strconv.Itoa(m.delay)
	name := m.Wait[0]
	sequience := strconv.Itoa(GetNumber(name))

	buf.WriteString("#EXTM3U\n")
	buf.WriteString("#EXT-X-VERSION:3\n")
	buf.WriteString("#EXT-X-TARGETDURATION:" + duration + "\n")
	buf.WriteString("#EXT-X-MEDIA-SEQUENCE:" + sequience + "\n")

	for _, it := range m.Wait {
		buf.WriteString("#EXTINF:" + duration + ",\n")
		buf.WriteString(it + "\n")
	}

	return buf.Bytes()
}

// String implement Stringer
func (m *MediaPlaylist) String() string {
	return string(m.Encode())
}
