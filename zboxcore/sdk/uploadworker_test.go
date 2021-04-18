package sdk

import (
	"bytes"
	"context"
	"errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/sdk/mocks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	tm "github.com/stretchr/testify/mock"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	encMocks "github.com/0chain/gosdk/zboxcore/encryption/mocks"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/stretchr/testify/assert"
)

const (
	uploadWorkerTestDir = configDir + "/uploadworker"
)

func TestPushDataIsEncrypted(t *testing.T) {
	assertion := assert.New(t)
	mec := &encMocks.EncryptionScheme{}
	mec.On("Encrypt", []byte{0x74, 0x65, 0x73, 0x74}).Return(&encryption.EncryptedMessage{}, nil).Once()
	var b = []byte{}
	wr := bytes.NewBuffer(b)

	var req = &UploadRequest{}
	req.remaining = 6
	req.fileHashWr = wr
	req.datashards = 1
	req.parityshards = 2
	req.isEncrypted = true
	req.uploadMask = zboxutil.NewUint128(2)
	req.encscheme = mec
	req.uploadDataCh = make([]chan []byte, 5)
	req.uploadDataCh[0] = make(chan []byte, 5)
	err := req.pushData([]byte("test"))
	mec.Test(t)
	mec.AssertExpectations(t)
	assertion.NoErrorf(err, "unexpected error but got: %v", err)
}

func TestPushData(t *testing.T) {
	assertion := assert.New(t)

	var b = []byte{}
	var wr = bytes.NewBuffer(b)
	var req = &UploadRequest{}
	req.remaining = 6
	req.fileHashWr = wr
	req.datashards = 1
	req.parityshards = 2

	err := req.pushData([]byte("test"))
	if err != nil {
		assertion.Errorf(err, "expected error != nil")
	}
}

func TestPushDataEncodeFail(t *testing.T) {
	assertion := assert.New(t)

	var b = []byte{}
	var wr = bytes.NewBuffer(b)
	var req = &UploadRequest{}
	req.remaining = 6
	test := wr
	req.fileHashWr = test

	err := req.pushData([]byte("test"))
	if err != nil {
		assertion.Errorf(err, "expected error != nil")
	}
}

func TestPrepareUpload(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()

	var b = []byte{}
	var wr = bytes.NewBuffer(b)
	var req = &UploadRequest{}
	req.remaining = 6
	req.fileHashWr = wr
	s := &sync.WaitGroup{}
	s.Add(1)
	req.filemeta = &UploadFileMeta{}
	req.filemeta.Size = 2
	req.prepareUpload(a, &blockchain.StorageNode{}, &fileref.FileRef{}, make(chan []byte, 1), make(chan []byte, 1), s)
}

func TestMaxBlobbersRequiredGreaterThanImplicitLimit128(t *testing.T) {
	var maxNumOfBlobbers = 129

	var req = &UploadRequest{}
	req.setUploadMask(maxNumOfBlobbers)
	req.fullconsensus = float32(maxNumOfBlobbers)

	if req.IsFullConsensusSupported() {
		t.Errorf("IsFullConsensusSupported() = %v, want %v", true, false)
	}
}

func TestMaxBlobbersRequiredEqualToImplicitLimit32(t *testing.T) {
	var maxNumOfBlobbers = 32

	var req = &UploadRequest{}
	req.setUploadMask(maxNumOfBlobbers)
	req.fullconsensus = float32(maxNumOfBlobbers)

	if !req.IsFullConsensusSupported() {
		t.Errorf("IsFullConsensusSupported() = %v, want %v", false, true)
	}
}

func TestNumBlobbersRequiredGreaterThanMask(t *testing.T) {
	var maxNumOfBlobbers = 5

	var req = &UploadRequest{}
	req.setUploadMask(maxNumOfBlobbers)
	req.fullconsensus = float32(6)

	if req.IsFullConsensusSupported() {
		t.Errorf("IsFullConsensusSupported() = %v, want %v", true, false)
	}
}

func TestNumBlobbersRequiredLessThanMask(t *testing.T) {
	var maxNumOfBlobbers = 5

	var req = &UploadRequest{}
	req.setUploadMask(maxNumOfBlobbers)
	req.fullconsensus = float32(4)

	if !req.IsFullConsensusSupported() {
		t.Errorf("IsFullConsensusSupported() = %v, want %v", false, true)
	}
}

func TestNumBlobbersRequiredZero(t *testing.T) {
	var maxNumOfBlobbers = 5

	var req = &UploadRequest{}
	req.setUploadMask(maxNumOfBlobbers)
	req.fullconsensus = float32(0)

	if !req.IsFullConsensusSupported() {
		t.Errorf("IsFullConsensusSupported() = %v, want %v", false, true)
	}
}
func TestSetupUpload(t *testing.T) {
	var maxNumOfBlobbers = 3
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()
	var req = &UploadRequest{}
	req.setUploadMask(maxNumOfBlobbers)
	req.fullconsensus = float32(0)
	req.filemeta = &UploadFileMeta{}
	req.filemeta.Name = "./allocation.go"
	if err := req.setupUpload(a); err != nil {
		t.Errorf("SetupUpload() = %v, want %v", err, nil)
	}
}
func TestSetupUploadIsEncrypted(t *testing.T) {
	var maxNumOfBlobbers = 3
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()
	var req = &UploadRequest{}
	req.setUploadMask(maxNumOfBlobbers)
	req.fullconsensus = float32(0)
	req.filemeta = &UploadFileMeta{}
	req.filemeta.Name = "1.txt"
	req.isEncrypted = true
	if err := req.setupUpload(a); err != nil {
		t.Errorf("SetupUpload() = %v, want %v", err, nil)
	}
}

