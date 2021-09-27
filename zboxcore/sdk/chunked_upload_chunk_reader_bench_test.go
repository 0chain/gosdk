package sdk

import (
	"bytes"
	"testing"

	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
)

func BenchmarkChunkedUploadChunkReader(b *testing.B) {

	KB := 1024
	MB := 1024 * KB
	//	GB := 1024 * MB

	benchmarks := []struct {
		Name string
		Size int

		ChunkSize       int
		DataShards      int
		ParityShards    int
		EncryptOnUpload bool
	}{
		{Name: "10M 64K 2+1", Size: MB * 10, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		{Name: "10M 64K 10+1", Size: MB * 10, ChunkSize: KB * 64, DataShards: 10, ParityShards: 3, EncryptOnUpload: false},
		{Name: "10M 6M 2+1", Size: MB * 10, ChunkSize: MB * 6, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		{Name: "10M 6M 10+1", Size: MB * 10, ChunkSize: MB * 6, DataShards: 10, ParityShards: 3, EncryptOnUpload: false},

		{Name: "100M 64K 2+1", Size: MB * 100, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		{Name: "100M 64K 10+1", Size: MB * 100, ChunkSize: KB * 64, DataShards: 10, ParityShards: 3, EncryptOnUpload: false},
		{Name: "100M 6M 2+1", Size: MB * 100, ChunkSize: MB * 6, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		{Name: "100M 6M 10+1", Size: MB * 100, ChunkSize: MB * 6, DataShards: 10, ParityShards: 3, EncryptOnUpload: false},

		{Name: "500M 64K 2+1", Size: MB * 500, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		{Name: "500M 64K 10+1", Size: MB * 500, ChunkSize: KB * 64, DataShards: 10, ParityShards: 3, EncryptOnUpload: false},
		{Name: "500M 6M 2+1", Size: MB * 500, ChunkSize: MB * 6, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		{Name: "500M 6M 10+1", Size: MB * 500, ChunkSize: MB * 6, DataShards: 10, ParityShards: 3, EncryptOnUpload: false},

		// {Name: "1G 64K 2+1", Size: GB * 1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// {Name: "1G 64K 10+1", Size: GB * 1, ChunkSize: KB * 64, DataShards: 10, ParityShards: 3, EncryptOnUpload: false},
		// {Name: "1G 60M 2+1", Size: GB * 1, ChunkSize: MB * 60, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// {Name: "1G 60M 10+1", Size: GB * 1, ChunkSize: MB * 60, DataShards: 10, ParityShards: 3, EncryptOnUpload: false},
	}

	for _, bm := range benchmarks {
		b.Run(bm.Name, func(b *testing.B) {

			buf := generateRandomBytes(bm.Size)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {

				uploadMask := zboxutil.NewUint128(1).Lsh(uint64(bm.DataShards + bm.ParityShards)).Sub64(1)
				erasureEncoder, _ := reedsolomon.New(bm.DataShards, bm.ParityShards, reedsolomon.WithAutoGoroutines(bm.ChunkSize))
				encscheme := encryption.NewEncryptionScheme()
				_, err := encscheme.Initialize("BenchmarkChunkReader")
				if err != nil {
					b.Fatal(err)
				}

				encscheme.InitForEncryption("filetype:audio")
				reader, err := createChunkReader(bytes.NewReader(buf), int64(bm.ChunkSize), bm.DataShards, bm.EncryptOnUpload, uploadMask, erasureEncoder, encscheme, CreateHasher(bm.ChunkSize))
				if err != nil {
					b.Fatal(err)
				}
				for {
					c, err := reader.Next()
					if err != nil {
						b.Fatal(err)
					}

					if c.IsFinal {
						break
					}
				}

			}
		})
	}

}
