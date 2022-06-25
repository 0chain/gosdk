package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/transaction"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// This file provides helper functions to manage list of starred (or favorited) files on allocations.
//
// Starred files will be stored on `/.starred` file on the allocation root path.
// The file contains a json listing the paths marked as starred on the allocation.
//
// It is up to the SDK clients to ensure that their copy of starred files is always up-to date (not stale) before
// doing a full overwrite of the saved starred list through UpdateStarredFiles(). Peeking
//
// It is recommended to flush any update to the list of starred files as often as possible to avoid list being stale for too long.

const StarredRegistryFilePath = `/.starred`

// StarredFiles defines the contents of starred registry file.
type StarredFiles struct {
	UpdatedAt common.Timestamp `json:"-"`
	Files     []StarredFile    `json:"files"`
}

// StarredFile defines the individual entry on starred registry file.
type StarredFile struct {
	Path string `json:"path"`
}

// UpdateStarredFiles writes the provided full list of starred files through the registry.
func UpdateStarredFiles(a *Allocation, files *StarredFiles) error {
	return newStarredFilesRegistry(a).UpdateStarredFiles(files)
}

// GetStarredFiles returns the full list of starred files through the registry.
func GetStarredFiles(a *Allocation) (*StarredFiles, error) {
	return newStarredFilesRegistry(a).GetStarredFiles()
}

// GetStarredFilesLastUpdateTimestamp retrieves the latest updated timestamp of the starred file registry.
func GetStarredFilesLastUpdateTimestamp(a *Allocation) (common.Timestamp, error) {
	return newStarredFilesRegistry(a).GetStarredFilesLastUpdateTimestamp()
}

// allocationFileStorer defines functions expected from Allocation for storing file.
// This interface enables easy mocking of file download for UT purposes.
type allocationFileStorer interface {
	ListDir(path string) (*ListResult, error)
	DownloadFile(localPath string, remotePath string, status StatusCallback) error
}

var _ allocationFileStorer = &Allocation{}

// chunkUploader defines the functions expected from ChunkedUpload for uploading file.
// This interface enables easy mocking of file upload for UT purposes.
type chunkUploader interface {
	Start() error
}

var _ chunkUploader = &ChunkedUpload{}

// starredFilesRegistry manages the storing of list of starred files.
// It stores the list at `/.starred` file saved on the allocation root.
type starredFilesRegistry struct {
	allocation   *Allocation
	fileStorer   allocationFileStorer
	fileUploader func(a *Allocation, meta FileMeta, reader io.Reader, isUpdate bool) (chunkUploader, error)
}

func newStarredFilesRegistry(a *Allocation) *starredFilesRegistry {
	return &starredFilesRegistry{
		allocation: a,
		fileStorer: a,
		fileUploader: func(a *Allocation, meta FileMeta, reader io.Reader, isUpdate bool) (chunkUploader, error) {
			return CreateChunkedUpload(os.TempDir(), a, meta, reader, isUpdate, false)
		},
	}
}

// UpdateStarredFiles writes the provided full list of starred files.
func (a *starredFilesRegistry) UpdateStarredFiles(files *StarredFiles) error {
	if files == nil {
		return errors.New("update_starred_files_failed", "Starred files is nil")
	}

	rf, err := a.getStarredRegistryFile()
	if err != nil {
		return errors.New("update_starred_files_failed", "Failed to check for starred files registry: "+err.Error())
	}

	// if no registry yet, upload new registry file.
	// otherwise, update registry file.
	isUpdate := rf != nil

	bt, err := json.Marshal(files)
	if err != nil {
		return errors.New("update_starred_files_failed", "Failed to marshal list of starred files: "+err.Error())
	}
	reader := bytes.NewReader(bt)

	fileMeta := FileMeta{
		Path:       StarredRegistryFilePath,
		ActualSize: int64(len(bt)),
		MimeType:   "application/json",
		RemoteName: filepath.Base(StarredRegistryFilePath),
		RemotePath: StarredRegistryFilePath,
	}

	cu, err := a.fileUploader(a.allocation, fileMeta, reader, isUpdate)
	if err != nil {
		return errors.New("update_starred_files_failed", "Failed to create chunked upload for starred files registry: "+err.Error())
	}

	err = cu.Start()
	if err != nil {
		return errors.New("update_starred_files_failed", "Failed to upload starred files registry: "+err.Error())
	}

	return nil
}

