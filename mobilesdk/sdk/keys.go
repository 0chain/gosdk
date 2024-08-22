package sdk

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/zcncrypto"
)

// SplitKeys split keys by private key, return json string of wallet
//
//	 {
//			"client_id":"3943b66200f4702145596de769c5137b333b42bf67c665677355843f233009e2",
//			"client_key":"3423af4e1d19b99d5997a7ce5ee9fb46372a5e6a2af18e4640e1aeb3fe116f18580e723ae6d6527ff4cf81b537d4bbf5de670169654fa8f9e54291fa930f9f9a",
//			"keys":[{
//				"public_key":"3423af4e1d19b99d5997a7ce5ee9fb46372a5e6a2af18e4640e1aeb3fe116f18580e723ae6d6527ff4cf81b537d4bbf5de670169654fa8f9e54291fa930f9f9a",
//				"private_key":"69a59735f39eae40b80d53a73c5d7b161998da0b7cea44c94672a01c9128da09"
//				}
//			],
//			"mnemonics":"guitar cook depend snake response major argue abstract script tonight energy soccer copy latin street bracket combine image fabric maximum mercy icon bitter dwarf",
//			"version":"1.0","date_created":"2022-11-02T10:21:32+08:00","nonce":0
//	 }
//		- privateKey: private key to use for signing
//		- signatureScheme: signature scheme to use for signing
//		- numSplits: number of splits
func SplitKeys(privateKey, signatureScheme string, numSplits int) ([]byte, error) {
	signScheme := zcncrypto.NewSignatureScheme(signatureScheme)
	if signScheme != nil {
		err := signScheme.SetPrivateKey(privateKey)
		if err != nil {
			return nil, err
		}
		wallet, err := signScheme.SplitKeys(numSplits)
		if err != nil {
			return nil, err

		}

		return json.Marshal(wallet)
	}
	return nil, ErrInvalidSignatureScheme
}
