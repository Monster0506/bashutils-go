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
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		deleteMode, _ := cmd.Flags().GetBool("delete")
		complement, _ := cmd.Flags().GetBool("complement")

		set1 := args[0]
		set2 := ""
		if len(args) == 2 {
			set2 = args[1]
		}

		// Error if delete mode but two sets provided
		if deleteMode && len(args) == 2 {
			fmt.Fprintf(os.Stderr, "tr: extra operand '%s'\n", set2)
			os.Exit(1)
		}

		expandedSet1 := expandCharSet(set1)
		expandedSet2 := expandCharSet(set2)

		set1Map := make(map[rune]bool)
		for _, r := range expandedSet1 {
			set1Map[r] = true
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

			if deleteMode {
				if complement {
					if !set1Map[r] {
						fmt.Print(string(r))
					}
				} else {
					if !set1Map[r] {
						fmt.Print(string(r))
					}
				}
			} else {
				idx := strings.IndexRune(expandedSet1, r)
				if idx != -1 {
					if idx < len(expandedSet2) {
						fmt.Print(string(expandedSet2[idx]))
					} else if len(expandedSet2) > 0 {
						fmt.Print(string(expandedSet2[len(expandedSet2)-1]))
					} else {
						// If set2 empty, delete char
					}
				} else {
					fmt.Print(string(r))
				}
			}
		}
	},
}

func init() {
	trCmd.Flags().BoolP("delete", "d", false, "delete characters in SET1")
	trCmd.Flags().BoolP("complement", "c", false, "use complement of SET1")
}

func expandCharSet(set string) string {
	var expanded []rune
	runes := []rune(set)

	for i := 0; i < len(runes); i++ {
		if i+2 < len(runes) && runes[i+1] == '-' {
			start := runes[i]
			end := runes[i+2]

			if start > end {
				start, end = end, start
			}

			for r := start; r <= end; r++ {
				expanded = append(expanded, r)
			}
			i += 2
		} else {
			expanded = append(expanded, runes[i])
		}
	}

	return string(expanded)
}
