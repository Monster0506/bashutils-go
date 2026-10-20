package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func splitOutputs(t *testing.T, dir, prefix string) (names []string, joined string) {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(dir, prefix+"*"))
	if err != nil {
		t.Fatal(err)
	}

	var sb strings.Builder
	for _, m := range matches {
		data, err := os.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, filepath.Base(m))
		sb.Write(data)
	}

	return names, sb.String()
}

func TestSplitFixedWidthSuffixes(t *testing.T) {
	var in strings.Builder
	for i := range 30 {
		fmt.Fprintf(&in, "line%d\n", i)
	}

	src := writeFile(t, "in.txt", in.String())
	dir := t.TempDir()

	run(t, "", "split", "-l", "1", src, filepath.Join(dir, "p"))

	names, joined := splitOutputs(t, dir, "p")
	if len(names) != 30 || names[0] != "paa" || names[1] != "pab" || names[26] != "pba" {
		t.Fatalf("unexpected names: %v", names)
	}
	if joined != in.String() {
		t.Fatalf("sorted pieces do not reassemble the input")
	}
}

func TestSplitBytesTerminates(t *testing.T) {
	src := writeFile(t, "in.txt", "abcdefghij")
	dir := t.TempDir()

	run(t, "", "split", "-b", "4", src, filepath.Join(dir, "p"))

	names, joined := splitOutputs(t, dir, "p")
	if len(names) != 3 || joined != "abcdefghij" {
		t.Fatalf("names %v, joined %q", names, joined)
	}
}

func TestSplitBytesFromStdin(t *testing.T) {
	dir := t.TempDir()

	run(t, "abcdefghij", "split", "-b", "4", "-", filepath.Join(dir, "p"))

	names, joined := splitOutputs(t, dir, "p")
	if len(names) != 3 || joined != "abcdefghij" {
		t.Fatalf("names %v, joined %q", names, joined)
	}
}

func TestSplitBytesRejectsZero(t *testing.T) {
	src := writeFile(t, "in.txt", "abc")
	dir := t.TempDir()

	r := run(t, "", "split", "-b", "0", src, filepath.Join(dir, "p"))

	names, _ := splitOutputs(t, dir, "p")
	if len(names) != 0 || r.stderr == "" {
		t.Fatalf("names %v, stderr %q", names, r.stderr)
	}
}
