package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall/js"

	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"go.uber.org/zap"
	"gopkg.in/cheggaaa/pb.v1"
)

func validateClientDetails(allocation, clientJSON string) error {
	if len(allocation) == 0 || len(clientJSON) == 0 {
		return NewError("invalid_param", "Please provide allocation and client_json for the client")
	}
	return nil
}

// This function try to execute wasm functions that are wrapped with "Promise"
// see: https://stackoverflow.com/questions/68426700/how-to-wait-a-js-async-function-from-golang-wasm/68427221#comment120939975_68427221
func await(awaitable js.Value) ([]js.Value, []js.Value) {
	then := make(chan []js.Value)
	defer close(then)
	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		then <- args
		return nil
	})
	defer thenFunc.Release()

	catch := make(chan []js.Value)
	defer close(catch)
	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		catch <- args
		return nil
	})
	defer catchFunc.Release()

	awaitable.Call("then", thenFunc).Call("catch", catchFunc)

	select {
	case result := <-then:
		return result, []js.Value{js.Null()}
	case err := <-catch:
		return []js.Value{js.Null()}, err
	}
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

/*InvalidRequest - create error messages that are needed when validating request input */
func InvalidRequest(msg string) error {
	return NewError("invalid_request", fmt.Sprintf("Invalid request (%v)", msg))
}

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
	chainJSON := p[2].String()
	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	remotePath := p[3].String()
	newName := p[4].String()
	if len(remotePath) == 0 || len(newName) == 0 {
		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path and new_name for rename").Error())
	}

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err = initSDK(clientJSON, chainJSON)
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

			resolve.Invoke(response)
		}()

		return nil
	})

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
	chainJSON := p[2].String()
	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	remotePath := p[3].String()
	destPath := p[4].String()
	if len(remotePath) == 0 || len(destPath) == 0 {
		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path and dest_path for copy").Error())
	}

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err = initSDK(clientJSON, chainJSON)
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

			resolve.Invoke(response)
		}()

		return nil
	})

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
	chainJSON := p[2].String()
	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	remotePath := p[3].String()
	if len(remotePath) == 0 {
		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path for share").Error())
	}

	refereeClientID := p[3].String()
	encryptionpublickey := p[4].String()
	expiry := p[5].Int()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err = initSDK(clientJSON, chainJSON)
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

			at, err := allocationObj.GetAuthTicket(remotePath, fileName, refType, refereeClientID, encryptionpublickey, int64(expiry))
			if err != nil {
				reject.Invoke(js.ValueOf("error: " + NewError("get_auth_ticket_failed", err.Error()).Error()))
				return
			}

			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(js.ValueOf(at))

			resolve.Invoke(response)
		}()

		return nil
	})

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
	chainJSON := p[2].String()
	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	remotePath := p[3].String()
	destPath := p[4].String()
	if len(remotePath) == 0 || len(destPath) == 0 {
		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path and dest_path for move").Error())
	}

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err = initSDK(clientJSON, chainJSON)
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

			resolve.Invoke(response)
		}()

		return nil
	})

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
	chainJSON := p[3].String()
	err := validateClientDetails(allocation, clientJSON)
	if err != nil {
		return js.ValueOf("error: " + err.Error())
	}

	remotePath := p[4].String()
	if len(remotePath) == 0 {
		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path for upload").Error())
	}

	workdir := p[5].String()

	Filename := p[5].String()
	file := p[6].String()
	// file, fileHeader, err := r.FormFile("file")
	// if err != nil {
	// 	js.ValueOf("error: " + NewError("invalid_params", "Unable to get file for upload :"+err.Error()).Error())
	// }
	// defer file.Close()
	encrypt := p[7].String()

	fileAttrs := p[8].String()
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

			err = initSDK(clientJSON, chainJSON)
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
					err = allocationObj.EncryptAndUploadFile(workdir, localFilePath, remotePath, attrs, statusBar)
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

			resolve.Invoke(response)
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/zproxycore/handler/download.go`
//-----------------------------------------------------------------------------

// Download is to download a file from dStorage
// TODO: this should be a dict-type, like a JSON, instead of a long list.
func Download(this js.Value, p []js.Value) interface{} {
	allocation := p[0].String()
	clientJSON := p[1].String()
	chainJSON := p[2].String()
	remotePath := p[3].String()
	authTicket := p[4].String()
	numBlocks := p[5].String()
	rx_pay := p[6].String()
	file_name := p[7].String()
	lookuphash := p[8].String()

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
			err = initSDK(clientJSON, chainJSON)
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
	chainJSON := p[2].String()
	remotePath := p[3].String()

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
			err = initSDK(clientJSON, chainJSON)
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

			resolve.Invoke(response)
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}
