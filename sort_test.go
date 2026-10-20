package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestSortIsStable(t *testing.T) {
	var in, want strings.Builder
	for i := range 100 {
		fmt.Fprintf(&in, "k%d line%03d\n", i%3, i)
	}
	for k := range 3 {
		for i := k; i < 100; i += 3 {
			fmt.Fprintf(&want, "k%d line%03d\n", k, i)
		}
	}

	got := run(t, in.String(), "sort", "-k", "1").stdout

	if got != want.String() {
		t.Fatalf("equal keys were reordered:\n%s", got)
	}
}
