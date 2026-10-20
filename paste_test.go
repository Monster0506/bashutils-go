package main

import "testing"

func TestPasteSharesStdinAcrossDashes(t *testing.T) {
	got := run(t, "1\n2\n3\n4\n", "paste", "-", "-").stdout

	if want := "1\t2\n3\t4\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
