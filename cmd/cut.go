package cmd

import (
	"bufio"
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var cutCmd = &cobra.Command{
	Use:   "cut [files...]",
	Short: "Extract specific columns or byte ranges from lines",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fields, _ := cmd.Flags().GetString("fields")
		delimiter, _ := cmd.Flags().GetString("delimiter")
		characters, _ := cmd.Flags().GetString("characters")

		if (fields == "" && characters == "") || (fields != "" && characters != "") {
			fmt.Fprintf(os.Stderr, "cut: specify either --fields or --characters\n")
			return
		}

		list := fields
		if list == "" {
			list = characters
		}
		spans, err := parseRanges(list)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cut: invalid list %q: %v\n", list, err)
			os.Exit(1)
		}

		if len(args) == 0 {
			// Read from stdin when no files provided
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				if fields != "" {
					printFields(line, delimiter, spans)
				} else if characters != "" {
					printCharacters(line, spans)
				}
			}
		} else {
			// Expand glob patterns in file argument
			expandedFiles, err := utils.ExpandGlobsForReading(args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cut: %v\n", err)
				return
			}

			for _, path := range expandedFiles {
				file, err := os.Open(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "cut: %v\n", err)
					continue
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					if fields != "" {
						printFields(line, delimiter, spans)
					} else if characters != "" {
						printCharacters(line, spans)
					}
				}

				if err := scanner.Err(); err != nil {
					fmt.Fprintf(os.Stderr, "cut: reading input: %v\n", err)
				}
			}
		}
	},
}

func init() {
	cutCmd.Flags().StringP("fields", "f", "", "select fields by delimiter (e.g. '1,3')")
	cutCmd.Flags().StringP("delimiter", "d", "\t", "specify delimiter (default is TAB)")
	cutCmd.Flags().StringP("characters", "c", "", "select character positions (e.g. '1-5,7')")
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

func indicesFor(spans []span, n int) []int {
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
		if picked[i] {
			indices = append(indices, i)
		}
	}
	return indices
}

func printFields(line, delimiter string, spans []span) {
	if !strings.Contains(line, delimiter) {
		fmt.Println(line)
		return
	}

	parts := strings.Split(line, delimiter)

	var selectedFields []string
	for _, idx := range indicesFor(spans, len(parts)) {
		selectedFields = append(selectedFields, parts[idx-1])
	}
	fmt.Println(strings.Join(selectedFields, delimiter))
}

func printCharacters(line string, spans []span) {
	runes := []rune(line)
	var selectedChars []rune
	for _, idx := range indicesFor(spans, len(runes)) {
		selectedChars = append(selectedChars, runes[idx-1])
	}
	fmt.Println(string(selectedChars))
}
