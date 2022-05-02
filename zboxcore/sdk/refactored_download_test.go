package sdk

import (
	"net/http"
	"testing"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/zcncrypto"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/require"
)

const (
	rdtMnemonic  = "critic earn bulb tribe swift soul upgrade endorse hire mesh girl grit enrich until gold chef day head strike like giant today fatigue marine"
	rdtClientID  = "cdc97ce18cfeb235689bb9afeefe3d8e3e1bde2714ec5ecf8df982242bc5c1f8"
	rdtClientKey = "fc7124bfae5ee2f19efe43123891a05038435c3ae5a881d1279aa3f09aec6d037f690176795a8a97b5deecceaa17508c88d7444ef5d31b05e77e88f9cbe0ec1a"
)

type mockClient struct {
	respond func(req *http.Request) (*http.Response, error)
}

func (client *mockClient) Do(req *http.Request) (*http.Response, error) {
	if client.respond == nil {
		return nil, errors.New("function_not_set", "")
	}
	return client.respond(req)
}

func TestDownloadBlock(t *testing.T) {
	zboxutil.Client = &mocks.HttpClient{}

}

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
			name: "Unknown download type",
			sdo: &StreamDownloadOption{
				DownloadType: "unknown",
			},
			wantErr: true,
			errMsg:  InvalidDownloadType,
		},
		{
			name: "Blocks per marker set to 0",
			sdo: &StreamDownloadOption{
				DownloadType:    "vertical",
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
