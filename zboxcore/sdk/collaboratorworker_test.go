package sdk

import (
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	collaboratorWorkerTestDir = configDir + "/collaboratorworker"
)

func TestCollaboratorRequest_UpdateCollaboratorToBlobbers(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, collaboratorWorkerTestDir, blobberMocks)
	defer cncl()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		want           bool
	}{
		{
			"Test_Update_Collaborator_To_Blobbers_Failure",
			nil,
			false,
		},
		{
			"Test_Update_Collaborator_To_Blobbers_Success",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				m.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
				}, nil)
				return nil
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			req := &CollaboratorRequest{
				a:              a,
				path:           "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
			}
			got := req.UpdateCollaboratorToBlobbers()
			var check = require.False
			if tt.want {
				check = require.True
			}
			check(got)
		})
	}
}

func TestCollaboratorRequest_updateCollaboratorToBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 1)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, collaboratorWorkerTestDir, blobberMocks)
	defer cncl()
	var wg sync.WaitGroup
	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		want           bool
	}{
		{
			"Test_update_Collaborator_To_Blobber_Failure",
			nil,
			false,
		},
		{
			"Test_Update_Collaborator_To_Blobbers_Success",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				m.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
				}, nil)
				return nil
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			req := &CollaboratorRequest{
				a:              a,
				path:           "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
				wg:             func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			}
			rspCh := make(chan bool, 1)
			go req.updateCollaboratorToBlobber(req.a.Blobbers[0], 0, rspCh)
			resp := <-rspCh
			var check = require.False
			if tt.want {
				check = require.True
			}
			check(resp)
		})
	}
}

func TestCollaboratorRequest_RemoveCollaboratorFromBlobbers(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, collaboratorWorkerTestDir, blobberMocks)
	defer cncl()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		want           bool
	}{
		{
			"Test_Remove_Collaborator_From_Blobbers_Failure",
			nil,
			false,
		},
		{
			"Test_Update_Collaborator_To_Blobbers_Success",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				m.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
				}, nil)
				return nil
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			req := &CollaboratorRequest{
				a:              a,
				path:           "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
			}
			got := req.RemoveCollaboratorFromBlobbers()
			var check = require.False
			if tt.want {
				check = require.True
			}
			check(got)
		})
	}
}

func TestCollaboratorRequest_removeCollaboratorFromBlobber(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, collaboratorWorkerTestDir, blobberMocks)
	defer cncl()
	var wg sync.WaitGroup
	tests := []struct {
		name           string
		additionalMock func(t *testing.T, testCaseName string) (teardown func(t *testing.T))
		want           bool
	}{
		{
			"Test_remove_Collaborator_From_Blobber_Failure",
			nil,
			false,
		},
		{
			"Test_Update_Collaborator_To_Blobbers_Success",
			func(t *testing.T, testCaseName string) (teardown func(t *testing.T)) {
				m := &mocks.HttpClient{}
				zboxutil.Client = m
				m.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
				}, nil)
				return nil
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if additionalMock := tt.additionalMock; additionalMock != nil {
				if teardown := additionalMock(t, tt.name); teardown != nil {
					defer teardown(t)
				}
			}
			req := &CollaboratorRequest{
				a:              a,
				path:           "/1.txt",
				collaboratorID: "9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468",
				wg:             func() *sync.WaitGroup { wg.Add(1); return &wg }(),
			}
			rspCh := make(chan bool, 1)
			go req.removeCollaboratorFromBlobber(req.a.Blobbers[0], 0, rspCh)
			resp := <-rspCh
			var check = require.False
			if tt.want {
				check = require.True
			}
			check(resp)
		})
	}
}
