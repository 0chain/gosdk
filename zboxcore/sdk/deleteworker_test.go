package sdk

import (
	"context"
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
	deleteWorkerTestDir = configDir + "/deleteworker"
)

func TestDeleteRequest_ProcessDelete(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, deleteWorkerTestDir, blobberMocks)
	defer cncl()
	tests := []struct {
		name            string
		additionalMock  func(t *testing.T) (teardown func(t *testing.T))
		wantFunc        func(require *require.Assertions, ar *DeleteRequest)
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "Test_All_Blobber_Delete_Attribute_Success",
			additionalMock: func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{},"commit_meta_txns":null,"content_hash":"adc83b19e793491b1c6ea0fd8b46cd9f32e592fc","created_at":"2021-03-23T18:47:03.367331Z","custom_meta":"","encrypted_key":"","hash":"b1fba32dfc8025a7390b05c5eb3aea2ec3a84ed5b3ce60b49093bc001cfc9710","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"6ed726c5aaf50067479a105ad9c4330bfa341f1fd889e3552af67303712ee0f0","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-23T18:47:03.367331Z"},"latest_write_marker":{"allocation_root":"7f2f17e2c946896933175e318413e9a2c7e91b54ef037288d20e0bea93589e65","prev_allocation_root":"19a0711f93583d614e5f4a57008c89d756cd315a085400d9b3e0a58f7d13a423","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":1,"blobber_id":"${blobber_id_1}","timestamp":1616525223,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"ee5d83a5f7d6582454e58bb0fb2bb76fa6f3e5a1c74e62e2489259d300ae25a2"}}`
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "GET" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
						}, nil}
					}
				})
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "DELETE" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader("")),
						}, nil}
					}
				})
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr: false,
			wantFunc: func(require *require.Assertions, ar *DeleteRequest) {
				require.NotNil(ar)
				require.Equal(uint32(15), ar.deleteMask)
				require.Equal(float32(4), ar.consensus)
			},
		},
		{
			name: "Test_Blobber_Index_0_Error_On_Delete_Attribute_Success",
			additionalMock: func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{},"commit_meta_txns":null,"content_hash":"adc83b19e793491b1c6ea0fd8b46cd9f32e592fc","created_at":"2021-03-23T18:47:03.367331Z","custom_meta":"","encrypted_key":"","hash":"b1fba32dfc8025a7390b05c5eb3aea2ec3a84ed5b3ce60b49093bc001cfc9710","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"6ed726c5aaf50067479a105ad9c4330bfa341f1fd889e3552af67303712ee0f0","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-23T18:47:03.367331Z"},"latest_write_marker":{"allocation_root":"7f2f17e2c946896933175e318413e9a2c7e91b54ef037288d20e0bea93589e65","prev_allocation_root":"19a0711f93583d614e5f4a57008c89d756cd315a085400d9b3e0a58f7d13a423","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":1,"blobber_id":"${blobber_id_1}","timestamp":1616525223,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"ee5d83a5f7d6582454e58bb0fb2bb76fa6f3e5a1c74e62e2489259d300ae25a2"}}`
				statusCode := http.StatusOK
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "GET" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
						}, nil}
					}
				})
				mockCall := m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "DELETE" }))
				mockCall.RunFn = func(args mock.Arguments) {
					req := args[0].(*http.Request)
					url := req.URL.Host
					switch url {
					case strings.ReplaceAll(a.Blobbers[0].Baseurl, "http://", ""):
						statusCode = http.StatusBadRequest
					default:
						statusCode = http.StatusOK
					}
					mockCall.ReturnArguments = mock.Arguments{&http.Response{
						StatusCode: statusCode,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil}
				}
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr: false,
			wantFunc: func(require *require.Assertions, ar *DeleteRequest) {
				require.NotNil(ar)
				require.Equal(uint32(14), ar.deleteMask)
				require.Equal(float32(3), ar.consensus)
			},
		},
		{
			name: "Test_Blobber_Index_0_2_Error_On_Delete_Attribute_Failure",
			additionalMock: func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{},"commit_meta_txns":null,"content_hash":"adc83b19e793491b1c6ea0fd8b46cd9f32e592fc","created_at":"2021-03-23T18:47:03.367331Z","custom_meta":"","encrypted_key":"","hash":"b1fba32dfc8025a7390b05c5eb3aea2ec3a84ed5b3ce60b49093bc001cfc9710","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"6ed726c5aaf50067479a105ad9c4330bfa341f1fd889e3552af67303712ee0f0","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-23T18:47:03.367331Z"},"latest_write_marker":{"allocation_root":"7f2f17e2c946896933175e318413e9a2c7e91b54ef037288d20e0bea93589e65","prev_allocation_root":"19a0711f93583d614e5f4a57008c89d756cd315a085400d9b3e0a58f7d13a423","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":1,"blobber_id":"${blobber_id_1}","timestamp":1616525223,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"ee5d83a5f7d6582454e58bb0fb2bb76fa6f3e5a1c74e62e2489259d300ae25a2"}}`
				statusCode := http.StatusOK
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "GET" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
						}, nil}
					}
				})
				mockCall := m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "DELETE" }))
				mockCall.RunFn = func(args mock.Arguments) {
					req := args[0].(*http.Request)
					url := req.URL.Host
					switch url {
					case strings.ReplaceAll(a.Blobbers[0].Baseurl, "http://", ""):
						statusCode = http.StatusBadRequest
					case strings.ReplaceAll(a.Blobbers[2].Baseurl, "http://", ""):
						statusCode = http.StatusBadRequest
					default:
						statusCode = http.StatusOK
					}
					mockCall.ReturnArguments = mock.Arguments{&http.Response{
						StatusCode: statusCode,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil}
				}
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr:         true,
			wantErrContains: "Delete failed",
		},
		{
			name: "Test_All_Blobber_Error_On_Delete_Attribute_Failure",
			additionalMock: func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{},"commit_meta_txns":null,"content_hash":"adc83b19e793491b1c6ea0fd8b46cd9f32e592fc","created_at":"2021-03-23T18:47:03.367331Z","custom_meta":"","encrypted_key":"","hash":"b1fba32dfc8025a7390b05c5eb3aea2ec3a84ed5b3ce60b49093bc001cfc9710","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"6ed726c5aaf50067479a105ad9c4330bfa341f1fd889e3552af67303712ee0f0","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-23T18:47:03.367331Z"},"latest_write_marker":{"allocation_root":"7f2f17e2c946896933175e318413e9a2c7e91b54ef037288d20e0bea93589e65","prev_allocation_root":"19a0711f93583d614e5f4a57008c89d756cd315a085400d9b3e0a58f7d13a423","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":1,"blobber_id":"${blobber_id_1}","timestamp":1616525223,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"ee5d83a5f7d6582454e58bb0fb2bb76fa6f3e5a1c74e62e2489259d300ae25a2"}}`
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "GET" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
						}, nil}
					}
				})
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "DELETE" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusBadRequest,
							Body:       ioutil.NopCloser(strings.NewReader("")),
						}, nil}
					}
				})
				willReturnCommitResult(&CommitResult{Success: true})
				return nil
			},
			wantErr:         true,
			wantErrContains: "Delete failed",
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
			req := &DeleteRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				ctx:          context.Background(),
				connectionID: zboxutil.NewConnectionId(),
			}

			err := req.ProcessDelete()

			if tt.wantFunc != nil {
				tt.wantFunc(require, req)
			}
			if tt.wantErr {
				require.Error(err, "expected error != nil")
				require.Contains(err.Error(), tt.wantErrContains, "expected error contains '%s'", tt.wantErrContains)
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
			if tt.wantFunc != nil {
				tt.wantFunc(require, req)
			}
		})
	}
}

