package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"os"
	"strings"
	"testing"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/sdk/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	allocationTestDir = configDir + "/allocation"
)
func TestThrowErrorWhenBlobbersRequiredGreaterThanImplicitLimit128(t *testing.T) {
	teardown := setupMocks()
	defer teardown()

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
	teardown := setupMocks()
	defer teardown()

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
	teardown := setupMocks()
	defer teardown()

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

func setupMocks() (teardown func()) {
	fn := GetFileInfo
	GetFileInfo = func(localpath string) (os.FileInfo, error) {
		return new(MockFile), nil
	}
	return func() {
		GetFileInfo = fn
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
			assertion := assert.New(t)
			var check = assertion.False
			if tt.want {
				check = assertion.True
			}
			check(got)
		})
	}
}

func TestAllocation_InitAllocation(t *testing.T) {
	a := Allocation{}
	a.InitAllocation()
	assert.New(t).NotZero(a)
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
	assert.New(t).Same(stats, got)
}

func TestAllocation_GetBlobberStats(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	blobbers := a.Blobbers

	tests := []struct {
		name     string
		blobbers []*blockchain.StorageNode
		want     map[string]*BlobberAllocationStats
	}{
		{
			"Test_Success",
			blobbers,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/"+"GetBlobberStats", tt.name, responseParamTypeCheck)
			expectedBytes := parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v.json", allocationTestDir, "GetBlobberStats", tt.name), nil)
			expectedStr := string(expectedBytes)
			for blobberIdx, blobber := range blobbers {
				expectedStr = strings.ReplaceAll(expectedStr, blobberIDMask(blobberIdx+1), blobber.ID)
				expectedStr = strings.ReplaceAll(expectedStr, blobberURLMask(blobberIdx+1), blobber.Baseurl)
			}
			expectedBytes = []byte(expectedStr)
			var expected map[string]*BlobberAllocationStats
			err := json.Unmarshal(expectedBytes, &expected)
			assertion.NoErrorf(err, "Error json.Unmarshal() cannot parse blobber stats result format: %v", err)
			got := a.GetBlobberStats()
			if expected == nil || len(expected) == 0 {
				assertion.EqualValues(expected, got)
				return
			}

			assertion.NotEmptyf(got, "Error no blobber stats result found")
			for key, val := range expected {
				assertion.NotNilf(got[key], "Error result map must be contain key %v", key)
				assertion.EqualValues(val, got[key])
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
			assertion := assert.New(t)
			if tt.want {
				assertion.True(got, `Error a.isInitialized() should returns "true"", but got "false"`)
				return
			}
			assertion.False(got, `Error a.isInitialized() should returns "false"", but got "true"`)
		})
	}
}

