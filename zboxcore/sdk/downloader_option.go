package sdk

import "github.com/0chain/gosdk/core/sys"

const DefaultBlocksPerMarker int = 100

// DownloadOption set download option
type DownloadOption func(do *DownloadOptions)

func WithVerifyDownload(shouldVerify bool) DownloadOption {
	return func(do *DownloadOptions) {
		do.verifyDownload = shouldVerify
	}
}

// WithAllocation set allocation object of the download option
func WithAllocation(obj *Allocation) DownloadOption {
	return func(do *DownloadOptions) {
		do.allocationObj = obj
	}
}

// WithBlocks set block range for the download request options
func WithBlocks(start, end int64, blocksPerMarker int) DownloadOption {
	return func(do *DownloadOptions) {
		if start > 0 && end > 0 && end >= start {
			do.isBlockDownload = true

			do.startBlock = start
			do.endBlock = end
		}

		do.blocksPerMarker = blocksPerMarker
		if do.blocksPerMarker < 1 {
			do.blocksPerMarker = DefaultBlocksPerMarker
		}

		SetNumBlockDownloads(do.blocksPerMarker)
	}
}

// WithOnlyThumbnail set thumbnail download option which makes the request download only the thumbnail.
func WithOnlyThumbnail(thumbnail bool) DownloadOption {
	return func(do *DownloadOptions) {
		do.isThumbnailDownload = thumbnail
	}
}

// WithAuthTicket set auth ticket and lookup hash for the download request options
func WithAuthticket(authTicket, lookupHash string) DownloadOption {
	return func(do *DownloadOptions) {
		if len(authTicket) > 0 {
			do.isViewer = true
			do.authTicket = authTicket
			do.lookupHash = lookupHash
		}
	}
}

// WithFileHandler set file handler for the download request options
func WithFileHandler(fileHandler sys.File) DownloadOption {
	return func(do *DownloadOptions) {
		do.fileHandler = fileHandler
		do.isFileHandlerDownload = true
	}
}
