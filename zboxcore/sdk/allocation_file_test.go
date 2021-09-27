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
	"sync"
	"testing"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/dev"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAllocation_UpdateFile(t *testing.T) {
	const mockLocalPath = "1.txt"
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_UpdateFile",
		ParityShards: 2,
		DataShards:   2,
	}
	setupMockAllocation(t, a)

	server := dev.NewBlobberServer()
	defer server.Close()

	for i := 0; i < numBlobbers; i++ {

		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	a.uploadChan = make(chan *UploadRequest, 4)
	err := a.UpdateFile(mockLocalPath, "/", fileref.Attributes{}, nil)
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
	}
	server := dev.NewBlobberServer()
	defer server.Close()
	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	err := a.UploadFile(mockLocalPath, "/", fileref.Attributes{}, nil)
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

	server := dev.NewBlobberServer()
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
				defer teardown(t)
			}
			a := &Allocation{
				Tx:           "TestAllocation_UpdateFileWithThumbnail",
				ParityShards: 2,
				DataShards:   2,
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
			err := a.UpdateFileWithThumbnail(tt.parameters.localPath, tt.parameters.remotePath, tt.parameters.thumbnailPath, fileref.Attributes{}, tt.parameters.status)
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
		mockLocalPath     = "1.txt"
		mockThumbnailPath = "thumbnail_alloc"
	)
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_UploadFileWithThumbnail",
		ParityShards: 2,
		DataShards:   2,
	}
	server := dev.NewBlobberServer()
	defer server.Close()
	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	err := a.UploadFileWithThumbnail(mockLocalPath, "/", mockThumbnailPath, fileref.Attributes{}, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUpdateFile(t *testing.T) {
	const mockLocalPath = "1.txt"
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_EncryptAndUpdateFile",
		ParityShards: 2,
		DataShards:   2,
	}

	server := dev.NewBlobberServer()
	defer server.Close()
	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	err := a.EncryptAndUpdateFile(mockLocalPath, "/", fileref.Attributes{}, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUploadFile(t *testing.T) {
	const mockLocalPath = "1.txt"
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_EncryptAndUploadFile",
		ParityShards: 2,
		DataShards:   2,
	}
	server := dev.NewBlobberServer()
	defer server.Close()
	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	err := a.EncryptAndUploadFile(mockLocalPath, "/", fileref.Attributes{}, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUpdateFileWithThumbnail(t *testing.T) {
	const (
		mockLocalPath     = "1.txt"
		mockThumbnailPath = "thumbnail_alloc"
	)
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_EncryptAndUpdateFileWithThumbnail",
		ParityShards: 2,
		DataShards:   2,
	}
	server := dev.NewBlobberServer()
	defer server.Close()
	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	err := a.EncryptAndUpdateFileWithThumbnail(mockLocalPath, "/", mockThumbnailPath, fileref.Attributes{}, nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUploadFileWithThumbnail(t *testing.T) {
	const (
		mockLocalPath     = "1.txt"
		mockThumbnailPath = "thumbnail_alloc"
	)
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_EncryptAndUploadFileWithThumbnail",
		ParityShards: 2,
		DataShards:   2,
	}
	server := dev.NewBlobberServer()
	defer server.Close()
	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}
	err := a.EncryptAndUploadFileWithThumbnail(mockLocalPath, "/", mockThumbnailPath, fileref.Attributes{}, nil)
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
		attrs         fileref.Attributes
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
				attrs:         fileref.Attributes{},
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
				attrs:         fileref.Attributes{},
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
				attrs:         fileref.Attributes{},
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
				attrs:         fileref.Attributes{},
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
				attrs:         fileref.Attributes{},
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
				attrs:         fileref.Attributes{},
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
			err := a.uploadOrUpdateFile(tt.parameters.localPath, tt.parameters.remotePath, tt.parameters.status, tt.parameters.isUpdate, tt.parameters.thumbnailPath, tt.parameters.encryption, tt.parameters.isRepair, tt.parameters.attrs)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {

				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}
