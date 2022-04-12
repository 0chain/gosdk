package model

type CommitResult struct {
	AllocationRoot string       `json:"allocation_root"`
	WriteMarker    *WriteMarker `json:"write_marker"`
	Success        bool         `json:"success"`
	ErrorMessage   string       `json:"error_msg,omitempty"`
	//	Changes        []*allocation.AllocationChange `json:"-"`
	//Result         []*UploadResult         `json:"result"`
}
