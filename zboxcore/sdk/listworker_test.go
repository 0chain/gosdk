package sdk

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/stretchr/testify/require"
)

const (
	listWorkerTestDir = configDir + "/listworker"
)

func TestListRequest_getListInfoFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, listWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, listWorkerTestDir+"/getListInfoFromBlobber", testcaseName)
		return nil
	}
	defer cncl()
	var wg sync.WaitGroup
	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		want           bool
		wantErr        bool
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
			false,
			true,
		},
		{
			"Test_HTTP_Response_Failed",
			nil,
			false,
			false,
		},
		{
			"Test_Error_HTTP_Response_Not_JSON_Format",
			blobbersResponseMock,
			false,
			false,
		},
		{
			"Test_Success",
			blobbersResponseMock,
			true,
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
			req := &ListRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				ctx:            context.Background(),
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				wg: func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			}
			rspCh := make(chan *listResponse, 1)
			req.getListInfoFromBlobber(req.blobbers[0], 0, rspCh)
			resp := <-rspCh
			var expectedResult *fileref.Ref
			parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__Test_Success.json", listWorkerTestDir, "getListInfoFromBlobber"), &expectedResult)
			if tt.wantErr {
				require.Error(resp.err, "expected error != nil")
				return
			}
			if !tt.want {
				require.NotEqual(expectedResult, resp.ref)
				return
			}
			require.EqualValues(expectedResult, resp.ref)
		})
	}
}

func TestListRequest_GetListFromBlobbers(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, listWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, listWorkerTestDir+"/GetListFromBlobbers", testcaseName)
		return nil
	}
	defer cncl()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		want           bool
		wantFunc       func(require *require.Assertions, req *ListRequest)
	}{
		{
			"Test_Error_Get_List_File_From_Blobbers_Failed",
			nil,
			false,
			func(require *require.Assertions, req *ListRequest) {
				require.NotNil(req)
				require.Equal(float32(0), req.consensus)
			},
		},
		{
			"Test_Success",
			blobbersResponseMock,
			true,
			func(require *require.Assertions, req *ListRequest) {
				require.NotNil(req)
				require.Equal(float32(4), req.consensus)
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
			req := &ListRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				ctx:            context.Background(),
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
			}
			got := req.GetListFromBlobbers()
			var expectedResult *ListResult
			parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__Test_Success.json", listWorkerTestDir, "GetListFromBlobbers"), &expectedResult)
			if tt.want {
				require.EqualValues(expectedResult, got)
				return
			}
			tt.wantFunc(require, req)
		})
	}
}
