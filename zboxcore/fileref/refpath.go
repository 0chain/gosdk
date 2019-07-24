package fileref

import (
	"github.com/0chain/gosdk/core/common"
	"github.com/mitchellh/mapstructure"
)

type ReferencePath struct {
	Meta map[string]interface{} `json:"meta_data"`
	List []*ReferencePath       `json:"list,omitempty"`
}

func (rp *ReferencePath) GetRefFromObjectTree(allocationID string) (RefEntity, error) {
	reftype := rp.Meta["type"].(string)
	if reftype == FILE {
		rootRef := &FileRef{}
		rootRef.Type = FILE
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
		err = decoder.Decode(rp.Meta)
		if err != nil {
			return nil, err
		}
		return rootRef, nil
	}
	return rp.GetDirTree(allocationID)
}

func (rp *ReferencePath) GetDirTree(allocationID string) (*Ref, error) {
	reftype := rp.Meta["type"].(string)
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
		err = decoder.Decode(rp.Meta)
		if err != nil {
			return nil, err
		}
		err = rp.populateChildren(rootRef)
		if err != nil {
			return nil, err
		}
		return rootRef, nil
	}
	return nil, common.NewError("invalid_ref_path", "Invalid reference path. root was not a directory type")
}

func (rp *ReferencePath) populateChildren(ref *Ref) error {
	for _, rpc := range rp.List {
		reftype := rpc.Meta["type"].(string)
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
		err = decoder.Decode(rpc.Meta)
		if err != nil {
			return err
		}
		ref.AddChild(childEntity) //append(ref.Children, childEntity)
		if childEntity.GetType() == DIRECTORY && rpc.List != nil && len(rpc.List) > 0 {
			err = rpc.populateChildren(childEntity.(*Ref))
		}
		if err != nil {
			return err
		}
	}
	return nil
}
