package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/pathutil"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/zcncrypto"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupHttpResponses(
	t *testing.T, mockClient *mocks.HttpClient, allocID string,
	refsInput, fileMetaInput []byte, hashes []string,
	numBlobbers, numCorrect int, isUpdate bool) {

	for i := 0; i < numBlobbers; i++ {
		metaBlobberBase := t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i) + zboxutil.FILE_META_ENDPOINT
		refsBlobberBase := t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i) + zboxutil.REFS_ENDPOINT
		uploadBlobberBase := t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i) + zboxutil.UPLOAD_ENDPOINT
		wmBlobberBase := t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i) + zboxutil.WM_LOCK_ENDPOINT
		commitBlobberBase := t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i) + zboxutil.COMMIT_ENDPOINT
		refPathBlobberBase := t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i) + zboxutil.REFERENCE_ENDPOINT
		latestMarkerBlobberBase := t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i) + zboxutil.LATEST_WRITE_MARKER_ENDPOINT
		rollbackBlobberBase := t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i) + zboxutil.ROLLBACK_ENDPOINT

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == "POST" &&
				strings.Contains(req.URL.String(), metaBlobberBase)
		})).Return(&http.Response{
			StatusCode: func() int {
				if i < numCorrect {
					return http.StatusOK
				}
				return http.StatusBadRequest
			}(),
			Body: ioutil.NopCloser(bytes.NewReader(fileMetaInput)),
		}, nil)

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == "GET" &&
				strings.Contains(req.URL.String(), refsBlobberBase)
		})).Return(&http.Response{
			StatusCode: func() int {
				if i < numCorrect {
					return http.StatusOK
				}
				return http.StatusBadRequest
			}(),
			Body: ioutil.NopCloser(bytes.NewReader(refsInput)),
		}, nil)

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if isUpdate {
				return req.Method == "PUT" &&
					strings.Contains(req.URL.String(), uploadBlobberBase)
			}
			return req.Method == "POST" &&
				strings.Contains(req.URL.String(), uploadBlobberBase)

		})).Return(&http.Response{
			StatusCode: func() int {
				if i < numCorrect {
					return http.StatusOK
				}
				return http.StatusBadRequest
			}(),
			Body: func() io.ReadCloser {
				hash := hashes[i]
				r := UploadResult{
					Filename: "1.txt",
					Hash:     hash,
				}
				b, _ := json.Marshal(r)
				return io.NopCloser(bytes.NewReader(b))
			}(),
		}, nil)

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == "POST" &&
				strings.Contains(req.URL.String(), wmBlobberBase)
		})).Return(&http.Response{
			StatusCode: func() int {
				if i < numCorrect {
					return http.StatusOK
				}
				return http.StatusBadRequest
			}(),
			Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"status":2}`))),
		}, nil)

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == "GET" &&
				strings.Contains(req.URL.String(), refPathBlobberBase)
		})).Return(&http.Response{
			StatusCode: func() int {
				if i < numCorrect {
					return http.StatusOK
				}
				return http.StatusBadRequest
			}(),
			Body: func() io.ReadCloser {
				s := `{"meta_data":{"chunk_size":0,"created_at":0,"hash":"","lookup_hash":"","name":"/","num_of_blocks":0,"path":"/","path_hash":"","size":0,"type":"d","updated_at":0},"Ref":{"ID":0,"Type":"d","AllocationID":"` + allocID + `","LookupHash":"","Name":"/","Path":"/","Hash":"","NumBlocks":0,"PathHash":"","ParentPath":"","PathLevel":1,"CustomMeta":"","ContentHash":"","Size":0,"MerkleRoot":"","ActualFileSize":0,"ActualFileHash":"","MimeType":"","WriteMarker":"","ThumbnailSize":0,"ThumbnailHash":"","ActualThumbnailSize":0,"ActualThumbnailHash":"","EncryptedKey":"","Children":null,"OnCloud":false,"CreatedAt":0,"UpdatedAt":0,"ChunkSize":0},"list":[{"meta_data":{"chunk_size":0,"created_at":0,"hash":"","lookup_hash":"","name":"1.txt","num_of_blocks":0,"path":"/1.txt","path_hash":"","size":0,"type":"f","updated_at":0},"Ref":{"ID":0,"Type":"f","AllocationID":"` + allocID + `","LookupHash":"","Name":"1.txt","Path":"/1.txt","Hash":"","NumBlocks":0,"PathHash":"","ParentPath":"/","PathLevel":1,"CustomMeta":"","ContentHash":"","Size":0,"MerkleRoot":"","ActualFileSize":0,"ActualFileHash":"","MimeType":"","WriteMarker":"","ThumbnailSize":0,"ThumbnailHash":"","ActualThumbnailSize":0,"ActualThumbnailHash":"","EncryptedKey":"","Children":null,"OnCloud":false,"CreatedAt":0,"UpdatedAt":0,"ChunkSize":0}}],"latest_write_marker":null}`
				return ioutil.NopCloser(bytes.NewReader([]byte(s)))
			}(),
		}, nil)

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == "GET" &&
				strings.Contains(req.URL.String(), latestMarkerBlobberBase)
		})).Return(&http.Response{
			StatusCode: func() int {
				if i < numCorrect {
					return http.StatusOK
				}
				return http.StatusBadRequest
			}(),
			Body: func() io.ReadCloser {
				s := `{"latest_write_marker":null,"prev_write_marker":null}`
				return ioutil.NopCloser(bytes.NewReader([]byte(s)))
			}(),
		}, nil)

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == "POST" &&
				strings.Contains(req.URL.String(), commitBlobberBase)
		})).Return(&http.Response{
			StatusCode: func() int {
				if i < numCorrect {
					return http.StatusOK
				}
				return http.StatusBadRequest
			}(),
			Body: ioutil.NopCloser(bytes.NewReader(nil)),
		}, nil)

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == "POST" &&
				strings.Contains(req.URL.String(), rollbackBlobberBase)
		})).Return(&http.Response{
			StatusCode: func() int {
				if i < numCorrect {
					return http.StatusOK
				}
				return http.StatusBadRequest
			}(),
			Body: ioutil.NopCloser(bytes.NewReader(nil)),
		}, nil)

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == "DELETE" &&
				strings.Contains(req.URL.String(), wmBlobberBase)
		})).Return(&http.Response{
			StatusCode: func() int {
				if i < numCorrect {
					return http.StatusOK
				}
				return http.StatusBadRequest
			}(),
		}, nil)
	}
}

func TestAllocation_UpdateFile(t *testing.T) {
	mockClient := mocks.HttpClient{}
	zboxutil.Client = &mockClient

	const mockLocalPath = "1.txt"

	a := &Allocation{
		ID:           "TestAllocation_UpdateFile",
		Tx:           "TestAllocation_UpdateFile",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}
	setupMockAllocation(t, a)

	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: "http://" + t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	teardown := setupMockFile(t, mockLocalPath)
	defer teardown(t)

	refsInput := map[string]interface{}{
		"total_pages": 1,
		"refs": []map[string]interface{}{
			{
				"file_id":          "2",
				"type":             "f",
				"allocation_id":    a.ID,
				"lookup_hash":      "lookup_hash",
				"name":             mockLocalPath,
				"path":             pathutil.Join("/", mockLocalPath),
				"path_hash":        "path_hash",
				"parent_path":      "/",
				"level":            1,
				"size":             65536,
				"actual_file_size": 65536 * int64(len(a.Blobbers)),
				"actual_file_hash": "actual_file_hash",
				"created_at":       common.Timestamp(time.Now().Unix()),
				"updated_at":       common.Timestamp(time.Now().Unix()),
				"id":               3,
			},
		},
	}

	resfsIn, err := json.Marshal(refsInput)
	require.NoError(t, err)

	fileMetaIn := []byte("{\"actual_file_size\":1}")

	hashes := []string{
		"5c84c73878159775992d20425c13bafc8bc10515c40e0365dde068626918fceb",
		"f8d78ca33bd3c532f4d9c56bcd969944b61350c1be64df22f9353f359e3a8ba4",
		"f435a42af309218e88196d4ed2e0c1977a701641b06434be0bb0263099f3faa9",
		"6b3e932bfd2b2c09e39d35e7c4928c42b73bee194045e545560229234d695669",
	}
	setupHttpResponses(t, &mockClient, a.ID, resfsIn, fileMetaIn, hashes, len(a.Blobbers), len(a.Blobbers), true)

	err = a.UpdateFile(os.TempDir(), mockLocalPath, "/", nil)
	require.NoErrorf(t, err, "Unexpected error %v", err)
}

func TestAllocation_UploadFile(t *testing.T) {
	mockClient := mocks.HttpClient{}
	zboxutil.Client = &mockClient

	const mockLocalPath = "1.txt"
	require := require.New(t)
	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_UploadFile",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}

	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: "http://" + t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i),
		})
	}
	hashes := []string{
		"5c84c73878159775992d20425c13bafc8bc10515c40e0365dde068626918fceb",
		"f8d78ca33bd3c532f4d9c56bcd969944b61350c1be64df22f9353f359e3a8ba4",
		"f435a42af309218e88196d4ed2e0c1977a701641b06434be0bb0263099f3faa9",
		"6b3e932bfd2b2c09e39d35e7c4928c42b73bee194045e545560229234d695669",
	}
	setupHttpResponses(t, &mockClient, a.ID, nil, nil, hashes, len(a.Blobbers), len(a.Blobbers), false)

	err := a.UploadFile(os.TempDir(), mockLocalPath, "/", nil)
	require.NoErrorf(err, "Unexpected error %v", err)
}

func TestAllocation_UpdateFileWithThumbnail(t *testing.T) {
	const (
		mockLocalPath     = "1.txt"
		mockThumbnailPath = "thumbnail_alloc"
	)

	mockClient := mocks.HttpClient{}
	zboxutil.Client = &mockClient

	a := &Allocation{
		ID:           "TestAllocation_UpdateFile_WithThumbNail",
		Tx:           "TestAllocation_UpdateFile_WithThumbNail",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}
	setupMockAllocation(t, a)

	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: "http://" + t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	teardown1 := setupMockFile(t, mockLocalPath)
	defer teardown1(t)
	teardown2 := setupMockFile(t, mockThumbnailPath)
	defer teardown2(t)

	refsInput := map[string]interface{}{
		"total_pages": 1,
		"refs": []map[string]interface{}{
			{
				"file_id":          "2",
				"type":             "f",
				"allocation_id":    a.ID,
				"lookup_hash":      "lookup_hash",
				"name":             mockLocalPath,
				"path":             pathutil.Join("/", mockLocalPath),
				"path_hash":        "path_hash",
				"parent_path":      "/",
				"level":            1,
				"size":             65536,
				"actual_file_size": 65536 * int64(len(a.Blobbers)),
				"actual_file_hash": "actual_file_hash",
				"created_at":       common.Timestamp(time.Now().Unix()),
				"updated_at":       common.Timestamp(time.Now().Unix()),
				"id":               3,
			},
		},
	}

	resfsIn, err := json.Marshal(refsInput)
	require.NoError(t, err)

	fileMetaIn := []byte("{\"actual_file_size\":1}")

	hashes := []string{
		"5c84c73878159775992d20425c13bafc8bc10515c40e0365dde068626918fceb",
		"f8d78ca33bd3c532f4d9c56bcd969944b61350c1be64df22f9353f359e3a8ba4",
		"f435a42af309218e88196d4ed2e0c1977a701641b06434be0bb0263099f3faa9",
		"6b3e932bfd2b2c09e39d35e7c4928c42b73bee194045e545560229234d695669",
	}
	setupHttpResponses(t, &mockClient, a.ID, resfsIn, fileMetaIn, hashes, len(a.Blobbers), len(a.Blobbers), true)

	err = a.UpdateFileWithThumbnail(os.TempDir(), mockLocalPath, "/", mockThumbnailPath, nil)
	require.NoErrorf(t, err, "Unexpected error %v", err)
}

func TestAllocation_UploadFileWithThumbnail(t *testing.T) {
	const (
		mockTmpPath       = "/tmp"
		mockLocalPath     = "1.txt"
		mockThumbnailPath = "thumbnail_alloc"
	)

	mockClient := mocks.HttpClient{}
	zboxutil.Client = &mockClient

	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}

	if teardown := setupMockFile(t, mockThumbnailPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_UploadFileWithThumbnail",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}

	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: "http://" + t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i),
		})
	}
	hashes := []string{
		"5c84c73878159775992d20425c13bafc8bc10515c40e0365dde068626918fceb",
		"f8d78ca33bd3c532f4d9c56bcd969944b61350c1be64df22f9353f359e3a8ba4",
		"f435a42af309218e88196d4ed2e0c1977a701641b06434be0bb0263099f3faa9",
		"6b3e932bfd2b2c09e39d35e7c4928c42b73bee194045e545560229234d695669",
	}
	setupHttpResponses(t, &mockClient, a.ID, nil, nil, hashes, len(a.Blobbers), len(a.Blobbers), false)

	err := a.UploadFileWithThumbnail(mockTmpPath, mockLocalPath, "/", mockThumbnailPath, nil)
	require.NoErrorf(t, err, "Unexpected error %v", err)
}

func TestAllocation_EncryptAndUpdateFile(t *testing.T) {
	mockClient := mocks.HttpClient{}
	zboxutil.Client = &mockClient

	const mockLocalPath = "1.txt"

	a := &Allocation{
		ID:           "TestAllocation_Encrypt_And_UpdateFile",
		Tx:           "TestAllocation_Encrypt_And_UpdateFile",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}
	setupMockAllocation(t, a)

	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: "http://" + t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	teardown := setupMockFile(t, mockLocalPath)
	defer teardown(t)

	refsInput := map[string]interface{}{
		"total_pages": 1,
		"refs": []map[string]interface{}{
			{
				"file_id":          "2",
				"type":             "f",
				"allocation_id":    a.ID,
				"lookup_hash":      "lookup_hash",
				"name":             mockLocalPath,
				"path":             pathutil.Join("/", mockLocalPath),
				"path_hash":        "path_hash",
				"parent_path":      "/",
				"level":            1,
				"size":             65536,
				"actual_file_size": 65536 * int64(len(a.Blobbers)),
				"actual_file_hash": "actual_file_hash",
				"created_at":       common.Timestamp(time.Now().Unix()),
				"updated_at":       common.Timestamp(time.Now().Unix()),
				"id":               3,
			},
		},
	}

	resfsIn, err := json.Marshal(refsInput)
	require.NoError(t, err)

	fileMetaIn := []byte("{\"actual_file_size\":1}")
	hashes := []string{
		"a9ad93057a092ebeeab2e34f16cd6c1135d08b5a165708d072e6d2da75b47e81",
		"bf116d80708522b6e006e818c05e1de4d6197e5882f17cd806702c4396100176",
		"3c4f6a43748f6b7cefee11216540414cb9b2563c294a5f7d633c2e9cda26f7bc",
		"249684daaeef1a8d38d0be0ea38777886e0b3ddf3deaef2eabe4117cc6e67256",
	}
	setupHttpResponses(t, &mockClient, a.ID, resfsIn, fileMetaIn, hashes, len(a.Blobbers), len(a.Blobbers), true)

	err = a.EncryptAndUpdateFile(os.TempDir(), mockLocalPath, "/", nil)
	require.NoError(t, err)
}

func TestAllocation_EncryptAndUploadFile(t *testing.T) {
	mockClient := mocks.HttpClient{}
	zboxutil.Client = &mockClient

	const (
		mockLocalPath = "1.txt"
		mockTmpPath   = "/tmp"
	)

	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}
	a := &Allocation{
		Tx:           "TestAllocation_EncryptAndUploadFile",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}

	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: "http://" + t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	hashes := []string{
		"5c84c73878159775992d20425c13bafc8bc10515c40e0365dde068626918fceb",
		"f8d78ca33bd3c532f4d9c56bcd969944b61350c1be64df22f9353f359e3a8ba4",
		"f435a42af309218e88196d4ed2e0c1977a701641b06434be0bb0263099f3faa9",
		"6b3e932bfd2b2c09e39d35e7c4928c42b73bee194045e545560229234d695669",
	}
	setupHttpResponses(t, &mockClient, a.ID, nil, nil, hashes, len(a.Blobbers), len(a.Blobbers), false)

	err := a.EncryptAndUploadFile(mockTmpPath, mockLocalPath, "/", nil)
	require.NoError(t, err)
}

func TestAllocation_EncryptAndUpdateFileWithThumbnail(t *testing.T) {
	mockClient := mocks.HttpClient{}
	zboxutil.Client = &mockClient

	const (
		mockLocalPath     = "1.txt"
		mockThumbnailPath = "thumbnail_alloc"
		mockTmpPath       = "/tmp"
	)

	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}

	if teardown := setupMockFile(t, mockThumbnailPath); teardown != nil {
		defer teardown(t)
	}

	a := &Allocation{
		ID:           "TestAllocation_EncryptAndUpdateFileWithThumbnail",
		Tx:           "TestAllocation_EncryptAndUpdateFileWithThumbnail",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
	}

	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: "http://" + t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	refsInput := map[string]interface{}{
		"total_pages": 1,
		"refs": []map[string]interface{}{
			{
				"file_id":          "2",
				"type":             "f",
				"allocation_id":    a.ID,
				"lookup_hash":      "lookup_hash",
				"name":             mockLocalPath,
				"path":             pathutil.Join("/", mockLocalPath),
				"path_hash":        "path_hash",
				"parent_path":      "/",
				"level":            1,
				"size":             65536,
				"actual_file_size": 65536 * int64(len(a.Blobbers)),
				"actual_file_hash": "actual_file_hash",
				"created_at":       common.Timestamp(time.Now().Unix()),
				"updated_at":       common.Timestamp(time.Now().Unix()),
				"id":               3,
			},
		},
	}

	resfsIn, err := json.Marshal(refsInput)
	require.NoError(t, err)

	fileMetaIn := []byte("{\"actual_file_size\":1}")

	hashes := []string{
		"a9ad93057a092ebeeab2e34f16cd6c1135d08b5a165708d072e6d2da75b47e81",
		"bf116d80708522b6e006e818c05e1de4d6197e5882f17cd806702c4396100176",
		"3c4f6a43748f6b7cefee11216540414cb9b2563c294a5f7d633c2e9cda26f7bc",
		"249684daaeef1a8d38d0be0ea38777886e0b3ddf3deaef2eabe4117cc6e67256",
	}
	setupHttpResponses(t, &mockClient, a.ID, resfsIn, fileMetaIn, hashes, len(a.Blobbers), len(a.Blobbers), true)
	err = a.EncryptAndUpdateFileWithThumbnail(mockTmpPath, mockLocalPath, "/", mockThumbnailPath, nil)

	require.NoError(t, err)
}

func TestAllocation_EncryptAndUploadFileWithThumbnail(t *testing.T) {
	mockClient := mocks.HttpClient{}
	zboxutil.Client = &mockClient

	const (
		mockLocalPath     = "1.txt"
		mockThumbnailPath = "thumbnail_alloc"
		mockTmpPath       = "/tmp"
	)

	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}

	if teardown := setupMockFile(t, mockThumbnailPath); teardown != nil {
		defer teardown(t)
	}

	a := &Allocation{
		Tx:           "TestAllocation_EncryptAndUploadFileWithThumbnail",
		ParityShards: 2,
		DataShards:   2,
		Size:         2 * GB,
		ctx:          context.TODO(),
	}

	setupMockAllocation(t, a)
	for i := 0; i < numBlobbers; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      mockBlobberId + strconv.Itoa(i),
			Baseurl: "http://" + t.Name() + "/" + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	hashes := []string{
		"5c84c73878159775992d20425c13bafc8bc10515c40e0365dde068626918fceb",
		"f8d78ca33bd3c532f4d9c56bcd969944b61350c1be64df22f9353f359e3a8ba4",
		"f435a42af309218e88196d4ed2e0c1977a701641b06434be0bb0263099f3faa9",
		"6b3e932bfd2b2c09e39d35e7c4928c42b73bee194045e545560229234d695669",
	}

	setupHttpResponses(t, &mockClient, a.ID, nil, nil, hashes, len(a.Blobbers), len(a.Blobbers), false)

	err := a.EncryptAndUploadFileWithThumbnail(mockTmpPath, mockLocalPath, "/", mockThumbnailPath, nil)
	require.NoError(t, err)
}

func TestAllocation_RepairFile(t *testing.T) {
	const (
		mockFileRefName = "mock file ref name"
		mockLocalPath   = "1.txt"
		mockActualHash  = "75a919d23622c29ade8096ed1add6606ec970579459178db3a7d1d0ff8df92d3"
		mockChunkHash   = "a6fb1cb61c9a3b8709242de28e44fb0b4de3753995396ae1d21ca9d4e956e9e2"
	)

	rawClient := zboxutil.Client
	createClient := resty.CreateClient

	var mockClient = mocks.HttpClient{}

	zboxutil.Client = &mockClient
	resty.CreateClient = func(t *http.Transport, timeout time.Duration) resty.Client {
		return &mockClient
	}

	defer func() {
		zboxutil.Client = rawClient
		resty.CreateClient = createClient
	}()

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	setupHttpResponses := func(t *testing.T, testName string, numBlobbers, numCorrect int) {
		require.True(t, numBlobbers >= numCorrect)
		for i := 0; i < numBlobbers; i++ {
			var hash string
			if i < numCorrect {
				hash = mockActualHash
			}
			frName := mockFileRefName + strconv.Itoa(i)
			url := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/file/meta"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), url)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileRefName, hash string) io.ReadCloser {
					jsonFR, err := json.Marshal(&fileref.FileRef{
						ActualFileHash: hash,
						Ref: fileref.Ref{
							Name: fileRefName,
						},
					})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(frName, hash),
			}, nil)
		}
	}

	setupHttpResponsesWithUpload := func(t *testing.T, testName string, numBlobbers, numCorrect int) {
		require.True(t, numBlobbers >= numCorrect)
		for i := 0; i < numBlobbers; i++ {
			var hash string
			if i < numCorrect {
				hash = mockActualHash
			}

			frName := mockFileRefName + strconv.Itoa(i)
			httpResponse := &http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileRefName, hash string) io.ReadCloser {
					jsonFR, err := json.Marshal(&fileref.FileRef{
						ActualFileHash: hash,
						ActualFileSize: 14,
						Ref: fileref.Ref{
							Name:         fileRefName,
							FileMetaHash: hash,
						},
					})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(frName, hash),
			}

			urlMeta := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/file/meta"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlMeta)
			})).Return(httpResponse, nil)

			urlUpload := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/file/upload"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlUpload)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileRefName, hash string) io.ReadCloser {
					jsonFR, err := json.Marshal(&UploadResult{
						Filename: mockLocalPath,
						Hash:     mockChunkHash,
					})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(frName, hash),
			}, nil)

			urlLatestWritemarker := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/file/latestwritemarker"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlLatestWritemarker)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func() io.ReadCloser {
					s := `{"latest_write_marker":null,"prev_write_marker":null}`
					return ioutil.NopCloser(bytes.NewReader([]byte(s)))
				}(),
			}, nil)

			urlRollback := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/connection/rollback"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlRollback)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil)

			urlFilePath := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/file/referencepath"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlFilePath)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileRefName, hash string) io.ReadCloser {
					jsonFR, err := json.Marshal(&ReferencePathResult{
						ReferencePath: &fileref.ReferencePath{
							Meta: map[string]interface{}{
								"type": "d",
							},
						},
						LatestWM: nil,
					})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(frName, hash),
			}, nil)

			urlCommit := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + "/v1/connection/commit"
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlCommit)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func(fileRefName, hash string) io.ReadCloser {
					jsonFR, err := json.Marshal(&ReferencePathResult{})
					require.NoError(t, err)
					return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
				}(frName, hash),
			}, nil)

			urlLock := "http://TestAllocation_RepairFile" + testName + mockBlobberUrl + strconv.Itoa(i) + zboxutil.WM_LOCK_ENDPOINT
			urlLock = strings.TrimRight(urlLock, "/")
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return strings.HasPrefix(req.URL.String(), urlLock)
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: func() io.ReadCloser {
					resp := &WMLockResult{
						Status: WMLockStatusOK,
					}
					respBuf, _ := json.Marshal(resp)
					return ioutil.NopCloser(bytes.NewReader(respBuf))
				}(),
			}, nil)
		}
	}

	type parameters struct {
		localPath  string
		remotePath string
		status     StatusCallback
	}
	tests := []struct {
		name        string
		parameters  parameters
		numBlobbers int
		numCorrect  int
		setup       func(*testing.T, string, int, int)
		wantErr     bool
		wantRepair  bool
		errMsg      string
	}{
		{
			name: "Test_Repair_Not_Required_Failed",
			parameters: parameters{
				localPath:  mockLocalPath,
				remotePath: "/",
			},
			numBlobbers: 4,
			numCorrect:  4,
			setup:       setupHttpResponses,
			wantRepair:  false,
		},
		{
			name: "Test_Repair_Required_Success",
			parameters: parameters{
				localPath:  mockLocalPath,
				remotePath: "/",
			},
			numBlobbers: 6,
			numCorrect:  5,
			setup:       setupHttpResponsesWithUpload,
			wantRepair:  true,
		},
	}

	if teardown := setupMockFile(t, mockLocalPath); teardown != nil {
		defer teardown(t)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			a := &Allocation{
				ParityShards: tt.numBlobbers / 2,
				DataShards:   tt.numBlobbers / 2,
				Size:         2 * GB,
			}
			a.downloadChan = make(chan *DownloadRequest, 10)
			a.repairChan = make(chan *RepairRequest, 1)
			a.ctx, a.ctxCancelF = context.WithCancel(context.Background())
			a.downloadProgressMap = make(map[string]*DownloadRequest)
			a.mutex = &sync.Mutex{}
			a.initialized = true
			sdkInitialized = true
			setupMockAllocation(t, a)
			for i := 0; i < tt.numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      mockBlobberId + strconv.Itoa(i),
					Baseurl: "http://TestAllocation_RepairFile" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			tt.setup(t, tt.name, tt.numBlobbers, tt.numCorrect)
			found, _, isRequired, ref, err := a.RepairRequired(tt.parameters.remotePath)
			require.Nil(err)
			require.Equal(tt.wantRepair, isRequired)
			if !tt.wantRepair {
				return
			}
			f, err := os.Open(tt.parameters.localPath)
			require.Nil(err)
			sz, err := f.Stat()
			require.Nil(err)
			require.NotNil(sz)
			ref.ActualSize = sz.Size()
			err = a.RepairFile(f, tt.parameters.remotePath, tt.parameters.status, found, ref)
			if tt.wantErr {
				require.NotNil(err)
			} else {
				require.Nil(err)
			}

			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "Unexpected error %v", err)
		})
	}
}
