package main

import (
	"strings"
	"testing"
)

func TestFatalErrorsPrintOnceAndExitNonZero(t *testing.T) {
	cases := [][]string{
		{"cal", "x"},
		{"cal", "13", "0"},
		{"cut", "-f", "x"},
		{"cut"},
		{"cut", "-f", "1", "-c", "1"},
	}

	for _, args := range cases {
		r := run(t, "a\n", args...)
		if r.code == 0 || r.stdout != "" {
			t.Errorf("%v: code %d, stdout %q", args, r.code, r.stdout)
		}
		if n := strings.Count(r.stderr, "\n"); n != 1 {
			t.Errorf("%v: want 1 stderr line, got %d: %q", args, n, r.stderr)
		}
	}
}
