package sdk

import (
	"github.com/0chain/gosdk/core/common"
	"github.com/stretchr/testify/require"
	"testing"
)

const tokenUnit = 10000000000.0

func Test_getMinWriteRead_success(t *testing.T) {
	var (
		ssc = newTestAllocation()
		err error
	)

	ssc.initialized = true
	sdkInitialized = true
	require.NotNil(t, ssc.BlobberDetails)

	minW, minR, err := ssc.GetMinWriteRead()
	require.NoError(t, err)
	require.Equal(t, 800000/tokenUnit, minW)
	require.Equal(t, 600000/tokenUnit, minR)
}

func Test_GetMaxWriteRead_success(t *testing.T) {
	var (
		ssc = newTestAllocation()
		err error
	)

	ssc.initialized = true
	sdkInitialized = true
	require.NotNil(t, ssc.BlobberDetails)

	maxW, maxR, err := ssc.GetMaxWriteRead()
	require.NoError(t, err)
	require.Equal(t, 900000/tokenUnit, maxW)
	require.Equal(t, 700000/tokenUnit, maxR)
}

func Test_GetMaxStorageCost_success(t *testing.T) {
	var (
		ssc = newTestAllocation()
		err error
	)

	ssc.initialized = true
	sdkInitialized = true
	cost, err := ssc.GetMaxStorageCost(100)
	require.NoError(t, err)
	require.Equal(t, 0.009000000000000001, cost)
}

func Test_GetMinStorageCost_success(t *testing.T) {
	var (
		ssc = newTestAllocation()
		err error
	)

	ssc.initialized = true
	sdkInitialized = true
	cost, err := ssc.GetMinStorageCost(100)
	require.NoError(t, err)
	require.Equal(t, 0.008, cost)
}

func Test_getMinWriteRead_noblobbers(t *testing.T) {
	var (
		ssc = newTestAllocationEmptyBlobbers()
		err error
	)
	ssc.initialized = true
	_, _, err = ssc.GetMinWriteRead()
	require.Error(t, err)
}

func Test_getMinWriteRead_notinitialized(t *testing.T) {
	var (
		ssc = newTestAllocation()
		err error
	)
	ssc.initialized = false
	_, _, err = ssc.GetMinWriteRead()
	require.Error(t, err)
}

func newTestAllocationEmptyBlobbers() (ssc *Allocation) {
	ssc = new(Allocation)
	ssc.Expiration = 0
	ssc.ID = "ID"
	ssc.BlobberDetails = make([]*BlobberAllocation, 0)
	return ssc
}

func newTestAllocation() (ssc *Allocation) {
	ssc = new(Allocation)
	ssc.Expiration = 0
	ssc.ID = "ID"
	ssc.BlobberDetails = newBlobbersDetails()
	return ssc
}

func newBlobbersDetails() (blobbers []*BlobberAllocation) {
	blobberDetails := make([]*BlobberAllocation, 0)

	var balloc BlobberAllocation
	balloc.BlobberID = "ID1"
	balloc.Size = 1000
	balloc.Terms = Terms{ReadPrice: common.Balance(700000), WritePrice: common.Balance(800000)}

	var balloc2 BlobberAllocation
	balloc2.BlobberID = "ID2"
	balloc2.Size = 2000
	balloc2.Terms = Terms{ReadPrice: common.Balance(600000), WritePrice: common.Balance(900000)}

	return append(append(blobberDetails, &balloc), &balloc2)
}
