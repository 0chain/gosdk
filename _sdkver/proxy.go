package main

import (
	"fmt"
	"testing"

	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/zcncore"

	"github.com/0chain/gosdk/bls"
	// "github.com/0chain/gosdk/miracl"
	// "github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/zcncrypto"

	"syscall/js"

	"github.com/0chain/gosdk/zboxcore/sdk"

	/// download.go imports.
	"path/filepath"
	"strconv"
	"sync"

	"go.uber.org/zap"

	/// file_operations.go imports.
	// "fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"

	/// util.go imports
	"gopkg.in/cheggaaa/pb.v1"

	// "0proxy.io/core/common"
	// "0proxy.io/zproxycore/handler"

	/// share.go imports
	"github.com/0chain/gosdk/zboxcore/fileref"

	/// upload.go imports
	"encoding/json"
	/// delete.go imports
	// All already imported or not needed.
)

var verifyPublickey = `041eeb1b4eb9b2456799d8e2a566877e83bc5d76ff38b964bd4b7796f6a6ccae6f1966a4d91d362669fafa3d95526b132a6341e3dfff6447e0e76a07b3a7cfa6e8034574266b382b8e5174477ab8a32a49a57eda74895578031cd2d41fd0aef446046d6e633f5eb68a93013dfac1420bf7a1e1bf7a87476024478e97a1cc115de9`
var signPrivatekey = `18c09c2639d7c8b3f26b273cdbfddf330c4f86c2ac3030a6b9a8533dc0c91f5e`
var data = `TEST`

func TestSSSignAndVerify(t *testing.T) {
	signScheme := zcncrypto.NewSignatureScheme("bls0chain")
	signScheme.SetPrivateKey(signPrivatekey)
	hash := zcncrypto.Sha3Sum256(data)

	fmt.Println("hash", hash)
	fmt.Println("privkey", signScheme.GetPrivateKey())

	var sk bls.SecretKey
	sk.DeserializeHexStr(signScheme.GetPrivateKey())
	pk := sk.GetPublicKey()
	fmt.Println("pubkey", pk.ToString())

	signature, err := signScheme.Sign(hash)

	fmt.Println("signature", signature)

	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	verifyScheme := zcncrypto.NewSignatureScheme("bls0chain")
	verifyScheme.SetPublicKey(verifyPublickey)
	if ok, err := verifyScheme.Verify(signature, hash); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	}
}

// Ported from `code/go/0proxy.io/zproxycore/handler/wallet.go`
// Promise code taken from:
// https://withblue.ink/2020/10/03/go-webassembly-http-requests-and-promises.html
func GetClientEncryptedPublicKey(this js.Value, p []js.Value) interface{} {
	clientJSON := p[0].String()
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			initSDK(clientJSON)
			key, err := sdk.GetClientEncryptedPublicKey()

			if err != nil {
				// fmt.Println("get_public_encryption_key_failed: " + err.Error())
				reject.Invoke(js.ValueOf("get_public_encryption_key_failed: " + err.Error()))
				return
			}

			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(js.ValueOf(key))

			// Resolve the Promise
			resolve.Invoke(response)
		}()

		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

