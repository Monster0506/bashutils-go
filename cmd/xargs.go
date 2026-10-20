package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

var xargsCmd = &cobra.Command{
	Use:   "xargs [command] [args...]",
	Short: "Build and execute command lines from standard input",
	Long: `xargs reads items from standard input, delimited by blanks (which can be protected
with double or single quotes or a backslash) or newlines, and executes the command
(default is /bin/echo) one or more times with any initial-arguments followed by
items read from standard input.`,
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, commandArgs, err := parseXargsArgs(args)
		if errors.Is(err, errXargsHelp) {
			return cmd.Help()
		}
		if err != nil {
			return fail(1, "xargs: %v", err)
		}

		delimiter := opts.delimiter
		if opts.replace != "" && !opts.null && delimiter == "" {
			delimiter = "\n"
		}

		items, err := readItems(os.Stdin, opts.null, delimiter)
		if err != nil {
			return fail(1, "xargs: error reading input: %v", err)
		}

		if len(commandArgs) == 0 {
			commandArgs = []string{"bashutils", "echo"}
		}

		if len(items) == 0 {
			if !opts.noRun {
				executeCommand(commandArgs, nil, opts.replace, opts.verbose)
			}
			return nil
		}

		batch := len(items)
		if opts.maxArgs > 0 {
			batch = opts.maxArgs
		}
		for i := 0; i < len(items); i += batch {
			end := min(i+batch, len(items))
			executeCommand(commandArgs, items[i:end], opts.replace, opts.verbose)
		}
		return nil
	},
}

var errXargsHelp = errors.New("help requested")

type xargsOpts struct {
	maxArgs   int
	replace   string
	null      bool
	delimiter string
	noRun     bool
	verbose   bool
}

const xargsValued = "nId"

var xargsLong = map[string]rune{
	"max-args":        'n',
	"replace":         'I',
	"null":            '0',
	"delimiter":       'd',
	"no-run-if-empty": 'r',
	"verbose":         't',
	"help":            'h',
}

func parseXargsArgs(args []string) (xargsOpts, []string, error) {
	var o xargsOpts
	i := 0
	for ; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			i++
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			break
		}

		if strings.HasPrefix(arg, "--") {
			name, value, hasValue := strings.Cut(arg[2:], "=")
			letter, ok := xargsLong[name]
			if !ok {
				return o, nil, fmt.Errorf("invalid option -- '%s'", arg)
			}
			used, err := o.apply(letter, value, hasValue, args[i+1:])
			if err != nil {
				return o, nil, err
			}
			i += used
			continue
		}

		for j := 1; j < len(arg); j++ {
			letter := rune(arg[j])
			if !strings.ContainsRune(xargsValued, letter) {
				if _, err := o.apply(letter, "", false, nil); err != nil {
					return o, nil, err
				}
				continue
			}

			used, err := o.apply(letter, arg[j+1:], j+1 < len(arg), args[i+1:])
			if err != nil {
				return o, nil, err
			}
			i += used
			break
		}
	}

	return o, args[i:], nil
}

func (o *xargsOpts) apply(letter rune, value string, hasValue bool, rest []string) (int, error) {
	used := 0
	if strings.ContainsRune(xargsValued, letter) && !hasValue {
		if len(rest) == 0 {
			return 0, fmt.Errorf("option requires an argument -- '%c'", letter)
		}
		value = rest[0]
		used = 1
	}

	switch letter {
	case 'n':
		n, err := strconv.Atoi(value)
		if err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid number %q for -n option", value)
		}
		o.maxArgs = n
	case 'I':
		o.replace = value
	case 'd':
		d, err := singleDelimiter(value)
		if err != nil {
			return 0, err
		}
		o.delimiter = d
	case '0':
		o.null = true
	case 'r':
		o.noRun = true
	case 't':
		o.verbose = true
	case 'h':
		return 0, errXargsHelp
	default:
		return 0, fmt.Errorf("invalid option -- '%c'", letter)
	}

	return used, nil
}

func readItems(r io.Reader, nullTerminated bool, delimiter string) ([]string, error) {
	if nullTerminated {
		delimiter = "\x00"
	}
	if delimiter == "" {
		return readQuotedItems(r)
	}

	scanner := utils.NewScanner(r)
	scanner.Split(splitOnByte(delimiter[0]))

	var items []string
	for scanner.Scan() {
		items = append(items, trimCR(scanner.Text(), delimiter))
	}
	return items, scanner.Err()
}

