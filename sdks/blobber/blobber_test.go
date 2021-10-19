package blobber

import (
	"net/http"
	"testing"

	"github.com/0chain/gosdk/dev"
	"github.com/0chain/gosdk/dev/blobber"
	"github.com/0chain/gosdk/sdks"
	"github.com/stretchr/testify/require"
)

const (
	zboxWallet = "{\"client_id\":\"9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85\",\"client_key\":\"40cd10039913ceabacf05a7c60e1ad69bb2964987bc50f77495e514dc451f907c3d8ebcdab20eedde9c8f39b9a1d66609a637352f318552fb69d4b3672516d1a\",\"keys\":[{\"public_key\":\"40cd10039913ceabacf05a7c60e1ad69bb2964987bc50f77495e514dc451f907c3d8ebcdab20eedde9c8f39b9a1d66609a637352f318552fb69d4b3672516d1a\",\"private_key\":\"a3a88aad5d89cec28c6e37c2925560ce160ac14d2cdcf4a4654b2bb358fe7514\"}],\"mnemonics\":\"inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown\",\"version\":\"1.0\",\"date_created\":\"2021-05-21 17:32:29.484657 +0545 +0545 m=+0.072791323\"}"
)

func NewMockZBox() *sdks.ZBox {
	z := sdks.New("9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85", "40cd10039913ceabacf05a7c60e1ad69bb2964987bc50f77495e514dc451f907c3d8ebcdab20eedde9c8f39b9a1d66609a637352f318552fb69d4b3672516d1a", "bls0chain")
	z.InitWallet(zboxWallet)

	return z
}

func TestCreateDir(t *testing.T) {
	list := []struct {
		TestName string

		ArgAllocationTx string
		ArgName         string

		Setup    func()
		Teardown func()

		Success bool
	}{
		{TestName: "create_dir_ok", ArgAllocationTx: "create_dir_ok", ArgName: "/test", Success: true, Setup: func() {
			blobber.MockCreateDir("create_dir_ok", "/test", http.StatusOK)

		}, Teardown: func() {
			blobber.UnmockCreateDir("create_dir_ok", "/test")
		}},
	}

	zbox := NewMockZBox()

	blobbers := []*dev.Server{
		dev.NewBlobberServer(),
		dev.NewBlobberServer(),
	}

	for _, it := range list {
		t.Run(it.TestName, func(test *testing.T) {

			it.Setup()

			blobberURLs := make([]string, len(blobbers))
			for i, b := range blobbers {
				blobberURLs[i] = b.URL
			}

			b := New(zbox, blobberURLs...)

			err := b.CreateDir(it.ArgAllocationTx, it.ArgName)

			r := require.New(test)
			if it.Success {
				r.Nil(err)
			} else {
				r.NotNil(err)
			}

		})
	}
}
