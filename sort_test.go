package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestSortIsStable(t *testing.T) {
	var in, want strings.Builder
	for i := range 100 {
		fmt.Fprintf(&in, "k%d line%03d\n", i%3, i)
	}
	for k := range 3 {
		for i := k; i < 100; i += 3 {
			fmt.Fprintf(&want, "k%d line%03d\n", k, i)
		}
	}

	got := run(t, in.String(), "sort", "-k", "1").stdout

	if got != want.String() {
		t.Fatalf("equal keys were reordered:\n%s", got)
	}
}

func TestSortNumericUsesLeadingNumber(t *testing.T) {
	got := run(t, "10 apples\n9 pears\n", "sort", "-n").stdout

	if want := "9 pears\n10 apples\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSortNumericTreatsNanAsText(t *testing.T) {
	got := run(t, "3\nnan\ninf\n1\n2\n", "sort", "-n").stdout

	if want := "1\n2\n3\ninf\nnan\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSortKeysAndModifiers(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"a 3\nb 1\nc 2\n", []string{"-k", "2"}, "b 1\nc 2\na 3\n"},
		{"a 10\nb 9\n", []string{"-k", "2,2n"}, "b 9\na 10\n"},
		{"a 10\nb 9\n", []string{"-k2n"}, "b 9\na 10\n"},
		{"a 10\na 9\nb 1\n", []string{"-k1,1", "-k2,2n"}, "a 9\na 10\nb 1\n"},
		{"a 1 z\na 1 y\n", []string{"-k1,2"}, "a 1 y\na 1 z\n"},
		{"x,2\ny,1\n", []string{"-t", ",", "-k2"}, "y,1\nx,2\n"},
		{"a 1\na 2\nb 3\n", []string{"-k1,1", "-u"}, "a 1\nb 3\n"},
		{"b\nA\na\nB\n", []string{"-f"}, "A\na\nB\nb\n"},
		{"  b\na\n", []string{"-b"}, "a\n  b\n"},
		{"  b\na\n", nil, "  b\na\n"},
		{"1\n3\n2\n", []string{"-nr"}, "3\n2\n1\n"},
		{"a 1\na 2\n", []string{"-k1,1", "-r"}, "a 2\na 1\n"},
		{"a 1\na 2\n", []string{"-k1,1", "-r", "-s"}, "a 1\na 2\n"},
		{"b 2\na 1\n", []string{"-k2,2r"}, "b 2\na 1\n"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"sort"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("sort %v: got %q, want %q", c.args, got, c.want)
		}
	}
}

func TestSortCheck(t *testing.T) {
	if r := run(t, "a\nb\n", "sort", "-c"); r.code != 0 || r.stdout != "" || r.stderr != "" {
		t.Errorf("sorted: code %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}

	r := run(t, "b\na\n", "sort", "-c")
	if r.code != 1 || r.stdout != "" || !strings.Contains(r.stderr, "disorder") {
		t.Errorf("unsorted: code %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}

func TestSortOutputFile(t *testing.T) {
	in := writeFile(t, "in.txt", "b\na\n")

	r := run(t, "", "sort", "-o", in, in)

	data, err := os.ReadFile(in)
	if err != nil {
		t.Fatal(err)
	}
	if r.stdout != "" || string(data) != "a\nb\n" {
		t.Fatalf("stdout %q, file %q", r.stdout, data)
	}
}

func TestSortRejectsBadKey(t *testing.T) {
	for _, key := range []string{"x", "0", "1,0", "1.2"} {
		if r := run(t, "a\n", "sort", "-k", key); r.code == 0 {
			t.Errorf("-k %s: expected failure", key)
		}
	}
}
