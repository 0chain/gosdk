package sdk

import (
	"context"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/stretchr/testify/assert"
)

const (
	filemetaWorkerTestDir = configDir + "/filemetaworker"
)

func TestListRequest_getFileMetaInfoFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, filemetaWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, filemetaWorkerTestDir+"/getFileConsensusFromBlobbers", testcaseName)
		return nil
	}
	defer cncl()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantFunc       func(assertions *assert.Assertions, ar *ListRequest)
	}{
		{
			name: "Test_Success",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantFunc: func(assertions *assert.Assertions, ar *ListRequest) {
				assertions.NotNil(ar)
				assertions.Equal(float32(0), ar.consensus)
			},
		},
		// {
		// 	name: "Test_All_Blobber_Delete_False",
		// 	additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
		// 		blobbersResponseMock(t, testCaseName)
		// 		willReturnCommitResult(&CommitResult{Success: true})
		// 		return nil
		// 	},
		// 	wantFunc: func(assertions *assert.Assertions, ar *DeleteRequest) {
		// 		assertions.NotNil(ar)
		// 		assertions.Equal(uint32(0), ar.deleteMask)
		// 		assertions.Equal(float32(0), ar.consensus)
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			req := &ListRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				authToken: &marker.AuthTicket{
					ClientID: "",
					OwnerID: "",
				},
				ctx: context.Background(),
			}
			rspCh := make(chan *fileMetaResponse, 1)

			// objectTreeRefs := make([]fileref.RefEntity, 1)

			req.wg = &sync.WaitGroup{}
			req.wg.Add(1)
			go req.getFileMetaInfoFromBlobber(req.blobbers[0], 0, rspCh)
			req.wg.Wait()

			if tt.wantFunc != nil {
				tt.wantFunc(assertion, req)
			}
		})
	}
}

func TestListRequest_getFileMetaFromBlobbers(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, filemetaWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, filemetaWorkerTestDir+"/getFileMetaFromBlobbers", testcaseName)
		return nil
	}
	defer cncl()
	tests := []struct {
		name            string
		additionalMock  func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantFunc        func(assertions *assert.Assertions, ar *ListRequest)
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "Test_Success",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr: false,
			wantFunc: func(assertions *assert.Assertions, ar *ListRequest) {
				assertions.NotNil(ar)
				assertions.Equal(float32(0), ar.consensus)
			},
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
			req := &ListRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				ctx:          context.Background(),
				// connectionID: zboxutil.NewConnectionId(),
			}

			req.getFileMetaFromBlobbers()
			if tt.wantFunc != nil {
				tt.wantFunc(assertion, req)
			}
		})
	}
}
func TestListRequest_getFileConsensusFromBlobbers(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, filemetaWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, filemetaWorkerTestDir+"/getFileMetaInfoFromBlobber", testcaseName)
		return nil
	}
	defer cncl()
	tests := []struct {
		name            string
		additionalMock  func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantFunc        func(assertions *assert.Assertions, ar *ListRequest)
		wantErr         bool
		wantErrContains string
		remotefilepath  string
	}{
		{
			name: "Test_Success",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantFunc: func(assertions *assert.Assertions, ar *ListRequest) {
				assertions.NotNil(ar)
				assertions.Equal(float32(1), ar.consensus)
			},
			remotefilepath: "1.txt",
		},
		{
			name: "Test_False",
			additionalMock: func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				blobbersResponseMock(t, testCaseName)
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantFunc: func(assertions *assert.Assertions, ar *ListRequest) {
				assertions.NotNil(ar)
				assertions.Equal(float32(0), ar.consensus)
			},
			remotefilepath: "1.txt",
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
			req := &ListRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: tt.remotefilepath,
				ctx:            context.Background(),
			}
	
			req.wg = &sync.WaitGroup{}
			req.getFileConsensusFromBlobbers()
			if tt.wantFunc != nil {
				tt.wantFunc(assertion, req)
			}
		})
	}
}
