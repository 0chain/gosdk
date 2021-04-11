package sdk

import (
	"context"
	"errors"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/sdk/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"math/rand"
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
	assert.Equal(t, rootPath+filePath, r.getLocalPath(&ListResult{Path: filePath}))
}

func Test_checkFileExists(t *testing.T) {
	var rootPath = repairWorkerTestDir + "/alloc"
	t.Run("Test_File_Not_Exists", func(t *testing.T) {
		assert.False(t, checkFileExists(rootPath+"/x.txt"))
	})
	t.Run("Test_Is_File_Exists", func(t *testing.T) {
		assert.True(t, checkFileExists(rootPath+"/file_existing.txt"))
	})
	t.Run("Test_Is_Dir", func(t *testing.T) {
		assert.False(t, checkFileExists(rootPath))
	})
}

func TestRepairRequest_repairFile(t *testing.T) {
	_, _, blobberMocks, cls := setupMockInitStorageSDK(t, configDir, 4)
	defer cls()
	a, cncl := setupMockAllocation(t, repairWorkerTestDir, blobberMocks)
	defer cncl()

	var (
		blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
			setupBlobberMockResponses(t, blobberMocks, repairWorkerTestDir+"/repairFile", testcaseName)
			return nil
		}
		rootPath = repairWorkerTestDir + "/repairFile/alloc"
		filePath = "/1.txt"
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
		name           string
		fields         fields
		file           *ListResult
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
	}{
		{
			"Test_Cancel_Repair_Process",
			fields{isRepairCanceled: true},
			nil,
			nil,
		},
		{
			"Test_Error_Remote_File_Not_Found",
			fields{},
			&ListResult{Path: filePath},
			blobbersResponseMock,
		},
		//{
		//	"Test_Not_Enough_Minimum_Found_Then_Success_To_Delete_The_File",
		//	fields{},
		//	&ListResult{Path: "/1.txt"},
		//	blobbersResponseMock,
		//},
		{
			"Test_Success",
			fields{
				localRootPath: rootPath,
				statusCB: func() StatusCallback {
					scm := &mocks.StatusCallback{}
					scm.On("Completed", a.ID, rootPath+filePath, "1.txt", mock.AnythingOfType("string"), 3, mock.AnythingOfType("int")).Twice()
					return scm
				}(),
			},
			&ListResult{Path: filePath},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				wf := &writeFile{
					FileName: rootPath + "/1.txt",
					IsDir:    false,
					Content:  []byte("abc"),
				}
				willDownloadSuccessFiles(wf)
				willReturnCommitResult(&CommitResult{Success: true})
				return func(t *testing.T) {
					willReturnCommitResult(nil)
					deleteFiles(wf)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			if st, ok := tt.fields.statusCB.(*mocks.StatusCallback); ok {
				st.Test(t)
				st.AssertExpectations(t)
			}
		})
	}
}

