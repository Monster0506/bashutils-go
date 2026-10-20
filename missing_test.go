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

func TestMissingFileExitsNonZero(t *testing.T) {
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
		{"split"},
	}

	for _, args := range cases {
		if r := run(t, "", append(args, missing)...); r.code == 0 {
			t.Errorf("%v: exit code 0 for a missing file", args)
		}
	}
}

func TestMissingFileDoesNotStopOtherFiles(t *testing.T) {
	good := writeFile(t, "good.txt", "hello\n")
	missing := filepath.Join(t.TempDir(), "missing.txt")

	r := run(t, "", "cat", missing, good)

	if r.stdout != "hello\n" || r.code == 0 {
		t.Fatalf("stdout %q, code %d", r.stdout, r.code)
	}
}
