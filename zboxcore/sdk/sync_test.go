package sdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/sdk/mock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const (
	configDir   = "test"
	syncTestDir = configDir + "/" + "sync"
	syncDir     = syncTestDir + "/" + "sync_alloc"
)

func blobberIDMask(idx int) string {
	return fmt.Sprintf("${blobber_id_%v}", idx)
}

func blobberURLMask(idx int) string {
	return fmt.Sprintf("${blobber_url_%v}", idx)
}

func parseFileContent(t *testing.T, fileName string, jsonUnmarshalerInterface interface{}) (fileContentBytes []byte) {
	fs, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	assert.NoErrorf(t, err, "Error os.OpenFile() %v: %v", fileName, err)

	defer fs.Close()
	bytes, err := ioutil.ReadAll(fs)
	assert.NoErrorf(t, err, "Error ioutil.ReadAll() cannot read file content of %v: %v", fileName, err)
	if jsonUnmarshalerInterface != nil {
		err = json.Unmarshal(bytes, jsonUnmarshalerInterface)
		assert.NoErrorf(t, err, "Error json.Unmarshal() cannot parse file content to %T object: %v", jsonUnmarshalerInterface, err)
	}

	return bytes
}

func writeFileContent(t *testing.T, fileName string, fileContentBytes []byte) {
	fs, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	assert.NoErrorf(t, err, "Error os.OpenFile() %v: %v", fileName, err)
	defer fs.Close()
	_, err = fs.Write(fileContentBytes)
	assert.NoErrorf(t, err, "Error fs.Write() cannot write file content to %v: %v", fileName, err)
}

type nodeConfig struct {
	BlockWorker       string   `yaml:"block_worker"`
	PreferredBlobbers []string `yaml:"preferred_blobbers"`
	SignScheme        string   `yaml:"signature_scheme"`
	ChainID           string   `yaml:"chain_id"`
}

func setupMockInitStorageSDK(t *testing.T, configDir string, minerHTTPMockURLs, sharderHTTPMockURLs, blobberHTTPMockURLs []string) {
	var nodeConfig *nodeConfig

	nodeConfigBytes := parseFileContent(t, configDir+"/"+"config.yaml", nil)
	err := yaml.Unmarshal(nodeConfigBytes, &nodeConfig)
	assert.NoErrorf(t, err, "Error yaml.Unmarshal(): %v", err)

	clientBytes := parseFileContent(t, configDir+"/"+"wallet.json", nil)
	clientConfig := string(clientBytes)

	blockWorker := nodeConfig.BlockWorker
	preferredBlobbers := nodeConfig.PreferredBlobbers
	signScheme := nodeConfig.SignScheme
	chainID := nodeConfig.ChainID

	if minerHTTPMockURLs != nil && len(minerHTTPMockURLs) > 0 && sharderHTTPMockURLs != nil && len(sharderHTTPMockURLs) > 0 {
		var close func()
		blockWorker, close = mock.NewBlockWorkerHTTPServer(t, minerHTTPMockURLs, sharderHTTPMockURLs)
		defer close()
		if blobberHTTPMockURLs != nil {
			preferredBlobbers = blobberHTTPMockURLs
		}
	}

	err = InitStorageSDK(clientConfig, blockWorker, chainID, signScheme, preferredBlobbers)
	assert.NoErrorf(t, err, "Error InitStorageSDK(): %v", err)
}

func setupMockAllocation(t *testing.T, dirPath string, blobberMocks []*mock.Blobber) *Allocation {
	blobbers := []*blockchain.StorageNode{}
	for _, blobberMock := range blobberMocks {
		if blobberMock != nil {
			blobbers = append(blobbers, &blockchain.StorageNode{
				ID:      blobberMock.ID,
				Baseurl: blobberMock.URL,
			})
		}
	}
	var allocation *Allocation
	contentBytes := parseFileContent(t, dirPath+"/"+"allocation.json", nil)
	for idx, blobber := range blobbers {
		contentBytes = []byte(strings.ReplaceAll(string(contentBytes), blobberIDMask(idx), blobber.ID))
		contentBytes = []byte(strings.ReplaceAll(string(contentBytes), blobberURLMask(idx), blobber.Baseurl))
	}
	err := json.Unmarshal(contentBytes, &allocation)
	assert.NoErrorf(t, err, "Error json.Unmarshal() cannot parse file content to %T object: %v", allocation, err)
	allocation.Blobbers = blobbers // inject mock blobbers
	allocation.InitAllocation()
	return allocation
}

