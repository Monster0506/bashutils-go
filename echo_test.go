package main

import "testing"

func TestEchoTreatsUnknownDashArgsAsText(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"-x"}, "-x\n"},
		{[]string{"-1", "a"}, "-1 a\n"},
		{[]string{"a", "-n"}, "a -n\n"},
		{[]string{"-nx"}, "-nx\n"},
		{[]string{"--", "a"}, "-- a\n"},
	}

	for _, c := range cases {
		got := run(t, "", append([]string{"echo"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("echo %v: got %q, want %q", c.args, got, c.want)
		}
	}
}

func TestEchoFlags(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"-n", "a"}, "a"},
		{[]string{"-ne", `a\tb`}, "a\tb"},
		{[]string{"-en", `a\tb`}, "a\tb"},
		{[]string{"-n", "-e", `a\n`}, "a\n"},
		{[]string{"-e", "-E", `a\n`}, "a\\n\n"},
		{[]string{`a\n`}, "a\\n\n"},
	}

	for _, c := range cases {
		got := run(t, "", append([]string{"echo"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("echo %v: got %q, want %q", c.args, got, c.want)
		}
	}
}

func TestEchoExpandEnvIsLongOnly(t *testing.T) {
	got := run(t, "", "echo", "--expand-env", "$HOME_UNSET_X").stdout

	if want := "$HOME_UNSET_X\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
