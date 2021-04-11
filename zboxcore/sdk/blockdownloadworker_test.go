package sdk

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/sdk/mocks"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sync"
	"testing"
)

const blockDownloadWorkerTestDir = configDir + "/blockdownloadworker"

func Test_getBlobberReadCtr(t *testing.T) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	var blobberID = hex.EncodeToString(bytes)
	tests := []struct {
		name           string
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		want           int64
	}{
		{
			"Test_Not_Found_Blobber_Read_Ctr_On_Map",
			func(t *testing.T) (teardown func(t *testing.T)) {
				br := blobberReadCounter
				blobberReadCounter = &sync.Map{}
				return func(t *testing.T) {
					blobberReadCounter = br
				}
			},
			0,
		},
		{
			"Test_Found_Blobber_Read_Ctr_On_Map",
			func(t *testing.T) (teardown func(t *testing.T)) {
				br := blobberReadCounter
				blobberReadCounter = &sync.Map{}
				blobberReadCounter.Store(blobberID, int64(1))
				return func(t *testing.T) {
					blobberReadCounter = br
				}
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if mock := tt.additionalMock; mock != nil {
				if teardown := mock(t); teardown != nil {
					defer teardown(t)
				}
			}
			assertion.Equal(tt.want, getBlobberReadCtr(&blockchain.StorageNode{ID: blobberID}))
		})
	}
}

func Test_incBlobberReadCtr(t *testing.T) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	var blobberID = hex.EncodeToString(bytes)
	tests := []struct {
		name           string
		numBlocks      int64
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		want           int64
	}{
		{
			"Test_Increasing_Not_Existed_Blobber_Read_Ctr_On_Map",
			2,
			func(t *testing.T) (teardown func(t *testing.T)) {
				br := blobberReadCounter
				blobberReadCounter = &sync.Map{}
				return func(t *testing.T) {
					blobberReadCounter = br
				}
			},
			2,
		},
		{
			"Test_Increasing_Existed_Blobber_Read_Ctr_On_Map",
			3,
			func(t *testing.T) (teardown func(t *testing.T)) {
				br := blobberReadCounter
				blobberReadCounter = &sync.Map{}
				blobberReadCounter.Store(blobberID, int64(1))
				return func(t *testing.T) {
					blobberReadCounter = br
				}
			},
			4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if mock := tt.additionalMock; mock != nil {
				if teardown := mock(t); teardown != nil {
					defer teardown(t)
				}
			}
			incBlobberReadCtr(&blockchain.StorageNode{ID: blobberID}, tt.numBlocks)
			value, ok := blobberReadCounter.Load(blobberID)
			assertion.Truef(ok, "block read counter must able to load %v", blobberID)
			v, ok := value.(int64)
			assertion.True(ok, "value loaded from block read counter must be 64 bit integer number")
			assertion.Equal(tt.want, v)
		})
	}
}

func Test_setBlobberReadCtr(t *testing.T) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	randNum := rand.Int63()
	var blobberID = hex.EncodeToString(bytes)
	tests := []struct {
		name           string
		counter        int64
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
	}{
		{
			"Test_Success_With_Setting_Counter_2",
			2,
			func(t *testing.T) (teardown func(t *testing.T)) {
				br := blobberReadCounter
				blobberReadCounter = &sync.Map{}
				return func(t *testing.T) {
					blobberReadCounter = br
				}
			},
		},
		{
			"Test_Success_With_Setting_Counter_Any_64_Bit_Integer_Number",
			randNum,
			func(t *testing.T) (teardown func(t *testing.T)) {
				br := blobberReadCounter
				blobberReadCounter = &sync.Map{}
				return func(t *testing.T) {
					blobberReadCounter = br
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if mock := tt.additionalMock; mock != nil {
				if teardown := mock(t); teardown != nil {
					defer teardown(t)
				}
			}
			setBlobberReadCtr(&blockchain.StorageNode{ID: blobberID}, tt.counter)
			value, ok := blobberReadCounter.Load(blobberID)
			assertion.Truef(ok, "block read counter must able to load %v", blobberID)
			v, ok := value.(int64)
			assertion.True(ok, "value loaded from block read counter must be 64 bit integer number")
			assertion.Equal(tt.counter, v)
		})
	}
}

func TestInitBlockDownloader(t *testing.T) {
	var bl = mocks.NewBlobberHTTPServer(t)
	defer bl.Close()
	var dbc = downloadBlockChan
	downloadBlockChan = nil
	defer func() { downloadBlockChan = dbc }()
	assertion := assert.New(t)
	InitBlockDownloader([]*blockchain.StorageNode{{Baseurl: bl.URL, ID: bl.ID}})
	assertion.Equal(1, len(downloadBlockChan))
	assertion.NotNil(downloadBlockChan[bl.ID])
	defer func() {
		close(downloadBlockChan[bl.ID])
		delete(downloadBlockChan, bl.ID)
	}()
}