func TestDeleteRequest_deleteBlobberFile(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, deleteWorkerTestDir, blobberMocks)
	defer cncl()
	var wg sync.WaitGroup
	tests := []struct {
		name           string
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		wantFunc       func(require *require.Assertions, ar *DeleteRequest)
	}{
		{
			"Test_Delete_Blobber_File_Failed",
			func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{},"commit_meta_txns":null,"content_hash":"adc83b19e793491b1c6ea0fd8b46cd9f32e592fc","created_at":"2021-03-23T18:47:03.367331Z","custom_meta":"","encrypted_key":"","hash":"b1fba32dfc8025a7390b05c5eb3aea2ec3a84ed5b3ce60b49093bc001cfc9710","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"6ed726c5aaf50067479a105ad9c4330bfa341f1fd889e3552af67303712ee0f0","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-23T18:47:03.367331Z"},"latest_write_marker":{"allocation_root":"7f2f17e2c946896933175e318413e9a2c7e91b54ef037288d20e0bea93589e65","prev_allocation_root":"19a0711f93583d614e5f4a57008c89d756cd315a085400d9b3e0a58f7d13a423","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":1,"blobber_id":"${blobber_id_1}","timestamp":1616525223,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"ee5d83a5f7d6582454e58bb0fb2bb76fa6f3e5a1c74e62e2489259d300ae25a2"}}`
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "GET" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
						}, nil}
					}
				})
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "DELETE" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusBadRequest,
							Body:       ioutil.NopCloser(strings.NewReader("")),
						}, nil}
					}
				})
				return nil
			},
			func(require *require.Assertions, ar *DeleteRequest) {
				require.NotNil(ar)
				require.Equal(uint32(0), ar.deleteMask)
				require.Equal(float32(0), ar.consensus)
			},
		},
		{
			"Test_Delete_Blobber_File_Success",
			func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{},"commit_meta_txns":null,"content_hash":"adc83b19e793491b1c6ea0fd8b46cd9f32e592fc","created_at":"2021-03-23T18:47:03.367331Z","custom_meta":"","encrypted_key":"","hash":"b1fba32dfc8025a7390b05c5eb3aea2ec3a84ed5b3ce60b49093bc001cfc9710","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"6ed726c5aaf50067479a105ad9c4330bfa341f1fd889e3552af67303712ee0f0","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-23T18:47:03.367331Z"},"latest_write_marker":{"allocation_root":"7f2f17e2c946896933175e318413e9a2c7e91b54ef037288d20e0bea93589e65","prev_allocation_root":"19a0711f93583d614e5f4a57008c89d756cd315a085400d9b3e0a58f7d13a423","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":1,"blobber_id":"${blobber_id_1}","timestamp":1616525223,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"ee5d83a5f7d6582454e58bb0fb2bb76fa6f3e5a1c74e62e2489259d300ae25a2"}}`
				m.On("Do", mock.AnythingOfType("*http.Request")).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
						}, nil}
					}
				})
				return nil
			},
			func(require *require.Assertions, ar *DeleteRequest) {
				require.NotNil(ar)
				require.Equal(uint32(1), ar.deleteMask)
				require.Equal(float32(1), ar.consensus)
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
			req := &DeleteRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				ctx:          context.Background(),
				connectionID: zboxutil.NewConnectionId(),
				wg:           func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			}
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

