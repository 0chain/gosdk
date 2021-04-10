package sdk

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	cmocks "github.com/0chain/gosdk/zboxcore/allocationchange/mocks"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/sdk/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"math/rand"
	"sync"
	"testing"
)

const commitWorkerTestDir = configDir + "/commitworker"

func TestErrorCommitResult(t *testing.T) {
	var errMsg = "some error message"
	assert.Equal(t, &CommitResult{Success: false, ErrorMessage: errMsg}, ErrorCommitResult(errMsg))
}

func TestSuccessCommitResult(t *testing.T) {
	assert.Equal(t, &CommitResult{Success: true}, SuccessCommitResult())
}

func TestInitCommitWorker(t *testing.T) {
	blobberMock := mocks.NewBlobberHTTPServer(t)
	defer blobberMock.Close()
	InitCommitWorker([]*blockchain.StorageNode{{ID: blobberMock.ID, Baseurl: blobberMock.URL}})
	defer close(commitChan[blobberMock.ID])
	assert.NotNil(t, commitChan[blobberMock.ID])
}

func Test_startCommitWorker(t *testing.T) {
	bytes := make([]byte, 32)
	var blobberID = hex.EncodeToString(bytes)
	if commitChan == nil {
		commitChan = make(map[string]chan *CommitRequest)
	}
	commitChan[blobberID] = make(chan *CommitRequest, 1)
	blobberChan := commitChan[blobberID]
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		startCommitWorker(blobberChan, blobberID)
		t.Log("Done")
	}(wg)
	blobberChan <- &CommitRequest{blobber: &blockchain.StorageNode{ID: blobberID, Baseurl: string([]byte{0x7f, 0, 0})}, wg: func() *sync.WaitGroup { wg := &sync.WaitGroup{}; wg.Add(1); return wg }()}
	close(blobberChan)
	wg.Wait()
}

