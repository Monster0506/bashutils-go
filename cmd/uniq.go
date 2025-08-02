package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

var uniqCmd = &cobra.Command{
	Use:   "uniq [file]",
	Short: "Filter out repeated lines",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		count, _ := cmd.Flags().GetBool("count")
		repeated, _ := cmd.Flags().GetBool("repeated")
		unique, _ := cmd.Flags().GetBool("unique")

		file, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "uniq: %v\n", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "uniq: reading input: %v\n", err)
			return
		}

		// uniq typically operates on sorted input.
		// For simplicity, we'll sort here if the input isn't guaranteed to be.
		// A more "unix-like" implementation would expect sorted input from a pipe.
		sort.Strings(lines)

		if len(lines) == 0 {
			return
		}

		currentLine := lines[0]
		currentCount := 1
		for i := 1; i < len(lines); i++ {
			if lines[i] == currentLine {
				currentCount++
			} else {
				printUniqLine(currentLine, currentCount, count, repeated, unique)
				currentLine = lines[i]
				currentCount = 1
			}
		}
		printUniqLine(currentLine, currentCount, count, repeated, unique)
	},
}

func init() {
	uniqCmd.Flags().BoolP("count", "c", false, "prefix lines with occurrence count")
	uniqCmd.Flags().BoolP("repeated", "d", false, "print only duplicate lines")
	uniqCmd.Flags().BoolP("unique", "u", false, "print only unique lines (non-repeated)")
}

func printUniqLine(line string, count int, showCount, showRepeated, showUnique bool) {
	if showRepeated && count == 1 {
		return // Skip unique lines if only repeated are requested
	}
	if showUnique && count > 1 {
		return // Skip repeated lines if only unique are requested
	}

	if showCount {
		fmt.Printf("%d %s\n", count, line)
	} else {
		fmt.Println(line)
	}
}
