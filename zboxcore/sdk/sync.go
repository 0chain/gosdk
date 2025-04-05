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

	count := 1
	for k := range remoteList {
		l.Logger.Debug("Remote list : ", count, " ", k)
		count++
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

	count := 0
	for k := range localMap {
		l.Logger.Debug("Local list : ", count, " ", k)
		count++
	}

	return localMap, err
}

func findDelta(remoteMap, localMap, prevRemoteMap map[string]FileInfo, localRootPath string) []FileDiff {
	var fileDiffs []FileDiff
	rMod, lMod := make(map[string]FileInfo), make(map[string]FileInfo)

	// Identify modified remote files
	for rFile, rInfo := range remoteMap {
		if prev, exists := prevRemoteMap[rFile]; exists && (prev.Hash != rInfo.Hash) {
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
	for rPath := range remoteMap {
		var op string
		if _, remoteModified := rMod[rPath]; remoteModified {
			if _, localModified := lMod[rPath]; localModified {
				op = Conflict
			} else {
				op = Download
			}
		} else if _, exists := localMap[rPath]; exists {
			delete(localMap, rPath)
			continue
		} else if _, existedBefore := prevRemoteMap[rPath]; existedBefore {
			op = Delete
		} else {
			op = Download
		}
		fileDiffs = append(fileDiffs, FileDiff{Path: rPath, Op: op, Type: remoteMap[rPath].Type})
	}

	// Determine sync actions for local files
	for lPath := range localMap {
		var op string
		if _, localModified := lMod[lPath]; localModified {
			if _, existOnRemote := remoteMap[lPath]; existOnRemote {
				op = Update
			} else {
				op = Upload
			}
		} else if _, existedInRemote := remoteMap[lPath]; !existedInRemote {
			if _, existedBefore := prevRemoteMap[lPath]; existedBefore {
				op = LocalDelete
			} else {
				op = Upload
			}
		} else {
			// This is a file that exists in both places but wasn't in local modified list
			// Skip it as it's already handled in remote file processing
			continue
		}

		// For directories: include all operations to ensure proper directory structure
		// is created before files are processed
		fileDiffs = append(fileDiffs, FileDiff{Path: lPath, Op: op, Type: localMap[lPath].Type})
	}

	// Sort paths to ensure parent directories are processed before their children
	sort.SliceStable(fileDiffs, func(i, j int) bool { return fileDiffs[i].Path < fileDiffs[j].Path })

	// Group operations by type to ensure proper sequence (create dirs first, then handle files)
	var dirCreateOps, fileOps, deleteOps []FileDiff
	for _, f := range fileDiffs {
		if f.Type == fileref.DIRECTORY && (f.Op == Upload || f.Op == Update) {
			// Directory creation operations first
			dirCreateOps = append(dirCreateOps, f)
		} else if f.Op == Delete || f.Op == LocalDelete {
			// Deletion operations last
			deleteOps = append(deleteOps, f)
		} else {
			// All other file operations
			fileOps = append(fileOps, f)
		}
		// Log the operation for debugging
		l.Logger.Debug("Sync operation detected:", f.Op, f.Path, f.Type)
	}

	// Combine operations in proper sequence: create dirs → file operations → deletions
	cleanedDiffs := append(dirCreateOps, fileOps...)
	cleanedDiffs = append(cleanedDiffs, deleteOps...)
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

	count := 1
	for k := range prevRemoteFileMap {
		l.Logger.Debug("Previous remote list : ", count, " ", k)
		count++
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