func TestDeleteRequest_getObjectTreeFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, deleteWorkerTestDir, blobberMocks)
	defer cncl()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Get_Object_Tree_From_Blobber_Failed",
			nil,
			true,
		},
		{
			"Test_Get_Object_Tree_From_Blobber_Success",
			func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{"who_pays_for_reads":1},"commit_meta_txns":null,"content_hash":"adc83b19e793491b1c6ea0fd8b46cd9f32e592fc","created_at":"2021-03-17T08:15:36.018029Z","custom_meta":"","encrypted_key":"","hash":"4fea25c7390d0d8374fd84c77345eee7037224c2b162e98950a9ea6d882e91e5","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"6ed726c5aaf50067479a105ad9c4330bfa341f1fd889e3552af67303712ee0f0","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-19T08:34:35.387704Z"},"latest_write_marker":{"allocation_root":"d069d04ba45c4f14764c699b69f07897d0973afaab5aa2cf9c302c849ccad955","prev_allocation_root":"fe2f7f060dd34adbfdcf2f142acd33f50401207f7363359dd35460be7a2bec2d","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":0,"blobber_id":"8eeb4d4d6621ea87ecf4d3ba61bb69d30db435c8ed7cbaccccd56b93ea050c42","timestamp":1616142875,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"ebea73ea057d334c7600ce65f323843d02ea4bd822c858ab004aa0f30043e11f"}}`
				m.On("Do", mock.AnythingOfType("*http.Request")).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
						}, nil}
					}
				})
				return nil
			},
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
			req := &DeleteRequest{
				allocationID: a.ID,
				allocationTx: a.Tx,
				blobbers:     a.Blobbers,
				ctx:          context.Background(),
				connectionID: zboxutil.NewConnectionId(),
			}
			objectTreeRefs := make([]fileref.RefEntity, 1)
			refEntity, _ := req.getObjectTreeFromBlobber(req.blobbers[0])
			objectTreeRefs[0] = refEntity
			_, err := req.getObjectTreeFromBlobber(req.blobbers[0])
			if tt.wantErr {
				require.Error(err, "expected error != nil")
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
		})
	}
}
