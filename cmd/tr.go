package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var trCmd = &cobra.Command{
	Use:   "tr [SET1] [SET2]",
	Short: "Translate or delete characters",
	Args:  cobra.RangeArgs(1, 2), // SET1 for delete, SET1 and SET2 for translate
	Run: func(cmd *cobra.Command, args []string) {
		deleteMode, _ := cmd.Flags().GetBool("delete")
		complement, _ := cmd.Flags().GetBool("complement")

		set1 := args[0]
		set2 := ""
		if len(args) == 2 {
			set2 = args[1]
		}

		if deleteMode && len(args) == 2 {
			fmt.Fprintf(os.Stderr, "tr: extra operand '%s'\n", set2)
			os.Exit(1)
		}

		inputReader := bufio.NewReader(os.Stdin)
		for {
			r, _, err := inputReader.ReadRune()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "tr: reading input: %v\n", err)
				os.Exit(1)
			}

			char := string(r)
			var processedChar string

			if deleteMode {
				if complement {
					if !strings.ContainsRune(expandCharSet(set1), r) {
						processedChar = char
					}
				} else {
					if !strings.ContainsRune(expandCharSet(set1), r) {
						processedChar = char
					}
				}
			} else { // Translate mode
				expandedSet1 := expandCharSet(set1)
				expandedSet2 := expandCharSet(set2)

				idx := strings.IndexRune(expandedSet1, r)
				if idx != -1 {
					if idx < len(expandedSet2) {
						processedChar = string(expandedSet2[idx])
					} else { // Handle cases where set2 is shorter than set1
						if len(expandedSet2) > 0 {
							processedChar = string(expandedSet2[len(expandedSet2)-1]) // Repeat last char
						} else {
							processedChar = "" // Delete if set2 is empty
						}
					}
				} else {
					processedChar = char
				}
			}
			fmt.Print(processedChar)
		}
	},
}

func init() {
	trCmd.Flags().BoolP("delete", "d", false, "delete characters in SET1")
	trCmd.Flags().BoolP("complement", "c", false, "use complement of SET1") // Common for tr, though not explicitly listed in original example
}

// expandCharSet expands character ranges like 'a-z' into a full string of characters.
func expandCharSet(set string) string {
	var expanded []rune
	for i := 0; i < len(set); i++ {
		if i+2 < len(set) && set[i+1] == '-' {
			start := rune(set[i])
			end := rune(set[i+2])
			if start > end {
				// Handle inverse ranges (e.g., 'Z-A') if needed, for now assume ascending
				start, end = end, start
			}
			for r := start; r <= end; r++ {
				expanded = append(expanded, r)
			}
			i += 2 // Skip the character after '-'
		} else {
			expanded = append(expanded, rune(set[i]))
		}
	}
	return string(expanded)
}
