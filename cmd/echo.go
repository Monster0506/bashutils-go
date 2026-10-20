package cmd

import (
	"fmt"
	"strings"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

type echoOpts struct {
	suppressNewline bool
	escapes         bool
	expandEnv       bool
}

var echoCmd = &cobra.Command{
	Use:   "echo [strings...]",
	Short: "Echo arguments to standard output",
	Long: `Echo arguments to standard output.

  -n           do not output the trailing newline
  -e           enable interpretation of backslash escapes
  -E           disable interpretation of backslash escapes (default)
  --expand-env expand environment variables ($VAR and %VAR%)`,
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "--help" {
			return cmd.Help()
		}

		opts, args := parseEchoArgs(args)
		out := strings.Join(args, " ")

		if opts.expandEnv {
			out = utils.ExpandEnvironmentVariables(out)
		}

		if opts.escapes {
			out = strings.ReplaceAll(out, "\\n", "\n")
			out = strings.ReplaceAll(out, "\\t", "\t")
		}

		if opts.suppressNewline {
			fmt.Print(out)
		} else {
			fmt.Println(out)
		}
		return nil
	},
}

func parseEchoArgs(args []string) (echoOpts, []string) {
	var o echoOpts
	for len(args) > 0 {
		a := args[0]
		if a == "--expand-env" {
			o.expandEnv = true
			args = args[1:]
			continue
		}
		if !isEchoFlags(a) {
			break
		}

		for _, letter := range a[1:] {
			switch letter {
			case 'n':
				o.suppressNewline = true
			case 'e':
				o.escapes = true
			case 'E':
				o.escapes = false
			}
		}
		args = args[1:]
	}
	return o, args
}

func isEchoFlags(a string) bool {
	if len(a) < 2 || a[0] != '-' {
		return false
	}
	return strings.Trim(a[1:], "neE") == ""
}
