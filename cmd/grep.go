package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

var grepCmd = &cobra.Command{
	Use:   "grep [pattern] [file]",
	Short: "Print lines matching a pattern",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		patternStr := args[0]
		filePath := args[1]

		ignoreCase, _ := cmd.Flags().GetBool("ignore-case")
		invertMatch, _ := cmd.Flags().GetBool("invert-match")
		lineNumber, _ := cmd.Flags().GetBool("line-number")
		regexpFlag, _ := cmd.Flags().GetString("regexp")

		if regexpFlag != "" {
			patternStr = regexpFlag
		}

		if ignoreCase {
			patternStr = "(?i)" + patternStr // Add case-insensitive flag to regex
		}

		re, err := regexp.Compile(patternStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: invalid regex pattern: %v\n", err)
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: %v\n", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			match := re.MatchString(line)

			if (match && !invertMatch) || (!match && invertMatch) {
				if lineNumber {
					fmt.Printf("%d:%s\n", lineNum, line)
				} else {
					fmt.Println(line)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "grep: reading input: %v\n", err)
		}
	},
}

func init() {
	grepCmd.Flags().BoolP("ignore-case", "i", false, "ignore case distinctions")
	grepCmd.Flags().BoolP("invert-match", "v", false, "select non-matching lines")
	grepCmd.Flags().BoolP("line-number", "n", false, "show line numbers")
	grepCmd.Flags().StringP("regexp", "e", "", "use a specific regex pattern")
}
