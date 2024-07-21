package sdk

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"

	"github.com/0chain/gosdk/zboxcore/client"

	"golang.org/x/crypto/sha3"
)

// ChunkedUploadFormBuilder build form data for uploading
type ChunkedUploadFormBuilder interface {
	// build form data
	Build(
		fileMeta *FileMeta, hasher Hasher, connectionID string,
		chunkSize int64, chunkStartIndex, chunkEndIndex int,
		isFinal bool, encryptedKey, encryptedKeyPoint string, fileChunksData [][]byte,
		thumbnailChunkData []byte, shardSize int64,
	) (blobberData, error)
}

// ChunkedUploadFormMetadata upload form metadata
type ChunkedUploadFormMetadata struct {
	FileBytesLen         int
	ThumbnailBytesLen    int
	ContentType          string
	ThumbnailContentHash string
	DataHash             string
}

// CreateChunkedUploadFormBuilder create ChunkedUploadFormBuilder instance
func CreateChunkedUploadFormBuilder() ChunkedUploadFormBuilder {
	return &chunkedUploadFormBuilder{}
}

type chunkedUploadFormBuilder struct {
}

const MAX_BLOCKS = 80 // 5MB(CHUNK_SIZE*80)

func (b *chunkedUploadFormBuilder) Build(
	fileMeta *FileMeta, hasher Hasher, connectionID string,
	chunkSize int64, chunkStartIndex, chunkEndIndex int,
	isFinal bool, encryptedKey, encryptedKeyPoint string, fileChunksData [][]byte,
	thumbnailChunkData []byte, shardSize int64,
) (blobberData, error) {

	metadata := ChunkedUploadFormMetadata{
		ThumbnailBytesLen: len(thumbnailChunkData),
	}

	var res blobberData

	if len(fileChunksData) == 0 {
		return res, nil
	}

	numBodies := len(fileChunksData) / MAX_BLOCKS
	if len(fileChunksData)%MAX_BLOCKS > 0 {
		numBodies++
	}
	dataBuffers := make([]*bytes.Buffer, 0, numBodies)
	contentSlice := make([]string, 0, numBodies)

	formData := UploadFormData{
		ConnectionID: connectionID,
		Filename:     fileMeta.RemoteName,
		Path:         fileMeta.RemotePath,

		ActualSize: fileMeta.ActualSize,

		ActualThumbHash: fileMeta.ActualThumbnailHash,
		ActualThumbSize: fileMeta.ActualThumbnailSize,

		MimeType: fileMeta.MimeType,

		// IsFinal:           isFinal,
		ChunkSize:         chunkSize,
		ChunkStartIndex:   chunkStartIndex,
		ChunkEndIndex:     chunkEndIndex,
		UploadOffset:      chunkSize * int64(chunkStartIndex),
		Size:              shardSize,
		EncryptedKeyPoint: encryptedKeyPoint,
		EncryptedKey:      encryptedKey,
		CustomMeta:        fileMeta.CustomMeta,
	}

	for i := 0; i < numBodies; i++ {

		startRange := i * MAX_BLOCKS
		endRange := startRange + MAX_BLOCKS
		if endRange > len(fileChunksData) {
			endRange = len(fileChunksData)
		}

		bodyBuf := make([]byte, 0, (CHUNK_SIZE*(endRange-startRange))+1024)

		body := bytes.NewBuffer(bodyBuf)
		formWriter := multipart.NewWriter(body)
		defer formWriter.Close()

		uploadFile, err := formWriter.CreateFormFile("uploadFile", formData.Filename)
		if err != nil {
			return res, err
		}

		for _, chunkBytes := range fileChunksData[startRange:endRange] {
			_, err = uploadFile.Write(chunkBytes)
			if err != nil {
				return res, err
			}

			err = hasher.WriteToBlockHasher(chunkBytes)
			if err != nil {
				return res, err
			}

			metadata.FileBytesLen += len(chunkBytes)
		}

		if isFinal && i == numBodies-1 {

			actualHashSignature, err := client.Sign(fileMeta.ActualHash)
			if err != nil {
				return res, err
			}

			formData.ActualHash = fileMeta.ActualHash
			formData.ActualFileHashSignature = actualHashSignature
			formData.ActualSize = fileMeta.ActualSize
			dataHash, err := hasher.GetBlockHash()
			if err != nil {
				return res, err
			}
			formData.DataHash = dataHash
			formData.DataHashSignature, err = client.Sign(dataHash)
			if err != nil {
				return res, err
			}
		}

		thumbnailSize := len(thumbnailChunkData)
		if thumbnailSize > 0 && i == 0 {

			uploadThumbnailFile, err := formWriter.CreateFormFile("uploadThumbnailFile", fileMeta.RemoteName+".thumb")
			if err != nil {

				return res, err
			}

			thumbnailHash := sha3.New256()
			thumbnailWriters := io.MultiWriter(uploadThumbnailFile, thumbnailHash)
			_, err = thumbnailWriters.Write(thumbnailChunkData)
			if err != nil {
				return res, err
			}
			_, err = thumbnailHash.Write([]byte(fileMeta.RemotePath))
			if err != nil {
				return res, err
			}
			formData.ActualThumbSize = fileMeta.ActualThumbnailSize
			formData.ThumbnailContentHash = hex.EncodeToString(thumbnailHash.Sum(nil))

		}
		if i > 0 {
			formData.UploadOffset = formData.UploadOffset + chunkSize*int64(MAX_BLOCKS)
		}

		err = formWriter.WriteField("connection_id", connectionID)
		if err != nil {
			return res, err
		}

		if isFinal && i == numBodies-1 {
			formData.IsFinal = true
		}

		uploadMeta, err := json.Marshal(formData)
		if err != nil {
			return res, err
		}

		err = formWriter.WriteField("uploadMeta", string(uploadMeta))
		if err != nil {
			return res, err
		}

		contentSlice = append(contentSlice, formWriter.FormDataContentType())
		dataBuffers = append(dataBuffers, body)
	}
	metadata.ThumbnailContentHash = formData.ThumbnailContentHash
	metadata.DataHash = formData.DataHash
	res.dataBuffers = dataBuffers
	res.contentSlice = contentSlice
	res.formData = metadata
	return res, nil
}
