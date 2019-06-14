package zcn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/0chain/gosdk/util"
)

type deleteFormData struct {
	ConnectionID string `json:"connection_id"`
	Filename     string `json:"filename"`
	Path         string `json:"filepath"`
}

func (obj *Allocation) deleteBlobberFile(blobber *util.Blobber, blobberIdx int, remotePath string) {
	defer obj.wg.Done()
	path, _ := filepath.Split(remotePath)
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}
	// TODO: Change it to File stats API
	lsData, err := obj.getDir(blobber, blobberIdx, path)
	if err != nil {
		Logger.Error(blobber.UrlRoot, remotePath, err)
		return
	}
	var blobberFileInfo map[string]interface{}
	for _, child := range lsData.Entities {
		if child["path"].(string) == remotePath {
			blobberFileInfo = child
			break
		}
	}
	if blobberFileInfo == nil {
		Logger.Error(blobber.UrlRoot, remotePath, " File not found")
		return
	}
	dt := util.NewDeleteToken()
	dt.FilePathHash = blobberFileInfo["path_hash"].(string)
	dt.FileRefHash = blobberFileInfo["hash"].(string)
	dt.AllocationID = obj.allocationId
	dt.Size = int64(blobberFileInfo["size"].(float64))
	dt.BlobberID = blobber.Id
	dt.Timestamp = util.Now()
	dt.ClientID = obj.client.Id
	err = dt.Sign(obj.client.PrivateKey)
	if err != nil {
		Logger.Error(blobber.UrlRoot, " Signing delete token", err)
		return
	}
	dtData, err := json.Marshal(dt)
	if err != nil {
		Logger.Error(blobber.UrlRoot, " Creating json delete token", err)
		return
	}
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	formData := deleteFormData{
		ConnectionID: fmt.Sprintf("%d", blobber.ConnObj.ConnectionId),
		Filename:     blobberFileInfo["name"].(string),
		Path:         blobberFileInfo["path"].(string),
	}
	var metaData []byte
	metaData, err = json.Marshal(formData)
	if err != nil {
		Logger.Error(blobber.UrlRoot, " creating delete formdata", err)
		return
	}
	formWriter.WriteField("uploadMeta", string(metaData))
	formWriter.WriteField("delete_token", string(dtData))
	formWriter.Close()
	req, err := util.NewDeleteRequest(blobber.UrlRoot, obj.allocationId, obj.client, body)
	if err != nil {
		Logger.Error(blobber.UrlRoot, "Error creating delete request", err)
		return
	}
	req.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	_ = httpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Delete : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			err = blobber.ConnObj.DeleteFile(remotePath, dt.Size)
			if err != nil {
				Logger.Error("Delete dirtree error ", err)
			}
			obj.consensus++
			Logger.Info(blobber.UrlRoot, " "+remotePath, " deleted.")
		} else {
			resp_body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				Logger.Error(blobber.UrlRoot, "Response: ", string(resp_body))
			}
		}
		return nil
	})
}

func (obj *Allocation) DeleteFile(remotePath string) error {
	fileInfo := util.GetFileInfo(&obj.dirTree, remotePath)
	if fileInfo == nil {
		// TODO: Use list API from blobbers to confirm
		return fmt.Errorf("Remote file doesn't exists")
	}
	obj.pauseBgSync()
	defer obj.resumeBgSync()
	obj.consensus = 0
	obj.wg.Add(len(obj.blobbers))
	for i := 0; i < len(obj.blobbers); i++ {
		go obj.deleteBlobberFile(&obj.blobbers[i], i, remotePath)
	}
	obj.wg.Wait()
	if !obj.isConsensusOk() {
		return fmt.Errorf("Delete failed: Success_rate:%2f, expected:%2f", obj.getConsensusRate(), obj.getConsensusRequiredForOk())
	}
	err := util.DeleteFile(&obj.dirTree, remotePath)
	if err != nil {
		Logger.Error("Deleting file from dir tree", err)
	}
	return nil
}
