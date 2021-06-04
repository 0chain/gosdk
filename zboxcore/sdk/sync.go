package sdk

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sort"

	"os"
	"path/filepath"
	"strings"

	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
)

// For sync app
const (
	Upload      = "Upload"
	Download    = "Download"
	Update      = "Update"
	Delete      = "Delete"
	Conflict    = "Conflict"
	LocalDelete = "LocalDelete"
)

type fileInfo struct {
	Size       int64  `json:"size"`
	ActualSize int64  `json:"actual_size"`
	Hash       string `json:"hash"`
	Type       string `json:"type"`
}

type FileDiff struct {
	Op   string `json:"operation"`
	Path string `json:"path"`
	Type string `json:"type"`
}

func (a *Allocation) getRemoteFilesAndDirs(dirList []string, fMap map[string]fileInfo, exclMap map[string]int) ([]string, error) {
	childDirList := make([]string, 0)
	for _, dir := range dirList {
		ref, err := a.ListDir(dir)
		if err != nil {
			return []string{}, err
		}
		for _, child := range ref.Children {
			if _, ok := exclMap[child.Path]; ok {
				continue
			}
			fMap[child.Path] = fileInfo{Size: child.Size, ActualSize: child.ActualSize, Hash: child.Hash, Type: child.Type}
			if child.Type == fileref.DIRECTORY {
				childDirList = append(childDirList, child.Path)
			}
		}
	}
	return childDirList, nil
}

func (a *Allocation) GetRemoteFileMap(exclMap map[string]int) (map[string]fileInfo, error) {
	// 1. Iteratively get dir and files separately till no more dirs left
	remoteList := make(map[string]fileInfo)
	dirs := []string{"/"}
	var err error
	for {
		dirs, err = a.getRemoteFilesAndDirs(dirs, remoteList, exclMap)
		if err != nil {
			Logger.Error(err.Error())
			break
		}
		if len(dirs) == 0 {
			break
		}
	}
	Logger.Debug("Remote List: ", remoteList)
	return remoteList, err
}

func calcFileHash(filePath string) string {
	fp, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	h := sha1.New()
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

func addLocalFileList(root string, fMap map[string]fileInfo, dirList *[]string, filter map[string]bool, exclMap map[string]int) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			Logger.Error("Local file list error for path", path, err.Error())
			return nil
		}
		// Filter out
		if _, ok := filter[info.Name()]; ok {
			return nil
		}
		lPath, err := filepath.Rel(root, path)
		if err != nil {
			Logger.Error("getting relative path failed", err)
		}
		lPath = "/" + lPath
		// Exclude
		if _, ok := exclMap[lPath]; ok {
			return nil
		}
		// Add to list
		if info.IsDir() {
			*dirList = append(*dirList, lPath)
		} else {
			fMap[lPath] = fileInfo{Size: info.Size(), Hash: calcFileHash(path), Type: fileref.FILE}
		}
		return nil
	}
}

