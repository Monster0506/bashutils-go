package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var trCmd = &cobra.Command{
	Use:   "tr [SET1] [SET2]",
	Short: "Translate or delete characters",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		deleteMode, _ := cmd.Flags().GetBool("delete")
		complement, _ := cmd.Flags().GetBool("complement")

		if deleteMode && len(args) == 2 {
			fmt.Fprintf(os.Stderr, "tr: extra operand '%s'\n", args[1])
			os.Exit(1)
		}

		set1 := expandCharSet(args[0])
		var set2 []rune
		if len(args) == 2 {
			set2 = expandCharSet(args[1])
		}

		index := make(map[rune]int, len(set1))
		for i, r := range set1 {
			if _, seen := index[r]; !seen {
				index[r] = i
			}
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

			idx, inSet := index[r]

			if deleteMode {
				if inSet == complement {
					fmt.Print(string(r))
				}
				continue
			}

			if complement {
				if inSet {
					fmt.Print(string(r))
					continue
				}
				if len(set2) > 0 {
					fmt.Print(string(set2[len(set2)-1]))
				}
				continue
			}

			if !inSet {
				fmt.Print(string(r))
				continue
			}
			if idx < len(set2) {
				fmt.Print(string(set2[idx]))
			} else if len(set2) > 0 {
				fmt.Print(string(set2[len(set2)-1]))
			}
		}
	},
}

func init() {
	trCmd.Flags().BoolP("delete", "d", false, "delete characters in SET1")
	trCmd.Flags().BoolP("complement", "c", false, "use complement of SET1")
}

func expandCharSet(set string) []rune {
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

	return expanded
}
