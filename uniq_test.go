package main

import "testing"

func TestUniqKeepsOrderAndAdjacency(t *testing.T) {
	got := run(t, "b\nb\na\nb\n", "uniq").stdout

	if want := "b\na\nb\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
