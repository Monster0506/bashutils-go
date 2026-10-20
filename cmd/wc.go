package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

type wcCount struct {
	label  string
	values map[string]int
}

const tabWidth = 8

var wcColumns = []string{"lines", "words", "chars", "bytes", "maxlen"}

var wcCmd = &cobra.Command{
	Use:   "wc [files...]",
	Short: "Print newline, word, byte, and char counts for each file",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var selected []string
		for _, name := range wcColumns {
			if on, _ := cmd.Flags().GetBool(name); on {
				selected = append(selected, name)
			}
		}
		if len(selected) == 0 {
			selected = []string{"lines", "words", "bytes"}
		}

		var counts []wcCount
		eachInput("wc", args, func(in input, r io.Reader) {
			data, err := io.ReadAll(r)
			if err != nil {
				warn("wc: %s: %v\n", in.name, err)
				return
			}
			counts = append(counts, countText(in.label, data))
		})

		printCounts(counts, selected)
	},
}

func init() {
	wcCmd.Flags().BoolP("lines", "l", false, "print newline count")
	wcCmd.Flags().BoolP("words", "w", false, "print word count")
	wcCmd.Flags().BoolP("bytes", "c", false, "print byte count")
	wcCmd.Flags().BoolP("chars", "m", false, "print character count")
	wcCmd.Flags().BoolP("maxlen", "L", false, "print the maximum display width")
}

func countText(label string, data []byte) wcCount {
	text := string(data)
	return wcCount{
		label: label,
		values: map[string]int{
			"lines":  strings.Count(text, "\n"),
			"words":  len(strings.Fields(text)),
			"chars":  utf8.RuneCountInString(text),
			"bytes":  len(data),
			"maxlen": maxLineWidth(text),
		},
	}
}

func maxLineWidth(text string) int {
	widest := 0
	for _, line := range strings.Split(text, "\n") {
		col := 0
		for _, r := range strings.TrimSuffix(line, "\r") {
			if r == '\t' {
				col += tabWidth - col%tabWidth
				continue
			}
			col++
		}
		widest = max(widest, col)
	}
	return widest
}

func printCounts(counts []wcCount, selected []string) {
	rows := counts
	if len(counts) > 1 {
		rows = append(rows, totalOf(counts, selected))
	}

	widths := make(map[string]int, len(selected))
	for _, row := range rows {
		for _, name := range selected {
			widths[name] = max(widths[name], len(strconv.Itoa(row.values[name])))
		}
	}

	for _, row := range rows {
		cells := make([]string, 0, len(selected)+1)
		for _, name := range selected {
			cells = append(cells, fmt.Sprintf("%*d", widths[name], row.values[name]))
		}
		if row.label != "" {
			cells = append(cells, row.label)
		}
		fmt.Fprintln(stdout, strings.Join(cells, " "))
	}
}

func totalOf(counts []wcCount, selected []string) wcCount {
	total := wcCount{label: "total", values: map[string]int{}}
	for _, c := range counts {
		for _, name := range selected {
			if name == "maxlen" {
				total.values[name] = max(total.values[name], c.values[name])
				continue
			}
			total.values[name] += c.values[name]
		}
	}
	return total
}
