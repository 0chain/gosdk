package BN254

import (
	"fmt"
	"github.com/0chain/gosdk/miracl/core"
	herumi "github.com/herumi/bls-go-binary/bls"
	"testing"
)

func SecretKeyToStr(sec *herumi.SecretKey) string {
	s := herumi.CastFromSecretKey(sec).GetString(16)
	return fmt.Sprintf("%064s", s)
}

func G1ToStr(P *herumi.G1) string {
	herumi.G1Normalize(P, P)
	return fmt.Sprintf("(%064s,%064s)", P.X.GetString(16), P.Y.GetString(16))
}

func SignToStr(sig *herumi.Sign) string {
	P := herumi.CastFromSign(sig)
	return G1ToStr(P)
}

func testMul(sStr string, xStr string, yStr string) {
	fmt.Printf("testMul")
	s := HexStrToBIG(sStr) //"0f535b6c36d87d91a4e13aa9810fb95e9ae79df5d64581345ec86503be613af7")
	x := HexStrToBIG(xStr) //"1b45dadc6bfb5ee1ed1ccc95ab154f4acc8d6ffca954fac1297f4ab33d89f311")
	y := HexStrToBIG(yStr) //"0b79d0f08ebc1e07f1df0d3d1a23c7438147976534cc21317d7cdef2dcc6ed37")
	P := NewECPbigs(x, y)
	fmt.Printf("P=%v\n", P.ToString())
	fmt.Printf("s=%v\n", s.ToString())
	sP := P.Mul(s)
	fmt.Printf("sP=%v\n", sP.ToString())
	fmt.Printf("---\n")
}

func TestMain(t *testing.T) {
	if herumi.Init(herumi.CurveFp254BNb) != nil {
		t.Fatal("herumi.Init")
	}
	if Init() == BLS_FAIL {
		fmt.Printf("err")
		return
	}
	// initialize at once
	H := NewHashAndMap()

	var oneSec herumi.SecretKey
	oneSec.SetHexString("1")
	var P2, Q2 herumi.G1

	const N = 1000
	for i := 0; i < N; i++ {
		key := fmt.Sprintf("sec%v\n", i)
		hash := core.NewHASH256()
		hash.Process_array([]byte(key))
		md := hash.Hash()

		var sec2 herumi.SecretKey

		// check secret key
		sec1 := H.copyAndMask(md)
		sec2.SetLittleEndian(md)

		str1 := sec1.ToString()
		str2 := SecretKeyToStr(&sec2)
		if str1 != str2 {
			t.Errorf("bad str i=%v\ns1=%v\ns2=%v\n", i, str1, str2)
		}

		msg := []byte(fmt.Sprintf("msg%v\n", i))

		// check hash-and-map function
		P1 := H.SetHashOf(msg)
		P2.HashAndMapTo(msg)

		str1 = P1.ToString()
		str2 = G1ToStr(&P2)
		if str1 != str2 {
			t.Errorf("bad map i=%v\nsig1=%s\nsig2=%s\n", i, str1, str2)
		}

		// check mul
		Q1 := P1.Mul(sec1)
		herumi.G1Mul(&Q2, &P2, herumi.CastFromSecretKey(&sec2))

		str1 = Q1.ToString()
		str2 = G1ToStr(&Q2)
		if str1 != str2 {
			t.Errorf("bad sig i=%v s=%s\nP=%s\nsig1=%s\nsig2=%s\n", i, sec1.ToString(), P1.ToString(), str1, str2)
		}
	}

	for i := 0; i < N; i++ {
		x := NewFPint(i)
		r := H.sq.Get(x)
		if r != nil {
			r.sqr()
			if !r.Equals(x) {
				fmt.Printf("err")
			}
		}
	}
}