func TestRepairRequest_iterateDir(t *testing.T) {
	_, _, blobberMocks, cls := setupMockInitStorageSDK(t, configDir, 4)
	defer cls()
	a, cncl := setupMockAllocation(t, repairWorkerTestDir, blobberMocks)
	defer cncl()

	var (
		blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
			setupBlobberMockResponses(t, blobberMocks, repairWorkerTestDir+"/iterateDir", testcaseName)
			return nil
		}
		rootPath = repairWorkerTestDir + "/iterateDir/alloc"
		filePath = "/1.txt"
	)

	type fields struct {
		isRepairCanceled  bool
		localRootPath     string
		statusCB          StatusCallback
		completedCallback func()
	}
	tests := []struct {
		name           string
		fields         fields
		listRs         *ListResult
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
	}{
		{
			"Test_Repair_With_List_Result_File_Type",
			fields{
				// quick pass repair file's call
				isRepairCanceled: true,
			},
			&ListResult{
				Type: fileref.FILE,
				Path: filePath,
			},
			nil,
		},
		{
			"Test_Repair_With_List_Result_Directory_Type_And_Zero_File_Child",
			fields{
				localRootPath: rootPath,
			},
			&ListResult{
				Type: fileref.DIRECTORY,
				Path: "/",
			},
			blobbersResponseMock,
		},
		{
			"Test_Repair_With_List_Result_Directory_Type_And_File_Child",
			fields{
				localRootPath: rootPath,
			},
			&ListResult{
				Type: fileref.DIRECTORY,
				Path: "/",
				Children: []*ListResult{
					{
						Type: fileref.FILE,
						Path: "/1.txt",
					},
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if m := tt.additionalMock; m != nil {
				if teardown := m(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			var wg sync.WaitGroup
			r := &RepairRequest{
				isRepairCanceled:  tt.fields.isRepairCanceled,
				localRootPath:     tt.fields.localRootPath,
				statusCB:          tt.fields.statusCB,
				completedCallback: tt.fields.completedCallback,
				wg:                &wg,
			}
			r.iterateDir(a, tt.listRs)
			if st, ok := tt.fields.statusCB.(*mocks.StatusCallback); ok {
				st.Test(t)
				st.AssertExpectations(t)
			}
		})
	}
}

func TestRepairRequest_processRepair(t *testing.T) {
	_, _, blobberMocks, cls := setupMockInitStorageSDK(t, configDir, 4)
	defer cls()
	a, cncl := setupMockAllocation(t, repairWorkerTestDir, blobberMocks)
	defer cncl()

	var (
		rootPath = repairWorkerTestDir + "/processRepair/alloc"
		filePath = "/1.txt"
	)

	var wg sync.WaitGroup
	var isCompletedCallbackCalled bool
	scm := &mocks.StatusCallback{}
	scm.On("RepairCompleted", 0).Once()
	r := &RepairRequest{
		listDir:       &ListResult{Path: filePath},
		localRootPath: rootPath,
		statusCB:      scm,
		completedCallback: func() {
			isCompletedCallbackCalled = true
		},
		wg: &wg,
	}
	r.processRepair(context.Background(), a)
	assert.True(t, isCompletedCallbackCalled)
	scm.Test(t)
	scm.AssertExpectations(t)
}

func TestRepairStatusCB_CommitMetaCompleted(t *testing.T) {
	err := errors.New("test_error")
	scm := &mocks.StatusCallback{}
	scm.On("CommitMetaCompleted", "test_request", "test_response", err).Once()
	cb := &RepairStatusCB{
		wg:       &sync.WaitGroup{},
		success:  false,
		err:      nil,
		statusCB: scm,
	}

	cb.CommitMetaCompleted("test_request", "test_response", err)
	assert.NoError(t, cb.err, "unexpected error but got %v", cb.err)
	scm.Test(t)
	scm.AssertExpectations(t)
}

func TestRepairStatusCB_Completed(t *testing.T) {
	scm := &mocks.StatusCallback{}
	scm.On("Completed", "test_allocation_id", "test_file_path", "test_file_name", "text/plain", 4, OpUpload).Once()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	cb := &RepairStatusCB{
		wg:       wg,
		success:  false,
		err:      nil,
		statusCB: scm,
	}

	cb.Completed("test_allocation_id", "test_file_path", "test_file_name", "text/plain", 4, OpUpload)
	wg.Wait()
	assert.NoErrorf(t, cb.err, "unexpected error but got %v", cb.err)
	assert.True(t, cb.success)
	scm.Test(t)
	scm.AssertExpectations(t)
}

func TestRepairStatusCB_Error(t *testing.T) {
	err := errors.New("test_error")
	scm := &mocks.StatusCallback{}
	scm.On("Error", "test_allocation_id", "test_file_path", OpUpload, err).Once()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	cb := &RepairStatusCB{
		wg:       wg,
		success:  false,
		err:      nil,
		statusCB: scm,
	}

	cb.Error("test_allocation_id", "test_file_path", OpUpload, err)
	wg.Wait()
	assert.Equalf(t, err, cb.err, "expected error %v but got %v", err, cb.err)
	assert.False(t, cb.success)
	scm.Test(t)
	scm.AssertExpectations(t)
}

func TestRepairStatusCB_InProgress(t *testing.T) {
	scm := &mocks.StatusCallback{}
	scm.On("InProgress", "test_allocation_id", "test_file_path", OpUpload, 3, []byte{97, 98, 99}).Once()
	wg := &sync.WaitGroup{}
	cb := &RepairStatusCB{
		wg:       wg,
		success:  false,
		err:      nil,
		statusCB: scm,
	}

	cb.InProgress("test_allocation_id", "test_file_path", OpUpload, 3, []byte{97, 98, 99})
	assert.NoErrorf(t, cb.err, "unexpected error but got %v", cb.err)
	assert.False(t, cb.success)
	scm.Test(t)
	scm.AssertExpectations(t)
}

func TestRepairStatusCB_RepairCompleted(t *testing.T) {
	randNum := rand.Int()
	scm := &mocks.StatusCallback{}
	scm.On("RepairCompleted", randNum).Once()
	wg := &sync.WaitGroup{}
	cb := &RepairStatusCB{
		wg:       wg,
		success:  false,
		err:      nil,
		statusCB: scm,
	}

	cb.RepairCompleted(randNum)
	assert.NoErrorf(t, cb.err, "unexpected error but got %v", cb.err)
	assert.False(t, cb.success)
	scm.Test(t)
	scm.AssertExpectations(t)
}

func TestRepairStatusCB_Started(t *testing.T) {
	scm := &mocks.StatusCallback{}
	scm.On("Started", "test_allocation_id", "test_file_path", OpUpload, 3).Once()
	wg := &sync.WaitGroup{}
	cb := &RepairStatusCB{
		wg:       wg,
		success:  false,
		err:      nil,
		statusCB: scm,
	}

	cb.Started("test_allocation_id", "test_file_path", OpUpload, 3)
	assert.NoErrorf(t, cb.err, "unexpected error but got %v", cb.err)
	assert.False(t, cb.success)
	scm.Test(t)
	scm.AssertExpectations(t)
}
