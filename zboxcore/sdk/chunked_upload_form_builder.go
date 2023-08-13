package sdk

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"sync"

	"github.com/0chain/gosdk/zboxcore/client"
	"golang.org/x/crypto/sha3"
)

// ChunkedUploadFormBuilder build form data for uploading
type ChunkedUploadFormBuilder interface {
	// build form data
	Build(
		fileMeta *FileMeta, hasher Hasher, connectionID string,
		chunkSize int64, chunkStartIndex, chunkEndIndex int,
		isFinal bool, encryptedKey string, fileChunksData [][]byte,
		thumbnailChunkData []byte,
	) (*bytes.Buffer, ChunkedUploadFormMetadata, error)
}

// ChunkedUploadFormMetadata upload form metadata
type ChunkedUploadFormMetadata struct {
	FileBytesLen         int
	ThumbnailBytesLen    int
	ContentType          string
	FixedMerkleRoot      string
	ValidationRoot       string
	ThumbnailContentHash string
}

// CreateChunkedUploadFormBuilder create ChunkedUploadFormBuilder instance
func CreateChunkedUploadFormBuilder() ChunkedUploadFormBuilder {
	return &chunkedUploadFormBuilder{}
}

type chunkedUploadFormBuilder struct {
}

func (b *chunkedUploadFormBuilder) Build(
	fileMeta *FileMeta, hasher Hasher, connectionID string,
	chunkSize int64, chunkStartIndex, chunkEndIndex int,
	isFinal bool, encryptedKey string, fileChunksData [][]byte,
	thumbnailChunkData []byte,
) (*bytes.Buffer, ChunkedUploadFormMetadata, error) {

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

	for _, chunkBytes := range fileChunksData {
		_, err = uploadFile.Write(chunkBytes)
		if err != nil {
			return nil, metadata, err
		}

		err = hasher.WriteToFixedMT(chunkBytes)
		if err != nil {
			return nil, metadata, err
		}

		err = hasher.WriteToValidationMT(chunkBytes)
		if err != nil {
			return nil, metadata, err
		}

		metadata.FileBytesLen += len(chunkBytes)
	}

	if isFinal {
		err = hasher.Finalize()
		if err != nil {
			return nil, metadata, err
		}

		var (
			wg      sync.WaitGroup
			errChan = make(chan error, 2)
		)
		wg.Add(2)
		go func() {
			formData.FixedMerkleRoot, err = hasher.GetFixedMerkleRoot()
			if err != nil {
				errChan <- err
			}
			wg.Done()
		}()
		go func() {
			formData.ValidationRoot, err = hasher.GetValidationRoot()
			if err != nil {
				errChan <- err
			}
			wg.Done()
		}()
		wg.Wait()
		close(errChan)
		for err := range errChan {
			return nil, metadata, err
		}

		actualHashSignature, err := client.Sign(fileMeta.ActualHash)
		if err != nil {
			return nil, metadata, err
		}

		validationRootSignature, err := client.Sign(actualHashSignature + formData.ValidationRoot)
		if err != nil {
			return nil, metadata, err
		}

		formData.ActualHash = fileMeta.ActualHash
		formData.ActualFileHashSignature = actualHashSignature
		formData.ValidationRootSignature = validationRootSignature
		formData.ActualSize = fileMeta.ActualSize

	}

	thumbnailSize := len(thumbnailChunkData)
	if thumbnailSize > 0 {

		uploadThumbnailFile, err := formWriter.CreateFormFile("uploadThumbnailFile", fileMeta.RemoteName+".thumb")
		if err != nil {

			return nil, metadata, err
		}

		thumbnailHash := sha3.New256()
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
	metadata.FixedMerkleRoot = formData.FixedMerkleRoot
	metadata.ValidationRoot = formData.ValidationRoot
	metadata.ThumbnailContentHash = formData.ThumbnailContentHash

	return body, metadata, nil
}
