package sdk

import (
	"context"
	"fmt"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

const commonTestDir = configDir + "/common"

func Test_getObjectTreeFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, commonTestDir+"/getObjectTreeFromBlobber", testcaseName)
		return nil
	}
	type args struct {
		allocationID   string
		allocationTx   string
		remotefilepath string
		blobber        *blockchain.StorageNode
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		want           bool
		wantErr        bool
	}{
		{
			"Test_Error_New_Object_Tree_HTTP_Request_Failed",
			args{
				allocationID:   "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx:   "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				remotefilepath: "/1.txt",
				blobber:        &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: string([]byte{0x7f, 0, 0})},
			},
			nil,
			false,
			true,
		},
		{
			"Test_Error_Object_Tree_HTTP_Response_Failed",
			args{
				allocationID:   "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx:   "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				remotefilepath: "/1.txt",
				blobber:        &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: blobberMocks[0].URL},
			},
			nil,
			false,
			true,
		},
		{
			"Test_Error_JSON_Unmarshal_Object_Tree_HTTP_Response_Failed",
			args{
				allocationID:   "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx:   "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				remotefilepath: "/1.txt",
				blobber:        &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: blobberMocks[0].URL},
			},
			blobbersResponseMock,
			false,
			true,
		},
		{
			"Test_Success",
			args{
				allocationID:   "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx:   "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				remotefilepath: "/1.txt",
				blobber:        &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: blobberMocks[0].URL},
			},
			blobbersResponseMock,
			true,
			false,
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
			got, err := getObjectTreeFromBlobber(context.Background(), tt.args.allocationID, tt.args.allocationTx, tt.args.remotefilepath, tt.args.blobber)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error %v", err)
			if tt.want {
				assertion.NotNil(got, "expected result object not nil")
				return
			}
			assertion.Nil(got, "unexpected result object")
		})
	}
}

func Test_getAllocationDataFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, commonTestDir+"/getAllocationDataFromBlobber", testcaseName)
		return nil
	}
	type args struct {
		blobber      *blockchain.StorageNode
		allocationTx string
		respCh       chan *BlobberAllocationStats
		wg           *sync.WaitGroup
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		want           func(t *testing.T, testCaseName string) *BlobberAllocationStats
	}{
		{
			"Test_Error_Create_Get_Allocation_HTTP_Request_Failed",
			args{
				blobber:      &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: string([]byte{0x7f, 0, 0})},
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				respCh:       func() chan *BlobberAllocationStats { var ch = make(chan *BlobberAllocationStats); return ch }(),
				wg:           func() *sync.WaitGroup { var wg = &sync.WaitGroup{}; wg.Add(1); return wg }(),
			},
			nil,
			nil,
		},
		{
			"Test_Error_Get_Allocation_HTTP_Response_Failed",
			args{
				blobber:      &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: blobberMocks[0].URL},
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				respCh:       func() chan *BlobberAllocationStats { var ch = make(chan *BlobberAllocationStats); return ch }(),
				wg:           func() *sync.WaitGroup { var wg = &sync.WaitGroup{}; wg.Add(1); return wg }(),
			},
			nil,
			nil,
		},
		{
			"Test_Error_JSON_Unmarshal_Get_Allocation_HTTP_Response_Failed",
			args{
				blobber:      &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: blobberMocks[0].URL},
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				respCh:       func() chan *BlobberAllocationStats { var ch = make(chan *BlobberAllocationStats); return ch }(),
				wg:           func() *sync.WaitGroup { var wg = &sync.WaitGroup{}; wg.Add(1); return wg }(),
			},
			blobbersResponseMock,
			nil,
		},
		{
			"Test_Success",
			args{
				blobber:      &blockchain.StorageNode{ID: "4a0ffbd42c64f44ec1cca858c7e5b5fd408911ed03df3b7009049cdb76e03ac2", Baseurl: blobberMocks[0].URL},
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				respCh:       func() chan *BlobberAllocationStats { var ch = make(chan *BlobberAllocationStats); return ch }(),
				wg:           func() *sync.WaitGroup { var wg = &sync.WaitGroup{}; wg.Add(1); return wg }(),
			},
			blobbersResponseMock,
			func(t *testing.T, testCaseName string) *BlobberAllocationStats {
				var b *BlobberAllocationStats
				parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v.json", commonTestDir, "getAllocationDataFromBlobber", testCaseName), &b)
				b.BlobberURL = blobberMocks[0].URL
				b.BlobberID = "4a0ffbd42c64f44ec1cca858c7e5b5fd408911ed03df3b7009049cdb76e03ac2"
				return b
			},
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
			if tt.args.respCh != nil {
				defer close(tt.args.respCh)
			}
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				if tt.want != nil {
					assertion.EqualValues(tt.want(t, tt.name), <-tt.args.respCh)
				}
			}()
			getAllocationDataFromBlobber(tt.args.blobber, tt.args.allocationTx, tt.args.respCh, tt.args.wg)
			wg.Wait()
		})
	}
}
