package cmd

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

var wcCmd = &cobra.Command{
	Use:   "wc [files...]",
	Short: "Print newline, word, byte, and char counts for each file",
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		showLines, _ := cmd.Flags().GetBool("lines")
		showWords, _ := cmd.Flags().GetBool("words")
		showBytes, _ := cmd.Flags().GetBool("bytes")
		showChars, _ := cmd.Flags().GetBool("chars")

		var totalLines, totalWords, totalBytes, totalChars int
		var validFiles []string
		var allCounts []struct {
			lines, words, bytes, chars int
			path                       string
		}

		type fileEntry struct {
			path string
			data []byte
		}
		var entries []fileEntry

		if len(args) == 0 {
			data, err := utils.ReadAllFromReader(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "wc: %v\n", err)
				return
			}
			entries = append(entries, fileEntry{"", []byte(data)})
		} else {
			expandedArgs, err := utils.ExpandGlobsForReading(args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "wc: %v\n", err)
				return
			}
			for _, path := range expandedArgs {
				data, err := os.ReadFile(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "wc: %v\n", err)
					continue
				}
				entries = append(entries, fileEntry{path, data})
			}
		}

		// First pass: collect all counts
		for _, entry := range entries {
			data := entry.data
			path := entry.path
			validFiles = append(validFiles, path)

			content := string(data)
			lines := strings.Count(content, "\n")
			words := len(strings.Fields(content))
			bytes := len(data)

			// Count **non-whitespace characters**
			chars := 0
			for _, r := range content {
				if !unicode.IsSpace(r) {
					chars++
				}
			}

			totalLines += lines
			totalWords += words
			totalBytes += bytes
			totalChars += chars

			allCounts = append(allCounts, struct {
				lines, words, bytes, chars int
				path                       string
			}{lines, words, bytes, chars, path})
		}

		// Determine column widths (for alignment)
		maxLinesWidth := 1
		maxWordsWidth := 1
		maxBytesWidth := 1
		maxCharsWidth := 1

		for _, count := range allCounts {
			if len(fmt.Sprintf("%d", count.lines)) > maxLinesWidth {
				maxLinesWidth = len(fmt.Sprintf("%d", count.lines))
			}
			if len(fmt.Sprintf("%d", count.words)) > maxWordsWidth {
				maxWordsWidth = len(fmt.Sprintf("%d", count.words))
			}
			if len(fmt.Sprintf("%d", count.bytes)) > maxBytesWidth {
				maxBytesWidth = len(fmt.Sprintf("%d", count.bytes))
			}
			if len(fmt.Sprintf("%d", count.chars)) > maxCharsWidth {
				maxCharsWidth = len(fmt.Sprintf("%d", count.chars))
			}
		}

		// Same thing for totals
		if len(fmt.Sprintf("%d", totalLines)) > maxLinesWidth {
			maxLinesWidth = len(fmt.Sprintf("%d", totalLines))
		}
		if len(fmt.Sprintf("%d", totalWords)) > maxWordsWidth {
			maxWordsWidth = len(fmt.Sprintf("%d", totalWords))
		}
		if len(fmt.Sprintf("%d", totalBytes)) > maxBytesWidth {
			maxBytesWidth = len(fmt.Sprintf("%d", totalBytes))
		}
		if len(fmt.Sprintf("%d", totalChars)) > maxCharsWidth {
			maxCharsWidth = len(fmt.Sprintf("%d", totalChars))
		}

		// Print per-file stats
		for _, count := range allCounts {
			out := []string{}
			if showLines {
				out = append(out, fmt.Sprintf("%*d", maxLinesWidth, count.lines))
			}
			if showWords {
				out = append(out, fmt.Sprintf("%*d", maxWordsWidth, count.words))
			}
			if showBytes {
				out = append(out, fmt.Sprintf("%*d", maxBytesWidth, count.bytes))
			}
			if showChars {
				out = append(out, fmt.Sprintf("%*d", maxCharsWidth, count.chars))
			}

			// Default if no flags provided: print all
			suffix := ""
			if count.path != "" {
				suffix = " " + count.path
			}
			if len(out) == 0 {
				fmt.Printf("%*d %*d %*d %*d%s\n",
					maxLinesWidth, count.lines,
					maxWordsWidth, count.words,
					maxBytesWidth, count.bytes,
					maxCharsWidth, count.chars,
					suffix,
				)
			} else {
				fmt.Printf("%s%s\n", strings.Join(out, " "), suffix)
			}
		}

		// Print totals if multiple files
		if len(validFiles) > 1 {
			out := []string{}
			if showLines {
				out = append(out, fmt.Sprintf("%*d", maxLinesWidth, totalLines))
			}
			if showWords {
				out = append(out, fmt.Sprintf("%*d", maxWordsWidth, totalWords))
			}
			if showBytes {
				out = append(out, fmt.Sprintf("%*d", maxBytesWidth, totalBytes))
			}
			if showChars {
				out = append(out, fmt.Sprintf("%*d", maxCharsWidth, totalChars))
			}

			if len(out) == 0 {
				fmt.Printf("%*d %*d %*d %*d total\n",
					maxLinesWidth, totalLines,
					maxWordsWidth, totalWords,
					maxBytesWidth, totalBytes,
					maxCharsWidth, totalChars,
				)
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
	wcCmd.Flags().BoolP("chars", "m", false, "print non-whitespace character count")
}
