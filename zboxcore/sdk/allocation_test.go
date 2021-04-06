package sdk

import (
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/stretchr/testify/require"
	"os"
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

func TestThrowErrorWhenBlobbersRequiredGreaterThanImplicitLimit32(t *testing.T) {
	setupMocks()

	var maxNumOfBlobbers = 33

	var allocation = &Allocation{}
	var blobbers = make([]*blockchain.StorageNode, maxNumOfBlobbers)
	allocation.initialized = true
	sdkInitialized = true
	allocation.Blobbers = blobbers
	allocation.DataShards = 16
	allocation.ParityShards = 17

	var file fileref.Attributes
	err := allocation.uploadOrUpdateFile("", "/", nil, false, "", false, false, file)

	var expectedErr = "allocation requires [33] blobbers, which is greater than the maximum permitted number of [32]. reduce number of data or parity shards and try again"
	if err == nil {
		t.Errorf("uploadOrUpdateFile() = expected error  but was %v", nil)
	} else if err.Error() != expectedErr {
		t.Errorf("uploadOrUpdateFile() = expected error message to be %v  but was %v", expectedErr, err.Error())
	}
}

func TestThrowErrorWhenBlobbersRequiredGreaterThanExplicitLimit(t *testing.T) {
	setupMocks()

	var maxNumOfBlobbers = 10

	var allocation = &Allocation{}
	var blobbers = make([]*blockchain.StorageNode, maxNumOfBlobbers)
	allocation.initialized = true
	sdkInitialized = true
	allocation.Blobbers = blobbers
	allocation.DataShards = 5
	allocation.ParityShards = 6

	var file fileref.Attributes
	err := allocation.uploadOrUpdateFile("", "/", nil, false, "", false, false, file)

	var expectedErr = "allocation requires [11] blobbers, which is greater than the maximum permitted number of [10]. reduce number of data or parity shards and try again"
	if err == nil {
		t.Errorf("uploadOrUpdateFile() = expected error  but was %v", nil)
	} else if err.Error() != expectedErr {
		t.Errorf("uploadOrUpdateFile() = expected error message to be %v  but was %v", expectedErr, err.Error())
	}
}

func TestDoNotThrowErrorWhenBlobbersRequiredLessThanLimit(t *testing.T) {
	setupMocks()

	var maxNumOfBlobbers = 10

	var allocation = &Allocation{}
	var blobbers = make([]*blockchain.StorageNode, maxNumOfBlobbers)
	allocation.initialized = true
	sdkInitialized = true
	allocation.Blobbers = blobbers
	allocation.DataShards = 5
	allocation.ParityShards = 4

	var file fileref.Attributes
	err := allocation.uploadOrUpdateFile("", "/", nil, false, "", false, false, file)

	if err != nil {
		t.Errorf("uploadOrUpdateFile() = expected no error but was %v", err)
	}
}

func setupMocks() {
	GetFileInfo = func(localpath string) (os.FileInfo, error) {
		return new(MockFile), nil
	}
}

type MockFile struct {
	os.FileInfo
	size int64
}

func (m MockFile) Size() int64 { return 10 }