package sdk

import (
	"bytes"
	"io"
	"strconv"
	"sync"
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
	sync.RWMutex
	delay    int
	capacity int

	writer M3u8Writer

	// wait wait to play
	wait []string
	// items full list
	items  []string
	offset int
	since  *time.Time
}

// NewMediaPlaylist create media playlist(.m3u8)
func NewMediaPlaylist(delay, capacity int, writer M3u8Writer) *MediaPlaylist {
	m3u8 := &MediaPlaylist{
		delay:    delay,
		capacity: capacity,
		wait:     make([]string, 0, capacity),
		items:    make([]string, 0, 5*capacity),
		offset:   0,
		writer:   writer,
	}

	m3u8.writer.Write([]byte("#EXTM3U\n"))
	m3u8.writer.Write([]byte("#EXT-X-VERSION:3\n"))
	m3u8.writer.Write([]byte("#EXT-X-TARGETDURATION:" + strconv.Itoa(delay) + "\n"))
	//m3u8.writer.Sync()
	//go m3u8.Play()

	return m3u8
}

// Append append new item
func (m *MediaPlaylist) Append(item string) {
	// m.Lock()
	// defer m.Unlock()
	//m.items = append(m.items, item)
	m.writer.Write([]byte("#EXTINF:" + strconv.Itoa(m.delay) + ",\n"))
	m.writer.Write([]byte(item + "\n"))
	//	m.writer.Sync()
}

// Play start to push item into playlist
func (m *MediaPlaylist) Play() {

	//duration := strconv.Itoa(m.delay)
	//write header first
	// m.writer.Truncate(0)
	// m.writer.Seek(0, 0)

	for {
		if m.flush() {
			time.Sleep(time.Duration(m.delay) * time.Second)
		}
	}

}

// flush try flush new ts file into playlist. return true if list is full
func (m *MediaPlaylist) flush() bool {
	now := time.Now()

	if m.since == nil {
		m.since = &now
	}

	m.RLock()
	defer m.RUnlock()
	//if m.offset < len(m.items) {

	// 	next := m.items[m.offset]

	// 	if len(m.wait) < m.capacity {

	// 		m.wait = append(m.wait, next)
	// 		// m.writer.Truncate(0)
	// 		// m.writer.Seek(0, 0)
	// 		m.writer.Write(m.Encode())
	// 		// m.writer.Sync()
	// 		return false

	// 	}

	// 	// first item is completed to play
	// 	if now.Sub(*m.since).Seconds() > float64(m.delay) {

	// 		m.wait = append(m.wait[1:], next)
	// 		m.offset++
	// 		m.writer.Truncate(0)
	// 		m.writer.Seek(0, 0)
	// 		m.writer.Write(m.Encode())
	// 		m.writer.Sync()
	// 		return true
	// 	}

	// }

	return false
}

// Encode encode m3u8
func (m *MediaPlaylist) Encode() []byte {
	var buf bytes.Buffer

	duration := strconv.Itoa(m.delay)

	buf.WriteString("#EXTM3U\n")
	buf.WriteString("#EXT-X-VERSION:3\n")
	buf.WriteString("#EXT-X-TARGETDURATION:" + duration + "\n")
	//	buf.WriteString("#EXT-X-MEDIA-SEQUENCE:" + strconv.Itoa(m.offset) + "\n")

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
