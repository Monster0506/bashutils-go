package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

type numbering int

const (
	numberNone numbering = iota
	numberAll
	numberNonBlank
)

type catOpts struct {
	numbering   numbering
	squeeze     bool
	showEnds    bool
	showTabs    bool
	lineNumber  int
	prevBlank   bool
	passthrough bool
}

var catCmd = &cobra.Command{
	Use:   "cat [file]",
	Short: "Concatenate and display files",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		opts := newCatOpts(cmd)

		eachInput("cat", args, func(in input, r io.Reader) {
			var err error
			if opts.passthrough {
				_, err = io.Copy(os.Stdout, r)
			} else {
				err = opts.copyLines(r)
			}
			if err != nil {
				warn("cat: %s: %v\n", in.name, err)
			}
		})
	},
}

func init() {
	catCmd.Flags().BoolP("number", "n", false, "number all output lines")
	catCmd.Flags().BoolP("number-nonblank", "b", false, "number nonempty output lines, overrides -n")
	catCmd.Flags().BoolP("squeeze-blank", "s", false, "suppress repeated empty output lines")
	catCmd.Flags().BoolP("show-ends", "E", false, "display $ at end of each line")
	catCmd.Flags().BoolP("show-tabs", "T", false, "display TAB characters as ^I")
}

func newCatOpts(cmd *cobra.Command) *catOpts {
	number, _ := cmd.Flags().GetBool("number")
	nonBlank, _ := cmd.Flags().GetBool("number-nonblank")
	squeeze, _ := cmd.Flags().GetBool("squeeze-blank")
	ends, _ := cmd.Flags().GetBool("show-ends")
	tabs, _ := cmd.Flags().GetBool("show-tabs")

	o := &catOpts{squeeze: squeeze, showEnds: ends, showTabs: tabs}
	switch {
	case nonBlank:
		o.numbering = numberNonBlank
	case number:
		o.numbering = numberAll
	}
	o.passthrough = o.numbering == numberNone && !squeeze && !ends && !tabs
	return o
}

func (o *catOpts) copyLines(r io.Reader) error {
	scanner := utils.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		blank := text == ""
		if o.squeeze && blank && o.prevBlank {
			continue
		}
		o.prevBlank = blank

		if o.numbering == numberAll || (o.numbering == numberNonBlank && !blank) {
			o.lineNumber++
			fmt.Printf("%6d\t", o.lineNumber)
		}
		if o.showTabs {
			text = strings.ReplaceAll(text, "\t", "^I")
		}
		fmt.Print(text)
		if o.showEnds && !scanner.Unterminated() {
			fmt.Print("$")
		}
		if !scanner.Unterminated() {
			fmt.Println()
		}
	}
	return scanner.Err()
}
