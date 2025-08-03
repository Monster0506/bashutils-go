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

		// First pass: collect all counts to determine column width
		var allCounts []struct {
			line  string
			count int
		}
		var maxCountWidth int

		currentLine := allLines[0]
		currentCount := 1
		for i := 1; i < len(allLines); i++ {
			if allLines[i] == currentLine {
				currentCount++
			} else {
				if shouldPrintLine(currentCount, repeated, unique) {
					allCounts = append(allCounts, struct {
						line  string
						count int
					}{currentLine, currentCount})
					
					countStr := fmt.Sprintf("%d", currentCount)
					if len(countStr) > maxCountWidth {
						maxCountWidth = len(countStr)
					}
				}
				currentLine = allLines[i]
				currentCount = 1
			}
		}
		// Don't forget the last group
		if shouldPrintLine(currentCount, repeated, unique) {
			allCounts = append(allCounts, struct {
				line  string
				count int
			}{currentLine, currentCount})
			
			countStr := fmt.Sprintf("%d", currentCount)
			if len(countStr) > maxCountWidth {
				maxCountWidth = len(countStr)
			}
		}

		// Second pass: print with proper alignment
		for _, item := range allCounts {
			if count {
				fmt.Printf("%*d %s\n", maxCountWidth, item.count, item.line)
			} else {
				fmt.Println(item.line)
			}
		}
	},
}

func init() {
	uniqCmd.Flags().BoolP("count", "c", false, "prefix lines with occurrence count")
	uniqCmd.Flags().BoolP("repeated", "d", false, "print only duplicate lines")
	uniqCmd.Flags().BoolP("unique", "u", false, "print only unique lines (non-repeated)")
}

func shouldPrintLine(count int, showRepeated, showUnique bool) bool {
	if showRepeated && count == 1 {
		return false // Skip unique lines if only repeated are requested
	}
	if showUnique && count > 1 {
		return false // Skip repeated lines if only unique are requested
	}
	return true
}
