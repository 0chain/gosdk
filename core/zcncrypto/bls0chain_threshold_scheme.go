package zcncrypto

import "github.com/0chain/gosdk/bls"

//BLS0ChainThresholdScheme - a scheme that can create threshold signature shares for BLS0Chain signature scheme
type BLS0ChainThresholdScheme struct {
	BLS0ChainScheme
	id  bls.ID
	Ids string `json:"threshold_scheme_id"`
}

//NewBLS0ChainThresholdScheme - create a new instance
func NewBLS0ChainThresholdScheme() *BLS0ChainThresholdScheme {
	return &BLS0ChainThresholdScheme{}
}

//SetID sets ID in HexString format
func (tss *BLS0ChainThresholdScheme) SetID(id string) error {
	tss.Ids = id
	return tss.id.SetHexString(id)
}

//GetID gets ID in hex string format
func (tss *BLS0ChainThresholdScheme) GetID() string {
	return tss.id.GetHexString()
}
