package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/0chain/gosdk/zboxcore/encryption"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/resty"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	tokenUnit          = 10000000000.0
	mockAllocationId   = "mock allocation id"
	mockAllocationTxId = "mock transaction id"
	mockClientId       = "mock client id"
	mockClientKey      = "mock client key"
	mockBlobberId      = "mock blobber id"
	mockBlobberUrl     = "mockBlobberUrl"
	mockLookupHash     = "mock lookup hash"
	mockAllocationRoot = "mock allocation root"
	mockFileRefName    = "mock file ref name"
	numBlobbers        = 4
)

func setupMockHttpResponse(t *testing.T, mockClient *mocks.HttpClient, funcName string, testCaseName string, a *Allocation, httpMethod string, statusCode int, body []byte) {
	for i := 0; i < numBlobbers; i++ {
		url := funcName + testCaseName + mockBlobberUrl + strconv.Itoa(i)
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == httpMethod &&
				strings.HasPrefix(req.URL.Path, url)
		})).Return(&http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, nil).Once()
	}
}

func setupMockCommitRequest(a *Allocation) {
	commitChan = make(map[string]chan *CommitRequest)
	for _, blobber := range a.Blobbers {
		if _, ok := commitChan[blobber.ID]; !ok {
			commitChan[blobber.ID] = make(chan *CommitRequest, 1)
			blobberChan := commitChan[blobber.ID]
			go func(c <-chan *CommitRequest, blID string) {
				for true {
					cm := <-c
					if cm != nil {
						cm.result = &CommitResult{
							Success: true,
						}
						if cm.wg != nil {
							cm.wg.Done()
						}
					}
				}
			}(blobberChan, blobber.ID)
		}
	}
}

func setupMockFile(t *testing.T, path string) (teardown func(t *testing.T)) {
	os.Create(path)
	ioutil.WriteFile(path, []byte("mockActualHash"), os.ModePerm)
	return func(t *testing.T) {
		os.Remove(path)
	}
}

func TestGetMinMaxWriteReadSuccess(t *testing.T) {
	var ssc = newTestAllocation()
	ssc.DataShards = 5
	ssc.ParityShards = 4

	ssc.initialized = true
	sdkInitialized = true
	require.NotNil(t, ssc.BlobberDetails)

	t.Run("Success minR, minW", func(t *testing.T) {
		minW, minR, err := ssc.GetMinWriteRead()
		require.NoError(t, err)
		require.Equal(t, 0.01, minW)
		require.Equal(t, 0.01, minR)
	})

	t.Run("Success maxR, maxW", func(t *testing.T) {
		maxW, maxR, err := ssc.GetMaxWriteRead()
		require.NoError(t, err)
		require.Equal(t, 0.01, maxW)
		require.Equal(t, 0.01, maxR)
	})

	t.Run("Error / No Blobbers", func(t *testing.T) {
		var (
			ssc = newTestAllocationEmptyBlobbers()
			err error
		)
		ssc.initialized = true
		_, _, err = ssc.GetMinWriteRead()
		require.Error(t, err)
	})

	t.Run("Error / Empty Blobbers", func(t *testing.T) {
		var err error
		ssc.initialized = false
		_, _, err = ssc.GetMinWriteRead()
		require.Error(t, err)
	})

	t.Run("Error / Not Initialized", func(t *testing.T) {
		var err error
		ssc.initialized = false
		_, _, err = ssc.GetMinWriteRead()
		require.Error(t, err)
	})
}

func TestGetMaxMinStorageCostSuccess(t *testing.T) {
	var ssc = newTestAllocation()
	ssc.DataShards = 4
	ssc.ParityShards = 2

	ssc.initialized = true
	sdkInitialized = true

	t.Run("Storage cost", func(t *testing.T) {
		cost, err := ssc.GetMaxStorageCost(100 * GB)
		require.NoError(t, err)
		require.Equal(t, 1.5, cost)
	})
}

func newTestAllocationEmptyBlobbers() (ssc *Allocation) {
	ssc = new(Allocation)
	ssc.Expiration = 0
	ssc.ID = "ID"
	ssc.BlobberDetails = make([]*BlobberAllocation, 0)
	return ssc
}

func newTestAllocation() (ssc *Allocation) {
	ssc = new(Allocation)
	ssc.Expiration = 0
	ssc.ID = "ID"
	ssc.BlobberDetails = newBlobbersDetails()
	return ssc
}

func newBlobbersDetails() (blobbers []*BlobberAllocation) {
	blobberDetails := make([]*BlobberAllocation, 0)

	for i := 1; i <= 1; i++ {
		var balloc BlobberAllocation
		balloc.Size = 1000

		balloc.Terms = Terms{ReadPrice: common.Balance(100000000), WritePrice: common.Balance(100000000)}
		blobberDetails = append(blobberDetails, &balloc)
	}

	return blobberDetails
}

func TestThrowErrorWhenBlobbersRequiredGreaterThanImplicitLimit128(t *testing.T) {
	setupMocks()

	var maxNumOfBlobbers = 129

	var allocation = &Allocation{}
	var blobbers = make([]*blockchain.StorageNode, maxNumOfBlobbers)
	allocation.initialized = true
	sdkInitialized = true
	allocation.Blobbers = blobbers
	allocation.DataShards = 64
	allocation.ParityShards = 65

	var file fileref.Attributes
	err := allocation.uploadOrUpdateFile("", "/", nil, false, "", false, false, file)

	var expectedErr = "allocation requires [129] blobbers, which is greater than the maximum permitted number of [128]. reduce number of data or parity shards and try again"
	if err == nil {
		t.Errorf("uploadOrUpdateFile() = expected error  but was %v", nil)
	} else if errors.Top(err) != expectedErr {
		t.Errorf("uploadOrUpdateFile() = expected error message to be %v  but was %v", expectedErr, errors.Top(err))
	}
}

func TestThrowErrorWhenBlobbersRequiredGreaterThanExplicitLimit(t *testing.T) {
	setupMocks()

	var maxNumOfBlobbers = 10

	var allocation = &Allocation{}
	var blobbers = make([]*blockchain.StorageNode, maxNumOfBlobbers)
	allocation.initialized = true
	sdkInitialized = true
	allocation.Blobbers = blobbers
	allocation.DataShards = 5
	allocation.ParityShards = 6

	var file fileref.Attributes
	err := allocation.uploadOrUpdateFile("", "/", nil, false, "", false, false, file)

	var expectedErr = "allocation requires [11] blobbers, which is greater than the maximum permitted number of [10]. reduce number of data or parity shards and try again"
	if err == nil {
		t.Errorf("uploadOrUpdateFile() = expected error  but was %v", nil)
	} else if errors.Top(err) != expectedErr {
		t.Errorf("uploadOrUpdateFile() = expected error message to be %v  but was %v", expectedErr, errors.Top(err))
	}
}

func TestDoNotThrowErrorWhenBlobbersRequiredLessThanLimit(t *testing.T) {
	setupMocks()

	var maxNumOfBlobbers = 10

	var allocation = &Allocation{}
	var blobbers = make([]*blockchain.StorageNode, maxNumOfBlobbers)
	allocation.initialized = true
	sdkInitialized = true
	allocation.Blobbers = blobbers
	allocation.DataShards = 5
	allocation.ParityShards = 4

	var file fileref.Attributes
	err := allocation.uploadOrUpdateFile("", "/", nil, false, "", false, false, file)

	if err != nil {
		t.Errorf("uploadOrUpdateFile() = expected no error but was %v", err)
	}
}

func setupMocks() {
	GetFileInfo = func(localpath string) (os.FileInfo, error) {
		return new(MockFile), nil
	}
}

type MockFile struct {
	os.FileInfo
	size int64
}

func (m MockFile) Size() int64 { return 10 }

func TestPriceRange_IsValid(t *testing.T) {
	type fields struct {
		Min int64
		Max int64
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"Test_Valid_InRange",
			fields{
				Min: 0,
				Max: 50,
			},
			true,
		},
		{
			"Test_Valid_At_Once_Value",
			fields{
				Min: 10,
				Max: 10,
			},
			true,
		},
		{
			"Test_Invalid_With_Negative_Value",
			fields{
				Min: -5,
				Max: 10,
			},
			false,
		},
		{
			"Test_Invalid_InRange",
			fields{
				Min: 10,
				Max: 5,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PriceRange{
				Min: tt.fields.Min,
				Max: tt.fields.Max,
			}
			got := pr.IsValid()
			require := require.New(t)
			var check = require.False
			if tt.want {
				check = require.True
			}
			check(got)
		})
	}
}

