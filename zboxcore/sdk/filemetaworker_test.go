package sdk

import (
	"context"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
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
		setupBlobberMockResponses(t, blobberMocks, filemetaWorkerTestDir+"/getFileMetaInfoFromBlobber", testcaseName)
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
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				authToken: &marker.AuthTicket{
					ClientID: "",
					OwnerID:  "",
				},
				ctx: context.Background(),
			}
			rspCh := make(chan *fileMetaResponse, 1)
			req.wg = &sync.WaitGroup{}
			req.wg.Add(1)
			go req.getFileMetaInfoFromBlobber(req.blobbers[0], 0, rspCh)
			req.wg.Wait()
			resp := <-rspCh
			if !tt.want {
				assertion.Nil(resp.fileref)
			}
			if tt.wantErr {
				assertion.Error(resp.err, "expected error != nil")
			}
		})
	}
}

func TestListRequest_getFileConsensusFromBlobbers(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
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
		wantFound      uint64
		wantErr        bool
	}{
		{
			"Test_All_Success",
			blobbersResponseMock,
			0xf,
			false,
		},
		{
			"Test_Index_3_Error",
			blobbersResponseMock,
			0x7,
			false,
		},
		{
			"Test_File_Consensus_Not_Found",
			blobbersResponseMock,
			0x0,
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

			found, _, _ := req.getFileConsensusFromBlobbers()
			assertion.Equal(zboxutil.NewUint128(tt.wantFound), found, "found value must be same")
		})
	}
}
