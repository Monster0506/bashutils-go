package main

import "testing"

func TestCutKeepsLinesWithoutDelimiter(t *testing.T) {
	got := run(t, "abc\na,b\n", "cut", "-d", ",", "-f", "2").stdout

	if want := "abc\nb\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
