package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/0chain/gosdk/dev"
	"github.com/0chain/gosdk/dev/mock"
	"github.com/0chain/gosdk/sdks/blobber"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/require"
)

func TestWriteMarkerMutext_Should_Work(t *testing.T) {
	a := &Allocation{
		ID:           "TestAllocation_UpdateFile",
		Tx:           "TestAllocation_UpdateFile",
		DataShards:   2,
		ParityShards: 1,
	}
	setupMockAllocation(t, a)

	require := require.New(t)

	resp := &WMLockResult{
		Status: WMLockStatusOK,
	}

	respBuf, _ := json.Marshal(resp)
	m := make(mock.ResponseMap)
	m[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = mock.Response{
		StatusCode: http.StatusOK,
		Body:       respBuf,
	}

	server := dev.NewBlobberServer(m)
	defer server.Close()

	for i := 0; i < 3; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      "write_marker_mutex_" + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}

	mutex := CreateWriteMarkerMutex(a)

	err := mutex.Lock(context.TODO(), zboxutil.NewConnectionId())
	require.Nil(err)

}
