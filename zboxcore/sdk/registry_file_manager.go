package sdk

import (
	"fmt"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/fileref"
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
	StartChunkedUpload(workdir, localPath string, remotePath string, status StatusCallback, isUpdate bool, isRepair bool, thumbnailPath string, encryption bool, attrs fileref.Attributes) error
}

var _ allocationFileStorer = &Allocation{}

type RegistryFileManager interface {
	Update(data []byte) error
	Get() ([]byte, common.Timestamp, error)
	GetLastUpdateTimestamp() (common.Timestamp, error)
}

// registryFileManager manages the storing of registry file on allocation.
type registryFileManager struct {
	registryFilePath string
	allocation       *Allocation
	fileStorer       allocationFileStorer
}

var _ RegistryFileManager = &registryFileManager{}

func newRegistryFileManager(a *Allocation, registryFilePath string) *registryFileManager {
	return &registryFileManager{
		registryFilePath: registryFilePath,
		allocation:       a,
		fileStorer:       a,
	}
}

// Update writes the registry file with data provided.
// This overwrites the current copy stored.
func (a *registryFileManager) Update(data []byte) error {
	rf, err := a.getRegistryFile()
	if err != nil {
		return errors.New("update_registry_file_failed", "Failed to check existence of registry file: "+err.Error())
	}

	// if no registry yet, upload new registry file.
	// otherwise, update registry file.
	isUpdate := rf != nil

	tempLocal := filepath.Join(os.TempDir(), fmt.Sprintf("registerfile_ul_%d", time.Now().UnixNano()))

	err = os.WriteFile(tempLocal, data, 0600)
	if err != nil {
		return errors.New("update_registry_file_failed", "Failed to create file locally for upload: "+err.Error())
	}

	err = a.fileStorer.StartChunkedUpload(os.TempDir(), tempLocal, a.registryFilePath, nil, isUpdate, false, "", false, fileref.Attributes{})
	if err != nil {
		return errors.New("update_registry_file_failed", "Failed to upload registry file: "+err.Error())
	}

	return nil
}

// Get retrieves the registry file contents and its last updated timestamp.
func (a *registryFileManager) Get() ([]byte, common.Timestamp, error) {
	rfLastUpdateTime, err := a.GetLastUpdateTimestamp()
	if err != nil {
		return nil, common.Timestamp(0), errors.New("get_registry_file_failed", "Failed to check existence of registry file: "+err.Error())
	}

	// no registry file yet
	// return empty
	if rfLastUpdateTime == common.Timestamp(0) {
		return []byte{}, rfLastUpdateTime, nil
	}

	// download registry file
	tempLocal := filepath.Join(os.TempDir(), fmt.Sprintf("starred_dl_%d", time.Now().UnixNano()))
	cb := newRegistryFileCB()

	err = a.fileStorer.DownloadFile(tempLocal, a.registryFilePath, cb)
	if err != nil {
		return nil, common.Timestamp(0), errors.New("get_registry_file_failed", "Failed to start download of registry file: "+err.Error())
	}

	<-cb.Wait() // wait for download to complete
	if cb.err != nil {
		return nil, common.Timestamp(0), errors.New("get_registry_file_failed", "Failed to download registry file: "+cb.err.Error())
	}

	f, err := os.Open(tempLocal)
	if err != nil {
		return nil, common.Timestamp(0), errors.New("get_registry_file_failed", "Failed to open downloaded registry file: "+err.Error())
	}

	defer os.Remove(tempLocal)
	defer f.Close()

	bt, err := io.ReadAll(f)
	if err != nil {
		return nil, common.Timestamp(0), errors.New("get_registry_file_failed", "Failed to read downloaded registry file: "+err.Error())
	}

	return bt, rfLastUpdateTime, nil
}

// GetLastUpdateTimestamp retrieves the last update timestamp of the registry file.
// Returns 0 as timestamp when no registry file exist.
func (a *registryFileManager) GetLastUpdateTimestamp() (common.Timestamp, error) {
	rf, err := a.getRegistryFile()
	if err != nil {
		return common.Timestamp(0), errors.New("get_last_update_timestamp_failed", "Failed to check existence of registry file: "+err.Error())
	}

	// when no registry yet, return beginning of epoch as default last update timestamp.
	if rf == nil {
		return common.Timestamp(0), nil
	}

	lastUpdate, err := time.Parse(time.RFC3339, rf.UpdatedAt)
	if err != nil {
		return common.Timestamp(0), errors.New("get_last_update_timestamp_failed", "Failed to parse last updated timestamp of registry file: "+err.Error())
	}

	return common.Timestamp(lastUpdate.Unix()), nil
}

func (a *registryFileManager) getRegistryFile() (*ListResult, error) {
	dir, err := a.fileStorer.ListDir(a.registryFilePath)
	if err != nil {
		return nil, err
	}

	if dir == nil {
		return nil, nil
	}

	for _, f := range dir.Children {
		if f.Path == a.registryFilePath {
			return f, nil
		}
	}

	return nil, nil
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
