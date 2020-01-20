package sdk

import (
	"strings"
)

func (a *Allocation) BatchUploader(lastSyncCachePath string, localRootPath string, localFileFilters []string, remoteExcludePath []string, status StatusCallback) ([]FileDiff, error) {
	lDiff, err := a.GetAllocationDiff(lastSyncCachePath, localRootPath, localFileFilters, remoteExcludePath)
	if err != nil {
		return nil, err
	}

	lDiff = filterOperations(lDiff)
	for _, f := range lDiff {
		localpath := strings.TrimRight(localRootPath, "/")
		lPath := localpath + f.Path
		switch f.Op {
		case Upload:
			err = a.UploadFile(lPath, f.Path, status)
		case Update:
			err = a.UpdateFile(lPath, f.Path, status)
		}
		if err != nil {
			return nil, err
		}
	}
	return lDiff, nil
}

func filterOperations(lDiff []FileDiff) (filterDiff []FileDiff) {
	for _, f := range lDiff {
		if f.Op == Update || f.Op == Upload {
			filterDiff = append(filterDiff, f)
		}
	}
	return
}
