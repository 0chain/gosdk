package sdk

import (
	"encoding/json"
	"fmt"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/sdk/mock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

const (
	configDir   = "test"
	syncTestDir = configDir + "/" + "sync"
	syncDir     = syncTestDir + "/" + "sync_alloc"
)

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

func SetupMockInitStorageSDK(t *testing.T, configDir string, minerHTTPMockURLs, sharderHTTPMockURLs, blobberHTTPMockURLs []string) {
	nodeConfig := viper.New()
	nodeConfig.SetConfigFile(configDir + "/" + "config.yaml")

	err := nodeConfig.ReadInConfig()
	assert.NoErrorf(t, err, "Error nodeConfig.ReadInConfig(): %v", err)

	clientBytes := parseFileContent(t, configDir+"/"+"wallet.json", nil)
	clientConfig := string(clientBytes)

	blockWorker := nodeConfig.GetString("block_worker")
	preferredBlobbers := nodeConfig.GetStringSlice("preferred_blobbers")
	signScheme := nodeConfig.GetString("signature_scheme")
	chainID := nodeConfig.GetString("chain_id")

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

func SetupMockAllocation(t *testing.T, blobberMocks []*mock.Blobber) *Allocation {
	blobbers := []*blockchain.StorageNode{}
	for _, blobberMock := range blobberMocks {
		blobbers = append(blobbers, &blockchain.StorageNode{
			ID:      blobberMock.ID,
			Baseurl: blobberMock.URL,
		})
	}
	var allocation *Allocation
	parseFileContent(t, syncTestDir+"/"+"allocation.json", &allocation)
	allocation.Blobbers = blobbers // inject mock blobbers
	allocation.InitAllocation()
	return allocation
}

func SetupBlobberMockResponses(t *testing.T, blobbers []*mock.Blobber, allocation, syncTestDir, testCaseName string) {
	var blobberMockPathHashResponses map[string][]interface{}
	parseFileContent(t, fmt.Sprintf("%v/blobbers_response__%v.json", syncTestDir, testCaseName), &blobberMockPathHashResponses)
	for idx, blobber := range blobbers {
		blobber.SetBlobberHandler(t, "/v1/file/list/"+allocation, func(w http.ResponseWriter, r *http.Request) {
			if respMock := blobberMockPathHashResponses[r.URL.Query().Get("path_hash")]; respMock != nil {
				w.WriteHeader(200)
				bytes, err := json.Marshal(respMock[idx])
				assert.NoErrorf(t, err, "Error %v invalid blobber json response: %v", r.URL.RawPath, err)
				w.Write(bytes)
				return
			}
			t.Logf("Warning blobber response is not initialized for %v", r.URL.String())
			w.WriteHeader(500)
			w.Write([]byte("Internal Server Error."))
		})
	}
}

func SetupExpectedResult(t *testing.T, syncTestDir, testCaseName string) []FileDiff {
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
	var closeBlobbers = []func(){}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx, closeIdx := mock.NewBlobberHTTPServer(t)
		blobbers = append(blobbers, blobberIdx)
		closeBlobbers = append(closeBlobbers, closeIdx)
	}

	defer func() {
		for _, f := range closeBlobbers {
			f()
		}
	}()

	// mock init sdk
	SetupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// mock allocation
	a := SetupMockAllocation(t, blobbers)
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
			SetupBlobberMockResponses(t, blobbers, a.ID, syncTestDir+"/"+"GetAllocationDiff", tt.name)
			if tt.additionalMock != nil {
				teardownAdditionalMock := tt.additionalMock(t)
				defer teardownAdditionalMock(t)
			}
			want := SetupExpectedResult(t, syncTestDir+"/"+"GetAllocationDiff", tt.name)
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
	var closeBlobbers = []func(){}
	var blobberNums = 4
	for i := 0; i < blobberNums; i++ {
		blobberIdx, closeIdx := mock.NewBlobberHTTPServer(t)
		blobbers = append(blobbers, blobberIdx)
		closeBlobbers = append(closeBlobbers, closeIdx)
	}

	defer func() {
		for _, f := range closeBlobbers {
			f()
		}
	}()

	// mock init sdk
	SetupMockInitStorageSDK(t, configDir, []string{miner}, []string{sharder}, []string{})
	// mock allocation
	a := SetupMockAllocation(t, blobbers)

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
		name    string
		args    args
		additionalMockLocalFile func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		wantErr bool
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
			SetupBlobberMockResponses(t, blobbers, a.ID, syncTestDir+"/"+"SaveRemoteSnapshot", tt.name)
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
