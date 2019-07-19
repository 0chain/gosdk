package util

import (
	"math/rand"
	"time"
)

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func MinInt64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func MaxInt64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func checkExists(c string, sl []string) bool {
	for _, s := range sl {
		if s == c {
			return true
		}
	}
	return false
}

//GetRandom returns n random slice from in
func GetRandom(in []string, n int) []string {
	out := make([]string, 0)
	nOut := MinInt(len(in), n)
	nOut = MaxInt(1, nOut)
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		i := randGen.Intn(len(in))
		if !checkExists(in[i], out) {
			out = append(out, in[i])
		}
		if len(out) >= nOut {
			break
		}
	}
	return out
}
