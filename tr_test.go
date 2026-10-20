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

func TestTrClassesEscapesAndSqueeze(t *testing.T) {
	cases := []struct {
		in   string
		args []string
		want string
	}{
		{"abc", []string{"[:lower:]", "[:upper:]"}, "ABC"},
		{"a1b2", []string{"-d", "[:digit:]"}, "ab"},
		{"a-b_1!", []string{"-cd", "[:alnum:]"}, "ab1"},
		{"a b\tc", []string{"[:space:]", "_"}, "a_b_c"},
		{"a,b", []string{"[:punct:]", " "}, "a b"},
		{"a\nb\n", []string{`\n`, " "}, "a b "},
		{"a\r\n", []string{"-d", `\r`}, "a\n"},
		{"a\tb", []string{`\t`, ","}, "a,b"},
		{"abc", []string{`\141`, "X"}, "Xbc"},
		{"a\\b", []string{`\\`, "/"}, "a/b"},
		{"a   b  c", []string{"-s", " "}, "a b c"},
		{"aabbc", []string{"-s", "a-b", "A-B"}, "ABc"},
		{"a\n\n\nb", []string{"-s", `\n`}, "a\nb"},
		{"aabbb", []string{"-ds", "a", "b"}, "b"},
		{"abc", []string{"-t", "abc", "x"}, "xbc"},
		{"abc", []string{"abc", "x"}, "xxx"},
		{"a-c", []string{"a\\-c", "xyz"}, "xyz"},
	}

	for _, c := range cases {
		got := run(t, c.in, append([]string{"tr"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("tr %q: got %q, want %q", c.args, got, c.want)
		}
	}
}

func TestTrRejectsReverseRange(t *testing.T) {
	r := run(t, "a", "tr", "z-a", "x")

	if r.code == 0 || r.stdout != "" || r.stderr == "" {
		t.Fatalf("code %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
}