func TestAllocation_UpdateFile(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/alloc/1.txt"
	assertion := assert.New(t)
	err := a.UpdateFile(localPath, "/", fileref.Attributes{}, nil)
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_UploadFile(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/alloc/1.txt"
	assertion := assert.New(t)
	err := a.UploadFile(localPath, "/", fileref.Attributes{}, nil)
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_RepairFile(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/alloc/1.txt"
	type args struct {
		localPath  string
		remotePath string
		status     StatusCallback
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Test_Repair_Required_Success",
			args{
				localPath:  localPath,
				remotePath: "/",
			},
			false,
		},
		{
			"Test_Repair_Not_Required_Failed",
			args{
				localPath:  localPath,
				remotePath: "/",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/"+"RepairFile", tt.name)
			err := a.RepairFile(tt.args.localPath, tt.args.remotePath, tt.args.status)
			if tt.wantErr {
				assertion.Errorf(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_UpdateFileWithThumbnail(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/alloc/1.txt"
	var thumbnailPath = allocationTestDir + "/thumbnail_alloc"
	type args struct {
		localPath, remotePath, thumbnailPath string
		status                               StatusCallback
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Test_Coverage",
			args{
				localPath:     localPath,
				remotePath:    "/",
				thumbnailPath: thumbnailPath,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			err := a.UpdateFileWithThumbnail(tt.args.localPath, tt.args.remotePath, tt.args.thumbnailPath, fileref.Attributes{}, tt.args.status)
			if tt.wantErr {
				assertion.Errorf(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_UploadFileWithThumbnail(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/alloc/1.txt"
	var thumbnailPath = allocationTestDir + "/thumbnail_alloc"

	assertion := assert.New(t)
	err := a.UploadFileWithThumbnail(localPath, "/", thumbnailPath, fileref.Attributes{}, nil)
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUpdateFile(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/alloc/1.txt"
	assertion := assert.New(t)
	err := a.EncryptAndUpdateFile(localPath, "/", fileref.Attributes{}, nil)
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUploadFile(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/alloc/1.txt"
	assertion := assert.New(t)
	err := a.EncryptAndUploadFile(localPath, "/", fileref.Attributes{}, nil)
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUpdateFileWithThumbnail(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/alloc/1.txt"
	var thumbnailPath = allocationTestDir + "/thumbnail_alloc"
	assertion := assert.New(t)
	err := a.EncryptAndUpdateFileWithThumbnail(localPath, "/", thumbnailPath, fileref.Attributes{}, nil)
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUploadFileWithThumbnail(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/alloc/1.txt"
	var thumbnailPath = allocationTestDir + "/thumbnail_alloc"
	assertion := assert.New(t)
	err := a.EncryptAndUploadFileWithThumbnail(localPath, "/", thumbnailPath, fileref.Attributes{}, nil)
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_uploadOrUpdateFile(t *testing.T) {
	var blockNums = 4
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, blockNums)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/alloc/1.txt"

	type args struct {
		localPath     string
		remotePath    string
		status        StatusCallback
		isUpdate      bool
		thumbnailPath string
		encryption    bool
		isRepair      bool
		attrs         fileref.Attributes
	}
	tests := []struct {
		name              string
		additionalSetupFn func(t *testing.T, testcaseName string) (teardown func(t *testing.T))
		args              args
		wantErr           bool
	}{
		{
			"Test_Not_Initialize_Failed",
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) { a.initialized = true }
			},
			args{
				localPath:     localPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
				attrs:         fileref.Attributes{},
			},
			true,
		},
		{
			"Test_Error_Local_File_Failed",
			nil,
			args{
				localPath:     "local_file_error",
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
				attrs:         fileref.Attributes{},
			},
			true,
		},
		{
			"Test_Thumbnail_File_Error_Success",
			nil,
			args{
				localPath:     localPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "thumbnail_file_error",
				encryption:    false,
				isRepair:      false,
				attrs:         fileref.Attributes{},
			},
			false,
		},
		{
			"Test_Invalid_Remote_Abs_Path_Failed",
			nil,
			args{
				localPath:     localPath,
				remotePath:    "",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
				attrs:         fileref.Attributes{},
			},
			true,
		},
		{
			"Test_Repair_Remote_File_Not_Found_Failed",
			nil,
			args{
				localPath:     localPath,
				remotePath:    "/x.txt",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      true,
				attrs:         fileref.Attributes{},
			},
			true,
		},
		{
			"Test_Repair_Content_Hash_Not_Matches_Failed",
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				setupBlobberMockResponses(t, blobberMocks, fmt.Sprintf("%v/%v", allocationTestDir, "uploadOrUpdateFile"), testcaseName)
				return nil
			},
			args{
				localPath:     localPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      true,
				attrs:         fileref.Attributes{},
			},
			true,
		},
		{
			"Test_Upload_Success",
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				willReturnCommitResult(&CommitResult{Success: true})
				return func(t *testing.T) {
					willReturnCommitResult(nil)
				}
			},
			args{
				localPath:     localPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
				attrs:         fileref.Attributes{},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if tt.additionalSetupFn != nil {
				if teardown := tt.additionalSetupFn(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.uploadOrUpdateFile(tt.args.localPath, tt.args.remotePath, tt.args.status, tt.args.isUpdate, tt.args.thumbnailPath, tt.args.encryption, tt.args.isRepair, tt.args.attrs)
			if tt.wantErr {
				assertion.Errorf(err, "Expected error != nil")
				return
			}
			assertion.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_RepairRequired(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var (
		blobberMockFn = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
			setupBlobberMockResponses(t, blobberMocks, fmt.Sprintf("%v/%v", allocationTestDir, "RepairRequired"), testcaseName, responseFormBodyTypeCheck)
			return func(t *testing.T) {
				for _, blobberMock := range blobberMocks {
					blobberMock.ResetHandler(t)
				}
			}
		}
		expectedFn = func(assertion *assert.Assertions, testcaseName string) *fileref.FileRef {
			var wantFileRef *fileref.FileRef
			parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v.json", allocationTestDir, "RepairRequired", testcaseName), &wantFileRef)
			return wantFileRef
		}
	)
	tests := []struct {
		name                          string
		additionalSetupFn             func(t *testing.T, testcaseName string) (teardown func(t *testing.T))
		remotePath                    string
		wantFound                     uint64
		wantFileRef                   func(assertion *assert.Assertions, testcaseName string) *fileref.FileRef
		wantMatchesConsensus, wantErr bool
	}{
		{
			"Test_Not_Repair_Required_Success",
			blobberMockFn,
			"/x.txt",
			0xf,
			expectedFn,
			false, false,
		},
		{
			"Test_Uninitialized_Failed",
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) { a.initialized = true }
			},
			"/",
			0,
			nil,
			false, true,
		},
		{
			"Test_Repair_Required_Success",
			blobberMockFn,
			"/",
			0x7,
			expectedFn,
			true, false,
		},
		{
			"Test_Remote_File_Not_Found_Failed",
			blobberMockFn,
			"/x.txt",
			0x0,
			func(assertion *assert.Assertions, testcaseName string) *fileref.FileRef {
				return nil
			},
			false, true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if tt.additionalSetupFn != nil {
				if teardown := tt.additionalSetupFn(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}

			found, matchesConsensus, fileRef, err := a.RepairRequired(tt.remotePath)
			assertion.Equal(zboxutil.NewUint128(tt.wantFound), found, "found value must be same")
			if tt.wantMatchesConsensus {
				assertion.True(tt.wantMatchesConsensus, matchesConsensus)
			} else {
				assertion.False(tt.wantMatchesConsensus, matchesConsensus)
			}
			if tt.wantErr {
				assertion.Errorf(err, "Expected error != nil")
				return
			}
			assertion.NoErrorf(err, "Unexpected error %v", err)
			assertion.EqualValues(tt.wantFileRef(assertion, tt.name), fileRef)
		})
	}
}

func TestAllocation_DownloadFile(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/DownloadFile"
	assertion := assert.New(t)
	err := a.DownloadFile(localPath, "/", nil)
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_DownloadFileByBlock(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/DownloadFileByBlock"
	assertion := assert.New(t)
	err := a.DownloadFileByBlock(localPath, "/", 1, 0, numBlockDownloads, nil)
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_DownloadThumbnail(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/DownloadThumbnail"

	assertion := assert.New(t)
	err := a.DownloadThumbnail(localPath, "/", nil)
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_downloadFile(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/downloadFile/alloc"
	var remotePath = "/1.txt"
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/downloadFile", testcaseName)
		return nil
	}

	type args struct {
		localPath, remotePath string
		contentMode           string
		startBlock, endBlock  int64
		numBlocks             int
		statusCallback        StatusCallback
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testcaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{
				localPath, remotePath,
				DOWNLOAD_CONTENT_FULL,
				1, 0,
				numBlockDownloads,
				nil,
			},
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) { a.initialized = true }
			},
			true,
		},
		{
			"Test_Local_Path_Is_Not_Dir_Failed",
			args{
				localPath:      allocationTestDir + "/downloadFile/Test_Local_Path_Is_Not_Dir_Failed",
				remotePath:     remotePath,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				statusCallback: nil,
			},
			nil,
			true,
		},
		{
			"Test_Local_File_Already_Existed_Failed",
			args{
				localPath:      allocationTestDir + "/alloc",
				remotePath:     remotePath,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				statusCallback: nil,
			},
			nil,
			true,
		},
		{
			"Test_No_Blobber_Failed",
			args{
				localPath:      localPath,
				remotePath:     remotePath,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				statusCallback: nil,
			},
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				blobbers := a.Blobbers
				a.Blobbers = []*blockchain.StorageNode{}
				return func(t *testing.T) {
					a.Blobbers = blobbers
				}
			},
			true,
		},
		{
			"Test_Download_File_Success",
			args{
				localPath:      localPath,
				remotePath:     remotePath,
				contentMode:    DOWNLOAD_CONTENT_FULL,
				startBlock:     1,
				endBlock:       0,
				numBlocks:      numBlockDownloads,
				statusCallback: nil,
			},
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testcaseName)
				return func(t *testing.T) {
					os.Remove(allocationTestDir + "/downloadFile/alloc/1.txt")
				}
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if m := tt.additionalMock; m != nil {
				if teardown := m(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.downloadFile(tt.args.localPath, tt.args.remotePath, tt.args.contentMode, tt.args.startBlock, tt.args.endBlock, tt.args.numBlocks, tt.args.statusCallback)
			if tt.wantErr {
				assertion.Error(err, "Expected error != nil")
				return
			}
			assertion.NoErrorf(err, "Unexpected error: %v", err)
		})
	}
}

func TestAllocation_DeleteFile(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	willReturnCommitResult(&CommitResult{Success: true})
	setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/DeleteFile", "TestAllocation_DeleteFile")
	assertion := assert.New(t)
	err := a.DeleteFile("/1.txt")
	assertion.NoErrorf(err, "unexpected error: %v", err)
}

func TestAllocation_deleteFile(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/deleteFile", testcaseName)
		return nil
	}

	type args struct {
		path string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Invalid_Path_Failed",
			args{},
			nil,
			true,
		},
		{
			"Test_Not_Abs_Path_Failed",
			args{"x.txt"},
			nil,
			true,
		},
		{
			"Test_Success",
			args{path: "/1.txt"},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{
					Success: true,
				})
				return func(t *testing.T) {
					willReturnCommitResult(nil)
				}
			},
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
			err := a.DeleteFile(tt.args.path)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_UpdateObjectAttributes(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	type args struct {
		path  string
		attrs fileref.Attributes
	}

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/UpdateObjectAttributes", testcaseName)
		return nil
	}

	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testcaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Invalid_Path_Failed",
			args{
				"",
				fileref.Attributes{WhoPaysForReads: common.WhoPaysOwner},
			},
			nil,
			true,
		},
		{
			"Test_Invalid_Remote_Abs_Path_Failed",
			args{
				"abc",
				fileref.Attributes{WhoPaysForReads: common.WhoPaysOwner},
			},
			nil,
			true,
		},
		{
			"Test_Update_Attributes_Failed",
			args{
				"/1.txt",
				fileref.Attributes{WhoPaysForReads: common.WhoPaysOwner},
			},
			blobbersResponseMock,
			true,
		},
		{
			"Test_Who_Pay_For_Read_Owner_Success",
			args{
				"/1.txt",
				fileref.Attributes{WhoPaysForReads: common.WhoPaysOwner},
			},
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testcaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return func(t *testing.T) {
					willReturnCommitResult(nil)
				}
			},
			false,
		},
		{
			"Test_Who_Pay_For_Read_3rd_Party_Success",
			args{
				"/1.txt",
				fileref.Attributes{WhoPaysForReads: common.WhoPays3rdParty},
			},
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testcaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return func(t *testing.T) {
					willReturnCommitResult(nil)
				}
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.UpdateObjectAttributes(tt.args.path, tt.args.attrs)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

// TestAllocation_MoveObject - the testing method implement two step:
// 1. copy file to other directory with same name using CopyObject method
// 2. delete old file using DeleteFile method
// Let's say the CopyObject Method and DeleteFile method are all tested and will returns all expected result, so that we
// just need to test the statement coverage rate in this case.
func TestAllocation_MoveObject(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/MoveObject", testcaseName)
		return nil
	}

	type args struct {
		path     string
		destPath string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Cover_Copy_Object",
			args{
				path:     "/1.txt",
				destPath: "/d",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Cover_Delete_Object",
			args{
				path:     "/1.txt",
				destPath: "/d",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{
					Success: true,
				})
				return func(t *testing.T) {
					willReturnCommitResult(nil)
				}
			},
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
			err := a.MoveObject(tt.args.path, tt.args.destPath)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_CopyObject(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/CopyObject", testcaseName)
		return nil
	}

	type args struct {
		path     string
		destPath string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Wrong_Path_Or_Destination_Path_Failed",
			args{
				path:     "",
				destPath: "",
			},
			nil,
			true,
		},
		{
			"Test_Invalid_Remote_Absolute_Path",
			args{
				path:     "abc",
				destPath: "/d",
			},
			nil,
			true,
		},
		{
			"Test_Success",
			args{
				path:     "/1.txt",
				destPath: "/d",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{
					Success: true,
				})
				return func(t *testing.T) {
					willReturnCommitResult(nil)
				}
			},
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
			err := a.CopyObject(tt.args.path, tt.args.destPath)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_RenameObject(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/RenameObject", testcaseName)
		return nil
	}

	type args struct {
		path     string
		destName string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Wrong_Path_Or_Destination_Path_Failed",
			args{
				path:     "",
				destName: "",
			},
			nil,
			true,
		},
		{
			"Test_Invalid_Remote_Absolute_Path",
			args{
				path:     "abc",
				destName: "/2.txt",
			},
			nil,
			true,
		},
		{
			"Test_Success",
			args{
				path:     "/1.txt",
				destName: "/2.txt",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{
					Success: true,
				})
				return func(t *testing.T) {
					willReturnCommitResult(nil)
				}
			},
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
			err := a.RenameObject(tt.args.path, tt.args.destName)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_AddCollaborator(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/AddCollaborator", testcaseName)
		return nil
	}

	type args struct {
		filePath       string
		collaboratorID string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Add_Collaborator_Error_Response_Failed",
			args{
				filePath:       "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
			},
			nil,
			true,
		},
		{
			"Test_Success",
			args{
				filePath:       "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
			},
			blobbersResponseMock,
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
			err := a.AddCollaborator(tt.args.filePath, tt.args.collaboratorID)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_RemoveCollaborator(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/RemoveCollaborator", testcaseName)
		return nil
	}

	type args struct {
		filePath       string
		collaboratorID string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Remove_Collaborator_Error_Response_Failed",
			args{
				filePath:       "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
			},
			nil,
			true,
		},
		{
			"Test_Success",
			args{
				filePath:       "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
			},
			blobbersResponseMock,
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
			err := a.RemoveCollaborator(tt.args.filePath, tt.args.collaboratorID)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_GetFileStats(t *testing.T) {
	var blobberNums = 4
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, blobberNums)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/GetFileStats", testcaseName)
		return nil
	}

	type args struct {
		path string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{
				path: "/1.txt",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Invalid_Path_Failed",
			args{
				path: "",
			},
			nil,
			true,
		},
		{
			"Test_Invalid_Remote_Absolute_Path_Failed",
			args{
				path: "x.txt",
			},
			nil,
			true,
		},
		{
			"Test_Error_Getting_File_Stats_From_Blobbers_Failed",
			args{
				path: "/1.txt",
			},
			nil,
			true,
		},
		{
			"Test_Success",
			args{
				path: "/1.txt",
			},
			blobbersResponseMock,
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
			got, err := a.GetFileStats(tt.args.path)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
			assertion.NotNil(got, "unexpected nullable file stats result")
			assertion.Equalf(blobberNums, len(got), "expected length of file stats result is %d, but got %v", blobberNums, len(got))
			for _, blobberMock := range blobberMocks {
				assertion.NotEmptyf(got[blobberMock.ID], "unexpected empty value of file stats related to blobber %v", blobberMock.ID)
			}
		})
	}
}

func TestAllocation_GetFileMeta(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	type args struct {
		path string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Error_Getting_File_Meta_Data_From_Blobbers_Failed",
			args{
				path: "/1.txt",
			},
			nil,
			true,
		},
		{
			"Test_Success",
			args{
				path: "/1.txt",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/GetFileMeta", testCaseName)
				return nil
			},
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
			got, err := a.GetFileMeta(tt.args.path)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
			var expectedResult *ConsolidatedFileMeta
			parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v.json", allocationTestDir, "GetFileMeta", tt.name), &expectedResult)
			assertion.EqualValues(expectedResult, got)
		})
	}
}

func TestAllocation_GetAuthTicketForShare(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	assertion := assert.New(t)
	at, err := a.GetAuthTicketForShare("/1.txt", "1.txt", fileref.FILE, client.GetClientID())
	assertion.NotEmptyf(at, "unexpected empty auth ticket")
	assertion.NoErrorf(err, "unexpected error: %v", err)
}

func TestAllocation_GetAuthTicket(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	type args struct {
		path                       string
		filename                   string
		referenceType              string
		refereeClientID            string
		refereeEncryptionPublicKey string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testcaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Success_File_Type_Directory",
			args{
				path:            "/",
				filename:        "1.txt",
				referenceType:   fileref.DIRECTORY,
				refereeClientID: client.GetClientID(),
			},
			nil,
			false,
		},
		{
			"Test_Success_With_Referee_Encryption_Public_Key",
			args{
				path:                       "/1.txt",
				filename:                   "1.txt",
				referenceType:              fileref.FILE,
				refereeClientID:            client.GetClientID(),
				refereeEncryptionPublicKey: "this is some public key",
			},
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/GetAuthTicket", testcaseName)
				return nil
			},
			false,
		},
		{
			"Test_Success_With_No_Referee_Encryption_Public_Key",
			args{
				path:                       "/1.txt",
				filename:                   "1.txt",
				referenceType:              fileref.FILE,
				refereeClientID:            client.GetClientID(),
				refereeEncryptionPublicKey: "",
			},
			nil,
			false,
		},
		{
			"Test_Invalid_Path_Failed",
			args{
				filename: "1.txt",
			},
			nil,
			true,
		},
		{
			"Test_Remote_Path_Not_Absolute_Failed",
			args{
				path: "x",
			},
			nil,
			true,
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
			at, err := a.GetAuthTicket(tt.args.path, tt.args.filename, tt.args.referenceType, tt.args.refereeClientID, tt.args.refereeEncryptionPublicKey)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
			assertion.NotEmptyf(at, "unexpected empty auth ticket")
		})
	}
}

func TestAllocation_CancelUpload(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/alloc/1.txt"
	type args struct {
		localpath string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Failed",
			args{},
			nil,
			true,
		},
		{
			"Test_Success",
			args{
				localpath: localPath,
			},
			func(t *testing.T) (teardown func(t *testing.T)) {
				a.uploadProgressMap[localPath] = &UploadRequest{}
				return nil
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if tt.additionalMock != nil {
				if teardown := tt.additionalMock(t); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.CancelUpload(tt.args.localpath)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_CancelDownload(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var remotePath = "/1.txt"
	type args struct {
		remotepath string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Failed",
			args{},
			nil,
			true,
		},
		{
			"Test_Success",
			args{
				remotepath: remotePath,
			},
			func(t *testing.T) (teardown func(t *testing.T)) {
				a.downloadProgressMap[remotePath] = &DownloadRequest{}
				return nil
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if tt.additionalMock != nil {
				if teardown := tt.additionalMock(t); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.CancelDownload(tt.args.remotepath)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_CommitFolderChange(t *testing.T) {
	// setup mock sdk
	miners, sharders, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 0)
	defer closeFn()
	// setup mock allocation
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var minerResponseMocks = func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
		setupMinerMockResponses(t, miners, allocationTestDir+"/CommitFolderChange", testCaseName)
		return nil
	}
	var sharderResponseMocks = func(t *testing.T, testCaseName string) {
		setupSharderMockResponses(t, sharders, allocationTestDir+"/CommitFolderChange", testCaseName)
	}

	type args struct {
		operation, preValue, currValue string
	}
	blockchain.SetQuerySleepTime(1)
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Sharder_Verify_Txn_Failed",
			args{},
			minerResponseMocks,
			true,
		},
		{
			"Test_Max_Retried_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				maxTxnQuery := blockchain.GetMaxTxnQuery()
				blockchain.SetMaxTxnQuery(0)
				return func(t *testing.T) {
					blockchain.SetMaxTxnQuery(maxTxnQuery)
				}
			},
			true,
		},
		{
			"Test_Success",
			args{
				operation: "Move",
				preValue:  "/1.txt",
				currValue: "/d/1.txt",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				minerResponseMocks(t, testCaseName)
				sharderResponseMocks(t, testCaseName)
				return nil
			},
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
			got, err := a.CommitFolderChange(tt.args.operation, tt.args.preValue, tt.args.currValue)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
			expectedBytes := parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v.json", allocationTestDir, "CommitFolderChange", tt.name), nil)
			assertion.Equal(string(expectedBytes), got)
		})
	}
}

func TestAllocation_ListDirFromAuthTicket(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var lookupHash = fileref.GetReferenceLookup(a.ID, "/1.txt")
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/ListDirFromAuthTicket", testcaseName)
		return nil
	}

	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, client.GetClientID(), "")
	assert.NoErrorf(t, err, "unexpected get auth ticket error: %v", err)
	assert.NotEmptyf(t, authTicket, "unexpected empty auth ticket")
	type args struct {
		authTicket, lookupHash string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Cannot_Decode_Auth_Ticket_Failed",
			args{
				authTicket: "some wrong auth ticket to decode",
			},
			nil,
			true,
		},
		{
			"Test_Cannot_Unmarshal_Auth_Ticket_Failed",
			args{
				authTicket: "c29tZSB3cm9uZyBhdXRoIHRpY2tldCB0byBtYXJzaGFs",
			},
			nil,
			true,
		},
		{
			"Test_Wrong_Auth_Ticket_File_Path_Hash_Or_Lookup_Hash_Failed",
			args{
				authTicket: authTicket,
				lookupHash: "",
			},
			nil,
			true,
		},
		{
			"Test_Error_Get_List_File_From_Blobbers_Failed",
			args{
				authTicket: authTicket,
				lookupHash: lookupHash,
			},
			nil,
			true,
		},
		{
			"Test_Success",
			args{
				authTicket: authTicket,
				lookupHash: lookupHash,
			},
			blobbersResponseMock,
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
			got, err := a.ListDirFromAuthTicket(tt.args.authTicket, tt.args.lookupHash)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
			var expectedResult *ListResult
			parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v.json", allocationTestDir, "ListDirFromAuthTicket", tt.name), &expectedResult)
			assertion.EqualValues(expectedResult, got)
		})
	}
}

func TestAllocation_downloadFromAuthTicket(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var localPath = allocationTestDir + "/downloadFromAuthTicket/alloc"
	var remoteFileName = "1.txt"
	var remoteLookupHash = fileref.GetReferenceLookup(a.ID, "/1.txt")
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/downloadFromAuthTicket", testcaseName)
		return nil
	}

	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, client.GetClientID(), "")
	assert.NoErrorf(t, err, "unexpected get auth ticket error: %v", err)
	assert.NotEmptyf(t, authTicket, "unexpected empty auth ticket")

	type args struct {
		localPath        string
		authTicket       string
		remoteLookupHash string
		startBlock       int64
		endBlock         int64
		numBlocks        int
		remoteFilename   string
		contentMode      string
		rxPay            bool
		statusCallback   StatusCallback
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Cannot_Decode_Auth_Ticket_Failed",
			args{
				authTicket: "some wrong auth ticket to decode",
			},
			nil,
			true,
		},
		{
			"Test_Cannot_Unmarshal_Auth_Ticket_Failed",
			args{
				authTicket: "c29tZSB3cm9uZyBhdXRoIHRpY2tldCB0byBtYXJzaGFs",
			},
			nil,
			true,
		},
		{
			"Test_Local_Path_Is_Not_Directory_Failed",
			args{
				localPath:  allocationTestDir + "/downloadFromAuthTicket/Test_Local_Path_Is_Not_Directory_Failed",
				authTicket: authTicket,
			},
			nil,
			true,
		},
		{
			"Test_Local_File_Already_Exists_Failed",
			args{
				localPath:      allocationTestDir + "/alloc",
				authTicket:     authTicket,
				remoteFilename: remoteFileName,
			},
			nil,
			true,
		},
		{
			"Test_Not_Enough_Minimum_Blobbers_Failed",
			args{
				localPath:      localPath,
				authTicket:     authTicket,
				remoteFilename: remoteFileName,
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbers := a.Blobbers
				a.Blobbers = []*blockchain.StorageNode{}
				return func(t *testing.T) {
					a.Blobbers = blobbers
				}
			},
			true,
		},
		{
			"Test_Download_File_Success",
			args{
				localPath:        localPath,
				remoteFilename:   remoteFileName,
				authTicket:       authTicket,
				contentMode:      DOWNLOAD_CONTENT_FULL,
				startBlock:       1,
				endBlock:         0,
				numBlocks:        numBlockDownloads,
				remoteLookupHash: remoteLookupHash,
			},
			func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testcaseName)
				return func(t *testing.T) {
					os.Remove(allocationTestDir + "/downloadFromAuthTicket/alloc/1.txt")
				}
			},
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
			err := a.downloadFromAuthTicket(tt.args.localPath, tt.args.authTicket, tt.args.remoteLookupHash, tt.args.startBlock, tt.args.endBlock, tt.args.numBlocks, tt.args.remoteFilename, tt.args.contentMode, tt.args.rxPay, tt.args.statusCallback)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_listDir(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()
	var path = "/1.txt"
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/ListDirFromAuthTicket", testcaseName)
		return nil
	}

	type args struct {
		path                           string
		consensusThresh, fullConsensus float32
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Invalid_Path_Failed",
			args{},
			nil,
			true,
		},
		{
			"Test_Invalid_Absolute_Path_Failed",
			args{
				path: "1.txt",
			},
			nil,
			true,
		},
		{
			"Test_Error_Get_List_File_From_Blobbers_Failed",
			args{
				path:            path,
				consensusThresh: 50,
				fullConsensus:   4,
			},
			nil,
			true,
		},
		{
			"Test_Success",
			args{
				path:            path,
				consensusThresh: 50,
				fullConsensus:   4,
			},
			blobbersResponseMock,
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
			got, err := a.listDir(tt.args.path, tt.args.consensusThresh, tt.args.fullConsensus)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
			var expectedResult *ListResult
			parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v.json", allocationTestDir, "ListDirFromAuthTicket", tt.name), &expectedResult)
			assertion.EqualValues(expectedResult, got)
		})
	}
}

