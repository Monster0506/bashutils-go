package cmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

const (
	grepNoMatch = 1
	grepError   = 2
)

type outputMode int

const (
	outLines outputMode = iota
	outCount
	outListFiles
	outListMissing
	outQuiet
)

type nameMode int

const (
	nameAuto nameMode = iota
	nameAlways
	nameNever
)

const stdinLabel = "(standard input)"

type grepper struct {
	re           *regexp.Regexp
	invert       bool
	lineNumber   bool
	showName     bool
	silent       bool
	onlyMatching bool
	output       outputMode
	maxCount     int
	matched      bool
	failed       bool
}

var grepExprs []string

type patternKind int

const (
	kindRegex patternKind = iota
	kindFixed
)

type matchScope int

const (
	scopeAny matchScope = iota
	scopeWord
	scopeLine
)

type caseMode int

const (
	matchCase caseMode = iota
	foldCase
)

var grepCmd = &cobra.Command{
	Use:   "grep [pattern] [files...]",
	Short: "Print lines matching a pattern",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		invertMatch, _ := cmd.Flags().GetBool("invert-match")
		lineNumber, _ := cmd.Flags().GetBool("line-number")
		maxCount, _ := cmd.Flags().GetInt("max-count")
		silent, _ := cmd.Flags().GetBool("no-messages")
		onlyMatching, _ := cmd.Flags().GetBool("only-matching")

		patterns, fileArgs := grepPatterns(cmd, args)

		re, err := compilePatterns(patterns, grepKind(cmd), grepScope(cmd), grepCase(cmd))
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: invalid regex pattern: %v\n", err)
			os.Exit(grepError)
		}

		g := &grepper{
			re:           re,
			invert:       invertMatch,
			lineNumber:   lineNumber,
			output:       grepOutput(cmd),
			maxCount:     maxCount,
			silent:       silent,
			onlyMatching: onlyMatching,
		}

		if len(fileArgs) == 0 {
			g.showName = grepNames(cmd, 1) == nameAlways
			g.searchStream(stdinLabel, os.Stdin)
		} else {
			files, err := utils.ExpandGlobs(fileArgs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "grep: %v\n", err)
				os.Exit(grepError)
			}

			g.showName = grepNames(cmd, len(files)) == nameAlways
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
	grepCmd.Flags().StringArrayVarP(&grepExprs, "regexp", "e", nil, "use PATTERN for matching (repeatable)")
	grepCmd.Flags().StringP("file", "f", "", "take patterns from FILE, one per line")
	grepCmd.Flags().BoolP("extended-regexp", "E", false, "PATTERN is an extended regular expression")
	grepCmd.Flags().BoolP("fixed-strings", "F", false, "PATTERN is a set of newline-separated strings")
	grepCmd.Flags().BoolP("word-regexp", "w", false, "match only whole words")
	grepCmd.Flags().BoolP("line-regexp", "x", false, "match only whole lines")
	grepCmd.Flags().BoolP("count", "c", false, "print only a count of matching lines per file")
	grepCmd.Flags().BoolP("files-with-matches", "l", false, "print only names of files with matches")
	grepCmd.Flags().BoolP("files-without-match", "L", false, "print only names of files without matches")
	grepCmd.Flags().BoolP("quiet", "q", false, "suppress all normal output")
	grepCmd.Flags().BoolP("no-messages", "s", false, "suppress error messages about missing files")
	grepCmd.Flags().IntP("max-count", "m", 0, "stop after NUM selected lines per file")
	grepCmd.Flags().BoolP("with-filename", "H", false, "print the file name for each match")
	grepCmd.Flags().BoolP("no-filename", "h", false, "suppress the file name prefix")
	grepCmd.Flags().Bool("help", false, "help for grep")
	grepCmd.Flags().BoolP("only-matching", "o", false, "print only the matched parts of a line")
}

func grepKind(cmd *cobra.Command) patternKind {
	if fixed, _ := cmd.Flags().GetBool("fixed-strings"); fixed {
		return kindFixed
	}
	return kindRegex
}

func grepScope(cmd *cobra.Command) matchScope {
	if line, _ := cmd.Flags().GetBool("line-regexp"); line {
		return scopeLine
	}
	if word, _ := cmd.Flags().GetBool("word-regexp"); word {
		return scopeWord
	}
	return scopeAny
}

