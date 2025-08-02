package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var tailCmd = &cobra.Command{
	Use:   "tail [file]",
	Short: "Output the last part of files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		n, _ := cmd.Flags().GetInt("lines")
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "tail: %v\n", err)
			return
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
	},
}

func init() {
	tailCmd.Flags().IntP("lines", "n", 10, "number of lines to show from end")
}
