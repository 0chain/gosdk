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
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteRequest_deleteBlobberFile(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockRemoteFilePath = "mock/remote/file/path"
		mockAllocationId   = "mock allocation id"
		mockBlobberId      = "mock blobber id"
		mockType           = "f"
		mockConnectionId   = "1234567890"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	var wg sync.WaitGroup

	type parameters struct {
		referencePathToRetrieve fileref.ReferencePath
		requestFields           map[string]string
	}

	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, parameters)
		wantFunc   func(*require.Assertions, *DeleteRequest)
	}{
		{
			name: "Test_Delete_Blobber_File_Failed",
			parameters: parameters{
				referencePathToRetrieve: fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				},
			},
			setup: func(t *testing.T, testName string, p parameters) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName) &&
						req.Method == "GET" &&
						req.Header.Get("X-App-Client-ID") == mockClientId &&
						req.Header.Get("X-App-Client-Key") == mockClientKey
				})).Run(func(args mock.Arguments) {
					for _, c := range mockClient.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body: func() io.ReadCloser {
								jsonFR, err := json.Marshal(p.referencePathToRetrieve)
								require.NoError(t, err)
								return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
							}(),
						}, nil}
					}
				})

				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName) &&
						req.Method == "DELETE" &&
						req.Header.Get("X-App-Client-ID") == mockClientId &&
						req.Header.Get("X-App-Client-Key") == mockClientKey
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
				require.Equal(0, req.deleteMask.CountOnes())
				require.Equal(0, req.consensus.consensus)
			},
		},
		{
			name: "Test_Delete_Blobber_File_Success",
			parameters: parameters{
				referencePathToRetrieve: fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				},
				requestFields: map[string]string{
					"connection_id": mockConnectionId,
					"path":          mockRemoteFilePath,
				},
			},
			setup: func(t *testing.T, testName string, p parameters) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName) &&
						req.Method == "GET" &&
						req.Header.Get("X-App-Client-ID") == mockClientId &&
						req.Header.Get("X-App-Client-Key") == mockClientKey
				})).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body: func() io.ReadCloser {
						jsonFR, err := json.Marshal(p.referencePathToRetrieve)
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
				}, nil)

				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {

					for k, v := range p.requestFields {
						q := req.URL.Query().Get(k)
						require.Equal(t, v, q)
					}

					return strings.HasPrefix(req.URL.Path, testName) &&
						req.Method == "DELETE" &&
						req.Header.Get("X-App-Client-ID") == mockClientId &&
						req.Header.Get("X-App-Client-Key") == mockClientKey
				})).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body: func() io.ReadCloser {
						jsonFR, err := json.Marshal(p.referencePathToRetrieve)
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
				}, nil)
			},
			wantFunc: func(require *require.Assertions, req *DeleteRequest) {
				require.NotNil(req)
				require.Equal(1, req.deleteMask.CountOnes())
				require.Equal(1, req.consensus.consensus)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name, tt.parameters)
			req := &DeleteRequest{
				allocationID:   mockAllocationId,
				allocationTx:   mockAllocationTxId,
				remotefilepath: mockRemoteFilePath,
				consensus: Consensus{
					consensusThresh: 2,
					fullconsensus:   4,
				},
				maskMu:       &sync.Mutex{},
				ctx:          context.TODO(),
				connectionID: mockConnectionId,
				wg:           func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			}
			req.blobbers = append(req.blobbers, &blockchain.StorageNode{
				Baseurl: tt.name,
			})
			req.deleteMask = zboxutil.NewUint128(1).Lsh(uint64(len(req.blobbers))).Sub64(1)
			objectTreeRefs := make([]fileref.RefEntity, 1)
			refEntity, _ := req.getObjectTreeFromBlobber(0)
			objectTreeRefs[0] = refEntity
			req.deleteBlobberFile(req.blobbers[0], 0) //nolint
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
		mockConnectionId   = "1234567890"
	)

	rawClient := zboxutil.Client
	createClient := resty.CreateClient

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	zboxutil.Client = &mockClient
	resty.CreateClient = func(t *http.Transport, timeout time.Duration) resty.Client {
		return &mockClient
	}

	defer func() {
		zboxutil.Client = rawClient
		resty.CreateClient = createClient
	}()

	setupHttpResponses := func(t *testing.T, testName string, numBlobbers int, numCorrect int, req DeleteRequest) { //nolint
		for i := 0; i < numBlobbers; i++ {
			url := mockBlobberUrl + strconv.Itoa(i)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == "GET" &&
					strings.Contains(req.URL.String(), testName+url)
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
					strings.Contains(req.URL.String(), testName+url)
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
		wantFunc    func(*require.Assertions, *DeleteRequest)
	}{
		{
			name:        "Test_All_Blobber_Delete_Attribute_Success",
			numBlobbers: 4,
			numCorrect:  4,
			setup:       setupHttpResponses,
			wantErr:     false,
			wantFunc: func(require *require.Assertions, req *DeleteRequest) {
				require.NotNil(req)
				require.Equal(4, req.deleteMask.CountOnes())
				require.Equal(4, req.consensus.consensus)
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
				require.Equal(3, req.deleteMask.CountOnes())
				require.Equal(3, req.consensus.consensus)
			},
		},
		{
			name:        "Test_Blobber_Index_2_3_Error_On_Delete_Attribute_Failure",
			numBlobbers: 4,
			numCorrect:  2,
			setup:       setupHttpResponses,
			wantErr:     true,
			errMsg:      "delete_failed",
		},
		{
			name:        "Test_All_Blobber_Error_On_Delete_Attribute_Failure",
			numBlobbers: 4,
			numCorrect:  0,
			setup:       setupHttpResponses,
			wantErr:     true,
			errMsg:      "delete_failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			req := &DeleteRequest{
				allocationID:   mockAllocationId,
				allocationTx:   mockAllocationTxId,
				remotefilepath: mockRemoteFilePath,
				consensus: Consensus{
					consensusThresh: 3,
					fullconsensus:   4,
				},
				maskMu:       &sync.Mutex{},
				connectionID: mockConnectionId,
			}
			req.ctx, req.ctxCncl = context.WithCancel(context.TODO())

			a := &Allocation{
				DataShards: numBlobbers,
			}

			for i := 0; i < tt.numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "http://" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			req.deleteMask = zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)

			setupMockAllocation(t, a)
			setupMockWriteLockRequest(a, &mockClient)

			req.allocationObj = a
			req.blobbers = a.Blobbers

			tt.setup(t, tt.name, tt.numBlobbers, tt.numCorrect, *req) //nolint
			err := req.ProcessDelete()
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(errors.Top(err), tt.errMsg, "expected error contains '%s'", tt.errMsg)
				return
			}
			if tt.wantFunc != nil {
				tt.wantFunc(require, req)
			}
		})
	}
}
