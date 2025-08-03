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
		var allCounts []struct {
			lines, words, bytes int
			path                string
		}

		// First pass: collect all counts to determine column widths
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

			allCounts = append(allCounts, struct {
				lines, words, bytes int
				path                string
			}{lines, words, bytes, path})
		}

		// Determine column widths
		maxLinesWidth := 1
		maxWordsWidth := 1
		maxBytesWidth := 1
		maxTotalLinesWidth := 1
		maxTotalWordsWidth := 1
		maxTotalBytesWidth := 1

		for _, count := range allCounts {
			linesStr := fmt.Sprintf("%d", count.lines)
			wordsStr := fmt.Sprintf("%d", count.words)
			bytesStr := fmt.Sprintf("%d", count.bytes)
			
			if len(linesStr) > maxLinesWidth {
				maxLinesWidth = len(linesStr)
			}
			if len(wordsStr) > maxWordsWidth {
				maxWordsWidth = len(wordsStr)
			}
			if len(bytesStr) > maxBytesWidth {
				maxBytesWidth = len(bytesStr)
			}
		}

		totalLinesStr := fmt.Sprintf("%d", totalLines)
		totalWordsStr := fmt.Sprintf("%d", totalWords)
		totalBytesStr := fmt.Sprintf("%d", totalBytes)
		
		if len(totalLinesStr) > maxTotalLinesWidth {
			maxTotalLinesWidth = len(totalLinesStr)
		}
		if len(totalWordsStr) > maxTotalWordsWidth {
			maxTotalWordsWidth = len(totalWordsStr)
		}
		if len(totalBytesStr) > maxTotalBytesWidth {
			maxTotalBytesWidth = len(totalBytesStr)
		}

		linesWidth := maxLinesWidth
		if maxTotalLinesWidth > linesWidth {
			linesWidth = maxTotalLinesWidth
		}
		wordsWidth := maxWordsWidth
		if maxTotalWordsWidth > wordsWidth {
			wordsWidth = maxTotalWordsWidth
		}
		bytesWidth := maxBytesWidth
		if maxTotalBytesWidth > bytesWidth {
			bytesWidth = maxTotalBytesWidth
		}

		for _, count := range allCounts {
			out := []string{}
			if showLines {
				out = append(out, fmt.Sprintf("%*d", linesWidth, count.lines))
			}
			if showWords {
				out = append(out, fmt.Sprintf("%*d", wordsWidth, count.words))
			}
			if showBytes {
				out = append(out, fmt.Sprintf("%*d", bytesWidth, count.bytes))
			}

			if len(out) == 0 {
				fmt.Printf("%*d %*d %*d %s\n", linesWidth, count.lines, wordsWidth, count.words, bytesWidth, count.bytes, count.path)
			} else {
				fmt.Printf("%s %s\n", strings.Join(out, " "), count.path)
			}
		}

		if len(validFiles) > 1 {
			out := []string{}
			if showLines {
				out = append(out, fmt.Sprintf("%*d", linesWidth, totalLines))
			}
			if showWords {
				out = append(out, fmt.Sprintf("%*d", wordsWidth, totalWords))
			}
			if showBytes {
				out = append(out, fmt.Sprintf("%*d", bytesWidth, totalBytes))
			}

			if len(out) == 0 {
				fmt.Printf("%*d %*d %*d total\n", linesWidth, totalLines, wordsWidth, totalWords, bytesWidth, totalBytes)
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
