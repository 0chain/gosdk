package sdk

import (
	"testing"

	"github.com/0chain/gosdk/zboxcore/fileref"
)

func BenchmarkFormBuilder(b *testing.B) {

	KB := 1024
	MB := 1024 * KB
	GB := 1024 * MB

	benchmarks := []struct {
		Name string
		Size int

		ChunkSize       int
		EncryptOnUpload bool
	}{
		{Name: "10M 64K", Size: MB * 10, ChunkSize: KB * 64, EncryptOnUpload: false},
		{Name: "10M 64K", Size: MB * 10, ChunkSize: KB * 64, EncryptOnUpload: false},
		{Name: "10M 6M", Size: MB * 10, ChunkSize: MB * 6, EncryptOnUpload: false},
		{Name: "10M 6M", Size: MB * 10, ChunkSize: MB * 6, EncryptOnUpload: false},

		{Name: "100M 64K", Size: MB * 100, ChunkSize: KB * 64, EncryptOnUpload: false},
		{Name: "100M 64K", Size: MB * 100, ChunkSize: KB * 64, EncryptOnUpload: false},
		{Name: "100M 6M", Size: MB * 100, ChunkSize: MB * 6, EncryptOnUpload: false},
		{Name: "100M 6M", Size: MB * 100, ChunkSize: MB * 6, EncryptOnUpload: false},

		{Name: "500M 64K", Size: MB * 500, ChunkSize: KB * 64, EncryptOnUpload: false},
		{Name: "500M 64K", Size: MB * 500, ChunkSize: KB * 64, EncryptOnUpload: false},
		{Name: "500M 6M", Size: MB * 500, ChunkSize: MB * 6, EncryptOnUpload: false},
		{Name: "500M 6M", Size: MB * 500, ChunkSize: MB * 6, EncryptOnUpload: false},

		{Name: "1G 64K", Size: GB * 1, ChunkSize: KB * 64, EncryptOnUpload: false},
		{Name: "1G 64K", Size: GB * 1, ChunkSize: KB * 64, EncryptOnUpload: false},
		{Name: "1G 60M", Size: GB * 1, ChunkSize: MB * 60, EncryptOnUpload: false},
		{Name: "1G 60M", Size: GB * 1, ChunkSize: MB * 60, EncryptOnUpload: false},
	}

	for _, bm := range benchmarks {
		b.Run(bm.Name, func(b *testing.B) {

			buf := generateRandomBytes(bm.Size)
			fileMeta := &FileMeta{
				Path:       "/tmp/" + bm.Name,
				ActualSize: int64(bm.Size),

				MimeType:   "plain/text",
				RemoteName: "/test.txt",
				RemotePath: "/test.txt",
				Attributes: fileref.Attributes{},
			}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {

				builder := createFormBuilder()

				isFinal := false

				for chunkIndex := 0; ; chunkIndex++ {
					begin := chunkIndex * bm.ChunkSize
					end := chunkIndex*bm.ChunkSize + bm.ChunkSize
					if end > bm.Size {
						end = bm.Size
						isFinal = true
					}

					fileBytes := buf[begin:end]

					builder.Build(fileMeta, createHasher(bm.ChunkSize), "connectionID", int64(bm.ChunkSize), chunkIndex, isFinal, "", fileBytes, nil)

					if isFinal {
						break
					}
				}
			}
		})
	}

}
