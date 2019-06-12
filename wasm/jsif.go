package main

import (
	"fmt"
	"syscall/js"

	"0chain.net/clientsdk/zcn"
	// "time"
)

const (
	SUCCESS int = 0
	FAILED  int = -1
)

const (
	STARTED    int = 0
	INPROGRESS int = 1
	COMPLETED  int = 2
)

type zcnJsIf struct {
	setLogLevelCb               js.Callback
	setLogFileCb                js.Callback
	createInstanceCb            js.Callback
	setConfigCb                 js.Callback
	uploadFileCb                js.Callback
	updateFileCb                js.Callback
	repairFileCb                js.Callback
	commitCb                    js.Callback
	downloadFileCb              js.Callback
	downloadCancelCb            js.Callback
	getBlobbersCb               js.Callback
	getDirTreeCb                js.Callback
	listDirCb                   js.Callback
	addDirCb                    js.Callback
	deleteFileCb                js.Callback
	getShareAuthTokenCb         js.Callback
	downloadFileFromShareLinkCb js.Callback
	getFileStatsCb              js.Callback
	unloadCh                    chan struct{}
	unloadCb                    js.Callback

	// Allocation object
	allocation *zcn.Allocation
}

type statusCb struct {
	cbFn string
}

// Singleton instance
var zjs zcnJsIf

func init() {
	zjs.setLogLevelCb = js.NewCallback(ZcnSetLogLevel)
	zjs.setLogFileCb = js.NewCallback(ZcnSetLogFile)
	zjs.createInstanceCb = js.NewCallback(ZcnCreateInstance)
	zjs.setConfigCb = js.NewCallback(ZcnSetConfig)
	zjs.uploadFileCb = js.NewCallback(ZcnUploadFile)
	zjs.updateFileCb = js.NewCallback(ZcnUpdateFile)
	zjs.repairFileCb = js.NewCallback(ZcnRepairFile)
	zjs.commitCb = js.NewCallback(ZcnCommit)
	zjs.downloadFileCb = js.NewCallback(ZcnDownloadFile)
	zjs.downloadCancelCb = js.NewCallback(ZcnDownloadCancel)
	zjs.getBlobbersCb = js.NewCallback(ZcnGetBlobbers)
	zjs.getDirTreeCb = js.NewCallback(ZcnGetDirTree)
	zjs.listDirCb = js.NewCallback(ZcnListDir)
	zjs.addDirCb = js.NewCallback(ZcnAddDir)
	zjs.deleteFileCb = js.NewCallback(ZcnDeleteFile)
	zjs.getShareAuthTokenCb = js.NewCallback(ZcnGetShareAuthToken)
	zjs.downloadFileFromShareLinkCb = js.NewCallback(ZcnDownloadFileFromShareLink)
	zjs.getFileStatsCb = js.NewCallback(ZcnGetFileStats)
	zjs.unloadCh = make(chan struct{})
	zjs.unloadCb = js.NewCallback(ZcnUnload)
}

func exportFunctions() {
	js.Global().Set("ZcnSetLogLevel", zjs.setLogLevelCb)
	js.Global().Set("ZcnSetLogFile", zjs.setLogFileCb)
	js.Global().Set("ZcnCreateInstance", zjs.createInstanceCb)
	js.Global().Set("ZcnSetConfig", zjs.setConfigCb)
	js.Global().Set("ZcnUploadFile", zjs.uploadFileCb)
	js.Global().Set("ZcnUpdateFile", zjs.updateFileCb)
	js.Global().Set("ZcnRepairFile", zjs.repairFileCb)
	js.Global().Set("ZcnCommit", zjs.commitCb)
	js.Global().Set("ZcnDownloadFile", zjs.downloadFileCb)
	js.Global().Set("ZcnDownloadCancel", zjs.downloadCancelCb)
	js.Global().Set("ZcnGetBlobbers", zjs.getBlobbersCb)
	js.Global().Set("ZcnGetDirTree", zjs.getDirTreeCb)
	js.Global().Set("ZcnListDir", zjs.listDirCb)
	js.Global().Set("ZcnAddDir", zjs.addDirCb)
	js.Global().Set("ZcnDeleteFile", zjs.deleteFileCb)
	js.Global().Set("ZcnGetShareAuthToken", zjs.getShareAuthTokenCb)
	js.Global().Set("ZcnDownloadFileFromShareLink", zjs.downloadFileFromShareLinkCb)
	js.Global().Set("ZcnGetFileStats", zjs.getFileStatsCb)
	js.Global().Set("ZcnUnload", zjs.unloadCb)
}

