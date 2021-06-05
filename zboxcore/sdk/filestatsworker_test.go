package sdk

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	filestatsWorkerTestDir = configDir + "/filestatsworker"
)

func TestListRequest_getFileStatsInfoFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, filestatsWorkerTestDir, blobberMocks)
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
			true,
		},
		{
			"Test_Success",
			func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"ID":294,"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{},"commit_meta_txns":null,"content_hash":"adc83b19e793491b1c6ea0fd8b46cd9f32e592fc","created_at":"2021-03-23T18:56:28.318478Z","custom_meta":"","encrypted_key":"","hash":"b1fba32dfc8025a7390b05c5eb3aea2ec3a84ed5b3ce60b49093bc001cfc9710","last_challenge_txn":"","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"6ed726c5aaf50067479a105ad9c4330bfa341f1fd889e3552af67303712ee0f0","mimetype":"application/octet-stream","name":"1.txt","num_of_block_downloads":0,"num_of_blocks":1,"num_of_challenges":0,"num_of_failed_challenges":0,"num_of_updates":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-23T18:56:28.318478Z","write_marker_txn":"e3eaaa98d374931b8fd3f52096e7e47f68eedc4267d387a4a8b999a52b0f603b"}`
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
			rspCh := make(chan *fileStatsResponse, 1)
			go req.getFileStatsInfoFromBlobber(req.blobbers[0], 0, rspCh)
			resp := <-rspCh
			if tt.wantErr {
				require.Error(resp.err, "expected error != nil")
				return
			}
			if !tt.want {
				require.Nil(resp.filestats, "expected nullable file stats result")
				return
			}
			require.NotNil(resp.filestats, "unexpected nullable file stats result")
		})
	}
}

func TestListRequest_getFileStatsFromBlobbers(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, filestatsWorkerTestDir, blobberMocks)
	defer cncl()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		want           bool
	}{
		{
			"Test_Error_Getting_File_Stats_From_Blobbers_Failed",
			nil,
			false,
		},
		{
			"Test_Success",
			func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"ID":${ID},"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{},"commit_meta_txns":null,"content_hash":"adc83b19e793491b1c6ea0fd8b46cd9f32e592fc","created_at":"2021-03-23T18:56:28.318478Z","custom_meta":"","encrypted_key":"","hash":"b1fba32dfc8025a7390b05c5eb3aea2ec3a84ed5b3ce60b49093bc001cfc9710","last_challenge_txn":"","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"6ed726c5aaf50067479a105ad9c4330bfa341f1fd889e3552af67303712ee0f0","mimetype":"application/octet-stream","name":"1.txt","num_of_block_downloads":0,"num_of_blocks":1,"num_of_challenges":0,"num_of_failed_challenges":0,"num_of_updates":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-23T18:56:28.318478Z","write_marker_txn":"e3eaaa98d374931b8fd3f52096e7e47f68eedc4267d387a4a8b999a52b0f603b"}`
				mockCall := m.On("Do", mock.AnythingOfType("*http.Request"))
				mockCall.RunFn = func(args mock.Arguments) {
					req := args[0].(*http.Request)
					url := req.URL.Host
					switch url {
					case strings.ReplaceAll(a.Blobbers[0].Baseurl, "http://", ""):
						bodyString = strings.ReplaceAll(bodyString, "${ID}", "294")
					case strings.ReplaceAll(a.Blobbers[1].Baseurl, "http://", ""):
						bodyString = strings.ReplaceAll(bodyString, "${ID}", "257")
					case strings.ReplaceAll(a.Blobbers[2].Baseurl, "http://", ""):
						bodyString = strings.ReplaceAll(bodyString, "${ID}", "53")
					case strings.ReplaceAll(a.Blobbers[3].Baseurl, "http://", ""):
						bodyString = strings.ReplaceAll(bodyString, "${ID}", "86")
					}
					mockCall.ReturnArguments = mock.Arguments{&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
					}, nil}
				}
				return nil
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.additionalMock != nil {
				if teardown := tt.additionalMock(t); teardown != nil {
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
			got := req.getFileStatsFromBlobbers()
			if !tt.want {
				for _, blobberMock := range blobberMocks {
					require.Emptyf(got[blobberMock.ID], "expected empty value of file stats related to blobber %v", blobberMock.ID)
				}
				return
			}
			require.NotNil(got, "unexpected nullable file stats result")
			require.Equalf(4, len(got), "expected length of file stats result is %d, but got %v", 4, len(got))
			for _, blobberMock := range blobberMocks {
				require.NotEmptyf(got[blobberMock.ID], "unexpected empty value of file stats related to blobber %v", blobberMock.ID)
			}
		})
	}
}
