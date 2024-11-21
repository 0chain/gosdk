// Miscellaneous utility functions.
package utils

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
)

func Encode(publicKey *ecdsa.PublicKey, privateKey *ecdsa.PrivateKey) (string, string) {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncoded), string(pemEncodedPub)
}
