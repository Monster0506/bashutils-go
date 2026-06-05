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
		ignoreCase, _ := cmd.Flags().GetBool("ignore-case")
		invertMatch, _ := cmd.Flags().GetBool("invert-match")
		lineNumber, _ := cmd.Flags().GetBool("line-number")
		regexpFlag, _ := cmd.Flags().GetString("regexp")

		var patternStr string
		var fileArgs []string

		if regexpFlag != "" {
			patternStr = regexpFlag
			fileArgs = args
		} else {
			if len(args) == 0 {
				fmt.Fprintf(os.Stderr, "grep: missing pattern\n")
				os.Exit(2)
			}
			patternStr = args[0]
			fileArgs = args[1:]
		}

		if ignoreCase {
			patternStr = "(?i)" + patternStr
		}

		re, err := regexp.Compile(patternStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: invalid regex pattern: %v\n", err)
			os.Exit(2)
		}

		matched := false
		exitErr := false

		printLine := func(lineNum int, line string) {
			if lineNumber {
				fmt.Printf("%d:%s\n", lineNum, line)
			} else {
				fmt.Println(line)
			}
		}

		if len(fileArgs) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				line := scanner.Text()
				match := re.MatchString(line)
				if (match && !invertMatch) || (!match && invertMatch) {
					matched = true
					printLine(lineNum, line)
				}
			}
		} else {
			expandedFiles, err := utils.ExpandGlobs(fileArgs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "grep: %v\n", err)
				os.Exit(2)
			}

			for _, path := range expandedFiles {
				if len(expandedFiles) > 1 {
					fmt.Printf("==> %s <==\n", path)
				}

				file, err := os.Open(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "grep: %v\n", err)
					exitErr = true
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
						matched = true
						printLine(lineNum, line)
					}
				}

				if err := scanner.Err(); err != nil {
					fmt.Fprintf(os.Stderr, "grep: reading input: %v\n", err)
					exitErr = true
				}

				if len(expandedFiles) > 1 && path != expandedFiles[len(expandedFiles)-1] {
					fmt.Println()
				}
			}
		}

		if exitErr {
			os.Exit(2)
		}
		if !matched {
			os.Exit(1)
		}
	},
}

func init() {
	grepCmd.Flags().BoolP("ignore-case", "i", false, "ignore case distinctions")
	grepCmd.Flags().BoolP("invert-match", "v", false, "select non-matching lines")
	grepCmd.Flags().BoolP("line-number", "n", false, "show line numbers")
	grepCmd.Flags().StringP("regexp", "e", "", "use a specific regex pattern")
}
