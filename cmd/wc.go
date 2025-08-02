package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var wcCmd = &cobra.Command{
	Use:   "wc [file]",
	Short: "Print newline, word, and byte counts for each file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showLines, _ := cmd.Flags().GetBool("lines")
		showWords, _ := cmd.Flags().GetBool("words")
		showBytes, _ := cmd.Flags().GetBool("bytes")

		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "wc: %v\n", err)
			return
		}
		content := string(data)
		lines := strings.Count(content, "\n")
		words := len(strings.Fields(content))
		bytes := len(data)

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
			fmt.Printf("%d %d %d %s\n", lines, words, bytes, args[0])
		} else {
			fmt.Printf("%s %s\n", strings.Join(out, " "), args[0])
		}
	},
}

func init() {
	wcCmd.Flags().BoolP("lines", "l", false, "print newline count")
	wcCmd.Flags().BoolP("words", "w", false, "print word count")
	wcCmd.Flags().BoolP("bytes", "c", false, "print byte count")
}
