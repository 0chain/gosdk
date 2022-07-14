package sdk

import (
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

// allocationFileStorer defines functions expected from Allocation for storing file.
// This interface enables easy mocking for UT purposes.
type allocationFileStorer interface {
	ListDir(path string) (*ListResult, error)
	DownloadFile(localPath string, remotePath string, status StatusCallback) error
	StartChunkedUpload(workdir, localPath string, remotePath string, status StatusCallback, isUpdate bool, isRepair bool, thumbnailPath string, encryption bool) error
}

var _ allocationFileStorer = &Allocation{}

type registryFileManager interface {
	Update(data []byte) error
	Get() ([]byte, common.Timestamp, error)
	GetLastUpdateTimestamp() (common.Timestamp, error)
}

// registryFileStore manages the storing of registry file on allocation.
type registryFileStore struct {
	registryFilePath string
	fileStorer       allocationFileStorer
}

var _ registryFileManager = &registryFileStore{}

func newRegistryFileManager(a *Allocation, registryFilePath string) *registryFileStore {
	return &registryFileStore{
		registryFilePath: registryFilePath,
		fileStorer:       a,
	}
}

// Update writes the registry file with data provided.
// This overwrites the current copy stored.
func (a *registryFileStore) Update(data []byte) error {
	lastUpdate, err := a.getRegistryFileLastUpdate()
	if err != nil {
		return errors.New("update_registry_file_failed", "failed to check existence of registry file: "+err.Error())
	}

	// if last update timestamp is the zero value, it means no registry file ye so will upload.
	// otherwise, update registry file.
	isUpdate := lastUpdate != common.Timestamp(0)

	tempLocal := filepath.Join(os.TempDir(), fmt.Sprintf("registerfile_ul_%d", time.Now().UTC().UnixNano()))

	err = os.WriteFile(tempLocal, data, 0600)
	if err != nil {
		return errors.New("update_registry_file_failed", "failed to create file locally for upload: "+err.Error())
	}

	err = a.fileStorer.StartChunkedUpload(os.TempDir(), tempLocal, a.registryFilePath, nil, isUpdate, false, "", false)
	if err != nil {
		return errors.New("update_registry_file_failed", "failed to upload registry file: "+err.Error())
	}

	return nil
}

// Get retrieves the registry file contents and its last updated timestamp.
func (a *registryFileStore) Get() ([]byte, common.Timestamp, error) {
	rfLastUpdateTime, err := a.GetLastUpdateTimestamp()
	if err != nil {
		return nil, common.Timestamp(0), errors.New("get_registry_file_failed", "failed to check existence of registry file: "+err.Error())
	}

	// no registry file yet
	// return empty
	if rfLastUpdateTime == common.Timestamp(0) {
		return []byte{}, rfLastUpdateTime, nil
	}

	// download registry file
	tempLocal := filepath.Join(os.TempDir(), fmt.Sprintf("starred_dl_%d", time.Now().UTC().UnixNano()))
	cb := newRegistryFileCB()

	err = a.fileStorer.DownloadFile(tempLocal, a.registryFilePath, cb)
	if err != nil {
		return nil, common.Timestamp(0), errors.New("get_registry_file_failed", "failed to start download of registry file: "+err.Error())
	}

	<-cb.Wait() // wait for download to complete
	if cb.err != nil {
		return nil, common.Timestamp(0), errors.New("get_registry_file_failed", "failed to download registry file: "+cb.err.Error())
	}

	f, err := os.Open(tempLocal)
	if err != nil {
		return nil, common.Timestamp(0), errors.New("get_registry_file_failed", "failed to open downloaded registry file: "+err.Error())
	}

	defer os.Remove(tempLocal)
	defer f.Close()

	bt, err := io.ReadAll(f)
	if err != nil {
		return nil, common.Timestamp(0), errors.New("get_registry_file_failed", "failed to read downloaded registry file: "+err.Error())
	}

	return bt, rfLastUpdateTime, nil
}

// GetLastUpdateTimestamp retrieves the last update timestamp of the registry file.
// Returns 0 as timestamp when no registry file exist.
func (a *registryFileStore) GetLastUpdateTimestamp() (common.Timestamp, error) {
	lastUpdate, err := a.getRegistryFileLastUpdate()
	if err != nil {
		return common.Timestamp(0), errors.New("get_last_update_timestamp_failed", "failed to get updated timestamp of registry file: "+err.Error())
	}

	return lastUpdate, nil
}

func (a *registryFileStore) getRegistryFileLastUpdate() (common.Timestamp, error) {
	dir, err := a.fileStorer.ListDir(a.registryFilePath)
	if err != nil {
		return common.Timestamp(0), err
	}

	if dir == nil {
		return common.Timestamp(0), nil
	}

	for _, f := range dir.Children {
		if f.Path == a.registryFilePath {
			lastUpdate, err := time.Parse(time.RFC3339, f.UpdatedAt)
			if err != nil {
				return common.Timestamp(0), err
			}

			return common.Timestamp(lastUpdate.Unix()), nil
		}
	}

	return common.Timestamp(0), nil
}

var _ StatusCallback = &registerFileCB{}

type registerFileCB struct {
	once sync.Once
	done chan struct{}
	err  error
}

func newRegistryFileCB() *registerFileCB {
	return &registerFileCB{
		done: make(chan struct{}),
	}
}

func (cb *registerFileCB) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	cb.once.Do(func() {
		close(cb.done)
	})
}

func (cb *registerFileCB) Error(allocationID string, filePath string, op int, err error) {
	cb.once.Do(func() {
		cb.err = err
		close(cb.done)
	})
}

func (cb *registerFileCB) Wait() <-chan struct{} {
	return cb.done
}

func (cb *registerFileCB) CommitMetaCompleted(request, response string, txn *transaction.Transaction, err error) {
	// noop
}

func (cb *registerFileCB) Started(allocationId, filePath string, op int, totalBytes int) {
	// noop
}

func (cb *registerFileCB) InProgress(allocationId, filePath string, op int, completedBytes int, data []byte) {
	// noop
}

func (cb *registerFileCB) RepairCompleted(filesRepaired int) {
	// noop
}
