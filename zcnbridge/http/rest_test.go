package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MakeURL(t *testing.T) {
	p := Params{
		"blobber_id": "1",
	}

	url := makeURL(p, "https://baseuri", "/relativePath").String()

	require.Equal(
		t,
		"https://baseuri/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/relativePath?blobber_id=1",
		url,
	)

	p = Params{
		"blobber_id": "1",
		"path":       "2",
	}

	url = makeURL(p, "https://baseuri", "/relativePath").String()

	require.Equal(
		t,
		"https://baseuri/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/relativePath?blobber_id=1&path=2",
		url,
	)
}
