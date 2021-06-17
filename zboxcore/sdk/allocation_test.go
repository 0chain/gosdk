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
	"testing"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/assert"
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
	mockType           = "d"
	numBlobbers        = 4
)

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
	} else if err.Error() != expectedErr {
		t.Errorf("uploadOrUpdateFile() = expected error message to be %v  but was %v", expectedErr, err.Error())
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
	} else if err.Error() != expectedErr {
		t.Errorf("uploadOrUpdateFile() = expected error message to be %v  but was %v", expectedErr, err.Error())
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
					return strings.HasPrefix(req.URL.Path, testName)
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
				Baseurl: tt.name + mockBlobberUrl,
			})
			got := a.GetBlobberStats()
			require.NotEmptyf(got, "Error no blobber stats result found")

			expected := make(map[string]*BlobberAllocationStats, 1)
			expected[tt.name+mockBlobberUrl] = &BlobberAllocationStats{
				ID:         mockAllocationId,
				Tx:         mockAllocationTxId,
				BlobberID:  tt.name + mockBlobberId,
				BlobberURL: tt.name + mockBlobberUrl,
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

func TestAllocation_UpdateFile(t *testing.T) {
	const mockLocalPath = "testdata/alloc/1.txt"
	require := require.New(t)
	a := &Allocation{}
	a.InitAllocation()
	sdkInitialized = true
	err := a.UpdateFile(mockLocalPath, "/", fileref.Attributes{}, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_UploadFile(t *testing.T) {
	const mockLocalPath = "testdata/alloc/1.txt"
	require := require.New(t)
	a := &Allocation{}
	a.InitAllocation()
	sdkInitialized = true
	err := a.UploadFile(mockLocalPath, "/", fileref.Attributes{}, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_RepairFile(t *testing.T) {
	const (
		mockFileRefName = "mock file ref name"
		mockLocalPath   = "testdata/alloc/1.txt"
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
			url := mockBlobberUrl + strconv.Itoa(i)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.Path, testName+url)
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
			a := &Allocation{}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < tt.numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			tt.setup(t, tt.name, tt.numBlobbers, tt.numCorrect)
			err := a.RepairFile(tt.parameters.localPath, tt.parameters.remotePath, tt.parameters.status)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
				return
			}
			require.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_UpdateFileWithThumbnail(t *testing.T) {
	const (
		mockLocalPath     = "testdata/alloc/1.txt"
		mockThumbnailPath = "thumbnail_alloc"
	)

	type parameters struct {
		localPath, remotePath, thumbnailPath string
		status                               StatusCallback
	}
	tests := []struct {
		name       string
		parameters parameters
		wantErr    bool
	}{
		{
			"Test_Coverage",
			parameters{
				localPath:     mockLocalPath,
				remotePath:    "/",
				thumbnailPath: mockThumbnailPath,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{}
			a.InitAllocation()
			sdkInitialized = true
			err := a.UpdateFileWithThumbnail(tt.parameters.localPath, tt.parameters.remotePath, tt.parameters.thumbnailPath, fileref.Attributes{}, tt.parameters.status)
			if tt.wantErr {
				require.Errorf(err, "expected error != nil")
				return
			}
			require.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_UploadFileWithThumbnail(t *testing.T) {
	const (
		mockLocalPath     = "testdata/alloc/1.txt"
		mockThumbnailPath = "thumbnail_alloc"
	)
	require := require.New(t)
	a := &Allocation{}
	a.InitAllocation()
	sdkInitialized = true
	err := a.UploadFileWithThumbnail(mockLocalPath, "/", mockThumbnailPath, fileref.Attributes{}, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUpdateFile(t *testing.T) {
	const mockLocalPath = "testdata/alloc/1.txt"
	require := require.New(t)
	a := &Allocation{}
	a.InitAllocation()
	sdkInitialized = true
	err := a.EncryptAndUpdateFile(mockLocalPath, "/", fileref.Attributes{}, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUploadFile(t *testing.T) {
	const mockLocalPath = "testdata/alloc/1.txt"
	require := require.New(t)
	a := &Allocation{}
	a.InitAllocation()
	sdkInitialized = true
	err := a.EncryptAndUploadFile(mockLocalPath, "/", fileref.Attributes{}, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUpdateFileWithThumbnail(t *testing.T) {
	const (
		mockLocalPath     = "testdata/alloc/1.txt"
		mockThumbnailPath = "thumbnail_alloc"
	)
	require := require.New(t)
	a := &Allocation{}
	a.InitAllocation()
	sdkInitialized = true
	err := a.EncryptAndUpdateFileWithThumbnail(mockLocalPath, "/", mockThumbnailPath, fileref.Attributes{}, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUploadFileWithThumbnail(t *testing.T) {
	const (
		mockLocalPath     = "testdata/alloc/1.txt"
		mockThumbnailPath = "thumbnail_alloc"
	)
	require := require.New(t)
	a := &Allocation{}
	a.InitAllocation()
	sdkInitialized = true
	err := a.EncryptAndUploadFileWithThumbnail(mockLocalPath, "/", mockThumbnailPath, fileref.Attributes{}, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_uploadOrUpdateFile(t *testing.T) {
	const (
		mockFileRefName   = "mock file ref name"
		mockLocalPath     = "testdata/alloc/1.txt"
		mockActualHash    = "4041e3eeb170751544a47af4e4f9d374e76cee1d"
		mockErrorHash     = "1041e3eeb170751544a47af4e4f9d374e76cee1d"
		mockThumbnailPath = "thumbnail_alloc"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponses := func(t *testing.T, testcaseName string, a *Allocation, hash string) (teardown func(t *testing.T)) {
		for i := 0; i < numBlobbers; i++ {
			var hash string
			if i < numBlobbers-1 {
				hash = mockErrorHash
			}
			frName := mockFileRefName + strconv.Itoa(i)
			url := mockBlobberUrl + strconv.Itoa(i)
			a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
				Baseurl: testcaseName + url,
			})
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.Path, testcaseName+url)
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
		return nil
	}

	type parameters struct {
		localPath     string
		remotePath    string
		status        StatusCallback
		isUpdate      bool
		thumbnailPath string
		encryption    bool
		isRepair      bool
		attrs         fileref.Attributes
		hash          string
	}
	tests := []struct {
		name       string
		setup      func(*testing.T, string, *Allocation, string) (teardown func(*testing.T))
		parameters parameters
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Not_Initialize_Failed",
			setup: func(t *testing.T, testcaseName string, a *Allocation, hash string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) { a.initialized = true }
			},
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
				attrs:         fileref.Attributes{},
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name:  "Test_Error_Local_File_Failed",
			setup: nil,
			parameters: parameters{
				localPath:     "local_file_error",
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
				attrs:         fileref.Attributes{},
			},
			wantErr: true,
			errMsg:  "Local file error: stat local_file_error: no such file or directory",
		},
		{
			name:  "Test_Thumbnail_File_Error_Success",
			setup: setupHttpResponses,
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: mockThumbnailPath,
				encryption:    false,
				isRepair:      false,
				attrs:         fileref.Attributes{},
				hash:          mockActualHash,
			},
		},
		{
			name:  "Test_Invalid_Remote_Abs_Path_Failed",
			setup: nil,
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
				attrs:         fileref.Attributes{},
			},
			wantErr: true,
			errMsg:  "invalid_path: Path should be valid and absolute",
		},
		{
			name:  "Test_Repair_Remote_File_Not_Found_Failed",
			setup: nil,
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "/x.txt",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      true,
				attrs:         fileref.Attributes{},
			},
			wantErr: true,
			errMsg:  "File not found for the given remotepath",
		},
		{
			name:  "Test_Repair_Content_Hash_Not_Matches_Failed",
			setup: setupHttpResponses,
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      true,
				attrs:         fileref.Attributes{},
				hash:          mockErrorHash,
			},
			wantErr: true,
			errMsg:  "Content hash doesn't match",
		},
		{
			name:  "Test_Upload_Success",
			setup: setupHttpResponses,
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
				attrs:         fileref.Attributes{},
				hash:          mockActualHash,
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
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a, tt.parameters.hash); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.uploadOrUpdateFile(tt.parameters.localPath, tt.parameters.remotePath, tt.parameters.status, tt.parameters.isUpdate, tt.parameters.thumbnailPath, tt.parameters.encryption, tt.parameters.isRepair, tt.parameters.attrs)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
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
			setup: func(t *testing.T, testcaseName string, a *Allocation) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					url := mockBlobberUrl + strconv.Itoa(i)
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						Baseurl: testcaseName + url,
					})
					mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return strings.HasPrefix(req.URL.Path, testcaseName+url)
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
			setup: func(t *testing.T, testcaseName string, a *Allocation) (teardown func(t *testing.T)) {
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
			setup: func(t *testing.T, testcaseName string, a *Allocation) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					var hash string
					if i < numBlobbers-1 {
						hash = mockActualHash
					}
					url := mockBlobberUrl + strconv.Itoa(i)
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						Baseurl: testcaseName + url,
					})
					mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return strings.HasPrefix(req.URL.Path, testcaseName+url)
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
			setup: func(t *testing.T, testcaseName string, a *Allocation) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					url := mockBlobberUrl + strconv.Itoa(i)
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						Baseurl: testcaseName + url,
					})
					mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return strings.HasPrefix(req.URL.Path, testcaseName+url)
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
				require.EqualValues(tt.errMsg, err.Error())
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
	require := require.New(t)
	a := &Allocation{}
	a.InitAllocation()
	sdkInitialized = true
	for i := 0; i < 4; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      strconv.Itoa(i),
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
	require := require.New(t)
	a := &Allocation{}
	a.InitAllocation()
	sdkInitialized = true
	for i := 0; i < 4; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      strconv.Itoa(i),
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
	a.InitAllocation()
	sdkInitialized = true
	for i := 0; i < 4; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      strconv.Itoa(i),
			Baseurl: mockBlobberUrl + strconv.Itoa(i),
		})
	}
	err := a.DownloadThumbnail(mockLocalPath, "/", nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_downloadFile(t *testing.T) {
	const (
		mockActualHash     = "4041e3eeb170751544a47af4e4f9d374e76cee1d"
		mockLocalPath      = "mock/local/path"
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
		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
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
			setup: func(t *testing.T, testcaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) { a.initialized = true }
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Local_Path_Is_Not_Dir_Failed",
			parameters: parameters{
				localPath:      "testdata/downloadFile/Test_Local_Path_Is_Not_Dir_Failed",
				remotePath:     mockRemoteFilePath,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				statusCallback: nil,
			},
			setup:   nil,
			wantErr: true,
			errMsg:  "Local path is not a directory 'testdata/downloadFile/Test_Local_Path_Is_Not_Dir_Failed'",
		},
		{
			name: "Test_Local_File_Already_Existed_Failed",
			parameters: parameters{
				localPath:      "testdata/alloc",
				remotePath:     mockRemoteFilePath,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				statusCallback: nil,
			},
			setup:   nil,
			wantErr: true,
			errMsg:  "Local file already exists 'testdata/alloc/1.txt'",
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
			setup: func(t *testing.T, testcaseName string, a *Allocation) (teardown func(t *testing.T)) {
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
			setup: func(t *testing.T, testcaseName string, a *Allocation) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					url := mockBlobberUrl + strconv.Itoa(i)
					mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
						return strings.HasPrefix(req.URL.Path, testcaseName+url)
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
					os.Remove("testdata/downloadFile/alloc/1.txt")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + strconv.Itoa(i),
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if m := tt.setup; m != nil {
				if teardown := m(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.downloadFile(tt.parameters.localPath, tt.parameters.remotePath, tt.parameters.contentMode, tt.parameters.startBlock, tt.parameters.endBlock, tt.parameters.numBlocks, tt.parameters.statusCallback)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
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
			Baseurl: mockBlobberUrl + strconv.Itoa(i),
		})
	}

	body, err := json.Marshal(&fileref.ReferencePath{
		Meta: map[string]interface{}{
			"type": mockType,
		},
	})
	require.NoError(err)
	setupMockHttpResponse(t, &mockClient, "", a, http.MethodGet, http.StatusOK, body)
	setupMockHttpResponse(t, &mockClient, "", a, http.MethodDelete, http.StatusOK, []byte(""))
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
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodDelete, http.StatusOK, []byte(""))
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
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
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
				require.EqualValues(tt.errMsg, err.Error())
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
			setup: func(t *testing.T, testcaseName string, p parameters, a *Allocation) (teardown func(t *testing.T)) {
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
				setupMockHttpResponse(t, &mockClient, testName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, testName, a, http.MethodPost, p.statusCode, []byte(""))
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
				setupMockHttpResponse(t, &mockClient, testName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, testName, a, http.MethodPost, p.statusCode, []byte(""))
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
				setupMockHttpResponse(t, &mockClient, testName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, testName, a, http.MethodPost, p.statusCode, []byte(""))
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
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
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
				require.EqualValues(tt.errMsg, err.Error())
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
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodPost, http.StatusOK, []byte(""))
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodDelete, http.StatusOK, []byte(""))
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
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
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
				require.EqualValues(tt.errMsg, err.Error())
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
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodPost, http.StatusOK, []byte(""))
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
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
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
				require.EqualValues(tt.errMsg, err.Error())
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
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodPost, http.StatusOK, []byte(""))
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
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
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
				require.EqualValues(tt.errMsg, err.Error())
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
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodPost, http.StatusBadRequest, []byte(""))
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
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodPost, http.StatusOK, []byte(""))
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
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
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
				require.EqualValues(tt.errMsg, err.Error())
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
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodDelete, http.StatusBadRequest, []byte(""))
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
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodDelete, http.StatusOK, []byte(""))
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
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
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
				require.EqualValues(tt.errMsg, err.Error())
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
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodPost, http.StatusBadRequest, body)
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
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodPost, http.StatusOK, body)
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
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
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
				require.EqualValues(tt.errMsg, err.Error())
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
	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}
	require := require.New(t)
	a := &Allocation{}
	a.InitAllocation()
	sdkInitialized = true
	at, err := a.GetAuthTicketForShare("/1.txt", "1.txt", fileref.FILE, mockClientId)
	require.NotEmptyf(at, "unexpected empty auth ticket")
	require.NoErrorf(err, "unexpected error: %v", err)
}

