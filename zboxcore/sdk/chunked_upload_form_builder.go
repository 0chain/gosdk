package sdk

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
)

// ChunkedUploadFormBuilder build form data for uploading
type ChunkedUploadFormBuilder interface {
	// build form data
	Build(fileMeta *FileMeta, hasher Hasher, connectionID string, chunkSize int64, chunkStartIndex, chunkEndIndex int, isFinal bool, encryptedKey string, fileChunksData [][]byte, thumbnailChunkData []byte) (*bytes.Buffer, ChunkedUploadFormMetadata, error)
}

// ChunkedUploadFormMetadata upload form metadata
type ChunkedUploadFormMetadata struct {
	FileBytesLen         int
	ThumbnailBytesLen    int
	ContentType          string
	ChunkHash            string
	ChallengeHash        string
	ContentHash          string
	ThumbnailContentHash string
}

// CreateChunkedUploadFormBuilder create ChunkedUploadFormBuilder instance
func CreateChunkedUploadFormBuilder() ChunkedUploadFormBuilder {
	return &chunkedUploadFormBuilder{}
}

type chunkedUploadFormBuilder struct {
}

func (b *chunkedUploadFormBuilder) Build(fileMeta *FileMeta, hasher Hasher, connectionID string, chunkSize int64, chunkStartIndex, chunkEndIndex int, isFinal bool, encryptedKey string, fileChunksData [][]byte, thumbnailChunkData []byte) (*bytes.Buffer, ChunkedUploadFormMetadata, error) {

	metadata := ChunkedUploadFormMetadata{
		ThumbnailBytesLen: len(thumbnailChunkData),
	}

	if len(fileChunksData) == 0 {
		return nil, metadata, nil
	}

	body := &bytes.Buffer{}

	formData := UploadFormData{
		ConnectionID: connectionID,
		Filename:     fileMeta.RemoteName,
		Path:         fileMeta.RemotePath,

		ActualSize: fileMeta.ActualSize,

		ActualThumbHash: fileMeta.ActualThumbnailHash,
		ActualThumbSize: fileMeta.ActualThumbnailSize,

		MimeType: fileMeta.MimeType,

		IsFinal:         isFinal,
		ChunkSize:       chunkSize,
		ChunkStartIndex: chunkStartIndex,
		ChunkEndIndex:   chunkEndIndex,
		UploadOffset:    chunkSize * int64(chunkStartIndex),
	}

	formWriter := multipart.NewWriter(body)
	defer formWriter.Close()

	uploadFile, err := formWriter.CreateFormFile("uploadFile", formData.Filename)
	if err != nil {
		return nil, metadata, err
	}

	chunkHashWriter := sha256.New()
	chunksHashWriter := sha256.New()
	chunksWriters := io.MultiWriter(uploadFile, chunkHashWriter, chunksHashWriter)

	for i, chunkBytes := range fileChunksData {
		_, err = chunksWriters.Write(chunkBytes)
		if err != nil {
			return nil, metadata, err
		}

		err = hasher.WriteToChallenge(chunkBytes, chunkStartIndex+i)
		if err != nil {
			return nil, metadata, err
		}

		err = hasher.WriteHashToContent(hex.EncodeToString(chunkHashWriter.Sum(nil)), chunkStartIndex+i)
		if err != nil {
			return nil, metadata, err
		}

		metadata.FileBytesLen += len(chunkBytes)
		chunkHashWriter.Reset()
	}

	formData.ChunkHash = hex.EncodeToString(chunksHashWriter.Sum(nil))
	formData.ContentHash = formData.ChunkHash

	if isFinal {

		//fixed shard data's info in last chunk for stream
		formData.ChallengeHash, err = hasher.GetChallengeHash()
		if err != nil {
			return nil, metadata, err
		}
		formData.ContentHash, err = hasher.GetContentHash()
		if err != nil {
			return nil, metadata, err
		}

		//fixed original file's info in last chunk for stream
		formData.ActualHash = fileMeta.ActualHash
		formData.ActualSize = fileMeta.ActualSize

	}

	thumbnailSize := len(thumbnailChunkData)
	if thumbnailSize > 0 {

		uploadThumbnailFile, err := formWriter.CreateFormFile("uploadThumbnailFile", fileMeta.RemoteName+".thumb")
		if err != nil {

			return nil, metadata, err
		}

		thumbnailHash := sha256.New()
		thumbnailWriters := io.MultiWriter(uploadThumbnailFile, thumbnailHash)
		_, err = thumbnailWriters.Write(thumbnailChunkData)
		if err != nil {
			return nil, metadata, err
		}

		formData.ActualThumbSize = fileMeta.ActualThumbnailSize
		formData.ThumbnailContentHash = hex.EncodeToString(thumbnailHash.Sum(nil))

	}

	formData.EncryptedKey = encryptedKey

	err = formWriter.WriteField("connection_id", connectionID)
	if err != nil {
		return nil, metadata, err
	}

	uploadMeta, err := json.Marshal(formData)
	if err != nil {
		return nil, metadata, err
	}

	err = formWriter.WriteField("uploadMeta", string(uploadMeta))
	if err != nil {
		return nil, metadata, err
	}
	metadata.ContentType = formWriter.FormDataContentType()
	metadata.ChunkHash = formData.ChunkHash
	metadata.ChallengeHash = formData.ChallengeHash
	metadata.ContentHash = formData.ContentHash
	metadata.ThumbnailContentHash = formData.ThumbnailContentHash

	return body, metadata, nil
}