func TestAllocation_InitAllocation(t *testing.T) {
	a := Allocation{}
	a.InitAllocation()
	require.New(t).NotZero(a)
}

func TestAllocation_dispatchWork(t *testing.T) {
	a := Allocation{DataShards: 2, ParityShards: 2, uploadChan: make(chan *UploadRequest), downloadChan: make(chan *DownloadRequest), repairChan: make(chan *RepairRequest)}
	t.Run("Test_Cover_Context_Canceled", func(t *testing.T) {
		ctx, cancelFn := context.WithCancel(context.Background())
		go a.dispatchWork(ctx)
		cancelFn()
	})
	t.Run("Test_Cover_Upload_Request", func(t *testing.T) {
		go a.dispatchWork(context.Background())
		a.uploadChan <- &UploadRequest{file: []*fileref.FileRef{}, filemeta: &UploadFileMeta{}}
	})
	t.Run("Test_Cover_Download_Request", func(t *testing.T) {
		go a.dispatchWork(context.Background())
		a.downloadChan <- &DownloadRequest{}
	})
	t.Run("Test_Cover_Repair_Request", func(t *testing.T) {
		go a.dispatchWork(context.Background())
		a.repairChan <- &RepairRequest{listDir: &ListResult{}}
	})
}

func TestAllocation_GetStats(t *testing.T) {
	stats := &AllocationStats{}
	a := &Allocation{
		Stats: stats,
	}
	got := a.GetStats()
	require.New(t).Same(stats, got)
}

func TestAllocation_GetBlobberStats(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	tests := []struct {
		name  string
		setup func(*testing.T, string)
	}{
		{
			name: "Test_Success",
			setup: func(t *testing.T, testName string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestAllocation_GetBlobberStats"+testName)
				})).Return(&http.Response{
					Body: func() io.ReadCloser {
						jsonFR, err := json.Marshal(&BlobberAllocationStats{
							ID: mockAllocationId,
							Tx: mockAllocationTxId,
						})
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
					StatusCode: http.StatusOK,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name)
			a := &Allocation{
				ID: mockAllocationId,
				Tx: mockAllocationTxId,
			}
			a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
				ID:      tt.name + mockBlobberId,
				Baseurl: "TestAllocation_GetBlobberStats" + tt.name + mockBlobberUrl,
			})
			got := a.GetBlobberStats()
			require.NotEmptyf(got, "Error no blobber stats result found")

			expected := make(map[string]*BlobberAllocationStats, 1)
			expected["TestAllocation_GetBlobberStats"+tt.name+mockBlobberUrl] = &BlobberAllocationStats{
				ID:         mockAllocationId,
				Tx:         mockAllocationTxId,
				BlobberID:  tt.name + mockBlobberId,
				BlobberURL: "TestAllocation_GetBlobberStats" + tt.name + mockBlobberUrl,
			}

			for key, val := range expected {
				require.NotNilf(got[key], "Error result map must be contain key %v", key)
				require.EqualValues(val, got[key])
			}
		})
	}
}

func TestAllocation_isInitialized(t *testing.T) {
	tests := []struct {
		name                                        string
		sdkInitialized, allocationInitialized, want bool
	}{
		{
			"Test_Initialized",
			true, true, true,
		},
		{
			"Test_SDK_Uninitialized",
			false, true, false,
		},
		{
			"Test_Allocation_Uninitialized",
			true, false, false,
		},
		{
			"Test_Both_SDK_And_Allocation_Uninitialized",
			false, false, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalSDKInitialized := sdkInitialized
			defer func() { sdkInitialized = originalSDKInitialized }()
			sdkInitialized = tt.sdkInitialized
			a := &Allocation{initialized: tt.allocationInitialized}
			got := a.isInitialized()
			require := require.New(t)
			if tt.want {
				require.True(got, `Error a.isInitialized() should returns "true"", but got "false"`)
				return
			}
			require.False(got, `Error a.isInitialized() should returns "false"", but got "true"`)
		})
	}
}

