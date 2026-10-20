package cmd

import (
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
	"os"
)

var tailCmd = &cobra.Command{
	Use:   "tail [files...]",
	Short: "Output the last part of files",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		n, _ := cmd.Flags().GetInt("lines")

		if len(args) == 0 {
			// Read from stdin when no files provided
			scanner := utils.NewScanner(os.Stdin)
			lines := []string{}
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
				if len(lines) > n {
					lines = lines[1:]
				}
			}
			for _, line := range lines {
				fmt.Println(line)
			}
		} else {
			// Expand glob patterns in arguments
			expandedArgs, err := utils.ExpandGlobsForReading(args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "tail: %v\n", err)
				return
			}

			for i, path := range expandedArgs {
				if len(expandedArgs) > 1 {
					fmt.Printf("==> %s <==\n", path)
				}

				f, err := os.Open(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "tail: %v\n", err)
					continue
				}
				defer f.Close()
				scanner := utils.NewScanner(f)
				lines := []string{}
				for scanner.Scan() {
					lines = append(lines, scanner.Text())
					if len(lines) > n {
						lines = lines[1:]
					}
				}
				for _, line := range lines {
					fmt.Println(line)
				}

				if i < len(expandedArgs)-1 {
					fmt.Println()
				}
			}
		}
	},
}

func init() {
	tailCmd.Flags().IntP("lines", "n", 10, "number of lines to show from end")
}
