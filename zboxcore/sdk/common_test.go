package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/sdk/mocks"
	"github.com/stretchr/testify/assert"

	"gopkg.in/yaml.v2"
)

const (
	configDir            = "testdata"
	syncTestDir          = configDir + "/" + "sync"
	syncDir              = syncTestDir + "/" + "sync_alloc"
	textPlainContentType = "text/plain"
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

func setupMockInitStorageSDK(t *testing.T, configDir string, blobberNums int) (miners, sharders []string, blobbers []*mocks.Blobber, close func()) {
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

	// setup mock miner, sharder and blobber http server
	miner, closeMinerServer := mocks.NewMinerHTTPServer(t)
	sharder, closeSharderServer := mocks.NewSharderHTTPServer(t)
	var closeBlockWorkerServer func()
	blockWorker, closeBlockWorkerServer = mocks.NewBlockWorkerHTTPServer(t, []string{miner}, []string{sharder})
	blobbers = make([]*mocks.Blobber, blobberNums)
	for i := 0; i < blobberNums; i++ {
		blobber := mocks.NewBlobberHTTPServer(t)
		blobbers[i] = blobber
	}

	err = InitStorageSDK(clientConfig, blockWorker, chainID, signScheme, preferredBlobbers)
	assert.NoErrorf(t, err, "Error InitStorageSDK(): %v", err)
	return []string{miner}, []string{sharder}, blobbers, func() {
		closeBlockWorkerServer()
		closeMinerServer()
		closeSharderServer()
		for _, bl := range blobbers {
			bl.Close()
		}
	}
}

type writeFile struct {
	FileName string
	IsDir    bool
	Content  []byte
	Child    []*writeFile
}

var (
	downloadSuccessFileChan chan []*writeFile
	commitResultChan        chan *CommitResult
)

func willDownloadSuccessFiles(wf ...*writeFile) {
	downloadSuccessFileChan <- wf
}

func deleteFiles(wfs ...*writeFile) {
	for _, wf := range wfs {
		os.Remove(wf.FileName)
	}
}

func willReturnCommitResult(c *CommitResult) {
	commitResultChan <- c
}

func writeFiles(wfs ...*writeFile) {
	for _, wf := range wfs {
		var err error
		if wf.IsDir {
			err = os.MkdirAll(wf.FileName, 0755)
			if len(wf.Child) > 0 {
				writeFiles(wf.Child...)
			}
		} else {
			err = ioutil.WriteFile(wf.FileName, wf.Content, 0644)
		}
		if err != nil {
			break
		}
	}
}

func setupMockAllocation(t *testing.T, dirPath string, blobberMocks []*mocks.Blobber) (allocation *Allocation, cncl func()) {
	blobbers := []*blockchain.StorageNode{}
	if blobberMocks != nil {
		for _, blobberMock := range blobberMocks {
			if blobberMock != nil {
				blobbers = append(blobbers, &blockchain.StorageNode{
					ID:      blobberMock.ID,
					Baseurl: blobberMock.URL,
				})
			}
		}
	}
	contentBytes := parseFileContent(t, dirPath+"/"+"allocation.json", nil)
	for idx, blobber := range blobbers {
		contentBytes = []byte(strings.ReplaceAll(string(contentBytes), blobberIDMask(idx), blobber.ID))
		contentBytes = []byte(strings.ReplaceAll(string(contentBytes), blobberURLMask(idx), blobber.Baseurl))
	}
	err := json.Unmarshal(contentBytes, &allocation)
	assert.NoErrorf(t, err, "Error json.Unmarshal() cannot parse file content to %T object: %v", allocation, err)
	allocation.Blobbers = blobbers // inject mock blobbers
	allocation.uploadChan = make(chan *UploadRequest, 10)
	allocation.downloadChan = make(chan *DownloadRequest, 10)
	allocation.repairChan = make(chan *RepairRequest, 1)
	allocation.ctx, allocation.ctxCancelF = context.WithCancel(context.Background())
	allocation.uploadProgressMap = make(map[string]*UploadRequest)
	allocation.downloadProgressMap = make(map[string]*DownloadRequest)
	allocation.mutex = &sync.Mutex{}

	// init mock test commit worker
	commitChan = make(map[string]chan *CommitRequest)
	commitResultChan = make(chan *CommitResult)

	var commitResult *CommitResult
	for _, blobber := range blobbers {
		if _, ok := commitChan[blobber.ID]; !ok {
			commitChan[blobber.ID] = make(chan *CommitRequest, 1)
			blobberChan := commitChan[blobber.ID]
			go func(c <-chan *CommitRequest, blID string) {
				for true {
					cm := <-c
					if cm != nil {
						cm.result = commitResult
						if cm.wg != nil {
							cm.wg.Done()
						}
					}
				}
			}(blobberChan, blobber.ID)
		}
	}

	downloadSuccessFileChan = make(chan []*writeFile, 5)
	var downloadWriteFiles []*writeFile

	// init mock test dispatcher, commit result, download success
	go func() {
		for true {
			select {
			case <-allocation.ctx.Done():
				t.Log("Upload cancelled by the parent")
				return
			case commitResult = <-commitResultChan:
			case wfs := <-downloadSuccessFileChan:
				downloadWriteFiles = wfs
			case uploadReq := <-allocation.uploadChan:
				if uploadReq.completedCallback != nil {
					uploadReq.completedCallback(uploadReq.filepath)
				}
				if uploadReq.statusCallback != nil {
					uploadReq.statusCallback.Completed(allocation.ID, uploadReq.filepath, uploadReq.filemeta.Name, uploadReq.filemeta.MimeType, int(uploadReq.filemeta.Size), OpUpload)
				}
				if uploadReq.wg != nil {
					uploadReq.wg.Done()
				}
				t.Logf("received a upload request for %v %v\n", uploadReq.filepath, uploadReq.remotefilepath)
			case downloadReq := <-allocation.downloadChan:
				if len(downloadWriteFiles) > 0 {
					writeFiles(downloadWriteFiles...)
				}
				if downloadReq.completedCallback != nil {
					downloadReq.completedCallback(downloadReq.remotefilepath, downloadReq.remotefilepathhash)
				}
				if downloadReq.statusCallback != nil {
					downloadReq.statusCallback.Completed(allocation.ID, downloadReq.localpath, "1.txt", "application/octet-stream", 3, OpDownload)
				}
				if downloadReq.wg != nil {
					downloadReq.wg.Done()
				}
				t.Logf("received a download request for %v\n", downloadReq.remotefilepath)
			case repairReq := <-allocation.repairChan:
				if repairReq.completedCallback != nil {
					repairReq.completedCallback()
				}
				if repairReq.wg != nil {
					repairReq.wg.Done()
				}
				t.Logf("received a repair request for %v\n", repairReq.listDir.Path)
			}
		}
	}()
	allocation.initialized = true
	return allocation, func() {}
}

type httpMockResponseDefinition struct {
	StatusCode  int         `json:"status"`
	Body        interface{} `json:"body"`
	ContentType string      `json:"content_type,omitempty"`
}

type httpMockDefinition struct {
	Method    string                          `json:"method"`
	Path      string                          `json:"path"`
	Params    []map[string]string             `json:"params"`
	Responses [][]*httpMockResponseDefinition `json:"responses"`
}

func responseParamTypeCheck(param map[string]string, r *http.Request) bool {
	for key, val := range param {
		if r.URL.Query().Get(key) != val {
			return false
		}
	}
	return true
}

func responseFormBodyTypeCheck(param map[string]string, r *http.Request) bool {
	for key, val := range param {
		r.ParseForm()
		if r.FormValue(key) != val {
			return false
		}
	}
	return true
}

//func responseBodyTypeCheck(param map[string]interface{}, r *http.Request) bool {
//	var bodyReq map[string]interface{}
//	json.NewDecoder(r.Body).Decode(&bodyReq)
//	return reflect.DeepEqual(body, bodyReq)
//}

func blobberMockMaskReplacing(blobbers []*mocks.Blobber) func(input string) (output string) {
	return func(input string) (output string) {
		for replacingIdx, replacingBlobber := range blobbers {
			input = strings.ReplaceAll(input, blobberIDMask(replacingIdx+1), replacingBlobber.ID)
			input = strings.ReplaceAll(input, blobberURLMask(replacingIdx+1), replacingBlobber.URL)
		}
		output = input
		return
	}
}

func mockResponseParser(t *testing.T, indx int, mapHttpMock map[string]*httpMockDefinition, replacingResponseFn func(respInput string) (respOutput string), checks ...func(params map[string]string, r *http.Request) bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if httpMock := mapHttpMock[r.Method+" "+r.URL.Path]; httpMock != nil {
			for paramIdx, param := range httpMock.Params {
				var matchesParam = true
				for _, check := range checks {
					if !check(param, r) {
						matchesParam = false
						break
					}
				}

				if matchesParam {
					if httpMock.Responses[paramIdx][indx].ContentType == textPlainContentType {
						w.WriteHeader(httpMock.Responses[paramIdx][indx].StatusCode)
						body := fmt.Sprintf("%v", httpMock.Responses[paramIdx][indx].Body)
						if body == "" {
							w.Write([]byte("."))
							return
						}
						w.Write([]byte(body))
						return
					}
					respBytes, err := json.Marshal(httpMock.Responses[paramIdx][indx].Body)
					assert.NoErrorf(t, err, "Error json.Marshal() cannot marshal blobber's response: %v", err)
					respStr := string(respBytes)
					if replacingResponseFn != nil {
						respStr = replacingResponseFn(respStr)
					}

					w.WriteHeader(httpMock.Responses[paramIdx][indx].StatusCode)
					w.Write([]byte(respStr))
					return
				}
			}
		}

		t.Logf("Warning response is not initialized for %v", r.URL.String())
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error."))
		return
	}
}

