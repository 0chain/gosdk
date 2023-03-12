package http

import (
	"testing"

	"github.com/0chain/gosdk/zcncore"

	"github.com/stretchr/testify/require"
)

func Test_MakeURL(t *testing.T) {
	p := Params{
		"blobber_id": "1",
	}

	url := makeURL(p, "https://baseuri", "/relativePath").String()
	require.Equal(t, "https://baseuri/v1/screst/"+zcncore.ZCNSCSmartContractAddress+"/relativePath?blobber_id=1", url)

	p = Params{
		"blobber_id": "1",
		"path":       "2",
	}

	url = makeURL(p, "https://baseuri", "/relativePath").String()
	require.Equal(t, "https://baseuri/v1/screst/"+zcncore.ZCNSCSmartContractAddress+"/relativePath?blobber_id=1&path=2", url)
}
