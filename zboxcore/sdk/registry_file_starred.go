package sdk

import (
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"strings"
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

// UpdateStarredFiles writes the provided full list of starred files through the registry.
func (a *Allocation) UpdateStarredFiles(paths []string) error {
	data := []byte(strings.Join(paths, "\n"))

	err := starredFileRegistryManager(a).Update(data)
	if err != nil {
		return errors.New("update_starred_files_failed", "failed to update registry file for starred files: "+err.Error())
	}

	return nil
}

// GetStarredFiles returns the full list of starred files through the registry.
func (a *Allocation) GetStarredFiles() ([]string, common.Timestamp, error) {
	data, lastUpdateTime, err := starredFileRegistryManager(a).Get()
	if err != nil {
		return nil, common.Timestamp(0), errors.New("get_starred_files_failed", "failed to retrieve starred files: "+err.Error())
	}

	paths := []string{}
	if strings.TrimSpace(string(data)) == "" {
		return paths, lastUpdateTime, nil
	}

	paths = strings.Split(string(data), "\n")

	return paths, lastUpdateTime, nil
}

// GetStarredFilesLastUpdateTimestamp retrieves the latest updated timestamp of the starred file registry.
func (a *Allocation) GetStarredFilesLastUpdateTimestamp() (common.Timestamp, error) {
	lastUpdateTime, err := starredFileRegistryManager(a).GetLastUpdateTimestamp()
	if err != nil {
		return common.Timestamp(0), errors.New("get_starred_files_last_update_timestamp_failed", "failed to get last update timestamp of registry file for starred files: "+err.Error())
	}

	return lastUpdateTime, nil
}

// starredFileRegistryManager is the factory method for managing registry file for starred files.
// this is a variable to easy mocking for UT purposes.
var starredFileRegistryManager = func(a *Allocation) registryFileManager {
	return newRegistryFileManager(a, StarredRegistryFilePath)
}
