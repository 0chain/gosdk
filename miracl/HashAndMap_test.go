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

func SignToStr(sig *herumi.Sign) string {
	P := herumi.CastFromSign(sig)
	herumi.G1Normalize(P, P)
	return fmt.Sprintf("(%064s,%064s)", P.X.GetString(16), P.Y.GetString(16))
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

	const N = 1000
	for i := 0; i < N; i++ {
		key := fmt.Sprintf("sec%v\n", i)
		hash := core.NewHASH256()
		hash.Process_array([]byte(key))
		md := hash.Hash()

		sec1 := H.copyAndMask(md)
		var sec2 herumi.SecretKey
		sec2.SetLittleEndian(md)
		s1 := sec1.ToString()
		s2 := SecretKeyToStr(&sec2)
		if s1 != s2 {
			t.Errorf("bad str i=%v\ns1=%v\ns2=%v\n", i, s1, s2)
		}

		msg := []byte(fmt.Sprintf("msg%v\n", i))

		P := H.SetHashOf(msg)
		sig1str := P.ToString()
		sig2 := oneSec.SignByte(msg)
		sig2str := SignToStr(sig2)
		if sig1str != sig2str {
			t.Errorf("bad map i=%v\nsig1=%s\nsig2=%s\n", i, sig1str, sig2str)
		}
		/*
			P = P.Mul(sec1.GetBIG())
			sig1str = P.ToString()
			sig2 = sec2.SignByte(msg)
			sig2str = SignToStr(sig2)
			if sig1str != sig2str {
				t.Errorf("bad sig i=%v s=%s\nsig1=%s\nsig2=%s\n", i, s1, sig1str, sig2str)
			}
		*/
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
