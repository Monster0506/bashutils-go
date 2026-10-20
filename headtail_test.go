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
