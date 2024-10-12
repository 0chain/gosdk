package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/dev"
	devMock "github.com/0chain/gosdk/dev/mock"
	"github.com/0chain/gosdk/sdks/blobber"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRenameRequest_renameBlobberObject(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockRemoteFilePath = "mock/remote/file/path"
		mockAllocationId   = "mock allocation id"
		mockBlobberId      = "mock blobber id"
		mockType           = "f"
		mockConnectionId   = "1234567890"
		mockNewName        = "mock new name"
	)

	var mockClient = mocks.HttpClient{}
	rawClient := zboxutil.Client
	zboxutil.Client = &mockClient
	defer func() {
		zboxutil.Client = rawClient
	}()

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		referencePathToRetrieve fileref.ReferencePath
		requestFields           map[string]string
	}

	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, parameters)
		wantErr    bool
		errMsg     string
		wantFunc   func(*require.Assertions, *RenameRequest)
	}{
		{
			name:    "Test_Error_New_HTTP_Failed_By_Containing_" + string([]byte{0x7f, 0, 0}),
			setup:   func(t *testing.T, testName string, p parameters) {},
			wantErr: true,
			errMsg:  `net/url: invalid control character in URL`,
		},
		{
			name: "Test_Error_Get_Object_Tree_From_Blobber_Failed",
			setup: func(t *testing.T, testName string, p parameters) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			wantErr: true,
			errMsg:  "400: Object tree error response: Body:",
		},
		{
			name: "Test_Rename_Blobber_Object_Failed",
			parameters: parameters{
				referencePathToRetrieve: fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				},
			},
			setup: func(t *testing.T, testName string, p parameters) {

				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.Contains(req.URL.Path, "latestwritemarker") &&
						req.Method == "GET"
				})).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body: func() io.ReadCloser {
						jsonFR := `{"latest_write_marker":null,"prev_write_marker":null}`
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
				}, nil)

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
					return strings.HasPrefix(req.URL.Path, testName) &&
						req.Method == "POST" &&
						req.Header.Get("X-App-Client-ID") == mockClientId &&
						req.Header.Get("X-App-Client-Key") == mockClientKey
				})).Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			wantErr: true,
			errMsg:  "response_error: ",
			wantFunc: func(require *require.Assertions, req *RenameRequest) {
				require.NotNil(req)
				require.Equal(0, req.renameMask.CountOnes())
				require.Equal(0, req.consensus.consensus)
			},
		},
		{
			name: "Test_Rename_Blobber_Object_Success",
			parameters: parameters{
				referencePathToRetrieve: fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				},
				requestFields: map[string]string{
					"connection_id": mockConnectionId,
					"path":          mockRemoteFilePath,
					"new_name":      mockNewName,
				},
			},
			setup: func(t *testing.T, testName string, p parameters) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.Contains(req.URL.Path, "latestwritemarker") &&
						req.Method == "GET"
				})).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body: func() io.ReadCloser {
						jsonFR := `{"latest_write_marker":null,"prev_write_marker":null}`
						return io.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
				}, nil)

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
					mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
					require.NoError(t, err)
					require.True(t, strings.HasPrefix(mediaType, "multipart/"))
					reader := multipart.NewReader(req.Body, params["boundary"])

					err = nil
					for {
						var part *multipart.Part
						part, err = reader.NextPart()
						if err != nil {
							break
						}
						expected, ok := p.requestFields[part.FormName()]
						require.True(t, ok)
						actual, err := ioutil.ReadAll(part)
						require.NoError(t, err)
						require.EqualValues(t, expected, string(actual))
					}
					require.Error(t, err)
					require.EqualValues(t, "EOF", errors.Top(err))

					return strings.HasPrefix(req.URL.Path, testName) &&
						req.Method == "POST" &&
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
			wantFunc: func(require *require.Assertions, req *RenameRequest) {
				require.NotNil(req)
				require.Equal(4, req.renameMask.CountOnes())
				require.Equal(1, req.consensus.consensus)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name, tt.parameters)
			req := &RenameRequest{
				allocationID:   mockAllocationId,
				allocationTx:   mockAllocationTxId,
				remotefilepath: mockRemoteFilePath,
				consensus: Consensus{
					RWMutex:         &sync.RWMutex{},
					consensusThresh: 2,
					fullconsensus:   4,
				},
				ctx:          context.TODO(),
				renameMask:   zboxutil.NewUint128(1).Lsh(uint64(4)).Sub64(1),
				maskMU:       &sync.Mutex{},
				connectionID: mockConnectionId,
				newName:      mockNewName,
			}
			req.blobbers = append(req.blobbers, &blockchain.StorageNode{
				Baseurl: tt.name,
			})
			err := req.renameBlobberObject(req.blobbers[0], 0)
			require.EqualValues(tt.wantErr, err != nil, "Error: ", err)
			if err != nil {
				require.Contains(errors.Top(err), tt.errMsg)
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
			if tt.wantFunc != nil {
				tt.wantFunc(require, req)
			}
		})
	}
}

