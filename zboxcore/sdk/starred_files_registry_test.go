package sdk

import (
	"fmt"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
	"time"
)

func TestStarredFilesRegistry_UpdateStarredFiles(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		input                 *StarredFiles
		listDirRes            *ListResult
		listDirErr            error
		createChunkUploadErr  error
		uploadErr             error
		wantErr               error
		wantUpdate            bool
		wantContentsForUpload string
	}{
		{
			name:                  "upload new /.starred successfully",
			input:                 &StarredFiles{UpdatedAt: common.Timestamp(1642816984), Files: []StarredFile{{Path: "/abc.txt"}, {Path: "/def.txt"}}},
			listDirRes:            &ListResult{Children: []*ListResult{}},
			wantContentsForUpload: `{"files":[{"path":"/abc.txt"},{"path":"/def.txt"}]}`,
		},
		{
			name:                  "update /.starred successfully",
			input:                 &StarredFiles{UpdatedAt: common.Timestamp(1642816984), Files: []StarredFile{{Path: "/abc.txt"}, {Path: "/def.txt"}}},
			listDirRes:            &ListResult{Children: []*ListResult{{Path: "/.starred", UpdatedAt: "2022-01-22T02:03:04Z"}}},
			wantUpdate:            true,
			wantContentsForUpload: `{"files":[{"path":"/abc.txt"},{"path":"/def.txt"}]}`,
		},
		{
			name:                  "ListDir throws error",
			input:                 &StarredFiles{UpdatedAt: common.Timestamp(1642816984), Files: []StarredFile{{Path: "/abc.txt"}, {Path: "/def.txt"}}},
			listDirErr:            fmt.Errorf("list error"),
			wantErr:               errors.New("update_starred_files_failed", "Failed to check for starred files registry: list error"),
			wantContentsForUpload: `{"files":[{"path":"/abc.txt"},{"path":"/def.txt"}]}`,
		},
		{
			name:                  "Create chunk upload error",
			input:                 &StarredFiles{UpdatedAt: common.Timestamp(1642816984), Files: []StarredFile{{Path: "/abc.txt"}, {Path: "/def.txt"}}},
			listDirRes:            &ListResult{Children: []*ListResult{}},
			createChunkUploadErr:  fmt.Errorf("server error"),
			wantErr:               errors.New("update_starred_files_failed", "Failed to create chunked upload for starred files registry: server error"),
			wantContentsForUpload: `{"files":[{"path":"/abc.txt"},{"path":"/def.txt"}]}`,
		},
		{
			name:                  "Chunk upload error",
			input:                 &StarredFiles{UpdatedAt: common.Timestamp(1642816984), Files: []StarredFile{{Path: "/abc.txt"}, {Path: "/def.txt"}}},
			listDirRes:            &ListResult{Children: []*ListResult{}},
			uploadErr:             fmt.Errorf("upload error"),
			wantErr:               errors.New("update_starred_files_failed", "Failed to upload starred files registry: upload error"),
			wantContentsForUpload: `{"files":[{"path":"/abc.txt"},{"path":"/def.txt"}]}`,
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			dummyAlloc := &Allocation{}

			r := starredFilesRegistry{
				allocation: dummyAlloc,
				fileStorer: &mockAllocationFileStorer{
					t:          t,
					listDirRes: tt.listDirRes,
					listDirErr: tt.listDirErr,
				},
				fileUploader: func(a *Allocation, meta FileMeta, reader io.Reader, isUpdate bool) (chunkUploader, error) {
					require.Equal(t, dummyAlloc, a)
					assert.Equal(t, tt.wantUpdate, isUpdate)
					assert.Equal(t, FileMeta{
						Path:       "/.starred",
						ActualSize: int64(51),
						MimeType:   "application/json",
						RemoteName: ".starred",
						RemotePath: "/.starred",
					}, meta)

					bt, err := io.ReadAll(reader)
					require.Nil(t, err)

					assert.Equal(t, tt.wantContentsForUpload, string(bt))

					cu := &mockChunkUploader{uploadErr: tt.uploadErr}

					return cu, tt.createChunkUploadErr
				},
			}

			err := r.UpdateStarredFiles(tt.input)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestStarredFilesRegistry_GetStarredFiles(t *testing.T) {
	for _, tc := range []struct {
		name                string
		listDirRes          *ListResult
		listDirErr          error
		downloadErr         error
		downloadCallbackErr error
		downloadContents    string
		wantErr             error
		want                *StarredFiles
	}{
		{
			name:             "with /.starred",
			listDirRes:       &ListResult{Children: []*ListResult{{Path: "/.starred", UpdatedAt: "2022-01-22T02:03:04Z"}}},
			downloadContents: `{"files":[{"path":"/abc.txt"},{"path":"/def.txt"}]}`,
			want:             &StarredFiles{UpdatedAt: common.Timestamp(1642816984), Files: []StarredFile{{Path: "/abc.txt"}, {Path: "/def.txt"}}},
		},
		{
			name:       "ListDir throws error",
			listDirErr: fmt.Errorf("list error"),
			wantErr:    errors.New("get_starred_files_failed", "Failed to check for starred files registry: get_starred_files_last_update_failed: Failed to check for starred files registry: list error"),
		},
		{
			name:       "ListDir returns no /.starred",
			listDirRes: &ListResult{Children: []*ListResult{}},
			want:       &StarredFiles{UpdatedAt: common.Timestamp(0), Files: []StarredFile{}},
		},
		{
			name:        "Download start error of /.starred",
			listDirRes:  &ListResult{Children: []*ListResult{{Path: "/.starred", UpdatedAt: "2022-01-22T02:03:04Z"}}},
			downloadErr: fmt.Errorf("download error"),
			wantErr:     errors.New("get_starred_files_failed", "Failed to start download of starred files registry: download error"),
		},
		{
			name:                "Download callback error of /.starred",
			listDirRes:          &ListResult{Children: []*ListResult{{Path: "/.starred", UpdatedAt: "2022-01-22T02:03:04Z"}}},
			downloadCallbackErr: fmt.Errorf("callback error"),
			wantErr:             errors.New("get_starred_files_failed", "Failed to download starred files registry: callback error"),
		},
		{
			name:             "Bad /.starred content",
			listDirRes:       &ListResult{Children: []*ListResult{{Path: "/.starred", UpdatedAt: "2022-01-22T02:03:04Z"}}},
			downloadContents: `not a json`,
			wantErr:          errors.New("get_starred_files_failed", "Failed to parse downloaded starred files registry: invalid character 'o' in literal null (expecting 'u')"),
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			r := starredFilesRegistry{
				fileStorer: &mockAllocationFileStorer{
					t:                   t,
					listDirRes:          tt.listDirRes,
					listDirErr:          tt.listDirErr,
					downloadErr:         tt.downloadErr,
					downloadCallbackErr: tt.downloadCallbackErr,
					downloadContents:    tt.downloadContents,
				},
			}

			got, err := r.GetStarredFiles()

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestStarredFilesRegistry_GetStarredFilesLastUpdateTimestamp(t *testing.T) {
	for _, tc := range []struct {
		name       string
		listDirRes *ListResult
		listDirErr error
		wantErr    error
		want       common.Timestamp
	}{
		{
			name:       "with /.starred",
			listDirRes: &ListResult{Children: []*ListResult{{Path: "/.starred", UpdatedAt: "2022-01-22T02:03:04Z"}}},
			want:       common.Timestamp(1642816984),
		},
		{
			name:       "ListDir throws error",
			listDirErr: fmt.Errorf("server error"),
			wantErr:    errors.New("get_starred_files_last_update_failed", "Failed to check for starred files registry: server error"),
			want:       common.Timestamp(0),
		},
		{
			name:       "ListDir returns no /.starred",
			listDirRes: &ListResult{Children: []*ListResult{}},
			want:       common.Timestamp(0),
		},
		{
			name:       "/.starred has invalid updated_at",
			listDirRes: &ListResult{Children: []*ListResult{{Path: "/.starred", UpdatedAt: "20220122T020304Z"}}},
			wantErr:    errors.New("get_starred_files_last_update_failed", "Failed to parse last updated timestamp of starred files registry: parsing time \"20220122T020304Z\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"0122T020304Z\" as \"-\""),
			want:       common.Timestamp(0),
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			r := starredFilesRegistry{
				fileStorer: &mockAllocationFileStorer{
					t:          t,
					listDirRes: tt.listDirRes,
					listDirErr: tt.listDirErr,
				},
			}

			got, err := r.GetStarredFilesLastUpdateTimestamp()

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

type mockAllocationFileStorer struct {
	t                   *testing.T
	listDirRes          *ListResult
	listDirErr          error
	downloadErr         error
	downloadCallbackErr error
	downloadContents    string
}

func (a *mockAllocationFileStorer) ListDir(path string) (*ListResult, error) {
	require.Equal(a.t, StarredRegistryFilePath, path)
	return a.listDirRes, a.listDirErr
}

func (a *mockAllocationFileStorer) DownloadFile(localPath string, remotePath string, status StatusCallback) error {
	require.Equal(a.t, StarredRegistryFilePath, remotePath)
	os.WriteFile(localPath, []byte(a.downloadContents), 0600)

	time.AfterFunc(time.Millisecond*200, func() {
		if a.downloadCallbackErr != nil {
			status.Error("dummy", "dummy", 0, a.downloadCallbackErr)
		} else {
			status.Completed("dummy", "dummy", "dummy", "dummy", 0, 0)
		}
	})

	return a.downloadErr
}

type mockChunkUploader struct {
	uploadErr error
}

func (u *mockChunkUploader) Start() error {
	return u.uploadErr
}
