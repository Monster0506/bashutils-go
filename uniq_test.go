package main

import "testing"

func TestUniqKeepsOrderAndAdjacency(t *testing.T) {
	got := run(t, "b\nb\na\nb\n", "uniq").stdout

	if want := "b\na\nb\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestUniqOptions(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"A\na\nb\n", []string{"-i"}, "A\nb\n"},
		{"x a\ny a\nz b\n", []string{"-f", "1"}, "x a\nz b\n"},
		{"xa\nya\nzb\n", []string{"-s", "1"}, "xa\nzb\n"},
		{"ab\nac\nbd\n", []string{"-w", "1"}, "ab\nbd\n"},
		{"a\na\nb\n", []string{"-c"}, "      2 a\n      1 b\n"},
		{"a\na\nb\n", []string{"-D"}, "a\na\n"},
		{"a\na\nb\n", []string{"-d"}, "a\n"},
		{"a\na\nb\n", []string{"-u"}, "b\n"},
		{"a\na\nb\n", []string{"-dc"}, "      2 a\n"},
		{"", nil, ""},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"uniq"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("uniq %v: got %q, want %q", c.args, got, c.want)
		}
	}
}
