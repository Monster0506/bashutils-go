package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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

func TestSplitLinesKeepsMissingFinalNewline(t *testing.T) {
	src := writeFile(t, "in.txt", "a\nb")
	dir := t.TempDir()

	run(t, "", "split", "-l", "1", src, filepath.Join(dir, "p"))

	_, joined := splitOutputs(t, dir, "p")
	if joined != "a\nb" {
		t.Fatalf("got %q", joined)
	}
}

func splitNames(t *testing.T, args []string, in string) ([]string, result) {
	t.Helper()

	src := writeFile(t, "in.txt", in)
	dir := t.TempDir()

	full := append([]string{"split"}, args...)
	full = append(full, src, filepath.Join(dir, "p"))
	r := run(t, "", full...)

	names, _ := splitOutputs(t, dir, "p")
	return names, r
}

func TestSplitSuffixOptions(t *testing.T) {
	cases := []struct {
		args []string
		want []string
	}{
		{[]string{"-l", "1", "-a", "3"}, []string{"paaa", "paab"}},
		{[]string{"-l", "1", "-d"}, []string{"p00", "p01"}},
		{[]string{"-l", "1", "-d", "-a", "3"}, []string{"p000", "p001"}},
		{[]string{"-l", "1", "--additional-suffix=.txt"}, []string{"paa.txt", "pab.txt"}},
		{[]string{"-b", "2", "-d"}, []string{"p00", "p01"}},
	}

	for _, c := range cases {
		names, r := splitNames(t, c.args, "a\nb\n")
		if !reflect.DeepEqual(names, c.want) || r.code != 0 {
			t.Errorf("split %v: got %v (code %d), want %v", c.args, names, r.code, c.want)
		}
	}
}

func TestSplitSuffixExhaustion(t *testing.T) {
	var in strings.Builder
	for i := range 11 {
		fmt.Fprintf(&in, "line%d\n", i)
	}

	names, r := splitNames(t, []string{"-l", "1", "-d", "-a", "1"}, in.String())

	if len(names) != 10 || r.code == 0 || !strings.Contains(r.stderr, "exhausted") {
		t.Fatalf("names %v, code %d, stderr %q", names, r.code, r.stderr)
	}
}
