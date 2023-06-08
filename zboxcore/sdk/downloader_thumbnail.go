package sdk

type thumbnailDownloader struct {
	baseDownloader
}

func (d *thumbnailDownloader) Start(status StatusCallback, isFinal bool) error {
	if d.isViewer {
		return d.allocationObj.DownloadThumbnailFromAuthTicket(d.localPath,
			d.authTicket, d.lookupHash, d.fileName, d.verifyDownload, status)

	}
	return d.allocationObj.DownloadThumbnail(d.localPath, d.remotePath, d.verifyDownload, status, isFinal)
}
