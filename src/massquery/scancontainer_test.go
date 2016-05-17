package main

import (
	"testing"
)

func TestNewScanContainer(t *testing.T) {

	cases := []struct {
		inSize   int
		wantSize int
	}{
		{0, 0},
		{5, 5},
		{-1, 0},
	}

	for _, c := range cases {
		got := newScanContainer(c.inSize)
		sizePointers, sizeValues := len(got.Pointers), len(got.Values)
		if sizePointers != c.wantSize {
			t.Errorf("newScanContainer(%q) return Pointers with size %q, want %q", c.inSize, sizePointers, c.wantSize)
		}
		if sizeValues != c.wantSize {
			t.Errorf("newScanContainer(%q) return Values with size %q, want %q", c.inSize, sizeValues, c.wantSize)
		}
	}
}
