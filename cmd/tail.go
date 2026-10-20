package cmd

import (
	"fmt"
	"io"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

var tailCmd = &cobra.Command{
	Use:   "tail [files...]",
	Short: "Output the last part of files",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		n, _ := cmd.Flags().GetInt("lines")

		eachInput("tail", args, func(in input, r io.Reader) {
			printHeader(in)

			scanner := utils.NewScanner(r)
			lines := []string{}
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
				if len(lines) > n {
					lines = lines[1:]
				}
			}

			if err := scanner.Err(); err != nil {
				warn("tail: %s: %v\n", in.name, err)
				return
			}

			for i, line := range lines {
				if i == len(lines)-1 && scanner.Unterminated() {
					fmt.Print(line)
					continue
				}
				fmt.Println(line)
			}
		})
	},
}

func init() {
	tailCmd.Flags().IntP("lines", "n", 10, "number of lines to show from end")
}
