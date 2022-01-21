package sdk

type fileDownloader struct {
	baseDownloader
}

func (d *fileDownloader) Start(status StatusCallback) error {
	if d.options.isViewer {
		return d.options.allocationObj.DownloadFromAuthTicket(d.options.localPath,
			d.options.authTicket, d.options.lookupHash, d.options.fileName, d.options.rxPay, status)
	}

	return d.options.allocationObj.DownloadFile(d.options.localPath, d.options.remotePath, status)
}
