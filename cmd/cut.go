package cmd

import (
	"bufio"
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var cutCmd = &cobra.Command{
	Use:   "cut [files...]",
	Short: "Extract specific columns or byte ranges from lines",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fields, _ := cmd.Flags().GetString("fields")
		delimiter, _ := cmd.Flags().GetString("delimiter")
		characters, _ := cmd.Flags().GetString("characters")

		if (fields == "" && characters == "") || (fields != "" && characters != "") {
			fmt.Fprintf(os.Stderr, "cut: specify either --fields or --characters\n")
			return
		}

		if len(args) == 0 {
			// Read from stdin when no files provided
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				if fields != "" {
					printFields(line, delimiter, fields)
				} else if characters != "" {
					printCharacters(line, characters)
				}
			}
		} else {
			// Expand glob patterns in file argument
			expandedFiles, err := utils.ExpandGlobsForReading(args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cut: %v\n", err)
				return
			}

			for _, path := range expandedFiles {
				file, err := os.Open(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "cut: %v\n", err)
					continue
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					if fields != "" {
						printFields(line, delimiter, fields)
					} else if characters != "" {
						printCharacters(line, characters)
					}
				}

				if err := scanner.Err(); err != nil {
					fmt.Fprintf(os.Stderr, "cut: reading input: %v\n", err)
				}
			}
		}
	},
}

func init() {
	cutCmd.Flags().StringP("fields", "f", "", "select fields by delimiter (e.g. '1,3')")
	cutCmd.Flags().StringP("delimiter", "d", "\t", "specify delimiter (default is TAB)")
	cutCmd.Flags().StringP("characters", "c", "", "select character positions (e.g. '1-5,7')")
}

func parseRanges(input string) ([]int, error) {
	var indices []int
	parts := strings.Split(input, ",")
	for _, part := range parts {
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, err
			}
			end := start
			if len(rangeParts) > 1 {
				end, err = strconv.Atoi(rangeParts[1])
				if err != nil {
					return nil, err
				}
			}
			for i := start; i <= end; i++ {
				indices = append(indices, i)
			}
		} else {
			idx, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
			indices = append(indices, idx)
		}
	}
	return indices, nil
}

func printFields(line, delimiter, fields string) {
	parts := strings.Split(line, delimiter)
	fieldIndices, err := parseRanges(fields)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cut: invalid field list: %v\n", err)
		return
	}

	var selectedFields []string
	for _, idx := range fieldIndices {
		if idx > 0 && idx <= len(parts) {
			selectedFields = append(selectedFields, parts[idx-1]) // 1-based index
		}
	}
	fmt.Println(strings.Join(selectedFields, delimiter))
}

func printCharacters(line, characters string) {
	charIndices, err := parseRanges(characters)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cut: invalid character list: %v\n", err)
		return
	}

	runes := []rune(line)
	var selectedChars []rune
	for _, idx := range charIndices {
		if idx > 0 && idx <= len(runes) {
			selectedChars = append(selectedChars, runes[idx-1]) // 1-based index
		}
	}
	fmt.Println(string(selectedChars))
}
