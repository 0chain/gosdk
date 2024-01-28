package sdk

import (
	"encoding/hex"
	"math/rand"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"
)

func TestRemoveFromMask(t *testing.T) {
	req := DownloadRequest{}
	req.maskMu = &sync.Mutex{}
	N := 30
	req.downloadMask = zboxutil.NewUint128(1).Lsh(uint64(30)).Sub64(1)

	require.Equal(t, N, req.downloadMask.CountOnes())

	n := 10
	for i := 0; i < n; i++ {
		req.removeFromMask(uint64(i))
	}

	expected := N - n
	require.Equal(t, expected, req.downloadMask.CountOnes())
}

func TestDecodeEC(t *testing.T) {
	type input struct {
		name        string
		req         *DownloadRequest
		shards      [][]byte
		wantErr     bool
		contentHash string
		errMsg      string
		setup       func(in *input)
	}

	tests := []*input{
		{
			name:    "should be ok",
			wantErr: false,
			setup: func(in *input) {
				req := DownloadRequest{}
				req.datashards = 4
				req.parityshards = 2
				req.effectiveBlockSize = 64 * 1024

				err := req.initEC()
				require.NoError(t, err)

				d, err := getDummyData(64 * 1024 * 4)
				require.NoError(t, err)

				h := sha3.New256()
				n, err := h.Write(d)
				require.NoError(t, err)
				require.Equal(t, len(d), n)
				in.contentHash = hex.EncodeToString(h.Sum(nil))

				shards, err := req.ecEncoder.Split(d)
				require.NoError(t, err)

				err = req.ecEncoder.Encode(shards)
				require.NoError(t, err)

				shards[0] = nil

				in.shards = shards
				in.req = &req
			},
		},
	}

	for _, test := range tests {

		test := test
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(test)
			}

			err := test.req.decodeEC(test.shards)
			if test.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.errMsg)
				return
			} else {
				require.NoError(t, err)
			}

			h := sha3.New256()
			total := 0
			for i := 0; i < test.req.datashards; i++ {
				n, err := h.Write(test.shards[i])
				require.NoError(t, err)
				total += n
			}
			hash := hex.EncodeToString(h.Sum(nil))
			require.Equal(t, test.contentHash, hash)

		})
	}
}

func TestFillShards(t *testing.T) {
	type input struct {
		name          string
		wantErr       bool
		totalBlocks   int
		totalBlobbers int
		blobberIdx    int
		expectedSize  int
		req           *DownloadRequest
		setup         func(in *input)
		shards        [][][]byte
		result        *downloadBlock
	}

	tests := []*input{
		{
			name:    "fill shards ok",
			wantErr: false,
			setup: func(in *input) {
				in.expectedSize = 64 * 1024
				in.totalBlobbers = 4
				in.totalBlocks = 2
				in.blobberIdx = 1
				d, err := getDummyData(in.expectedSize * in.totalBlocks)
				require.NoError(t, err)
				in.req = &DownloadRequest{}
				in.req.maskMu = &sync.Mutex{}
				shards := make([][]byte, in.totalBlocks)
				for i := 0; i < in.totalBlocks; i++ {
					index := i * in.expectedSize
					shards[i] = d[index : index+in.expectedSize]
				}

				in.result = &downloadBlock{
					BlockChunks: shards,
					Success:     true,
					idx:         in.blobberIdx,
				}

				in.shards = make([][][]byte, in.totalBlocks)
				for i := range in.shards {
					in.shards[i] = make([][]byte, in.totalBlobbers)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(test)
			}

			// maskCount := test.req.downloadMask.CountOnes()
			err := test.req.fillShards(test.shards, test.result)
			if test.wantErr {
				require.Error(t, err)
				// require.Equal(t, maskCount-1, test.req.downloadMask.CountOnes())
				return
			}

			require.NoError(t, err)
			for i := 0; i < test.totalBlocks; i++ {
				data := test.shards[i][test.blobberIdx]
				require.Equal(t, test.expectedSize, len(data))
			}

		})
	}
}

func getDummyData(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b) //nolint
	if err != nil {
		return nil, err
	}
	return b, nil
}