func TestAllocation_GetFileMetaFromAuthTicket(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var lookupHash = fileref.GetReferenceLookup(a.ID, "/1.txt")
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/GetFileMetaFromAuthTicket", testcaseName)
		return nil
	}

	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, client.GetClientID(), "")
	assert.NoErrorf(t, err, "unexpected get auth ticket error: %v", err)
	assert.NotEmptyf(t, authTicket, "unexpected empty auth ticket")

	type args struct {
		authTicket, lookupHash string
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Cannot_Decode_Auth_Ticket_Failed",
			args{
				authTicket: "some wrong auth ticket to decode",
			},
			nil,
			true,
		},
		{
			"Test_Cannot_Unmarshal_Auth_Ticket_Failed",
			args{
				authTicket: "c29tZSB3cm9uZyBhdXRoIHRpY2tldCB0byBtYXJzaGFs",
			},
			nil,
			true,
		},
		{
			"Test_Wrong_Auth_Ticket_File_Path_Hash_Or_Lookup_Hash_Failed",
			args{
				authTicket: authTicket,
				lookupHash: "",
			},
			nil,
			true,
		},
		{
			"Test_Error_Get_File_Meta_From_Blobbers_Failed",
			args{
				authTicket: authTicket,
				lookupHash: lookupHash,
			},
			nil,
			true,
		},
		{
			"Test_Success",
			args{
				authTicket: authTicket,
				lookupHash: lookupHash,
			},
			blobbersResponseMock,
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
			got, err := a.GetFileMetaFromAuthTicket(tt.args.authTicket, tt.args.lookupHash)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
			var expectedResult *ConsolidatedFileMeta
			parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v.json", allocationTestDir, "GetFileMetaFromAuthTicket", tt.name), &expectedResult)
			assertion.EqualValues(expectedResult, got)
		})
	}
}

