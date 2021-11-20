// +build !js,!wasm

package zcncrypto

import "github.com/herumi/bls-go-binary/bls"

//NewHerumiThresholdScheme - create a new instance
func NewHerumiThresholdScheme() *HerumiThresholdScheme {
	return &HerumiThresholdScheme{}
}

//HerumiThresholdScheme - a scheme that can create threshold signature shares for BLS0Chain signature scheme
type HerumiThresholdScheme struct {
	HerumiScheme
	id  bls.ID
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
