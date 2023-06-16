package sdk

type blockDownloader struct {
	baseDownloader
}

func (d *blockDownloader) Start(status StatusCallback, isFinal bool) error {
	if d.isViewer {
		return d.allocationObj.DownloadFromAuthTicketByBlocks(
			d.localPath, d.authTicket,
			d.startBlock, d.endBlock, d.blocksPerMarker,
			d.lookupHash, d.fileName, d.verifyDownload, status, isFinal)
	}

	return d.allocationObj.DownloadFileByBlock(d.localPath, d.remotePath,
		d.startBlock, d.endBlock, d.blocksPerMarker, d.verifyDownload,
		status, isFinal)
}
