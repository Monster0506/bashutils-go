package main

import (
	"path/filepath"
	"testing"
)

func TestGrepMultiFilePrefixesNames(t *testing.T) {
	f1 := writeFile(t, "f1.txt", "a\nb\n")
	f2 := writeFile(t, "f2.txt", "xa\n")

	got := run(t, "", "grep", "-n", "a", f1, f2).stdout

	want := f1 + ":1:a\n" + f2 + ":1:xa\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGrepSingleFileHasNoPrefix(t *testing.T) {
	f := writeFile(t, "f.txt", "a\nb\n")

	got := run(t, "", "grep", "a", f).stdout

	if want := "a\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGrepMissingFileAmongOthers(t *testing.T) {
	good := writeFile(t, "good.txt", "a\n")
	missing := filepath.Join(t.TempDir(), "missing.txt")

	r := run(t, "", "grep", "a", missing, good)

	if r.code != 2 || r.stdout != good+":a\n" {
		t.Fatalf("code %d, stdout %q", r.code, r.stdout)
	}
}

func TestGrepEmptyPatternViaE(t *testing.T) {
	r := run(t, "a\nb\n", "grep", "-e", "")

	if r.code != 0 || r.stdout != "a\nb\n" {
		t.Fatalf("code %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}
