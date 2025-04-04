package sdk

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
)

// For sync app
const (
	// Upload - Upload file to remote
	Upload = "Upload"

	// Download - Download file from remote
	Download = "Download"

	// Update - Update file in remote
	Update = "Update"

	// Delete - Delete file from remote
	Delete = "Delete"

	// Conflict - Conflict in file
	Conflict = "Conflict"

	// LocalDelete - Delete file from local
	LocalDelete = "LocalDelete"
)

// FileInfo file information representation for sync
type FileInfo struct {
	Size         int64            `json:"size"`
	MimeType     string           `json:"mimetype"`
	ActualSize   int64            `json:"actual_size"`
	Hash         string           `json:"hash"`
	Type         string           `json:"type"`
	EncryptedKey string           `json:"encrypted_key"`
	LookupHash   string           `json:"lookup_hash"`
	CreatedAt    common.Timestamp `json:"created_at"`
	UpdatedAt    common.Timestamp `json:"updated_at"`
}

// FileDiff file difference representation for sync
type FileDiff struct {
	Op   string `json:"operation"`
	Path string `json:"path"`
	Type string `json:"type"`
}

func (a *Allocation) getRemoteFilesAndDirs(dirList []string, fMap map[string]FileInfo, exclMap map[string]int, remotePath string) ([]string, error) {
	childDirList := make([]string, 0)
	remotePath = strings.TrimRight(remotePath, "/")
	for _, dir := range dirList {
		ref, err := a.ListDir(dir)
		if err != nil {
			return []string{}, err
		}
		for _, child := range ref.Children {
			if _, ok := exclMap[child.Path]; ok {
				continue
			}
			relativePathFromRemotePath := strings.TrimPrefix(child.Path, remotePath)
			fMap[relativePathFromRemotePath] = FileInfo{
				Size:         child.Size,
				ActualSize:   child.ActualSize,
				Hash:         child.Hash,
				MimeType:     child.MimeType,
				Type:         child.Type,
				EncryptedKey: child.EncryptionKey,
				LookupHash:   child.LookupHash,
				CreatedAt:    child.CreatedAt,
				UpdatedAt:    child.UpdatedAt,
			}
			if child.Type == fileref.DIRECTORY {
				childDirList = append(childDirList, child.Path)
			}
		}
	}
	return childDirList, nil
}

// GetRemoteFileMap retrieve the remote file map
//   - exclMap is the exclude map, a map of paths to exclude
//   - remotepath is the remote path to get the file map
func (a *Allocation) GetRemoteFileMap(exclMap map[string]int, remotepath string) (map[string]FileInfo, error) {
	// 1. Iteratively get dir and files separately till no more dirs left
	remoteList := make(map[string]FileInfo)
	dirs := []string{remotepath}
	var err error
	for {
		dirs, err = a.getRemoteFilesAndDirs(dirs, remoteList, exclMap, remotepath)
		if err != nil {
			l.Logger.Error(err.Error())
			break
		}
		if len(dirs) == 0 {
			break
		}
	}
	for k, v := range remoteList {
		l.Logger.Debug("Remote list : ", k, " ", v)
	}
	return remoteList, err
}

func calcFileHash(filePath string) string {
	fp, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	h := md5.New()
	if _, err := io.Copy(h, fp); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func getRemoteExcludeMap(exclPath []string) map[string]int {
	exclMap := make(map[string]int)
	for idx, path := range exclPath {
		exclMap[strings.TrimRight(path, "/")] = idx
	}
	return exclMap
}

func addLocalFileList(root string, fMap map[string]FileInfo, dirList *[]string, filter map[string]bool, exclMap map[string]int) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			l.Logger.Error("Local file list error for path", path, err.Error())
			return nil
		}
		// Filter out
		if _, ok := filter[info.Name()]; ok {
			return nil
		}
		lPath, err := filepath.Rel(root, path)
		if err != nil {
			l.Logger.Error("getting relative path failed", err)
		}
		// Allocation paths are like unix, so we modify all the backslashes
		// to forward slashes. File path in windows contain backslashes.
		lPath = "/" + strings.ReplaceAll(lPath, "\\", "/")
		// Exclude
		if _, ok := exclMap[lPath]; ok {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		// Add to list
		if info.IsDir() {
			*dirList = append(*dirList, lPath)
		} else {
			fMap[lPath] = FileInfo{Size: info.Size(), Hash: calcFileHash(path), Type: fileref.FILE}
		}
		return nil
	}
}

