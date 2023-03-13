package sdk

import (
	"testing"
)

func BenchmarkChunkedUploadFormBuilder(b *testing.B) {

	benchmarks := []struct {
		Name            string
		Size            int64
		Hasher          Hasher
		ChunkSize       int
		EncryptOnUpload bool
	}{
		{Name: "1M 1K", Size: MB * 1, ChunkSize: MB * 1, EncryptOnUpload: false, Hasher: CreateHasher(KB * 1)},
		{Name: "1M 64K", Size: MB * 1, ChunkSize: KB * 64, EncryptOnUpload: false, Hasher: CreateHasher(KB * 64)},

		{Name: "10M 64K", Size: MB * 10, ChunkSize: KB * 64, EncryptOnUpload: false, Hasher: CreateHasher(KB * 64)},
		{Name: "10M 6M", Size: MB * 10, ChunkSize: MB * 6, EncryptOnUpload: false, Hasher: CreateHasher(KB * 64)},
		// {Name: "10M 64K NoHash", Size: MB * 10, ChunkSize: KB * 64, EncryptOnUpload: false, Hasher: &nopeHasher{}},
		// {Name: "10M 6M  NoHash", Size: MB * 10, ChunkSize: MB * 6, EncryptOnUpload: false, Hasher: &nopeHasher{}},

		{Name: "100M 64K", Size: MB * 100, ChunkSize: KB * 64, EncryptOnUpload: false, Hasher: CreateHasher(KB * 64)},
		{Name: "100M 6M", Size: MB * 100, ChunkSize: MB * 6, EncryptOnUpload: false, Hasher: CreateHasher(KB * 64)},
		// {Name: "100M 64K NoHash", Size: MB * 100, ChunkSize: KB * 64, EncryptOnUpload: false, Hasher: &nopeHasher{}},
		// {Name: "100M 6M  NoHash", Size: MB * 100, ChunkSize: MB * 6, EncryptOnUpload: false, Hasher: &nopeHasher{}},

		{Name: "500M 64K", Size: MB * 500, ChunkSize: KB * 64, EncryptOnUpload: false, Hasher: CreateHasher(KB * 64)},
		{Name: "500M 6M", Size: MB * 500, ChunkSize: MB * 6, EncryptOnUpload: false, Hasher: CreateHasher(KB * 64)},
		// {Name: "500M 64K NoHash", Size: MB * 500, ChunkSize: KB * 64, EncryptOnUpload: false, Hasher: &nopeHasher{}},
		// {Name: "500M 6M  NoHash", Size: MB * 500, ChunkSize: MB * 6, EncryptOnUpload: false, Hasher: &nopeHasher{}},

		{Name: "1G 64K", Size: GB * 1, ChunkSize: KB * 64, EncryptOnUpload: false, Hasher: CreateHasher(KB * 64)},
		{Name: "1G 60M", Size: GB * 1, ChunkSize: MB * 60, EncryptOnUpload: false, Hasher: CreateHasher(KB * 64)},
		// {Name: "1G 64K NoHash", Size: GB * 1, ChunkSize: KB * 64, EncryptOnUpload: false, Hasher: &nopeHasher{}},
		// {Name: "1G 60M NoHash", Size: GB * 1, ChunkSize: MB * 60, EncryptOnUpload: false, Hasher: &nopeHasher{}},
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
			}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {

				builder := CreateChunkedUploadFormBuilder()

				isFinal := false

				hasher := CreateHasher(getShardSize(fileMeta.ActualSize, 1, false))
				for chunkIndex := 0; ; chunkIndex++ {
					begin := int64(chunkIndex * bm.ChunkSize)
					end := int64(chunkIndex*bm.ChunkSize + bm.ChunkSize)
					if end > bm.Size {
						end = bm.Size
						isFinal = true
					}

					fileBytes := buf[begin:end]

					_, _, err := builder.Build(fileMeta, hasher, "connectionID", int64(bm.ChunkSize), chunkIndex, chunkIndex, isFinal, "", [][]byte{fileBytes}, nil)
					if err != nil {
						b.Fatal(err)
						return
					}

					if isFinal {
						break
					}
				}
			}
		})
	}

}