func TestAllocation_CreateDir(t *testing.T) {
	const mockLocalPath = "/test"
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		ParityShards: 2,
		DataShards:   2,
	}
	setupMockAllocation(t, a)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return strings.HasPrefix(req.URL.Path, "TestAllocation_CreateDir")
	})).Return(&http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
	}, nil)

	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: "TestAllocation_CreateDir" + mockBlobberUrl + strconv.Itoa(i),
		})
	}
	err := a.CreateDir(mockLocalPath)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_RepairFile(t *testing.T) {
	const (
		mockFileRefName = "mock file ref name"
		mockLocalPath   = "1.txt"
		mockActualHash  = "4041e3eeb170751544a47af4e4f9d374e76cee1d"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponses := func(t *testing.T, testName string, numBlobbers, numCorrect int) {
		require.True(t, numBlobbers >= numCorrect)
		for i := 0; i < numBlobbers; i++ {
			var hash string
			if i < numCorrect {
				hash = mockActualHash
			}
			frName := mockFileRefName + strconv.Itoa(i)
			url := "TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.Path, url)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileRefName, hash string) io.ReadCloser {
					jsonFR, err := json.Marshal(&fileref.FileRef{
						ActualFileHash: hash,
						Ref: fileref.Ref{
							Name: fileRefName,
						},
					})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(frName, hash),
			}, nil)
		}
	}

	type parameters struct {
		localPath  string
		remotePath string
		status     StatusCallback
	}
	tests := []struct {
		name        string
		parameters  parameters
		numBlobbers int
		numCorrect  int
		setup       func(*testing.T, string, int, int)
		wantErr     bool
		errMsg      string
	}{
		{
			name: "Test_Repair_Not_Required_Failed",
			parameters: parameters{
				localPath:  mockLocalPath,
				remotePath: "/",
			},
			numBlobbers: 4,
			numCorrect:  4,
			setup:       setupHttpResponses,
			wantErr:     true,
			errMsg:      "Repair not required",
		},
		{
			name: "Test_Repair_Required_Success",
			parameters: parameters{
				localPath:  mockLocalPath,
				remotePath: "/",
			},
			numBlobbers: 4,
			numCorrect:  3,
			setup:       setupHttpResponses,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
				defer teardown(t)
			}
			a := &Allocation{
				ParityShards: 2,
				DataShards:   2,
			}
			a.uploadChan = make(chan *UploadRequest, 10)
			a.downloadChan = make(chan *DownloadRequest, 10)
			a.repairChan = make(chan *RepairRequest, 1)
			a.ctx, a.ctxCancelF = context.WithCancel(context.Background())
			a.uploadProgressMap = make(map[string]*UploadRequest)
			a.downloadProgressMap = make(map[string]*DownloadRequest)
			a.mutex = &sync.Mutex{}
			a.initialized = true
			sdkInitialized = true
			setupMockAllocation(t, a)
			for i := 0; i < tt.numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					Baseurl: "TestAllocation_RepairFile" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			tt.setup(t, tt.name, tt.numBlobbers, tt.numCorrect)
			err := a.RepairFile(tt.parameters.localPath, tt.parameters.remotePath, tt.parameters.status)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_RepairRequired(t *testing.T) {
	const (
		mockActualHash = "4041e3eeb170751544a47af4e4f9d374e76cee1d"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	tests := []struct {
		name                          string
		setup                         func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		remotePath                    string
		wantFound                     uint64
		wantFileRef                   *fileref.FileRef
		wantMatchesConsensus, wantErr bool
		errMsg                        string
	}{
		{
			name: "Test_Not_Repair_Required_Success",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					url := "TestAllocation_RepairRequired" + testCaseName + mockBlobberUrl + strconv.Itoa(i)
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						Baseurl: url,
					})
					mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return strings.HasPrefix(req.URL.Path, url)
					})).Return(&http.Response{
						StatusCode: http.StatusOK,
						Body: func() io.ReadCloser {
							jsonFR, err := json.Marshal(&fileref.FileRef{
								ActualFileHash: mockActualHash,
							})
							require.NoError(t, err)
							return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
						}(),
					}, nil)
				}
				return nil
			},
			remotePath:           "/x.txt",
			wantFound:            0xf,
			wantMatchesConsensus: false,
			wantErr:              false,
		},
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) { a.initialized = true }
			},
			remotePath:           "/",
			wantFound:            0,
			wantMatchesConsensus: false,
			wantErr:              true,
			errMsg:               "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Repair_Required_Success",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					var hash string
					if i < numBlobbers-1 {
						hash = mockActualHash
					}
					url := "TestAllocation_RepairRequired" + testCaseName + mockBlobberUrl + strconv.Itoa(i)
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						Baseurl: url,
					})
					mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return strings.HasPrefix(req.URL.Path, url)
					})).Return(&http.Response{
						StatusCode: http.StatusOK,
						Body: func(hash string) io.ReadCloser {
							jsonFR, err := json.Marshal(&fileref.FileRef{
								ActualFileHash: hash,
							})
							require.NoError(t, err)
							return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
						}(hash),
					}, nil)
				}
				return nil
			},
			remotePath:           "/",
			wantFound:            0x7,
			wantMatchesConsensus: true,
			wantErr:              false,
		},
		{
			name: "Test_Remote_File_Not_Found_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					url := "TestAllocation_RepairRequired" + testCaseName + mockBlobberUrl + strconv.Itoa(i)
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						Baseurl: url,
					})
					mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return strings.HasPrefix(req.URL.Path, url)
					})).Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
					}, nil)
				}
				return nil
			},
			remotePath:           "/x.txt",
			wantFound:            0x0,
			wantMatchesConsensus: false,
			wantErr:              true,
			errMsg:               "File not found for the given remotepath",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
			}
			a.InitAllocation()
			sdkInitialized = true
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			found, matchesConsensus, fileRef, err := a.RepairRequired(tt.remotePath)
			require.Equal(zboxutil.NewUint128(tt.wantFound), found, "found value must be same")
			if tt.wantMatchesConsensus {
				require.True(tt.wantMatchesConsensus, matchesConsensus)
			} else {
				require.False(tt.wantMatchesConsensus, matchesConsensus)
			}
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			expected := &fileref.FileRef{
				ActualFileHash: mockActualHash,
			}
			require.EqualValues(expected, fileRef)
			require.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_DownloadFile(t *testing.T) {
	const (
		mockLocalPath = "DownloadFile"
	)
	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	require := require.New(t)
	a := &Allocation{
		ParityShards: 2,
		DataShards:   2,
	}
	setupMockAllocation(t, a)

	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: mockBlobberUrl + strconv.Itoa(i),
		})
	}
	err := a.DownloadFile(mockLocalPath, "/", nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_DownloadFileByBlock(t *testing.T) {
	const (
		mockLocalPath = "DownloadFileByBlock"
	)
	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	require := require.New(t)
	a := &Allocation{
		ParityShards: 2,
		DataShards:   2,
	}
	setupMockAllocation(t, a)

	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: mockBlobberUrl + strconv.Itoa(i),
		})
	}
	err := a.DownloadFileByBlock(mockLocalPath, "/", 1, 0, numBlockDownloads, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_DownloadThumbnail(t *testing.T) {
	const (
		mockLocalPath = "DownloadThumbnail"
	)
	require := require.New(t)
	a := &Allocation{}
	setupMockAllocation(t, a)

	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      strconv.Itoa(i),
			Baseurl: "TestAllocation_DownloadThumbnail" + mockBlobberUrl + strconv.Itoa(i),
		})
	}
	err := a.DownloadThumbnail(mockLocalPath, "/", nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_downloadFile(t *testing.T) {
	const (
		mockActualHash     = "mockActualHash"
		mockLocalPath      = "alloc"
		mockRemoteFilePath = "1.txt"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		localPath, remotePath string
		contentMode           string
		startBlock, endBlock  int64
		numBlocks             int
		statusCallback        StatusCallback
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, parameters, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			parameters: parameters{
				mockLocalPath, mockRemoteFilePath,
				DOWNLOAD_CONTENT_FULL,
				1, 0,
				numBlockDownloads,
				nil,
			},
			setup: func(t *testing.T, testCaseName string, p parameters, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) { a.initialized = true }
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Local_Path_Is_Not_Dir_Failed",
			parameters: parameters{
				localPath:      "Test_Local_Path_Is_Not_Dir_Failed",
				remotePath:     mockRemoteFilePath,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				statusCallback: nil,
			},
			setup: func(t *testing.T, testCaseName string, p parameters, a *Allocation) (teardown func(t *testing.T)) {
				os.Create(p.localPath)
				return func(t *testing.T) {
					os.Remove(p.localPath)
				}
			},
			wantErr: true,
			errMsg:  "Local path is not a directory 'Test_Local_Path_Is_Not_Dir_Failed'",
		},
		{
			name: "Test_Local_File_Already_Existed_Failed",
			parameters: parameters{
				localPath:      "alloc",
				remotePath:     mockRemoteFilePath,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				statusCallback: nil,
			},
			setup: func(t *testing.T, testCaseName string, p parameters, a *Allocation) (teardown func(t *testing.T)) {
				os.Mkdir(p.localPath, 0755)
				os.Create(p.localPath + "/" + p.remotePath)
				return func(t *testing.T) {
					os.RemoveAll(p.localPath + "/" + p.remotePath)
					os.RemoveAll(p.localPath)
				}
			},
			wantErr: true,
			errMsg:  "Local file already exists 'alloc/1.txt'",
		},
		{
			name: "Test_No_Blobber_Failed",
			parameters: parameters{
				localPath:      mockLocalPath,
				remotePath:     mockRemoteFilePath,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				statusCallback: nil,
			},
			setup: func(t *testing.T, testCaseName string, p parameters, a *Allocation) (teardown func(t *testing.T)) {
				blobbers := a.Blobbers
				a.Blobbers = []*blockchain.StorageNode{}
				return func(t *testing.T) {
					a.Blobbers = blobbers
				}
			},
			wantErr: true,
			errMsg:  "No Blobbers set in this allocation",
		},
		{
			name: "Test_Download_File_Success",
			parameters: parameters{
				localPath:      mockLocalPath,
				remotePath:     mockRemoteFilePath,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				statusCallback: nil,
			},
			setup: func(t *testing.T, testCaseName string, p parameters, a *Allocation) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					url := "TestAllocation_downloadFile" + mockBlobberUrl + strconv.Itoa(i)
					mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return strings.HasPrefix(req.URL.Path, url)
					})).Return(&http.Response{
						StatusCode: http.StatusOK,
						Body: func() io.ReadCloser {
							jsonFR, err := json.Marshal(&fileref.FileRef{
								ActualFileHash: mockActualHash,
							})
							require.NoError(t, err)
							return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
						}(),
					}, nil)
				}
				return func(t *testing.T) {
					os.Remove("alloc/1.txt")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{}
			a.uploadChan = make(chan *UploadRequest, 10)
			a.downloadChan = make(chan *DownloadRequest, 10)
			a.repairChan = make(chan *RepairRequest, 1)
			a.ctx, a.ctxCancelF = context.WithCancel(context.Background())
			a.uploadProgressMap = make(map[string]*UploadRequest)
			a.downloadProgressMap = make(map[string]*DownloadRequest)
			a.mutex = &sync.Mutex{}
			a.initialized = true
			sdkInitialized = true
			setupMockAllocation(t, a)
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + strconv.Itoa(i),
					Baseurl: "TestAllocation_downloadFile" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if m := tt.setup; m != nil {
				if teardown := m(t, tt.name, tt.parameters, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.downloadFile(tt.parameters.localPath, tt.parameters.remotePath, tt.parameters.contentMode, tt.parameters.startBlock, tt.parameters.endBlock, tt.parameters.numBlocks, tt.parameters.statusCallback)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "Unexpected error: %v", err)
		})
	}
}

