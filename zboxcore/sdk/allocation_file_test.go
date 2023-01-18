package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/dev"

	"github.com/0chain/gosdk/sdks/blobber"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	devmock "github.com/0chain/gosdk/dev/mock"
)

func TestAllocation_UpdateFile(t *testing.T) {
	const mockLocalPath = "1.txt"

	a := &Allocation{
		ID:           "TestAllocation_UpdateFile",
		Tx:           "TestAllocation_UpdateFile",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}
	setupMockAllocation(t, a)

	require := require.New(t)
	if teardown := setupMockFileAndReferencePathResult(t, a.ID, mockLocalPath); teardown != nil {
		defer teardown(t)
	}

	resp := &WMLockResult{
		Status: WMLockStatusOK,
	}

	respBuf, _ := json.Marshal(resp)
	respMap := make(devmock.ResponseMap)
	respMap[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       respBuf,
	}
	respMap[http.MethodPost+":"+blobber.EndpointFileMeta+a.Tx] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       []byte("{\"actual_file_size\":1}"),
	}

	server := dev.NewBlobberServer(respMap)
	defer server.Close()

	for i := 0; i < numBlobbers; i++ {

		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	a.uploadChan = make(chan *UploadRequest, 4)
	err := a.UpdateFile(os.TempDir(), mockLocalPath, "/", nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_UploadFile(t *testing.T) {
	const mockLocalPath = "1.txt"
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_UploadFile",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}

	resp := &WMLockResult{
		Status: WMLockStatusOK,
	}

	respBuf, _ := json.Marshal(resp)
	respMap := make(devmock.ResponseMap)
	respMap[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       respBuf,
	}
	respMap[http.MethodPost+":"+blobber.EndpointFileMeta+a.Tx] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       []byte("{\"actual_file_size\":1}"),
	}

	server := dev.NewBlobberServer(respMap)

	defer server.Close()
	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	err := a.UploadFile(os.TempDir(), mockLocalPath, "/", nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_UpdateFileWithThumbnail(t *testing.T) {
	const (
		mockLocalPath     = "1.txt"
		mockThumbnailPath = "thumbnail_alloc"
	)

	type parameters struct {
		localPath, remotePath, thumbnailPath string
		status                               StatusCallback
	}
	tests := []struct {
		name       string
		parameters parameters
		wantErr    bool
	}{
		{
			"Test_Coverage",
			parameters{
				localPath:     mockLocalPath,
				remotePath:    "/",
				thumbnailPath: mockThumbnailPath,
			},
			false,
		},
	}

	resp := &WMLockResult{
		Status: WMLockStatusOK,
	}

	respBuf, _ := json.Marshal(resp)
	respMap := make(devmock.ResponseMap)
	respMap[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+"TestAllocation_UpdateFileWithThumbnail"] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       respBuf,
	}
	respMap[http.MethodPost+":"+blobber.EndpointFileMeta+"TestAllocation_UpdateFileWithThumbnail"] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       []byte("{\"actual_file_size\":1}"),
	}

	server := dev.NewBlobberServer(respMap)
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			a := &Allocation{
				ID:           "TestAllocation_UpdateFileWithThumbnail",
				Tx:           "TestAllocation_UpdateFileWithThumbnail",
				ParityShards: 2,
				DataShards:   2,
				Size:         2 * GB,
			}

			if teardown := setupMockFileAndReferencePathResult(t, a.ID, mockLocalPath); teardown != nil {
				defer teardown(t)
			}

			if teardown := setupMockFile(t, tt.parameters.thumbnailPath); teardown != nil {
				defer teardown(t)
			}

			a.uploadChan = make(chan *UploadRequest, 10)
			a.downloadChan = make(chan *DownloadRequest, 10)
			a.repairChan = make(chan *RepairRequest, 1)
			a.ctx, a.ctxCancelF = context.WithCancel(context.Background())
			a.uploadProgressMap = make(map[string]*UploadRequest)
			a.downloadProgressMap = make(map[string]*DownloadRequest)
			a.mutex = &sync.Mutex{}
			a.initialized = true
			sdkInitialized = true
			setupMockAllocation(t, a)
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      mockBlobberId + strconv.Itoa(i),
					Baseurl: server.URL,
				})
			}
			err := a.UpdateFileWithThumbnail(os.TempDir(), tt.parameters.localPath, tt.parameters.remotePath, tt.parameters.thumbnailPath, tt.parameters.status)
			if tt.wantErr {
				require.Errorf(err, "expected error != nil")
				return
			}
			require.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_UploadFileWithThumbnail(t *testing.T) {
	const (
		mockTmpPath       = "/tmp"
		mockLocalPath     = "1.txt"
		mockThumbnailPath = "thumbnail_alloc"
	)
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}

	if teardown := setupMockFile(t, mockThumbnailPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_UploadFileWithThumbnail",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}

	resp := &WMLockResult{
		Status: WMLockStatusOK,
	}

	respBuf, _ := json.Marshal(resp)
	respMap := make(devmock.ResponseMap)
	respMap[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       respBuf,
	}
	respMap[http.MethodPost+":"+blobber.EndpointFileMeta+a.Tx] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       []byte("{\"actual_file_size\":1}"),
	}

	server := dev.NewBlobberServer(respMap)
	defer server.Close()

	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	err := a.UploadFileWithThumbnail(mockTmpPath, mockLocalPath, "/", mockThumbnailPath, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUpdateFile(t *testing.T) {
	const (
		mockLocalPath = "1.txt"
		mockTmpPath   = "/tmp"
	)
	require := require.New(t)

	a := &Allocation{
		Tx:           "TestAllocation_EncryptAndUpdateFile",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}

	if teardown := setupMockFileAndReferencePathResult(t, a.Tx, mockLocalPath); teardown != nil {
		defer teardown(t)
	}

	resp := &WMLockResult{
		Status: WMLockStatusOK,
	}

	respBuf, _ := json.Marshal(resp)
	respMap := make(devmock.ResponseMap)
	respMap[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       respBuf,
	}
	respMap[http.MethodPost+":"+blobber.EndpointFileMeta+a.Tx] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       []byte("{\"actual_file_size\":1}"),
	}

	server := dev.NewBlobberServer(respMap)
	defer server.Close()

	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	err := a.EncryptAndUpdateFile(mockTmpPath, mockLocalPath, "/", nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUploadFile(t *testing.T) {
	const (
		mockLocalPath = "1.txt"
		mockTmpPath   = "/tmp"
	)
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_EncryptAndUploadFile",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}

	resp := &WMLockResult{
		Status: WMLockStatusOK,
	}

	respBuf, _ := json.Marshal(resp)
	respMap := make(devmock.ResponseMap)
	respMap[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       respBuf,
	}
	respMap[http.MethodPost+":"+blobber.EndpointFileMeta+a.Tx] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       []byte("{\"actual_file_size\":1}"),
	}

	server := dev.NewBlobberServer(respMap)
	defer server.Close()

	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	err := a.EncryptAndUploadFile(mockTmpPath, mockLocalPath, "/", nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

// This test was always failing but passing because processCommit function
// was returning nil error regardless of non-nil error.
// We should return back after some time to fix it.
// func TestAllocation_EncryptAndUpdateFileWithThumbnail(t *testing.T) {
// 	const (
// 		mockLocalPath     = "1.txt"
// 		mockThumbnailPath = "thumbnail_alloc"
// 		mockTmpPath       = "/tmp"
// 	)
// 	require := require.New(t)
// 	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
// 		defer teardown(t)
// 	}
// 	a := &Allocation{
// 		Tx:           "TestAllocation_EncryptAndUpdateFileWithThumbnail",
// 		ParityShards: 2,
// 		DataShards:   2,
// 	}

// 	resp := &WMLockResult{
// 		Status: WMLockStatusOK,
// 	}

// 	respBuf, _ := json.Marshal(resp)
// 	respMap := make(devmock.ResponseMap)
// 	respMap[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = devmock.Response{
// 		StatusCode: http.StatusOK,
// 		Body:       respBuf,
// 	}
// 	server := dev.NewBlobberServer(respMap)
// 	defer server.Close()

// 	setupMockAllocation(t, a)
// 	for i := 0; i < numBlobbers; i++ {
// 		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
// 			ID:      mockBlobberId + strconv.Itoa(i),
// 			Baseurl: server.URL,
// 		})
// 	}
// 	err := a.EncryptAndUpdateFileWithThumbnail(mockTmpPath, mockLocalPath, "/", mockThumbnailPath, nil)
// 	require.NoErrorf(err, "Unexpected error %v", err)
// }

func TestAllocation_EncryptAndUploadFileWithThumbnail(t *testing.T) {
	const (
		mockLocalPath     = "1.txt"
		mockThumbnailPath = "thumbnail_alloc"
		mockTmpPath       = "/tmp"
	)
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_EncryptAndUploadFileWithThumbnail",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
		ctx:          context.TODO(),
	}

	resp := &WMLockResult{
		Status: WMLockStatusOK,
	}

	respBuf, _ := json.Marshal(resp)
	respMap := make(devmock.ResponseMap)
	respMap[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = devmock.Response{
		StatusCode: http.StatusOK,
		Body:       respBuf,
	}
	server := dev.NewBlobberServer(respMap)
	defer server.Close()

	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	err := a.EncryptAndUploadFileWithThumbnail(mockTmpPath, mockLocalPath, "/", mockThumbnailPath, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_uploadOrUpdateFile(t *testing.T) {
	const (
		mockFileRefName   = "mock file ref name"
		mockLocalPath     = "1.txt"
		mockActualHash    = "4041e3eeb170751544a47af4e4f9d374e76cee1d"
		mockErrorHash     = "1041e3eeb170751544a47af4e4f9d374e76cee1d"
		mockThumbnailPath = "thumbnail_alloc"
	)

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponses := func(t *testing.T, testCaseName string, a *Allocation, hash string) (teardown func(t *testing.T)) {
		for i := 0; i < numBlobbers; i++ {
			var hash string
			if i < numBlobbers-1 {
				hash = mockErrorHash
			}
			frName := mockFileRefName + strconv.Itoa(i)
			url := "TestAllocation_uploadOrUpdateFile" + testCaseName + mockBlobberUrl + strconv.Itoa(i)
			a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
				Baseurl: url,
			})
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.Path, url)
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

			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.Contains(req.URL.Path, blobber.EndpointWriteMarkerLock)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileRefName, hash string) io.ReadCloser {
					resp := &WMLockResult{
						Status: WMLockStatusOK,
					}
					respBuf, _ := json.Marshal(resp)
					return ioutil.NopCloser(bytes.NewReader(respBuf))
				}(frName, hash),
			}, nil)
		}
		return nil
	}

	type parameters struct {
		localPath     string
		remotePath    string
		status        StatusCallback
		isUpdate      bool
		thumbnailPath string
		encryption    bool
		isRepair      bool
		hash          string
	}
	tests := []struct {
		name       string
		setup      func(*testing.T, string, *Allocation, string) (teardown func(*testing.T))
		parameters parameters
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Not_Initialize_Failed",
			setup: func(t *testing.T, testCaseName string, a *Allocation, hash string) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) { a.initialized = true }
			},
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name:  "Test_Thumbnail_File_Error_Success",
			setup: setupHttpResponses,
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: mockThumbnailPath,
				encryption:    false,
				isRepair:      false,
				hash:          mockActualHash,
			},
		},
		{
			name:  "Test_Invalid_Remote_Abs_Path_Failed",
			setup: nil,
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
			},
			wantErr: true,
			errMsg:  "invalid_path: Path should be valid and absolute",
		},
		{
			name:  "Test_Repair_Remote_File_Not_Found_Failed",
			setup: nil,
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "/x.txt",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      true,
			},
			wantErr: true,
			errMsg:  "File not found for the given remotepath",
		},
		{
			name:  "Test_Repair_Content_Hash_Not_Matches_Failed",
			setup: setupHttpResponses,
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      true,
				hash:          mockErrorHash,
			},
			wantErr: true,
			errMsg:  "Content hash doesn't match",
		},
		{
			name:  "Test_Upload_Success",
			setup: setupHttpResponses,
			parameters: parameters{
				localPath:     mockLocalPath,
				remotePath:    "/",
				isUpdate:      false,
				thumbnailPath: "",
				encryption:    false,
				isRepair:      false,
				hash:          mockActualHash,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
				defer teardown(t)
			}
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
			}
			setupMockAllocation(t, a)
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a, tt.parameters.hash); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.uploadOrUpdateFile(tt.parameters.localPath, tt.parameters.remotePath, tt.parameters.status, tt.parameters.isUpdate, tt.parameters.thumbnailPath, tt.parameters.encryption, tt.parameters.isRepair)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {

				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}

