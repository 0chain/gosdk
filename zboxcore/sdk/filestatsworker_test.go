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
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListRequest_getFileStatsInfoFromBlobber(t *testing.T) {
	const (
		mockFileStatsName  = "mock fileStats name"
		mockAllocationTxId = "mock transaction id"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
		mockRemoteFilePath = "mock/remote/file/path"
		mockAllocationId   = "mock allocation id"
		mockErrorMessage   = "mock error message"
		mockBlobberId      = "mock blobber Id"
		mockBlobberIndex   = 87
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	var client = zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	type parameters struct {
		fileStatsHttpResp FileStats
		fileStatsFinal    FileStats
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
			errMsg:  "file stats response parse error",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				fileStatsHttpResp: FileStats{
					Name: mockFileStatsName,
				},
				fileStatsFinal: FileStats{
					Name:       mockFileStatsName,
					BlobberID:  mockBlobberId,
					BlobberURL: "Test_Success",
				},
				blobberIdx:     mockBlobberIndex,
				respStatusCode: 200,
				requestFields: map[string]string{
					"path_hash": fileref.GetReferenceLookup(mockAllocationId, mockRemoteFilePath),
				},
			},
			setup: func(t *testing.T, testName string, p parameters, _ string) {
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
					require.NoError(t, err)
					require.True(t, strings.HasPrefix(mediaType, "multipart/"))
					reader := multipart.NewReader(req.Body, params["boundary"])
					var part *multipart.Part
					part, err = reader.NextPart()
					require.NoError(t, err)
					expected, ok := p.requestFields[part.FormName()]
					require.True(t, ok)
					actual, err := ioutil.ReadAll(part)
					require.NoError(t, err)
					require.EqualValues(t, expected, string(actual))

					sign, err := zclient.Sign(encryption.Hash(mockAllocationTxId))
					return req.URL.Path == "Test_Success"+zboxutil.FILE_STATS_ENDPOINT+mockAllocationTxId &&
						req.URL.RawPath == "Test_Success"+zboxutil.FILE_STATS_ENDPOINT+mockAllocationTxId &&
						req.Method == "POST" &&
						req.Header.Get("X-App-Client-ID") == mockClientId &&
						req.Header.Get("X-App-Client-Key") == mockClientKey &&
						req.Header.Get(zboxutil.CLIENT_SIGNATURE_HEADER) == sign &&
						testName == "Test_Success"
				})).Return(&http.Response{
					StatusCode: p.respStatusCode,
					Body: func(p parameters) io.ReadCloser {
						jsonFR, err := json.Marshal(p.fileStatsHttpResp)
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(p),
				}, nil).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t, tt.name, tt.parameters, tt.errMsg)
			blobber := blockchain.StorageNode{
				ID:      mockBlobberId,
				Baseurl: tt.name,
			}
			req := &ListRequest{
				allocationID:   mockAllocationId,
				allocationTx:   mockAllocationTxId,
				remotefilepath: mockRemoteFilePath,
				ctx:            context.Background(),
				wg:             &sync.WaitGroup{},
			}
			rspCh := make(chan *fileStatsResponse, 1)
			req.wg.Add(1)
			go req.getFileStatsInfoFromBlobber(&blobber, mockBlobberIndex, rspCh)
			req.wg.Wait()
			resp := <-rspCh
			require.EqualValues(t, tt.wantErr, resp.err != nil)
			if resp.err != nil {
				require.EqualValues(t, tt.errMsg, errors.Top(resp.err))
				return
			}
			require.EqualValues(t, tt.parameters.fileStatsFinal, *resp.filestats)
			require.EqualValues(t, tt.parameters.blobberIdx, resp.blobberIdx)
		})
	}
}

func TestListRequest_getFileStatsFromBlobbers(t *testing.T) {
	const (
		mockAllocationTxId = "mock transaction id"
		mockAllocationId   = "mock allocation id"
		mockFileRefName    = "mock file ref name"
		mockBlobberUrl     = "mockBlobberUrl"
		mockClientId       = "mock client id"
		mockClientKey      = "mock client key"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponses := func(t *testing.T, name string, numBlobbers, numCorrect int) {
		for i := 0; i < numBlobbers; i++ {
			frName := mockFileRefName + strconv.Itoa(i)
			url := name + mockBlobberUrl + strconv.Itoa(i)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.Path, url)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileStatsName string) io.ReadCloser {
					jsonFR, err := json.Marshal(&FileStats{
						Name: fileStatsName,
					})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(frName),
			}, nil)
		}
	}

	tests := []struct {
		name        string
		numBlobbers int
		numCorrect  int

		setup             func(*testing.T, string, int, int)
		httpRespFileStats []FileStats
		wantFileStats     []FileStats
		wantErr           bool
	}{
		{
			name:        "Test_Success",
			numBlobbers: 10,
			setup:       setupHttpResponses,
			httpRespFileStats: func(number int) []FileStats {
				var fileStats []FileStats
				for i := 0; i < number; i++ {
					fileStats = append(fileStats, FileStats{
						Name: mockFileRefName + strconv.Itoa(i),
					})
				}
				return fileStats
			}(10),
			wantFileStats: func(number int) []FileStats {
				var fileStats []FileStats
				for i := 0; i < number; i++ {
					fileStats = append(fileStats, FileStats{
						Name:       mockFileRefName + strconv.Itoa(i),
						BlobberID:  strconv.Itoa(i),
						BlobberURL: "Test_Success" + mockBlobberUrl + strconv.Itoa(i),
					})
				}
				return fileStats
			}(10),
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
			}
			for i := 0; i < tt.numBlobbers; i++ {
				req.blobbers = append(req.blobbers, &blockchain.StorageNode{
					ID:      strconv.Itoa(i),
					Baseurl: tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			mapResp := req.getFileStatsFromBlobbers()
			for _, fs := range mapResp {
				index, err := strconv.Atoi(fs.BlobberID)
				require.NoError(t, err)
				require.EqualValues(t, tt.wantFileStats[index], *fs)
			}
		})
	}
}
