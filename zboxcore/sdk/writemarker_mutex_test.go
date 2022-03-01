package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/dev"
	"github.com/0chain/gosdk/dev/mock"
	"github.com/0chain/gosdk/sdks/blobber"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/require"
)

func TestWriteMarkerMutext_Should_Lock(t *testing.T) {
	a := &Allocation{
		ID:           "TestWriteMarkerMutext",
		Tx:           "TestWriteMarkerMutext",
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

func TestWriteMarkerMutext_Pending_Should_Lock(t *testing.T) {
	a := &Allocation{
		ID:           "TestWriteMarkerMutext",
		Tx:           "TestWriteMarkerMutext",
		DataShards:   2,
		ParityShards: 1,
	}
	setupMockAllocation(t, a)

	require := require.New(t)

	respStatusOK, _ := json.Marshal(&WMLockResult{
		Status: WMLockStatusOK,
	})

	respStatusPending, _ := json.Marshal(&WMLockResult{
		Status: WMLockStatusPending,
	})

	m := make(mock.ResponseMap)

	statusOK := mock.Response{
		StatusCode: http.StatusOK,
		Body:       respStatusOK,
	}
	statusPending := mock.Response{
		StatusCode: http.StatusOK,
		Body:       respStatusPending,
	}

	m[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = statusPending

	server := dev.NewBlobberServer(m)
	defer server.Close()

	for i := 0; i < 3; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      "write_marker_mutex_" + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}

	go func() {
		time.Sleep(1 * time.Second)
		m[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = statusOK
	}()

	mutex := CreateWriteMarkerMutex(a)

	err := mutex.Lock(context.TODO(), zboxutil.NewConnectionId())
	require.Nil(err)
}

func TestWriteMarkerMutext_Some_Blobbers_Down_Should_Lock(t *testing.T) {
	a := &Allocation{
		ID:           "TestWriteMarkerMutext",
		Tx:           "TestWriteMarkerMutext",
		DataShards:   2,
		ParityShards: 1,
	}
	setupMockAllocation(t, a)

	require := require.New(t)

	respStatusOK, _ := json.Marshal(&WMLockResult{
		Status: WMLockStatusOK,
	})

	m := make(mock.ResponseMap)

	statusOK := mock.Response{
		StatusCode: http.StatusOK,
		Body:       respStatusOK,
	}

	m[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = statusOK

	server := dev.NewBlobberServer(m)
	defer server.Close()

	for i := 0; i < 3; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      "write_marker_mutex_" + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}

	// 1st blobber is unreachable
	a.Blobbers[0].Baseurl = "http://127.0.0.1:5003"

	mutex := CreateWriteMarkerMutex(a)

	err := mutex.Lock(context.TODO(), zboxutil.NewConnectionId())
	require.Nil(err)
}

func TestWriteMarkerMutext_Too_Less_Blobbers_Response_Should_Not_Lock(t *testing.T) {
	a := &Allocation{
		ID:           "TestWriteMarkerMutext",
		Tx:           "TestWriteMarkerMutext",
		DataShards:   2,
		ParityShards: 1,
	}
	setupMockAllocation(t, a)

	require := require.New(t)

	respStatusOK, _ := json.Marshal(&WMLockResult{
		Status: WMLockStatusOK,
	})

	m := make(mock.ResponseMap)

	statusOK := mock.Response{
		StatusCode: http.StatusOK,
		Body:       respStatusOK,
	}

	m[http.MethodPost+":"+blobber.EndpointWriteMarkerLock+a.Tx] = statusOK

	server := dev.NewBlobberServer(m)
	defer server.Close()

	for i := 0; i < 3; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      "write_marker_mutex_" + strconv.Itoa(i),
			Baseurl: server.URL,
		})
	}

	//  blobber 1/2 are unreachable
	a.Blobbers[0].Baseurl = "http://127.0.0.1:5003"
	a.Blobbers[1].Baseurl = "http://127.0.0.1:5003"

	mutex := CreateWriteMarkerMutex(a)

	err := mutex.Lock(context.TODO(), zboxutil.NewConnectionId())
	require.ErrorIs(constants.ErrNotLockedWritMarker, err)
}
