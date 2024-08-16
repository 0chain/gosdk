package sdk

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/0chain/gosdk/dev"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type nopeChunkedUploadProgressStorer struct {
	up *UploadProgress
}

func (nope *nopeChunkedUploadProgressStorer) Load(id string) *UploadProgress {
	return nope.up
}

func (nope *nopeChunkedUploadProgressStorer) Save(up UploadProgress) {
	nope.up = &up
}

func (nope *nopeChunkedUploadProgressStorer) Remove(id string) error {
	nope.up = nil
	return nil
}

func (nope *nopeChunkedUploadProgressStorer) Update(id string, chunkIndex int, upMask zboxutil.Uint128) {
}

func generateRandomBytes(n int64) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil
	}

	return b
}

func BenchmarkChunkedUpload(b *testing.B) {

	SetLogFile("cmdlog.log", false)

	logger.Logger.SetLevel(2)

	server := dev.NewBlobberServer(nil)
	defer server.Close()

	benchmarks := []struct {
		Name            string
		Size            int64
		ChunkSize       int
		EncryptOnUpload bool
	}{
		{Name: "1M 1K", Size: MB * 1, ChunkSize: KB * 1, EncryptOnUpload: false},
		{Name: "1M 64K", Size: MB * 1, ChunkSize: KB * 64, EncryptOnUpload: false},

		{Name: "10M 64K", Size: MB * 10, ChunkSize: KB * 64, EncryptOnUpload: false},
		{Name: "10M 6M", Size: MB * 10, ChunkSize: MB * 6, EncryptOnUpload: false},

		{Name: "100M 64K", Size: MB * 100, ChunkSize: KB * 64, EncryptOnUpload: false},
		{Name: "100M 6M", Size: MB * 100, ChunkSize: MB * 6, EncryptOnUpload: false},

		{Name: "500M 64K", Size: MB * 500, ChunkSize: KB * 64, EncryptOnUpload: false},
		{Name: "500M 6M", Size: MB * 500, ChunkSize: MB * 6, EncryptOnUpload: false},

		// {Name: "1G 64K", Size: GB * 1, ChunkSize: KB * 64, EncryptOnUpload: false},
		// {Name: "1G 60M", Size: GB * 1, ChunkSize: MB * 60, EncryptOnUpload: false},
	}

	for n, bm := range benchmarks {
		b.Run(bm.Name, func(b *testing.B) {

			buf := generateRandomBytes(bm.Size)

			b.ResetTimer()

			a := &Allocation{
				ID:           "1a0190c411f3e742c881b7b84c964dc1bb435d459bd3beca74a6c0ae8ececd92",
				Tx:           "1a0190c411f3e742c881b7b84c964dc1bb435d459bd3beca74a6c0ae8ececd92",
				DataShards:   2,
				ParityShards: 1,
				ctx:          context.TODO(),
			}
			a.fullconsensus, a.consensusThreshold = a.getConsensuses()
			for i := 0; i < (a.DataShards + a.ParityShards); i++ {

				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      fmt.Sprintf("blobber_%v_%v_", n, i),
					Baseurl: server.URL,
				})
			}

			for i := 0; i < b.N; i++ {
				name := strings.Replace(bm.Name, " ", "_", -1)

				fileName := "test_" + name + ".txt"

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
				}

				chunkedUpload, err := CreateChunkedUpload(a.ctx, "/tmp", a, fileMeta, reader, false, false, false, zboxutil.NewConnectionId())
				if err != nil {
					b.Fatal(err)
					return
				}
				chunkedUpload.progressStorer = &nopeChunkedUploadProgressStorer{}

				err = chunkedUpload.Start()
				if err != nil {
					b.Fatal(err)
					return
				}

			}
		})
	}
}
