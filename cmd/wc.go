package cmd

import (
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var wcCmd = &cobra.Command{
	Use:   "wc [files...]",
	Short: "Print newline, word, and byte counts for each file",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showLines, _ := cmd.Flags().GetBool("lines")
		showWords, _ := cmd.Flags().GetBool("words")
		showBytes, _ := cmd.Flags().GetBool("bytes")

		// Expand glob patterns in arguments
		expandedArgs, err := utils.ExpandGlobsForReading(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "wc: %v\n", err)
			return
		}

		var totalLines, totalWords, totalBytes int
		var validFiles []string

		for _, path := range expandedArgs {
			data, err := os.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "wc: %v\n", err)
				continue
			}
			validFiles = append(validFiles, path)
			
			content := string(data)
			lines := strings.Count(content, "\n")
			words := len(strings.Fields(content))
			bytes := len(data)

			totalLines += lines
			totalWords += words
			totalBytes += bytes

			out := []string{}
			if showLines {
				out = append(out, fmt.Sprintf("%d", lines))
			}
			if showWords {
				out = append(out, fmt.Sprintf("%d", words))
			}
			if showBytes {
				out = append(out, fmt.Sprintf("%d", bytes))
			}

			if len(out) == 0 {
				fmt.Printf("%d %d %d %s\n", lines, words, bytes, path)
			} else {
				fmt.Printf("%s %s\n", strings.Join(out, " "), path)
			}
		}

		if len(validFiles) > 1 {
			out := []string{}
			if showLines {
				out = append(out, fmt.Sprintf("%d", totalLines))
			}
			if showWords {
				out = append(out, fmt.Sprintf("%d", totalWords))
			}
			if showBytes {
				out = append(out, fmt.Sprintf("%d", totalBytes))
			}

			if len(out) == 0 {
				fmt.Printf("%d %d %d total\n", totalLines, totalWords, totalBytes)
			} else {
				fmt.Printf("%s total\n", strings.Join(out, " "))
			}
		}
	},
}

func init() {
	wcCmd.Flags().BoolP("lines", "l", false, "print newline count")
	wcCmd.Flags().BoolP("words", "w", false, "print word count")
	wcCmd.Flags().BoolP("bytes", "c", false, "print byte count")
}
