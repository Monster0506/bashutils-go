package cmd

import (
	"fmt"
	"strings"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

var echoCmd = &cobra.Command{
	Use:   "echo [strings...]",
	Short: "Echo arguments to standard output",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		suppressNewline, _ := cmd.Flags().GetBool("newline")
		enableEscape, _ := cmd.Flags().GetBool("escape")
		expandEnv, _ := cmd.Flags().GetBool("expand-env")

		out := strings.Join(args, " ")
		
		if expandEnv {
			out = utils.ExpandEnvironmentVariables(out)
		}
		
		if enableEscape {
			out = strings.ReplaceAll(out, "\\n", "\n")
			out = strings.ReplaceAll(out, "\\t", "\t")
		}

		if suppressNewline {
			fmt.Print(out)
		} else {
			fmt.Println(out)
		}
	},
}

func init() {
	echoCmd.Flags().BoolP("newline", "n", false, "do not output the trailing newline")
	echoCmd.Flags().BoolP("escape", "e", false, "enable interpretation of backslash escapes")
	echoCmd.Flags().BoolP("expand-env", "E", false, "expand environment variables ($VAR and %VAR%)")
}
