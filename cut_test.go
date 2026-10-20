package main

import (
	"strings"
	"testing"
)

func TestCutKeepsLinesWithoutDelimiter(t *testing.T) {
	got := run(t, "abc\na,b\n", "cut", "-d", ",", "-f", "2").stdout

	if want := "abc\nb\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCutOpenRanges(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"a,b,c,d\n", []string{"-d", ",", "-f", "2-"}, "b,c,d\n"},
		{"a,b,c,d\n", []string{"-d", ",", "-f", "-2"}, "a,b\n"},
		{"abcdef\n", []string{"-c", "3-"}, "cdef\n"},
		{"abcdef\n", []string{"-c", "-2"}, "ab\n"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"cut"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("cut %v: got %q, want %q", c.args, got, c.want)
		}
	}
}

func TestCutOutputsInInputOrderWithoutDuplicates(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"a,b,c\n", []string{"-d", ",", "-f", "3,1"}, "a,c\n"},
		{"a,b,c\n", []string{"-d", ",", "-f", "1,1"}, "a\n"},
		{"abcd\n", []string{"-c", "2-3,1-2"}, "abc\n"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"cut"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("cut %v: got %q, want %q", c.args, got, c.want)
		}
	}
}

func TestCutRejectsBadListOnce(t *testing.T) {
	r := run(t, "a\nb\nc\n", "cut", "-f", "x")

	if r.code == 0 {
		t.Error("expected non-zero exit code")
	}
	if r.stdout != "" {
		t.Errorf("unexpected stdout %q", r.stdout)
	}
	if n := strings.Count(r.stderr, "\n"); n != 1 {
		t.Errorf("want 1 error line, got %d: %q", n, r.stderr)
	}
}

func TestCutDashReadsStdin(t *testing.T) {
	got := run(t, "ab\n", "cut", "-c", "1", "-").stdout

	if want := "a\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
