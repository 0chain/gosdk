package util

import (
	"testing"
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
