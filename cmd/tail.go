package cmd

import (
	"bufio"
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
	"os"
)

var tailCmd = &cobra.Command{
	Use:   "tail [file]",
	Short: "Output the last part of files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		n, _ := cmd.Flags().GetInt("lines")
		
		// Expand glob patterns in arguments
		expandedArgs, err := utils.ExpandGlobsForReading(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tail: %v\n", err)
			return
		}
		
		for _, path := range expandedArgs {
			if len(expandedArgs) > 1 {
				fmt.Printf("==> %s <==\n", path)
			}
			
			f, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "tail: %v\n", err)
				continue
			}
			defer f.Close()
			scanner := bufio.NewScanner(f)
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
			
			if len(expandedArgs) > 1 && path != expandedArgs[len(expandedArgs)-1] {
				fmt.Println()
			}
		}
	},
}

func init() {
	tailCmd.Flags().IntP("lines", "n", 10, "number of lines to show from end")
}
