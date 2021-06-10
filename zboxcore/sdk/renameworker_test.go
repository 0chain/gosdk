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

const (
	renameWorkerTestDir = configDir + "/renameworker"
)

func TestRenameRequest_getObjectTreeFromBlobber(t *testing.T) {
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
				})).Return(&http.Response{
					Body: func() io.ReadCloser {
						jsonFR, err := json.Marshal(fileref.ReferencePath{
							Meta: map[string]interface{}{
								"type": mockType,
							},
						})
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
					StatusCode: http.StatusOK,
				}, nil)
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
			ar := &RenameRequest{
				allocationID:   mockAllocationId,
				allocationTx:   mockAllocationTxId,
				remotefilepath: mockRemoteFilePath,
				Consensus: Consensus{
					consensusThresh: 50,
					fullconsensus:   4,
				},
				ctx:          context.TODO(),
				renameMask:   0,
				connectionID: zboxutil.NewConnectionId(),
			}
			_, err := ar.getObjectTreeFromBlobber(blobber)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
		})
	}
}

func TestRenameRequest_renameBlobberObject(t *testing.T) {
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

	tests := []struct {
		name     string
		setup    func(t *testing.T, name string)
		wantErr  bool
		errMsg   string
		wantFunc func(require *require.Assertions, ar *RenameRequest)
	}{
		{
			name:    "Test_Error_New_HTTP_Failed_By_Containing_" + string([]byte{0x7f, 0, 0}),
			setup:   func(t *testing.T, testName string) {},
			wantErr: true,
			errMsg:  `parse "Test_Error_New_HTTP_Failed_By_Containing_\u007f\x00\x00": net/url: invalid control character in URL`,
		},
		{
			name: "Test_Error_Get_Object_Tree_From_Blobber_Failed",
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
			name: "Test_Rename_Blobber_Object_Failed",
			setup: func(t *testing.T, testName string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == "GET" &&
						strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
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
				}, nil)
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == "POST" &&
						strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			wantFunc: func(require *require.Assertions, ar *RenameRequest) {
				require.NotNil(ar)
				require.Equal(uint32(0), ar.renameMask)
				require.Equal(float32(0), ar.consensus)
			},
		},
		{
			name: "Test_Rename_Blobber_Object_Success",
			setup: func(t *testing.T, testName string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
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
				}, nil)
			},
			wantFunc: func(require *require.Assertions, ar *RenameRequest) {
				require.NotNil(ar)
				require.Equal(uint32(1), ar.renameMask)
				require.Equal(float32(1), ar.consensus)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name)
			ar := &RenameRequest{
				allocationID:   mockAllocationId,
				allocationTx:   mockAllocationTxId,
				remotefilepath: mockRemoteFilePath,
				Consensus: Consensus{
					consensusThresh: 50,
					fullconsensus:   4,
				},
				ctx:          context.TODO(),
				renameMask:   0,
				connectionID: zboxutil.NewConnectionId(),
			}
			ar.blobbers = append(ar.blobbers, &blockchain.StorageNode{
				Baseurl: tt.name,
			})
			_, err := ar.renameBlobberObject(ar.blobbers[0], 0)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
			if tt.wantFunc != nil {
				tt.wantFunc(require, ar)
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
	)

	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, renameWorkerTestDir, blobberMocks)
	defer cncl()

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponses := func(t *testing.T, testName string, numBlobbers int, numCorrect int) {
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
				return req.Method == "POST" &&
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
		willReturnCommitResult(&CommitResult{Success: true})
	}

	tests := []struct {
		name        string
		numBlobbers int
		numCorrect  int
		setup       func(*testing.T, string, int, int)
		wantErr     bool
		errMsg      string
		wantFunc    func(require *require.Assertions, ar *RenameRequest)
	}{
		{
			name:        "Test_All_Blobber_Rename_Success",
			numBlobbers: 4,
			numCorrect:  4,
			setup:       setupHttpResponses,
			wantErr:     false,
			wantFunc: func(require *require.Assertions, ar *RenameRequest) {
				require.NotNil(ar)
				require.Equal(uint32(15), ar.renameMask)
				require.Equal(float32(4), ar.consensus)
			},
		},
		{
			name:        "Test_Blobber_Index_3_Error_On_Rename_Success",
			numBlobbers: 4,
			numCorrect:  3,
			setup:       setupHttpResponses,
			wantErr:     false,
			wantFunc: func(require *require.Assertions, ar *RenameRequest) {
				require.NotNil(ar)
				require.Equal(uint32(7), ar.renameMask)
				require.Equal(float32(3), ar.consensus)
			},
		},
		{
			name:        "Test_Blobber_Index_2_3_Error_On_Rename_Failure",
			numBlobbers: 4,
			numCorrect:  2,
			setup:       setupHttpResponses,
			wantErr:     true,
			errMsg:      "Rename failed: Rename request failed. Operation failed.",
		},
		{
			name:        "Test_All_Blobber_Error_On_Rename_Failure",
			numBlobbers: 4,
			numCorrect:  0,
			setup:       setupHttpResponses,
			wantErr:     true,
			errMsg:      "Rename failed: Rename request failed. Operation failed.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name, tt.numBlobbers, tt.numCorrect)
			ar := &RenameRequest{
				allocationID:   mockAllocationId,
				allocationTx:   mockAllocationTxId,
				remotefilepath: mockRemoteFilePath,
				Consensus: Consensus{
					consensusThresh: 50,
					fullconsensus:   4,
				},
				ctx:          context.TODO(),
				renameMask:   0,
				connectionID: zboxutil.NewConnectionId(),
			}
			for i := 0; i < tt.numBlobbers; i++ {
				ar.blobbers = append(ar.blobbers, &blockchain.StorageNode{
					ID:      a.Blobbers[i].ID,
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			err := ar.ProcessRename()
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
				return
			}
			if tt.wantFunc != nil {
				tt.wantFunc(require, ar)
			}
		})
	}
}
