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

func TestWcDashReadsStdinAndLabelsIt(t *testing.T) {
	got := run(t, "a b\n", "wc", "-").stdout

	if want := "1 2 4 -\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWcTotalsMultipleFiles(t *testing.T) {
	f1 := writeFile(t, "a.txt", "a\n")
	f2 := writeFile(t, "b.txt", "b c\nd\n")

	got := run(t, "", "wc", "-lw", f1, f2).stdout

	want := "1 1 " + f1 + "\n2 3 " + f2 + "\n3 4 total\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
