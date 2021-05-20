package sdk

import (
	"context"
	"testing"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/require"
)

const (
	renameWorkerTestDir = configDir + "/renameworker"
)

func TestRenameRequest_getObjectTreeFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, renameWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, renameWorkerTestDir+"/getObjectTreeFromBlobber", testcaseName)
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
			blobbersResponseMock,
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
			ar := &RenameRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				ctx:          context.Background(),
				renameMask:   0,
				connectionID: zboxutil.NewConnectionId(),
			}
			_, err := ar.getObjectTreeFromBlobber(a.Blobbers[0])
			if tt.wantErr {
				require.Error(err, "expected error != nil")
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
		})
	}
}

func TestRenameRequest_renameBlobberObject(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, renameWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, renameWorkerTestDir+"/renameBlobberObject", testcaseName)
		return nil
	}
	defer cncl()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
		wantFunc       func(require *require.Assertions, ar *RenameRequest)
	}{
		{
			"Test_Error_New_HTTP_Failed",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				url := a.Blobbers[0].Baseurl
				a.Blobbers[0].Baseurl = string([]byte{0x7f, 0, 0})
				return func(t *testing.T) {
					a.Blobbers[0].Baseurl = url
				}
			},
			true,
			nil,
		},
		{
			"Test_Error_Get_Object_Tree_From_Blobber_Failed",
			nil,
			true,
			nil,
		},
		{
			"Test_Rename_Blobber_Object_Failed",
			blobbersResponseMock,
			false,
			func(require *require.Assertions, ar *RenameRequest) {
				require.NotNil(ar)
				require.Equal(uint32(0), ar.renameMask)
				require.Equal(float32(0), ar.consensus)
			},
		},
		{
			"Test_Rename_Blobber_Object_Success",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			false,
			func(require *require.Assertions, ar *RenameRequest) {
				require.NotNil(ar)
				require.Equal(uint32(1), ar.renameMask)
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
			ar := &RenameRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				ctx:          context.Background(),
				renameMask:   0,
				connectionID: zboxutil.NewConnectionId(),
			}
			_, err := ar.renameBlobberObject(a.Blobbers[0], 0)
			if tt.wantErr {
				require.Error(err, "expected error != nil")
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
			if tt.wantFunc != nil {
				tt.wantFunc(require, ar)
			}
		})
	}
}

func TestRenameRequest_ProcessRename(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, renameWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, renameWorkerTestDir+"/ProcessRename", testcaseName)
		return nil
	}
	defer cncl()
	tests := []struct {
		name            string
		additionalMock  func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr         bool
		wantErrContains string
		wantFunc        func(require *require.Assertions, ar *RenameRequest)
	}{
		{
			name: "Test_All_Blobber_Rename_Success",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr: false,
			wantFunc: func(require *require.Assertions, ar *RenameRequest) {
				require.NotNil(ar)
				require.Equal(uint32(15), ar.renameMask)
				require.Equal(float32(4), ar.consensus)
			},
		},
		{
			name: "Test_Blobber_Index_0_Error_On_Rename_Success",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr: false,
			wantFunc: func(require *require.Assertions, ar *RenameRequest) {
				require.NotNil(ar)
				require.Equal(uint32(14), ar.renameMask)
				require.Equal(float32(3), ar.consensus)
			},
		},
		{
			name: "Test_Blobber_Index_0_2_Error_On_Rename_Failure",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr:         true,
			wantErrContains: "Rename failed",
		},
		{
			name: "Test_All_Blobber_Error_On_Rename_Failure",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr:         true,
			wantErrContains: "Rename failed",
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
			ar := &RenameRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				ctx:          context.Background(),
				renameMask:   0,
				connectionID: zboxutil.NewConnectionId(),
			}
			err := ar.ProcessRename()
			if tt.wantErr {
				require.Error(err, "expected error != nil")
				require.Contains(err.Error(), tt.wantErrContains, "expected error contains '%s'", tt.wantErrContains)
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
			if tt.wantFunc != nil {
				tt.wantFunc(require, ar)
			}
		})
	}
}
