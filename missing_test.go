package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMissingFileIsReported(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.txt")

	cases := [][]string{
		{"cat"},
		{"head"},
		{"tail"},
		{"wc"},
		{"sort"},
		{"uniq"},
		{"cut", "-c", "1"},
		{"grep", "x"},
		{"paste"},
	}

	for _, args := range cases {
		r := run(t, "", append(args, missing)...)
		if !strings.Contains(r.stderr, "missing.txt") {
			t.Errorf("%v: stderr %q does not mention the file", args, r.stderr)
		}
	}
}

func TestCatDashReadsStdin(t *testing.T) {
	got := run(t, "hi\n", "cat", "-").stdout

	if want := "hi\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
