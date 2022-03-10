package sdk

type thumbnailDownloader struct {
	baseDownloader
}

func (d *thumbnailDownloader) Start(status StatusCallback) error {
	if d.options.isViewer {
		return d.options.allocationObj.DownloadThumbnailFromAuthTicket(d.options.localPath,
			d.options.authTicket, d.options.lookupHash, d.options.fileName, d.options.rxPay, status)

	}
	return d.options.allocationObj.DownloadThumbnail(d.options.localPath, d.options.remotePath, status)
}