func TestAllocation_GetAuthTicket(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

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
			name: "Test_Success_File_Type_Directory",
			parameters: parameters{
				path:            "/",
				filename:        "1.txt",
				referenceType:   fileref.DIRECTORY,
				refereeClientID: mockClientId,
			},
		},
		{
			name: "Test_Success_With_Referee_Encryption_Public_Key",
			parameters: parameters{
				path:                       "/1.txt",
				filename:                   "1.txt",
				referenceType:              fileref.FILE,
				refereeClientID:            mockClientId,
				refereeEncryptionPublicKey: "this is some public key",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": "f",
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodPost, http.StatusOK, body)
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
			require := require.New(t)
			a := &Allocation{}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			at, err := a.GetAuthTicket(tt.parameters.path, tt.parameters.filename, tt.parameters.referenceType, tt.parameters.refereeClientID, tt.parameters.refereeEncryptionPublicKey)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			require.NotEmptyf(at, "unexpected empty auth ticket")
		})
	}
}

func TestAllocation_CancelUpload(t *testing.T) {
	const localPath = "testdata/alloc/1.txt"
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
			errMsg:  "local_path_not_found: Invalid path. No upload in progress for the path ",
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
				require.EqualValues(tt.errMsg, err.Error())
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
			errMsg:  "local_path_not_found: Invalid path. No upload in progress for the path ",
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

// func TestAllocation_CommitFolderChange(t *testing.T) {
// 	const (
// 		mockPath = "/1.txt"
// 		mockType = "d"
// 	)

// 	var mockClient = mocks.HttpClient{}
// 	zboxutil.Client = &mockClient

// 	client := zclient.GetClient()
// 	client.Wallet = &zcncrypto.Wallet{
// 		ClientID:  mockClientId,
// 		ClientKey: mockClientKey,
// 	}

// 	type parameters struct {
// 		operation, preValue, currValue string
// 	}
// 	blockchain.SetQuerySleepTime(1)
// 	blockchain.SetMiners([]string{"http://127.0.0.1:46671"})
// 	blockchain.SetSharders([]string{"http://127.0.0.1:46461"})
// 	tests := []struct {
// 		name       string
// 		parameters parameters
// 		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
// 		wantErr    bool
// 		errMsg     string
// 	}{
// 		// {
// 		// 	name: "Test_Uninitialized_Failed",
// 		// 	setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
// 		// 		a.initialized = false
// 		// 		return func(t *testing.T) {
// 		// 			a.initialized = true
// 		// 		}
// 		// 	},
// 		// 	wantErr: true,
// 		// 	errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
// 		// },
// 		// {
// 		// 	name: "Test_Sharder_Verify_Txn_Failed",
// 		// 	setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
// 		// 		// 	body, err := json.Marshal(&transaction.Transaction{
// 		// 		// 		Hash: "1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23",
// 		// 		// 	})
// 		// 		// require.NoError(t, err)
// 		// 		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
// 		// 			return req.Method == "POST"
// 		// 		})).Return(&http.Response{
// 		// 			StatusCode: http.StatusOK,
// 		// 			Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
// 		// 					"txn": {
// 		// 						"hash": "1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23",
// 		// 						"version": "1.0",
// 		// 						"client_id": "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
// 		// 						"public_key": "eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24",
// 		// 						"transaction_data": "{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}",
// 		// 						"transaction_value": 0,
// 		// 						"signature": "98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92",
// 		// 						"creation_date": 1617159987,
// 		// 						"transaction_type": 10,
// 		// 						"transaction_fee": 0,
// 		// 						"txn_output_hash": ""
// 		// 					},
// 		// 					"block_hash": "4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"
// 		// 				}`))),
// 		// 		}, nil).Once()
// 		// 		return nil
// 		// 	},
// 		// 	wantErr: true,
// 		// 	errMsg:  "missing_transaction_detail: No transaction detail was found on any of the sharders",
// 		// },
// 		// {
// 		// 	name: "Test_Max_Retried_Failed",
// 		// 	setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
// 		// 		maxTxnQuery := blockchain.GetMaxTxnQuery()
// 		// 		blockchain.SetMaxTxnQuery(0)
// 		// 		return func(t *testing.T) {
// 		// 			blockchain.SetMaxTxnQuery(maxTxnQuery)
// 		// 		}
// 		// 	},
// 		// 	wantErr: true,
// 		// 	errMsg:  "transaction_validation_failed: Failed to get the transaction confirmation",
// 		// },
// 		{
// 			name: "Test_Success",
// 			parameters: parameters{
// 				operation: "Move",
// 				preValue:  "/1.txt",
// 				currValue: "/d/1.txt",
// 			},
// 			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
// 				body, err := json.Marshal(&transaction.Transaction{
// 					Hash: "1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23",
// 				})
// 				require.NoError(t, err)
// 				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
// 					return req.Method == "POST"
// 				})).Return(&http.Response{
// 					StatusCode: http.StatusOK,
// 					Body:       ioutil.NopCloser(bytes.NewReader(body)),
// 				}, nil).Once()

// 				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
// 					return req.Method == "GET"
// 				})).Return(&http.Response{
// 					StatusCode: http.StatusOK,
// 					Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
// 							"txn": {
// 								"hash": "1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23",
// 								"version": "1.0",
// 								"client_id": "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
// 								"public_key": "eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24",
// 								"transaction_data": "{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}",
// 								"transaction_value": 0,
// 								"signature": "98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92",
// 								"creation_date": 1617159987,
// 								"transaction_type": 10,
// 								"transaction_fee": 0,
// 								"txn_output_hash": ""
// 							},
// 							"block_hash": "4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"
// 						}`))),
// 				}, nil).Once()
// 				return nil
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			require := require.New(t)
// 			a := &Allocation{
// 				ID:           mockAllocationId,
// 				Tx:           mockAllocationTxId,
// 				DataShards:   2,
// 				ParityShards: 2,
// 			}
// 			a.InitAllocation()
// 			sdkInitialized = true
// 			if tt.setup != nil {
// 				if teardown := tt.setup(t, tt.name, a); teardown != nil {
// 					defer teardown(t)
// 				}
// 			}
// 			_, err := a.CommitFolderChange(tt.parameters.operation, tt.parameters.preValue, tt.parameters.currValue)
// 			require.EqualValues(tt.wantErr, err != nil)
// 			if err != nil {
// 				require.EqualValues(tt.errMsg, err.Error())
// 				return
// 			}
// 			require.NoErrorf(err, "unexpected error: %v", err)
// 			// require.Equal(string(expectedBytes), got)
// 		})
// 	}
// }

func TestAllocation_ListDirFromAuthTicket(t *testing.T) {
	const (
		mockLookupHash = "mock lookup hash"
		mockType       = "d"
	)

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
	a.InitAllocation()
	sdkInitialized = true
	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, mockClientId, "")
	require.NoErrorf(t, err, "unexpected get auth ticket error: %v", err)
	require.NotEmptyf(t, authTicket, "unexpected empty auth ticket")

	type parameters struct {
		authTicket     string
		lookupHash     string
		expectedResult *ListResult
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
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
			name: "Test_Cannot_Decode_Auth_Ticket_Failed",
			parameters: parameters{
				authTicket: "some wrong auth ticket to decode",
			},
			setup:   nil,
			wantErr: true,
			errMsg:  "auth_ticket_decode_error: Error decoding the auth ticket.illegal base64 data at input byte 4",
		},
		{
			name: "Test_Cannot_Unmarshal_Auth_Ticket_Failed",
			parameters: parameters{
				authTicket: "c29tZSB3cm9uZyBhdXRoIHRpY2tldCB0byBtYXJzaGFs",
			},
			setup:   nil,
			wantErr: true,
			errMsg:  "auth_ticket_decode_error: Error unmarshaling the auth ticket.invalid character 's' looking for beginning of value",
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
			setup: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						Baseurl: testCaseName + mockBlobberUrl + strconv.Itoa(i),
					})
				}
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodGet, http.StatusBadRequest, []byte(""))
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
			setup: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						ID:      testCaseName + mockBlobberId + strconv.Itoa(i),
						Baseurl: testCaseName + mockBlobberUrl + strconv.Itoa(i),
					})
				}
				body, err := json.Marshal(&fileref.ListResult{
					AllocationRoot: mockAllocationRoot,
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodGet, http.StatusOK, body)
				return nil
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
			got, err := a.ListDirFromAuthTicket(tt.parameters.authTicket, tt.parameters.lookupHash)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
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

	a := &Allocation{
		ID: mockAllocationId,
		Tx: mockAllocationTxId,
	}
	a.InitAllocation()
	sdkInitialized = true
	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, mockClientId, "")
	require.NoErrorf(t, err, "unexpected get auth ticket error: %v", err)
	require.NotEmptyf(t, authTicket, "unexpected empty auth ticket")

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
				localPath:  "testdata/downloadFromAuthTicket/Test_Local_Path_Is_Not_Directory_Failed",
				authTicket: authTicket,
			},
			wantErr: true,
			errMsg:  "Local path is not a directory 'testdata/downloadFromAuthTicket/Test_Local_Path_Is_Not_Directory_Failed'",
		},
		{
			name: "Test_Local_File_Already_Exists_Failed",
			parameters: parameters{
				localPath:      "testdata/alloc",
				authTicket:     authTicket,
				remoteFilename: mockRemoteFilePath,
			},
			wantErr: true,
			errMsg:  "Local file already exists 'testdata/alloc/1.txt'",
		},
		{
			name: "Test_Not_Enough_Minimum_Blobbers_Failed",
			parameters: parameters{
				localPath:      mockLocalPath,
				authTicket:     authTicket,
				remoteFilename: mockRemoteFilePath,
			},
			setup: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
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
				remoteFilename: mockRemoteFilePath,
				authTicket:     authTicket,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				lookupHash:     mockLookupHash,
			},
			setup: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						ID:      testCaseName + strconv.Itoa(i),
						Baseurl: testCaseName + mockBlobberUrl + strconv.Itoa(i),
					})
				}
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodGet, http.StatusOK, body)
				return func(t *testing.T) {
					os.Remove("alloc/1.txt")
				}
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
			err := a.downloadFromAuthTicket(tt.parameters.localPath, tt.parameters.authTicket, tt.parameters.lookupHash, tt.parameters.startBlock, tt.parameters.endBlock, tt.parameters.numBlocks, tt.parameters.remoteFilename, tt.parameters.contentMode, tt.parameters.rxPay, tt.parameters.statusCallback)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
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

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		path                           string
		consensusThresh, fullConsensus float32
		expectedResult                 *ListResult
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
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodGet, http.StatusBadRequest, []byte(""))
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
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.ListResult{
					AllocationRoot: mockAllocationRoot,
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodGet, http.StatusOK, body)
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			got, err := a.listDir(tt.parameters.path, tt.parameters.consensusThresh, tt.parameters.fullConsensus)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
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
	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, mockClientId, "")
	require.NoErrorf(t, err, "unexpected get auth ticket error: %v", err)
	require.NotEmptyf(t, authTicket, "unexpected empty auth ticket")

	type parameters struct {
		authTicket, lookupHash string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
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
			setup: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				for i := 0; i < numBlobbers; i++ {
					a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
						ID:      testCaseName + mockBlobberId + strconv.Itoa(i),
						Baseurl: testCaseName + mockBlobberUrl + strconv.Itoa(i),
					})
				}
				body, err := json.Marshal(&fileref.FileRef{
					ActualFileHash: mockActualHash,
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodPost, http.StatusOK, body)
				return nil
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
			got, err := a.GetFileMetaFromAuthTicket(tt.parameters.authTicket, tt.parameters.lookupHash)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
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
	a.InitAllocation()
	sdkInitialized = true
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      strconv.Itoa(i),
			Baseurl: mockBlobberUrl + strconv.Itoa(i),
		})
	}

	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, mockClientId, "")
	require.NoErrorf(err, "unexpected get auth ticket error: %v", err)
	require.NotEmptyf(authTicket, "unexpected auth ticket")

	body, err := json.Marshal(&fileref.ReferencePath{
		Meta: map[string]interface{}{
			"type": mockType,
		},
	})
	require.NoError(err)
	setupMockHttpResponse(t, &mockClient, "", a, http.MethodGet, http.StatusOK, body)

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
	a.InitAllocation()
	sdkInitialized = true
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      strconv.Itoa(i),
			Baseurl: mockBlobberUrl + strconv.Itoa(i),
		})
	}

	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, mockClientId, "")
	require.NoErrorf(err, "unexpected get auth ticket error: %v", err)
	require.NotEmptyf(authTicket, "unexpected auth ticket")

	err = a.DownloadFromAuthTicket(mockLocalPath, authTicket, mockLookupHash, mockRemoteFilePath, true, nil)
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
			ID:      strconv.Itoa(i),
			Baseurl: mockBlobberUrl + strconv.Itoa(i),
		})
	}

	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, mockClientId, "")
	require.NoErrorf(err, "unexpected get auth ticket error: %v", err)
	require.NotEmptyf(authTicket, "unexpected auth ticket")

	err = a.DownloadFromAuthTicketByBlocks(mockLocalPath, authTicket, 1, 0, numBlockDownloads, mockLookupHash, mockRemoteFilePath, true, nil)
	defer os.Remove("alloc/1.txt")
	require.NoErrorf(err, "unexpected error: %v", err)
}