func Test_startBlockDownloadWorker(t *testing.T) {
	var ch = make(chan *BlockDownloadRequest, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		startBlockDownloadWorker(ch)
	}(wg)
	rsCh := make(chan *downloadBlock, 1)
	defer close(rsCh)
	ch <- &BlockDownloadRequest{wg: func() *sync.WaitGroup { wg := &sync.WaitGroup{}; wg.Add(1); return wg }(), result: rsCh}
	result := <-rsCh
	assert.False(t, result.Success)
	assert.Error(t, result.err)
	close(ch)
	wg.Wait()
}

func TestBlockDownloadRequest_splitData(t *testing.T) {
	type args struct {
		buf []byte
		lim int
	}
	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		{
			"Test_Empty_Buffer_Success",
			args{
				[]byte{},
				1,
			},
			[][]byte{},
		},
		{
			"Test_Valuable_Buffer_Length_With_Less_Limit_Number_Success",
			args{
				[]byte("abcde"),
				2,
			},
			[][]byte{[]byte("ab"), []byte("cd"), []byte("e")},
		},
		{
			"Test_Valuable_Limit_Number_With_Less_Buffer_Length_Success",
			args{
				[]byte("abcd"),
				6,
			},
			[][]byte{[]byte("abcd")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			req := &BlockDownloadRequest{}
			assertion.EqualValues(tt.want, req.splitData(tt.args.buf, tt.args.lim))
		})
	}
}

