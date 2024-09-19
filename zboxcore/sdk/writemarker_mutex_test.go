package sdk

import (
	"bytes"
	"context"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/0chain/gosdk/zboxcore/blockchain"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWriteMarkerMutext_Should_Lock(t *testing.T) {
	rawClient := zboxutil.Client
	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	defer func() {
		zboxutil.Client = rawClient
	}()

	a := &Allocation{
		ID:           "TestWriteMarkerMutext",
		Tx:           "TestWriteMarkerMutext",
		DataShards:   2,
		ParityShards: 1,
		Owner:        mockClientId,
	}
	setupMockAllocation(t, a)

	setupHttpResponses := func(t *testing.T, testName string, numBlobbers int, numCorrect int) {
		for i := 0; i < numBlobbers; i++ {
			url := mockBlobberUrl + strconv.Itoa(i)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == "POST" &&
					strings.Contains(req.URL.String(), testName+url)
			})).Return(&http.Response{
				StatusCode: func() int {
					if i < numCorrect {
						return http.StatusOK
					}
					return http.StatusBadRequest
				}(),
				Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"status":2}`))),
			}, nil)

			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == "DELETE" &&
					strings.Contains(req.URL.String(), testName+url)
			})).Return(&http.Response{
				StatusCode: func() int {
					if i < numCorrect {
						return http.StatusOK
					}
					return http.StatusBadRequest
				}(),
			}, nil)
		}
	}

	for i := 0; i < 3; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      "write_marker_mutex_" + strconv.Itoa(i),
			Baseurl: "http://" + t.Name() + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	setupHttpResponses(t, t.Name(), len(a.Blobbers), len(a.Blobbers))

	mask := zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
	mu := &sync.Mutex{}
	mutex, _ := CreateWriteMarkerMutex(a)
	consensus := &Consensus{RWMutex: &sync.RWMutex{}}
	consensus.Init(a.consensusThreshold, a.fullconsensus)

	err := mutex.Lock(context.TODO(), &mask, mu, a.Blobbers,
		consensus, 0, time.Minute, zboxutil.NewConnectionId())
	require.Nil(t, err)

}

func TestWriteMarkerMutext_Some_Blobbers_Down_Should_Lock(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	rawClient := zboxutil.Client
	zboxutil.Client = &mockClient

	defer func() {
		zboxutil.Client = rawClient
	}()

	a := &Allocation{
		ID:           "TestWriteMarkerMutext",
		Tx:           "TestWriteMarkerMutext",
		DataShards:   2,
		ParityShards: 2,
		Owner:        mockClientId,
	}
	setupMockAllocation(t, a)

	require := require.New(t)

	setupHttpResponses := func(t *testing.T, testName string, numBlobbers int, numCorrect int) {
		for i := 0; i < numBlobbers; i++ {
			url := mockBlobberUrl + strconv.Itoa(i)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == "POST" &&
					strings.Contains(req.URL.String(), testName+url)
			})).Return(&http.Response{
				StatusCode: func() int {
					if i < numCorrect {
						return http.StatusOK
					}
					return http.StatusBadRequest
				}(),
				Body: io.NopCloser(bytes.NewReader([]byte(`{"status":2}`))),
			}, nil)

			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == "DELETE" &&
					strings.Contains(req.URL.String(), testName+url)
			})).Return(&http.Response{
				StatusCode: func() int {
					if i < numCorrect {
						return http.StatusOK
					}
					return http.StatusBadRequest
				}(),
				Body: io.NopCloser(bytes.NewReader([]byte(""))),
			}, nil)
		}
	}

	for i := 0; i < 4; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      "write_marker_mutex_" + strconv.Itoa(i),
			Baseurl: "http://" + t.Name() + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	setupHttpResponses(t, t.Name(), len(a.Blobbers), len(a.Blobbers)-1)
	mutex, _ := CreateWriteMarkerMutex(a)
	mask := zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
	mu := &sync.Mutex{}
	consensus := &Consensus{RWMutex: &sync.RWMutex{}}
	consensus.Init(a.consensusThreshold, a.fullconsensus)
	err := mutex.Lock(context.TODO(), &mask, mu, a.Blobbers,
		consensus, 0, time.Minute, zboxutil.NewConnectionId())
	require.Nil(err)
}

func TestWriteMarkerMutext_Too_Less_Blobbers_Response_Should_Not_Lock(t *testing.T) {
	var mockClient = mocks.HttpClient{}

	rawClient := zboxutil.Client
	zboxutil.Client = &mockClient

	defer func() {
		zboxutil.Client = rawClient
	}()

	a := &Allocation{
		ID:           "TestWriteMarkerMutext",
		Tx:           "TestWriteMarkerMutext",
		DataShards:   2,
		ParityShards: 1,
		Owner:        mockClientId,
	}
	setupMockAllocation(t, a)

	setupHttpResponses := func(t *testing.T, testName string, numBlobbers int, numCorrect int) {
		for i := 0; i < numBlobbers; i++ {
			url := mockBlobberUrl + strconv.Itoa(i)
			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == "POST" &&
					strings.Contains(req.URL.String(), testName+url)
			})).Return(&http.Response{
				StatusCode: func() int {
					if i < numCorrect {
						return http.StatusOK
					}
					return http.StatusBadRequest
				}(),
				Body: ioutil.NopCloser(bytes.NewReader([]byte(""))),
			}, nil)

			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == "DELETE" &&
					strings.Contains(req.URL.String(), testName+url)
			})).Return(&http.Response{
				StatusCode: func() int {
					if i < numCorrect {
						return http.StatusOK
					}
					return http.StatusBadRequest
				}(),
				Body: io.NopCloser(bytes.NewReader([]byte(""))),
			}, nil)
		}
	}

	for i := 0; i < a.DataShards+a.ParityShards; i++ {
		a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
			ID:      "write_marker_mutex_" + strconv.Itoa(i),
			Baseurl: "http://" + t.Name() + mockBlobberUrl + strconv.Itoa(i),
		})
	}

	setupHttpResponses(t, t.Name(), len(a.Blobbers), a.consensusThreshold-1)
	mutex, err := CreateWriteMarkerMutex(a)
	require.NoError(t, err)
	mask := zboxutil.NewUint128(1).Lsh(uint64(len(a.Blobbers))).Sub64(1)
	mu := &sync.Mutex{}
	consensus := &Consensus{RWMutex: &sync.RWMutex{}}
	consensus.Init(a.consensusThreshold, a.fullconsensus)
	err = mutex.Lock(context.TODO(), &mask, mu, a.Blobbers,
		consensus, 0, time.Minute, zboxutil.NewConnectionId())
	if err != nil {
		require.Contains(t, err.Error(), "lock_consensus_not_met")
	}
}
