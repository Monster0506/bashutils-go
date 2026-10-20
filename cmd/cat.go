package cmd

import (
	"io"
	"os"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

var catCmd = &cobra.Command{
	Use:   "cat [file]",
	Short: "Concatenate and display files",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
				warn("cat: %v\n", err)
			}
			return
		}

		expandedArgs, err := utils.ExpandGlobs(args)
		if err != nil {
			warn("cat: %v\n", err)
			return
		}

		for _, path := range expandedArgs {
			if path == "-" {
				if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
					warn("cat: %v\n", err)
				}
				continue
			}
			data, err := os.ReadFile(path)
			if err != nil {
				warn("cat: %v\n", err)
				continue
			}
			os.Stdout.Write(data)
		}
	},
}