type httpMockResponseDefinition struct {
	StatusCode int         `json:"status"`
	Body       interface{} `json:"body"`
}

type httpMockDefinition struct {
	Method    string                     `json:"method"`
	Path      string                     `json:"path"`
	Params    []map[string]string        `json:"params"`
	Responses [][]*httpMockResponseDefinition `json:"responses"`
}

func blobberResponseParamCheck(param map[string]string, r *http.Request) bool {
	for key, val := range param {
		if r.URL.Query().Get(key) != val {
			return false
		}
	}
	return true
}

func blobberResponseFormBodyCheck(param map[string]string, r *http.Request) bool {
	for key, val := range param {
		if r.PostForm.Get(key) != val {
			return false
		}
	}
	return true
}

func setupBlobberMockResponses(t *testing.T, blobbers []*mock.Blobber, dirPath, testCaseName string, checks ...func(params map[string]string, r *http.Request) bool) {
	var blobberHTTPMocks []*httpMockDefinition
	parseFileContent(t, fmt.Sprintf("%v/blobbers_response__%v.json", dirPath, testCaseName), &blobberHTTPMocks)
	for _, blobberHTTPMock := range blobberHTTPMocks {
		blobberMockHandler := func(blobberIdx int) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if strings.ToLower(r.Method) == strings.ToLower(blobberHTTPMock.Method) {
					for paramIdx, param := range blobberHTTPMock.Params {
						var matchesParam = true
						for _, check := range checks {
							if !check(param, r) {
								matchesParam = false
								break
							}
						}

						if matchesParam {
							respBytes, err := json.Marshal(blobberHTTPMock.Responses[paramIdx][blobberIdx].Body)
							assert.NoErrorf(t, err, "Error json.Marshal() cannot marshal blobber's response: %v", err)
							respStr := string(respBytes)
							for replacingIdx, replacingBlobber := range blobbers {
								respStr = strings.ReplaceAll(respStr, blobberIDMask(replacingIdx+1), replacingBlobber.ID)
								respStr = strings.ReplaceAll(respStr, blobberURLMask(replacingIdx+1), replacingBlobber.URL)
							}

							w.WriteHeader(blobberHTTPMock.Responses[paramIdx][blobberIdx].StatusCode)
							w.Write([]byte(respStr))
							return
						}
					}
				}

				t.Logf("Warning blobber response is not initialized for %v", r.URL.String())
				w.WriteHeader(500)
				w.Write([]byte("Internal Server Error."))
				return
			}
		}
		for idx, blobber := range blobbers {
			blobber.SetHandler(t, blobberHTTPMock.Path, blobberMockHandler(idx))
		}
	}
}

func setupExpectedResult(t *testing.T, syncTestDir, testCaseName string) []FileDiff {
	var expectedResult []FileDiff
	parseFileContent(t, fmt.Sprintf("%v/expected_result__%v.json", syncTestDir, testCaseName), &expectedResult)
	return expectedResult
}

