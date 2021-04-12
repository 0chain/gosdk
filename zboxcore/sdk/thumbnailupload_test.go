package sdk

import (
	"github.com/0chain/gosdk/zboxcore/encryption"
	encMocks "github.com/0chain/gosdk/zboxcore/encryption/mocks"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushThumbnailData(t *testing.T) {

	var req = &UploadRequest{}
	test := &TestStructImplIOWrite{}
	req.thumbnailHashWr = test
	req.datashards = 2
	req.parityshards = 2
	err := req.pushThumbnailData([]byte("test"))
	if err != nil {
		t.Errorf("pushThumbnailData() = %v, want %v", err, "nil")
	}
}

func TestPushThumbnailDataIsEncrypted(t *testing.T) {
	assertion := assert.New(t)
	mec := &encMocks.EncryptionScheme{}
	mec.On("Encrypt", []byte{0x65}).Return(&encryption.EncryptedMessage{}, nil).Once()

	var req = &UploadRequest{}
	test := &TestStructImplIOWrite{}
	req.thumbnailHashWr = test
	req.datashards = 200
	req.parityshards = 20
	req.isEncrypted = true
	req.uploadMask = 2
	req.encscheme = mec
	req.uploadThumbCh = make([]chan []byte, 5)
	req.uploadThumbCh[0] = make(chan []byte, 5)
	err := req.pushThumbnailData([]byte("test123"))
	mec.Test(t)
	mec.AssertExpectations(t)
	assertion.NoErrorf(err, "unexpected error but got: %v", err)
}

func TestPushThumbnailDataEncodeFail(t *testing.T) {
	assertion := assert.New(t)

	var req = &UploadRequest{}
	test := &TestStructImplIOWrite{}
	req.thumbnailHashWr = test
	err := req.pushThumbnailData([]byte("test"))
	if err != nil {
		assertion.Errorf(err, "expected error != nil")
	}
}

func TestCompleteThumbnailPush(t *testing.T) {
	assertion := assert.New(t)
	var req = &UploadRequest{}
	test := &mocklHash{}
	req.thumbnailHash = test
	req.uploadMask = 0
	req.filemeta = &UploadFileMeta{}
	err := req.completeThumbnailPush()
	if err != nil {
		assertion.Errorf(err, "expected error != nil")
	}
}

func TestProcessThumbnail(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var req = &UploadRequest{}
	test := &TestStructImplIOWrite{}
	req.thumbnailHashWr = test
	req.datashards = 2
	req.parityshards = 2
	req.thumbnailpath = "./allocation.go"
	req.filemeta = &UploadFileMeta{}
	req.thumbnailHash = &mocklHash{}
	s := &sync.WaitGroup{}
	s.Add(1)
	req.processThumbnail(a, s)
}
func TestProcessThumbnailChunksPerShard(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	a.DataShards = 5000

	defer cncl()

	var req = &UploadRequest{}
	test := &TestStructImplIOWrite{}
	req.thumbnailHashWr = test
	req.datashards = 2
	req.parityshards = 2
	req.thumbnailpath = "./allocation.go"
	req.filemeta = &UploadFileMeta{}
	req.thumbnailHash = &mocklHash{}
	req.filemeta.ThumbnailSize = 999999
	s := &sync.WaitGroup{}
	s.Add(1)
	req.processThumbnail(a, s)
}
func TestProcessThumbnailFail(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var req = &UploadRequest{}
	test := &TestStructImplIOWrite{}
	req.thumbnailHashWr = test
	req.datashards = 2
	req.parityshards = 2
	req.thumbnailpath = "./failpath"
	req.filemeta = &UploadFileMeta{}
	req.thumbnailHash = &mocklHash{}
	s := &sync.WaitGroup{}
	s.Add(1)
	req.processThumbnail(a, s)
}

func TestProcessThumbnailIsEncrypted(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var req = &UploadRequest{}
	test := &TestStructImplIOWrite{}
	req.thumbnailHashWr = test
	req.datashards = 2
	req.parityshards = 2
	req.thumbnailpath = "./allocation.go"
	req.filemeta = &UploadFileMeta{}
	req.thumbnailHash = mocklHash{}
	s := &sync.WaitGroup{}
	s.Add(1)
	req.isEncrypted = true
	req.processThumbnail(a, s)
}

func (t mocklHash) Write(p []byte) (n int, err error) {
	return 1, nil
}

type mocklHash struct {
}

func (t mocklHash) Sum(b []byte) []byte {
	return []byte("test")
}

func (t mocklHash) Reset() {

}

func (t mocklHash) Size() int {
	return 2
}

func (t mocklHash) BlockSize() int {
	return 2
}
