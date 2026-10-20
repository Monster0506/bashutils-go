package main

import "testing"

func TestXargsRejectsUnknownOption(t *testing.T) {
	r := run(t, "x\n", "xargs", "-Z", binPath, "echo", "hi")

	if r.code == 0 || r.stdout != "" || r.stderr == "" {
		t.Fatalf("code %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}

func TestXargsVerboseIgnoresChildArgs(t *testing.T) {
	r := run(t, "x\n", "xargs", binPath, "echo", "--", "-t")

	if r.stdout != "-t x\n" || r.stderr != "" {
		t.Fatalf("stdout %q, stderr %q", r.stdout, r.stderr)
	}
}

func TestXargsVerbosePrintsOnlyCommandLine(t *testing.T) {
	r := run(t, "x\n", "xargs", "-t", binPath, "echo")

	if want := binPath + " echo x\n"; r.stderr != want {
		t.Fatalf("stderr %q, want %q", r.stderr, want)
	}
}

func TestXargsKeepsWhitespaceInDelimitedItems(t *testing.T) {
	cases := []struct {
		in   string
		flag []string
	}{
		{" a \x00b\x00", []string{"-0"}},
		{" a ,b,", []string{"-d", ","}},
	}

	for _, c := range cases {
		args := append(append([]string{"xargs"}, c.flag...), binPath, "echo", "-n")
		got := run(t, c.in, args...).stdout

		if want := " a  b"; got != want {
			t.Errorf("%v: got %q, want %q", c.flag, got, want)
		}
	}
}

func TestXargsReplaceKeepsItemsWhole(t *testing.T) {
	got := run(t, "a b\nc\n", "xargs", "-I", "{}", binPath, "echo", "[{}]").stdout

	if want := "[a b]\n[c]\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXargsContinuesAfterFailure(t *testing.T) {
	good := writeFile(t, "good.txt", "hello\n")
	missing := good + ".missing"

	r := run(t, missing+"\n"+good+"\n", "xargs", "-n", "1", binPath, "cat")

	if r.stdout != "hello\n" || r.code != 123 {
		t.Fatalf("stdout %q, code %d", r.stdout, r.code)
	}
}
