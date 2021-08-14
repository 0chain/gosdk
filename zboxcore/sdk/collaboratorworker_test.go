package sdk

import (
	"bytes"
	"context"
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
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCollaboratorRequest_UpdateCollaboratorToBlobbers(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockRemoteFilePath = "mock/remote/file/path"
		mockBlobberUrl     = "mockblobberurl"
		mockCollaboratorID = "mock collaborator id"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	tests := []struct {
		name        string
		numBlobbers int
		setup       func(*testing.T, string)
		want        bool
	}{
		{
			name:        "Test_Update_Collaborator_To_Blobbers_Failure",
			numBlobbers: 4,
			setup: func(t *testing.T, testName string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			want: false,
		},
		{
			name:        "Test_Update_Collaborator_To_Blobbers_Success",
			numBlobbers: 4,
			setup: func(t *testing.T, testName string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name)
			req := &CollaboratorRequest{
				a: &Allocation{
					Tx:  mockAllocationTxId,
					ctx: context.TODO(),
				},
				path:           mockRemoteFilePath,
				collaboratorID: mockCollaboratorID,
			}
			for i := 0; i < tt.numBlobbers; i++ {
				req.a.Blobbers = append(req.a.Blobbers, &blockchain.StorageNode{
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			got := req.UpdateCollaboratorToBlobbers()
			var check = require.False
			if tt.want {
				check = require.True
			}
			check(got)
		})
	}
}

func TestCollaboratorRequest_updateCollaboratorToBlobber(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockRemoteFilePath = "mock/remote/file/path"
		mockBlobberUrl     = "mockblobberurl"
		mockCollaboratorID = "mock collaborator id"
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
		requestFields map[string]string
	}

	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, parameters)
		want       bool
	}{
		{
			name: "Test_update_Collaborator_To_Blobber_Failure",
			setup: func(t *testing.T, testName string, p parameters) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			want: false,
		},
		{
			name: "Test_update_Collaborator_To_Blobber_Success",
			parameters: parameters{
				requestFields: map[string]string{
					"path":      mockRemoteFilePath,
					"collab_id": mockCollaboratorID,
				},
			},
			setup: func(t *testing.T, testName string, p parameters) {
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

					return strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name, tt.parameters)
			req := &CollaboratorRequest{
				a: &Allocation{
					Tx:  mockAllocationTxId,
					ctx: context.TODO(),
				},
				path:           mockRemoteFilePath,
				collaboratorID: mockCollaboratorID,
				wg:             func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			}
			req.a.Blobbers = append(req.a.Blobbers, &blockchain.StorageNode{
				Baseurl: tt.name + mockBlobberUrl,
			})
			rspCh := make(chan bool, 1)
			go req.updateCollaboratorToBlobber(req.a.Blobbers[0], 0, rspCh)
			resp := <-rspCh
			var check = require.False
			if tt.want {
				check = require.True
			}
			check(resp)
		})
	}
}

func TestCollaboratorRequest_RemoveCollaboratorFromBlobbers(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockRemoteFilePath = "mock/remote/file/path"
		mockBlobberUrl     = "mockblobberurl"
		mockCollaboratorID = "mock collaborator id"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	tests := []struct {
		name        string
		numBlobbers int
		setup       func(*testing.T, string)
		want        bool
	}{
		{
			name:        "Test_Remove_Collaborator_From_Blobbers_Failure",
			numBlobbers: 4,
			setup: func(t *testing.T, testName string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			want: false,
		},
		{
			name:        "Test_Remove_Collaborator_From_Blobbers_Success",
			numBlobbers: 4,
			setup: func(t *testing.T, testName string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name)
			req := &CollaboratorRequest{
				a: &Allocation{
					Tx:  mockAllocationTxId,
					ctx: context.TODO(),
				},
				path:           mockRemoteFilePath,
				collaboratorID: mockCollaboratorID,
			}
			for i := 0; i < tt.numBlobbers; i++ {
				req.a.Blobbers = append(req.a.Blobbers, &blockchain.StorageNode{
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			got := req.RemoveCollaboratorFromBlobbers()
			var check = require.False
			if tt.want {
				check = require.True
			}
			check(got)
		})
	}
}

func TestCollaboratorRequest_removeCollaboratorFromBlobber(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockRemoteFilePath = "mock/remote/file/path"
		mockBlobberUrl     = "mockblobberurl"
		mockCollaboratorID = "mock collaborator id"
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
		requestFields map[string]string
	}

	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, parameters)
		want       bool
	}{
		{
			name: "Test_remove_Collaborator_From_Blobber_Failure",
			setup: func(t *testing.T, testName string, p parameters) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			want: false,
		},
		{
			name: "Test_remove_Collaborator_From_Blobber_Success",
			parameters: parameters{
				requestFields: map[string]string{
					"path":      mockRemoteFilePath,
					"collab_id": mockCollaboratorID,
				},
			},
			setup: func(t *testing.T, testName string, p parameters) {
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

					return strings.HasPrefix(req.URL.Path, testName)
				})).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			tt.setup(t, tt.name, tt.parameters)
			req := &CollaboratorRequest{
				a: &Allocation{
					Tx:  mockAllocationTxId,
					ctx: context.TODO(),
				},
				path:           mockRemoteFilePath,
				collaboratorID: mockCollaboratorID,
				wg:             func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			}
			req.a.Blobbers = append(req.a.Blobbers, &blockchain.StorageNode{
				Baseurl: tt.name + mockBlobberUrl,
			})
			rspCh := make(chan bool, 1)
			go req.removeCollaboratorFromBlobber(req.a.Blobbers[0], 0, rspCh)
			resp := <-rspCh
			var check = require.False
			if tt.want {
				check = require.True
			}
			check(resp)
		})
	}
}
