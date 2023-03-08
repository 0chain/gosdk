package sdk

type thumbnailDownloader struct {
	baseDownloader
}

func (d *thumbnailDownloader) Start(status StatusCallback) error {
	if d.isViewer {
		return d.allocationObj.DownloadThumbnailFromAuthTicket(d.localPath,
			d.authTicket, d.lookupHash, d.fileName, status)

	}
	return d.allocationObj.DownloadThumbnail(d.localPath, d.remotePath, status)
}
