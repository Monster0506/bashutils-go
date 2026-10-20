package main

import "testing"

func TestPasteSharesStdinAcrossDashes(t *testing.T) {
	got := run(t, "1\n2\n3\n4\n", "paste", "-", "-").stdout

	if want := "1\t2\n3\t4\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPasteDelimiterEscapes(t *testing.T) {
	cases := []struct {
		delim string
		want  string
	}{
		{`\n`, "1\n2\n3\n4\n"},
		{`\t`, "1\t2\n3\t4\n"},
		{`\\`, "1\\2\n3\\4\n"},
		{`\0`, "12\n34\n"},
	}

	for _, c := range cases {
		got := run(t, "1\n2\n3\n4\n", "paste", "-d", c.delim, "-", "-").stdout
		if got != c.want {
			t.Errorf("-d %q: got %q, want %q", c.delim, got, c.want)
		}
	}
}

func TestPasteSerial(t *testing.T) {
	f1 := writeFile(t, "a.txt", "1\n2\n3\n")
	f2 := writeFile(t, "b.txt", "4\n5\n")

	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"", []string{"-s", f1, f2}, "1\t2\t3\n4\t5\n"},
		{"", []string{"-sd,", f1, f2}, "1,2,3\n4,5\n"},
		{"a\nb\n", []string{"-s"}, "a\tb\n"},
		{"1\n2\n3\n4\n", []string{"-s", "-d", ",;"}, "1,2;3,4\n"},
		{"", []string{"-s", f2, f1}, "4\t5\n1\t2\t3\n"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"paste"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("paste %v: got %q, want %q", c.args, got, c.want)
		}
	}
}
