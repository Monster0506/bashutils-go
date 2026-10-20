package main

import "testing"

func TestTrDeleteComplement(t *testing.T) {
	got := run(t, "abcxyz\n", "tr", "-d", "-c", "a-c").stdout

	if want := "abc"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTrTranslateComplement(t *testing.T) {
	got := run(t, "abcxyz", "tr", "-c", "a-c", "_").stdout

	if want := "abc___"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTrMultibyte(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"hello", []string{"e", "é"}, "héllo"},
		{"héllo", []string{"é", "e"}, "hello"},
		{"aéb", []string{"éab", "xyz"}, "yxz"},
		{"héllo", []string{"-d", "é"}, "hllo"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"tr"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("tr %v: got %q, want %q", c.args, got, c.want)
		}
	}
}