func grepCase(cmd *cobra.Command) caseMode {
	if fold, _ := cmd.Flags().GetBool("ignore-case"); fold {
		return foldCase
	}
	return matchCase
}

func grepPatterns(cmd *cobra.Command, args []string) ([]string, []string) {
	exprs := grepExprs
	file, _ := cmd.Flags().GetString("file")

	if !cmd.Flags().Changed("regexp") && !cmd.Flags().Changed("file") {
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "grep: missing pattern\n")
			os.Exit(grepError)
		}
		exprs = []string{args[0]}
		args = args[1:]
	}

	var patterns []string
	for _, e := range exprs {
		patterns = append(patterns, strings.Split(e, "\n")...)
	}

	if file != "" {
		lines, err := readPatternFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: %v\n", err)
			os.Exit(grepError)
		}
		patterns = append(patterns, lines...)
	}

	return patterns, args
}

func readPatternFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return utils.ReadLines(f)
}

func compilePatterns(patterns []string, kind patternKind, scope matchScope, fold caseMode) (*regexp.Regexp, error) {
	if len(patterns) == 0 {
		return regexp.Compile(`[^\s\S]`)
	}

	parts := make([]string, len(patterns))
	for i, p := range patterns {
		if kind == kindFixed {
			p = regexp.QuoteMeta(p)
		}
		parts[i] = "(?:" + p + ")"
	}

	expr := strings.Join(parts, "|")
	switch scope {
	case scopeWord:
		expr = `\b(?:` + expr + `)\b`
	case scopeLine:
		expr = `^(?:` + expr + `)$`
	}
	if fold == foldCase {
		expr = "(?i)" + expr
	}

	return regexp.Compile(expr)
}

func grepOutput(cmd *cobra.Command) outputMode {
	flags := []struct {
		name string
		mode outputMode
	}{
		{"quiet", outQuiet},
		{"files-with-matches", outListFiles},
		{"files-without-match", outListMissing},
		{"count", outCount},
	}

	for _, f := range flags {
		if on, _ := cmd.Flags().GetBool(f.name); on {
			return f.mode
		}
	}
	return outLines
}

func grepNames(cmd *cobra.Command, fileCount int) nameMode {
	if never, _ := cmd.Flags().GetBool("no-filename"); never {
		return nameNever
	}
	if always, _ := cmd.Flags().GetBool("with-filename"); always || fileCount > 1 {
		return nameAlways
	}
	return nameNever
}

func (g *grepper) searchFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		if !g.silent {
			fmt.Fprintf(os.Stderr, "grep: %v\n", err)
		}
		g.failed = true
		return
	}
	defer f.Close()

	g.searchStream(path, f)
}

func (g *grepper) searchStream(name string, r io.Reader) {
	scanner := utils.NewScanner(r)
	lineNum, count := 0, 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if g.re.MatchString(line) == g.invert {
			continue
		}

		count++
		g.matched = true
		if g.output == outQuiet {
			os.Exit(0)
		}
		if g.output == outLines {
			g.emit(name, lineNum, line)
		}
		if g.output == outListFiles || g.output == outListMissing {
			break
		}
		if g.maxCount > 0 && count >= g.maxCount {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		if !g.silent {
			fmt.Fprintf(os.Stderr, "grep: reading input: %v\n", err)
		}
		g.failed = true
	}

	g.finishStream(name, count)
}

func (g *grepper) finishStream(name string, count int) {
	switch {
	case g.output == outCount:
		g.printPrefix(name)
		fmt.Println(count)
	case g.output == outListFiles && count > 0:
		fmt.Println(name)
	case g.output == outListMissing && count == 0:
		fmt.Println(name)
	}
}

func (g *grepper) printPrefix(name string) {
	if g.showName {
		fmt.Printf("%s:", name)
	}
}

func (g *grepper) printLine(name string, lineNum int, line string) {
	g.printPrefix(name)
	if g.lineNumber {
		fmt.Printf("%d:", lineNum)
	}
	fmt.Println(line)
}

func (g *grepper) emit(name string, lineNum int, line string) {
	if !g.onlyMatching || g.invert {
		g.printLine(name, lineNum, line)
		return
	}

	for _, m := range g.re.FindAllString(line, -1) {
		if m != "" {
			g.printLine(name, lineNum, m)
		}
	}
}
