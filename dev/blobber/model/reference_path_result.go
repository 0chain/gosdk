package model

import "context"

type ReferencePathResult struct {
	*ReferencePath
	LatestWM *WriteMarker `json:"latest_write_marker"`
}

func BuildReferencePathResult(rootRef *Ref) *ReferencePathResult {
	rootRef.CalculateHash(context.TODO()) //nolint

	refPath := &ReferencePath{Ref: rootRef}

	refsToProcess := []*ReferencePath{refPath}

	//convert Ref tree to ReferencePath tree
	for len(refsToProcess) > 0 {
		refToProcess := refsToProcess[0]
		refToProcess.Meta = refToProcess.Ref.GetListingData(context.TODO())
		if len(refToProcess.Ref.Children) > 0 {
			refToProcess.List = make([]*ReferencePath, len(refToProcess.Ref.Children))
		}
		for idx, child := range refToProcess.Ref.Children {
			childRefPath := &ReferencePath{Ref: child}
			refToProcess.List[idx] = childRefPath
			refsToProcess = append(refsToProcess, childRefPath)
		}
		refsToProcess = refsToProcess[1:]
	}

	return &ReferencePathResult{
		ReferencePath: refPath,
	}
}