func getLocalFileMap(rootPath string, filters []string, exclMap map[string]int) (map[string]fileInfo, error) {
	localMap := make(map[string]fileInfo)
	var dirList []string
	filterMap := make(map[string]bool)
	for _, f := range filters {
		filterMap[f] = true
	}
	err := filepath.Walk(rootPath, addLocalFileList(rootPath, localMap, &dirList, filterMap, exclMap))
	// Add the dirs at the end of the list for dir deletiion after all file deletion
	for _, d := range dirList {
		localMap[d] = fileInfo{Type: fileref.DIRECTORY}
	}
	Logger.Debug("Local List: ", localMap)
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

func findDelta(rMap map[string]fileInfo, lMap map[string]fileInfo, prevMap map[string]fileInfo, localRootPath string) []FileDiff {
	var lFDiff []FileDiff

	// Create a remote hash map and find modifications
	rMod := make(map[string]fileInfo)
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
	lMod := make(map[string]fileInfo)
	for lFile, lInfo := range lMap {
		if pm, ok := prevMap[lFile]; ok {
			// Local file existed in previous sync also
			if pm.Hash != lInfo.Hash {
				// File modified in local
				lMod[lFile] = lInfo
			}
		}
	}

	// Iterate remote list and get diff
	rDelMap := make(map[string]string)
	for rPath, _ := range rMap {
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
	for lPath, _ := range lMap {
		op := Upload
		if _, ok := lMod[lPath]; ok {
			op = Update
		} else if _, ok := prevMap[lPath]; ok {
			op = LocalDelete
		}
		if op != LocalDelete {
			// Skip if it is a directory
			lAbsPath := filepath.Join(localRootPath, lPath)
			fInfo, err := os.Stat(lAbsPath)
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
		Logger.Debug("Sorted diff: ", lFDiff)
		var newlFDiff []FileDiff
		for _, f := range lFDiff {
			if f.Op == LocalDelete || f.Op == Delete {
				if isParentFolderExists(newlFDiff, f.Path) == false {
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

func (a *Allocation) GetAllocationDiff(lastSyncCachePath string, localRootPath string, localFileFilters []string, remoteExcludePath []string) ([]FileDiff, error) {
	var lFdiff []FileDiff
	prevRemoteFileMap := make(map[string]fileInfo)
	// 1. Validate localSycnCachePath
	if len(lastSyncCachePath) > 0 {
		// Validate cache path
		fileInfo, err := os.Stat(lastSyncCachePath)
		if err == nil {
			if fileInfo.IsDir() {
				return lFdiff, fmt.Errorf("invalid file cache. %v", err)
			}
			content, err := ioutil.ReadFile(lastSyncCachePath)
			if err != nil {
				return lFdiff, fmt.Errorf("can't read cache file.")
			}
			err = json.Unmarshal(content, &prevRemoteFileMap)
			if err != nil {
				return lFdiff, fmt.Errorf("invalid cache content.")
			}
		}
	}

	// 2. Build a map for exclude path
	exclMap := getRemoteExcludeMap(remoteExcludePath)

	// 3. Get flat file list from remote
	remoteFileMap, err := a.GetRemoteFileMap(exclMap)
	if err != nil {
		return lFdiff, fmt.Errorf("error getting list dir from remote. %v", err)
	}

	// 4. Get flat file list on the local filesystem
	localRootPath = strings.TrimRight(localRootPath, "/")
	localFileList, err := getLocalFileMap(localRootPath, localFileFilters, exclMap)
	if err != nil {
		return lFdiff, fmt.Errorf("error getting list dir from local. %v", err)
	}

	// 5. Get the file diff with operation
	lFdiff = findDelta(remoteFileMap, localFileList, prevRemoteFileMap, localRootPath)
	Logger.Debug("Diff: ", lFdiff)
	return lFdiff, nil
}

// SaveRemoteSnapShot - Saves the remote current information to the given file
// This file can be passed to GetAllocationDiff to exactly find the previous sync state to current.
func (a *Allocation) SaveRemoteSnapshot(pathToSave string, remoteExcludePath []string) error {
	bIsFileExists := false
	// Validate path
	fileInfo, err := os.Stat(pathToSave)
	if err == nil {
		if fileInfo.IsDir() {
			return fmt.Errorf("invalid file path to save. %v", err)
		}
		bIsFileExists = true
	}

	// Get flat file list from remote
	exclMap := getRemoteExcludeMap(remoteExcludePath)
	remoteFileList, err := a.GetRemoteFileMap(exclMap)
	if err != nil {
		return fmt.Errorf("error getting list dir from remote. %v", err)
	}

	// Now we got the list from remote, delete the file if exists
	if bIsFileExists {
		err = os.Remove(pathToSave)
		if err != nil {
			return fmt.Errorf("error deleting previous cache. %v", err)
		}
	}
	by, err := json.Marshal(remoteFileList)
	if err != nil {
		return fmt.Errorf("failed to convert JSON. %v", err)
	}
	err = ioutil.WriteFile(pathToSave, by, 0644)
	if err != nil {
		return fmt.Errorf("error saving file. %v", err)
	}
	// Successfully saved
	return nil
}
