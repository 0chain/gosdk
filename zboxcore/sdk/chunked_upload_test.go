package sdk

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"testing/fstest"

	"github.com/0chain/gosdk/dev"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
)

type nopeChunkedUploadProgressStorer struct {
	up *UploadProgress
}

func (nope *nopeChunkedUploadProgressStorer) Load(id string) *UploadProgress {
	return nope.up
}

func (nope *nopeChunkedUploadProgressStorer) Save(up *UploadProgress) {
	nope.up = up
}

func (nope *nopeChunkedUploadProgressStorer) Remove(id string) error {
	nope.up = nil
	return nil
}

type echoClient struct {
}

func (echo *echoClient) Do(req *http.Request) (*http.Response, error) {
	uploadMeta := req.FormValue("uploadMeta")

	var form *UploadFormData
	err := json.Unmarshal([]byte(uploadMeta), &form)

	if err != nil {
		return nil, err
	}

	buf, err := json.Marshal(&UploadResult{
		Filename:   form.Filename,
		Hash:       form.ChunkHash,
		MerkleRoot: form.ChallengeHash,
	})
	if err != nil {
		return nil, err
	}

	w := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(buf)),
	}

	return w, nil

}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil
	}

	return b
}

func chunkedUploadTest(b *testing.B, buf []byte, n int) {

	fileName := "test_" + strconv.Itoa(n) + ".txt"

	m := fstest.MapFS{
		fileName: {
			Data: buf,
		},
	}

	reader, err := m.Open(fileName)

	if err != nil {
		b.Fatal(err)
		return
	}

	fi, _ := reader.Stat()

	fileMeta := FileMeta{
		Path:       "/tmp/" + fileName,
		ActualSize: fi.Size(),

		MimeType:   "plain/text",
		RemoteName: "/test.txt",
		RemotePath: "/test.txt",
		Attributes: fileref.Attributes{},
	}

	a := &Allocation{
		Tx:           "chunkedUploadTest",
		ParityShards: 2,
		DataShards:   2,
	}
	a.ctx, a.ctxCancelF = context.WithCancel(context.Background())

	server := dev.NewBlobberServer()
	defer server.Close()

	for i := 0; i < 2; i++ {

		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      fmt.Sprintf("blobber_%v_%v", n, i),
			Baseurl: server.URL,
		})
	}

	chunkedUpload, err := CreateChunkedUpload("/tmp", a, fileMeta, reader, false)
	if err != nil {
		b.Fatal(err)
		return
	}
	chunkedUpload.progressStorer = &nopeChunkedUploadProgressStorer{}
	chunkedUpload.client = &echoClient{}

	err = chunkedUpload.Start()
	if err != nil {
		b.Fatal(err)
		return
	}

	// for i := 0; ; i++ {

	// 	_, _, _, isFinal, err := chunkedUpload.readNextChunks(i)
	// 	if err != nil {
	// 		b.Fatal(err)
	// 		return
	// 	}

	// 	if isFinal {
	// 		break
	// 	}
	// }
}

func BenchmarkChunkedUpload(b *testing.B) {

	SetLogFile("cmdlog.log", false)

	logger.Logger.SetLevel(2)

	buf := generateRandomBytes(1024)

	for i := 0; i < b.N; i++ {
		chunkedUploadTest(b, buf, i)
	}
}

// func BenchmarkChunkedUpload10M(b *testing.B) {
// 	m := fstest.MapFS{
// 		"hello.txt": {
// 			Data: []byte("hello, world"),
// 		},
// 	}
// }

// func BenchmarkChunkedUpload100M(b *testing.B) {
// 	m := fstest.MapFS{
// 		"hello.txt": {
// 			Data: []byte("hello, world"),
// 		},
// 	}
// }

// func BenchmarkChunkedUpload1G(b *testing.B) {
// 	m := fstest.MapFS{
// 		"hello.txt": {
// 			Data: []byte("hello, world"),
// 		},
// 	}
// }
