package util

import (
	"github.com/0chain/gosdk/zboxcore/logger"

	"github.com/google/uuid"
)

// GetSHA1Uuid will give version 5 uuid. It depends on already existing uuid.
// The main idea is to synchronize uuid among blobbers. So for example, if
// a file is being uploaded to 4 blobbers then we require that file in each blobber
// have same uuid. All the goroutine that assigns uuid, calculates hashes and commits to blobber
// will use initial version 1 uuid and updates the uuid with recently calculated uuid so that
// the file will get same uuid in all blobbers.
func GetSHA1Uuid(u uuid.UUID, name string) uuid.UUID {
	return uuid.NewSHA1(u, []byte(name))
}

// GetNewUUID will give new version1 uuid. It will ignore first error if any but will panic
// if there is error twice.
func GetNewUUID() uuid.UUID {
	var uid uuid.UUID
	var err error
	for i := 0; i < 2; i++ {
		uid, err = uuid.NewUUID()
		if err == nil {
			break
		}
		logger.Logger.Error(err)

	}
	if err != nil {
		panic("could not get new uuid. Error: " + err.Error())
	}
	return uid
}
