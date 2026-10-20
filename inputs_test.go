package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDashReadsStdinEverywhere(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"sort", "-"}, "a\nb\n"},
		{[]string{"uniq", "-"}, "b\na\n"},
		{[]string{"cat", "-"}, "b\na\n"},
		{[]string{"grep", ".", "-"}, "b\na\n"},
		{[]string{"paste", "-"}, "b\na\n"},
		{[]string{"head", "-"}, "b\na\n"},
		{[]string{"tail", "-"}, "b\na\n"},
		{[]string{"cut", "-c", "1", "-"}, "b\na\n"},
	}

	for _, c := range cases {
		got := run(t, "b\na\n", c.args...).stdout
		if got != c.want {
			t.Errorf("%v: got %q, want %q", c.args, got, c.want)
		}
	}
}

func TestMissingFileDoesNotStopOtherInputs(t *testing.T) {
	good := writeFile(t, "good.txt", "hello\n")
	missing := filepath.Join(t.TempDir(), "missing.txt")

	cases := [][]string{
		{"cat"},
		{"sort"},
		{"uniq"},
		{"head"},
		{"tail"},
		{"cut", "-c", "1-"},
		{"grep", "hello"},
	}

	for _, args := range cases {
		r := run(t, "", append(args, missing, good)...)
		if !strings.Contains(r.stdout, "hello") || r.code == 0 || r.stderr == "" {
			t.Errorf("%v: stdout %q, code %d, stderr %q", args, r.stdout, r.code, r.stderr)
		}
	}
}
