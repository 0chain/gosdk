package sdk

import (
	"context"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/sdk/mocks"
	"github.com/stretchr/testify/assert"
	tm "github.com/stretchr/testify/mock"
	"sync"
	"testing"
)

const commitMetaWorkerTestDir = configDir + "/commitmetaworker"

func TestCommitMetaRequest_processCommitMetaRequest(t *testing.T) {
	// setup mock sdk
	miners, sharders, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a , cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var minerResponseMocks = func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
		setupMinerMockResponses(t, miners, commitMetaWorkerTestDir+"/processCommitMetaRequest", testCaseName)
		return nil
	}
	var sharderResponseMocks = func(t *testing.T, testCaseName string) {
		setupSharderMockResponses(t, sharders, commitMetaWorkerTestDir+"/processCommitMetaRequest", testCaseName)
	}
	var blobbersResponseMock = func(t *testing.T, testcaseName string) (teardown func(t *testing.T)) {
		setupBlobberMockResponses(t, blobberMocks, commitMetaWorkerTestDir+"/processCommitMetaRequest", testcaseName)
		return nil
	}
	var authTicket, err = a.GetAuthTicket("/1.txt", "1.txt", fileref.FILE, client.GetClientID(), "")
	assert.NoErrorf(t, err, "unexpected get auth ticket error: %v", err)
	assert.NotEmptyf(t, authTicket, "unexpected empty auth ticket")
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(1)
	type fields struct {
		CommitMetaData CommitMetaData
		status         func(t *testing.T) StatusCallback
		a              *Allocation
		authToken      string
	}
	tests := []struct {
		name           string
		fields         fields
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
	}{
		{
			"Test_Compute_Hash_And_Sign_Failed",
			fields{
				CommitMetaData: CommitMetaData{},
				status: func(t *testing.T) StatusCallback {
					scm := &mocks.StatusCallback{}
					scm.On("CommitMetaCompleted", "{\"CrudType\":\"\",\"MetaData\":null}", "", tm.AnythingOfType("*errors.errorString")).Once()
					return scm
				},
				a:         a,
				authToken: "",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				cl := client.GetClient()
				keys := cl.Keys
				cl.Keys = []zcncrypto.KeyPair{{}}
				return func(t *testing.T) {
					cl.Keys = keys
				}
			},
		},
		{
			"Test_Sharder_Verify_Txn_Failed",
			fields{
				CommitMetaData: CommitMetaData{},
				status: func(t *testing.T) StatusCallback {
					scm := &mocks.StatusCallback{}
					scm.On("CommitMetaCompleted", "{\"CrudType\":\"\",\"MetaData\":null}", "", tm.AnythingOfType("*common.Error")).Once()
					return scm
				},
				a:         a,
				authToken: "",
			},
			minerResponseMocks,
		},
		{
			"Test_Max_Retried_Failed",
			fields{
				CommitMetaData: CommitMetaData{},
				status: func(t *testing.T) StatusCallback {
					scm := &mocks.StatusCallback{}
					scm.On("CommitMetaCompleted", "{\"CrudType\":\"\",\"MetaData\":null}", "", tm.AnythingOfType("*common.Error")).Once()
					return scm
				},
				a:         a,
				authToken: "",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				maxTxnQuery := blockchain.GetMaxTxnQuery()
				blockchain.SetMaxTxnQuery(0)
				return func(t *testing.T) {
					blockchain.SetMaxTxnQuery(maxTxnQuery)
				}
			},
		},
		{
			"Test_Error_Update_Commit_Meta_Txn_To_Blobber_Success",
			fields{
				CommitMetaData: CommitMetaData{MetaData: &ConsolidatedFileMeta{
					LookupHash: fileref.GetReferenceLookup(a.ID, "/1.txt"),
				}},
				status: func(t *testing.T) StatusCallback {
					scm := &mocks.StatusCallback{}
					scm.On(
						"CommitMetaCompleted",
						"{\"CrudType\":\"\",\"MetaData\":{\"Name\":\"\",\"Type\":\"\",\"Path\":\"\",\"LookupHash\":\"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697\",\"Hash\":\"\",\"MimeType\":\"\",\"Size\":0,\"ActualFileSize\":0,\"ActualNumBlocks\":0,\"EncryptedKey\":\"\",\"CommitMetaTxns\":null,\"Collaborators\":null,\"Attributes\":{}}}",
						"{\"TxnID\":\"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23\",\"MetaData\":{\"Name\":\"\",\"Type\":\"\",\"Path\":\"\",\"LookupHash\":\"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697\",\"Hash\":\"\",\"MimeType\":\"\",\"Size\":0,\"ActualFileSize\":0,\"ActualNumBlocks\":0,\"EncryptedKey\":\"\",\"CommitMetaTxns\":null,\"Collaborators\":null,\"Attributes\":{}}}",
						nil,
					).Once()
					return scm
				},
				a:         a,
				authToken: "",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				minerResponseMocks(t, testCaseName)
				sharderResponseMocks(t, testCaseName)
				blobbersResponseMock(t, testCaseName)
				return nil
			},
		},
		{
			"Test_Success",
			fields{
				CommitMetaData: CommitMetaData{MetaData: &ConsolidatedFileMeta{
					LookupHash: fileref.GetReferenceLookup(a.ID, "/1.txt"),
				}},
				status: func(t *testing.T) StatusCallback {
					scm := &mocks.StatusCallback{}
					scm.On("CommitMetaCompleted",
						"{\"CrudType\":\"\",\"MetaData\":{\"Name\":\"\",\"Type\":\"\",\"Path\":\"\",\"LookupHash\":\"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697\",\"Hash\":\"\",\"MimeType\":\"\",\"Size\":0,\"ActualFileSize\":0,\"ActualNumBlocks\":0,\"EncryptedKey\":\"\",\"CommitMetaTxns\":null,\"Collaborators\":null,\"Attributes\":{}}}",
						"{\"TxnID\":\"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23\",\"MetaData\":{\"Name\":\"\",\"Type\":\"\",\"Path\":\"\",\"LookupHash\":\"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697\",\"Hash\":\"\",\"MimeType\":\"\",\"Size\":0,\"ActualFileSize\":0,\"ActualNumBlocks\":0,\"EncryptedKey\":\"\",\"CommitMetaTxns\":null,\"Collaborators\":null,\"Attributes\":{}}}",
						nil).
						Once()
					return scm
				},
				a:         a,
				authToken: "",
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				minerResponseMocks(t, testCaseName)
				sharderResponseMocks(t, testCaseName)
				blobbersResponseMock(t, testCaseName)
				return nil
			},
		},
		{
			"Test_Success_With_Auth_Ticket",
			fields{
				CommitMetaData: CommitMetaData{MetaData: &ConsolidatedFileMeta{
					LookupHash: fileref.GetReferenceLookup(a.ID, "/1.txt"),
				}},
				status: func(t *testing.T) StatusCallback {
					scm := &mocks.StatusCallback{}
					scm.On("CommitMetaCompleted",
						"{\"CrudType\":\"\",\"MetaData\":{\"Name\":\"\",\"Type\":\"\",\"Path\":\"\",\"LookupHash\":\"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697\",\"Hash\":\"\",\"MimeType\":\"\",\"Size\":0,\"ActualFileSize\":0,\"ActualNumBlocks\":0,\"EncryptedKey\":\"\",\"CommitMetaTxns\":null,\"Collaborators\":null,\"Attributes\":{}}}",
						"{\"TxnID\":\"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23\",\"MetaData\":{\"Name\":\"\",\"Type\":\"\",\"Path\":\"\",\"LookupHash\":\"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697\",\"Hash\":\"\",\"MimeType\":\"\",\"Size\":0,\"ActualFileSize\":0,\"ActualNumBlocks\":0,\"EncryptedKey\":\"\",\"CommitMetaTxns\":null,\"Collaborators\":null,\"Attributes\":{}}}",
						nil).
						Once()
					return scm
				},
				a:         a,
				authToken: authTicket,
			},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				minerResponseMocks(t, testCaseName)
				sharderResponseMocks(t, testCaseName)
				blobbersResponseMock(t, testCaseName)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status StatusCallback
			if st := tt.fields.status; st != nil {
				status = st(t)
			}
			req := &CommitMetaRequest{
				CommitMetaData: tt.fields.CommitMetaData,
				status:         status,
				a:              tt.fields.a,
				authToken:      tt.fields.authToken,
			}
			if mock := tt.additionalMock; mock != nil {
				if teardown := mock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			req.processCommitMetaRequest()
			if st, ok := status.(*mocks.StatusCallback); ok {
				st.Test(t)
				st.AssertExpectations(t)
			}
		})
	}
}