// Ported from `code/go/0proxy.io/zproxycore/zproxy/main.go`
// TODO: should be passing in JSON. Better than a long arg list.
func initializeConfig(this js.Value, p []js.Value) interface{} {
	Configuration.ChainID = p[0].String()
	Configuration.SignatureScheme = p[1].String()
	Configuration.Port = p[2].Int()
	Configuration.BlockWorker = p[3].String()
	Configuration.CleanUpWorkerMinutes = p[4].Int()
	return nil
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/zproxycore/handler/util.go`
//-----------------------------------------------------------------------------

func initStorageSDK(this js.Value, p []js.Value) interface{} {
	clientJSON := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err := initSDK(clientJSON)
			if err != nil {
				reject.Invoke(err.Error())
			}

			resolve.Invoke(true)
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func initSDK(clientJSON string) error {
	if len(Configuration.BlockWorker) == 0 ||
		len(Configuration.ChainID) == 0 ||
		len(Configuration.SignatureScheme) == 0 {
		return NewError("invalid_param", "Configuration is empty")
	}

	err := sdk.InitStorageSDK(clientJSON,
		Configuration.BlockWorker,
		Configuration.ChainID,
		Configuration.SignatureScheme,
		nil)
	if err != nil {
		return err
	}

	return zcncore.Init(clientJSON)
}

func validateClientDetails(allocation, clientJSON string) error {
	if len(allocation) == 0 || len(clientJSON) == 0 {
		return NewError("invalid_param", "Please provide allocation and client_json for the client")
	}
	return nil
}

// StatusBar is to check status of any operation
type StatusBar struct {
	b       *pb.ProgressBar
	wg      *sync.WaitGroup
	success bool
	err     error
}

// Started for statusBar
func (s *StatusBar) Started(allocationID, filePath string, op int, totalBytes int) {
	s.b = pb.StartNew(totalBytes)
	s.b.Set(0)
}

// InProgress for statusBar
func (s *StatusBar) InProgress(allocationID, filePath string, op int, completedBytes int, todo_name_var []byte) {
	s.b.Set(completedBytes)
}

// Completed for statusBar
func (s *StatusBar) Completed(allocationID, filePath string, filename string, mimetype string, size int, op int) {
	if s.b != nil {
		s.b.Finish()
	}
	s.success = true
	defer s.wg.Done()
	fmt.Println("Status completed callback. Type = " + mimetype + ". Name = " + filename)
}

// Error for statusBar
func (s *StatusBar) Error(allocationID string, filePath string, op int, err error) {
	if s.b != nil {
		s.b.Finish()
	}
	s.success = false
	s.err = err
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in statusBar Error", r)
		}
	}()
	PrintError("Error in file operation." + err.Error())
	s.wg.Done()
}

// CommitMetaCompleted when commit meta completes
func (s *StatusBar) CommitMetaCompleted(request, response string, err error) {
}

// RepairCompleted when repair is completed
func (s *StatusBar) RepairCompleted(filesRepaired int) {
}

// PrintError is to print error
func PrintError(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/zproxycore/handler/download.go`
//-----------------------------------------------------------------------------

// Download is to download a file from dStorage
// TODO: this should be a dict-type, like a JSON, instead of a long list.
func Download(this js.Value, p []js.Value) interface{} {
	allocation := p[0].String()
	clientJSON := p[1].String()
	remotePath := p[2].String()
	authTicket := p[3].String()
	numBlocks := p[4].String()
	rx_pay := p[5].String()
	file_name := p[6].String()
	lookuphash := p[7].String()

	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	if len(remotePath) == 0 && len(authTicket) == 0 {
		return js.ValueOf("error: " + NewError("invalid_params", "Please provide remote_path OR auth_ticket to download").Error())
	}

	numBlocksInt, _ := strconv.Atoi(numBlocks)
	if numBlocksInt == 0 {
		numBlocksInt = 10
	}

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err = initSDK(clientJSON)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("sdk_not_initialized", "Unable to initialize gosdk with the given client details").Error()))
				return
			}
			sdk.SetNumBlockDownloads(numBlocksInt)

			var at *sdk.AuthTicket
			downloadUsingAT := false
			if len(authTicket) > 0 {
				downloadUsingAT = true
				at = sdk.InitAuthTicket(authTicket)
			}

			// var localFilePath, fileName string
			var localFilePath string
			wg := &sync.WaitGroup{}
			statusBar := &StatusBar{wg: wg}
			wg.Add(1)
			if downloadUsingAT {
				rxPay, _ := strconv.ParseBool(rx_pay)
				allocationObj, err := sdk.GetAllocationFromAuthTicket(authTicket)
				if err != nil {
					reject.Invoke(js.ValueOf("error: " + NewError("get_allocation_failed", err.Error()).Error()))
					return
				}
				fileName := file_name
				if len(fileName) == 0 {
					fileName, err = at.GetFileName()
					if err != nil {
						reject.Invoke(js.ValueOf("error: " + NewError("get_file_name_failed", err.Error()).Error()))
						return
					}
				}

				// In wasm running in browser, we cannot assume the file system exists.
				// createDirIfNotExists(allocationObj.ID)
				// localFilePath = getPath(allocationObj.ID, fileName)
				// deleletFile(localFilePath)
				if len(lookuphash) == 0 {
					lookuphash, err = at.GetLookupHash()
					if err != nil {
						reject.Invoke(js.ValueOf("error: " + NewError("get_lookuphash_failed", err.Error()).Error()))
						return
					}
				}

				// Logger.Info("Doing file download using authTicket", zap.Any("filename", fileName), zap.Any("allocation", allocationObj.ID), zap.Any("lookuphash", lookuphash))
				fmt.Println("Doing file download using authTicket", zap.Any("filename", fileName), zap.Any("allocation", allocationObj.ID), zap.Any("lookuphash", lookuphash))
				localFilePath = "b.txt"
				err = allocationObj.DownloadFromAuthTicket(localFilePath, authTicket, lookuphash, fileName, rxPay, statusBar)
				if err != nil {
					reject.Invoke(js.ValueOf("error: " + NewError("download_from_auth_ticket_failed", err.Error()).Error()))
					return
				}
			} else {

				// In wasm running in browser, we cannot assume the file system exists.
				// createDirIfNotExists(allocation)
				// fileName = filepath.Base(remotePath)
				// localFilePath = getPath(allocation, fileName)
				// deleletFile(localFilePath)

				allocationObj, err := sdk.GetAllocation(allocation)
				if err != nil {
					reject.Invoke(js.ValueOf("error: " + NewError("get_allocation_failed", err.Error()).Error()))
					return
				}

				// Logger.Info("Doing file download", zap.Any("remotepath", remotePath), zap.Any("allocation", allocation))
				fmt.Println("Doing file download", zap.Any("remotepath", remotePath), zap.Any("allocation", allocation))
				// localFilePath = "asdf"
				fmt.Println("dl debug", remotePath)
				// remotePath += "b.txt"
				localFilePath = "b.txt"
				err = allocationObj.DownloadFile(localFilePath, remotePath, statusBar)
				if err != nil {
					reject.Invoke(js.ValueOf("error: " + NewError("download_file_failed", err.Error()).Error()))
					return
				}
			}
			wg.Wait()
			if !statusBar.success {
				reject.Invoke(js.ValueOf("error: " + statusBar.err.Error()))
				return
			}

			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(js.ValueOf(localFilePath))

			resolve.Invoke(response)
		}()

		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/zproxycore/handler/delete.go`
//-----------------------------------------------------------------------------

// Delete is to delete a file in dStorage
func Delete(this js.Value, p []js.Value) interface{} {
	allocation := p[0].String()
	clientJSON := p[1].String()
	remotePath := p[2].String()

	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	if len(remotePath) == 0 {
		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path for delete").Error())
	}

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err = initSDK(clientJSON)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("sdk_not_initialized", "Unable to initialize gosdk with the given client details").Error()))
				return
			}

			allocationObj, err := sdk.GetAllocation(allocation)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("get_allocation_failed", err.Error()).Error()))
				return
			}

			err = allocationObj.DeleteFile(remotePath)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("delete_object_failed", err.Error()).Error()))
				return
			}

			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(js.ValueOf("Delete done successfully"))

			// Resolve the Promise
			resolve.Invoke(response)
		}()

		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/core/common/errors.go`
