package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"testing"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
	"github.com/stretchr/testify/require"
)

const (
	rdtMnemonic  = "critic earn bulb tribe swift soul upgrade endorse hire mesh girl grit enrich until gold chef day head strike like giant today fatigue marine"
	rdtClientID  = "cdc97ce18cfeb235689bb9afeefe3d8e3e1bde2714ec5ecf8df982242bc5c1f8"
	rdtClientKey = "fc7124bfae5ee2f19efe43123891a05038435c3ae5a881d1279aa3f09aec6d037f690176795a8a97b5deecceaa17508c88d7444ef5d31b05e77e88f9cbe0ec1a"
	blockSize    = 65536
)

func TestGetDstorageFileReader(t *testing.T) {
	type input struct {
		name       string
		sdo        *StreamDownloadOption
		ref        *ORef
		wantErr    bool
		errMsg     string
		allocation *Allocation
	}

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  rdtClientID,
		ClientKey: rdtClientKey,
		Mnemonic:  rdtMnemonic,
	}

	encscheme := encryption.NewEncryptionScheme()
	mnemonic := zclient.GetClient().Mnemonic
	_, err := encscheme.Initialize(mnemonic)
	require.Nil(t, err)

	encscheme.InitForEncryption("filetype:audio")
	encryptedKey := encscheme.GetEncryptedKey()

	tests := []input{
		{
			name: "Blocks per marker set to 0",
			sdo: &StreamDownloadOption{
				BlocksPerMarker: 0,
			},
			wantErr: true,
			errMsg:  InvalidBlocksPerMarker,
		},
		{
			name:       "Wrong encrypted key",
			allocation: &Allocation{},
			sdo: &StreamDownloadOption{
				BlocksPerMarker: 1,
			},
			ref: &ORef{
				SimilarField: SimilarField{EncryptedKey: "wrong encrypted key"},
			},
			wantErr: true,
		},
		{
			name:       "Ok",
			allocation: &Allocation{},
			sdo: &StreamDownloadOption{
				BlocksPerMarker: 1,
			},
			ref: &ORef{
				SimilarField: SimilarField{EncryptedKey: encryptedKey},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := GetDStorageFileReader(test.allocation, test.ref, test.sdo)
			if test.wantErr {
				require.NotNil(t, err)
				require.Contains(t, err.Error(), test.errMsg)
				return
			}

			require.Nil(t, err)
		})
	}
}

func TestSetOffset(t *testing.T) {
	s := StreamDownload{}
	s.SetOffset(65536)
	require.EqualValues(t, s.offset, 65536)
}

func TestGetBlobberStartingIdx(t *testing.T) {
	type input struct {
		name string
		sd   StreamDownload
		want int
	}

	tests := []input{
		{
			name: "Test#1",
			sd: StreamDownload{
				offset:             65536,
				effectiveBlockSize: 65536,
				dataShards:         3,
			},

			want: 1,
		},
		{
			name: "Test#2",
			sd: StreamDownload{
				offset:             655360,
				effectiveBlockSize: 65536,
				dataShards:         3,
			},
			want: 1,
		},
		{
			name: "Test#3",
			sd: StreamDownload{
				offset:             655360 - 65536,
				effectiveBlockSize: 65536,
				dataShards:         3,
			},
			want: 0,
		},
		{
			name: "Test#4",
			sd: StreamDownload{
				offset:             655360 + 65536,
				effectiveBlockSize: 65536,
				dataShards:         3,
			},
			want: 2,
		},
		{
			name: "Test#5",
			sd: StreamDownload{
				offset:             655360 + 2719,
				effectiveBlockSize: 65536 - 272, // test for when file is encrypted
				dataShards:         3,
			},
			want: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.sd.getBlobberStartingIdx()
			require.Equal(t, test.want, got)
		})
	}
}

func TestGetBlobberEndIdx(t *testing.T) {
	type input struct {
		name    string
		sd      StreamDownload
		wantIdx int
		size    int
	}

	tests := []input{
		{
			name: "Test#1",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				dataShards:         3,
			},
			wantIdx: 0,
			size:    655360,
		},
		{
			name: "Test#2",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				dataShards:         3,
			},
			wantIdx: 1,
			size:    655360 + 1,
		},
		{
			name: "Test#3",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				dataShards:         3,
			},
			wantIdx: 2,
			size:    655360 - 65536,
		},
		{
			name: "Test#4",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				dataShards:         4,
			},
			wantIdx: 0,
			size:    655360 - 65536,
		},
		{
			name: "Test#5",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				dataShards:         35,
			},
			wantIdx: 9,
			size:    655360,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.sd.getBlobberEndIdx(test.size)
			require.Equal(t, test.wantIdx, got)
		})
	}
}