func TestCommitRequest_processCommit(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, commitWorkerTestDir+"/processCommit", testcaseName)
		return nil
	}
	var wg = &sync.WaitGroup{}

	type fields struct {
		changes      []allocationchange.AllocationChange
		blobber      *blockchain.StorageNode
		allocationID string
		allocationTx string
		connectionID string
		wg           *sync.WaitGroup
	}
	tests := []struct {
		name           string
		fields         fields
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		want           *CommitResult
	}{
		{
			"Test_Error_New_HTTP_Request_Failed",
			fields{
				changes: func() []allocationchange.AllocationChange {
					acmock := &cmocks.AllocationChange{}
					acmock.On("GetAffectedPath").Return("/1.txt").Once()
					return []allocationchange.AllocationChange{acmock}
				}(),
				blobber: &blockchain.StorageNode{
					ID:      blobberMocks[0].ID,
					Baseurl: string([]byte{0x7f, 0, 0}),
				},
				allocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				connectionID: zboxutil.NewConnectionId(),
				wg:           func() *sync.WaitGroup { wg.Add(1); return wg }(),
			},
			nil,
			nil,
		},
		{
			"Test_Error_Blobber_Referrence_Path_Response_Failed",
			fields{
				changes: func() []allocationchange.AllocationChange {
					acmock := &cmocks.AllocationChange{}
					acmock.On("GetAffectedPath").Return("/1.txt").Once()
					return []allocationchange.AllocationChange{acmock}
				}(),
				blobber: &blockchain.StorageNode{
					ID:      blobberMocks[0].ID,
					Baseurl: blobberMocks[0].URL,
				},
				allocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				connectionID: zboxutil.NewConnectionId(),
				wg:           func() *sync.WaitGroup { wg.Add(1); return wg }(),
			},
			nil,
			ErrorCommitResult("Reference path error response: Status: 500 - Internal Server Error!"),
		},
		{
			"Test_Error_JSON_Marshaling_Blobber_Referrence_Path_Response_Failed",
			fields{
				changes: func() []allocationchange.AllocationChange {
					acmock := &cmocks.AllocationChange{}
					acmock.On("GetAffectedPath").Return("/1.txt").Once()
					return []allocationchange.AllocationChange{acmock}
				}(),
				blobber: &blockchain.StorageNode{
					ID:      blobberMocks[0].ID,
					Baseurl: blobberMocks[0].URL,
				},
				allocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				connectionID: zboxutil.NewConnectionId(),
				wg:           func() *sync.WaitGroup { wg.Add(1); return wg }(),
			},
			blobbersResponseMock,
			ErrorCommitResult("json: cannot unmarshal string into Go value of type sdk.ReferencePathResult"),
		},
		{
			"Test_Error_Reference_Path_Result_Get_Dir_Tree_Failed",
			fields{
				changes: func() []allocationchange.AllocationChange {
					acmock := &cmocks.AllocationChange{}
					acmock.On("GetAffectedPath").Return("/1.txt").Once()
					return []allocationchange.AllocationChange{acmock}
				}(),
				blobber: &blockchain.StorageNode{
					ID:      blobberMocks[0].ID,
					Baseurl: blobberMocks[0].URL,
				},
				allocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				connectionID: zboxutil.NewConnectionId(),
				wg:           func() *sync.WaitGroup { wg.Add(1); return wg }(),
			},
			blobbersResponseMock,
			ErrorCommitResult("invalid_ref_path: Invalid reference path. root was not a directory type"),
		},
		{
			"Test_Error_Allocation_Change_Process_Failed",
			fields{
				changes: func() []allocationchange.AllocationChange {
					acmock := &cmocks.AllocationChange{}
					acmock.On("GetAffectedPath").Return("/1.txt").Once()
					acmock.On("ProcessChange", mock.AnythingOfType("*fileref.Ref")).Return(errors.New("some error message")).Once()
					return []allocationchange.AllocationChange{acmock}
				}(),
				blobber: &blockchain.StorageNode{
					ID:      blobberMocks[0].ID,
					Baseurl: blobberMocks[0].URL,
				},
				allocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				connectionID: zboxutil.NewConnectionId(),
				wg:           func() *sync.WaitGroup { wg.Add(1); return wg }(),
			},
			blobbersResponseMock,
			ErrorCommitResult("some error message"),
		},
		{
			"Test_Error_Commit_Blobber_Failed",
			fields{
				changes: func() []allocationchange.AllocationChange {
					acmock := &cmocks.AllocationChange{}
					acmock.On("GetAffectedPath").Return("/1.txt").Once()
					acmock.On("ProcessChange", mock.AnythingOfType("*fileref.Ref")).Return(nil).Once()
					acmock.On("GetSize").Return(rand.Int63())
					return []allocationchange.AllocationChange{acmock}
				}(),
				blobber: &blockchain.StorageNode{
					ID:      blobberMocks[0].ID,
					Baseurl: blobberMocks[0].URL,
				},
				allocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				connectionID: zboxutil.NewConnectionId(),
				wg:           func() *sync.WaitGroup { wg.Add(1); return wg }(),
			},
			blobbersResponseMock,
			ErrorCommitResult("commit_error: Internal Server Error!"),
		},
		{
			"Test_Success",
			fields{
				changes: func() []allocationchange.AllocationChange {
					acmock := &cmocks.AllocationChange{}
					acmock.On("GetAffectedPath").Return("/1.txt").Once()
					acmock.On("ProcessChange", mock.AnythingOfType("*fileref.Ref")).Return(nil).Once()
					acmock.On("GetSize").Return(rand.Int63())
					return []allocationchange.AllocationChange{acmock}
				}(),
				blobber: &blockchain.StorageNode{
					ID:      blobberMocks[0].ID,
					Baseurl: blobberMocks[0].URL,
				},
				allocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				connectionID: zboxutil.NewConnectionId(),
				wg:           func() *sync.WaitGroup { wg.Add(1); return wg }(),
			},
			blobbersResponseMock,
			SuccessCommitResult(),
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
			commitreq := &CommitRequest{
				changes:      tt.fields.changes,
				blobber:      tt.fields.blobber,
				allocationID: tt.fields.allocationID,
				allocationTx: tt.fields.allocationTx,
				connectionID: tt.fields.connectionID,
				wg:           tt.fields.wg,
			}
			commitreq.processCommit()
			assertion.EqualValues(tt.want, commitreq.result)
		})
	}
}

