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

// The functions on this file is for the maintenance of starred/favorited files on allocations.
//
// Starred files will be stored on `.starred` file stored on the allocation root path.
// The file contains a list of paths  that is marked as starred on the allocation.

const STARRED_REGISTRY_FILE = `./starred`

type StarredFiles struct {
	Files []StarredFile `json:"files"`
}

type StarredFile struct {
	Path string `json:"path"`
}

// UpdateStarredFiles writes the provided full list of starred files.
func (a *Allocation) UpdateStarredFiles(files *StarredFiles) error {
	if !a.isInitialized() {
		return notInitialized
	}

	if files == nil {
		return errors.New("update_starred_files_failed", "Starred files is nil")
	}

	registry, err := a.getStarredRegistryFile()
	if err != nil {
		return errors.New("update_starred_files_failed", "Failed to check for starred files registry: "+err.Error())
	}

	bt, err := json.Marshal(files)
	if err != nil {
		return errors.New("update_starred_files_failed", "Failed to marshal list of starred files: "+err.Error())
	}

	// if no registry yet, upload a registry.
	// otherwise, update registry.
	isUpdate := registry == nil
	reader := bytes.NewReader(bt)

	fileMeta := FileMeta{
		Path:       STARRED_REGISTRY_FILE,
		ActualSize: int64(len(bt)),
		MimeType:   "application/json",
		RemoteName: filepath.Base(STARRED_REGISTRY_FILE),
		RemotePath: STARRED_REGISTRY_FILE,
	}

	cu, err := CreateChunkedUpload(os.TempDir(), a, fileMeta, reader, isUpdate, false)
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
func (a *Allocation) GetStarredFiles() (*StarredFiles, error) {
	if !a.isInitialized() {
		return nil, notInitialized
	}

	registry, err := a.getStarredRegistryFile()
	if err != nil {
		return nil, errors.New("get_starred_files_failed", "Failed to check for starred files registry: "+err.Error())
	}

	// no registry
	if registry == nil {
		return &StarredFiles{Files: []StarredFile{}}, nil
	}

	// download
	tempLocal := filepath.Join(os.TempDir(), "starred_dl_"+fmt.Sprintf("%d", time.Now().UnixNano()))
	cb := newStarredFileCB()

	err = a.DownloadFile(tempLocal, STARRED_REGISTRY_FILE, cb)
	if err != nil {
		return nil, errors.New("get_starred_files_failed", "Failed to start download of starred files registry: "+err.Error())
	}

	cb.Wait() // wait for download to complete
	if cb.err != nil {
		return nil, errors.New("get_starred_files_failed", "Failed to download starred files registry: "+cb.err.Error())
	}

	starred := &StarredFiles{}

	f, err := os.Open(tempLocal)
	if f != nil {
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

	return starred, nil
}

// GetStarredFilesLastUpdateTimestamp retrieves the latest update timestamp of the starred file registry.
func (a *Allocation) GetStarredFilesLastUpdateTimestamp() (common.Timestamp, error) {
	if !a.isInitialized() {
		return common.Timestamp(0), notInitialized
	}

	registry, err := a.getStarredRegistryFile()
	if err != nil {
		return common.Timestamp(0), errors.New("get_starred_files_last_update_failed", "Failed to check for starred files registry: "+err.Error())
	}

	// when no registry yet, return beginning of epoch as default last update timestamp.
	if registry == nil {
		return common.Timestamp(0), nil
	}

	lastUpdate, err := time.Parse(time.RFC3339, registry.UpdatedAt)
	if err != nil {
		return common.Timestamp(0), errors.New("get_starred_files_last_update_failed", "Failed to parse last updated timestamp of starred files registry: "+err.Error())
	}

	return common.Timestamp(lastUpdate.Unix()), nil
}

func (a *Allocation) getStarredRegistryFile() (*ListResult, error) {
	// list root directory
	rootDir, err := a.ListDir("/")
	if err != nil {
		return nil, err
	}

	if rootDir == nil {
		return nil, nil
	}

	for _, f := range rootDir.Children {
		if f.Path == STARRED_REGISTRY_FILE {
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
