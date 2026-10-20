package cmd

import (
	"fmt"
	"io"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

var headCmd = &cobra.Command{
	Use:   "head [files...]",
	Short: "Output the first part of files",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		lines, _ := cmd.Flags().GetInt("lines")

		eachInput("head", args, func(in input, r io.Reader) {
			printHeader(in)

			scanner := utils.NewScanner(r)
			for i := 0; i < lines && scanner.Scan(); i++ {
				fmt.Println(scanner.Text())
			}

			if err := scanner.Err(); err != nil {
				warn("head: %s: %v\n", in.name, err)
			}
		})
	},
}

func init() {
	headCmd.Flags().IntP("lines", "n", 10, "number of lines to show from start")
}