func TestAllocation_GetAllocationDiff(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()
	var blobbers = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobber := mock.NewBlobberHTTPServer(t)
		blobbers = append(blobbers, blobber)
	}

	defer func() {
		for _, blobber := range blobbers {
			blobber.Close(t)
		}
	}()

	// mock init sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// mock allocation
	a := setupMockAllocation(t, syncTestDir, blobbers)
	type args struct {
		lastSyncCachePath func(t *testing.T, testcaseName string) string
		localRootPath     string
		localFileFilters  []string
		remoteExcludePath []string
	}
	var getLastSyncCachePath = func(t *testing.T, testCaseName string) string {
		return syncTestDir + "/" + "GetAllocationDiff" + "/" + "localcache__" + testCaseName + ".json"
	}
	var localRootPath = syncDir
	var additionalMockLocalFile = func(fileName string) func(t *testing.T) (teardown func(t *testing.T)) {
		return func(t *testing.T) (teardown func(t *testing.T)) {
			teardown = func(t *testing.T) {}
			fullFileName := syncDir + "/" + fileName
			writeFileContent(t, fullFileName, []byte("abcd1234")) // create additional sync file
			return func(t *testing.T) {
				defer os.Remove(fullFileName)
			}
		}
	}

	var resetBlobberMock = func(t *testing.T) {
		for _, blobber := range blobbers {
			blobber.ResetHandler(t)
		}
	}

	tests := []struct {
		name           string
		args           args
		additionalMock func(t *testing.T) (teardown func(t *testing.T))
		wantErr        bool
	}{
		{
			"Test_Matches_Local_And_Remote_Sync_Files",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			nil,
			false,
		},
		{
			"Test_Local_Delete_File",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			nil,
			false,
		},
		{
			// This test case make sure that the local update should be sync with blobber's storage. This test case modifying content of /3.txt.
			// 1: localcache_Test_Update_File.json and blobber's storage mock response should have same file hash of /3.txt (/3.txt file synced before)
			// 2: when update /3.txt content in local => the current hash of local /3.txt file content is different previous version that is stored in localcache_Test_Update_File.json.
			// 3: when update /3.txt content in local => the current hash of local /3.txt file content is different /3.txt content hash showing in blobber's storage list file api's response.
			"Test_Update_File",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			additionalMockLocalFile("3.txt"),
			false,
		},
		{
			"Test_Download_File",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			nil,
			false,
		},
		{
			"Test_Delete_File",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			nil,
			false,
		},
		{
			"Test_Upload_File",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			additionalMockLocalFile("3.txt"),
			false,
		},
		{
			// this test case make sure the test method ignore the check of local additional file /3.txt which is doesn't existed in blobber's storage
			"Test_Matches_Local_And_Remote_Sync_Files_With_Local_File_Filter",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git", "3.txt"},
				remoteExcludePath: []string{},
			},
			additionalMockLocalFile("3.txt"),
			false,
		},
		{
			// this test cases using the blobber's mock response that file /3.txt is already in blobber's storage, but it's not contained in local
			// this test cases make sure that the test method ignore the check of remote /3.txt path from blobber's response
			"Test_Matches_Local_And_Remote_Sync_Files_With_Remote_Exclude_Path",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{"/3.txt"},
			},
			nil,
			false,
		},
		{
			"Test_Remote_Modified_File",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			nil,
			false,
		},
		{
			"Test_Conflict_File",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			additionalMockLocalFile("3.txt"),
			false,
		},
		{
			"Test_Upload_All_Local_Files",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			nil,
			false,
		},
		{
			"Test_Root_Path_Failed",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath + "/" + "some_failed_path",
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			nil,
			true,
		},
		{
			"Test_Last_Sync_File_Cache_Is_Directory_Failed",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			nil,
			true,
		},
		{
			"Test_Last_Sync_File_Cache_Content_Not_JSON_Format_Failed",
			args{
				lastSyncCachePath: getLastSyncCachePath,
				localRootPath:     localRootPath,
				localFileFilters:  []string{".DS_Store", ".git"},
				remoteExcludePath: []string{},
			},
			nil,
			true,
		},
		//{
		//	"Test_Blobber's_HTTP_Response_Error_Failed",
		//	args{
		//		lastSyncCachePath: getLastSyncCachePath,
		//		localRootPath:     localRootPath,
		//		localFileFilters:  []string{".DS_Store", ".git"},
		//		remoteExcludePath: []string{},
		//	},
		//	nil,
		//	true,
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupBlobberMockResponses(t, blobbers, syncTestDir+"/"+"GetAllocationDiff", tt.name, blobberResponseParamCheck)
			defer resetBlobberMock(t)
			if tt.additionalMock != nil {
				teardownAdditionalMock := tt.additionalMock(t)
				defer teardownAdditionalMock(t)
			}
			want := setupExpectedResult(t, syncTestDir+"/"+"GetAllocationDiff", tt.name)
			got, err := a.GetAllocationDiff(tt.args.lastSyncCachePath(t, tt.name), tt.args.localRootPath, tt.args.localFileFilters, tt.args.remoteExcludePath)
			if tt.wantErr {
				assert.Error(t, err, "expected error != nil")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.EqualValues(t, want, got)
			}
		})
	}
}

