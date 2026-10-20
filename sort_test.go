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

func TestSortNumericUsesLeadingNumber(t *testing.T) {
	got := run(t, "10 apples\n9 pears\n", "sort", "-n").stdout

	if want := "9 pears\n10 apples\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSortNumericTreatsNanAsText(t *testing.T) {
	got := run(t, "3\nnan\ninf\n1\n2\n", "sort", "-n").stdout

	if want := "1\n2\n3\ninf\nnan\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
