package sdk

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
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
	l.Logger.Debug("Remote List: ", remoteList)
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
	l.Logger.Debug("Local List: ", localMap)
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

func findDelta(rMap map[string]FileInfo, lMap map[string]FileInfo, prevMap map[string]FileInfo, localRootPath string) []FileDiff {
	var lFDiff []FileDiff

	// Create a remote hash map and find modifications
	rMod := make(map[string]FileInfo)
	for rFile, rInfo := range rMap {
		if pm, ok := prevMap[rFile]; ok {
			// Remote file existed in previous sync also
			if pm.Hash != rInfo.Hash {
				// File modified in remote
				rMod[rFile] = rInfo
			}
		}
	}

	// Create a local hash map and find modification
	lMod := make(map[string]FileInfo)
	for lFile, lInfo := range lMap {
		if pm, ok := rMap[lFile]; ok {
			// Local file existed in previous sync also
			if pm.Hash != lInfo.Hash {
				// File modified in local
				lMod[lFile] = lInfo
			}
		}
	}

	// Iterate remote list and get diff
	rDelMap := make(map[string]string)
	for rPath := range rMap {
		op := Download
		bRemoteModified := false
		bLocalModified := false
		if _, ok := rMod[rPath]; ok {
			bRemoteModified = true
		}
		if _, ok := lMod[rPath]; ok {
			bLocalModified = true
			delete(lMap, rPath)
		}
		if bRemoteModified && bLocalModified {
			op = Conflict
		} else if bLocalModified {
			op = Update
		} else if _, ok := lMap[rPath]; ok {
			// No conflicts and file exists locally
			delete(lMap, rPath)
			continue
		} else if _, ok := prevMap[rPath]; ok {
			op = Delete
			// Remote allows delete directory skip individual file deletion
			rDelMap[rPath] = "d"
			rDir, _ := filepath.Split(rPath)
			rDir = strings.TrimRight(rDir, "/")
			if _, ok := rDelMap[rDir]; ok {
				continue
			}
		}
		lFDiff = append(lFDiff, FileDiff{Path: rPath, Op: op, Type: rMap[rPath].Type})
	}

	// Upload all local files
	for lPath := range lMap {
		op := Upload
		if _, ok := lMod[lPath]; ok {
			op = Update
		} else if _, ok := prevMap[lPath]; ok {
			op = LocalDelete
		}
		if op != LocalDelete {
			// Skip if it is a directory
			lAbsPath := filepath.Join(localRootPath, lPath)
			fInfo, err := sys.Files.Stat(lAbsPath)
			if err != nil {
				continue
			}
			if fInfo.IsDir() {
				continue
			}
		}
		lFDiff = append(lFDiff, FileDiff{Path: lPath, Op: op, Type: lMap[lPath].Type})
	}

	// If there are differences, remove childs if the parent folder is deleted
	if len(lFDiff) > 0 {
		sort.SliceStable(lFDiff, func(i, j int) bool { return lFDiff[i].Path < lFDiff[j].Path })
		l.Logger.Debug("Sorted diff: ", lFDiff)
		var newlFDiff []FileDiff
		for _, f := range lFDiff {
			if f.Op == LocalDelete || f.Op == Delete {
				if !isParentFolderExists(newlFDiff, f.Path) {
					newlFDiff = append(newlFDiff, f)
				}
			} else {
				// Add only files for other Op
				if f.Type == fileref.FILE {
					newlFDiff = append(newlFDiff, f)
				}
			}
		}
		return newlFDiff
	}
	return lFDiff
}

// GetAllocationDiff retrieves the difference between the remote and local filesystem representation of the allocation
//   - lastSyncCachePath is the path to the last sync cache file, which carries exact state of the remote filesystem
//   - localRootPath is the local root path of the allocation
//   - localFileFilters is the list of local file filters
//   - remoteExcludePath is the list of remote exclude paths
//   - remotePath is the remote path of the allocation
func (a *Allocation) GetAllocationDiff(lastSyncCachePath string, localRootPath string, localFileFilters []string, remoteExcludePath []string, remotePath string) ([]FileDiff, error) {
	var lFdiff []FileDiff
	prevRemoteFileMap := make(map[string]FileInfo)
	// 1. Validate localSycnCachePath
	if len(lastSyncCachePath) > 0 {
		// Validate cache path
		fileInfo, err := sys.Files.Stat(lastSyncCachePath)
		if err == nil {
			if fileInfo.IsDir() {
				return lFdiff, errors.Wrap(err, "invalid file cache.")
			}
			content, err := ioutil.ReadFile(lastSyncCachePath)
			if err != nil {
				return lFdiff, errors.New("", "can't read cache file.")
			}
			err = json.Unmarshal(content, &prevRemoteFileMap)
			if err != nil {
				return lFdiff, errors.New("", "invalid cache content.")
			}
		}
	}

	// 2. Build a map for exclude path
	exclMap := getRemoteExcludeMap(remoteExcludePath)

	// 3. Get flat file list from remote
	remoteFileMap, err := a.GetRemoteFileMap(exclMap, remotePath)
	if err != nil {
		return lFdiff, errors.Wrap(err, "error getting list dir from remote.")
	}

	// 4. Get flat file list on the local filesystem
	localRootPath = strings.TrimRight(localRootPath, "/")
	localFileList, err := getLocalFileMap(localRootPath, localFileFilters, exclMap)
	if err != nil {
		return lFdiff, errors.Wrap(err, "error getting list dir from local.")
	}

	// 5. Get the file diff with operation
	lFdiff = findDelta(remoteFileMap, localFileList, prevRemoteFileMap, localRootPath)
	l.Logger.Debug("Diff: ", lFdiff)
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
	err = ioutil.WriteFile(pathToSave, by, 0644)
	if err != nil {
		return errors.Wrap(err, "error saving file.")
	}
	// Successfully saved
	return nil
}
