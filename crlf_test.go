package main

import "testing"

func TestCRLFIsPreserved(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"a\r\nb\r\n", []string{"head"}, "a\r\nb\r\n"},
		{"a\r\nb\r\n", []string{"tail"}, "a\r\nb\r\n"},
		{"b\r\na\r\n", []string{"sort"}, "a\r\nb\r\n"},
		{"a\r\na\r\nb\r\n", []string{"uniq"}, "a\r\nb\r\n"},
		{"a\r\nb\r\n", []string{"grep", "a"}, "a\r\n"},
		{"a\r\nb\r\n", []string{"paste"}, "a\r\nb\r\n"},
		{"a\r\n", []string{"cut", "-c", "1-"}, "a\r\n"},
	}

	for _, c := range cases {
		got := run(t, c.in, c.args...).stdout
		if got != c.want {
			t.Errorf("%v: got %q, want %q", c.args, got, c.want)
		}
	}
}
