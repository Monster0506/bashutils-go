package main

import "testing"

func TestTrDeleteComplement(t *testing.T) {
	got := run(t, "abcxyz\n", "tr", "-d", "-c", "a-c").stdout

	if want := "abc"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTrTranslateComplement(t *testing.T) {
	got := run(t, "abcxyz", "tr", "-c", "a-c", "_").stdout

	if want := "abc___"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