func TestAllocation_DeleteFile(t *testing.T) {
	const (
		mockType = "f"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	require := require.New(t)

	a := &Allocation{
		DataShards:   2,
		ParityShards: 2,
	}
	a.InitAllocation()
	sdkInitialized = true

	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: "TestAllocation_DeleteFile" + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	body, err := json.Marshal(&fileref.ReferencePath{
		Meta: map[string]interface{}{
			"type": mockType,
		},
	})
	require.NoError(err)
	setupMockHttpResponse(t, &mockClient, "TestAllocation_DeleteFile", "", a, http.MethodGet, http.StatusOK, body)
	setupMockHttpResponse(t, &mockClient, "TestAllocation_DeleteFile", "", a, http.MethodDelete, http.StatusOK, []byte(""))
	setupMockCommitRequest(a)

	err = a.DeleteFile("/1.txt")
	require.NoErrorf(err, "unexpected error: %v", err)
}

func TestAllocation_deleteFile(t *testing.T) {
	const (
		mockType = "f"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		path string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name:    "Test_Invalid_Path_Failed",
			wantErr: true,
			errMsg:  "invalid_path: Invalid path for the list",
		},
		{
			name: "Test_Not_Abs_Path_Failed",
			parameters: parameters{
				path: "x.txt",
			},
			wantErr: true,
			errMsg:  "invalid_path: Path should be valid and absolute",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				path: "/1.txt",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_deleteFile", testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_deleteFile", testCaseName, a, http.MethodDelete, http.StatusOK, []byte(""))
				setupMockCommitRequest(a)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
			}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "TestAllocation_deleteFile" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.DeleteFile(tt.parameters.path)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_UpdateObjectAttributes(t *testing.T) {
	const (
		mockType = "f"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		path       string
		attrs      fileref.Attributes
		statusCode int
	}

	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, parameters, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, p parameters, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Invalid_Path_Failed",
			parameters: parameters{
				path:  "",
				attrs: fileref.Attributes{WhoPaysForReads: common.WhoPaysOwner},
			},
			wantErr: true,
			errMsg:  "update_attrs: Invalid path for the list",
		},
		{
			name: "Test_Invalid_Remote_Abs_Path_Failed",
			parameters: parameters{
				path:  "abc",
				attrs: fileref.Attributes{WhoPaysForReads: common.WhoPaysOwner},
			},
			wantErr: true,
			errMsg:  "update_attrs: Path should be valid and absolute",
		},
		{
			name: "Test_Update_Attributes_Failed",
			parameters: parameters{
				path:       "/1.txt",
				attrs:      fileref.Attributes{WhoPaysForReads: common.WhoPaysOwner},
				statusCode: 400,
			},
			setup: func(t *testing.T, testName string, p parameters, a *Allocation) (teardown func(*testing.T)) {
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_UpdateObjectAttributes", testName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_UpdateObjectAttributes", testName, a, http.MethodPost, p.statusCode, []byte(""))
				setupMockCommitRequest(a)
				return nil
			},
			wantErr: true,
			errMsg:  "Update attributes failed: request failed, operation failed",
		},
		{
			name: "Test_Who_Pay_For_Read_Owner_Success",
			parameters: parameters{
				path:       "/1.txt",
				attrs:      fileref.Attributes{WhoPaysForReads: common.WhoPaysOwner},
				statusCode: 200,
			},
			setup: func(t *testing.T, testName string, p parameters, a *Allocation) (teardown func(*testing.T)) {
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_UpdateObjectAttributes", testName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_UpdateObjectAttributes", testName, a, http.MethodPost, p.statusCode, []byte(""))
				setupMockCommitRequest(a)
				return nil
			},
		},
		{
			name: "Test_Who_Pay_For_Read_3rd_Party_Success",
			parameters: parameters{
				path:       "/1.txt",
				attrs:      fileref.Attributes{WhoPaysForReads: common.WhoPays3rdParty},
				statusCode: 200,
			},
			setup: func(t *testing.T, testName string, p parameters, a *Allocation) (teardown func(*testing.T)) {
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_UpdateObjectAttributes", testName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_UpdateObjectAttributes", testName, a, http.MethodPost, p.statusCode, []byte(""))
				setupMockCommitRequest(a)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
			}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "TestAllocation_UpdateObjectAttributes" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if setup := tt.setup; setup != nil {
				if teardown := setup(t, tt.name, tt.parameters, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.UpdateObjectAttributes(tt.parameters.path, tt.parameters.attrs)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_MoveObject(t *testing.T) {
	const (
		mockType = "f"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		path     string
		destPath string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Cover_Copy_Object",
			parameters: parameters{
				path:     "/1.txt",
				destPath: "/d",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Cover_Delete_Object",
			parameters: parameters{
				path:     "/1.txt",
				destPath: "/d",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodPost, http.StatusOK, []byte(""))
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodDelete, http.StatusOK, []byte(""))
				setupMockCommitRequest(a)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
			}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "TestAllocation_MoveObject" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.MoveObject(tt.parameters.path, tt.parameters.destPath)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_CopyObject(t *testing.T) {
	const (
		mockType = "f"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		path     string
		destPath string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Wrong_Path_Or_Destination_Path_Failed",
			parameters: parameters{
				path:     "",
				destPath: "",
			},
			wantErr: true,
			errMsg:  "invalid_path: Invalid path for copy",
		},
		{
			name: "Test_Invalid_Remote_Absolute_Path",
			parameters: parameters{
				path:     "abc",
				destPath: "/d",
			},
			wantErr: true,
			errMsg:  "invalid_path: Path should be valid and absolute",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				path:     "/1.txt",
				destPath: "/d",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_CopyObject", testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_CopyObject", testCaseName, a, http.MethodPost, http.StatusOK, []byte(""))
				setupMockCommitRequest(a)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
			}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "TestAllocation_CopyObject" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.CopyObject(tt.parameters.path, tt.parameters.destPath)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_RenameObject(t *testing.T) {
	const (
		mockType = "f"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		path     string
		destName string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Wrong_Path_Or_Destination_Path_Failed",
			parameters: parameters{
				path:     "",
				destName: "",
			},
			wantErr: true,
			errMsg:  "invalid_path: Invalid path for the list",
		},
		{
			name: "Test_Invalid_Remote_Absolute_Path",
			parameters: parameters{
				path:     "abc",
				destName: "/2.txt",
			},
			wantErr: true,
			errMsg:  "invalid_path: Path should be valid and absolute",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				path:     "/1.txt",
				destName: "/2.txt",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_RenameObject", testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_RenameObject", testCaseName, a, http.MethodPost, http.StatusOK, []byte(""))
				setupMockCommitRequest(a)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
			}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "TestAllocation_RenameObject" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.RenameObject(tt.parameters.path, tt.parameters.destName)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_AddCollaborator(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		filePath       string
		collaboratorID string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Add_Collaborator_Error_Response_Failed",
			parameters: parameters{
				filePath:       "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				setupMockHttpResponse(t, &mockClient, "TestAllocation_AddCollaborator", testCaseName, a, http.MethodPost, http.StatusBadRequest, []byte(""))
				return nil
			},
			wantErr: true,
			errMsg:  "add_collaborator_failed: Failed to add collaborator on all blobbers.",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				filePath:       "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				setupMockHttpResponse(t, &mockClient, "TestAllocation_AddCollaborator", testCaseName, a, http.MethodPost, http.StatusOK, []byte(""))
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
			}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "TestAllocation_AddCollaborator" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.AddCollaborator(tt.parameters.filePath, tt.parameters.collaboratorID)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_RemoveCollaborator(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		filePath       string
		collaboratorID string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Remove_Collaborator_Error_Response_Failed",
			parameters: parameters{
				filePath:       "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				setupMockHttpResponse(t, &mockClient, "TestAllocation_RemoveCollaborator", testCaseName, a, http.MethodDelete, http.StatusBadRequest, []byte(""))
				return nil
			},
			wantErr: true,
			errMsg:  "remove_collaborator_failed: Failed to remove collaborator on all blobbers.",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				filePath:       "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				setupMockHttpResponse(t, &mockClient, "TestAllocation_RemoveCollaborator", testCaseName, a, http.MethodDelete, http.StatusOK, []byte(""))
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
			}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "TestAllocation_RemoveCollaborator" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.RemoveCollaborator(tt.parameters.filePath, tt.parameters.collaboratorID)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_GetFileMeta(t *testing.T) {
	const (
		mockType       = "f"
		mockActualHash = "mockActualHash"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		path string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Test_Uninitialized_Failed",
			parameters: parameters{},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Error_Getting_File_Meta_Data_From_Blobbers_Failed",
			parameters: parameters{
				path: "/1.txt",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.FileRef{
					ActualFileHash: mockActualHash,
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_GetFileMeta", testCaseName, a, http.MethodPost, http.StatusBadRequest, body)
				return nil
			},
			wantErr: true,
			errMsg:  "file_meta_error: Error getting the file meta data from blobbers",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				path: "/1.txt",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.FileRef{
					ActualFileHash: mockActualHash,
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_GetFileMeta", testCaseName, a, http.MethodPost, http.StatusOK, body)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
			}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "TestAllocation_GetFileMeta" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			got, err := a.GetFileMeta(tt.parameters.path)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			expectedResult := &ConsolidatedFileMeta{
				Hash: mockActualHash,
			}
			require.EqualValues(expectedResult, got)
		})
	}
}

func TestAllocation_GetAuthTicketForShare(t *testing.T) {
	const mockContentHash = "mock content hash"
	const numberBlobbers = 10

	var mockClient = mocks.HttpClient{}
	httpResponse := http.Response{
		StatusCode: http.StatusOK,
		Body: func() io.ReadCloser {
			jsonFR, err := json.Marshal(fileref.FileRef{
				Ref: fileref.Ref{
					Name: mockFileRefName,
				},
				ContentHash: mockContentHash,
			})
			require.NoError(t, err)
			return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
		}(),
	}
	zboxutil.Client = &mockClient
	for i := 0; i < numBlobbers; i++ {
		mockClient.On("Do", mock.Anything).Return(&httpResponse, nil)
	}

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}
	require := require.New(t)
	a := &Allocation{DataShards: 1, ParityShards: 1}
	a.InitAllocation()
	for i := 0; i < numberBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{})
	}
	sdkInitialized = true
	at, err := a.GetAuthTicketForShare("/1.txt", "1.txt", fileref.FILE, mockClientId)
	require.NotEmptyf(at, "unexpected empty auth ticket")
	require.NoErrorf(err, "unexpected error: %v", err)
}

