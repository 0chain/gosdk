package fileref

import (
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/mitchellh/mapstructure"
)

type ListResult struct {
	AllocationRoot string                   `json:"allocation_root"`
	Meta           map[string]interface{}   `json:"meta_data"`
	Entities       []map[string]interface{} `json:"list"`
}

func (lr *ListResult) GetDirTree(allocationID string) (*Ref, error) {
	reftype := lr.Meta["type"].(string)
	if reftype == DIRECTORY {
		rootRef := &Ref{Type: DIRECTORY}
		rootRef.AllocationID = allocationID

		rootRef.CreatedAt = parseTime(lr.Meta["created_at"].(string))
		rootRef.UpdatedAt = parseTime(lr.Meta["updated_at"].(string))
		delete(lr.Meta, "created_at")
		delete(lr.Meta, "updated_at")

		var md mapstructure.Metadata
		config := &mapstructure.DecoderConfig{
			Metadata: &md,
			Result:   rootRef,
			TagName:  "json",
		}
		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return nil, err
		}
		err = decoder.Decode(lr.Meta)
		if err != nil {
			return nil, err
		}
		err = lr.populateChildren(rootRef)
		if err != nil {
			return nil, err
		}
		return rootRef, nil
	}
	return nil, common.NewError("invalid_list_path", "Invalid list path. list was not for a directory")
}

func (lr *ListResult) populateChildren(ref *Ref) error {
	for _, rpc := range lr.Entities {
		reftype := rpc["type"].(string)
		var childEntity RefEntity
		if reftype == DIRECTORY {
			dref := &Ref{Type: DIRECTORY}
			dref.AllocationID = ref.AllocationID
			dref.CreatedAt = parseTime(rpc["created_at"].(string))
			dref.UpdatedAt = parseTime(rpc["updated_at"].(string))
			childEntity = dref
		} else {
			fref := &FileRef{}
			fref.Type = FILE
			fref.AllocationID = ref.AllocationID
			fref.CreatedAt = parseTime(rpc["created_at"].(string))
			fref.UpdatedAt = parseTime(rpc["updated_at"].(string))
			childEntity = fref
		}
		delete(rpc, "created_at")
		delete(rpc, "updated_at")
		var md mapstructure.Metadata
		config := &mapstructure.DecoderConfig{
			Metadata: &md,
			Result:   childEntity,
			TagName:  "json",
		}
		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return err
		}
		err = decoder.Decode(rpc)
		if err != nil {
			return err
		}
		ref.Children = append(ref.Children, childEntity)
	}
	return nil
}

func parseTime(timeString string) time.Time {
	layout := "2006-01-02T15:04:05.000000Z"
	t, _ := time.Parse(layout, timeString)
	return t
}
