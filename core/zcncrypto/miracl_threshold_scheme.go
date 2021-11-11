package zcncrypto

import "github.com/herumi/bls-go-binary/bls"

//NewMiraclThresholdScheme - create a new instance
func NewMiraclThresholdScheme() *MiraclThresholdScheme {
	return &MiraclThresholdScheme{}
}

//MiraclThresholdScheme - a scheme that can create threshold signature shares for BLS0Chain signature scheme
type MiraclThresholdScheme struct {
	MiraclScheme
	id  bls.ID
	Ids string `json:"threshold_scheme_id"`
}

//SetID sets ID in HexString format
func (mts *MiraclThresholdScheme) SetID(id string) error {
	mts.Ids = id
	return mts.id.SetHexString(id)
}

//GetID gets ID in hex string format
func (mts *MiraclThresholdScheme) GetID() string {
	return mts.id.GetHexString()
}