func setupBlobberMockResponses(t *testing.T, blobbers []*mocks.Blobber, dirPath, testCaseName string, checks ...func(params map[string]string, r *http.Request) bool) {
	var blobberHTTPMocks []*httpMockDefinition
	parseFileContent(t, fmt.Sprintf("%v/blobbers_response__%v.json", dirPath, testCaseName), &blobberHTTPMocks)
	var mapBlobberHTTPMocks = make(map[string]*httpMockDefinition, len(blobberHTTPMocks))
	for _, blobberHTTPMock := range blobberHTTPMocks {
		mapBlobberHTTPMocks[blobberHTTPMock.Method+" "+blobberHTTPMock.Path] = blobberHTTPMock
	}

	for idx, blobber := range blobbers {
		for _, blobberMock := range blobberHTTPMocks {
			blobber.SetHandler(t, blobberMock.Path, mockResponseParser(t, idx, mapBlobberHTTPMocks, blobberMockMaskReplacing(blobbers), checks...))
		}
	}
}

func setupMinerMockResponses(t *testing.T, miners []string, dirPath, testCaseName string, checks ...func(params map[string]string, r *http.Request) bool) {
	var minerHTTPMocks []*httpMockDefinition
	parseFileContent(t, fmt.Sprintf("%v/miners_response__%v.json", dirPath, testCaseName), &minerHTTPMocks)
	var mapMinerHTTPMocks = make(map[string]*httpMockDefinition, len(minerHTTPMocks))
	for _, minerHTTPMock := range minerHTTPMocks {
		mapMinerHTTPMocks[minerHTTPMock.Method+" "+minerHTTPMock.Path] = minerHTTPMock
	}

	for idx, _ := range miners {
		for _, minerMock := range minerHTTPMocks {
			mocks.SetMinerHandler(t, minerMock.Path, mockResponseParser(t, idx, mapMinerHTTPMocks, nil, checks...))
		}
	}
}

func setupSharderMockResponses(t *testing.T, sharders []string, dirPath, testCaseName string, checks ...func(params map[string]string, r *http.Request) bool) {
	var sharderHTTPMocks []*httpMockDefinition
	parseFileContent(t, fmt.Sprintf("%v/sharders_response__%v.json", dirPath, testCaseName), &sharderHTTPMocks)
	var mapSharderHTTPMocks = make(map[string]*httpMockDefinition, len(sharderHTTPMocks))
	for _, sharderHTTPMock := range sharderHTTPMocks {
		mapSharderHTTPMocks[sharderHTTPMock.Method+" "+sharderHTTPMock.Path] = sharderHTTPMock
	}

	for idx, _ := range sharders {
		for _, sharderMock := range sharderHTTPMocks {
			mocks.SetSharderHandler(t, sharderMock.Path, mockResponseParser(t, idx, mapSharderHTTPMocks, nil, checks...))
		}
	}
}

func setupExpectedResult(t *testing.T, syncTestDir, testCaseName string) []FileDiff {
	var expectedResult []FileDiff
	parseFileContent(t, fmt.Sprintf("%v/expected_result__%v.json", syncTestDir, testCaseName), &expectedResult)
	return expectedResult
}
