package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusCallback(t *testing.T) {

	status := &StatusCallback{
		status: make(map[string]*Status),
	}

	status.Started("allocationId", "/remotePath", 0, 100)
	status.InProgress("allocationId", "/remotePath", 0, 100, nil)
	status.Completed("allocationId", "/remotePath", "", "", 100, 0)
	status.Error("allocationId", "/remotePath", 0, errors.New("err"))

	s := status.GetStatus("/remotePath")

	require.True(t, s.Started)
	require.True(t, s.Completed)
	require.EqualValues(t, 100, s.TotalBytes)
	require.EqualValues(t, 100, s.CompletedBytes)
	require.Equal(t, "err", s.Error)
}
