package sdk

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/logger"
)

// ChunkedUploadProgressStorer load and save upload progress
type ChunkedUploadProgressStorer interface {
	// Load load upload progress by id
	Load(id string) *UploadProgress
	// Save save upload progress
	Save(up *UploadProgress)
	// Remove remove upload progress by id
	Remove(id string) error
}

// fsChunkedUploadProgressStorer load and save upload progress in file system
type fsChunkedUploadProgressStorer struct {
	isRemoved bool
	up        *UploadProgress
}

func createFsChunkedUploadProgress(ctx context.Context) *fsChunkedUploadProgressStorer {
	up := &fsChunkedUploadProgressStorer{}

	go up.start()

	return up
}

func (fs *fsChunkedUploadProgressStorer) start() {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {

		<-ticker.C

		if fs.up == nil {
			continue
		}

		if fs.isRemoved {
			break
		}

		buf, err := json.Marshal(fs.up)
		if err != nil {
			logger.Logger.Error("[progress] save ", fs.up, err)
			continue
		}

		err = FS.WriteFile(fs.up.ID, buf, 0644)
		if err != nil {
			logger.Logger.Error("[progress] save ", fs.up, err)
			continue
		}

	}
}

// Load load upload progress from file system
func (fs *fsChunkedUploadProgressStorer) Load(progressID string) *UploadProgress {

	progress := new(UploadProgress)

	buf, err := FS.ReadFile(progressID)

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	// progress storerer Map String Interface
	psMSI := make(map[string]interface{})
	if err := json.Unmarshal(buf, &psMSI); err != nil {
		return nil
	}

	idI, ok := psMSI["id"]
	if !ok {
		return nil
	}

	progress.ID = idI.(string)

	chunkSizeI, ok := psMSI["chunk_size"]
	if !ok {
		return nil
	}

	progress.ChunkSize = int64(chunkSizeI.(float64))

	connectionIDI, ok := psMSI["connection_id"]
	if !ok {
		return nil
	}
	progress.ConnectionID = connectionIDI.(string)

	merkleHashersI, ok := psMSI["merkle_hashers"]
	if !ok {
		return nil
	}

	uploadBlobberStatuses := make([]*UploadBlobberStatus, 0)

	merkleHashers, ok := merkleHashersI.([]interface{})
	if !ok {
		return nil
	}
	for _, merkleHashI := range merkleHashers {
		merkleHashMap, ok := merkleHashI.(map[string]interface{})
		if !ok {
			return nil
		}
		uploadLength, ok := merkleHashMap["upload_length"].(float64)
		if !ok {
			return nil
		}
		hasherMap, ok := merkleHashMap["Hasher"].(map[string]interface{})
		if !ok {
			return nil
		}

		challenge := new(util.FixedMerkleTree)
		content := new(util.CompactMerkleTree)

		for key, value := range hasherMap {
			switch key {
			case "file":
				continue
			case "challenge":
				marshalledValue, err := json.Marshal((value))
				if err != nil {
					return nil
				}
				if err := json.Unmarshal(marshalledValue, challenge); err != nil {
					return nil
				}
			case "content":
				marshaledvalue, err := json.Marshal(value)
				if err != nil {
					return nil
				}
				if err := json.Unmarshal(marshaledvalue, content); err != nil {
					return nil
				}
			}
		}

		h := hasher{}
		h.File = sha256.New()
		h.Challenge = challenge
		h.Content = content

		ubs := UploadBlobberStatus{ // UploadBlobberStatus
			Hasher:       &h,
			UploadLength: int64(uploadLength),
		}

		uploadBlobberStatuses = append(uploadBlobberStatuses, &ubs)
	}

	progress.Blobbers = uploadBlobberStatuses
	return progress
}

// Save save upload progress in file system
func (fs *fsChunkedUploadProgressStorer) Save(up *UploadProgress) {
	fs.up = up
}

// Remove remove upload progress from file system
func (fs *fsChunkedUploadProgressStorer) Remove(progressID string) error {
	fs.isRemoved = true
	err := FS.Remove(progressID)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	return nil
}
