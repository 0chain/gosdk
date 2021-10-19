package sdk

import (
	"context"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/sdks/blobber"
)

// CreateDir create intermediate directories on blobbers
func CreateDir(ctx context.Context, allocationID string, name string) error {
	if allocationID == "" {
		return errors.Throw(constants.ErrInvalidParameter, "allocationID")
	}

	if name == "" {
		return errors.Throw(constants.ErrInvalidParameter, "name")
	}

	alloc, err := GetAllocation(allocationID)

	if err != nil {
		return err
	}

	b := blobber.New(zbox, alloc.getBlobberUrls()...)

	return b.CreateDir(allocationID, name)
}
