//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"syscall/js"
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
// Ported over from `code/go/0proxy.io/zproxycore/handler/upload.go`
//-----------------------------------------------------------------------------

// Upload is to upload file to dStorage
// func Upload(this js.Value, p []js.Value) interface{} {

// 	method := p[0].String() // POST or PUT
// 	allocation := p[1].String()
// 	clientJSON := p[2].String()
// 	chainJSON := p[3].String()
// 	err := validateClientDetails(allocation, clientJSON)
// 	if err != nil {
// 		return js.ValueOf("error: " + err.Error())
// 	}

// 	remotePath := p[4].String()
// 	if len(remotePath) == 0 {
// 		return js.ValueOf("error: " + NewError("invalid_param", "Please provide remote_path for upload").Error())
// 	}

// 	workdir := p[5].String()

// 	Filename := p[5].String()
// 	file := p[6].String()
// 	// file, fileHeader, err := r.FormFile("file")
// 	// if err != nil {
// 	// 	js.ValueOf("error: " + NewError("invalid_params", "Unable to get file for upload :"+err.Error()).Error())
// 	// }
// 	// defer file.Close()
// 	encrypt := p[7].String()

// 	fileAttrs := p[8].String()
// 	var attrs fileref.Attributes
// 	if len(fileAttrs) > 0 {
// 		err := json.Unmarshal([]byte(fileAttrs), &attrs)
// 		if err != nil {
// 			return js.ValueOf("error: " + NewError("failed_to_parse_file_attrs", err.Error()).Error())
// 		}
// 	}

// 	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
// 		resolve := args[0]
// 		reject := args[1]

// 		go func() {
// 			createDirIfNotExists(allocation)

// 			// localFilePath, err := writeFile(file, getPath(allocation, fileHeader.Filename))
// 			localFilePath, err := writeFile2(file, getPath(allocation, Filename))
// 			if err != nil {
// 				reject.Invoke(js.ValueOf("error: " + NewError("write_local_temp_file_failed", err.Error()).Error()))
// 				return
// 			}
// 			defer deleletFile(localFilePath)

// 			err = initSDK(clientJSON, chainJSON)
// 			if err != nil {
// 				reject.Invoke(js.ValueOf("error: " + NewError("sdk_not_initialized", "Unable to initialize gosdk with the given client details").Error()))
// 				return
// 			}

// 			allocationObj, err := sdk.GetAllocation(allocation)
// 			if err != nil {
// 				reject.Invoke(js.ValueOf("error: " + NewError("get_allocation_failed", err.Error()).Error()))
// 				return
// 			}

// 			wg := &sync.WaitGroup{}
// 			statusBar := &StatusBar{wg: wg}
// 			wg.Add(1)
// 			if method == "POST" {
// 				encryptBool, _ := strconv.ParseBool(encrypt)
// 				if encryptBool {
// 					// Logger.Info("Doing encrypted file upload with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
// 					fmt.Println("Doing encrypted file upload with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
// 					err = allocationObj.EncryptAndUploadFile(workdir, localFilePath, remotePath, attrs, statusBar)
// 				} else {
// 					// Logger.Info("Doing file upload with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
// 					fmt.Println("Doing file upload with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
// 					err = allocationObj.UploadFile(os.TempDir(), localFilePath, remotePath, attrs, statusBar)
// 				}
// 			} else {
// 				// Logger.Info("Doing file update with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
// 				fmt.Println("Doing file update with", zap.Any("remotepath", remotePath), zap.Any("allocation", allocationObj.ID))
// 				err = allocationObj.UpdateFile(os.TempDir(), localFilePath, remotePath, attrs, statusBar)
// 			}
// 			if err != nil {
// 				reject.Invoke(js.ValueOf("error: " + NewError("upload_file_failed", err.Error()).Error()))
// 				return
// 			}

// 			wg.Wait()
// 			if !statusBar.success {
// 				reject.Invoke(js.ValueOf("error: " + statusBar.err.Error()))
// 				return
// 			}

// 			responseConstructor := js.Global().Get("Response")
// 			response := responseConstructor.New(js.ValueOf("Upload done successfully"))

// 			resolve.Invoke(response)
// 		}()

// 		return nil
// 	})

// 	promiseConstructor := js.Global().Get("Promise")
// 	return promiseConstructor.New(handler)
// }
