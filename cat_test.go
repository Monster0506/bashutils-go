package main

import "testing"

func TestCatOptions(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"a\nb\n", []string{"-n"}, "     1\ta\n     2\tb\n"},
		{"a\n\nb\n", []string{"-n"}, "     1\ta\n     2\t\n     3\tb\n"},
		{"a\n\nb\n", []string{"-b"}, "     1\ta\n\n     2\tb\n"},
		{"a\n\n\n\nb\n", []string{"-s"}, "a\n\nb\n"},
		{"\n\na\n", []string{"-s"}, "\na\n"},
		{"a\n", []string{"-E"}, "a$\n"},
		{"a\tb\n", []string{"-T"}, "a^Ib\n"},
		{"a", []string{"-n"}, "     1\ta"},
		{"a\n\n\nb\n", []string{"-sn"}, "     1\ta\n     2\t\n     3\tb\n"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"cat"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("cat %v: got %q, want %q", c.args, got, c.want)
		}
	}
}

func TestCatNumbersContinueAcrossFiles(t *testing.T) {
	f1 := writeFile(t, "a.txt", "a\n")
	f2 := writeFile(t, "b.txt", "b\n")

	got := run(t, "", "cat", "-n", f1, f2).stdout

	if want := "     1\ta\n     2\tb\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
