package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

const stdinName = "standard input"

type input struct {
	name  string
	label string
	index int
	count int
}

type errorMode int

const (
	reportErrors errorMode = iota
	silenceErrors
)

func inputError(mode errorMode, format string, args ...any) {
	if mode == silenceErrors {
		exitStatus = 1
		return
	}
	warn(format, args...)
}

func resolveInputs(cmdName string, args []string, mode errorMode) []input {
	implicit := len(args) == 0
	if implicit {
		args = []string{"-"}
	}

	paths, err := utils.ExpandGlobs(args)
	if err != nil {
		inputError(mode, "%s: %v\n", cmdName, err)
		return nil
	}

	inputs := make([]input, len(paths))
	for i, path := range paths {
		inputs[i] = input{name: path, label: path, index: i, count: len(paths)}
		if path != "-" {
			continue
		}

		inputs[i].name = stdinName
		if implicit {
			inputs[i].label = ""
		}
	}
	return inputs
}

func openInput(cmdName string, in input, mode errorMode) io.ReadCloser {
	if in.name == stdinName {
		return io.NopCloser(os.Stdin)
	}

	f, err := os.Open(in.name)
	if err != nil {
		inputError(mode, "%s: %v\n", cmdName, err)
		return nil
	}
	return f
}

func eachInput(cmdName string, args []string, fn func(in input, r io.Reader)) {
	eachInputMode(cmdName, args, reportErrors, fn)
}

func eachInputMode(cmdName string, args []string, mode errorMode, fn func(in input, r io.Reader)) {
	runInputs(cmdName, resolveInputs(cmdName, args, mode), mode, fn)
}

func runInputs(cmdName string, inputs []input, mode errorMode, fn func(in input, r io.Reader)) {
	for _, in := range inputs {
		r := openInput(cmdName, in, mode)
		if r == nil {
			continue
		}

		fn(in, r)
		r.Close()
	}
}

func readLines(cmdName string, args []string) []string {
	var all []string
	eachInput(cmdName, args, func(in input, r io.Reader) {
		lines, err := utils.ReadLines(r)
		if err != nil {
			warn("%s: %s: %v\n", cmdName, in.name, err)
			return
		}
		all = append(all, lines...)
	})
	return all
}

type headerMode int

const (
	headerAuto headerMode = iota
	headerAlways
	headerNever
)

func headerModeOf(cmd *cobra.Command) headerMode {
	if quiet, _ := cmd.Flags().GetBool("quiet"); quiet {
		return headerNever
	}
	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		return headerAlways
	}
	return headerAuto
}

func printHeader(in input, mode headerMode) {
	if mode == headerNever || (mode == headerAuto && in.count < 2) {
		return
	}

	if in.index > 0 {
		fmt.Fprintln(stdout)
	}
	fmt.Fprintf(stdout, "==> %s <==\n", in.name)
}

func printLine(s *utils.LineScanner) {
	fmt.Fprint(stdout, s.Text())
	if !s.Unterminated() {
		fmt.Fprintln(stdout)
	}
}
