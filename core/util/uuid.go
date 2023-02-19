package util

import (
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

// GetNewUUID will give new version1 uuid. It will panic if any error occurred
func GetNewUUID() uuid.UUID {
	uid, err := uuid.NewUUID()
	if err != nil {
		panic("could not get new uuid. Error: " + err.Error())
	}
	return uid
}
