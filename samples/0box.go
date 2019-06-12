package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/trace"

	"github.com/0chain/gosdk/util"
	"github.com/0chain/clientsdk/zcn"
	"github.com/gosuri/uiprogress"
)

const Black = "\u001b[30m"
const Red = "\u001b[31m"
const Green = "\u001b[32m"
const Yellow = "\u001b[33m"
const Blue = "\u001b[34m"
const Magenta = "\u001b[35m"
const Cyan = "\u001b[36m"
const White = "\u001b[37m"
const Reset = "\u001b[0m"

const CONFIG_FILE = "out/config.json"

// var blobberStr string = `
// [
// {"id":"ccde4b06d02e24113889164d94a9692284f60c8701f10e600bde58889d335055","url":"http://localhost:5051"},
// {"id":"f45305566d865b06ff045e877d6daa76425ee04c84053d35aa462c4962351f45","url":"http://localhost:5052"},
// {"id":"73c154fc8347296880da8eaa9a8ad428d43196b438ceb28e64de1e19724e7e20","url":"http://localhost:5053"},
// {"id":"e417067e383ab7a46b404e300bf5a036ea50dc3a46e95e8a25c6d6e905fa8326","url":"http://localhost:5054"}
// ]
// `

var blobberStr string = `
[
{"id":"8fb6c86507b2c1618d0003db0d39a32f63a2ec2819a86a600c2e3f7fc60c6115","url":"http://b002.jaydevstorage.testnet-0chain.net:5051"},
{"id":"df48d4817bf439997caeed2a39310e8629db68bb3df22fac6cc2dc830fb9108c","url":"http://b003.jaydevstorage.testnet-0chain.net:5051"},
{"id":"5d1b7c76e4556be1757ed0576fd116fdca197d257ec57d4af9fc3d792d6a1f5a","url":"http://b000.jaydevstorage.testnet-0chain.net:5051"},
{"id":"7677d2ba522129b17fa65dff9a2c6717564d922177d848a20c8e08f50f378a5a","url":"http://b001.jaydevstorage.testnet-0chain.net:5051"}
]
`

var clientStr string = `{
	"id" : "8cd930c50b8e06d9ba2ab6a86ca9e3c6d073974d6976312f36a766a7443efd55",
	"public_key" : "78d4cd4d6edbfb3f0a1c7dec479d5b295b9abc2c1d2d332e67a115a62a3c1fd0",
	"private_key" : "c8e05e590c3beddf0c2a239d04a92c20323e660d92e9d2a096e46577f4595b1478d4cd4d6edbfb3f0a1c7dec479d5b295b9abc2c1d2d332e67a115a62a3c1fd0"
}`

const NUMDATASHARDS = 2
const NUMPARITYSHARDS = 2

var TOTALSHARDS int = NUMDATASHARDS + NUMPARITYSHARDS

func (s *StatusBar) Started(allocationId, filePath string, op int, totalBytes int) {
	s.b = uiprogress.AddBar(totalBytes)
	s.b.AppendCompleted()
	s.b.PrependElapsed()
}

func (s *StatusBar) InProgress(allocationId, filePath string, op int, completedBytes int) {
	s.b.Set(completedBytes)
}

func (s *StatusBar) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	// Not required
	// s.b.PrependElapsed()
	fmt.Println("Status completed callback. Type = " + mimetype + ". Name = " + filename)
}

type StatusBar struct {
	b *uiprogress.Bar
}

func getClientConfig(configFile string) string {
	if _, err := os.Stat(configFile); err == nil {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			fmt.Println("Error opening config", err)
		}
		if len(data) > 0 {
			return string(data)
		}
	}
	return clientStr
}

func getConfig(clientID string) (string, string) {
	configFileName := CONFIG_FILE + "-" + clientID
	if _, err := os.Stat(configFileName); err == nil {
		data, err := ioutil.ReadFile(configFileName)
		if err != nil {
			fmt.Println("Error opening config", err)
		}
		mapData := make(map[string]string, 1)
		err = json.Unmarshal(data, &mapData)
		if err == nil {
			return mapData["blobbers"], mapData["dirtree"]
		}
	}
	return blobberStr, ""
}

func saveConfig(alloc *zcn.Allocation, clientID string) {
	mapData := make(map[string]string, 2)
	mapData["blobbers"] = alloc.GetBlobbers()
	mapData["dirtree"] = alloc.GetDirTree()
	js, _ := json.Marshal(mapData)
	configFileName := CONFIG_FILE + "-" + clientID
	err := ioutil.WriteFile(configFileName, js, 0644)
	if err != nil {
		fmt.Println("Save config failed.")
	}
}

