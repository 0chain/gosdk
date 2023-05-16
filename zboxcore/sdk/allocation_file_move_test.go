package sdk

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/require"
)

func TestAllocation_MoveObject(t *testing.T) {

	const (
		mockType = "f"
	)

	rawClient := zboxutil.Client

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	zboxutil.Client = &mockClient

	defer func() {
		zboxutil.Client = rawClient
	}()

	type parameters struct {
		path     string
		destPath string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Cover_Copy_Object",
			parameters: parameters{
				path:     "/1.txt",
				destPath: "/d",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				a.initialized = false
				return func(t *testing.T) {
					a.initialized = true
				}
			},
			wantErr: true,
			errMsg:  "sdk_not_initialized: Please call InitStorageSDK Init and use GetAllocation to get the allocation object",
		},
		{
			name: "Test_Cover_Delete_Object",
			parameters: parameters{
				path:     "/1.txt",
				destPath: "/d",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodPost, http.StatusOK, []byte(""))
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodPost, http.StatusOK, []byte(`{"status":2}`))
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodPost, http.StatusOK, []byte(`{"status":2}`))
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodDelete, http.StatusOK, []byte(""))
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodDelete, http.StatusOK, []byte(""))
				setupMockHttpResponse(t, &mockClient, "TestAllocation_MoveObject", testCaseName, a, http.MethodDelete, http.StatusOK, []byte(""))
				setupMockCommitRequest(a)
				setupMockRollback(a, &mockClient)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				Tx:           "TestAllocation_MoveObject",
				DataShards:   2,
				ParityShards: 2,
			}

			setupMockAllocation(t, a)

			for i := 0; i < numBlobbers; i++ {

				path := "/TestAllocation_MoveObject" + tt.name + mockBlobberUrl + strconv.Itoa(i)

				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: path,
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.MoveObject(tt.parameters.path, tt.parameters.destPath)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}
