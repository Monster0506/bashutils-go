package cmd

import (
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var pasteCmd = &cobra.Command{
	Use:   "paste [files...]",
	Short: "Merge lines from files",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		delimitersStr, _ := cmd.Flags().GetString("delimiters")
		serial, _ := cmd.Flags().GetBool("serial")

		delimiters := []string{"\t"}
		if delimitersStr != "" {
			delimiters = parseDelimiters(delimitersStr)
		}

		if serial {
			warn("paste: --serial flag is not yet implemented.\n")
			os.Exit(1)
			return
		}

		if len(args) == 0 {
			args = []string{"-"}
		}

		// Expand glob patterns, but preserve "-" entries
		var expandedArgs []string
		for _, arg := range args {
			if arg == "-" {
				expandedArgs = append(expandedArgs, "-")
				continue
			}
			expanded, err := utils.ExpandGlobs([]string{arg})
			if err != nil {
				warn("paste: %v\n", err)
				return
			}
			expandedArgs = append(expandedArgs, expanded...)
		}

		files := make([]*os.File, len(expandedArgs))
		scanners := make([]*utils.LineScanner, len(expandedArgs))
		stdinScanner := utils.NewScanner(os.Stdin)
		for i, filePath := range expandedArgs {
			if filePath == "-" {
				files[i] = nil
				scanners[i] = stdinScanner
				continue
			}
			file, err := os.Open(filePath)
			if err != nil {
				warn("paste: %v\n", err)
				return
			}
			files[i] = file
			scanners[i] = utils.NewScanner(file)
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
						warn("paste: reading input: %v\n", err)
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
						outputBuilder.WriteString(delimiters[i%len(delimiters)])
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

var delimiterEscapes = map[rune]string{
	'n':  "\n",
	't':  "\t",
	'0':  "",
	'\\': "\\",
}

func parseDelimiters(s string) []string {
	var delims []string
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] != '\\' || i+1 == len(runes) {
			delims = append(delims, string(runes[i]))
			continue
		}

		i++
		if d, ok := delimiterEscapes[runes[i]]; ok {
			delims = append(delims, d)
			continue
		}
		delims = append(delims, string(runes[i]))
	}
	return delims
}
