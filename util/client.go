package util

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/0chain/gosdk/encryption"
)

type OperationStatus struct {
	Path string
	Size int64
}

type Connection struct {
	ConnectionId int64
	DirTree      FileDirInfo
	uploadList   []OperationStatus
	deleteList   []OperationStatus
}

type Blobber struct {
	Id          string      `json:"id"`
	UrlRoot     string      `json:"url"`
	ReadCounter int64       `json:"counter"`
	DirTree     FileDirInfo `json:"tree,omit_empty"`
	ConnObj     Connection  `json:"-"`
}

type ClientConfig struct {
	Id         string `json:"id"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type MarkerI interface {
	GetHash() string
	Sign(privateKey string) error
}

type WriteMarker struct {
	AllocationRoot         string `json:"allocation_root"`
	PreviousAllocationRoot string `json:"prev_allocation_root"`
	AllocationID           string `json:"allocation_id"`
	Size                   int64  `json:"size"`
	BlobberID              string `json:"blobber_id"`
	Timestamp              int64  `json:"timestamp"`
	ClientID               string `json:"client_id"`
	Signature              string `json:"signature"`
}

type ReadMarker struct {
	ClientID        string `json:"client_id"`
	ClientPublicKey string `json:"client_public_key"`
	BlobberID       string `json:"blobber_id"`
	AllocationID    string `json:"allocation_id"`
	OwnerID         string `json:"owner_id"`
	Timestamp       int64  `json:"timestamp"`
	ReadCounter     int64  `json:"counter"`
	Signature       string `json:"signature"`
}

type DeleteToken struct {
	FilePathHash string `json:"file_path_hash"`
	FileRefHash  string `json:"file_ref_hash"`
	AllocationID string `json:"allocation_id"`
	Size         int64  `json:"size"`
	BlobberID    string `json:"blobber_id"`
	Timestamp    int64  `json:"timestamp"`
	ClientID     string `json:"client_id"`
	Signature    string `json:"signature"`
	Status       int    `json:"status"`
}

const UPLOAD_ENDPOINT = "/v1/file/upload/"
const LIST_ENDPOINT = "/v1/file/list/"
const CONNECTION_ENDPOINT = "/v1/connection/details/"
const COMMIT_ENDPOINT = "/v1/connection/commit/"
const DOWNLOAD_ENDPOINT = "/v1/file/download/"
const LATEST_READ_MARKER = "/v1/readmarker/latest"
const FILE_META_ENDPOINT = "/v1/file/meta/"
const FILE_STATS_ENDPOINT = "/v1/file/stats/"

/*Now - current datetime */
func Now() int64 {
	return time.Now().Unix()
}

func NewConnectionId() int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(0xffffffff))
	if err != nil {
		panic(err)
	}
	return nBig.Int64()
}

func (c *Connection) Reset() {
	c.ConnectionId = NewConnectionId()
	c.uploadList = make([]OperationStatus, 0)
	c.deleteList = make([]OperationStatus, 0)
}

func (c *Connection) GetAllocationRoot(timestamp int64) string {
	_ = CalculateDirHash(&c.DirTree)
	c.DirTree.Hash = encryption.Hash(c.DirTree.Hash + ":" + strconv.FormatInt(timestamp, 10))
	return c.DirTree.Hash
}

func (c *Connection) AddFile(path, hash string, size int64) error {
	_, err := InsertFile(&c.DirTree, path, hash, size)
	if err != nil {
		return err
	}
	status := OperationStatus{Path: path, Size: size}
	c.uploadList = append(c.uploadList, status)
	return nil
}

func (c *Connection) UpdateFile(path, hash string, size int64) error {
	flInfo := GetFileInfo(&c.DirTree, path)
	if flInfo == nil {
		return fmt.Errorf("File not found")
	}
	flInfo.Hash = hash
	// New upload size is previo
	status := OperationStatus{Path: path, Size: (size - flInfo.Size)}
	c.uploadList = append(c.uploadList, status)
	flInfo.Size = size
	return nil
}

func (c *Connection) DeleteFile(path string, size int64) error {
	err := DeleteFile(&c.DirTree, path)
	if err != nil {
		return err
	}
	status := OperationStatus{Path: path, Size: size}
	c.deleteList = append(c.deleteList, status)
	return nil
}

func (c *Connection) GetSize() int64 {
	totalSize := int64(0)
	for _, file := range c.uploadList {
		totalSize += file.Size
	}
	for _, file := range c.deleteList {
		totalSize -= file.Size
	}
	return totalSize
}

func (c *Connection) GetCommitData() string {
	var s strings.Builder
	s.WriteString("Upload/Update: ")
	for _, file := range c.uploadList {
		fmt.Fprintf(&s, "%v ", file)
	}
	s.WriteString("Delete: ")
	for _, file := range c.deleteList {
		fmt.Fprintf(&s, "%v", file)
	}
	return s.String()
}

func GetBlobbers(j string) ([]Blobber, error) {
	dec := json.NewDecoder(strings.NewReader(j))

	// read open bracket
	_, err := dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	blobbers := make([]Blobber, 0)
	// while the array contains values
	for dec.More() {
		var b Blobber
		// decode an array value (Message)
		err := dec.Decode(&b)
		if err != nil {
			log.Fatal(err)
		}
		// Valid dir tree
		if b.DirTree.Type != "d" || b.DirTree.Name != "/" {
			b.DirTree.Type = "d"
			b.DirTree.Name = "/"
		}
		blobbers = append(blobbers, b)
	}

	// read closing bracket
	_, err = dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	return blobbers, nil
}

func SetBlobberDirTree(b *Blobber, j string) error {
	dir, err := GetDirTreeFromJson(j)
	if err == nil {
		b.DirTree = dir
	}
	return err
}

func GetBlobberJson(b *Blobber) string {
	by, err := json.Marshal(b)
	if err != nil {
		return "{}"
	}
	return string(by)
}

func GetClientConfig(j string) (ClientConfig, error) {
	var client ClientConfig
	err := json.Unmarshal([]byte(j), &client)
	return client, err
}

func setClientInfo(req *http.Request, err error, client ClientConfig) (*http.Request, error) {
	if err == nil {
		req.Header.Set("X-App-Client-ID", client.Id)
		req.Header.Set("X-App-Client-Key", client.PublicKey)
	}
	return req, err
}

func NewUploadRequest(baseUrl, allocation string, client ClientConfig, body io.Reader, update bool) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, UPLOAD_ENDPOINT, allocation)
	var req *http.Request
	var err error
	if update {
		req, err = http.NewRequest(http.MethodPut, url, body)
	} else {
		req, err = http.NewRequest(http.MethodPost, url, body)
	}
	return setClientInfo(req, err, client)
}

func NewCommitRequest(baseUrl, allocation string, client ClientConfig, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, COMMIT_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	return setClientInfo(req, err, client)
}

func NewDownloadRequest(baseUrl, allocation string, client ClientConfig, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, DOWNLOAD_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	return setClientInfo(req, err, client)
}

func NewListRequest(baseUrl, allocation string, client ClientConfig, path string) (*http.Request, error) {
	nurl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	nurl.Path += LIST_ENDPOINT + allocation
	params := url.Values{}
	params.Add("path", path)
	//url := fmt.Sprintf("%s%s%s?path=%s", baseUrl, LIST_ENDPOINT, allocation, path)
	nurl.RawQuery = params.Encode() // Escape Query Parameters
	req, err := http.NewRequest(http.MethodGet, nurl.String(), nil)
	return setClientInfo(req, err, client)
}

func NewStatsRequest(baseUrl, allocation string, client ClientConfig, path string) (*http.Request, error) {
	//url := fmt.Sprintf("%s%s%s?path=%s", baseUrl, FILE_STATS_ENDPOINT, allocation, path)
	//req, err := http.NewRequest(http.MethodGet, url, nil)
	nurl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	nurl.Path += FILE_STATS_ENDPOINT + allocation
	params := url.Values{}
	params.Add("path", path)
	//url := fmt.Sprintf("%s%s%s?path=%s", baseUrl, LIST_ENDPOINT, allocation, path)
	nurl.RawQuery = params.Encode() // Escape Query Parameters
	req, err := http.NewRequest(http.MethodGet, nurl.String(), nil)
	return setClientInfo(req, err, client)
}

func NewLatestReadMarkerRequest(baseUrl string, client ClientConfig) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", baseUrl, LATEST_READ_MARKER)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	return setClientInfo(req, err, client)
}

func NewFileMetaRequest(baseUrl string, allocation string, client ClientConfig, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, FILE_META_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	return setClientInfo(req, err, client)
}

func NewDeleteRequest(baseUrl, allocation string, client ClientConfig, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, UPLOAD_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodDelete, url, body)
	return setClientInfo(req, err, client)
}

func NewWriteMarker() *WriteMarker {
	return &WriteMarker{}
}

func (wm *WriteMarker) GetHash() string {
	sigData := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v", wm.AllocationRoot, wm.PreviousAllocationRoot, wm.AllocationID, wm.BlobberID, wm.ClientID, wm.Size, wm.Timestamp)
	return encryption.Hash(sigData)
}

func (wm *WriteMarker) Sign(privateKey string) error {
	var err error
	wm.Signature, err = encryption.Sign(privateKey, wm.GetHash())
	return err
}

func NewReadMarker() *ReadMarker {
	return &ReadMarker{}
}

func (rm *ReadMarker) GetHash() string {
	sigData := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v", rm.AllocationID, rm.BlobberID, rm.ClientID, rm.ClientPublicKey, rm.OwnerID, rm.ReadCounter, rm.Timestamp)
	return encryption.Hash(sigData)
}

func (rm *ReadMarker) Sign(privateKey string) error {
	var err error
	rm.Signature, err = encryption.Sign(privateKey, rm.GetHash())
	return err
}

func NewDeleteToken() *DeleteToken {
	return &DeleteToken{}
}

func (dt *DeleteToken) GetHash() string {
	sigData := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v", dt.FileRefHash, dt.FilePathHash, dt.AllocationID, dt.BlobberID, dt.ClientID, dt.Size, dt.Timestamp)
	return encryption.Hash(sigData)
}

func (dt *DeleteToken) Sign(privateKey string) error {
	var err error
	dt.Signature, err = encryption.Sign(privateKey, dt.GetHash())
	return err
}
