package main

import "testing"

func TestWcCharsCountsAllRunes(t *testing.T) {
	got := run(t, "héllo wörld\n", "wc", "-m").stdout

	if want := "12\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWcDefaultColumns(t *testing.T) {
	got := run(t, "a b\n", "wc").stdout

	if want := "1 2 4\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
