package main

import (
	"strings"
	"testing"
)

func TestCalMonthEndingOnSaturday(t *testing.T) {
	out := run(t, "", "cal", "9", "2023").stdout

	if strings.HasSuffix(out, "\n\n") {
		t.Fatalf("trailing blank line: %q", out)
	}
}

func TestCalMonthLayout(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"2", "2024"}, "   February 2024\nSu Mo Tu We Th Fr Sa\n             1  2  3\n 4  5  6  7  8  9 10\n11 12 13 14 15 16 17\n18 19 20 21 22 23 24\n25 26 27 28 29\n"},
		{[]string{"-m", "2", "2024"}, "   February 2024\nMo Tu We Th Fr Sa Su\n          1  2  3  4\n 5  6  7  8  9 10 11\n12 13 14 15 16 17 18\n19 20 21 22 23 24 25\n26 27 28 29\n"},
	}

	for _, c := range cases {
		got := run(t, "", append([]string{"cal"}, c.args...)...).stdout
		if got != c.want {
			t.Errorf("cal %v:\ngot  %q\nwant %q", c.args, got, c.want)
		}
	}
}

func TestCalThreeMonthsAndYear(t *testing.T) {
	three := run(t, "", "cal", "-3", "2", "2024").stdout
	first := strings.SplitN(three, "\n", 2)[0]
	for _, name := range []string{"January 2024", "February 2024", "March 2024"} {
		if !strings.Contains(first, name) {
			t.Errorf("-3 title line %q misses %q", first, name)
		}
	}

	for _, args := range [][]string{{"-y", "2024"}, {"2024"}} {
		year := run(t, "", append([]string{"cal"}, args...)...).stdout
		for _, name := range []string{"January", "June", "December", "2024"} {
			if !strings.Contains(year, name) {
				t.Errorf("cal %v misses %q", args, name)
			}
		}
	}
}

func TestCalThreeMonthsWrapsYears(t *testing.T) {
	got := run(t, "", "cal", "-3", "1", "2024").stdout
	first := strings.SplitN(got, "\n", 2)[0]

	if !strings.Contains(first, "December 2023") || !strings.Contains(first, "February 2024") {
		t.Fatalf("title line %q", first)
	}
}