//-----------------------------------------------------------------------------

/*Error type for a new application error */
type Error struct {
	Code string `json:"code,omitempty"`
	Msg  string `json:"msg"`
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s: %s", err.Code, err.Msg)
}

/*NewError - create a new error */
func NewError(code string, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

/*InvalidRequest - create error messages that are needed when validating request input */
func InvalidRequest(msg string) error {
	return NewError("invalid_request", fmt.Sprintf("Invalid request (%v)", msg))
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/core/config/config.go`
//-----------------------------------------------------------------------------

/*Config - all the config options passed from the command line*/
type Config struct {
	Port                 int
	ChainID              string
	DeploymentMode       byte
	SignatureScheme      string
	BlockWorker          string
	CleanUpWorkerMinutes int
}

/*Configuration of the system */
var Configuration Config

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/zproxycore/handler/file_operations.go`
//-----------------------------------------------------------------------------

const FilesRepo = "files/"

func writeFile(file multipart.File, filePath string) (string, error) {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	return f.Name(), err
}

// Same as writeFile, but takes a string
func writeFile2(file string, filePath string) (string, error) {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.Write([]byte(file))
	// _, err = io.Copy(f, file)
	return f.Name(), err
}

func deleletFile(filePath string) error {
	return os.RemoveAll(filePath)
}

func readFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func getPath(allocation, fileName string) string {
	return FilesRepo + allocation + "/" + fileName
}

func getPathForStream(allocation, fileName string, start, end int) string {
	return FilesRepo + allocation + "/" + fmt.Sprintf("%d-%d-%s", start, end, fileName)
}

func createDirIfNotExists(allocation string) {
	allocationDir := FilesRepo + allocation
	if _, err := os.Stat(allocationDir); os.IsNotExist(err) {
		os.Mkdir(allocationDir, 0777)
	} else {
		fmt.Println("WARN: error in createDirIfNotExists: ", err.Error())
	}
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/zproxycore/handler/rename.go`
//-----------------------------------------------------------------------------

// Rename is to rename file in dStorage
func Rename(this js.Value, p []js.Value) interface{} {
	allocation := p[0].String()
	clientJSON := p[1].String()
	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	remotePath := p[2].String()
	newName := p[3].String()
	if len(remotePath) == 0 || len(newName) == 0 {
		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path and new_name for rename").Error())
	}

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err = initSDK(clientJSON)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("sdk_not_initialized", "Unable to initialize gosdk with the given client details").Error()))
				return
			}

			allocationObj, err := sdk.GetAllocation(allocation)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("get_allocation_failed", err.Error()).Error()))
				return
			}

			err = allocationObj.RenameObject(remotePath, newName)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("rename_object_failed", err.Error()).Error()))
				return
			}

			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(js.ValueOf("Rename done successfully"))

			// Resolve the Promise
			resolve.Invoke(response)
		}()

		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/zproxycore/handler/rename.go`
//-----------------------------------------------------------------------------

// Copy is to copy a file from remotePath to destPath in dStorage
func Copy(this js.Value, p []js.Value) interface{} {
	allocation := p[0].String()
	clientJSON := p[1].String()
	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	remotePath := p[2].String()
	destPath := p[3].String()
	if len(remotePath) == 0 || len(destPath) == 0 {
		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path and dest_path for copy").Error())
	}

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err = initSDK(clientJSON)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("sdk_not_initialized", "Unable to initialize gosdk with the given client details").Error()))
				return
			}

			allocationObj, err := sdk.GetAllocation(allocation)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("get_allocation_failed", err.Error()).Error()))
				return
			}

			err = allocationObj.CopyObject(remotePath, destPath)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("copy_object_failed", err.Error()).Error()))
				return
			}

			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(js.ValueOf("Copy done successfully"))

			// Resolve the Promise
			resolve.Invoke(response)
		}()

		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/zproxycore/handler/share.go`
//-----------------------------------------------------------------------------

// Share is to share file in dStorage
func Share(this js.Value, p []js.Value) interface{} {
	allocation := p[0].String()
	clientJSON := p[1].String()
	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	remotePath := p[2].String()
	if len(remotePath) == 0 {
		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path for share").Error())
	}

	refereeClientID := p[3].String()
	encryptionpublickey := p[4].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err = initSDK(clientJSON)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("sdk_not_initialized", "Unable to initialize gosdk with the given client details").Error()))
				return
			}

			allocationObj, err := sdk.GetAllocation(allocation)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("get_allocation_failed", err.Error()).Error()))
				return

			}

			refType := fileref.FILE
			statsMap, err := allocationObj.GetFileStats(remotePath)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("get_file_stats_failed", err.Error()).Error()))
				return
			}

			isFile := false
			for _, v := range statsMap {
				if v != nil {
					isFile = true
					break
				}
			}
			if !isFile {
				refType = fileref.DIRECTORY
			}

			var fileName string
			_, fileName = filepath.Split(remotePath)

			at, err := allocationObj.GetAuthTicket(remotePath, fileName, refType, refereeClientID, encryptionpublickey)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("get_auth_ticket_failed", err.Error()).Error()))
				return
			}

			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(js.ValueOf(at))

			// var response struct {
			// 	AuthTicket string `json:"auth_ticket"`
			// }
			// response.AuthTicket = at
			resolve.Invoke(response)
		}()

		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/zproxycore/handler/move.go`
