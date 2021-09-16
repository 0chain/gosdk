package sdk

import (
	"testing"

	"github.com/0chain/gosdk/zboxcore/fileref"
)

type nopeHasher struct {
}

// GetFileHash get file hash
func (h *nopeHasher) GetFileHash() (string, error) {
	return "", nil
}

// WriteToFile write bytes to file hasher
func (h *nopeHasher) WriteToFile(buf []byte, chunkIndex int) error {
	return nil
}

// GetChallengeHash get challenge hash
func (h *nopeHasher) GetChallengeHash() (string, error) {
	return "", nil
}

// WriteToChallenge write bytes to challenge hasher
func (h *nopeHasher) WriteToChallenge(buf []byte, chunkIndex int) error {
	return nil
}

// GetContentHash get content hash
func (h *nopeHasher) GetContentHash() (string, error) {
	return "", nil
}

// WriteHashToContent write hash leaf to content hasher
func (h *nopeHasher) WriteHashToContent(hash string, chunkIndex int) error {
	return nil
}

func BenchmarkChunkedUploadFormBuilder(b *testing.B) {

	KB := 1024
	MB := 1024 * KB
	GB := 1024 * MB

	benchmarks := []struct {
		Name            string
		Size            int
		Hasher          Hasher
		ChunkSize       int
		EncryptOnUpload bool
	}{
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
				Attributes: fileref.Attributes{},
			}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {

				builder := CreateChunkedUploadFormBuilder()

				isFinal := false

				hasher := CreateHasher(bm.ChunkSize)
				for chunkIndex := 0; ; chunkIndex++ {
					begin := chunkIndex * bm.ChunkSize
					end := chunkIndex*bm.ChunkSize + bm.ChunkSize
					if end > bm.Size {
						end = bm.Size
						isFinal = true
					}

					fileBytes := buf[begin:end]

					_, _, err := builder.Build(fileMeta, hasher, "connectionID", int64(bm.ChunkSize), chunkIndex, isFinal, "", fileBytes, nil)
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
