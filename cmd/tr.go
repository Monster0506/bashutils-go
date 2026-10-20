package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

type translator struct {
	set1       []rune
	set2       []rune
	index      map[rune]int
	delete     bool
	complement bool
	squeeze    func(rune) bool
	last       rune
	haveLast   bool
}

var trCmd = &cobra.Command{
	Use:   "tr [SET1] [SET2]",
	Short: "Translate or delete characters",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		deleteMode, _ := cmd.Flags().GetBool("delete")
		complement, _ := cmd.Flags().GetBool("complement")
		squeeze, _ := cmd.Flags().GetBool("squeeze-repeats")
		truncate, _ := cmd.Flags().GetBool("truncate-set1")

		if err := checkTrOperands(args, deleteMode, squeeze); err != nil {
			return err
		}

		set1, err := parseCharSet(args[0])
		if err != nil {
			return fail(1, "tr: %v", err)
		}
		var set2 []rune
		if len(args) == 2 {
			if set2, err = parseCharSet(args[1]); err != nil {
				return fail(1, "tr: %v", err)
			}
		}
		if truncate && len(set2) > 0 && len(set2) < len(set1) {
			set1 = set1[:len(set2)]
		}

		t := &translator{set1: set1, set2: set2, delete: deleteMode, complement: complement}
		t.index = make(map[rune]int, len(set1))
		for i, r := range set1 {
			if _, seen := t.index[r]; !seen {
				t.index[r] = i
			}
		}
		if squeeze {
			t.squeeze = squeezeTest(t, len(args))
		}

		if err := t.run(os.Stdin); err != nil {
			return fail(1, "tr: reading input: %v", err)
		}
		return nil
	},
}

func init() {
	trCmd.Flags().BoolP("delete", "d", false, "delete characters in SET1")
	trCmd.Flags().BoolP("complement", "c", false, "use complement of SET1")
	trCmd.Flags().BoolP("squeeze-repeats", "s", false, "replace each run of a repeated character with one occurrence")
	trCmd.Flags().BoolP("truncate-set1", "t", false, "first truncate SET1 to length of SET2")
}

func checkTrOperands(args []string, deleteMode, squeeze bool) error {
	switch {
	case deleteMode && !squeeze && len(args) == 2:
		return fail(1, "tr: extra operand '%s'", args[1])
	case deleteMode && squeeze && len(args) == 1:
		return fail(1, "tr: missing operand after '%s'", args[0])
	case !deleteMode && !squeeze && len(args) == 1:
		return fail(1, "tr: missing operand after '%s'", args[0])
	}
	return nil
}

func squeezeTest(t *translator, operands int) func(rune) bool {
	if operands == 2 {
		set := make(map[rune]bool, len(t.set2))
		for _, r := range t.set2 {
			set[r] = true
		}
		return func(r rune) bool { return set[r] }
	}

	return func(r rune) bool {
		_, inSet := t.index[r]
		return inSet != t.complement
	}
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

		out, keep := t.mapRune(c)
		if !keep {
			continue
		}
		if t.squeeze != nil && t.haveLast && out == t.last && t.squeeze(out) {
			continue
		}

		fmt.Print(string(out))
		t.last, t.haveLast = out, true
	}
}

func (t *translator) mapRune(r rune) (rune, bool) {
	idx, inSet := t.index[r]
	switch {
	case t.delete:
		return r, inSet == t.complement
	case len(t.set2) == 0:
		return r, true
	case t.complement:
		if inSet {
			return r, true
		}
		return t.lastOfSet2()
	case !inSet:
		return r, true
	case idx < len(t.set2):
		return t.set2[idx], true
	default:
		return t.lastOfSet2()
	}
}

func (t *translator) lastOfSet2() (rune, bool) {
	return t.set2[len(t.set2)-1], true
}

type setItem struct {
	r       rune
	literal bool
	class   bool
}

var trEscapes = map[rune]rune{
	'a': '\a', 'b': '\b', 'f': '\f', 'n': '\n', 'r': '\r', 't': '\t', 'v': '\v', '\\': '\\',
}

var trClasses = map[string]func(rune) bool{
	"alpha":  unicode.IsLetter,
	"digit":  unicode.IsDigit,
	"alnum":  func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) },
	"upper":  unicode.IsUpper,
	"lower":  unicode.IsLower,
	"space":  unicode.IsSpace,
	"blank":  func(r rune) bool { return r == ' ' || r == '\t' },
	"punct":  func(r rune) bool { return unicode.IsPunct(r) || unicode.IsSymbol(r) },
	"cntrl":  unicode.IsControl,
	"print":  unicode.IsPrint,
	"graph":  func(r rune) bool { return unicode.IsGraphic(r) && r != ' ' },
	"xdigit": func(r rune) bool { return strings.ContainsRune("0123456789abcdefABCDEF", r) },
}

const asciiLimit = 128

func parseCharSet(set string) ([]rune, error) {
	items, err := tokenizeSet([]rune(set))
	if err != nil {
		return nil, err
	}

	var out []rune
	for i := 0; i < len(items); i++ {
		isRange := i+2 < len(items) && items[i+1].r == '-' && !items[i+1].literal &&
			!items[i].class && !items[i+2].class
		if !isRange {
			out = append(out, items[i].r)
			continue
		}

		start, end := items[i].r, items[i+2].r
		if start > end {
			return nil, fmt.Errorf("range-endpoints of '%c-%c' are in reverse collating sequence order", start, end)
		}
		for r := start; r <= end; r++ {
			out = append(out, r)
		}
		i += 2
	}
	return out, nil
}

func tokenizeSet(runes []rune) ([]setItem, error) {
	var items []setItem
	for i := 0; i < len(runes); i++ {
		switch {
		case runes[i] == '\\' && i+1 < len(runes):
			r, used := parseEscape(runes[i+1:])
			items = append(items, setItem{r: r, literal: true})
			i += used
		case runes[i] == '[' && i+1 < len(runes) && runes[i+1] == ':':
			name, ok := className(runes[i+2:])
			if !ok {
				items = append(items, setItem{r: runes[i]})
				continue
			}
			members, err := classMembers(name)
			if err != nil {
				return nil, err
			}
			items = append(items, members...)
			i += len(name) + 3
		default:
			items = append(items, setItem{r: runes[i]})
		}
	}
	return items, nil
}

func parseEscape(rest []rune) (rune, int) {
	if rest[0] >= '0' && rest[0] <= '7' {
		digits := 0
		for digits < len(rest) && digits < 3 && rest[digits] >= '0' && rest[digits] <= '7' {
			digits++
		}
		n, _ := strconv.ParseInt(string(rest[:digits]), 8, 32)
		return rune(n), digits
	}
	if esc, ok := trEscapes[rest[0]]; ok {
		return esc, 1
	}
	return rest[0], 1
}

func className(rest []rune) (string, bool) {
	for i := 0; i+1 < len(rest); i++ {
		if rest[i] == ':' && rest[i+1] == ']' {
			return string(rest[:i]), true
		}
	}
	return "", false
}

func classMembers(name string) ([]setItem, error) {
	test, ok := trClasses[name]
	if !ok {
		return nil, fmt.Errorf("invalid character class '%s'", name)
	}

	var items []setItem
	for r := rune(0); r < asciiLimit; r++ {
		if test(r) {
			items = append(items, setItem{r: r, class: true})
		}
	}
	return items, nil
}