func releaseFunctions() {
	zjs.setLogLevelCb.Release()
	zjs.setLogFileCb.Release()
	zjs.createInstanceCb.Release()
	zjs.setConfigCb.Release()
	zjs.uploadFileCb.Release()
	zjs.updateFileCb.Release()
	zjs.repairFileCb.Release()
	zjs.commitCb.Release()
	zjs.downloadFileCb.Release()
	zjs.downloadCancelCb.Release()
	zjs.getBlobbersCb.Release()
	zjs.getDirTreeCb.Release()
	zjs.listDirCb.Release()
	zjs.addDirCb.Release()
	zjs.deleteFileCb.Release()
	zjs.setLogLevelCb.Release()
	zjs.getShareAuthTokenCb.Release()
	zjs.downloadFileFromShareLinkCb.Release()
	zjs.getFileStatsCb.Release()
}

// Callback function for upload/download
func (s *statusCb) Started(allocationId, filePath string, op int, totalBytes int) {
	js.Global().Call(s.cbFn, STARTED, op, SUCCESS, filePath, "", "", totalBytes, "")
}
func (s *statusCb) InProgress(allocationId, filePath string, op int, completedBytes int) {
	js.Global().Call(s.cbFn, INPROGRESS, op, SUCCESS, filePath, "", "", completedBytes, "")
}
func (s *statusCb) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	js.Global().Call(s.cbFn, COMPLETED, op, SUCCESS, filePath, filename, mimetype, 0, "")
}

// args[0]: Log level (int)
func ZcnSetLogLevel(args []js.Value) {
	if len(args) < 1 {
		return
	}
	zcn.SetLogLevel(args[0].Int())
}

// args[0]: Log file
// args[1]: verbose - true - console output; false - no console output
func ZcnSetLogFile(args []js.Value) {
	if len(args) < 2 {
		return
	}
	zcn.SetLogFile(args[0].String(), args[1].Bool())
}

// args[0] : Allocation ID string
func ZcnCreateInstance(args []js.Value) {
	if len(args) < 1 {
		return
	}
	allocationId := js.ValueOf(args[0]).String()
	var err error
	zjs.allocation, err = zcn.CreateInstance(allocationId)
	if err != nil {
		fmt.Println("AllocationId :", allocationId, "creation failed")
	}
}

// args[0] : Client config JSON
// args[1] : Dir Tree JSON
// args[2] : Blobber config JSON
// args[3] : Number Data shards
// args[4] : Number of parity shards
// args[5] : Callback function
func ZcnSetConfig(args []js.Value) {
	if len(args) < 6 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[5].String(), FAILED, "No allocation found")
		return
	}
	go func() {
		err := zjs.allocation.SetConfig(args[0].String(),
			args[1].String(),
			args[2].String(),
			args[3].Int(), args[4].Int())
		if err != nil {
			js.Global().Call(args[5].String(), FAILED, fmt.Sprintf("Invalid config: %s", err.Error()))
			return
		}
		js.Global().Call(args[5].String(), SUCCESS, "")
	}()
}

// args[0] : path
// args[1] : Callback function - fn(status, string)
func ZcnListDir(args []js.Value) {
	if len(args) < 2 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[1].String(), FAILED, "No allocation found")
		return
	}
	go func() {
		str := zjs.allocation.ListDir(args[0].String())
		js.Global().Call(args[1].String(), SUCCESS, str)
	}()
}

// args[0] : path
// args[1] : Callback function - fn(status, string)
func ZcnAddDir(args []js.Value) {
	if len(args) < 2 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[1].String(), FAILED, "No allocation found")
		return
	}
	go func() {
		err := zjs.allocation.AddDir(args[0].String())
		if err != nil {
			js.Global().Call(args[1].String(), FAILED, err.Error())
		} else {
			js.Global().Call(args[1].String(), SUCCESS, "")
		}
	}()
}

// args[0] : Local file path
// args[1] : Remote path
// args[2] : Callback function
func ZcnUploadFile(args []js.Value) {
	if len(args) < 3 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[2].String(), COMPLETED, zcn.OpUpload, FAILED, args[1].String(), 0, "No allocation found")
		return
	}
	go func() {
		err := zjs.allocation.UploadFile(args[0].String(), args[1].String(), &statusCb{cbFn: args[2].String()})
		if err != nil {
			js.Global().Call(args[2].String(), COMPLETED, zcn.OpUpload, FAILED, args[1].String(), 0, fmt.Sprintf("Upload failed in SDK: %s", err.Error()))
			return
		}
	}()
}

// args[0] : Local file path
// args[1] : Remote path
// args[2] : Callback function
func ZcnUpdateFile(args []js.Value) {
	if len(args) < 3 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[2].String(), COMPLETED, zcn.OpUpload, FAILED, args[1].String(), 0, "No allocation found")
		return
	}
	go func() {
		err := zjs.allocation.UpdateFile(args[0].String(), args[1].String(), &statusCb{cbFn: args[2].String()})
		if err != nil {
			js.Global().Call(args[2].String(), COMPLETED, zcn.OpUpload, FAILED, args[1].String(), 0, fmt.Sprintf("Upload failed in SDK: %s", err.Error()))
			return
		}
	}()
}

