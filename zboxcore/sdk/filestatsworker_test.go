package sdk

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	filestatsWorkerTestDir = configDir + "/filestatsworker"
)

func TestListRequest_getFileStatsInfoFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, filestatsWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, filestatsWorkerTestDir+"/getFileStatsInfoFromBlobber", testcaseName)
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
			rspCh := make(chan *fileStatsResponse, 1)
			req.wg = &sync.WaitGroup{}
			req.wg.Add(1)
			go req.getFileStatsInfoFromBlobber(req.blobbers[0], 0, rspCh)
			req.wg.Wait()
			resp := <-rspCh
			if tt.wantErr {
				assertion.Error(resp.err, "expected error != nil")
				return
			}
			if !tt.want {
				assertion.Nil(resp.filestats, "expected nullable file stats result")
				return
			}
			assertion.NotNil(resp.filestats, "unexpected nullable file stats result")
		})
	}
}

func TestListRequest_getFileStatsFromBlobbers(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, filestatsWorkerTestDir, blobberMocks)
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, filestatsWorkerTestDir+"/getFileStatsFromBlobbers", testcaseName)
		return nil
	}
	defer cncl()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		want           bool
	}{
		{
			"Test_Error_Getting_File_Stats_From_Blobbers_Failed",
			nil,
			false,
		},
		{
			"Test_Success",
			blobbersResponseMock,
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
			got := req.getFileStatsFromBlobbers()
			if !tt.want {
				for _, blobberMock := range blobberMocks {
					assertion.Emptyf(got[blobberMock.ID], "expected empty value of file stats related to blobber %v", blobberMock.ID)
				}
				return
			}
			assertion.NotNil(got, "unexpected nullable file stats result")
			assertion.Equalf(4, len(got), "expected length of file stats result is %d, but got %v", 4, len(got))
			for _, blobberMock := range blobberMocks {
				assertion.NotEmptyf(got[blobberMock.ID], "unexpected empty value of file stats related to blobber %v", blobberMock.ID)
			}
		})
	}
}
