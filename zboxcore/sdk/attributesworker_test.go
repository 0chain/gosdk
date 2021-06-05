package sdk

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"

	// "github.com/0chain/gosdk/zboxcore/zboxutil/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	attributeWorkerTestDir = configDir + "/attributesworker"
)

func TestAttributesRequest_ProcessAttributes(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, attributeWorkerTestDir, blobberMocks)
	defer cncl()
	tests := []struct {
		name            string
		additionalMock  func(t *testing.T) (teardown func(t *testing.T))
		wantErr         bool
		wantErrContains string
		wantFunc        func(require *require.Assertions, ar *AttributesRequest)
	}{
		{
			name: "Test_All_Blobber_Update_Attribute_Success",
			additionalMock: func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{"who_pays_for_reads":1},"commit_meta_txns":null,"content_hash":"3a52ce780950d4d969792a2559cd519d7ee8c727","created_at":"2021-03-17T08:15:36.137135Z","custom_meta":"","encrypted_key":"","hash":"49f57d8a02ebcc96df36ef676201d7f96d79365a83a6c239432e7b63a03d5d36","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"ea13052ab648c94a2fc001ce4f6f5f2d8bb699d4b69264b361c45324c88da744","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-19T08:34:35.765008Z"},"latest_write_marker":{"allocation_root":"418132ae676240069a77e3785e046c45aa022fc771a33ab74da3b19f3b44d5f9","prev_allocation_root":"3e89fba8c0e47c24f6bd06bbef9b4eef466a7bdbd0da72ae9e511c5768d86f5c","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":0,"blobber_id":"9cb4a21362291d2d2ca8ebca1877bd60d63d51d8f12ecec1e55964df452d0e4a","timestamp":1616142875,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"15deae584c0f8ea3809e383c61c30e51c09b3af732878e8f2da7c08a21b4b19f"}}`
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "GET" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
						}, nil}
					}
				})
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "POST" })).Run(func(args mock.Arguments) {
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
			wantFunc: func(require *require.Assertions, ar *AttributesRequest) {
				require.NotNil(ar)
				require.Equal(uint32(15), ar.attributesMask)
				require.Equal(float32(4), ar.consensus)
			},
		},
		{
			name: "Test_Blobber_Index_0_Error_On_Update_Attribute_Success",
			additionalMock: func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{"who_pays_for_reads":1},"commit_meta_txns":null,"content_hash":"3a52ce780950d4d969792a2559cd519d7ee8c727","created_at":"2021-03-17T08:15:36.137135Z","custom_meta":"","encrypted_key":"","hash":"49f57d8a02ebcc96df36ef676201d7f96d79365a83a6c239432e7b63a03d5d36","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"ea13052ab648c94a2fc001ce4f6f5f2d8bb699d4b69264b361c45324c88da744","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-19T08:34:35.765008Z"},"latest_write_marker":{"allocation_root":"418132ae676240069a77e3785e046c45aa022fc771a33ab74da3b19f3b44d5f9","prev_allocation_root":"3e89fba8c0e47c24f6bd06bbef9b4eef466a7bdbd0da72ae9e511c5768d86f5c","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":0,"blobber_id":"9cb4a21362291d2d2ca8ebca1877bd60d63d51d8f12ecec1e55964df452d0e4a","timestamp":1616142875,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"15deae584c0f8ea3809e383c61c30e51c09b3af732878e8f2da7c08a21b4b19f"}}`
				statusCode := http.StatusOK
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "GET" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
						}, nil}
					}
				})
				mockCall := m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "POST" }))
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
			wantFunc: func(require *require.Assertions, ar *AttributesRequest) {
				require.NotNil(ar)
				require.Equal(uint32(14), ar.attributesMask)
				require.Equal(float32(3), ar.consensus)
			},
		},
		{
			name: "Test_Blobber_Index_0_2_Error_On_Update_Attribute_Failure",
			additionalMock: func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{"who_pays_for_reads":1},"commit_meta_txns":null,"content_hash":"3a52ce780950d4d969792a2559cd519d7ee8c727","created_at":"2021-03-17T08:15:36.137135Z","custom_meta":"","encrypted_key":"","hash":"49f57d8a02ebcc96df36ef676201d7f96d79365a83a6c239432e7b63a03d5d36","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"ea13052ab648c94a2fc001ce4f6f5f2d8bb699d4b69264b361c45324c88da744","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-19T08:34:35.765008Z"},"latest_write_marker":{"allocation_root":"418132ae676240069a77e3785e046c45aa022fc771a33ab74da3b19f3b44d5f9","prev_allocation_root":"3e89fba8c0e47c24f6bd06bbef9b4eef466a7bdbd0da72ae9e511c5768d86f5c","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":0,"blobber_id":"9cb4a21362291d2d2ca8ebca1877bd60d63d51d8f12ecec1e55964df452d0e4a","timestamp":1616142875,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"15deae584c0f8ea3809e383c61c30e51c09b3af732878e8f2da7c08a21b4b19f"}}`
				statusCode := http.StatusOK
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "GET" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
						}, nil}
					}
				})
				mockCall := m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "POST" }))
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
			wantErrContains: "Update attributes failed",
		},
		{
			name: "Test_All_Blobber_Error_On_Update_Attribute_Failure",
			additionalMock: func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{"who_pays_for_reads":1},"commit_meta_txns":null,"content_hash":"3a52ce780950d4d969792a2559cd519d7ee8c727","created_at":"2021-03-17T08:15:36.137135Z","custom_meta":"","encrypted_key":"","hash":"49f57d8a02ebcc96df36ef676201d7f96d79365a83a6c239432e7b63a03d5d36","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"ea13052ab648c94a2fc001ce4f6f5f2d8bb699d4b69264b361c45324c88da744","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-19T08:34:35.765008Z"},"latest_write_marker":{"allocation_root":"418132ae676240069a77e3785e046c45aa022fc771a33ab74da3b19f3b44d5f9","prev_allocation_root":"3e89fba8c0e47c24f6bd06bbef9b4eef466a7bdbd0da72ae9e511c5768d86f5c","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":0,"blobber_id":"9cb4a21362291d2d2ca8ebca1877bd60d63d51d8f12ecec1e55964df452d0e4a","timestamp":1616142875,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"15deae584c0f8ea3809e383c61c30e51c09b3af732878e8f2da7c08a21b4b19f"}}`
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "GET" })).Run(func(args mock.Arguments) {
					for _, c := range m.ExpectedCalls {
						c.ReturnArguments = mock.Arguments{&http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
						}, nil}
					}
				})
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "POST" })).Run(func(args mock.Arguments) {
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
			wantErrContains: "Update attributes failed",
		},
	}
	attrs := fileref.Attributes{WhoPaysForReads: common.WhoPays3rdParty}
	var attrsb []byte
	attrsb, _ = json.Marshal(attrs)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t); teardown != nil {
					defer teardown(t)
				}
			}
			ar := &AttributesRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				attributes:     string(attrsb),
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				ctx:            context.Background(),
				attributesMask: 0,
				connectionID:   zboxutil.NewConnectionId(),
			}
			err := ar.ProcessAttributes()
			if tt.wantErr {
				require.Error(err, "expected error != nil")
				require.Contains(err.Error(), tt.wantErrContains, "expected error contains '%s'", tt.wantErrContains)
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
			if tt.wantFunc != nil {
				tt.wantFunc(require, ar)
			}
		})
	}
}

