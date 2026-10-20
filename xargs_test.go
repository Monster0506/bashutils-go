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