func TestGetDataOffset(t *testing.T) {
	type input struct {
		name string
		sd   StreamDownload
		want int
	}

	tests := []input{
		{
			name: "Test#1",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				offset:             0,
			},
			want: 0,
		},
		{
			name: "Test#2",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				offset:             1,
			},
			want: 1,
		},
		{
			name: "Test#3",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				offset:             65536 + 1,
			},
			want: 1,
		},
		{
			name: "Test#4",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				offset:             65536 + 65535,
			},
			want: 65535,
		},
		{
			name: "Test#5",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				offset:             65536 + 65536 + 2,
			},
			want: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.sd.getDataOffset()
			require.Equal(t, test.want, got)
		})
	}
}

func TestGetChunksRequired(t *testing.T) {
	type input struct {
		name          string
		sd            StreamDownload
		startingIndex int
		size          int
		want          int
	}

	tests := []input{
		{
			name: "Test#1",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				effectiveChunkSize: 65536 * 3,
				dataShards:         3,
			},
			size:          65536 + 10,
			startingIndex: 2,
			want:          2,
		},
		{
			name: "Test#2",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				effectiveChunkSize: 65536 * 3,
				dataShards:         3,
			},
			size:          65536 + 10,
			startingIndex: 1,
			want:          1,
		},
		{
			name: "Test#3",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				effectiveChunkSize: 65536 * 3,
				dataShards:         3,
			},
			size:          65536 + 10,
			startingIndex: 0,
			want:          1,
		},
		{
			name: "Test#4",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				effectiveChunkSize: 65536 * 3,
				dataShards:         3,
			},
			size:          655360,
			startingIndex: 2,
			want:          4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.sd.getChunksRequired(test.startingIndex, test.size)
			require.Equal(t, test.want, got)
		})
	}
}

func TestGetEndOffsetChunkIndex(t *testing.T) {
	type input struct {
		name string
		sd   StreamDownload
		size int
		want int
	}

	tests := []input{
		{
			name: "Test#1",
			sd: StreamDownload{
				offset:             0,
				fileSize:           6553600,
				effectiveBlockSize: 65536,
				dataShards:         3,
			},
			size: 655360,
			want: 4,
		},
		{
			name: "Test#2",
			sd: StreamDownload{
				offset:             0,
				fileSize:           6553600,
				effectiveBlockSize: 65536,
				dataShards:         3,
			},
			size: 655360 + 65536*2,
			want: 4,
		},
		{
			name: "Test#3",
			sd: StreamDownload{
				offset:             65536 * 2,
				fileSize:           6553600,
				effectiveBlockSize: 65536,
				dataShards:         3,
			},
			size: 655360 + 65536,
			want: 5,
		},
		{
			name: "Test#4",
			sd: StreamDownload{
				fileSize:           6553600,
				effectiveBlockSize: 65536,
				dataShards:         3,
			},
			size: 65536,
			want: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.sd.getEndOffsetChunkIndex(test.size)
			require.Equal(t, test.want, got)
		})
	}
}

func TestGetStartOffsetChunkIndex(t *testing.T) {
	type input struct {
		name string
		sd   StreamDownload
		want int
	}

	tests := []input{
		{
			name: "Test#1",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				dataShards:         3,
			},
			want: 1,
		},
		{
			name: "Test#2",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				dataShards:         3,
				offset:             655360,
			},
			want: 4,
		},
		{
			name: "Test#3",
			sd: StreamDownload{
				effectiveBlockSize: 65536,
				dataShards:         3,
				offset:             65536*3 - 1,
			},
			want: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.sd.getStartOffsetChunkIndex()
			require.Equal(t, test.want, got)
		})
	}
}

