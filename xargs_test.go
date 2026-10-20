package main

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestXargsRejectsUnknownOption(t *testing.T) {
	r := run(t, "x\n", "xargs", "-Z", binPath, "echo", "hi")

	if r.code == 0 || r.stdout != "" || r.stderr == "" {
		t.Fatalf("code %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}

func TestXargsVerboseIgnoresChildArgs(t *testing.T) {
	r := run(t, "x\n", "xargs", binPath, "echo", "-t")

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

func TestXargsRunsBuiltinCommands(t *testing.T) {
	got := run(t, "a\n", "xargs", "echo", "hi").stdout

	if want := "hi a\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXargsRunsBuiltinWhenInvokedByLinkName(t *testing.T) {
	got := runAs(t, "xargs.exe", "a\n", "echo").stdout

	if want := "a\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXargsRejectsBadOptionValues(t *testing.T) {
	cases := [][]string{
		{"xargs", "-n", "x", "echo"},
		{"xargs", "-n", "0", "echo"},
		{"xargs", "-n"},
		{"xargs", "-I"},
		{"xargs", "-d"},
	}

	for _, args := range cases {
		r := run(t, "a b\n", args...)
		if r.code == 0 || r.stdout != "" || r.stderr == "" {
			t.Errorf("%v: code %d, stdout %q, stderr %q", args, r.code, r.stdout, r.stderr)
		}
	}
}

func TestXargsAcceptsAttachedAndLongOptions(t *testing.T) {
	cases := [][]string{
		{"xargs", "-n1", "echo"},
		{"xargs", "--max-args=1", "echo"},
		{"xargs", "--max-args", "1", "echo"},
		{"xargs", "-rn1", "echo"},
	}

	for _, args := range cases {
		got := run(t, "a b\n", args...).stdout
		if want := "a\nb\n"; got != want {
			t.Errorf("%v: got %q, want %q", args, got, want)
		}
	}
}

func TestXargsRunsDefaultEchoOnEmptyInput(t *testing.T) {
	r := run(t, "", "xargs")

	if r.stdout != "\n" || r.code != 0 {
		t.Fatalf("stdout %q, code %d", r.stdout, r.code)
	}
}

func TestXargsNoRunIfEmptySkipsCommand(t *testing.T) {
	r := run(t, "", "xargs", "-r")

	if r.stdout != "" || r.code != 0 {
		t.Fatalf("stdout %q, code %d", r.stdout, r.code)
	}
}

func TestXargsReplaceStripsCarriageReturns(t *testing.T) {
	got := run(t, "a\r\nb\r\n", "xargs", "-I", "{}", "echo", "[{}]").stdout

	if want := "[a]\n[b]\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXargsHonorsQuotesAndBackslashes(t *testing.T) {
	got := run(t, "'a b' \"c d\" e\\ f g\n", "xargs", "-n", "1", "echo").stdout

	if want := "a b\nc d\ne f\ng\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXargsRejectsUnmatchedQuote(t *testing.T) {
	r := run(t, "'a b\n", "xargs", "echo")

	if r.code == 0 || r.stdout != "" || r.stderr == "" {
		t.Fatalf("code %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}

func TestXargsKeepsBackslashesInPaths(t *testing.T) {
	got := run(t, "C:\\Users\\me \\\\srv\\share\n", "xargs", "-n", "1", "echo").stdout

	if want := "C:\\Users\\me\n\\\\srv\\share\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXargsDelimiterMustBeSingleCharacter(t *testing.T) {
	r := run(t, "a,,b", "xargs", "-d", ",,", "echo")

	if r.code == 0 || r.stdout != "" || r.stderr == "" {
		t.Fatalf("code %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}

func TestXargsDelimiterEscapes(t *testing.T) {
	got := run(t, "a\tb\tc", "xargs", "-d", `\t`, "-n", "1", "echo").stdout

	if want := "a\nb\nc\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXargsArgFile(t *testing.T) {
	f := writeFile(t, "items.txt", "a b\n")

	got := run(t, "ignored\n", "xargs", "-a", f, "echo").stdout

	if want := "a b\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXargsArgFileMissing(t *testing.T) {
	r := run(t, "", "xargs", "-a", writeFile(t, "x", "")+".missing", "echo")

	if r.code == 0 || r.stderr == "" {
		t.Fatalf("code %d, stderr %q", r.code, r.stderr)
	}
}

func TestXargsMaxLines(t *testing.T) {
	got := run(t, "a b\nc d\n\ne\n", "xargs", "-L", "2", "echo").stdout

	if want := "a b c d\ne\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXargsEOFString(t *testing.T) {
	got := run(t, "a b STOP c\n", "xargs", "-E", "STOP", "echo").stdout

	if want := "a b\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestXargsRejectsBadMaxLines(t *testing.T) {
	if r := run(t, "a\n", "xargs", "-L", "0", "echo"); r.code == 0 {
		t.Fatal("expected failure")
	}
}

func TestXargsParallelRunsEveryCommand(t *testing.T) {
	r := run(t, "a b c d e f\n", "xargs", "-P", "3", "-n", "1", "echo")

	got := strings.Split(strings.TrimSpace(r.stdout), "\n")
	sort.Strings(got)
	if want := []string{"a", "b", "c", "d", "e", "f"}; !reflect.DeepEqual(got, want) || r.code != 0 {
		t.Fatalf("got %q (code %d)", got, r.code)
	}
}

func TestXargsParallelReportsFailure(t *testing.T) {
	good := writeFile(t, "good.txt", "hello\n")
	r := run(t, good+".missing\n"+good+"\n", "xargs", "-P", "2", "-n", "1", binPath, "cat")

	if r.stdout != "hello\n" || r.code != 123 {
		t.Fatalf("stdout %q, code %d", r.stdout, r.code)
	}
}

func TestXargsRejectsBadParallelism(t *testing.T) {
	for _, v := range []string{"x", "-1"} {
		if r := run(t, "a\n", "xargs", "-P", v, "echo"); r.code == 0 {
			t.Errorf("-P %s: expected failure", v)
		}
	}
}

func TestXargsParallelZeroMeansUnlimited(t *testing.T) {
	got := run(t, "a b\n", "xargs", "-P", "0", "-n", "1", "echo")

	if got.code != 0 || len(strings.Fields(got.stdout)) != 2 {
		t.Fatalf("code %d, stdout %q", got.code, got.stdout)
	}
}
