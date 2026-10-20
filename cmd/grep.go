package cmd

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

const (
	grepNoMatch = 1
	grepError   = 2
)

type grepper struct {
	re         *regexp.Regexp
	invert     bool
	lineNumber bool
	showName   bool
	matched    bool
	failed     bool
}

var grepCmd = &cobra.Command{
	Use:   "grep [pattern] [files...]",
	Short: "Print lines matching a pattern",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ignoreCase, _ := cmd.Flags().GetBool("ignore-case")
		invertMatch, _ := cmd.Flags().GetBool("invert-match")
		lineNumber, _ := cmd.Flags().GetBool("line-number")
		regexpFlag, _ := cmd.Flags().GetString("regexp")

		var patternStr string
		var fileArgs []string

		if cmd.Flags().Changed("regexp") {
			patternStr = regexpFlag
			fileArgs = args
		} else {
			if len(args) == 0 {
				fmt.Fprintf(os.Stderr, "grep: missing pattern\n")
				os.Exit(grepError)
			}
			patternStr = args[0]
			fileArgs = args[1:]
		}

		if ignoreCase {
			patternStr = "(?i)" + patternStr
		}

		re, err := regexp.Compile(patternStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: invalid regex pattern: %v\n", err)
			os.Exit(grepError)
		}

		g := &grepper{re: re, invert: invertMatch, lineNumber: lineNumber}

		if len(fileArgs) == 0 {
			g.searchStream("", os.Stdin)
		} else {
			files, err := utils.ExpandGlobs(fileArgs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "grep: %v\n", err)
				os.Exit(grepError)
			}

			g.showName = len(files) > 1
			for _, path := range files {
				g.searchFile(path)
			}
		}

		if g.failed {
			os.Exit(grepError)
		}
		if !g.matched {
			os.Exit(grepNoMatch)
		}
	},
}

func init() {
	grepCmd.Flags().BoolP("ignore-case", "i", false, "ignore case distinctions")
	grepCmd.Flags().BoolP("invert-match", "v", false, "select non-matching lines")
	grepCmd.Flags().BoolP("line-number", "n", false, "show line numbers")
	grepCmd.Flags().StringP("regexp", "e", "", "use a specific regex pattern")
}

func (g *grepper) searchFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grep: %v\n", err)
		g.failed = true
		return
	}
	defer f.Close()

	g.searchStream(path, f)
}

func (g *grepper) searchStream(name string, r io.Reader) {
	scanner := utils.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if g.re.MatchString(line) == g.invert {
			continue
		}

		g.matched = true
		g.printLine(name, lineNum, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "grep: reading input: %v\n", err)
		g.failed = true
	}
}

func (g *grepper) printLine(name string, lineNum int, line string) {
	if g.showName {
		fmt.Printf("%s:", name)
	}
	if g.lineNumber {
		fmt.Printf("%d:", lineNum)
	}
	fmt.Println(line)
}
