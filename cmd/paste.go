package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var pasteCmd = &cobra.Command{
	Use:   "paste [files...]",
	Short: "Merge lines from files",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		delimitersStr, _ := cmd.Flags().GetString("delimiters")
		serial, _ := cmd.Flags().GetBool("serial")

		delimiters := []rune{'\t'} // Default delimiter
		if delimitersStr != "" {
			delimiters = []rune(delimitersStr)
		}

		if serial {
			fmt.Fprintf(os.Stderr, "paste: --serial flag is not yet implemented.\n")
			os.Exit(1)
			return
		}

		files := make([]*os.File, len(args))
		scanners := make([]*bufio.Scanner, len(args))
		for i, filePath := range args {
			file, err := os.Open(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "paste: %v\n", err)
				return
			}
			files[i] = file
			scanners[i] = bufio.NewScanner(file)
		}
		defer func() {
			for _, file := range files {
				if file != nil {
					file.Close()
				}
			}
		}()

		var currentLines []string
		moreData := true
		for moreData {
			currentLines = make([]string, len(scanners))
			moreData = false
			for i, scanner := range scanners {
				if scanner.Scan() {
					currentLines[i] = scanner.Text()
					moreData = true
				} else {
					if err := scanner.Err(); err != nil {
						fmt.Fprintf(os.Stderr, "paste: reading input: %v\n", err)
						return
					}
					currentLines[i] = "" // Pad with empty string if file finished
				}
			}
			if moreData {
				var outputBuilder strings.Builder
				for i, line := range currentLines {
					outputBuilder.WriteString(line)
					if i < len(currentLines)-1 {
						outputBuilder.WriteRune(delimiters[i%len(delimiters)]) // Cycle through delimiters
					}
				}
				fmt.Println(outputBuilder.String())
			}
		}
	},
}

func init() {
	pasteCmd.Flags().StringP("delimiters", "d", "", "use specified delimiters instead of TAB")
	pasteCmd.Flags().BoolP("serial", "s", false, "paste one file at a time instead of in parallel (not yet implemented)")
}
