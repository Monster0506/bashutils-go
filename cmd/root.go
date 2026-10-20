package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bashutils",
	Short: "A Go-based reimplementation of bash coreutils",
	Long:  `bashutils is a CLI tool written in Go that mimics common bash coreutils commands.`,
	Args:  cobra.MinimumNArgs(1),
}

var exitStatus int

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *exitError) Unwrap() error {
	return e.err
}

func fail(code int, format string, args ...any) error {
	return &exitError{code: code, err: fmt.Errorf(format, args...)}
}

func exitWith(code int) error {
	return &exitError{code: code}
}

func warn(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	exitStatus = 1
}

func Execute() {
	rootCmd.SetArgs(normalizeArgs(invocationArgs()))
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	os.Exit(exitCode(rootCmd.Execute()))
}

func exitCode(err error) int {
	if err == nil {
		return exitStatus
	}

	var ee *exitError
	if !errors.As(err, &ee) {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}

	if ee.err != nil {
		fmt.Fprintln(os.Stderr, ee.err)
	}
	return ee.code
}

func invocationArgs() []string {
	base := filepath.Base(os.Args[0])
	name := strings.ToLower(strings.TrimSuffix(base, filepath.Ext(base)))

	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			return append([]string{name}, os.Args[1:]...)
		}
	}
	return os.Args[1:]
}

var numericOptionCmds = map[string]bool{"head": true, "tail": true}

func normalizeArgs(args []string) []string {
	if len(args) == 0 || !numericOptionCmds[args[0]] {
		return args
	}

	out := []string{args[0]}
	expectValue := false
	for i := 1; i < len(args); i++ {
		a := args[i]
		switch {
		case expectValue:
			expectValue = false
		case a == "--" || !strings.HasPrefix(a, "-") || a == "-":
			return append(out, args[i:]...)
		case isNumericOption(a):
			out = append(out, "-n", a[1:])
			continue
		default:
			expectValue = a == "-n" || a == "-c"
		}
		out = append(out, a)
	}
	return out
}

func isNumericOption(a string) bool {
	if len(a) < 2 || a[0] != '-' {
		return false
	}
	return strings.Trim(a[1:], "0123456789") == ""
}

func init() {
	rootCmd.AddCommand(echoCmd)
	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(headCmd)
	rootCmd.AddCommand(tailCmd)
	rootCmd.AddCommand(wcCmd)
	rootCmd.AddCommand(cutCmd)
	rootCmd.AddCommand(sortCmd)
	rootCmd.AddCommand(uniqCmd)
	rootCmd.AddCommand(grepCmd)
	rootCmd.AddCommand(trCmd)
	rootCmd.AddCommand(pasteCmd)
	rootCmd.AddCommand(splitCmd)
	rootCmd.AddCommand(xargsCmd)
	rootCmd.AddCommand(calCmd)
	rootCmd.AddCommand(installCmd)
}
