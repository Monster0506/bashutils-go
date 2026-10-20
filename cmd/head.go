package cmd

import (
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
	"os"
)

var headCmd = &cobra.Command{
	Use:   "head [files...]",
	Short: "Output the first part of files",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		lines, _ := cmd.Flags().GetInt("lines")

		if len(args) == 0 {
			// Read from stdin when no files provided
			scanner := utils.NewScanner(os.Stdin)
			for i := 0; i < lines && scanner.Scan(); i++ {
				fmt.Println(scanner.Text())
			}
		} else {
			// Expand glob patterns in arguments
			expandedArgs, err := utils.ExpandGlobs(args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "head: %v\n", err)
				return
			}

			for i, path := range expandedArgs {
				if len(expandedArgs) > 1 {
					fmt.Printf("==> %s <==\n", path)
				}

				f, err := os.Open(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "head: %v\n", err)
					continue
				}
				defer f.Close()
				scanner := utils.NewScanner(f)
				for i := 0; i < lines && scanner.Scan(); i++ {
					fmt.Println(scanner.Text())
				}

				if i < len(expandedArgs)-1 {
					fmt.Println()
				}
			}
		}
	},
}

func init() {
	headCmd.Flags().IntP("lines", "n", 10, "number of lines to show from start")
}