func TestReadError(t *testing.T) {
	type input struct {
		name string
		sd   StreamDownload
		b    []byte
	}

	tests := []input{
		{
			name: "Closed Reader",
			sd: StreamDownload{
				opened: false,
			},
		},
		{
			name: "EOF reached",
			sd: StreamDownload{
				opened:     true,
				eofReached: true,
			},
		},
		{
			name: "Offset greater than file size",
			sd: StreamDownload{
				opened:     true,
				eofReached: true,
				offset:     655360,
				fileSize:   655360,
			},
		},
		{
			name: "Exceeding failed blobbers",
			sd: StreamDownload{
				opened:   true,
				fileSize: 1,
				failedBlobbers: map[int]*blockchain.StorageNode{
					1: {},
					2: {},
				},
				parityShards: 1,
			},
		},
		{
			name: "Want size 0",
			sd: StreamDownload{
				opened:   true,
				fileSize: 10,
			},
			b: make([]byte, 0),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.sd.Read(test.b)
			require.NotNil(t, err)
		})
	}
}

func getErasureEncodedData(t *testing.T, blobbers []*blockchain.StorageNode, fileSize, data, parity int) map[string][]byte {
	p := make([]byte, fileSize)
	n, err := rand.Read(p)
	require.Nil(t, err)
	require.Equal(t, n, fileSize)

	chunkSize := data * blockSize

	numberOfChunks := int(math.Ceil(float64(fileSize) / float64(chunkSize)))
	resultantFileSize := numberOfChunks * chunkSize
	numPad := resultantFileSize - fileSize

	r := io.MultiReader(bytes.NewReader(p), bytes.NewReader(make([]byte, numPad)))
	dataMap := make(map[string][]byte, data+parity)

	for i := 0; i < numberOfChunks; i++ {
		chunkdata := make([]byte, chunkSize)
		n, err := r.Read(chunkdata)
		require.Nil(t, err)
		require.Equal(t, n, chunkSize)

		encoder, err := reedsolomon.New(data, parity)
		require.Nil(t, err)

		splittedData, err := encoder.Split(chunkdata)
		require.Nil(t, err)

		err = encoder.Encode(splittedData)
		require.Nil(t, err)

		for i, d := range splittedData {
			blobber := blobbers[i]
			dataMap[blobber.ID] = append(dataMap[blobber.ID], d...)
		}
	}

	return dataMap
}

type mockClient struct {
	data map[string][]byte
}

func (client *mockClient) Do(req *http.Request) (*http.Response, error) {
	if client.data == nil {
		return nil, errors.New("data_not_set", "")
	}

	rmStr := req.Header.Get("X-Read-Marker")
	rm := new(marker.ReadMarker)
	err := json.Unmarshal([]byte(rmStr), rm)
	if err != nil {
		return nil, err
	}

	blData := client.data[rm.BlobberID]
	if blData == nil {
		return nil, errors.New("data_not_set", fmt.Sprintf("Data not set for blobber %s", rm.BlobberID))
	}

	startBlockStr := req.Header.Get("X-Block-Num")
	numBlocksStr := req.Header.Get("X-Num-Blocks")

	startBlock := 0
	if startBlockStr != "" {
		startBlock, err = strconv.Atoi(startBlockStr)
		if err != nil {
			return nil, err
		}
	}

	numBlocks := 1
	if numBlocksStr != "" {
		numBlocks, err = strconv.Atoi(numBlocksStr)
		if err != nil {
			return nil, err
		}
	}

	totalBlocks := len(blData) / blockSize
	if startBlock < 0 || numBlocks < 0 || startBlock >= totalBlocks {
		return nil, errors.New("invalid_block_num", "")
	}

	offset := startBlock * blockSize
	limit := offset + numBlocks*blockSize
	limit = int(math.Min(float64(len(blData)), float64(limit)))

	return &http.Response{
		StatusCode: http.StatusOK,
		Body: func() io.ReadCloser {
			return io.NopCloser(bytes.NewReader(blData[offset:limit]))
		}(),
	}, nil
}

func getBlobbers(n int) (blobbers []*blockchain.StorageNode) {
	for i := 0; i < n; i++ {
		blobbers = append(blobbers, &blockchain.StorageNode{
			ID:      fmt.Sprintf("blobberID#%d", i),
			Baseurl: "https://localhost/blobber" + fmt.Sprint(i),
		})
	}

	return
}

