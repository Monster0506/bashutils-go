package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var sortCmd = &cobra.Command{
	Use:   "sort [files...]",
	Short: "Sort lines of text files",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		reverse, _ := cmd.Flags().GetBool("reverse")
		numeric, _ := cmd.Flags().GetBool("numeric-sort")
		unique, _ := cmd.Flags().GetBool("unique")
		column, _ := cmd.Flags().GetInt("key")
		separator, _ := cmd.Flags().GetString("field-separator")

		allLines := readLines("sort", args)

		sort.SliceStable(allLines, func(i, j int) bool {
			var keyI, keyJ string

			if column > 0 {
				var columnsI, columnsJ []string
				if separator != "" {
					columnsI = strings.Split(allLines[i], separator)
					columnsJ = strings.Split(allLines[j], separator)
				} else {
					columnsI = strings.Fields(allLines[i])
					columnsJ = strings.Fields(allLines[j])
				}

				if column <= len(columnsI) {
					keyI = columnsI[column-1]
				}
				if column <= len(columnsJ) {
					keyJ = columnsJ[column-1]
				}
			} else {
				keyI = allLines[i]
				keyJ = allLines[j]
			}

			if numeric {
				numI, okI := leadingNumber(keyI)
				numJ, okJ := leadingNumber(keyJ)

				if !okI && !okJ {
					return keyI < keyJ
				}
				if !okI {
					return false
				}
				if !okJ {
					return true
				}
				return numI < numJ
			}

			return keyI < keyJ
		})

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
	sortCmd.Flags().IntP("key", "k", 0, "sort by the specified column (1-based index)")
	sortCmd.Flags().StringP("field-separator", "t", "", "use specified character as field separator")
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