// args[0] : Local file path
// args[1] : Remote path
// args[2] : Callback function
func ZcnRepairFile(args []js.Value) {
	if len(args) < 3 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[2].String(), COMPLETED, zcn.OpRepair, FAILED, args[1].String(), 0, "No allocation found")
		return
	}
	go func() {
		err := zjs.allocation.RepairFile(args[0].String(), args[1].String(), &statusCb{cbFn: args[2].String()})
		if err != nil {
			js.Global().Call(args[2].String(), COMPLETED, zcn.OpRepair, FAILED, args[1].String(), 0, fmt.Sprintf("Upload failed in SDK: %s", err.Error()))
			return
		}
	}()
}

// args[0] : Callback function
func ZcnCommit(args []js.Value) {
	if len(args) < 1 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[0].String(), FAILED, "No allocation found")
		return
	}
	go func() {
		err := zjs.allocation.Commit()
		if err != nil {
			js.Global().Call(args[0].String(), FAILED, fmt.Sprintf("Commit failed in SDK: %s", err.Error()))
			return
		}
		js.Global().Call(args[0].String(), SUCCESS, "")
	}()
}

// args[0] : Remote path
// args[1] : Local path to save file
// args[2] : Callback function
func ZcnDownloadFile(args []js.Value) {
	if len(args) < 3 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[2].String(), COMPLETED, zcn.OpDownload, FAILED, args[1].String(), 0, "No allocation found")
		return
	}
	go func() {
		err := zjs.allocation.DownloadFile(args[0].String(), args[1].String(), &statusCb{cbFn: args[2].String()})
		if err != nil {
			js.Global().Call(args[2].String(), COMPLETED, zcn.OpDownload, FAILED, args[1].String(), 0, fmt.Sprintf("Download failed in SDK: %s", err.Error()))
			return
		}
	}()
}

// No arguments
func ZcnDownloadCancel(args []js.Value) {
	if zjs.allocation == nil {
		return
	}
	zjs.allocation.DownloadCancel()
}

// args[0] : Remote path
// args[1] : Callback function - fn(status, string)
func ZcnDeleteFile(args []js.Value) {
	if len(args) < 2 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[1].String(), FAILED, "No allocation found")
		return
	}
	go func() {
		err := zjs.allocation.DeleteFile(args[0].String())
		if err != nil {
			js.Global().Call(args[1].String(), FAILED, err.Error())
		} else {
			js.Global().Call(args[1].String(), SUCCESS, "")
		}
	}()
}

// args[0] : Callback function
func ZcnGetBlobbers(args []js.Value) {
	if len(args) < 1 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[0].String(), FAILED, "No allocation found")
		return
	}
	blobberStr := zjs.allocation.GetBlobbers()
	js.Global().Call(args[0].String(), SUCCESS, blobberStr)
}

// args[0] : Callback function
func ZcnGetDirTree(args []js.Value) {
	if len(args) < 1 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[0].String(), FAILED, "No allocation found")
		return
	}
	dirTree := zjs.allocation.GetDirTree()
	js.Global().Call(args[0].String(), SUCCESS, dirTree)
}

// args[0] : Remote path
// args[1] : ClientID
// args[2] : Callback function
func ZcnGetShareAuthToken(args []js.Value) {
	if len(args) < 3 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[2].String(), FAILED, "No allocation found")
		return
	}
	go func() {
		authtoken := zjs.allocation.GetShareAuthToken(args[0].String(), args[1].String())
		js.Global().Call(args[2].String(), SUCCESS, authtoken)
	}()
}

// args[0] : Local path to save file
// args[1] : Auth Token
// args[2] : Callback function
func ZcnDownloadFileFromShareLink(args []js.Value) {
	if len(args) < 3 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[2].String(), COMPLETED, zcn.OpDownload, FAILED, args[1].String(), 0, "No allocation found")
		return
	}
	go func() {
		err := zjs.allocation.DownloadFile(args[0].String(), args[1].String(), &statusCb{cbFn: args[2].String()})
		if err != nil {
			js.Global().Call(args[2].String(), COMPLETED, zcn.OpDownload, FAILED, args[1].String(), 0, fmt.Sprintf("Download failed in SDK: %s", err.Error()))
			return
		}
	}()
}

// args[0] : Remote path
// args[1] : Callback function
func ZcnGetFileStats(args []js.Value) {
	if len(args) < 2 {
		return
	}
	if zjs.allocation == nil {
		js.Global().Call(args[1].String(), FAILED, "No allocation found")
		return
	}
	go func() {
		stats := zjs.allocation.GetFileStats(args[0].String())
		js.Global().Call(args[1].String(), SUCCESS, stats)
	}()
}

// No argument
func ZcnUnload(args []js.Value) {
	zjs.unloadCh <- struct{}{}
}

func main() {
	exportFunctions()
	fmt.Println("0Chain SDK WASM Initialized!!")
	// Wait forever event to cleanup resource
	<-zjs.unloadCh
	releaseFunctions()
	fmt.Println("0Chain SDK WASM Uninitialized!!")
}
