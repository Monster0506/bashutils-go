package cmd

import (
	"fmt"
	"io"
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

const stdinLabel = "(standard input)"

type grepper struct {
	re           *regexp.Regexp
	invert       bool
	lineNumber   bool
	names        nameMode
	errors       errorMode
	onlyMatching bool
	output       outputMode
	maxCount     int
	matched      bool
}

var grepExprs []string

var grepCmd = &cobra.Command{
	Use:   "grep [pattern] [files...]",
	Short: "Print lines matching a pattern",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		patterns, fileArgs, err := grepPatterns(cmd, args)
		if err != nil {
			return err
		}

		re, err := compilePatterns(patterns, grepKind(cmd), grepScope(cmd), grepCase(cmd))
		if err != nil {
			return fail(grepError, "grep: invalid regex pattern: %v", err)
		}

		g := newGrepper(cmd, re)
		eachInputMode("grep", fileArgs, g.errors, func(in input, r io.Reader) {
			g.searchStream(in, r)
		})

		if g.output == outQuiet && g.matched {
			return nil
		}
		if exitStatus != 0 {
			return exitWith(grepError)
		}
		if !g.matched {
			return exitWith(grepNoMatch)
		}
		return nil
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

func newGrepper(cmd *cobra.Command, re *regexp.Regexp) *grepper {
	invert, _ := cmd.Flags().GetBool("invert-match")
	lineNumber, _ := cmd.Flags().GetBool("line-number")
	maxCount, _ := cmd.Flags().GetInt("max-count")
	onlyMatching, _ := cmd.Flags().GetBool("only-matching")

	g := &grepper{
		re:           re,
		invert:       invert,
		lineNumber:   lineNumber,
		names:        grepNames(cmd),
		errors:       reportErrors,
		onlyMatching: onlyMatching,
		output:       grepOutput(cmd),
		maxCount:     maxCount,
	}
	if silent, _ := cmd.Flags().GetBool("no-messages"); silent {
		g.errors = silenceErrors
	}
	return g
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

func grepNames(cmd *cobra.Command) nameMode {
	if never, _ := cmd.Flags().GetBool("no-filename"); never {
		return nameNever
	}
	if always, _ := cmd.Flags().GetBool("with-filename"); always {
		return nameAlways
	}
	return nameAuto
}

func grepPatterns(cmd *cobra.Command, args []string) ([]string, []string, error) {
	file, _ := cmd.Flags().GetString("file")

	exprs := grepExprs
	if !cmd.Flags().Changed("regexp") && !cmd.Flags().Changed("file") {
		if len(args) == 0 {
			return nil, nil, fail(grepError, "grep: missing pattern")
		}
		exprs = []string{args[0]}
		args = args[1:]
	}

	var patterns []string
	for _, e := range exprs {
		patterns = append(patterns, strings.Split(e, "\n")...)
	}

	if file != "" {
		lines := readLines("grep", []string{file})
		if exitStatus != 0 {
			return nil, nil, exitWith(grepError)
		}
		patterns = append(patterns, lines...)
	}

	return patterns, args, nil
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

func (g *grepper) displayName(in input) string {
	if in.name == stdinName {
		return stdinLabel
	}
	return in.name
}

func (g *grepper) showsName(in input) bool {
	return g.names == nameAlways || (g.names == nameAuto && in.count > 1)
}

func (g *grepper) searchStream(in input, r io.Reader) {
	if g.output == outQuiet && g.matched {
		return
	}

	name := g.displayName(in)
	show := g.showsName(in)
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
			return
		}
		if g.output == outLines {
			g.emit(name, show, lineNum, line)
		}
		if g.output == outListFiles || g.output == outListMissing {
			break
		}
		if g.maxCount > 0 && count >= g.maxCount {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		inputError(g.errors, "grep: %s: %v\n", name, err)
	}

	g.finishStream(name, show, count)
}

func (g *grepper) finishStream(name string, show bool, count int) {
	switch {
	case g.output == outCount:
		printPrefix(name, show)
		fmt.Println(count)
	case g.output == outListFiles && count > 0:
		fmt.Println(name)
	case g.output == outListMissing && count == 0:
		fmt.Println(name)
	}
}

func printPrefix(name string, show bool) {
	if show {
		fmt.Printf("%s:", name)
	}
}

func (g *grepper) printLine(name string, show bool, lineNum int, line string) {
	printPrefix(name, show)
	if g.lineNumber {
		fmt.Printf("%d:", lineNum)
	}
	fmt.Println(line)
}

func (g *grepper) emit(name string, show bool, lineNum int, line string) {
	if !g.onlyMatching || g.invert {
		g.printLine(name, show, lineNum, line)
		return
	}

	for _, m := range g.re.FindAllString(line, -1) {
		if m != "" {
			g.printLine(name, show, lineNum, m)
		}
	}
}
