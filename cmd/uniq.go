package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type uniqGroup struct {
	lines []string
}

type uniqOpts struct {
	count      bool
	repeated   bool
	unique     bool
	all        bool
	ignoreCase bool
	skipFields int
	skipChars  int
	checkChars int
}

var uniqCmd = &cobra.Command{
	Use:   "uniq [files...]",
	Short: "Filter out repeated lines",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		o := newUniqOpts(cmd)

		for _, g := range o.groups(readLines("uniq", args)) {
			o.print(g)
		}
	},
}

func init() {
	uniqCmd.Flags().BoolP("count", "c", false, "prefix lines by the number of occurrences")
	uniqCmd.Flags().BoolP("repeated", "d", false, "only print duplicate lines, one for each group")
	uniqCmd.Flags().BoolP("all-repeated", "D", false, "print all duplicate lines")
	uniqCmd.Flags().BoolP("unique", "u", false, "only print unique lines")
	uniqCmd.Flags().BoolP("ignore-case", "i", false, "ignore differences in case when comparing")
	uniqCmd.Flags().IntP("skip-fields", "f", 0, "avoid comparing the first N fields")
	uniqCmd.Flags().IntP("skip-chars", "s", 0, "avoid comparing the first N characters")
	uniqCmd.Flags().IntP("check-chars", "w", 0, "compare no more than N characters in lines")
}

func newUniqOpts(cmd *cobra.Command) *uniqOpts {
	o := &uniqOpts{}
	o.count, _ = cmd.Flags().GetBool("count")
	o.repeated, _ = cmd.Flags().GetBool("repeated")
	o.all, _ = cmd.Flags().GetBool("all-repeated")
	o.unique, _ = cmd.Flags().GetBool("unique")
	o.ignoreCase, _ = cmd.Flags().GetBool("ignore-case")
	o.skipFields, _ = cmd.Flags().GetInt("skip-fields")
	o.skipChars, _ = cmd.Flags().GetInt("skip-chars")
	o.checkChars, _ = cmd.Flags().GetInt("check-chars")
	return o
}

func (o *uniqOpts) key(line string) string {
	rest := line
	for i := 0; i < o.skipFields; i++ {
		rest = strings.TrimLeft(rest, " \t")
		if end := strings.IndexAny(rest, " \t"); end >= 0 {
			rest = rest[end:]
		} else {
			rest = ""
		}
	}

	runes := []rune(rest)
	if o.skipChars < len(runes) {
		runes = runes[o.skipChars:]
	} else {
		runes = nil
	}
	if o.checkChars > 0 && o.checkChars < len(runes) {
		runes = runes[:o.checkChars]
	}

	key := string(runes)
	if o.ignoreCase {
		key = strings.ToLower(key)
	}
	return key
}

func (o *uniqOpts) groups(lines []string) []uniqGroup {
	var groups []uniqGroup
	prevKey := ""
	for i, line := range lines {
		key := o.key(line)
		if i > 0 && key == prevKey {
			last := &groups[len(groups)-1]
			last.lines = append(last.lines, line)
			continue
		}
		groups = append(groups, uniqGroup{lines: []string{line}})
		prevKey = key
	}
	return groups
}

func (o *uniqOpts) selected(g uniqGroup) bool {
	n := len(g.lines)
	switch {
	case o.all:
		return n > 1
	case o.repeated && o.unique:
		return false
	case o.repeated:
		return n > 1
	case o.unique:
		return n == 1
	}
	return true
}

func (o *uniqOpts) print(g uniqGroup) {
	if !o.selected(g) {
		return
	}

	if o.all {
		for _, line := range g.lines {
			fmt.Fprintln(stdout, line)
		}
		return
	}

	if o.count {
		fmt.Fprintf(stdout, "%7d %s\n", len(g.lines), g.lines[0])
		return
	}
	fmt.Fprintln(stdout, g.lines[0])
}
