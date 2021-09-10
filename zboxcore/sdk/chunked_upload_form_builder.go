package sdk

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
)

// FormBuilder build form data for uploading
type FormBuilder interface {
	// build form data
	Build(fileMeta *FileMeta, hasher Hasher, connectionID string, chunkSize int64, chunkIndex int, isFinal bool, encryptedKey string, fileBytes, thumbnailBytes []byte) (*bytes.Buffer, FormMetadata, error)
}

type FormMetadata struct {
	ContentType          string
	ChunkHash            string
	ChallengeHash        string
	ContentHash          string
	ThumbnailContentHash string
}

func createFormBuilder() FormBuilder {
	return &formBuilder{}
}

type formBuilder struct {
}

func (b *formBuilder) Build(fileMeta *FileMeta, hasher Hasher, connectionID string, chunkSize int64, chunkIndex int, isFinal bool, encryptedKey string, fileBytes, thumbnailBytes []byte) (*bytes.Buffer, FormMetadata, error) {

	metadata := FormMetadata{}

	if len(fileBytes) == 0 {
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

		MimeType:   fileMeta.MimeType,
		Attributes: fileMeta.Attributes,

		IsFinal:      isFinal,
		ChunkSize:    chunkSize,
		ChunkIndex:   chunkIndex,
		UploadOffset: chunkSize * int64(chunkIndex),
	}

	formWriter := multipart.NewWriter(body)
	defer formWriter.Close()

	uploadFile, err := formWriter.CreateFormFile("uploadFile", formData.Filename)
	if err != nil {
		return nil, metadata, err
	}

	chunkHashWriter := sha1.New()
	chunkWriters := io.MultiWriter(uploadFile, chunkHashWriter)

	chunkWriters.Write(fileBytes)

	formData.ChunkHash = hex.EncodeToString(chunkHashWriter.Sum(nil))
	formData.ContentHash = formData.ChunkHash

	hasher.WriteToChallenge(fileBytes, chunkIndex)
	hasher.WriteHashToContent(formData.ChunkHash, chunkIndex)

	if isFinal {

		//fixed shard data's info in last chunk for stream
		formData.MerkleRoot, err = hasher.GetChallengeHash()
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

	thumbnailSize := len(thumbnailBytes)
	if thumbnailSize > 0 {

		uploadThumbnailFile, err := formWriter.CreateFormFile("uploadThumbnailFile", fileMeta.RemoteName+".thumb")
		if err != nil {

			return nil, metadata, err
		}

		thumbnailHash := sha1.New()
		thumbnailWriters := io.MultiWriter(uploadThumbnailFile, thumbnailHash)
		thumbnailWriters.Write(thumbnailBytes)

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
	metadata.ChallengeHash = formData.MerkleRoot
	metadata.ContentHash = formData.ContentHash
	metadata.ThumbnailContentHash = formData.ThumbnailContentHash

	return body, metadata, nil
}
