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
