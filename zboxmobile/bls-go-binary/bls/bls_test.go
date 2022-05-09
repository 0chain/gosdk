package bls

import (
	"fmt"
	"testing"
)

// e(P1, Q1) == e(P2, Q2)
func comparePairing1(P1 *G1, Q1 *G2, P2 *G1, Q2 *G2) bool {
	var e1, e2 GT
	Pairing(&e1, P1, Q1)
	Pairing(&e2, P2, Q2)
	return e1.IsEqual(&e2)
}

// FinalExp(ML(P1, Q1) ML(-P2, Q2)) == 1
func comparePairing2(P1 *G1, Q1 *G2, P2 *G1, Q2 *G2) bool {
	var e1, e2 GT
	MillerLoop(&e1, P1, Q1)
	var negP2 G1
	G1Neg(&negP2, P2)
	MillerLoop(&e2, &negP2, Q2)
	GTMul(&e1, &e1, &e2)
	FinalExp(&e1, &e1)
	return e1.IsOne()
}

// FinalExp(MLvec(P1, Q1, -P2, Q2)) == 1
func comparePairing3(P1 *G1, Q1 *G2, P2 *G1, Q2 *G2) bool {
	var e GT
	v1 := make([]G1, 2)
	v2 := make([]G2, 2)
	v1[0] = *P1
	G1Neg(&v1[1], P2)
	v2[0] = *Q1
	v2[1] = *Q2
	MillerLoopVec(&e, v1, v2)
	FinalExp(&e, &e)
	return e.IsOne()
}

// set (P1, Q1, P2, Q2) s.t. e(P1, Q1) == e(P2, Q2)
func initPQ(P1 *G1, Q1 *G2, P2 *G1, Q2 *G2) {
	P1.HashAndMapTo([]byte("abc"))
	Q1.HashAndMapTo([]byte("abc"))
	var a Fr
	var rev Fr
	a.SetInt64(123)
	FrInv(&rev, &a)
	G1Mul(P2, P1, &a)
	G2Mul(Q2, Q1, &rev)
}

func TestSign(t *testing.T) {
	Init(BLS12_381)
	var sec SecretKey
	sec.SetByCSPRNG()
	fmt.Printf("sec:%s\n", sec.SerializeToHexStr())
	pub := sec.GetPublicKey()
	fmt.Printf("pub:%s\n", pub.SerializeToHexStr())
	msgTbl := []string{"abc", "def", "123"}
	n := len(msgTbl)
	sigVec := make([]*Sign, n)
	for i := 0; i < n; i++ {
		m := msgTbl[i]
		sigVec[i] = sec.Sign(m)
		fmt.Printf("%d. sign(%s)=%s\n", i, m, sigVec[i].SerializeToHexStr())
	}
	agg := sigVec[0]
	for i := 1; i < n; i++ {
		agg.Add(sigVec[i])
	}
	hashPt := HashAndMapToSignature([]byte(msgTbl[0]))
	for i := 1; i < n; i++ {
		hashPt.Add(HashAndMapToSignature([]byte(msgTbl[i])))
	}
	if !VerifyPairing(agg, hashPt, pub) {
		t.Error("VerifyPairing")
	}
}

func TestPairing(t *testing.T) {
	Init(BLS12_381)
	var P1, P2 G1
	var Q1, Q2 G2
	initPQ(&P1, &Q1, &P2, &Q2)
	var b1, b2, b3 bool
	b1 = comparePairing1(&P1, &Q1, &P2, &Q2)
	b2 = comparePairing2(&P1, &Q1, &P2, &Q2)
	b3 = comparePairing3(&P1, &Q1, &P2, &Q2)
	if !(b1 && b2 && b3) {
		t.Error("must be true")
	}
	G1Dbl(&P1, &P1)
	// e(P1, Q1) != e(P2, Q2)
	b1 = comparePairing1(&P1, &Q1, &P2, &Q2)
	b2 = comparePairing2(&P1, &Q1, &P2, &Q2)
	b3 = comparePairing3(&P1, &Q1, &P2, &Q2)
	if b1 || b2 || b3 {
		t.Error("must be false")
	}
}

func TestCast(t *testing.T) {
	Init(BLS12_381)
	var sec SecretKey
	sec.SetByCSPRNG()
	{
		x := *CastFromSecretKey(&sec)
		sec2 := *CastToSecretKey(&x)
		if !sec.IsEqual(&sec2) {
			t.Error("sec is not equal")
		}
	}
	pub := *sec.GetPublicKey()
	g2 := *CastFromPublicKey(&pub)
	G2Add(&g2, &g2, &g2)
	pub.Add(&pub)
	if !pub.IsEqual(CastToPublicKey(&g2)) {
		t.Error("pub not equal")
	}
	sig := sec.Sign("abc")
	g1 := *CastFromSign(sig)
	G1Add(&g1, &g1, &g1)
	sig.Add(sig)
	if !sig.IsEqual(CastToSign(&g1)) {
		t.Error("sig not equal")
	}
}


func BenchmarkPairing(b *testing.B) {
	Init(BLS12_381)
	var P G1
	var Q G2
	var e GT
	P.HashAndMapTo([]byte("abc"))
	Q.HashAndMapTo([]byte("abc"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Pairing(&e, &P, &Q)
	}
}

func BenchmarkPairing1(b *testing.B) {
	var P1, P2 G1
	var Q1, Q2 G2
	initPQ(&P1, &Q1, &P2, &Q2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comparePairing1(&P1, &Q1, &P2, &Q2)
	}
}

func BenchmarkPairing2(b *testing.B) {
	var P1, P2 G1
	var Q1, Q2 G2
	initPQ(&P1, &Q1, &P2, &Q2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comparePairing2(&P1, &Q1, &P2, &Q2)
	}
}

func BenchmarkPairing3(b *testing.B) {
	var P1, P2 G1
	var Q1, Q2 G2
	initPQ(&P1, &Q1, &P2, &Q2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comparePairing3(&P1, &Q1, &P2, &Q2)
	}
}
