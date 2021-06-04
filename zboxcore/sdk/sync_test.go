package sdk

import (
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/0chain/gosdk/zboxcore/zboxutil/mocks"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestAllocation_getRemoteFilesAndDirs(t *testing.T) {
	m := &mocks.HttpClient{}
	zboxutil.Client = m
	bodyStringSuccess := `{"allocation": "$allocation_id"}`
	bodyStringFailed := `{"error": "xxx"}`
	var bodyString string
	statusCode := http.StatusOK
	m.On("Do", mock.AnythingOfType("*http.Request")).Run(func(args mock.Arguments){
		req := args[0].(*http.Request)
		//check path
		url := req.URL.Host
		switch url {
		case blobber[0].URL:
			bodyString = strings.ReplaceAll(bodyStringSuccess, "$blobber_id", blobber[0].ID)
			statusCode = http.StatusOK
		case blobber[1].URL:
			bodyString = strings.ReplaceAll(bodyStringFailed, "$blobber_id", blobber[0].ID)
			statusCode = http.StatusBadRequest
		}
		allocationID := strings.Split(url, "/")[3]
		bodyString = strings.ReplaceAll(bodyString, "$allocation_id", allocationID)
	}).Return(&http.Response{
		StatusCode: statusCode,
		Body: ioutil.NopCloser(strings.NewReader(``)),
	}, nil)
}
