package sdk

import (
	"context"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
)

// CreateDirectory create intermediate directories on blobbers
func CreateDirectory(ctx context.Context, allocationID string, dir string) error {

	if allocationID == "" {
		return errors.Throw(constants.ErrInvalidParameter, "allocationID")
	}

	if dir == "" {
		return errors.Throw(constants.ErrInvalidParameter, "dir")
	}

	alloc, err := GetAllocation(allocationID)

	if err != nil {
		return err
	}

	alloc.CreateDir(dir)

	return nil
}
