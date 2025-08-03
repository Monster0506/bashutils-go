package cmd

import (
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"os"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
)

var sortCmd = &cobra.Command{
	Use:   "sort [files...]",
	Short: "Sort lines of text files",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		reverse, _ := cmd.Flags().GetBool("reverse")
		numeric, _ := cmd.Flags().GetBool("numeric-sort")
		unique, _ := cmd.Flags().GetBool("unique")

		allLines, err := utils.ReadLinesFromFilesOrStdin(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sort: %v\n", err)
			return
		}

		if numeric {
			sort.Slice(allLines, func(i, j int) bool {
				numI, errI := strconv.ParseFloat(allLines[i], 64)
				numJ, errJ := strconv.ParseFloat(allLines[j], 64)

				if errI != nil && errJ != nil {
					return allLines[i] < allLines[j] // Fallback to string sort if both are not numbers
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
			sort.Strings(allLines)
		}

		if reverse {
			for i, j := 0, len(allLines)-1; i < j; i, j = i+1, j-1 {
				allLines[i], allLines[j] = allLines[j], allLines[i]
			}
		}

		if unique {
			var uniqueLines []string
			if len(allLines) > 0 {
				uniqueLines = append(uniqueLines, allLines[0])
				for i := 1; i < len(allLines); i++ {
					if allLines[i] != allLines[i-1] {
						uniqueLines = append(uniqueLines, allLines[i])
					}
				}
			}
			allLines = uniqueLines
		}

		for _, line := range allLines {
			fmt.Println(line)
		}
	},
}

func init() {
	sortCmd.Flags().BoolP("reverse", "r", false, "sort in reverse order")
	sortCmd.Flags().BoolP("numeric-sort", "n", false, "compare according to string numerical value")
	sortCmd.Flags().BoolP("unique", "u", false, "output only the first of an equal run")
}