func TestAllocation_DownloadThumbnailFromAuthTicket(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/DownloadThumbnailFromAuthTicket", "TestAllocation_DownloadThumbnailFromAuthTicket")
	assertion := assert.New(t)
	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, client.GetClientID(), "")
	assertion.NoErrorf(err, "unexpected get auth ticket error: %v", err)
	assertion.NotEmptyf(authTicket, "unexpected auth ticket")
	var lookupHash = fileref.GetReferenceLookup(a.ID, "/1.txt")
	err = a.DownloadThumbnailFromAuthTicket(allocationTestDir+"/DownloadThumbnailFromAuthTicket/alloc", authTicket, lookupHash, "1.txt", true, nil)
	defer os.Remove(allocationTestDir + "/DownloadThumbnailFromAuthTicket/alloc/1.txt")
	assertion.NoErrorf(err, "unexpected error: %v", err)
}

func TestAllocation_DownloadFromAuthTicket(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/DownloadFromAuthTicket", "TestAllocation_DownloadFromAuthTicket")
	assertion := assert.New(t)
	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, client.GetClientID(), "")
	assertion.NoErrorf(err, "unexpected get auth ticket error: %v", err)
	assertion.NotEmptyf(authTicket, "unexpected auth ticket")
	var lookupHash = fileref.GetReferenceLookup(a.ID, "/1.txt")
	err = a.DownloadFromAuthTicket(allocationTestDir+"/DownloadFromAuthTicket/alloc", authTicket, lookupHash, "1.txt", true, nil)
	defer os.Remove(allocationTestDir + "/DownloadFromAuthTicket/alloc/1.txt")
	assertion.NoErrorf(err, "unexpected error: %v", err)
}

