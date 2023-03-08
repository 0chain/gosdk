package sdk

type fileDownloader struct {
	baseDownloader
}

func (d *fileDownloader) Start(status StatusCallback) error {
	if d.isViewer {
		return d.allocationObj.DownloadFromAuthTicket(d.localPath,
			d.authTicket, d.lookupHash, d.fileName, status)
	}

	return d.allocationObj.DownloadFile(d.localPath, d.remotePath, status)
}
