package sdk

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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
			"Test_Failed",
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
				ctx:            context.Background(),
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
			}
			rspCh := make(chan *listResponse, 1)
			req.wg = &sync.WaitGroup{}
			req.wg.Add(1)
			req.getListInfoFromBlobber(req.blobbers[0], 0, rspCh)
			req.wg.Wait()
			resp := <-rspCh
			if tt.wantErr {
				assertion.Error(resp.err, "expected error != nil")
				return
			}
			if !tt.want {
				assertion.Empty(resp.ref.Type, "expected nullable type result")
				return
			}
			assertion.NotEmpty(resp.ref.Type, "unexpected nullable type result")
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
	}{
		// {
		// 	"Test_Error_Get_List_File_From_Blobbers_Failed",
		// 	nil,
		// 	false,
		// },
		{
			"Test_Success",
			blobbersResponseMock,
			true,
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
				assertion.EqualValues(expectedResult, got)
				return
			}
			assertion.NotEqual(expectedResult, got)
		})
	}
}