func TestDownloadBlock(t *testing.T) {
	type input struct {
		name           string
		client         *mockClient
		sd             StreamDownload
		wantSize       int
		wantDataLength int
		wantErr        bool
		errMsg         string
	}

	tests := []input{
		{
			name:   "OK Download",
			client: &mockClient{},
			sd: StreamDownload{
				dataShards:         10,
				parityShards:       5,
				fileSize:           65536 * 10 * 10,
				effectiveBlockSize: 65536,
				effectiveChunkSize: 65536 * 10,
				opened:             true,
				blobbers:           getBlobbers(15),
				blocksPerMarker:    5,
				retry:              3,
				ctx:                context.Background(),
			},
			wantSize:       65536 * 10 * 5,
			wantDataLength: 65536 * 10 * 5,
		},
	}
	for _, test := range tests {
		test.client.data = getErasureEncodedData(
			t, test.sd.blobbers, int(test.sd.fileSize), test.sd.dataShards, test.sd.parityShards)

		zboxutil.Client = test.client

		t.Run(test.name, func(t *testing.T) {
			t.Logf("Running test: %s", test.name)
			zboxutil.Client = test.client
			var (
				data []byte
				err  error
			)

			data, err = test.sd.getData(test.wantSize)

			if test.wantErr {
				require.NotNil(t, err)
				require.Contains(t, err.Error(), test.errMsg)
				require.Len(t, data, test.wantDataLength)
				return
			}

			require.Nil(t, err)
			require.Len(t, data, test.wantDataLength)
		})
	}

}

type reconstructionMockClient struct {
	mockClient
}

func (mc reconstructionMockClient) NilData(indexes []string) {
	if mc.data == nil {
		return
	}

	for _, i := range indexes {
		mc.data[i] = nil
	}
}

func TestReconstruction(t *testing.T) {
	type input struct {
		name               string
		sd                 StreamDownload
		wantErr            bool
		errMsg             string
		client             *reconstructionMockClient
		offsetBlock        int
		failedBlobbersList []int
	}

	tests := []input{
		{
			name:   "OK Reconstruction",
			client: &reconstructionMockClient{},
			sd: StreamDownload{
				dataShards:         10,
				parityShards:       5,
				fileSize:           65536 * 10 * 10,
				effectiveBlockSize: 65536,
				effectiveChunkSize: 65536 * 10,
				opened:             true,
				blobbers:           getBlobbers(15),
				blocksPerMarker:    5,
				retry:              3,
				ctx:                context.Background(),
			},
			failedBlobbersList: []int{0, 3, 5},
		},
		{
			name:   "Fail Reconstruction",
			client: &reconstructionMockClient{},
			sd: StreamDownload{
				dataShards:         10,
				parityShards:       5,
				fileSize:           65536 * 10 * 10,
				effectiveBlockSize: 65536,
				effectiveChunkSize: 65536 * 10,
				opened:             true,
				blobbers:           getBlobbers(15),
				blocksPerMarker:    5,
				retry:              3,
				ctx:                context.Background(),
			},
			failedBlobbersList: []int{0, 1, 3, 5, 7, 9},
			wantErr:            true,
			errMsg:             ErrNoRequiredShards.Code,
		},
	}

	for _, test := range tests {
		test.client.data = getErasureEncodedData(t,
			test.sd.blobbers, int(test.sd.fileSize), test.sd.dataShards, test.sd.parityShards)

		var blList []string
		for _, i := range test.failedBlobbersList {
			blb := test.sd.blobbers[i]
			blList = append(blList, blb.ID)
		}

		test.client.NilData(blList)
		zboxutil.Client = test.client

		results := make([]*blobberStreamDownloadRequest, test.sd.dataShards+test.sd.parityShards)
		var count int
		for i, blb := range test.sd.blobbers {
			if count == 10 {
				break
			}

			bsdl := &blobberStreamDownloadRequest{
				blobberIdx:      i,
				blobberID:       blb.ID,
				blobberUrl:      blb.Baseurl,
				offsetBlock:     test.offsetBlock,
				blocksPerMarker: test.sd.blocksPerMarker,
			}

			limit := test.sd.blocksPerMarker * blockSize
			offset := test.offsetBlock

			blobberData := test.client.data[blb.ID]
			if blobberData != nil {
				for start := offset; start < limit; start = start + blockSize {
					bsdl.result.data = append(bsdl.result.data, blobberData[start:start+blockSize])
				}
			} else {
				bsdl.result.err = errors.New("No data", "")
			}

			results[i] = bsdl
			count++
		}

		t.Run(test.name, func(t *testing.T) {
			err := test.sd.reconstruct(results, len(test.failedBlobbersList), test.offsetBlock, test.sd.blocksPerMarker)

			if test.wantErr {
				require.NotNil(t, err)
				require.Contains(t, err.Error(), test.errMsg)
				return
			}

			require.Nil(t, err)
		})
	}
}
