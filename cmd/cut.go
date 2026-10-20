package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

type cutMode int

const (
	cutFields cutMode = iota
	cutChars
	cutBytes
)

type selection int

const (
	selectListed selection = iota
	selectComplement
)

type cutter struct {
	mode          cutMode
	spans         []span
	delimiter     string
	outDelimiter  string
	hasOutDelim   bool
	onlyDelimited bool
	selection     selection
}

var cutCmd = &cobra.Command{
	Use:   "cut [files...]",
	Short: "Extract specific columns or byte ranges from lines",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newCutter(cmd)
		if err != nil {
			return err
		}

		eachInput("cut", args, func(in input, r io.Reader) {
			scanner := utils.NewScanner(r)
			for scanner.Scan() {
				if out, ok := c.cutLine(scanner.Text()); ok {
					fmt.Println(out)
				}
			}

			if err := scanner.Err(); err != nil {
				warn("cut: %s: %v\n", in.name, err)
			}
		})
		return nil
	},
}

func init() {
	cutCmd.Flags().StringP("fields", "f", "", "select fields by delimiter (e.g. '1,3')")
	cutCmd.Flags().StringP("delimiter", "d", "\t", "specify delimiter (default is TAB)")
	cutCmd.Flags().StringP("characters", "c", "", "select character positions (e.g. '1-5,7')")
	cutCmd.Flags().StringP("bytes", "b", "", "select byte positions (e.g. '1-5,7')")
	cutCmd.Flags().BoolP("only-delimited", "s", false, "do not print lines not containing delimiters")
	cutCmd.Flags().Bool("complement", false, "complement the set of selected fields, characters or bytes")
	cutCmd.Flags().String("output-delimiter", "", "use STRING as the output delimiter")
}

func newCutter(cmd *cobra.Command) (*cutter, error) {
	modes := []struct {
		flag string
		mode cutMode
	}{
		{"fields", cutFields},
		{"characters", cutChars},
		{"bytes", cutBytes},
	}

	c := &cutter{}
	list, chosen := "", 0
	for _, m := range modes {
		if v, _ := cmd.Flags().GetString(m.flag); v != "" {
			list, c.mode = v, m.mode
			chosen++
		}
	}
	if chosen != 1 {
		return nil, fail(1, "cut: specify exactly one of --fields, --characters or --bytes")
	}

	spans, err := parseRanges(list)
	if err != nil {
		return nil, fail(1, "cut: invalid list %q: %v", list, err)
	}
	c.spans = spans

	c.delimiter, _ = cmd.Flags().GetString("delimiter")
	c.outDelimiter, _ = cmd.Flags().GetString("output-delimiter")
	c.hasOutDelim = cmd.Flags().Changed("output-delimiter")
	c.onlyDelimited, _ = cmd.Flags().GetBool("only-delimited")
	if complement, _ := cmd.Flags().GetBool("complement"); complement {
		c.selection = selectComplement
	}
	return c, nil
}

func (c *cutter) cutLine(line string) (string, bool) {
	switch c.mode {
	case cutFields:
		return c.cutFields(line)
	case cutChars:
		runes := []rune(line)
		return c.cutUnits(len(runes), func(i int) string { return string(runes[i]) }), true
	default:
		return c.cutUnits(len(line), func(i int) string { return line[i : i+1] }), true
	}
}

func (c *cutter) cutFields(line string) (string, bool) {
	if !strings.Contains(line, c.delimiter) {
		return line, !c.onlyDelimited
	}

	parts := strings.Split(line, c.delimiter)
	var selected []string
	for _, idx := range indicesFor(c.spans, len(parts), c.selection) {
		selected = append(selected, parts[idx-1])
	}

	sep := c.delimiter
	if c.hasOutDelim {
		sep = c.outDelimiter
	}
	return strings.Join(selected, sep), true
}

func (c *cutter) cutUnits(n int, unit func(i int) string) string {
	var sb strings.Builder
	prev := 0
	for _, idx := range indicesFor(c.spans, n, c.selection) {
		if c.hasOutDelim && prev > 0 && idx != prev+1 {
			sb.WriteString(c.outDelimiter)
		}
		sb.WriteString(unit(idx - 1))
		prev = idx
	}
	return sb.String()
}

const openEnd = -1

type span struct {
	start, end int
}

func parseRanges(input string) ([]span, error) {
	var spans []span
	for _, part := range strings.Split(input, ",") {
		first, last, isRange := strings.Cut(part, "-")

		if !isRange {
			idx, err := strconv.Atoi(first)
			if err != nil {
				return nil, err
			}
			spans = append(spans, span{idx, idx})
			continue
		}

		start := 1
		if first != "" {
			n, err := strconv.Atoi(first)
			if err != nil {
				return nil, err
			}
			start = n
		}

		end := openEnd
		if last != "" {
			n, err := strconv.Atoi(last)
			if err != nil {
				return nil, err
			}
			end = n
		}

		spans = append(spans, span{start, end})
	}
	return spans, nil
}

func indicesFor(spans []span, n int, sel selection) []int {
	picked := make([]bool, n+1)
	for _, s := range spans {
		end := s.end
		if end == openEnd || end > n {
			end = n
		}
		for i := max(s.start, 1); i <= end; i++ {
			picked[i] = true
		}
	}

	var indices []int
	for i := 1; i <= n; i++ {
		if picked[i] == (sel == selectListed) {
			indices = append(indices, i)
		}
	}
	return indices
}
