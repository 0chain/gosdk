package zcncrypto

import "github.com/0chain/gosdk/bls"

//WasmThresholdScheme - a scheme that can create threshold signature shares for BLS0Chain signature scheme
type WasmThresholdScheme struct {
	WasmScheme
	id  bls.ID
	Ids string `json:"threshold_scheme_id"`
}

//NewWasmThresholdScheme - create a new instance
func NewWasmThresholdScheme() *WasmThresholdScheme {
	return &WasmThresholdScheme{}
}

//SetID sets ID in HexString format
func (tss *WasmThresholdScheme) SetID(id string) error {
	tss.Ids = id
	return tss.id.SetHexString(id)
}

//GetID gets ID in hex string format
func (tss *WasmThresholdScheme) GetID() string {
	return tss.id.GetHexString()
}