func TestCommitMetaRequest_updatCommitMetaTxnToBlobber(t *testing.T) {
	blobberMock := mocks.NewBlobberHTTPServer(t)
	defer blobberMock.Close()
	var blobber = &blockchain.StorageNode{
		ID:      blobberMock.ID,
		Baseurl: blobberMock.URL,
	}
	var txnHash = "1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23"
	var wg sync.WaitGroup
	var respCh chan bool
	type fields struct {
		CommitMetaData CommitMetaData
		a              *Allocation
		authToken      string
		wg             *sync.WaitGroup
	}
	type args struct {
		txnHash string
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
	}{
		{
			"Test_Error_Decode_Auth_Ticket_Failed",
			fields{
				CommitMetaData: CommitMetaData{
					MetaData: &ConsolidatedFileMeta{
						LookupHash: fileref.GetReferenceLookup("69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", "/1.txt"),
					},
				},
				authToken: "some wrong auth ticket to decode",
				wg:        func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			},
			args{},
			nil,
		},
		{
			"Test_Error_New_HTTP_Failed",
			fields{
				CommitMetaData: CommitMetaData{
					MetaData: &ConsolidatedFileMeta{
						LookupHash: fileref.GetReferenceLookup("69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", "/1.txt"),
					},
				},
				a: &Allocation{
					Tx:  "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f",
					ctx: context.Background(),
				},
				wg: func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			},
			args{},
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				url := blobber.Baseurl
				blobber.Baseurl = string([]byte{0x7f, 0, 0})
				return func(t *testing.T) {
					blobber.Baseurl = url
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CommitMetaRequest{
				CommitMetaData: tt.fields.CommitMetaData,
				a:              tt.fields.a,
				authToken:      tt.fields.authToken,
				wg:             tt.fields.wg,
			}
			if mock := tt.additionalMock; mock != nil {
				if teardown := mock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			req.updatCommitMetaTxnToBlobber(blobber, 0, txnHash, respCh)
		})
	}
}