func TestProcessUploadCompletedCallback(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()
	ctx := context.Background()
	var req = &UploadRequest{}
	req.filemeta = &UploadFileMeta{}
	req.filemeta.MimeType = "mime type"
	req.completedCallback = func(filepath string) {}
	req.processUpload(ctx, a)
}

func TestProcessUploadStatusCallback(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()
	scm := &mocks.StatusCallback{}
	scm.On("Error", "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", "", 0, tm.AnythingOfType("*common.Error")).Once()
	ctx := context.Background()
	var req = &UploadRequest{}
	req.filemeta = &UploadFileMeta{}
	req.filemeta.MimeType = "mime type"
	req.statusCallback = scm
	req.processUpload(ctx, a)
	scm.Test(t)
	scm.AssertExpectations(t)
}

func TestProcessUploadThumbnailPath(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()
	ctx := context.Background()
	var req = &UploadRequest{}
	req.filemeta = &UploadFileMeta{}
	req.filemeta.MimeType = "mime type"
	req.thumbnailpath = "thumbnail path"
	req.processUpload(ctx, a)
}

func TestProcessUploadIsEncryptedWithConsensusFailed(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()
	var filePath = uploadWorkerTestDir + "/alloc/1.txt"
	var remotePath = "/1.txt"
	scm := &mocks.StatusCallback{}
	scm.On("Started", "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", remotePath, OpUpload, 0).Once()
	scm.On("Error", "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", remotePath, OpUpload, errors.New("Upload failed: Consensus_rate:NaN, expected:10.000000")).Once()
	scm.On("Error", "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", remotePath, OpUpload, common.NewError("commit_consensus_failed", "Upload failed as there was no commit consensus")).Once()

	ctx := context.Background()
	var req = &UploadRequest{}
	req.filemeta = &UploadFileMeta{}
	req.filemeta.MimeType = "application/octet-stream"
	req.isEncrypted = true
	req.filepath = filePath
	req.remotefilepath = remotePath
	req.statusCallback = scm
	req.processUpload(ctx, a)
	scm.Test(t)
	scm.AssertExpectations(t)
}

func TestProcessUploadOpenFileFailed(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()
	var filePath = "x"
	scm := &mocks.StatusCallback{}
	scm.On("Error", "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", filePath, OpUpload, &common.Error{Code: "open_file_failed", Msg: "open x: no such file or directory"}).Once()

	ctx := context.Background()
	var req = &UploadRequest{}
	req.filemeta = &UploadFileMeta{}
	req.filemeta.MimeType = "plain/text"
	req.uploadMask = zboxutil.NewUint128(0xf)
	req.statusCallback = scm
	req.filepath = filePath
	req.processUpload(ctx, a)
	scm.Test(t)
	scm.AssertExpectations(t)
}

func TestProcessUploadChunksPerShard(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()
	var filePath = uploadWorkerTestDir + "/alloc/1.txt"
	var remotePath = "/1.txt"
	scm := &mocks.StatusCallback{}
	scm.On("Started", "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", remotePath, OpUpload, 18000).Once()
	scm.On("Completed", "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", remotePath, "", "application/octet-stream", 18000, 0).Once()
	scm.On("InProgress", "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", remotePath, OpUpload, 18000, []byte(nil)).Maybe()
	scm.On("Error", "69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f", remotePath, OpUpload, tm.AnythingOfType("*errors.errorString")).Maybe()
	willReturnCommitResult(&CommitResult{Success: true})

	defer willReturnCommitResult(nil)
	ctx := context.Background()
	var req = &UploadRequest{}
	req.filemeta = &UploadFileMeta{}
	req.filemeta.MimeType = "plain/text"
	req.datashards = 2
	req.parityshards = 2
	req.uploadMask = zboxutil.NewUint128(0xf)
	req.filemeta.Size = 9000
	req.Consensus = Consensus{
		consensusThresh: 50,
		fullconsensus:   4,
	}
	req.remotefilepath = remotePath
	req.filepath = filePath
	req.statusCallback = scm
	req.processUpload(ctx, a)
	scm.Test(t)
	scm.AssertExpectations(t)
}

func TestProcessUploadIsUpdate(t *testing.T) {
	// setup mock sdk
	_, _, blobberMocks, closeFn := setupMockInitStorageSDK(t, configDir, 4)
	defer closeFn()
	// setup mock allocation
	a, cncl := setupMockAllocation(t, allocationTestDir, blobberMocks)
	defer cncl()
	ctx := context.Background()
	var req = &UploadRequest{}
	req.filemeta = &UploadFileMeta{}
	req.filemeta.MimeType = "mime type"
	req.uploadMask = zboxutil.NewUint128(2)
	req.isUpdate = true
	req.processUpload(ctx, a)
}
