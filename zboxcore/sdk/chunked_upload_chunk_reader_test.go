package sdk

import (
	"bytes"
	"math"
	"testing"

	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
	"github.com/stretchr/testify/require"
)

func TestReadChunks(t *testing.T) {
	tests := []struct {
		Name      string
		Size      int64
		ChunkSize int64

		EncryptOnUpload bool
		DataShards      int
		ParityShards    int
	}{

		// size < chunk_size
		{Name: "Size_Less_ChunkSize", Size: KB*64 - KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size == chunk_size
		{Name: "size_Equals_ChunkSize", Size: KB * 64 * 1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size > chunk_size
		{Name: "Size_Greater_ChunkSize", Size: KB*64 + KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},

		// size < datashards * chunk_size
		{Name: "Size_Less_DataShards_x_ChunkSize", Size: KB*64*2 - KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size == datashards * chunk_size
		{Name: "size_Equals_DataShards_x_ChunkSize", Size: KB * 64 * 2, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size > datashards * chunk_size
		{Name: "Size_Greater_DataShards_x_ChunkSize", Size: KB*64*2 + KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},

		// size = 3 * datashards * chunk_size
		{Name: "Size_Less_3_x_DataShards_x_ChunkSize", Size: KB*64*2*3 - KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size = 3 * datashards * chunk_size
		{Name: "Size_Equals_3_x_DataShards_x_ChunkSize", Size: KB * 64 * 2 * 3, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size > 3 * datashards * chunk_size
		{Name: "Size_Greater_3_x_DataShards_x_ChunkSize", Size: KB*64*2*3 + KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},

		// size < chunk_size
		{Name: "Size_Less_ChunkSize_Encrypt", Size: KB*64 - KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size == chunk_size
		{Name: "size_Equals_ChunkSize_Encrypt", Size: KB * 64 * 1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size > chunk_size
		{Name: "Size_Greater_ChunkSize_Encrypt", Size: KB*64 + KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},

		// size < datashards * chunk_size
		{Name: "Size_Less_DataShards_x_ChunkSize_Encrypt", Size: KB*64*2 - KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size == datashards * chunk_size
		{Name: "size_Equals_DataShards_x_ChunkSize_Encrypt", Size: KB * 64 * 2, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size > datashards * chunk_size
		{Name: "Size_Greater_DataShards_x_ChunkSize_Encrypt", Size: KB*64*2 + KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},

		// size = 3 * datashards * chunk_size
		{Name: "Size_Less_3_x_DataShards_x_ChunkSize_Encrypt", Size: KB*64*2*3 - KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size = 3 * datashards * chunk_size
		{Name: "Size_Equals_3_x_DataShards_x_ChunkSize_Encrypt", Size: KB * 64 * 2 * 3, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
		// size > 3 * datashards * chunk_size
		{Name: "Size_Greater_3_x_DataShards_x_ChunkSize_Encrypt", Size: KB*64*2*3 + KB*1, ChunkSize: KB * 64, DataShards: 2, ParityShards: 1, EncryptOnUpload: false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			uploadMask := zboxutil.NewUint128(1).Lsh(uint64(test.DataShards + test.ParityShards)).Sub64(1)
			erasureEncoder, _ := reedsolomon.New(test.DataShards, test.ParityShards, reedsolomon.WithAutoGoroutines(int(test.ChunkSize)))
			encscheme := encryption.NewEncryptionScheme()
			_, err := encscheme.Initialize(test.Name)
			if err != nil {
				t.Fatal(err)
			}
			encscheme.InitForEncryption("filetype:audio")

			buf := generateRandomBytes(test.Size)

			reader, err := createChunkReader(bytes.NewReader(buf), int64(test.Size), int64(test.ChunkSize), test.DataShards, test.EncryptOnUpload, uploadMask, erasureEncoder, encscheme, CreateHasher(int(test.ChunkSize)))

			require := require.New(t)

			lastChunkIndex := 0
			var totalReadSize int64
			var totalFragmentSize int64
			for {
				chunk, err := reader.Next()
				if err != nil {
					t.Fatal(err)
				}

				lastChunkIndex = chunk.Index

				totalReadSize += chunk.ReadSize
				totalFragmentSize += chunk.FragmentSize * int64(test.DataShards)

				if chunk.IsFinal {
					break
				}
			}

			var chunkDataSize int64
			if test.EncryptOnUpload {
				chunkDataSize = test.ChunkSize - 16 - 2*1024
			} else {
				chunkDataSize = test.ChunkSize
			}

			chunkDataSizePerRead := chunkDataSize * int64(test.DataShards)
			chunkNumber := int(math.Ceil(float64(test.Size) / float64(chunkDataSizePerRead)))

			totalSize := test.Size

			if test.EncryptOnUpload {
				totalSize += 16 + 2*1024*int64(test.DataShards)
			}

			require.Equal(chunkNumber, lastChunkIndex+1)
			require.Equal(totalSize, totalFragmentSize)
			require.Equal(totalSize, totalReadSize)
		})
	}
}