func main() {
	var allocation string
	flag.StringVar(&allocation, "allocation", "", "Allocation for the upload")

	var cmd string
	flag.StringVar(&cmd, "cmd", "", "upload/download/delete")

	var clientConfigPath string
	flag.StringVar(&clientConfigPath, "clientconfig", "", "path to the client configuration")

	var localPath string
	flag.StringVar(&localPath, "localpath", "", "Local file path to upload/download")

	var remotePath string
	flag.StringVar(&remotePath, "remotepath", "/", "Remote path to upload/download/delete")

	var authToken string
	flag.StringVar(&authToken, "authToken", "", "auth token for the shared file")

	var logFile string
	flag.StringVar(&logFile, "logFile", "out/0Box.log", "Full file path to write log")

	var logVerbose bool
	flag.BoolVar(&logVerbose, "logverbose", false, "Full file path to write log")

	flag.Parse()

	switch cmd {
	case "upload":
	case "download":
	case "delete":
	case "share":
	case "downloadshared":
	case "repair":
	case "stats":
	case "update":
	default:
		fmt.Println("Unsupported command:", cmd)
		flag.Usage()
		return
	}

	// ======== Trace Start ============
	f, err := os.Create("out/trace.out")
	if err != nil {
		log.Fatalf("failed to create trace output file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("failed to close trace file: %v", err)
		}
	}()

	if err := trace.Start(f); err != nil {
		log.Fatalf("failed to start trace: %v", err)
	}
	defer trace.Stop()
	// ========== Trace End ==============

	uiprogress.Start() // start rendering
	fmt.Println("0Box version : ", zcn.GetVersion())
	zcn.SetLogLevel(4)
	zcn.SetLogFile(logFile, logVerbose)
	alloc, err := zcn.CreateInstance(allocation)

	clientConfigStr := getClientConfig(clientConfigPath)

	//get client object from config str
	clientObj, _ := util.GetClientConfig(clientConfigStr)

	blobStr, dirStr := getConfig(clientObj.Id)
	err = alloc.SetConfig(clientConfigStr, dirStr, blobStr, NUMDATASHARDS, NUMPARITYSHARDS)
	if err != nil {
		return
	}

	var bar *uiprogress.Bar

	switch cmd {
	case "upload":
		err = alloc.UploadFile(localPath, remotePath, &StatusBar{b: bar})
		if err != nil {
			fmt.Println("Upload failed in SDK:", err)
			return
		}
		err = alloc.Commit()
		if err != nil {
			fmt.Println("Commit failed:", err)
			return
		}
		fmt.Println(Green, "File upload and commit success!!", Reset)
	case "download":
		err = alloc.DownloadFile(remotePath, localPath, &StatusBar{b: bar})
		if err != nil {
			fmt.Println("Download failed in SDK:", err)
			saveConfig(alloc, clientObj.Id)
			return
		}
		fmt.Println(Green, "File download success!!", Reset)
	case "delete":
		err = alloc.DeleteFile(remotePath)
		if err != nil {
			fmt.Println("Delete failed in SDK:", err)
			return
		}
		err = alloc.Commit()
		if err != nil {
			fmt.Println("Commit failed:", err)
			return
		}
		fmt.Println(Green, "File delete and commit success!!", Reset)
	case "share":
		authtoken := alloc.GetShareAuthToken(remotePath, "")
		fmt.Println("auth token : " + authtoken)
	case "downloadshared":
		err := alloc.DownloadFileFromShareLink(localPath, authToken, &StatusBar{b: bar})
		if err != nil {
			fmt.Println("Download from share link failed in SDK:", err)
			saveConfig(alloc, clientObj.Id)
			return
		}
		fmt.Println(Green, "File download success!!", Reset)
	case "repair":
		err = alloc.RepairFile(localPath, remotePath, &StatusBar{b: bar})
		if err != nil {
			fmt.Println("Repair failed in SDK:", err)
			return
		}
		err = alloc.Commit()
		if err != nil {
			fmt.Println("Commit failed:", err)
			return
		}
		fmt.Println(Green, "File repair and commit success!!", Reset)
	case "stats":
		stats := alloc.GetFileStats(remotePath)
		fmt.Println("Stats: ", stats)
	case "update":
		err = alloc.UpdateFile(localPath, remotePath, &StatusBar{b: bar})
		if err != nil {
			fmt.Println("Upload failed in SDK:", err)
			return
		}
		err = alloc.Commit()
		if err != nil {
			fmt.Println("Commit failed:", err)
			return
		}
		fmt.Println(Green, "File update and commit success!!", Reset)
	}
	zcn.CloseLog()
	saveConfig(alloc, clientObj.Id)
}