func TestAllocation_DownloadFromAuthTicketByBlocks(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/DownloadFromAuthTicketByBlocks", "TestAllocation_DownloadFromAuthTicketByBlocks")
	assertion := assert.New(t)
	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, client.GetClientID(), "")
	assertion.NoErrorf(err, "unexpected get auth ticket error: %v", err)
	assertion.NotEmptyf(authTicket, "unexpected auth ticket")
	var lookupHash = fileref.GetReferenceLookup(a.ID, "/1.txt")
	err = a.DownloadFromAuthTicketByBlocks(allocationTestDir+"/DownloadFromAuthTicketByBlocks/alloc", authTicket, 1, 0, numBlockDownloads, lookupHash, "1.txt", true, nil)
	defer os.Remove(allocationTestDir + "/DownloadFromAuthTicketByBlocks/alloc/1.txt")
	assertion.NoErrorf(err, "unexpected error: %v", err)
}

// TestAllocation_CommitMetaTransaction	- calling 3 dependence method: GetFileMeta, GetFileMetaFromAuthTicket and processCommitMetaRequest
// Let's says both that method are all tested itself. So we can ignore the test on these method, just need to make sure the statement are able to covered
func TestAllocation_CommitMetaTransaction(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, client.GetClientID(), "")
	assert.NoErrorf(t, err, "unexpected get auth ticket error: %v", err)
	assert.NotEmptyf(t, authTicket, "unexpected empty auth ticket")
	type args struct {
		path          string
		crudOperation string
		authTicket    string
		lookupHash    string
		fileMeta      func(t *testing.T, testCaseName string) *ConsolidatedFileMeta
		status        func(t *testing.T) StatusCallback
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_No_File_Meta_With_Path_Args_Failed",
			args{
				path:          "/1.txt",
				crudOperation: "",
				authTicket:    "",
				lookupHash:    fileref.GetReferenceLookup(a.ID, "/1.txt"),
				fileMeta:      nil,
			},
			nil,
			true,
		},
		{
			"Test_No_File_Meta_With_Auth_Ticket_Args_Failed",
			args{
				path:          "",
				crudOperation: "",
				authTicket:    authTicket,
				lookupHash:    fileref.GetReferenceLookup(a.ID, "/1.txt"),
				fileMeta:      nil,
			},
			nil,
			true,
		},
		{
			"Test_No_File_Meta_With_No_Path_And_No_Auth_Ticket_Args_Coverage",
			args{
				path:          "",
				crudOperation: "",
				authTicket:    "",
				lookupHash:    fileref.GetReferenceLookup(a.ID, "/1.txt"),
				fileMeta:      nil,
				status: func(t *testing.T) StatusCallback {
					scm := &mocks.StatusCallback{}
					scm.On("CommitMetaCompleted", mock.Anything, mock.Anything, mock.Anything).Maybe()
					return scm
				},
			},
			nil,
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

			var fileMeta *ConsolidatedFileMeta
			if tt.args.fileMeta != nil {
				fileMeta = tt.args.fileMeta(t, tt.name)
			}

			var status StatusCallback
			if tt.args.status != nil {
				status = tt.args.status(t)
			}
			err := a.CommitMetaTransaction(tt.args.path, tt.args.crudOperation, tt.args.authTicket, tt.args.lookupHash, fileMeta, status)
			if st, ok := status.(*mocks.StatusCallback); ok {
				st.Test(t)
				st.AssertExpectations(t)
			}
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_StartRepair(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/StartRepair", testcaseName)
		return nil
	}
	var localPath = allocationTestDir + "/alloc/1.txt"

	type args struct {
		localPath, pathToRepair string
		statusCallback          StatusCallback
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			true,
		},
		{
			"Test_Cannot_Get_List_From_Blobbers_Failed",
			args{
				localPath:    localPath,
				pathToRepair: "/1.txt",
			},
			nil,
			true,
		},
		{
			"Test_Repair_Success",
			args{
				localPath:    localPath,
				pathToRepair: "/1.txt",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return func(t *testing.T) {
					willReturnCommitResult(nil)
				}
			},
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

			err := a.StartRepair(tt.args.localPath, tt.args.pathToRepair, tt.args.statusCallback)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestAllocation_CancelRepair(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	tests := []struct {
		name           string
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Failed",
			nil,
			true,
		},
		{
			"Test_Success",
			func(t *testing.T) (teardown func(t *testing.T)) {
				a.repairRequestInProgress = &RepairRequest{}
				return nil
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if tt.additionalMock != nil {
				if teardown := tt.additionalMock(t); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.CancelRepair()
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}
