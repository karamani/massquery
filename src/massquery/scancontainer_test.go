package main

import (
	"database/sql"
	"strconv"
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

func TestAsString(t *testing.T) {

	size := 5

	c := newScanContainer(size)

	for i := range c.Pointers {
		*c.Pointers[i].(*sql.RawBytes) = sql.RawBytes(strconv.Itoa(i))
	}

	if len(c.AsStrings()) != size {
		t.Errorf("scanContainer.AsString: size = %d, want %d", len(c.AsStrings()), size)
	}

	for i, s := range c.AsStrings() {
		if strconv.Itoa(i) != s {
			t.Errorf("scanContainer.AsString: got %s, want %s", s, strconv.Itoa(i))
		}
	}
}
