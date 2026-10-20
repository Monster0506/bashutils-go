package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

var pasteCmd = &cobra.Command{
	Use:   "paste [files...]",
	Short: "Merge lines from files",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		delimitersStr, _ := cmd.Flags().GetString("delimiters")
		serial, _ := cmd.Flags().GetBool("serial")

		delimiters := []string{"\t"}
		if delimitersStr != "" {
			delimiters = parseDelimiters(delimitersStr)
		}

		scanners, closeAll, err := openScanners(resolveInputs("paste", args, reportErrors))
		defer closeAll()
		if err != nil {
			return err
		}

		if serial {
			return pasteSerial(scanners, delimiters)
		}
		return pasteParallel(scanners, delimiters)
	},
}

func pasteSerial(scanners []*utils.LineScanner, delimiters []string) error {
	for _, scanner := range scanners {
		var out strings.Builder
		for i := 0; scanner.Scan(); i++ {
			if i > 0 {
				out.WriteString(delimiters[(i-1)%len(delimiters)])
			}
			out.WriteString(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return fail(1, "paste: reading input: %v", err)
		}
		fmt.Println(out.String())
	}
	return nil
}

func pasteParallel(scanners []*utils.LineScanner, delimiters []string) error {
	for {
		currentLines := make([]string, len(scanners))
		moreData := false
		for i, scanner := range scanners {
			if scanner.Scan() {
				currentLines[i] = scanner.Text()
				moreData = true
				continue
			}
			if err := scanner.Err(); err != nil {
				return fail(1, "paste: reading input: %v", err)
			}
		}

		if !moreData {
			return nil
		}

		var out strings.Builder
		for i, line := range currentLines {
			out.WriteString(line)
			if i < len(currentLines)-1 {
				out.WriteString(delimiters[i%len(delimiters)])
			}
		}
		fmt.Println(out.String())
	}
}

func init() {
	pasteCmd.Flags().StringP("delimiters", "d", "", "use specified delimiters instead of TAB")
	pasteCmd.Flags().BoolP("serial", "s", false, "paste one file at a time instead of in parallel")
}

func openScanners(inputs []input) ([]*utils.LineScanner, func(), error) {
	var closers []io.Closer
	closeAll := func() {
		for _, c := range closers {
			c.Close()
		}
	}

	var stdin *utils.LineScanner
	scanners := make([]*utils.LineScanner, len(inputs))
	for i, in := range inputs {
		if in.name == stdinName && stdin != nil {
			scanners[i] = stdin
			continue
		}

		f := openInput("paste", in, reportErrors)
		if f == nil {
			return nil, closeAll, exitWith(1)
		}
		closers = append(closers, f)
		scanners[i] = utils.NewScanner(f)
		if in.name == stdinName {
			stdin = scanners[i]
		}
	}
	return scanners, closeAll, nil
}

var delimiterEscapes = map[rune]string{
	'n':  "\n",
	't':  "\t",
	'0':  "",
	'\\': "\\",
}

func parseDelimiters(s string) []string {
	var delims []string
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] != '\\' || i+1 == len(runes) {
			delims = append(delims, string(runes[i]))
			continue
		}

		i++
		if d, ok := delimiterEscapes[runes[i]]; ok {
			delims = append(delims, d)
			continue
		}
		delims = append(delims, string(runes[i]))
	}
	return delims
}