// GetStarredFiles returns the full list of starred files.
// If no registry file yet, an empty list will be returned.
func (a *starredFilesRegistry) GetStarredFiles() (*StarredFiles, error) {
	rfLastUpdateTime, err := a.GetStarredFilesLastUpdateTimestamp()
	if err != nil {
		return nil, errors.New("get_starred_files_failed", "Failed to check for starred files registry: "+err.Error())
	}

	// no registry file yet
	if rfLastUpdateTime == common.Timestamp(0) {
		return &StarredFiles{UpdatedAt: common.Timestamp(0), Files: []StarredFile{}}, nil
	}

	// download registry file
	tempLocal := filepath.Join(os.TempDir(), "starred_dl_"+fmt.Sprintf("%d", time.Now().UnixNano()))
	cb := newStarredFileCB()

	err = a.fileStorer.DownloadFile(tempLocal, StarredRegistryFilePath, cb)
	if err != nil {
		return nil, errors.New("get_starred_files_failed", "Failed to start download of starred files registry: "+err.Error())
	}

	<-cb.Wait() // wait for download to complete
	if cb.err != nil {
		return nil, errors.New("get_starred_files_failed", "Failed to download starred files registry: "+cb.err.Error())
	}

	starred := &StarredFiles{}

	f, err := os.Open(tempLocal)
	if err != nil {
		return nil, errors.New("get_starred_files_failed", "Failed to open downloaded starred files registry: "+err.Error())
	}

	defer os.Remove(tempLocal)
	defer f.Close()

	bt, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.New("get_starred_files_failed", "Failed to read downloaded starred files registry: "+err.Error())
	}

	err = json.Unmarshal(bt, starred)
	if err != nil {
		return nil, errors.New("get_starred_files_failed", "Failed to parse downloaded starred files registry: "+err.Error())
	}

	starred.UpdatedAt = rfLastUpdateTime

	return starred, nil
}

// GetStarredFilesLastUpdateTimestamp retrieves the latest update timestamp of the starred file registry.
func (a *starredFilesRegistry) GetStarredFilesLastUpdateTimestamp() (common.Timestamp, error) {
	rf, err := a.getStarredRegistryFile()
	if err != nil {
		return common.Timestamp(0), errors.New("get_starred_files_last_update_failed", "Failed to check for starred files registry: "+err.Error())
	}

	// when no registry yet, return beginning of epoch as default last update timestamp.
	if rf == nil {
		return common.Timestamp(0), nil
	}

	lastUpdate, err := time.Parse(time.RFC3339, rf.UpdatedAt)
	if err != nil {
		return common.Timestamp(0), errors.New("get_starred_files_last_update_failed", "Failed to parse last updated timestamp of starred files registry: "+err.Error())
	}

	return common.Timestamp(lastUpdate.Unix()), nil
}

func (a *starredFilesRegistry) getStarredRegistryFile() (*ListResult, error) {
	dir, err := a.fileStorer.ListDir(StarredRegistryFilePath)
	if err != nil {
		return nil, err
	}

	if dir == nil {
		return nil, nil
	}

	for _, f := range dir.Children {
		if f.Path == StarredRegistryFilePath {
			return f, nil
		}
	}

	return nil, nil
}

var _ StatusCallback = &starredFileCB{}

type starredFileCB struct {
	once sync.Once
	done chan struct{}
	err  error
}

func newStarredFileCB() *starredFileCB {
	return &starredFileCB{
		done: make(chan struct{}),
	}
}

func (cb *starredFileCB) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	cb.once.Do(func() {
		close(cb.done)
	})
}

func (cb *starredFileCB) Error(allocationID string, filePath string, op int, err error) {
	cb.once.Do(func() {
		cb.err = err
		close(cb.done)
	})
}

func (cb *starredFileCB) Wait() <-chan struct{} {
	return cb.done
}

func (cb *starredFileCB) CommitMetaCompleted(request, response string, txn *transaction.Transaction, err error) {
	// noop
}

func (cb *starredFileCB) Started(allocationId, filePath string, op int, totalBytes int) {
	// noop
}

func (cb *starredFileCB) InProgress(allocationId, filePath string, op int, completedBytes int, data []byte) {
	// noop
}

func (cb *starredFileCB) RepairCompleted(filesRepaired int) {
	// noop
}
