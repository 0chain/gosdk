package fileref

import (
	"github.com/0chain/errors"
	"github.com/mitchellh/mapstructure"
)

type ListResult struct {
	AllocationRoot string                   `json:"allocation_root"`
	Meta           map[string]interface{}   `json:"meta_data"`
	Entities       []map[string]interface{} `json:"list"`
}

func (lr *ListResult) GetDirTree(allocationID string) (*Ref, error) {
	if lr.Meta == nil {
		return nil, errors.New("invalid_list_path", "badly formatted list result, nil meta")
	}
	reftype := lr.Meta["type"].(string)
	if reftype == DIRECTORY {
		rootRef := &Ref{Type: DIRECTORY}
		rootRef.AllocationID = allocationID
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
	return nil, errors.New("invalid_list_path", "Invalid list path. list was not for a directory")
}

func (lr *ListResult) populateChildren(ref *Ref) error {
	for _, rpc := range lr.Entities {
		reftype := rpc["type"].(string)
		var childEntity RefEntity
		if reftype == DIRECTORY {
			dref := &Ref{Type: DIRECTORY}
			dref.AllocationID = ref.AllocationID
			childEntity = dref
		} else {
			fref := &FileRef{}
			fref.Type = FILE
			fref.AllocationID = ref.AllocationID
			childEntity = fref
		}
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
