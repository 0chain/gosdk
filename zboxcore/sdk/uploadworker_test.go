package sdk

import (
	"fmt"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMaxBlobbersRequiredGreaterThanImplicitLimit128(t *testing.T) {
	var maxNumOfBlobbers = 129

	var req = &UploadRequest{}
	req.setUploadMask(maxNumOfBlobbers)
	req.fullconsensus = float32(maxNumOfBlobbers)

	if req.IsFullConsensusSupported() {
		t.Errorf("IsFullConsensusSupported() = %v, want %v", true, false)
	}
}

func TestNaxBlobbersRequiredEqualToImplicitLimit32(t *testing.T) {
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

func TestMarketplaceApi(t *testing.T) {
	mnemonic, err := zboxutil.GetMarketplaceMnemonic("http://localhost:5051")
	if err != nil {
		fmt.Println("Got error", err)
		return
		// handle error
	}
	assert.NotEmpty(t, mnemonic)
}