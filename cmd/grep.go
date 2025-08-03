package cmd

import (
	"bufio"
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

var grepCmd = &cobra.Command{
	Use:   "grep [pattern] [files...]",
	Short: "Print lines matching a pattern",
	Args:  cobra.ArbitraryArgs,
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

		if len(args) < 2 {
			// Read from stdin when no files provided
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				match := re.MatchString(line)
				if (match && !invertMatch) || (!match && invertMatch) {
					if lineNumber {
						fmt.Printf("%d:%s\n", 0, line)
					} else {
						fmt.Println(line)
					}
				}
			}
			return
		}

		// Expand glob patterns in file argument
		expandedFiles, err := utils.ExpandGlobsForReading([]string{filePath})
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: %v\n", err)
			return
		}

		for _, path := range expandedFiles {
			if len(expandedFiles) > 1 {
				fmt.Printf("==> %s <==\n", path)
			}

			file, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "grep: %v\n", err)
				continue
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

			if len(expandedFiles) > 1 && path != expandedFiles[len(expandedFiles)-1] {
				fmt.Println()
			}
		}
	},
}

func init() {
	grepCmd.Flags().BoolP("ignore-case", "i", false, "ignore case distinctions")
	grepCmd.Flags().BoolP("invert-match", "v", false, "select non-matching lines")
	grepCmd.Flags().BoolP("line-number", "n", false, "show line numbers")
	grepCmd.Flags().StringP("regexp", "e", "", "use a specific regex pattern")
}
