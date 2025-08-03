package cmd

import (
	"bufio"
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
			scanner := bufio.NewScanner(os.Stdin)
			for i := 0; i < lines && scanner.Scan(); i++ {
				fmt.Println(scanner.Text())
			}
		} else {
			// Expand glob patterns in arguments
			expandedArgs, err := utils.ExpandGlobsForReading(args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "head: %v\n", err)
				return
			}
			
			for _, path := range expandedArgs {
				if len(expandedArgs) > 1 {
					fmt.Printf("==> %s <==\n", path)
				}
				
				f, err := os.Open(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "head: %v\n", err)
					continue
				}
				defer f.Close()
				scanner := bufio.NewScanner(f)
				for i := 0; i < lines && scanner.Scan(); i++ {
					fmt.Println(scanner.Text())
				}
				
				if len(expandedArgs) > 1 && path != expandedArgs[len(expandedArgs)-1] {
					fmt.Println()
				}
			}
		}
	},
}

func init() {
	headCmd.Flags().IntP("lines", "n", 10, "number of lines to show from start")
}
