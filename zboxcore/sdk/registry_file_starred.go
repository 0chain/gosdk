package sdk

import (
	"encoding/json"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
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
	UpdatedAt common.Timestamp `json:"-"` // on read of `.starred`, will be populated with `updated_at`, on write the value is ignored
	Files     []StarredFile    `json:"files"`
}

// StarredFile defines the individual entry on starred registry file.
type StarredFile struct {
	Path string `json:"path"`
}

// UpdateStarredFiles writes the provided full list of starred files through the registry.
func (a *Allocation) UpdateStarredFiles(files *StarredFiles) error {
	if files == nil {
		return errors.New("update_starred_files_failed", "Starred files is nil")
	}

	bt, err := json.Marshal(files)
	if err != nil {
		return errors.New("update_starred_files_failed", "Failed to marshal starred files: "+err.Error())
	}

	err = starredFileRegistryManager(a).Update(bt)
	if err != nil {
		return errors.New("update_starred_files_failed", "Failed to update registry file for starred files: "+err.Error())
	}

	return nil
}

// GetStarredFiles returns the full list of starred files through the registry.
func (a *Allocation) GetStarredFiles() (*StarredFiles, error) {
	data, lastUpdateTime, err := starredFileRegistryManager(a).Get()
	if err != nil {
		return nil, errors.New("get_starred_files_failed", "Failed to retrieve starred files: "+err.Error())
	}

	if len(data) == 0 {
		return &StarredFiles{UpdatedAt: lastUpdateTime, Files: []StarredFile{}}, nil
	}

	starred := &StarredFiles{}

	err = json.Unmarshal(data, starred)
	if err != nil {
		return nil, errors.New("get_starred_files_failed", "Failed to parse downloaded registry file for starred files: "+err.Error())
	}

	starred.UpdatedAt = lastUpdateTime

	return starred, nil
}

// GetStarredFilesLastUpdateTimestamp retrieves the latest updated timestamp of the starred file registry.
func (a *Allocation) GetStarredFilesLastUpdateTimestamp() (common.Timestamp, error) {
	lastUpdateTime, err := starredFileRegistryManager(a).GetLastUpdateTimestamp()
	if err != nil {
		return common.Timestamp(0), errors.New("get_starred_files_last_update_timestamp_failed", "Failed to get last update timestamp of registry file for starred files: "+err.Error())
	}

	return lastUpdateTime, nil
}

// starredFileRegistryManager is the factory method for managing registry file for starred files.
// this is a variable to easy mocking for UT purposes.
var starredFileRegistryManager = func(a *Allocation) RegistryFileManager {
	return newRegistryFileManager(a, StarredRegistryFilePath)
}
