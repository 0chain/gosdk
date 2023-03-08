package sdk

type blockDownloader struct {
	baseDownloader
}

func (d *blockDownloader) Start(status StatusCallback) error {
	if d.isViewer {
		return d.allocationObj.DownloadFromAuthTicketByBlocks(
			d.localPath, d.authTicket,
			d.startBlock, d.endBlock, d.blocksPerMarker,
			d.lookupHash, d.fileName, status)
	}

	return d.allocationObj.DownloadFileByBlock(d.localPath, d.remotePath,
		d.startBlock, d.endBlock, d.blocksPerMarker,
		status)
}
