package main

import (
	"fmt"
	"github.com/herumi/bls-go-binary/bls"
)

var g_Qcoeff []uint64

func initQcoeff() {
	bls.Init(bls.BLS12_381)
	n := bls.GetUint64NumToPrecompute()
	g_Qcoeff = make([]uint64, n)
	var Q bls.PublicKey
	bls.BlsGetGeneratorOfPublicKey(&Q)
	bls.PrecomputeG2(g_Qcoeff, bls.CastFromPublicKey(&Q))
}

func toG(h []byte) (P bls.G1) {
	var t bls.Fp
	t.SetLittleEndian(h)
	bls.MapToG1(&P, &t)
	return P
}

func splitVAH_init(agg *bls.Sign) (e bls.GT) {
	P := *bls.CastFromSign(agg)
	bls.G1Neg(&P, &P)
	bls.PrecomputedMillerLoop(&e, &P, g_Qcoeff)
	return e
}

// return the processed number
func splitVAH_exec(e *bls.GT, pubs []bls.PublicKey, hs [][]byte) (m int) {
	N := 16
	g1s := make([]bls.G1, N)
	g2s := make([]bls.G2, N)
	m = len(pubs)
	if m > N {
		m = N
	}
	for i := 0; i < m; i++ {
		g1s[i] = toG(hs[i])
		g2s[i] = *bls.CastFromPublicKey(&pubs[i])
	}
	var e1 bls.GT
	bls.MillerLoopVec(&e1, g1s[0:m], g2s[0:m])
	bls.GTMul(e, e, &e1)
	return m
}

func splitVAH_final(e *bls.GT) bool {
	bls.FinalExp(e, e)
	return e.IsOne()
}

func splitVerifyAggregateHashes(agg *bls.Sign, pubs []bls.PublicKey, hs [][]byte) bool {
	n := len(pubs)
	if n != len(hs) {
		return false
	}

	e := splitVAH_init(agg)
	for {
		m := splitVAH_exec(&e, pubs, hs)
		pubs = pubs[m:]
		hs = hs[m:]
		n = n - m
		if n == 0 {
			break
		}
	}
	/*
		N := 16
		g1work := make([]bls.G1, N)
		g2work := make([]bls.G2, N)
		for {
			m := N
			if n < m {
				m = n
			}
			for i := 0; i < m; i++ {
				g1work[i] = toG(hs[i])
				g2work[i] = *bls.CastFromPublicKey(&pubs[i])
			}
			var e1 bls.GT
			bls.MillerLoopVec(&e1, g1work[0:m], g2work[0:m])
			bls.GTMul(&e, &e, &e1)
			pubs = pubs[m:]
			hs = hs[m:]
			n = n - m
			if n == 0 {
				break
			}
		}
	*/
	return splitVAH_final(&e)
}

func testVerifyAggregateHashes() {
	n := 100
	secs := make([]bls.SecretKey, n)
	pubs := make([]bls.PublicKey, n)
	hs := make([][]byte, n)
	sigs := make([]bls.Sign, n)
	for i := 0; i < n; i++ {
		hs[i] = make([]byte, 32)
		hs[i][0] = byte(i + 1)
		secs[i].SetByCSPRNG()
		pubs[i] = *secs[i].GetPublicKey()
		sigs[i] = *secs[i].SignHash(hs[i])
		if !sigs[i].VerifyHash(&pubs[i], hs[i]) {
			fmt.Printf("err %v\n", i)
		}
	}
	var agg bls.Sign
	agg.Aggregate(sigs)
	fmt.Printf("verify=%v\n", agg.VerifyAggregateHashes(pubs, hs))
	fmt.Printf("verify split=%v\n", splitVerifyAggregateHashes(&agg, pubs, hs))
}

func main() {
	initQcoeff()
	testVerifyAggregateHashes()
}
