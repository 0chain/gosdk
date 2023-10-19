//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"fmt"

	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

type FilePlayer struct {
	allocationID  string
	remotePath    string
	authTicket    string
	lookupHash    string
	startBlock    int64
	reqC          chan int64
	numBlocks     int
	isViewer      bool
	allocationObj *sdk.Allocation
	authTicketObj *marker.AuthTicket
	playlistFile  *sdk.PlaylistFile

	downloadedChunks chan *downloadChunks
	downloadedLen    int
	ctx              context.Context
	cancel           context.CancelFunc
	prefetchQty      int
}

type downloadChunks struct {
	data       []byte
	startBlock int64
}

func (p *FilePlayer) Start() error {
	if p.cancel != nil {
		p.cancel()
	}

	p.ctx, p.cancel = context.WithCancel(context.TODO())

	file, err := p.loadPlaylistFile()
	if err != nil {
		return err
	}

	p.reqC = make(chan int64, p.prefetchQty)
	p.reqC <- 1
	// p.startBlock = 1
	p.playlistFile = file

	p.downloadedChunks = make(chan *downloadChunks, p.prefetchQty)

	go p.startDownload()

	return nil
}

func (p *FilePlayer) Stop() {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
}

func (p *FilePlayer) download(startBlock int64) {
	endBlock := startBlock + int64(p.numBlocks) - 1

	if endBlock > p.playlistFile.NumBlocks {
		endBlock = p.playlistFile.NumBlocks
	}
	fmt.Println("### start:", startBlock, "end:", endBlock, "numBlocks:", p.numBlocks, "total:", p.playlistFile.NumBlocks)

	data, err := downloadBlocks(p.allocationObj, p.remotePath, p.authTicket, p.lookupHash, startBlock, endBlock)
	if err != nil {
		PrintError(err.Error())
		return
	}
	withRecover(func() {
		if p.downloadedChunks != nil {
			p.downloadedChunks <- &downloadChunks{
				data:       data,
				startBlock: startBlock,
			}
		}
	})
}

func (p *FilePlayer) startDownload() {
	fmt.Println("### start download")
	if p.playlistFile.NumBlocks < 1 {
		PrintError("### playlist: numBlocks is invalid")
		return
	}

	var prevBlockNum int64

	for {
		select {
		case <-p.ctx.Done():
			PrintInfo("### playlist: download is cancelled")
			return
		default:
			startBlock := <-p.reqC
			if startBlock < prevBlockNum {
				continue
			}
			fmt.Println(">>> download start:", startBlock)
			p.download(startBlock)

			prevBlockNum = startBlock
			startBlock = startBlock + int64(p.numBlocks)
			if startBlock+int64(p.numBlocks) > p.playlistFile.NumBlocks {
				go func() {
					// trigger js to close stream
					fmt.Println("### end of file")
					close(p.downloadedChunks)
				}()
				return
			}

			fmt.Println("<<< download end, new start:", startBlock)
			p.reqC <- startBlock
		}
	}
}

func (p *FilePlayer) loadPlaylistFile() (*sdk.PlaylistFile, error) {
	if p.isViewer {
		//get playlist file from auth ticket
		return sdk.GetPlaylistFileByAuthTicket(p.ctx, p.allocationObj, p.authTicket, p.lookupHash)
	}

	d, err := p.allocationObj.ListDir(p.remotePath)
	if err != nil {
		fmt.Println("could not list dir:", p.remotePath)
		return nil, err
	}
	f := d.Children[0]
	var (
		dataShards            = p.allocationObj.DataShards
		effectivePerShardSize = (int(f.ActualSize) + dataShards - 1) / dataShards
		totalBlocks           = (effectivePerShardSize + sdk.DefaultChunkSize - 1) / sdk.DefaultChunkSize
		// chunkDuration         = (p.numBlocks * sdk.DefaultChunkSize * p.allocationObj.DataShards * 100) / int(f.ActualSize)
	)

	fmt.Println("totalBlocks:", totalBlocks)
	fmt.Println("file size:", f.Size)
	// fmt.Println("chunk duration:", chunkDuration)

	return &sdk.PlaylistFile{
		Name:           f.Name,
		Path:           f.Path,
		LookupHash:     f.LookupHash,
		NumBlocks:      int64(totalBlocks),
		Size:           f.Size,
		ActualFileSize: f.ActualSize,
		MimeType:       f.MimeType,
		Type:           f.Type,
	}, nil
}

func (p *FilePlayer) GetNext() []byte {
	chunks, ok := <-p.downloadedChunks
	if ok {
		b := chunks.data
		fmt.Println("### get next block data:", chunks.startBlock)
		if chunks.startBlock+int64(p.numBlocks) >= p.playlistFile.NumBlocks {
			startIndex := int(chunks.startBlock-1) * sdk.DefaultChunkSize * p.allocationObj.DataShards

			// already read data is startIndex
			rest := p.playlistFile.ActualFileSize - int64(startIndex)
			b = b[:rest]
		}

		return b
	}
	return nil
}

func (p *FilePlayer) VideoSeek(position int64) []byte {
	seekStartBlock := p.playlistFile.NumBlocks * position / 46 // 46 is the duration of the video, this is for testing the sample video file
	// set startBlock to new position
	fmt.Println("### sdk seek to:", position, " new start:", seekStartBlock)
	go func() {
		p.reqC <- seekStartBlock
	}()
	return nil
}

// createFilePalyer create player for remotePath
func createFilePalyer(allocationID, remotePath, authTicket, lookupHash string) (*FilePlayer, error) {
	player := &FilePlayer{}
	player.prefetchQty = 3
	player.remotePath = remotePath
	player.authTicket = authTicket
	player.lookupHash = lookupHash
	player.numBlocks = 5 // change back to 10 after debugging, 5 is for not downloading too fast to test the playback
	player.allocationID = allocationID

	//player is viewer
	if len(authTicket) > 0 {
		//player is viewer via shared authticket
		at, err := sdk.InitAuthTicket(authTicket).Unmarshall()

		if err != nil {
			PrintError(err)
			return nil, err
		}

		allocationObj, err := sdk.GetAllocationFromAuthTicket(authTicket)
		if err != nil {
			PrintError("Error fetching the allocation", err)
			return nil, err
		}

		player.isViewer = true
		player.allocationObj = allocationObj
		player.authTicketObj = at
		player.lookupHash = at.FilePathHash

		return player, nil

	}

	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	player.isViewer = false
	player.allocationObj = allocationObj

	return player, nil
}
