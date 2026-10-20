package main

import "testing"

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
