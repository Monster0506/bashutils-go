package main

import "testing"

func TestEchoEscapes(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`a\\b`, "a\\b\n"},
		{`a\tb`, "a\tb\n"},
		{`a\rb`, "a\rb\n"},
		{`\a\b\f\v`, "\a\b\f\v\n"},
		{`\e`, "\x1b\n"},
		{`\0101`, "A\n"},
		{`\x41`, "A\n"},
		{`\x4`, "\x04\n"},
		{`a\cb`, "a"},
		{`a\qb`, "a\\qb\n"},
		{`a\`, "a\\\n"},
	}

	for _, c := range cases {
		got := run(t, "", "echo", "-e", c.in).stdout
		if got != c.want {
			t.Errorf("echo -e %q: got %q, want %q", c.in, got, c.want)
		}
	}
}
