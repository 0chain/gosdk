package sdk

// LiveUpload live streaming video upload manager
type LiveUpload struct {
	allocationObj *Allocation
	homedir       string

	// delay  delay to upload video
	delay int

	liveMeta   LiveMeta
	liveReader LiveUploadReader

	// encryptOnUpload encrypt data on upload or not.
	encryptOnUpload bool
	// chunkSize how much bytes a chunk has. 64KB is default value.
	chunkSize int64

	clipsIndex int

	// statusCallback trigger progress on StatusCallback
	statusCallback func() StatusCallback
}

// CreateLiveUpload create a LiveChunkedUpload instance
func CreateLiveUpload(homedir string, allocationObj *Allocation, liveMeta LiveMeta, liveReader LiveUploadReader, opts ...LiveUploadOption) *LiveUpload {
	u := &LiveUpload{
		allocationObj: allocationObj,
		homedir:       homedir,
		//delay:         5 * time.Second,
		//clipsSize:    1024 * 1024 * 20, //50M
		liveMeta:   liveMeta,
		liveReader: liveReader,
		clipsIndex: 1,
	}

	for _, opt := range opts {
		opt(u)
	}

	return u
}

// Start start live streaming upload
func (lu *LiveUpload) Start() error {

	var err error
	var clipsUpload *ChunkedUpload
	for {

		clipsUpload, err = lu.createClipsUpload(lu.clipsIndex, lu.liveReader)
		if err != nil {
			return err
		}

		err = clipsUpload.Start()

		if err != nil {
			return err
		}

		lu.clipsIndex++

	}

}

func (lu *LiveUpload) createClipsUpload(clipsIndex int, reader LiveUploadReader) (*ChunkedUpload, error) {
	fileMeta := FileMeta{
		Path:       reader.GetClipsFile(clipsIndex),
		ActualSize: reader.Size(),

		MimeType:   lu.liveMeta.MimeType,
		RemoteName: reader.GetClipsFileName(clipsIndex),
		RemotePath: lu.liveMeta.RemotePath + "/" + reader.GetClipsFileName(clipsIndex),
		Attributes: lu.liveMeta.Attributes,
	}

	return CreateChunkedUpload(lu.homedir, lu.allocationObj, fileMeta, reader, false,
		WithChunkSize(lu.chunkSize),
		WithEncrypt(lu.encryptOnUpload),
		WithStatusCallback(lu.statusCallback()))
}
