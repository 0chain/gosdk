package sdk

type fileHandlerDownloader struct {
	baseDownloader
}

func (d *fileHandlerDownloader) Start(status StatusCallback, isFinal bool) error {
	if d.isThumbnailDownload {
		if d.isViewer {
			return d.allocationObj.DownloadThumbnailToFileHandlerFromAuthTicket(d.fileHandler,
				d.authTicket, d.lookupHash, d.fileName, d.verifyDownload, status, isFinal)
		}

		return d.allocationObj.DownloadThumbnailToFileHandler(d.fileHandler,
			d.remotePath, d.verifyDownload, status, isFinal, d.reqOpts...)
	} else if d.isBlockDownload {
		if d.isViewer {
			return d.allocationObj.DownloadByBlocksToFileHandlerFromAuthTicket(d.fileHandler,
				d.authTicket, d.lookupHash, d.startBlock, d.endBlock, d.blocksPerMarker,
				d.fileName, d.verifyDownload, status, isFinal)
		}

		return d.allocationObj.DownloadByBlocksToFileHandler(d.fileHandler,
			d.remotePath, d.startBlock, d.endBlock, d.blocksPerMarker,
			d.verifyDownload, status, isFinal, d.reqOpts...)
	}
	if d.isViewer {
		return d.allocationObj.DownloadFileToFileHandlerFromAuthTicket(d.fileHandler,
			d.authTicket, d.lookupHash, d.fileName, d.verifyDownload, status, isFinal)
	}

	return d.allocationObj.DownloadFileToFileHandler(d.fileHandler,
		d.remotePath, d.verifyDownload, status, isFinal, d.reqOpts...)
}
