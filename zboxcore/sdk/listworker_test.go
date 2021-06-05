package sdk

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	listWorkerTestDir = configDir + "/listworker"
)

func TestListRequest_getListInfoFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, listWorkerTestDir, blobberMocks)
	defer cncl()
	var wg sync.WaitGroup
	tests := []struct {
		name           string
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		want           bool
		wantErr        bool
	}{
		{
			"Test_Error_New_HTTP_Failed",
			func(t *testing.T) (teardown func(t *testing.T)) {
				url := a.Blobbers[0].Baseurl
				a.Blobbers[0].Baseurl = string([]byte{0x7f, 0, 0})
				return func(t *testing.T) {
					a.Blobbers[0].Baseurl = url
				}
			},
			false,
			true,
		},
		{
			"Test_HTTP_Response_Failed",
			nil,
			false,
			false,
		},
		{
			"Test_Error_HTTP_Response_Not_JSON_Format",
			func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				m.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("This is not JSON format")),
				}, nil)
				return nil
			},
			false,
			false,
		},
		{
			"Test_Success",
			func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"allocation_root":"47786ba91e51589e67bc3c8de7d8aff0dfddb399f826f6332c8f2b23e0e26420","meta_data":{"created_at":"0001-01-01T00:00:00Z","hash":"","lookup_hash":"","name":"","num_of_blocks":0,"path":"/1.txt","path_hash":"","size":0,"type":"d","updated_at":"0001-01-01T00:00:00Z"},"list":[]}`
				m.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
				}, nil)
				return nil
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t); teardown != nil {
					defer teardown(t)
				}
			}
			req := &ListRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				ctx:            context.Background(),
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				wg: func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			}
			rspCh := make(chan *listResponse, 1)
			req.getListInfoFromBlobber(req.blobbers[0], 0, rspCh)
			resp := <-rspCh
			var expectedResult *fileref.Ref
			parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__Test_Success.json", listWorkerTestDir, "getListInfoFromBlobber"), &expectedResult)
			if tt.wantErr {
				require.Error(resp.err, "expected error != nil")
				return
			}
			if !tt.want {
				require.NotEqual(expectedResult, resp.ref)
				return
			}
			require.EqualValues(expectedResult, resp.ref)
		})
	}
}

func TestListRequest_GetListFromBlobbers(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, listWorkerTestDir, blobberMocks)
	defer cncl()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		want           bool
		wantFunc       func(require *require.Assertions, req *ListRequest)
	}{
		{
			"Test_Error_Get_List_File_From_Blobbers_Failed",
			nil,
			false,
			func(require *require.Assertions, req *ListRequest) {
				require.NotNil(req)
				require.Equal(float32(0), req.consensus)
			},
		},
		{
			"Test_Success",
			func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"allocation_root":"1c6870fa9d3c371d9e219c73c4a52cf04b8ab6036460879fee598b68a02febf2","meta_data":{"created_at":"0001-01-01T00:00:00Z","hash":"","lookup_hash":"","name":"","num_of_blocks":0,"path":"/1.txt","path_hash":"","size":0,"type":"d","updated_at":"0001-01-01T00:00:00Z"},"list":[]}`
				statusCode := http.StatusOK
				mockCall := m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "GET" }))
				mockCall.RunFn = func(args mock.Arguments) {
					req := args[0].(*http.Request)
					url := req.URL.Host
					switch url {
					case strings.ReplaceAll(a.Blobbers[2].Baseurl, "http://", ""):
						statusCode = http.StatusBadRequest
					case strings.ReplaceAll(a.Blobbers[3].Baseurl, "http://", ""):
						statusCode = http.StatusBadRequest
					default:
						statusCode = http.StatusOK
					}
					mockCall.ReturnArguments = mock.Arguments{&http.Response{
						StatusCode: statusCode,
						Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
					}, nil}
				}
				return nil
			},
			true,
			func(require *require.Assertions, req *ListRequest) {
				require.NotNil(req)
				require.Equal(float32(4), req.consensus)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t); teardown != nil {
					defer teardown(t)
				}
			}
			req := &ListRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				ctx:            context.Background(),
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
			}
			got := req.GetListFromBlobbers()
			var expectedResult *ListResult
			parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__Test_Success.json", listWorkerTestDir, "GetListFromBlobbers"), &expectedResult)
			if tt.want {
				require.EqualValues(expectedResult, got)
				return
			}
			tt.wantFunc(require, req)
		})
	}
}