func TestBlockDownloadRequest_downloadBlobberBlock(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	req := &BlockDownloadRequest{
		blobber:            a.Blobbers[0],
		allocationID:       a.ID,
		allocationTx:       a.Tx,
		blobberIdx:         0,
		remotefilepath:     "/1.txt",
		remotefilepathhash: fileref.GetReferenceLookup(a.ID, "/1.txt"),
		blockNum:           1,
		encryptedKey:       "",
		contentMode:        DOWNLOAD_CONTENT_FULL,
		numBlocks:          int64(numBlockDownloads),
		rxPay:              true,
		authTicket:         nil,
		wg:                 &sync.WaitGroup{},
		ctx:                context.Background(),
		result:             nil,
	}
	var wClient = client.GetClient()
	var resultChan chan *downloadBlock
	var prepairAndResetTestCasesData = func(t *testing.T) (teardown func(t *testing.T)) {
		req.wg.Add(1)
		resultChan = make(chan *downloadBlock)
		rsChan := req.result
		req.result = resultChan
		return func(t *testing.T) {
			close(resultChan)
			req.result = rsChan
		}
	}
	var prepairAndResetBlobberReadCounterTestCaseData = func(t *testing.T) (teardown func(t *testing.T)) {
		brc := blobberReadCounter
		blobberReadCounter = &sync.Map{}
		return func(t *testing.T) {
			blobberReadCounter = brc
		}
	}
	var wantFailedFn = func(t *testing.T, testCaseName string, wg *sync.WaitGroup) {
		defer wg.Done()
		rs := <-req.result
		assert.False(t, rs.Success)
		assert.Equal(t, req.blobberIdx, 0)
		assert.NotNil(t, rs.err)
	}
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, blockDownloadWorkerTestDir+"/downloadBlobberBlock", testcaseName)
		return nil
	}

	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantFn         func(t *testing.T, testCaseName string, wg *sync.WaitGroup)
	}{
		{
			"Test_Zero_NumBlocks_Failed",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				td := prepairAndResetTestCasesData(t)
				numBlocks := req.numBlocks
				req.numBlocks = 0
				return func(t *testing.T) {
					req.numBlocks = numBlocks
					td(t)
				}
			},
			wantFailedFn,
		},
		{
			"Test_Skip_Blobber_Failed",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				td := prepairAndResetTestCasesData(t)
				isKip := a.Blobbers[0].IsSkip()
				a.Blobbers[0].SetSkip(true)
				return func(t *testing.T) {
					a.Blobbers[0].SetSkip(isKip)
					td(t)
				}
			},
			wantFailedFn,
		},
		{
			"Test_Error_Sign_Failed",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				keys := wClient.Keys
				wClient.Keys = []zcncrypto.KeyPair{{}}
				td1 := prepairAndResetTestCasesData(t)
				td2 := prepairAndResetBlobberReadCounterTestCaseData(t)
				return func(t *testing.T) {
					wClient.Keys = keys
					td1(t)
					td2(t)
				}
			},
			wantFailedFn,
		},
		{
			"Test_Error_Create_Http_Request_Failed",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				url := req.blobber.Baseurl
				req.blobber.Baseurl = string([]byte{0x7f, 0, 0})
				td1 := prepairAndResetTestCasesData(t)
				td2 := prepairAndResetBlobberReadCounterTestCaseData(t)
				return func(t *testing.T) {
					req.blobber.Baseurl = url
					td1(t)
					td2(t)
				}
			},
			wantFailedFn,
		},
		{
			"Test_Error_Blobber_Download_Block_Failed",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				td1 := prepairAndResetTestCasesData(t)
				td2 := prepairAndResetBlobberReadCounterTestCaseData(t)
				return func(t *testing.T) {
					td1(t)
					td2(t)
				}
			},
			wantFailedFn,
		},
		{
			"Test_Blobber_Response_Not_Success_With_Status_200_Failed",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				td1 := prepairAndResetTestCasesData(t)
				td2 := prepairAndResetBlobberReadCounterTestCaseData(t)
				skip := req.blobber.IsSkip()
				return func(t *testing.T) {
					td1(t)
					td2(t)
					req.blobber.SetSkip(skip)
				}
			},
			nil,
		},
		{
			"Test_Error_Blobber_Response_With_Status_400_Failed",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				td1 := prepairAndResetTestCasesData(t)
				td2 := prepairAndResetBlobberReadCounterTestCaseData(t)
				skip := req.blobber.IsSkip()
				return func(t *testing.T) {
					td1(t)
					td2(t)
					req.blobber.SetSkip(skip)
				}
			},
			wantFailedFn,
		},
		{
			"Test_Success_With_Auth_Ticket",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				td1 := prepairAndResetTestCasesData(t)
				td2 := prepairAndResetBlobberReadCounterTestCaseData(t)
				authTicket, err := a.GetAuthTicket("/1/txt", "1.txt", fileref.FILE, client.GetClientID(), "")
				assert.NoErrorf(t, err, "unexpected get auth ticket error: %v", err)
				assert.NotEmptyf(t, authTicket, "unexpected empty auth ticket")
				sEnc, err := base64.StdEncoding.DecodeString(authTicket)
				assert.NoErrorf(t, err, "unexpected decode auth ticket error: %v", err)
				err = json.Unmarshal(sEnc, &req.authTicket)
				assert.NoErrorf(t, err, "unexpected error when marshaling auth ticket error: %v", err)
				return func(t *testing.T) {
					td1(t)
					td2(t)
					req.authTicket = nil
				}
			},
			func(t *testing.T, testCaseName string, wg *sync.WaitGroup) {
				defer wg.Done()
				rs := <-req.result
				assert.True(t, rs.Success)
				assert.Equal(t, req.blobberIdx, 0)
				assert.Nil(t, rs.err)
				expectedBytes := parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v", blockDownloadWorkerTestDir, "downloadBlobberBlock", testCaseName), nil)
				assert.Equal(t, expectedBytes, rs.BlockChunks[0])
			},
		},
		{
			"Test_Success",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				td1 := prepairAndResetTestCasesData(t)
				td2 := prepairAndResetBlobberReadCounterTestCaseData(t)
				return func(t *testing.T) {
					td1(t)
					td2(t)
				}
			},
			func(t *testing.T, testCaseName string, wg *sync.WaitGroup) {
				defer wg.Done()
				rs := <-req.result
				assert.True(t, rs.Success)
				assert.Equal(t, req.blobberIdx, 0)
				assert.Nil(t, rs.err)
				expectedBytes := parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v", blockDownloadWorkerTestDir, "downloadBlobberBlock", testCaseName), nil)
				assert.Equal(t, expectedBytes, rs.BlockChunks[0])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if mock := tt.additionalMock; mock != nil {
				if teardown := mock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			wg := sync.WaitGroup{}
			if tt.wantFn != nil {
				wg.Add(1)
				go tt.wantFn(t, tt.name, &wg)
			}
			req.downloadBlobberBlock()
			wg.Wait()
		})
	}
}

func TestAddBlockDownloadReq(t *testing.T) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	var blobberID = hex.EncodeToString(bytes)
	downloadBlockChan = make(map[string]chan *BlockDownloadRequest)
	ch := make(chan *BlockDownloadRequest)
	defer close(ch)
	downloadBlockChan[blobberID] = ch
	defer delete(downloadBlockChan, blobberID)
	in := &BlockDownloadRequest{blobber: &blockchain.StorageNode{ID: blobberID}}
	go func() {
		out := <-downloadBlockChan[blobberID]
		assert.Equal(t, in, out)
	}()
	AddBlockDownloadReq(in)
}
