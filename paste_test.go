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
