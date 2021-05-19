package sdk

import (
	"context"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/require"
)

const (
	deleteWorkerTestDir = configDir + "/deleteworker"
)

func TestDeleteRequest_ProcessDelete(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, deleteWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, deleteWorkerTestDir+"/ProcessDelete", testcaseName)
		return nil
	}
	defer cncl()
	tests := []struct {
		name            string
		additionalMock  func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantFunc        func(require *require.Assertions, ar *DeleteRequest)
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "Test_All_Blobber_Delete_Attribute_Success",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr: false,
			wantFunc: func(require *require.Assertions, ar *DeleteRequest) {
				require.NotNil(ar)
				require.Equal(uint32(15), ar.deleteMask)
				require.Equal(float32(4), ar.consensus)
			},
		},
		{
			name: "Test_Blobber_Index_0_Error_On_Delete_Attribute_Success",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr: false,
			wantFunc: func(require *require.Assertions, ar *DeleteRequest) {
				require.NotNil(ar)
				require.Equal(uint32(14), ar.deleteMask)
				require.Equal(float32(3), ar.consensus)
			},
		},
		{
			name: "Test_Blobber_Index_0_2_Error_On_Delete_Attribute_Failure",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr:         true,
			wantErrContains: "Delete failed",
		},
		{
			name: "Test_All_Blobber_Error_On_Delete_Attribute_Failure",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr:         true,
			wantErrContains: "Delete failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			req := &DeleteRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				ctx:          context.Background(),
				connectionID: zboxutil.NewConnectionId(),
			}

			err := req.ProcessDelete()

			if tt.wantFunc != nil {
				tt.wantFunc(require, req)
			}
			if tt.wantErr {
				require.Error(err, "expected error != nil")
				require.Contains(err.Error(), tt.wantErrContains, "expected error contains '%s'", tt.wantErrContains)
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
			if tt.wantFunc != nil {
				tt.wantFunc(require, req)
			}
		})
	}
}

func TestDeleteRequest_deleteBlobberFile(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, deleteWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, deleteWorkerTestDir+"/deleteBlobberFile", testcaseName)
		return nil
	}
	defer cncl()
	var wg sync.WaitGroup
	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantFunc       func(require *require.Assertions, ar *DeleteRequest)
	}{
		{
			"Test_Delete_Blobber_File_Failed",
			blobbersResponseMock,
			func(require *require.Assertions, ar *DeleteRequest) {
				require.NotNil(ar)
				require.Equal(uint32(0), ar.deleteMask)
				require.Equal(float32(0), ar.consensus)
			},
		},
		{
			"Test_Delete_Blobber_File_Success",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			func(require *require.Assertions, ar *DeleteRequest) {
				require.NotNil(ar)
				require.Equal(uint32(1), ar.deleteMask)
				require.Equal(float32(1), ar.consensus)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			req := &DeleteRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				ctx:          context.Background(),
				connectionID: zboxutil.NewConnectionId(),
				wg:           func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			}
			objectTreeRefs := make([]fileref.RefEntity, 1)
			refEntity, _ := req.getObjectTreeFromBlobber(req.blobbers[0])
			objectTreeRefs[0] = refEntity
			req.deleteBlobberFile(req.blobbers[0], 0, objectTreeRefs[0])
			if tt.wantFunc != nil {
				tt.wantFunc(require, req)
			}
		})
	}
}

func TestDeleteRequest_getObjectTreeFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, deleteWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, deleteWorkerTestDir+"/getObjectTreeFromBlobber", testcaseName)
		return nil
	}
	defer cncl()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Get_Object_Tree_From_Blobber_Failed",
			nil,
			true,
		},
		{
			"Test_Get_Object_Tree_From_Blobber_Success",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			req := &DeleteRequest{
				allocationID: a.ID,
				allocationTx: a.Tx,
				blobbers:     a.Blobbers,
				ctx:          context.Background(),
				connectionID: zboxutil.NewConnectionId(),
			}
			objectTreeRefs := make([]fileref.RefEntity, 1)
			refEntity, _ := req.getObjectTreeFromBlobber(req.blobbers[0])
			objectTreeRefs[0] = refEntity
			_, err := req.getObjectTreeFromBlobber(req.blobbers[0])
			if tt.wantErr {
				require.Error(err, "expected error != nil")
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
		})
	}
}