func splitOnByte(delim byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		if i := bytes.IndexByte(data, delim); i >= 0 {
			return i + 1, data[:i], nil
		}
		if atEOF && len(data) > 0 {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
}

var delimiterArgEscapes = map[string]string{
	`\n`: "\n",
	`\t`: "\t",
	`\0`: "\x00",
	`\\`: "\\",
}

func singleDelimiter(value string) (string, error) {
	if d, ok := delimiterArgEscapes[value]; ok {
		return d, nil
	}
	if len(value) != 1 {
		return "", fmt.Errorf("delimiter must be a single byte, got %q", value)
	}
	return value, nil
}

func trimCR(item, delimiter string) string {
	if delimiter == "\n" {
		return strings.TrimSuffix(item, "\r")
	}
	return item
}

var errUnmatchedQuote = errors.New("unmatched quote")

func readQuotedItems(r io.Reader) ([]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var items []string
	var cur strings.Builder
	var quote rune
	inItem := false
	runes := []rune(string(data))

	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch {
		case quote != 0:
			if c == '\n' {
				return nil, errUnmatchedQuote
			}
			if c == quote {
				quote = 0
				continue
			}
			cur.WriteRune(c)
		case c == '\\' && i+1 < len(runes) && isEscapable(runes[i+1]):
			inItem = true
			i++
			cur.WriteRune(runes[i])
		case c == '\'' || c == '"':
			quote = c
			inItem = true
		case unicode.IsSpace(c):
			if inItem {
				items = append(items, cur.String())
				cur.Reset()
				inItem = false
			}
		default:
			cur.WriteRune(c)
			inItem = true
		}
	}

	if quote != 0 {
		return nil, errUnmatchedQuote
	}
	if inItem {
		items = append(items, cur.String())
	}
	return items, nil
}

func isEscapable(r rune) bool {
	return r == '\'' || r == '"' || unicode.IsSpace(r)
}
func executeCommand(command []string, args []string, replaceStr string, verbose bool) {
	if replaceStr != "" {
		for _, arg := range args {
			replaced := make([]string, len(command))
			for i, part := range command {
				replaced[i] = strings.ReplaceAll(part, replaceStr, arg)
			}
			executeSingleCommand(replaced, verbose)
		}
	} else {
		// Build the final command arguments
		var finalArgs []string
		finalArgs = append(finalArgs, command...)
		finalArgs = append(finalArgs, args...)
		executeSingleCommand(finalArgs, verbose)
	}
}

const (
	childFailedStatus = 123
	selfName          = "bashutils"
)

func executeSingleCommand(args []string, verbose bool) {
	if verbose {
		fmt.Fprintln(os.Stderr, strings.Join(args, " "))
	}

	cmd := buildCommand(args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "xargs: command failed: %v\n", err)
		exitStatus = childFailedStatus
	}
}

func buildCommand(args []string) *exec.Cmd {
	if args[0] == selfName && len(args) > 1 {
		args = args[1:]
	}

	exe, err := os.Executable()
	if err != nil || !isBuiltin(args[0]) {
		return exec.Command(args[0], args[1:]...)
	}

	cmd := exec.Command(exe, args...)
	cmd.Args[0] = selfName
	return cmd
}

func isBuiltin(name string) bool {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name && !nonToolCmds[name] {
			return true
		}
	}
	return false
}

func init() {
	xargsCmd.Flags().IntP("max-args", "n", 0, "use at most max-args arguments per command line")
	xargsCmd.Flags().StringP("replace", "I", "", "replace occurrences of replace-str in the initial-arguments with names read from standard input")
	xargsCmd.Flags().BoolP("null", "0", false, "input items are terminated by a null character instead of by whitespace")
	xargsCmd.Flags().StringP("delimiter", "d", "", "input items are terminated by the specified character")
	xargsCmd.Flags().BoolP("no-run-if-empty", "r", false, "if the standard input does not contain any nonblanks, do not run the command")
	xargsCmd.Flags().BoolP("verbose", "t", false, "print the command line on the standard error output before executing it")
}