func TestAllocation_GetAuthTicket(t *testing.T) {
	var testTitle = "TestAllocation_GetAuthTicket"
	type parameters struct {
		path                       string
		filename                   string
		referenceType              string
		refereeClientID            string
		refereeEncryptionPublicKey string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation, *mocks.HttpClient) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation, mockClient *mocks.HttpClient) (teardown func(t *testing.T)) {
				a.initialized = false
				httpResponse := http.Response{
					StatusCode: http.StatusOK,
					Body: func() io.ReadCloser {
						jsonFR, err := json.Marshal(fileref.FileRef{
							Ref: fileref.Ref{
								Name: mockFileRefName,
							},
							ContentHash: "mock content hash",
						})
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
				}
				for i := 0; i < numBlobbers; i++ {
					mockClient.On("Do", mock.Anything).Return(&httpResponse, nil)
				}
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Success_File_Type_Directory",
			parameters: parameters{
				path:            "/",
				filename:        "1.txt",
				referenceType:   fileref.DIRECTORY,
				refereeClientID: mockClientId,
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation, mockClient *mocks.HttpClient) (teardown func(t *testing.T)) {
				httpResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body: func() io.ReadCloser {
						jsonFR, err := json.Marshal(fileref.FileRef{
							Ref: fileref.Ref{
								Name: mockFileRefName,
							},
							ContentHash: "mock content hash",
						})
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
				}
				for i := 0; i < numBlobbers; i++ {
					mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return strings.HasPrefix(req.URL.Path, testTitle+testCaseName)
					})).Return(httpResponse, nil)
				}
				return nil
			},
		},
		{
			name: "Test_Success_With_Referee_Encryption_Public_Key",
			parameters: parameters{
				path:            "/1.txt",
				filename:        "1.txt",
				referenceType:   fileref.FILE,
				refereeClientID: mockClientId,
				refereeEncryptionPublicKey: func() string {
					client_mnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
					client_encscheme := encryption.NewEncryptionScheme()
					client_encscheme.Initialize(client_mnemonic)
					client_encscheme.InitForEncryption("filetype:audio")
					client_enc_pub_key, err := client_encscheme.GetPublicKey()
					require.NoError(t, err)
					return client_enc_pub_key
				}(),
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation, mockClient *mocks.HttpClient) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": "f",
					},
				})
				httpResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body: func() io.ReadCloser {
						jsonFR, err := json.Marshal(fileref.FileRef{
							Ref: fileref.Ref{
								Name: mockFileRefName,
							},
							ContentHash: "mock content hash",
						})
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
				}
				for i := 0; i < numBlobbers; i++ {
					mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return strings.HasPrefix(req.URL.Path, testTitle+testCaseName)
					})).Return(httpResponse, nil)
				}
				require.NoError(t, err)
				setupMockHttpResponse(t, mockClient, "TestAllocation_GetAuthTicket", testCaseName, a, http.MethodPost, http.StatusOK, body)
				return nil
			},
		},
		{
			name: "Test_Success_With_No_Referee_Encryption_Public_Key",
			parameters: parameters{
				path:                       "/1.txt",
				filename:                   "1.txt",
				referenceType:              fileref.FILE,
				refereeClientID:            mockClientId,
				refereeEncryptionPublicKey: "",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation, mockClient *mocks.HttpClient) (teardown func(t *testing.T)) {
				httpResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body: func() io.ReadCloser {
						jsonFR, err := json.Marshal(fileref.FileRef{
							Ref: fileref.Ref{
								Name: mockFileRefName,
							},
							ContentHash: "mock content hash",
						})
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
				}
				for i := 0; i < numBlobbers; i++ {
					mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return strings.HasPrefix(req.URL.Path, testTitle+testCaseName)
					})).Return(httpResponse, nil)
				}
				return nil
			},
		},
		{
			name: "Test_Invalid_Path_Failed",
			parameters: parameters{
				filename: "1.txt",
			},
			wantErr: true,
			errMsg:  "invalid_path: Invalid path for the list",
		},
		{
			name: "Test_Remote_Path_Not_Absolute_Failed",
			parameters: parameters{
				path: "x",
			},
			wantErr: true,
			errMsg:  "invalid_path: Path should be valid and absolute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockClient = mocks.HttpClient{}
			zboxutil.Client = &mockClient

			client := zclient.GetClient()
			client.Wallet = &zcncrypto.Wallet{
				ClientID:  mockClientId,
				ClientKey: mockClientKey,
			}

			require := require.New(t)
			a := &Allocation{}
			a.InitAllocation()
			sdkInitialized = true

			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "TestAllocation_GetAuthTicket" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			const mockContentHash = "mock content hash"
			const numberBlobbers = 10

			zboxutil.Client = &mockClient

			a.InitAllocation()
			a.DataShards = 1
			a.ParityShards = 1
			sdkInitialized = true

			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a, &mockClient); teardown != nil {
					defer teardown(t)
				}
			}
			at, err := a.GetAuthTicket(tt.parameters.path, tt.parameters.filename, tt.parameters.referenceType, tt.parameters.refereeClientID, tt.parameters.refereeEncryptionPublicKey, 0)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			require.NotEmptyf(at, "unexpected empty auth ticket")
		})
	}
}

func TestAllocation_CancelUpload(t *testing.T) {
	const localPath = "alloc"
	type parameters struct {
		localpath string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name:    "Test_Failed",
			wantErr: true,
			errMsg:  "local_path_not_found: Invalid path. No upload in progress for the path",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				localpath: localPath,
			},
			setup: func(t *testing.T, a *Allocation) (teardown func(t *testing.T)) {
				a.uploadProgressMap[localPath] = &UploadRequest{}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{}
			a.InitAllocation()
			sdkInitialized = true
			if tt.setup != nil {
				if teardown := tt.setup(t, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.CancelUpload(tt.parameters.localpath)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_CancelDownload(t *testing.T) {
	const remotePath = "/1.txt"
	type parameters struct {
		remotepath string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name:    "Test_Failed",
			wantErr: true,
			errMsg:  "local_path_not_found: Invalid path. No upload in progress for the path",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				remotepath: remotePath,
			},
			setup: func(t *testing.T, a *Allocation) (teardown func(t *testing.T)) {
				a.downloadProgressMap[remotePath] = &DownloadRequest{}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{}
			a.InitAllocation()
			sdkInitialized = true
			if tt.setup != nil {
				if teardown := tt.setup(t, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.CancelDownload(tt.parameters.remotepath)
			if tt.wantErr {
				require.Error(err, "expected error != nil")
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_CommitFolderChange(t *testing.T) {
	const (
		mockHash      = "mock hash"
		mockSignature = "mock signature"
	)

	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	resty.CreateClient = func(t *http.Transport, timeout time.Duration) resty.Client {
		return &mockClient
	}

	conf.InitClientConfig(&conf.Config{
		MinConfirmation: 50,
	})

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponse := func(t *testing.T, name string, httpMethod string, statusCode int, body []byte) {
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == httpMethod &&
				strings.Index(req.URL.String(), name) > -1
		})).Return(&http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(body))),
		}, nil)
	}

	type parameters struct {
		operation, preValue, currValue string
	}
	blockchain.SetQuerySleepTime(1)

	tests := []struct {
		name          string
		parameters    parameters
		setup         func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		wantErr       bool
		errMsg        string
		exceptedError error
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false

				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Sharder_Verify_Txn_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {

				body, err := json.Marshal(&transaction.Transaction{
					Hash: mockHash,
				})
				require.NoError(t, err)
				setupHttpResponse(t, testCaseName+"mockMiners", http.MethodPost, http.StatusOK, body)
				setupHttpResponse(t, testCaseName+"mockSharders", http.MethodGet, http.StatusBadRequest, []byte(""))
				return nil
			},
			wantErr:       true,
			exceptedError: transaction.ErrTooLessConfirmation,
			errMsg:        "transaction_not_found: Transaction was not found on any of the sharders",
		},
		{
			name: "Test_Max_Retried_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				setupHttpResponse(t, testCaseName+"mockMiners", http.MethodPost, http.StatusBadRequest, []byte(""))
				maxTxnQuery := blockchain.GetMaxTxnQuery()
				blockchain.SetMaxTxnQuery(0)
				return func(t *testing.T) {
					blockchain.SetMaxTxnQuery(maxTxnQuery)
				}
			},
			wantErr: true,
			errMsg:  "transaction_validation_failed: Failed to get the transaction confirmation",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				operation: "Move",
				preValue:  "/1.txt",
				currValue: "/d/1.txt",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&transaction.Transaction{
					Hash: mockHash,
				})
				require.NoError(t, err)
				setupHttpResponse(t, testCaseName+"mockMiners", http.MethodPost, http.StatusOK, body)
				setupHttpResponse(t, testCaseName+"mockSharders", http.MethodGet, http.StatusOK, []byte(`{
					"txn": {
						"hash": "`+mockHash+`",
						"signature": "`+mockSignature+`"
					}
				}`))
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				ID:           mockAllocationId,
				Tx:           mockAllocationTxId,
				DataShards:   2,
				ParityShards: 2,
			}
			a.InitAllocation()
			sdkInitialized = true
			blockchain.SetMiners([]string{"http://" + tt.name + "mockMiners"})
			blockchain.SetSharders([]string{"http://" + tt.name + "mockSharders"})
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}

			_, err := a.CommitFolderChange(tt.parameters.operation, tt.parameters.preValue, tt.parameters.currValue)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {

				// test it by predefined error variable instead of error message
				if tt.exceptedError != nil {
					require.ErrorIs(err, tt.exceptedError)
				} else {
					require.EqualValues(tt.errMsg, errors.Top(err))
				}

				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			// require.Equal(string(expectedBytes), got)
		})
	}
}

