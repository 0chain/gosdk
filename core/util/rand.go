package util

import (
	"errors"
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

func Shuffle(in []string) (shuffle []string) {
	shuffle = make([]string, len(in))
	copy(shuffle, in)
	var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(len(in), func(i, j int) {
		shuffle[i], shuffle[j] = shuffle[j], shuffle[i]
	})
	return
}

// GetRandom returns n random slice from in
// If n > len(in), then this will return a shuffled version of in
func GetRandom(in []string, n int) []string {
	n = MinInt(len(in), n)
	out := make([]string, 0)

	rand.Seed(time.Now().UnixNano()) //nolint
	perm := rand.Perm(len(in))
	for i := 0; i < n; i++ {
		out = append(out, in[perm[i]])
	}
	return out
}

var (
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	// ErrNoItem there is no item anymore
	ErrNoItem = errors.New("rand: there is no item anymore")
)

// Rand a progressive rand
type Rand struct {
	items []int
}

// Next get next random item
func (r *Rand) Next() (int, error) {
	if len(r.items) > 0 {
		i := randGen.Intn(len(r.items))

		it := r.items[i]

		copy(r.items[i:], r.items[i+1:])
		r.items = r.items[:len(r.items)-1]

		return it, nil
	}

	return -1, ErrNoItem

}

// NewRand create a ProgressiveRand instance
func NewRand(max int) Rand {
	r := Rand{
		items: make([]int, max),
	}

	for i := 0; i < max; i++ {
		r.items[i] = i
	}

	return r

}
