package cmd

import (
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

var uniqCmd = &cobra.Command{
	Use:   "uniq [files...]",
	Short: "Filter out repeated lines",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		count, _ := cmd.Flags().GetBool("count")
		repeated, _ := cmd.Flags().GetBool("repeated")
		unique, _ := cmd.Flags().GetBool("unique")

		allLines, err := utils.ReadLinesFromFilesOrStdin(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "uniq: %v\n", err)
			return
		}

		// uniq typically operates on sorted input.
		// For simplicity, we'll sort here if the input isn't guaranteed to be.
		// A more "unix-like" implementation would expect sorted input from a pipe.
		sort.Strings(allLines)

		if len(allLines) == 0 {
			return
		}

		currentLine := allLines[0]
		currentCount := 1
		for i := 1; i < len(allLines); i++ {
			if allLines[i] == currentLine {
				currentCount++
			} else {
				printUniqLine(currentLine, currentCount, count, repeated, unique)
				currentLine = allLines[i]
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