func TestAllocation_ListDirFromAuthTicket(t *testing.T) {
	const (
		mockLookupHash = "mock lookup hash"
		mockType       = "d"
	)

	authTicket := getMockAuthTicket(t)

	type parameters struct {
		authTicket     string
		lookupHash     string
		expectedResult *ListResult
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation, *mocks.HttpClient) (teardown func(t *testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation, mockClient *mocks.HttpClient) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "invalid_path: Invalid path for the list",
		},
		{
			name: "Test_Cannot_Decode_Auth_Ticket_Failed",
			parameters: parameters{
				authTicket: "some wrong auth ticket to decode",
			},
			setup:   nil,
			wantErr: true,
			errMsg: "invalid_path: Invalid path for the list" +
				"",
		},
		{
			name: "Test_Cannot_Unmarshal_Auth_Ticket_Failed",
			parameters: parameters{
				authTicket: "c29tZSB3cm9uZyBhdXRoIHRpY2tldCB0byBtYXJzaGFs",
			},
			setup:   nil,
			wantErr: true,
			errMsg:  "invalid_path: Invalid path for the list",
		},
		{
			name: "Test_Wrong_Auth_Ticket_File_Path_Hash_Or_Lookup_Hash_Failed",
			parameters: parameters{
				authTicket: authTicket,
				lookupHash: "",
			},
			setup:   nil,
			wantErr: true,
			errMsg:  "invalid_path: Invalid path for the list",
		},
		{
			name: "Test_Error_Get_List_File_From_Blobbers_Failed",
			parameters: parameters{
				authTicket:     authTicket,
				lookupHash:     mockLookupHash,
				expectedResult: &ListResult{},
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation, mockClient *mocks.HttpClient) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						Baseurl: "TestAllocation_ListDirFromAuthTicket" + testCaseName + mockBlobberUrl + strconv.Itoa(i),
					})
				}
				setupMockHttpResponse(t, mockClient, "TestAllocation_ListDirFromAuthTicket", testCaseName, a, http.MethodGet, http.StatusBadRequest, []byte(""))
				return func(t *testing.T) {
					a.Blobbers = nil
				}
			},
		},
		{
			name: "Test_Success",
			parameters: parameters{
				authTicket: authTicket,
				lookupHash: mockLookupHash,
				expectedResult: &ListResult{
					Type: mockType,
					Size: -1,
				},
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation, mockClient *mocks.HttpClient) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						ID:      testCaseName + mockBlobberId + strconv.Itoa(i),
						Baseurl: "TestAllocation_ListDirFromAuthTicket" + testCaseName + mockBlobberUrl + strconv.Itoa(i),
					})
				}
				body, err := json.Marshal(&fileref.ListResult{
					AllocationRoot: mockAllocationRoot,
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, mockClient, "TestAllocation_ListDirFromAuthTicket", testCaseName, a, http.MethodGet, http.StatusOK, body)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			var mockClient = mocks.HttpClient{}
			zboxutil.Client = &mockClient

			client := zclient.GetClient()
			client.Wallet = &zcncrypto.Wallet{
				ClientID:  mockClientId,
				ClientKey: mockClientKey,
			}
			a := &Allocation{
				ID: mockAllocationId,
				Tx: mockAllocationTxId,
			}

			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a, &mockClient); teardown != nil {
					defer teardown(t)
				}
			}

			setupMockGetFileInfoResponse(t, &mockClient)
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{})
			}

			got, err := a.ListDirFromAuthTicket(authTicket, tt.parameters.lookupHash)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			require.EqualValues(tt.parameters.expectedResult, got)
		})
	}
}

