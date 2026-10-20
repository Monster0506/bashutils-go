package cmd

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type keyFlags struct {
	numeric bool
	reverse bool
	fold    bool
	blanks  bool
}

type sortKey struct {
	start, end int
	flags      keyFlags
}

type sortRecord struct {
	line string
	keys []string
}

type stability int

const (
	lastResort stability = iota
	stableOnly
)

type sorter struct {
	keys      []sortKey
	separator string
	stability stability
	reverse   bool
}

var sortKeyDefs []string

var sortCmd = &cobra.Command{
	Use:   "sort [files...]",
	Short: "Sort lines of text files",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSorter(cmd)
		if err != nil {
			return err
		}

		unique, _ := cmd.Flags().GetBool("unique")
		check, _ := cmd.Flags().GetBool("check")
		output, _ := cmd.Flags().GetString("output")

		records := s.records(readLines("sort", args))
		if check {
			return s.check(records, args)
		}

		sort.SliceStable(records, func(i, j int) bool {
			return s.compare(records[i], records[j]) < 0
		})
		if unique {
			records = s.dedupe(records)
		}

		return writeSorted(records, output)
	},
}

func init() {
	sortCmd.Flags().BoolP("reverse", "r", false, "reverse the result of comparisons")
	sortCmd.Flags().BoolP("numeric-sort", "n", false, "compare according to string numerical value")
	sortCmd.Flags().BoolP("unique", "u", false, "output only the first of an equal run")
	sortCmd.Flags().BoolP("ignore-case", "f", false, "fold lower case to upper case characters")
	sortCmd.Flags().BoolP("ignore-leading-blanks", "b", false, "ignore leading blanks")
	sortCmd.Flags().BoolP("stable", "s", false, "disable last-resort comparison")
	sortCmd.Flags().BoolP("check", "c", false, "check whether input is sorted; do not sort")
	sortCmd.Flags().StringP("output", "o", "", "write result to FILE instead of standard output")
	sortCmd.Flags().StringArrayVarP(&sortKeyDefs, "key", "k", nil, "sort via a key; KEYDEF is F[,F][nrfb]")
	sortCmd.Flags().StringP("field-separator", "t", "", "use SEP instead of non-blank to blank transition")
}

func newSorter(cmd *cobra.Command) (*sorter, error) {
	global := keyFlags{}
	global.reverse, _ = cmd.Flags().GetBool("reverse")
	global.numeric, _ = cmd.Flags().GetBool("numeric-sort")
	global.fold, _ = cmd.Flags().GetBool("ignore-case")
	global.blanks, _ = cmd.Flags().GetBool("ignore-leading-blanks")

	s := &sorter{reverse: global.reverse}
	s.separator, _ = cmd.Flags().GetString("field-separator")
	if stable, _ := cmd.Flags().GetBool("stable"); stable {
		s.stability = stableOnly
	}
	if unique, _ := cmd.Flags().GetBool("unique"); unique {
		s.stability = stableOnly
	}

	for _, def := range sortKeyDefs {
		key, err := parseKeyDef(def, global)
		if err != nil {
			return nil, fail(2, "sort: %v", err)
		}
		s.keys = append(s.keys, key)
	}
	if len(s.keys) == 0 {
		s.keys = []sortKey{{flags: global}}
	}
	return s, nil
}

func parseKeyDef(def string, global keyFlags) (sortKey, error) {
	startPart, endPart, hasEnd := strings.Cut(def, ",")

	start, startFlags, err := parseKeyPart(startPart)
	if err != nil {
		return sortKey{}, fmt.Errorf("invalid key %q: %v", def, err)
	}

	key := sortKey{start: start, flags: startFlags}
	if hasEnd {
		end, endFlags, err := parseKeyPart(endPart)
		if err != nil {
			return sortKey{}, fmt.Errorf("invalid key %q: %v", def, err)
		}
		key.end = end
		key.flags = mergeFlags(startFlags, endFlags)
	}

	if key.flags == (keyFlags{}) {
		key.flags = global
	}
	return key, nil
}

