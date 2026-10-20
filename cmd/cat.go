package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

var catCmd = &cobra.Command{
	Use:   "cat [file]",
	Short: "Concatenate and display files",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		eachInput("cat", args, func(in input, r io.Reader) {
			if _, err := io.Copy(os.Stdout, r); err != nil {
				warn("cat: %s: %v\n", in.name, err)
			}
		})
	},
}
