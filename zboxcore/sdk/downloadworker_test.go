package sdk

import (
	"context"
	"errors"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/stretchr/testify/assert"
	"math/bits"
	"sync"
	"testing"
)

const downloadWorkerTestDir = configDir + "/downloadworker"

func TestDownloadRequest_downloadBlock(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()

	var blobbers = make([]*blockchain.StorageNode, 0)
	for _, bl := range blobberMocks {
		blobbers = append(blobbers, &blockchain.StorageNode{ID: bl.ID, Baseurl: bl.URL})
	}

	type fields struct {
		allocationID       string
		allocationTx       string
		blobbers           []*blockchain.StorageNode
		datashard          int
		localpath          string
		statusCallback     StatusCallback
		authTicket         *marker.AuthTicket
		downloadMask       uint32
		encryptedKey       string
		isDownloadCanceled bool
		completedCallback  func(remotepath string, remotepathhash string)
		contentMode        string
	}
	tests := []struct {
		name           string
		fields         fields
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		resultChData   []*downloadBlock
		want           []byte
		wantErr        bool
	}{
		{
			"Test_Result_Error_With_Zero_Data_Shard_And_Encrypted_Key_Failed",
			fields{
				blobbers:          blobbers,
				localpath:         downloadWorkerTestDir + "/alloc",
				datashard:         0,
				downloadMask:      1,
				encryptedKey:      "F01uReOJTdgFOMxhleNYqpOpyFbFSltEcwv8G8kwHJo=",
				completedCallback: func(remotepath string, remotepathhash string) {},
				contentMode:       DOWNLOAD_CONTENT_FULL,
			},
			nil,
			[]*downloadBlock{{
				Success: false,
				idx:     0,
				err:     errors.New("some thing error"),
			}},
			[]byte{},
			true,
		},
		{
			"Test_Success_With_No_Encrypted_Key",
			fields{
				blobbers:          blobbers,
				localpath:         downloadWorkerTestDir + "/alloc",
				downloadMask:      3,
				datashard:         2,
				encryptedKey:      "",
				completedCallback: func(remotepath string, remotepathhash string) {},
				contentMode:       DOWNLOAD_CONTENT_FULL,
			},
			nil,
			[]*downloadBlock{
				{
					Success:     true,
					BlockChunks: [][]byte{[]byte("ab"), []byte("ef")},
					idx:         0,
				},
				{
					Success:     true,
					BlockChunks: [][]byte{[]byte("cd"), []byte("gh")},
					idx:         1,
				}},
			[]byte("abcdefgh"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if tt.additionalMock != nil {
				if teardown := tt.additionalMock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			req := &DownloadRequest{
				allocationID:       "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx:       "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				blobbers:           tt.fields.blobbers,
				datashards:         tt.fields.datashard,
				parityshards:       2,
				remotefilepath:     "/1.txt",
				remotefilepathhash: fileref.GetReferenceLookup("69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", "/1.txt"),
				localpath:          tt.fields.localpath,
				startBlock:         1,
				endBlock:           0,
				numBlocks:          2,
				downloadMask:       tt.fields.downloadMask,
				encryptedKey:       tt.fields.encryptedKey,
				completedCallback:  tt.fields.completedCallback,
				contentMode:        tt.fields.contentMode,
				Consensus: Consensus{
					consensusThresh: 50,
					fullconsensus:   4,
				},
				ctx: context.Background(),
			}
			var d = downloadBlockChan
			downloadBlockChan = make(map[string]chan *BlockDownloadRequest)
			for _, bl := range blobbers {
				downloadBlockChan[bl.ID] = make(chan *BlockDownloadRequest)
			}
			defer func() { downloadBlockChan = d }()
			var wg = &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				pos := 0
				counter := 0
				for i := req.downloadMask; i != 0; i &= ^(1 << uint32(pos)) {
					pos = bits.TrailingZeros32(i)
					d := <-downloadBlockChan[tt.fields.blobbers[pos].ID]
					if tt.resultChData != nil {
						if len(tt.resultChData) > counter {
							d.result <- tt.resultChData[counter]
						}
					}
					counter++
				}
			}()

			got, err := req.downloadBlock(4)
			wg.Wait()
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error %v", err)
			assertion.EqualValues(tt.want, got)
		})
	}
}
