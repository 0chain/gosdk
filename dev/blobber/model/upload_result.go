package model

type UploadResult struct {
	Filename        string `json:"filename"`
	ShardSize       int64  `json:"size"`
	ValidationRoot  string `json:"validation_root,omitempty"`
	FixedMerkleRoot string `json:"fixed_merkle_root,omitempty"`
}
