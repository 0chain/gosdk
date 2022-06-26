package sdk

import (
	"fmt"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateStarredFiles(t *testing.T) {
	for _, tc := range []struct {
		name           string
		input          *StarredFiles
		updateErr      error
		wantUpdateData []byte
		wantErr        error
	}{
		{
			name:           "update successfully",
			input:          &StarredFiles{UpdatedAt: common.Timestamp(1642816984), Files: []StarredFile{{Path: "/abc.txt"}, {Path: "/def.txt"}}},
			wantUpdateData: []byte(`{"files":[{"path":"/abc.txt"},{"path":"/def.txt"}]}`),
		},
		{
			name:           "update error",
			input:          &StarredFiles{UpdatedAt: common.Timestamp(1642816984), Files: []StarredFile{{Path: "/abc.txt"}, {Path: "/def.txt"}}},
			updateErr:      fmt.Errorf("update error"),
			wantUpdateData: []byte(`{"files":[{"path":"/abc.txt"},{"path":"/def.txt"}]}`),
			wantErr:        errors.New("update_starred_files_failed", "Failed to update registry file for starred files: update error"),
		},
		{
			name:    "missing input",
			input:   nil,
			wantErr: errors.New("update_starred_files_failed", "Starred files is nil"),
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			dummyAlloc := &Allocation{}

			starredFileRegistryManager = func(a *Allocation) RegistryFileManager {
				assert.Equal(t, dummyAlloc, a)
				return &mockRegistryFileManager{
					t:              t,
					updateErr:      tt.updateErr,
					wantUpdateData: tt.wantUpdateData,
				}
			}

			err := UpdateStarredFiles(dummyAlloc, tt.input)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestGetStarredFiles(t *testing.T) {
	for _, tc := range []struct {
		name            string
		getResData      []byte
		getResTimestamp common.Timestamp
		getErr          error
		wantErr         error
		want            *StarredFiles
	}{
		{
			name:            "get successfully",
			getResData:      []byte(`{"files":[{"path":"/abc.txt"},{"path":"/def.txt"}]}`),
			getResTimestamp: common.Timestamp(1642816984),
			want:            &StarredFiles{UpdatedAt: common.Timestamp(1642816984), Files: []StarredFile{{Path: "/abc.txt"}, {Path: "/def.txt"}}},
		},
		{
			name:    "get throws error",
			getErr:  fmt.Errorf("get error"),
			wantErr: errors.New("get_starred_files_failed", "Failed to retrieve starred files: get error"),
		},
		{
			name:            "Bad data retrieved",
			getResData:      []byte(`not a json`),
			getResTimestamp: common.Timestamp(1642816984),
			wantErr:         errors.New("get_starred_files_failed", "Failed to parse downloaded registry file for starred files: invalid character 'o' in literal null (expecting 'u')"),
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			dummyAlloc := &Allocation{}

			starredFileRegistryManager = func(a *Allocation) RegistryFileManager {
				assert.Equal(t, dummyAlloc, a)
				return &mockRegistryFileManager{
					t:               t,
					getErr:          tt.getErr,
					getResData:      tt.getResData,
					getResTimestamp: tt.getResTimestamp,
				}
			}

			got, err := GetStarredFiles(dummyAlloc)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestGetStarredFilesLastUpdateTimestamp(t *testing.T) {
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
			wantErr: errors.New("get_starred_files_last_update_timestamp_failed", "Failed to get last update timestamp of registry file for starred files: server error"),
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			dummyAlloc := &Allocation{}

			starredFileRegistryManager = func(a *Allocation) RegistryFileManager {
				assert.Equal(t, dummyAlloc, a)
				return &mockRegistryFileManager{
					t:               t,
					getErr:          tt.getErr,
					getResTimestamp: tt.getResTimestamp,
				}
			}

			got, err := GetStarredFilesLastUpdateTimestamp(dummyAlloc)

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
