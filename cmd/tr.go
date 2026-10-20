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
	RunE: func(cmd *cobra.Command, args []string) error {
		deleteMode, _ := cmd.Flags().GetBool("delete")
		complement, _ := cmd.Flags().GetBool("complement")

		if deleteMode && len(args) == 2 {
			return fail(1, "tr: extra operand '%s'", args[1])
		}

		t := &translator{
			set1:       expandCharSet(args[0]),
			delete:     deleteMode,
			complement: complement,
		}
		if len(args) == 2 {
			t.set2 = expandCharSet(args[1])
		}
		t.index = make(map[rune]int, len(t.set1))
		for i, r := range t.set1 {
			if _, seen := t.index[r]; !seen {
				t.index[r] = i
			}
		}

		if err := t.run(os.Stdin); err != nil {
			return fail(1, "tr: reading input: %v", err)
		}
		return nil
	},
}

type translator struct {
	set1       []rune
	set2       []rune
	index      map[rune]int
	delete     bool
	complement bool
}

func (t *translator) run(r io.Reader) error {
	reader := bufio.NewReader(r)
	for {
		c, _, err := reader.ReadRune()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if out, keep := t.mapRune(c); keep {
			fmt.Print(string(out))
		}
	}
}

func (t *translator) mapRune(r rune) (rune, bool) {
	idx, inSet := t.index[r]
	switch {
	case t.delete:
		return r, inSet == t.complement
	case t.complement:
		if inSet {
			return r, true
		}
		return t.last()
	case !inSet:
		return r, true
	case idx < len(t.set2):
		return t.set2[idx], true
	default:
		return t.last()
	}
}

func (t *translator) last() (rune, bool) {
	if len(t.set2) == 0 {
		return 0, false
	}
	return t.set2[len(t.set2)-1], true
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
