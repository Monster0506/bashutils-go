package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var headCmd = &cobra.Command{
	Use:   "head [file]",
	Short: "Output the first part of files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lines, _ := cmd.Flags().GetInt("lines")
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "head: %v\n", err)
			return
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for i := 0; i < lines && scanner.Scan(); i++ {
			fmt.Println(scanner.Text())
		}
	},
}

func init() {
	headCmd.Flags().IntP("lines", "n", 10, "number of lines to show from start")
}
