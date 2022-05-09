package main

import (
	"fmt"
	"github.com/herumi/bls-go-binary/bls"
)

func PutPub(pub *bls.PublicKey) {
	g2 := bls.CastFromPublicKey(pub)
	bls.G2Normalize(g2, g2)
	fmt.Println("x.a", g2.X.D[0].GetString(16))
	fmt.Println("x.b", g2.X.D[1].GetString(16))
	fmt.Println("y.a", g2.Y.D[0].GetString(16))
	fmt.Println("y.b", g2.Y.D[1].GetString(16))
}

func main() {
	bls.Init(bls.CurveFp254BNb)
	var sk bls.SecretKey
	sk.SetHexString("1")
	pk := sk.GetPublicKey()
	fmt.Println("sk", sk.GetHexString())
	fmt.Println("pk", pk.GetHexString())
	PutPub(pk)
}
