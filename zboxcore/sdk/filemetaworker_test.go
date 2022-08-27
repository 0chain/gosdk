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
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListRequest_getFileMetaInfoFromBlobber(t *testing.T) {
	const mockFileRefName = "mock fileRef name"
	const mockAllocationTxId = "mock transaction id"
	const mockClientId = "mock client id"
	const mockClientKey = "mock client key"
	const mockRemoteFilePath = "mock/remote/file/path"
	const mockSignature = "mock signature"
	const mockAllocationId = "mock allocation id"
	const mockErrorMessage = "mock error message"

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		fileRefToRetrieve fileref.FileRef
		respStatusCode    int
		requestFields     map[string]string
		blobberIdx        int
	}

	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, parameters, string)
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Http_Error",
			parameters: parameters{
				respStatusCode: 0,
			},
			setup: func(t *testing.T, name string, p parameters, errMsg string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "Test_Http_Error")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
					StatusCode: p.respStatusCode,
				}, errors.New("", mockErrorMessage))
			},
			wantErr: true,
			errMsg:  mockErrorMessage,
		},
		{
			name: "Test_Badly_Formatted",
			parameters: parameters{
				respStatusCode: 200,
			},
			setup: func(t *testing.T, name string, p parameters, errMsg string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "Test_Badly_Formatted")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
					StatusCode: p.respStatusCode,
				}, nil)
			},
			wantErr: true,
			errMsg:  "file meta data response parse error",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				fileRefToRetrieve: fileref.FileRef{
					Ref: fileref.Ref{
						Name: mockFileRefName,
					},
				},
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
				respStatusCode: http.StatusOK,
				blobberIdx:     73,
			},
			setup: func(t *testing.T, testName string, p parameters, errMsg string) {
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

					return req.URL.Path == "Test_Success"+zboxutil.FILE_META_ENDPOINT+mockAllocationTxId &&
						req.URL.RawPath == "Test_Success"+zboxutil.FILE_META_ENDPOINT+mockAllocationTxId &&
						req.Method == "POST" &&
						req.Header.Get("X-App-Client-ID") == mockClientId &&
						req.Header.Get("X-App-Client-Key") == mockClientKey &&
						testName == "Test_Success"
				})).Return(&http.Response{
					StatusCode: p.respStatusCode,
					Body: func(p parameters) io.ReadCloser {
						jsonFR, err := json.Marshal(p.fileRefToRetrieve)
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(p),
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.setup(t, tt.name, tt.parameters, tt.errMsg)
			blobber := &blockchain.StorageNode{
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
			rspCh := make(chan *fileMetaResponse, 1)
			req.wg.Add(1)
			go req.getFileMetaInfoFromBlobber(blobber, 73, rspCh)
			req.wg.Wait()
			resp := <-rspCh
			require.EqualValues(t, tt.wantErr, resp.err != nil)
			if resp.err != nil {
				require.EqualValues(t, tt.errMsg, errors.Top(resp.err))
				return
			}
			require.EqualValues(t, tt.parameters.fileRefToRetrieve, *resp.fileref)
			require.EqualValues(t, tt.parameters.blobberIdx, resp.blobberIdx)
		})
	}
}

func TestListRequest_getFileConsensusFromBlobbers(t *testing.T) {
	const mockAllocationTxId = "mock transaction id"
	const mockAllocationId = "mock allocation id"
	const mockFileRefName = "mock file ref name"
	const mockBlobberUrl = "mockBlobberUrl"
	const mockActualHash = "mockActualHash"

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	const mockClientId = "mock client id"
	const mockClientKey = "mock client key"
	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponses := func(t *testing.T, name string, numBlobbers, numCorrect int) {
		require.True(t, numBlobbers >= numCorrect)
		for i := 0; i < numBlobbers; i++ {
			var hash string
			if i < numCorrect {
				hash = mockActualHash
			}
			frName := mockFileRefName + strconv.Itoa(i)
			url := mockBlobberUrl + strconv.Itoa(i)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.Path, name+url)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileRefName, hash string) io.ReadCloser {
					jsonFR, err := json.Marshal(&fileref.FileRef{
						ActualFileHash: hash,
						Ref: fileref.Ref{
							Name: fileRefName,
						},
					})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(frName, hash),
			}, nil)
		}
	}

	tests := []struct {
		name        string
		numBlobbers int
		consensus   Consensus
		numCorrect  int

		setup   func(*testing.T, string, int, int)
		wantErr bool
	}{
		{
			name:        "Fail_Consensus",
			numBlobbers: 10,
			consensus: Consensus{
				consensusThresh: 2,
				fullconsensus:   50,
			},
			numCorrect: 5,
			setup:      setupHttpResponses,
			wantErr:    true,
		},
		{
			name:        "Pass_Consensus",
			numBlobbers: 10,
			consensus: Consensus{
				consensusThresh: 2,
				fullconsensus:   50,
			},
			numCorrect: 6,
			setup:      setupHttpResponses,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t, tt.name, tt.numBlobbers, tt.numCorrect)
			req := &ListRequest{
				allocationID: mockAllocationId,
				allocationTx: mockAllocationTxId,
				ctx:          context.TODO(),
				blobbers:     []*blockchain.StorageNode{},
				wg:           &sync.WaitGroup{},
				Consensus:    tt.consensus,
			}
			for i := 0; i < tt.numBlobbers; i++ {
				req.blobbers = append(req.blobbers, &blockchain.StorageNode{
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}

			_, mask, fileRefs := req.getFileConsensusFromBlobbers()

			if mask == nil && fileRefs == nil {
				require.True(t, tt.wantErr)
				return
			}
			require.Len(t, fileRefs, tt.numBlobbers)
			for i, actual := range fileRefs {
				expected := fileref.FileRef{}
				expected.Name = mockFileRefName + strconv.Itoa(i)
				if i < tt.numCorrect {
					expected.ActualFileHash = mockActualHash
				}
				require.EqualValues(t, expected, *actual.fileref)
			}
		})
	}
}
