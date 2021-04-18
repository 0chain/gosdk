package sdk

import (
	"context"
	"errors"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/sdk/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
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

	var (
		allocationID       = "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f"
		allocationTx       = "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f"
		localPath          = downloadWorkerTestDir + "/downloadBlock/alloc/1.txt"
		remoteFilePath     = "/1.txt"
		remoteFilePathHash = fileref.GetReferenceLookup(allocationID, remoteFilePath)
	)

	type fields struct {
		allocationID       string
		allocationTx       string
		blobbers           []*blockchain.StorageNode
		datashards          int
		localpath          string
		statusCallback     StatusCallback
		authTicket         *marker.AuthTicket
		downloadMask       uint64
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
				localpath:         localPath,
				datashards:         0,
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
				localpath:         localPath,
				downloadMask:      3,
				datashards:         2,
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
				allocationID:       allocationID,
				allocationTx:       allocationTx,
				blobbers:           tt.fields.blobbers,
				datashards:         tt.fields.datashards,
				parityshards:       2,
				remotefilepath:     remoteFilePath,
				remotefilepathhash: remoteFilePathHash,
				localpath:          tt.fields.localpath,
				startBlock:         0,
				endBlock:           1,
				numBlocks:          int64(numBlockDownloads),
				downloadMask:       zboxutil.NewUint128(tt.fields.downloadMask),
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
				for i := req.downloadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(uint64(pos)).Not()) {
					pos = i.TrailingZeros()
					d := <-downloadBlockChan[tt.fields.blobbers[pos].ID]
					if tt.resultChData != nil {
						if len(tt.resultChData) > counter {
							d.result <- tt.resultChData[counter]
						}
					}
					counter++
				}
			}()

			got, err := req.downloadBlock(1, 10)
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

func TestDownloadRequest_processDownload(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()

	var blobbers = make([]*blockchain.StorageNode, 0)
	for _, bl := range blobberMocks {
		blobbers = append(blobbers, &blockchain.StorageNode{ID: bl.ID, Baseurl: bl.URL})
	}

	var (
		allocationID         = "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f"
		allocationTx         = "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f"
		localPath            = downloadWorkerTestDir + "/processDownload/alloc/1.txt"
		remoteFilePath       = "/1.txt"
		remoteFilePathHash   = fileref.GetReferenceLookup(allocationID, remoteFilePath)
		blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
			setupBlobberMockResponses(t, blobberMocks, downloadWorkerTestDir+"/processDownload", testcaseName)
			return nil
		}
		blockRespData map[string]*downloadBlock
	)

	type fields struct {
		blobbers           []*blockchain.StorageNode
		remotefilepath     string
		remotefilepathhash string
		localpath          string
		statusCallback     StatusCallback
		contentMode        string
		completedCallback  func(remotepath string, remotepathhash string)
	}
	tests := []struct {
		name                      string
		fields                    fields
		additionalMock            func(t *testing.T, testCase string) (teardown func(t *testing.T))
		blockDownloadResponseMock func(blobber *blockchain.StorageNode, wg *sync.WaitGroup)
		assertionFn               func(assertions *assert.Assertions, testCase string)
	}{
		{
			"Test_Error_Get_File_Consensus_From_Blobbers_Failed",
			fields{
				blobbers:           blobbers,
				remotefilepath:     remoteFilePath,
				remotefilepathhash: remoteFilePathHash,
				localpath:          localPath,
				statusCallback: func() StatusCallback {
					scm := &mocks.StatusCallback{}
					scm.On("Error", allocationTx, remoteFilePath, OpDownload, mock.AnythingOfType("*errors.errorString")).Once()
					return scm
				}(),
				completedCallback: func(remotepath string, remotepathhash string) {},
			},
			nil,
			func(blobber *blockchain.StorageNode, wg *sync.WaitGroup) {
				defer wg.Done()
			},
			func(assertion *assert.Assertions, testCase string) {
				_, err := os.Stat(localPath)
				assertion.Error(err)
			},
		},
		{
			"Test_Success",
			fields{
				blobbers:           blobbers,
				remotefilepath:     remoteFilePath,
				remotefilepathhash: remoteFilePathHash,
				localpath:          localPath,
				statusCallback: func() StatusCallback {
					scm := &mocks.StatusCallback{}
					scm.On("Started", allocationTx, remoteFilePath, OpDownload, 4).Once()
					scm.On("InProgress", allocationTx, remoteFilePath, OpDownload, 4, mock.AnythingOfType("[]uint8")).Once()
					scm.On("Completed", allocationTx, remoteFilePath, "1.txt", "application/octet-stream", 4, OpDownload).Once()
					return scm
				}(),
				contentMode:       DOWNLOAD_CONTENT_FULL,
				completedCallback: func(remotepath string, remotepathhash string) {},
			},
			func(t *testing.T, testCase string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCase)
				blockRespData = map[string]*downloadBlock{
					blobbers[0].ID: {
						BlockChunks: [][]byte{{97, 98}},
						Success:     true,
						idx:         0,
					},
					blobbers[1].ID: {
						BlockChunks: [][]byte{{99, 10}},
						Success:     true,
						idx:         1,
					},
					blobbers[2].ID: {
						BlockChunks: [][]byte{{101, 178}},
						Success:     true,
						idx:         2,
					},
					blobbers[3].ID: {
						BlockChunks: [][]byte{{103, 218}},
						Success:     true,
						idx:         3,
					},
				}
				return func(t *testing.T) {
					blockRespData = nil
					_ = os.Remove(localPath)
				}
			},
			func(blobber *blockchain.StorageNode, wg *sync.WaitGroup) {
				defer wg.Done()
				d := <-downloadBlockChan[blobber.ID]
				d.result <- blockRespData[blobber.ID]
			},
			func(assertion *assert.Assertions, testCase string) {
				_, err := os.Stat(localPath)
				assertion.NoError(err)
			},
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
				allocationID:       allocationID,
				allocationTx:       allocationTx,
				blobbers:           tt.fields.blobbers,
				datashards:         2,
				parityshards:       2,
				remotefilepath:     tt.fields.remotefilepath,
				remotefilepathhash: tt.fields.remotefilepathhash,
				localpath:          tt.fields.localpath,
				numBlocks:          int64(numBlockDownloads),
				rxPay:              false,
				statusCallback:     tt.fields.statusCallback,
				authTicket:         nil,
				ctx:                context.Background(),
				completedCallback:  tt.fields.completedCallback,
				contentMode:        tt.fields.contentMode,
				Consensus: Consensus{
					consensus:       0,
					consensusThresh: 50,
					fullconsensus:   4,
				},
			}
			var d = downloadBlockChan
			downloadBlockChan = make(map[string]chan *BlockDownloadRequest)
			for _, bl := range tt.fields.blobbers {
				downloadBlockChan[bl.ID] = make(chan *BlockDownloadRequest)
			}
			defer func() {
				for _, bl := range tt.fields.blobbers {
					close(downloadBlockChan[bl.ID])
				}
				downloadBlockChan = d
			}()
			var wg = &sync.WaitGroup{}
			for i := 0; i < len(tt.fields.blobbers); i++ {
				wg.Add(1)
				go tt.blockDownloadResponseMock(tt.fields.blobbers[i], wg)
			}

			req.processDownload(context.Background())
			wg.Wait()
			if st, ok := tt.fields.statusCallback.(*mocks.StatusCallback); ok {
				st.Test(t)
				st.AssertExpectations(t)
			}
			tt.assertionFn(assertion, tt.name)
		})
	}
}
