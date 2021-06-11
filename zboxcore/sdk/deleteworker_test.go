package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteRequest_getObjectTreeFromBlobber(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockAllocationId   = "mock allocation id"
		mockBlobberId      = "mock blobber id"
		mockType           = "f"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	tests := []struct {
		name    string
		setup   func(*testing.T, string)
		wantErr bool
		errMsg  string
	}{
		{
			name: "Test_Get_Object_Tree_From_Blobber_Failed",
			setup: func(t *testing.T, testName string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			wantErr: true,
			errMsg:  "Object tree error response: Status: 400 -  ",
		},
		{
			name: "Test_Get_Object_Tree_From_Blobber_Success",
			setup: func(t *testing.T, testName string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName)
				})).Run(func(args mock.Arguments) {
					for _, c := range mockClient.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body: func() io.ReadCloser {
								jsonFR, err := json.Marshal(fileref.ReferencePath{
									Meta: map[string]interface{}{
										"type": mockType,
									},
								})
								require.NoError(t, err)
								return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
							}(),
						}, nil}
					}
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name)
			blobber := &blockchain.StorageNode{
				ID:      mockBlobberId,
				Baseurl: tt.name,
			}
			req := &DeleteRequest{
				allocationID: mockAllocationId,
				allocationTx: mockAllocationTxId,
				Consensus: Consensus{
					consensusThresh: 50,
					fullconsensus:   4,
				},
				ctx:          context.TODO(),
				connectionID: zboxutil.NewConnectionId(),
			}
			objectTreeRefs := make([]fileref.RefEntity, 1)
			refEntity, _ := req.getObjectTreeFromBlobber(blobber)
			objectTreeRefs[0] = refEntity
			_, err := req.getObjectTreeFromBlobber(blobber)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
		})
	}
}

func TestDeleteRequest_deleteBlobberFile(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockRemoteFilePath = "mock/remote/file/path"
		mockAllocationId   = "mock allocation id"
		mockBlobberId      = "mock blobber id"
		mockType           = "f"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	var wg sync.WaitGroup

	tests := []struct {
		name     string
		setup    func(t *testing.T, name string)
		wantFunc func(require *require.Assertions, req *DeleteRequest)
	}{
		{
			name: "Test_Delete_Blobber_File_Failed",
			setup: func(t *testing.T, testName string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == "GET" &&
						strings.HasPrefix(req.URL.Path, testName)
				})).Run(func(args mock.Arguments) {
					for _, c := range mockClient.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body: func() io.ReadCloser {
								jsonFR, err := json.Marshal(fileref.ReferencePath{
									Meta: map[string]interface{}{
										"type": mockType,
									},
								})
								require.NoError(t, err)
								return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
							}(),
						}, nil}
					}
				})
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == "DELETE" &&
						strings.HasPrefix(req.URL.Path, testName)
				})).Run(func(args mock.Arguments) {
					for _, c := range mockClient.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusBadRequest,
							Body:       ioutil.NopCloser(strings.NewReader("")),
						}, nil}
					}
				})
			},
			wantFunc: func(require *require.Assertions, req *DeleteRequest) {
				require.NotNil(req)
				require.Equal(uint32(0), req.deleteMask)
				require.Equal(float32(0), req.consensus)
			},
		},
		{
			name: "Test_Delete_Blobber_File_Success",
			setup: func(t *testing.T, testName string) {
				mockClient.On("Do", mock.AnythingOfType("*http.Request")).Run(func(args mock.Arguments) {
					for _, c := range mockClient.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body: func() io.ReadCloser {
								jsonFR, err := json.Marshal(fileref.ReferencePath{
									Meta: map[string]interface{}{
										"type": mockType,
									},
								})
								require.NoError(t, err)
								return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
							}(),
						}, nil}
					}
				})
			},
			wantFunc: func(require *require.Assertions, req *DeleteRequest) {
				require.NotNil(req)
				require.Equal(uint32(1), req.deleteMask)
				require.Equal(float32(1), req.consensus)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name)
			req := &DeleteRequest{
				allocationID:   mockAllocationId,
				allocationTx:   mockAllocationTxId,
				remotefilepath: mockRemoteFilePath,
				Consensus: Consensus{
					consensusThresh: 50,
					fullconsensus:   4,
				},
				ctx:          context.TODO(),
				connectionID: zboxutil.NewConnectionId(),
				wg:           func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			}
			req.blobbers = append(req.blobbers, &blockchain.StorageNode{
				Baseurl: tt.name,
			})
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

