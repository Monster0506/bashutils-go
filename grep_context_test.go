package main

import "testing"

func TestGrepContext(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"a\nb\nc\nd\n", []string{"-A", "1", "a"}, "a\nb\n"},
		{"a\nb\nc\nd\n", []string{"-B", "1", "c"}, "b\nc\n"},
		{"1\n2\nm\n4\n5\n6\nm\n8\n", []string{"-C", "1", "m"}, "2\nm\n4\n--\n6\nm\n8\n"},
		{"a\nb\n", []string{"-n", "-A", "1", "a"}, "1:a\n2-b\n"},
		{"m\nx\nm\n", []string{"-A", "1", "m"}, "m\nx\nm\n"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"grep"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("grep %v: got %q, want %q", c.args, got, c.want)
		}
	}
}
