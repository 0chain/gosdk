package sdk

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	zclient "github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/require"
)

func TestAllocation_RenameObject(t *testing.T) {
	const (
		mockType = "f"
	)

	rawClient := zboxutil.Client
	createClient := resty.CreateClient

	var mockClient = mocks.HttpClient{}
	zboxutil.Client = &mockClient

	client := zclient.GetClient()
	client.Wallet = &zcncrypto.Wallet{
		ClientID:  mockClientId,
		ClientKey: mockClientKey,
	}

	zboxutil.Client = &mockClient
	resty.CreateClient = func(t *http.Transport, timeout time.Duration) resty.Client {
		return &mockClient
	}

	defer func() {
		zboxutil.Client = rawClient
		resty.CreateClient = createClient
	}()
	type parameters struct {
		path     string
		destName string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T, string, *Allocation) (teardown func(*testing.T))
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Uninitialized_Failed",
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
			name: "Test_Wrong_Path_Or_Destination_Path_Failed",
			parameters: parameters{
				path:     "",
				destName: "",
			},
			wantErr: true,
			errMsg:  "invalid_path: Invalid path for the list",
		},
		{
			name: "Test_Invalid_Remote_Absolute_Path",
			parameters: parameters{
				path:     "abc",
				destName: "/2.txt",
			},
			wantErr: true,
			errMsg:  "invalid_path: Path should be valid and absolute",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				path:     "/1.txt",
				destName: "/2.txt",
			},
			setup: func(t *testing.T, testCaseName string, a *Allocation) (teardown func(t *testing.T)) {
				body, err := json.Marshal(&fileref.ReferencePath{
					Meta: map[string]interface{}{
						"type": mockType,
					},
				})
				require.NoError(t, err)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_RenameObject", testCaseName, a, http.MethodGet, http.StatusOK, body)
				setupMockHttpResponse(t, &mockClient, "TestAllocation_RenameObject", testCaseName, a, http.MethodPost, http.StatusOK, []byte(""))
				setupMockCommitRequest(a)
				setupMockRollback(a, &mockClient)
				setupMockWriteLockRequest(a, &mockClient)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			a := &Allocation{
				DataShards:   2,
				ParityShards: 2,
				FileOptions:  63,
			}
			a.InitAllocation()
			sdkInitialized = true
			for i := 0; i < numBlobbers; i++ {
				a.Blobbers = append(a.Blobbers, &blockchain.StorageNode{
					ID:      tt.name + mockBlobberId + strconv.Itoa(i),
					Baseurl: "http://TestAllocation_RenameObject" + tt.name + mockBlobberUrl + strconv.Itoa(i),
				})
			}
			if tt.setup != nil {
				if teardown := tt.setup(t, tt.name, a); teardown != nil {
					defer teardown(t)
				}
			}
			err := a.RenameObject(tt.parameters.path, tt.parameters.destName)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, errors.Top(err))
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}