func TestAllocation_SaveRemoteSnapshot(t *testing.T) {
	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mock.NewMinerHTTPServer(t)
	defer closeMinerServer()
	sharder, closeSharderServer := mock.NewSharderHTTPServer(t)
	defer closeSharderServer()
	var blobbers = []*mock.Blobber{}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx := mock.NewBlobberHTTPServer(t)
		blobbers = append(blobbers, blobberIdx)
	}

	defer func() {
		for _, blobber := range blobbers {
			blobber.Close(t)
		}
	}()

	// mock init sdk
	setupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// mock allocation
	a := setupMockAllocation(t, syncTestDir, blobbers)

	var additionalMockLocalFile = func(t *testing.T, fullFileName string) (teardown func(t *testing.T)) {
		teardown = func(t *testing.T) {}
		writeFileContent(t, fullFileName, []byte("abcd1234")) // create additional localcache file
		return func(t *testing.T) {
			defer os.Remove(fullFileName)
		}
	}

	type args struct {
		pathToSavePrefix  string
		remoteExcludePath []string
	}

	tests := []struct {
		name                    string
		args                    args
		additionalMockLocalFile func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr                 bool
	}{
		{
			"Test_Save_Last_Sync_File_Cache_Success",
			args{
				pathToSavePrefix:  "",
				remoteExcludePath: []string{},
			},
			nil,
			false,
		},
		{
			"Test_Remove_Existing_File_Success",
			args{
				pathToSavePrefix:  "",
				remoteExcludePath: []string{},
			},
			additionalMockLocalFile,
			false,
		},
		{
			// this test cases using file path to save is an existing directory
			"Test_Invalid_File_Path_To_Save_Failed",
			args{
				pathToSavePrefix:  "",
				remoteExcludePath: []string{},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupBlobberMockResponses(t, blobbers, syncTestDir+"/"+"SaveRemoteSnapshot", tt.name, blobberResponseParamCheck)

			var pathToSave string
			if tt.args.pathToSavePrefix == "" {
				pathToSave = fmt.Sprintf("%v/%v/localcache__%v.json", syncTestDir, "SaveRemoteSnapshot", tt.name)
			} else {
				pathToSave = fmt.Sprintf("%v/%v/%v/localcache__%v.json", syncTestDir, "SaveRemoteSnapshot", tt.args.pathToSavePrefix, tt.name)
			}
			defer os.Remove(pathToSave)
			if tt.additionalMockLocalFile != nil {
				teardownAdditionalMock := tt.additionalMockLocalFile(t, pathToSave)
				defer teardownAdditionalMock(t)
			}
			err := a.SaveRemoteSnapshot(pathToSave, tt.args.remoteExcludePath)
			if tt.wantErr {
				assert.Error(t, err, "expected error != nil")
			} else {
				assert.NoError(t, err, "expected no error")
				expectedFileContentBytes := parseFileContent(t, fmt.Sprintf("%v/%v/expected_result__%v.json", syncTestDir, "SaveRemoteSnapshot", tt.name), nil)
				savedDileContentBytes := parseFileContent(t, pathToSave, nil)
				assert.EqualValues(t, expectedFileContentBytes, savedDileContentBytes)
			}
		})
	}
}