func getLocalFileMap(rootPath string, filters []string, exclMap map[string]int) (map[string]FileInfo, error) {
	localMap := make(map[string]FileInfo)
	var dirList []string
	filterMap := make(map[string]bool)
	for _, f := range filters {
		filterMap[f] = true
	}
	err := filepath.Walk(rootPath, addLocalFileList(rootPath, localMap, &dirList, filterMap, exclMap))
	// Add the dirs at the end of the list for dir deletiion after all file deletion
	for _, d := range dirList {
		localMap[d] = FileInfo{Type: fileref.DIRECTORY}
	}

	for k, v := range localMap {
		l.Logger.Debug("Local list : ", k, " ", v)
	}

	return localMap, err
}

func isParentFolderExists(lFDiff []FileDiff, path string) bool {
	subdirs := strings.Split(path, "/")
	p := "/"
	for _, dir := range subdirs {
		p = filepath.Join(p, dir)
		for _, f := range lFDiff {
			if f.Path == p {
				return true
			}
		}
	}
	return false
}

func findDelta(remoteMap, localMap, prevMap map[string]FileInfo, localRootPath string) []FileDiff {
	var fileDiffs []FileDiff
	rMod, lMod := make(map[string]FileInfo), make(map[string]FileInfo)
	noCachePrevious := len(prevMap) == 0

	// Identify modified remote files
	for rFile, rInfo := range remoteMap {
		if prev, exists := prevMap[rFile]; exists && (prev.Hash != rInfo.Hash) {
			rMod[rFile] = rInfo
		}
	}

	// Identify modified local files
	for lFile, lInfo := range localMap {
		if remote, exists := remoteMap[lFile]; exists && (remote.Hash != lInfo.Hash) {
			lMod[lFile] = lInfo
		}
	}

	// Determine sync actions for remote files
	for rPath, rInfo := range remoteMap {
		var op = Download
		if _, remoteModified := rMod[rPath]; remoteModified {
			if _, localModified := lMod[rPath]; localModified {
				// In case of conflict and no cache files, use timestamp to decide
				if noCachePrevious {
					// Get local file info for timestamp comparison
					lAbsPath := filepath.Join(localRootPath, rPath)
					if lInfo, err := os.Stat(lAbsPath); err == nil {
						// Compare timestamps - newer wins
						localTime := common.Timestamp(lInfo.ModTime().Unix())
						if localTime > rInfo.UpdatedAt {
							op = Update // Local is newer, upload to remote
						} else {
							op = Download // Remote is newer, download to local
						}
					} else {
						op = Download // If can't get local info, default to download
					}
				} else {
					op = Conflict // Standard conflict when we have previous cache
				}
			} else {
				op = Download // Remote is modified but local is not
			}
		} else if _, exists := localMap[rPath]; exists {
			// Files exist in both places and are identical - no action needed
			delete(localMap, rPath)
			continue
		} else if _, existedBefore := prevMap[rPath]; existedBefore && !noCachePrevious {
			// Only mark for deletion if it existed in previous cache (and we have a cache)
			op = Delete
		} else if noCachePrevious {
			// For initial sync, download all files from remote that don't exist locally
			op = Download
		}
		fileDiffs = append(fileDiffs, FileDiff{Path: rPath, Op: op, Type: remoteMap[rPath].Type})
	}

	// Determine sync actions for local files
	for lPath := range localMap {
		var op = Upload
		if _, modified := lMod[lPath]; modified {
			op = Update
		} else if _, existedBefore := prevMap[lPath]; existedBefore {
			op = LocalDelete
		}

		// Ensure directories are not added for upload unless explicitly required
		if op != LocalDelete {
			lAbsPath := filepath.Join(localRootPath, lPath)
			if fInfo, err := os.Stat(lAbsPath); err == nil && fInfo.IsDir() {
				// Only add directory for upload in first sync
				if !noCachePrevious {
					continue
				}
			}
		}
		fileDiffs = append(fileDiffs, FileDiff{Path: lPath, Op: op, Type: localMap[lPath].Type})
	}

	// Remove child files if parent directory is deleted
	sort.SliceStable(fileDiffs, func(i, j int) bool { return fileDiffs[i].Path < fileDiffs[j].Path })
	var cleanedDiffs []FileDiff
	for _, f := range fileDiffs {
		if (f.Op == Delete || f.Op == LocalDelete) && isParentFolderExists(cleanedDiffs, f.Path) {
			continue
		}
		if f.Type == fileref.FILE || f.Op == Delete || f.Op == LocalDelete {
			cleanedDiffs = append(cleanedDiffs, f)
		}
	}
	return cleanedDiffs
}