func TestDeleteRequest_ProcessDelete(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockRemoteFilePath = "mock/remote/file/path"
		mockAllocationId   = "mock allocation id"
		mockBlobberId      = "mock blobber id"
		mockBlobberUrl     = "mockblobberurl"
		mockType           = "f"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponses := func(t *testing.T, testName string, numBlobbers int, numCorrect int, req DeleteRequest) {
		for i := 0; i < numBlobbers; i++ {
			url := mockBlobberUrl + strconv.Itoa(i)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == "GET" &&
					strings.HasPrefix(req.URL.Path, testName+url)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func() io.ReadCloser {
					jsonFR, err := json.Marshal(&fileref.ReferencePath{
						Meta: map[string]interface{}{
							"type": mockType,
						},
					})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(),
			}, nil)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == "DELETE" &&
					strings.HasPrefix(req.URL.Path, testName+url)
			})).Return(&http.Response{
				StatusCode: func() int {
					if i < numCorrect {
						return http.StatusOK
					}
					return http.StatusBadRequest
				}(),
				Body: ioutil.NopCloser(bytes.NewReader([]byte(""))),
			}, nil)
		}

		commitChan = make(map[string]chan *CommitRequest)
		for _, blobber := range req.blobbers {
			if _, ok := commitChan[blobber.ID]; !ok {
				commitChan[blobber.ID] = make(chan *CommitRequest, 1)
			}
		}
		blobberChan := commitChan
		go func() {
			cm0 := <-blobberChan[req.blobbers[0].ID]
			require.EqualValues(t, cm0.blobber.ID, testName+mockBlobberId+strconv.Itoa(0))
			cm0.result = &CommitResult{
				Success: true,
			}
			if cm0.wg != nil {
				cm0.wg.Done()
			}
		}()
		go func() {
			cm1 := <-blobberChan[req.blobbers[1].ID]
			require.EqualValues(t, cm1.blobber.ID, testName+mockBlobberId+strconv.Itoa(1))
			cm1.result = &CommitResult{
				Success: true,
			}
			if cm1.wg != nil {
				cm1.wg.Done()
			}
		}()
		go func() {
			cm2 := <-blobberChan[req.blobbers[2].ID]
			require.EqualValues(t, cm2.blobber.ID, testName+mockBlobberId+strconv.Itoa(2))
			cm2.result = &CommitResult{
				Success: true,
			}
			if cm2.wg != nil {
				cm2.wg.Done()
			}
		}()
		go func() {
			cm3 := <-blobberChan[req.blobbers[3].ID]
			require.EqualValues(t, cm3.blobber.ID, testName+mockBlobberId+strconv.Itoa(3))
			cm3.result = &CommitResult{
				Success: true,
			}
			if cm3.wg != nil {
				cm3.wg.Done()
			}
		}()
	}

	tests := []struct {
		name        string
		numBlobbers int
		numCorrect  int
		setup       func(*testing.T, string, int, int, DeleteRequest)
		wantErr     bool
		errMsg      string
		wantFunc    func(require *require.Assertions, req *DeleteRequest)
	}{
		{
			name:        "Test_All_Blobber_Delete_Attribute_Success",
			numBlobbers: 4,
			numCorrect:  4,
			setup:       setupHttpResponses,
			wantErr:     false,
			wantFunc: func(require *require.Assertions, req *DeleteRequest) {
				require.NotNil(req)
				require.Equal(uint32(15), req.deleteMask)
				require.Equal(float32(4), req.consensus)
			},
		},
		{
			name:        "Test_Blobber_Index_3_Error_On_Delete_Attribute_Success",
			numBlobbers: 4,
			numCorrect:  3,
			setup:       setupHttpResponses,
			wantErr:     false,
			wantFunc: func(require *require.Assertions, req *DeleteRequest) {
				require.NotNil(req)
				require.Equal(uint32(7), req.deleteMask)
				require.Equal(float32(3), req.consensus)
			},
		},
		{
			name:        "Test_Blobber_Index_2_3_Error_On_Delete_Attribute_Failure",
			numBlobbers: 4,
			numCorrect:  2,
			setup:       setupHttpResponses,
			wantErr:     true,
			errMsg:      "Delete failed",
		},
		{
			name:        "Test_All_Blobber_Error_On_Delete_Attribute_Failure",
			numBlobbers: 4,
			numCorrect:  0,
			setup:       setupHttpResponses,
			wantErr:     true,
			errMsg:      "Delete failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			req := &DeleteRequest{
				allocationID:   mockAllocationId,
				allocationTx:   mockAllocationTxId,
				remotefilepath: mockRemoteFilePath,
				Consensus: Consensus{
					consensusThresh: 50,
					fullconsensus:   4,
				},
				ctx:          context.TODO(),
				connectionID: zboxutil.NewConnectionId(),
			}
			for i := 0; i < tt.numBlobbers; i++ {
				req.blobbers = append(req.blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			tt.setup(t, tt.name, tt.numBlobbers, tt.numCorrect, *req)
			err := req.ProcessDelete()
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg, "expected error contains '%s'", tt.errMsg)
				return
			}
			if tt.wantFunc != nil {
				tt.wantFunc(require, req)
			}
		})
	}
}