func TestCommitRequest_commitBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, commitWorkerTestDir+"/commitBlobber", testcaseName)
		return nil
	}

	type fields struct {
		blobber      *blockchain.StorageNode
		allocationID string
		allocationTx string
		connectionID string
	}
	type args struct {
		rootRef  *fileref.Ref
		latestWM *marker.WriteMarker
		size     int64
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Error_Sign_Write_Marker_Failed",
			fields{
				blobber:      &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: blobberMocks[0].URL},
				allocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				connectionID: zboxutil.NewConnectionId(),
			},
			args{
				rootRef: &fileref.Ref{
					Type:         "d",
					AllocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
					Hash:         "e1deaa241bfb0ea70d29910d6cfec3bf4d9ab35b3a8ed9f4c2e7409c4740953d",
				},
				latestWM: &marker.WriteMarker{AllocationRoot: "bf634d1d05e0ed35088d91a96fa1b11549449356247f46b22949497852f85510"},
				size:     rand.Int63(),
			},
			nil,
			true,
		},
		{
			"Test_Error_New_Commit_HTTP_Request_Failed",
			fields{
				blobber:      &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: string([]byte{0x7f, 0, 0})},
				allocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				connectionID: zboxutil.NewConnectionId(),
			},
			args{
				rootRef: &fileref.Ref{
					Type:         "d",
					AllocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
					Hash:         "e1deaa241bfb0ea70d29910d6cfec3bf4d9ab35b3a8ed9f4c2e7409c4740953d",
				},
				latestWM: &marker.WriteMarker{AllocationRoot: "bf634d1d05e0ed35088d91a96fa1b11549449356247f46b22949497852f85510"},
				size:     rand.Int63(),
			},
			nil,
			true,
		},
		{
			"Test_Not_Existed_Latest_Write_Marker_Success",
			fields{
				blobber:      &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: blobberMocks[0].URL},
				allocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
				connectionID: zboxutil.NewConnectionId(),
			},
			args{
				rootRef: &fileref.Ref{
					Type:         "d",
					AllocationID: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
					Hash:         "e1deaa241bfb0ea70d29910d6cfec3bf4d9ab35b3a8ed9f4c2e7409c4740953d",
				},
				latestWM: nil,
				size:     rand.Int63(),
			},
			blobbersResponseMock,
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
			req := &CommitRequest{
				blobber:      tt.fields.blobber,
				allocationID: tt.fields.allocationID,
				allocationTx: tt.fields.allocationTx,
				connectionID: tt.fields.connectionID,
			}
			err := req.commitBlobber(tt.args.rootRef, tt.args.latestWM, tt.args.size)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error %v", err)
		})
	}
}

func TestCommitRequest_calculateHashRequest(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()

	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, commitWorkerTestDir+"/calculateHashRequest", testcaseName)
		return nil
	}
	type fields struct {
		blobber      *blockchain.StorageNode
		allocationTx string
	}
	type args struct {
		ctx   context.Context
		paths []string
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_No_Path_Coverage",
			fields{},
			args{paths: []string{}},
			nil,
			false,
		},
		{
			"Test_Error_New_Calculate_Hash_HTTP_Request_Failed",
			fields{
				blobber:      &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: string([]byte{0x7f, 0, 0})},
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
			},
			args{paths: []string{"/1.txt"}},
			nil,
			true,
		},
		{
			"Test_Error_Calculate_Hash_Response_Failed",
			fields{
				blobber:      &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: blobberMocks[0].URL},
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
			},
			args{paths: []string{"/1.txt"}},
			nil,
			true,
		},
		{
			"Test_Success",
			fields{
				blobber:      &blockchain.StorageNode{ID: blobberMocks[0].ID, Baseurl: blobberMocks[0].URL},
				allocationTx: "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
			},
			args{paths: []string{"/1.txt"}},
			blobbersResponseMock,
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
			commitreq := &CommitRequest{
				blobber:      tt.fields.blobber,
				allocationTx: tt.fields.allocationTx,
			}
			err := commitreq.calculateHashRequest(context.Background(), tt.args.paths)
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error %v", err)
		})
	}
}