//-----------------------------------------------------------------------------

// Move is to move file in dStorage
func Move(this js.Value, p []js.Value) interface{} {
	allocation := p[0].String()
	clientJSON := p[1].String()
	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	remotePath := p[2].String()
	destPath := p[3].String()
	if len(remotePath) == 0 || len(destPath) == 0 {
		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path and dest_path for move").Error())
	}

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err = initSDK(clientJSON)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("sdk_not_initialized", "Unable to initialize gosdk with the given client details").Error()))
				return
			}

			allocationObj, err := sdk.GetAllocation(allocation)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("get_allocation_failed", err.Error()).Error()))
				return
			}

			err = allocationObj.MoveObject(remotePath, destPath)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("move_object_failed", err.Error()).Error()))
				return
			}

			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(js.ValueOf("Move done successfully"))

			// resp := response{
			// 	Message: "Move done successfully",
			// }
			resolve.Invoke(response)
		}()

		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/zproxycore/handler/upload.go`
//-----------------------------------------------------------------------------

// Upload is to upload file to dStorage
func Upload(this js.Value, p []js.Value) interface{} {
	method := p[0].String() // POST or PUT
	allocation := p[1].String()
	clientJSON := p[2].String()
	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	remotePath := p[3].String()
	if len(remotePath) == 0 {
		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path for upload").Error())
	}

	Filename := p[4].String()
	file := p[5].String()
	// file, fileHeader, err := r.FormFile("file")
	// if err != nil {
	// 	js.ValueOf("error: " + NewError("invalid_params", "Unable to get file for upload :"+err.Error()).Error())
	// }
	// defer file.Close()
	encrypt := p[6].String()

	fileAttrs := p[7].String()
	var attrs fileref.Attributes
	if len(fileAttrs) > 0 {
		err := json.Unmarshal([]byte(fileAttrs), &attrs)
		if err != nil {
			return js.ValueOf("error: " + NewError("failed_to_parse_file_attrs", err.Error()).Error())
		}
	}

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			createDirIfNotExists(allocation)

			// localFilePath, err := writeFile(file, getPath(allocation, fileHeader.Filename))
			localFilePath, err := writeFile2(file, getPath(allocation, Filename))
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("write_local_temp_file_failed", err.Error()).Error()))
				return
			}
			defer deleletFile(localFilePath)

			err = initSDK(clientJSON)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("sdk_not_initialized", "Unable to initialize gosdk with the given client details").Error()))
				return
			}

			allocationObj, err := sdk.GetAllocation(allocation)
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("get_allocation_failed", err.Error()).Error()))
				return
			}

			wg := &sync.WaitGroup{}
			statusBar := &StatusBar{wg: wg}
			wg.Add(1)
			if method == "POST" {
				encryptBool, _ := strconv.ParseBool(encrypt)
				if encryptBool {
					// Logger.Info("Doing encrypted file upload with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
					fmt.Println("Doing encrypted file upload with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
					err = allocationObj.EncryptAndUploadFile(localFilePath, remotePath, attrs, statusBar)
				} else {
					// Logger.Info("Doing file upload with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
					fmt.Println("Doing file upload with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
					err = allocationObj.UploadFile(localFilePath, remotePath, attrs, statusBar)
				}
			} else {
				// Logger.Info("Doing file update with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
				fmt.Println("Doing file update with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
				err = allocationObj.UpdateFile(localFilePath, remotePath, attrs, statusBar)
			}
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("upload_file_failed", err.Error()).Error()))
				return
			}

			wg.Wait()
			if !statusBar.success {
				reject.Invoke(js.ValueOf("error: " + statusBar.err.Error()))
				return
			}

			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(js.ValueOf("Upload done successfully"))

			// resp := response{
			// 	Message: "Upload done successfully",
			// }
			resolve.Invoke(response)
		}()

		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

//-----------------------------------------------------------------------------

func main() {
	// Ported over a basic unit test to make sure it runs in the browser.
	// TestSSSignAndVerify(new(testing.T))

	fmt.Printf("0CHAIN - GOSDK (version=%v)\n", version.VERSIONSTR)

	c := make(chan struct{}, 0)

	// Just functions for 0proxy.
	js.Global().Set("initializeConfig", js.FuncOf(initializeConfig))
	js.Global().Set("initStorageSDK", js.FuncOf(initStorageSDK))
	js.Global().Set("Upload", js.FuncOf(Upload))
	js.Global().Set("Download", js.FuncOf(Download))
	js.Global().Set("Share", js.FuncOf(Share))
	js.Global().Set("Rename", js.FuncOf(Rename))
	js.Global().Set("Copy", js.FuncOf(Copy))
	js.Global().Set("Delete", js.FuncOf(Delete))
	js.Global().Set("Move", js.FuncOf(Move))
	js.Global().Set("GetClientEncryptedPublicKey", js.FuncOf(GetClientEncryptedPublicKey))

	// ethwallet.go
	js.Global().Set("TokensToEth", js.FuncOf(TokensToEth))
	js.Global().Set("EthToTokens", js.FuncOf(EthToTokens))
	js.Global().Set("GTokensToEth", js.FuncOf(GTokensToEth))
	js.Global().Set("GEthToTokens", js.FuncOf(GEthToTokens))
	js.Global().Set("ConvertZcnTokenToETH", js.FuncOf(ConvertZcnTokenToETH))
	js.Global().Set("SuggestEthGasPrice", js.FuncOf(SuggestEthGasPrice))
	js.Global().Set("TransferEthTokens", js.FuncOf(TransferEthTokens))

	js.Global().Set("GetWalletAddrFromEthMnemonic", js.FuncOf(GetWalletAddrFromEthMnemonic))
	js.Global().Set("IsValidEthAddress", js.FuncOf(IsValidEthAddress))
	js.Global().Set("CreateWalletFromEthMnemonic", js.FuncOf(CreateWalletFromEthMnemonic))

	// wallet.go
	js.Global().Set("InitZCNSDK", js.FuncOf(InitZCNSDK))

	<-c
}
