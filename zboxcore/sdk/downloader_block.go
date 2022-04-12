package sdk

type blockDownloader struct {
	baseDownloader
}

func (d *blockDownloader) Start(status StatusCallback) error {
	if d.options.isBlockDownload {
		return d.options.allocationObj.DownloadFromAuthTicketByBlocks(
			d.options.localPath, d.options.authTicket,
			d.options.startBlock, d.options.endBlock, d.options.blocksPerMarker,
			d.options.lookupHash, d.options.fileName, d.options.rxPay, status)
	}

	return d.options.allocationObj.DownloadFileByBlock(d.options.localPath, d.options.remotePath,
		d.options.startBlock, d.options.endBlock, d.options.blocksPerMarker,
		status)
}
