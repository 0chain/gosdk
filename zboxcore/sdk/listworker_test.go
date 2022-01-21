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

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListRequest_getListInfoFromBlobber(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockRemoteFilePath = "mock/remote/file/path"
		mockSignature      = "mock signature"
		mockAllocationId   = "mock allocation id"
		mockErrorMessage   = "mock error message"
		mockBlobberId      = "mock blobber Id"
		mockAllocationRoot = "mock allocation root"
		mockType           = "d"
	)
	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient
	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}
	type parameters struct {
		listHttpResp   listResponse
		ListResult     fileref.ListResult
		respStatusCode int
		blobberIdx     int
		requestFields  map[string]string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(t *testing.T, name string, p parameters, errMsg string)
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Error_New_HTTP_Failed_By_Containing_" + string([]byte{0x7f, 0, 0}),
			parameters: parameters{
				blobberIdx:     41,
				respStatusCode: 0,
			},
			setup:   func(t *testing.T, name string, p parameters, errMsg string) {},
			wantErr: true,
			errMsg:  `parse "Test_Error_New_HTTP_Failed_By_Containing_\u007f\x00\x00": net/url: invalid control character in URL`,
		},
		{
			name: "Test_HTTP_Error",
			parameters: parameters{
				blobberIdx:     41,
				respStatusCode: 0,
			},
			setup: func(t *testing.T, name string, p parameters, errMsg string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, name)
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
					StatusCode: p.respStatusCode,
				}, errors.New("", mockErrorMessage))
			},
			wantErr: true,
			errMsg:  mockErrorMessage,
		},
		{
			name: "Test_HTTP_Response_Failed",
			parameters: parameters{
				blobberIdx:     41,
				respStatusCode: http.StatusInternalServerError,
			},
			setup: func(t *testing.T, name string, p parameters, errMsg string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, name)
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(mockErrorMessage))),
					StatusCode: p.respStatusCode,
				}, nil)
			},
			wantErr: true,
			errMsg:  "error from server list response: " + mockErrorMessage,
		},
		{
			name: "Test_Error_HTTP_Response_Not_JSON_Format",
			parameters: parameters{
				blobberIdx:     41,
				respStatusCode: http.StatusOK,
			},
			setup: func(t *testing.T, name string, p parameters, errMsg string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, name)
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("this is not json format"))),
					StatusCode: p.respStatusCode,
				}, nil)
			},
			wantErr: true,
			errMsg:  "list entities response parse error:",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				listHttpResp: listResponse{
					ref: &fileref.Ref{
						AllocationID: mockAllocationId,
						Type:         mockType,
					},
				},
				ListResult: fileref.ListResult{
					AllocationRoot: mockAllocationRoot,
					Meta: map[string]interface{}{
						"type": mockType,
					},
				},
				blobberIdx:     41,
				respStatusCode: http.StatusOK,
				requestFields: map[string]string{
					"auth_token": func() string {
						authBytes, err := json.Marshal(&marker.AuthTicket{
							Signature: mockSignature,
						})
						require.NoError(t, err)
						return string(authBytes)
					}(),
					"path_hash": fileref.GetReferenceLookup(mockAllocationId, mockRemoteFilePath),
				},
			},
			setup: func(t *testing.T, testName string, p parameters, _ string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					query := req.URL.Query()
					require.EqualValues(t, query.Get("auth_token"), p.requestFields["auth_token"])
					require.EqualValues(t, query.Get("path_hash"), p.requestFields["path_hash"])
					return req.URL.Path == "Test_Success"+zboxutil.LIST_ENDPOINT+mockAllocationTxId &&
						req.Method == "GET" &&
						req.Header.Get("X-App-Client-ID") == mockClientId &&
						req.Header.Get("X-App-Client-Key") == mockClientKey &&
						testName == "Test_Success"
				})).Return(&http.Response{
					StatusCode: p.respStatusCode,
					Body: func(p parameters) io.ReadCloser {
						jsonFR, err := json.Marshal(p.ListResult)
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(p),
				}, nil).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name, tt.parameters, tt.errMsg)
			blobber := &blockchain.StorageNode{
				ID:      mockBlobberId,
				Baseurl: tt.name,
			}
			req := &ListRequest{
				allocationID:   mockAllocationId,
				allocationTx:   mockAllocationTxId,
				ctx:            context.TODO(),
				remotefilepath: mockRemoteFilePath,
				authToken: &marker.AuthTicket{
					Signature: mockSignature,
				},
				wg: &sync.WaitGroup{},
			}
			rspCh := make(chan *listResponse, 1)
			req.wg.Add(1)
			go req.getListInfoFromBlobber(blobber, 41, rspCh)
			req.wg.Wait()
			resp := <-rspCh
			require.EqualValues(tt.wantErr, resp.err != nil)
			if resp.err != nil {
				require.EqualValues(tt.errMsg, errors.Top(resp.err))
				return
			}
			require.EqualValues(tt.parameters.listHttpResp.ref, resp.ref)
			require.EqualValues(tt.parameters.blobberIdx, resp.blobberIdx)
		})
	}
}

func TestListRequest_GetListFromBlobbers(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockAllocationId   = "mock allocation id"
		mockBlobberUrl     = "mockBlobberUrl"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockAllocationRoot = "mock allocation root"
		mockType           = "d"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponses := func(t *testing.T, name string, numBlobbers int) {
		for i := 0; i < numBlobbers; i++ {
			url := mockBlobberUrl + strconv.Itoa(i)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.Path, name+url)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func() io.ReadCloser {
					jsonFR, err := json.Marshal(&fileref.ListResult{
						AllocationRoot: mockAllocationRoot,
						Meta: map[string]interface{}{
							"type": mockType,
						},
					})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(),
			}, nil)
		}
	}

	tests := []struct {
		name        string
		numBlobbers int

		setup    func(*testing.T, string, int)
		wantFunc func(require *require.Assertions, req *ListRequest)
		wantErr  bool
	}{
		{
			name:  "Test_Failed",
			setup: nil,
			wantFunc: func(require *require.Assertions, req *ListRequest) {
				require.NotNil(req)
				require.Equal(float32(0), req.consensus)
			},
			wantErr: true,
		},
		{
			name:        "Test_Success",
			numBlobbers: 4,
			setup:       setupHttpResponses,
			wantFunc: func(require *require.Assertions, req *ListRequest) {
				require.NotNil(req)
				require.Equal(float32(4), req.consensus)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if setup := tt.setup; setup != nil {
				tt.setup(t, tt.name, tt.numBlobbers)
			}
			req := &ListRequest{
				allocationID: mockAllocationId,
				allocationTx: mockAllocationTxId,
				ctx:          context.TODO(),
				blobbers:     []*blockchain.StorageNode{},
				wg:           &sync.WaitGroup{},
				Consensus: Consensus{
					consensusThresh:        50,
					fullconsensus:          4,
					consensusRequiredForOk: 60,
				},
			}
			for i := 0; i < tt.numBlobbers; i++ {
				req.blobbers = append(req.blobbers, &blockchain.StorageNode{
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			got := req.GetListFromBlobbers()
			expectedResult := &ListResult{
				Type: mockType,
				Size: -1,
			}
			if !tt.wantErr {
				require.EqualValues(expectedResult, got)
				return
			}
			tt.wantFunc(require, req)
		})
	}
}
