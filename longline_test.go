package main

import (
	"strings"
	"testing"
)

func TestLongLinesSurvive(t *testing.T) {
	line := strings.Repeat("x", 200_000) + "\n"

	cases := [][]string{
		{"head"},
		{"tail"},
		{"grep", "x"},
		{"sort"},
		{"uniq"},
		{"paste"},
		{"cut", "-c", "1-"},
	}

	for _, args := range cases {
		r := run(t, line, args...)
		if r.stdout != line {
			t.Errorf("%v: got %d bytes, want %d (stderr %q)", args, len(r.stdout), len(line), r.stderr)
		}
	}
}
