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
