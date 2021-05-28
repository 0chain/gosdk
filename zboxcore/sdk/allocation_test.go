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

func TestGetMinMaxWriteReadSuccess(t *testing.T) {
	var ssc = newTestAllocation()
	ssc.DataShards = 5
	ssc.ParityShards = 4

	ssc.initialized = true
	sdkInitialized = true
	require.NotNil(t, ssc.BlobberDetails)

	t.Run("Success minR, minW", func(t *testing.T) {
		minW, minR, err := ssc.GetMinWriteRead()
		require.NoError(t, err)
		require.Equal(t, 0.01, minW)
		require.Equal(t, 0.01, minR)
	})

	t.Run("Success maxR, maxW", func(t *testing.T) {
		maxW, maxR, err := ssc.GetMaxWriteRead()
		require.NoError(t, err)
		require.Equal(t, 0.01, maxW)
		require.Equal(t, 0.01, maxR)
	})

	t.Run("Error / No Blobbers", func(t *testing.T) {
		var (
			ssc = newTestAllocationEmptyBlobbers()
			err error
		)
		ssc.initialized = true
		_, _, err = ssc.GetMinWriteRead()
		require.Error(t, err)
	})

	t.Run("Error / Empty Blobbers", func(t *testing.T) {
		var err error
		ssc.initialized = false
		_, _, err = ssc.GetMinWriteRead()
		require.Error(t, err)
	})

	t.Run("Error / Not Initialized", func(t *testing.T) {
		var err error
		ssc.initialized = false
		_, _, err = ssc.GetMinWriteRead()
		require.Error(t, err)
	})

}

func TestGetMaxMinStorageCostSuccess(t *testing.T) {
	var ssc = newTestAllocation()
	ssc.DataShards = 4
	ssc.ParityShards = 2

	ssc.initialized = true
	sdkInitialized = true

	t.Run("Storage cost", func(t *testing.T) {
		cost, err := ssc.GetMaxStorageCost(100 * GB)
		require.NoError(t, err)
		require.Equal(t, 1.5, cost)
	})
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

	for i := 1; i <= 1; i++ {
		var balloc BlobberAllocation
		balloc.Size = 1000

		balloc.Terms = Terms{ReadPrice: common.Balance(100000000), WritePrice: common.Balance(100000000)}
		blobberDetails = append(blobberDetails, &balloc)
	}

	return blobberDetails
}

func TestThrowErrorWhenBlobbersRequiredGreaterThanImplicitLimit128(t *testing.T) {
	setupMocks()

	var maxNumOfBlobbers = 129

	var allocation = &Allocation{}
	var blobbers = make([]*blockchain.StorageNode, maxNumOfBlobbers)
	allocation.initialized = true
	sdkInitialized = true
	allocation.Blobbers = blobbers
	allocation.DataShards = 64
	allocation.ParityShards = 65

	var file fileref.Attributes
	err := allocation.uploadOrUpdateFile("", "/", nil, false, "", false, false, file)

	var expectedErr = "allocation requires [129] blobbers, which is greater than the maximum permitted number of [128]. reduce number of data or parity shards and try again"
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
