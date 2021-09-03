package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRandomSlice(t *testing.T) {
	s := []string{"hello", "world", "beuty"}
	ns := GetRandom(s, 2)
	if len(ns) != 2 {
		t.Fatalf("Getrandom() failed")
	}
	ns = GetRandom(s, 4)
	if len(ns) != 3 {
		t.Fatalf("Getrandom() failed")
	}
}

func TestRand(t *testing.T) {
	s := []string{"a", "b", "c", "d", "e", "f", "h", "i", "j", "k"}

	r := NewRand(len(s))

	selected := make(map[int]string)

	for i := 0; i < len(s); i++ {
		index, err := r.Next()

		require.Equal(t, nil, err)

		_, ok := selected[index]

		require.Equal(t, false, ok)

		selected[index] = s[index]
	}

	for i := 0; i < len(s); i++ {
		require.Equal(t, s[i], selected[i])
	}

	_, err := r.Next()

	require.Equal(t, ErrNoItem, err)

}
