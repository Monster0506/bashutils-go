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
	before       int
	after        int
	printed      bool
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
		inputs, sawDir := grepInputs(cmd, fileArgs, g.errors)
		if sawDir && g.names == nameAuto {
			g.names = nameAlways
		}
		runInputs("grep", inputs, g.errors, func(in input, r io.Reader) {
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
	grepCmd.Flags().IntP("after-context", "A", 0, "print NUM lines of trailing context")
	grepCmd.Flags().IntP("before-context", "B", 0, "print NUM lines of leading context")
	grepCmd.Flags().IntP("context", "C", 0, "print NUM lines of output context")
	grepCmd.Flags().BoolP("recursive", "r", false, "read all files under each directory, recursively")
	grepCmd.Flags().BoolP("dereference-recursive", "R", false, "like -r")
	grepCmd.Flags().StringArrayVar(&grepIncludes, "include", nil, "search only files whose base name matches GLOB")
	grepCmd.Flags().StringArrayVar(&grepExcludes, "exclude", nil, "skip files whose base name matches GLOB")
	grepCmd.Flags().StringArrayVar(&grepExcludeDirs, "exclude-dir", nil, "skip directories whose base name matches GLOB")
}

func newGrepper(cmd *cobra.Command, re *regexp.Regexp) *grepper {
	invert, _ := cmd.Flags().GetBool("invert-match")
	lineNumber, _ := cmd.Flags().GetBool("line-number")
	maxCount, _ := cmd.Flags().GetInt("max-count")
	onlyMatching, _ := cmd.Flags().GetBool("only-matching")

	before, after := grepContext(cmd)
	g := &grepper{
		re:           re,
		invert:       invert,
		lineNumber:   lineNumber,
		names:        grepNames(cmd),
		errors:       reportErrors,
		onlyMatching: onlyMatching,
		output:       grepOutput(cmd),
		maxCount:     maxCount,
		before:       before,
		after:        after,
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

const (
	matchSep   = ":"
	contextSep = "-"
)

type numberedLine struct {
	num  int
	text string
}

func (g *grepper) searchStream(in input, r io.Reader) {
	if g.output == outQuiet && g.matched {
		return
	}

	name := g.displayName(in)
	show := g.showsName(in)
	scanner := utils.NewScanner(r)
	var before []numberedLine
	lineNum, count, afterLeft, lastPrinted := 0, 0, 0, 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if g.re.MatchString(line) == g.invert {
			if g.output != outLines {
				continue
			}
			if afterLeft > 0 {
				g.printLine(name, show, lineNum, contextSep, line)
				lastPrinted = lineNum
				afterLeft--
				continue
			}
			if g.before > 0 {
				before = append(before, numberedLine{lineNum, line})
				if len(before) > g.before {
					before = before[1:]
				}
			}
			continue
		}

		count++
		g.matched = true
		if g.output == outQuiet {
			return
		}
		if g.output == outLines {
			first := lineNum
			if len(before) > 0 {
				first = before[0].num
			}
			g.printSeparator(first, lastPrinted)
			for _, b := range before {
				g.printLine(name, show, b.num, contextSep, b.text)
			}
			before = before[:0]
			g.emit(name, show, lineNum, line)
			lastPrinted = lineNum
			afterLeft = g.after
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

func (g *grepper) printSeparator(first, lastPrinted int) {
	if g.before == 0 && g.after == 0 {
		return
	}
	if g.printed && (lastPrinted == 0 || first > lastPrinted+1) {
		fmt.Println("--")
	}
}

func (g *grepper) finishStream(name string, show bool, count int) {
	switch {
	case g.output == outCount:
		printPrefix(name, show, matchSep)
		fmt.Println(count)
	case g.output == outListFiles && count > 0:
		fmt.Println(name)
	case g.output == outListMissing && count == 0:
		fmt.Println(name)
	}
}

func printPrefix(name string, show bool, sep string) {
	if show {
		fmt.Print(name + sep)
	}
}

func (g *grepper) printLine(name string, show bool, lineNum int, sep, line string) {
	printPrefix(name, show, sep)
	if g.lineNumber {
		fmt.Printf("%d%s", lineNum, sep)
	}
	fmt.Println(line)
	g.printed = true
}

func (g *grepper) emit(name string, show bool, lineNum int, line string) {
	if !g.onlyMatching || g.invert {
		g.printLine(name, show, lineNum, matchSep, line)
		return
	}

	for _, m := range g.re.FindAllString(line, -1) {
		if m != "" {
			g.printLine(name, show, lineNum, matchSep, m)
		}
	}
}

func grepContext(cmd *cobra.Command) (int, int) {
	both, _ := cmd.Flags().GetInt("context")
	before, _ := cmd.Flags().GetInt("before-context")
	after, _ := cmd.Flags().GetInt("after-context")

	if !cmd.Flags().Changed("before-context") {
		before = both
	}
	if !cmd.Flags().Changed("after-context") {
		after = both
	}
	return before, after
}
