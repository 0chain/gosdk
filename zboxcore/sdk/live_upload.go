package sdk

import (
	"io"
	"strconv"
	"time"
)

// LiveUpload live streaming video upload manager
type LiveUpload struct {
	allocationObj *Allocation

	// delay  delay to upload video
	delay time.Duration
	// clipsSize how much bytes in a video clips
	clipsSize int

	liveMeta     LiveMeta
	streamReader io.Reader

	// encryptOnUpload encrypt data on upload or not.
	encryptOnUpload bool
	// chunkSize how much bytes a chunk has. 64KB is default value.
	chunkSize int

	clipsIndex int

	// statusCallback trigger progress on StatusCallback
	statusCallback func() StatusCallback
}

// CreateLiveUpload create a LiveStreamUpload instance
func CreateLiveUpload(allocationObj *Allocation, liveMeta LiveMeta, streamReader io.Reader, opts ...LiveUploadOption) *LiveUpload {
	u := &LiveUpload{
		allocationObj: allocationObj,
		delay:         5 * time.Second,
		liveMeta:      liveMeta,
		streamReader:  streamReader,
		clipsIndex:    1,
	}

	for _, opt := range opts {
		opt(u)
	}

	return u
}

// Start start live streaming upload
func (lu *LiveUpload) Start() error {

	reader := createLiveUploadReader(lu.streamReader, lu.delay, lu.clipsSize)

	var err error
	var clipsUpload *StreamUpload
	for {

		clipsUpload = lu.createClipsUpload(lu.clipsIndex, reader)

		err = clipsUpload.Start()

		if err != nil {
			return err
		}

		lu.clipsIndex++

	}

}

func (lu *LiveUpload) createClipsUpload(clipsIndex int, reader io.Reader) *StreamUpload {
	fileMeta := FileMeta{

		MimeType:   lu.liveMeta.MimeType,
		RemoteName: lu.liveMeta.RemoteName + "." + strconv.Itoa(clipsIndex),
		RemotePath: lu.liveMeta.RemotePath + "." + strconv.Itoa(clipsIndex),
		Attributes: lu.liveMeta.Attributes,
	}

	return CreateStreamUpload(lu.allocationObj, fileMeta, reader,
		WithChunkSize(lu.chunkSize),
		WithEncrypt(lu.encryptOnUpload),
		WithStatusCallback(lu.statusCallback()))
}