func TestRenameRequest_ProcessRename(t *testing.T) {
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
		mockNewName        = "mock new name"
	)

	var mockClient = mocks.HttpClient{}
	rawClient := zboxutil.Client
	zboxutil.Client = &mockClient
	defer func() {
		zboxutil.Client = rawClient
	}()

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponses := func(t *testing.T, testName string, numBlobbers int, numCorrect int, req *RenameRequest) {

		for i := 0; i < numBlobbers; i++ {
			url := mockBlobberUrl + strconv.Itoa(i)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == "GET" &&
					strings.Contains(req.URL.String(), testName+url) && !strings.Contains(req.URL.String(), "latestwritemarker")
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
				return req.Method == "POST" &&
					strings.Contains(req.URL.String(), zboxutil.RENAME_ENDPOINT) &&
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

			if i < numCorrect {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == "POST" &&
						strings.Contains(req.URL.String(), zboxutil.WM_LOCK_ENDPOINT) &&
						strings.Contains(req.URL.String(), testName+url)
				})).Return(&http.Response{
					StatusCode: func() int {
						if i < numCorrect {
							return http.StatusOK
						}
						return http.StatusBadRequest
					}(),
					Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"status":2}`))),
				}, nil)

				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == "POST" &&
						strings.Contains(req.URL.String(), zboxutil.COMMIT_ENDPOINT) &&
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

			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == "DELETE" &&
					strings.Contains(req.URL.String(), testName+url)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
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
		setup       func(*testing.T, string, int, int, *RenameRequest)
		wantErr     bool
		errMsg      string
		wantFunc    func(require *require.Assertions, req *RenameRequest)
	}{
		{
			name:        "Test_All_Blobber_Rename_Success",
			numBlobbers: 4,
			numCorrect:  4,
			setup:       setupHttpResponses,
			wantErr:     false,
			wantFunc: func(require *require.Assertions, req *RenameRequest) {
				require.NotNil(req)
				require.Equal(4, req.renameMask.CountOnes())
				require.Equal(4, req.consensus.consensus)
			},
		},
		{
			name:        "Test_Blobber_Index_3_Error_On_Rename_Success",
			numBlobbers: 4,
			numCorrect:  3,
			setup:       setupHttpResponses,
			wantErr:     false,
			wantFunc: func(require *require.Assertions, req *RenameRequest) {
				require.NotNil(req)
				require.Equal(3, req.renameMask.CountOnes())
				require.Equal(3, req.consensus.consensus)
			},
		},
		{
			name:        "Test_Blobber_Index_2_3_Error_On_Rename_Failure",
			numBlobbers: 4,
			numCorrect:  2,
			setup:       setupHttpResponses,
			wantErr:     true,
			errMsg:      "rename_failed",
		},
		{
			name:        "Test_All_Blobber_Error_On_Rename_Failure",
			numBlobbers: 4,
			numCorrect:  0,
			setup:       setupHttpResponses,
			wantErr:     true,
			errMsg:      "rename_failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			a := &Allocation{
				Tx:         "TestRenameRequest_ProcessRename",
				DataShards: numBlobbers,
			}

			setupMockAllocation(t, a)

			resp := &WMLockResult{
				Status: WMLockStatusOK,
			}

			respBuf, _ := json.Marshal(resp)
			m := make(devMock.ResponseMap)

			server := dev.NewBlobberServer(m)
			defer server.Close()

			for i := 0; i < numBlobbers; i++ {
				path := "/TestRenameRequest_ProcessRename" + tt.name + mockBlobberUrl + strconv.Itoa(i)

				m[http.MethodPost+":"+path+blobber.EndpointWriteMarkerLock+a.ID] = devMock.Response{
					StatusCode: http.StatusOK,
					Body:       respBuf,
				}

				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: server.URL + path,
				})
			}

			setupMockRollback(a, &mockClient)

			req := &RenameRequest{
				allocationObj:  a,
				blobbers:       a.Blobbers,
				allocationID:   mockAllocationId,
				allocationTx:   mockAllocationTxId,
				remotefilepath: mockRemoteFilePath,
				consensus: Consensus{
					RWMutex:         &sync.RWMutex{},
					consensusThresh: 3,
					fullconsensus:   4,
				},
				renameMask:   zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1),
				maskMU:       &sync.Mutex{},
				connectionID: mockConnectionId,
				newName:      mockNewName,
			}
			req.ctx, req.ctxCncl = context.WithCancel(context.TODO())
			tt.setup(t, tt.name, tt.numBlobbers, tt.numCorrect, req)
			err := req.ProcessRename()
			require.EqualValues(tt.wantErr, err != nil, err)
			if err != nil {
				require.Contains(errors.Top(err), tt.errMsg)
				return
			}
			if tt.wantFunc != nil {
				tt.wantFunc(require, req)
			}
		})
	}
}
