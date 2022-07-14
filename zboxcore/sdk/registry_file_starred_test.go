package sdk

import (
	"fmt"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllocation_UpdateStarredFiles(t *testing.T) {
	for _, tc := range []struct {
		name           string
		inputPaths     []string
		updateErr      error
		wantUpdateData []byte
		wantErr        error
	}{
		{
			name:           "update successfully",
			inputPaths:     []string{"/abc.txt", "/def.txt"},
			wantUpdateData: []byte("/abc.txt\n/def.txt"),
		},
		{
			name:           "update successfully with single-path input",
			inputPaths:     []string{"/abc.txt"},
			wantUpdateData: []byte("/abc.txt"),
		},
		{
			name:           "update successfully with empty input",
			inputPaths:     []string{},
			wantUpdateData: []byte(""),
		},
		{
			name:           "update successfully with nil input",
			inputPaths:     nil,
			wantUpdateData: []byte(""),
		},
		{
			name:           "update error",
			inputPaths:     []string{"/abc.txt", "/def.txt"},
			updateErr:      fmt.Errorf("update error"),
			wantUpdateData: []byte("/abc.txt\n/def.txt"),
			wantErr:        errors.New("update_starred_files_failed", "failed to update registry file for starred files: update error"),
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			dummyAlloc := &Allocation{}

			starredFileRegistryManager = func(a *Allocation) registryFileManager {
				assert.Equal(t, dummyAlloc, a)
				return &mockRegistryFileManager{
					t:              t,
					updateErr:      tt.updateErr,
					wantUpdateData: tt.wantUpdateData,
				}
			}

			err := dummyAlloc.UpdateStarredFiles(tt.inputPaths)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestAllocation_GetStarredFiles(t *testing.T) {
	for _, tc := range []struct {
		name                string
		getResData          []byte
		getResTimestamp     common.Timestamp
		getErr              error
		wantErr             error
		wantPaths           []string
		wantUpdateTimestamp common.Timestamp
	}{
		{
			name:                "get successfully",
			getResData:          []byte("/abc.txt\n/def.txt"),
			getResTimestamp:     common.Timestamp(1642816984),
			wantPaths:           []string{"/abc.txt", "/def.txt"},
			wantUpdateTimestamp: common.Timestamp(1642816984),
		},
		{
			name:                "get successfully with empty data",
			getResData:          []byte{},
			getResTimestamp:     common.Timestamp(0),
			wantPaths:           []string{},
			wantUpdateTimestamp: common.Timestamp(0),
		},
		{
			name:    "get throws error",
			getErr:  fmt.Errorf("get error"),
			wantErr: errors.New("get_starred_files_failed", "failed to retrieve starred files: get error"),
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			dummyAlloc := &Allocation{}

			starredFileRegistryManager = func(a *Allocation) registryFileManager {
				assert.Equal(t, dummyAlloc, a)
				return &mockRegistryFileManager{
					t:               t,
					getErr:          tt.getErr,
					getResData:      tt.getResData,
					getResTimestamp: tt.getResTimestamp,
				}
			}

			gotPaths, gotTimestamp, err := dummyAlloc.GetStarredFiles()

			assert.Equal(t, tt.wantPaths, gotPaths)
			assert.Equal(t, tt.wantUpdateTimestamp, gotTimestamp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestAllocation_GetStarredFilesLastUpdateTimestamp(t *testing.T) {
	for _, tc := range []struct {
		name            string
		getResTimestamp common.Timestamp
		getErr          error
		wantErr         error
		wantTimestamp   common.Timestamp
	}{
		{
			name:            "get last update timestamp successfully",
			getResTimestamp: common.Timestamp(1642816984),
			wantTimestamp:   common.Timestamp(1642816984),
		},
		{
			name:    "get throws error",
			getErr:  fmt.Errorf("server error"),
			wantErr: errors.New("get_starred_files_last_update_timestamp_failed", "failed to get last update timestamp of registry file for starred files: server error"),
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			dummyAlloc := &Allocation{}

			starredFileRegistryManager = func(a *Allocation) registryFileManager {
				assert.Equal(t, dummyAlloc, a)
				return &mockRegistryFileManager{
					t:               t,
					getErr:          tt.getErr,
					getResTimestamp: tt.getResTimestamp,
				}
			}

			got, err := dummyAlloc.GetStarredFilesLastUpdateTimestamp()

			assert.Equal(t, tt.wantTimestamp, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

type mockRegistryFileManager struct {
	t               *testing.T
	updateErr       error
	getResData      []byte
	getResTimestamp common.Timestamp
	getErr          error
	wantUpdateData  []byte
}

func (m *mockRegistryFileManager) Update(data []byte) error {
	assert.Equal(m.t, m.wantUpdateData, data)
	return m.updateErr

}

func (m *mockRegistryFileManager) Get() ([]byte, common.Timestamp, error) {
	return m.getResData, m.getResTimestamp, m.getErr
}

func (m *mockRegistryFileManager) GetLastUpdateTimestamp() (common.Timestamp, error) {
	return m.getResTimestamp, m.getErr
}