func parseKeyPart(part string) (int, keyFlags, error) {
	digits := strings.TrimRight(part, "nrfb")
	mods := part[len(digits):]

	n, err := strconv.Atoi(digits)
	if err != nil || n < 1 {
		return 0, keyFlags{}, fmt.Errorf("field number must be a positive integer")
	}

	var f keyFlags
	for _, m := range mods {
		switch m {
		case 'n':
			f.numeric = true
		case 'r':
			f.reverse = true
		case 'f':
			f.fold = true
		case 'b':
			f.blanks = true
		}
	}
	return n, f, nil
}

func mergeFlags(a, b keyFlags) keyFlags {
	return keyFlags{
		numeric: a.numeric || b.numeric,
		reverse: a.reverse || b.reverse,
		fold:    a.fold || b.fold,
		blanks:  a.blanks || b.blanks,
	}
}

func (s *sorter) records(lines []string) []sortRecord {
	records := make([]sortRecord, len(lines))
	for i, line := range lines {
		keys := make([]string, len(s.keys))
		for j, k := range s.keys {
			keys[j] = s.extract(line, k)
		}
		records[i] = sortRecord{line: line, keys: keys}
	}
	return records
}

func (s *sorter) extract(line string, k sortKey) string {
	text := line
	if k.start > 0 {
		var fields []string
		sep := " "
		if s.separator != "" {
			fields = strings.Split(line, s.separator)
			sep = s.separator
		} else {
			fields = strings.Fields(line)
		}

		if k.start > len(fields) {
			return ""
		}
		end := len(fields)
		if k.end > 0 && k.end < end {
			end = k.end
		}
		text = strings.Join(fields[k.start-1:end], sep)
	}

	if k.flags.blanks {
		text = strings.TrimLeft(text, " \t")
	}
	return text
}

func (s *sorter) compare(a, b sortRecord) int {
	for i, k := range s.keys {
		c := compareKey(a.keys[i], b.keys[i], k.flags)
		if k.flags.reverse {
			c = -c
		}
		if c != 0 {
			return c
		}
	}

	if s.stability == stableOnly {
		return 0
	}

	c := strings.Compare(a.line, b.line)
	if s.reverse {
		c = -c
	}
	return c
}

func compareKey(a, b string, f keyFlags) int {
	if f.fold {
		a, b = strings.ToLower(a), strings.ToLower(b)
	}
	if !f.numeric {
		return strings.Compare(a, b)
	}

	numA, okA := leadingNumber(a)
	numB, okB := leadingNumber(b)
	switch {
	case !okA && !okB:
		return strings.Compare(a, b)
	case !okA:
		return 1
	case !okB:
		return -1
	case numA < numB:
		return -1
	case numA > numB:
		return 1
	}
	return 0
}

func (s *sorter) dedupe(records []sortRecord) []sortRecord {
	var out []sortRecord
	for _, r := range records {
		if len(out) > 0 && s.compare(out[len(out)-1], r) == 0 {
			continue
		}
		out = append(out, r)
	}
	return out
}

func (s *sorter) check(records []sortRecord, args []string) error {
	name := "-"
	if len(args) > 0 {
		name = args[0]
	}

	for i := 1; i < len(records); i++ {
		if s.compare(records[i-1], records[i]) > 0 {
			return fail(1, "sort: %s:%d: disorder: %s", name, i+1, records[i].line)
		}
	}
	return nil
}

func writeSorted(records []sortRecord, output string) error {
	var sb strings.Builder
	for _, r := range records {
		sb.WriteString(r.line)
		sb.WriteByte('\n')
	}

	if output == "" {
		fmt.Fprint(stdout, sb.String())
		return nil
	}
	if err := os.WriteFile(output, []byte(sb.String()), 0o644); err != nil {
		return fail(2, "sort: %v", err)
	}
	return nil
}

var leadingNumberRe = regexp.MustCompile(`^\s*[+-]?(\d+(\.\d*)?|\.\d+)`)

func leadingNumber(s string) (float64, bool) {
	match := leadingNumberRe.FindString(s)
	if match == "" {
		return 0, false
	}

	n, err := strconv.ParseFloat(strings.TrimSpace(match), 64)
	return n, err == nil
}
