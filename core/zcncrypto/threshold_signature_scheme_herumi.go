//go:build !js && !wasm
// +build !js,!wasm

package zcncrypto

//NewHerumiThresholdScheme - create a new instance
func NewHerumiThresholdScheme() *HerumiThresholdScheme {
	return &HerumiThresholdScheme{
		id: blsInstance.NewID(),
	}
}

//HerumiThresholdScheme - a scheme that can create threshold signature shares for BLS0Chain signature scheme
type HerumiThresholdScheme struct {
	HerumiScheme
	id  ID
	Ids string `json:"threshold_scheme_id"`
}

//SetID sets ID in HexString format
func (mts *HerumiThresholdScheme) SetID(id string) error {
	mts.Ids = id
	return mts.id.SetHexString(id)
}

//GetID gets ID in hex string format
func (mts *HerumiThresholdScheme) GetID() string {
	return mts.id.GetHexString()
}
