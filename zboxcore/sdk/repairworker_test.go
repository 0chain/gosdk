package sdk

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

const (
	repairWorkerTestDir = configDir + "/repairworker"
)

func TestRepairRequest_getLocalPath(t *testing.T) {
	var (
		rootPath = repairWorkerTestDir + "/alloc"
		filePath = "/1.txt"
	)
	r := &RepairRequest{localRootPath: rootPath}
	assert.Equal(t, rootPath + filePath, r.getLocalPath(&ListResult{Path: filePath}))
}

func Test_checkFileExists(t *testing.T) {
	var rootPath = repairWorkerTestDir + "/alloc"
	t.Run("Test_File_Not_Exists", func(t *testing.T) {
		assert.False(t, checkFileExists(rootPath + "/x.txt"))
	})
	t.Run("Test_Is_File_Exists", func(t *testing.T) {
		assert.True(t, checkFileExists(rootPath + "/1.txt"))
	})
	t.Run("Test_Is_Dir", func(t *testing.T) {
		assert.False(t, checkFileExists(rootPath))
	})
}

func TestRepairRequest_repairFile(t *testing.T) {
	_, _, blobberMocks, cls := setupMockInitStorageSDK(t, configDir, 4)
	defer cls()
	a := setupMockAllocation(t, repairWorkerTestDir, blobberMocks)
	var (
		blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
			setupBlobberMockResponses(t, blobberMocks, repairWorkerTestDir+"/repairFile", testcaseName)
			return nil
		}

	)
	type fields struct {
		listDir           *ListResult
		isRepairCanceled  bool
		localRootPath     string
		statusCB          StatusCallback
		completedCallback func()
		filesRepaired     int
		wg                *sync.WaitGroup
	}
	tests := []struct {
		name   string
		fields fields
		file *ListResult
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantFn func(assertion *assert.Assertions)
	}{
		{
			"Test_Cancel_Repair_Process",
			fields{isRepairCanceled: true},
			nil,
			nil,
			nil,
		},
		{
			"Test_Error_Remote_File_Not_Found",
			fields{},
			&ListResult{Path: "/1.txt"},
			blobbersResponseMock,
			nil,
		},
		//{
		//	"Test_Not_Enough_Minimum_Found_Then_Success_To_Delete_The_File",
		//	fields{},
		//	&ListResult{Path: "/1.txt"},
		//	blobbersResponseMock,
		//	nil,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if m := tt.additionalMock; m != nil {
				if teardown := m(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			r := &RepairRequest{
				listDir:           tt.fields.listDir,
				isRepairCanceled:  tt.fields.isRepairCanceled,
				localRootPath:     tt.fields.localRootPath,
				statusCB:          tt.fields.statusCB,
				completedCallback: tt.fields.completedCallback,
				filesRepaired:     tt.fields.filesRepaired,
				wg:                tt.fields.wg,
			}
			r.repairFile(a, tt.file)
			if tt.wantFn != nil {
				tt.wantFn(assertion)
			}
		})
	}
}
