package sdk

import (
	"fmt"
	"path"

	"errors"

	"github.com/0chain/gosdk/zboxcore/fileref"
)

// Downloader downloader for file, blocks and thumbnail
type Downloader interface {
	GetAllocation() *Allocation
	GetFileName() string
	Start(status StatusCallback) error
}

// DownloadOptions download options
type DownloadOptions struct {
	allocationObj *Allocation
	fileName      string

	localPath  string
	remotePath string

	isViewer   bool
	authTicket string
	lookupHash string

	isBlockDownload bool
	blocksPerMarker int
	startBlock      int64
	endBlock        int64

	isThumbnailDownload bool
	verifyDownload      bool
}

// CreateDownloader create a downloander
func CreateDownloader(allocationID, localPath, remotePath string, opts ...DownloadOption) (Downloader, error) {
	do := &DownloadOptions{
		localPath:  localPath,
		remotePath: remotePath,
	}

	if len(remotePath) > 0 {
		do.fileName = path.Base(remotePath)
	}

	for _, option := range opts {
		option(do)
	}

	var err error
	if do.allocationObj == nil {
		if do.isViewer {
			do.allocationObj, err = GetAllocationFromAuthTicket(do.authTicket)
			if err != nil {
				return nil, err
			}
		} else {
			do.allocationObj, err = GetAllocation(allocationID)
			if err != nil {
				return nil, err
			}
		}
	}

	// fixed fileName if only auth ticket/lookup are known
	if len(do.fileName) == 0 {
		if do.isViewer {
			at, err := InitAuthTicket(do.authTicket).Unmarshall()

			if err != nil {
				return nil, err
			}

			if at.RefType == fileref.FILE {
				do.fileName = at.FileName
				do.lookupHash = at.FilePathHash
			} else if len(do.lookupHash) > 0 {
				fileMeta, err := do.allocationObj.GetFileMetaFromAuthTicket(do.authTicket, do.lookupHash)
				if err != nil {
					return nil, err
				}
				do.fileName = fileMeta.Name
			} else if len(remotePath) > 0 {
				do.lookupHash = fileref.GetReferenceLookup(do.allocationObj.Tx, remotePath)
				do.fileName = path.Base(remotePath)
			} else {
				return nil, errors.New("Either remotepath or lookuphash is required when using authticket of directory type")
			}
		} else {
			return nil, errors.New("remotepath is required")
		}
	}

	fmt.Println("download: ", do.isViewer, do.allocationObj.ID, do.fileName, do.lookupHash)

	if do.isThumbnailDownload {
		return &thumbnailDownloader{
			baseDownloader: baseDownloader{
				options: do,
			},
		}, nil
	} else if do.isBlockDownload {
		return &blockDownloader{
			baseDownloader: baseDownloader{
				options: do,
			},
		}, nil
	}

	return &fileDownloader{
		baseDownloader: baseDownloader{
			options: do,
		},
	}, nil
}

type baseDownloader struct {
	options *DownloadOptions
}

func (d *baseDownloader) GetAllocation() *Allocation {
	if d == nil || d.options == nil {
		return nil
	}
	return d.options.allocationObj
}

func (d *baseDownloader) GetFileName() string {
	if d == nil || d.options == nil {
		return ""
	}

	return d.options.fileName
}