// GetAllocationDiff retrieves the difference between the remote and local filesystem representation of the allocation
//   - lastSyncCachePath is the path to the last sync cache file, which carries exact state of the remote filesystem
//   - localRootPath is the local root path of the allocation
//   - localFileFilters is the list of local file filters
//   - remoteExcludePath is the list of remote exclude paths
//   - remotePath is the remote path of the allocation
func (a *Allocation) GetAllocationDiff(lastSyncCachePath, localRootPath string, localFileFilters, remoteExcludePath []string, remotePath string) ([]FileDiff, error) {
	var lFdiff []FileDiff
	prevRemoteFileMap := make(map[string]FileInfo)

	// 1. Validate and Load Previous Sync Cache
	if len(lastSyncCachePath) > 0 {
		if content, err := os.ReadFile(lastSyncCachePath); err == nil {
			if err := json.Unmarshal(content, &prevRemoteFileMap); err != nil {
				return lFdiff, errors.Wrap(err, "Invalid cache content.")
			}
		} else if os.IsNotExist(err) {
			// If cache file is deleted, initialize empty map and log a warning
			prevRemoteFileMap = make(map[string]FileInfo)
			l.Logger.Info("Sync cache file not found. Performing full sync.")
		} else {
			return lFdiff, errors.Wrap(err, "Error reading sync cache file.")
		}
	}

	// 2. Build Exclusion Map
	exclMap := getRemoteExcludeMap(remoteExcludePath)

	// 3. Get Remote and Local File Maps
	remoteFileMap, err := a.GetRemoteFileMap(exclMap, remotePath)
	if err != nil {
		return lFdiff, errors.Wrap(err, "Error retrieving remote directory list.")
	}

	localRootPath = strings.TrimRight(localRootPath, "/")
	localFileList, err := getLocalFileMap(localRootPath, localFileFilters, exclMap)
	if err != nil {
		return lFdiff, errors.Wrap(err, "Error retrieving local directory list.")
	}

	// 4. Compute File Differences
	lFdiff = findDelta(remoteFileMap, localFileList, prevRemoteFileMap, localRootPath)
	return lFdiff, nil
}

// SaveRemoteSnapshot saves the remote current information to the given file.
// This file can be passed to GetAllocationDiff to exactly find the previous sync state to current.
//   - pathToSave is the path to save the remote snapshot
//   - remoteExcludePath is the list of paths to exclude
func (a *Allocation) SaveRemoteSnapshot(pathToSave string, remoteExcludePath []string) error {
	bIsFileExists := false
	// Validate path
	fileInfo, err := sys.Files.Stat(pathToSave)
	if err == nil {
		if fileInfo.IsDir() {
			return errors.Wrap(err, "invalid file path to save.")
		}
		bIsFileExists = true
	}

	// Get flat file list from remote
	exclMap := getRemoteExcludeMap(remoteExcludePath)
	remoteFileList, err := a.GetRemoteFileMap(exclMap, "/")
	if err != nil {
		return errors.Wrap(err, "error getting list dir from remote.")
	}

	// Now we got the list from remote, delete the file if exists
	if bIsFileExists {
		err = os.Remove(pathToSave)
		if err != nil {
			return errors.Wrap(err, "error deleting previous cache.")
		}
	}
	by, err := json.Marshal(remoteFileList)
	if err != nil {
		return errors.Wrap(err, "failed to convert JSON.")
	}
	err = os.WriteFile(pathToSave, by, 0644)
	if err != nil {
		return errors.Wrap(err, "error saving file.")
	}
	// Successfully saved
	return nil
}