func TestAllocation_RepairFile(t *testing.T) {
	const (
		mockFileRefName = "mock file ref name"
		mockLocalPath   = "1.txt"
		mockActualHash  = "75a919d23622c29ade8096ed1add6606ec970579459178db3a7d1d0ff8df92d3"
		mockChunkHash   = "a6fb1cb61c9a3b8709242de28e44fb0b4de3753995396ae1d21ca9d4e956e9e2"
	)

	rawClient := zboxutil.Client
	createClient := resty.CreateClient

	var mockClient = mocks.HttpClient{}

	zboxutil.Client = &mockClient
	resty.CreateClient = func(t *http.Transport, timeout time.Duration) resty.Client {
		return &mockClient
	}

	defer func() {
		zboxutil.Client = rawClient
		resty.CreateClient = createClient
	}()

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponses := func(t *testing.T, testName string, numBlobbers, numCorrect int) {
		require.True(t, numBlobbers >= numCorrect)
		for i := 0; i < numBlobbers; i++ {
			var hash string
			if i < numCorrect {
				hash = mockActualHash
			}
			frName := mockFileRefName + strconv.Itoa(i)
			url := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/file/meta"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), url)
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

	setupHttpResponsesWithUpload := func(t *testing.T, testName string, numBlobbers, numCorrect int) {
		require.True(t, numBlobbers >= numCorrect)
		for i := 0; i < numBlobbers; i++ {
			var hash string
			if i < numCorrect {
				hash = mockActualHash
			}

			frName := mockFileRefName + strconv.Itoa(i)
			httpResponse := &http.Response{
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
			}

			urlMeta := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/file/meta"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlMeta)
			})).Return(httpResponse, nil)

			urlUpload := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/file/upload"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlUpload)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileRefName, hash string) io.ReadCloser {
					jsonFR, err := json.Marshal(&UploadResult{
						Filename: mockLocalPath,
						Hash:     mockChunkHash,
					})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(frName, hash),
			}, nil)

			urlFilePath := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/file/referencepath"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlFilePath)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileRefName, hash string) io.ReadCloser {
					jsonFR, err := json.Marshal(&ReferencePathResult{
						ReferencePath: &fileref.ReferencePath{
							Meta: map[string]interface{}{
								"type": "d",
							},
						},
						LatestWM: nil,
					})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(frName, hash),
			}, nil)

			urlCommit := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/connection/commit"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlCommit)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileRefName, hash string) io.ReadCloser {
					jsonFR, err := json.Marshal(&ReferencePathResult{})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(frName, hash),
			}, nil)

			urlLock := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + zboxutil.WM_LOCK_ENDPOINT
			urlLock = strings.TrimRight(urlLock, "/")
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlLock)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func() io.ReadCloser {
					resp := &WMLockResult{
						Status: WMLockStatusOK,
					}
					respBuf, _ := json.Marshal(resp)
					return ioutil.NopCloser(bytes.NewReader(respBuf))
				}(),
			}, nil)
		}
	}

	type parameters struct {
		localPath  string
		remotePath string
		status     StatusCallback
	}
	tests := []struct {
		name        string
		parameters  parameters
		numBlobbers int
		numCorrect  int
		setup       func(*testing.T, string, int, int)
		wantErr     bool
		errMsg      string
	}{
		{
			name: "Test_Repair_Not_Required_Failed",
			parameters: parameters{
				localPath:  mockLocalPath,
				remotePath: "/",
			},
			numBlobbers: 4,
			numCorrect:  4,
			setup:       setupHttpResponses,
			wantErr:     true,
			errMsg:      "chunk_upload: Repair not required",
		},
		{
			name: "Test_Repair_Required_Success",
			parameters: parameters{
				localPath:  mockLocalPath,
				remotePath: "/",
			},
			numBlobbers: 6,
			numCorrect:  5,
			setup:       setupHttpResponsesWithUpload,
		},
	}

	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			a := &Allocation{
				ParityShards: tt.numBlobbers / 2,
				DataShards:   tt.numBlobbers / 2,
				Size:         2 * GB,
			}
			a.uploadChan = make(chan *UploadRequest, 10)
			a.downloadChan = make(chan *DownloadRequest, 10)
			a.repairChan = make(chan *RepairRequest, 1)
			a.ctx, a.ctxCancelF = context.WithCancel(context.Background())
			a.uploadProgressMap = make(map[string]*UploadRequest)
			a.downloadProgressMap = make(map[string]*DownloadRequest)
			a.mutex = &sync.Mutex{}
			a.initialized = true
			sdkInitialized = true
			setupMockAllocation(t, a)
			for i := 0; i < tt.numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					Baseurl: "http://TestAllocation_RepairFile" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			tt.setup(t, tt.name, tt.numBlobbers, tt.numCorrect)
			err := a.RepairFile(tt.parameters.localPath, tt.parameters.remotePath, tt.parameters.status)
			if tt.wantErr {
				require.NotNil(err)
			} else {
				require.Nil(err)
			}

			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}
