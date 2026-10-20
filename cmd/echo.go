package cmd

import (
	"fmt"
	"strconv"
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
			var stop bool
			out, stop = interpretEscapes(out)
			if stop {
				opts.suppressNewline = true
			}
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

var echoSimpleEscapes = map[byte]byte{
	'\\': '\\',
	'a':  '\a',
	'b':  '\b',
	'e':  0x1b,
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
}

func interpretEscapes(s string) (string, bool) {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 == len(s) {
			sb.WriteByte(s[i])
			continue
		}

		next := s[i+1]
		switch {
		case next == 'c':
			return sb.String(), true
		case next == '0':
			b, used := parseNumber(s[i+2:], 8, 3)
			sb.WriteByte(b)
			i += 1 + used
		case next == 'x':
			b, used := parseNumber(s[i+2:], 16, 2)
			if used == 0 {
				sb.WriteString(`\x`)
			} else {
				sb.WriteByte(b)
			}
			i += 1 + used
		default:
			if esc, ok := echoSimpleEscapes[next]; ok {
				sb.WriteByte(esc)
			} else {
				sb.WriteByte('\\')
				sb.WriteByte(next)
			}
			i++
		}
	}
	return sb.String(), false
}

func parseNumber(s string, base, maxDigits int) (byte, int) {
	used := 0
	value := 0
	for used < len(s) && used < maxDigits {
		d, err := strconv.ParseInt(s[used:used+1], base, 32)
		if err != nil {
			break
		}
		value = value*base + int(d)
		used++
	}
	return byte(value), used
}

func isEchoFlags(a string) bool {
	if len(a) < 2 || a[0] != '-' {
		return false
	}
	return strings.Trim(a[1:], "neE") == ""
}
