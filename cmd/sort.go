package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
)

var sortCmd = &cobra.Command{
	Use:   "sort [file]",
	Short: "Sort lines of text files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		reverse, _ := cmd.Flags().GetBool("reverse")
		numeric, _ := cmd.Flags().GetBool("numeric-sort")
		unique, _ := cmd.Flags().GetBool("unique")

		file, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "sort: %v\n", err)
			return
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "sort: reading input: %v\n", err)
			return
		}

		if numeric {
			sort.Slice(lines, func(i, j int) bool {
				numI, errI := strconv.ParseFloat(lines[i], 64)
				numJ, errJ := strconv.ParseFloat(lines[j], 64)

				if errI != nil && errJ != nil {
					return lines[i] < lines[j] // Fallback to string sort if both are not numbers
				}
				if errI != nil {
					return false // Non-numeric treated as larger if compare to numeric
				}
				if errJ != nil {
					return true // Non-numeric treated as larger if compare to numeric
				}
				return numI < numJ
			})
		} else {
			sort.Strings(lines)
		}

		if reverse {
			for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
				lines[i], lines[j] = lines[j], lines[i]
			}
		}

		if unique {
			var uniqueLines []string
			if len(lines) > 0 {
				uniqueLines = append(uniqueLines, lines[0])
				for i := 1; i < len(lines); i++ {
					if lines[i] != lines[i-1] {
						uniqueLines = append(uniqueLines, lines[i])
					}
				}
			}
			lines = uniqueLines
		}

		for _, line := range lines {
			fmt.Println(line)
		}
	},
}

func init() {
	sortCmd.Flags().BoolP("reverse", "r", false, "sort in reverse order")
	sortCmd.Flags().BoolP("numeric-sort", "n", false, "compare according to string numerical value")
	sortCmd.Flags().BoolP("unique", "u", false, "output only the first of an equal run")
}
