package sdk

import (
	"fmt"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
	"time"
)

func TestRegistryFileManager_Update(t *testing.T) {
	const registryFile = "/.registry"
	var registryData = []byte("registry contents")

	for _, tc := range []struct {
		name       string
		listDirRes *ListResult
		listDirErr error
		uploadErr  error
		wantErr    error
		wantUpdate bool
	}{
		{
			name:       "upload new registry file successfully",
			listDirRes: &ListResult{Children: []*ListResult{}},
		},
		{
			name:       "update registry file successfully",
			listDirRes: &ListResult{Children: []*ListResult{{Path: registryFile, UpdatedAt: "2022-01-22T02:03:04Z"}}},
			wantUpdate: true,
		},
		{
			name:       "ListDir throws error",
			listDirErr: fmt.Errorf("list error"),
			wantErr:    errors.New("update_registry_file_failed", "Failed to check existence of registry file: list error"),
		},
		{
			name:       "Chunk upload error",
			listDirRes: &ListResult{Children: []*ListResult{}},
			uploadErr:  fmt.Errorf("upload error"),
			wantErr:    errors.New("update_registry_file_failed", "Failed to upload registry file: upload error"),
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			dummyAlloc := &Allocation{}

			r := registryFileStore{
				registryFilePath: registryFile,
				allocation:       dummyAlloc,
				fileStorer: &mockAllocationFileStorer{
					t:                    t,
					listDirRes:           tt.listDirRes,
					listDirErr:           tt.listDirErr,
					uploadErr:            tt.uploadErr,
					wantRegistryFilePath: registryFile,
					wantUploadIsUpdate:   tt.wantUpdate,
					wantUploadContents:   registryData,
				},
			}

			err := r.Update(registryData)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestRegistryFileManager_Get(t *testing.T) {
	const registryFile = "/.registry"
	var registryData = []byte("registry contents")

	for _, tc := range []struct {
		name                    string
		listDirRes              *ListResult
		listDirErr              error
		downloadErr             error
		downloadCallbackErr     error
		downloadContents        []byte
		wantErr                 error
		wantData                []byte
		wantLastUpdateTimestamp common.Timestamp
	}{
		{
			name:                    "with registry file",
			listDirRes:              &ListResult{Children: []*ListResult{{Path: registryFile, UpdatedAt: "2022-01-22T02:03:04Z"}}},
			downloadContents:        registryData,
			wantData:                registryData,
			wantLastUpdateTimestamp: common.Timestamp(1642816984),
		},
		{
			name:       "ListDir throws error",
			listDirErr: fmt.Errorf("list error"),
			wantErr:    errors.New("get_registry_file_failed", "Failed to check existence of registry file: get_last_update_timestamp_failed: Failed to check existence of registry file: list error"),
		},
		{
			name:                    "ListDir returns no registry file",
			listDirRes:              &ListResult{Children: []*ListResult{}},
			wantData:                []byte{},
			wantLastUpdateTimestamp: common.Timestamp(0),
		},
		{
			name:        "Download start error of registry file",
			listDirRes:  &ListResult{Children: []*ListResult{{Path: registryFile, UpdatedAt: "2022-01-22T02:03:04Z"}}},
			downloadErr: fmt.Errorf("download error"),
			wantErr:     errors.New("get_registry_file_failed", "Failed to start download of registry file: download error"),
		},
		{
			name:                "Download callback error of registry file",
			listDirRes:          &ListResult{Children: []*ListResult{{Path: registryFile, UpdatedAt: "2022-01-22T02:03:04Z"}}},
			downloadCallbackErr: fmt.Errorf("callback error"),
			wantErr:             errors.New("get_registry_file_failed", "Failed to download registry file: callback error"),
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			r := registryFileStore{
				registryFilePath: registryFile,
				fileStorer: &mockAllocationFileStorer{
					t:                    t,
					listDirRes:           tt.listDirRes,
					listDirErr:           tt.listDirErr,
					downloadErr:          tt.downloadErr,
					downloadCallbackErr:  tt.downloadCallbackErr,
					downloadContents:     tt.downloadContents,
					wantRegistryFilePath: registryFile,
				},
			}

			gotData, gotTimestamp, err := r.Get()

			assert.Equal(t, tt.wantData, gotData)
			assert.Equal(t, tt.wantLastUpdateTimestamp, gotTimestamp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestRegistryFileManager_GetLastUpdateTimestamp(t *testing.T) {
	const registryFile = "/.registry"

	for _, tc := range []struct {
		name       string
		listDirRes *ListResult
		listDirErr error
		wantErr    error
		want       common.Timestamp
	}{
		{
			name:       "with registry file",
			listDirRes: &ListResult{Children: []*ListResult{{Path: registryFile, UpdatedAt: "2022-01-22T02:03:04Z"}}},
			want:       common.Timestamp(1642816984),
		},
		{
			name:       "ListDir throws error",
			listDirErr: fmt.Errorf("server error"),
			wantErr:    errors.New("get_last_update_timestamp_failed", "Failed to check existence of registry file: server error"),
			want:       common.Timestamp(0),
		},
		{
			name:       "ListDir returns no registry file",
			listDirRes: &ListResult{Children: []*ListResult{}},
			want:       common.Timestamp(0),
		},
		{
			name:       "registry file has invalid updated_at",
			listDirRes: &ListResult{Children: []*ListResult{{Path: registryFile, UpdatedAt: "20220122T020304Z"}}},
			wantErr:    errors.New("get_last_update_timestamp_failed", "Failed to parse last updated timestamp of registry file: parsing time \"20220122T020304Z\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"0122T020304Z\" as \"-\""),
			want:       common.Timestamp(0),
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			r := registryFileStore{
				registryFilePath: registryFile,
				fileStorer: &mockAllocationFileStorer{
					t:                    t,
					listDirRes:           tt.listDirRes,
					listDirErr:           tt.listDirErr,
					wantRegistryFilePath: registryFile,
				},
			}

			got, err := r.GetLastUpdateTimestamp()

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

type mockAllocationFileStorer struct {
	t                    *testing.T
	listDirRes           *ListResult
	listDirErr           error
	downloadErr          error
	downloadCallbackErr  error
	downloadContents     []byte
	uploadErr            error
	wantRegistryFilePath string
	wantUploadIsUpdate   bool
	wantUploadContents   []byte
}

func (a *mockAllocationFileStorer) ListDir(path string) (*ListResult, error) {
	assert.Equal(a.t, a.wantRegistryFilePath, path)
	return a.listDirRes, a.listDirErr
}

func (a *mockAllocationFileStorer) DownloadFile(localPath string, remotePath string, status StatusCallback) error {
	assert.Equal(a.t, a.wantRegistryFilePath, remotePath)
	os.WriteFile(localPath, a.downloadContents, 0600)

	time.AfterFunc(time.Millisecond*200, func() {
		if a.downloadCallbackErr != nil {
			status.Error("dummy", "dummy", 0, a.downloadCallbackErr)
		} else {
			status.Completed("dummy", "dummy", "dummy", "dummy", 0, 0)
		}
	})

	return a.downloadErr
}

func (a *mockAllocationFileStorer) StartChunkedUpload(workdir, localPath, remotePath string, status StatusCallback, isUpdate, isRepair bool, thumbnailPath string, encryption bool, attrs fileref.Attributes) error {
	assert.Equal(a.t, os.TempDir(), workdir)
	assert.Equal(a.t, a.wantRegistryFilePath, remotePath)
	assert.Nil(a.t, status)
	assert.Equal(a.t, a.wantUploadIsUpdate, isUpdate)
	assert.False(a.t, isRepair)
	assert.False(a.t, encryption)
	assert.Equal(a.t, "", thumbnailPath)
	assert.Equal(a.t, fileref.Attributes{}, attrs)

	f, err := os.Open(localPath)
	require.Nil(a.t, err)

	bt, err := io.ReadAll(f)
	require.Nil(a.t, err)

	assert.Equal(a.t, a.wantUploadContents, bt)

	return a.uploadErr
}
