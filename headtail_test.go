package main

import "testing"

func TestHeadTailSeparateRepeatedFile(t *testing.T) {
	f := writeFile(t, "f.txt", "x\n")

	for _, name := range []string{"head", "tail"} {
		got := run(t, "", name, f, f).stdout
		header := "==> " + f + " <==\n"

		if want := header + "x\n\n" + header + "x\n"; got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

func TestHeadTailDashReadsStdin(t *testing.T) {
	for _, name := range []string{"head", "tail"} {
		got := run(t, "a\nb\n", name, "-").stdout
		if want := "a\nb\n"; got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

func TestHeadTailReportReadErrors(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"head", "tail"} {
		r := run(t, "", name, dir)
		if r.code == 0 || r.stderr == "" {
			t.Errorf("%s: code %d, stderr %q", name, r.code, r.stderr)
		}
	}
}

const tenLines = "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n"

func TestHeadOptions(t *testing.T) {
	f := writeFile(t, "f.txt", "a\n")

	cases := []struct {
		in   string
		args []string
		want string
	}{
		{tenLines, []string{"-5"}, "1\n2\n3\n4\n5\n"},
		{tenLines, []string{"-n5"}, "1\n2\n3\n4\n5\n"},
		{tenLines, []string{"-n", "-8"}, "1\n2\n"},
		{tenLines, []string{"-c", "5"}, "1\n2\n3"},
		{tenLines, []string{"-c", "-17"}, "1\n2\n"},
		{"abc", []string{"-c", "0"}, ""},
		{tenLines, []string{"-n", "0"}, ""},
		{"", []string{"-v", f}, "==> " + f + " <==\na\n"},
		{"", []string{"-q", f, f}, "a\na\n"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"head"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("head %v: got %q, want %q", c.args, got, c.want)
		}
	}
}

func TestTailOptions(t *testing.T) {
	f := writeFile(t, "f.txt", "a\n")

	cases := []struct {
		in   string
		args []string
		want string
	}{
		{tenLines, []string{"-3"}, "8\n9\n10\n"},
		{tenLines, []string{"-n3"}, "8\n9\n10\n"},
		{tenLines, []string{"-n", "+9"}, "9\n10\n"},
		{tenLines, []string{"-n", "+1"}, tenLines},
		{tenLines, []string{"-c", "3"}, "10\n"},
		{tenLines, []string{"-c", "+17"}, "9\n10\n"},
		{tenLines, []string{"-n", "0"}, ""},
		{"", []string{"-v", f}, "==> " + f + " <==\na\n"},
		{"", []string{"-q", f, f}, "a\na\n"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"tail"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("tail %v: got %q, want %q", c.args, got, c.want)
		}
	}
}
