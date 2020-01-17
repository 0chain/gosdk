package sdk

import (
	"strings"
)

func (a *Allocation) BatchUploader(lastSyncCachePath string, localRootPath string, localFileFilters []string, remoteExcludePath []string, status StatusCallback) error {
	lDiff, err := a.GetAllocationDiff(lastSyncCachePath, localRootPath, localFileFilters, remoteExcludePath)
	if err != nil {
		return err
	}

	for _, f := range lDiff {
		localpath := strings.TrimRight(localRootPath, "/")
		lPath := localpath + f.Path
		switch f.Op {
		case Upload:
			err = a.UploadFile(lPath, f.Path, status)
		case Update:
			err = a.UpdateFile(lPath, f.Path, status)
		}
	}
	return a.saveCache(lastSyncCachePath, remoteExcludePath)
}

func (a *Allocation) saveCache(path string, exclPath []string) error {
	if len(path) > 0 {
		err := a.SaveRemoteSnapshot(path, exclPath)
		if err != nil {
			return err
		}
	}
	return nil
}
