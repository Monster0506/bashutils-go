package main

import (
	"path/filepath"
	"testing"
)

func TestGrepMultiFilePrefixesNames(t *testing.T) {
	f1 := writeFile(t, "f1.txt", "a\nb\n")
	f2 := writeFile(t, "f2.txt", "xa\n")

	got := run(t, "", "grep", "-n", "a", f1, f2).stdout

	want := f1 + ":1:a\n" + f2 + ":1:xa\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGrepSingleFileHasNoPrefix(t *testing.T) {
	f := writeFile(t, "f.txt", "a\nb\n")

	got := run(t, "", "grep", "a", f).stdout

	if want := "a\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGrepMissingFileAmongOthers(t *testing.T) {
	good := writeFile(t, "good.txt", "a\n")
	missing := filepath.Join(t.TempDir(), "missing.txt")

	r := run(t, "", "grep", "a", missing, good)

	if r.code != 2 || r.stdout != good+":a\n" {
		t.Fatalf("code %d, stdout %q", r.code, r.stdout)
	}
}

func TestGrepEmptyPatternViaE(t *testing.T) {
	r := run(t, "a\nb\n", "grep", "-e", "")

	if r.code != 0 || r.stdout != "a\nb\n" {
		t.Fatalf("code %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}

func TestGrepPatternModes(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"a.b\naxb\n", []string{"-F", "a.b"}, "a.b\n"},
		{"a\nb\nc\n", []string{"-E", "a|b"}, "a\nb\n"},
		{"foo bar\nfoobar\n", []string{"-w", "foo"}, "foo bar\n"},
		{"foo\nfoo bar\n", []string{"-x", "foo"}, "foo\n"},
		{"a\nb\nc\n", []string{"-e", "a", "-e", "c"}, "a\nc\n"},
		{"A\nb\n", []string{"-i", "-F", "a"}, "A\n"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"grep"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("grep %v: got %q, want %q", c.args, got, c.want)
		}
	}
}

func TestGrepPatternFile(t *testing.T) {
	pats := writeFile(t, "pats.txt", "a\nc\n")

	got := run(t, "a\nb\nc\n", "grep", "-f", pats).stdout

	if want := "a\nc\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGrepOutputModes(t *testing.T) {
	f1 := writeFile(t, "f1.txt", "a\nb\na\n")
	f2 := writeFile(t, "f2.txt", "b\n")

	cases := []struct {
		in   string
		args []string
		want string
		code int
	}{
		{"a\nb\na\n", []string{"-c", "a"}, "2\n", 0},
		{"", []string{"-c", "a", f1, f2}, f1 + ":2\n" + f2 + ":0\n", 0},
		{"", []string{"-l", "a", f1, f2}, f1 + "\n", 0},
		{"", []string{"-L", "a", f1, f2}, f2 + "\n", 0},
		{"a\n", []string{"-q", "a"}, "", 0},
		{"b\n", []string{"-q", "a"}, "", 1},
		{"a\na\na\n", []string{"-m", "2", "a"}, "a\na\n", 0},
		{"", []string{"-H", "b", f2}, f2 + ":b\n", 0},
		{"", []string{"-h", "b", f1, f2}, "b\nb\n", 0},
		{"a\n", []string{"-H", "a"}, "(standard input):a\n", 0},
	}

	for _, c := range cases {
		r := run(t, c.in, append([]string{"grep"}, c.args...)...)
		if r.stdout != c.want || r.code != c.code {
			t.Errorf("grep %v: got %q (code %d), want %q (code %d)", c.args, r.stdout, r.code, c.want, c.code)
		}
	}
}

func TestGrepSilentSuppressesMissingFileErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.txt")

	r := run(t, "", "grep", "-s", "a", missing)

	if r.stderr != "" || r.code != 2 {
		t.Fatalf("stderr %q, code %d", r.stderr, r.code)
	}
}

func TestGrepOnlyMatching(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"abac\n", []string{"-o", "a."}, "ab\nac\n"},
		{"x\nabac\n", []string{"-on", "a."}, "2:ab\n2:ac\n"},
		{"abc\n", []string{"-o", "z"}, ""},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"grep"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("grep %v: got %q, want %q", c.args, got, c.want)
		}
	}
}
