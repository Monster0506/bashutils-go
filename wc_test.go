package main

import "testing"

func TestWcCharsCountsAllRunes(t *testing.T) {
	got := run(t, "héllo wörld\n", "wc", "-m").stdout

	if want := "12\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWcDefaultColumns(t *testing.T) {
	got := run(t, "a b\n", "wc").stdout

	if want := "1 2 4\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWcDashReadsStdinAndLabelsIt(t *testing.T) {
	got := run(t, "a b\n", "wc", "-").stdout

	if want := "1 2 4 -\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWcTotalsMultipleFiles(t *testing.T) {
	f1 := writeFile(t, "a.txt", "a\n")
	f2 := writeFile(t, "b.txt", "b c\nd\n")

	got := run(t, "", "wc", "-lw", f1, f2).stdout

	want := "1 1 " + f1 + "\n2 3 " + f2 + "\n3 4 total\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWcMaxLineLength(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"ab\nabcd\n", []string{"-L"}, "4\n"},
		{"a\tb\n", []string{"-L"}, "9\n"},
		{"héé\n", []string{"-L"}, "3\n"},
		{"a b\n", []string{"-lL"}, "1 3\n"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"wc"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("wc %v: got %q, want %q", c.args, got, c.want)
		}
	}
}

func TestWcMaxLineLengthTotalIsMaximum(t *testing.T) {
	f1 := writeFile(t, "a.txt", "abcdef\n")
	f2 := writeFile(t, "b.txt", "xy\n")

	got := run(t, "", "wc", "-L", f1, f2).stdout

	want := "6 " + f1 + "\n2 " + f2 + "\n6 total\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
