package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"0chain.net/clientsdk/encryption"
)

type FileDirInfo struct {
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Hash     string                 `json:"hash"`
	Size     int64                  `json:"size,omit_empty"`
	Meta     map[string]interface{} `json:"meta_data",omit_empty`
	Children []FileDirInfo          `json:"children,omit_empty"`
}

func (fi *FileDirInfo) GetInfoHash() string {
	return encryption.Hash(fi.Type + ":" + fi.Name + ":" + strconv.FormatInt(fi.Size, 10) + ":" + fi.Hash)
}

type FileConfig struct {
	Name       string `json:"Name"`
	Path       string `json:"Path"`
	Size       int64  `json:"Size"`
	Type       string `json:"Type"`
	ActualHash string
	FileHash   hash.Hash
	FileHashWr io.Writer
	Remaining  int64
}

var (
	FILEEXISTS = errors.New("File already exists")
)

func NewDirTree() FileDirInfo {
	var fD FileDirInfo
	fD.Type = "d"
	fD.Name = "/"
	return fD
}

func GetDirTreeFromJson(j string) (FileDirInfo, error) {
	var root FileDirInfo
	dec := json.NewDecoder(strings.NewReader(j))
	for {
		if err := dec.Decode(&root); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal("Parsing directory failed: ", err)
			return FileDirInfo{}, err
		}
	}
	return root, nil
}

func CalculateDirHash(d *FileDirInfo) string {
	childHashes := make([]string, 0)
	for i := 0; i < len(d.Children); i++ {
		child := &d.Children[i]
		if child.Type == "d" {
			if len(child.Children) > 0 {
				child.Hash = CalculateDirHash(child)
			} else {
				child.Hash = encryption.Hash("")
			}
		}
		childHashes = append(childHashes, child.Hash)
	}
	d.Hash = encryption.Hash(strings.Join(childHashes, ":"))
	return d.Hash
}

func checkAndAddDir(d *FileDirInfo, s string) *FileDirInfo {
	for i := 0; i < len(d.Children); i++ {
		child := &d.Children[i]
		if s == child.Name {
			return child
		}
	}
	return appendChild(d, "d", s, "", 0)
}

func isFileExists(d *FileDirInfo, s string) *FileDirInfo {
	for i := 0; i < len(d.Children); i++ {
		child := &d.Children[i]
		if s == child.Name {
			return child
		}
	}
	return nil
}

func appendChild(d *FileDirInfo, t, n, h string, l int64) *FileDirInfo {
	fileInfo := FileDirInfo{
		Type: t,
		Name: n,
		Hash: h,
		Size: l,
	}
	if t == "f" {
		fileInfo.Meta = make(map[string]interface{})
	}
	d.Children = append(d.Children, fileInfo)
	return &d.Children[(len(d.Children) - 1)]
}

func getSubDirs(p string) []string {
	subDirs := strings.Split(p, "/")
	tSubDirs := make([]string, 0)
	for _, s := range subDirs {
		if s != "" {
			tSubDirs = append(tSubDirs, s)
		}
	}
	return tSubDirs
}

func InsertFile(d *FileDirInfo, path, hash string, size int64) (*FileDirInfo, error) {
	path, fileName := filepath.Split(path)
	child := AddDir(d, path)
	if isFileExists(child, fileName) != nil {
		return nil, FILEEXISTS
	}
	return appendChild(child, "f", fileName, hash, size), nil
}

func AddDir(d *FileDirInfo, path string) *FileDirInfo {
	tSubDirs := getSubDirs(path)
	child := d
	for _, subPath := range tSubDirs {
		child = checkAndAddDir(child, subPath)
	}
	return child
}

func getChildDir(d *FileDirInfo, s string) *FileDirInfo {
	for i := 0; i < len(d.Children); i++ {
		child := &d.Children[i]
		if s == child.Name && "d" == child.Type {
			return child
		}
	}
	return nil
}

func ListDir(d *FileDirInfo, path string) []FileDirInfo {
	tSubDirs := getSubDirs(path)
	child := d
	for _, subPath := range tSubDirs {
		child = getChildDir(child, subPath)
		if child == nil {
			return nil
		}
	}
	return child.Children
}

func GetFileInfo(d *FileDirInfo, path string) *FileDirInfo {
	path, fileName := filepath.Split(path)
	tSubDirs := getSubDirs(path)
	child := d
	for _, subPath := range tSubDirs {
		child = getChildDir(child, subPath)
		if child == nil {
			return nil
		}
	}
	return isFileExists(child, fileName)
}

func DeleteFile(d *FileDirInfo, path string) error {
	path, fileName := filepath.Split(path)
	tSubDirs := getSubDirs(path)
	child := d
	for _, subPath := range tSubDirs {
		child = getChildDir(child, subPath)
		if child == nil {
			return fmt.Errorf("File folder doesn't exist")
		}
	}
	childIdx := int(-1)
	for i := 0; i < len(child.Children); i++ {
		sd := &child.Children[i]
		if fileName == sd.Name {
			childIdx = i
		}
	}
	if childIdx == -1 {
		return fmt.Errorf("File doesn't exist")
	}
	child.Hash = ""
	child.Children = append(child.Children[0:childIdx], child.Children[childIdx+1:]...)
	return nil
}

func GetJsonFromDirTree(d *FileDirInfo) string {
	by, err := json.Marshal(d)
	if err != nil {
		return "{}"
	}
	return string(by)
}

func GetFileConfig(j string) (FileConfig, error) {
	var file FileConfig
	err := json.Unmarshal([]byte(j), &file)
	return file, err
}

func GetFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)
	// fmt.Println("Found content type : " + contentType)

	return contentType, nil
}