func TestAllocation_downloadFromAuthTicket(t *testing.T) {
	const (
		mockLookupHash     = "mock lookup hash"
		mockLocalPath      = "alloc"
		mockRemoteFileName = "1.txt"
		mockType           = "f"
		mockActualHash     = "mockActualHash"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	a := &Allocation{
		ID:           mockAllocationId,
		Tx:           mockAllocationTxId,
		DataShards:   2,
		ParityShards: 2,
	}
	setupMockAllocation(t, a)
	setupMockGetFileInfoResponse(t, &mockClient)
	authTicket := getMockAuthTicket(t)

	type parameters struct {
		localPath      string
		authTicket     string
		lookupHash     string
		startBlock     int64
		endBlock       int64
		numBlocks      int
		remoteFilename string
		contentMode    string
		rxPay          bool
		statusCallback StatusCallback
	}
	tests := []struct {
		name                      string
		parameters                parameters
		setup                     func(*testing.T, string, parameters) (teardown func(*testing.T))
		blockDownloadResponseMock func(blobber *blockchain.StorageNode, wg *sync.WaitGroup)
		wantErr                   bool
		errMsg                    string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, p parameters) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Cannot_Decode_Auth_Ticket_Failed",
			parameters: parameters{
				authTicket: "some wrong auth ticket to decode",
			},
			wantErr: true,
			errMsg:  "auth_ticket_decode_error: Error decoding the auth ticket.illegal base64 data at input byte 4",
		},
		{
			name: "Test_Cannot_Unmarshal_Auth_Ticket_Failed",
			parameters: parameters{
				authTicket: "c29tZSB3cm9uZyBhdXRoIHRpY2tldCB0byBtYXJzaGFs",
			},
			wantErr: true,
			errMsg:  "auth_ticket_decode_error: Error unmarshaling the auth ticket.invalid character 's' looking for beginning of value",
		},
		{
			name: "Test_Local_Path_Is_Not_Directory_Failed",
			parameters: parameters{
				localPath:  "Test_Local_Path_Is_Not_Directory_Failed",
				authTicket: authTicket,
			},
			setup: func(t *testing.T, testCaseName string, p parameters) (teardown func(t *testing.T)) {
				os.Create(p.localPath)
				return func(t *testing.T) {
					os.Remove(p.localPath)
				}
			},
			wantErr: true,
			errMsg:  "Local path is not a directory 'Test_Local_Path_Is_Not_Directory_Failed'",
		},
		{
			name: "Test_Local_File_Already_Exists_Failed",
			parameters: parameters{
				localPath:      mockLocalPath,
				authTicket:     authTicket,
				remoteFilename: mockRemoteFileName,
			},
			setup: func(t *testing.T, testCaseName string, p parameters) (teardown func(t *testing.T)) {
				os.Mkdir(p.localPath, 0755)
				os.Create(p.localPath + "/" + p.remoteFilename)
				return func(t *testing.T) {
					os.RemoveAll(p.localPath + "/" + p.remoteFilename)
					os.RemoveAll(p.localPath)
				}
			},
			wantErr: true,
			errMsg:  "Local file already exists 'alloc/1.txt'",
		},
		{
			name: "Test_Not_Enough_Minimum_Blobbers_Failed",
			parameters: parameters{
				localPath:      mockLocalPath,
				authTicket:     authTicket,
				remoteFilename: mockRemoteFileName,
			},
			setup: func(t *testing.T, testCaseName string, p parameters) (teardown func(t *testing.T)) {
				blobbers := a.Blobbers
				a.Blobbers = []*blockchain.StorageNode{}
				return func(t *testing.T) {
					a.Blobbers = blobbers
				}
			},
			wantErr: true,
			errMsg:  "No Blobbers set in this allocation",
		},
		{
			name: "Test_Download_File_Success",
			parameters: parameters{
				localPath:      mockLocalPath,
				remoteFilename: mockRemoteFileName,
				authTicket:     authTicket,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				lookupHash:     mockLookupHash},
			setup: func(t *testing.T, testCaseName string, p parameters) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						ID:      testCaseName + mockBlobberId + strconv.Itoa(i),
						Baseurl: "TestAllocation_downloadFromAuthTicket" + testCaseName + mockBlobberUrl + strconv.Itoa(i),
					})
				}
				body, err := json.Marshal(&fileref.FileRef{
					Ref: fileref.Ref{
						Name: mockFileRefName,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_downloadFromAuthTicket", testCaseName, a, http.MethodPost, http.StatusOK, body)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, tt.parameters); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.downloadFromAuthTicket(tt.parameters.localPath, tt.parameters.authTicket, tt.parameters.lookupHash, tt.parameters.startBlock, tt.parameters.endBlock, tt.parameters.numBlocks, tt.parameters.remoteFilename, tt.parameters.contentMode, tt.parameters.rxPay, tt.parameters.statusCallback)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_listDir(t *testing.T) {
	const (
		mockPath = "/1.txt"
		mockType = "d"
	)

	type parameters struct {
		path                           string
		consensusThresh, fullConsensus float32
		expectedResult                 *ListResult
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation, *mocks.HttpClient) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation, mockClient *mocks.HttpClient) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name:    "Test_Invalid_Path_Failed",
			wantErr: true,
			errMsg:  "invalid_path: Invalid path for the list",
		},
		{
			name: "Test_Invalid_Absolute_Path_Failed",
			parameters: parameters{
				path: "1.txt",
			},
			wantErr: true,
			errMsg:  "invalid_path: Path should be valid and absolute",
		},
		{
			name: "Test_Error_Get_List_File_From_Blobbers_Failed",
			parameters: parameters{
				path:            mockPath,
				consensusThresh: 50,
				fullConsensus:   4,
				expectedResult:  &ListResult{},
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation, mockClient *mocks.HttpClient) (teardown func(t *testing.T)) {
				setupMockHttpResponse(t, mockClient, "TestAllocation_listDir", testCaseName, a, http.MethodGet, http.StatusBadRequest, []byte(""))
				return nil
			},
		},
		{
			name: "Test_Success",
			parameters: parameters{
				path:            mockPath,
				consensusThresh: 50,
				fullConsensus:   4,
				expectedResult: &ListResult{
					Type: mockType,
					Size: -1,
				},
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation, mockClient *mocks.HttpClient) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.ListResult{
					AllocationRoot: mockAllocationRoot,
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, mockClient, "TestAllocation_listDir", testCaseName, a, http.MethodGet, http.StatusOK, body)
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockClient = mocks.HttpClient{}
			zboxutil.Client = &mockClient

			client := zclient.GetClient()
			client.Wallet = &zcncrypto.Wallet{
				ClientID:  mockClientId,
				ClientKey: mockClientKey,
			}

			require := require.New(t)
			a := &Allocation{
				ID: mockAllocationId,
				Tx: mockAllocationTxId,
			}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "TestAllocation_listDir" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a, &mockClient); teardown != nil {
					defer teardown(t)
				}
			}
			got, err := a.listDir(tt.parameters.path, tt.parameters.consensusThresh, tt.parameters.fullConsensus)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			require.EqualValues(tt.parameters.expectedResult, got)
		})
	}
}

func TestAllocation_GetFileMetaFromAuthTicket(t *testing.T) {
	const (
		mockLookupHash = "mock lookup hash"
		mockActualHash = "mockActualHash"
		mockType       = "f"
	)

	var authTicket = getMockAuthTicket(t)

	type parameters struct {
		authTicket, lookupHash string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation, *mocks.HttpClient) (teardown func(t *testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation, _ *mocks.HttpClient) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Cannot_Decode_Auth_Ticket_Failed",
			parameters: parameters{
				authTicket: "some wrong auth ticket to decode",
			},
			wantErr: true,
			errMsg:  "auth_ticket_decode_error: Error decoding the auth ticket.illegal base64 data at input byte 4",
		},
		{
			name: "Test_Cannot_Unmarshal_Auth_Ticket_Failed",
			parameters: parameters{
				authTicket: "c29tZSB3cm9uZyBhdXRoIHRpY2tldCB0byBtYXJzaGFs",
			},
			wantErr: true,
			errMsg:  "auth_ticket_decode_error: Error unmarshaling the auth ticket.invalid character 's' looking for beginning of value",
		},
		{
			name: "Test_Wrong_Auth_Ticket_File_Path_Hash_Or_Lookup_Hash_Failed",
			parameters: parameters{
				authTicket: authTicket,
				lookupHash: "",
			},
			wantErr: true,
			errMsg:  "invalid_path: Invalid path for the list",
		},
		{
			name: "Test_Error_Get_File_Meta_From_Blobbers_Failed",
			parameters: parameters{
				authTicket: authTicket,
				lookupHash: mockLookupHash,
			},
			wantErr: true,
			errMsg:  "file_meta_error: Error getting the file meta data from blobbers",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				authTicket: authTicket,
				lookupHash: mockLookupHash,
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation, mockClient *mocks.HttpClient) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						ID:      testCaseName + mockBlobberId + strconv.Itoa(i),
						Baseurl: "TestAllocation_GetFileMetaFromAuthTicket" + testCaseName + mockBlobberUrl + strconv.Itoa(i),
					})
				}
				body, err := json.Marshal(&fileref.FileRef{
					ActualFileHash: mockActualHash,
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, mockClient, "TestAllocation_GetFileMetaFromAuthTicket", testCaseName, a, http.MethodPost, http.StatusOK, body)
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockClient = mocks.HttpClient{}
			zboxutil.Client = &mockClient

			client := zclient.GetClient()
			client.Wallet = &zcncrypto.Wallet{
				ClientID:  mockClientId,
				ClientKey: mockClientKey,
			}

			a := &Allocation{
				ID:           mockAllocationId,
				Tx:           mockAllocationTxId,
				DataShards:   2,
				ParityShards: 2,
			}
			a.InitAllocation()
			sdkInitialized = true
			a.initialized = true

			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a, &mockClient); teardown != nil {
					defer teardown(t)
				}
			}
			got, err := a.GetFileMetaFromAuthTicket(tt.parameters.authTicket, tt.parameters.lookupHash)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			expectedResult := &ConsolidatedFileMeta{
				Hash: mockActualHash,
			}
			require.EqualValues(expectedResult, got)
		})
	}
}