func TestAllocation_CommitMetaTransaction(t *testing.T) {
	const (
		mockLookupHash = "mock lookup hash"
		mockType       = "d"
	)

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

	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, mockClientId, "")
	assert.NoErrorf(t, err, "unexpected get auth ticket error: %v", err)
	assert.NotEmptyf(t, authTicket, "unexpected empty auth ticket")

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
				require.EqualValues(tt.errMsg, err.Error())
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_StartRepair(t *testing.T) {
	const (
		mockLookupHash   = "mock lookup hash"
		mockLocalPath    = "/alloc/1.txt"
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
				setupMockHttpResponse(t, &mockClient, testCaseName, a, http.MethodGet, http.StatusOK, body)
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
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
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
				require.EqualValues(tt.errMsg, err.Error())
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
			a.InitAllocation()
			sdkInitialized = true
			if tt.setup != nil {
				if teardown := tt.setup(t, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.CancelRepair()
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func setupMockHttpResponse(t *testing.T, mockClient *mocks.HttpClient, testCaseName string, a *Allocation, httpMethod string, statusCode int, body []byte) {
	for i := 0; i < numBlobbers; i++ {
		url := mockBlobberUrl + strconv.Itoa(i)
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == httpMethod &&
				strings.HasPrefix(req.URL.Path, testCaseName+url)
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
