package cmd

import (
	"bufio"
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var splitCmd = &cobra.Command{
	Use:   "split [file] [prefix]",
	Short: "Split files into pieces",
	Args:  cobra.RangeArgs(1, 2), // file and optional prefix
	Run: func(cmd *cobra.Command, args []string) {
		linesPerFile, _ := cmd.Flags().GetInt64("lines")
		bytesPerFile, _ := cmd.Flags().GetString("bytes")
		numericSuffixes, _ := cmd.Flags().GetBool("numeric-suffixes")

		filePath := args[0]
		prefix := "x"
		if len(args) > 1 {
			prefix = args[1]
		}

		if linesPerFile == 0 && bytesPerFile == "" {
			linesPerFile = 1000
		}
		if linesPerFile > 0 && bytesPerFile != "" {
			fmt.Fprintf(os.Stderr, "split: cannot split by lines and bytes simultaneously\n")
			return
		}

		var inputFile *os.File
		if filePath == "-" {
			inputFile = os.Stdin
		} else {
			expandedFiles, err := utils.ExpandGlobsForReading([]string{filePath})
			if err != nil {
				fmt.Fprintf(os.Stderr, "split: %v\n", err)
				return
			}
			if len(expandedFiles) == 0 {
				fmt.Fprintf(os.Stderr, "split: no matching files found\n")
				return
			}
			f, err := os.Open(expandedFiles[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "split: %v\n", err)
				return
			}
			defer f.Close()
			inputFile = f
		}

		if linesPerFile > 0 {
			splitByLines(inputFile, prefix, linesPerFile, numericSuffixes)
		} else if bytesPerFile != "" {
			splitByBytes(inputFile, prefix, bytesPerFile, numericSuffixes)
		}
	},
}

func init() {
	splitCmd.Flags().Int64P("lines", "l", 0, "split by number of lines")
	splitCmd.Flags().StringP("bytes", "b", "", "split by number of bytes (e.g., '1K', '1M')")
	splitCmd.Flags().BoolP("numeric-suffixes", "d", false, "use numeric suffixes instead of alphabetic")
}

func generateSuffix(index int, numeric bool) string {
	if numeric {
		return fmt.Sprintf("%02d", index) // Pad with leading zeros for consistency
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	const minWidth = 2
	suf := ""
	for ; index > 0 || len(suf) < minWidth; index /= len(alphabet) {
		suf = string(alphabet[index%len(alphabet)]) + suf
	}
	return suf
}

func splitByLines(inputFile *os.File, prefix string, linesPerFile int64, numericSuffixes bool) {
	scanner := bufio.NewScanner(inputFile)
	fileIndex := 0
	currentLineCount := int64(0)
	var outputFile *os.File
	var outputWriter *bufio.Writer
	var err error

	for scanner.Scan() {
		if currentLineCount == 0 {
			if outputFile != nil {
				outputWriter.Flush()
				outputFile.Close()
			}
			suffix := generateSuffix(fileIndex, numericSuffixes)
			outputFileName := prefix + suffix
			outputFile, err = os.Create(outputFileName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "split: creating output file: %v\n", err)
				return
			}
			outputWriter = bufio.NewWriter(outputFile)
			fileIndex++
		}

		_, err = outputWriter.WriteString(scanner.Text() + "\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "split: writing to output file: %v\n", err)
			return
		}
		currentLineCount++

		if currentLineCount >= linesPerFile {
			currentLineCount = 0 // Reset for the next file
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "split: reading input: %v\n", err)
	}
	if outputFile != nil {
		outputWriter.Flush()
		outputFile.Close()
	}
}

func parseBytesString(bytesStr string) (int64, error) {
	bytesStr = strings.TrimSpace(bytesStr)
	lastChar := ' '
	if len(bytesStr) > 0 {
		lastChar = rune(bytesStr[len(bytesStr)-1])
	}

	multiplier := int64(1)
	valueStr := bytesStr

	switch lastChar {
	case 'k', 'K':
		multiplier = 1024
		valueStr = bytesStr[:len(bytesStr)-1]
	case 'm', 'M':
		multiplier = 1024 * 1024
		valueStr = bytesStr[:len(bytesStr)-1]
	case 'g', 'G':
		multiplier = 1024 * 1024 * 1024
		valueStr = bytesStr[:len(bytesStr)-1]
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid byte size format: %s", bytesStr)
	}
	if value <= 0 {
		return 0, fmt.Errorf("invalid byte size: %s", bytesStr)
	}
	return value * multiplier, nil
}

func splitByBytes(inputFile *os.File, prefix string, bytesPerFileStr string, numericSuffixes bool) {
	bytesPerFile, err := parseBytesString(bytesPerFileStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "split: %v\n", err)
		return
	}

	reader := bufio.NewReader(inputFile)
	for fileIndex := 0; ; fileIndex++ {
		if _, err := reader.Peek(1); err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "split: reading input: %v\n", err)
			}
			return
		}

		outputFile, err := os.Create(prefix + generateSuffix(fileIndex, numericSuffixes))
		if err != nil {
			fmt.Fprintf(os.Stderr, "split: creating output file: %v\n", err)
			return
		}

		_, copyErr := io.CopyN(outputFile, reader, bytesPerFile)
		if err := outputFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "split: closing output file: %v\n", err)
			return
		}

		if copyErr == io.EOF {
			return
		}
		if copyErr != nil {
			fmt.Fprintf(os.Stderr, "split: copying to output file: %v\n", copyErr)
			return
		}
	}
}