func TestAllocation_DownloadThumbnailFromAuthTicket(t *testing.T) {
	const (
		mockLookupHash     = "mock lookup hash"
		mockLocalPath      = "alloc"
		mockRemoteFilePath = "1.txt"
		mockType           = "d"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	require := require.New(t)

	a := &Allocation{}
	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      strconv.Itoa(i),
			Baseurl: "TestAllocation_DownloadThumbnailFromAuthTicket" + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	var authTicket = getMockAuthTicket(t)

	body, err := json.Marshal(&fileref.ReferencePath{
		Meta: map[string]interface{}{
			"type": mockType,
		},
	})
	require.NoError(err)
	setupMockHttpResponse(t, &mockClient, "TestAllocation_DownloadThumbnailFromAuthTicket", "", a, http.MethodGet, http.StatusOK, body)

	err = a.DownloadThumbnailFromAuthTicket(mockLocalPath, authTicket, mockLookupHash, mockRemoteFilePath, true, nil)
	defer os.Remove("alloc/1.txt")
	require.NoErrorf(err, "unexpected error: %v", err)
}

func TestAllocation_DownloadFromAuthTicket(t *testing.T) {
	const (
		mockLookupHash     = "mock lookup hash"
		mockLocalPath      = "alloc"
		mockRemoteFilePath = "1.txt"
		mockType           = "d"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	require := require.New(t)

	a := &Allocation{}
	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      strconv.Itoa(i),
			Baseurl: "TestAllocation_DownloadFromAuthTicket" + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	var authTicket = getMockAuthTicket(t)

	err := a.DownloadFromAuthTicket(mockLocalPath, authTicket, mockLookupHash, mockRemoteFilePath, true, nil)
	defer os.Remove("alloc/1.txt")
	require.NoErrorf(err, "unexpected error: %v", err)
}

func TestAllocation_DownloadFromAuthTicketByBlocks(t *testing.T) {
	const (
		mockLookupHash     = "mock lookup hash"
		mockLocalPath      = "alloc"
		mockRemoteFilePath = "1.txt"
		mockType           = "d"
	)

	var authTicket = getMockAuthTicket(t)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	require := require.New(t)

	a := &Allocation{}
	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      strconv.Itoa(i),
			Baseurl: "TestAllocation_DownloadFromAuthTicketByBlocks" + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	setupMockHttpResponse(t, &mockClient, "TestAllocation_DownloadFromAuthTicketByBlocks", "", a, http.MethodPost, http.StatusBadRequest, []byte(""))

	err := a.DownloadFromAuthTicketByBlocks(mockLocalPath, authTicket, 1, 0, numBlockDownloads, mockLookupHash, mockRemoteFilePath, true, nil)
	defer os.Remove("alloc/1.txt")
	require.NoErrorf(err, "unexpected error: %v", err)
}

func TestAllocation_CommitMetaTransaction(t *testing.T) {
	const (
		mockLookupHash = "mock lookup hash"
		mockType       = "d"
	)

	var authTicket = getMockAuthTicket(t)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	a := &Allocation{}
	a.InitAllocation()
	sdkInitialized = true

	type parameters struct {
		path          string
		crudOperation string
		authTicket    string
		lookupHash    string
		fileMeta      func(t *testing.T, testCaseName string) *ConsolidatedFileMeta
		status        func(t *testing.T) StatusCallback
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_No_File_Meta_With_Path_parameters_Failed",
			parameters: parameters{
				path:          "/1.txt",
				crudOperation: "",
				authTicket:    "",
				lookupHash:    mockLookupHash,
				fileMeta:      nil,
			},
			wantErr: true,
			errMsg:  "file_meta_error: Error getting the file meta data from blobbers",
		},
		{
			name: "Test_No_File_Meta_With_Auth_Ticket_parameters_Failed",
			parameters: parameters{
				path:          "",
				crudOperation: "",
				authTicket:    authTicket,
				lookupHash:    mockLookupHash,
				fileMeta:      nil,
			},
			wantErr: true,
			errMsg:  "file_meta_error: Error getting the file meta data from blobbers",
		},
		{
			name: "Test_No_File_Meta_With_No_Path_And_No_Auth_Ticket_parameters_Coverage",
			parameters: parameters{
				path:          "",
				crudOperation: "",
				authTicket:    "",
				lookupHash:    mockLookupHash,
				fileMeta:      nil,
				status: func(t *testing.T) StatusCallback {
					scm := &mocks.StatusCallback{}
					scm.On("CommitMetaCompleted", mock.Anything, mock.Anything, mock.Anything).Maybe()
					return scm
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			var fileMeta *ConsolidatedFileMeta
			if tt.parameters.fileMeta != nil {
				fileMeta = tt.parameters.fileMeta(t, tt.name)
			}
			var status StatusCallback
			if tt.parameters.status != nil {
				status = tt.parameters.status(t)
			}
			err := a.CommitMetaTransaction(tt.parameters.path, tt.parameters.crudOperation, tt.parameters.authTicket, tt.parameters.lookupHash, fileMeta, status)
			if st, ok := status.(*mocks.StatusCallback); ok {
				st.Test(t)
				st.AssertExpectations(t)
			}
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_StartRepair(t *testing.T) {
	const (
		mockLookupHash   = "mock lookup hash"
		mockLocalPath    = "alloc"
		mockPathToRepair = "/1.txt"
		mockType         = "d"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		localPath, pathToRepair string
		statusCallback          StatusCallback
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Repair_Success",
			parameters: parameters{
				localPath:    mockLocalPath,
				pathToRepair: "/1.txt",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.ListResult{
					AllocationRoot: mockAllocationRoot,
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_StartRepair", testCaseName, a, http.MethodGet, http.StatusOK, body)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
			}
			setupMockAllocation(t, a)
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "TestAllocation_StartRepair" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.StartRepair(tt.parameters.localPath, tt.parameters.pathToRepair, tt.parameters.statusCallback)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_CancelRepair(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T, *Allocation) (teardown func(*testing.T))
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Test_Failed",
			wantErr: true,
			errMsg:  "invalid_cancel_repair_request: No repair in progress for the allocation",
		},
		{
			name: "Test_Success",
			setup: func(t *testing.T, a *Allocation) (teardown func(t *testing.T)) {
				a.repairRequestInProgress = &RepairRequest{}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{}
			setupMockAllocation(t, a)
			if tt.setup != nil {
				if teardown := tt.setup(t, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.CancelRepair()
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func setupMockAllocation(t *testing.T, a *Allocation) {
	a.uploadChan = make(chan *UploadRequest, 10)
	a.downloadChan = make(chan *DownloadRequest, 10)
	a.repairChan = make(chan *RepairRequest, 1)
	a.ctx, a.ctxCancelF = context.WithCancel(context.Background())
	a.uploadProgressMap = make(map[string]*UploadRequest)
	a.downloadProgressMap = make(map[string]*DownloadRequest)
	a.mutex = &sync.Mutex{}
	a.initialized = true
	sdkInitialized = true
	go func() {
		for true {
			select {
			case <-a.ctx.Done():
				t.Log("Upload cancelled by the parent")
				return
			case uploadReq := <-a.uploadChan:
				if uploadReq.completedCallback != nil {
					uploadReq.completedCallback(uploadReq.filepath)
				}
				if uploadReq.statusCallback != nil {
					uploadReq.statusCallback.Completed(a.ID, uploadReq.filepath, uploadReq.filemeta.Name, uploadReq.filemeta.MimeType, int(uploadReq.filemeta.Size), OpUpload)
				}
				if uploadReq.wg != nil {
					uploadReq.wg.Done()
				}
				t.Logf("received a upload request for %v %v\n", uploadReq.filepath, uploadReq.remotefilepath)
			case downloadReq := <-a.downloadChan:
				if downloadReq.completedCallback != nil {
					downloadReq.completedCallback(downloadReq.remotefilepath, downloadReq.remotefilepathhash)
				}
				if downloadReq.statusCallback != nil {
					downloadReq.statusCallback.Completed(a.ID, downloadReq.localpath, "1.txt", "application/octet-stream", 3, OpDownload)
				}
				if downloadReq.wg != nil {
					downloadReq.wg.Done()
				}
				t.Logf("received a download request for %v\n", downloadReq.remotefilepath)
			case repairReq := <-a.repairChan:
				if repairReq.completedCallback != nil {
					repairReq.completedCallback()
				}
				if repairReq.wg != nil {
					repairReq.wg.Done()
				}
				t.Logf("received a repair request for %v\n", repairReq.listDir.Path)
			}
		}
	}()
}

func setupMockGetFileInfoResponse(t *testing.T, mockClient *mocks.HttpClient) {
	httpResponse := http.Response{
		StatusCode: http.StatusOK,
		Body: func() io.ReadCloser {
			jsonFR, err := json.Marshal(fileref.FileRef{
				Ref: fileref.Ref{
					Name: mockFileRefName,
				},
				ContentHash: "mock content hash",
			})
			require.NoError(t, err)
			return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
		}(),
	}
	for i := 0; i < numBlobbers; i++ {
		mockClient.On("Do", mock.Anything).Return(&httpResponse, nil)
	}
}

func getMockAuthTicket(t *testing.T) string {
	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}
	a := &Allocation{
		ID:           mockAllocationId,
		Tx:           mockAllocationTxId,
		DataShards:   1,
		ParityShards: 1,
	}
	setupMockGetFileInfoResponse(t, &mockClient)
	a.InitAllocation()
	sdkInitialized = true
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      strconv.Itoa(i),
			Baseurl: "TestAllocation_getMockAuthTicket" + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	httpResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body: func() io.ReadCloser {
			jsonFR, err := json.Marshal(fileref.FileRef{
				Ref: fileref.Ref{
					Name: mockFileRefName,
				},
				ContentHash: "mock content hash",
			})
			require.NoError(t, err)
			return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
		}(),
	}
	for i := 0; i < numBlobbers; i++ {
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return strings.HasPrefix(req.URL.Path, "TestAllocation_ListDirFromAuthTicket")
		})).Return(httpResponse, nil)

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return strings.HasPrefix(req.URL.Path, "TestAllocation_getMockAuthTicket")
		})).Return(httpResponse, nil)
	}
	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, mockClientId, "", 0)
	require.NoErrorf(t, err, "unexpected get auth ticket error: %v", err)
	require.NotEmptyf(t, authTicket, "unexpected empty auth ticket")
	return authTicket
}
