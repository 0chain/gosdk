package model

import "github.com/0chain/gosdk/core/common"

type WriteMarker struct {
	AllocationRoot         string           `gorm:"column:allocation_root;primary_key" json:"allocation_root"`
	PreviousAllocationRoot string           `gorm:"column:prev_allocation_root" json:"prev_allocation_root"`
	AllocationID           string           `gorm:"column:allocation_id" json:"allocation_id"`
	Size                   int64            `gorm:"column:size" json:"size"`
	BlobberID              string           `gorm:"column:blobber_id" json:"blobber_id"`
	Timestamp              common.Timestamp `gorm:"column:timestamp" json:"timestamp"`
	ClientID               string           `gorm:"column:client_id" json:"client_id"`
	Signature              string           `gorm:"column:signature" json:"signature"`
}