func TestAttributesRequest_updateBlobberObjectAttributes(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, attributeWorkerTestDir, blobberMocks)
	defer cncl()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		wantErr        bool
		wantFunc       func(require *require.Assertions, ar *AttributesRequest)
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
			true,
			nil,
		},
		{
			"Test_Error_Get_Object_Tree_From_Blobber_Failed",
			nil,
			true,
			nil,
		},
		{
			"Test_Update_Blobber_Object_Attributes_Failed",
			func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{"who_pays_for_reads":1},"commit_meta_txns":null,"content_hash":"adc83b19e793491b1c6ea0fd8b46cd9f32e592fc","created_at":"2021-03-17T08:15:36.018029Z","custom_meta":"","encrypted_key":"","hash":"4fea25c7390d0d8374fd84c77345eee7037224c2b162e98950a9ea6d882e91e5","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"6ed726c5aaf50067479a105ad9c4330bfa341f1fd889e3552af67303712ee0f0","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-19T08:34:35.387704Z"},"latest_write_marker":{"allocation_root":"d069d04ba45c4f14764c699b69f07897d0973afaab5aa2cf9c302c849ccad955","prev_allocation_root":"fe2f7f060dd34adbfdcf2f142acd33f50401207f7363359dd35460be7a2bec2d","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":0,"blobber_id":"8eeb4d4d6621ea87ecf4d3ba61bb69d30db435c8ed7cbaccccd56b93ea050c42","timestamp":1616142875,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"ebea73ea057d334c7600ce65f323843d02ea4bd822c858ab004aa0f30043e11f"}}`
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "GET" })).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
				}, nil)
				m.On("Do", mock.MatchedBy(func(req *http.Request) bool { return req.Method == "POST" })).Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(strings.NewReader("")),
				}, nil)
				return nil
			},
			false,
			func(require *require.Assertions, ar *AttributesRequest) {
				require.NotNil(ar)
				require.Equal(uint32(0), ar.attributesMask)
				require.Equal(float32(0), ar.consensus)
			},
		},
		{
			"Test_Update_Blobber_Object_Attributes_Success",
			func(t *testing.T) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				bodyString := `{"meta_data":{"actual_file_hash":"03cfd743661f07975fa2f1220c5194cbaff48451","actual_file_size":4,"actual_thumbnail_hash":"","actual_thumbnail_size":0,"attributes":{"who_pays_for_reads":1},"commit_meta_txns":null,"content_hash":"adc83b19e793491b1c6ea0fd8b46cd9f32e592fc","created_at":"2021-03-17T08:15:36.018029Z","custom_meta":"","encrypted_key":"","hash":"4fea25c7390d0d8374fd84c77345eee7037224c2b162e98950a9ea6d882e91e5","lookup_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","merkle_root":"6ed726c5aaf50067479a105ad9c4330bfa341f1fd889e3552af67303712ee0f0","mimetype":"application/octet-stream","name":"1.txt","num_of_blocks":1,"on_cloud":false,"path":"/1.txt","path_hash":"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697","size":1,"thumbnail_hash":"","thumbnail_size":0,"type":"f","updated_at":"2021-03-19T08:34:35.387704Z"},"latest_write_marker":{"allocation_root":"d069d04ba45c4f14764c699b69f07897d0973afaab5aa2cf9c302c849ccad955","prev_allocation_root":"fe2f7f060dd34adbfdcf2f142acd33f50401207f7363359dd35460be7a2bec2d","allocation_id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","size":0,"blobber_id":"8eeb4d4d6621ea87ecf4d3ba61bb69d30db435c8ed7cbaccccd56b93ea050c42","timestamp":1616142875,"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","signature":"ebea73ea057d334c7600ce65f323843d02ea4bd822c858ab004aa0f30043e11f"}}`
				m.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
				}, nil)
				return nil
			},
			false,
			func(require *require.Assertions, ar *AttributesRequest) {
				require.NotNil(ar)
				require.Equal(uint32(1), ar.attributesMask)
				require.Equal(float32(1), ar.consensus)
			},
		},
	}
	attrs := fileref.Attributes{WhoPaysForReads: common.WhoPays3rdParty}
	var attrsb []byte
	attrsb, _ = json.Marshal(attrs)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t); teardown != nil {
					defer teardown(t)
				}
			}
			ar := &AttributesRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				attributes:     string(attrsb),
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				ctx:            context.Background(),
				attributesMask: 0,
				connectionID:   zboxutil.NewConnectionId(),
			}
			_, err := ar.updateBlobberObjectAttributes(a.Blobbers[0], 0)
			if tt.wantErr {
				require.Error(err, "expected error != nil")
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
			if tt.wantFunc != nil {
				tt.wantFunc(require, ar)
			}
		})
	}
}

func TestAttributesRequest_getObjectTreeFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, attributeWorkerTestDir, blobberMocks)
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
				m.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(bodyString)),
				}, nil)
				return nil
			},
			false,
		},
	}
	attrs := fileref.Attributes{WhoPaysForReads: common.WhoPays3rdParty}
	var attrsb []byte
	attrsb, _ = json.Marshal(attrs)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t); teardown != nil {
					defer teardown(t)
				}
			}
			ar := &AttributesRequest{
				allocationID:   a.ID,
				allocationTx:   a.Tx,
				blobbers:       a.Blobbers,
				remotefilepath: "/1.txt",
				attributes:     string(attrsb),
				Consensus: Consensus{
					consensusThresh: (float32(a.DataShards) * 100) / float32(a.DataShards+a.ParityShards),
					fullconsensus:   float32(a.DataShards + a.ParityShards),
				},
				ctx:            context.Background(),
				attributesMask: 0,
				connectionID:   zboxutil.NewConnectionId(),
			}
			_, err := ar.getObjectTreeFromBlobber(a.Blobbers[0])
			if tt.wantErr {
				require.Error(err, "expected error != nil")
				return
			}
			require.NoErrorf(err, "expected no error but got %v", err)
		})
	}
}
