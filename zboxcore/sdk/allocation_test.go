package sdk

import (
	"encoding/json"
	"fmt"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/sdk/mock"
	"github.com/stretchr/testify/assert"
	tm "github.com/stretchr/testify/mock"
	"strings"
	"testing"
)

const (
	allocationTestDir = configDir + "/allocation"
)

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

func TestAllocation_GetStats(t *testing.T) {
	stats := &AllocationStats{}
	a := &Allocation{
		Stats: stats,
	}
	got := a.GetStats()
	assert.New(t).Same(stats, got)
}

func TestAllocation_GetBlobberStats(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()

	blobbers := []*blockchain.StorageNode{}
	for _, blobberMock := range blobberMocks {
		blobbers = append(blobbers, &blockchain.StorageNode{
			ID:      blobberMock.ID,
			Baseurl: blobberMock.URL,
		})
	}

	// mock init sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	type fields struct {
		Blobbers []*blockchain.StorageNode
	}

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
			setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/"+"GetBlobberStats", tt.name, blobberResponseParamCheck)
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
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/alloc/1.txt"
	var status = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, "/1.txt", tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}
	assertion := assert.New(t)
	err := a.UpdateFile(localPath, "/", fileref.Attributes{}, status(t))
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_UploadFile(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/alloc/1.txt"
	var status = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, "/1.txt", tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}
	assertion := assert.New(t)
	err := a.UploadFile(localPath, "/", fileref.Attributes{}, status(t))
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_RepairFile(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/alloc/1.txt"
	type args struct {
		localPath  string
		remotePath string
		status     func(t *testing.T) StatusCallback
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
				status: func(t *testing.T) StatusCallback {
					scm := &mock.StatusCallback{}
					scm.Test(t)
					scm.On("Started", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
					scm.On("InProgress", a.ID, "/1.txt", tm.Anything, tm.Anything, tm.Anything).Maybe()
					scm.On("Error", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
					scm.On("Completed", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
					scm.AssertExpectations(t)
					return scm
				},
			},
			false,
		},
		{
			"Test_Repair_Not_Required_Failed",
			args{
				localPath:  localPath,
				remotePath: "/",
				status: func(t *testing.T) StatusCallback {
					scm := &mock.StatusCallback{}
					scm.Test(t)
					scm.On("Started", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
					scm.On("InProgress", a.ID, "/1.txt", tm.Anything, tm.Anything, tm.Anything).Maybe()
					scm.On("Error", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
					scm.On("Completed", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
					scm.AssertExpectations(t)
					return scm
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/"+"RepairFile", tt.name)
			err := a.RepairFile(tt.args.localPath, tt.args.remotePath, tt.args.status(t))
			if tt.wantErr {
				assertion.Errorf(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_UpdateFileWithThumbnail(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/alloc/1.txt"
	var thumbnailPath = allocationTestDir + "/thumbnail_alloc"
	type args struct {
		localPath, remotePath, thumbnailPath string
		status                               func(t *testing.T) StatusCallback
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
				status: func(t *testing.T) StatusCallback {
					scm := &mock.StatusCallback{}
					scm.Test(t)
					scm.On("Started", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
					scm.On("InProgress", a.ID, "/1.txt", tm.Anything, tm.Anything, tm.Anything).Maybe()
					scm.On("Error", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
					scm.On("Completed", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
					scm.AssertExpectations(t)
					return scm
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			err := a.UpdateFileWithThumbnail(tt.args.localPath, tt.args.remotePath, tt.args.thumbnailPath, fileref.Attributes{}, tt.args.status(t))
			if tt.wantErr {
				assertion.Errorf(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_UploadFileWithThumbnail(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/alloc/1.txt"
	var thumbnailPath = allocationTestDir + "/thumbnail_alloc"
	var status = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, "/1.txt", tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}

	assertion := assert.New(t)
	err := a.UploadFileWithThumbnail(localPath, "/", thumbnailPath, fileref.Attributes{}, status(t))
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUpdateFile(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/alloc/1.txt"
	var status = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, "/1.txt", tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}
	assertion := assert.New(t)
	err := a.EncryptAndUpdateFile(localPath, "/", fileref.Attributes{}, status(t))
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUploadFile(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/alloc/1.txt"
	var status = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, "/1.txt", tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}
	assertion := assert.New(t)
	err := a.EncryptAndUploadFile(localPath, "/", fileref.Attributes{}, status(t))
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUpdateFileWithThumbnail(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/alloc/1.txt"
	var thumbnailPath = allocationTestDir + "/thumbnail_alloc"
	var status = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, "/1.txt", tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}
	assertion := assert.New(t)
	err := a.EncryptAndUpdateFileWithThumbnail(localPath, "/", thumbnailPath, fileref.Attributes{}, status(t))
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUploadFileWithThumbnail(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/alloc/1.txt"
	var thumbnailPath = allocationTestDir + "/thumbnail_alloc"
	var status = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, "/1.txt", tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}
	assertion := assert.New(t)
	err := a.EncryptAndUploadFileWithThumbnail(localPath, "/", thumbnailPath, fileref.Attributes{}, status(t))
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_uploadOrUpdateFile(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/alloc/1.txt"
	var status = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, "/1.txt", tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, "/1.txt", tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}
	type args struct {
		localPath     string
		remotePath    string
		status        func(t *testing.T) StatusCallback
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
				status:        status,
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
				attrs:         fileref.Attributes{},
			},
			true,
		},
		{
			"Test_Local_File_Error_Failed",
			nil,
			args{
				localPath:     "local_file_error",
				remotePath:    "/",
				status:        status,
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
				status:        status,
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
				status:        status,
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
				status:        status,
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
				status:        status,
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      true,
				attrs:         fileref.Attributes{},
			},
			true,
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
			err := a.uploadOrUpdateFile(tt.args.localPath, tt.args.remotePath, tt.args.status(t), tt.args.isUpdate, tt.args.thumbnailPath, tt.args.encryption, tt.args.isRepair, tt.args.attrs)
			if tt.wantErr {
				assertion.Errorf(err, "Expected error != nil")
				return
			}
			assertion.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_RepairRequired(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var blobberMockFn = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, fmt.Sprintf("%v/%v", allocationTestDir, "RepairRequired"), testcaseName, blobberResponseFormBodyCheck)
		return func(t *testing.T) {
			for _, blobberMock := range blobberMocks {
				blobberMock.ResetHandler(t)
			}
		}
	}
	tests := []struct {
		name                          string
		additionalSetupFn             func(t *testing.T, testcaseName string) (teardown func(t *testing.T))
		remotePath                    string
		wantFound                     uint32
		wantMatchesConsensus, wantErr bool
	}{
		{
			"Test_Not_Repair_Required_Success",
			blobberMockFn,
			"/x.txt",
			0xf,
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
			false, true,
		},
		{
			"Test_Repair_Required_Success",
			blobberMockFn,
			"/",
			0x7,
			true, false,
		},
		{
			"Test_Remote_File_Not_Found_Failed",
			blobberMockFn,
			"/x.txt",
			0x0,
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
			var wantFileRef *fileref.FileRef
			parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v.json", allocationTestDir, "RepairRequired", tt.name), &wantFileRef)
			found, matchesConsensus, fileRef, err := a.RepairRequired(tt.remotePath)
			assertion.Equal(tt.wantFound, found, "found value must be same")
			if tt.wantMatchesConsensus {
				assertion.True(tt.wantMatchesConsensus, matchesConsensus)
			} else {
				assertion.False(tt.wantMatchesConsensus, matchesConsensus)
			}
			if tt.wantErr {
				assertion.Errorf(err, "Expected error != nil")
			} else {
				assertion.NoErrorf(err, "Unexpected error %v", err)
			}

			assertion.EqualValues(wantFileRef, fileRef)
		})
	}
}

func TestAllocation_DownloadFile(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/DownloadFile"
	var status = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, tm.Anything, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}
	assertion := assert.New(t)
	err := a.DownloadFile(localPath, "/", status(t))
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_DownloadFileByBlock(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/DownloadFileByBlock"
	var status = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, tm.Anything, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}
	assertion := assert.New(t)
	err := a.DownloadFileByBlock(localPath, "/", 1, 0, numBlockDownloads, status(t))
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_DownloadThumbnail(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var localPath = allocationTestDir + "/DownloadThumbnail"

	var status = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, tm.Anything, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}
	assertion := assert.New(t)
	err := a.DownloadThumbnail(localPath, "/", status(t))
	assertion.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_downloadFile(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()

	var blobberMocks = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobberMocks = append(blobberMocks, blobberIdx)
	}

	defer func() {
		for _, blobberMock := range blobberMocks {
			blobberMock.Close(t)
		}
	}()
	// setup mock sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// setup mock allocation
	a := setupMockAllocation(t, allocationTestDir, blobberMocks)

	var statusMock = func(t *testing.T) StatusCallback {
		scm := &mock.StatusCallback{}
		scm.Test(t)
		scm.On("Started", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("InProgress", a.ID, tm.Anything, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Error", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.On("Completed", a.ID, tm.Anything, tm.Anything, tm.Anything).Maybe()
		scm.AssertExpectations(t)
		return scm
	}

	var localPath = allocationTestDir + "/downloadFile/alloc"
	var remotePath = "/"

	type args struct {
		localPath, remotePath string
		contentMode           string
		startBlock, endBlock  int64
		numBlocks             int
		statusCallback        func(t *testing.T) StatusCallback
	}

	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Uninitialized_Failed",
			args{
				localPath, remotePath,
				DOWNLOAD_CONTENT_FULL,
				1, 0,
				numBlockDownloads,
				statusMock,
			},
			func(t *testing.T) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) { a.initialized = true }
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			setupBlobberMockResponses(t, blobberMocks, allocationTestDir+"/downloadFile", tt.name)
			if m := tt.additionalMock; m != nil {
				if teardown := m(t); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.downloadFile(tt.args.localPath, tt.args.remotePath, tt.args.contentMode, tt.args.startBlock, tt.args.endBlock, tt.args.numBlocks, tt.args.statusCallback(t))
			if tt.wantErr {
				assertion.Error(err, "Expected error != nil")
				return
			}
			assertion.NoErrorf(err, "Unexpected error: %v", err)
		})
	}
}
